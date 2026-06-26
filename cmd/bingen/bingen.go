package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/opencost/bingen/internal/generator"
	"github.com/opencost/bingen/internal/types"
)

const DefaultBufferPackage string = "github.com/opencost/bingen/pkg/util"

var (
	packageName = flag.String("package", "", "package name to generate binary codecs for")
	buffer      = flag.String("buffer", "github.com/opencost/bingen/pkg/util", "qualified package for the Buffer type")
	version     = flag.Uint("version", 1, "the versioning to use for the binary generator")
	//output      = flag.String("output", "", "output file name; default srcdir/<pkg>_codecs.go")
)

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

	if len(*buffer) == 0 {
		s := DefaultBufferPackage
		buffer = &s
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

	codecPath := path.Join(dir, fmt.Sprintf("%s_codecs.go", *packageName))
	if fileExists(codecPath) {
		if err := os.Remove(codecPath); err != nil {
			log.Fatalf("failed to remove existing codec file %q: %v", codecPath, err)
		}
	}

	defaultVersion := uint8(*version)

	tc, err := types.LoadTypes(dir, *packageName, defaultVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse @bingen annotations: %s", err)
		return
	}
	generator.Generate(dir, *packageName, *buffer, tc)
}
