#!/bin/bash

go get .
go build -o main
chmod +x main
./main --file sublinks.txt

git add output.txt
git commit -m "update output.txt, date: $(date)"
git push origin master
