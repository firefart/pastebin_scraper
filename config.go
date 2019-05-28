package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type configuration struct {
	Mailserver  string    `json:"mailserver"`
	Mailport    int       `json:"mailport"`
	Mailfrom    string    `json:"mailfrom"`
	Mailonerror bool      `json:"mailonerror"`
	Mailtoerror string    `json:"mailtoerror"`
	Mailto      string    `json:"mailto"`
	Mailsubject string    `json:"mailsubject"`
	Timeout     string    `json:"timeout"`
	Keywords    []keyword `json:"keywords"`
	CIDRs       []string  `json:"cidrs"`
}

type keyword struct {
	Keyword    string   `json:"keyword"`
	Exceptions []string `json:"exceptions"`
}

func getConfig(f string) (*configuration, error) {
	if f == "" {
		return nil, fmt.Errorf("please provide a valid config file")
	}

	b, err := ioutil.ReadFile(f) // nolint: gosec
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)

	decoder := json.NewDecoder(reader)
	c := configuration{}
	if err = decoder.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
