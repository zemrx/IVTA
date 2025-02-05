package validator

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var DefaultSymbols = []string{
	"!",
	"\"",
	"#",
	"$",
	"%",
	"&",
	"'",
	"(",
	")",
	"*",
	"+",
	",",
	"-",
	".",
	"/",
	":",
	";",
	"<",
	"=",
	">",
	"?",
	"@",
	"[",
	"\\",
	"]",
	"^",
	"_",
	"`",
	"{",
	"|",
	"}",
	"~",
}

func ValidateParameter(targetURL, param, marker, method string, headers, data map[string]string) (bool, string, error) {
	baselineData := cloneMap(data)
	baselineData[param] = "baseline"

	baselineResp, err := makeRequest(targetURL, method, headers, baselineData)
	if err != nil {
		return false, "", fmt.Errorf("baseline request failed: %w", err)
	}
	baselineBody, err := io.ReadAll(baselineResp.Body)
	baselineResp.Body.Close()
	if err != nil {
		return false, "", fmt.Errorf("reading baseline response failed: %w", err)
	}
	baselineStr := string(baselineBody)

	markerData := cloneMap(data)
	markerData[param] = marker

	markerResp, err := makeRequest(targetURL, method, headers, markerData)
	if err != nil {
		return false, "", fmt.Errorf("marker request failed: %w", err)
	}
	markerBody, err := io.ReadAll(markerResp.Body)
	markerResp.Body.Close()
	if err != nil {
		return false, "", fmt.Errorf("reading marker response failed: %w", err)
	}
	markerStr := string(markerBody)

	if strings.Contains(markerStr, marker) && !strings.Contains(baselineStr, marker) {
		return true, fmt.Sprintf("Parameter '%s' is unsanitized; marker '%s' was reflected.", param, marker), nil
	}
	return false, fmt.Sprintf("Parameter '%s' appears sanitized with marker '%s'.", param, marker), nil
}

func CheckReflectedSymbols(targetURL, param string, symbols []string, method string, headers, data map[string]string) ([]string, error) {
	var reflected []string
	for _, sym := range symbols {
		unsanitized, _, err := ValidateParameter(targetURL, param, sym, method, headers, data)
		if err != nil {
			return nil, err
		}
		if unsanitized {
			reflected = append(reflected, sym)
		}
	}
	return reflected, nil
}

func makeRequest(targetURL, method string, headers, data map[string]string) (*http.Response, error) {
	var req *http.Request
	var err error

	method = strings.ToUpper(method)
	if method == "GET" {
		req, err = http.NewRequest("GET", targetURL, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		for k, v := range data {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	} else {
		formData := url.Values{}
		for k, v := range data {
			formData.Add(k, v)
		}
		req, err = http.NewRequest(method, targetURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return client.Do(req)
}

func cloneMap(original map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range original {
		newMap[k] = v
	}
	return newMap
}
