#!/bin/bash

# Check if both arguments are provided
if [ $# -ne 3 ]; then
    echo "Usage: $0 <API_TOKEN> <CHAT_ID> <FILE> <N_ITEMS>"
    exit 1
fi

API_TOKEN="$1"
CHAT_ID="$2"
FILE="$3"

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