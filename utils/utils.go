package utils

import (
	"bufio"
	"log"
	"os"
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
