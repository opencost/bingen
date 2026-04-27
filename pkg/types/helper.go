package types

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"

	"github.com/opencost/bingen/pkg/meta"
)

// goIdentifier matches a single valid Go identifier. Used to validate names
// supplied via @bingen:define so we don't splice arbitrary text into a Go
// source file we then re-parse.
var goIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ParseDefined turns the bingen type definition into a go ast.TypeSpec by generating the type
// definition code, running it through the go parser, and then extracting the parsed TypeSpec.
// This allows us to pipe the resulting TypeSpec through the existing bingen machinery as if we
// were reading the type from the referenced package directly.
//
// Like the @bingen:imports command, this allows us to link in external package code without
// requiring direct access to the code for that type. This is safe to allow because if the type
// definition is wrong, then the resulting generated code will fail to compile.
func ParseDefined(td *meta.TypeDefinition) (*ast.TypeSpec, error) {
	if !goIdentifier.MatchString(td.Name) {
		return nil, fmt.Errorf("bingen: define name %q is not a valid Go identifier", td.Name)
	}

	if _, err := parser.ParseExpr(td.Type); err != nil {
		return nil, fmt.Errorf("bingen: define type %q is not a valid Go expression: %w", td.Type, err)
	}

	return ParseDecl(fmt.Sprintf("package %s; type %s %s", td.PackageSelector(), td.Name, td.Type))
}

// ParseDecl extracts a single package scoped type definition and returns the ast.TypeSpec. This
// implementation is meant to handle exactly one type for a single package.
func ParseDecl(declCode string) (*ast.TypeSpec, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", declCode, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse decl code: %s due to error: %w", declCode, err)
	}

	if len(f.Decls) != 1 {
		return nil, fmt.Errorf("failed to parse single type declaration from: %s", declCode)
	}

	genDecl, ok := f.Decls[0].(*ast.GenDecl)
	if !ok {
		return nil, fmt.Errorf("failed to locate GenDecl type in parsed result from: %s", declCode)
	}

	if genDecl.Tok != token.TYPE {
		return nil, fmt.Errorf("first discovered token was not a type from: %s", declCode)
	}

	if len(genDecl.Specs) != 1 {
		return nil, fmt.Errorf("failed to locate single type spec from: %s", declCode)
	}

	typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("failed to correctly parse single type spec from: %s", declCode)
	}

	return typeSpec, nil
}
