package main

import (
	"fmt"
	"golang.org/x/crypto/sha3"
	"strings"
	"time"
)

type Collector interface {
	collectEvent() string
	estimateEntropy() int
}

func process(c Collector) {
	db := connectDB()
	defer db.Close()

	t1 := strings.Join(strings.Split(time.Now().UTC().String(), " ")[0:2], " ")
	data := collectEvent(c)
	t2 := strings.Join(strings.Split(time.Now().UTC().String(), " ")[0:2], " ")

	digest := generateDigest(data)
	estimatedEntropy := estimateEntropy(c)

	addEventStatement := `INSERT INTO events (timestamp_init, timestamp_end, raw_value, digest, entropy_est, status) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(addEventStatement, t1, t2, data, digest, estimatedEntropy, 0)
	if err != nil {
		panic(err)
	}

}

func generateDigest(s string) string {
	return fmt.Sprintf("%x", sha3.Sum512([]byte(s)))
}
