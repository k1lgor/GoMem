package gomem

import (
	"fmt"
	"os"
	"testing"
)

func newStoreForTest(t *testing.T) *Store {
	t.Helper()
	dir, err := os.MkdirTemp("", "gomem-store-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	s, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStoreRememberAndSearch(t *testing.T) {
	s := newStoreForTest(t)

	if err := s.Remember("key1", "Go is a statically typed compiled language"); err != nil {
		t.Fatal(err)
	}
	if err := s.Remember("key2", "Rust is a systems programming language"); err != nil {
		t.Fatal(err)
	}

	hits, total, err := s.Search("Go language", 10)
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 {
		t.Fatal("expected search results")
	}

	found := false
	for _, hit := range hits {
		if hit.ID == "key1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected key1 in search results")
	}
}

func TestStoreDelete(t *testing.T) {
	s := newStoreForTest(t)

	if err := s.Remember("del-key", "delete me"); err != nil {
		t.Fatal(err)
	}

	// Verify it exists
	_, total, _ := s.Search("delete", 10)
	if total == 0 {
		t.Fatal("expected doc before delete")
	}

	// Delete
	if err := s.Delete("del-key"); err != nil {
		t.Fatal(err)
	}

	// Verify gone
	_, total, _ = s.Search("delete", 10)
	if total != 0 {
		t.Fatal("expected no results after delete")
	}
}

func TestStoreSearchLimit(t *testing.T) {
	s := newStoreForTest(t)

	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("limit-key-%d", i)
		if err := s.Remember(id, "search limit test document"); err != nil {
			t.Fatal(err)
		}
	}

	// Default limit
	hits, total, _ := s.Search("search limit", 0)
	if total < 20 {
		t.Fatalf("expected total >= 20, got %d", total)
	}
	if len(hits) != 10 {
		t.Fatalf("expected 10 hits with default limit, got %d", len(hits))
	}
}

func TestStoreDocCount(t *testing.T) {
	s := newStoreForTest(t)

	count, err := s.DocCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 docs, got %d", count)
	}

	s.Remember("c1", "doc one")
	s.Remember("c2", "doc two")

	count, _ = s.DocCount()
	if count != 2 {
		t.Fatalf("expected 2 docs, got %d", count)
	}
}

func TestStorePersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "gomem-persist-store-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	// First session
	s1, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.Remember("persist-me", "this should survive"); err != nil {
		t.Fatal(err)
	}
	s1.Close()

	// Second session
	s2, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	hits, total, err := s2.Search("survive", 10)
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 {
		t.Fatal("expected data to survive store close/reopen")
	}
	if hits[0].ID != "persist-me" {
		t.Fatalf("expected persist-me, got %q", hits[0].ID)
	}
}
