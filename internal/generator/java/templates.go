package java

import (
	"embed"
	"fmt"
	"io"
	"text/template"

	"github.com/opencost/bingen/internal/types"
)

// Template file names emitted by the java exporter. Additional per-type and
// runtime templates are added as the exporter is filled in.
const (
	// PackageInfoTemplate is the package-info.java placeholder emitted for every
	// generated package.
	PackageInfoTemplate = "package-info.java.tmpl"

	// RecordTemplate is the immutable record + builder emitted per struct type.
	RecordTemplate = "record.java.tmpl"

	// EncoderTemplate is the BinEncoder implementation emitted per struct type.
	EncoderTemplate = "encoder.java.tmpl"

	// DecoderTemplate is the BinDecoder implementation emitted per struct type.
	DecoderTemplate = "decoder.java.tmpl"

	// StreamerTemplate is the push streamer emitted per streamable struct type.
	StreamerTemplate = "streamer.java.tmpl"

	// InterfaceTemplate is the empty marker interface emitted per go interface.
	InterfaceTemplate = "interface.java.tmpl"

	// RegistryTemplate is the wire-string -> codec dispatch table.
	RegistryTemplate = "registry.java.tmpl"
)

// InterfaceParams is the template data for a generated marker interface.
type InterfaceParams struct {
	Package string
	Name    string
}

// RegistryEntry registers one concrete type keyed by its go wire string.
type RegistryEntry struct {
	Wire string
	Type string
}

// RegistryParams is the template data for the generated TypeRegistry.
type RegistryParams struct {
	Package        string
	RuntimePackage string
	Entries        []RegistryEntry
}

// RecordParams is the template data for a generated record.
type RecordParams struct {
	Package          string
	Imports          []string
	ImplementsClause string
	Type             *types.StructType
}

// CodecParams is the template data for a generated encoder or decoder.
type CodecParams struct {
	Package        string
	RuntimePackage string
	Imports        []string
	Version        uint8
	StringTable    bool
	PreProcess     bool
	PostProcess    bool
	Migration      bool
	Body           string
	Type           *types.StructType
}

// runtimeTemplates maps each shared-runtime output file name to the template
// that produces it. These are emitted into the runtime package on every run.
var runtimeTemplates = map[string]string{
	"ReadBuffer.java":            "ReadBuffer.java.tmpl",
	"WriteBuffer.java":           "WriteBuffer.java.tmpl",
	"BinEncoder.java":            "BinEncoder.java.tmpl",
	"EncodingContext.java":       "EncodingContext.java.tmpl",
	"BinDecoder.java":            "BinDecoder.java.tmpl",
	"DecodingContext.java":       "DecodingContext.java.tmpl",
	"StringTableWriter.java":     "StringTableWriter.java.tmpl",
	"StringTableReader.java":     "StringTableReader.java.tmpl",
	"GoTime.java":                "GoTime.java.tmpl",
	"BingenFieldInfo.java":       "BingenFieldInfo.java.tmpl",
	"BingenValue.java":           "BingenValue.java.tmpl",
	"BingenConsumer.java":        "BingenConsumer.java.tmpl",
	"BingenStreamFn.java":        "BingenStreamFn.java.tmpl",
	"BingenElement.java":         "BingenElement.java.tmpl",
	"BingenStreamException.java": "BingenStreamException.java.tmpl",
	"BingenPullStream.java":      "BingenPullStream.java.tmpl",
}

// RuntimeParams is the template data for the shared-runtime templates.
type RuntimeParams struct {
	RuntimePackage string
}

//go:embed templates/java/*.tmpl
var javaTemplatesFS embed.FS

// javaTemplates holds all parsed java templates.
var javaTemplates *template.Template

func init() {
	var err error
	javaTemplates = template.New("java-bingen-templates").Funcs(template.FuncMap{
		"JavaType":      JavaType,
		"JavaFieldName": JavaFieldName,
	})
	javaTemplates, err = javaTemplates.ParseFS(javaTemplatesFS, "templates/java/*.tmpl")
	if err != nil {
		panic(err)
	}
}

// writeTemplate executes the named template with the provided data into the
// writer.
func writeTemplate(writer io.Writer, name string, data any) error {
	if err := javaTemplates.ExecuteTemplate(writer, name, data); err != nil {
		return fmt.Errorf("executing template %q: %w", name, err)
	}
	return nil
}
