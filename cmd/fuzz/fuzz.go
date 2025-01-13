package fuzz

import (
	"fmt"
	"os"

	"ivta/config"
	"ivta/fuzzer"
)

func Execute() {
	if len(os.Args) < 3 {
		Help()
		os.Exit(1)
	}

	cfg := config.LoadFuzzConfig()

	fmt.Println("Target URL:", cfg.TargetURL)

	validPaths := fuzzer.FuzzDirectories(cfg.TargetURL, cfg.DirWordlistFile, cfg.Concurrency)
	fmt.Printf("Found %d valid paths.\n", len(validPaths))

	var validParams []string
	for _, path := range validPaths {
		params := fuzzer.FuzzParameters(path, cfg.ParamWordlistFile, cfg.Concurrency, cfg.CustomSymbol)
		validParams = append(validParams, params...)
	}
	fmt.Printf("Found %d valid parameters.\n", len(validParams))

	config.SaveResults(cfg.OutputFile, nil, nil, nil, validPaths, validParams)
	fmt.Println("Results saved to", cfg.OutputFile)
}

func Help() {
	fmt.Println("Usage: .\\ivta.exe fuzz -u <target_url> [options]")
	fmt.Println("Fuzz a website for directories and parameters.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required)")
	fmt.Println("  -w       Path to the directory wordlist file (default: wordlist.txt)")
	fmt.Println("  -p       Path to the parameter wordlist file (default: param_wordlist.txt)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Path to the output file (default: fuzz_results.json)")
	fmt.Println("  -s       Custom symbol to test for reflection (default: test)")
	fmt.Println("  -h, --help   Display this help message")
}
