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
		log.Printf("Failed to fetch sitemap: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch sitemap: Status code %d\n", resp.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read sitemap: %v\n", err)
		return nil
	}

	var sitemap Sitemap
	err = xml.Unmarshal(body, &sitemap)
	if err != nil {
		log.Printf("Failed to parse sitemap: %v\n", err)
		return nil
	}

	var urls []string
	for _, url := range sitemap.URLs {
		urls = append(urls, url.Loc)
	}

	return urls
}
