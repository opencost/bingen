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

	reader := &FileStringTableReader{
		f: tmp,
		refs: []fileStringRef{
			{off: 0, length: 5},
		},
		memo:         make([]atomic.Pointer[string], 1),
		memoMaxBytes: 16,
	}
	defer reader.Close()

	if got := reader.At(0); got != "hello" {
		t.Fatalf("baseline string mismatch, got %q", got)
	}

	if err := tmp.Truncate(0); err != nil {
		t.Fatalf("truncate temp file: %v", err)
	}

	if got := reader.At(0); got != "hello" {
		t.Fatalf("expected memoized value after truncate, got %q", got)
	}
}

func TestFileStringTableReader_EvictLeastUsedMemoEntries(t *testing.T) {
	s1 := "aaaa"
	s2 := "bbbb"
	s3 := "cccc"
	s4 := "dddd"
	reader := &FileStringTableReader{
		refs: []fileStringRef{
			{length: len(s1)},
			{length: len(s2)},
			{length: len(s3)},
			{length: len(s4)},
		},
		memo:         make([]atomic.Pointer[string], 4),
		memoHits:     make([]atomic.Uint64, 4),
		evictScratch: make([]memoEvictionCandidate, 4),
		memoMaxBytes: 16,
	}

	reader.memo[0].Store(&s1)
	reader.memo[1].Store(&s2)
	reader.memo[2].Store(&s3)
	reader.memo[3].Store(&s4)
	reader.memoHits[0].Store(10)
	reader.memoHits[1].Store(1)
	reader.memoHits[2].Store(3)
	reader.memoHits[3].Store(2)
	reader.memoBytes.Store(int64(len(s1) + len(s2) + len(s3) + len(s4)))

	reader.evictLeastUsedMemoEntries(0.10, 0.40)

	if got := reader.memo[1].Load(); got != nil {
		t.Fatalf("expected index 1 to be evicted first, got %q", *got)
	}
	if got := reader.memo[3].Load(); got == nil {
		t.Fatalf("expected index 3 to remain cached")
	}
}
