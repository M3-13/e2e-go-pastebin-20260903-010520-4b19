package store

import (
	"sync"
	"testing"
	"time"
)

func TestCreateAndGet(t *testing.T) {
	s := NewStore()
	p := Paste{
		ID:        "abc",
		Content:   "hello",
		Language:  "go",
		CreatedAt: time.Now(),
	}

	if err := s.Create(p); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, ok := s.Get("abc")
	if !ok {
		t.Fatalf("Get returned ok=false for existing id")
	}
	if got.Content != "hello" || got.ID != "abc" || got.Language != "go" {
		t.Fatalf("Get returned wrong paste: %+v", got)
	}
}

func TestCreateDuplicateReturnsError(t *testing.T) {
	s := NewStore()
	p := Paste{ID: "dup", Content: "x", CreatedAt: time.Now()}

	if err := s.Create(p); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}
	if err := s.Create(p); err == nil {
		t.Fatalf("expected error on duplicate id, got nil")
	}
}

func TestGetUnknownID(t *testing.T) {
	s := NewStore()

	got, ok := s.Get("missing")
	if ok {
		t.Fatalf("Get returned ok=true for unknown id")
	}
	if got.ID != "" || got.Content != "" {
		t.Fatalf("Get returned non-zero paste for unknown id: %+v", got)
	}
}

func TestGetExpired(t *testing.T) {
	s := NewStore()
	past := time.Now().Add(-time.Hour)
	p := Paste{
		ID:        "expired",
		Content:   "secret",
		CreatedAt: time.Now(),
		ExpiresAt: &past,
	}
	if err := s.Create(p); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, ok := s.Get("expired")
	if ok {
		t.Fatalf("Get returned ok=true for expired paste: %+v", got)
	}

	// lazy removal: after a failed Get the entry must be gone
	if _, ok := s.Get("expired"); ok {
		t.Fatalf("expired paste still present after lazy removal")
	}
}

func TestListFiltersExpired(t *testing.T) {
	s := NewStore()
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	active := Paste{ID: "active", Content: "keep", CreatedAt: time.Now(), ExpiresAt: &future}
	expired := Paste{ID: "old", Content: "drop", CreatedAt: time.Now(), ExpiresAt: &past}
	noExpiry := Paste{ID: "forever", Content: "keep2", CreatedAt: time.Now()}

	for _, p := range []Paste{active, expired, noExpiry} {
		if err := s.Create(p); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	list := s.List()
	ids := map[string]bool{}
	for _, p := range list {
		ids[p.ID] = true
	}
	if !ids["active"] || !ids["forever"] {
		t.Fatalf("List missing active pastes: %+v", list)
	}
	if ids["old"] {
		t.Fatalf("List included expired paste: %+v", list)
	}
	if len(list) != 2 {
		t.Fatalf("List returned %d pastes, want 2: %+v", len(list), list)
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()
	p := Paste{ID: "del", Content: "x", CreatedAt: time.Now()}
	if err := s.Create(p); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if !s.Delete("del") {
		t.Fatalf("Delete returned false for existing paste")
	}
	if _, ok := s.Get("del"); ok {
		t.Fatalf("paste still present after Delete")
	}
	if s.Delete("del") {
		t.Fatalf("Delete returned true for already-removed paste")
	}
	if s.Delete("never-existed") {
		t.Fatalf("Delete returned true for unknown id")
	}
}

func TestExpiredRemovedOnAccessAllowsRecreate(t *testing.T) {
	s := NewStore()
	past := time.Now().Add(-time.Hour)
	p := Paste{ID: "reuse", Content: "old", CreatedAt: time.Now(), ExpiresAt: &past}
	if err := s.Create(p); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, ok := s.Get("reuse"); ok {
		t.Fatalf("Get returned ok=true for expired paste")
	}

	if err := s.Create(Paste{ID: "reuse", Content: "new", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Create after lazy removal returned error: %v", err)
	}

	got, ok := s.Get("reuse")
	if !ok || got.Content != "new" {
		t.Fatalf("recreated paste not retrievable: %+v ok=%v", got, ok)
	}
}

func TestDeleteExpired(t *testing.T) {
	s := NewStore()
	past := time.Now().Add(-time.Hour)
	p := Paste{ID: "exp", Content: "x", CreatedAt: time.Now(), ExpiresAt: &past}
	if err := s.Create(p); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if s.Delete("exp") {
		t.Fatalf("Delete returned true for expired paste")
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	s := NewStore()
	ids := []string{"a", "b", "c", "d"}

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			p := Paste{ID: id, Content: "content-" + id, CreatedAt: time.Now()}
			_ = s.Create(p)
		}(id)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.List()
			for _, id := range ids {
				_, _ = s.Get(id)
			}
		}()
	}

	wg.Wait()

	list := s.List()
	if len(list) != len(ids) {
		t.Fatalf("after concurrent writes, List returned %d pastes, want %d", len(list), len(ids))
	}
}
