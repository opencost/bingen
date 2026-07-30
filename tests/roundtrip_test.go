package test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/opencost/bingen/tests/aliases"
	"github.com/opencost/bingen/tests/containerv2"
	"github.com/opencost/bingen/tests/opencost"
)

// TestRoundTrip_Aliases is a wire-level safety net for the generated aliases
// codec: it exercises alias types, maps, pointers, slices, and external define
// types. Encode -> decode must reproduce the original value. This guards the
// generator (template or core backend) against wire regressions.
func TestRoundTrip_Aliases(t *testing.T) {
	orig := aliases.NewGeneratedParent()

	data, err := orig.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	decoded := &aliases.Parent{}
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	if !reflect.DeepEqual(orig, decoded) {
		t.Fatalf("round-trip mismatch:\n orig:    %+v\n decoded: %+v", orig, decoded)
	}
}

// TestStream_Container exercises the generated streaming decoder at runtime: it
// marshals a Container, then streams its fields back, flattening the Children
// slice one element per yield. This validates the core-generated streamer, not
// just that it compiles.
func TestStream_Container(t *testing.T) {
	value := 2.5
	orig := &containerv2.Container{
		Name:     "box",
		Children: []string{"a", "b", "c"},
		Value:    &value,
	}

	data, err := orig.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	stream := containerv2.NewContainerStream(bytes.NewReader(data))
	defer stream.Close()

	var fields []string
	var children []string
	for fi, v := range stream.Stream() {
		fields = append(fields, fi.Name)
		if fi.Name == "Children" && v != nil {
			children = append(children, v.Value.(string))
		}
		if fi.Name == "Value" && v != nil {
			if got := *(v.Value.(*float64)); got != 2.5 {
				t.Errorf("Value = %v, want 2.5", got)
			}
		}
	}
	if err := stream.Error(); err != nil {
		t.Fatalf("stream error: %v", err)
	}

	wantFields := []string{"Name", "Children", "Children", "Children", "oldValue", "Value"}
	if !reflect.DeepEqual(fields, wantFields) {
		t.Fatalf("streamed fields = %v, want %v", fields, wantFields)
	}
	if !reflect.DeepEqual(children, []string{"a", "b", "c"}) {
		t.Fatalf("streamed children = %v, want [a b c]", children)
	}
}

// TestRoundTrip_Opencost is a wire-level safety net for the full opencost codec:
// an AssetSet exercises a string table, a map of interface values
// (map[string]Asset), nested time.Time (via Window), nested structs, and
// pre/post-process hooks. Encode -> decode -> re-encode must succeed and be
// length-stable (the encoded length is deterministic even though map and
// string-table ordering are not).
func TestRoundTrip_Opencost(t *testing.T) {
	start := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)
	win := opencost.NewWindow(&start, &end)

	orig := opencost.NewAssetSet(start, end,
		opencost.NewCloud("compute", "provider-1", start, end, win),
		opencost.NewDisk("disk-1", "cluster-a", "provider-2", start, end, win),
		opencost.NewNode("node-1", "cluster-a", "provider-3", start, end, win),
	)

	b1, err := orig.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	decoded := &opencost.AssetSet{}
	if err := decoded.UnmarshalBinary(b1); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	b2, err := decoded.MarshalBinary()
	if err != nil {
		t.Fatalf("re-encode MarshalBinary: %v", err)
	}

	if len(b1) != len(b2) {
		t.Fatalf("round-trip encoded length changed: first=%d re-encoded=%d", len(b1), len(b2))
	}
}
