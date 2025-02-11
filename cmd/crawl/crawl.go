package crawl

import (
	"fmt"
	"os"

	"ivta/config"
	"ivta/crawler"
	"ivta/parser"
	"ivta/utils"
)

func Execute() {
	cfg := config.LoadCrawlConfig()

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

		htmlLinks := crawler.RunCrawler(target, cfg.MaxDepth, cfg.Concurrency)
		fmt.Printf("Crawled %d links using HTML parsing.\n", len(htmlLinks))

		jsLinks := crawler.RunCrawlerWithJS(target, cfg.MaxDepth, cfg.Concurrency)
		fmt.Printf("Crawled %d links using JavaScript rendering.\n", len(jsLinks))

		crawler.SubmitForms(target)

		config.SaveResults(cfg.OutputFile, sitemapURLs, htmlLinks, jsLinks, nil, nil)
		fmt.Println("Results saved to", cfg.OutputFile)
	}
}

func Help() {
	fmt.Println("Usage: ivta crawl -u <target_url> [options]")
	fmt.Println("       ivta crawl -tl <target_list_file> [options]")
	fmt.Println()
	fmt.Println("Crawl a website and discover links, sitemaps, and forms.")
	fmt.Println("Options:")
	fmt.Println("  -u       Target URL (required if -tl is not used)")
	fmt.Println("  -tl      Path to a file containing a list of target URLs (required if -u is not used)")
	fmt.Println("  -d       Maximum depth for recursive discovery (default: 2)")
	fmt.Println("  -c       Number of concurrent requests (default: 5)")
	fmt.Println("  -v       Enable verbose mode")
	fmt.Println("  -o       Path to the output file (default: crawl_results.json)")
	fmt.Println("  -h       Display this help message")
}
