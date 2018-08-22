package main

import (
	"path"
	"testing"
)

func TestGetConfig(t *testing.T) {
	c, err := getConfig(path.Join("testdata", "test.json"))
	if err != nil {
		t.Fatalf("got error when reading config file: %v", err)
	}
	if c == nil {
		t.Fatal("got a nil config object")
	}
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
