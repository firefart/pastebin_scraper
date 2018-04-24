BINARY := pastebin_scraper
VERSION ?= latest
OUTDIR := out
ARCH := amd64
PLATFORMS := windows linux darwin
os = $(word 1, $@)

.DEFAULT_GOAL := build

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(os) GOARCH=$(ARCH) go build -o $(OUTDIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH)
	zip -j $(OUTDIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH).zip $(OUTDIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH)

.PHONY: release
release: clean deps $(PLATFORMS)

.PHONY: build
build: deps
	go build .

.PHONY: deps
deps:
	go get -u gopkg.in/gomail.v2

.PHONY: clean
clean:
	rm -rf $(OUTDIR)/*
