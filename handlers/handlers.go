package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"example.com/pastebin/idgen"
	"example.com/pastebin/store"
)

type API struct {
	Store        *store.Store
	MaxBodyBytes int64
}

func NewAPI(s *store.Store, maxBodyBytes int64) *API {
	return &API{Store: s, MaxBodyBytes: maxBodyBytes}
}

type createPasteRequest struct {
	Content          string `json:"content"`
	Language         string `json:"language"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
}

type pasteMeta struct {
	ID        string     `json:"id"`
	Language  string     `json:"language,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *API) CreatePaste(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.MaxBodyBytes)

	var req createPasteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusUnprocessableEntity, "content is required and must not be empty")
		return
	}
	if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "expires_in_seconds must be greater than 0")
		return
	}

	id, err := idgen.NewID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate id")
		return
	}

	paste := store.Paste{
		ID:        id,
		Content:   req.Content,
		Language:  req.Language,
		CreatedAt: time.Now(),
	}
	if req.ExpiresInSeconds != nil {
		exp := time.Now().Add(time.Duration(*req.ExpiresInSeconds) * time.Second)
		paste.ExpiresAt = &exp
	}

	if err := a.Store.Create(paste); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store paste")
		return
	}

	writeJSON(w, http.StatusCreated, paste)
}

func (a *API) GetPaste(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, ok := a.Store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "paste not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (a *API) ListPastes(w http.ResponseWriter, r *http.Request) {
	pastes := a.Store.List()
	meta := make([]pasteMeta, 0, len(pastes))
	for _, p := range pastes {
		meta = append(meta, pasteMeta{
			ID:        p.ID,
			Language:  p.Language,
			CreatedAt: p.CreatedAt,
			ExpiresAt: p.ExpiresAt,
		})
	}
	writeJSON(w, http.StatusOK, meta)
}

func (a *API) DeletePaste(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !a.Store.Delete(id) {
		writeError(w, http.StatusNotFound, "paste not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
