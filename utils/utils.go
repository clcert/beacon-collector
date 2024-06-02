package utils

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/clcert/beacon-collector/collectors"
	"github.com/clcert/beacon-collector/db"
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

	// VDF execution
	log.Info("executing vdf...")
	seed := getRandomBytes(16) // Challenge
	vdfOutput, vdfProof := VDFeval(hashedEvents[:], seed)
	log.Info("vdf done, saving...")

	vdfOutStr := hex.EncodeToString(vdfOutput)
	externalValue := collectors.GenerateDigest(vdfOutStr)
	addEventStatement := `INSERT INTO external_values (vdf_preimage, vdf_seed, vdf_output, vdf_proof, external_value, pulse_timestamp, status) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := dbConn.Exec(
		addEventStatement,
		hex.EncodeToString(hashedEvents[:]), // preimage
		hex.EncodeToString(seed),            // VDF seed
		vdfOutStr,                           // VDF output
		hex.EncodeToString(vdfProof),        // VDF proof
		externalValue,                       // external value
		timestamp, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": timestamp,
		}).Error("failed to add external value to database")
	}
	log.Info("external value added to database")
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

func CleanOldEvents(wg *sync.WaitGroup) {
	defer wg.Done()
	// Clean events that:
	// 1. are older than 1 hour
	// 2. are not in the first pulse of an hour
	// TODO: we could also delete events that have some specific status
	dbConn := db.ConnectDB()
	defer dbConn.Close()

	var defaultMessage = "DELETED"
	log.Info("cleaning old events from database...")

	// 1 hour old and not minute 0
	replaceRawEventsStatement :=
		`UPDATE events SET raw_event = $1, canonical_form = $1 ` +
			`WHERE pulse_timestamp >= (NOW() - INTERVAL '1 hour 1 min') ` +
			`AND pulse_timestamp < (NOW() - INTERVAL '1 hour') ` +
			`AND EXTRACT(minute FROM pulse_timestamp) != 0`
	_, err := dbConn.Exec(replaceRawEventsStatement, defaultMessage)
	if err != nil {
		log.Error("failed in deleting raw events")
		panic(err)
	}

	log.Info("old events cleaned!")
}
