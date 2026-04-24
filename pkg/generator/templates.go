package generator

import (
	"embed"
	"fmt"
	"io"
	"text/template"

	"github.com/opencost/bingen/pkg/generator/vars"
	"github.com/opencost/bingen/pkg/meta"
	"github.com/opencost/bingen/pkg/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// BingenSupportTemplate is the template name for supporting bingen code, including the file
	// header, package declaration, imports, and various utilities.
	BingenSupportTemplate = "support.go.tmpl"

	// BingenMarshallerTemplate is the template name for the binary marshalling generation. It contains
	// the MarshalBinary() implementation as well as various sub-templates for writing different types.
	BingenMarshallerTemplate = "marshaller.go.tmpl"

	// BingenUnmarshallerTemplate is the template name for the binary unmarshalling generation. It contains
	// the UnmarshalBinary() implementation as well as various sub-templates for reading different types.
	BingenUnmarshallerTemplate = "unmarshaller.go.tmpl"

	// BingenStreamerTemplate is the template name for the binary streaming generation. It contains variations
	// to the unmarshalling generation as well as calls into the unmarshalling template.
	BingenStreamerTemplate = "streamer.go.tmpl"
)

//go:embed templates/go/*.tmpl
var goBingenTemplatesFS embed.FS

var (
	goBingenTemplates *template.Template
	titleCaser        cases.Caser = cases.Title(language.Und, cases.NoLower)
)

// initialize the templating engine with the go template files, provide external support via
// func maps
func init() {
	var err error

	goBingenTemplates = template.New("go-bingen-templates")
	goBingenTemplates = goBingenTemplates.Funcs(template.FuncMap{
		"AsTarget":            vars.AsTarget,
		"NewFieldTarget":      vars.NewFieldTarget,
		"NewIndexedTarget":    vars.NewIndexedTarget,
		"NewTypeCastedTarget": vars.NewTypeCastedTarget,
		"NewCastedTarget":     vars.NewCastedTarget,
		"ToWriteParams":       ToWriteParams,
		"ToReadParams":        ToReadParams,
		"ToStreamParams":      ToStreamParams,
		"ToTitle":             titleCaser.String,
		"ToDefaultValue":      ToDefaultValue,
		"Map":                 MakeMap,
	})

	goBingenTemplates, err = goBingenTemplates.ParseFS(goBingenTemplatesFS, "templates/go/*.tmpl")
	if err != nil {
		panic(err)
	}
}

// MarshallerParams contains the general set of parameters to pass to any type marshalling generator.
// It contains the current generator context, the type currently being marshalled, and the name of the
// target to write from.
type MarshallerParams struct {
	Context GeneratorContext
	Type    types.GenType
	Target  vars.Target
}

// UnmarshallerParams contains the general set of parameters to pass to any type unmarshalling generator.
// It contains the current generator context, the type currently being unmarshalled, the target to read
// into, and whether or not that target requires initialization (Set should be false if the target already
// exists).
type UnmarshallerParams struct {
	Context GeneratorContext
	Type    types.GenType
	Target  vars.Target
	Set     bool
}

// StreamParams contains the general set of parameters to pass to any type streaming generator. It contains
// the current generator context and the type currently being generated.
type StreamParams struct {
	Context GeneratorContext
	Type    types.GenType
}

// SupportParams contains the general set of parameters to pass to the support generator. This includes code
// that supports all bingen operations.
type SupportParams struct {
	Package         string
	Imports         []string
	BufferImport    string
	VersionSets     []meta.VersionSet
	Types           []types.GenType
	StreamableTypes []*types.StructType
}

// ToWriteParams is a helper function made available to the templating engine to create MarshallerParams inline.
func ToWriteParams(context GeneratorContext, t types.GenType, target vars.Target) MarshallerParams {
	return MarshallerParams{
		Context: context,
		Type:    t,
		Target:  target,
	}
}

// ToReadParams is a helper function made available to the templating engine to create UnmarshallerParams inline.
func ToReadParams(context GeneratorContext, t types.GenType, target vars.Target, set bool) UnmarshallerParams {
	return UnmarshallerParams{
		Context: context,
		Type:    t,
		Target:  target,
		Set:     set,
	}
}

// ToStreamParams is a helper function made available to the templating engine to create StreamParams inline.
func ToStreamParams(context GeneratorContext, t types.GenType) StreamParams {
	return StreamParams{
		Context: context,
		Type:    t,
	}
}

// MakeMap is a helper function made available to the teplating engine to create dynamic parameters to pass
// to various templates. ie: (Map "Context" .Context "Type" .Type.InnerType)
//
// Templates accept a single object called a "pipeline". From that object, all properties can be accessed via
// . notation. So, for the above pipeline, .Context and .Type. In scenarios where templates create parameters
// on the fly, in order to pass those parameters to different templates, you can use this func to build the pipeline.
func MakeMap(kvPairs ...any) map[string]any {
	m := make(map[string]any)
	for i := 0; i < len(kvPairs); i += 2 {
		key, ok := kvPairs[i].(string)
		value := kvPairs[i+1]

		if !ok {
			continue
		}
		m[key] = value
	}
	return m
}

// default type format mapping with specific casts
var defaultValues map[uint8]string = map[uint8]string{
	types.TypeInt:     "int(%s)",
	types.TypeInt8:    "int8(%s)",
	types.TypeInt16:   "int16(%s)",
	types.TypeInt32:   "int32(%s)",
	types.TypeInt64:   "int64(%s)",
	types.TypeUInt:    "uint(%s)",
	types.TypeUInt8:   "uint8(%s)",
	types.TypeUInt16:  "uint16(%s)",
	types.TypeUInt32:  "uint32(%s)",
	types.TypeUInt64:  "uint64(%s)",
	types.TypeFloat32: "float32(%s)",
	types.TypeFloat64: "float64(%s)",
}

// ToDefaultValue provides a primitive default value mapping to string formatters for
// quickly generating the type casted default value for a type and default.
func ToDefaultValue(typeCode uint8, value string) string {
	if typeCode == types.TypeString {
		return fmt.Sprintf("\"%s\"", value)
	}

	if typeCode == types.TypeBool {
		if value == "" {
			return "false"
		}
		return value
	}

	if value == "" {
		return fmt.Sprintf(defaultValues[typeCode], "0")
	}
	return fmt.Sprintf(defaultValues[typeCode], value)
}

// WriteSupportTemplate writes the file header, package declaration, imports, and all supporting bingen
// code to the provided `io.Writer`
func WriteSupportTemplate(writer io.Writer, params SupportParams) error {
	return goBingenTemplates.ExecuteTemplate(writer, BingenSupportTemplate, params)
}

// WriteMarshallerTemplate writes the binary marshalling code to the provided `io.Writer`
func WriteMarshallerTemplate(writer io.Writer, params MarshallerParams) error {
	return goBingenTemplates.ExecuteTemplate(writer, BingenMarshallerTemplate, params)
}

// WriteUnmarshallerTemplate writes the binary unmarshalling code to the provided `io.Writer`
func WriteUnmarshallerTemplate(writer io.Writer, params UnmarshallerParams) error {
	return goBingenTemplates.ExecuteTemplate(writer, BingenUnmarshallerTemplate, params)
}

// WriteStreamerTemplate writes all of the streaming unmarshalling code to the provided `io.Writer`
func WriteStreamerTemplate(writer io.Writer, params StreamParams) error {
	return goBingenTemplates.ExecuteTemplate(writer, BingenStreamerTemplate, params)
}
