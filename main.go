package main

import (
	"fmt"
	"io"
	mine "ivta/cmd/Param_miner"
	"ivta/cmd/crawl"
	"ivta/cmd/fuzz"
	"ivta/cmd/hybrid"
	"ivta/cmd/validate"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp(os.Stdout)
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
		printHelp(os.Stdout)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp(os.Stdout)
		os.Exit(1)
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: .\\ivta.exe <command> [options]")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  crawl    Run the crawler")
	fmt.Fprintln(w, "  fuzz     Run the fuzzer")
	fmt.Fprintln(w, "  hybrid   Run the hybrid crawler and fuzzer")
	fmt.Fprintln(w, "  validate Run the validator")
	fmt.Fprintln(w, "  mine     Run the parameter miner")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  -h, --help   Display this help message")
}
