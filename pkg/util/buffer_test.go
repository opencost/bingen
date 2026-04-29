package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
)

// TestRemainingReaderModeIsUnknown documents the contract that Remaining() in
// reader-mode returns -1 ("unknown"). Generated codecs use this sentinel so
// that length-prefix sanity checks don't false-reject legitimate large
// payloads on an io.Reader where the bufio.Reader's incidental buffer size
// would otherwise look like the entire payload.
func TestRemainingReaderModeIsUnknown(t *testing.T) {
	r := NewBufferFromReader(io.NopCloser(bytes.NewReader([]byte{1, 2, 3})))
	if got := r.Remaining(); got != -1 {
		t.Errorf("reader-mode Remaining() = %d, want -1", got)
	}
}

// TestRemainingByteBufferModeIsExact verifies that byte-buffer mode keeps the
// existing semantics: Remaining returns the unread byte count.
func TestRemainingByteBufferModeIsExact(t *testing.T) {
	b := NewBufferFromBytes([]byte{1, 2, 3, 4, 5})
	if got := b.Remaining(); got != 5 {
		t.Errorf("byte-buffer Remaining() = %d, want 5", got)
	}
	b.ReadUInt8()
	if got := b.Remaining(); got != 4 {
		t.Errorf("byte-buffer Remaining() after 1 read = %d, want 4", got)
	}
}

// TestRoundTripIntFullRange verifies that int values across the full int64 range
// survive a write/read pair. The previous implementation truncated to int32 and
// silently corrupted any value outside that range.
//
// The over-int32 values are gated on strconv.IntSize so the test compiles on
// 32-bit platforms (where math.MaxInt64 isn't representable as an int constant).
// The values themselves are constructed via int64-typed locals to avoid
// untyped-constant overflow at compile time.
func TestRoundTripIntFullRange(t *testing.T) {
	values := []int{
		0,
		1,
		-1,
		math.MaxInt32,
		math.MinInt32,
	}
	if strconv.IntSize == 64 {
		var v int64
		v = math.MaxInt32 + 1
		values = append(values, int(v))
		v = math.MinInt32 - 1
		values = append(values, int(v))
		v = math.MaxInt64
		values = append(values, int(v))
		v = math.MinInt64
		values = append(values, int(v))
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

// TestRoundTripUintFullRange verifies the corresponding behavior for uint. The
// over-uint32 values are gated on strconv.IntSize for 32-bit compile compat.
func TestRoundTripUintFullRange(t *testing.T) {
	values := []uint{
		0,
		1,
		math.MaxUint32,
	}
	if strconv.IntSize == 64 {
		var v uint64
		v = math.MaxUint32 + 1
		values = append(values, uint(v))
		v = math.MaxUint64
		values = append(values, uint(v))
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
// a multi-gigabyte string returns an empty string instead of allocating, and
// records the failure on Buffer.Err() so callers can detect the rejection.
func TestReadStringRejectsHugeLength(t *testing.T) {
	// Encode a uint32 string length of MaxInt32 with no trailing payload.
	var buf bytes.Buffer
	buf.WriteByte(0xFF)
	buf.WriteByte(0xFF)
	buf.WriteByte(0xFF)
	buf.WriteByte(0x7F)

	r := NewBufferFromBytes(buf.Bytes())
	got := r.ReadString()
	if got != "" {
		t.Errorf("ReadString with bogus huge length should return \"\", got %q", got)
	}
	if !errors.Is(r.Err(), ErrStringTooLarge) {
		t.Errorf("Err() = %v, want ErrStringTooLarge", r.Err())
	}
}

// TestReadStringRejectsHugeLengthReaderMode verifies the same cap is applied
// in reader-mode, where Remaining() can't bound the length.
func TestReadStringRejectsHugeLengthReaderMode(t *testing.T) {
	// uint32 length = MaxStringLength + 1, just over the cap. We encode the
	// length as little-endian bytes and feed them through an io.Reader.
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(MaxStringLength+1))

	r := NewBufferFromReader(bytes.NewReader(hdr[:]))
	got := r.ReadString()
	if got != "" {
		t.Errorf("ReadString reader-mode with oversize length should return \"\", got len=%d", len(got))
	}
	if !errors.Is(r.Err(), ErrStringTooLarge) {
		t.Errorf("Err() = %v, want ErrStringTooLarge", r.Err())
	}
}

// TestReadIntOverflowSetsErr verifies that decoding an int that doesn't fit
// the host int width records the helper's overflow error on Buffer.Err().
// On a 64-bit host the helpers can't overflow (every int64 fits in int), so
// this test only runs on 32-bit. We synthesize the test by writing an int64
// and asserting the reader-mode int read sets b.err.
func TestReadIntOverflowSetsErr(t *testing.T) {
	if strconv.IntSize >= 64 {
		t.Skip("int is 64-bit on this host; readInt cannot overflow")
	}

	w := NewBuffer()
	w.WriteInt64(math.MaxInt64) // 8 bytes — wire format for int is also 8 bytes
	r := NewBufferFromBytes(w.Bytes())
	_ = r.ReadInt()
	if r.Err() == nil {
		t.Errorf("ReadInt with int64 value above host int range should set Err()")
	}
}

// TestReadStringTruncatedByteBuffer verifies that a payload whose declared
// length exceeds the unread bytes (in byte-buffer mode) is rejected by
// returning "" AND records io.ErrUnexpectedEOF on Err(), so the caller can
// distinguish corruption from a legitimate empty string.
func TestReadStringTruncatedByteBuffer(t *testing.T) {
	// uint32 length of 100 with no following payload.
	var buf bytes.Buffer
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], 100)
	buf.Write(hdr[:])

	r := NewBufferFromBytes(buf.Bytes())
	got := r.ReadString()
	if got != "" {
		t.Errorf("ReadString on truncated payload should return \"\", got %q", got)
	}
	if !errors.Is(r.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("Err() = %v, want io.ErrUnexpectedEOF", r.Err())
	}
}

// TestReadStringTruncatedReaderMode verifies the reader-mode counterpart:
// readBuffFull failure now records the wrapped error on Err().
func TestReadStringTruncatedReaderMode(t *testing.T) {
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], 100) // claims 100 bytes
	// Provide only the header — no payload follows.
	r := NewBufferFromReader(bytes.NewReader(hdr[:]))
	got := r.ReadString()
	if got != "" {
		t.Errorf("ReadString on truncated reader should return \"\", got %q", got)
	}
	if r.Err() == nil {
		t.Errorf("Err() should be non-nil on truncated reader, got nil")
	}
}

// TestReadStringAtCap verifies that a string exactly at MaxStringLength bytes
// in byte-buffer mode round-trips successfully (the cap is "<=" allowed).
func TestReadStringAtCap(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates 64 MiB")
	}
	// Encode a uint32 length of MaxStringLength followed by that many bytes.
	w := NewBuffer()
	s := strings.Repeat("a", MaxStringLength)
	w.WriteString(s)
	r := NewBufferFromBytes(w.Bytes())
	got := r.ReadString()
	if len(got) != MaxStringLength {
		t.Errorf("ReadString at cap: got len=%d, want %d", len(got), MaxStringLength)
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v, want nil", r.Err())
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
