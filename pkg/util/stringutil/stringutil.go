package stringutil

import (
	"sync"
	"time"
)

// StringBank is a thread-safe string interner. Implementations must be safe to
// call concurrently.
type StringBank interface {
	LoadOrStore(key, value string) (string, bool)
	LoadOrStoreFunc(key string, f func() string) (string, bool)
	Clear()
	// Close releases any background resources held by the bank (for example, the
	// background eviction goroutine in lruStringBank). It is safe to call more
	// than once.
	Close() error
}

const (
	// defaultLruCapacity bounds the default string bank to prevent unbounded
	// memory growth from interning every distinct string seen on the wire.
	defaultLruCapacity = 100_000

	// defaultLruEvictionInterval is the period of the background eviction goroutine
	// for the default string bank.
	defaultLruEvictionInterval = 30 * time.Second
)

var (
	lock sync.RWMutex

	// strings is the active StringBank used to intern decoded strings. The default
	// is a bounded LRU bank — previous releases used an unbounded map, which
	// behaved as a memory leak when fed unique input.
	strings StringBank = NewLruStringBank(defaultLruCapacity, defaultLruEvictionInterval)
)

// UpdateStringBank swaps the active StringBank, closing the previous one to
// avoid leaking goroutines or other background resources.
func UpdateStringBank(sb StringBank) {
	lock.Lock()
	defer lock.Unlock()

	if strings != nil {
		_ = strings.Close()
	}
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
