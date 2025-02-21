package miner

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	reWords            = regexp.MustCompile(`[A-Za-z][A-Za-z0-9_]*`)
	reNotJunk          = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	reInputs           = regexp.MustCompile(`(?i)<(?:input|textarea|select)[^>]+?(?:id|name)=[\"']?([^\"'\s>]+)`)
	reJSVarDeclaration = regexp.MustCompile(`\b(?:var|let|const)\s+(\w+)\s*=`)
	reEmptyVars        = regexp.MustCompile(`(?:[;\n]|\bvar|\blet|\bconst)\s+(\w+)\s*=\s*(?:["']{1,2}|true|false|null)`)
	reMapKeys          = regexp.MustCompile(`[\"'](\w+?)[\"']\s*:\s*[\"']`)
)

func isNotJunk(param string) bool {
	return reNotJunk.MatchString(param)
}

func extractKeysFromJSON(data interface{}) []string {
	keys := []string{}
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			keys = append(keys, k)
			keys = append(keys, extractKeysFromJSON(val)...)
		}
	case []interface{}:
		for _, item := range v {
			keys = append(keys, extractKeysFromJSON(item)...)
		}
	}
	return keys
}

func ExtractPotentialParams(response string, headers http.Header, wordlist []string) ([]string, bool) {
	wordsExist := false
	var potentialParams []string

	contentType := headers.Get("Content-Type")
	lowerResponse := strings.ToLower(response)

	if strings.HasPrefix(contentType, "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
			jsonKeys := extractKeysFromJSON(jsonData)
			potentialParams = append(potentialParams, jsonKeys...)
			wordsExist = len(jsonKeys) > 0
		}
	} else if strings.HasPrefix(contentType, "text/plain") {
		if len(response) < 200 {
			if (strings.Contains(lowerResponse, "required") ||
				strings.Contains(lowerResponse, "missing") ||
				strings.Contains(lowerResponse, "not found") ||
				strings.Contains(lowerResponse, "requires")) &&
				(strings.Contains(lowerResponse, "param") ||
					strings.Contains(lowerResponse, "parameter") ||
					strings.Contains(lowerResponse, "field")) {
				wordsExist = true
			}
			potentialParams = append(potentialParams, reWords.FindAllString(response, -1)...)
		}
	}

	inputMatches := reInputs.FindAllStringSubmatch(response, -1)
	for _, m := range inputMatches {
		if len(m) > 1 {
			potentialParams = append(potentialParams, m[1])
		}
	}

	jsVarMatches := reJSVarDeclaration.FindAllStringSubmatch(response, -1)
	for _, m := range jsVarMatches {
		if len(m) > 1 {
			potentialParams = append(potentialParams, m[1])
		}
	}

	emptyVarMatches := reEmptyVars.FindAllStringSubmatch(response, -1)
	for _, m := range emptyVarMatches {
		if len(m) > 1 {
			potentialParams = append(potentialParams, m[1])
		}
	}

	mapKeyMatches := reMapKeys.FindAllStringSubmatch(response, -1)
	for _, m := range mapKeyMatches {
		if len(m) > 1 {
			potentialParams = append(potentialParams, m[1])
		}
	}

	found := make(map[string]struct{})
	var result []string
	for _, word := range potentialParams {
		if isNotJunk(word) {
			if _, exists := found[word]; !exists {
				found[word] = struct{}{}
				for i, w := range wordlist {
					if w == word {
						wordlist = append(wordlist[:i], wordlist[i+1:]...)
						break
					}
				}
				wordlist = append([]string{word}, wordlist...)
				result = append(result, word)
			}
		}
	}

	if len(result) > 0 {
		wordsExist = true
	}

	return result, wordsExist
}

func ExtractParamsFromBaseline(resp *http.Response, wordlist []string) ([]string, bool, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	responseText := string(body)
	params, wordsExist := ExtractPotentialParams(responseText, resp.Header, wordlist)
	return params, wordsExist, nil
}
