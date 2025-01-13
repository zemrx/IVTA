package crawler

import (
	"net/url"
	"sync"

	"github.com/gocolly/colly/v2"
)

func RunCrawler(targetURL string, depth int, concurrency int) []string {
	var (
		discoveredLinks []string
		visited         = make(map[string]bool)
		mutex           = &sync.Mutex{}
	)

	c := colly.NewCollector(
		colly.AllowedDomains(extractDomain(targetURL)),
		colly.MaxDepth(depth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: concurrency,
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		mutex.Lock()
		if !visited[link] {
			visited[link] = true
			discoveredLinks = append(discoveredLinks, link)
		}
		mutex.Unlock()

		e.Request.Visit(link)
	})

	c.OnError(func(r *colly.Response, err error) {

	})

	err := c.Visit(targetURL)
	if err != nil {

	}

	c.Wait()

	return discoveredLinks
}

func extractDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {

	}
	return parsedURL.Hostname()
}
