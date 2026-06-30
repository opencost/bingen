package table

import (
	"fmt"

	"github.com/opencost/bingen/pkg/util"
)

// SliceStringTableReader is a basic pre-loaded []string that provides index-based access.
// The cost of this implementation is holding all strings in memory, which provides faster
// lookup performance at the expense of memory usage.
type SliceStringTableReader struct {
	table []string
}

// NewSliceStringTableReaderFrom creates a new SliceStringTableReader instance loading
// data directly from the buffer. The buffer's position should start at the table length.
func NewSliceStringTableReaderFrom(buffer *util.Buffer) StringTableReader {
	// table length
	tl := buffer.ReadInt()

	var table []string
	if tl > 0 {
		table = make([]string, tl)
		for i := range tl {
			table[i] = buffer.ReadString()
		}
	}

	return &SliceStringTableReader{
		table: table,
	}
}

// At returns the string entry at a specific index, or panics on out of bounds.
func (sstr *SliceStringTableReader) At(index int) string {
	if index < 0 || index >= len(sstr.table) {
		panic(fmt.Errorf("string table index out of bounds: %d", index))
	}

	return sstr.table[index]
}

// Len returns the total number of strings loaded in the string table.
func (sstr *SliceStringTableReader) Len() int {
	if sstr == nil {
		return 0
	}

	return len(sstr.table)
}

// Close for the slice tables just nils out the slice and returns
func (sstr *SliceStringTableReader) Close() error {
	sstr.table = nil
	return nil
}
