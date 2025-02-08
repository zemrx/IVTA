package validator

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type ParamCheck struct {
	URL   string
	Param string
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 1 * time.Second,
		}).DialContext,
	},
	Timeout: 30 * time.Second,
}

func main() {
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	scanner := bufio.NewScanner(os.Stdin)
	initialChecks := make(chan ParamCheck, 40)

	reflectedChecks := makePool(initialChecks, processReflectedParams)
	appendedChecks := makePool(reflectedChecks, processAppendedCheck)
	done := makePool(appendedChecks, processCharInjection)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			initialChecks <- ParamCheck{URL: line}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
	close(initialChecks)

	for range done {
	}
}

func processReflectedParams(pc ParamCheck, out chan<- ParamCheck) {
	params, err := checkReflected(pc.URL)
	if err != nil {
		log.Printf("checkReflected error for URL %s: %v", pc.URL, err)
		return
	}
	if len(params) == 0 {
		return
	}
	for _, p := range params {
		out <- ParamCheck{URL: pc.URL, Param: p}
	}
}

func processAppendedCheck(pc ParamCheck, out chan<- ParamCheck) {
	const testSuffix = "iy3j4h234hjb23234"
	reflected, err := checkAppend(pc.URL, pc.Param, testSuffix)
	if err != nil {
		log.Printf("checkAppend error for URL %s, param %s: %v", pc.URL, pc.Param, err)
		return
	}
	if reflected {
		out <- pc
	}
}

func processCharInjection(pc ParamCheck, out chan<- ParamCheck) {
	result := []string{pc.URL, pc.Param}
	specialChars := []string{`"`, `'`, "<", ">", "$", "|", "(", ")", "`", ":", ";", "{", "}"}
	for _, char := range specialChars {
		testSuffix := "aprefix" + char + "asuffix"
		reflected, err := checkAppend(pc.URL, pc.Param, testSuffix)
		if err != nil {
			log.Printf("checkAppend error for URL %s, param %s with char %s: %v", pc.URL, pc.Param, char, err)
			continue
		}
		if reflected {
			result = append(result, char)
		}
	}
	if len(result) > 2 {
		fmt.Printf("URL: %s  Param: %s  Unfiltered: %v\n", result[0], result[1], result[2:])
	}
}

func checkReflected(targetURL string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.100 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return nil, nil
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "html") {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	var reflectedParams []string
	for key, values := range parsedURL.Query() {
		for _, value := range values {
			if strings.Contains(body, value) {
				reflectedParams = append(reflectedParams, key)
				break
			}
		}
	}
	return reflectedParams, nil
}

func checkAppend(targetURL, param, suffix string) (bool, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}
	queryValues := parsedURL.Query()
	originalValue := queryValues.Get(param)
	queryValues.Set(param, originalValue+suffix)
	parsedURL.RawQuery = queryValues.Encode()

	reflectedParams, err := checkReflected(parsedURL.String())
	if err != nil {
		return false, err
	}
	for _, p := range reflectedParams {
		if p == param {
			return true, nil
		}
	}
	return false, nil
}

type workerFunc func(ParamCheck, chan<- ParamCheck)

func makePool(input <-chan ParamCheck, fn workerFunc) <-chan ParamCheck {
	var wg sync.WaitGroup
	output := make(chan ParamCheck)
	workerCount := 40

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range input {
				fn(item, output)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}

func IdentifyReflectedParams(targetURL string) ([]string, error) {
	return checkReflected(targetURL)
}
