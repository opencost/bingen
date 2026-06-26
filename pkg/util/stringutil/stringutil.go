package stringutil

import (
	"strings"
	"sync"
)

// StringBank provides a string interning interface that will check for the existence of a heap allocated
// string using a temporary string (from a byte stream).
type StringBank interface {
	// LoadOrStoreFunc provides a temporary string (casted from a byte stream) as the `key`. If the temporary string
	// does not have a heap allocated variant, the provided `func() string` should allocate and return the representation.
	// If the temporary string _does_ have a heap allocated variant, the `f func() string` is not called and the previously
	// allocated string is returned.
	LoadOrStoreFunc(key string, f func() string) (string, bool)

	// Clears all heap allocated strings from the bank.
	Clear()
}

var (
	lock sync.RWMutex

	// stringBank is an unbounded string cache that is thread-safe. It is especially useful if
	// storing a large frequency of dynamically allocated duplicate bank.
	bank StringBank = NewStringBank()
)

func UpdateStringBank(sb StringBank) {
	lock.Lock()
	defer lock.Unlock()

	bank.Clear()
	bank = sb
}

// GetStringBank returns the _current_ StringBank implementation. Note that the read-lock is
// not held for the duration of usage, so the returned string bank could be swapped out
// after being retrieved.
func GetStringBank() StringBank {
	lock.RLock()
	defer lock.RUnlock()

	return bank
}

// Bank will return a non-copy of a string if it has been used before. Otherwise, it will store
// the string as the unique instance.
func Bank(s string) string {
	return BankFunc(s, func() string {
		return strings.Clone(s)
	})
}

// BankFunc will use the provided s string to check for an existing allocation of the string. However,
// if no allocation exists, the f parameter will be used to create the string and store in the bank.
func BankFunc(s string, f func() string) string {
	ss, _ := GetStringBank().LoadOrStoreFunc(s, f)
	return ss
}

func ClearBank() {
	GetStringBank().Clear()
}
