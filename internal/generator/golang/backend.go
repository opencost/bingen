package golang

import (
	"fmt"
	"strings"

	"github.com/opencost/bingen/internal/generator/core"
	"github.com/opencost/bingen/internal/types"
)

// GoBackend renders Go marshal/unmarshal statements for the core codec walk,
// mirroring internal/generator/templates/go/{marshaller,unmarshaller}.go.tmpl.
// The generated file is run through go/format + goimports, so this backend does
// not track imports.
type GoBackend struct {
	// streaming selects the error-handling form: streaming decoders run inside a
	// range-over-func that returns nothing, so errors set stream.err and return
	// bare rather than `return err`.
	streaming bool
}

func NewGoBackend() *GoBackend {
	return &GoBackend{}
}

func NewGoStreamBackend() *GoBackend {
	return &GoBackend{
		streaming: true,
	}
}

// primitive type to write methods on the buffer
var goWrite = map[uint8]string{
	types.TypeBool:    "WriteBool",
	types.TypeInt:     "WriteInt",
	types.TypeInt8:    "WriteInt8",
	types.TypeInt16:   "WriteInt16",
	types.TypeInt32:   "WriteInt32",
	types.TypeInt64:   "WriteInt64",
	types.TypeUInt:    "WriteUInt",
	types.TypeUInt8:   "WriteUInt8",
	types.TypeUInt16:  "WriteUInt16",
	types.TypeUInt32:  "WriteUInt32",
	types.TypeUInt64:  "WriteUInt64",
	types.TypeFloat32: "WriteFloat32",
	types.TypeFloat64: "WriteFloat64",
}

// primitive type to read methods on the buffer
var goRead = map[uint8]string{
	types.TypeBool:    "ReadBool",
	types.TypeInt:     "ReadInt",
	types.TypeInt8:    "ReadInt8",
	types.TypeInt16:   "ReadInt16",
	types.TypeInt32:   "ReadInt32",
	types.TypeInt64:   "ReadInt64",
	types.TypeUInt:    "ReadUInt",
	types.TypeUInt8:   "ReadUInt8",
	types.TypeUInt16:  "ReadUInt16",
	types.TypeUInt32:  "ReadUInt32",
	types.TypeUInt64:  "ReadUInt64",
	types.TypeFloat32: "ReadFloat32",
	types.TypeFloat64: "ReadFloat64",
}

// goDefaultCast maps a numeric primitive to its cast form, reproducing the
// template's ToDefaultValue.
var goDefaultCast = map[uint8]string{
	types.TypeInt:     "int",
	types.TypeInt8:    "int8",
	types.TypeInt16:   "int16",
	types.TypeInt32:   "int32",
	types.TypeInt64:   "int64",
	types.TypeUInt:    "uint",
	types.TypeUInt8:   "uint8",
	types.TypeUInt16:  "uint16",
	types.TypeUInt32:  "uint32",
	types.TypeUInt64:  "uint64",
	types.TypeFloat32: "float32",
	types.TypeFloat64: "float64",
}

func (g *GoBackend) TypeName(t types.GenType, boxed bool) string {
	return t.TypeName()
}

func (g *GoBackend) EncodeFieldSource(field string) string {
	return "target." + field
}

func (g *GoBackend) CommitField(e *core.Emit, field, valueExpr string) {
	e.Linef("target.%s = %s", field, valueExpr)
}

func (g *GoBackend) Default(t types.GenType, opts *types.FieldOpts) string {
	if t.IsNilable() {
		return "nil"
	}

	def := ""
	if opts != nil {
		def = opts.Default
	}

	code := t.Code()

	// bool default
	if code == types.TypeBool {
		if def == "" {
			return "false"
		}
		return def
	}

	// number primitive default
	if types.IsNumericalPrimitive(code) {
		if def == "" {
			def = "0"
		}

		return fmt.Sprintf("%s(%s)", goDefaultCast[code], def)
	}

	// string primitive default
	if code == types.TypeString {
		return fmt.Sprintf("%q", def)
	}

	// all other type defaults
	if def != "" {
		return def
	}
	return t.Name() + "{}"
}

