package java

import (
	"fmt"
	"strings"

	"github.com/opencost/bingen/internal/generator/core"
	"github.com/opencost/bingen/internal/types"
)

// JavaBackend renders Java syntax for the core codec walk. It holds the runtime
// package so it can import runtime codec classes (ie:  GoTime).
type JavaBackend struct {
	runtimePkg string
}

func NewJavaBackend(runtimePkg string) *JavaBackend {
	return &JavaBackend{
		runtimePkg: runtimePkg,
	}
}

func (j *JavaBackend) TypeName(t types.GenType, boxed bool) string {
	return javaType(t, boxed)
}

func (j *JavaBackend) EncodeFieldSource(field string) string {
	return fmt.Sprintf("value.%s()", JavaFieldName(field))
}

func (j *JavaBackend) CommitField(e *core.Emit, field, valueExpr string) {
	e.Linef("builder.%s(%s);", JavaFieldName(field), valueExpr)
}

func (j *JavaBackend) Default(t types.GenType, opts *types.FieldOpts) string {
	return javaDefault(t, opts)
}

// Comment writes a comment to the generated output with the provided text
func (j *JavaBackend) Comment(e *core.Emit, text string) {
	for line := range strings.Lines(text) {
		e.Linef("// %s", line)
	}
}

// useExternal records the imports needed to reference an external type's runtime
// codec class and returns that class name.
func (j *JavaBackend) useExternal(e *core.Emit, ext externalType) string {
	for _, imp := range ext.typeImports {
		e.Import(imp)
	}
	e.Import(j.runtimePkg + "." + ext.runtimeClass)
	return ext.runtimeClass
}

//--------------------------------------------------------------------------
//  Encode
//--------------------------------------------------------------------------

func (j *JavaBackend) WriteNilFlag(e *core.Emit, present bool) {
	if present {
		e.Line("buff.writeUInt8(1); // write non-nil byte")
	} else {
		e.Line("buff.writeUInt8(0); // write nil byte")
	}
}

func (j *JavaBackend) NilGuard(e *core.Emit, target string, whenNil, whenPresent func()) {
	e.Linef("if (%s == null) {", target)
	e.Indented(whenNil)
	e.Line("} else {")
	e.Indented(whenPresent)
	e.Line("}")
}

func (j *JavaBackend) WritePrimitive(e *core.Emit, t types.GenType, target string) {
	e.Linef("buff.%s(%s);", primitives[t.Code()].write, target)
}

func (j *JavaBackend) WriteString(e *core.Emit, t types.GenType, target string) {
	e.Line("if (ctx.isStringTable()) {")
	e.Indented(func() {
		e.Linef("buff.writeInt(ctx.table().addOrGet(%s)); // string table index", target)
	})
	e.Line("} else {")
	e.Indented(func() {
		e.Linef("buff.writeString(%s); // string", target)
	})
	e.Line("}")
}

func (j *JavaBackend) WriteSlice(e *core.Emit, t *types.SliceType, target string, body func(elem string)) {
	elem := e.NextVar()

	e.Linef("buff.writeInt(%s.size()); // slice length", target)
	e.Linef("for (%s %s : %s) {", javaType(t.InnerType, true), elem, target)
	e.Indented(func() {
		body(elem)
	})
	e.Line("}")
}

func (j *JavaBackend) WriteMap(e *core.Emit, t *types.MapType, target string, body func(key, value string)) {
	e.Import("java.util.Map")

	entry := e.NextVar()

	e.Linef("buff.writeInt(%s.size()); // map length", target)
	e.Linef(
		"for (Map.Entry<%s, %s> %s : %s.entrySet()) {",
		javaType(t.KeyType, true),
		javaType(t.ValueType, true),
		entry,
		target)
	e.Indented(func() {
		body(entry+".getKey()", entry+".getValue()")
	})
	e.Line("}")
}

func (j *JavaBackend) WriteStruct(e *core.Emit, t types.GenType, target string) {
	e.Line("buff.writeInt(0); // [compatibility, unused]")
	e.Linef("%sEncoder.INSTANCE.encode(%s, ctx);", t.Name(), target)
}

