package miner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	defaultResponseTimeThreshold  = 1.3
	defaultContentLengthThreshold = 100
	flagScoreThreshold            = 2
)

var (
	tagRegex   = regexp.MustCompile(`<.*?>`)
	jsVarRegex = regexp.MustCompile(`\b(var|let|const)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
)

type ResponseFactors struct {
	SameCode          int
	SameBody          string
	SamePlaintext     string
	LinesNum          int
	LinesDiff         []string
	SameHeaders       []string
	SameRedirect      string
	ParamMissing      []string
	ValueMissing      bool
	ResponseTimeDiff  float64
	ContentLengthDiff int
	HeaderChanges     map[string]string
	JavaScriptVars    []string
}

type RequestOptions struct {
	Method  string
	Headers map[string]string
	Data    map[string]string
	Verbose bool
}

func verbosePrintf(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format, args...)
	}
}

func AnalyzeParameter(ctx context.Context, targetURL, param, value string, wordlist []string, options RequestOptions) (ResponseFactors, error) {
	var factors ResponseFactors

	verbosePrintf(options.Verbose, "Sending first request to %s with options: %+v\n", targetURL, options)
	startTime := time.Now()
	resp1, err := DoRequest(ctx, targetURL, options)
	if err != nil {
		return factors, fmt.Errorf("error making initial request: %w", err)
	}
	defer resp1.Body.Close()
	responseTime1 := time.Since(startTime).Seconds()
	verbosePrintf(options.Verbose, "Received first response with status %d in %f seconds\n", resp1.StatusCode, responseTime1)

	modifiedOptions := options
	if modifiedOptions.Data == nil {
		modifiedOptions.Data = make(map[string]string)
	}
	modifiedOptions.Data[param] = value

	verbosePrintf(options.Verbose, "Sending second request with modified parameter '%s=%s'\n", param, value)
	startTime = time.Now()
	resp2, err := DoRequest(ctx, targetURL, modifiedOptions)
	if err != nil {
		return factors, fmt.Errorf("error making second request: %w", err)
	}
	defer resp2.Body.Close()
	responseTime2 := time.Since(startTime).Seconds()
	verbosePrintf(options.Verbose, "Received second response with status %d in %f seconds\n", resp2.StatusCode, responseTime2)

	factors = compareResponses(resp1, resp2, param, value, wordlist, responseTime1, responseTime2, options.Verbose)
	return factors, nil
}

func DoRequest(ctx context.Context, targetURL string, options RequestOptions) (*http.Response, error) {
	method := strings.ToUpper(options.Method)
	var req *http.Request
	var err error

	switch method {
	case "GET":
		req, err = http.NewRequestWithContext(ctx, "GET", targetURL, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		for key, val := range options.Data {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	case "POST":
		formData := url.Values{}
		for key, val := range options.Data {
			formData.Add(key, val)
		}
		req, err = http.NewRequestWithContext(ctx, "POST", targetURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case "JSON":
		jsonData, err := json.Marshal(options.Data)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	case "XML":
		xmlContent, ok := options.Data["xml"]
		if !ok {
			return nil, fmt.Errorf("missing xml content in options.Data")
		}
		req, err = http.NewRequestWithContext(ctx, "POST", targetURL, strings.NewReader(xmlContent))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/xml")
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", options.Method)
	}

	for key, val := range options.Headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	verbosePrintf(options.Verbose, "Executing HTTP request: %s %s\n", req.Method, req.URL.String())
	return client.Do(req)
}

func compareResponses(resp1, resp2 *http.Response, param, value string, wordlist []string, time1, time2 float64, verbose bool) ResponseFactors {
	factors := ResponseFactors{
		HeaderChanges: make(map[string]string),
	}

	verbosePrintf(verbose, "Comparing status codes: %d vs %d\n", resp1.StatusCode, resp2.StatusCode)
	if resp1.StatusCode == resp2.StatusCode {
		factors.SameCode = resp1.StatusCode
	}

	factors.SameHeaders = make([]string, 0, len(resp1.Header))
	for key, values1 := range resp1.Header {
		factors.SameHeaders = append(factors.SameHeaders, key)
		value2 := resp2.Header.Get(key)
		if len(values1) > 0 && values1[0] != value2 {
			factors.HeaderChanges[key] = fmt.Sprintf("%s -> %s", values1[0], value2)
			verbosePrintf(verbose, "Header '%s' changed: %s -> %s\n", key, values1[0], value2)
		}
	}

	body1, err1 := io.ReadAll(resp1.Body)
	body2, err2 := io.ReadAll(resp2.Body)
	if err1 != nil || err2 != nil {
		verbosePrintf(verbose, "Error reading response bodies: err1=%v, err2=%v\n", err1, err2)
		return factors
	}

	if bytes.Equal(body1, body2) {
		factors.SameBody = string(body1)
		verbosePrintf(verbose, "Response bodies are identical\n")
	} else {
		plaintext1 := removeHTMLTags(string(body1))
		plaintext2 := removeHTMLTags(string(body2))
		if plaintext1 == plaintext2 {
			factors.SamePlaintext = plaintext1
			verbosePrintf(verbose, "Response plaintexts are identical after removing HTML tags\n")
		}
	}

	baselineContainsParam := strings.Contains(string(body1), param)
	alteredContainsParam := strings.Contains(string(body2), param)
	if baselineContainsParam && !alteredContainsParam {
		factors.ParamMissing = wordlist
		verbosePrintf(verbose, "Parameter '%s' missing in second response\n", param)
	}

	baselineContainsValue := strings.Contains(string(body1), value)
	alteredContainsValue := strings.Contains(string(body2), value)
	if baselineContainsValue && !alteredContainsValue {
		factors.ValueMissing = true
		verbosePrintf(verbose, "Value '%s' missing in second response\n", value)
	}

	location1 := resp1.Header.Get("Location")
	location2 := resp2.Header.Get("Location")
	if location1 != "" && location2 != "" && location1 == location2 {
		factors.SameRedirect = location1
		verbosePrintf(verbose, "Both responses redirect to %s\n", location1)
	}

	factors.ResponseTimeDiff = time2 - time1
	factors.ContentLengthDiff = len(body2) - len(body1)
	verbosePrintf(verbose, "Response time difference: %f seconds; Content length difference: %d bytes\n", factors.ResponseTimeDiff, factors.ContentLengthDiff)
	factors.JavaScriptVars = extractJSVariables(string(body2))
	if len(factors.JavaScriptVars) > 0 {
		verbosePrintf(verbose, "Extracted JavaScript variables: %v\n", factors.JavaScriptVars)
	}

	return factors
}

func removeHTMLTags(html string) string {
	return tagRegex.ReplaceAllString(html, "")
}

func extractJSVariables(html string) []string {
	matches := jsVarRegex.FindAllStringSubmatch(html, -1)
	var variables []string
	for _, match := range matches {
		if len(match) > 2 {
			variables = append(variables, match[2])
		}
	}
	return variables
}

func isFlagged(factors ResponseFactors) bool {
	score := 0
	if factors.ValueMissing {
		score++
	}
	if len(factors.ParamMissing) > 0 {
		score++
	}
	if math.Abs(factors.ResponseTimeDiff) > factors.ResponseTimeDiff*defaultResponseTimeThreshold {
		score++
	}
	if math.Abs(float64(factors.ContentLengthDiff)) > defaultContentLengthThreshold {
		score++
	}
	return score >= flagScoreThreshold
}

func BruteForce(ctx context.Context, targetURL string, wordlist []string, options RequestOptions, concurrency int) ([]string, error) {
	var results []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	jobs := make(chan string, len(wordlist))
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for param := range jobs {
				verbosePrintf(options.Verbose, "Testing parameter: %s\n", param)
				factors, err := AnalyzeParameter(ctx, targetURL, param, "test-value", wordlist, options)
				if err != nil {
					verbosePrintf(options.Verbose, "Error testing parameter '%s': %v\n", param, err)
					continue
				}
				if isFlagged(factors) {
					verbosePrintf(options.Verbose, "Parameter '%s' flagged. Factors: %+v\n", param, factors)
					mu.Lock()
					results = append(results, param)
					mu.Unlock()
				}
			}
		}()
	}

	for _, param := range wordlist {
		select {
		case jobs <- param:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results, ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()

	return results, nil
}
