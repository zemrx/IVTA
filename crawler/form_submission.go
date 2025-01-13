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
		action, _ := form.Attr("action")
		method, _ := form.Attr("method")
		if method == "" {
			method = "GET"
		}

		formData := url.Values{}
		form.Find("input").Each(func(i int, input *goquery.Selection) {
			name, _ := input.Attr("name")
			value, _ := input.Attr("value")
			if name != "" {
				formData.Set(name, value)
			}
		})

		submitURL := targetURL
		if action != "" {
			submitURL = action
		}

		var resp *http.Response
		if strings.ToUpper(method) == "POST" {
			resp, err = http.PostForm(submitURL, formData)
		} else {
			resp, err = http.Get(submitURL + "?" + formData.Encode())
		}

		if err != nil {
			log.Printf("Failed to submit form: %v", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Form submitted to %s (Status: %d)\n", submitURL, resp.StatusCode)
	})
}
