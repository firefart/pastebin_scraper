package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	userAgent = "Pastebin Scraper (https://firefart.at)"
)

var (
	client = &http.Client{
		// default timeout
		Timeout: 10 * time.Second,
	}
)

func httpRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	return resp, err
}

func httpRespBodyToString(resp *http.Response) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("response is nil")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	res := string(body)
	return res, nil
}
