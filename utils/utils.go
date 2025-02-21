package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func CountLines(filename string) int64 {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open wordlist file: %v", err)
	}
	defer file.Close()

	var lineCount int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading wordlist file: %v", err)
	}

	return lineCount
}

func MergeUnique(a, b []string) []string {
	combined := append(a, b...)
	uniqueMap := make(map[string]bool, len(combined))
	uniqueList := make([]string, 0, len(combined))

	for _, item := range combined {
		if !uniqueMap[item] {
			uniqueMap[item] = true
			uniqueList = append(uniqueList, item)
		}
	}
	return uniqueList
}
func ReadTargetList(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		target := scanner.Text()
		if target != "" {
			targets = append(targets, target)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}
func ReadWordlist(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open wordlist file: %v", err)
	}
	defer file.Close()

	var wordlist []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wordlist = append(wordlist, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading wordlist file: %v", err)
	}

	return wordlist, nil
}

func ParseIntSlice(s string) []int {
	var result []int
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		result = append(result, n)
	}
	return result
}
