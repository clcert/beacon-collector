package collectors

import (
	"encoding/json"
	"fmt"
	"github.com/clcert/beacon-collector-go/db"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Collector interface {
	sourceName() string
	collectEvent() (string, string)
	estimateEntropy() int
}

type Source struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func Process(c Collector, recordTimestamp time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	dbConn := db.ConnectDB()
	defer dbConn.Close()

	data, metadata := c.collectEvent()

	digest := generateDigest(data)
	sourceId := sourceId(c)
	estimatedEntropy := c.estimateEntropy()

	addEventStatement := `INSERT INTO events_collected (source_id, raw_event, metadata, entropy_estimation, digest, event_status, pulse_timestamp) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := dbConn.Exec(addEventStatement, sourceId, data, metadata, estimatedEntropy, digest, 0, recordTimestamp)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": recordTimestamp,
			"sourceId":       sourceId,
		}).Error("Failed to add event to database")
		panic(err)
	}

}

func sourceId(c Collector) int {
	// Open our dbConfig
	jsonFile, err := os.Open("collectors/sourcesConfig.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our dbConfig so that we can parse it later on
	defer jsonFile.Close()

	var sources []Source
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &sources)

	for _, source := range sources {
		if source.Name == c.sourceName() {
			return source.Id
		}
	}

	return -1

}

func generateDigest(s string) string {
	return fmt.Sprintf("%x", sha3.Sum512([]byte(s)))
}
