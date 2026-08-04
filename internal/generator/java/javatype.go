package java

import (
	"sort"
	"strings"
	"unicode"

	"github.com/opencost/bingen/internal/types"
)

// primitiveInfo holds the Java syntax for a primitive IR type: the unboxed and
// boxed type names and the ReadBuffer/WriteBuffer method suffixes. Unsigned
// values widen one step (uint/uint32 -> long, uint8/uint16 -> int) to preserve
// their full range. Strings are handled separately because of string-table
// support.
type primitiveInfo struct {
	prim  string
	boxed string
	read  string
	write string
}

var primitives = map[uint8]primitiveInfo{
	types.TypeBool:    {"boolean", "Boolean", "readBool", "writeBool"},
	types.TypeInt:     {"int", "Integer", "readInt", "writeInt"},
	types.TypeInt8:    {"byte", "Byte", "readInt8", "writeInt8"},
	types.TypeInt16:   {"short", "Short", "readInt16", "writeInt16"},
	types.TypeInt32:   {"int", "Integer", "readInt32", "writeInt32"},
	types.TypeInt64:   {"long", "Long", "readInt64", "writeInt64"},
	types.TypeUInt:    {"long", "Long", "readUInt", "writeUInt"},
	types.TypeUInt8:   {"int", "Integer", "readUInt8", "writeUInt8"},
	types.TypeUInt16:  {"int", "Integer", "readUInt16", "writeUInt16"},
	types.TypeUInt32:  {"long", "Long", "readUInt32", "writeUInt32"},
	types.TypeUInt64:  {"long", "Long", "readUInt64", "writeUInt64"},
	types.TypeFloat32: {"float", "Float", "readFloat32", "writeFloat32"},
	types.TypeFloat64: {"double", "Double", "readFloat64", "writeFloat64"},
}

// JavaType returns the Java type string for a field of the given IR type.
// Primitive fields are unboxed unless they are pointers (nilable), in which case
// the boxed wrapper is used. Slice/map element types are always boxed because
// Java generics cannot hold primitives.
func JavaType(t types.GenType) string {
	return javaType(t, false)
}

func javaType(t types.GenType, forceBox bool) string {
	box := forceBox || t.IsPtr()

	if p, ok := primitives[t.Code()]; ok {
		return toPrimitive(box, p.prim, p.boxed)
	}

	switch t.Code() {
	case types.TypeString:
		return "String"
	case types.TypeSlice:
		st := t.(*types.SliceType)
		return "List<" + javaType(st.InnerType, true) + ">"
	case types.TypeMap:
		mt := t.(*types.MapType)
		return "Map<" + javaType(mt.KeyType, true) + ", " + javaType(mt.ValueType, true) + ">"
	case types.TypeInterface:
		if name := t.Name(); name != "interface{}" {
			return name
		}
		return "Object"
	case types.TypeStruct:
		return t.Name()
	case types.TypeAlias:
		return javaType(t.(*types.AliasType).Alias, box)
	case types.TypeReference:
		if resolved := t.(*types.ReferenceType).Resolve(); resolved != nil {
			return javaType(resolved, box)
		}

		if ext, ok := externalTypes[t.Name()]; ok {
			return ext.javaType
		}

		return externalRefName(t.Name())
	}

	return "Object"
}

// toPrimitive returns the boxed wrapper when box is true, otherwise the primitive.
func toPrimitive(box bool, primitive string, boxed string) string {
	if box {
		return boxed
	}
	return primitive
}

// externalRefName reduces a qualified go type name (ie:  "shared.Name") to its
// bare Java type name.
func externalRefName(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// JavaFieldName converts an exported go field name to a conventional Java
// camelCase name, guarding against reserved words.
func JavaFieldName(goName string) string {
	if goName == "" {
		return goName
	}

	runes := []rune(goName)
	runes[0] = unicode.ToLower(runes[0])
	name := string(runes)
	if javaReservedWords[name] {
		return name + "_"
	}
	return name
}

// requiredImports returns the sorted set of imports required by the struct's
// fields: java.util collections and any external-type imports (ie:  java.time),
// recursing into nested element types.
func requiredImports(st *types.StructType) []string {
	need := make(map[string]bool)

	for _, f := range st.Fields {
		types.WalkType(f.Type, func(t types.GenType) {
			switch t.Code() {
			case types.TypeSlice:
				need["java.util.List"] = true
			case types.TypeMap:
				need["java.util.Map"] = true
			case types.TypeReference:
				if ext, ok := externalTypes[t.Name()]; ok {
					for _, imp := range ext.typeImports {
						need[imp] = true
					}
				}
			}
		})
	}

	out := make([]string, 0, len(need))
	for imp := range need {
		out = append(out, imp)
	}
	sort.Strings(out)

	return out
}
