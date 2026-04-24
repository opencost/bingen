package stringutil

import (
	"sync"
)

type StringBank interface {
	LoadOrStore(key, value string) (string, bool)
	LoadOrStoreFunc(key string, f func() string) (string, bool)
	Clear()
}

var (
	lock sync.RWMutex

	// stringBank is an unbounded string cache that is thread-safe. It is especially useful if
	// storing a large frequency of dynamically allocated duplicate strings.
	strings StringBank = NewStringBank()
)

func UpdateStringBank(sb StringBank) {
	lock.Lock()
	defer lock.Unlock()

	strings.Clear()
	strings = sb
}

// GetStringBank returns the _current_ StringBank implementation. Note that the read-lock is
// not held for the duration of usage, so the returned string bank could be swapped out
// after being retrieved.
func GetStringBank() StringBank {
	lock.RLock()
	defer lock.RUnlock()

	return strings
}

// Bank will return a non-copy of a string if it has been used before. Otherwise, it will store
// the string as the unique instance.
func Bank(s string) string {
	ss, _ := GetStringBank().LoadOrStore(s, s)
	return ss
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
