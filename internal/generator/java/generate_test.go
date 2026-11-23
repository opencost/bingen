package java

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_EmitsPackageInfo(t *testing.T) {
	out := t.TempDir()

	opts, err := FromConfig("container", map[string]string{
		OptBasePackage: "com.opencost",
		OptOutputDir:   out,
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}

	if err := Generate(".", &fakeTypeCollection{}, opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	pkgInfo := filepath.Join(out, "com", "opencost", "container", "package-info.java")
	data, err := os.ReadFile(pkgInfo)
	if err != nil {
		t.Fatalf("reading generated package-info.java: %v", err)
	}

	if !strings.Contains(string(data), "package com.opencost.container;") {
		t.Fatalf("package-info.java missing package declaration, got:\n%s", data)
	}
}

func TestGenerate_NilOptions(t *testing.T) {
	if err := Generate(".", &fakeTypeCollection{}, nil); err == nil {
		t.Fatal("expected error for nil options")
	}
}
