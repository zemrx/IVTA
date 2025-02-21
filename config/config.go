package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
)

type CrawlConfig struct {
	TargetURL      string
	TargetListFile string
	MaxDepth       int
	Concurrency    int
	Verbose        bool
	OutputFile     string
}

type FuzzConfig struct {
	TargetURL            string
	TargetListFile       string
	DirWordlistFile      string
	ParamWordlistFile    string
	MaxDepth             int
	Concurrency          int
	Verbose              bool
	OutputFile           string
	CustomSymbol         string
	BlacklistStatusCodes string
	BlacklistLengths     string
	BlacklistWordCounts  string
	BlacklistLineCounts  string
	BlacklistSearchWords string
	BlacklistRegex       string
}
type HybridConfig struct {
	TargetURL         string
	TargetListFile    string
	DirWordlistFile   string
	ParamWordlistFile string
	MaxDepth          int
	Concurrency       int
	Verbose           bool
	OutputFile        string
	CustomSymbol      string
	Headers           string
	Data              string
	Method            string
}

type MinerConfig struct {
	TargetURL      string
	TargetListFile string
	WordlistFile   string
	Method         string
	Headers        string
	Data           string
	ctx            int
	Concurrency    int
	Verbose        bool
	OutputFile     string
	InjectValue    string
}

type ValidatorConfig struct {
	TargetURL      string
	TargetListFile string
	Concurrency    int
	Verbose        bool
	OutputFile     string
}

