# ultra-lookup

High-density lookup demo for mapping customer/account IDs to the correct cluster under a strict memory and latency budget.

## Problem

Build a migration lookup table for:

- `6,000,000` Customer IDs
- `12,000,000` Account IDs
- `18,000,000` total mappings

Target:

- `>= 1,000,000 TPS`
- very low and stable `p99`
- total memory usage `< 200MB`
- exact lookup result: if a key does not exist, return `not found`

## Requirements

Input/domain requirements:

- `CustomerID` and `AccountID` are 10-character strings.
- Valid characters are `[0-9A-Z]`.
- Customer and account IDs are managed independently, so the same string may exist in both namespaces.
- IDs are increasing over time, but they are not guaranteed to be contiguous or evenly ranged.
- Cluster/shard number is small and expected to fit in `0..255`.
- The migration table is mostly static; updates are rare and mostly append-only.

Runtime requirements:

- Memory budget means total RAM, regardless of heap or off-heap.
- TPS target is for a real system with concurrent execution, expected around `16` workers/threads.
- Lookup must be exact, not probabilistic.
- Lookup hot path should avoid allocations and object-per-entry storage.

## Why A Regular HashMap Fails

An object-heavy hashmap such as `java.util.HashMap` is too expensive for this budget.

A single hashmap node can cost roughly `56 bytes` or more after object headers, references, key/value objects, hash metadata, and alignment. At `18M` records, that easily reaches `~1GB+`, far above the `<200MB` target.

It also creates many objects for the garbage collector to manage, which increases the risk of unstable p99 latency.

The data structure needs:

- primitive/packed arrays
- contiguous memory
- exact membership check
- zero allocation on lookup
- good cache locality

## Data Model

IDs are encoded from base36-10 strings into integers:

```text
36^10 = 3,656,158,440,062,976 < 2^52
```

So each ID fits in `52 bits`.

The demo stores:

- key: packed 52-bit value
- shard: `uint8`
- lookup metadata: compact primitive arrays/bitsets

This keeps memory close to the required density:

```text
200MB / 18M records ~= 11 bytes/record
```

## Solution

This demo focuses on one production-oriented structure: **Static Packed Robin Hood Hash Table**.

It is a static open-addressing hash table built once and used read-only on the hot lookup path.

Lookup flow:

```text
hash(encodedID) -> probe contiguous slots -> exact key compare -> shard
```

Why this fits the problem:

- packed 52-bit keys avoid storing ID strings
- `uint8` shard values keep value storage small
- contiguous arrays improve cache locality
- Robin Hood probing keeps probe distance balanced, helping p99 stability
- exact key comparison prevents false positives
- no object-per-entry storage

## Append Strategy

The main table is static by design:

```text
build once -> serve many lookups
```

For rare append-only updates, the recommended production model is **base + delta**:

```text
lookup(key):
  if delta contains key:
    return delta[key]
  return base[key]
```

- `base`: large Static Packed Robin Hood table for the main dataset.
- `delta`: small append table for newly added records.
- periodic rebuild: merge `base + delta` into a new base table.
- atomic swap: replace the old base table without blocking lookup traffic.

This avoids resizing or mutating the large table on the lookup path, keeping p99 stable while still supporting new records.

## Current Result

Local benchmark on Apple M1, synthetic dataset matching the target scale:

```text
rh-load    = 0.90
customers  = 6,000,000
accounts   = 12,000,000
entries    = 18,000,000
workers    = 16
memory     ~= 164.51 MiB
throughput ~= 21M TPS
p99        ~= 750ns
```

This measures the in-process core lookup engine. A production system still needs end-to-end validation including request parsing, RPC, logging, metrics, scheduling, and deployment topology.

## Run

Run the fixed production-like demo profile:

```bash
make run
```

Run the HTTP service:

```bash
make server
```

Run HTTP load with append traffic:

```bash
make load-http
```

Run the full E2E observability stack:

```bash
make observability
```

This starts:

- ultra-lookup HTTP server on `http://localhost:8080`
- Prometheus on `http://localhost:9090`
- Grafana on `http://localhost:3000` (`admin/admin`)

The server exposes:

- `GET /lookup?kind=customer&id=000000000A`
- `POST /append`
- `GET /healthz`
- `GET /readyz`
- `GET /metrics`

Append requests go into a small delta table. Lookup checks delta first, then the static base table, so new records can be served without mutating or resizing the large table.

## Make Targets

```bash
make run            # default production-like RobinHood profile
make server         # HTTP API with Prometheus metrics
make load-http      # HTTP load generator with append traffic
make observability  # app + Prometheus + Grafana
make down           # stop Docker Compose stack
make test           # correctness tests
make bench          # microbenchmarks
make tidy           # go mod tidy
```

## Important Flags

- `-rh-load`: RobinHood load factor, default `0.90` in Makefile
- `-sample-every`: latency sampling interval

## Production Readiness Signals

The E2E demo exports Prometheus metrics for:

- lookup throughput
- lookup p50/p95/p99 latency
- hit/miss/delta-hit ratio
- append request rate and latency
- base and delta table entries
- estimated table memory
- Go heap, runtime memory, GC pause total, and goroutines
