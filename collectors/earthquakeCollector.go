package collectors

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type EarthquakeCollector struct{}

func (e EarthquakeCollector) collectEvent() string {
	now := time.Now().UTC()

	currentYear := strconv.Itoa(now.Year())
	currentMonth := strings.Split(now.Format("2006#01#02"), "#")[1]
	baseURL := "http://sismologia.cl/events/listados/"
	dateFormat := now.Format("20060102")
	url := baseURL + currentYear + "/" + currentMonth + "/" + dateFormat + ".html"

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
	if content != nil {
		content = append(content[:1], content[2:]...)
		content[4] = cleanMagnitude(content)
		return fmt.Sprint(content)
	} else {
		log.Error("Failed to get Earthquake event")
		return "0"
	}

}

func cleanMagnitude(data []string) string {
	magnitude := data[len(data)-1]
	return strings.Split(magnitude, " ")[0]
}

func (e EarthquakeCollector) estimateEntropy() int {
	return 0
}

func (e EarthquakeCollector) sourceID() int {
	return 0
}
