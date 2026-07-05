package di

import (
	"net/http"
	"time"

	"ultra-lookup/internal/lookup"
	"ultra-lookup/internal/metrics"
	"ultra-lookup/internal/server/adapters/driven/inmemory"
	driverhttp "ultra-lookup/internal/server/adapters/driver/http"
	"ultra-lookup/internal/server/core/service"
	"ultra-lookup/internal/server/infrastructure"
)

type Container struct {
	Server        *http.Server
	BuildDuration time.Duration
	MemoryBytes   uint64
}

func Build(cfg infrastructure.Config) (*Container, error) {
	metrics.Register()

	records := lookup.GenerateSyntheticEncoded(lookup.SyntheticConfig{
		Customers: cfg.Customers,
		Accounts:  cfg.Accounts,
		Shards:    cfg.Shards,
		Seed:      cfg.Seed,
	})

	buildStart := time.Now()
	table, err := lookup.BuildEncoded(records, lookup.Config{RobinHoodLoadF: cfg.RHLoad})
	if err != nil {
		return nil, err
	}
	buildDuration := time.Since(buildStart)
	stats := table.MemoryStats()

	metrics.BuildDuration.Set(buildDuration.Seconds())
	metrics.TableEntries.WithLabelValues("base").Set(float64(table.Len()))
	metrics.TableEntries.WithLabelValues("delta").Set(0)
	metrics.TableMemoryBytes.WithLabelValues("keys").Set(float64(stats.KeyBytes))
	metrics.TableMemoryBytes.WithLabelValues("values").Set(float64(stats.ValueBytes))
	metrics.TableMemoryBytes.WithLabelValues("index").Set(float64(stats.IndexBytes))
	metrics.TableMemoryBytes.WithLabelValues("total").Set(float64(stats.TotalEstimatedB))
	metrics.ObserveRuntime()

	deltaStore := inmemory.NewDeltaStore()
	svc := service.NewLookupService(table, deltaStore)
	handler := driverhttp.NewHandler(svc)
	mux := driverhttp.NewMux(handler)
	server := driverhttp.NewServer(cfg.Addr, mux)

	return &Container{
		Server:        server,
		BuildDuration: buildDuration,
		MemoryBytes:   stats.TotalEstimatedB,
	}, nil
}
