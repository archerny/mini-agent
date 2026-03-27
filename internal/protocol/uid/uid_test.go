package uid

import (
	"strings"
	"testing"
)

func TestNew_Format(t *testing.T) {
	id := New()

	// UUID format: 8-4-4-4-12
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 parts, got %d: %s", len(parts), id)
	}
	if len(parts[0]) != 8 {
		t.Errorf("part 0 should be 8 chars, got %d", len(parts[0]))
	}
	if len(parts[1]) != 4 {
		t.Errorf("part 1 should be 4 chars, got %d", len(parts[1]))
	}
	if len(parts[2]) != 4 {
		t.Errorf("part 2 should be 4 chars, got %d", len(parts[2]))
	}
	if len(parts[3]) != 4 {
		t.Errorf("part 3 should be 4 chars, got %d", len(parts[3]))
	}
	if len(parts[4]) != 12 {
		t.Errorf("part 4 should be 12 chars, got %d", len(parts[4]))
	}

	// Version 7: third group starts with '7'
	if parts[2][0] != '7' {
		t.Errorf("expected version 7 (third group starts with '7'), got '%c'", parts[2][0])
	}

	// Variant: fourth group first char should be 8, 9, a, or b
	c := parts[3][0]
	if c != '8' && c != '9' && c != 'a' && c != 'b' {
		t.Errorf("expected variant bits (8/9/a/b), got '%c'", c)
	}
}

func TestNew_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 10000)
	for i := range 10000 {
		id := New()
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID at iteration %d: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestNew_TimeOrdered(t *testing.T) {
	id1 := New()
	id2 := New()
	// UUID v7 with same millisecond uses counter, so id2 >= id1 lexicographically
	if id2 < id1 {
		t.Errorf("expected id2 >= id1 (time ordering), got id1=%s id2=%s", id1, id2)
	}
}
