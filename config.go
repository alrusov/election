package election

import "time"

type RedisMode string

const (
	RedisStandalone RedisMode = "standalone"
	RedisSentinel   RedisMode = "sentinel"
	RedisCluster    RedisMode = "cluster"
)

type Config struct {
	NodeID            string
	HeartbeatInterval time.Duration
	TTL               time.Duration
	JitterPercent     int // ±%, usually 10

	EnableFencing bool
	RedisMode     RedisMode
}
