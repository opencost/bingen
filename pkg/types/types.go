package types

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/opencost/bingen/pkg/meta"
)

//--------------------------------------------------------------------------
//  Basic Generator Types
//--------------------------------------------------------------------------

const (
	TypeBool = iota
	TypeInt
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUInt
	TypeUInt8
	TypeUInt16
	TypeUInt32
	TypeUInt64
	TypeFloat32
	TypeFloat64
	TypeString
	TypeMap
	TypeSlice
	TypeInterface
	TypeStruct
	TypeReference
	TypeAlias
)

// BasicTypes
var BasicTypes = []*BasicType{
	NewBasicType("", "bool", TypeBool, false, false),
	NewBasicType("", "int", TypeInt, false, false),
	NewBasicType("", "int8", TypeInt8, false, false),
	NewBasicType("", "int16", TypeInt16, false, false),
	NewBasicType("", "int32", TypeInt32, false, false),
	NewBasicType("", "int64", TypeInt64, false, false),
	NewBasicType("", "uint", TypeUInt, false, false),
	NewBasicType("", "uint8", TypeUInt8, false, false),
	NewBasicType("", "uint16", TypeUInt16, false, false),
	NewBasicType("", "uint32", TypeUInt32, false, false),
	NewBasicType("", "uint64", TypeUInt64, false, false),
	NewBasicType("", "float32", TypeFloat32, false, false),
	NewBasicType("", "float64", TypeFloat64, false, false),
	NewBasicType("", "string", TypeString, false, false),
}

//--------------------------------------------------------------------------
//  AnnotatedTypeOpts
//--------------------------------------------------------------------------

// GenerateTypeOpts sets the predefined generator operations that can be applied
// when encoding/decoding the type
type GenerateTypeOpts struct {
	SetName               string
	SetVersion            uint8
	IsStreamable          bool
	IsGenerateStringTable bool
	IsPreProcess          bool
	IsPostProcess         bool
	IsMigration           bool
}

// FieldOpts represents any specific field metadata parsed by bingen
type FieldOpts struct {
	Version uint8
	Default string
	Ignore  bool
}

//--------------------------------------------------------------------------
//  AnnotatedType
//--------------------------------------------------------------------------

// AnnotatedType represents a generator targetted type with generator options defined
type AnnotatedType struct {
	T    *ast.TypeSpec
	Opts *GenerateTypeOpts
}

//--------------------------------------------------------------------------
//  AnnotatedField
//--------------------------------------------------------------------------

// AnnotatedField represents a generator targetted type field with specific field opts defined
type AnnotatedField struct {
	F    *ast.Field
	Opts *FieldOpts
}

//--------------------------------------------------------------------------
//  Import
//--------------------------------------------------------------------------

// Import represents a package import either passed in via the import annotation, or referenced
// implicitly by the type.
type Import struct {
	// The short name for the package
	Name string

	// The full import path for the package
	Path string
}

// ImportFromPath returns a new Import instance from a package import path.
func ImportFromPath(path string) *Import {
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return &Import{
			Name: path,
			Path: path,
		}
	}

	return &Import{
		Name: parts[len(parts)-1],
		Path: path,
	}
}

//--------------------------------------------------------------------------
//  Generator Intermediate Representation
//--------------------------------------------------------------------------

// GenType is the intermediate representation of
type GenType interface {
	// Name returns the name of the type.
	Name() string

	// TypeName returns the pointer qualified type name
	TypeName() string

	// FullName returns the full pkg/name of the type
	FullName() string

	// Code returns the byte enumeration type id.
	Code() uint8

	// IsMap returns true if the type is a map.
	IsMap() bool

	// IsArray returns true if the type is an array or slice.
	IsArray() bool

	// IsPtr returns true if the type is a pointer type.
	IsPtr() bool

	// IsInterface returns true if the type is an interface type
	IsInterface() bool

	// IsNilable returns true if the nil is an acceptable value to assign to an instance
	// of this type.
	IsNilable() bool

	// CreatePtr creates a new pointer GenType implementation
	CreatePtr() GenType
}

// Creates a map of primitive types returned for a dynamically growing
func NewBasicTypes() map[string]GenType {
	m := make(map[string]GenType)
	for i := 0; i < len(BasicTypes); i++ {
		m[BasicTypes[i].name] = BasicTypes[i]
	}
	return m
}

//--------------------------------------------------------------------------
//  GenType implementations
//--------------------------------------------------------------------------

type BasicType struct {
	pkg  string
	name string
	code uint8
	ptr  bool
	intf bool
}

