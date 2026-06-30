package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"

	importsfmt "golang.org/x/tools/imports"

	"github.com/opencost/bingen/internal/types"
)

func Generate(dir string, pkg string, bufferImport string, tc types.TypeCollection) {
	var out bytes.Buffer

	ctxFactory := NewGeneratorContextFactory(ErrorHandler)
	targetTypes := tc.Types()

	// Sort Types to ensure Deterministic Ordering
	sort.Slice(targetTypes, func(i, j int) bool {
		return targetTypes[i].Name() < targetTypes[j].Name()
	})

	err := WriteSupportTemplate(&out, SupportParams{
		Package:         pkg,
		Imports:         tc.Imports(),
		BufferImport:    bufferImport,
		VersionSets:     tc.VersionSets(),
		Types:           targetTypes,
		StreamableTypes: toStreamableTypes(targetTypes),
	})
	if err != nil {
		panic(err)
	}

	for _, t := range targetTypes {
		if st, ok := t.(*types.StructType); ok {
			err = WriteMarshallerTemplate(&out, MarshallerParams{
				Context: ctxFactory.NewContext(st.Opts),
				Type:    t,
			})
			if err != nil {
				panic(err)
			}

			err = WriteUnmarshallerTemplate(&out, UnmarshallerParams{
				Context: ctxFactory.NewContext(st.Opts),
				Type:    t,
			})
			if err != nil {
				panic(err)
			}

			if st.Opts.IsStreamable {
				err = WriteStreamerTemplate(&out, StreamParams{
					Context: ctxFactory.NewContext(st.Opts).WithErrorHandler(StreamErrorHandler),
					Type:    t,
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}

	// Formatting can be turned on and off to see raw generated output for testing, but
	// this should generally be left on.
	const formatOn = true

	outFile := filepath.Join(dir, fmt.Sprintf("%s_codecs.go", pkg))

	var result []byte
	if formatOn {
		result, err = format.Source(out.Bytes())
		if err != nil {
			fmt.Println("Failed to format:", err)
			return
		}

		result, err = importsfmt.Process(outFile, result, nil)
		if err != nil {
			fmt.Println("Failed to process imports:", err)
			return
		}
	} else {
		result = out.Bytes()
	}

	err = os.WriteFile(outFile, result, 0600)
	if err != nil {
		return
	}
}

func ErrorHandler(newErr string) string {
	return fmt.Sprintf("return %s", newErr)
}

func StreamErrorHandler(errString string) string {
	return "stream.err = " + errString + "\nreturn\n"
}

func toStreamableTypes(ts []types.GenType) []*types.StructType {
	streamTypes := []*types.StructType{}

	for _, t := range ts {
		if st, ok := t.(*types.StructType); ok {
			if st.Opts.IsStreamable {
				streamTypes = append(streamTypes, st)
			}
		}
	}

	return streamTypes
}
