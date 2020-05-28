package collectors

import (
	"fmt"
	"github.com/clcert/beacon-collector-go/db"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"sync"
	"time"
)

type Collector interface {
	collectEvent() (string, string)
	estimateEntropy() int
	sourceID() int
}

func Process(c Collector, recordTimestamp time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	dbConn := db.ConnectDB()
	defer dbConn.Close()

	data, metadata := c.collectEvent()

	digest := generateDigest(data)
	estimatedEntropy := c.estimateEntropy()

	addEventStatement := `INSERT INTO events_collected (source_id, raw_event, metadata, entropy_estimation, digest, event_status, pulse_timestamp) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := dbConn.Exec(addEventStatement, c.sourceID(), data, metadata, estimatedEntropy, digest, 0, recordTimestamp)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": recordTimestamp,
			"sourceId":       c.sourceID(),
		}).Error("Failed to add event to database")
		panic(err)
	}
	log.Debugf("Source #%d Completed!", c.sourceID())
}

func generateDigest(s string) string {
	return fmt.Sprintf("%x", sha3.Sum512([]byte(s)))
}