// Comment writes a comment to the generated output with the provided text
func (g *GoBackend) Comment(e *core.Emit, text string) {
	for line := range strings.Lines(text) {
		e.Linef("// %s", line)
	}
}

//--------------------------------------------------------------------------
//  Encode
//--------------------------------------------------------------------------

func (g *GoBackend) WriteNilFlag(e *core.Emit, present bool) {
	if present {
		e.Line("buff.WriteUInt8(uint8(1)) // write non-nil byte")
	} else {
		e.Line("buff.WriteUInt8(uint8(0)) // write nil byte")
	}
}

func (g *GoBackend) NilGuard(e *core.Emit, target string, whenNil, whenPresent func()) {
	e.Linef("if %s == nil {", target)
	e.Indented(whenNil)
	e.Line("} else {")
	e.Indented(whenPresent)
	e.Line("}")
}

func (g *GoBackend) WritePrimitive(e *core.Emit, t types.GenType, target string) {
	if t.IsPtr() {
		target = "*" + target
	}

	e.Linef("buff.%s(%s)", goWrite[t.Code()], target)
}

func (g *GoBackend) WriteString(e *core.Emit, t types.GenType, target string) {
	if t.IsPtr() {
		target = "*" + target
	}

	idx := e.NextVar()

	e.Line("if ctx.IsStringTable() {")
	e.Indented(func() {
		e.Linef("%s := ctx.Table.AddOrGet(%s)", idx, target)
		e.Linef("buff.WriteInt(%s) // write table index", idx)
	})
	e.Line("} else {")
	e.Indented(func() {
		e.Linef("buff.WriteString(%s) // write string", target)
	})
	e.Line("}")
}

func (g *GoBackend) WriteSlice(e *core.Emit, t *types.SliceType, target string, body func(elem string)) {
	loop := e.NextVar()

	ds := core.BeginDebugScope(g, e, "write", "slice", t.Name())
	defer ds.End()

	e.Linef("buff.WriteInt(len(%s)) // slice length", target)
	e.Linef("for %s := range %s {", loop, target)
	e.Indented(func() {
		body(fmt.Sprintf("%s[%s]", target, loop))
	})
	e.Line("}")
}

func (g *GoBackend) WriteMap(e *core.Emit, t *types.MapType, target string, body func(key, value string)) {
	k := e.NextVar()
	v := e.NextVar()

	ds := core.BeginDebugScope(g, e, "write", "map", t.TypeName())
	defer ds.End()

	e.Linef("buff.WriteInt(len(%s)) // map length", target)
	e.Linef("for %s, %s := range %s {", k, v, target)
	e.Indented(func() {
		body(k, v)
	})
	e.Line("}")
}

func (g *GoBackend) WriteStruct(e *core.Emit, t types.GenType, target string) {
	err := e.NextVar()

	ds := core.BeginDebugScope(g, e, "write", "struct", t.Name())
	defer ds.End()

	e.Line("buff.WriteInt(0) // [compatibility, unused]")
	e.Linef("%s := %s.MarshalBinaryWithContext(ctx)", err, target)
	g.returnIfErr(e, err)
}

func (g *GoBackend) WriteInterface(e *core.Emit, t types.GenType, target string) {
	iv := e.NextVar()
	mv := e.NextVar()
	ok := e.NextVar()
	err := e.NextVar()

	ds := core.BeginDebugScope(g, e, "write", "interface", t.Name())
	defer ds.End()

	e.Linef("%s := reflect.ValueOf(%s).Interface()", iv, target)
	e.Linef("%s, %s := %s.(BinEncoder)", mv, ok, iv)
	e.Linef("if !%s {", ok)
	e.Indented(func() {
		e.Linef(`return fmt.Errorf("type: %%s does not implement %%s.BinEncoder", typeToString(%s), GeneratorPackageName)`, target)
	})
	e.Line("}")

	e.Linef("buff.WriteString(typeToString(%s))", target)
	e.Line("buff.WriteInt(0) // [compatibility, unused]")
	e.Linef("%s := %s.MarshalBinaryWithContext(ctx)", err, mv)
	g.returnIfErr(e, err)
}

