package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var lookupLatencyBuckets = []float64{
	0.00000005,
	0.0000001,
	0.00000025,
	0.0000005,
	0.000001,
	0.0000025,
	0.000005,
	0.00001,
	0.000025,
	0.00005,
	0.0001,
	0.00025,
	0.0005,
	0.001,
	0.0025,
	0.005,
	0.01,
}

var (
	LookupRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ultra_lookup_requests_total",
			Help: "Total lookup requests.",
		},
		[]string{"kind", "result", "source"},
	)

	LookupLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ultra_lookup_latency_seconds",
			Help:    "Lookup request latency.",
			Buckets: lookupLatencyBuckets,
		},
		[]string{"kind", "result", "source"},
	)

	AppendRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ultra_lookup_append_requests_total",
			Help: "Total append requests.",
		},
		[]string{"kind", "result"},
	)

	AppendLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ultra_lookup_append_latency_seconds",
			Help:    "Append request latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"kind", "result"},
	)

	BuildDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ultra_lookup_build_duration_seconds",
			Help: "Initial base table build duration.",
		},
	)

	TableEntries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ultra_lookup_table_entries",
			Help: "Number of entries by table layer.",
		},
		[]string{"layer"},
	)

	TableMemoryBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ultra_lookup_table_memory_bytes",
			Help: "Estimated memory usage by table part.",
		},
		[]string{"part"},
	)

	RuntimeMemoryBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ultra_lookup_runtime_memory_bytes",
			Help: "Go runtime memory statistics.",
		},
		[]string{"part"},
	)

	GCPauseTotalSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ultra_lookup_gc_pause_total_seconds",
			Help: "Total GC pause time reported by Go runtime.",
		},
	)

	Goroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ultra_lookup_goroutines",
			Help: "Current number of goroutines.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		LookupRequests,
		LookupLatency,
		AppendRequests,
		AppendLatency,
		BuildDuration,
		TableEntries,
		TableMemoryBytes,
		RuntimeMemoryBytes,
		GCPauseTotalSeconds,
		Goroutines,
	)
}

func ObserveRuntime() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	RuntimeMemoryBytes.WithLabelValues("alloc").Set(float64(mem.Alloc))
	RuntimeMemoryBytes.WithLabelValues("heap_alloc").Set(float64(mem.HeapAlloc))
	RuntimeMemoryBytes.WithLabelValues("heap_sys").Set(float64(mem.HeapSys))
	RuntimeMemoryBytes.WithLabelValues("stack_sys").Set(float64(mem.StackSys))
	RuntimeMemoryBytes.WithLabelValues("sys").Set(float64(mem.Sys))
	GCPauseTotalSeconds.Set(float64(mem.PauseTotalNs) / float64(time.Second))
	Goroutines.Set(float64(runtime.NumGoroutine()))
}
