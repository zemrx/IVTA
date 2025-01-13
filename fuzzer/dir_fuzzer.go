package fuzzer

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

func FuzzDirectories(targetURL string, wordlistFile string, concurrency int, totalWords int64) []string {
	file, err := os.Open(wordlistFile)
	if err != nil {
		log.Fatalf("Failed to open wordlist file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	var validPaths []string
	var mutex sync.Mutex

	var wordCounter int64

	fmt.Printf("Starting directory fuzzing on: %s\n", targetURL)
	fmt.Printf("Total words in wordlist: %d\n", totalWords)

	for scanner.Scan() {
		word := scanner.Text()
		fullURL := fmt.Sprintf("%s/%s", targetURL, word)

		wg.Add(1)
		semaphore <- struct{}{}

		go func(url string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error requesting %s: %v", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Printf("\n[+] Found: %s (Status: %d)\n", url, resp.StatusCode)
				mutex.Lock()
				validPaths = append(validPaths, url)
				mutex.Unlock()
			}

			atomic.AddInt64(&wordCounter, 1)
			fmt.Printf("\rProgress: %d/%d words processed", atomic.LoadInt64(&wordCounter), totalWords)
		}(fullURL)
	}

	wg.Wait()
	fmt.Println("\nDirectory fuzzing completed.")
	return validPaths
}
