#!/bin/bash

go get -d -v ./...
go build -o main
chmod +x main
./main

git add output.txt
git commit -m "update output.txt, date: $(date)"
git push origin master
