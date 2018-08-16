package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"
)

const (
	apiEndpoint = "https://scrape.pastebin.com/api_scraping.php"
)

type paste struct {
	FullURL         string `json:"full_url"`
	ScrapeURL       string `json:"scrape_url"`
	Date            string `json:"date"`
	Key             string `json:"key"`
	Size            string `json:"size"`
	Expire          string `json:"expire"`
	Title           string `json:"title"`
	Syntax          string `json:"syntax"`
	User            string `json:"user"`
	Content         string
	MatchedKeywords map[string]string
}

func (p *paste) String() string {
	keywords := strings.Join(getKeysFromMap(p.MatchedKeywords), ",")
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Pastebin Alert for Keywords %s\n\n", keywords)
	if p.Title != "" {
		fmt.Fprintf(buf, "Title: %s\n", p.Title)
	}
	if p.FullURL != "" {
		fmt.Fprintf(buf, "URL: %s\n", p.FullURL)
	}
	if p.User != "" {
		fmt.Fprintf(buf, "User: %s\n", p.User)
	}
	if p.Date != "" {
		fmt.Fprintf(buf, "Date: %s\n", p.Date)
	}
	if p.Size != "" {
		fmt.Fprintf(buf, "Size: %s\n", p.Size)
	}
	if p.Expire != "" {
		fmt.Fprintf(buf, "Expire: %s\n", p.Expire)
	}
	if p.Syntax != "" {
		fmt.Fprintf(buf, "Syntax: %s\n", p.Syntax)
	}

	for k, v := range p.MatchedKeywords {
		fmt.Fprintf(buf, "\nFirst match of Keyword: %s\n%s", k, v)
	}

	return buf.String()
}

func (p *paste) sendPasteMessage(config *configuration) (err error) {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Mailfrom)
	m.SetHeader("To", config.Mailto)
	keywords := strings.Join(getKeysFromMap(p.MatchedKeywords), ",")
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

func (p paste) fetch(ctx context.Context) error {
	debugOutput("checking paste %s", p.Key)
	alredyChecked[p.Key] = time.Now()
	resp, err := httpRequest(ctx, p.ScrapeURL)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.ContentLength > 0 {
		b, err := httpRespBodyToString(resp)
		if err != nil {
			return err
		}
		found, key := checkKeywords(b)
		if found {
			p.Content = b
			p.MatchedKeywords = key
			chanOutput <- p
		}
	} else {
		b, err := httpRespBodyToString(resp)
		return fmt.Errorf("Output: %s, Error: %v", b, err)
	}
	return nil
}

func fetchPasteList(ctx context.Context) ([]paste, error) {
	var list []paste
	debugOutput("fetching paste list")
	url := fmt.Sprintf("%s?limit=100", apiEndpoint)
	resp, err := httpRequest(ctx, url)
	if err != nil {
		return list, err
	}

	body, err := httpRespBodyToString(resp)
	if err != nil {
		return list, err
	}
	if strings.Contains(body, "DOES NOT HAVE ACCESS") {
		panic("You do not have access to the scrape API from this IP address!")
	}

	jsonErr := json.Unmarshal([]byte(body), &list)
	if jsonErr != nil {
		return list, fmt.Errorf("error on parsing json: %v. json: %s", jsonErr, body)
	}
	return list, nil
}
