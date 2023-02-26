package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/clcert/beacon-collector-go/db"
	"github.com/clcert/vdf/govdf"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

func getEventsCollectedHashed(timestamp time.Time) []string {
	dbConn := db.ConnectDB()
	defer dbConn.Close()

	var eventsCollectedHashed []string

	getEventsCollectedHashedStatement := `SELECT digest FROM events WHERE pulse_timestamp = $1`
	rows, err := dbConn.Query(getEventsCollectedHashedStatement, timestamp)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": timestamp,
		}).Error("failed to get events collected")
	}
	defer rows.Close()
	for rows.Next() {
		var eventCollectedHashed string
		err = rows.Scan(&eventCollectedHashed)
		if err != nil {
			log.WithFields(log.Fields{
				"pulseTimestamp": timestamp,
			}).Error("no events collected for this pulse")
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
	log.Debug("Computing VDF...")
	seed := govdf.GetRandomSeed()
	vdfOutput, vdfProof := vdf(seed, hashedEvents)
	log.Debug("VDF done. Saving...")
	externalValue := sha3.Sum512(vdfOutput)
	addEventStatement := `INSERT INTO external_values (vdf_preimage, vdf_seed, vdf_output, vdf_proof, external_value, pulse_timestamp, status) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := dbConn.Exec(
		addEventStatement,
		hex.EncodeToString(hashedEvents[:]),  // preimage
		hex.EncodeToString(seed),             // VDF seed
		hex.EncodeToString(vdfOutput),        // VDF output
		hex.EncodeToString(vdfProof),         // VDF proof
		hex.EncodeToString(externalValue[:]), // external value
		timestamp, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": timestamp,
		}).Error("failed to add external value to database")
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

func vdf(seed []byte, events [64]byte) ([]byte, []byte) {
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

	govdf.SetServer(vdfConfig["vdfServer"])

	lbda := 1024
	T := 2000000

	out, proof := govdf.Eval(T, lbda, events[:], seed)
	return out, proof
}

type EventSimplified struct {
	Id             int       `json:"id"`
	PulseTimestamp time.Time `json:"pulseTimestamp"`
}

func CleanOldEvents(wg *sync.WaitGroup) {
	defer wg.Done()
	// Clean old events that don't fulfill the following requirements:
	// 1. are less than X hour old
	// 2. his output value is not present in previous_hour, previous_day, previous_month or previous_year of any pulse
	dbConn := db.ConnectDB()
	defer dbConn.Close()

	now := time.Now().UTC()
	limitTimestamp := now.Add(-time.Hour)

	var defaultMessage = "delete"
	// getPossiblesStatement := `SELECT id, pulse_timestamp FROM events WHERE digest != $1 AND pulse_timestamp < $2`
	getPossiblesStatement := `SELECT id, pulse_timestamp FROM events WHERE digest != $1 AND pulse_timestamp < $2 AND (event_status & 1) != 1`
	rows, err := dbConn.Query(getPossiblesStatement, defaultMessage, limitTimestamp)
	if err != nil {
		// something
		panic(err)
	}
	var allPossibleEvents []EventSimplified
	defer rows.Close()
	for rows.Next() {
		var possibleEventToDelete EventSimplified
		err = rows.Scan(&possibleEventToDelete.Id, &possibleEventToDelete.PulseTimestamp)
		if err != nil {
			panic(err)
		}
		allPossibleEvents = append(allPossibleEvents, possibleEventToDelete)
	}
	var eventsIDToDelete []int
	for _, event := range allPossibleEvents {
		if event.PulseTimestamp.Minute() != 0 {
			eventsIDToDelete = append(eventsIDToDelete, event.Id)
		}
	}

	deleteRawStatement := `UPDATE events SET raw_event = $1, canonical_form = $1 WHERE id = ANY($2)`
	_, err = dbConn.Exec(deleteRawStatement, defaultMessage, pq.Array(eventsIDToDelete))
	if err != nil {
		panic(err)
	}

}
