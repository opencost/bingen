package java

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opencost/bingen/tests/container"
)

// encodeMainHarness builds a Container via the generated builder, encodes it
// with the generated encoder, and writes the bytes to args[0].
const encodeMainHarness = `import com.opencost.container.Container;
import com.opencost.container.ContainerEncoder;
import com.bingen.EncodingContext;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

public final class EncodeMain {
    public static void main(String[] args) throws Exception {
        Container c = Container.builder()
            .name("TestContainer")
            .children(List.of("a", "b", "c"))
            .value(3.14)
            .build();

        EncodingContext ctx = EncodingContext.create();
        ContainerEncoder.INSTANCE.encode(c, ctx);
        Files.write(Path.of(args[0]), ctx.toBytes());
    }
}
`

// TestEncoder_JavaToGoRoundTrip generates the java encoder for Container, runs
// it to produce bytes, then decodes those bytes with the real go container
// codec and asserts the values survived the crossing.
func TestEncoder_JavaToGoRoundTrip(t *testing.T) {
	const goPackage = "container"
	const harnessName = "EncodeMain"

	checkJava(t)

	tc := newTypesCollectionWith(containerStruct())
	out := buildJavaTestFromTypes(t, ".", goPackage, BaseJavaPackage, tc, &javaSource{
		FileName: harnessName,
		Source:   encodeMainHarness,
	})

	if o, err := runJavaTest(t, out, harnessName, "encoded.bin"); err != nil {
		t.Fatalf("EncodeMain failed: %v\n%s", err, o)
	}

	binPath := filepath.Join(out, "encoded.bin")
	data, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("reading encoded bytes: %v", err)
	}

	var c container.Container
	if err := c.UnmarshalBinary(data); err != nil {
		t.Fatalf("go UnmarshalBinary of java-encoded bytes failed: %v", err)
	}

	if c.Name != "TestContainer" {
		t.Errorf("Name = %q, want %q", c.Name, "TestContainer")
	}
	if want := []string{"a", "b", "c"}; len(c.Children) != len(want) {
		t.Fatalf("Children = %v, want %v", c.Children, want)
	} else {
		for i, v := range want {
			if c.Children[i] != v {
				t.Errorf("Children[%d] = %q, want %q", i, c.Children[i], v)
			}
		}
	}
	if c.Value != 3.14 {
		t.Errorf("Value = %v, want 3.14", c.Value)
	}
}
