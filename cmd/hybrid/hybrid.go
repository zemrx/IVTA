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
	htmlLinks := crawler.RunCrawler(cfg.TargetURL, cfg.MaxDepth, cfg.Concurrency)
	fmt.Printf("Crawled %d links using HTML parsing.\n", len(htmlLinks))

	jsLinks := crawler.RunCrawlerWithJS(cfg.TargetURL)
	fmt.Printf("Crawled %d links using JavaScript rendering.\n", len(jsLinks))

	crawler.SubmitForms(cfg.TargetURL)

	validPaths := fuzzer.FuzzDirectories(cfg.TargetURL, cfg.DirWordlistFile, cfg.Concurrency, 0)
	fmt.Printf("Found %d valid paths.\n", len(validPaths))

	validParams := fuzzer.FuzzParameters(cfg.TargetURL, cfg.ParamWordlistFile, cfg.Concurrency, 0, cfg.CustomSymbol)
	fmt.Printf("Found %d valid parameters.\n", len(validParams))

	config.SaveResults(cfg.OutputFile, sitemapURLs, htmlLinks, jsLinks, validPaths, validParams)
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
