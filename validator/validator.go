package validator

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 1 * time.Second,
		}).DialContext,
	},
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type ReflectionContext string

const (
	ContextHTML       ReflectionContext = "HTML Body"
	ContextAttribute  ReflectionContext = "HTML Attribute"
	ContextJavaScript ReflectionContext = "JavaScript"
	ContextURL        ReflectionContext = "URL/Href"
	ContextCSS        ReflectionContext = "CSS"
	ContextComment    ReflectionContext = "HTML Comment"
	ContextUnknown    ReflectionContext = "Unknown"
)

type RiskLevel string

const (
	RiskCritical RiskLevel = "CRITICAL"
	RiskHigh     RiskLevel = "HIGH"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskLow      RiskLevel = "LOW"
	RiskInfo     RiskLevel = "INFO"
)

type Finding struct {
	Parameter        string
	Context          ReflectionContext
	UnfilteredChars  []string
	RiskLevel        RiskLevel
	RiskDescription  string
	ExploitScenario  string
	ReflectionSample string
}

func IdentifyReflectedParams(targetURL string) ([]string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	if len(parsedURL.Query()) == 0 {
		fmt.Println("No parameters found in URL. Please provide a URL with parameters.")
		return nil, nil
	}

	var results []string
	for param := range parsedURL.Query() {
		originalValue := parsedURL.Query().Get(param)
		if originalValue == "" {
			continue
		}

		body, err := fetchResponseBody(targetURL)
		if err != nil {
			continue
		}

		if !strings.Contains(body, originalValue) {
			continue
		}

		findings := analyzeParameter(targetURL, param, originalValue, body)
		if len(findings) > 0 {
			results = append(results, param)
			printFindings(targetURL, param, findings)
		}
	}

	return results, nil
}

func fetchResponseBody(targetURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return "", fmt.Errorf("redirect response")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func analyzeParameter(targetURL, param, originalValue, body string) []Finding {
	var findings []Finding

	contexts := detectReflectionContexts(body, originalValue)

	unfilteredChars := testCharacterInjection(targetURL, param)

	for _, ctx := range contexts {
		finding := assessRisk(param, ctx, unfilteredChars, body, originalValue)
		findings = append(findings, finding)
	}

	return findings
}

func detectReflectionContexts(body, value string) []ReflectionContext {
	var contexts []ReflectionContext
	contextMap := make(map[ReflectionContext]bool)

	escapedValue := regexp.QuoteMeta(value)

	htmlBodyPattern := regexp.MustCompile(`>([^<]*` + escapedValue + `[^<]*)<`)
	if htmlBodyPattern.MatchString(body) {
		contextMap[ContextHTML] = true
	}

	attrPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<[^>]+\s+\w+\s*=\s*"[^"]*` + escapedValue + `[^"]*"`),
		regexp.MustCompile(`<[^>]+\s+\w+\s*=\s*'[^']*` + escapedValue + `[^']*'`),
		regexp.MustCompile(`<[^>]+\s+\w+\s*=\s*` + escapedValue),
	}
	for _, pattern := range attrPatterns {
		if pattern.MatchString(body) {
			contextMap[ContextAttribute] = true
			break
		}
	}

	jsPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<script[^>]*>[\s\S]*` + escapedValue + `[\s\S]*</script>`),
		regexp.MustCompile(`\bonon\w+\s*=\s*["']?[^"']*` + escapedValue),
		regexp.MustCompile(`javascript:[^"']*` + escapedValue),
	}
	for _, pattern := range jsPatterns {
		if pattern.MatchString(body) {
			contextMap[ContextJavaScript] = true
			break
		}
	}

	urlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`href\s*=\s*["']?[^"']*` + escapedValue),
		regexp.MustCompile(`src\s*=\s*["']?[^"']*` + escapedValue),
		regexp.MustCompile(`action\s*=\s*["']?[^"']*` + escapedValue),
	}
	for _, pattern := range urlPatterns {
		if pattern.MatchString(body) {
			contextMap[ContextURL] = true
			break
		}
	}

	cssPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<style[^>]*>[\s\S]*` + escapedValue + `[\s\S]*</style>`),
		regexp.MustCompile(`style\s*=\s*["'][^"']*` + escapedValue),
	}
	for _, pattern := range cssPatterns {
		if pattern.MatchString(body) {
			contextMap[ContextCSS] = true
			break
		}
	}

	commentPattern := regexp.MustCompile(`<!--[\s\S]*` + escapedValue + `[\s\S]*-->`)
	if commentPattern.MatchString(body) {
		contextMap[ContextComment] = true
	}

	for ctx := range contextMap {
		contexts = append(contexts, ctx)
	}

	if len(contexts) == 0 && strings.Contains(body, value) {
		contexts = append(contexts, ContextUnknown)
	}

	return contexts
}

func assessRisk(param string, context ReflectionContext, unfilteredChars []string, body, value string) Finding {
	finding := Finding{
		Parameter:       param,
		Context:         context,
		UnfilteredChars: unfilteredChars,
	}

	finding.ReflectionSample = extractReflectionSample(body, value)

	switch context {
	case ContextHTML:
		finding.RiskLevel, finding.RiskDescription, finding.ExploitScenario = assessHTMLRisk(unfilteredChars)
	case ContextAttribute:
		finding.RiskLevel, finding.RiskDescription, finding.ExploitScenario = assessAttributeRisk(unfilteredChars)
	case ContextJavaScript:
		finding.RiskLevel, finding.RiskDescription, finding.ExploitScenario = assessJavaScriptRisk(unfilteredChars)
	case ContextURL:
		finding.RiskLevel, finding.RiskDescription, finding.ExploitScenario = assessURLRisk(unfilteredChars)
	case ContextCSS:
		finding.RiskLevel, finding.RiskDescription, finding.ExploitScenario = assessCSSRisk(unfilteredChars)
	case ContextComment:
		finding.RiskLevel, finding.RiskDescription, finding.ExploitScenario = assessCommentRisk(unfilteredChars)
	default:
		finding.RiskLevel = RiskInfo
		finding.RiskDescription = "Parameter is reflected but context is unclear"
		finding.ExploitScenario = "Manual analysis required"
	}

	return finding
}

