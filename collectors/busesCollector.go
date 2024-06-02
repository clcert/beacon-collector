package collectors

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

const (
	timeLayout           = "02-01-2006 15:04:05"
	canonicalTimeLayout  = "15:04:05 02/01/2006"
	directionFieldLength = 4
)

type BusesCollector struct{}

type BusesLocation struct {
	BusesLocation []BusLocation `json:"busesLocations"`
}

type BusLocation struct {
	LicensePlate string  `json:"licensePlate"`
	Utc          string  `json:"utc"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Speed        float64 `json:"speed"`
	Route        string  `json:"route"`
	Direction    string  `json:"direction"`
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
		log.Errorf("error getting buses locations: %s", err)
		ch <- Event{"", "", FLES_SourceFail}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading buses locations response: %s", err)
		ch <- Event{"", "", FLES_SourceFail}
		return
	}

	// Parse the response
	var responseContent BusesResponse
	err = json.Unmarshal(body, &responseContent)
	if err != nil {
		log.Errorf("error unmarshalling buses locations json: %s", err)
		ch <- Event{"", "", FLES_SourceFail}
		return
	}

	log.Info("bus locations json unmarshalled")
	log.Infof("buses found: %d", len(responseContent.RawPositions))

	servicesParser := getServicesParser()
	if servicesParser == nil {
		log.Error("error loading services parser")
		ch <- Event{"", "", FLES_SourceFail}
		return
	}

	// Parse the buses locations
	var buses BusesLocation
	for _, rawBusLocation := range responseContent.RawPositions {
		var busLocation BusLocation
		rawData := strings.Split(rawBusLocation, ";")

		locationUTC, err := time.Parse(timeLayout, rawData[0])
		if err != nil {
			log.Error(err)
			continue
		}
		busLocation.LicensePlate = rawData[1]
		busLocation.Lat, _ = strconv.ParseFloat(rawData[2], 64)
		busLocation.Lon, _ = strconv.ParseFloat(rawData[3], 64)
		busLocation.Speed, _ = strconv.ParseFloat(rawData[4], 64)
		busLocation.Route = rawData[9]
		busLocation.Direction = rawData[8]

		// Discard old records
		twoMinutesAgo := time.Now().UTC().Add(-2 * time.Minute)
		if locationUTC.Before(twoMinutesAgo) {
			continue
		}
		busLocation.Utc = locationUTC.Format(canonicalTimeLayout)

		// Assign user route code if possible
		// (it may differ from internal route code)
		rawRoute := busLocation.Route
		if len(rawRoute) < 4 { // Invalid route code
			busLocation.Route = "Unknown"
		} else {
			undirectedRouteCode := rawRoute[:len(rawRoute)-4]
			// In transit to service buses have a different route code
			if undirectedRouteCode[len(undirectedRouteCode)-2:] == "TS" {
				busLocation.Route = "InTransit"
			} else {
				busLocation.Route = servicesParser[undirectedRouteCode]
			}
			// Not recognized route code
			if busLocation.Route == "" {
				busLocation.Route = "Unknown"
			}
		}
		buses.BusesLocation = append(buses.BusesLocation, busLocation)
	}
	log.Infof("buses successfully parsed: %d", len(buses.BusesLocation))

	busesRaw, _ := json.Marshal(buses)
	busesRawStr := string(busesRaw)
	busesMetadata := generateBusesMetadata(busesRawStr)
	ch <- Event{busesRawStr, busesMetadata, 0}
}

func generateBusesMetadata(busesRaw string) string {
	digest := sha3.Sum512([]byte(busesRaw))
	return hex.EncodeToString(digest[:])
}

/*
  - Loads the services.csv file and returns a dictionary that allows to
    map the route code to the user route code.
*/
func getServicesParser() map[string]string {
	serviceFile, err := os.Open("collectors/services.csv")
	if err != nil {
		log.Error(err)
		return nil
	}
	defer serviceFile.Close()
	servicesParser := make(map[string]string)
	byteValue, _ := io.ReadAll(serviceFile)
	servicesData := strings.Split(string(byteValue), "\n")
	for _, serviceData := range servicesData {
		service := strings.Split(serviceData, ";")
		if len(service) < 3 {
			continue
		}
		userRouteCode := service[0]
		serviceName := service[2]
		undirectedServiceName := serviceName[:len(serviceName)-directionFieldLength]
		servicesParser[undirectedServiceName] = userRouteCode
	}

	return servicesParser
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
	var buses BusesLocation
	err := json.Unmarshal([]byte(data), &buses)
	if err != nil {
		log.Error(err)
	}
	return buses.CanonicalForm()
}

func (buses BusesLocation) CanonicalForm() string {
	var busesCanonLst []string
	for _, busLocation := range buses.BusesLocation {
		busesCanonLst = append(busesCanonLst, busLocation.CanonicalForm())
	}
	busesCanon := strings.Join(busesCanonLst, ";")
	return busesCanon
}

func (b BusLocation) CanonicalForm() string {
	latStr := strconv.FormatFloat(b.Lat, 'f', -1, 64)
	lonStr := strconv.FormatFloat(b.Lon, 'f', -1, 64)
	speedStr := strconv.FormatFloat(b.Speed, 'f', -1, 64)
	values := []string{b.LicensePlate, b.Utc, latStr, lonStr, speedStr, b.Route, b.Direction}
	return strings.Join(values, ";")
}
