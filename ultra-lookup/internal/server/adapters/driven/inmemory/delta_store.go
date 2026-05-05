package inmemory

import (
	"fmt"
	"sync"

	"ultra-lookup/internal/lookup"
	"ultra-lookup/internal/server/core/entity"
)

type DeltaStore struct {
	mu        sync.RWMutex
	customers map[uint64]uint8
	accounts  map[uint64]uint8
}

func NewDeltaStore() *DeltaStore {
	return &DeltaStore{
		customers: make(map[uint64]uint8),
		accounts:  make(map[uint64]uint8),
	}
}

func (d *DeltaStore) Lookup(entry entity.Entry) (uint8, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	switch entry.Kind {
	case lookup.KindCustomer:
		v, ok := d.customers[entry.ID]
		return v, ok
	case lookup.KindAccount:
		v, ok := d.accounts[entry.ID]
		return v, ok
	default:
		return 0, false
	}
}

func (d *DeltaStore) Append(entry entity.Entry) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var target map[uint64]uint8
	switch entry.Kind {
	case lookup.KindCustomer:
		target = d.customers
	case lookup.KindAccount:
		target = d.accounts
	default:
		return false, fmt.Errorf("unsupported kind: %s", entry.Kind)
	}

	if _, exists := target[entry.ID]; exists {
		return false, fmt.Errorf("duplicate delta id: %d", entry.ID)
	}
	target[entry.ID] = entry.Shard
	return true, nil
}

func (d *DeltaStore) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.customers) + len(d.accounts)
}
