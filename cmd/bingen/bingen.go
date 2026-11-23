package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/opencost/bingen/internal/generator/golang"
	"github.com/opencost/bingen/internal/generator/java"
	"github.com/opencost/bingen/internal/types"
)

const DefaultBufferPackage string = "github.com/opencost/bingen/pkg/util"

// optionFlag collects repeatable -opt key=value language-specific options into
// a map.
type optionFlag map[string]string

func (o optionFlag) String() string {
	pairs := make([]string, 0, len(o))
	for k, v := range o {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ",")
}

func (o optionFlag) Set(s string) error {
	k, v, ok := strings.Cut(s, "=")
	if !ok {
		return fmt.Errorf("invalid -opt %q, expected key=value", s)
	}
	k = strings.TrimSpace(k)
	if k == "" {
		return fmt.Errorf("invalid -opt %q, empty key", s)
	}
	o[k] = strings.TrimSpace(v)
	return nil
}

var (
	packageName = flag.String("package", "", "package name to generate binary codecs for")
	buffer      = flag.String("buffer", DefaultBufferPackage, "[DEPRECATED] qualified package for the Buffer type")
	version     = flag.Uint("version", 1, "the versioning to use for the binary generator")
	lang        = flag.String("lang", "go", "target language: go|java")
	options     = optionFlag{}
	//output      = flag.String("output", "", "output file name; default srcdir/<pkg>_codecs.go")
)

func init() {
	flag.Var(options, "opt", "language-specific option key=value (repeatable)")
}

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of bingen:\n")
	fmt.Fprintf(os.Stderr, "\tbingen [flags] -package P [directory]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

// fileExists returns whether or not a file exists
func fileExists(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("bingen: ")
	flag.Usage = Usage
	flag.Parse()

	if len(*packageName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// check to see if buffer is non-default before printing deprecation message.
	if !strings.EqualFold(*buffer, DefaultBufferPackage) {
		fmt.Fprintf(os.Stderr, "DEPRECATED use of the -buffer option. Supplied value of: %s is ignored.\nBingen always uses github.com/opencost/bingen/pkg/util", *buffer)
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	var dir string
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
	} else {
		dir = filepath.Dir(args[0])
	}

	targetLang := strings.ToLower(strings.TrimSpace(*lang))

	// The go exporter overwrites a single <pkg>_codecs.go file; remove any stale
	// copy up front. Other languages manage their own output layout.
	if targetLang == "" || targetLang == "go" {
		codecPath := path.Join(dir, fmt.Sprintf("%s_codecs.go", *packageName))
		if fileExists(codecPath) {
			if err := os.Remove(codecPath); err != nil {
				log.Fatalf("failed to remove existing codec file %q: %v", codecPath, err)
			}
		}
	}

	defaultVersion := uint8(*version)

	tc, err := types.LoadTypes(dir, *packageName, defaultVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse @bingen annotations: %s", err)
		return
	}

	switch targetLang {
	case "", "go":
		golang.Generate(dir, *packageName, DefaultBufferPackage, tc)
	case "java":
		opts, err := java.FromConfig(*packageName, options)
		if err != nil {
			log.Fatalf("%s", err)
		}
		if err := java.Generate(dir, tc, opts); err != nil {
			log.Fatalf("%s", err)
		}
	default:
		log.Fatalf("unsupported -lang %q (want go|java)", *lang)
	}
}
