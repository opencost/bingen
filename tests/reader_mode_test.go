package test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/opencost/bingen/tests/container"
)

// TestUnmarshalFromReaderLargeSlice round-trips a slice longer than the
// bufio.Reader default buffer (4 KiB) through the reader-mode unmarshal path.
//
// An earlier draft of the length-prefix sanity check used Buffer.Remaining()
// for both byte-buffer and reader-mode buffers. In reader-mode that returned
// only the bufio.Reader's currently-buffered byte count, which falsely
// rejected legitimate large slices/maps/string-tables. Now Remaining() returns
// -1 in reader-mode and the upper-bound check is skipped.
func TestUnmarshalFromReaderLargeSlice(t *testing.T) {
	const n = 8192 // larger than bufio.Reader default 4096-byte buffer

	children := make([]string, n)
	for i := range children {
		children[i] = "child-" + strconv.Itoa(i)
	}

	c := &container.Container{
		Name:     "big",
		Children: children,
		Value:    1.5,
	}

	data, err := c.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %s", err)
	}

	got := &container.Container{}
	if err := got.UnmarshalBinaryFromReader(bytes.NewReader(data)); err != nil {
		t.Fatalf("UnmarshalBinaryFromReader: %s", err)
	}

	if got.Name != c.Name || got.Value != c.Value {
		t.Errorf("scalar fields mismatch: got %+v want %+v", got, c)
	}
	if len(got.Children) != len(c.Children) {
		t.Fatalf("child count: got %d want %d", len(got.Children), len(c.Children))
	}
	for i := range c.Children {
		if got.Children[i] != c.Children[i] {
			t.Fatalf("Children[%d]: got %q want %q", i, got.Children[i], c.Children[i])
		}
	}
}
