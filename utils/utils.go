package utils

import (
	"encoding/hex"
	"github.com/clcert/beacon-collector-go/db"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"sync"
	"time"
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
	db := db.ConnectDB()
	defer db.Close()

	hashedEvents := hashEvents(eventsCollected)
	externalEvent := vdf(hashedEvents)
	addEventStatement := `INSERT INTO external_values (value, pulse_timestamp, status) VALUES ($1, $2, $3)`

	_, err := db.Exec(addEventStatement, hex.EncodeToString(externalEvent[:]), timestamp, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": timestamp,
		}).Error("failed to add external value to database")
	}
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

func vdf(events [64]byte) [64]byte {
	return sha3.Sum512(events[:])
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
