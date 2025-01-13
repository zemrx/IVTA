package hybrid

import (
	"flag"
	"fmt"
	"log"

	"ivta/config"
	"ivta/crawler"
	"ivta/fuzzer"
	"ivta/parser"
)

func Execute() {
	cfg := config.HybridConfig{}

	flag.StringVar(&cfg.TargetURL, "u", "", "Target URL for hybrid discovery (required)")
	flag.StringVar(&cfg.DirWordlistFile, "w", "wordlist.txt", "Path to the directory wordlist file")
	flag.StringVar(&cfg.ParamWordlistFile, "p", "param_wordlist.txt", "Path to the parameter wordlist file")
	flag.IntVar(&cfg.MaxDepth, "d", 2, "Maximum depth for recursive discovery")
	flag.IntVar(&cfg.Concurrency, "c", 5, "Number of concurrent requests")
	flag.BoolVar(&cfg.Verbose, "v", false, "Enable verbose mode")
	flag.StringVar(&cfg.OutputFile, "o", "hybrid_results.json", "Path to the output file (JSON format)")
	flag.StringVar(&cfg.CustomSymbol, "s", "test", "Custom symbol to test for reflection")

	flag.Parse()

	if cfg.TargetURL == "" {
		log.Fatal("Please provide a target URL using the -u flag")
	}

	sitemapURLs := parser.ParseSitemap(cfg.TargetURL + "/sitemap.xml")
	fmt.Printf("Parsed %d URLs from sitemap.\n", len(sitemapURLs))

	jsLinks := crawler.RunCrawlerWithJS(cfg.TargetURL)
	fmt.Printf("Crawled %d links using JavaScript rendering.\n", len(jsLinks))

	crawler.SubmitForms(cfg.TargetURL)

	validPaths := fuzzer.FuzzDirectories(cfg.TargetURL, cfg.DirWordlistFile, cfg.Concurrency, 0)
	fmt.Printf("Found %d valid paths.\n", len(validPaths))

	validParams := fuzzer.FuzzParameters(cfg.TargetURL, cfg.ParamWordlistFile, cfg.Concurrency, 0, cfg.CustomSymbol)
	fmt.Printf("Found %d valid parameters.\n", len(validParams))

	config.SaveResults(cfg.OutputFile, sitemapURLs, jsLinks, validPaths, validParams)
	fmt.Println("Results saved to", cfg.OutputFile)
}
