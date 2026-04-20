package stringutil

import (
	"sync"
	"testing"
)

func TestLoadOrStoreStoresValueUnderKey(t *testing.T) {
	sb := newStringBank()

	got, loaded := sb.LoadOrStore("k1", "v1")
	if loaded {
		t.Fatalf("expected loaded=false on initial store, got true")
	}
	if got != "v1" {
		t.Fatalf("expected stored value %q, got %q", "v1", got)
	}

	got2, loaded2 := sb.LoadOrStore("k1", "different")
	if !loaded2 {
		t.Fatalf("expected loaded=true on subsequent LoadOrStore")
	}
	if got2 != "v1" {
		t.Fatalf("expected previously stored value %q, got %q", "v1", got2)
	}
}

// TestLoadOrStoreFuncStoresUnderKey guards against a regression where the
// computed value was stored at sb.m[value] instead of sb.m[key], which broke
// subsequent lookups when key != value.
func TestLoadOrStoreFuncStoresUnderKey(t *testing.T) {
	sb := newStringBank()

	key := "alias"
	computed := "canonical-string"

	calls := 0
	got, loaded := sb.LoadOrStoreFunc(key, func() string {
		calls++
		return computed
	})
	if loaded {
		t.Fatalf("expected loaded=false on initial LoadOrStoreFunc, got true")
	}
	if got != computed {
		t.Fatalf("expected %q, got %q", computed, got)
	}
	if calls != 1 {
		t.Fatalf("expected f to be called exactly once, got %d", calls)
	}

	got2, loaded2 := sb.LoadOrStoreFunc(key, func() string {
		calls++
		return "should-not-be-used"
	})
	if !loaded2 {
		t.Fatalf("expected loaded=true on subsequent call with the same key")
	}
	if got2 != computed {
		t.Fatalf("expected cached %q, got %q", computed, got2)
	}
	if calls != 1 {
		t.Fatalf("expected f to NOT be called on cache hit, total calls=%d", calls)
	}

	if _, ok := sb.m[key]; !ok {
		t.Fatalf("expected value to be stored under key %q, not found", key)
	}
	if _, ok := sb.m[computed]; ok {
		t.Fatalf("value must not be stored under the computed value as a key")
	}
}

func TestClear(t *testing.T) {
	sb := newStringBank()
	sb.LoadOrStore("a", "A")
	sb.LoadOrStore("b", "B")
	if len(sb.m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(sb.m))
	}

	sb.Clear()
	if len(sb.m) != 0 {
		t.Fatalf("expected 0 entries after Clear, got %d", len(sb.m))
	}
}

func TestBankDeduplicatesStrings(t *testing.T) {
	ClearBank()
	a := Bank("hello")
	b := Bank("hello")
	if a != b {
		t.Errorf("expected Bank to return equal strings, got %q and %q", a, b)
	}
}

func TestBankFuncUsesKey(t *testing.T) {
	ClearBank()

	calls := 0
	first := BankFunc("k", func() string {
		calls++
		return "stored"
	})
	if first != "stored" {
		t.Fatalf("expected %q, got %q", "stored", first)
	}

	second := BankFunc("k", func() string {
		calls++
		return "other"
	})
	if second != "stored" {
		t.Fatalf("expected cached %q, got %q", "stored", second)
	}
	if calls != 1 {
		t.Fatalf("expected f called once, got %d", calls)
	}
}

func TestConcurrentLoadOrStore(t *testing.T) {
	sb := newStringBank()

	var wg sync.WaitGroup
	const goroutines = 32
	const iters = 500

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				sb.LoadOrStore("shared-key", "shared-value")
			}
		}()
	}
	wg.Wait()

	if v, ok := sb.m["shared-key"]; !ok || v != "shared-value" {
		t.Fatalf("concurrent LoadOrStore: got (%q, %v), want (shared-value, true)", v, ok)
	}
}