func (j *JavaBackend) WriteInterface(e *core.Emit, t types.GenType, target string) {
	e.Linef("TypeRegistry.encode(%s, ctx);", target)
}

func (j *JavaBackend) WriteExternal(e *core.Emit, t types.GenType, target string) bool {
	ext, ok := externalTypes[t.Name()]
	if !ok {
		return false
	}

	e.Linef("%s.encode(%s, buff);", j.useExternal(e, ext), target)
	return true
}

// EncodeAlias for the Java backend flattens aliases to their underlying type, so the target is
// passed through with no cast.
func (j *JavaBackend) EncodeAlias(e *core.Emit, at *types.AliasType, target string, body func(underlying string)) {
	body(target)
}

//--------------------------------------------------------------------------
//  Decode
//--------------------------------------------------------------------------

func (j *JavaBackend) DeclareVar(e *core.Emit, t types.GenType, v string) {
	e.Linef("%s %s;", javaType(t, false), v)
}

func (j *JavaBackend) VersionGate(e *core.Emit, fieldVersion uint8, whenCompatible, setDefault func()) {
	e.Linef("if (%d <= version) {", fieldVersion)
	e.Indented(whenCompatible)
	e.Line("} else {")
	e.Indented(setDefault)
	e.Line("}")
}

func (j *JavaBackend) ReadNilGuard(e *core.Emit, v string, whenPresent func()) {
	e.Line("if (buff.readUInt8() == 0) {")
	e.Indented(func() {
		e.Linef("%s = null;", v)
	})
	e.Line("} else {")
	e.Indented(whenPresent)
	e.Line("}")
}

func (j *JavaBackend) ReadPrimitive(e *core.Emit, t types.GenType, v string) {
	e.Linef("%s = buff.%s();", v, primitives[t.Code()].read)
}

func (j *JavaBackend) ReadString(e *core.Emit, t types.GenType, v string) {
	e.Line("if (ctx.isStringTable()) {")
	e.Indented(func() {
		e.Linef("%s = ctx.table().at(buff.readInt()); // string table index", v)
	})
	e.Line("} else {")
	e.Indented(func() {
		e.Linef("%s = buff.readString(); // read string", v)
	})
	e.Line("}")
}

func (j *JavaBackend) ReadSlice(e *core.Emit, t *types.SliceType, v string, readElem func() string) {
	e.Import("java.util.ArrayList")
	e.Import("java.util.List")

	lenVar := e.NextVar()
	idx := e.NextVar()

	e.Linef("int %s = buff.readInt(); // slice length", lenVar)
	e.Linef("%s = new ArrayList<>(%s);", v, lenVar)
	e.Linef("for (int %s = 0; %s < %s; %s++) {", idx, idx, lenVar, idx)
	e.Indented(func() {
		e.Linef("%s.add(%s);", v, readElem())
	})
	e.Line("}")
}

func (j *JavaBackend) ReadMap(e *core.Emit, t *types.MapType, v string, readPair func() (key, value string)) {
	e.Import("java.util.LinkedHashMap")
	e.Import("java.util.Map")

	lenVar := e.NextVar()
	idx := e.NextVar()

	e.Linef("int %s = buff.readInt(); // map length", lenVar)
	e.Linef("%s = new LinkedHashMap<>(%s);", v, lenVar)
	e.Linef("for (int %s = 0; %s < %s; %s++) {", idx, idx, lenVar, idx)
	e.Indented(func() {
		key, value := readPair()
		e.Linef("%s.put(%s, %s);", v, key, value)
	})
	e.Line("}")
}

func (j *JavaBackend) ReadStruct(e *core.Emit, t types.GenType, v string) {
	e.Line("buff.readInt(); // [compatibility, unused]")
	e.Linef("%s = %sDecoder.INSTANCE.decode(ctx);", v, t.Name())
}

