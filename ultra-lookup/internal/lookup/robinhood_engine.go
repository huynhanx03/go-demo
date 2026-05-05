package lookup

import "fmt"

type robinHoodEngine struct {
	customers robinHoodKindTable
	accounts  robinHoodKindTable
	entries   int
}

type robinHoodKindTable struct {
	cap    uint32
	keys   packed52
	values shardValues
	dist   []uint8
	occ    []uint64
	size   int
}

func buildRobinHoodEngine(records []EncodedRecord, cfg Config) (*robinHoodEngine, error) {
	customers := make([]kindRecord, 0, len(records)/3)
	accounts := make([]kindRecord, 0, len(records)-len(records)/3)

	for _, rec := range records {
		switch rec.Kind {
		case KindCustomer:
			customers = append(customers, kindRecord{id: rec.ID, shard: rec.Shard})
		case KindAccount:
			accounts = append(accounts, kindRecord{id: rec.ID, shard: rec.Shard})
		default:
			return nil, fmt.Errorf("lookup: unsupported kind: %v", rec.Kind)
		}
	}

	customerTable, err := buildRobinHoodKindTable(customers, KindCustomer, cfg.RobinHoodLoadF)
	if err != nil {
		return nil, err
	}
	accountTable, err := buildRobinHoodKindTable(accounts, KindAccount, cfg.RobinHoodLoadF)
	if err != nil {
		return nil, err
	}

	return &robinHoodEngine{
		customers: customerTable,
		accounts:  accountTable,
		entries:   len(records),
	}, nil
}

func buildRobinHoodKindTable(records []kindRecord, kind Kind, loadFactor float64) (robinHoodKindTable, error) {
	n := len(records)
	if n == 0 {
		return robinHoodKindTable{}, nil
	}

	capacity := uint64(float64(n)/loadFactor) + 1
	if capacity < 16 {
		capacity = 16
	}
	if capacity > uint64(^uint32(0)) {
		return robinHoodKindTable{}, fmt.Errorf("lookup: robinhood capacity too large for %s", kind)
	}

	t := robinHoodKindTable{
		cap:    uint32(capacity),
		keys:   newPacked52(int(capacity)),
		values: newShardValues(int(capacity)),
		dist:   make([]uint8, int(capacity)),
		occ:    make([]uint64, (capacity+63)/64),
		size:   n,
	}

	for _, rec := range records {
		if err := t.insert(rec.id, rec.shard, kind); err != nil {
			return robinHoodKindTable{}, err
		}
	}
	return t, nil
}

func (e *robinHoodEngine) LookupEncoded(kind Kind, encodedID uint64) (uint8, bool) {
	switch kind {
	case KindCustomer:
		return e.customers.lookup(encodedID)
	case KindAccount:
		return e.accounts.lookup(encodedID)
	default:
		return 0, false
	}
}

func (e *robinHoodEngine) Len() int {
	return e.entries
}

func (e *robinHoodEngine) MemoryStats() MemoryStats {
	keyBytes := e.customers.keys.byteSize() + e.accounts.keys.byteSize()
	valueBytes := e.customers.values.byteSize() + e.accounts.values.byteSize()
	indexBytes := uint64(len(e.customers.dist)+len(e.accounts.dist)) +
		uint64(len(e.customers.occ)+len(e.accounts.occ))*8
	return MemoryStats{
		Entries:         e.entries,
		KeyBytes:        keyBytes,
		ValueBytes:      valueBytes,
		IndexBytes:      indexBytes,
		TotalEstimatedB: keyBytes + valueBytes + indexBytes,
	}
}

func (t *robinHoodKindTable) insert(id uint64, shard uint8, kind Kind) error {
	slot := uint32(splitmix64(indexKey(id)) % uint64(t.cap))
	curID := id
	curShard := shard
	curDist := uint8(0)

	for steps := uint32(0); steps < t.cap; steps++ {
		i := int(slot)
		if !alreadySet(t.occ, uint64(slot)) {
			markSet(t.occ, uint64(slot))
			t.keys.set(i, curID)
			t.values.set(uint64(slot), curShard)
			t.dist[i] = curDist
			return nil
		}

		existingID := t.keys.get(i)
		if existingID == curID {
			return fmt.Errorf("lookup: duplicate id (%s:%d)", kind, id)
		}

		existingDist := t.dist[i]
		if existingDist < curDist {
			existingShard := t.values.get(uint64(slot))
			t.keys.set(i, curID)
			t.values.set(uint64(slot), curShard)
			t.dist[i] = curDist

			curID = existingID
			curShard = existingShard
			curDist = existingDist
		}

		slot++
		if slot == t.cap {
			slot = 0
		}
		if curDist == ^uint8(0) {
			return fmt.Errorf("lookup: robinhood probe overflow (%s)", kind)
		}
		curDist++
	}

	return fmt.Errorf("lookup: robinhood table full (%s)", kind)
}

func (t *robinHoodKindTable) lookup(id uint64) (uint8, bool) {
	if len(t.dist) == 0 {
		return 0, false
	}

	slot := uint32(splitmix64(indexKey(id)) % uint64(t.cap))
	searchDist := uint8(0)
	for steps := uint32(0); steps < t.cap; steps++ {
		i := int(slot)
		if !alreadySet(t.occ, uint64(slot)) {
			return 0, false
		}
		if t.dist[i] < searchDist {
			return 0, false
		}
		if t.keys.get(i) == id {
			return t.values.get(uint64(slot)), true
		}

		slot++
		if slot == t.cap {
			slot = 0
		}
		if searchDist == ^uint8(0) {
			return 0, false
		}
		searchDist++
	}
	return 0, false
}
