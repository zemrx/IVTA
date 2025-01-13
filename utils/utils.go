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
