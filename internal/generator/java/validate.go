package java

import (
	"fmt"
	"strings"
)

var javaReserved []string = []string{
	"abstract",
	"assert",
	"boolean",
	"break",
	"byte",
	"case",
	"catch",
	"char",
	"class",
	"const",
	"continue",
	"default",
	"do",
	"double",
	"else",
	"enum",
	"extends",
	"final",
	"finally",
	"float",
	"for",
	"goto",
	"if",
	"implements",
	"import",
	"instanceof",
	"int",
	"interface",
	"long",
	"native",
	"new",
	"package",
	"private",
	"protected",
	"public",
	"return",
	"short",
	"static",
	"strictfp",
	"super",
	"switch",
	"synchronized",
	"this",
	"throw",
	"throws",
	"transient",
	"try",
	"void",
	"volatile",
	"while",
	"true",
	"false",
	"null",
	"_",
}

// javaReservedWords are the Java keywords (and literals) that may not appear as
// a package segment.
var javaReservedWords map[string]bool

func init() {
	javaReservedWords = make(map[string]bool, len(javaReserved))
	for _, word := range javaReserved {
		javaReservedWords[word] = true
	}
}

// validateJavaPackage ensures every segment is a legal Java identifier that is
// not a reserved word.
func validateJavaPackage(pkg string) error {
	for seg := range strings.SplitSeq(pkg, ".") {
		if !isJavaIdentifier(seg) {
			return fmt.Errorf("segment %q is not a valid java identifier", seg)
		}
		if javaReservedWords[seg] {
			return fmt.Errorf("segment %q is a reserved word", seg)
		}
	}
	return nil
}

// isJavaIdentifier reports whether s is a syntactically valid Java identifier.
func isJavaIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !isJavaIdentStart(r) {
				return false
			}
			continue
		}
		if !isJavaIdentStart(r) && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func isJavaIdentStart(r rune) bool {
	return r == '_' || r == '$' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z')
}
