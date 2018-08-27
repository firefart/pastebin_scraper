package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	t := true
	test = &t
}

func httpServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	return ts
}

func TestHttpRequest(t *testing.T) {
	h := httpServer(t, "test")
	defer h.Close()
	_, err := httpRequest(context.Background(), h.URL)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
}

func TestHttpRespBodyToString(t *testing.T) {
	h := httpServer(t, "test")
	defer h.Close()
	r, err := httpRequest(context.Background(), h.URL)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	x, err := httpRespBodyToString(r)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if x != "test" {
		t.Fatalf("Content does not match. Got %q, expected %q", x, "test")
	}
}
