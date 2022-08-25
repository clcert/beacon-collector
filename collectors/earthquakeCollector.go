package collectors

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	goquery "github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
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

func (e EarthquakeCollector) collectEvent(ch chan Event) {
	prefixURL := "http://www.sismologia.cl"
	resp, err := http.Get(prefixURL + "/index.html")
	// handle the error if there is one
	if err != nil {
		log.Error("failed to get earthquake event")
		// return "", "", 2
		ch <- Event{"", "", 2}
		return
	}
	if resp.StatusCode != 200 {
		log.Error("earthquake error response code: " + strconv.Itoa(resp.StatusCode))
		// return "", "", 2
		ch <- Event{"", "", 2}
		return
	}
	body := resp.Body
	defer body.Close()

	log.Debug("Looking for earthquake events...")

	docIndex, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	// Get last earthquakes URLs
	var lastEarthquakesURL []string
	docIndex.Find(".sismologia").Each(
		func(i int, s *goquery.Selection) {
			s.Find("a").Each(
				func(i int, s *goquery.Selection) {
					url, _ := s.Attr("href")
					lastEarthquakesURL = append(lastEarthquakesURL, url)
				},
			)
		},
	)

	var lastEarthquakeURL string
	var status = 0
	for _, v := range lastEarthquakesURL {
		resp, err = http.Get(prefixURL + v)
		if err != nil {
			log.Error("failed to get earthquake event")
			// return "", "", 2
			ch <- Event{"", "", 2}
			return
		}
		if resp.StatusCode != 200 {
			log.Error("earthquake error response code: " + strconv.Itoa(resp.StatusCode))
			status = 8
		} else {
			lastEarthquakeURL = v
			break
		}
	}
	body = resp.Body
	defer body.Close()
	var content []string

	docEarthquake, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	ommit := true
	docEarthquake.Find(".sismologia.informe").Each(
		func(i int, s *goquery.Selection) {
			s.Find("td").Each(
				func(i int, s *goquery.Selection) {
					text := strings.TrimSpace(s.Text())
					if !ommit {
						content = append(content, text)
					}
					ommit = !ommit
				},
			)
		},
	)

	lastEarthquakeID := getIDFromURL(lastEarthquakeURL)
	log.Debug("Last earthquake ID: " + lastEarthquakeID)
	lastEarthquake := createEarthquakeObject(content, lastEarthquakeID)
	log.Debug("Last earthquake: " + EarthquakeCanonicalForm(lastEarthquake))
	lastEarthquakeMetadata := generateEarthquakeMetadata(lastEarthquake)
	lastEarthquakeAsJSONBytes, _ := json.Marshal(lastEarthquake)
	lastEarthquakeAsJSONString := string(lastEarthquakeAsJSONBytes)
	// return lastEarthquakeAsJSONString, lastEarthquakeMetadata, status
	log.Debug("Done EarthquakeCollector")
	ch <- Event{lastEarthquakeAsJSONString, lastEarthquakeMetadata, status}
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

func createEarthquakeObject(data []string, id string) Earthquake {
	var lastEarthquake Earthquake
	lastEarthquake.ID = id
	lastEarthquake.UTC = data[2]
	lastEarthquake.Latitude = data[3]
	lastEarthquake.Longitude = data[4]
	lastEarthquake.Depth = cleanProperty(data[5])
	lastEarthquake.Magnitude = cleanProperty(data[6])
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
