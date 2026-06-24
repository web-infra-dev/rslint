package typesnapshot

import (
	"reflect"
	"testing"
)

// TestEncodeBinaryRoundTrip checks Encode→Decode is lossless and that
// node2type comes back sorted by (tokenStart, end) for the worker's binary
// search. Covers every TypeBlock field (flags, name, isArray, member/arg/
// callSig id lists) plus a deliberately unsorted input.
func TestEncodeBinaryRoundTrip(t *testing.T) {
	snap := Snapshot{
		Node2Type: []NodeTypeEntry{
			{TokenStart: 10, End: 20, TypeID: 5},
			{TokenStart: 5, End: 8, TypeID: 3},
			{TokenStart: 10, End: 15, TypeID: 4}, // same start as the first, smaller end
		},
		Types: map[TypeID]TypeBlock{
			3: {ID: 3, Flags: 134217728, Name: "A | B", MemberTypes: []TypeID{4, 5}},
			4: {ID: 4, Flags: 4, Name: "string"},
			5: {ID: 5, Flags: 1, Name: "Foo<Bar>[]", IsArray: true, TypeArgs: []TypeID{4}, CallSigReturns: []TypeID{3}},
		},
		PrimTypes: PrimTypes{Undefined: 1, Null: 2},
	}

	got, ok := DecodeBinary(EncodeBinary(snap))
	if !ok {
		t.Fatal("DecodeBinary returned ok=false on a valid payload")
	}

	wantN2T := []NodeTypeEntry{
		{TokenStart: 5, End: 8, TypeID: 3},
		{TokenStart: 10, End: 15, TypeID: 4},
		{TokenStart: 10, End: 20, TypeID: 5},
	}
	if !reflect.DeepEqual(got.Node2Type, wantN2T) {
		t.Errorf("node2type round-trip mismatch:\n got %+v\nwant %+v", got.Node2Type, wantN2T)
	}
	if !reflect.DeepEqual(got.Types, snap.Types) {
		t.Errorf("types round-trip mismatch:\n got %+v\nwant %+v", got.Types, snap.Types)
	}
	if got.PrimTypes != snap.PrimTypes {
		t.Errorf("primTypes round-trip mismatch: got %+v want %+v", got.PrimTypes, snap.PrimTypes)
	}
}

// TestEncodeBinaryRoundTripFromBuild round-trips a snapshot built from real
// tsgo type-checking (the same fixture snapshot_test.go uses), so the encoder
// is exercised against actual type shapes, not just synthetic blocks.
func TestEncodeBinaryRoundTripFromBuild(t *testing.T) {
	snap := buildFixtureSnapshot(t)
	if len(snap.Node2Type) == 0 {
		t.Fatal("fixture snapshot is empty")
	}

	got, ok := DecodeBinary(EncodeBinary(snap))
	if !ok {
		t.Fatal("DecodeBinary returned ok=false on a built snapshot")
	}
	if len(got.Node2Type) != len(snap.Node2Type) {
		t.Errorf("node2type count: got %d want %d", len(got.Node2Type), len(snap.Node2Type))
	}
	if !reflect.DeepEqual(got.Types, snap.Types) {
		t.Errorf("types round-trip mismatch on built snapshot")
	}
	if got.PrimTypes != snap.PrimTypes {
		t.Errorf("primTypes mismatch: got %+v want %+v", got.PrimTypes, snap.PrimTypes)
	}
	// Every node2type entry must resolve to a present type block.
	for _, e := range got.Node2Type {
		if _, present := got.Types[e.TypeID]; !present {
			t.Errorf("node2type entry (%d,%d) → type %d missing from decoded types", e.TokenStart, e.End, e.TypeID)
		}
	}
}

// TestDecodeBinaryRejectsBadInput verifies the decoder fails closed (ok=false,
// no panic) on truncated or foreign payloads — the worker faces the same bytes
// over IPC and must never crash on a malformed frame.
func TestDecodeBinaryRejectsBadInput(t *testing.T) {
	valid := EncodeBinary(Snapshot{
		Node2Type: []NodeTypeEntry{{TokenStart: 1, End: 2, TypeID: 3}},
		Types:     map[TypeID]TypeBlock{3: {ID: 3, Name: "x"}},
	})

	cases := map[string][]byte{
		"empty":               {},
		"shorter than header": valid[:headerSize-1],
		"bad magic":           append([]byte{0, 0, 0, 0}, valid[4:]...),
		"truncated body":      valid[:len(valid)-1],
	}
	for name, data := range cases {
		if _, ok := DecodeBinary(data); ok {
			t.Errorf("%s: DecodeBinary returned ok=true, want false", name)
		}
	}
}
