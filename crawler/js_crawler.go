package crawler

import (
	"context"
	"net/url"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

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
	_ = chromedp.Run(chromedpCtx,
		chromedp.Navigate(currentURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('a')).map(a => a.href).filter(href => href !== null && href !== '');
        `, &fetchedLinks),
	)

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

func RunCrawlerWithJS(targetURL string, depth int, concurrency int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var (
		links   = make(chan string, 1000)
		visited = &sync.Map{}
		wg      sync.WaitGroup
		result  []string
	)

	wg.Add(1)
	go fetchLinks(ctx, targetURL, 0, depth, visited, links, &wg, targetURL)

	go func() {
		for link := range links {
			result = append(result, link)
		}
	}()

	wg.Wait()
	close(links)

	return result
}

func resolveURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	relative, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(relative).String(), nil
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
