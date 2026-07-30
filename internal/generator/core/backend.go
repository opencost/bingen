package core

import "github.com/opencost/bingen/internal/types"

// Backend renders the target-language syntax for each node the core codec walk
// visits. Backend owns type mapping, naming, and how each operation is output.
// Each method documents it's binary contract such that the core's guarantees hold
// across languages.
//
// Callback parameters (whenNil, whenPresent, body, readElem, readPair) are
// invoked by the Backend at the point the recursed code belongs, letting the
// Backend own the surrounding control-flow syntax.
type Backend interface {
	// TypeName returns the target type for t. When boxed is true a nullable form
	// is required (ie: Java Integer rather than int). The core requests boxed
	// for collection elements and pointer values.
	TypeName(t types.GenType, boxed bool) string

	// EncodeFieldSource returns the expression reading struct field from the
	// value being encoded.
	EncodeFieldSource(field string) string

	// CommitField assigns valueExpr into field on the value being built.
	CommitField(e *Emit, field string, valueExpr string)

	// Default returns the value expression for a field whose version exceeds the
	// payload version.
	Default(t types.GenType, opts *types.FieldOpts) string

	// Comment writes a comment to the generated output with the provided text
	Comment(e *Emit, text string)

	//--------------------------------------------------------------------------
	//  Encode
	//--------------------------------------------------------------------------

	// WriteNilFlag writes the one-byte presence marker: present=1, absent=0.
	WriteNilFlag(e *Emit, present bool)

	// NilGuard renders "if target is nil { whenNil } else { whenPresent }".
	NilGuard(e *Emit, target string, whenNil, whenPresent func())

	// WritePrimitive writes a non-string primitive value.
	WritePrimitive(e *Emit, t types.GenType, target string)

	// WriteString writes a string as a string-table index or a raw string,
	// depending on the runtime context.
	WriteString(e *Emit, t types.GenType, target string)

	// WriteSlice writes the element count then iterates, invoking body(elem) to
	// encode each element expression.
	WriteSlice(e *Emit, t *types.SliceType, target string, body func(elem string))

	// WriteMap writes the entry count then iterates, invoking body(key, value)
	// to encode each pair (key before value).
	WriteMap(e *Emit, t *types.MapType, target string, body func(key, value string))

	// WriteStruct writes the compatibility marker then delegates to the nested
	// type's encoder.
	WriteStruct(e *Emit, t types.GenType, target string)

	// WriteInterface writes the concrete type name and delegates through the
	// registry.
	WriteInterface(e *Emit, t types.GenType, target string)

	// WriteExternal writes a non-generated reference type handled by the runtime
	// (ie: time.Time). It returns false if name is unsupported.
	WriteExternal(e *Emit, t types.GenType, target string) bool

	// EncodeAlias encodes an alias value. body encodes the underlying value; the
	// Backend passes it the target cast to the underlying type when the language
	// requires it.
	EncodeAlias(e *Emit, at *types.AliasType, target string, body func(underlying string))

	//--------------------------------------------------------------------------
	//  Decode
	//--------------------------------------------------------------------------

	// DeclareVar declares an uninitialized local of t's type named v.
	DeclareVar(e *Emit, t types.GenType, v string)

	// VersionGate renders "if fieldVersion <= version { whenCompatible } else {
	// setDefault }".
	VersionGate(e *Emit, fieldVersion uint8, whenCompatible, setDefault func())

	// ReadNilGuard reads the presence byte, assigning nil to v in the absent
	// branch and running whenPresent otherwise.
	ReadNilGuard(e *Emit, v string, whenPresent func())

	// ReadPrimitive reads a non-string primitive into v.
	ReadPrimitive(e *Emit, t types.GenType, v string)

	// ReadString reads a string (table index or raw) into v. t is the (possibly
	// pointer) string type.
	ReadString(e *Emit, t types.GenType, v string)

	// ReadSlice reads the element count then iterates; readElem reads one element
	// and returns the variable holding it, which the Backend appends to v.
	ReadSlice(e *Emit, t *types.SliceType, v string, readElem func() string)

	// ReadMap reads the entry count then iterates; readPair reads one key/value
	// (key before value) and returns their variables, which the Backend puts in v.
	ReadMap(e *Emit, t *types.MapType, v string, readPair func() (key, value string))

	// ReadStruct reads the compatibility marker then delegates to the nested
	// type's decoder, into v.
	ReadStruct(e *Emit, t types.GenType, v string)

	// ReadInterface reads the concrete type name and delegates through the
	// registry, into v.
	ReadInterface(e *Emit, t types.GenType, v string)

	// ReadExternal reads a runtime-handled reference type into v. It returns
	// false if name is unsupported.
	ReadExternal(e *Emit, t types.GenType, v string) bool

	// DecodeAlias decodes into v (declared as the alias type). body reads the
	// underlying value into the variable name it is given; the Backend then
	// converts it to the alias type as needed (Java passes v through unchanged).
	DecodeAlias(e *Emit, at *types.AliasType, v string, body func(underlying string))
}
