package splice

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/solo-038/internal/model"
)

type Lease struct {
	ID        model.LeaseID `json:"id"`
	Owner     string        `json:"owner"`
	Fence     uint64        `json:"fence"`
	ExpiresAt time.Time     `json:"expires_at"`
}

type LeaseManager struct {
	mu     sync.Mutex
	fence  uint64
	active Lease
}

func NewLeaseManager() *LeaseManager {
	return &LeaseManager{}
}

func (manager *LeaseManager) Acquire(owner string, ttl time.Duration, now time.Time) (Lease, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if owner == "" || ttl <= 0 {
		return Lease{}, fmt.Errorf("splice lease owner and ttl are required")
	}
	if manager.active.ID.Valid() && now.Before(manager.active.ExpiresAt) && manager.active.Owner != owner {
		return Lease{}, fmt.Errorf("splice window is leased by %s", manager.active.Owner)
	}
	manager.fence++
	manager.active = Lease{ID: model.NewLeaseID(), Owner: owner, Fence: manager.fence, ExpiresAt: now.Add(ttl).UTC()}
	return manager.active, nil
}

func (manager *LeaseManager) Validate(lease Lease, now time.Time) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return lease.ID == manager.active.ID && lease.Fence == manager.active.Fence && lease.Owner == manager.active.Owner && now.Before(manager.active.ExpiresAt)
}

func (manager *LeaseManager) Release(lease Lease) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if lease.ID != manager.active.ID || lease.Fence != manager.active.Fence {
		return false
	}
	manager.active = Lease{}
	return true
}
