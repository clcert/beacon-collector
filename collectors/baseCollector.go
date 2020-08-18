package collectors

import (
	"database/sql"
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
	collectEvent() (string, string, int)
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

	data, metadata, statusCollection := c.collectEvent()
	status := comparePrevious(metadata, statusCollection, c)

	canonical := c.getCanonicalForm(data)
	digest := generateDigest(canonical)
	sourceName := c.sourceName()
	estimatedEntropy := c.estimateEntropy()

	addEventStatement := `INSERT INTO events (source_name, raw_event, metadata, entropy_estimation, digest, event_status, pulse_timestamp, canonical_form) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := dbConn.Exec(addEventStatement, sourceName, data, metadata, estimatedEntropy, digest, status, recordTimestamp, canonical)
	if err != nil {
		log.WithFields(log.Fields{
			"pulseTimestamp": recordTimestamp,
			"sourceName":     sourceName,
		}).Error("failed to add event to database")
		log.Error(err)
	}
	log.Debugf("complete %s collection", c.sourceName())
}

func generateDigest(canonical string) string {
	if canonical == "" {
		return strings.Repeat("0", 128)
	}
	return fmt.Sprintf("%x", sha3.Sum512([]byte(canonical)))
}

func comparePrevious(currentMetadata string, currentStatus int, c Collector) int {
	dbConn := db.ConnectDB()
	defer dbConn.Close()
	var prevMetadata string

	getLastMetadataStatement := `SELECT metadata FROM events WHERE source_name = $1 ORDER BY id DESC LIMIT 1`
	lastMetadataRow := dbConn.QueryRow(getLastMetadataStatement, c.sourceName())
	switch err := lastMetadataRow.Scan(&prevMetadata); err {
	case sql.ErrNoRows:
		return currentStatus
	case nil:
		if currentMetadata == prevMetadata {
			return currentStatus | 4
		} else {
			return currentStatus
		}
	default:
		return currentStatus
	}
}
