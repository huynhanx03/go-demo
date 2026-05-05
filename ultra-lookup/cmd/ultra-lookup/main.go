package main

import (
	"flag"
	"fmt"
	"runtime"
	"slices"
	"sync"
	"time"

	"ultra-lookup/internal/lookup"
)

const (
	throughputScale        = 1_000_000.0
	workerIDSkewMultiplier = int64(7919)
)

func main() {
	var (
		customers   = flag.Int("customers", 6_000_000, "number of customer IDs")
		accounts    = flag.Int("accounts", 12_000_000, "number of account IDs")
		shards      = flag.Int("shards", 256, "number of shard values in synthetic data (<=256)")
		lookups     = flag.Int64("lookups", 3_000_000, "number of lookup operations for benchmark")
		workers     = flag.Int("workers", runtime.NumCPU(), "concurrent lookup workers")
		hitRatio    = flag.Int("hit-ratio", 95, "percentage of lookup hits [0..100]")
		sampleEvery = flag.Int("sample-every", 1024, "sample one latency per N requests")
		seed        = flag.Uint64("seed", 42, "seed for deterministic synthetic data")
		rhLoad      = flag.Float64("rh-load", 0.95, "robinhood load factor (0.5..0.99)")
	)
	flag.Parse()

	cfg := lookup.DefaultConfig()
	cfg.RobinHoodLoadF = *rhLoad

	fmt.Printf("building synthetic dataset: algo=robinhood customers=%d accounts=%d shards=%d\n", *customers, *accounts, *shards)
	records := lookup.GenerateSyntheticEncoded(lookup.SyntheticConfig{
		Customers: *customers,
		Accounts:  *accounts,
		Shards:    *shards,
		Seed:      *seed,
	})

	buildStart := time.Now()
	table, err := lookup.BuildEncoded(records, cfg)
	if err != nil {
		panic(err)
	}
	buildDur := time.Since(buildStart)

	stats := table.MemoryStats()
	fmt.Printf("build done in %s\n", buildDur)
	fmt.Printf("entries=%d keys=%s value=%s index=%s total=%s\n",
		stats.Entries,
		humanBytes(stats.KeyBytes),
		humanBytes(stats.ValueBytes),
		humanBytes(stats.IndexBytes),
		humanBytes(stats.TotalEstimatedB),
	)

	bench := runLookupBenchmark(table, benchmarkConfig{
		customers:   *customers,
		accounts:    *accounts,
		lookups:     *lookups,
		workers:     *workers,
		hitRatio:    *hitRatio,
		sampleEvery: *sampleEvery,
	})

	fmt.Printf("lookups=%d workers=%d elapsed=%s throughput=%.2f M TPS\n",
		bench.totalOps, bench.workers, bench.elapsed, bench.tps/throughputScale)
	fmt.Printf("hits=%d misses=%d sampled=%d p50=%s p99=%s\n",
		bench.hits, bench.misses, len(bench.samples), bench.p50, bench.p99)
}

type benchmarkConfig struct {
	customers   int
	accounts    int
	lookups     int64
	workers     int
	hitRatio    int
	sampleEvery int
}

type benchmarkResult struct {
	totalOps int64
	workers  int
	hits     int64
	misses   int64
	elapsed  time.Duration
	tps      float64
	samples  []int64
	p50      time.Duration
	p99      time.Duration
}

func runLookupBenchmark(table *lookup.Table, cfg benchmarkConfig) benchmarkResult {
	if cfg.workers <= 0 {
		cfg.workers = 1
	}
	if cfg.sampleEvery <= 0 {
		cfg.sampleEvery = 1024
	}
	if cfg.hitRatio < 0 {
		cfg.hitRatio = 0
	}
	if cfg.hitRatio > 100 {
		cfg.hitRatio = 100
	}

	perWorker := cfg.lookups / int64(cfg.workers)
	extra := cfg.lookups % int64(cfg.workers)

	type workerResult struct {
		hits    int64
		misses  int64
		samples []int64
	}
	results := make([]workerResult, cfg.workers)
	var wg sync.WaitGroup

	start := time.Now()

	for w := 0; w < cfg.workers; w++ {
		ops := perWorker
		if int64(w) < extra {
			ops++
		}

		wg.Add(1)
		go func(workerID int, ops int64) {
			defer wg.Done()

			localSamples := make([]int64, 0, max(1, int(ops)/cfg.sampleEvery+1))
			var localHits int64

			for i := int64(0); i < ops; i++ {
				hitRequest := (i % 100) < int64(cfg.hitRatio)
				kind := lookup.KindCustomer
				limit := cfg.customers
				if i&1 == 1 {
					kind = lookup.KindAccount
					limit = cfg.accounts
				}
				if limit <= 0 {
					continue
				}

				var id uint64
				if hitRequest {
					id = uint64((i + int64(workerID)*workerIDSkewMultiplier) % int64(limit))
					id++
				} else {
					id = uint64(limit) + uint64(i) + 1
				}

				if i%int64(cfg.sampleEvery) == 0 {
					t0 := time.Now()
					_, ok := table.LookupEncoded(kind, id)
					localSamples = append(localSamples, time.Since(t0).Nanoseconds())
					if ok {
						localHits++
					}
				} else {
					_, ok := table.LookupEncoded(kind, id)
					if ok {
						localHits++
					}
				}
			}

			localMisses := ops - localHits
			results[workerID] = workerResult{
				hits:    localHits,
				misses:  localMisses,
				samples: localSamples,
			}
		}(w, ops)
	}

	wg.Wait()
	elapsed := time.Since(start)

	var (
		totalHits   int64
		totalMisses int64
		allSamples  []int64
	)
	for _, r := range results {
		totalHits += r.hits
		totalMisses += r.misses
		allSamples = append(allSamples, r.samples...)
	}

	p50, p99 := percentiles(allSamples)

	return benchmarkResult{
		totalOps: cfg.lookups,
		workers:  cfg.workers,
		hits:     totalHits,
		misses:   totalMisses,
		elapsed:  elapsed,
		tps:      float64(cfg.lookups) / elapsed.Seconds(),
		samples:  allSamples,
		p50:      p50,
		p99:      p99,
	}
}

func percentiles(samples []int64) (time.Duration, time.Duration) {
	if len(samples) == 0 {
		return 0, 0
	}
	slices.Sort(samples)
	p50Idx := (len(samples) - 1) * 50 / 100
	p99Idx := (len(samples) - 1) * 99 / 100
	return time.Duration(samples[p50Idx]), time.Duration(samples[p99Idx])
}

func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
