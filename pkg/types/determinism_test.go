package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findFixtureDir walks upward from the test working directory until it finds
// a sibling tests/<name> directory. The bundled bingen fixtures live under
// repo-root/tests/.
func findFixtureDir(t *testing.T, name string) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}

	for cur := wd; ; {
		candidate := filepath.Join(cur, "tests", name)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			t.Fatalf("could not find tests/%s starting from %s", name, wd)
		}
		cur = parent
	}
}

// TestLoadTypesIsDeterministic asserts that LoadTypes produces a deterministic
// ordering of types, imports, and version sets across runs. Without sorting,
// the generator's output would shift between runs because of randomized map
// iteration in pkg/meta.LoadAnnotations and pkg/types.LoadTypes.
//
// The check is repeated several times because Go's map iteration order is
// randomized per-iteration (not per-process), so a single pass is not enough
// to catch regressions.
func TestLoadTypesIsDeterministic(t *testing.T) {
	const iterations = 25

	// Use the opencost fixture because it has several files each contributing
	// annotated types and multiple version sets, so it exercises both the
	// package/file ordering in LoadTypes and the version set ordering in
	// LoadAnnotations.
	//
	// We copy the fixture into a t.TempDir() first so that when this test runs
	// in parallel with tests/generator_test.go (which rewrites
	// tests/opencost/opencost_codecs.go in place via runGenerator), the live
	// fixture flicker doesn't make our parse results racy. The fingerprint
	// depends on the parsed AST of the files under `dir`, so the directory
	// has to be stable for the duration of the test.
	src := findFixtureDir(t, "opencost")
	dir := copyFixture(t, src)

	first := fingerprintLoad(t, dir, "opencost", 16)

	for i := 1; i < iterations; i++ {
		got := fingerprintLoad(t, dir, "opencost", 16)
		if got != first {
			t.Fatalf("LoadTypes(opencost) was non-deterministic on iteration %d:\nfirst:    %s\niter %02d: %s",
				i, first, i, got)
		}
	}
}

// copyFixture clones every regular file in src into a fresh t.TempDir() and
// returns the destination path. It deliberately copies only files in the top
// level of src (the bingen fixtures don't nest), and skips any pre-existing
// .go.bak / editor backups that might be present.
func copyFixture(t *testing.T, src string) string {
	t.Helper()

	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("os.ReadDir(%s) failed: %v", src, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if isFixtureBackup(name) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatalf("os.ReadFile(%s) failed: %v", filepath.Join(src, name), err)
		}
		if err := os.WriteFile(filepath.Join(dst, name), data, 0o600); err != nil {
			t.Fatalf("os.WriteFile(%s) failed: %v", filepath.Join(dst, name), err)
		}
	}
	return dst
}

func isFixtureBackup(name string) bool {
	return strings.HasSuffix(name, ".go.bak") ||
		strings.HasSuffix(name, "~") ||
		strings.HasSuffix(name, ".swp")
}

func TestCopyFixtureSkipsBackups(t *testing.T) {
	src := t.TempDir()
	for _, name := range []string{"valid.go", "backup.go.bak", "emacs.go~", "swap.swp"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte(name), 0o600); err != nil {
			t.Fatalf("os.WriteFile(%s) failed: %v", name, err)
		}
	}

	dst := copyFixture(t, src)

	for _, name := range []string{"backup.go.bak", "emacs.go~", "swap.swp"} {
		if _, err := os.Stat(filepath.Join(dst, name)); err == nil {
			t.Fatalf("expected %q to be skipped, but it was copied", name)
		} else if !os.IsNotExist(err) {
			t.Fatalf("os.Stat(%s) failed: %v", name, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dst, "valid.go"))
	if err != nil {
		t.Fatalf("os.ReadFile(valid.go) failed: %v", err)
	}
	if string(data) != "valid.go" {
		t.Fatalf("valid.go: got %q, want %q", data, "valid.go")
	}
}

// fingerprintLoad runs LoadTypes once and produces a stable string capturing
// every ordered enumeration that flows into the generator (type names, import
// names, version-set name+version pairs).
func fingerprintLoad(t *testing.T, dir, pkg string, defaultVersion uint8) string {
	t.Helper()

	tc, err := LoadTypes(dir, pkg, defaultVersion)
	if err != nil {
		t.Fatalf("LoadTypes(%s, %s) failed: %v", dir, pkg, err)
	}

	var b strings.Builder

	b.WriteString("types:")
	for _, gt := range tc.Types() {
		fmt.Fprintf(&b, " %s", gt.Name())
	}
	b.WriteString("\nimports:")
	for _, im := range tc.Imports() {
		fmt.Fprintf(&b, " %s", im)
	}
	b.WriteString("\nsets:")
	for _, vs := range tc.VersionSets() {
		fmt.Fprintf(&b, " %s@%d", vs.Name(), vs.Version())
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
