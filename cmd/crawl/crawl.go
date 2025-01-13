package crawl

import (
	"fmt"
	"os"

	"ivta/config"
	"ivta/crawler"
	"ivta/parser"
)

func Execute() {
	if len(os.Args) < 3 {
		Help()
		os.Exit(1)
	}

	cfg := config.LoadCrawlConfig()

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

	results := crawler.RunCrawlerWithJS(cfg.TargetURL)
	jsLinks := results["links"]
	fmt.Printf("Crawled %d links using JavaScript rendering.\n", len(jsLinks))

	crawler.SubmitForms(cfg.TargetURL)

	config.SaveResults(cfg.OutputFile, sitemapURLs, htmlLinks, jsLinks, nil, nil)
	fmt.Println("Results saved to", cfg.OutputFile)
}

func Help() {
	fmt.Println("Usage: .\\ivta.exe crawl -u <target_url> [options]")
	fmt.Println("Crawl a website and discover links.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required)")
	fmt.Println("  -d       Maximum depth for recursive discovery (default: 2)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Path to the output file (default: crawl_results.json)")
	fmt.Println("  -h, --help   Display this help message")
}
