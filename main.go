package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	debug = flag.Bool("debug", false, "Print debug output")
	test  = flag.Bool("test", false, "do not send mails, print them instead")

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	// capture IPs (only v4)
	// https://www.regular-expressions.info/ip.html
	regexIP = regexp.MustCompile(`(\b(?:\d{1,3}\.){3}\d{1,3}\b)`)
)

type keywordType struct {
	regexp     *regexp.Regexp
	exceptions []string
}

type cidrType struct {
	ipNet *net.IPNet
}

func debugOutput(s string, a ...interface{}) {
	if *debug {
		log.Printf("[DEBUG] %s", fmt.Sprintf(s, a...))
	}
}

func checkKeywords(body string, keywords *map[string]keywordType) (bool, map[string][]string) {
	found := make(map[string][]string)
	status := false
	for k, v := range *keywords {
		var x []string
		s := v.regexp.FindAllString(body, -1)
		// we have a match
		if len(s) > 0 {
			// check for exceptions
			for _, m := range s {
				match := strings.TrimSpace(m)
				if !checkExceptions(match, v.exceptions) {
					x = append(x, match)
					status = true
				}
			}
		}
		// append result if not in exceptions
		if len(x) > 0 {
			found[k] = x
		}
	}
	return status, found
}

func checkCIDRs(body string, cidrs *[]cidrType) (bool, map[string][]string) {
	found := make(map[string][]string)
	status := false
	for _, cidr := range *cidrs {
		var x []string
		s := regexIP.FindAllString(body, -1)
		// we have a match
		if len(s) > 0 {
			for _, m := range s {
				match := strings.TrimSpace(m)
				ip := net.ParseIP(match)
				// invalid IP matched
				if ip == nil {
					debugOutput("%q is not a valid ip", match)
					continue
				}
				if cidr.ipNet.Contains(ip) {
					debugOutput("%v contains %s", cidr.ipNet, ip)
					x = append(x, match)
					status = true
				}
			}
		}
		if len(x) > 0 {
			found[cidr.ipNet.String()] = x
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

func parseKeywords(k []keyword) *map[string]keywordType {
	keywords := make(map[string]keywordType)
	for _, k := range k {
		// use a boundary for keyword searching
		r := fmt.Sprintf(`(?im)^(.*\b%s.*)$`, regexp.QuoteMeta(k.Keyword))
		keywords[k.Keyword] = keywordType{
			regexp:     regexp.MustCompile(r),
			exceptions: k.Exceptions,
		}
	}
	return &keywords
}

func parseCIDRs(cidrs []string) (*[]cidrType, error) {
	var ret []cidrType
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return nil, fmt.Errorf("could not parse cidr %s: %v", c, err)
		}
		ret = append(ret, cidrType{ipNet: n})
	}
	return &ret, nil
}

// nolint: gocyclo
func main() {
	configFile := flag.String("config", "", "Config File to use")
	var lastCheck time.Time

	chanError := make(chan error)
	chanOutput := make(chan paste)
	// we run in an endless loop so no need for a waitgroup here
	defer close(chanOutput)
	defer close(chanError)

	alredyChecked := make(map[string]time.Time)

	flag.Parse()

	log.Println("Starting Pastebin Scraper")
	config, err := getConfig(*configFile)
	if err != nil {
		log.Fatalf("could not read config file %s: %v", *configFile, err)
	}

	keywords := parseKeywords(config.Keywords)
	cidrs, err := parseCIDRs(config.CIDRs)
	if err != nil {
		log.Fatalf("could not parse cidrs: %v", err)
	}
	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		log.Fatalf("invalid value for timeout: %q - %v", config.Timeout, err)
	}
	client.Timeout = timeout

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(c configuration) {
		for p := range chanOutput {
			debugOutput("found paste:\n%+v", p)
			err = p.sendPasteMessage(c)
			if err != nil {
				chanError <- fmt.Errorf("sendPasteMessage: %v", err)
			}
		}
	}(*config)

	go func(c configuration) {
		for err := range chanError {
			log.Printf("%v", err)
			if c.Mailonerror {
				err2 := sendErrorMessage(c, err)
				if err2 != nil {
					log.Printf("ERROR on sending error mail: %v", err2)
				}
			}
		}
	}(*config)

	for {
		// Only fetch the main list once a minute
		sleepTime := time.Until(lastCheck.Add(1 * time.Minute))
		if sleepTime > 0 {
			debugOutput("sleeping for %s", sleepTime)
			time.Sleep(sleepTime)
		}

		lastCheck = time.Now()
		pastes, err := fetchPasteList(ctx)
		if err != nil {
			chanError <- fmt.Errorf("fetchPasteList: %v", err)
			continue
		}

		for _, p := range pastes {
			if _, ok := alredyChecked[p.Key]; ok {
				debugOutput("skipping key %s as it was already checked", p.Key)
			} else {
				alredyChecked[p.Key] = time.Now()
				p2, err := p.fetch(ctx, keywords, cidrs)
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
}
