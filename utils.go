package main

import (
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"time"
)

func getExternalEvents(timestamp time.Time) []string {
	db := connectDB()
	defer db.Close()

	var externalEvents []string

	getEventsStatement := `SELECT digest FROM events WHERE record_timestamp = $1`
	rows, err := db.Query(getEventsStatement, timestamp)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var externalEvent string
		err = rows.Scan(&externalEvent)
		if err != nil {
			panic(err)
		}
		externalEvents = append(externalEvents, externalEvent)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return externalEvents

}

func generateExternalEventField(externalEvents []string, timestamp time.Time) {
	db := connectDB()
	defer db.Close()

	hashedEvents := hashEvents(externalEvents)
	externalEvent := vdf(hashedEvents)
	addEventStatement := `INSERT INTO external_events (value, record_timestamp) VALUES ($1, $2)`

	_, err := db.Exec(addEventStatement, hex.EncodeToString(externalEvent[:]), timestamp)
	if err != nil {
		panic(err)
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
