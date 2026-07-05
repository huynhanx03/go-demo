package lookup

import "fmt"

type lookupEngine interface {
	LookupEncoded(kind Kind, encodedID uint64) (uint8, bool)
	Len() int
	MemoryStats() MemoryStats
}

type Table struct {
	engine lookupEngine
}

func Build(records []Record, cfg Config) (*Table, error) {
	encoded := make([]EncodedRecord, 0, len(records))
	for _, rec := range records {
		id, err := EncodeBase36ID(rec.ID)
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, EncodedRecord{
			Kind:  rec.Kind,
			ID:    id,
			Shard: rec.Shard,
		})
	}
	return BuildEncoded(encoded, cfg)
}

func BuildEncoded(records []EncodedRecord, cfg Config) (*Table, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("lookup: empty input")
	}

	cfg = cfg.normalized()

	engine, err := buildRobinHoodEngine(records, cfg)
	if err != nil {
		return nil, err
	}

	return &Table{engine: engine}, nil
}

func (t *Table) Lookup(kind Kind, id string) (uint8, bool) {
	encoded, err := EncodeBase36ID(id)
	if err != nil {
		return 0, false
	}
	return t.LookupEncoded(kind, encoded)
}

func (t *Table) LookupEncoded(kind Kind, encodedID uint64) (uint8, bool) {
	return t.engine.LookupEncoded(kind, encodedID)
}

func (t *Table) Len() int {
	return t.engine.Len()
}

func (t *Table) MemoryStats() MemoryStats {
	return t.engine.MemoryStats()
}
