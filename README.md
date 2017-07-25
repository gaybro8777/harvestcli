# harvestcli

Scripts/CLI to manipulate harvest.
At the moment it allows, for a given appID and day, to aggregate queries and upload the result to GCS and BQ.

## Requirements

* `go`, `gsutil` and `bq` must be available
* You must have read/write access to GCP project `alg-insights` (GCS and BQ)

## Usage

`./run.sh APP_ID YYYY/MM/DD`

e.g: `./run.sh UJ5WYC0L7X 2017/07/24`

## Explanation

Under the hood, the script performs the following actions:

* fetch events for the given APP_ID and DATE (in `tmp/APP_ID/DATE/EVENT_TYPE/`
* merge queries together, outputing only terminal searches
* upload result to GCS (in 1/aggr_search/APP_ID/DATE, e.g: `gs://harvestevents/1/aggr_search/UJ5WYC0L7X/2017/07/24/aggr_searches_UJ5WYC0L7X_20170724.json`)
* load data into BQ (e.g in `alg-insights:harvestevents.1_aggr_search_UJ5WYC0L7X_20170724`)

See `run.sh` for more details.
