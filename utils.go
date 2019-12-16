package main

import (
	"encoding/hex"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"time"
)

func getEventsCollectedHashed(timestamp time.Time) []string {
	db := connectDB()
	defer db.Close()

	var eventsCollectedHashed []string

	getEventsCollectedHashedStatement := `SELECT digest FROM events_collected WHERE pulse_timestamp = $1`
	rows, err := db.Query(getEventsCollectedHashedStatement, timestamp)
	if err != nil {
		log.WithFields(logrus.Fields{
			"pulseTimestamp": timestamp,
		}).Panic("Failed to get events collected")
	}
	defer rows.Close()
	for rows.Next() {
		var eventCollectedHashed string
		err = rows.Scan(&eventCollectedHashed)
		if err != nil {
			log.WithFields(logrus.Fields{
				"pulseTimestamp": timestamp,
			}).Panic("No events collected for this pulse")
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
	db := connectDB()
	defer db.Close()

	hashedEvents := hashEvents(eventsCollected)
	externalEvent := vdf(hashedEvents)
	addEventStatement := `INSERT INTO external_events (value, pulse_timestamp, status) VALUES ($1, $2, $3)`

	_, err := db.Exec(addEventStatement, hex.EncodeToString(externalEvent[:]), timestamp, 0)
	if err != nil {
		log.WithFields(logrus.Fields{
			"pulseTimestamp": timestamp,
		}).Panic("Failed to add External Events to database")
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
