package main

import (
	"github.com/sirupsen/logrus"
	"time"
)

const secondMarkInit = 0

var log = logrus.New()

func main() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.WarnLevel)
	oneSecondInNs := 1000000000
	for {
		now := time.Now()
		timeToWait := time.Duration(oneSecondInNs*(60-now.Second()) + (1000000000 - now.Nanosecond()) - oneSecondInNs + oneSecondInNs*secondMarkInit)
		log.Debug("Waiting for the minute to end...")
		time.Sleep(timeToWait)
		collectEvents()
	}
}
