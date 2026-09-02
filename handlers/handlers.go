package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/pastebin/store"
)

type API struct {
	Store        *store.Store
	MaxBodyBytes int64
}

func NewAPI(s *store.Store, maxBodyBytes int64) *API {
	return &API{Store: s, MaxBodyBytes: maxBodyBytes}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (a *API) CreatePaste(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (a *API) GetPaste(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (a *API) ListPastes(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (a *API) DeletePaste(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
