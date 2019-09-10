package main

import "time"

type Event struct {
	Id            int
	TimestampInit time.Time
	TimestampEnd  time.Time
	RawValue      string
	Digest        string
	EntropyEst    float64
	Status        int
	RecordId      int
	SourceId      int
}
