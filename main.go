package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	debug = flag.Bool("debug", false, "Print debug output")

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	keywordsRegex = make(map[string]keywordRegexType)
)

type keywordRegexType struct {
	regexp     *regexp.Regexp
	exceptions []string
}

func debugOutput(s string, a ...interface{}) {
	if *debug {
		log.Printf("[DEBUG] %s", fmt.Sprintf(s, a...))
	}
}

func checkKeywords(body string) (bool, map[string]string) {
	found := make(map[string]string)
	status := false
	for k, v := range keywordsRegex {
		if v.regexp != nil {
			if s := v.regexp.FindStringSubmatch(body); s != nil {
				match := strings.TrimSpace(s[1])
				if !checkExceptions(match, v.exceptions) {
					found[k] = match
					status = true
				}
			}
		}
	}
	return status, found
}

func checkExceptions(s string, exceptions []string) bool {
	for _, x := range exceptions {
		if strings.Contains(s, x) {
			debugOutput("String %q contains exception %q", s, x)
			return true
		}
	}
	return false
}

// nolint: gocyclo
func main() {
	configFile := flag.String("config", "", "Config File to use")
	var lastCheck time.Time

	chanError := make(chan error)
	chanOutput := make(chan paste)

	alredyChecked := make(map[string]time.Time)

	var wgOutput sync.WaitGroup
	var wgError sync.WaitGroup

	flag.Parse()

	log.Println("Starting Pastebin Scraper")
	config, err := getConfig(*configFile)
	if err != nil {
		log.Fatalf("could not read config file %s: %v", *configFile, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wgOutput.Add(1)
	wgError.Add(1)

	go func() {
		defer wgOutput.Done()
		for p := range chanOutput {
			debugOutput("found paste:\n%s", p)
			err = p.sendPasteMessage(config)
			if err != nil {
				chanError <- fmt.Errorf("sendPasteMessage: %v", err)
			}
		}
	}()

	go func() {
		defer wgError.Done()
		for err := range chanError {
			log.Printf("%v", err)
			if config.Mailonerror {
				err2 := sendErrorMessage(config, err)
				if err2 != nil {
					log.Printf("ERROR on sending error mail: %v", err2)
				}
			}
		}
	}()

	// use a boundary for keyword searching
	for _, k := range config.Keywords {
		r := fmt.Sprintf(`(?im)^(.*\b%s.*)$`, regexp.QuoteMeta(k.Keyword))
		keywordsRegex[k.Keyword] = keywordRegexType{
			regexp:     regexp.MustCompile(r),
			exceptions: k.Exceptions,
		}
	}

	for {
		// Only fetch the main list once a minute
		sleepTime := time.Until(lastCheck.Add(1 * time.Minute))
		if sleepTime > 0 {
			debugOutput("sleeping for %s", sleepTime)
			time.Sleep(sleepTime)
		}

		pastes, err := fetchPasteList(ctx)
		if err != nil {
			chanError <- fmt.Errorf("fetchPasteList: %v", err)
		}
		lastCheck = time.Now()

		for _, p := range pastes {
			if _, ok := alredyChecked[p.Key]; ok {
				debugOutput("skipping key %s as it was already checked", p.Key)
			} else {
				alredyChecked[p.Key] = time.Now()
				p2, err := p.fetch(ctx)
				if err != nil {
					chanError <- fmt.Errorf("fetch: %v", err)
				} else if p2 != nil {
					chanOutput <- *p2
				}
				// do not hammer the API
				time.Sleep(1 * time.Second)
			}
		}
		// clean up old items in alreadyChecked map
		// delete everything older than 10 minutes
		threshold := time.Now().Add(-10 * time.Minute)
		for k, v := range alredyChecked {
			if v.Before(threshold) {
				debugOutput("deleting expired entry %s", k)
				delete(alredyChecked, k)
			}
		}
	}

	close(chanOutput)
	wgOutput.Wait()
	close(chanError)
	wgError.Wait()
}
