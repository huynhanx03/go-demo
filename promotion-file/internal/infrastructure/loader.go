package infrastructure

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/huynhanx03/go-demo/promotion-file/internal/adapters/driven"
	"github.com/huynhanx03/go-demo/promotion-file/internal/ports"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/datastructs/bloom"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/workerpool"
)

const (
	// Estimated number of elements for Bloom Filter
	EstimatedCapacity = 10_000_000
	// High accuracy
	FalsePositiveRate = 0.0001
)

// InitIntersectionRepository initializes the Bloom Filter by finding the intersection of two files.
func InitIntersectionRepository(campaignPath, memberPath string) (ports.IPromotionRepository, error) {
	bfCampaign, err := bloom.New(EstimatedCapacity, FalsePositiveRate)
	if err != nil {
		return nil, err
	}
	loadFileToBloom(campaignPath, bfCampaign)

	bfFinal, err := bloom.New(EstimatedCapacity, FalsePositiveRate)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(memberPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var count int64
	for scanner.Scan() {
		code := scanner.Text()
		if bfCampaign.HasString(code) {
			bfFinal.AddString(code)
			count++
		}
	}
	log.Printf("Intersection Complete. Loaded %d valid codes.", count)

	return driven.NewBloomRepository(bfFinal), nil
}

func loadFileToBloom(path string, bf *bloom.Bloom) int64 {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", path, err)
	}
	defer f.Close()

	var count int64
	workers := runtime.NumCPU()

	pool, _ := workerpool.NewGenericPool[string](workers, func(code string) {
		bf.AddString(code)
		atomic.AddInt64(&count, 1)
	})
	defer pool.Release()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pool.Invoke(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning file: %v", err)
	}

	log.Printf("Loaded %d codes from %s", count, path)
	return count
}
