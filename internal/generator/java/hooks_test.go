package java

import (
	"testing"

	"github.com/opencost/bingen/internal/types"
)

const noteHooksSource = `package com.opencost.note;

public final class NoteHooks {
    private NoteHooks() {
    }

    // preProcess derives Length from Text before encoding.
    public static Note preProcess(Note n) {
        return new Note(n.text(), n.text().length());
    }

    // postProcess upper-cases Text after decoding.
    public static Note postProcess(Note n) {
        return new Note(n.text().toUpperCase(), n.length());
    }
}
`

const noteMainSource = `import com.opencost.note.Note;
import com.opencost.note.NoteEncoder;
import com.opencost.note.NoteDecoder;

public final class NoteMain {
    static void check(boolean ok, String what) {
        if (!ok) {
            System.err.println("FAIL: " + what);
            System.exit(1);
        }
    }

    public static void main(String[] args) {
        // Length starts at 0; preProcess should set it to text length (3) on encode.
        byte[] bytes = NoteEncoder.toBytes(new Note("abc", 0));
        Note out = NoteDecoder.fromBytes(bytes);
        check(out.text().equals("ABC"), "postProcess upper-case; text=" + out.text());
        check(out.length() == 3, "preProcess length; length=" + out.length());
        System.out.println("OK");
    }
}
`

func noteStruct() *types.StructType {
	return &types.StructType{
		BasicType: types.NewBasicType("", "Note", types.TypeStruct, false, false),
		Fields: []*types.StructField{
			{Name: "Text", Type: types.BasicTypes[types.TypeString]},
			{Name: "Length", Type: types.BasicTypes[types.TypeInt]},
		},
		Opts: &types.GenerateTypeOpts{
			SetName:       "Note",
			SetVersion:    1,
			IsPreProcess:  true,
			IsPostProcess: true,
		},
	}
}

// TestHooks_PreAndPostProcess verifies the encoder pre-process and decoder
// post-process hook points fire around the wire.
func TestHooks_PreAndPostProcess(t *testing.T) {
	const goPackage = "note"
	const harnessName = "NoteMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   noteMainSource,
	}

	javaHooks := &javaSource{
		FileName: "NoteHooks.java",
		Source:   noteHooksSource,
	}

	out := buildJavaTestFromTypes(
		t,
		".",
		goPackage,
		BaseJavaPackage,
		newTypesCollectionWith(noteStruct()),
		javaMain,
		javaHooks,
	)

	o, err := runJavaTest(t, out, harnessName)
	if err != nil {
		t.Fatalf("NoteMain failed: %v\n%s", err, o)
	}
	t.Logf("%s", string(o))
}
