#!/bin/bash

# Ensure the script is running in the correct directory
cd "$(dirname "$0")"

# Ensure Go is available in the cron job environment
export PATH=$PATH:/usr/local/go/bin

bash docker-entrypoint.sh
git pull --rebase

# Check if 'output.txt' has changed
if git diff --exit-code data/output.txt > /dev/null; then
    echo "No changes detected in output.txt. Skipping commit and push."
else
    echo "Changes detected in output.txt. Committing and pushing changes."
    git add data/output.txt
    git commit -m "update output.txt, date: $(date)"
    git push origin master
fi

