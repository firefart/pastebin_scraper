package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	debug      = flag.Bool("debug", false, "Print debug output")
	configFile = flag.String("config", "", "Config File to use")

	config configuration

	lastCheck     time.Time
	terminate     = false
	alredyChecked = make(map[string]time.Time)

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	chanError  = make(chan error)
	chanOutput = make(chan paste)
	chanSignal = make(chan os.Signal)

	wgOutput sync.WaitGroup
	wgError  sync.WaitGroup

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	keywordsRegex = make(map[string]*regexp.Regexp)
	ctx           context.Context
)

func debugOutput(s string) {
	if *debug {
		log.Printf("[DEBUG] %s", s)
	}
}

func checkKeywords(body string) (bool, map[string]string) {
	found := make(map[string]string)
	status := false
	for k, v := range keywordsRegex {
		if v != nil {
			if s := v.FindStringSubmatch(body); s != nil {
				found[k] = strings.TrimSpace(s[1])
				status = true
			}
		}
	}
	return status, found
}

// nolint: gocyclo
func main() {
	flag.Parse()

	if *configFile == "" {
		log.Fatalln("Please provide a valid config file")
	}

	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}

	defer func() {
		rerr := file.Close()
		if rerr != nil {
			log.Fatalf("Error closing config file: %v", rerr)
		}
	}()

	decoder := json.NewDecoder(file)
	config = configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	log.Println("Starting Pastebin Scraper")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Clean exit when pressing CTRL+C
	signal.Notify(chanSignal, os.Interrupt)
	go func() {
		for range chanSignal {
			fmt.Println("CTRL+C pressed")
			terminate = true
		}
		cancel()
	}()

	wgOutput.Add(1)
	wgError.Add(1)

	go func() {
		defer wgOutput.Done()
		for p := range chanOutput {
			debugOutput(fmt.Sprintf("found paste:\n%s", p))
			err = p.sendPasteMessage()
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
				err2 := sendErrorMessage(err)
				if err2 != nil {
					log.Printf("ERROR on sending error mail: %v", err2)
				}
			}
		}
	}()

	// use a boundary for keyword searching
	for _, k := range config.Keywords {
		r := fmt.Sprintf(`(?im)^(.*\b%s.*)$`, regexp.QuoteMeta(k))
		keywordsRegex[k] = regexp.MustCompile(r)
	}

	for {
		if terminate {
			break
		}

		// Only fetch the main list once a minute
		sleepTime := time.Until(lastCheck.Add(1 * time.Minute))
		if sleepTime > 0 {
			debugOutput(fmt.Sprintf("sleeping for %s", sleepTime))
			time.Sleep(sleepTime)
		}

		pastes, err := fetchPasteList(ctx)
		if err != nil {
			chanError <- fmt.Errorf("fetchPasteList: %v", err)
		}

		for _, p := range pastes {
			if terminate {
				break
			}
			if _, ok := alredyChecked[p.Key]; ok {
				debugOutput(fmt.Sprintf("skipping key %s as it was already checked", p.Key))
			} else {
				err := p.fetch(ctx)
				if err != nil {
					chanError <- fmt.Errorf("fetch: %v", err)
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
				debugOutput(fmt.Sprintf("deleting expired entry %s", k))
				delete(alredyChecked, k)
			}
		}
	}

	close(chanOutput)
	wgOutput.Wait()
	close(chanError)
	wgError.Wait()
}
