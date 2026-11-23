package java

import (
	"fmt"
	"testing"
	"time"

	"github.com/opencost/bingen/internal/types"
	"github.com/opencost/bingen/tests/aliasnil"
	"github.com/opencost/bingen/tests/container"
	"github.com/opencost/bingen/tests/shape"
	"github.com/opencost/bingen/tests/sttable"
	"github.com/opencost/bingen/tests/timerec"
)

// RoundTripJavaSourceFor generates a java program which will read a binary file from the first argument,
// decode it into the proper type, then re-encode it back into binary and write it to the second argument.
func RoundTripJavaSourceFor(baseJavaPackage, goPackage, typeName, harnessName string) string {
	fullPkg := fmt.Sprintf("%s.%s", baseJavaPackage, goPackage)
	encoder := fmt.Sprintf("%sEncoder", typeName)
	decoder := fmt.Sprintf("%sDecoder", typeName)

	return fmt.Sprintf(`import %s.%s;
import %s.%s;
import %s.%s;
import com.bingen.DecodingContext;
import com.bingen.EncodingContext;
import java.nio.file.Files;
import java.nio.file.Path;

public final class %s {
    public static void main(String[] args) throws Exception {
        byte[] in = Files.readAllBytes(Path.of(args[0]));
        %s c = %s.fromBytes(in);
        Files.write(Path.of(args[1]), %s.toBytes(c));
    }
}
`,
		fullPkg, typeName,
		fullPkg, decoder,
		fullPkg, encoder,
		harnessName,
		typeName, decoder,
		encoder,
	)
}

// TestRoundTrip_Decoder marshals a Container in go, has java decode and
// re-encode it, and asserts the bytes are identical and the values survived.
func TestRoundTrip_Decoder(t *testing.T) {
	const goPackage = "container"
	const harnessName = "DecodeReencodeMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   RoundTripJavaSourceFor(BaseJavaPackage, goPackage, "Container", harnessName),
	}

	testRunner := newRoundTripTestBuilder[container.Container](goPackage, javaMain).
		WithAssertion(func(t *testing.T, c1, c2 *container.Container) {
			if c2.Name != c1.Name || c2.Value != c1.Value || len(c2.Children) != len(c1.Children) {
				t.Fatalf("values not preserved: got %+v want %+v", c2, c1)
			}
		}).
		Build(t)

	defer testRunner.Cleanup()

	instance := &container.Container{
		Name:     "TestContainer",
		Children: []string{"a", "b", "c"},
		Value:    3.14,
	}

	testRunner.Run(instance)
}

// TestInterface_ShapeRoundTrip encodes a Drawing with a heterogeneous []Shape in
// go, has java decode + re-encode it through the TypeRegistry, and asserts the
// bytes are identical and the concrete types survived.
func TestRoundTripInterface_Shape(t *testing.T) {
	const goPackage = "shape"
	const harnessName = "DrawingReencodeMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   RoundTripJavaSourceFor(BaseJavaPackage, goPackage, "Drawing", harnessName),
	}
	testRunner := newRoundTripTestBuilder[shape.Drawing](goPackage, javaMain).
		WithAssertion(func(t *testing.T, t1 *shape.Drawing, t2 *shape.Drawing) {
			t.Helper()

			if t2.Name != t1.Name || len(t2.Shapes) != len(t1.Shapes) {
				t.Fatalf("unexpected drawing: %+v", t2)
			}
			c, ok := t2.Shapes[0].(*shape.Circle)
			if !ok || c.Radius != 2 {
				t.Fatalf("Shapes[0] = %#v, want *Circle{Radius:2}", t2.Shapes[0])
			}
			s, ok := t2.Shapes[1].(*shape.Square)
			if !ok || s.Side != 3 {
				t.Fatalf("Shapes[1] = %#v, want *Square{Side:3}", t2.Shapes[1])
			}
		}).
		Build(t)

	defer testRunner.Cleanup()

	instance := &shape.Drawing{
		Name: "pic",
		Shapes: []shape.Shape{
			&shape.Circle{
				Radius: 2,
			},
			&shape.Square{
				Side: 3,
			},
		},
	}

	testRunner.Run(instance)
}

// TestStringTable_GoToJavaToGo encodes a string-table Doc in go (repeated strings
// deduplicated into a BGST-prefixed table), has java decode + re-encode it, and
// asserts the bytes are identical and the values survived.
func TestRoundTrip_StringTable(t *testing.T) {
	const goPackage = "sttable"
	const harnessName = "DocReencodeMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   RoundTripJavaSourceFor(BaseJavaPackage, goPackage, "Doc", harnessName),
	}
	testRunner := newRoundTripTestBuilder[sttable.Doc](goPackage, javaMain).
		WithAssertion(func(t *testing.T, t1 *sttable.Doc, t2 *sttable.Doc) {
			t.Helper()

			if t2.Title != t1.Title || t2.Author != t1.Author || len(t2.Tags) != len(t1.Tags) {
				t.Fatalf("values not preserved: %+v", t2)
			}
			for i, tag := range t2.Tags {
				if t1.Tags[i] != tag {
					t.Fatalf("values not preserved: %+v", t2)
				}
			}
		}).
		Build(t)

	defer testRunner.Cleanup()

	instance := &sttable.Doc{
		Title: "report",
		Tags: []string{
			"a",
			"report",
			"b",
		},
		Author: "report",
	}

	testRunner.Run(instance)
}

