package util

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

// TestRoundTripIntFullRange verifies that int values across the full int64 range
// survive a write/read pair. The previous implementation truncated to int32 and
// silently corrupted any value outside that range.
func TestRoundTripIntFullRange(t *testing.T) {
	values := []int{
		0,
		1,
		-1,
		math.MaxInt32,
		math.MinInt32,
		math.MaxInt32 + 1,
		math.MinInt32 - 1,
		math.MaxInt64,
		math.MinInt64,
	}

	for _, v := range values {
		w := NewBuffer()
		w.WriteInt(v)
		r := NewBufferFromBytes(w.Bytes())
		got := r.ReadInt()
		if got != v {
			t.Errorf("WriteInt/ReadInt round-trip: wrote %d, read %d", v, got)
		}
	}
}

// TestRoundTripUintFullRange verifies the corresponding behavior for uint.
func TestRoundTripUintFullRange(t *testing.T) {
	values := []uint{
		0,
		1,
		math.MaxUint32,
		math.MaxUint32 + 1,
		math.MaxUint64,
	}

	for _, v := range values {
		w := NewBuffer()
		w.WriteUInt(v)
		r := NewBufferFromBytes(w.Bytes())
		got := r.ReadUInt()
		if got != v {
			t.Errorf("WriteUInt/ReadUInt round-trip: wrote %d, read %d", v, got)
		}
	}
}

// TestRoundTripLargeString verifies that strings beyond the old uint16 limit
// survive a round-trip. Previously WriteString silently truncated at 64KiB.
func TestRoundTripLargeString(t *testing.T) {
	for _, n := range []int{0, 1, 32, math.MaxUint16, math.MaxUint16 + 1, 100_000} {
		s := strings.Repeat("a", n)
		w := NewBuffer()
		w.WriteString(s)
		r := NewBufferFromBytes(w.Bytes())
		got := r.ReadString()
		if len(got) != len(s) {
			t.Errorf("WriteString/ReadString len mismatch for n=%d: got %d", n, len(got))
		}
	}
}

// TestReadStringRejectsHugeLength verifies that a hand-crafted payload claiming
// a multi-gigabyte string returns an empty string instead of allocating.
func TestReadStringRejectsHugeLength(t *testing.T) {
	// Encode a uint32 string length of MaxInt32 with no trailing payload.
	var buf bytes.Buffer
	// little-endian uint32 = MaxInt32
	buf.WriteByte(0xFF)
	buf.WriteByte(0xFF)
	buf.WriteByte(0xFF)
	buf.WriteByte(0x7F)

	r := NewBufferFromBytes(buf.Bytes())
	got := r.ReadString()
	if got != "" {
		t.Errorf("ReadString with bogus huge length should return \"\", got %q", got)
	}
}

// TestReadBytesRejectsHugeLength verifies that asking for more bytes than
// remain returns nil rather than panicking or allocating.
func TestReadBytesRejectsHugeLength(t *testing.T) {
	r := NewBufferFromBytes([]byte{1, 2, 3})
	got := r.ReadBytes(math.MaxInt32)
	if got != nil {
		t.Errorf("ReadBytes(MaxInt32) on 3-byte buffer should return nil, got len=%d", len(got))
	}
}

// FuzzBufferRoundtrip ensures that no input causes the read paths on a
// reader-mode buffer to panic. The fuzz target only pulls a handful of values
// off the buffer; the contract is "no panics on any input."
func FuzzBufferRoundtrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0})
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	f.Fuzz(func(t *testing.T, data []byte) {
		buf := NewBufferFromBytes(append([]byte(nil), data...))
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()
		_ = buf.ReadInt()
		_ = buf.ReadString()
		_ = buf.ReadBool()
		_ = buf.ReadInt64()
		_ = buf.ReadUInt32()
	})
}
