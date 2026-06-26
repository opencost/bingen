package table

import (
	"cmp"
	"slices"

	"github.com/opencost/bingen/pkg/util"
)

type indexed struct {
	s     string
	count uint64
	index int
}

func newIndexed(s string, index int) *indexed {
	return &indexed{
		s:     s,
		count: 1,
		index: index,
	}
}

// PrepassStringTableWriter maps strings to specific indices for encoding, sorted by the total
// number of times they're accessed
type PrepassStringTableWriter struct {
	prepass map[string]*indexed
	next    int
}

// NewPrepassStringTableWriter creates a new PrepassStringTableWriter instance.
func NewPrepassStringTableWriter() *PrepassStringTableWriter {
	return &PrepassStringTableWriter{
		prepass: make(map[string]*indexed),
	}
}

// AddOrGet retrieves a string entry's index if it exists. Otherwise, it adds the entry and returns the new index.
func (st *PrepassStringTableWriter) AddOrGet(s string) int {
	if ind, ok := st.prepass[s]; ok {
		ind.count += 1
		return ind.index
	}

	current := st.next
	st.next++

	st.prepass[s] = newIndexed(s, current)
	return current
}

// WriteSortedTo sorts the string table by the number of accesses, writes the table in that
// order, then returns a new StringTableWriter implementation that can be used for the new
// sorted order index lookups.
func (st *PrepassStringTableWriter) WriteSortedTo(buff *util.Buffer) StringTableWriter {
	sl := make([]*indexed, st.next)
	for _, ind := range st.prepass {
		sl[ind.index] = ind
	}

	slices.SortFunc(sl, func(a *indexed, b *indexed) int {
		return -cmp.Compare(a.count, b.count)
	})

	sti := NewIndexedStringTableWriter()
	for _, ind := range sl {
		sti.AddOrGet(ind.s)
	}

	sti.WriteTo(buff)
	return sti
}

// WriteTo will write the StringTable data (with the header) to the provided
// Buffer starting a the current write position
func (st *PrepassStringTableWriter) WriteTo(buff *util.Buffer) {
	panic("Prepass StringTableWriter cannot write directly")
}
