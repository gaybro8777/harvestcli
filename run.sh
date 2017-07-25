#!/bin/sh

APP_ID=$1
FULL_DATE=$2
DATE_STR=$(echo ${FULL_DATE////})

OUTPUT_DIR=tmp/$APP_ID/$FULL_DATE
OUTPUT_FILE=aggr_searches_${APP_ID}_${DATE_STR}.json
GCS_TARGET=gs://harvestevents/1/aggr_search/$APP_ID/$FULL_DATE/$OUTPUT_FILE
BQ_TABLE=alg-insights:harvestevents.1_aggr_search_${APP_ID}_${DATE_STR}

printf "Processing logs for App:%s on:%s%s\n" $APP_ID $FULL_DATE

if [ -d "$OUTPUT_DIR" ]; then
  printf "Temp directory %s already exists, overwriting...\n" $OUTPUT_DIR
  rm -rf $OUTPUT_DIR
fi

mkdir -p $OUTPUT_DIR/clicks
mkdir -p $OUTPUT_DIR/searches

printf "Fetching events...\n"
gsutil -m cp -r gs://harvestevents/1/click/$APP_ID/$FULL_DATE/* $OUTPUT_DIR/clicks
gsutil -m cp -r gs://harvestevents/1/search/$APP_ID/$FULL_DATE/* $OUTPUT_DIR/searches

printf "Merging logs...\n"
find $OUTPUT_DIR/clicks -name '*.json' -exec cat {} \; > $OUTPUT_DIR/clicks.json
find $OUTPUT_DIR/searches -name '*.json' -exec cat {} \; > $OUTPUT_DIR/searches.json

printf "Converting JSON input into CSV...\n"
go run main.go convert-csv -i $OUTPUT_DIR/searches.json -o $OUTPUT_DIR/searches.csv
go run main.go convert-csv -i $OUTPUT_DIR/clicks.json -o $OUTPUT_DIR/clicks.csv

printf "Associating clicks to searches...\n"
go run main.go associate -c $OUTPUT_DIR/clicks.csv -s $OUTPUT_DIR/searches.csv -o $OUTPUT_DIR/associated.csv

printf "Sorting searches...\n"
sort --field-separator=',' --key 3,3 --key 5,5 --key 1n,1n $OUTPUT_DIR/associated.csv > $OUTPUT_DIR/sorted.csv

printf "Figuring out terminal searches...\n"
go run main.go merge -s $OUTPUT_DIR/sorted.csv -o $OUTPUT_DIR/merged.csv

printf "Converting CSV input into JSON...\n"
go run main.go convert-json -i $OUTPUT_DIR/merged.csv -o $OUTPUT_DIR/$OUTPUT_FILE

printf "Uploading to GCS...\n"
gsutil cp $OUTPUT_DIR/$OUTPUT_FILE $GCS_TARGET

if [ $? -eq 0 ]
then
  printf "Done. Resulting file is available at %s\n" $GCS_TARGET
else
  exit
fi

printf "Loading into BQ...\n"


printf "Target table is %s, checking it does not already exists...\n" $BQ_TABLE
bq show $BQ_TABLE

if [ $? -eq 0 ]
then
  printf "BQ Table %s already exists. Not loading.\n" $BQ_TABLE
  printf "To manually populate this table, run: %s\n" "bq load --source_format=NEWLINE_DELIMITED_JSON $BQ_TABLE $GCS_TARGET timestamp:integer,index:string,appID:string,queryID:string,userID:string,context:string,query:string,queryParameters:string"
  exit
fi

bq load --source_format=NEWLINE_DELIMITED_JSON $BQ_TABLE $GCS_TARGET timestamp:integer,index:string,appID:string,queryID:string,userID:string,context:string,query:string,queryParameters:string

if [ $? -eq 0 ]
then
  printf "Done. Data loaded into %s\n" $BQ_TABLE
fi
