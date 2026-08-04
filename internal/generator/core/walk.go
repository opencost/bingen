package core

import "github.com/opencost/bingen/internal/types"

// EncodeStruct renders the field encoding body for the provided types.StructType in
// field order, through the nil-flag and type dispatch. The surrounding method's version
// and hook handling is provided by the Backend implementation.
func EncodeStruct(b Backend, st *types.StructType) (string, []string, error) {
	e := NewEmit(2)
	for _, f := range st.Fields {
		encodeType(e, b, f.Type, b.EncodeFieldSource(f.Name))
	}
	return e.Result()
}

// encodeType writes a leading presence byte for nilable types, then the raw
// value. Aliases re-enter here (not encodeRaw) so an alias of a nilable type
// keeps its nil flag
func encodeType(e *Emit, b Backend, t types.GenType, target string) {
	if !t.IsNilable() {
		encodeRaw(e, b, t, target)
		return
	}

	b.NilGuard(e, target,
		func() {
			b.WriteNilFlag(e, false)
		},
		func() {
			b.WriteNilFlag(e, true)
			encodeRaw(e, b, t, target)
		})
}

func encodeRaw(e *Emit, b Backend, t types.GenType, target string) {
	code := t.Code()

	// handle non-string primitive values first
	if types.IsNonStringPrimitive(code) {
		b.WritePrimitive(e, t, target)
		return
	}

	// handle remaining types
	switch code {
	case types.TypeString:
		b.WriteString(e, t, target)
	case types.TypeSlice:
		st := t.(*types.SliceType)
		b.WriteSlice(e, st, target, func(elem string) {
			encodeType(e, b, st.InnerType, elem)
		})
	case types.TypeMap:
		mt := t.(*types.MapType)
		b.WriteMap(e, mt, target, func(key, value string) {
			encodeType(e, b, mt.KeyType, key)
			encodeType(e, b, mt.ValueType, value)
		})
	case types.TypeStruct:
		b.WriteStruct(e, t, target)
	case types.TypeInterface:
		b.WriteInterface(e, t, target)
	case types.TypeAlias:
		at := t.(*types.AliasType)
		b.EncodeAlias(e, at, target, func(underlying string) {
			encodeType(e, b, at.Alias, underlying)
		})
	case types.TypeReference:
		if resolved := t.(*types.ReferenceType).Resolve(); resolved != nil {
			encodeRaw(e, b, resolved, target)
			return
		}
		if !b.WriteExternal(e, t, target) {
			e.Errf("unsupported external reference type %q", t.Name())
		}

	default:
		e.Errf("unsupported type %q (code %d)", t.Name(), code)
	}
}

// DecodeStruct renders the field-decoding body for st, committing each field to
// the Backend's builder. The surrounding method (version read/check, build,
// migrate/post-process hooks) belongs to the Backend's file scaffolding.
func DecodeStruct(b Backend, st *types.StructType) (string, []string, error) {
	e := NewEmit(2)
	for _, f := range st.Fields {
		decodeField(e, b, f)
	}
	return e.Result()
}

// decodeField applies version gating: fields newer than the payload version get
// the default; otherwise the value is read and committed.
func decodeField(e *Emit, b Backend, f *types.StructField) {
	var fieldVersion uint8
	if f.Opts != nil {
		fieldVersion = f.Opts.Version
	}

	if fieldVersion == 0 {
		b.CommitField(e, f.Name, decodeValue(e, b, f.Type))
		return
	}

	b.VersionGate(
		e,
		fieldVersion,
		func() {
			b.CommitField(e, f.Name, decodeValue(e, b, f.Type))
		},
		func() {
			b.CommitField(e, f.Name, b.Default(f.Type, f.Opts))
		})
}

// decodeValue declares a new variable, reads into it, and returns its name.
func decodeValue(e *Emit, b Backend, t types.GenType) string {
	v := e.NextVar()

	b.DeclareVar(e, t, v)
	decodeInto(e, b, t, v)

	return v
}

// decodeInto reads a leading presence byte for nilable types, then the raw
// value. Aliases re-enter here (not decodeRaw) so an alias of a nilable type
// consumes its nil flag, mirroring encodeType.
func decodeInto(e *Emit, b Backend, t types.GenType, v string) {
	if !t.IsNilable() {
		decodeRaw(e, b, t, v)
		return
	}

	b.ReadNilGuard(e, v, func() {
		decodeRaw(e, b, t, v)
	})
}

func decodeRaw(e *Emit, b Backend, t types.GenType, v string) {
	code := t.Code()

	// handle non-string primitive values first
	if types.IsNonStringPrimitive(code) {
		b.ReadPrimitive(e, t, v)
		return
	}

	switch code {
	case types.TypeString:
		b.ReadString(e, t, v)
	case types.TypeSlice:
		st := t.(*types.SliceType)
		b.ReadSlice(e, st, v, func() string {
			return decodeValue(e, b, st.InnerType)
		})
	case types.TypeMap:
		mt := t.(*types.MapType)
		b.ReadMap(e, mt, v, func() (string, string) {
			return decodeValue(e, b, mt.KeyType), decodeValue(e, b, mt.ValueType)
		})
	case types.TypeStruct:
		b.ReadStruct(e, t, v)
	case types.TypeInterface:
		b.ReadInterface(e, t, v)
	case types.TypeAlias:
		at := t.(*types.AliasType)
		b.DecodeAlias(e, at, v, func(underlying string) {
			decodeInto(e, b, at.Alias, underlying)
		})
	case types.TypeReference:
		if resolved := t.(*types.ReferenceType).Resolve(); resolved != nil {
			decodeRaw(e, b, resolved, v)
			return
		}
		if !b.ReadExternal(e, t, v) {
			e.Errf("unsupported external reference type %q", t.Name())
		}

	default:
		e.Errf("unsupported type %q (code %d)", t.Name(), code)
	}
}
