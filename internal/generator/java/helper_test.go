package java

import (
	"bytes"
	"encoding"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencost/bingen/internal/meta"
	"github.com/opencost/bingen/internal/types"
)

var (
	javaExec string
	javaCC   string
)

func init() {
	var err error
	javaExec, err = exec.LookPath("java")
	if err != nil {
		javaExec = ""
	}

	javaCC, err = exec.LookPath("javac")
	if err != nil {
		javaCC = ""
	}
}

const BaseJavaPackage string = "com.opencost"

// containerStruct builds the IR for the container test struct without touching
// the filesystem: Name string, Children []string, Value float64.
func containerStruct() *types.StructType {
	stringT := types.BasicTypes[types.TypeString]
	return &types.StructType{
		BasicType: types.NewBasicType("", "Container", types.TypeStruct, false, false),
		Fields: []*types.StructField{
			{Name: "Name", Type: stringT},
			{Name: "Children", Type: &types.SliceType{
				BasicType: types.NewBasicType("", "[]string", types.TypeSlice, false, false),
				InnerType: stringT,
			}},
			{Name: "Value", Type: types.BasicTypes[types.TypeFloat64]},
		},
		Opts: &types.GenerateTypeOpts{SetName: "ContainerExample", SetVersion: 1},
	}
}

// fakeTypeCollection is a minimal TypeCollection used to drive the generation
// pipeline without parsing a real go package.
type fakeTypeCollection struct {
	types []types.GenType
}

func (f *fakeTypeCollection) AddStructType(*types.AnnotatedType, []*types.AnnotatedField) {}

func (f *fakeTypeCollection) AddInterface(*types.AnnotatedType) {}

func (f *fakeTypeCollection) AddAlias(*types.AnnotatedType, bool) {}

func (f *fakeTypeCollection) Types() []types.GenType {
	return f.types
}

func (f *fakeTypeCollection) Imports() []string {
	return nil
}

func (f *fakeTypeCollection) VersionSets() []meta.VersionSet {
	return nil
}

func newTypesCollectionWith(typs ...types.GenType) types.TypeCollection {
	return &fakeTypeCollection{
		types: typs,
	}
}

// javaSource represents a java source file with a test harness.
type javaSource struct {
	FileName string
	Source   string
}

func (js *javaSource) HarnessName() string {
	return toHarnessName(js.FileName)
}

// test helper which writes the java sources to a specific directory
func writeJavaSource(t *testing.T, dir string, file *javaSource, other ...*javaSource) {
	t.Helper()

	files := append([]*javaSource{file}, other...)
	for _, jfile := range files {
		fileName := jfile.FileName
		if !strings.Contains(fileName, ".java") {
			fileName += ".java"
		}

		fullPath := filepath.Join(dir, fileName)
		if err := os.WriteFile(fullPath, []byte(jfile.Source), 0o600); err != nil {
			t.Fatalf("writing harness: %v", err)
		}
	}
}

// test helper which checks to see if java and javac are reachable on the path.
// if not, tests are skipped
func checkJava(t *testing.T) {
	t.Helper()

	if javaExec == "" {
		t.Skip("java not found on PATH; skipping java run test")
	}

	if javaCC == "" {
		t.Skip("javac not found on PATH; skipping java compile test")
	}
}

// findJavaSources returns every .java file beneath root.
func findJavaSources(t *testing.T, root string) []string {
	t.Helper()

	var sources []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".java") {
			sources = append(sources, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("walking %q: %v", root, err)
	}

	return sources
}

// compileJavaTree compiles every .java file beneath root, skipping the test if
// javac is unavailable.
func compileJavaTree(t *testing.T, root string) {
	t.Helper()

	sources := findJavaSources(t, root)
	if len(sources) == 0 {
		t.Fatal("no .java sources generated")
	}

	compileJava(t, root, sources)
}

func compileJava(t *testing.T, dir string, sources []string) {
	t.Helper()

	classes := filepath.Join(dir, "classes")
	if err := os.MkdirAll(classes, 0o750); err != nil {
		t.Fatalf("mkdir classes: %v", err)
	}

	args := append([]string{"-encoding", "UTF-8", "-d", classes}, sources...)
	cmd := exec.Command(javaCC, args...)
	cmd.Dir = dir

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("javac failed: %v\n%s", err, out)
	}
}

