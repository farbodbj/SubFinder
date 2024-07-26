#!/bin/bash

# Define the cron job
NEW_CRON_JOB="36 18 * * * $(pwd)/test_configs_and_push.sh >> $(pwd)/job.log 2>&1"

# Remove the old cron job if it exists
(crontab -l | grep -vF "test_configs_and_push.sh") | crontab -

# Add the new cron job
(crontab -l; echo "$NEW_CRON_JOB") | crontab -

echo "Cron job updated to: $NEW_CRON_JOB"

