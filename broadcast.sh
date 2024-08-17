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

# Ensure optional proxy environment variables are set
USE_PROXY=false
if [ -n "$PROXY_IP" ] && [ -n "$PROXY_PORT" ]; then
    USE_PROXY=true
    PROXY_URL="http://$PROXY_IP:$PROXY_PORT"
fi

HEADER="V2ray Config:"

# Function to send message via Telegram with retry logic
send_message() {
    local message="$1"
    local retries=3
    local count=0

    while [ $count -lt $retries ]; do
        if [ "$USE_PROXY" = true ]; then
            curl -s -X POST "https://api.telegram.org/bot$API_TOKEN/sendMessage" \
            -d chat_id="$CHAT_ID" \
            -d text="$message" \
	    -d parse_mode='MarkdownV2' \
            --proxy "$PROXY_URL"
        else
            curl -s -X POST "https://api.telegram.org/bot$API_TOKEN/sendMessage" \
            -d chat_id="$CHAT_ID" \
            -d text="$message"
	    -d parse_mode='MarkdownV2'
        fi

        if [ $? -eq 0 ]; then
            echo "Message sent successfully."
            return 0
        else
            echo "Failed to send message. Retrying... ($((count + 1))/$retries)"
            sleep 2
            count=$((count + 1))
        fi
    done

    echo "Failed to send message after $retries attempts."
    return 1
}

# Check if the file exists and read the top 10 lines
if [ -f "$FILE" ]; then
    count=0
    while IFS= read -r line && [ $count -lt 10 ]; do
        # Build the message
	fence='`'
        MESSAGE="${HEADER}%0A%0A${fence}${line}${fence}%0A%0AFollow us on: $CHAT_ID"

        # Send each line as a separate message with retry
        send_message "$MESSAGE"
        count=$((count + 1))
    done < "$FILE"
else
    echo "File $FILE not found."
    exit 1
fi
