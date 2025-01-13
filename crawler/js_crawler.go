package crawler

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
)

func RunCrawlerWithJS(targetURL string) []string {
	var links []string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('a')).map(a => {
                const href = a.getAttribute('href');
                if (href) {
                    if (href.startsWith('/') || href.startsWith('./') || href.startsWith('../')) {
                        return href; 
                    } else if (href.startsWith('http://') || href.startsWith('https://')) {
                        return href; 
                    }
                }
                return null; 
            }).filter(link => link !== null); 
        `, &links),
	)
	if err != nil {
		log.Fatalf("Failed to run browser: %v", err)
	}

	for i, link := range links {
		if strings.HasPrefix(link, "/") || strings.HasPrefix(link, "./") || strings.HasPrefix(link, "../") {
			resolvedURL, err := resolveURL(targetURL, link)
			if err != nil {
				log.Printf("Failed to resolve URL %s: %v", link, err)
				continue
			}
			links[i] = resolvedURL
		}
	}

	return links
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
