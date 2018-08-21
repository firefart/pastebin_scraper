@echo off
echo Updating Dependencies
go get -u gopkg.in/gomail.v2

echo Running gometalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install > nul
gometalinter ./...

echo Running Tests
go test -v -race ./...

echo Running build
set GOOS=linux
set GOARCH=amd64
go build -o pastebin_scraper