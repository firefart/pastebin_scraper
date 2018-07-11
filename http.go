package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func httpRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	return resp, err
}

func httpRespBodyToString(resp *http.Response) (res string, err error) {
	if resp == nil {
		return "", fmt.Errorf("response is nil")
	}

	// catch errors when closing and return them
	defer func() {
		rerr := resp.Body.Close()
		if rerr != nil {
			err = rerr
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	res = string(body)
	return res, nil
}
