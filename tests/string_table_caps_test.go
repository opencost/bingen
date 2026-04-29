package test

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/opencost/bingen/tests/aliases"
	"github.com/opencost/bingen/tests/container"
)

// TestUnmarshalRejectsOversizedSlice exercises the unconditional
// MaxContainerLength cap by hand-crafting a payload that advertises an
// oversized slice in reader-mode (where Remaining() returns -1, so the
// remaining-bytes upper bound is not in play). Without the unconditional cap
// the generated unmarshaller would call make() with the attacker-supplied
// length and OOM before any read failed.
func TestUnmarshalRejectsOversizedSlice(t *testing.T) {
	// container.Container layout (v0.2): version(1) + Name(uint32 len) +
	// Children-non-nil-byte(1) + Children-len(int=8 bytes) + entries +
	// Value(float64). We craft the prefix up through the slice length and
	// stop — the cap check fires before any further reads.
	var buf bytes.Buffer
	buf.WriteByte(container.ContainerExampleCodecVersion) // version
	// Name: empty string (uint32 len = 0).
	var nameLen [4]byte
	binary.LittleEndian.PutUint32(nameLen[:], 0)
	buf.Write(nameLen[:])
	// Children non-nil indicator.
	buf.WriteByte(1)
	// Children: int64 length far above MaxContainerLength.
	var childLen [8]byte
	binary.LittleEndian.PutUint64(childLen[:], uint64(container.MaxContainerLength)+1)
	buf.Write(childLen[:])

	// Use a streaming reader to force reader-mode (Remaining() == -1).
	got := &container.Container{}
	err := got.UnmarshalBinaryFromReader(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatalf("UnmarshalBinaryFromReader should have rejected oversized slice")
	}
	if !strings.Contains(err.Error(), "MaxContainerLength") {
		t.Errorf("error should reference MaxContainerLength, got: %s", err)
	}
}

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
