# photon-download-workflow
Temporal workflow to download photon data

Create scheduled workflow for photon download

```
go run ./src/cmd create
```

Start worker that downloads latest photon archive.

```
S3_ENDPOINT_URL="REPLACE_ME" \
S3_REGION="us-west-1" \
BUCKET="butcket-name" \
ARCHIVE_URL="https://download1.graphhopper.com/public/extracts/by-country-code/us/photon-db-us-latest.tar.bz2" \
ARCHIVE_KEY="north-america/us/photon-db-us-latest.tar.bz2" \
AWS_ACCESS_KEY_ID=REPLACE_ME \
AWS_SECRET_ACCESS_KEY=REPLACE_ME \
go run ./src/cmd worker
```