package crawler

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func RunCrawler(targetURL string, depth int, concurrency int) []string {
	var discoveredLinks []string

	resp, err := http.Get(targetURL)
	if err != nil {
		log.Fatalf("Failed to fetch page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse page: %v", err)
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			discoveredLinks = append(discoveredLinks, href)
		}
	})

	return discoveredLinks
}
