package java

import (
	"testing"

	"github.com/opencost/bingen/internal/types"
)

// personPetHarness round-trips a struct with a map and a nested pointer struct
// entirely in java and asserts record equality.
const personPetHarness = `import com.example.model.Person;
import com.example.model.Pet;
import com.example.model.PersonEncoder;
import com.example.model.PersonDecoder;
import com.bingen.DecodingContext;
import com.bingen.EncodingContext;
import java.util.Map;

public final class PersonPetMain {
    static void check(boolean ok, String what) {
        if (!ok) {
            System.err.println("FAIL: " + what);
            System.exit(1);
        }
    }

    public static void main(String[] args) {
        Person p = Person.builder()
            .name("Ada")
            .age(36)
            .scores(Map.of("math", 95))
            .pet(Pet.builder().kind("cat").build())
            .build();

        EncodingContext ec = EncodingContext.create();
        PersonEncoder.INSTANCE.encode(p, ec);
        Person back = PersonDecoder.INSTANCE.decode(DecodingContext.fromBytes(ec.toBytes()));

        check(p.equals(back), "round-trip equality; got " + back);
        System.out.println("OK");
    }
}
`

// personPetStructs builds Person{Name, Age, Scores map[string]int, Pet *Pet} and
// Pet{Kind}, exercising maps and nested pointer structs.
func personPetStructs() []types.GenType {
	stringT := types.BasicTypes[types.TypeString]
	intT := types.BasicTypes[types.TypeInt]

	pet := &types.StructType{
		BasicType: types.NewBasicType("", "Pet", types.TypeStruct, false, false),
		Fields:    []*types.StructField{{Name: "Kind", Type: stringT}},
		Opts:      &types.GenerateTypeOpts{SetName: "Model", SetVersion: 1},
	}

	person := &types.StructType{
		BasicType: types.NewBasicType("", "Person", types.TypeStruct, false, false),
		Fields: []*types.StructField{
			{Name: "Name", Type: stringT},
			{Name: "Age", Type: intT},
			{Name: "Scores", Type: &types.MapType{
				BasicType: types.NewBasicType("", "map[string]int", types.TypeMap, false, false),
				KeyType:   stringT,
				ValueType: intT,
			}},
			{Name: "Pet", Type: pet.CreatePtr()},
		},
		Opts: &types.GenerateTypeOpts{SetName: "Model", SetVersion: 1},
	}

	return []types.GenType{person, pet}
}

// TestCodec_MapAndNestedStruct round-trips a map + nested pointer struct in java.
func TestCodec_MapAndNestedStruct(t *testing.T) {
	const goPackage = "model"
	const harnessName = "PersonPetMain"
	const exampleJavaPackageBase = "com.example"

	checkJava(t)

	javaMain := &javaSource{
		FileName: harnessName,
		Source:   personPetHarness,
	}

	out := buildJavaTestFromTypes(
		t,
		".",
		goPackage,
		exampleJavaPackageBase,
		newTypesCollectionWith(personPetStructs()...),
		javaMain,
	)

	if o, err := runJavaTest(t, out, harnessName); err != nil {
		t.Fatalf("PersonPetMain failed: %v\n%s", err, o)
	}
}
