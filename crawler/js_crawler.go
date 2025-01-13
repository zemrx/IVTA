package crawler

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func RunCrawlerWithJS(targetURL string) []string {
	var links []string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &links),
	)
	if err != nil {
		log.Fatalf("Failed to run browser: %v", err)
	}

	return links
}
