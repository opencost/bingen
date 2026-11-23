package java

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_RecordCompiles(t *testing.T) {
	out := t.TempDir()
	opts, err := FromConfig("container", map[string]string{
		OptBasePackage: "com.opencost",
		OptOutputDir:   out,
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}

	st := containerStruct()
	if err := Generate(".", newTypesCollectionWith(st), opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	recordPath := filepath.Join(out, "com", "opencost", "container", "Container.java")
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("reading generated record: %v", err)
	}
	src := string(data)
	for _, want := range []string{
		"public record Container(String name, List<String> children, double value)",
		"public static Builder builder()",
		"public Container build()",
		"import java.util.List;",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("Container.java missing %q\n---\n%s", want, src)
		}
	}

	compileJavaTree(t, out)
}
