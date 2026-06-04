package types

import (
	"go/ast"
	"testing"

	"github.com/opencost/bingen/pkg/meta"
)

func TestDefinitionToAst(t *testing.T) {
	td := &meta.TypeDefinition{
		Package: "github.com/opencost/bingen/shared",
		Name:    "FloatList",
		Type:    "[]float64",
	}

	if td.PackageSelector() != "shared" {
		t.Fatalf("PackageShort() should return 'shared'. Got: %s", td.PackageSelector())
	}

	if td.FullName() != "shared.FloatList" {
		t.Fatalf("FullName() should return 'shared.FloatList'. Got: %s", td.FullName())
	}

	typeSpec, err := ParseDefined(td)
	if err != nil {
		t.Fatal(err)
	}

	if typeSpec.Name.Name != "FloatList" {
		t.Fatalf("Expected type name to equal 'FloatList'. Got: %s", typeSpec.Name.Name)
	}

	sliceType, ok := typeSpec.Type.(*ast.ArrayType)
	if !ok {
		t.Fatalf("Expected the type to be a slice/array. Instead got: %T", typeSpec.Type)
	}

	ident, ok := sliceType.Elt.(*ast.Ident)
	if !ok {
		t.Fatalf("Expected the inner type to be an ident. Instead got: %T", sliceType.Elt)
	}

	if ident.Name != "float64" {
		t.Fatalf("Expected the inner type to be float64. Got: %s", ident.Name)
	}
}
