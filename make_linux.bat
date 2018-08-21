@echo off
echo Updating Dependencies
go get -u gopkg.in/gomail.v2

echo Running gometalinter
gometalinter ./...

echo Running Tests
go test -v ./...

echo Running build
set GOOS=linux
set GOARCH=amd64
go build -o pastebin_scraper