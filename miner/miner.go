package miner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
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
}

func Define(targetURL, param, value string, wordlist []string, options RequestOptions) (ResponseFactors, error) {
	var factors ResponseFactors

	startTime := time.Now()
	resp1, err := makeMineRequest(targetURL, options)
	if err != nil {
		return factors, fmt.Errorf("error making initial request: %w", err)
	}
	defer resp1.Body.Close()
	responseTime1 := time.Since(startTime).Seconds()

	modifiedOptions := options
	if modifiedOptions.Data == nil {
		modifiedOptions.Data = make(map[string]string)
	}
	modifiedOptions.Data[param] = value

	startTime = time.Now()
	resp2, err := makeMineRequest(targetURL, modifiedOptions)
	if err != nil {
		return factors, fmt.Errorf("error making second request: %w", err)
	}
	defer resp2.Body.Close()
	responseTime2 := time.Since(startTime).Seconds()

	factors = compareResponses(resp1, resp2, param, value, wordlist, responseTime1, responseTime2)
	return factors, nil
}

func makeMineRequest(targetURL string, options RequestOptions) (*http.Response, error) {
	var req *http.Request
	var err error

	method := strings.ToUpper(options.Method)
	switch method {
	case "GET":
		req, err = http.NewRequest("GET", targetURL, nil)
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
		req, err = http.NewRequest("POST", targetURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case "JSON":
		jsonData, err := json.Marshal(options.Data)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	case "XML":
		xmlContent, ok := options.Data["xml"]
		if !ok {
			return nil, fmt.Errorf("missing xml content in options.Data")
		}
		req, err = http.NewRequest("POST", targetURL, strings.NewReader(xmlContent))
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}

	return resp, nil
}

func compareResponses(resp1, resp2 *http.Response, param, value string, wordlist []string, time1, time2 float64) ResponseFactors {
	factors := ResponseFactors{
		HeaderChanges: make(map[string]string),
	}

	if resp1.StatusCode == resp2.StatusCode {
		factors.SameCode = resp1.StatusCode
	}

	factors.SameHeaders = make([]string, 0, len(resp1.Header))
	for key, values1 := range resp1.Header {
		factors.SameHeaders = append(factors.SameHeaders, key)
		value2 := resp2.Header.Get(key)
		if len(values1) > 0 && values1[0] != value2 {
			factors.HeaderChanges[key] = fmt.Sprintf("%s -> %s", values1[0], value2)
		}
	}

	body1, err1 := io.ReadAll(resp1.Body)
	body2, err2 := io.ReadAll(resp2.Body)
	if err1 != nil || err2 != nil {
		return factors
	}

	if bytes.Equal(body1, body2) {
		factors.SameBody = string(body1)
	} else {
		plaintext1 := removeTags(string(body1))
		plaintext2 := removeTags(string(body2))
		if plaintext1 == plaintext2 {
			factors.SamePlaintext = plaintext1
		}
	}

	if !strings.Contains(string(body2), param) {
		factors.ParamMissing = wordlist
	}
	if !strings.Contains(string(body2), value) {
		factors.ValueMissing = true
	}

	location1 := resp1.Header.Get("Location")
	location2 := resp2.Header.Get("Location")
	if location1 != "" && location2 != "" && location1 == location2 {
		factors.SameRedirect = location1
	}

	factors.ResponseTimeDiff = time2 - time1
	factors.ContentLengthDiff = len(body2) - len(body1)
	factors.JavaScriptVars = extractJavaScriptVariables(string(body2))

	return factors
}

func removeTags(html string) string {
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(html, "")
}

func extractJavaScriptVariables(html string) []string {
	re := regexp.MustCompile(`\b(var|let|const)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	matches := re.FindAllStringSubmatch(html, -1)
	var variables []string
	for _, match := range matches {
		variables = append(variables, match[2])
	}
	return variables
}

func BruteForce(targetURL string, wordlist []string, options RequestOptions, concurrency int) ([]string, error) {
	var results []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	jobs := make(chan string, len(wordlist))
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for param := range jobs {
				factors, err := Define(targetURL, param, "test-value", wordlist, options)
				if err != nil {
					continue
				}
				if factors.ValueMissing || len(factors.ParamMissing) > 0 || factors.ResponseTimeDiff > 1.0 || factors.ContentLengthDiff > 100 {
					mu.Lock()
					results = append(results, param)
					mu.Unlock()
				}
			}
		}()
	}

	for _, param := range wordlist {
		jobs <- param
	}
	close(jobs)
	wg.Wait()

	return results, nil
}
