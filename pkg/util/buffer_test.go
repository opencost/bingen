package util

import (
	"bytes"
	"io"
	"math"
	"strings"
	"testing"
)

// TestBufferRoundTrip exercises each Write*/Read* pair on the in-memory
// (bytes.Buffer backed) variant of Buffer to ensure the error-checked
// Read* paths still produce the expected values.
func TestBufferRoundTrip(t *testing.T) {
	b := NewBuffer()

	b.WriteBool(true)
	b.WriteInt(-42)
	b.WriteInt8(-8)
	b.WriteInt16(-16)
	b.WriteInt32(-32)
	b.WriteInt64(-64)
	b.WriteUInt(42)
	b.WriteUInt8(8)
	b.WriteUInt16(16)
	b.WriteUInt32(32)
	b.WriteUInt64(64)
	b.WriteFloat32(3.25)
	b.WriteFloat64(6.5)
	b.WriteString("hello")

	if got := b.ReadBool(); got != true {
		t.Errorf("ReadBool = %v, want true", got)
	}
	if got := b.ReadInt(); got != -42 {
		t.Errorf("ReadInt = %d, want -42", got)
	}
	if got := b.ReadInt8(); got != -8 {
		t.Errorf("ReadInt8 = %d, want -8", got)
	}
	if got := b.ReadInt16(); got != -16 {
		t.Errorf("ReadInt16 = %d, want -16", got)
	}
	if got := b.ReadInt32(); got != -32 {
		t.Errorf("ReadInt32 = %d, want -32", got)
	}
	if got := b.ReadInt64(); got != -64 {
		t.Errorf("ReadInt64 = %d, want -64", got)
	}
	if got := b.ReadUInt(); got != 42 {
		t.Errorf("ReadUInt = %d, want 42", got)
	}
	if got := b.ReadUInt8(); got != 8 {
		t.Errorf("ReadUInt8 = %d, want 8", got)
	}
	if got := b.ReadUInt16(); got != 16 {
		t.Errorf("ReadUInt16 = %d, want 16", got)
	}
	if got := b.ReadUInt32(); got != 32 {
		t.Errorf("ReadUInt32 = %d, want 32", got)
	}
	if got := b.ReadUInt64(); got != 64 {
		t.Errorf("ReadUInt64 = %d, want 64", got)
	}
	if got := b.ReadFloat32(); got != 3.25 {
		t.Errorf("ReadFloat32 = %v, want 3.25", got)
	}
	if got := b.ReadFloat64(); got != 6.5 {
		t.Errorf("ReadFloat64 = %v, want 6.5", got)
	}
	if got := b.ReadString(); got != "hello" {
		t.Errorf("ReadString = %q, want %q", got, "hello")
	}
}

// TestBufferReaderRoundTrip exercises the bufio.Reader-backed
// (NewBufferFromReader) variant of Buffer.
func TestBufferReaderRoundTrip(t *testing.T) {
	src := NewBuffer()
	src.WriteBool(false)
	src.WriteInt(7)
	src.WriteInt8(-3)
	src.WriteInt16(-300)
	src.WriteInt32(-30000)
	src.WriteInt64(-3_000_000_000)
	src.WriteUInt(9)
	src.WriteUInt8(17)
	src.WriteUInt16(4096)
	src.WriteUInt32(1 << 20)
	src.WriteUInt64(1 << 40)
	src.WriteFloat32(-1.5)
	src.WriteFloat64(-9.75)
	src.WriteString("world")

	r := NewBufferFromReader(bytes.NewReader(src.Bytes()))

	if got := r.ReadBool(); got != false {
		t.Errorf("ReadBool = %v, want false", got)
	}
	if got := r.ReadInt(); got != 7 {
		t.Errorf("ReadInt = %d, want 7", got)
	}
	if got := r.ReadInt8(); got != -3 {
		t.Errorf("ReadInt8 = %d, want -3", got)
	}
	if got := r.ReadInt16(); got != -300 {
		t.Errorf("ReadInt16 = %d, want -300", got)
	}
	if got := r.ReadInt32(); got != -30000 {
		t.Errorf("ReadInt32 = %d, want -30000", got)
	}
	if got := r.ReadInt64(); got != -3_000_000_000 {
		t.Errorf("ReadInt64 = %d, want -3_000_000_000", got)
	}
	if got := r.ReadUInt(); got != 9 {
		t.Errorf("ReadUInt = %d, want 9", got)
	}
	if got := r.ReadUInt8(); got != 17 {
		t.Errorf("ReadUInt8 = %d, want 17", got)
	}
	if got := r.ReadUInt16(); got != 4096 {
		t.Errorf("ReadUInt16 = %d, want 4096", got)
	}
	if got := r.ReadUInt32(); got != 1<<20 {
		t.Errorf("ReadUInt32 = %d, want %d", got, 1<<20)
	}
	if got := r.ReadUInt64(); got != 1<<40 {
		t.Errorf("ReadUInt64 = %d, want %d", got, uint64(1<<40))
	}
	if got := r.ReadFloat32(); got != -1.5 {
		t.Errorf("ReadFloat32 = %v, want -1.5", got)
	}
	if got := r.ReadFloat64(); got != -9.75 {
		t.Errorf("ReadFloat64 = %v, want -9.75", got)
	}
	if got := r.ReadString(); got != "world" {
		t.Errorf("ReadString = %q, want %q", got, "world")
	}
}

