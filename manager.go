package election

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Manager struct {
	cfg       Config
	storage   Storage
	callbacks Callbacks

	mu      sync.Mutex
	masters map[string]bool
	tokens  map[string]int64
	cancel  map[string]context.CancelFunc
}

func NewManager(cfg Config, storage Storage, cb Callbacks) *Manager {
	if cfg.JitterPercent == 0 {
		cfg.JitterPercent = 10
	}
	return &Manager{
		cfg:       cfg,
		storage:   storage,
		callbacks: cb,
		masters:   map[string]bool{},
		tokens:    map[string]int64{},
		cancel:    map[string]context.CancelFunc{},
	}
}

func (m *Manager) Start(ctx context.Context, resource string) {
	m.mu.Lock()
	if _, ok := m.cancel[resource]; ok {
		m.mu.Unlock()
		return
	}
	cctx, cancel := context.WithCancel(ctx)
	m.cancel[resource] = cancel
	m.mu.Unlock()

	go m.loop(cctx, resource)
}

func (m *Manager) loop(ctx context.Context, resource string) {
	for {
		d := jitterDuration(m.cfg.HeartbeatInterval, m.cfg.JitterPercent)
		timer := time.NewTimer(d)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			m.tick(resource)
		}
	}
}

func jitterDuration(base time.Duration, percent int) time.Duration {
	if percent <= 0 {
		return base
	}
	maxJitter := float64(base) * float64(percent) / 100
	delta := (rand.Float64()*2 - 1) * maxJitter
	return time.Duration(float64(base) + delta)
}

func (m *Manager) tick(resource string) {
	token, isMaster, err := m.storage.TryAcquireOrRenew(
		context.Background(),
		resource,
		m.cfg.NodeID,
		m.cfg.TTL.Milliseconds(),
	)
	if err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.masters[resource]

	if isMaster {
		m.masters[resource] = true
		m.tokens[resource] = token
		if !prev && m.callbacks.OnElected != nil {
			go m.callbacks.OnElected(resource, token)
		}
	} else {
		if prev {
			m.masters[resource] = false
			if m.callbacks.OnRevoked != nil {
				go m.callbacks.OnRevoked(resource)
			}
		}
	}
}

func (m *Manager) IsMaster(resource string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.masters[resource]
}

func (m *Manager) Token(resource string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tokens[resource]
}

func (m *Manager) Resign(resource string) error {
	return m.storage.Resign(context.Background(), resource, m.cfg.NodeID)
}

func (m *Manager) ForceReelection(resource string) error {
	return m.storage.ForceReelection(context.Background(), resource)
}

func (m *Manager) Stop(resource string) {
	m.mu.Lock()
	if cancel, ok := m.cancel[resource]; ok {
		cancel()
		delete(m.cancel, resource)
		delete(m.masters, resource)
		delete(m.tokens, resource)
	}
	m.mu.Unlock()
}
