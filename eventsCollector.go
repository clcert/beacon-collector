package main

import (
	"time"
)

func collectEvents() {
	db := connectDB()
	defer db.Close()

	now := time.Now().UTC()
	nextRecordTimestamp := now.Add(time.Minute)
	year, month, day := nextRecordTimestamp.Date()

	// Set the second and millisecond to 0
	nextRecordTimestamp = time.Date(year, month, day,
		nextRecordTimestamp.Hour(), nextRecordTimestamp.Minute(), 0, 0, nextRecordTimestamp.Location())

	// TODO: Parallelize this

	var eqCollector EarthquakeCollector
	go process(eqCollector, nextRecordTimestamp)

	var twCollector TwitterCollector
	go process(twCollector, nextRecordTimestamp)

	// var radioCollector RadioCollector
	// go process(radioCollector, nextRecordTimestamp)

	var ethCollector EthCollector
	go process(ethCollector, nextRecordTimestamp)

}
