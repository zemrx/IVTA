package crawler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func SubmitForms(targetURL string) {
	resp, err := http.Get(targetURL)
	if err != nil {
		log.Fatalf("Failed to fetch page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse page: %v", err)
	}

	doc.Find("form").Each(func(i int, form *goquery.Selection) {
		action, exists := form.Attr("action")
		if !exists || action == "" {
			action = targetURL
		} else {
			action = resolveFormAction(targetURL, action)
		}

		method, exists := form.Attr("method")
		if !exists || method == "" {
			method = "GET"
		}

		formData := url.Values{}
		form.Find("input").Each(func(i int, input *goquery.Selection) {
			name, exists := input.Attr("name")
			if !exists || name == "" {
				return
			}
			value, _ := input.Attr("value")
			formData.Set(name, value)
		})

		var resp *http.Response
		if strings.ToUpper(method) == "POST" {
			resp, err = http.PostForm(action, formData)
		} else {
			resp, err = http.Get(action + "?" + formData.Encode())
		}

		if err != nil {
			log.Printf("Failed to submit form (action: %s): %v", action, err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Form submitted to %s (Status: %d)\n", action, resp.StatusCode)
	})
}

func resolveFormAction(baseURL, action string) string {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing base URL: %v", err)
		return action
	}
	parsedAction, err := url.Parse(action)
	if err != nil {
		log.Printf("Error parsing action URL: %v", err)
		return action
	}
	return parsedBase.ResolveReference(parsedAction).String()
}
