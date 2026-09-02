package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/pastebin/store"
)

func newMux(api *API) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pastes", api.CreatePaste)
	mux.HandleFunc("GET /pastes/{id}", api.GetPaste)
	mux.HandleFunc("GET /pastes", api.ListPastes)
	mux.HandleFunc("DELETE /pastes/{id}", api.DeletePaste)
	return mux
}

func newTestMux(maxBody int64) (*http.ServeMux, *store.Store) {
	s := store.NewStore()
	api := NewAPI(s, maxBody)
	return newMux(api), s
}

func doRequest(t *testing.T, mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func assertContentType(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func assertErrorBody(t *testing.T, rr *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	if rr.Code != wantStatus {
		t.Fatalf("status = %d, want %d (body: %s)", rr.Code, wantStatus, rr.Body.String())
	}
	assertContentType(t, rr)
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("error body is not valid JSON: %v", err)
	}
	if len(body) != 1 {
		t.Errorf("error body must have exactly 1 field, got %d: %v", len(body), body)
	}
	msg, ok := body["error"].(string)
	if !ok || msg == "" {
		t.Errorf("error field must be a non-empty string, got %v", body["error"])
	}
}

func isURLSafeID(s string) bool {
	if len(s) != 22 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '-' || c == '_':
		default:
			return false
		}
	}
	return true
}

func TestCreatePaste(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"hello","language":"go"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", rr.Code, rr.Body.String())
	}
	assertContentType(t, rr)
	var p store.Paste
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !isURLSafeID(p.ID) {
		t.Errorf("id %q is not a 22-char URL-safe id", p.ID)
	}
	if p.Content != "hello" {
		t.Errorf("content = %q, want hello", p.Content)
	}
	if p.Language != "go" {
		t.Errorf("language = %q, want go", p.Language)
	}
	if p.CreatedAt.IsZero() {
		t.Error("created_at must not be zero")
	}
	if p.ExpiresAt != nil {
		t.Errorf("expires_at must be nil when not requested, got %v", p.ExpiresAt)
	}
}

func TestCreatePaste_Validation(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		status int
	}{
		{"missing content", `{"language":"go"}`, http.StatusUnprocessableEntity},
		{"empty content", `{"content":""}`, http.StatusUnprocessableEntity},
		{"invalid json", `{not json`, http.StatusBadRequest},
		{"expires_in_seconds zero", `{"content":"x","expires_in_seconds":0}`, http.StatusUnprocessableEntity},
		{"expires_in_seconds negative", `{"content":"x","expires_in_seconds":-5}`, http.StatusUnprocessableEntity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux, _ := newTestMux(1 << 20)
			rr := doRequest(t, mux, http.MethodPost, "/pastes", tc.body)
			assertErrorBody(t, rr, tc.status)
		})
	}
}

func TestCreatePaste_BodyTooLarge(t *testing.T) {
	mux, _ := newTestMux(10)
	rr := doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"this body is far larger than the ten byte limit"}`)
	assertErrorBody(t, rr, http.StatusRequestEntityTooLarge)
}

func TestCreatePaste_WithExpiry(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"x","expires_in_seconds":3600}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", rr.Code, rr.Body.String())
	}
	var p store.Paste
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.ExpiresAt == nil {
		t.Fatal("expires_at must be set when expires_in_seconds is given")
	}
	if !p.ExpiresAt.After(time.Now()) {
		t.Errorf("expires_at must be in the future, got %v", p.ExpiresAt)
	}
}

func TestGetPaste(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"hello","language":"txt"}`)
	var created store.Paste
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	rr = doRequest(t, mux, http.MethodGet, "/pastes/"+created.ID, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}
	assertContentType(t, rr)
	var got store.Paste
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("id = %q, want %q", got.ID, created.ID)
	}
	if got.Content != "hello" {
		t.Errorf("content = %q, want hello", got.Content)
	}
	if got.Language != "txt" {
		t.Errorf("language = %q, want txt", got.Language)
	}
}

func TestGetPaste_NotFound(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodGet, "/pastes/nonexistent", "")
	assertErrorBody(t, rr, http.StatusNotFound)
}

func TestListPastes_MetadataOnly(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"secret-one","language":"go"}`)
	doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"secret-two"}`)

	rr := doRequest(t, mux, http.MethodGet, "/pastes", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}
	assertContentType(t, rr)
	var list []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	for _, item := range list {
		if _, ok := item["content"]; ok {
			t.Errorf("list item must not contain content: %v", item)
		}
		if _, ok := item["id"]; !ok {
			t.Errorf("list item missing id: %v", item)
		}
		if _, ok := item["created_at"]; !ok {
			t.Errorf("list item missing created_at: %v", item)
		}
	}
}

func TestListPastes_Empty(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodGet, "/pastes", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "[]" {
		t.Errorf("empty list must be [], got %q", body)
	}
}

func TestDeletePaste(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"x"}`)
	var created store.Paste
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	rr = doRequest(t, mux, http.MethodDelete, "/pastes/"+created.ID, "")
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (body: %s)", rr.Code, rr.Body.String())
	}
	if rr.Body.Len() != 0 {
		t.Errorf("204 response must have empty body, got %q", rr.Body.String())
	}

	rr = doRequest(t, mux, http.MethodGet, "/pastes/"+created.ID, "")
	assertErrorBody(t, rr, http.StatusNotFound)
}

func TestDeletePaste_NotFound(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodDelete, "/pastes/nonexistent", "")
	assertErrorBody(t, rr, http.StatusNotFound)
}

func TestExpiredPaste(t *testing.T) {
	mux, _ := newTestMux(1 << 20)
	rr := doRequest(t, mux, http.MethodPost, "/pastes", `{"content":"x","expires_in_seconds":1}`)
	var created store.Paste
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ExpiresAt == nil {
		t.Fatal("expires_at must be set")
	}

	time.Sleep(1500 * time.Millisecond)

	rr = doRequest(t, mux, http.MethodGet, "/pastes/"+created.ID, "")
	assertErrorBody(t, rr, http.StatusNotFound)

	rr = doRequest(t, mux, http.MethodDelete, "/pastes/"+created.ID, "")
	assertErrorBody(t, rr, http.StatusNotFound)

	rr = doRequest(t, mux, http.MethodGet, "/pastes", "")
	var list []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, item := range list {
		if item["id"] == created.ID {
			t.Errorf("expired paste must not appear in list")
		}
	}
}
