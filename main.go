package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

const secondMarkInit = 0

func main() {
	log.SetLevel(log.WarnLevel)
	oneSecondInNs := 1000000000
	for {
		now := time.Now()
		timeToWait := time.Duration(oneSecondInNs*(60-now.Second()) + (1000000000 - now.Nanosecond()) - oneSecondInNs + oneSecondInNs*secondMarkInit)
		log.Info("Waiting for the minute to end...")
		time.Sleep(timeToWait)
		log.Info("Collecting events...")
		collectEvents()
		log.Info("Events collected!")
	}
}
