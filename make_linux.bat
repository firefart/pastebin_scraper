@echo off
set GOOS=linux
set GOARCH=amd64

go test -v -race ./...
go build -o pastebin_scraper