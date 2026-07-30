package streamfeed

// Feed is a streamable type using only string/int/collection fields so that a
// streamed value's textual form is identical in Go and Java, enabling a
// cross-language stream-sequence comparison.
type Feed struct {
	Name  string
	Tags  []string
	Count int
	Meta  map[string]string
}
