package stringutil

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// TestLruEvictionUsesKey verifies that LoadOrStore actually shrinks the map
// when the eviction threshold is crossed. The previous implementation deleted
// by `value` rather than `key`, which silently no-op'd on the LoadOrStore path
// and let the bank grow without bound.
func TestLruEvictionUsesKey(t *testing.T) {
	bank := newLruForTest(t, 4)

	// Insert enough entries to force the in-line eviction (capacity + capacity/2).
	// With capacity=4 the threshold is 6, so add 7 distinct (key,value) pairs
	// where key != value to exercise the buggy path.
	for i := 0; i < 7; i++ {
		key := fmt.Sprintf("key-%d", i)
		val := fmt.Sprintf("val-%d", i)
		bank.LoadOrStore(key, val)
	}

	got := bankSize(bank)
	if got > 4 {
		t.Errorf("expected eviction to bring map to <= capacity (4), got size %d", got)
	}
}

// TestUpdateStringBankClosesPrevious verifies that swapping the active bank
// stops the previous LRU bank's background goroutine. Without Close, every
// swap would leak a goroutine.
func TestUpdateStringBankClosesPrevious(t *testing.T) {
	// settle goroutines before capturing the baseline.
	time.Sleep(10 * time.Millisecond)
	runtime.Gosched()

	before := runtime.NumGoroutine()

	for i := 0; i < 25; i++ {
		UpdateStringBank(NewLruStringBank(8, 50*time.Millisecond))
	}

	// Replace once more with a no-op to release the last LRU's goroutine too.
	UpdateStringBank(NewNoOpStringBank())

	// Give scheduled goroutines a chance to exit.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond)
		if runtime.NumGoroutine() <= before+2 {
			return
		}
	}

	if delta := runtime.NumGoroutine() - before; delta > 2 {
		t.Errorf("UpdateStringBank leaked goroutines: delta=%d", delta)
	}
}

// helpers

func newLruForTest(t *testing.T, capacity int) *lruStringBank {
	t.Helper()
	// Use a long eviction interval so the background goroutine doesn't race
	// with the in-line eviction we're trying to observe.
	sb := NewLruStringBank(capacity, time.Hour).(*lruStringBank)
	t.Cleanup(func() { _ = sb.Close() })
	return sb
}

func bankSize(sb *lruStringBank) int {
	sb.lock.Lock()
	defer sb.lock.Unlock()
	return len(sb.m)
}
