package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/pastebin/handlers"
	"example.com/pastebin/store"
)

func TestHealthz(t *testing.T) {
	api := handlers.NewAPI(store.NewStore(), 1048576)
	mux := newMux(api)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET /healthz: expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("GET /healthz: expected Content-Type application/json, got %q", ct)
	}
}