func (g *GoBackend) WriteExternal(e *core.Emit, t types.GenType, target string) bool {
	// Go handles any unresolved reference generically via its BinaryMarshaler.
	data := e.NextVar()
	err := e.NextVar()

	ds := core.BeginDebugScope(g, e, "write", "reference", t.Name())
	defer ds.End()

	e.Linef("%s, %s := %s.MarshalBinary()", data, err, target)
	g.returnIfErr(e, err)
	e.Linef("buff.WriteInt(len(%s))", data)
	e.Linef("buff.WriteBytes(%s)", data)

	return true
}

func (g *GoBackend) EncodeAlias(e *core.Emit, at *types.AliasType, target string, body func(underlying string)) {
	inner := target
	if at.IsPtr() {
		inner = "*" + target
	}

	var casted string
	if at.Alias.IsPtr() {
		casted = fmt.Sprintf("((%s)(%s))", at.Alias.TypeName(), inner)
	} else {
		casted = fmt.Sprintf("%s(%s)", at.Alias.TypeName(), inner)
	}

	ds := core.BeginDebugScope(g, e, "write", "alias", at.Name())
	defer ds.End()

	body(casted)
}

//--------------------------------------------------------------------------
//  Decode
//--------------------------------------------------------------------------

func (g *GoBackend) DeclareVar(e *core.Emit, t types.GenType, v string) {
	e.Linef("var %s %s", v, t.TypeName())
}

func (g *GoBackend) VersionGate(e *core.Emit, fieldVersion uint8, whenCompatible, setDefault func()) {
	e.Linef("if uint8(%d) <= version {", fieldVersion)
	e.Indented(whenCompatible)
	e.Line("} else {")
	e.Indented(setDefault)
	e.Line("}")
}

func (g *GoBackend) ReadNilGuard(e *core.Emit, v string, whenPresent func()) {
	e.Line("if buff.ReadUInt8() == uint8(0) {")
	e.Indented(func() {
		e.Linef("%s = nil", v)
	})
	e.Line("} else {")
	e.Indented(whenPresent)
	e.Line("}")
}

func (g *GoBackend) ReadPrimitive(e *core.Emit, t types.GenType, v string) {
	if t.IsPtr() {
		tmp := e.NextVar()

		e.Linef("%s := buff.%s()", tmp, goRead[t.Code()])
		e.Linef("%s = &%s", v, tmp)
		return
	}

	e.Linef("%s = buff.%s()", v, goRead[t.Code()])
}

func (g *GoBackend) ReadString(e *core.Emit, t types.GenType, v string) {
	dest := v

	if t.IsPtr() {
		tmp := e.NextVar()

		e.Linef("var %s string", tmp)
		g.readStringInto(e, tmp)
		e.Linef("%s = &%s", v, tmp)
		return
	}

	g.readStringInto(e, dest)
}

func (g *GoBackend) readStringInto(e *core.Emit, dest string) {
	idx := e.NextVar()

	e.Line("if ctx.IsStringTable() {")
	e.Indented(func() {
		e.Linef("%s := buff.ReadInt() // read string index", idx)
		e.Linef("%s = ctx.Table.At(%s)", dest, idx)
	})
	e.Line("} else {")
	e.Indented(func() {
		e.Linef("%s = buff.ReadString() // read string", dest)
	})
	e.Line("}")
}

