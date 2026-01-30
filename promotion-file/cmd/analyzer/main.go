package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/huynhanx03/go-demo/promotion-file/internal/di"
	"github.com/huynhanx03/go-demo/promotion-file/internal/infrastructure"
	"github.com/huynhanx03/go-demo/promotion-file/internal/ports"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/constants"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/datastructs/bloom"
)

func main() {
	dataDir := flag.String("data-dir", "data", "Directory containing data files")
	flag.Parse()

	startTotal := time.Now()

	loadStart := time.Now()
	container := di.SetupDependencies(*dataDir)
	service := container.PromotionContainer.Service
	loadDuration := time.Since(loadStart)

	bfValid, _ := bloom.New(infrastructure.EstimatedCapacity, infrastructure.FalsePositiveRate)

	processStart := time.Now()
	processIntersection(filepath.Join(*dataDir, constants.FileMember), service, bfValid)
	processDuration := time.Since(processStart)

	verifyStart := time.Now()
	verifyStats := verifyTestCases(filepath.Join(*dataDir, constants.FileTestCase), bfValid)
	verifyDuration := time.Since(verifyStart)

	totalDuration := time.Since(startTotal)

	fmt.Println("\n================ STATISTICS ================")
	fmt.Printf("Load Time:              %s\n", loadDuration)
	fmt.Printf("Process Time:           %s\n", processDuration)
	fmt.Printf("Verification Time:      %s\n", verifyDuration)

	accuracy := float64(verifyStats.Correct) * 100.0 / float64(verifyStats.Total)
	fmt.Printf("Test Accuracy:          %.2f%% (Correct: %d, Total: %d)\n", accuracy, verifyStats.Correct, verifyStats.Total)

	if verifyStats.FalsePositives > 0 {
		fpRate := float64(verifyStats.FalsePositives) * 100.0 / float64(verifyStats.Total-verifyStats.TruePositives)
		fmt.Printf("False Positives:        %d (%.2f%%)\n", verifyStats.FalsePositives, fpRate)
	}
	fmt.Printf("Total Execution Time:   %s\n", totalDuration)
	fmt.Println("============================================")
}

func processIntersection(inPath string, service ports.IPromotionService, validSet *bloom.Bloom) int64 {
	inFile, err := os.Open(inPath)
	if err != nil {
		log.Fatalf("Failed to open input %s: %v", inPath, err)
	}
	defer inFile.Close()

	var validCount int64
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		code := scanner.Text()
		if service.IsEligible(code) {
			validSet.AddString(code)
			validCount++
		}
	}
	return validCount
}

type VerifyStats struct {
	Total          int64
	Correct        int64
	TruePositives  int64
	FalsePositives int64
	TrueNegatives  int64
	FalseNegatives int64
}

func verifyTestCases(path string, bf *bloom.Bloom) VerifyStats {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: Could not open test cases file: %v", err)
		return VerifyStats{}
	}
	defer f.Close()

	var stats VerifyStats
	scanner := bufio.NewScanner(f)

	if scanner.Scan() {
	}

	for scanner.Scan() {
		line := scanner.Text()
		var code string
		var expectedStr string
		for i := 0; i < len(line); i++ {
			if line[i] == ',' {
				code = line[:i]
				expectedStr = line[i+1:]
				break
			}
		}

		expected := expectedStr == "true"
		actual := bf.HasString(code)

		stats.Total++
		if actual == expected {
			stats.Correct++
			if expected {
				stats.TruePositives++
			} else {
				stats.TrueNegatives++
			}
		} else {
			if actual && !expected {
				stats.FalsePositives++
			} else {
				stats.FalseNegatives++
			}
		}
	}
	return stats
}