// builds the java runtime files and returns the temp output directory
func createAndEmitRuntime(t *testing.T, goPackage, baseJavaPackage string) string {
	t.Helper()

	dir := t.TempDir()

	opts, err := FromConfig(goPackage, map[string]string{
		OptBasePackage: baseJavaPackage,
		OptOutputDir:   dir,
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}

	// Emit the generated runtime under com/bingen.
	if err := emitRuntime(opts); err != nil {
		t.Fatalf("emitRuntime: %v", err)
	}
	return dir
}

// test helper
func buildJavaTestFromTypes(t *testing.T, goSrcDir string, goPackage string, baseJavaPackage string, tc types.TypeCollection, javaSrc *javaSource, other ...*javaSource) string {
	t.Helper()

	out := t.TempDir()
	opts, err := FromConfig(goPackage, map[string]string{
		OptBasePackage: baseJavaPackage,
		OptOutputDir:   out,
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if err := Generate(goSrcDir, tc, opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	writeJavaSource(t, out, javaSrc, other...)
	compileJavaTree(t, out)

	return out
}

// loads the test types in <root>/tests/<goPackage> and returns the directory and type collection
func loadTestTypes(t *testing.T, goPackage string) (string, types.TypeCollection) {
	t.Helper()

	// load types from the source go directory (<root>/tests/<package_name>)
	dir := goSourceToDirectory(t, goPackage)
	tc, err := types.LoadTypes(dir, goPackage, 1)
	if err != nil {
		t.Fatalf("LoadTypes: %v", err)
	}

	return dir, tc
}

func buildJavaTest(t *testing.T, goPackage string, javaSrc *javaSource, other ...*javaSource) string {
	t.Helper()

	dir, tc := loadTestTypes(t, goPackage)
	return buildJavaTestFromTypes(t, dir, goPackage, BaseJavaPackage, tc, javaSrc, other...)
}

func runJavaTest(t *testing.T, runDir string, harnessName string, args ...string) ([]byte, error) {
	t.Helper()

	allArgs := append(
		[]string{
			"-cp",
			"classes",
			harnessName,
		},
		args...,
	)

	run := exec.Command(javaExec, allArgs...)
	run.Dir = runDir

	return run.CombinedOutput()
}

func buildAndRunJavaTest(t *testing.T, goSrcTestDir string, javaSrc *javaSource) {
	t.Helper()

	out := buildJavaTest(t, goSrcTestDir, javaSrc)
	o, err := runJavaTest(t, out, javaSrc.HarnessName())
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", javaSrc.HarnessName(), err, o)
	}
	t.Logf("%s", o)
}

func goSourceToDirectory(t *testing.T, goSrc string) string {
	t.Helper()

	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "tests", goSrc))
	if err != nil {
		t.Fatalf("abs go-src dir: %v", err)
	}
	return dir
}

func toHarnessName(name string) string {
	r, _ := strings.CutSuffix(name, ".java")
	return r
}

type binaryPtr[T any] interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	*T
}

type roundTripTestBuilder[T any, U binaryPtr[T]] struct {
	goPackage   string
	harnessName string
	javaMain    *javaSource
	sources     []*javaSource
	tc          types.TypeCollection
	assertFn    func(*testing.T, *T, *T)
}

func newRoundTripTestBuilder[T any, U binaryPtr[T]](goPackage string, javaMain *javaSource) *roundTripTestBuilder[T, U] {
	return &roundTripTestBuilder[T, U]{
		goPackage:   goPackage,
		harnessName: javaMain.HarnessName(),
		javaMain:    javaMain,
		sources:     []*javaSource{},
	}
}

func (rtt *roundTripTestBuilder[T, U]) AddSource(javaSrc *javaSource) *roundTripTestBuilder[T, U] {
	rtt.sources = append(rtt.sources, javaSrc)
	return rtt
}

func (rtt *roundTripTestBuilder[T, U]) WithIRTypes(tc types.TypeCollection) *roundTripTestBuilder[T, U] {
	rtt.tc = tc
	return rtt
}

func (rtt *roundTripTestBuilder[T, U]) WithAssertion(assertFn func(*testing.T, *T, *T)) *roundTripTestBuilder[T, U] {
	rtt.assertFn = assertFn
	return rtt
}

func (rtt *roundTripTestBuilder[T, U]) Build(t *testing.T) *roundTripTest[T, U] {
	var out string
	if rtt.tc != nil {
		out = buildJavaTestFromTypes(t, ".", rtt.goPackage, BaseJavaPackage, rtt.tc, rtt.javaMain, rtt.sources...)
	} else {
		out = buildJavaTest(t, rtt.goPackage, rtt.javaMain, rtt.sources...)
	}

	return &roundTripTest[T, U]{
		t:           t,
		goPackage:   rtt.goPackage,
		harnessName: rtt.harnessName,
		outDir:      out,
		assertFn:    rtt.assertFn,
	}
}

type roundTripTest[T any, U binaryPtr[T]] struct {
	t           *testing.T
	outDir      string
	goPackage   string
	harnessName string
	assertFn    func(*testing.T, *T, *T)
}

func (rtt *roundTripTest[T, U]) Run(instance *T) {
	t := rtt.t
	t.Helper()

	inPath := filepath.Join(rtt.outDir, "in.bin")
	outPath := filepath.Join(rtt.outDir, "out.bin")

	var orig U = instance
	goBytes, err := orig.MarshalBinary()
	if err != nil {
		t.Fatalf("go MarshalBinary: %v", err)
		return
	}

	if err := os.WriteFile(inPath, goBytes, 0o600); err != nil {
		t.Fatalf("writing in.bin: %v", err)
		return
	}

	if o, err := runJavaTest(t, rtt.outDir, rtt.harnessName, "in.bin", "out.bin"); err != nil {
		t.Fatalf("%s failed: %v\n%s", rtt.harnessName, err, o)
		return
	}

	reBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading out.bin: %v", err)
	}
	if !bytes.Equal(goBytes, reBytes) {
		t.Fatalf("roundtrip byte comparison failed\n go:   %x\n java: %x", goBytes, reBytes)
		return
	}

	var back U = new(T)
	if err := back.UnmarshalBinary(reBytes); err != nil {
		t.Fatalf("go decode of java-reencoded bytes: %v", err)
		return
	}

	if rtt.assertFn != nil {
		rtt.assertFn(t, orig, back)
	}
}

func (rtt *roundTripTest[T, U]) Cleanup() {
	_ = os.RemoveAll(rtt.outDir)
}