// Name returns the name of the type.
func (bt *BasicType) Name() string {
	return bt.name
}

// TypeName returns the pointer qualified type name
func (bt *BasicType) TypeName() string {
	name := bt.Name()
	if bt.IsPtr() {
		name = "*" + name
	}
	return name
}

// FullName returns the full pkg/name of the type
func (bt *BasicType) FullName() string {
	return bt.name
}

// Code returns the byte enumeration type id.
func (bt *BasicType) Code() uint8 {
	return bt.code
}

// IsMap returns true if the type is a map.
func (bt *BasicType) IsMap() bool {
	return bt.code == TypeMap
}

// IsArray returns true if the type is an array or slice.
func (bt *BasicType) IsArray() bool {
	return bt.code == TypeSlice
}

// IsPtr returns true if the type is a pointer type.
func (bt *BasicType) IsPtr() bool {
	return bt.ptr
}

// IsInterface returns true if the type is an interface type
func (bt *BasicType) IsInterface() bool {
	return bt.intf
}

// IsNilable returns true if the nil is an acceptable value to assign to an instance
// of this type.
func (bt *BasicType) IsNilable() bool {
	return bt.IsPtr() || bt.IsMap() || bt.IsArray() || bt.IsInterface()
}

// CreatePtr creates a new pointer GenType implementation
func (bt *BasicType) CreatePtr() GenType {
	return &BasicType{
		pkg:  bt.pkg,
		name: bt.name,
		code: bt.code,
		ptr:  true,
	}
}

// NewBasicType Creates a new basic type
func NewBasicType(pkg string, name string, code uint8, ptr bool, intf bool) *BasicType {
	return &BasicType{pkg, name, code, ptr, intf}
}

// SliceType is the GenType implementation for a slice.
type SliceType struct {
	*BasicType
	InnerType GenType
}

// CreatePtr creates a new slice pointer
func (at *SliceType) CreatePtr() GenType {
	return &SliceType{
		BasicType: at.BasicType.CreatePtr().(*BasicType),
		InnerType: at.InnerType,
	}
}

// MapType is the GenType implementation for a map.
type MapType struct {
	*BasicType
	KeyType   GenType
	ValueType GenType
}

// CreatePtr creates a new map pointer
func (mt *MapType) CreatePtr() GenType {
	return &MapType{
		BasicType: mt.BasicType.CreatePtr().(*BasicType),
		KeyType:   mt.KeyType,
		ValueType: mt.ValueType,
	}
}

// StructField represents a field, imported or exported within a struct
type StructField struct {
	Name string
	Type GenType
	Opts *FieldOpts
}

// StructType represents a struct type
type StructType struct {
	*BasicType
	Fields []*StructField
	Opts   *GenerateTypeOpts
}

// CreatePtr creates a new struct ptr
func (ut *StructType) CreatePtr() GenType {
	return &StructType{
		BasicType: ut.BasicType.CreatePtr().(*BasicType),
		Fields:    ut.Fields,
		Opts:      ut.Opts,
	}
}

// ReferenceType represents a reference to a struct/type that hasn't been resolved
// yet. Once the full set of types has been collected, these can be resolved, but
// normally suffice as just a stand-in.
type ReferenceType struct {
	*BasicType
	Resolve func() GenType
}

// AliasType represents type alias
type AliasType struct {
	*BasicType
	Alias GenType
}

func (at *AliasType) CreatePtr() GenType {
	return &AliasType{
		BasicType: at.BasicType.CreatePtr().(*BasicType),
		Alias:     at.Alias,
	}
}

// InterfaceType represents an interface
type InterfaceType struct {
	*BasicType
}

//--------------------------------------------------------------------------
//  TypeCollection
//--------------------------------------------------------------------------

// TypeCollection is used to convert ast.TypeSpec types into GenType implementations
// which represent intermediate forms of the types.
type TypeCollection interface {
	// AddStructType adds a struct type with the provided fields
	AddStructType(t *AnnotatedType, fields []*AnnotatedField)

	// AddInterface adds an interface type to the known types in the collection
	AddInterface(t *AnnotatedType)

	// AddAlias adds a type alias to the known types in the collection by
	// resolving the alias target.
	AddAlias(t *AnnotatedType, isPtr bool)

	// Types returns all of the GenType types to generate
	Types() []GenType

	// Imports returns all additional imports to include.
	Imports() []string

	// VersionSets contains a list of version sets defined
	VersionSets() []meta.VersionSet
}

