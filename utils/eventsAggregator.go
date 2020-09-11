package utils

import (
	"github.com/clcert/beacon-collector-go/collectors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

func AggregateEvents() {
	log.Debug("start events collection")

	now := time.Now().UTC()
	nextRecordTimestamp := now.Add(time.Minute)
	year, month, day := nextRecordTimestamp.Date()

	// Set the second and millisecond to 0
	nextRecordTimestamp = time.Date(year, month, day,
		nextRecordTimestamp.Hour(), nextRecordTimestamp.Minute(), 0, 0, nextRecordTimestamp.Location())

	var wg sync.WaitGroup

	var eqCollector collectors.EarthquakeCollector
	wg.Add(1)
	go collectors.Process(eqCollector, nextRecordTimestamp, &wg)

	var twCollector collectors.TwitterCollector
	wg.Add(1)
	go collectors.Process(twCollector, nextRecordTimestamp, &wg)

	var radioCollector collectors.RadioCollector
	wg.Add(1)
	go collectors.Process(radioCollector, nextRecordTimestamp, &wg)

	var ethCollector collectors.EthereumCollector
	wg.Add(1)
	go collectors.Process(ethCollector, nextRecordTimestamp, &wg)

	// wg.Add(1)
	// go CleanOldEvents(&wg)

	wg.Wait()

	log.Debug("finish events collection")

	generateExternalValue(getEventsCollectedHashed(nextRecordTimestamp), nextRecordTimestamp)
}
