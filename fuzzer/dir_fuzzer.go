package fuzzer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type FuzzResult struct {
	URL           string
	StatusCode    int
	ContentLength int
	WordCount     int
	LineCount     int
}

type FuzzOptions struct {
	Depth                int
	BlacklistStatusCodes []int
	BlacklistLengths     []int
	BlacklistWordCounts  []int
	BlacklistLineCounts  []int
	BlacklistSearchWords []string
	BlacklistRegex       []string
}

func readWordlist(wordlistFile string) ([]string, error) {
	file, err := os.Open(wordlistFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			words = append(words, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

func FuzzDirectories(targetURL string, wordlistFile string, concurrency int, options FuzzOptions) []FuzzResult {
	words, err := readWordlist(wordlistFile)
	if err != nil {
		log.Fatalf("Failed to read wordlist file: %v", err)
	}
	totalWords := len(words)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	var results []FuzzResult
	var mutex sync.Mutex
	var wordCounter int64

	fmt.Printf("Starting directory fuzzing on: %s\n", targetURL)
	fmt.Printf("Total words in wordlist: %d\n", totalWords)

	var fuzz func(baseURL string, depth int)
	fuzz = func(baseURL string, depth int) {
		if depth > options.Depth {
			return
		}
		for _, word := range words {
			fullURL := fmt.Sprintf("%s/%s", baseURL, word)
			wg.Add(1)
			semaphore <- struct{}{}

			go func(url string, currentDepth int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				client := http.Client{
					Timeout: 10 * time.Second,
				}
				resp, err := client.Get(url)
				if err != nil {
					log.Printf("Error requesting %s: %v", url, err)
					atomic.AddInt64(&wordCounter, 1)
					return
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Error reading response body for %s: %v", url, err)
					atomic.AddInt64(&wordCounter, 1)
					return
				}
				bodyStr := string(body)

				if !passesBlacklistFilters(resp, bodyStr, options) {
					atomic.AddInt64(&wordCounter, 1)
					return
				}

				wordCount := len(strings.Fields(bodyStr))
				lines := strings.Split(bodyStr, "\n")
				nonEmptyLines := 0
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						nonEmptyLines++
					}
				}

				result := FuzzResult{
					URL:           url,
					StatusCode:    resp.StatusCode,
					ContentLength: len(bodyStr),
					WordCount:     wordCount,
					LineCount:     nonEmptyLines,
				}

				fmt.Printf("\n[+] Found: %s (Status: %d, Length: %d, Words: %d, Lines: %d)\n",
					result.URL, result.StatusCode, result.ContentLength, result.WordCount, result.LineCount)

				mutex.Lock()
				results = append(results, result)
				mutex.Unlock()

				if currentDepth < options.Depth {
					fuzz(url, currentDepth+1)
				}

				atomic.AddInt64(&wordCounter, 1)
				fmt.Printf("\rProgress: %d/%d words processed", atomic.LoadInt64(&wordCounter), totalWords)
			}(fullURL, depth)
		}
	}

	fuzz(targetURL, 1)
	wg.Wait()
	fmt.Println("\nDirectory fuzzing completed.")
	return results
}

func passesBlacklistFilters(resp *http.Response, body string, options FuzzOptions) bool {
	for _, code := range options.BlacklistStatusCodes {
		if resp.StatusCode == code {
			return false
		}
	}

	bodyLen := len(body)
	for _, blen := range options.BlacklistLengths {
		if bodyLen == blen {
			return false
		}
	}

	words := strings.Fields(body)
	wordCount := len(words)
	for _, cnt := range options.BlacklistWordCounts {
		if wordCount == cnt {
			return false
		}
	}

	lines := strings.Split(body, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}
	for _, cnt := range options.BlacklistLineCounts {
		if nonEmptyLines == cnt {
			return false
		}
	}

	for _, word := range options.BlacklistSearchWords {
		if strings.Contains(body, word) {
			return false
		}
	}

	for _, pattern := range options.BlacklistRegex {
		re, err := regexp.Compile(pattern)
		if err == nil && re.MatchString(body) {
			return false
		}
	}

	return true
}
