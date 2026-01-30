package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/huynhanx03/go-demo/promotion-file/internal/di"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

func main() {
	dataDir := flag.String("data-dir", "data", "Directory containing data files")
	flag.Parse()

	log.Println("Initializing system, please wait...")
	container := di.SetupDependencies(*dataDir)
	service := container.PromotionContainer.Service

	fmt.Println("System Ready! Enter promotion code to check (or 'exit' to quit).")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if text == "" {
			continue
		}

		if len(text) < 1 || len(text) > 5 {
			fmt.Println(ColorRed + "Invalid (Length must be 1-5)" + ColorReset)
			continue
		}

		isValidChar := true
		for _, r := range text {
			if r < 'a' || r > 'z' {
				isValidChar = false
				break
			}
		}
		if !isValidChar {
			fmt.Println(ColorRed + "Invalid (Characters must be a-z)" + ColorReset)
			continue
		}

		isValid := service.IsEligible(text)
		if isValid {
			fmt.Println(ColorGreen + "Valid Code" + ColorReset)
		} else {
			fmt.Println(ColorRed + "Invalid Code" + ColorReset)
		}
	}
}
