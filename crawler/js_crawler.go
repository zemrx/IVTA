package crawler

import (
	"context"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

func RunCrawlerWithJS(targetURL string, depth int, concurrency int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var (
		links   = make(chan string, 1000)
		visited = &sync.Map{}
		wg      sync.WaitGroup
	)

	wg.Add(1)
	go fetchLinks(ctx, targetURL, 0, depth, visited, links, &wg, targetURL)

	wg.Wait()
	close(links)

	var results []string
	for link := range links {
		results = append(results, link)
	}
	return results
}

func fetchLinks(ctx context.Context, currentURL string, currentDepth int, depth int, visited *sync.Map, links chan<- string, wg *sync.WaitGroup, targetURL string) {
	defer wg.Done()

	if currentDepth > depth {
		return
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	chromedpCtx, cancelChromedp := chromedp.NewContext(fetchCtx)
	defer cancelChromedp()

	var fetchedLinks []string
	if err := chromedp.Run(chromedpCtx,
		chromedp.Navigate(currentURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href).filter(href => href);`, &fetchedLinks),
	); err != nil {
		log.Printf("chromedp error at %s: %v", currentURL, err)
		return
	}

	for _, link := range fetchedLinks {
		resolvedLink, err := resolveURL(currentURL, link)
		if err != nil {
			continue
		}

		if isSameDomain(targetURL, resolvedLink) {
			if _, loaded := visited.LoadOrStore(resolvedLink, true); !loaded {
				links <- resolvedLink

				wg.Add(1)
				go fetchLinks(ctx, resolvedLink, currentDepth+1, depth, visited, links, wg, targetURL)
			}
		}
	}
}

func resolveURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	rel, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(rel).String(), nil
}

func isSameDomain(baseURL, targetURL string) bool {
	base, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	return base.Hostname() == target.Hostname()
}
