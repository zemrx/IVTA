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

// FuzzResult contains detailed response information for a target URL.
type FuzzResult struct {
	URL           string // The target URL that passed filtering.
	StatusCode    int    // HTTP status code of the response.
	ContentLength int    // Length of the response body (in bytes).
	WordCount     int    // Number of words in the response body.
	LineCount     int    // Number of non-empty lines in the response body.
}

// FuzzOptions holds the blacklist filtering and recursion options for directory fuzzing.
type FuzzOptions struct {
	Depth                int      // How deep to recurse into subdirectories.
	BlacklistStatusCodes []int    // Reject responses with any of these status codes.
	BlacklistLengths     []int    // Reject responses whose body length exactly equals any of these.
	BlacklistWordCounts  []int    // Reject responses whose word count exactly equals any of these.
	BlacklistLineCounts  []int    // Reject responses whose non-empty line count exactly equals any of these.
	BlacklistSearchWords []string // Reject responses that contain any of these words.
	BlacklistRegex       []string // Reject responses whose body matches any of these regex patterns.
}

// readWordlist reads all nonempty lines from the given file.
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

	// Check blacklisted word count.
	words := strings.Fields(body)
	wordCount := len(words)
	for _, cnt := range options.BlacklistWordCounts {
		if wordCount == cnt {
			return false
		}
	}

	// Check blacklisted non-empty line count.
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

	// Check blacklisted search words: if any appear in the body, reject.
	for _, word := range options.BlacklistSearchWords {
		if strings.Contains(body, word) {
			return false
		}
	}

	// Check blacklisted regex patterns.
	for _, pattern := range options.BlacklistRegex {
		re, err := regexp.Compile(pattern)
		if err == nil && re.MatchString(body) {
			return false
		}
	}

	return true
}
