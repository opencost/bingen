package java

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opencost/bingen/tests/container"
)

// containerHooksSource is the user-supplied migration hook, mirroring the go
// migrateContainer func: when the legacy oldValue is set, promote it to Value.
const containerHooksSource = `package com.opencost.containerv2;

public final class ContainerHooks {
    private ContainerHooks() {
    }

    public static Container migrate(Container c, int from, int to) {
        if (c.oldValue() != 0.0) {
            return new Container(c.name(), c.children(), c.oldValue(), c.oldValue());
        }
        return c;
    }
}
`

const migrateMainSource = `import com.opencost.containerv2.Container;
import com.opencost.containerv2.ContainerDecoder;
import java.nio.file.Files;
import java.nio.file.Path;

public final class MigrateMain {
    static void check(boolean ok, String what) {
        if (!ok) {
            System.err.println("FAIL: " + what);
            System.exit(1);
        }
    }

    public static void main(String[] args) throws Exception {
        byte[] in = Files.readAllBytes(Path.of(args[0]));
        Container c = ContainerDecoder.fromBytes(in);
        check(c.name().equals("legacy"), "name=" + c.name());
        check(c.oldValue() == 42.5, "oldValue=" + c.oldValue());
        check(c.value() != null && c.value() == 42.5, "value=" + c.value());
        System.out.println("OK");
    }
}
`

// TestVersioning_MigrateV1UnderV2 encodes a v1 Container in go, then decodes it
// with the java v2 codec: the field-version gate defaults the new Value to null,
// oldValue absorbs the v1 float, and the migrate hook promotes it to Value.
func TestVersioning_MigrateV1UnderV2(t *testing.T) {
	checkJava(t)

	// compile java test harness

	javaMain := &javaSource{
		FileName: "MigrateMain.java",
		Source:   migrateMainSource,
	}
	javaHook := &javaSource{
		FileName: "ContainerHooks.java",
		Source:   containerHooksSource,
	}

	v2Dir, tc := loadTestTypes(t, "containerv2")
	out := buildJavaTestFromTypes(t, v2Dir, "containerv2", BaseJavaPackage, tc, javaMain, javaHook)

	// build a v1 container, marshal binary, write to java test harness dir
	v1 := container.Container{Name: "legacy", Children: []string{"a", "b"}, Value: 42.5}
	v1Bytes, err := v1.MarshalBinary()
	if err != nil {
		t.Fatalf("go v1 MarshalBinary: %v", err)
	}

	inPath := filepath.Join(out, "v1.bin")
	if err := os.WriteFile(inPath, v1Bytes, 0o600); err != nil {
		t.Fatalf("writing v1.bin: %v", err)
	}

	o, err := runJavaTest(t, out, javaMain.HarnessName(), "v1.bin")
	if err != nil {
		t.Fatalf("MigrateMain failed: %v\n%s", err, o)
	}

	t.Logf("%s", string(o))
}
