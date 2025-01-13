package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
)

type CrawlConfig struct {
	TargetURL   string
	MaxDepth    int
	Concurrency int
	Verbose     bool
	OutputFile  string
}

type FuzzConfig struct {
	TargetURL         string
	DirWordlistFile   string
	ParamWordlistFile string
	Concurrency       int
	Verbose           bool
	OutputFile        string
	CustomSymbol      string
}

type HybridConfig struct {
	TargetURL         string
	DirWordlistFile   string
	ParamWordlistFile string
	MaxDepth          int
	Concurrency       int
	Verbose           bool
	OutputFile        string
	CustomSymbol      string
}

func LoadCrawlConfig() CrawlConfig {
	cfg := CrawlConfig{}

	crawlFlagSet := flag.NewFlagSet("crawl", flag.ExitOnError)

	crawlFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for crawling (required)")
	crawlFlagSet.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	crawlFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	crawlFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	crawlFlagSet.StringVar(&cfg.OutputFile, "o", "crawl_results.json", "Path to the output file (JSON format)")

	if err := crawlFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	validateURL(cfg.TargetURL)

	return cfg
}

func LoadFuzzConfig() FuzzConfig {
	cfg := FuzzConfig{}

	fuzzFlagSet := flag.NewFlagSet("fuzz", flag.ExitOnError)

	fuzzFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for fuzzing (required)")
	fuzzFlagSet.StringVar(&cfg.DirWordlistFile, "w", "wordlist.txt", "Path to the directory wordlist file")
	fuzzFlagSet.StringVar(&cfg.ParamWordlistFile, "p", "param_wordlist.txt", "Path to the parameter wordlist file")
	fuzzFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	fuzzFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	fuzzFlagSet.StringVar(&cfg.OutputFile, "o", "fuzz_results.json", "Path to the output file (JSON format)")
	fuzzFlagSet.StringVar(&cfg.CustomSymbol, "s", "test", "Custom symbol to test for reflection")

	if err := fuzzFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	validateURL(cfg.TargetURL)

	return cfg
}

func LoadHybridConfig() HybridConfig {
	cfg := HybridConfig{}

	hybridFlagSet := flag.NewFlagSet("hybrid", flag.ExitOnError)

	hybridFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for hybrid discovery (required)")
	hybridFlagSet.StringVar(&cfg.DirWordlistFile, "w", "wordlist.txt", "Path to the directory wordlist file")
	hybridFlagSet.StringVar(&cfg.ParamWordlistFile, "p", "param_wordlist.txt", "Path to the parameter wordlist file")
	hybridFlagSet.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	hybridFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	hybridFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	hybridFlagSet.StringVar(&cfg.OutputFile, "o", "hybrid_results.json", "Path to the output file (JSON format)")
	hybridFlagSet.StringVar(&cfg.CustomSymbol, "s", "test", "Custom symbol to test for reflection")

	if err := hybridFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	validateURL(cfg.TargetURL)

	return cfg
}

func validateURL(url string) {
	if url == "" {
		log.Fatal("Please provide a target URL using the -u flag")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		log.Fatal("Error: The URL must start with http:// or https://")
	}
}

func SaveResults(outputFile string, sitemapURLs []string, htmlLinks []string, jsLinks []string, validPaths []string, validParams []string) {
	var allLinks []string
	allLinks = append(allLinks, htmlLinks...)
	allLinks = append(allLinks, jsLinks...)
	uniqueLinks := make(map[string]bool)
	for _, link := range allLinks {
		uniqueLinks[link] = true
	}

	var result []string
	for link := range uniqueLinks {
		result = append(result, link)
	}
	results := struct {
		Links []string `json:"links"`
	}{
		Links: result,
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(results)
	if err != nil {
		log.Fatalf("Failed to encode results to JSON: %v", err)
	}
}