// TestRoundTrip_Time encodes an Event with a UTC time.Time in go, has java
// decode + re-encode it through GoTime, and asserts the bytes are identical and
// the instant survived.
func TestRoundTrip_Time(t *testing.T) {
	const goPackage = "timerec"
	const harnessName = "EventReencodeMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   RoundTripJavaSourceFor(BaseJavaPackage, goPackage, "Event", harnessName),
	}

	testRunner := newRoundTripTestBuilder[timerec.Event](goPackage, javaMain).
		WithAssertion(func(t *testing.T, t1 *timerec.Event, t2 *timerec.Event) {
			t.Helper()

			if t2.Name != t1.Name || t2.Seq != t1.Seq {
				t.Fatalf("values not preserved: %+v (want %+v)", t2, t1)
			}
			if !t2.At.Equal(t1.At) {
				t.Fatalf("Time values not preserved. Got: %s. Expected: %s", t1.At, t2.At)
			}
		}).
		Build(t)

	defer testRunner.Cleanup()

	cases := map[string]time.Time{
		"utc_nanos":  time.Date(2023, 1, 2, 3, 4, 5, 123456789, time.UTC),
		"utc_flat":   time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
		"offset_ist": time.Date(2023, 1, 2, 3, 4, 5, 0, time.FixedZone("IST", 5*3600+30*60)),
		"offset_pst": time.Date(2023, 1, 2, 3, 4, 5, 500, time.FixedZone("PST", -8*3600)),
	}

	for name, at := range cases {
		t.Run(name, func(t *testing.T) {
			instance := &timerec.Event{
				Name: name,
				At:   at,
				Seq:  9,
			}

			testRunner.Run(instance)
		})
	}
}

// TestRoundTrip_NilableAlias covers an alias whose underlying type is nilable
// (Tags = []string). The nil-flag byte must be read symmetrically or the
// trailing Count field is corrupted.
func TestRoundTrip_NilableAlias(t *testing.T) {
	const goPackage = "aliasnil"
	const harnessName = "HolderReencodeMain"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   RoundTripJavaSourceFor(BaseJavaPackage, goPackage, "Holder", harnessName),
	}

	testRunner := newRoundTripTestBuilder[aliasnil.Holder](goPackage, javaMain).
		WithAssertion(func(t *testing.T, h1, h2 *aliasnil.Holder) {
			if h2.Name != "h" || len(h2.Tags) != 2 || h2.Tags[0] != "a" || h2.Tags[1] != "b" || h2.Count != 7 {
				t.Fatalf("values not preserved: %+v", h2)
			}
		}).
		Build(t)

	defer testRunner.Cleanup()

	instance := &aliasnil.Holder{
		Name:  "h",
		Tags:  aliasnil.Tags{"a", "b"},
		Count: 7,
	}

	testRunner.Run(instance)
}

// feedStruct builds a streamable Feed{Name string, Items []string, Score int}.
func feedStruct() *types.StructType {
	stringT := types.BasicTypes[types.TypeString]
	return &types.StructType{
		BasicType: types.NewBasicType("", "Feed", types.TypeStruct, false, false),
		Fields: []*types.StructField{
			{Name: "Name", Type: stringT},
			{Name: "Items", Type: &types.SliceType{
				BasicType: types.NewBasicType("", "[]string", types.TypeSlice, false, false),
				InnerType: stringT,
			}},
			{Name: "Score", Type: types.BasicTypes[types.TypeInt]},
		},
		Opts: &types.GenerateTypeOpts{SetName: "Feed", SetVersion: 1, IsStreamable: true},
	}
}

// TestStreamer_JavaPushRoundTrip encodes a Feed and streams it back through the
// generated push streamer, asserting the fields arrive in order with the Items
// slice flattened one element per callback.
func TestRoundTrip_JavaStreamPush(t *testing.T) {
	const feedStreamHarness = `import com.opencost.feed.Feed;
import com.opencost.feed.FeedEncoder;
import com.opencost.feed.FeedStreamer;
import java.util.ArrayList;
import java.util.List;

public final class FeedStreamMain {
    static void check(boolean ok, String what) {
        if (!ok) {
            System.err.println("FAIL: " + what);
            System.exit(1);
        }
    }

    public static void main(String[] args) {
        byte[] data = FeedEncoder.toBytes(new Feed("news", List.of("a", "b", "c"), 7));

        List<String> fields = new ArrayList<>();
        List<String> items = new ArrayList<>();
        int[] score = {-1};

        FeedStreamer.stream(data, (fi, v) -> {
            fields.add(fi.name());
            if (fi.name().equals("Items") && v != null) {
                items.add((String) v.value());
            }
            if (fi.name().equals("Score") && v != null) {
                score[0] = (Integer) v.value();
            }
            return true;
        });

        check(fields.equals(List.of("Name", "Items", "Items", "Items", "Score")), "fields=" + fields);
        check(items.equals(List.of("a", "b", "c")), "items=" + items);
        check(score[0] == 7, "score=" + score[0]);
        System.out.println("OK");
    }
}
`

	const goPackage = "feed"
	const harnessName = "FeedStreamMain"

	checkJava(t)

	tc := newTypesCollectionWith(feedStruct())
	out := buildJavaTestFromTypes(t, ".", goPackage, BaseJavaPackage, tc, &javaSource{
		FileName: harnessName,
		Source:   feedStreamHarness,
	})

	o, err := runJavaTest(t, out, harnessName)
	if err != nil {
		t.Fatalf("FeedStreamMain failed: %v\n%s", err, o)
	}
	t.Logf("%s", o)
}
