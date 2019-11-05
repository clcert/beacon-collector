package main

import (
	"sync"
	"time"
)

func collectEvents() {
	//db := connectDB()
	//defer db.Close()

	now := time.Now().UTC()
	nextRecordTimestamp := now.Add(time.Minute)
	year, month, day := nextRecordTimestamp.Date()

	// Set the second and millisecond to 0
	nextRecordTimestamp = time.Date(year, month, day,
		nextRecordTimestamp.Hour(), nextRecordTimestamp.Minute(), 0, 0, nextRecordTimestamp.Location())

	var wg sync.WaitGroup

	var eqCollector EarthquakeCollector
	wg.Add(1)
	go process(eqCollector, nextRecordTimestamp, &wg)

	var twCollector TwitterCollector
	wg.Add(1)
	go process(twCollector, nextRecordTimestamp, &wg)

	var radioCollector RadioCollector
	wg.Add(1)
	process(radioCollector, nextRecordTimestamp, &wg)

	var ethCollector EthereumCollector
	wg.Add(1)
	go process(ethCollector, nextRecordTimestamp, &wg)

	wg.Wait()

	externalEventsCollected := getExternalEvents(nextRecordTimestamp)
	generateExternalEventField(externalEventsCollected, nextRecordTimestamp)

}
