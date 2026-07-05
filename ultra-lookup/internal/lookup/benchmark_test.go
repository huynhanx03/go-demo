package lookup

import (
	"sync"
	"testing"
)

const (
	benchShardModulo      = 256
	benchAccountShardSkew = 17
)

var (
	benchTable     *Table
	benchBuildErr  error
	benchTableOnce sync.Once
	benchSinkValue uint8
	benchSinkFound bool
)

func buildBenchTable(b *testing.B) *Table {
	b.Helper()

	benchTableOnce.Do(func() {
		const customers = 200_000
		const accounts = 400_000
		records := make([]EncodedRecord, 0, customers+accounts)

		for i := 0; i < customers; i++ {
			records = append(records, EncodedRecord{
				Kind:  KindCustomer,
				ID:    uint64(i + 1),
				Shard: uint8(i % benchShardModulo),
			})
		}

		for i := 0; i < accounts; i++ {
			records = append(records, EncodedRecord{
				Kind:  KindAccount,
				ID:    uint64(i + 1),
				Shard: uint8((i + benchAccountShardSkew) % benchShardModulo),
			})
		}

		tbl, err := BuildEncoded(records, DefaultConfig())
		if err != nil {
			benchBuildErr = err
			return
		}
		benchTable = tbl
	})

	if benchBuildErr != nil {
		b.Fatalf("Build(): %v", benchBuildErr)
	}
	return benchTable
}

func BenchmarkLookupHitRobinHood(b *testing.B) {
	tbl := buildBenchTable(b)
	benchmarkLookupHit(b, tbl)
}

func benchmarkLookupHit(b *testing.B, tbl *Table) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i uint64 = 1
		for pb.Next() {
			v, ok := tbl.LookupEncoded(KindCustomer, i)
			benchSinkValue = v
			benchSinkFound = ok
			i++
			if i > 200_000 {
				i = 1
			}
		}
	})
}

func BenchmarkLookupMissRobinHood(b *testing.B) {
	tbl := buildBenchTable(b)
	benchmarkLookupMiss(b, tbl)
}

func benchmarkLookupMiss(b *testing.B, tbl *Table) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i uint64 = 9_000_000_000
		for pb.Next() {
			v, ok := tbl.LookupEncoded(KindAccount, i)
			benchSinkValue = v
			benchSinkFound = ok
			i++
		}
	})
}
