package election

import (
	"context"
	"testing"
	"time"
)

type mockStorage struct {
	owner string
	token int64
}

func (m *mockStorage) TryAcquireOrRenew(ctx context.Context, r, node string, ttl int64) (int64, bool, error) {
	if m.owner == "" {
		m.owner = node
		m.token++
		return m.token, true, nil
	}
	if m.owner == node {
		return m.token, true, nil
	}
	return 0, false, nil
}

func (m *mockStorage) ForceReelection(ctx context.Context, r string) error {
	m.owner = ""
	return nil
}

func (m *mockStorage) Resign(ctx context.Context, r, node string) error {
	if m.owner == node {
		m.owner = ""
	}
	return nil
}

func (m *mockStorage) GetOwner(ctx context.Context, r string) (string, error) {
	return m.owner, nil
}

func TestElection(t *testing.T) {
	st := &mockStorage{}
	elected := false

	m := NewManager(Config{
		NodeID:            "node1",
		HeartbeatInterval: 10 * time.Millisecond,
		TTL:               100 * time.Millisecond,
		JitterPercent:     10,
	}, st, Callbacks{
		OnElected: func(r string, token int64) {
			elected = true
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx, "res1")
	time.Sleep(50 * time.Millisecond)

	if !elected {
		t.Fatal("not elected")
	}
	if !m.IsMaster("res1") {
		t.Fatal("should be master")
	}
}
