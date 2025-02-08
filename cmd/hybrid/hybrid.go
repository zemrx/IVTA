package hybrid

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ivta/config"
	"ivta/crawler"
	"ivta/fuzzer"
	"ivta/miner"
	"ivta/parser"
	"ivta/utils"
)

func Execute() {
	targetURLFlag := flag.String("u", "", "Target URL for hybrid discovery (required if -tl is not used)")
	targetListFlag := flag.String("tl", "", "Path to a file containing target URLs (required if -u is not used)")
	dirWordlistFlag := flag.String("w", "wordlist.txt", "Path to the directory wordlist file (default: wordlist.txt)")
	paramWordlistFlag := flag.String("p", "param_wordlist.txt", "Path to the parameter wordlist file (default: param_wordlist.txt)")
	maxDepthFlag := flag.Int("d", 2, "Maximum depth for recursive discovery (default: 2)")
	concurrencyFlag := flag.Int("c", 5, "Number of concurrent requests (default: 5)")
	verboseFlag := flag.Bool("v", false, "Enable verbose mode")
	outputFileFlag := flag.String("o", "hybrid_results.json", "Path to the output file (default: hybrid_results.json)")
	customSymbolFlag := flag.String("s", "test", "Custom symbol to test for reflection (default: test)")
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

	cfg := config.LoadHybridConfig()

	if *targetURLFlag != "" {
		cfg.TargetURL = *targetURLFlag
	}
	if *targetListFlag != "" {
		cfg.TargetListFile = *targetListFlag
	}
	cfg.DirWordlistFile = *dirWordlistFlag
	cfg.ParamWordlistFile = *paramWordlistFlag
	cfg.MaxDepth = *maxDepthFlag
	cfg.Concurrency = *concurrencyFlag
	cfg.OutputFile = *outputFileFlag
	cfg.CustomSymbol = *customSymbolFlag
	cfg.Verbose = *verboseFlag

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

		sitemapURL := target + "/sitemap.xml"
		sitemapURLs := parser.ParseSitemap(sitemapURL)
		if sitemapURLs == nil || len(sitemapURLs) == 0 {
			fmt.Println("No sitemap found or failed to parse sitemap.")
		} else {
			fmt.Printf("Parsed %d URLs from sitemap.\n", len(sitemapURLs))
		}

		fmt.Println("Starting initial crawling...")
		htmlLinks := crawler.RunCrawler(target, cfg.MaxDepth, cfg.Concurrency)
		fmt.Printf("Crawled %d links using HTML parsing.\n", len(htmlLinks))

		jsLinks := crawler.RunCrawlerWithJS(target, cfg.MaxDepth, cfg.Concurrency)
		fmt.Printf("Crawled %d links using JavaScript rendering.\n", len(jsLinks))

		allLinks := append(htmlLinks, jsLinks...)

		fmt.Println("Starting fuzzing on discovered links...")
		var validPaths []string
		for _, link := range allLinks {
			paths := fuzzer.FuzzDirectories(link, cfg.DirWordlistFile, cfg.Concurrency, fuzzer.FuzzOptions{
				Depth: cfg.MaxDepth,
			})
			for _, res := range paths {
				validPaths = append(validPaths, res.URL)
			}
		}
		fmt.Printf("Found %d valid directory paths.\n", len(validPaths))

		fmt.Println("Starting parameter fuzzing...")
		paramWordlist, err := utils.ReadWordlist(cfg.ParamWordlistFile)
		if err != nil {
			fmt.Printf("Error reading parameter wordlist: %v\n", err)
			continue
		}

		headers := make(map[string]string)
		if cfg.Headers != "" {
			for _, pair := range strings.Split(cfg.Headers, ",") {
				parts := strings.Split(pair, ":")
				if len(parts) == 2 {
					headers[parts[0]] = parts[1]
				}
			}
		}
		data := make(map[string]string)
		if cfg.Data != "" {
			for _, pair := range strings.Split(cfg.Data, ",") {
				parts := strings.Split(pair, ":")
				if len(parts) == 2 {
					data[parts[0]] = parts[1]
				}
			}
		}
		options := miner.RequestOptions{
			Method:  cfg.Method,
			Headers: headers,
			Data:    data,
		}

		validParams, err := miner.BruteForce(target, paramWordlist, options, cfg.Concurrency)
		if err != nil {
			fmt.Printf("Error during parameter fuzzing: %v\n", err)
			continue
		}
		fmt.Printf("Found %d valid parameters.\n", len(validParams))
		for _, param := range validParams {
			fmt.Println(param)
		}

		config.SaveResults(cfg.OutputFile, sitemapURLs, allLinks, jsLinks, validPaths, validParams)
		fmt.Println("Results saved to", cfg.OutputFile)
	}
}

func Help() {
	fmt.Println("Usage: ivta.exe hybrid -u <target_url> [options]")
	fmt.Println("       ivta.exe hybrid -tl <target_list_file> [options]")
	fmt.Println()
	fmt.Println("Run a hybrid discovery (crawling + fuzzing) on a website.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required if -tl is not used)")
	fmt.Println("  -tl      Path to a file containing target URLs (required if -u is not used)")
	fmt.Println("  -w       Path to the directory wordlist file (default: wordlist.txt)")
	fmt.Println("  -p       Path to the parameter wordlist file (default: param_wordlist.txt)")
	fmt.Println("  -d       Maximum depth for recursive discovery (default: 2)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Output file (default: hybrid_results.json)")
	fmt.Println("  -s       Custom symbol to test for reflection (default: test)")
	fmt.Println("  -h       Display this help message")
}
