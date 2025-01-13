package config

import (
	"encoding/json"
	"log"
	"os"
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

func SaveResults(outputFile string, sitemapURLs []string, jsLinks []string, validPaths []string, validParams []string) {
	results := struct {
		SitemapURLs []string `json:"sitemap_urls"`
		JSLinks     []string `json:"js_links"`
		ValidPaths  []string `json:"valid_paths"`
		ValidParams []string `json:"valid_params"`
	}{
		SitemapURLs: sitemapURLs,
		JSLinks:     jsLinks,
		ValidPaths:  validPaths,
		ValidParams: validParams,
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
