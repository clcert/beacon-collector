package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/clcert/beacon-collector-go/db"
	"github.com/clcert/vdf/govdf"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"os"
	"time"
)

func getEventsCollectedHashed(timestamp time.Time) []string {
	dbConn := db.ConnectDB()
	defer dbConn.Close()

	var eventsCollectedHashed []string

	getEventsCollectedHashedStatement := `SELECT digest FROM events_collected WHERE pulse_timestamp = $1`
	rows, err := dbConn.Query(getEventsCollectedHashedStatement, timestamp)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": timestamp,
		}).Error("Failed to get events collected")
	}
	defer rows.Close()
	for rows.Next() {
		var eventCollectedHashed string
		err = rows.Scan(&eventCollectedHashed)
		if err != nil {
			log.WithFields(log.Fields{
				"pulseTimestamp": timestamp,
			}).Error("No events collected for this pulse")
		}
		eventsCollectedHashed = append(eventsCollectedHashed, eventCollectedHashed)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return eventsCollectedHashed

}

func generateExternalValue(eventsCollected []string, timestamp time.Time) {
	dbConn := db.ConnectDB()
	defer dbConn.Close()

	hashedEvents := hashEvents(eventsCollected)
	log.Debug("Calculating VDF...")
	externalEvent := vdf(hashedEvents)
	addEventStatement := `INSERT INTO external_events (value, pulse_timestamp, status_collected) VALUES ($1, $2, $3)`

	_, err := dbConn.Exec(addEventStatement, hex.EncodeToString(externalEvent), timestamp, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": timestamp,
		}).Error("Failed to add External Events to database")
	}
	log.Debug("Process Finalized!")
}

// H(e1 || e2 || ... || en)
func hashEvents(events []string) [64]byte {
	var concatenatedEvents string
	for _, l := range events {
		concatenatedEvents = concatenatedEvents + l
	}
	byteEvents := []byte(concatenatedEvents)
	return sha3.Sum512(byteEvents)
}

func vdf(events [64]byte) []byte {
	// Open our vdfConfig
	jsonFile, err := os.Open("utils/vdfConfig.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our dbConfig so that we can parse it later on
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var vdfConfig map[string]string
	json.Unmarshal(byteValue, &vdfConfig)

	seed := govdf.GetRandomSeed()
	govdf.SetServer(vdfConfig["vdfServer"])

	lbda := 1024
	T := 2000000

	y, _ := govdf.Eval(T, lbda, events[:], seed)
	return y
}
