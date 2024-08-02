#!/bin/bash

# Check if both arguments are provided
if [ $# -ne 3 ]; then
    echo "Usage: $0 <API_TOKEN> <CHAT_ID> <CONFIG>"
    exit 1
fi

API_TOKEN="$1"
CHAT_ID="$2"
CONFIG="$3"

# Set the message text
MESSAGE="$CONFIG"

# Use curl to send the message
curl -s -X POST "https://api.telegram.org/bot$API_TOKEN/sendMessage" -d chat_id="$CHAT_ID" -d text="$MESSAGE"