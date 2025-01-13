package fuzzer

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// CountLines counts the number of lines in a file.
func CountLines(filename string) int64 {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open wordlist file: %v", err)
	}
	defer file.Close()

	var lineCount int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading wordlist file: %v", err)
	}

	return lineCount
}

func FuzzParameters(targetURL string, wordlistFile string, concurrency int, customSymbol string) []string {
	totalWords := CountLines(wordlistFile)
	file, err := os.Open(wordlistFile)
	if err != nil {
		log.Fatalf("Failed to open wordlist file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	var validParams []string
	var mutex sync.Mutex

	var wordCounter int64

	fmt.Printf("Starting parameter fuzzing on: %s\n", targetURL)
	fmt.Printf("Total words in wordlist: %d\n", totalWords)

	for scanner.Scan() {
		param := scanner.Text()
		fullURL := addParamToURL(targetURL, param, customSymbol)

		wg.Add(1)
		semaphore <- struct{}{}

		go func(url string, param string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error requesting %s: %v", url, err)
				return
			}
			defer resp.Body.Close()

			if isReflected(resp, customSymbol) {
				fmt.Printf("\n[+] Found reflected parameter: %s (Status: %d)\n", url, resp.StatusCode)
				mutex.Lock()
				validParams = append(validParams, url)
				mutex.Unlock()
			}

			atomic.AddInt64(&wordCounter, 1)
			fmt.Printf("\rProgress: %d/%d words processed", atomic.LoadInt64(&wordCounter), totalWords)
		}(fullURL, param)
	}

	wg.Wait()
	fmt.Println("\nParameter fuzzing completed.")
	return validParams
}

func addParamToURL(targetURL string, param string, customSymbol string) string {
	if strings.Contains(targetURL, "?") {
		return fmt.Sprintf("%s&%s=%s", targetURL, param, customSymbol)
	}
	return fmt.Sprintf("%s?%s=%s", targetURL, param, customSymbol)
}

func isReflected(resp *http.Response, customSymbol string) bool {
	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	bodyStr := string(body[:n])

	return strings.Contains(bodyStr, customSymbol)
}
