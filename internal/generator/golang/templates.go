package golang

import (
	"embed"
	"io"
	"text/template"

	"github.com/opencost/bingen/internal/meta"
	"github.com/opencost/bingen/internal/types"
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
		"ToTitle": titleCaser.String,
	})

	goBingenTemplates, err = goBingenTemplates.ParseFS(goBingenTemplatesFS, "templates/go/*.tmpl")
	if err != nil {
		panic(err)
	}
}

// MarshallerParams contains the general set of parameters to pass to the marshaller
// method skeleton: the generator context, the type being marshalled, and the
// field-encoding Body produced by the core Go backend.
type MarshallerParams struct {
	Context GeneratorContext
	Type    types.GenType
	Body    string
}

// UnmarshallerParams contains the parameters passed to the unmarshaller method
// skeleton: the generator context, the type being unmarshalled, and the
// field-decoding Body produced by the core Go backend.
type UnmarshallerParams struct {
	Context GeneratorContext
	Type    types.GenType
	Body    string
}

// StreamParams contains the general set of parameters to pass to any type streaming generator. It contains
// the current generator context and the type currently being generated.
type StreamParams struct {
	Context GeneratorContext
	Type    types.GenType
	Body    string
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
