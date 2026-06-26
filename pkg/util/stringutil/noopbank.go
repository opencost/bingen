package stringutil

type noOpStringBank struct{}

// NewNoOpStringBank returns a no-op implementation of the `StringBank` which provides no interning
// and always allocates the string data.
func NewNoOpStringBank() StringBank {
	return new(noOpStringBank)
}

// LoadOrStoreFunc for the no-op implementation _always_ runs the allocating func and returns false,
// indicating that the string was not cached and the allocation ran.
func (nsb *noOpStringBank) LoadOrStoreFunc(key string, f func() string) (string, bool) {
	return f(), false
}

// Clear performs no action on the no-op string bank.
func (nsb *noOpStringBank) Clear() {}
