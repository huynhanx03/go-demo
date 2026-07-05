package service

import (
	"ultra-lookup/internal/lookup"
	"ultra-lookup/internal/metrics"
	"ultra-lookup/internal/server/core/domain"
	"ultra-lookup/internal/server/core/entity"
	"ultra-lookup/internal/server/ports/driven"
)

type LookupService struct {
	base  *lookup.Table
	delta driven.DeltaStore
}

func NewLookupService(base *lookup.Table, delta driven.DeltaStore) *LookupService {
	return &LookupService{
		base:  base,
		delta: delta,
	}
}

func (s *LookupService) Lookup(kind lookup.Kind, id uint64) entity.LookupResult {
	entry := entity.Entry{Kind: kind, ID: id}
	if shard, ok := s.delta.Lookup(entry); ok {
		return entity.LookupResult{
			Found:  true,
			Shard:  shard,
			Source: domain.LookupSourceDelta,
		}
	}

	if shard, ok := s.base.LookupEncoded(kind, id); ok {
		return entity.LookupResult{
			Found:  true,
			Shard:  shard,
			Source: domain.LookupSourceBase,
		}
	}
	return entity.LookupResult{
		Found:  false,
		Source: domain.LookupSourceNone,
	}
}

func (s *LookupService) Append(entry entity.Entry) (bool, int, error) {
	added, err := s.delta.Append(entry)
	if err != nil {
		return false, s.delta.Len(), err
	}
	totalDelta := s.delta.Len()
	metrics.TableEntries.WithLabelValues("delta").Set(float64(totalDelta))
	metrics.ObserveRuntime()
	return added, totalDelta, nil
}
