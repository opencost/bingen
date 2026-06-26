package table

import (
	"github.com/opencost/bingen/pkg/util"
)

const (
	// BinaryTagStringTable is written and/or read prior to the existence of a string
	// table (where each index is encoded as a string entry in the resource
	BinaryTagStringTable string = "BGST"
)

//--------------------------------------------------------------------------
//  String Table Reader
//--------------------------------------------------------------------------

// StringTableReader is the interface used to read the string table from the decoding.
type StringTableReader interface {
	// At returns the string entry at a specific index, or panics on out of bounds.
	At(index int) string

	// Len returns the total number of strings loaded in the string table.
	Len() int

	// Close will clear the loaded table, and drop any external resources used.
	Close() error
}

//--------------------------------------------------------------------------
//  String Table Writer
//--------------------------------------------------------------------------

// StringTableWriter is the interface used to write the string table for encoding.
type StringTableWriter interface {
	// AddOrGet adds a string to the string table and returns the new index or
	// an existing index.
	AddOrGet(s string) int

	// WriteTo will write the StringTable data (with the header) to the provided
	// Buffer starting a the current write position
	WriteTo(b *util.Buffer)
}
