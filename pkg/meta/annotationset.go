package meta

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
)

//--------------------------------------------------------------------------
//  MetaData
//--------------------------------------------------------------------------

// BingenAnnotated contains the globally annotated parsed data.
type BingenAnnotated struct {
	// Imports contains a list of any additional imports needed by the generated code.
	Imports []string

	// VersionSets contains a list of all version set definitions annotated for generation.
	VersionSets []VersionSet
}

//--------------------------------------------------------------------------
//  VersionSet
//--------------------------------------------------------------------------

// VersionSet is a named and versioned collection interface which is capable of collecting annotations
// within a package for query using commands and targets.
type VersionSet interface {
	// Name is the name of the set
	Name() string

	// Version is the version of the set
	Version() uint8

	// IsAnnotation is used to determine if a specific target exists for an annotation
	IsAnnotation(command string, target string) bool

	// IsCommandOption Determines if a specific target command has a specific option.
	IsCommandOption(command string, target string, option string) bool

	// IsGenerate is short-hand for IsAnnotation("generate", target)
	IsGenerate(target string) bool
}

//--------------------------------------------------------------------------
//  package types
//--------------------------------------------------------------------------

type annotationSet struct {
	name        string
	version     uint8
	annotations map[string]map[string]*Annotation
}

// newAnnotationSet creates a new set of annotations which are grouped by name and version
func newAnnotationSet(name string, version uint8) *annotationSet {
	return &annotationSet{
		name:    name,
		version: version,
		annotations: map[string]map[string]*Annotation{
			AnnotationGenerate: {},
		},
	}
}

// annotationCollector keeps a mapping of annotation commands to targets for key lookup
type annotationCollector struct {
	sets    map[string]*annotationSet
	current *annotationSet
	imports []string
}

// newAnnotationsCollector creates a new annotation collector with ignore and generate commands
// initialized
func newAnnotationsCollector() *annotationCollector {
	return &annotationCollector{
		sets:    map[string]*annotationSet{},
		current: nil,
		imports: []string{},
	}
}

// Name is the name of the set
func (ac *annotationSet) Name() string {
	return ac.name
}

// Version is the version of the set
func (ac *annotationSet) Version() uint8 {
	return ac.version
}

// IsAnnotation is used to determine if a specific target exists for an annotation
func (ac *annotationSet) IsAnnotation(command string, target string) bool {
	if set, ok := ac.annotations[command]; ok {
		_, ook := set[target]
		return ook
	}

	return false
}

// IsCommandOption Determines if a specific target command has a specific option.
func (ac *annotationSet) IsCommandOption(command string, target string, option string) bool {
	if set, ok := ac.annotations[command]; ok {
		if a, ook := set[target]; ook {
			return a.Options[option]
		}
		return false
	}

	return false
}

// IsGenerate is short-hand for IsAnnotation("generate", target)
func (ac *annotationSet) IsGenerate(target string) bool {
	return ac.IsAnnotation(AnnotationGenerate, target)
}

// Collect collects all annotations for an ast.File.
func (ac *annotationCollector) Collect(file *ast.File) error {
	for _, clist := range file.Comments {
		for _, comment := range clist.List {
			if a, ok := ParseAnnotation(comment.Text); ok {
				// we collect file scoped comments as global commands, any contextual
				// commands will be included in the comments list, but we'll ignore them
				// and rely on the context aware parsers to handle them
				if !IsGlobalScopedCommand(a.Command) {
					continue
				}

				switch a.Command {
				// AnnotationImport (@bingen:import)
				case AnnotationImport:
					ac.imports = append(ac.imports, a.Target)

				// AnnotationSet (@bingen:set)
				case AnnotationSet:
					if ac.current != nil {
						// pop the current annotation set if it's the default
						if ac.current.name == AnnotationDefaultSetName {
							ac.current = nil
						} else {
							return fmt.Errorf("found bingen set inside existing set scope")
						}
					}
					n, v, err := nameVersionFor(a)
					if err != nil {
						return err
					}
					ac.current = newAnnotationSet(n, v)
					ac.sets[n] = ac.current

				// AnnotationEndSet (@bingen:end)
				case AnnotationEndSet:
					ac.current = nil

				// Other
				default:
					if ac.current == nil {
						if _, ok := ac.sets[AnnotationDefaultSetName]; !ok {
							ac.sets[AnnotationDefaultSetName] = newAnnotationSet(AnnotationDefaultSetName, 0)
						}

						ac.current = ac.sets[AnnotationDefaultSetName]
					}

					if ac.current.annotations[a.Command] == nil {
						ac.current.annotations[a.Command] = map[string]*Annotation{}
					}

					ac.current.annotations[a.Command][a.Target] = a
				}
			}
		}
	}

	return nil
}

// nameVersionFor finds the name and version from the annotation options.
func nameVersionFor(a *Annotation) (string, uint8, error) {
	var name string
	var version uint8

	for option := range a.Options {
		strs := strings.Split(option, "=")
		if len(strs) < 2 {
			return "", 0, fmt.Errorf("parse error: failed to parse set option: %s", option)
		}

		prop := strings.TrimSpace(strs[0])
		value := strings.TrimSpace(strs[1])
		if prop == AnnotationSetName {
			name = value
		}
		if prop == AnnotationSetVersion {
			r, err := strconv.ParseUint(value, 10, 8)
			if err != nil {
				return "", 0, fmt.Errorf("parse error: illegal version value for set: %s", err)
			}
			version = uint8(r)
		}
	}

	if name == "" {
		return "", 0, fmt.Errorf("failed to parse name from @bingen:set option; use @bingen:set[name=] to apply a name")
	}
	// version will just inherit the default if 0

	return name, version, nil
}

//--------------------------------------------------------------------------
//  package funcs
//--------------------------------------------------------------------------

// LoadAnnotations collects all annotations from the files within the package and returns
// an slice of VersionSet implementations
//
//nolint:staticcheck // parser.ParseDir returns ast.Package map; migrating to go/types is a larger refactor.
func LoadAnnotations(packages map[string]*ast.Package, defaultVersion uint8) (*BingenAnnotated, error) {
	ac := newAnnotationsCollector()

	for _, v := range packages {
		for _, file := range v.Files {
			err := ac.Collect(file)
			if err != nil {
				return nil, err
			}
		}
	}

	sets := []VersionSet{}
	for _, v := range ac.sets {
		if v.name == AnnotationDefaultSetName {
			v.version = defaultVersion
		}
		sets = append(sets, v)
	}

	imports := []string{}
	imports = append(imports, ac.imports...)

	return &BingenAnnotated{
		Imports:     imports,
		VersionSets: sets,
	}, nil
}
