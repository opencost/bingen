package java

import (
	"testing"

	"github.com/opencost/bingen/internal/types"
)

func basic(code uint8) types.GenType {
	return types.BasicTypes[code]
}

func TestJavaType_Primitives(t *testing.T) {
	cases := map[uint8]string{
		types.TypeBool:    "boolean",
		types.TypeInt:     "int",
		types.TypeInt8:    "byte",
		types.TypeInt16:   "short",
		types.TypeInt32:   "int",
		types.TypeInt64:   "long",
		types.TypeUInt:    "long",
		types.TypeUInt8:   "int",
		types.TypeUInt16:  "int",
		types.TypeUInt32:  "long",
		types.TypeUInt64:  "long",
		types.TypeFloat32: "float",
		types.TypeFloat64: "double",
		types.TypeString:  "String",
	}

	for code, want := range cases {
		if got := JavaType(basic(code)); got != want {
			t.Errorf("JavaType(code=%d) = %q, want %q", code, got, want)
		}
	}
}

func TestJavaType_PointerBoxes(t *testing.T) {
	if got := JavaType(basic(types.TypeInt).CreatePtr()); got != "Integer" {
		t.Errorf("JavaType(*int) = %q, want Integer", got)
	}
	if got := JavaType(basic(types.TypeFloat64).CreatePtr()); got != "Double" {
		t.Errorf("JavaType(*float64) = %q, want Double", got)
	}
	// String is a reference type; a pointer stays String.
	if got := JavaType(basic(types.TypeString).CreatePtr()); got != "String" {
		t.Errorf("JavaType(*string) = %q, want String", got)
	}
}

func TestJavaType_SliceAndMapBoxInner(t *testing.T) {
	slice := &types.SliceType{
		BasicType: types.NewBasicType("", "[]int", types.TypeSlice, false, false),
		InnerType: basic(types.TypeInt),
	}
	if got := JavaType(slice); got != "List<Integer>" {
		t.Errorf("JavaType([]int) = %q, want List<Integer>", got)
	}

	m := &types.MapType{
		BasicType: types.NewBasicType("", "map[string]float64", types.TypeMap, false, false),
		KeyType:   basic(types.TypeString),
		ValueType: basic(types.TypeFloat64),
	}
	if got := JavaType(m); got != "Map<String, Double>" {
		t.Errorf("JavaType(map[string]float64) = %q, want Map<String, Double>", got)
	}
}

func TestJavaFieldName(t *testing.T) {
	cases := map[string]string{
		"Name":     "name",
		"Children": "children",
		"ID":       "iD",
		"Value":    "value",
	}

	for in, want := range cases {
		if got := JavaFieldName(in); got != want {
			t.Errorf("JavaFieldName(%q) = %q, want %q", in, got, want)
		}
	}
}
