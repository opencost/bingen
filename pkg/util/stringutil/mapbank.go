package stringutil

import "sync"

type stringBank struct {
	lock sync.Mutex
	m    map[string]string
}

// NewStringBank creates a set backed cache that allows for allocated strings existance to be
// checked by temporary strings.
func NewStringBank() StringBank {
	return &stringBank{
		m: make(map[string]string),
	}
}

// LoadOrStoreFunc provides a temporary string (casted from a byte stream) as the `key`. If the temporary string
// does not have a heap allocated variant, the provided `func() string` should allocate and return the representation.
// If the temporary string _does_ have a heap allocated variant, the `f func() string` is not called and the previously
// allocated string is returned.
func (sb *stringBank) LoadOrStoreFunc(key string, f func() string) (string, bool) {
	sb.lock.Lock()

	if v, ok := sb.m[key]; ok {
		sb.lock.Unlock()
		return v, ok
	}

	// create the key and value using the func (the key could be deallocated later)
	value := f()
	sb.m[value] = value
	sb.lock.Unlock()
	return value, false
}

// Clear deallocates the string cache and creates a new internal storage.
func (sb *stringBank) Clear() {
	sb.lock.Lock()
	sb.m = make(map[string]string)
	sb.lock.Unlock()
}
