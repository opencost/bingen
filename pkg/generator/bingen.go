package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"

	"github.com/opencost/bingen/pkg/types"
)

func Generate(dir string, pkg string, bufferImport string, tc types.TypeCollection) (err error) {
	// Recover panics from template rendering or type-resolution code so callers
	// see a normal error rather than a process crash.
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("bingen: code generation panicked: %w", e)
				return
			}
			err = fmt.Errorf("bingen: code generation panicked: %v", r)
		}
	}()

	var out bytes.Buffer

	ctxFactory := NewGeneratorContextFactory(ErrorHandler)
	targetTypes := tc.Types()

	// Sort Types to ensure Deterministic Ordering
	sort.Slice(targetTypes, func(i, j int) bool {
		return targetTypes[i].Name() < targetTypes[j].Name()
	})

	if err = WriteSupportTemplate(&out, SupportParams{
		Package:         pkg,
		Imports:         tc.Imports(),
		BufferImport:    bufferImport,
		VersionSets:     tc.VersionSets(),
		Types:           targetTypes,
		StreamableTypes: toStreamableTypes(targetTypes),
	}); err != nil {
		return fmt.Errorf("bingen: failed to write support template: %w", err)
	}

	for _, t := range targetTypes {
		if st, ok := t.(*types.StructType); ok {
			if err = WriteMarshallerTemplate(&out, MarshallerParams{
				Context: ctxFactory.NewContext(st.Opts),
				Type:    t,
			}); err != nil {
				return fmt.Errorf("bingen: failed to write marshaller for %s: %w", t.Name(), err)
			}

			if err = WriteUnmarshallerTemplate(&out, UnmarshallerParams{
				Context: ctxFactory.NewContext(st.Opts),
				Type:    t,
			}); err != nil {
				return fmt.Errorf("bingen: failed to write unmarshaller for %s: %w", t.Name(), err)
			}

			if st.Opts.IsStreamable {
				if err = WriteStreamerTemplate(&out, StreamParams{
					Context: ctxFactory.NewContext(st.Opts).WithErrorHandler(StreamErrorHandler),
					Type:    t,
				}); err != nil {
					return fmt.Errorf("bingen: failed to write streamer for %s: %w", t.Name(), err)
				}
			}
		}
	}

	// Formatting can be turned on and off to see raw generated output for testing, but
	// this should generally be left on.
	const formatOn = true

	var result []byte
	if formatOn {
		result, err = format.Source(out.Bytes())
		if err != nil {
			return fmt.Errorf("bingen: failed to format generated source: %w", err)
		}
	} else {
		result = out.Bytes()
	}

	outPath := fmt.Sprintf("%s/%s_codecs.go", dir, pkg)
	if err = os.WriteFile(outPath, result, 0600); err != nil {
		return fmt.Errorf("bingen: failed to write codec file %q: %w", outPath, err)
	}
	return nil
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
