package typeservice

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func buildBinaryHeader(session string, gen uint32, skip bool, op byte) []byte {
	var b []byte
	b = binary.LittleEndian.AppendUint16(b, uint16(len(session)))
	b = append(b, session...)
	b = binary.LittleEndian.AppendUint32(b, gen)
	var flags byte
	if skip {
		flags |= binaryFlagSkipRecords
	}
	b = append(b, flags, op)
	return b
}

func appendBinaryNodeRef(b []byte, fileIdx uint32, file string, index uint32, tok *SyntaxToken) []byte {
	b = binary.LittleEndian.AppendUint32(b, fileIdx)
	if fileIdx == BinaryInlineFileSentinel {
		b = binary.LittleEndian.AppendUint32(b, uint32(len(file)))
		b = append(b, file...)
	}
	b = binary.LittleEndian.AppendUint32(b, index)
	if tok == nil {
		b = append(b, 0)
	} else {
		b = append(b, 1)
		b = binary.LittleEndian.AppendUint32(b, uint32(int32(tok.Kind)))
		b = binary.LittleEndian.AppendUint32(b, uint32(int32(tok.Pos)))
		b = binary.LittleEndian.AppendUint32(b, uint32(int32(tok.End)))
	}
	return b
}

func TestDecodeBinaryQueryShapesAndEdgeCases(t *testing.T) {
	// 1. NodeRequest with fileIdx + SyntaxToken + skipRecords
	tok := &SyntaxToken{Kind: ast.KindIdentifier, Pos: 10, End: 15}
	p1 := buildBinaryHeader("session-abc", 7, true, 20) // type
	p1 = appendBinaryNodeRef(p1, 3, "", 42, tok)
	d1, err := DecodeBinaryQuery(p1)
	if err != nil {
		t.Fatalf("decode node request: %v", err)
	}
	if d1.Session != "session-abc" || d1.Generation != 7 || !d1.Query.SkipRecords || d1.FileIdx != 3 {
		t.Fatalf("unexpected header: %+v", d1)
	}
	nr, ok := d1.Query.Request.(*NodeRequest)
	if !ok || nr.Op != "type" || nr.Node.Index != 42 || nr.Node.Token == nil || *nr.Node.Token != *tok {
		t.Fatalf("unexpected NodeRequest: %#v", d1.Query.Request)
	}

	// 2. PropertyRequest with UTF-16 lone surrogate name
	p2 := buildBinaryHeader("s", 1, false, 100) // property
	p2 = binary.LittleEndian.AppendUint32(p2, 99)
	units := []uint16{'a', 0xD800, 'b'} // lone high surrogate
	p2 = binary.LittleEndian.AppendUint32(p2, binaryNameUTF16Flag|uint32(len(units)))
	for _, u := range units {
		p2 = binary.LittleEndian.AppendUint16(p2, u)
	}
	d2, err := DecodeBinaryQuery(p2)
	if err != nil {
		t.Fatalf("decode utf16 property: %v", err)
	}
	pr, ok := d2.Query.Request.(*PropertyRequest)
	if !ok || pr.Op != "property" || pr.ID != 99 {
		t.Fatalf("unexpected PropertyRequest: %#v", d2.Query.Request)
	}
	// Verify lossless match against JSON Name unmarshaling.
	var jsonName Name
	rawUnits, err := json.Marshal(units)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rawUnits, &jsonName); err != nil {
		t.Fatal(err)
	}
	if pr.Name != jsonName {
		t.Fatalf("lone surrogate mismatch: got %q want %q", pr.Name, jsonName)
	}

	// 3. PrintRequest distinguishing flags=nil vs flags=0
	p3 := buildBinaryHeader("s", 2, false, 120) // typeString
	p3 = binary.LittleEndian.AppendUint32(p3, 55)
	p3 = append(p3, 0) // presence=0 (no node, no flags)
	p3 = binary.LittleEndian.AppendUint32(p3, 0)
	d3, err := DecodeBinaryQuery(p3)
	if err != nil {
		t.Fatalf("decode print nil flags: %v", err)
	}
	pr3 := d3.Query.Request.(*PrintRequest)
	if pr3.Flags != nil || pr3.Node != nil {
		t.Fatalf("expected nil flags and nil node, got %+v", pr3)
	}

	p4 := buildBinaryHeader("s", 2, false, 120)
	p4 = binary.LittleEndian.AppendUint32(p4, 55)
	p4 = append(p4, 2) // presence=2 (flags present = 0)
	p4 = binary.LittleEndian.AppendUint32(p4, 0)
	p4 = binary.LittleEndian.AppendUint32(p4, 0) // explicit flags=0
	d4, err := DecodeBinaryQuery(p4)
	if err != nil {
		t.Fatalf("decode print zero flags: %v", err)
	}
	pr4 := d4.Query.Request.(*PrintRequest)
	if pr4.Flags == nil || *pr4.Flags != 0 {
		t.Fatalf("expected explicit flags=0, got %+v", pr4.Flags)
	}

	// 4. Signed negative int32 (symbolsInScope meaning = -1)
	p5 := buildBinaryHeader("s", 3, false, 113)
	p5 = appendBinaryNodeRef(p5, BinaryInlineFileSentinel, "/a.ts", 7, nil)
	negOne := int32(-1)
	p5 = binary.LittleEndian.AppendUint32(p5, uint32(negOne))
	d5, err := DecodeBinaryQuery(p5)
	if err != nil {
		t.Fatalf("decode symbolsInScope: %v", err)
	}
	sr := d5.Query.Request.(*ScopeSymbolsRequest)
	if sr.Meaning != -1 || sr.Node.File != "/a.ts" || d5.FileIdx != BinaryInlineFileSentinel {
		t.Fatalf("unexpected ScopeSymbolsRequest: %+v fileIdx=%d", sr, d5.FileIdx)
	}

	// 5. OP_RAW_JSON accepts resolveName and rejects ProgramRequest / suggestionDiagnostics
	rawGood := buildBinaryHeader("s", 4, false, binaryOpRawJSON)
	rawGood = append(rawGood, `{"generation":4,"op":"resolveName","args":{"name":[120],"kind":-1,"excludeGlobals":false}}`...)
	d6, err := DecodeBinaryQuery(rawGood)
	if err != nil {
		t.Fatalf("decode raw json resolveName: %v", err)
	}
	rn := d6.Query.Request.(*ResolveNameRequest)
	if rn.Name != "x" || rn.Kind != -1 {
		t.Fatalf("unexpected ResolveNameRequest: %+v", rn)
	}

	rawProg := buildBinaryHeader("s", 4, false, binaryOpRawJSON)
	rawProg = append(rawProg, `{"generation":4,"op":"rootFileNames","args":{}}`...)
	if _, err := DecodeBinaryQuery(rawProg); err == nil {
		t.Fatal("expected ProgramRequest to be rejected on binary RPC")
	}

	rawSuggestion := buildBinaryHeader("s", 4, false, binaryOpRawJSON)
	rawSuggestion = append(rawSuggestion, `{"generation":4,"op":"suggestionDiagnostics","args":{"node":{"file":"/a.ts","index":1}}}`...)
	if _, err := DecodeBinaryQuery(rawSuggestion); err == nil {
		t.Fatal("expected suggestionDiagnostics to be rejected on binary RPC")
	}

	rawSkipMismatch := buildBinaryHeader("s", 4, true, binaryOpRawJSON)
	rawSkipMismatch = append(rawSkipMismatch, `{"generation":4,"op":"resolveName","args":{"name":[120],"kind":-1,"excludeGlobals":false}}`...)
	if _, err := DecodeBinaryQuery(rawSkipMismatch); err == nil {
		t.Fatal("expected skipRecords mismatch on OP_RAW_JSON to be rejected")
	}

	// Out-of-range int16 SyntaxToken.Kind must be rejected.
	badTok := buildBinaryHeader("s", 1, false, 20)
	badTok = binary.LittleEndian.AppendUint32(badTok, 0)
	badTok = binary.LittleEndian.AppendUint32(badTok, 1)
	badTok = append(badTok, 1)                               // hasToken=1
	badTok = binary.LittleEndian.AppendUint32(badTok, 70000) // > math.MaxInt16
	badTok = binary.LittleEndian.AppendUint32(badTok, 0)
	badTok = binary.LittleEndian.AppendUint32(badTok, 1)
	if _, err := DecodeBinaryQuery(badTok); err == nil {
		t.Fatal("expected out-of-range SyntaxToken.Kind to be rejected")
	}

	// 6. Trailing bytes and truncated payloads must fail
	if _, err := DecodeBinaryQuery(append(p1, 0x00)); err == nil {
		t.Fatal("expected trailing byte error")
	}
	for cut := range p1 {
		if _, err := DecodeBinaryQuery(p1[:cut]); err == nil {
			t.Fatalf("expected truncated payload at %d to fail", cut)
		}
	}
}