// typeCollector is a TypeCollection implementation that creates the type representation
// used for generation
type typeCollector struct {
	knownTypes  map[string]GenType
	collected   []GenType
	imports     map[string]string
	versionSets []meta.VersionSet
}

// NewTypeCollection creates a TypeCollection pre-populated with basic types.
func NewTypeCollection(annotations *meta.BingenAnnotated) TypeCollection {
	imports := make(map[string]string)
	for _, imp := range annotations.Imports {
		im := ImportFromPath(imp)
		imports[im.Name] = im.Path
	}

	tc := &typeCollector{
		knownTypes:  NewBasicTypes(),
		collected:   []GenType{},
		imports:     imports,
		versionSets: annotations.VersionSets,
	}

	for _, def := range annotations.Definitions {
		tc.addTypeDefinition(def)
	}

	return tc
}

// AddStructType adds a struct type with the provided fields
func (tc *typeCollector) AddStructType(t *AnnotatedType, fields []*AnnotatedField) {
	gt := tc.toStructType(t, fields)
	tc.knownTypes[gt.Name()] = gt
	tc.collected = append(tc.collected, gt)
}

// AddInterface adds an interface type to the known types in the collection
func (tc *typeCollector) AddInterface(t *AnnotatedType) {
	it := &InterfaceType{
		BasicType: NewBasicType("", t.T.Name.Name, TypeInterface, false, true),
	}

	tc.knownTypes[it.Name()] = it
}

// AddAlias adds a type alias to the known types in the collection by
// resolving the alias target.
func (tc *typeCollector) AddAlias(t *AnnotatedType, isPtr bool) {
	resolved := tc.toGenType(t.T.Type, "", isPtr)
	gt := &AliasType{
		BasicType: NewBasicType("", t.T.Name.Name, TypeAlias, false, false),
		Alias:     resolved,
	}

	tc.knownTypes[gt.Name()] = gt
}

// type definitions, by design, are treated as an alias since there is no way to express
// an external type definition within the bingen command syntax. It only differs from an
// in-package alias by a package prefix. ie: <package>.<alias-name>. This allows field
// references to easily lookup the type definition at generation time.
func (tc *typeCollector) addTypeDefinition(def *meta.TypeDefinition) {
	typeSpec, err := ParseDefined(def)
	if err != nil {
		panic(fmt.Errorf("failed to parse type definition %s: %w", def.Name, err))
	}

	resolved := tc.toGenType(typeSpec.Type, "", false)
	gt := &AliasType{
		BasicType: NewBasicType(def.Package, def.FullName(), TypeAlias, false, false),
		Alias:     resolved,
	}

	tc.knownTypes[gt.Name()] = gt
}

// Types returns all of the GenType types to generate
func (tc *typeCollector) Types() []GenType {
	return tc.collected
}

// Imports returns a slice of imports to include in the generated source
func (tc *typeCollector) Imports() []string {
	im := make([]string, 0, len(tc.imports))
	for _, v := range tc.imports {
		im = append(im, v)
	}
	// go/format reorders the import block in generated *_codecs.go; this sort
	// only stabilizes tc.Imports() for callers/tests, not final file layout.
	sort.Strings(im)
	return im
}

// toStructType creates a struct type representation with a set of fields
func (tc *typeCollector) toStructType(at *AnnotatedType, typeFields []*AnnotatedField) GenType {
	fields := []*StructField{}

	t := at.T
	opts := at.Opts

	for _, typeField := range typeFields {
		if typeField.Opts != nil && typeField.Opts.Ignore {
			continue
		}

		fields = append(fields, tc.toStructField(t.Name.Name, typeField)...)
	}

	userType := &StructType{
		BasicType: NewBasicType("", t.Name.Name, TypeStruct, false, false),
		Fields:    fields,
		Opts:      opts,
	}

	return userType
}

// toStructField creates the field type for a specific type
func (tc *typeCollector) toStructField(typeName string, typeField *AnnotatedField) []*StructField {
	fields := []*StructField{}

	f := typeField.F
	opts := typeField.Opts

	// filter fields to ignore
	names := []*ast.Ident{}
	names = append(names, f.Names...)
	if len(names) == 0 {
		return fields
	}

	// Resolve Type after filtering names in case we filter all names of this type
	fieldType := tc.toGenType(f.Type, "", false)
	for _, i := range names {
		fields = append(fields, &StructField{
			Name: i.Name,
			Type: fieldType,
			Opts: opts,
		})
	}

	return fields
}

