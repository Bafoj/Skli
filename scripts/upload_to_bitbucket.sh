#!/bin/bash

# script to upload GoReleaser artifacts to Bitbucket Downloads
# Uses BITBUCKET_API_KEY (App Password) and BITBUCKET_USERNAME from .env

set -e

REPO="cuatroochenta/skli"
DIST_DIR="dist"

# Load .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

if [ -z "$BITBUCKET_USERNAME" ] || [ -z "$BITBUCKET_API_KEY" ]; then
    echo "Error: BITBUCKET_USERNAME and BITBUCKET_API_KEY must be set in .env."
    echo "Note: BITBUCKET_API_KEY should be a Bitbucket App Password."
    exit 1
fi

echo "Uploading artifacts to Bitbucket Downloads ($REPO)..."

# Upload archives and checksums
for file in "$DIST_DIR"/*.tar.gz "$DIST_DIR"/*.zip "$DIST_DIR"/checksums.txt; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        echo "  -> $filename"
        
        # Bitbucket Downloads API: POST /repositories/{workspace}/{repo_slug}/downloads
        # Using Basic Auth (-u) for App Passwords
        tmp_resp="/tmp/bitbucket_resp_$$.txt"
        http_code=$(curl -s -o "$tmp_resp" -w "%{http_code}" -X POST -u "$BITBUCKET_USERNAME:$BITBUCKET_API_KEY" \
             "https://api.bitbucket.org/2.0/repositories/$REPO/downloads" \
             -F files=@"$file")
        
        body=$(cat "$tmp_resp")
        rm -f "$tmp_resp"
        
        if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
            # Success
            continue
        elif [ "$http_code" -eq 409 ]; then
            echo "     [SKIP] El archivo ya existe en Bitbucket."
        else
            echo "     [ERROR] Falló la subida de $filename (HTTP $http_code)"
            echo "     Detalles: $body"
            exit 1
        fi
    fi
done

echo "Done! Artifacts are available in the Downloads section of Bitbucket."