func (j *JavaBackend) ReadInterface(e *core.Emit, t types.GenType, v string) {
	name := e.NextVar()
	e.Linef("String %s = buff.readString();", name)
	e.Linef("%s = (%s) TypeRegistry.decode(%s, ctx);", v, javaType(t, false), name)
}

func (j *JavaBackend) ReadExternal(e *core.Emit, t types.GenType, v string) bool {
	ext, ok := externalTypes[t.Name()]
	if !ok {
		return false
	}

	e.Linef("%s = %s.decode(buff);", v, j.useExternal(e, ext))
	return true
}

// DecodeAlias for the Java backend flattens aliases to their underlying type, so v is read into
// directly with no conversion.
func (j *JavaBackend) DecodeAlias(e *core.Emit, at *types.AliasType, v string, body func(underlying string)) {
	body(v)
}

//--------------------------------------------------------------------------
//  Streaming
//--------------------------------------------------------------------------

func (j *JavaBackend) FieldInfo(e *core.Emit, name string, t types.GenType) {
	e.Linef("fi = new BingenFieldInfo(%q);", name)
}

func (j *JavaBackend) Yield(e *core.Emit, value string) {
	j.accept(e, fmt.Sprintf("BingenValue.single(%s)", value))
}

func (j *JavaBackend) YieldNil(e *core.Emit) {
	j.accept(e, "null")
}

func (j *JavaBackend) YieldIndexed(e *core.Emit, index, value string) {
	j.accept(e, fmt.Sprintf("BingenValue.pair(%s, %s)", index, value))
}

func (j *JavaBackend) accept(e *core.Emit, valueExpr string) {
	e.Linef("if (!consumer.accept(fi, %s)) {", valueExpr)
	e.Indented(func() { e.Line("return;") })
	e.Line("}")
}

func (j *JavaBackend) StreamNilGuard(e *core.Emit, whenNil, whenPresent func()) {
	e.Line("if (buff.readUInt8() == 0) {")
	e.Indented(whenNil)
	e.Line("} else {")
	e.Indented(whenPresent)
	e.Line("}")
}

func (j *JavaBackend) StreamSlice(e *core.Emit, t *types.SliceType, body func(index string)) {
	lenVar := e.NextVar()
	idx := e.NextVar()
	e.Linef("int %s = buff.readInt();", lenVar)
	e.Linef("for (int %s = 0; %s < %s; %s++) {", idx, idx, lenVar, idx)
	e.Indented(func() { body(idx) })
	e.Line("}")
}

func (j *JavaBackend) StreamMap(e *core.Emit, t *types.MapType, body func()) {
	lenVar := e.NextVar()
	idx := e.NextVar()
	e.Linef("int %s = buff.readInt();", lenVar)
	e.Linef("for (int %s = 0; %s < %s; %s++) {", idx, idx, lenVar, idx)
	e.Indented(body)
	e.Line("}")
}

func (j *JavaBackend) StreamDefault(e *core.Emit, t types.GenType, opts *types.FieldOpts) {
	if t.IsNilable() {
		j.YieldNil(e)
		return
	}

	v := e.NextVar()
	e.Linef("%s %s = %s;", javaType(t, false), v, javaDefault(t, opts))
	j.Yield(e, v)
}

// javaDefault returns the Java expression assigned to a field whose version
// exceeds the payload version. Nilable types default to null and primitives to
// their zero value (or the raw annotation default when present); go-style cast
// defaults are not translated.
func javaDefault(t types.GenType, opts *types.FieldOpts) string {
	if t.IsNilable() {
		return "null"
	}

	def := ""
	if opts != nil {
		def = opts.Default
	}

	switch t.Code() {
	case types.TypeBool:
		if def == "" {
			return "false"
		}
		return def

	case types.TypeString:
		return fmt.Sprintf("%q", def)

	case types.TypeAlias:
		return javaDefault(t.(*types.AliasType).Alias, opts)

	case types.TypeStruct, types.TypeInterface, types.TypeReference:
		return "null"

	default:
		if def == "" {
			return "0"
		}

		return def
	}
}