// TestBufferReadPanicsOnUnderflow verifies that when the underlying reader is
// exhausted mid-read, the Read* methods surface the error as a panic (rather
// than silently returning zero), which is the contract of the void-returning
// Read* APIs.
func TestBufferReadPanicsOnUnderflow(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*Buffer)
	}{
		{"ReadBool", func(b *Buffer) { b.ReadBool() }},
		{"ReadInt", func(b *Buffer) { b.ReadInt() }},
		{"ReadInt8", func(b *Buffer) { b.ReadInt8() }},
		{"ReadInt16", func(b *Buffer) { b.ReadInt16() }},
		{"ReadInt32", func(b *Buffer) { b.ReadInt32() }},
		{"ReadInt64", func(b *Buffer) { b.ReadInt64() }},
		{"ReadUInt", func(b *Buffer) { b.ReadUInt() }},
		{"ReadUInt8", func(b *Buffer) { b.ReadUInt8() }},
		{"ReadUInt16", func(b *Buffer) { b.ReadUInt16() }},
		{"ReadUInt32", func(b *Buffer) { b.ReadUInt32() }},
		{"ReadUInt64", func(b *Buffer) { b.ReadUInt64() }},
		{"ReadFloat32", func(b *Buffer) { b.ReadFloat32() }},
		{"ReadFloat64", func(b *Buffer) { b.ReadFloat64() }},
	}

	for _, c := range cases {
		t.Run("bw/"+c.name, func(t *testing.T) {
			b := NewBufferFromBytes(nil)
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("%s did not panic on empty buffer", c.name)
				}
			}()
			c.fn(b)
		})

		t.Run("reader/"+c.name, func(t *testing.T) {
			b := NewBufferFromReader(strings.NewReader(""))
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("%s did not panic on empty reader", c.name)
				}
			}()
			c.fn(b)
		})
	}
}

// TestBufferWriteStringClampsToMaxUint16 ensures that very long strings are
// truncated to the uint16 max as documented.
func TestBufferWriteStringClampsToMaxUint16(t *testing.T) {
	b := NewBuffer()
	long := strings.Repeat("a", math.MaxUint16+10)
	b.WriteString(long)

	got := b.ReadString()
	if len(got) != math.MaxUint16 {
		t.Errorf("ReadString length = %d, want %d", len(got), math.MaxUint16)
	}
}

// TestBufferPeekUnsupportedOnRW confirms that Peek is only valid on
// reader-backed buffers.
func TestBufferPeekUnsupportedOnRW(t *testing.T) {
	b := NewBuffer()
	if _, err := b.Peek(1); err == nil {
		t.Fatalf("expected error from Peek on read/write buffer, got nil")
	}
}

// TestBufferBytesDrainsReader ensures that Bytes() on a reader-backed buffer
// drains remaining input, even when the underlying reader is an EOF-only
// source after the first chunk.
func TestBufferBytesDrainsReader(t *testing.T) {
	payload := []byte{1, 2, 3, 4}
	r := NewBufferFromReader(bytes.NewReader(payload))
	got := r.Bytes()
	if !bytes.Equal(got, payload) {
		t.Errorf("Bytes = %v, want %v", got, payload)
	}

	// Second call should return empty since underlying reader is drained.
	if got := r.Bytes(); len(got) != 0 {
		t.Errorf("Bytes after drain = %v, want empty", got)
	}
}

// Ensure Buffer value type satisfies expectations against io.Reader-returning
// helpers.
var _ io.Reader = (*bytes.Reader)(nil)
