package hybrid

import (
	"fmt"
	"os"

	"ivta/config"
	"ivta/crawler"
	"ivta/fuzzer"
	"ivta/parser"
)

func Execute() {
	if len(os.Args) < 3 {
		Help()
		os.Exit(1)
	}

	cfg := config.LoadHybridConfig()

	fmt.Println("Target URL:", cfg.TargetURL)

	sitemapURL := cfg.TargetURL + "/sitemap.xml"
	sitemapURLs := parser.ParseSitemap(sitemapURL)
	if sitemapURLs == nil {
		fmt.Println("No sitemap found or failed to parse sitemap.")
	} else {
		fmt.Printf("Parsed %d URLs from sitemap.\n", len(sitemapURLs))
	}

	var allLinks []string
	var validPaths []string
	var validParams []string

	hybridDepth := cfg.MaxDepth

	fmt.Println("Starting initial crawling...")
	htmlLinks := crawler.RunCrawler(cfg.TargetURL, hybridDepth, cfg.Concurrency)
	fmt.Printf("Crawled %d links using HTML parsing.\n", len(htmlLinks))

	results := crawler.RunCrawlerWithJS(cfg.TargetURL)
	jsLinks := results["links"]

	allLinks = append(allLinks, htmlLinks...)
	allLinks = append(allLinks, jsLinks...)

	fmt.Println("Starting fuzzing on discovered links...")
	for _, link := range allLinks {
		paths := fuzzer.FuzzDirectories(link, cfg.DirWordlistFile, cfg.Concurrency)
		validPaths = append(validPaths, paths...)
	}
	fmt.Printf("Found %d valid paths.\n", len(validPaths))

	for depth := 1; depth <= hybridDepth; depth++ {
		fmt.Printf("Starting hybrid crawling and fuzzing at depth %d...\n", depth)

		var newLinks []string
		for _, path := range validPaths {
			links := crawler.RunCrawler(path, 1, cfg.Concurrency)
			newLinks = append(newLinks, links...)
		}
		fmt.Printf("Crawled %d new links at depth %d.\n", len(newLinks), depth)

		var newValidPaths []string
		for _, link := range newLinks {
			paths := fuzzer.FuzzDirectories(link, cfg.DirWordlistFile, cfg.Concurrency)
			newValidPaths = append(newValidPaths, paths...)
		}
		fmt.Printf("Found %d new valid paths at depth %d.\n", len(newValidPaths), depth)

		validPaths = append(validPaths, newValidPaths...)
	}

	fmt.Println("Starting parameter fuzzing...")
	for _, path := range validPaths {
		params := fuzzer.FuzzParameters(path, cfg.ParamWordlistFile, cfg.Concurrency, cfg.CustomSymbol)
		validParams = append(validParams, params...)
	}
	fmt.Printf("Found %d valid parameters.\n", len(validParams))

	config.SaveResults(cfg.OutputFile, sitemapURLs, allLinks, jsLinks, validPaths, validParams)
	fmt.Println("Results saved to", cfg.OutputFile)
}

func Help() {
	fmt.Println("Usage: .\\ivta.exe hybrid -u <target_url> [options]")
	fmt.Println("Run a hybrid discovery (crawling + fuzzing) on a website.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required)")
	fmt.Println("  -w       Path to the directory wordlist file (default: wordlist.txt)")
	fmt.Println("  -p       Path to the parameter wordlist file (default: param_wordlist.txt)")
	fmt.Println("  -d       Maximum depth for recursive discovery (default: 2)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Path to the output file (default: hybrid_results.json)")
	fmt.Println("  -s       Custom symbol to test for reflection (default: test)")
	fmt.Println("  -h, --help   Display this help message")
}
