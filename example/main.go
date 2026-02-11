package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alrusov/election"

	"github.com/redis/go-redis/v9"
)

func main() {
	//rdb := redis.NewFailoverClient(&redis.FailoverOptions{
	//	MasterName:    "mymaster",
	//	SentinelAddrs: []string{"localhost:26379"},
	//})

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	storage := election.NewRedisStorage(rdb)

	mgr := election.NewManager(
		election.Config{
			NodeID:            "node-1", // set a different value for another node
			HeartbeatInterval: 2 * time.Second,
			TTL:               6 * time.Second,
			JitterPercent:     10,
			EnableFencing:     true,
			//RedisMode:         election.RedisSentinel,
			RedisMode: election.RedisStandalone,
		},
		storage,
		election.Callbacks{
			OnElected: func(res string, token int64) {
				fmt.Println("MASTER:", res, "token", token)
			},
			OnRevoked: func(res string) {
				fmt.Println("LOST MASTER:", res)
			},
		},
	)

	ctx := context.Background()
	mgr.Start(ctx, "resource-1")

	select {}
}
