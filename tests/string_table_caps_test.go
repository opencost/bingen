package test

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/opencost/bingen/tests/aliases"
)

// TestUnmarshalRejectsOversizedStringTable hand-crafts a payload whose string
// table header advertises a length above MaxStringTableEntries and verifies
// the unmarshal fails with a normal error rather than panicking the host.
//
// NewSliceStringTableReaderFrom previously panicked on attacker-controlled
// input; it now returns an error that propagates through
// NewDecodingContextFromBytes -> UnmarshalBinary.
func TestUnmarshalRejectsOversizedStringTable(t *testing.T) {
	const tag = "BGST"
	// 4-byte tag + 8-byte little-endian int64 length prefix exceeding the cap.
	var buf bytes.Buffer
	buf.WriteString(tag)
	var lenBytes [8]byte
	// aliases.MaxStringTableEntries == 1<<20; pick something above it.
	binary.LittleEndian.PutUint64(lenBytes[:], uint64(aliases.MaxStringTableEntries+1))
	buf.Write(lenBytes[:])

	got := &aliases.Parent{}
	err := got.UnmarshalBinary(buf.Bytes())
	if err == nil {
		t.Fatalf("UnmarshalBinary should have rejected oversized string table")
	}
	if !strings.Contains(err.Error(), "MaxStringTableEntries") {
		t.Errorf("error should reference MaxStringTableEntries, got: %s", err)
	}
}
