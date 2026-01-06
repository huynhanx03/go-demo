# Go Radius Search: High-Performance Benchmark (MySQL vs Redis)

## Abstract
This project proposes a high-performance architecture for Location-Based Services (LBS), specifically addressing the "K-Nearest Neighbors" (KNN) problem in a distributed environment. The system leverages in-memory spatial indexing techniques to achieve sub-millisecond latency for spatial queries, ensuring real-time responsiveness for user-centric applications.

## Methodology
The core mechanism utilizes **Redis Geospatial**, which implements a **Geohash** encoding scheme mapped to a **Sorted Set (ZSET)** data structure.
- **Spatial Encoding**: 2D coordinates (latitude, longitude) are interleaved into a 52-bit integer (Geohash), reducing multi-dimensional spatial data into a linear 1D index.
- **Query Efficiency**: Range queries (e.g., `GEOSEARCH`) operate with $O(N + \log M)$ time complexity, where $N$ is the number of elements in the radius and $M$ is the total number of items in the set. This in-memory approach provides a significant throughput advantage over traditional B-Tree based disk storage systems (e.g., PostGIS, MongoDB).

## Optimization Strategy
To address the limitations of vertical scaling (single-node memory constraints), we introduce a **Geo-Sharding** partitioning strategy.
- **Partitioning Logic**: The spatial domain is discretized into grid cells using coarse-grained Geohashes. Data points falling within a specific grid cell are routed to a dedicated Redis shard.
- **Load Balancing**: This approach distributes the dataset and read/write throughput across multiple nodes, ensuring the system remains performant under high-concurrency loads.

## Trade-offs

### Advantages
1.  **Ultra-Low Latency**: The utilization of in-memory data structures minimizes I/O overhead, delivering deterministic performance suitable for real-time SLA requirements.
2.  **Linear Scalability**: The shared-nothing architecture enabled by Geo-Sharding allows linear scaling of throughput by adding additional nodes.

### Limitations & Challenges
1.  **Volatile Memory Cost**: The dependency on DRAM for storage introduces higher operational costs compared to SSD-backed solutions.
2.  **Boundary & Edge Effects**:
    - **Cross-Shard Queries**: Users located near the boundary of a shard necessitate "Scatter-Gather" query patterns, where requests must be fanned out to adjacent shards.
    - **Complexity**: Aggregating results from multiple shards requires complex application-side logic to handle deduplication and sorting, impacting the simplicity of the implementation.

## Experimental Results

### Dataset Configuration
- **Volume**: 1,000,000 records
- **Distribution**: Randomly distributed across Vietnam (Lat: 8.5-23.5, Lng: 102.0-109.5)
- **Environment**: Local Docker Containers (MacBook M1 2020 16GB RAM)

### Latency Benchmark

| Test Case | Region | Coordinates | Radius | MySQL Latency | Redis Latency | Speedup |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Case 1** | Hanoi (Dense) | `21.0285, 105.8542` | **5 km** | 823 ms | **24 ms** | **~34x** |
| **Case 2** | HCM (Dense) | `10.7769, 106.7009` | **10 km** | 797 ms | **21 ms** | **~38x** |
| **Case 3** | Da Nang (Wide) | `16.0610, 108.2270` | **50 km** | 791 ms | **214 ms** | **~3.7x** |

### Analysis
1.  **Recall & Precision**: Both methods return identical result sets, ensuring data integrity.
2.  **Performance Gap**:
    - **Redis (In-Memory)**: Delivers sub-millisecond to double-digit millisecond responses. The use of Geohash grouping in Sorted Sets allows `O(k+log(N))` search complexity.
    - **MySQL (Disk-Based)**: Consistent latency around 800ms. While `ST_Distance_Sphere` is optimized with spatial indexes (R-Tree), it incurs I/O overhead on large datasets (1M+ rows).
3.  **Scalability**: Redis demonstrates superior scalability for high-concurrency read workloads, effectively serving as a query offload engine.

## Quick Start

### 1. Requirements
*   **Docker** & **Go 1.22+** installed.

### 2. Sync Data
Activate the CDC connection to stream data from MySQL to Redis:
```bash
make cdc-init
```

### 3. Boot Up
Start the infrastructure and API. This automatically seeds **1,000,000 shops** if the database is empty.
```bash
make run
```

> **Note**: Wait until you see `shop data check/init completed` in the logs (~2-3 mins).

### 4. Benchmark
Compare the performance difference yourself:

**MySQL (Standard)**
```bash
curl -X POST http://localhost:8080/api/v1/shops/search-radius \
  -H "Content-Type: application/json" \
  -d '{"lat": 21.0285, "lng": 105.8542, "radius": 5.0}'
```

**Redis (Optimized)**
```bash
curl -X POST http://localhost:8080/api/v1/shops/search-radius-fast \
  -H "Content-Type: application/json" \
  -d '{"lat": 21.0285, "lng": 105.8542, "radius": 5.0}'
```

## Conclusion
The hybrid approach (MySQL for storage + Redis for query) with CDC synchronization successfully solves the trade-off between consistency and performance, achieving **>30x speed improvement** for critical LBS queries.