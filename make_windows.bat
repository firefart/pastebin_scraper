@echo off
SET BUILDARGS=-ldflags="-s -w" -gcflags="all=-trimpath=%GOPATH%\src" -asmflags="all=-trimpath=%GOPATH%\src"

echo Updating Dependencies
go get -u gopkg.in/gomail.v2

echo Running gometalinter
go get -u github.com/alecthomas/gometalinter
gometalinter --install > nul
gometalinter ./...

echo Running Tests
go test -v ./...

echo Running build
set GOOS=windows
set GOARCH=amd64
go build %BUILDARGS% -o pastebin_scraper.exe
