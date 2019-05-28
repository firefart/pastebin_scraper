package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	gomail "gopkg.in/gomail.v2"
)

const (
	apiEndpoint = "https://scrape.pastebin.com/api_scraping.php"
)

type paste struct {
	FullURL   string `json:"full_url"`
	ScrapeURL string `json:"scrape_url"`
	Date      string `json:"date"`
	Key       string `json:"key"`
	Size      string `json:"size"`
	Expire    string `json:"expire"`
	Title     string `json:"title"`
	Syntax    string `json:"syntax"`
	User      string `json:"user"`
	Content   string
	Matches   map[string][]string
}

func (p *paste) String() string {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	keywords := strings.Join(getKeysFromMap(p.Matches), ", ")
	if _, err := fmt.Fprintf(tw, "Pastebin Alert for Keywords %s\n\n", keywords); err != nil {
		return fmt.Sprintf("error on tostring: %v", err)
	}

	fields := []struct {
		prefix  string
		content string
	}{
		{"Title", p.Title},
		{"URL", p.FullURL},
		{"User", p.User},
		{"Date", dateToString(p.Date)},
		{"Size", p.Size},
		{"Expire", dateToString(p.Expire)},
		{"Syntax", p.Syntax},
	}

	for _, x := range fields {
		if x.content != "" {
			if _, err := fmt.Fprintf(tw, "%s:\t%s\n", x.prefix, x.content); err != nil {
				return fmt.Sprintf("error on tostring: %v", err)
			}
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Sprintf("error on tostring: %v", err)
	}

	for k, v := range p.Matches {
		if _, err := fmt.Fprintf(bw, "\nMatches for %s:\n", k); err != nil {
			return fmt.Sprintf("error on tostring: %v", err)
		}
		for _, m := range v {
			if _, err := fmt.Fprintf(bw, "%s\n", m); err != nil {
				return fmt.Sprintf("error on tostring: %v", err)
			}
		}
	}

	if err := bw.Flush(); err != nil {
		return fmt.Sprintf("error on tostring: %v", err)
	}
	return buffer.String()
}

func (p *paste) sendPasteMessage(config configuration) (err error) {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Mailfrom)
	m.SetHeader("To", config.Mailto)
	keywords := strings.Join(getKeysFromMap(p.Matches), ",")
	m.SetHeader("Subject", fmt.Sprintf("Pastebin Alert for %s", keywords))

	filename := fmt.Sprintf("%s.zip", randomString(10))
	fullPath := path.Join(os.TempDir(), filename)
	zipFile, err := createZip("content.txt", p.Content)
	if err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() {
		rerr := os.Remove(fullPath)
		if rerr != nil {
			err = rerr
		}
	}()

	_, err = f.Write(zipFile)
	if err != nil {
		return err
	}

	defer func() {
		rerr := f.Close()
		if rerr != nil {
			err = rerr
		}
	}()

	m.Attach(fullPath)

	m.SetBody("text/plain", p.String())
	err = sendEmail(config, m)
	return err
}

func (p paste) fetch(ctx context.Context, keywords *map[string]keywordType, cidrs *[]cidrType) (*paste, error) {
	debugOutput("checking paste %s", p.Key)
	resp, err := httpRequest(ctx, p.ScrapeURL)
	if err != nil {
		// Ignore HTTP based errors like timeout and connection reset
		return nil, nil
	}

	if resp.StatusCode == http.StatusOK || resp.ContentLength > 0 {
		b, err := httpRespBodyToString(resp)
		if err != nil {
			return nil, err
		}
		found, key := checkKeywords(b, keywords)
		found2, key2 := checkCIDRs(b, cidrs)
		if found || found2 {
			// merge key1 and key2
			for k, v := range key2 {
				key[k] = v
			}

			p.Content = b
			p.Matches = key
			return &p, nil
		}
	} else {
		b, err := httpRespBodyToString(resp)
		return nil, fmt.Errorf("Output: %s, Error: %v", b, err)
	}
	// nothing found
	return nil, nil
}

func fetchPasteList(ctx context.Context) ([]paste, error) {
	var list []paste
	debugOutput("fetching paste list")
	url := fmt.Sprintf("%s?limit=100", apiEndpoint)
	resp, err := httpRequest(ctx, url)
	if err != nil {
		// Ignore HTTP based errors like timeout and connection reset
		return list, nil
	}

	body, err := httpRespBodyToString(resp)
	if err != nil {
		return list, err
	}
	// ip does not have access. Do not panic so error mail will be sent
	if strings.Contains(body, "DOES NOT HAVE ACCESS") {
		return list, fmt.Errorf(body)
	}

	jsonErr := json.Unmarshal([]byte(body), &list)
	if jsonErr != nil {
		return list, fmt.Errorf("error on parsing json: %v. json: %s", jsonErr, body)
	}
	return list, nil
}