func (g *GoBackend) ReadSlice(e *core.Emit, t *types.SliceType, v string, readElem func() string) {
	ln := e.NextVar()
	loop := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "slice", t.Name())
	defer ds.End()

	e.Linef("%s := buff.ReadInt() // slice len", ln)
	e.Linef("%s = make(%s, %s)", v, t.TypeName(), ln)
	e.Linef("for %s := range %s {", loop, ln)
	e.Indented(func() {
		e.Linef("%s[%s] = %s", v, loop, readElem())
	})
	e.Line("}")
}

func (g *GoBackend) ReadMap(e *core.Emit, t *types.MapType, v string, readPair func() (key, value string)) {
	ln := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "map", t.Name())
	defer ds.End()

	e.Linef("%s := buff.ReadInt() // map len ", ln)
	e.Linef("%s = make(%s, %s)", v, t.TypeName(), ln)
	e.Linef("for range %s {", ln)
	e.Indented(func() {
		key, value := readPair()
		e.Linef("%s[%s] = %s", v, key, value)
	})
	e.Line("}")
}

func (g *GoBackend) ReadStruct(e *core.Emit, t types.GenType, v string) {
	ub := e.NextVar()
	err := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "struct", t.Name())
	defer ds.End()

	e.Linef("%s := new(%s)", ub, t.Name())
	e.Line("buff.ReadInt() // [compatibility, unused]")
	e.Linef("%s := %s.UnmarshalBinaryWithContext(ctx)", err, ub)
	g.returnIfErr(e, err)
	if t.IsPtr() {
		e.Linef("%s = %s", v, ub)
	} else {
		e.Linef("%s = *%s", v, ub)
	}
}

func (g *GoBackend) ReadInterface(e *core.Emit, t types.GenType, v string) {
	tn := e.NextVar()
	nm := e.NextVar()
	ub := e.NextVar()
	ok := e.NextVar()
	err := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "interface", t.Name())
	defer ds.End()

	e.Linef("%s := buff.ReadString()", tn)
	e.Linef("_, %s, _ := resolveType(%s)", nm, tn)
	e.Linef("if _, ok := typeMap[%s]; !ok {", nm)
	e.Indented(func() {
		g.fail(e, fmt.Sprintf(`fmt.Errorf("unknown type: %%s", %s)`, nm))
	})
	e.Line("}")
	e.Linef("%s, %s := reflect.New(typeMap[%s]).Interface().(BinDecoder)", ub, ok, nm)
	e.Linef("if !%s {", ok)
	e.Indented(func() {
		g.fail(e, fmt.Sprintf(`fmt.Errorf("type: %%s does not implement %%s.BinDecoder.", %s, GeneratorPackageName)`, nm))
	})
	e.Line("}")
	e.Line("buff.ReadInt() // [compatibility, unused]")
	e.Linef("%s := %s.UnmarshalBinaryWithContext(ctx)", err, ub)
	g.returnIfErr(e, err)
	e.Linef("%s = %s.(%s)", v, ub, t.Name())
}

func (g *GoBackend) ReadExternal(e *core.Emit, t types.GenType, v string) bool {
	ub := e.NextVar()
	ln := e.NextVar()
	ba := e.NextVar()
	err := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "reference", t.Name())
	defer ds.End()

	e.Linef("%s := new(%s)", ub, t.Name())
	e.Linef("%s := buff.ReadInt() // byte array length", ln)
	e.Linef("%s := buff.ReadBytes(%s)", ba, ln)
	e.Linef("%s := %s.UnmarshalBinary(%s)", err, ub, ba)
	g.returnIfErr(e, err)
	if t.IsPtr() {
		e.Linef("%s = %s", v, ub)
	} else {
		e.Linef("%s = *%s", v, ub)
	}

	return true
}

