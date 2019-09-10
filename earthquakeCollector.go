package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

func collectEarthquake() {
	url := "http://sismologia.cl/events/listados/2019/09/20190910.html"

	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		panic(err)
	}

	body := resp.Body
	defer body.Close()

	z := html.NewTokenizer(body)
	var content []string

	// While have not hit the </html> tag
	for z.Token().Data != "html" {
		var tt = z.Next()
		if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "td" {
				inner := z.Next()
				if inner == html.StartTagToken {
					t := z.Token()
					isAnchor := t.Data == "a"
					if isAnchor {
						z.Next()
						text := (string)(z.Text())
						t := strings.TrimSpace(text)
						content = append(content, t)
					}
				}
				if inner == html.TextToken {
					text := (string)(z.Text())
					t := strings.TrimSpace(text)
					content = append(content, t)
					if len(content) == 6 {
						break
					}
				}
			}
		}
	}
	// Print to check the slice's content
	fmt.Println(content)
}
