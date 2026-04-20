package meta

import "strings"

// Annotation Command Directives
const (
	// AnnotationPrefix is the prefix used to trigger parsing an annotation
	AnnotationPrefix = "@bingen"

	// AnnotationGenerate is the annotation command used to generate a target
	AnnotationGenerate = "generate"

	// AnnotationImport is the annotation command used to add a package import to the generated code.
	AnnotationImport = "import"

	// AnnotationSet pushes a new annotation set on the stack for command collection
	AnnotationSet = "set"

	// AnnotationEndSet pops the annotation set
	AnnotationEndSet = "end"

	// AnnotationField is an context aware annotation that is made on a single struct field
	AnnotationField = "field"
)

// AnnotationSet options
const (
	AnnotationDefaultSetName = "default"
	AnnotationSetName        = "name"
	AnnotationSetVersion     = "version"
)

// AnnotationField options
const (
	AnnotationFieldVersion = "version"
	AnnotationFieldDefault = "default"
	AnnotationFieldIgnore  = "ignore"
)

// Annotation command options
const (
	// AnnotationOptionStreamable is used to specify whether or not to generate BingenStream
	// implementations for the type. This allows the users to stream in the fields of an encoded
	// type without having to populate all the fields at once.
	AnnotationOptionStreamable = "streamable"

	// AnnotationOptionStringTable is used to specify whether or not there exists a string table
	// for the type. This is typically applied to a parent type.
	AnnotationOptionStringTable = "stringtable"

	// AnnotationOptionPreProcess is used to specify a pre-processing function to execute just before
	// the type is encoded.
	AnnotationOptionPreProcess = "preprocess"

	// AnnotationOptionPostProcess is used to specify a post processing function to execute just before
	// a decoding result is returned.
	AnnotationOptionPostProcess = "postprocess"

	// AnnotationOptionMigrate is used to specify a migration function to execute when the unmarshalled
	// data version is different from the current version.
	AnnotationOptionMigrate = "migrate"
)

//--------------------------------------------------------------------------
//  package types
//--------------------------------------------------------------------------

// Annotation is used to represent a parsed bingen Annotation begining with the prefix
// @bingen
type Annotation struct {
	Command string
	Options map[string]bool
	Target  string
}

// ParseAnnotation accepts a single comment and returns the annotation and bool
// status representing whether an annotation was parsed or not.
func ParseAnnotation(comment string) (*Annotation, bool) {
	idx := strings.Index(comment, AnnotationPrefix)
	if idx < 0 {
		return nil, false
	}

	s := strings.Split(comment[idx:], ":")
	if len(s) < 2 {
		return nil, false
	}

	command := strings.TrimSpace(s[1])
	opts := make(map[string]bool)
	if strings.Contains(command, "[") && strings.Contains(command, "]") {
		o := strings.Index(command, "[")
		c := strings.LastIndex(command, "]")

		options := command[o+1 : c]
		for _, opt := range strings.Split(options, ",") {
			opt = strings.TrimSpace(opt)
			opts[opt] = true
		}

		command = command[:o]

	}

	// command directive requires a :<target>
	var target string
	if HasTarget(command) && len(s) == 3 {
		target = strings.TrimSpace(s[2])
		sIdx := strings.Index(target, " ")
		if sIdx >= 0 {
			target = target[:sIdx]
		}
	}

	return &Annotation{command, opts, target}, true
}

// IsGlobalScopedCommand returns true if the command specified is allowed to be used in global scope, or
// scope not relative to a specific field or type
func IsGlobalScopedCommand(command string) bool {
	return strings.EqualFold(command, AnnotationGenerate) ||
		strings.EqualFold(command, AnnotationImport) ||
		strings.EqualFold(command, AnnotationSet) ||
		strings.EqualFold(command, AnnotationEndSet)
}

// HasTarget returns true if the annotation includes a command, options, and also a target
func HasTarget(command string) bool {
	return strings.EqualFold(command, AnnotationGenerate) ||
		strings.EqualFold(command, AnnotationImport)
}
