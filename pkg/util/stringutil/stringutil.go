package stringutil

import (
	"math/rand"
	"sync"
	"time"
)

type stringBank struct {
	lock sync.Mutex
	m    map[string]string
}

func newStringBank() *stringBank {
	return &stringBank{
		m: make(map[string]string),
	}
}

func (sb *stringBank) LoadOrStore(key, value string) (string, bool) {
	sb.lock.Lock()

	if v, ok := sb.m[key]; ok {
		sb.lock.Unlock()
		return v, ok
	}

	sb.m[key] = value
	sb.lock.Unlock()
	return value, false
}

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

func (sb *stringBank) Clear() {
	sb.lock.Lock()
	sb.m = make(map[string]string)
	sb.lock.Unlock()
}

// stringBank is an unbounded string cache that is thread-safe. It is especially useful if
// storing a large frequency of dynamically allocated duplicate strings.
var strings = newStringBank() // sync.Map

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Bank will return a non-copy of a string if it has been used before. Otherwise, it will store
// the string as the unique instance.
func Bank(s string) string {
	ss, _ := strings.LoadOrStore(s, s)
	return ss
}

// BankFunc will use the provided s string to check for an existing allocation of the string. However,
// if no allocation exists, the f parameter will be used to create the string and store in the bank.
func BankFunc(s string, f func() string) string {
	ss, _ := strings.LoadOrStoreFunc(s, f)
	return ss
}

func ClearBank() {
	strings.Clear()
}
