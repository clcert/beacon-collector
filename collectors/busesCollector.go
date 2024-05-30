package collectors

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const timeLayout = "02-01-2006 15:04:05"

type BusesCollector struct{}

type BusesLocation struct {
	BusesLocation []BusLocation `json:"busesLocations"`
}

type BusLocation struct {
	DateTime     time.Time `json:"dateTime"`
	LicensePlate string    `json:"licensePlate"`
	Lat          float64   `json:"lat"`
	Lon          float64   `json:"lon"`
	Speed        float64   `json:"speed"`
	Route        string    `json:"route"`
	Direction    string    `json:"direction"`
}

type BusesResponse struct {
	Date         string   `json:"fecha_consulta"`
	RawPositions []string `json:"posiciones"`
}

func (e BusesCollector) sourceName() string {
	return "buses"
}

func (e BusesCollector) collectEvent(ch chan Event) {
	busLocationURL := "http://www.dtpmetropolitano.cl/posiciones"
	username, password := getBusesCredentials()

	// Create a new request
	req, err := http.NewRequest("GET", busLocationURL, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Add the Basic Auth header
	auth := username + ":" + password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", basicAuth)

	// Create a client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		ch <- Event{"", "", FLES_SourceFail}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		ch <- Event{"", "", FLES_SourceFail}
		return
	}

	// Parse the response
	var responseContent BusesResponse
	err = json.Unmarshal(body, &responseContent)
	if err != nil {
		log.Fatal(err)
		ch <- Event{"", "", FLES_SourceFail}
		return
	}

	log.Info("JSON Unmarshalled")
	log.Info("Positions Found: " + strconv.Itoa(len(responseContent.RawPositions)))

	// Parse the buses locations
	var buses BusesLocation
	for _, rawBusLocation := range responseContent.RawPositions {
		var busLocation BusLocation
		rawData := strings.Split(rawBusLocation, ";")

		busLocation.DateTime, err = time.Parse(timeLayout, rawData[0])
		if err != nil {
			log.Fatal(err)
			continue
		}
		busLocation.LicensePlate = rawData[1]
		busLocation.Lat, _ = strconv.ParseFloat(rawData[2], 64)
		busLocation.Lon, _ = strconv.ParseFloat(rawData[3], 64)
		busLocation.Speed, _ = strconv.ParseFloat(rawData[4], 64)
		busLocation.Route = rawData[7]
		busLocation.Direction = rawData[8]

		// Discard old records
		oneMinuteAgo := time.Now().UTC().Add(-5 * time.Minute)
		if busLocation.DateTime.Before(oneMinuteAgo) {
			continue
		}

		log.Info("bus location: " + busLocation.CanonicalForm())
		buses.BusesLocation = append(buses.BusesLocation, busLocation)
	}

	log.Info("Positions Parsed: " + strconv.Itoa(len(buses.BusesLocation)))
}

func getBusesCredentials() (string, string) {
	configJsonFile, err := os.Open("collectors/dtpmConfig.json")
	if err != nil {
		log.Error(err)
		return "", ""
	}
	defer configJsonFile.Close()
	busesCredentials := make(map[string]string)
	byteValue, _ := io.ReadAll(configJsonFile)
	json.Unmarshal(byteValue, &busesCredentials)

	return busesCredentials["username"], busesCredentials["password"]
}

func (e BusesCollector) estimateEntropy() int {
	return 0
}

func (e BusesCollector) getCanonicalForm(data string) string {
	if data == "" {
		return ""
	}
	return ""
}

func (buses BusesLocation) BusesCanonicalForm() string {
	var busesCanonLst []string
	for _, busLocation := range buses.BusesLocation {
		busesCanonLst = append(busesCanonLst, busLocation.CanonicalForm())
	}
	busesCanon := strings.Join(busesCanonLst, "\n")
	return busesCanon
}

func (b BusLocation) CanonicalForm() string {
	latStr := strconv.FormatFloat(b.Lat, 'f', -1, 64)
	lonStr := strconv.FormatFloat(b.Lon, 'f', -1, 64)
	speedStr := strconv.FormatFloat(b.Speed, 'f', -1, 64)
	values := []string{b.DateTime.Local().Format(timeLayout), b.LicensePlate, latStr, lonStr, speedStr, b.Route, b.Direction}
	return strings.Join(values, ";")
}
