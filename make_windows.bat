@echo off

echo Updating Dependencies
go get -u gopkg.in/gomail.v2

echo Running gometalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install > nul
gometalinter ./...

echo Running tests
go test -v ./...

echo Building program
set GOOS=windows
set GOARCH=amd64
go build -o pastebin_scraper.exe