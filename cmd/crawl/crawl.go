package crawl

import (
	"flag"
	"fmt"
	"log"

	"ivta/config"
	"ivta/crawler"
	"ivta/parser"
)

func Execute() {
	cfg := config.CrawlConfig{}

	flag.StringVar(&cfg.TargetURL, "u", "", "Target URL for crawling (required)")
	flag.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	flag.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	flag.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	flag.StringVar(&cfg.OutputFile, "o", "crawl_results.json", "Path to the output file (JSON format)")

	flag.Parse()

	if cfg.TargetURL == "" {
		log.Fatal("Please provide a target URL using the -u flag")
	}

	sitemapURLs := parser.ParseSitemap(cfg.TargetURL + "/sitemap.xml")
	fmt.Printf("Parsed %d URLs from sitemap.\n", len(sitemapURLs))

	jsLinks := crawler.RunCrawlerWithJS(cfg.TargetURL)
	fmt.Printf("Crawled %d links using JavaScript rendering.\n", len(jsLinks))

	crawler.SubmitForms(cfg.TargetURL)

	config.SaveResults(cfg.OutputFile, sitemapURLs, jsLinks, nil, nil)
	fmt.Println("Results saved to", cfg.OutputFile)
}
