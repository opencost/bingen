package test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGeneratorDeterministic regenerates the codecs for a representative test
// package twice and asserts the bytes are identical. Map iteration ordering
// previously made the output non-deterministic across runs.
func TestGeneratorDeterministic(t *testing.T) {
	td := getTestDir()
	dir := filepath.Join(td, "container")
	codecPath := filepath.Join(dir, "container_codecs.go")

	if err := runGenerator(dir, "container"); err != nil {
		t.Fatalf("first generator run failed: %s", err)
	}
	first, err := os.ReadFile(codecPath)
	if err != nil {
		t.Fatalf("read first codec: %s", err)
	}

	if err := runGenerator(dir, "container"); err != nil {
		t.Fatalf("second generator run failed: %s", err)
	}
	second, err := os.ReadFile(codecPath)
	if err != nil {
		t.Fatalf("read second codec: %s", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("generated codec output is not deterministic: byte lengths %d vs %d", len(first), len(second))
	}
}