func assessHTMLRisk(chars []string) (RiskLevel, string, string) {
	hasAngleBrackets := contains(chars, "<") && contains(chars, ">")

	if hasAngleBrackets {
		return RiskCritical, "XSS via HTML injection", "Attacker can inject arbitrary HTML tags like <script>, <img>, <iframe>"
	}
	return RiskLow, "Limited HTML injection risk", "Angle brackets are filtered, reducing XSS risk"
}

func assessAttributeRisk(chars []string) (RiskLevel, string, string) {
	hasQuotes := contains(chars, "\"") || contains(chars, "'")
	hasAngleBrackets := contains(chars, "<") && contains(chars, ">")

	if hasQuotes && hasAngleBrackets {
		return RiskCritical, "XSS via attribute breakout", "Attacker can break out of attribute and inject new tags or event handlers"
	}
	if hasQuotes {
		return RiskHigh, "XSS via event handler injection", "Attacker can inject event handlers like onload, onerror, onclick"
	}
	return RiskMedium, "Attribute injection with limited impact", "Cannot easily break out of attribute context"
}

func assessJavaScriptRisk(chars []string) (RiskLevel, string, string) {
	hasQuotes := contains(chars, "\"") || contains(chars, "'")
	hasSemicolon := contains(chars, ";")

	if hasQuotes || hasSemicolon {
		return RiskCritical, "JavaScript injection", "Attacker can break out of string context and execute arbitrary JavaScript"
	}
	return RiskMedium, "Limited JavaScript injection", "Difficult to break out of current context"
}

func assessURLRisk(chars []string) (RiskLevel, string, string) {
	hasQuotes := contains(chars, "\"") || contains(chars, "'")

	if hasQuotes {
		return RiskHigh, "Open redirect or XSS via URL manipulation", "Attacker can inject javascript: URLs or redirect to malicious sites"
	}
	return RiskMedium, "URL manipulation possible", "May allow redirects but limited XSS risk"
}

func assessCSSRisk(chars []string) (RiskLevel, string, string) {
	hasSemicolon := contains(chars, ";")
	hasParens := contains(chars, "(") && contains(chars, ")")

	if hasSemicolon || hasParens {
		return RiskMedium, "CSS injection", "Attacker may inject CSS properties or use expression() for XSS in older browsers"
	}
	return RiskLow, "Limited CSS injection", "Difficult to exploit in modern browsers"
}

func assessCommentRisk(chars []string) (RiskLevel, string, string) {
	hasCommentEnd := contains(chars, ">")

	if hasCommentEnd {
		return RiskMedium, "HTML comment breakout", "Attacker can break out of comment and inject HTML"
	}
	return RiskLow, "Reflected in comment", "Low risk as content is commented out"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractReflectionSample(body, value string) string {
	index := strings.Index(body, value)
	if index == -1 {
		return ""
	}

	start := index - 50
	if start < 0 {
		start = 0
	}
	end := index + len(value) + 50
	if end > len(body) {
		end = len(body)
	}

	sample := body[start:end]
	sample = strings.ReplaceAll(sample, "\n", " ")
	sample = strings.ReplaceAll(sample, "\t", " ")
	sample = regexp.MustCompile(`\s+`).ReplaceAllString(sample, " ")

	return "..." + sample + "..."
}

func printFindings(targetURL, param string, findings []Finding) {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("URL: %s\n", targetURL)
	fmt.Printf("Parameter: %s\n", param)
	fmt.Printf("Unfiltered Characters: %v\n", findings[0].UnfilteredChars)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for i, finding := range findings {
		fmt.Printf("\n[Finding #%d]\n", i+1)
		fmt.Printf("  Context: %s\n", finding.Context)
		fmt.Printf("  Risk Level: %s\n", finding.RiskLevel)
		fmt.Printf("  Description: %s\n", finding.RiskDescription)
		fmt.Printf("  Exploit Scenario: %s\n", finding.ExploitScenario)
		if finding.ReflectionSample != "" {
			fmt.Printf("  Sample: %s\n", finding.ReflectionSample)
		}
	}
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func getResponseWithParam(targetURL, param string) (string, string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", "", err
	}

	originalValue := parsedURL.Query().Get(param)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	return string(bodyBytes), originalValue, nil
}

func checkReflected(targetURL string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
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
			if value != "" && strings.Contains(body, value) {
				reflectedParams = append(reflectedParams, key)
				break
			}
		}
	}
	return reflectedParams, nil
}

func testCharacterInjection(targetURL, param string) []string {
	specialChars := []string{`"`, `'`, "<", ">", "$", "|", "(", ")", "`", ":", ";", "{", "}"}
	var unfilteredChars []string

	for _, char := range specialChars {
		testSuffix := "aprefix" + char + "asuffix"
		reflected, err := checkAppend(targetURL, param, testSuffix)
		if err != nil {
			continue
		}
		if reflected {
			unfilteredChars = append(unfilteredChars, char)
		}
	}

	return unfilteredChars
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
