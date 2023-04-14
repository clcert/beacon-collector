#!/bin/sh
FILE="../README.md"
EXT_TXT_SRC_DESC=$(go run parse_md.py $FILE parse)
EXT_SRC_ID=$(go run apply_sha3.py $FILE hash)

# DATABASE CONFIGURATIONS
DB_HOST=$(grep -o '"host": "[^"]*' ../db/dbConfig.json | grep -o '[^"]*$')
DB_PORT=$(grep -o '"port": "[^"]*' ../db/dbConfig.json | grep -o '[^"]*$')
DB_USER=$(grep -o '"user": "[^"]*' ../db/dbConfig.json | grep -o '[^"]*$')
DB_PASS=$(grep -o '"password": "[^"]*' ../db/dbConfig.json | grep -o '[^"]*$')
DB_NAME=$(grep -o '"dbname": "[^"]*' ../db/dbConfig.json | grep -o '[^"]*$')

export PGPASSWORD=$DB_PASS

psql -h $DB_HOST -U $DB_USER -d $DB_NAME \
     -c "INSERT INTO external_sources_info (source_id, source_description, source_status) VALUES ('$EXT_SRC_ID', '$EXT_TXT_SRC_DESC', 1);"