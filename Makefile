BINARY := pastebin_scraper
VERSION ?= latest
OUTDIR := out
ARCH := amd64
PLATFORMS := windows linux darwin
os = $(word 1, $@)

GOPATH := $(or $(GOPATH), $(HOME)/go)
BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

.DEFAULT_GOAL := build

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(os) GOARCH=$(ARCH) go build -o $(OUTDIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH)
	zip -j $(OUTDIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH).zip $(OUTDIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH)

.PHONY: release
release: clean deps lint $(PLATFORMS)

.PHONY: build
build:
	go build .

.PHONY: deps
deps:
	go get -u gopkg.in/gomail.v2

.PHONY: clean
clean:
	rm -rf $(OUTDIR)/*

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	$(GOMETALINTER) --install &> /dev/null

.PHONY: lint
lint: deps $(GOMETALINTER)
	$(BIN_DIR)/gometalinter .
