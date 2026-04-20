package test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/opencost/bingen/pkg/generator"
	"github.com/opencost/bingen/pkg/types"
)

var (
	findTestDir sync.Once
	testDir     string
)

func getTestDir() string {
	findTestDir.Do(func() {
		// Start with current working directory
		wd, err := os.Getwd()
		if err != nil {
			return
		}

		tDir := filepath.Join(wd, "tests")
		if dirExists(tDir) {
			testDir = tDir
			return
		}

		if filepath.Base(wd) == "tests" {
			testDir = wd
			return
		}

		// Walk up parent directories looking for a "tests" sibling directory.
		current := wd
		for {
			candidate := filepath.Join(current, "tests")
			if dirExists(candidate) {
				testDir = candidate
				return
			}

			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}

		testDir = ""
	})
	return testDir
}

const (
	V          uint8  = uint8(16)
	BinGenUtil string = "github.com/opencost/bingen/pkg/util"
)

func fileExists(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

func dirExists(dir string) bool {
	stat, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

// basic generator run providing a directory and package
func runGenerator(dir string, pkg string) error {
	codecPath := path.Join(dir, fmt.Sprintf("%s_codecs.go", pkg))
	if fileExists(codecPath) {
		os.Remove(codecPath)
	}

	tc, err := types.LoadTypes(dir, pkg, V)
	if err != nil {
		return err
	}

	generator.Generate(dir, pkg, BinGenUtil, tc)
	return nil
}

func TestGenerateAliasBinCodecs(t *testing.T) {
	td := getTestDir()

	err := runGenerator(filepath.Join(td, "aliases"), "aliases")
	if err != nil {
		t.Errorf("\n%s", err)
	}
}

func TestGenerateContainerBinCodecs(t *testing.T) {
	td := getTestDir()

	err := runGenerator(filepath.Join(td, "container"), "container")
	if err != nil {
		t.Errorf("Failed to generate container pkg: %s", err)
	}

	err = runGenerator(filepath.Join(td, "containerv2"), "containerv2")
	if err != nil {
		t.Errorf("Failed to generated container-v2: %s", err)
	}
}

func TestGenerateOpencostBinCodecs(t *testing.T) {
	td := getTestDir()

	err := runGenerator(filepath.Join(td, "opencost"), "opencost")
	if err != nil {
		t.Errorf("\n%s", err)
	}
}

func TestGenerateFailingBinCodecs(t *testing.T) {
	td := getTestDir()

	err := runGenerator(filepath.Join(td, "failing"), "failing")
	t.Logf("\n%s", err)

	if err == nil {
		t.Errorf("Expected Failures, but completed successfully.")
	}
}

func TestImportFromPath(t *testing.T) {
	const FullPath = "github.com/opencost/opencost-special-path/pkg/opencost"
	ip := types.ImportFromPath(FullPath)
	if ip.Name != "opencost" {
		t.Errorf("Expected opencost, got %s", ip.Name)
	}
	if ip.Path != FullPath {
		t.Errorf("Expected %s, got %s", FullPath, ip.Path)
	}

	const ShortPath = "opencost"
	ip = types.ImportFromPath(ShortPath)
	if ip.Name != "opencost" {
		t.Errorf("Expected opencost, got %s", ip.Name)
	}
	if ip.Path != ShortPath {
		t.Errorf("Expected %s, got %s", ShortPath, ip.Path)
	}
}
