package java

import (
	"path/filepath"
	"testing"
)

func TestFromConfig_DerivesJavaPackage(t *testing.T) {
	opts, err := FromConfig("container", map[string]string{
		OptBasePackage: "com.opencost",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The last segment of the java package must equal the go package name.
	if opts.JavaPackage != "com.opencost.container" {
		t.Fatalf("JavaPackage = %q, want %q", opts.JavaPackage, "com.opencost.container")
	}
	if opts.GoPackage != "container" {
		t.Fatalf("GoPackage = %q, want %q", opts.GoPackage, "container")
	}
	if opts.OutputDir != "." {
		t.Fatalf("OutputDir = %q, want default %q", opts.OutputDir, ".")
	}
}

func TestFromConfig_RuntimePackageDefaultAndOverride(t *testing.T) {
	opts, err := FromConfig("container", map[string]string{OptBasePackage: "com.opencost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.RuntimePackage != DefaultRuntimePackage {
		t.Fatalf("RuntimePackage = %q, want default %q", opts.RuntimePackage, DefaultRuntimePackage)
	}

	opts, err = FromConfig("container", map[string]string{
		OptBasePackage:    "com.opencost",
		OptRuntimePackage: "com.acme.rt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.RuntimePackage != "com.acme.rt" {
		t.Fatalf("RuntimePackage = %q, want %q", opts.RuntimePackage, "com.acme.rt")
	}

	if _, err := FromConfig("container", map[string]string{
		OptBasePackage:    "com.opencost",
		OptRuntimePackage: "com.9bad",
	}); err == nil {
		t.Fatal("expected error for invalid runtimePackage")
	}
}

func TestRuntimePackageDir(t *testing.T) {
	opts, err := FromConfig("container", map[string]string{
		OptBasePackage: "com.opencost",
		OptOutputDir:   "build/gen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("build", "gen", "com", "bingen")
	if got := opts.RuntimePackageDir(); got != want {
		t.Fatalf("RuntimePackageDir() = %q, want %q", got, want)
	}
}

func TestFromConfig_RequiresBasePackage(t *testing.T) {
	if _, err := FromConfig("container", map[string]string{}); err == nil {
		t.Fatal("expected error when basePackage is missing")
	}
}

func TestFromConfig_TrimsAndValidatesBasePackage(t *testing.T) {
	cases := map[string]bool{
		"com.opencost":      true,
		"  com.opencost  ":  true,
		"com.opencost.core": true,
		"com._x$y.z9":       true,
		"":                  false,
		"com..opencost":     false,
		"com.9bad":          false,
		"com.opencost.":     false,
		"com.op encost":     false,
		"com.class":         false, // reserved word
	}

	for base, wantOK := range cases {
		_, err := FromConfig("container", map[string]string{OptBasePackage: base})
		if wantOK && err != nil {
			t.Errorf("FromConfig(base=%q) unexpected error: %v", base, err)
		}
		if !wantOK && err == nil {
			t.Errorf("FromConfig(base=%q) expected error, got nil", base)
		}
	}
}

func TestFromConfig_HonorsOutputDirAndGoImport(t *testing.T) {
	opts, err := FromConfig("container", map[string]string{
		OptBasePackage: "com.opencost",
		OptOutputDir:   "build/gen",
		OptGoImport:    "github.com/opencost/bingen/tests/container",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.OutputDir != "build/gen" {
		t.Fatalf("OutputDir = %q, want %q", opts.OutputDir, "build/gen")
	}
	if opts.GoImport != "github.com/opencost/bingen/tests/container" {
		t.Fatalf("GoImport = %q", opts.GoImport)
	}
}

func TestPackageDir(t *testing.T) {
	opts, err := FromConfig("container", map[string]string{
		OptBasePackage: "com.opencost",
		OptOutputDir:   "build/gen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join("build", "gen", "com", "opencost", "container")
	if got := opts.PackageDir(); got != want {
		t.Fatalf("PackageDir() = %q, want %q", got, want)
	}
}
