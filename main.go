package main

import (
	"time"

	"github.com/clcert/beacon-collector-go/utils"
	log "github.com/sirupsen/logrus"
)

const secondMarkInit = 0

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
	oneSecondInNs := 1000000000
	for {
		now := time.Now()
		timeToWait := time.Duration(oneSecondInNs*(60-now.Second()) + (1000000000 - now.Nanosecond()) - oneSecondInNs + oneSecondInNs*secondMarkInit)
		log.Debug("waiting for the next minute...")
		time.Sleep(timeToWait)
		utils.AggregateEvents()
	}
}
