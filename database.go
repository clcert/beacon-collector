package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"os"
)

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"dbname"`
}

func connectDB() *sql.DB {
	// Open our dbConfig
	jsonFile, err := os.Open("dbConfig.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our dbConfig so that we can parse it later on
	defer jsonFile.Close()

	var dbConfig DBConfig
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &dbConfig)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}
