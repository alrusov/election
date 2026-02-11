package election

import "context"

type Storage interface {
	TryAcquireOrRenew(ctx context.Context, resource, nodeID string, ttlMs int64) (token int64, isMaster bool, err error)
	ForceReelection(ctx context.Context, resource string) error
	Resign(ctx context.Context, resource, nodeID string) error
	GetOwner(ctx context.Context, resource string) (string, error)
}
