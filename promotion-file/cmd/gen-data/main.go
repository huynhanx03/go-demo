package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/huynhanx03/go-demo/promotion-file/pkg/constants"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/workerpool"
)

const (
	TotalUnique   = 10_000_000
	CommonCount   = 2_500_000
	CodeLength    = 5
	Charset       = "abcdefghijklmnopqrstuvwxyz"
	TestCaseCount = 1_000_000
	BatchSize     = 1000
)

func main() {
	outDir := flag.String("out-dir", "data", "Output directory")
	flag.Parse()

	os.MkdirAll(*outDir, 0755)

	pathCampaign := filepath.Join(*outDir, constants.FileCampaign)
	pathMember := filepath.Join(*outDir, constants.FileMember)
	pathTestCase := filepath.Join(*outDir, constants.FileTestCase)

	fmt.Printf("Generating %d unique codes (Common: %d)... using %d workers\n", TotalUnique, CommonCount, runtime.NumCPU())

	allCodes := generateWorkerPool(TotalUnique)
	fmt.Printf("Generated %d unique codes\n", len(allCodes))

	common := allCodes[:CommonCount]
	remaining := allCodes[CommonCount:]

	halfRem := len(remaining) / 2
	campaignUnique := remaining[:halfRem]
	memberUnique := remaining[halfRem:]

	fmt.Printf("Distribution:\n- Common: %d\n- Campaign Unique: %d\n- Member Unique: %d\n", len(common), len(campaignUnique), len(memberUnique))

	write(pathCampaign, merge(common, campaignUnique))
	write(pathMember, merge(common, memberUnique))
	genTestCases(pathTestCase, common, campaignUnique, memberUnique)

	fmt.Println("Done.")
}

func generateWorkerPool(target int) []string {
	res := make(map[string]bool, target)
	list := make([]string, 0, target)

	workers := runtime.NumCPU()
	resultChan := make(chan string, 10000)

	pool, _ := workerpool.NewGenericPool[int](workers, func(batchSize int) {
		for i := 0; i < batchSize; i++ {
			resultChan <- randStr()
		}
	})
	defer pool.Release()

	done := make(chan bool)
	go func() {
		for code := range resultChan {
			if !res[code] {
				res[code] = true
				list = append(list, code)
				if len(list)%1_000_000 == 0 {
					fmt.Printf("Collected %d codes...\n", len(list))
				}
				if len(list) >= target {
					close(done)
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				pool.Invoke(100)
			}
		}
	}()

	<-done
	return list
}

func randStr() string {
	n := rand.Intn(5) + 1
	b := make([]byte, n)
	for i := range b {
		b[i] = Charset[rand.Intn(len(Charset))]
	}
	return string(b)
}

func merge(s1, s2 []string) []string {
	res := make([]string, 0, len(s1)+len(s2))
	res = append(res, s1...)
	res = append(res, s2...)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(res), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}

func write(path string, data []string) {
	fmt.Printf("Writing %d lines to %s...\n", len(data), path)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, s := range data {
		w.WriteString(s + "\n")
	}
	w.Flush()
}

func genTestCases(path string, common, diff1, diff2 []string) {
	fmt.Printf("Generating %d test cases...\n", TestCaseCount)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString("code,expected\n")

	targetTrue := TestCaseCount / 2
	for i := 0; i < targetTrue && i < len(common); i++ {
		w.WriteString(common[i] + ",true\n")
	}

	targetFalse1 := TestCaseCount / 4
	for i := 0; i < targetFalse1 && i < len(diff1); i++ {
		w.WriteString(diff1[i] + ",false\n")
	}

	targetFalse2 := TestCaseCount / 4
	for i := 0; i < targetFalse2 && i < len(diff2); i++ {
		w.WriteString(diff2[i] + ",false\n")
	}

	w.Flush()
	fmt.Println("Generated test cases.")
}
