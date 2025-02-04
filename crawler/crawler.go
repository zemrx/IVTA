package crawler

import (
	"log"
	"net/url"
	"sync"

	"github.com/gocolly/colly/v2"
)

// RunCrawler crawls the target URL (using Colly) to extract links up to the given depth
// and with the given concurrency. It returns a slice of discovered links.
func RunCrawler(targetURL string, depth int, concurrency int) []string {
	var (
		discoveredLinks []string
		visited         = make(map[string]bool)
		mutex           = &sync.Mutex{}
	)

	// Create a new collector with allowed domains and asynchronous processing.
	c := colly.NewCollector(
		colly.AllowedDomains(extractDomain(targetURL)),
		colly.MaxDepth(depth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: concurrency,
	})

	// OnHTML callback to extract and store links.
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}
		mutex.Lock()
		if !visited[link] {
			visited[link] = true
			discoveredLinks = append(discoveredLinks, link)
		}
		mutex.Unlock()
		// Visit link in case it hasn’t been visited.
		_ = e.Request.Visit(link)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Colly error: %v (URL: %s)", err, r.Request.URL)
	})

	if err := c.Visit(targetURL); err != nil {
		log.Printf("Error visiting %s: %v", targetURL, err)
	}

	c.Wait()
	return discoveredLinks
}

// extractDomain extracts the hostname from a URL string.
func extractDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Failed to parse URL '%s': %v", rawURL, err)
		return ""
	}
	return parsedURL.Hostname()
}
