package idgen

import (
	"strings"
	"testing"
)

func TestNewIDLength(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID() returned error: %v", err)
	}
	if len(id) != idLength {
		t.Fatalf("NewID() length = %d, want %d", len(id), idLength)
	}
}

func TestNewIDAlphabet(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID() returned error: %v", err)
	}
	for _, c := range id {
		if !strings.ContainsRune(alphabet, c) {
			t.Fatalf("NewID() contains invalid character %q", c)
		}
	}
}

func TestNewIDDistinct(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := NewID()
		if err != nil {
			t.Fatalf("NewID() returned error: %v", err)
		}
		if seen[id] {
			t.Fatalf("NewID() returned duplicate id %q", id)
		}
		seen[id] = true
	}
}