func LoadCrawlConfig() CrawlConfig {
	cfg := CrawlConfig{}

	crawlFlagSet := flag.NewFlagSet("crawl", flag.ExitOnError)
	crawlFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for crawling (required if -tl is not used)")
	crawlFlagSet.StringVar(&cfg.TargetListFile, "tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	crawlFlagSet.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	crawlFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	crawlFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	crawlFlagSet.StringVar(&cfg.OutputFile, "o", "crawl_results.json", "Path to the output file (JSON format)")

	if err := crawlFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	if cfg.TargetURL == "" && cfg.TargetListFile == "" {
		log.Fatal("Please provide a target URL using the -u flag or a target list file using the -tl flag")
	}

	if cfg.TargetURL != "" {
		validateURL(cfg.TargetURL)
	}

	return cfg
}
func LoadFuzzConfig() FuzzConfig {
	cfg := FuzzConfig{}

	fuzzFlagSet := flag.NewFlagSet("fuzz", flag.ExitOnError)

	fuzzFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for fuzzing (required if -tl is not used)")
	fuzzFlagSet.StringVar(&cfg.TargetListFile, "tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	fuzzFlagSet.StringVar(&cfg.DirWordlistFile, "w", "wordlist.txt", "Path to the directory wordlist file")
	fuzzFlagSet.StringVar(&cfg.ParamWordlistFile, "p", "param_wordlist.txt", "Path to the parameter wordlist file")
	fuzzFlagSet.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	fuzzFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	fuzzFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	fuzzFlagSet.StringVar(&cfg.OutputFile, "o", "fuzz_results.json", "Path to the output file (JSON format)")
	fuzzFlagSet.StringVar(&cfg.CustomSymbol, "s", "test", "Custom symbol to test for reflection")

	fuzzFlagSet.StringVar(&cfg.BlacklistStatusCodes, "bs", "", "Comma-separated list of blacklisted status codes (e.g. 404,500)")
	fuzzFlagSet.StringVar(&cfg.BlacklistLengths, "bl", "", "Comma-separated list of blacklisted body lengths (in bytes)")
	fuzzFlagSet.StringVar(&cfg.BlacklistWordCounts, "bw", "", "Comma-separated list of blacklisted word counts")
	fuzzFlagSet.StringVar(&cfg.BlacklistLineCounts, "blc", "", "Comma-separated list of blacklisted non-empty line counts")
	fuzzFlagSet.StringVar(&cfg.BlacklistSearchWords, "bsw", "", "Comma-separated list of blacklisted search words")
	fuzzFlagSet.StringVar(&cfg.BlacklistRegex, "br", "", "Comma-separated list of blacklisted regex patterns")

	if err := fuzzFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	if cfg.TargetURL == "" && cfg.TargetListFile == "" {
		log.Fatal("Please provide a target URL using the -u flag or a target list file using the -tl flag")
	}

	if cfg.TargetURL != "" {
		validateURL(cfg.TargetURL)
	}

	return cfg
}
func LoadHybridConfig() HybridConfig {
	cfg := HybridConfig{}

	hybridFlagSet := flag.NewFlagSet("hybrid", flag.ExitOnError)

	hybridFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for hybrid discovery (required if -tl is not used)")
	hybridFlagSet.StringVar(&cfg.TargetListFile, "tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	hybridFlagSet.StringVar(&cfg.DirWordlistFile, "w", "wordlist.txt", "Path to the directory wordlist file")
	hybridFlagSet.StringVar(&cfg.ParamWordlistFile, "p", "param_wordlist.txt", "Path to the parameter wordlist file")
	hybridFlagSet.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	hybridFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	hybridFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	hybridFlagSet.StringVar(&cfg.OutputFile, "o", "hybrid_results.json", "Path to the output file (JSON format)")
	hybridFlagSet.StringVar(&cfg.CustomSymbol, "s", "test", "Custom symbol to test for reflection")
	hybridFlagSet.StringVar(&cfg.Headers, "H", "", "Custom headers (e.g., 'Header1:Value1,Header2:Value2')")
	hybridFlagSet.StringVar(&cfg.Data, "ddata", "", "Custom data (e.g., 'key1:value1,key2:value2')")
	hybridFlagSet.StringVar(&cfg.Method, "m", "GET", "HTTP method (GET, POST, JSON, XML)")

	if err := hybridFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	if cfg.TargetURL == "" && cfg.TargetListFile == "" {
		log.Fatal("Please provide a target URL using the -u flag or a target list file using the -tl flag")
	}

	cfg.TargetURL = strings.TrimSpace(cfg.TargetURL)
	cfg.TargetListFile = strings.TrimSpace(cfg.TargetListFile)
	cfg.DirWordlistFile = strings.TrimSpace(cfg.DirWordlistFile)
	cfg.ParamWordlistFile = strings.TrimSpace(cfg.ParamWordlistFile)
	cfg.OutputFile = strings.TrimSpace(cfg.OutputFile)
	cfg.CustomSymbol = strings.TrimSpace(cfg.CustomSymbol)
	cfg.Headers = strings.TrimSpace(cfg.Headers)
	cfg.Data = strings.TrimSpace(cfg.Data)
	cfg.Method = strings.ToUpper(strings.TrimSpace(cfg.Method))

	if cfg.TargetURL != "" {
		validateURL(cfg.TargetURL)
	}

	return cfg
}

func LoadMinerConfig() MinerConfig {
	cfg := MinerConfig{}

	minerFlagSet := flag.NewFlagSet("miner", flag.ExitOnError)

	minerFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for parameter mining (required if -tl is not used)")
	minerFlagSet.StringVar(&cfg.TargetListFile, "tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	minerFlagSet.StringVar(&cfg.WordlistFile, "w", "wordlist.txt", "Path to the wordlist file")
	minerFlagSet.StringVar(&cfg.Method, "m", "GET", "HTTP method (GET, POST, JSON, XML)")
	minerFlagSet.StringVar(&cfg.Headers, "H", "", "Custom headers (e.g., 'Header1:Value1,Header2:Value2')")
	minerFlagSet.StringVar(&cfg.Data, "d", "", "Custom data (e.g., 'key1:value1,key2:value2')")
	minerFlagSet.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	minerFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	minerFlagSet.StringVar(&cfg.OutputFile, "o", "miner_results.json", "Path to the output file (JSON format)")
	minerFlagSet.StringVar(&cfg.InjectValue, "i", "test-value", "Injection value to use when fuzzing (default: test-value)")

	if err := minerFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	if cfg.TargetURL == "" && cfg.TargetListFile == "" {
		log.Fatal("Please provide a target URL using the -u flag or a target list file using the -tl flag")
	}

	if cfg.TargetURL != "" {
		validateURL(cfg.TargetURL)
	}

	cfg.Method = strings.ToUpper(cfg.Method)
	cfg.Headers = strings.TrimSpace(cfg.Headers)
	cfg.Data = strings.TrimSpace(cfg.Data)

	return cfg
}

func LoadValidatorConfig() ValidatorConfig {
	cfg := ValidatorConfig{}

	validatorFlagSet := flag.NewFlagSet("validator", flag.ExitOnError)

	validatorFlagSet.StringVar(&cfg.TargetURL, "u", "", "Target URL for validation (required if -tl is not used)")
	validatorFlagSet.StringVar(&cfg.TargetListFile, "tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	validatorFlagSet.IntVar(&cfg.Concurrency, "c", 40, "Number of concurrent requests")
	validatorFlagSet.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	validatorFlagSet.StringVar(&cfg.OutputFile, "o", "validator_results.json", "Path to the output file (JSON format)")

	if err := validatorFlagSet.Parse(os.Args[2:]); err != nil {
		log.Fatal("Failed to parse flags:", err)
	}

	if cfg.TargetURL == "" && cfg.TargetListFile == "" {
		log.Fatal("Please provide a target URL using the -u flag or a target list file using the -tl flag")
	}

	if cfg.TargetURL != "" {
		validateURL(cfg.TargetURL)
	}

	return cfg
}
func validateURL(url string) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		log.Fatal("Error: The URL must start with http:// or https://")
	}
}
func SaveResults(outputFile string, sitemapURLs []string, htmlLinks []string, jsLinks []string, validPaths []string, validParams []string) {
	results := struct {
		SitemapURLs []string `json:"sitemap_urls"`
		HTMLLinks   []string `json:"html_links"`
		JSLinks     []string `json:"js_links"`
		ValidPaths  []string `json:"valid_paths"`
		ValidParams []string `json:"valid_params"`
	}{
		SitemapURLs: sitemapURLs,
		HTMLLinks:   removeDuplicates(htmlLinks),
		JSLinks:     removeDuplicates(jsLinks),
		ValidPaths:  removeDuplicates(validPaths),
		ValidParams: removeDuplicates(validParams),
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		log.Fatalf("Failed to encode results to JSON: %v", err)
	}
}

func removeDuplicates(input []string) []string {
	unique := make(map[string]bool)
	var result []string
	for _, item := range input {
		if !unique[item] {
			unique[item] = true
			result = append(result, item)
		}
	}
	return result
}
