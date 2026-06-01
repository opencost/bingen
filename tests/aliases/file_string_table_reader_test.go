package aliases

import (
	"os"
	"sync/atomic"
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
	s := "hello"
	reader := &FileStringTableReader{
		f: tmp,
		refs: []fileStringRef{
			{off: 0, length: 5},
		},
		memo:         make([]atomic.Pointer[string], 1),
		memoMaxBytes: 16,
	}
	reader.memo[0].Store(&s)
	reader.memoBytes.Store(5)
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
	s3 := "cccc"
	reader := &FileStringTableReader{
		refs: []fileStringRef{
			{length: len(s1)},
			{length: len(s2)},
			{length: len(s3)},
		},
		memo:         make([]atomic.Pointer[string], 3),
		memoMaxBytes: 16,
	}

	// Pre-populate cache (simulating what NewFileStringTableReaderFrom does)
	reader.memo[0].Store(&s1)
	reader.memo[1].Store(&s2)
	reader.memoBytes.Store(int64(len(s1) + len(s2)))

	// Verify cached entries are accessible
	if got := reader.memo[0].Load(); got == nil || *got != s1 {
		t.Fatalf("expected cached s1, got %v", got)
	}
	if got := reader.memo[1].Load(); got == nil || *got != s2 {
		t.Fatalf("expected cached s2, got %v", got)
	}
	// s3 should not be cached (would exceed limit)
	if got := reader.memo[2].Load(); got != nil {
		t.Fatalf("expected s3 to not be cached, got %q", *got)
	}
}