// VersionSets contains a list of version sets defined
func (tc *typeCollector) VersionSets() []meta.VersionSet {
	return tc.versionSets
}

func (tc *typeCollector) toGenType(t ast.Expr, selectorName string, isPtr bool) GenType {
	switch vt := t.(type) {
	// StartExpr defines a pointer type
	case *ast.StarExpr:
		return tc.toGenType(vt.X, "", true)

	// ArrayType defines an array or slice type
	case *ast.ArrayType:
		innerType := tc.toGenType(vt.Elt, "", isPtr)
		innerName := innerType.Name()
		if innerType.IsPtr() {
			innerName = "*" + innerName
		}
		combinedName := "[]" + innerName

		return &SliceType{
			BasicType: NewBasicType("", combinedName, TypeSlice, isPtr, false),
			InnerType: innerType,
		}

	// MapType defines a map[KeyType]ValueType
	case *ast.MapType:
		keyType := tc.toGenType(vt.Key, "", isPtr)
		keyName := keyType.Name()
		if keyType.IsPtr() {
			keyName = "*" + keyName
		}

		valueType := tc.toGenType(vt.Value, "", isPtr)
		valueName := valueType.Name()
		if valueType.IsPtr() {
			valueName = "*" + valueName
		}
		combinedName := "map[" + keyName + "]" + valueName

		return &MapType{
			BasicType: NewBasicType("", combinedName, TypeMap, isPtr, false),
			KeyType:   keyType,
			ValueType: valueType,
		}

	// InterfaceType is either an empty interface, or a dynamic interface.
	case *ast.InterfaceType:
		return &InterfaceType{
			BasicType: NewBasicType("", "interface{}", TypeInterface, isPtr, false),
		}

	// Represents a <package>.<ident>
	case *ast.SelectorExpr:
		return tc.toGenType(vt.X, vt.Sel.Name, isPtr)

	// Identifier which represents another type, which we can lookup to determine
	// if there is already generated type metadata for the name. If there is, we can
	// simply return the GenType result. Otherwise, we'll need to defer resolution
	// until after we've completed adding types. This specific case is treated as a
	// ReferenceType, a placeholder for a type which we'll resolve when requested later.
	case *ast.Ident:
		var fullName string
		if selectorName == "" {
			fullName = vt.Name
		} else {
			fullName = vt.Name + "." + selectorName
			if _, ok := tc.imports[vt.Name]; !ok {
				tc.imports[vt.Name] = vt.Name
			}
		}
		if kt, ok := tc.knownTypes[fullName]; ok {
			if isPtr {
				return kt.CreatePtr()
			}

			return kt
		}

		// Type was not part of the collection, so create a ReferenceType and
		// initialize with a resolution function.
		var rt *ReferenceType
		rt = &ReferenceType{
			BasicType: NewBasicType("", fullName, TypeReference, isPtr, false),
			Resolve: func() GenType {
				if resolvedType, ok := tc.knownTypes[rt.Name()]; ok {
					if rt.IsPtr() {
						return resolvedType.CreatePtr()
					}
					return resolvedType
				}

				return nil
			},
		}
		return rt
	}

	return BasicTypes[0]
}

func findAnnotationSet(typeName string, annotations []meta.VersionSet) (meta.VersionSet, bool) {
	for _, set := range annotations {
		if set.IsGenerate(typeName) {
			return set, true
		}
	}
	return nil, false
}

// findTypes locates types to generate based on annotation comments in the
// source file
func findTypes(file *ast.File, annotations []meta.VersionSet, defaultVersion uint8) []*AnnotatedType {
	types := []*AnnotatedType{}

	ast.Inspect(file, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		name := t.Name.String()
		annotationSet, ok := findAnnotationSet(name, annotations)
		if !ok {
			return true
		}

		fmt.Println("Found Type: ", t.Name.String())

		// Set the default version if the annotation set has a 0 version
		setVersion := annotationSet.Version()
		if setVersion == 0 {
			setVersion = defaultVersion
		}

		types = append(types, &AnnotatedType{
			T: t,
			Opts: &GenerateTypeOpts{
				SetName:               annotationSet.Name(),
				SetVersion:            setVersion,
				IsStreamable:          annotationSet.IsCommandOption(meta.AnnotationGenerate, name, meta.AnnotationOptionStreamable),
				IsGenerateStringTable: annotationSet.IsCommandOption(meta.AnnotationGenerate, name, meta.AnnotationOptionStringTable),
				IsPreProcess:          annotationSet.IsCommandOption(meta.AnnotationGenerate, name, meta.AnnotationOptionPreProcess),
				IsPostProcess:         annotationSet.IsCommandOption(meta.AnnotationGenerate, name, meta.AnnotationOptionPostProcess),
				IsMigration:           annotationSet.IsCommandOption(meta.AnnotationGenerate, name, meta.AnnotationOptionMigrate),
			},
		})
		return true
	})

	return types
}

