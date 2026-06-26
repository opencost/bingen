package table

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/opencost/bingen/pkg/util"
)

const StringHeaderSize = int64(unsafe.Sizeof(""))

// fileStringRef maps a bingen string-table index to a payload stored in a temp file.
type fileStringRef struct {
	off    int64
	length int
}

// FileStringTableReader leverages a local file to write string table data for lookup. On
// memory focused systems, this allows a slower parse with a significant decrease in memory
// usage. This implementation is often pair with streaming readers for high throughput with
// reduced memory usage.
type FileStringTableReader struct {
	filePrefix string
	f          *os.File
	refs       []fileStringRef
	memo       []string
}

// NewFileStringTableFromBuffer reads exactly tl length-prefixed (uint16) string payloads from buffer
// and appends each payload to a new temp file. It does not retain full strings in memory.
func NewFileStringTableReaderFrom(buffer *util.Buffer, dir string, filePrefix string, memoMaxBytes int64) StringTableReader {
	// helper func to cast a string in-place to a byte slice.
	// NOTE: Return value is READ-ONLY. DO NOT MODIFY!
	byteSliceFor := func(s string) []byte {
		return unsafe.Slice(unsafe.StringData(s), len(s))
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Errorf("%s: failed to create string table directory: %w", filePrefix, err))
	}

	f, err := os.CreateTemp(dir, fmt.Sprintf("%s-bgst-*", filePrefix))
	if err != nil {
		panic(fmt.Errorf("%s: failed to create string table file: %w", filePrefix, err))
	}

	var writeErr error
	defer func() {
		if writeErr != nil {
			_ = f.Close()
		}
	}()

	// table length
	tl := buffer.ReadInt()

	var refs []fileStringRef
	if tl > 0 {
		refs = make([]fileStringRef, tl)

		for i := range tl {
			payload := byteSliceFor(buffer.ReadString())

			var off int64
			if len(payload) > 0 {
				off, err = f.Seek(0, io.SeekEnd)
				if err != nil {
					writeErr = fmt.Errorf("%s: failed to seek string table file: %w", filePrefix, err)
					panic(writeErr)
				}
				if _, err := f.Write(payload); err != nil {
					writeErr = fmt.Errorf("%s: failed to write string table entry %d: %w", filePrefix, i, err)
					panic(writeErr)
				}
			}

			refs[i] = fileStringRef{
				off:    off,
				length: len(payload),
			}
		}
	}

	var memo []string

	// Pre-load cache with strings up to memoMaxBytes, respecting string boundaries
	if memoMaxBytes > 0 && len(refs) > 0 {
		memo = make([]string, len(refs))
		var cumulativeSize int64
		for i, ref := range refs {
			// Check if adding this string would exceed the limit
			if cumulativeSize+int64(ref.length)+StringHeaderSize > memoMaxBytes {
				// Would exceed limit, stop here
				break
			}

			// Read string from file and cache it
			if ref.length > 0 {
				b := make([]byte, ref.length)
				_, err := f.ReadAt(b, ref.off)
				if err != nil {
					// If we can't read, skip this entry but continue
					continue
				}

				// Cast the allocated bytes to a string in-place
				str := unsafe.String(unsafe.SliceData(b), len(b))
				memo[i] = str
				cumulativeSize += int64(ref.length) + StringHeaderSize
			}
		}
	}

	return &FileStringTableReader{
		filePrefix: filePrefix,
		f:          f,
		refs:       refs,
		memo:       memo,
	}
}

// At returns the string from the internal file using the reference's offset and length.
func (fstr *FileStringTableReader) At(index int) string {
	if fstr == nil || fstr.f == nil {
		panic(fmt.Errorf("%s: failed to read file string table data", fstr.filePrefix))
	}
	if index < 0 || index >= len(fstr.refs) {
		panic(fmt.Errorf("%s: string table index out of bounds: %d", fstr.filePrefix, index))
	}

	ref := fstr.refs[index]
	if ref.length == 0 {
		return ""
	}

	// Check cache first
	if fstr.memo != nil && len(fstr.memo) > index && fstr.memo[index] != "" {
		return fstr.memo[index]
	}

	// Cache miss - read from file
	b := make([]byte, ref.length)
	_, err := fstr.f.ReadAt(b, ref.off)
	if err != nil {
		return ""
	}

	// Cast the allocated bytes to a string in-place, as we were the ones that allocated the bytes
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// Len returns the total number of strings loaded in the string table.
func (fstr *FileStringTableReader) Len() int {
	if fstr == nil {
		return 0
	}

	return len(fstr.refs)
}

// Close for the file string table reader closes the file and deletes it.
func (fstr *FileStringTableReader) Close() error {
	if fstr == nil || fstr.f == nil {
		return nil
	}

	path := fstr.f.Name()
	err := fstr.f.Close()
	fstr.f = nil
	fstr.refs = nil
	fstr.memo = nil

	if path != "" {
		_ = os.Remove(path)
	}

	return err
}
