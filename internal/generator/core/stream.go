package core

import "github.com/opencost/bingen/internal/types"

// StreamBackend renders a streaming decoder. Instead of building a value, it
// yields each field and a single depth of all slices and maps on the type to
// consumer. It reuses the read-value methods from Backend for scalars, structs,
// interfaces, and references, and adds the yield/collection primitives below.
type StreamBackend interface {
	Backend

	// FieldInfo emits the per-field descriptor (name + type) that subsequent
	// yields reference.
	FieldInfo(e *Emit, name string, t types.GenType)

	// Yield yields a single value for the current field.
	Yield(e *Emit, value string)

	// YieldNil yields a nil value for the current field.
	YieldNil(e *Emit)

	// YieldIndexed yields one element of a flattened collection with its index or
	// key.
	YieldIndexed(e *Emit, index, value string)

	// StreamNilGuard reads the presence byte, running whenNil in the absent
	// branch and whenPresent otherwise.
	StreamNilGuard(e *Emit, whenNil, whenPresent func())

	// StreamSlice reads the element count then iterates, invoking body(idx) per
	// element (body reads the element and yields it indexed).
	StreamSlice(e *Emit, t *types.SliceType, body func(index string))

	// StreamMap reads the entry count then iterates, invoking body() per entry
	// (body reads the key/value and yields them keyed).
	StreamMap(e *Emit, t *types.MapType, body func())

	// StreamDefault yields the default for a field newer than the payload
	// version (nil for nilable, the zero/annotation default for basics, nothing
	// for other types without an explicit default).
	StreamDefault(e *Emit, t types.GenType, opts *types.FieldOpts)
}

// StreamStruct renders the streaming-decode body for the provided StructType by yielding
// each of it's fields, in order.
func StreamStruct(b StreamBackend, st *types.StructType) (string, []string, error) {
	e := NewEmit(2)
	for _, f := range st.Fields {
		streamField(e, b, f)
	}
	return e.Result()
}

func streamField(e *Emit, b StreamBackend, f *types.StructField) {
	b.FieldInfo(e, f.Name, f.Type)

	var fieldVersion uint8
	if f.Opts != nil {
		fieldVersion = f.Opts.Version
	}

	if fieldVersion == 0 {
		streamType(e, b, f.Type)
		return
	}

	b.VersionGate(
		e,
		fieldVersion,
		func() {
			streamType(e, b, f.Type)
		},
		func() {
			b.StreamDefault(e, f.Type, f.Opts)
		})
}

func streamType(e *Emit, b StreamBackend, t types.GenType) {
	if !t.IsNilable() {
		streamRaw(e, b, t)
		return
	}

	b.StreamNilGuard(
		e,
		func() {
			b.YieldNil(e)
		},
		func() {
			streamRaw(e, b, t)
		})
}

func streamRaw(e *Emit, b StreamBackend, t types.GenType) {
	code := t.Code()

	switch code {
	case types.TypeSlice:
		st := t.(*types.SliceType)
		b.StreamSlice(e, st, func(index string) {
			b.YieldIndexed(e, index, decodeValue(e, b, st.InnerType))
		})

	case types.TypeMap:
		mt := t.(*types.MapType)
		b.StreamMap(e, mt, func() {
			key := decodeValue(e, b, mt.KeyType)
			value := decodeValue(e, b, mt.ValueType)
			b.YieldIndexed(e, key, value)
		})

	case types.TypeAlias:
		// An alias of a collection streams its underlying collection element by
		// element; any other alias is fully read and yielded as a single value.
		inner := t.(*types.AliasType).Alias
		innerCode := inner.Code()

		if innerCode == types.TypeSlice || innerCode == types.TypeMap {
			streamRaw(e, b, inner)
			return
		}
		streamSingle(e, b, t)

	default:
		streamSingle(e, b, t)
	}
}

// streamSingle reads a whole value into a new variable and yields it to the consumer.
func streamSingle(e *Emit, b StreamBackend, t types.GenType) {
	v := e.NextVar()
	b.DeclareVar(e, t, v)
	decodeRaw(e, b, t, v)
	b.Yield(e, v)
}
