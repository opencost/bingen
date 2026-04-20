package vars

import (
	"fmt"
	"strings"
)

// Target is used to reference a target property or identifier
type Target string

// Deref returns a dereferenced target.
func (t Target) Deref() Target {
	return "*" + t
}

// CastAs returns the target casted as a specific type.
func (t Target) CastAs(typeName string) Target {
	return Target(fmt.Sprintf("%s(%s)", typeName, t))
}

// String returns the string representation of the Target
func (t Target) String() string {
	return string(t)
}

func AsTarget(name string) Target {
	return Target(name)
}

// NewFieldTarget creates and returns an <obj>.<field> Target
func NewFieldTarget(obj string, field string) Target {
	return Target(fmt.Sprintf("%s.%s", obj, field))
}

func NewIndexedTarget(target Target, idx string) Target {
	return Target(fmt.Sprintf("%s[%s]", target, idx))
}

func NewCastedTarget(cast string, target Target) Target {
	return Target(fmt.Sprintf("%s(%s)", cast, target))
}

func NewTypeCastedTarget(cast string, target Target) Target {
	return Target(fmt.Sprintf("((%s)(%s))", cast, target))
}

// VarNames generates variable names
type VarNames interface {
	// Generates a returns the next variable name
	Next() string
}

// Alpha Var Runes
const (
	A rune = iota + 97
	B
	C
	D
	E
	F
	G
	H
	I
	J
	K
	L
	M
	N
	O
	P
	Q
	R
	S
	T
	U
	V
	W
	X
	Y
	Z
)

// string of all lowercase alpha
const alphaChars string = "abcdefghijklmnopqrstuvwxyz"

// AlphaCharFor returns the lower case alpha character for a rune
func AlphaCharFor(r rune) string {
	if r < A || r > Z {
		return ""
	}

	i := int32(r - A)
	return alphaChars[i : i+1]
}

// ToAlphaChars accepts a var list of runes and returns the strings
func ToAlphaChars(rs ...rune) []string {
	s := []string{}
	for _, r := range rs {
		c := AlphaCharFor(r)
		if c != "" {
			s = append(s, c)
		}
	}
	return s
}

// All Alpha Runes
func AllAlphaRunes() []rune {
	var alpha []rune
	for i := A; i <= Z; i++ {
		alpha = append(alpha, i)
	}
	return alpha
}

// OnlyAlphaRunes returns a skip list to pass omitting all but the provided runes
func SkipAllBut(allow ...rune) []rune {
	var skip []rune
	alphas := AllAlphaRunes()
	for _, r := range alphas {
		found := false
		for _, ar := range allow {
			if ar == r {
				found = true
				break
			}
		}
		if !found {
			skip = append(skip, r)
		}
	}
	return skip
}

// runeSet is used to store runes
type runeSet map[rune]bool

// creates a new rune set provided the slice of runes
func newRuneSet(runes []rune) runeSet {
	m := make(map[rune]bool)
	for _, r := range runes {
		m[r] = true
	}

	return runeSet(m)
}

// Has returns true if the rune set contains the rune
func (rs runeSet) Has(r rune) bool {
	return rs[r]
}

// VarNames implementation which creates vars a-z, aa-zz, aaa-zzz, etc...
type alphaVarNames struct {
	inc     int32
	cycle   int
	skipSet runeSet
}

// NewAlphaVarNames returns a new VarNames implementation capable of ignoring specific runes.
func NewAlphaVarNames(skip ...rune) VarNames {
	avn := &alphaVarNames{
		inc:     0,
		cycle:   1,
		skipSet: newRuneSet(skip),
	}

	avn.skip()
	return avn
}

// Next returns a valid string representing the next variable name.
func (vn *alphaVarNames) Next() string {
	varName := strings.Repeat(vn.current(), vn.cycle)
	vn.incrementAndSkip()

	return varName
}

// current string for the current rune
func (vn *alphaVarNames) current() string {
	return string([]rune{vn.currentRune()})
}

// current rune given position
func (vn *alphaVarNames) currentRune() rune {
	return rune(A + vn.inc)
}

// skip any ignored runes. Panics if the skip set contains every alpha rune,
// which would otherwise cause an infinite loop.
func (vn *alphaVarNames) skip() {
	checked := 0
	for vn.skipSet.Has(rune(A + vn.inc)) {
		checked++
		if checked >= 26 {
			panic("vars: no available variable names; skip set contains all letters")
		}
		vn.increment()
	}
}

// skip any ignored runes and increment the position if necessary
func (vn *alphaVarNames) incrementAndSkip() {
	vn.increment()
	vn.skip()
}

// increment the rune position and cycle if necessary
func (vn *alphaVarNames) increment() {
	vn.inc = (vn.inc + 1) % 26
	if vn.inc == 0 {
		vn.cycle++
	}
}
