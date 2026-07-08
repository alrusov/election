package election

import "time"

type Config struct {
	NodeID            string
	HeartbeatInterval time.Duration
	TTL               time.Duration
	JitterPercent     int // ±%, usually 10
}
