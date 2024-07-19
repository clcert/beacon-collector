package main

import (
	"github.com/clcert/beacon-collector/utils"
	log "github.com/sirupsen/logrus"
)

const secondMarkInit = 0

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
	utils.AggregateEvents()
}
