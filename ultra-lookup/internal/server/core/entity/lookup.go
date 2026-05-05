package entity

import "ultra-lookup/internal/lookup"

type Entry struct {
	Kind  lookup.Kind
	ID    uint64
	Shard uint8
}

type LookupResult struct {
	Found  bool
	Shard  uint8
	Source string
}