func (g *GoBackend) DecodeAlias(e *core.Emit, at *types.AliasType, v string, body func(underlying string)) {
	tmp := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "alias", at.Name())
	defer ds.End()

	e.Linef("var %s %s", tmp, at.Alias.TypeName())

	body(tmp)

	if at.IsPtr() {
		cast := e.NextVar()
		e.Linef("%s := %s(%s)", cast, at.Name(), tmp)
		e.Linef("%s = &%s", v, cast)
	} else {
		e.Linef("%s = %s(%s)", v, at.Name(), tmp)
	}
}

func (g *GoBackend) returnIfErr(e *core.Emit, err string) {
	e.Linef("if %s != nil {", err)
	e.Indented(func() {
		g.fail(e, err)
	})
	e.Line("}")
}

// fail emits an early exit with an error: `return err` normally, or
// `stream.err = err; return` inside a streaming range-over-func.
func (g *GoBackend) fail(e *core.Emit, errExpr string) {
	if g.streaming {
		e.Linef("stream.err = %s", errExpr)
		e.Line("return")
		return
	}

	e.Linef("return %s", errExpr)
}

//--------------------------------------------------------------------------
//  Streaming
//--------------------------------------------------------------------------

func (g *GoBackend) FieldInfo(e *core.Emit, name string, t types.GenType) {
	e.Line("fi = bstream.BingenFieldInfo{")
	e.Indented(func() {
		e.Linef("Type: reflect.TypeFor[%s](),", t.TypeName())
		e.Linef("Name: %q,", name)
	})
	e.Line("}")
}

func (g *GoBackend) Yield(e *core.Emit, value string) {
	g.yield(e, fmt.Sprintf("bstream.SingleV(%s)", value))
}

func (g *GoBackend) YieldNil(e *core.Emit) {
	g.yield(e, "nil")
}

func (g *GoBackend) YieldIndexed(e *core.Emit, index, value string) {
	g.yield(e, fmt.Sprintf("bstream.PairV(%s, %s)", index, value))
}

func (g *GoBackend) yield(e *core.Emit, valueExpr string) {
	e.Linef("if !yield(fi, %s) {", valueExpr)
	e.Indented(func() {
		e.Line("return")
	})
	e.Line("}")
}

func (g *GoBackend) StreamNilGuard(e *core.Emit, whenNil, whenPresent func()) {
	e.Line("if buff.ReadUInt8() == uint8(0) {")
	e.Indented(whenNil)
	e.Line("} else {")
	e.Indented(whenPresent)
	e.Line("}")
}

func (g *GoBackend) StreamSlice(e *core.Emit, t *types.SliceType, body func(index string)) {
	ln := e.NextVar()
	loop := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "streaming-slice", t.Name())
	defer ds.End()

	e.Linef("%s := buff.ReadInt() // slice len", ln)
	e.Linef("for %s := range %s {", loop, ln)
	e.Indented(func() {
		body(loop)
	})
	e.Line("}")
}

func (g *GoBackend) StreamMap(e *core.Emit, t *types.MapType, body func()) {
	ln := e.NextVar()

	ds := core.BeginDebugScope(g, e, "read", "streaming-map", t.Name())
	defer ds.End()

	e.Linef("%s := buff.ReadInt() // map len", ln)
	e.Linef("for range %s {", ln)
	e.Indented(body)
	e.Line("}")
}

func (g *GoBackend) StreamDefault(e *core.Emit, t types.GenType, opts *types.FieldOpts) {
	if t.IsNilable() {
		g.YieldNil(e)
		return
	}

	if t.Code() <= types.TypeString {
		v := e.NextVar()
		e.Linef("var %s %s = %s // default", v, t.TypeName(), g.Default(t, opts))
		g.Yield(e, v)
		return
	}

	def := ""
	if opts != nil {
		def = opts.Default
	}

	if def != "" {
		v := e.NextVar()
		e.Linef("var %s %s = %s // default", v, t.TypeName(), def)
		g.Yield(e, v)
	}
}
