package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	c, err := getConfig(path.Join("testdata", "test.json"))
	if err != nil {
		t.Fatalf("got error when reading config file: %v", err)
	}
	if c == nil {
		t.Fatal("got a nil config object")
	}
	assert.NotEmpty(t, c.Mailserver)
	assert.NotZero(t, c.Mailport)
	assert.NotEmpty(t, c.MailUsername)
	assert.NotEmpty(t, c.MailPassword)
	assert.True(t, c.MailSkipTLS)
	assert.NotEmpty(t, c.Mailfrom)
	assert.NotEmpty(t, c.Mailto)
	assert.True(t, c.Mailonerror)
	assert.NotEmpty(t, c.Mailtoerror)
	assert.Len(t, c.Keywords, 3)
	assert.NotEmpty(t, c.Keywords[0].Keyword)
	assert.Len(t, c.Keywords[0].Exceptions, 3)
	assert.NotEmpty(t, c.Keywords[0].Exceptions[0])
	assert.Len(t, c.CIDRs, 2)
	assert.NotEmpty(t, c.CIDRs[0])
}

func TestGetConfigErrors(t *testing.T) {
	_, err := getConfig("")
	if err == nil {
		t.Fatal("expected error on empty filename")
	}
	_, err = getConfig("this_does_not_exist")
	if err == nil {
		t.Fatal("expected error on invalid file")
	}
}

func TestGetConfigInvalid(t *testing.T) {
	_, err := getConfig(path.Join("testdata", "invalid.json"))
	if err == nil {
		t.Fatal("expected error when reading config file but got none")
	}
}
