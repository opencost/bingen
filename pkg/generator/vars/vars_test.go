package vars

import "testing"

// TestAlphaVarNamesPanicsOnFullSkipSet guards against a regression where the
// internal skip() loop would run forever if the caller asked to skip every
// letter in the alphabet. We now expect an explicit panic.
func TestAlphaVarNamesPanicsOnFullSkipSet(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when all alpha runes are skipped")
		}
	}()

	NewAlphaVarNames(AllAlphaRunes()...)
}

func TestAlphaVarNamesBasicSequence(t *testing.T) {
	vn := NewAlphaVarNames()
	want := []string{"a", "b", "c", "d"}
	for i, w := range want {
		if got := vn.Next(); got != w {
			t.Errorf("iteration %d: expected %q, got %q", i, w, got)
		}
	}
}

// TestAlphaVarNamesSkipAllButOne verifies that the generator produces a
// deterministic sequence of a single rune with increasing repetition count
// when all-but-one letters are skipped.
func TestAlphaVarNamesSkipAllButOne(t *testing.T) {
	vn := NewAlphaVarNames(SkipAllBut(M)...)
	want := []string{"m", "mm", "mmm", "mmmm"}
	for i, w := range want {
		if got := vn.Next(); got != w {
			t.Errorf("iteration %d: expected %q, got %q", i, w, got)
		}
	}
}

func TestSkipAllBut(t *testing.T) {
	skip := SkipAllBut(A, Z)
	if len(skip) != 24 {
		t.Fatalf("expected 24 skipped runes, got %d", len(skip))
	}
	for _, r := range skip {
		if r == A || r == Z {
			t.Errorf("rune %q should not be in skip set", string(r))
		}
	}
}

func TestAlphaCharFor(t *testing.T) {
	if got := AlphaCharFor(A); got != "a" {
		t.Errorf("expected 'a', got %q", got)
	}
	if got := AlphaCharFor(Z); got != "z" {
		t.Errorf("expected 'z', got %q", got)
	}
	if got := AlphaCharFor('0'); got != "" {
		t.Errorf("expected empty string for non-alpha rune, got %q", got)
	}
}
