package main

import (
	"time"
)

type Record struct {
	Id             int       `json:"id"`
	OutputValue    string    `json:"outputValue"`
	Timestamp      time.Time `json:"timestamp"`
	SignatureValue string    `json:"signatureValue"`
	Status         int       `json:"status"`
	HashedMessage  string    `json:"hashedMessage"`
	Witness        string    `json:"witness"`
	PulseIndex     int       `json:"pulseIndex"`
	ChainIndex     int       `json:"chainIndex"`
}
