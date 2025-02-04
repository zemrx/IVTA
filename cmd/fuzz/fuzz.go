package fuzz

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ivta/config"
	"ivta/fuzzer"
	"ivta/utils"
)

func Execute() {
	targetURLFlag := flag.String("u", "", "Target URL to fuzz (required if -tl is not used)")
	targetListFlag := flag.String("tl", "", "Path to a file containing a list of target URLs (required if -u is not used)")
	dirWordlistFlag := flag.String("w", "wordlist.txt", "Path to the directory wordlist file (default: wordlist.txt)")
	maxDepthFlag := flag.Int("d", 2, "Maximum recursion depth for fuzzing (default: 2)")
	concurrencyFlag := flag.Int("c", 5, "Number of concurrent requests (default: 5)")
	outputFileFlag := flag.String("o", "fuzz_results.json", "Path to the output file (default: fuzz_results.json)")
	verboseFlag := flag.Bool("v", false, "Enable verbose mode")

	blacklistStatusFlag := flag.String("bs", "", "Comma-separated list of blacklisted status codes (e.g. 404,500)")
	blacklistLengthsFlag := flag.String("bl", "", "Comma-separated list of blacklisted body lengths (in bytes)")
	blacklistWordCountsFlag := flag.String("bw", "", "Comma-separated list of blacklisted word counts")
	blacklistLineCountsFlag := flag.String("blc", "", "Comma-separated list of blacklisted non-empty line counts")
	blacklistSearchWordsFlag := flag.String("bsw", "", "Comma-separated list of blacklisted search words")
	blacklistRegexFlag := flag.String("br", "", "Comma-separated list of blacklisted regex patterns")

	helpFlag := flag.Bool("h", false, "Display this help message")
	flag.Parse()

	if *helpFlag {
		Help()
		os.Exit(0)
	}

	if *targetURLFlag == "" && *targetListFlag == "" {
		fmt.Println("Error: You must provide either a target URL (-u) or a target list file (-tl).")
		Help()
		os.Exit(1)
	}

	cfg := config.LoadFuzzConfig()

	if *targetURLFlag != "" {
		cfg.TargetURL = *targetURLFlag
	}
	if *targetListFlag != "" {
		cfg.TargetListFile = *targetListFlag
	}
	cfg.DirWordlistFile = *dirWordlistFlag
	cfg.MaxDepth = *maxDepthFlag
	cfg.Concurrency = *concurrencyFlag
	cfg.OutputFile = *outputFileFlag
	cfg.Verbose = *verboseFlag

	var bs []int
	if *blacklistStatusFlag != "" {
		bs = utils.ParseIntSlice(*blacklistStatusFlag)
	}
	var bl []int
	if *blacklistLengthsFlag != "" {
		bl = utils.ParseIntSlice(*blacklistLengthsFlag)
	}
	var bw []int
	if *blacklistWordCountsFlag != "" {
		bw = utils.ParseIntSlice(*blacklistWordCountsFlag)
	}
	var blc []int
	if *blacklistLineCountsFlag != "" {
		blc = utils.ParseIntSlice(*blacklistLineCountsFlag)
	}
	var bsw []string
	if *blacklistSearchWordsFlag != "" {
		parts := strings.Split(*blacklistSearchWordsFlag, ",")
		for _, p := range parts {
			if word := strings.TrimSpace(p); word != "" {
				bsw = append(bsw, word)
			}
		}
	}
	var br []string
	if *blacklistRegexFlag != "" {
		parts := strings.Split(*blacklistRegexFlag, ",")
		for _, p := range parts {
			if pat := strings.TrimSpace(p); pat != "" {
				br = append(br, pat)
			}
		}
	}

	options := fuzzer.FuzzOptions{
		Depth:                cfg.MaxDepth,
		BlacklistStatusCodes: bs,
		BlacklistLengths:     bl,
		BlacklistWordCounts:  bw,
		BlacklistLineCounts:  blc,
		BlacklistSearchWords: bsw,
		BlacklistRegex:       br,
	}

	var targets []string
	if cfg.TargetListFile != "" {
		var err error
		targets, err = utils.ReadTargetList(cfg.TargetListFile)
		if err != nil {
			fmt.Println("Error reading target list:", err)
			os.Exit(1)
		}
	} else {
		targets = []string{cfg.TargetURL}
	}

	for _, target := range targets {
		fmt.Println("Processing target:", target)
		results := fuzzer.FuzzDirectories(target, cfg.DirWordlistFile, cfg.Concurrency, options)
		fmt.Printf("Found %d valid paths.\n", len(results))

		var resultURLs []string
		for _, res := range results {
			resultURLs = append(resultURLs, res.URL)
		}

		config.SaveResults(cfg.OutputFile, nil, nil, nil, resultURLs, nil)
		fmt.Println("Results saved to", cfg.OutputFile)
	}
}

func Help() {
	fmt.Println("Usage: ivta fuzz -u <target_url> [options]")
	fmt.Println("       ivta fuzz -tl <target_list_file> [options]")
	fmt.Println()
	fmt.Println("Fuzz a website for directories using blacklist filtering.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required if -tl is not used)")
	fmt.Println("  -tl      Path to a file containing target URLs (required if -u is not used)")
	fmt.Println("  -w       Path to the directory wordlist file (default: wordlist.txt)")
	fmt.Println("  -d       Maximum recursion depth for fuzzing (default: 2)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Output file (default: fuzz_results.json)")
	fmt.Println("  -bs      Comma-separated list of blacklisted status codes (e.g. 404,500)")
	fmt.Println("  -bl      Comma-separated list of blacklisted body lengths (in bytes)")
	fmt.Println("  -bw      Comma-separated list of blacklisted word counts")
	fmt.Println("  -blc     Comma-separated list of blacklisted non-empty line counts")
	fmt.Println("  -bsw     Comma-separated list of blacklisted search words")
	fmt.Println("  -br      Comma-separated list of blacklisted regex patterns")
	fmt.Println("  -h       Display this help message")
}
