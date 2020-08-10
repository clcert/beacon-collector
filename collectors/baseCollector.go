package collectors

import (
	"fmt"
	"github.com/clcert/beacon-collector-go/db"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"strings"
	"sync"
	"time"
)

type Collector interface {
	sourceName() string
	collectEvent() (string, string)
	estimateEntropy() int
	getCanonicalForm(string) string
}

type Source struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func Process(c Collector, recordTimestamp time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	dbConn := db.ConnectDB()
	defer dbConn.Close()

	data, metadata := c.collectEvent()

	canonical := c.getCanonicalForm(data)
	digest := generateDigest(canonical)
	sourceName := c.sourceName()
	estimatedEntropy := c.estimateEntropy()

	addEventStatement := `INSERT INTO events_collected (source_name, raw_event, metadata, entropy_estimation, digest, event_status, pulse_timestamp) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := dbConn.Exec(addEventStatement, sourceName, data, metadata, estimatedEntropy, digest, 0, recordTimestamp)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": recordTimestamp,
			"sourceName":     sourceName,
		}).Error("failed to add event to database")
		panic(err)
	}
	log.Debugf("complete %s collection", c.sourceName())
}

func generateDigest(canonical string) string {
	if canonical == "" {
		return strings.Repeat("0", 128)
	}
	return fmt.Sprintf("%x", sha3.Sum512([]byte(canonical)))
}
