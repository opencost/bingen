package timerec

import "time"

// Event exercises a time.Time field, which bingen encodes as an external
// reference (length-prefixed time.Time.MarshalBinary output).
type Event struct {
	Name string
	At   time.Time
	Seq  int
}
