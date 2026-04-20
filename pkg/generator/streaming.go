package generator

import (
	"slices"

	"github.com/opencost/bingen/pkg/generator/vars"
	"github.com/opencost/bingen/pkg/types"
)

func GetStreamableTypes(ts []types.GenType) []*types.StructType {
	streamTypes := []*types.StructType{}

	for _, t := range ts {
		if st, ok := t.(*types.StructType); ok {
			if st.Opts.IsStreamable {
				streamTypes = append(streamTypes, st)
			}
		}
	}

	return streamTypes
}

func StreamYieldNil(ctx Scope, fiVar string) {
	inner := ctx.Push("if !yield(%s, nil) {", fiVar)
	inner.Writeln("return")
	inner.Pop("}")
}

func StreamYieldSingle(ctx Scope, fiVar string, valueVar string) {
	inner := ctx.Push("if !yield(%s, singleV(%s)) {", fiVar, valueVar)
	inner.Writeln("return")
	inner.Pop("}")
}

func StreamYieldIndexed(ctx Scope, fiVar string, valueVar string, indexVar string) {
	inner := ctx.Push("if !yield(%s, pairV(%s, %s)) {", fiVar, indexVar, valueVar)
	inner.Writeln("return")
	inner.Pop("}")
}

func StreamErrorHandler(errCheck Scope, newErr string) {
	errCheck.Writeln("stream.err = %s", newErr)
	errCheck.Writeln("return")
}

func WriteStreamReader(ctx GeneratorContext, t types.GenType) {
	ctx.Writeln(BingenStreamFormatCode, slices.Repeat([]any{t.Name()}, 12)...)

	WriteStreamMethod(ctx, t)
}

func WriteStreamMethod(ctx GeneratorContext, t types.GenType) {
	st := t.(*types.StructType)

	ctx.Writeln("// Stream returns the iterator which will stream each field of the target type.")
	funcScope := ctx.PushScope("func (stream *%sStream) Stream() iter.Seq2[BingenFieldInfo, *BingenValue] {", t.Name())
	scope := funcScope.Push("return func(yield func(BingenFieldInfo, *BingenValue) bool) {")

	scope.Writeln("var fi BingenFieldInfo\n")

	scope.Writeln("ctx := stream.ctx")
	scope.Writeln("buff := ctx.Buffer")
	scope.Writeln("version := buff.ReadUInt8()\n")

	cond := scope.Push("if version > %s {", ctx.VersionSetConst())
	cond.Writeln("stream.err = fmt.Errorf(\"Invalid Version Unmarshaling %s. Expected %%d or less, got %%d\", %s, version)", t.Name(), ctx.VersionSetConst())
	cond.Writeln("return")
	cond.Pop("}\n")

	StreamReadFullStructType(ctx, st)

	scope.Pop("}")
	funcScope.Pop("}\n")
}

func StreamReadFullStructType(ctx GeneratorContext, t *types.StructType) {
	for _, f := range t.Fields {
		var fieldVersion uint8 = 0
		if f.Opts != nil {
			fieldVersion = f.Opts.Version
		}

		fiScope := ctx.PushScope("fi = BingenFieldInfo{")
		fiScope.Writeln("Type: reflect.TypeFor[%s](),", f.Type.Name())
		fiScope.Writeln("Name: \"%s\",", f.Name)
		fiScope.Pop("}\n")

		if fieldVersion != 0 {
			ctx.Writeln("// field version check")
			cond := ctx.PushScope("if uint8(%d) <= version {", fieldVersion)
			StreamReadType(ctx, f.Type)
			cond.PopPush("} else {")
			StreamSetTypeDefault(ctx, f.Type, f.Opts)
			cond.Pop("}\n")
		} else {
			StreamReadType(ctx, f.Type)
		}
	}
}

func StreamReadType(ctx GeneratorContext, t types.GenType) {
	if t.IsNilable() {
		cond := ctx.PushScope("if buff.ReadUInt8() == uint8(0) {")
		StreamYieldNil(cond, "fi")
		cond.PopPush("} else {")

		StreamReadTypeRaw(ctx, t)

		cond.Pop("}")
		return
	}

	StreamReadTypeRaw(ctx, t)
}

