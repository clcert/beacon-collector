package main

import "time"

func collectEvents() {
	db := connectDB()
	defer db.Close()

	now := time.Now().UTC()
	nextRecordTimestamp := now.Add(time.Minute)
	var nextRecordId int

	createRecordStatement := `INSERT INTO records (id, timestamp) VALUES ($1, $2)`
	_, err := db.Exec(createRecordStatement, nextRecordId, nextRecordTimestamp)
	if err != nil {
		panic(err)
	}

	var eqCollector EarthquakeCollector
	process(eqCollector.eqCollector)
}
