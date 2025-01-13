package fuzz

import (
	"flag"
	"fmt"
	"log"

	"ivta/config"
	"ivta/fuzzer"
)

func Execute() {
	cfg := config.FuzzConfig{}

	flag.StringVar(&cfg.TargetURL, "u", "", "Target URL for fuzzing (required)")
	flag.StringVar(&cfg.DirWordlistFile, "w", "wordlist.txt", "Path to the directory wordlist file")
	flag.StringVar(&cfg.ParamWordlistFile, "p", "param_wordlist.txt", "Path to the parameter wordlist file")
	flag.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	flag.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	flag.StringVar(&cfg.OutputFile, "o", "fuzz_results.json", "Path to the output file (JSON format)")
	flag.StringVar(&cfg.CustomSymbol, "s", "test", "Custom symbol to test for reflection")

	flag.Parse()

	if cfg.TargetURL == "" {
		log.Fatal("Please provide a target URL using the -u flag")
	}

	validPaths := fuzzer.FuzzDirectories(cfg.TargetURL, cfg.DirWordlistFile, cfg.Concurrency, 0)
	fmt.Printf("Found %d valid paths.\n", len(validPaths))

	validParams := fuzzer.FuzzParameters(cfg.TargetURL, cfg.ParamWordlistFile, cfg.Concurrency, 0, cfg.CustomSymbol)
	fmt.Printf("Found %d valid parameters.\n", len(validParams))

	config.SaveResults(cfg.OutputFile, nil, nil, validPaths, validParams)
	fmt.Println("Results saved to", cfg.OutputFile)
}
