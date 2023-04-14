package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/clcert/beacon-collector-go/db"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

func parseMd(content string) string {
	cleanned := strings.ReplaceAll(content, "*", "")
	cleanned = strings.ReplaceAll(cleanned, "`", "")
	cleanned = strings.ReplaceAll(cleanned, "# ", "")
	cleanned = strings.ReplaceAll(cleanned, "#", "")
	cleanned = strings.ReplaceAll(cleanned, "[", "<")
	cleanned = strings.ReplaceAll(cleanned, "]", ">")
	return cleanned
}

func insertDescriptionInDB(content string, digestedContent string) {
	dbConn := db.ConnectDB()
	defer dbConn.Close()

	addExternalDescription := `INSERT INTO external_sources_info (source_id, source_description, source_status) 
		 VALUES ($1, $2, 0) ON CONFLICT DO NOTHING`
	_, err := dbConn.Exec(addExternalDescription, digestedContent, content)
	if err != nil {
		log.WithFields(log.Fields{
			"source_id":          digestedContent,
			"source_description": content,
		}).Error("failed to add externals description to database")
		log.Error(err)
	}
	log.Info("external values description added to database")
}

func main() {
	filename := os.Args[1]
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	parsedContent := parseMd(string(content))

	hashObject := sha3.New512()
	hashObject.Write([]byte(parsedContent))
	digest := hashObject.Sum(nil)
	digestedContent := fmt.Sprintf("%x", digest)
	insertDescriptionInDB(parsedContent, digestedContent)
}