func findIdentifierIndex(t *testing.T, sf *ast.SourceFile, name string) uint32 {
	t.Helper()
	table := encoder.GetNodeIndexTable(sf)
	for i, n := range table.Nodes {
		if n != nil && ast.IsIdentifier(n) && n.Text() == name {
			return uint32(i)
		}
	}
	t.Fatalf("identifier %q not found in %s", name, sf.FileName())
	return 0
}

func TestSessionQuerySlotsIsolationAndSentMaskDedupe(t *testing.T) {
	p, _, root := buildTestProgram(t, map[string]string{
		"a.ts": "export const x: number = 1; export const y: number = 2;\n",
	}, []string{"a.ts"})
	sf := p.GetSourceFile(tspath.ResolvePath(root, "a.ts"))
	session, err := New(context.Background(), p, []*ast.SourceFile{sf})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	desc, err := session.Descriptor()
	if err != nil {
		t.Fatal(err)
	}
	fileIdx := BinaryInlineFileSentinel
	for i, src := range desc.Sources {
		if src.FileName == sf.FileName() {
			fileIdx = uint32(i)
			break
		}
	}
	if fileIdx == BinaryInlineFileSentinel {
		t.Fatal("expected /src/a.ts in Descriptor.Sources")
	}

	// Find identifier `x` node index via session's own query helper.
	xIdx := findIdentifierIndex(t, sf, "x")
	yIdx := findIdentifierIndex(t, sf, "y")

	// Mix a valid query (slot 0), an invalid type handle (slot 1), and a valid query (slot 2).
	// Slot 1's error MUST NOT fail slots 0 or 2.
	results := session.QuerySlots(context.Background(), []SlotQuery{
		{
			Slot:        0,
			WorkerEpoch: 1,
			FileIdx:     fileIdx,
			Query:       Query{Scope: Scope{Generation: desc.Generation}, Request: &NodeRequest{Op: "type", Node: NodeRef{Index: xIdx}}},
		},
		{
			Slot:        1,
			WorkerEpoch: 1,
			FileIdx:     BinaryInlineFileSentinel,
			Query:       Query{Scope: Scope{Generation: desc.Generation}, Request: &TypeRequest{Op: "properties", ID: 999999}},
		},
		{
			Slot:        2,
			WorkerEpoch: 1,
			FileIdx:     fileIdx,
			Query:       Query{Scope: Scope{Generation: desc.Generation}, Request: &NodeRequest{Op: "type", Node: NodeRef{Index: yIdx}}},
		},
	})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Err != nil || len(results[0].Packet) == 0 {
		t.Fatalf("slot 0 failed unexpectedly: %v", results[0].Err)
	}
	if results[1].Err == nil || !strings.Contains(results[1].Err.Error(), "invalid type handle") {
		t.Fatalf("expected slot 1 invalid type handle error, got %v", results[1].Err)
	}
	if results[2].Err != nil || len(results[2].Packet) == 0 {
		t.Fatalf("slot 2 failed unexpectedly: %v", results[2].Err)
	}

	// First call on slot 0 should bundle supplemental typeRecord (RTB1).
	if string(results[0].Packet[:4]) != "RTB1" {
		t.Fatalf("expected first call on slot 0 to carry RTB1 bundle, got %q", results[0].Packet[:4])
	}
	// Second call for same `number` type on slot 0 at same epoch should suppress
	// already-delivered supplemental record and return bare RTS5 handle packet (12B).
	second := session.QuerySlots(context.Background(), []SlotQuery{
		{
			Slot:        0,
			WorkerEpoch: 1,
			FileIdx:     fileIdx,
			Query:       Query{Scope: Scope{Generation: desc.Generation}, Request: &NodeRequest{Op: "type", Node: NodeRef{Index: yIdx}}},
		},
	})
	if second[0].Err != nil {
		t.Fatalf("second slot 0 query failed: %v", second[0].Err)
	}
	if string(second[0].Packet[:4]) != "RTS5" {
		t.Fatalf("expected second call on slot 0 to suppress duplicate supplemental record (RTS5), got %q", second[0].Packet[:4])
	}

	// Worker respawn (WorkerEpoch = 2) on slot 0 MUST reset sentMask bit 0 and resend RTB1.
	respawned := session.QuerySlots(context.Background(), []SlotQuery{
		{
			Slot:        0,
			WorkerEpoch: 2,
			FileIdx:     fileIdx,
			Query:       Query{Scope: Scope{Generation: desc.Generation}, Request: &NodeRequest{Op: "type", Node: NodeRef{Index: xIdx}}},
		},
	})
	if respawned[0].Err != nil {
		t.Fatalf("respawned slot 0 query failed: %v", respawned[0].Err)
	}
	if string(respawned[0].Packet[:4]) != "RTB1" {
		t.Fatalf("expected respawned worker epoch to receive fresh RTB1 bundle, got %q", respawned[0].Packet[:4])
	}
}
