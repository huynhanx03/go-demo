package main

import (
	"errors"
	"flag"
	"log/slog"
	"net/http"

	"ultra-lookup/internal/server/di"
	"ultra-lookup/internal/server/infrastructure"
)

func main() {
	cfg := infrastructure.Config{}
	flag.StringVar(&cfg.Addr, "addr", ":8080", "HTTP listen address")
	flag.IntVar(&cfg.Customers, "customers", 6_000_000, "number of synthetic customer IDs")
	flag.IntVar(&cfg.Accounts, "accounts", 12_000_000, "number of synthetic account IDs")
	flag.IntVar(&cfg.Shards, "shards", 256, "number of shard values")
	flag.Uint64Var(&cfg.Seed, "seed", 42, "synthetic data seed")
	flag.Float64Var(&cfg.RHLoad, "rh-load", 0.90, "robinhood load factor")
	flag.Parse()

	container, err := di.Build(cfg)
	if err != nil {
		panic(err)
	}

	slog.Info("server ready", "addr", cfg.Addr, "build", container.BuildDuration.String(), "memory", container.MemoryBytes)
	if err := container.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "error", err)
		panic(err)
	}
}
