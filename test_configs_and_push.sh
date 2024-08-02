#!/bin/bash

# Ensure the script is running in the correct directory
cd "$(dirname "$0")"

# Ensure Go is available in the cron job environment
export PATH=$PATH:/usr/local/go/bin

go get .
go build -o main
chmod +x main
./main --file sublinks.txt

git add output.txt
git commit -m "update output.txt, date: $(date)"
git push origin master