// findFields locates field definitions within the type and returns them
func findFields(t *AnnotatedType) ([]*AnnotatedField, error) {
	var errs []string
	fields := []*AnnotatedField{}

	ast.Inspect(t.T, func(nn ast.Node) bool {
		f, ok := nn.(*ast.Field)
		if !ok {
			return true
		}

		var fieldOpts *FieldOpts

		// locate a field annotation if it exists, create field opts from annotation
		if f.Comment != nil {
			for _, c := range f.Comment.List {
				if a, ok := meta.ParseAnnotation(c.Text); ok {
					var err error
					fieldOpts, err = toFieldOpts(a)
					if err != nil {
						errs = append(errs, fmt.Sprintf("[%s.%s] Failed to parse field annotation: %s\n", t.T.Name, f.Names[0], err))
					}
				}
			}
		}

		// NOTE: Perfectly OK if fieldOpts is nil here - indicative of no field annotation,
		// NOTE: which is allowed.

		fields = append(fields, &AnnotatedField{
			F:    f,
			Opts: fieldOpts,
		})
		return true
	})

	if len(errs) != 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}

	return fields, nil
}

// create field opts from the field annotation
func toFieldOpts(a *meta.Annotation) (*FieldOpts, error) {
	if a == nil {
		return nil, fmt.Errorf("nil annotation")
	}
	if a.Command != meta.AnnotationField {
		return nil, fmt.Errorf("field annotation does not use 'field' command")
	}

	var version uint8
	var defaultVal string
	ignore := false

	for option := range a.Options {
		if option == meta.AnnotationFieldIgnore {
			ignore = true
			break
		}

		strs := strings.Split(option, "=")
		if len(strs) < 2 {
			return nil, fmt.Errorf("no set(=) detected for option")
		}

		prop := strings.TrimSpace(strs[0])
		value := strings.TrimSpace(strs[1])
		if prop == meta.AnnotationFieldVersion {
			r, err := strconv.ParseUint(value, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("failed to parse uint8 version from %q", value)
			}
			version = uint8(r)
		}
		if prop == meta.AnnotationFieldDefault {
			defaultVal = value
		}
	}

	return &FieldOpts{
		Version: version,
		Default: defaultVal,
		Ignore:  ignore,
	}, nil
}

//--------------------------------------------------------------------------
//  package funcs
//--------------------------------------------------------------------------

// Collect gathers an intermediate type representation of all the package types
// annotated for generation.
func LoadTypes(dir string, pkg string, defaultVersion uint8) (TypeCollection, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse: %s", err))
	}

	annotations, err := meta.LoadAnnotations(packages, defaultVersion)
	if err != nil {
		return nil, err
	}

	typeCollector := NewTypeCollection(annotations)

	// Visit packages and files in sorted order. generator.Generate sorts types
	// alphabetically before emission, but registration order here still affects
	// whether cross-package struct fields are captured as ReferenceType vs a
	// resolved StructType, which can change generated unmarshaller text.
	pkgNames := make([]string, 0, len(packages))
	for k := range packages {
		pkgNames = append(pkgNames, k)
	}
	sort.Strings(pkgNames)

	for _, k := range pkgNames {
		v := packages[k]
		fmt.Printf("Package: %s\n", k)

		fileNames := make([]string, 0, len(v.Files))
		for kk := range v.Files {
			fileNames = append(fileNames, kk)
		}
		sort.Strings(fileNames)

		for _, kk := range fileNames {
			file := v.Files[kk]
			fmt.Printf("File: %s\n", kk)

			types := findTypes(file, annotations.VersionSets, defaultVersion)
			for _, at := range types {
				t := at.T
				fields, err := findFields(at)
				if err != nil {
					return nil, err
				}

				switch t.Type.(type) {
				case *ast.StructType:
					typeCollector.AddStructType(at, fields)
				case *ast.InterfaceType:
					typeCollector.AddInterface(at)
				// TODO: This may be too broad of a catchall, and we may have other type
				// TODO: specific nodes that follow typing.
				default:
					typeCollector.AddAlias(at, false)
				}
			}
		}
	}

	return typeCollector, nil
}
