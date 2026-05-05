package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"ultra-lookup/internal/lookup"
)

const (
	defaultDuration        = 60 * time.Second
	defaultWorkers         = 16
	defaultHitRatio        = 95
	defaultAppendEvery     = 1000
	defaultSampleEvery     = 1024
	httpClientTimeout      = 3 * time.Second
	localSamplesPerSecond  = 100
	localAppendSamplesInit = 128
	accountWorkerOffset    = int64(7919)
	newIDBaseOffset        = uint64(10_000_000_000)
	newIDWorkerStride      = uint64(1_000_000)
	maxShardValuePlusOne   = int64(256)
)

func main() {
	var (
		baseURL     = flag.String("url", "http://localhost:8080", "server base URL")
		duration    = flag.Duration("duration", defaultDuration, "load duration")
		workers     = flag.Int("workers", defaultWorkers, "concurrent workers")
		customers   = flag.Int("customers", 6_000_000, "base customer IDs")
		accounts    = flag.Int("accounts", 12_000_000, "base account IDs")
		hitRatio    = flag.Int("hit-ratio", defaultHitRatio, "lookup hit ratio")
		appendEvery = flag.Int("append-every", defaultAppendEvery, "append one new record every N operations per worker; 0 disables append")
		sampleEvery = flag.Int("sample-every", defaultSampleEvery, "sample one latency every N operations")
	)
	flag.Parse()

	if *workers <= 0 {
		*workers = 1
	}
	if *sampleEvery <= 0 {
		*sampleEvery = defaultSampleEvery
	}

	var stats loadStats
	samples := make([][]int64, *workers)
	appendSamples := make([][]int64, *workers)
	deadline := time.Now().Add(*duration)
	start := time.Now()

	var wg sync.WaitGroup
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: httpClientTimeout}
			localSamples := make([]int64, 0, int(*duration/time.Second)*localSamplesPerSecond)
			localAppendSamples := make([]int64, 0, localAppendSamplesInit)

			for i := int64(0); time.Now().Before(deadline); i++ {
				if *appendEvery > 0 && i > 0 && i%int64(*appendEvery) == 0 {
					t0 := time.Now()
					if err := appendRecord(client, *baseURL, workerID, i, *customers, *accounts); err != nil {
						atomic.AddInt64(&stats.errors, 1)
					} else {
						atomic.AddInt64(&stats.appends, 1)
					}
					localAppendSamples = append(localAppendSamples, time.Since(t0).Nanoseconds())
					continue
				}

				t0 := time.Now()
				found, err := lookupRecord(client, *baseURL, workerID, i, *customers, *accounts, *hitRatio)
				dur := time.Since(t0)
				if i%int64(*sampleEvery) == 0 {
					localSamples = append(localSamples, dur.Nanoseconds())
				}
				if err != nil {
					atomic.AddInt64(&stats.errors, 1)
					continue
				}
				atomic.AddInt64(&stats.lookups, 1)
				if found {
					atomic.AddInt64(&stats.hits, 1)
				} else {
					atomic.AddInt64(&stats.misses, 1)
				}
			}

			samples[workerID] = localSamples
			appendSamples[workerID] = localAppendSamples
		}(w)
	}

	wg.Wait()
	elapsed := time.Since(start)
	allSamples := flatten(samples)
	allAppendSamples := flatten(appendSamples)
	p50, p95, p99 := percentiles(allSamples)
	appendP50, appendP95, appendP99 := percentiles(allAppendSamples)

	fmt.Printf("duration=%s workers=%d lookups=%d appends=%d errors=%d throughput=%.2f req/s\n",
		elapsed.Round(time.Millisecond),
		*workers,
		stats.lookups,
		stats.appends,
		stats.errors,
		float64(stats.lookups+stats.appends)/elapsed.Seconds(),
	)
	fmt.Printf("lookup hits=%d misses=%d sampled=%d p50=%s p95=%s p99=%s\n",
		stats.hits, stats.misses, len(allSamples), p50, p95, p99)
	fmt.Printf("append sampled=%d p50=%s p95=%s p99=%s\n",
		len(allAppendSamples), appendP50, appendP95, appendP99)
}

type loadStats struct {
	lookups int64
	appends int64
	hits    int64
	misses  int64
	errors  int64
}

func lookupRecord(client *http.Client, baseURL string, workerID int, i int64, customers, accounts, hitRatio int) (bool, error) {
	kind := "customer"
	limit := customers
	if i&1 == 1 {
		kind = "account"
		limit = accounts
	}

	hit := (i % 100) < int64(hitRatio)
	idNum := uint64(limit) + uint64(i) + 1
	if hit {
		idNum = uint64((i + int64(workerID)*accountWorkerOffset) % int64(limit))
		idNum++
	}

	id, err := lookup.FormatBase36ID(idNum)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/lookup?kind="+kind+"&id="+id, nil)
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return false, fmt.Errorf("lookup status=%d", resp.StatusCode)
	}

	var out struct {
		Found bool `json:"found"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Found, nil
}

func appendRecord(client *http.Client, baseURL string, workerID int, i int64, customers, accounts int) error {
	kind := "customer"
	base := customers
	if i&1 == 1 {
		kind = "account"
		base = accounts
	}
	idNum := uint64(base) + newIDBaseOffset + uint64(workerID)*newIDWorkerStride + uint64(i)
	id, err := lookup.FormatBase36ID(idNum)
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]any{
		"kind":  kind,
		"id":    id,
		"shard": uint8((i + int64(workerID)) % maxShardValuePlusOne),
	})
	if err != nil {
		return err
	}

	resp, err := client.Post(baseURL+"/append", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("append status=%d", resp.StatusCode)
	}
	return nil
}

func flatten(xs [][]int64) []int64 {
	var total int
	for _, x := range xs {
		total += len(x)
	}
	out := make([]int64, 0, total)
	for _, x := range xs {
		out = append(out, x...)
	}
	return out
}

func percentiles(samples []int64) (time.Duration, time.Duration, time.Duration) {
	if len(samples) == 0 {
		return 0, 0, 0
	}
	slices.Sort(samples)
	p50 := samples[(len(samples)-1)*50/100]
	p95 := samples[(len(samples)-1)*95/100]
	p99 := samples[(len(samples)-1)*99/100]
	return time.Duration(p50), time.Duration(p95), time.Duration(p99)
}
