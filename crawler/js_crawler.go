package crawler

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
)

func RunCrawlerWithJS(targetURL string) map[string][]string {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var (
		links          []string
		textContent    []string
		imageSources   []string
		metadata       map[string]string
		dynamicContent []string
	)

	metadata = make(map[string]string)

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),

		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('a')).map(a => {
                const href = a.getAttribute('href');
                if (href) {
                    try {
                        const url = new URL(href, window.location.origin);
                        return url.href;
                    } catch (e) {
                        return null;
                    }
                }
                return null;
            }).filter(link => link !== null);
        `, &links),

		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('div.content, span.text, p')).map(el => el.textContent.trim());
        `, &textContent),

		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('img')).map(img => img.src);
        `, &imageSources),

		chromedp.Evaluate(`
            const metadata = {};
            metadata.title = document.title;
            metadata.description = document.querySelector('meta[name="description"]')?.content || '';
            metadata.ogTitle = document.querySelector('meta[property="og:title"]')?.content || '';
            metadata.ogDescription = document.querySelector('meta[property="og:description"]')?.content || '';
            metadata.ogImage = document.querySelector('meta[property="og:image"]')?.content || '';
            metadata;
        `, &metadata),

		chromedp.WaitVisible(".dynamic-content"),
		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('.dynamic-content')).map(el => el.textContent.trim());
        `, &dynamicContent),
	)
	if err != nil {
		log.Printf("Failed to run browser: %v", err)
		return map[string][]string{
			"links":          []string{},
			"textContent":    []string{},
			"imageSources":   []string{},
			"dynamicContent": []string{},
			"metadata":       []string{},
		}
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

	links = removeDuplicates(links)
	textContent = removeDuplicates(textContent)
	imageSources = removeDuplicates(imageSources)
	dynamicContent = removeDuplicates(dynamicContent)

	metadataSlice := mapToSlice(metadata)

	result := map[string][]string{
		"links":          links,
		"textContent":    textContent,
		"imageSources":   imageSources,
		"dynamicContent": dynamicContent,
		"metadata":       metadataSlice,
	}

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

func removeDuplicates(slice []string) []string {
	unique := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !unique[item] {
			unique[item] = true
			result = append(result, item)
		}
	}

	return result
}

func mapToSlice(m map[string]string) []string {
	var result []string
	for key, value := range m {
		result = append(result, key+": "+value)
	}
	return result
}
