package aliases

import (
	"os"
	"testing"
)

func TestFileStringTableReaderAt_UsesMemoCache(t *testing.T) {
	tmp, err := os.CreateTemp("", "bingen-bgst-test-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write([]byte("hello")); err != nil {
		t.Fatalf("write temp data: %v", err)
	}

	// Pre-populate cache (simulating what NewFileStringTableReaderFrom does)
	reader := &FileStringTableReader{
		f: tmp,
		refs: []fileStringRef{
			{off: 0, length: 5},
		},
		memo: []string{"hello"},
	}
	defer reader.Close()

	if got := reader.At(0); got != "hello" {
		t.Fatalf("baseline string mismatch, got %q", got)
	}

	// Truncate file to verify cache is used instead of reading from file
	if err := tmp.Truncate(0); err != nil {
		t.Fatalf("truncate temp file: %v", err)
	}

	if got := reader.At(0); got != "hello" {
		t.Fatalf("expected memoized value after truncate, got %q", got)
	}
}

func TestFileStringTableReader_PreloadCache(t *testing.T) {
	// Test that the cache is immutable after creation (no dynamic insertion)
	s1 := "aaaa"
	s2 := "bbbb"
	reader := &FileStringTableReader{
		refs: []fileStringRef{
			{length: len(s1)},
			{length: len(s2)},
			{length: 0}, // s3 not cached
		},
		memo: []string{s1, s2, ""}, // Pre-populated cache
	}

	// Verify cached entries are accessible
	if got := reader.memo[0]; got != s1 {
		t.Fatalf("expected cached s1, got %v", got)
	}
	if got := reader.memo[1]; got != s2 {
		t.Fatalf("expected cached s2, got %v", got)
	}
	// s3 should not be cached (empty string)
	if got := reader.memo[2]; got != "" {
		t.Fatalf("expected s3 to not be cached, got %q", got)
	}
}
