package main

import (
	"fmt"
	mine "ivta/cmd/Param_miner"
	"ivta/cmd/crawl"
	"ivta/cmd/fuzz"
	"ivta/cmd/hybrid"
	"ivta/cmd/validate"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "crawl":
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			crawl.Help()
		} else {
			crawl.Execute()
		}
	case "fuzz":
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			fuzz.Help()
		} else {
			fuzz.Execute()
		}
	case "hybrid":
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			hybrid.Help()
		} else {
			hybrid.Execute()
		}
	case "validate":
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			validate.Help()
		} else {
			validate.Execute()
		}
	case "mine":
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			mine.Help()
		} else {
			mine.Execute()
		}
	case "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: .\\ivta.exe <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  crawl    Run the crawler")
	fmt.Println("  fuzz     Run the fuzzer")
	fmt.Println("  hybrid   Run the hybrid crawler and fuzzer")
	fmt.Println("Options:")
	fmt.Println("  -h, --help   Display this help message")
}
