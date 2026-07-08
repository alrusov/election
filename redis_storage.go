package election

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	rdb redis.UniversalClient
}

func NewRedisStorage(rdb redis.UniversalClient) *RedisStorage {
	return &RedisStorage{rdb: rdb}
}

func lockKey(r string) string  { return "election:" + r + ":lock" }
func tokenKey(r string) string { return "election:" + r + ":token" }

var electionScript = redis.NewScript(`
local owner = redis.call("GET", KEYS[1])
if not owner then
  local token = redis.call("INCR", KEYS[2])
  redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
  return token
end
if owner == ARGV[1] then
  redis.call("PEXPIRE", KEYS[1], ARGV[2])
  return redis.call("GET", KEYS[2])
end
return 0
`)

func (r *RedisStorage) TryAcquireOrRenew(ctx context.Context, resource, nodeID string, ttlMs int64) (int64, bool, error) {
	res, err := electionScript.Run(
		ctx,
		r.rdb,
		[]string{lockKey(resource), tokenKey(resource)},
		nodeID,
		ttlMs,
	).Int64()

	if err != nil {
		return 0, false, err
	}
	if res > 0 {
		return res, true, nil
	}
	return 0, false, nil
}

func (r *RedisStorage) ForceReelection(ctx context.Context, resource string) error {
	return r.rdb.Del(ctx, lockKey(resource)).Err()
}

func (r *RedisStorage) Resign(ctx context.Context, resource, nodeID string) error {
	owner, err := r.rdb.Get(ctx, lockKey(resource)).Result()
	if err != nil {
		return err
	}
	if owner == nodeID {
		return r.rdb.Del(ctx, lockKey(resource)).Err()
	}
	return fmt.Errorf("not owner")
}

func (r *RedisStorage) GetOwner(ctx context.Context, resource string) (string, error) {
	return r.rdb.Get(ctx, lockKey(resource)).Result()
}
