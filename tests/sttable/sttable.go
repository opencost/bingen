package sttable

// Doc exercises string-table encoding: repeated strings across fields should be
// deduplicated into a shared table.
type Doc struct {
	Title  string
	Tags   []string
	Author string
}
