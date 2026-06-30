package table

import "github.com/opencost/bingen/pkg/util"

// IndexedStringTableWriter maps strings to specific indices for encoding
type IndexedStringTableWriter struct {
	indices map[string]int
	next    int
}

// NewIndexedStringTableWriter Creates a new IndexedStringTableWriter instance.
func NewIndexedStringTableWriter() *IndexedStringTableWriter {
	return &IndexedStringTableWriter{
		indices: make(map[string]int),
		next:    0,
	}
}

// AddOrGet retrieves a string entry's index if it exists. Otherwise, it adds the entry and returns the new index.
func (st *IndexedStringTableWriter) AddOrGet(s string) int {
	if ind, ok := st.indices[s]; ok {
		return ind
	}

	current := st.next
	st.next++

	st.indices[s] = current
	return current
}

// ToSlice Converts the contents to a string array for encoding.
func (st *IndexedStringTableWriter) ToSlice() []string {
	if st.next == 0 {
		return []string{}
	}

	sl := make([]string, st.next)
	for s, i := range st.indices {
		sl[i] = s
	}
	return sl
}

// ToBytes Converts the contents to a binary encoded representation
func (st *IndexedStringTableWriter) ToBytes() []byte {
	buff := util.NewBuffer()
	st.WriteTo(buff)
	return buff.Bytes()
}

// WriteTo will write the StringTable data (with the header) to the provided
// Buffer starting a the current write position
func (st *IndexedStringTableWriter) WriteTo(buff *util.Buffer) {
	// bingen string table header
	buff.WriteBytes([]byte(BinaryTagStringTable))

	// get an ordered string slice to encode
	strs := st.ToSlice()

	buff.WriteInt(len(strs)) // table length
	for _, s := range strs {
		buff.WriteString(s)
	}
}
