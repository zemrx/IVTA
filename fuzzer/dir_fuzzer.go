package fuzzer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
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
	Depth                  int
	Extensions             []string
	UserAgent              string
	BlacklistStatusCodes   []int
	BlacklistLengths       []int
	BlacklistWordCounts    []int
	BlacklistLineCounts    []int
	BlacklistSearchWords   []string
	BlacklistRegex         []string
	BlacklistRegexCompiled []*regexp.Regexp
	Verbose                bool
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

func FuzzDirectories(ctx context.Context, targetURL string, wordlistFile string, concurrency int, options FuzzOptions) []FuzzResult {
	words, err := readWordlist(wordlistFile)
	if err != nil {
		log.Fatalf("Failed to read wordlist file: %v", err)
	}

	for i, pattern := range options.BlacklistRegex {
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err == nil {
			options.BlacklistRegexCompiled = append(options.BlacklistRegexCompiled, re)
		} else {
			log.Printf("Invalid blacklist regex %q: %v", pattern, err)
		}
		_ = i
	}

	var fuzzWords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		word = strings.TrimLeft(word, "/")
		if word == "" {
			continue
		}
		fuzzWords = append(fuzzWords, word)
		for _, ext := range options.Extensions {
			if ext != "" {
				if !strings.HasPrefix(ext, ".") {
					ext = "." + ext
				}
				fuzzWords = append(fuzzWords, word+ext)
			}
		}
	}
	words = fuzzWords
	totalWords := len(words)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	var results []FuzzResult
	var mutex sync.Mutex
	var wordCounter int64

	fmt.Printf("Starting directory fuzzing on: %s\n", targetURL)
	fmt.Printf("Total words in wordlist (including extensions): %d\n", totalWords)

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var fuzz func(baseURL string, depth int)
	fuzz = func(baseURL string, depth int) {
		if depth > options.Depth {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, word := range words {
			select {
			case <-ctx.Done():
				return
			default:
			}

			u, err := url.Parse(baseURL)
			if err != nil {
				if options.Verbose {
					log.Printf("Error parsing base URL %s: %v", baseURL, err)
				}
				continue
			}
			segment := url.PathEscape(word)
			u.Path = path.Join(strings.TrimRight(u.Path, "/"), segment)
			fullURL := u.String()

			wg.Add(1)
			semaphore <- struct{}{}

			go func(urlStr string, currentDepth int) {
				atomic.AddInt64(&wordCounter, 1)
				defer wg.Done()
				defer func() { <-semaphore }()

				select {
				case <-ctx.Done():
					return
				default:
				}

				req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
				if err != nil {
					if options.Verbose {
						log.Printf("Error creating request for %s: %v", urlStr, err)
					}
					return
				}

				if options.UserAgent != "" {
					req.Header.Set("User-Agent", options.UserAgent)
				} else {
					req.Header.Set("User-Agent", "IVTA-Fuzzer/1.0")
				}

				resp, err := client.Do(req)
				if err != nil {
					if ctx.Err() == nil && options.Verbose {
						log.Printf("Error requesting %s: %v", urlStr, err)
					}
					return
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					if options.Verbose {
						log.Printf("Error reading response body for %s: %v", urlStr, err)
					}
					return
				}
				bodyStr := string(body)

				if !passesBlacklistFilters(resp, bodyStr, options) {
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
					URL:           urlStr,
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

				ct := resp.Header.Get("Content-Type")
				isHTML := strings.Contains(strings.ToLower(ct), "text/html") || strings.Contains(strings.ToLower(ct), "application/xhtml+xml")

				if currentDepth < options.Depth && (resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 || resp.StatusCode == 403) && isHTML {
					isFile := false
					for _, ext := range options.Extensions {
						if ext != "" && strings.HasSuffix(urlStr, ext) {
							isFile = true
							break
						}
					}
					if !isFile && (strings.Contains(bodyStr, "Index of") || strings.Contains(bodyStr, "<html")) {
						fuzz(urlStr, currentDepth+1)
					}
				}

				current := atomic.LoadInt64(&wordCounter)
				if current%10 == 0 {
					fmt.Printf("\rProgress: %d words processed", current)
				}
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

	bodyLower := strings.ToLower(body)
	for _, word := range options.BlacklistSearchWords {
		if strings.Contains(bodyLower, strings.ToLower(word)) {
			return false
		}
	}

	for _, re := range options.BlacklistRegexCompiled {
		if re.MatchString(body) {
			return false
		}
	}

	return true
}
