package collectors

import (
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

type EarthquakeCollector struct{}

type Earthquake struct {
	ID        string `json:"id"`
	UTC       string `json:"utc"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Depth     string `json:"depth"`
	Magnitude string `json:"magnitude"`
}

func (e EarthquakeCollector) sourceName() string {
	return "earthquake"
}

func (e EarthquakeCollector) collectEvent() (string, string) {
	prefixURL := "http://www.sismologia.cl"
	resp, err := http.Get(prefixURL + "/links/ultimos_sismos.html")
	// handle the error if there is one
	if err != nil {
		log.Error("failed to get earthquake event")
		return "", ""
	}
	body := resp.Body
	defer body.Close()

	// Get last earthquake URL
	var lastEarthquakeURL string
	z := html.NewTokenizer(body)
	for z.Token().Data != "html" {
		var tt = z.Next()
		if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						lastEarthquakeURL = a.Val
						break
					}
				}
				break
			}
		}
	}

	resp, err = http.Get(prefixURL + lastEarthquakeURL)
	// handle the error if there is one
	if err != nil {
		log.Error("failed to get earthquake event")
		return "", ""
	}
	body = resp.Body
	defer body.Close()
	var content []string

	// Get data from last earthquake
	z = html.NewTokenizer(body)
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
				}
			}
		}
	}

	lastEarthquakeID := getIDFromURL(lastEarthquakeURL)
	lastEarthquake := createEarthquakeObject(cleanData(content), lastEarthquakeID)
	lastEarthquakeMetadata := generateEarthquakeMetadata(lastEarthquake)
	lastEarthquakeAsJSONBytes, _ := json.Marshal(lastEarthquake)
	lastEarthquakeAsJSONString := string(lastEarthquakeAsJSONBytes)
	return lastEarthquakeAsJSONString, lastEarthquakeMetadata
}

func generateEarthquakeMetadata(eq Earthquake) string {
	digest := sha3.Sum512([]byte(EarthquakeCanonicalForm(eq)))
	return hex.EncodeToString(digest[:])
}

func getIDFromURL(url string) string {
	var a = strings.Split(url, "/")
	id := a[len(a)-1]
	return strings.Split(id, ".html")[0]
}

func cleanData(c []string) []string {
	var ret []string
	for i := 0; i < len(c); i++ {
		if i%2 != 0 {
			ret = append(ret, c[i])
		}
	}
	return ret
}

func createEarthquakeObject(data []string, id string) Earthquake {
	var lastEarthquake Earthquake
	lastEarthquake.ID = id
	lastEarthquake.UTC = data[1]
	lastEarthquake.Latitude = data[2]
	lastEarthquake.Longitude = data[3]
	lastEarthquake.Depth = cleanProperty(data[4])
	lastEarthquake.Magnitude = cleanProperty(data[5])
	return lastEarthquake
}

func cleanProperty(data string) string {
	return strings.Split(data, " ")[0]
}

func (e EarthquakeCollector) estimateEntropy() int {
	return 0
}

func (e EarthquakeCollector) getCanonicalForm(s string) string {
	if s == "" {
		return ""
	}
	var earthquake Earthquake
	err := json.Unmarshal([]byte(s), &earthquake)
	if err != nil {
		log.Error(err)
	}
	return EarthquakeCanonicalForm(earthquake)
}

func EarthquakeCanonicalForm(eq Earthquake) string {
	values := []string{eq.ID, eq.UTC, eq.Latitude, eq.Longitude, eq.Depth, eq.Magnitude}
	return strings.Join(values, ";")
}
