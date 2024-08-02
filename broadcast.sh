#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    source .env
else
    echo "Error: .env file not found."
    exit 1
fi

# Ensure required environment variables are set
if [ -z "$API_TOKEN" ] || [ -z "$CHAT_ID" ] || [ -z "$FILE" ]; then
    echo "Error: One or more required environment variables are missing in .env file."
    exit 1
fi

HEADER="V2ray Config:"

# Check if the file exists and read the top 10 lines
if [ -f "$FILE" ]; then
    count=0
    while IFS= read -r line && [ $count -lt 1 ]; do
        # Build the message
        MESSAGE="${HEADER}%0A%0A${line}%0A%0AFollow us on: $CHAT_ID"

        # Send each line as a separate message
        curl -s -X POST "https://api.telegram.org/bot$API_TOKEN/sendMessage" \
        -d chat_id="$CHAT_ID" \
        -d text="$MESSAGE"
        count=$((count + 1))
    done < "$FILE"
else
    echo "File $FILE not found."
    exit 1
fi