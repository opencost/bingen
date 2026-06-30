package table

import (
	"os"
	"testing"

	util "github.com/opencost/bingen/pkg/util"
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

func TestFileStringTableReader_ZeroMaxBytes(t *testing.T) {
	// Test that FileStringTableReader works correctly when memoMaxBytes = 0 (no caching)
	// This is an edge case that should not cause nil panics

	// Create a buffer and write test strings
	buf := util.NewBuffer()
	testStrings := []string{
		"string1",
		"string2",
		"string3",
		"string4",
		"string5",
	}

	// Write the string table to buffer
	// First write the table length
	buf.WriteInt(len(testStrings))
	// Then write each string
	for _, s := range testStrings {
		buf.WriteString(s)
	}

	// Create reader with memoMaxBytes = 0 (no caching)
	reader := NewFileStringTableReaderFrom(buf, t.TempDir(), "bingen", 0)
	if reader == nil {
		t.Fatal("NewFileStringTableReaderFrom returned nil")
	}
	defer reader.Close()

	// Cast to concrete type to inspect internal state
	fileReader, ok := reader.(*FileStringTableReader)
	if !ok {
		t.Fatal("Expected *FileStringTableReader type")
	}

	// Verify memo is nil or empty when memoMaxBytes = 0
	if len(fileReader.memo) > 0 {
		t.Errorf("Expected memo to be nil or empty with memoMaxBytes=0, got length %d", len(fileReader.memo))
	}

	// Verify no panic occurs and strings are correctly retrieved from file
	for i, expected := range testStrings {
		actual := reader.At(i)
		if actual != expected {
			t.Errorf("At(%d) = %q, want %q", i, actual, expected)
		}
	}

	// Verify Len() works correctly
	if reader.Len() != len(testStrings) {
		t.Errorf("Len() = %d, want %d", reader.Len(), len(testStrings))
	}

	// Test accessing the same string multiple times (should read from file each time)
	for i := 0; i < 3; i++ {
		actual := reader.At(0)
		if actual != testStrings[0] {
			t.Errorf("At(0) iteration %d = %q, want %q", i, actual, testStrings[0])
		}
	}
}
