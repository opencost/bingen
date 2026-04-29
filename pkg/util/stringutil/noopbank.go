package stringutil

type noOpStringBank struct{}

func NewNoOpStringBank() StringBank {
	return new(noOpStringBank)
}

// LoadOrStore for the no-op bank always reports "not loaded" (false): the caller
// is treated as the first writer for every key, since nothing is cached.
func (nsb *noOpStringBank) LoadOrStore(key, value string) (string, bool) {
	return value, false
}

// LoadOrStoreFunc mirrors LoadOrStore: the function is always invoked and the
// result is reported as freshly stored.
func (nsb *noOpStringBank) LoadOrStoreFunc(key string, f func() string) (string, bool) {
	return f(), false
}

func (nsb *noOpStringBank) Clear() {}

func (nsb *noOpStringBank) Close() error { return nil }