func StreamSetTypeDefault(ctx GeneratorContext, t types.GenType, opts *types.FieldOpts) {
	if t.IsNilable() {
		StreamYieldNil(ctx.AsScope(), "fi")
		return
	}

	var defaultValue string
	if opts == nil {
		defaultValue = ""
	} else {
		defaultValue = opts.Default
	}

	switch rt := t.(type) {
	case *types.BasicType:
		target := ctx.NextVar()
		ctx.Writeln("var %s %s = %s // default", target, t.TypeName(), ToDefaultValue(rt.Code(), defaultValue))
		StreamYieldSingle(ctx.AsScope(), "fi", target)
	default:
		if defaultValue == "" {
			return
		}
		target := ctx.NextVar()
		ctx.Writeln("var %s %s = %s // default", target, t.TypeName(), defaultValue)
		StreamYieldSingle(ctx.AsScope(), "fi", target)
	}
}

func StreamReadTypeRaw(ctx GeneratorContext, t types.GenType) {
	switch rt := t.(type) {
	case *types.BasicType:
		target := ctx.NextVar()
		ReadBasicType(ctx, rt, vars.AsTarget(target), true)
		StreamYieldSingle(ctx.AsScope(), "fi", target)
	case *types.SliceType:
		StreamReadArrayType(ctx, rt)
	case *types.MapType:
		StreamReadMapType(ctx, rt)
	case *types.StructType:
		target := ctx.NextVar()
		ReadStructType(ctx, rt, vars.AsTarget(target), true)
		StreamYieldSingle(ctx.AsScope(), "fi", target)
	case *types.ReferenceType:
		resolvedType := rt.Resolve()
		if resolvedType != nil {
			StreamReadTypeRaw(ctx, resolvedType)
		} else {
			target := ctx.NextVar()
			ReadReferenceType(ctx, rt, vars.AsTarget(target), true)
			StreamYieldSingle(ctx.AsScope(), "fi", target)
		}
	case *types.AliasType:
		StreamReadAliasType(ctx, rt)
	case *types.InterfaceType:
		target := ctx.NextVar()
		ReadInterfaceType(ctx, rt, vars.AsTarget(target), true)
		StreamYieldSingle(ctx.AsScope(), "fi", target)
	}
}

func StreamReadArrayType(ctx GeneratorContext, at *types.SliceType) {
	d := ctx.PushDebugRead("streaming-slice", at.Name())
	defer d.End()

	lenVar := ctx.NextVar()
	loopVar := ctx.NextLoopVar()
	innerVar := ctx.NextVar()

	ctx.Writeln("%s := buff.ReadInt() // array len", lenVar)
	loop := ctx.PushScope("for %s := 0; %s < %s; %s++ {", loopVar, loopVar, lenVar, loopVar)
	ReadType(ctx, at.InnerType, vars.AsTarget(innerVar), true)
	StreamYieldIndexed(loop, "fi", innerVar, loopVar)
	loop.Pop("}")
}

func StreamReadMapType(ctx GeneratorContext, mt *types.MapType) {
	d := ctx.PushDebugRead("streaming-map", mt.Name())
	defer d.End()

	lenVar := ctx.NextVar()
	idxVar := ctx.NextLoopVar()
	keyVar := ctx.NextMapVar()
	valVar := ctx.NextMapVar()

	ctx.Writeln("%s := buff.ReadInt() // map len", lenVar)
	loop := ctx.PushScope("for %s := 0; %s < %s; %s++ {", idxVar, idxVar, lenVar, idxVar)
	ReadType(ctx, mt.KeyType, vars.AsTarget(keyVar), true)
	ReadType(ctx, mt.ValueType, vars.AsTarget(valVar), true)
	StreamYieldIndexed(loop, "fi", valVar, keyVar)
	loop.Pop("}")
}

func StreamReadAliasType(ctx GeneratorContext, t *types.AliasType) {
	d := ctx.PushDebugRead("streamng-alias", t.Name())
	defer d.End()

	var aliasTypeName string = t.Name()
	if t.IsPtr() {
		aliasTypeName = "*" + t.Name()
	}

	switch aliasType := t.Alias.(type) {
	case *types.SliceType:
		StreamReadArrayType(ctx, aliasType)
	case *types.MapType:
		StreamReadMapType(ctx, aliasType)
	default:
		target := vars.AsTarget(ctx.NextVar())
		ReadType(ctx, t.Alias, target, true)
		StreamYieldSingle(ctx.AsScope(), "fi", target.CastAs(aliasTypeName).String())
	}

}
