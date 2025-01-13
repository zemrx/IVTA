package parser

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
)

type Sitemap struct {
	URLs []URL `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

func ParseSitemap(sitemapURL string) []string {
	resp, err := http.Get(sitemapURL)
	if err != nil {
		log.Fatalf("Failed to fetch sitemap: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read sitemap: %v", err)
	}

	var sitemap Sitemap
	err = xml.Unmarshal(body, &sitemap)
	if err != nil {
		log.Fatalf("Failed to parse sitemap: %v", err)
	}

	var urls []string
	for _, url := range sitemap.URLs {
		urls = append(urls, url.Loc)
	}

	return urls
}
