package typeservice

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func readU32(data []byte, offset uint32) uint32 {
	return binary.LittleEndian.Uint32(data[offset : offset+4])
}

func readUTF16(t *testing.T, data []byte, offset uint32) string {
	t.Helper()
	n := readU32(data, offset)
	units := make([]uint16, n)
	for i := range n {
		units[i] = binary.LittleEndian.Uint16(data[offset+4+i*2:])
	}
	runes := make([]rune, len(units))
	for i, u := range units {
		runes[i] = rune(u)
	}
	return string(runes)
}

func TestEncodePacketSymbolFactSymmetricNodes(t *testing.T) {
	declA := NodeRef{File: "/proj/a.ts", Index: 7}
	declB := NodeRef{File: "/proj/b.ts", Index: 19}
	fact := &SymbolFact{
		Name:             "Widget",
		EscapedName:      "Widget",
		Flags:            uint32(ast.SymbolFlagsInterface | ast.SymbolFlagsTransient),
		CheckFlags:       42,
		Declarations:     []NodeRef{declA, declB},
		ValueDeclaration: &declB,
	}
	pkt, err := EncodePacket(fact)
	if err != nil {
		t.Fatal(err)
	}
	if string(pkt[:4]) != "RTS5" || readU32(pkt, 4) != packetSymbol {
		t.Fatalf("unexpected header: %q kind=%d", pkt[:4], readU32(pkt, 4))
	}
	if got := readUTF16(t, pkt, readU32(pkt, 8)); got != "Widget" {
		t.Fatalf("name=%q, want Widget", got)
	}
	if got := readU32(pkt, 28); got != 43 {
		t.Fatalf("checkFlags+1=%d, want 43", got)
	}
	declAt := readU32(pkt, 20)
	if count := readU32(pkt, declAt); count != 2 {
		t.Fatalf("decl count=%d, want 2", count)
	}
	if f0, i0 := readUTF16(t, pkt, readU32(pkt, declAt+4)), readU32(pkt, declAt+8); f0 != declA.File || i0 != declA.Index {
		t.Fatalf("decl[0]=(%q,%d), want (%q,%d)", f0, i0, declA.File, declA.Index)
	}
	valAt := readU32(pkt, 24)
	if vf, vi := readUTF16(t, pkt, readU32(pkt, valAt)), readU32(pkt, valAt+4); vf != declB.File || vi != declB.Index {
		t.Fatalf("valueDecl=(%q,%d), want (%q,%d)", vf, vi, declB.File, declB.Index)
	}
}

func TestEncodePacketOptionalCollectionsAndLiterals(t *testing.T) {
	absentMembers, err := EncodePacket(OptionalNamedSymbols{Present: false})
	if err != nil {
		t.Fatal(err)
	}
	if readU32(absentMembers, 4) != packetOptionalNamedSymbols || readU32(absentMembers, 8) != 0 {
		t.Fatalf("expected absent OptionalNamedSymbols offset 0")
	}

	presentMembers, err := EncodePacket(OptionalNamedSymbols{
		Present: true,
		Symbols: NamedSymbols{{ID: 9, Name: "foo"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	at := readU32(presentMembers, 8)
	if at == 0 || readU32(presentMembers, at) != 1 || readU32(presentMembers, at+4) != 9 {
		t.Fatalf("unexpected present OptionalNamedSymbols encoding")
	}

	absentIdx, err := EncodePacket(OptionalIndexInfos{Present: false})
	if err != nil {
		t.Fatal(err)
	}
	if readU32(absentIdx, 4) != packetOptionalIndexInfos || readU32(absentIdx, 8) != 0 {
		t.Fatalf("expected absent OptionalIndexInfos offset 0")
	}

	numLit, err := EncodePacket(Literal{Kind: 2, Number: 3.5})
	if err != nil {
		t.Fatal(err)
	}
	if got := math.Float64frombits(binary.LittleEndian.Uint64(numLit[12:20])); got != 3.5 {
		t.Fatalf("number literal=%v, want 3.5", got)
	}
}

func TestEncodeReplyBundlesSupplementalRecords(t *testing.T) {
	reply, err := EncodeReply(Handle(5), []Record{
		{Op: "typeRecord", ID: 5, Fact: &TypeFact{Flags: []string{"String"}}},
		{Op: "symbolRecord", ID: 12, Fact: &SymbolFact{Name: "x", EscapedName: "x"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(reply[:4]) != "RTB1" {
		t.Fatalf("expected RTB1 bundle header, got %q", reply[:4])
	}
	if count := readU32(reply, 8); count != 2 {
		t.Fatalf("record count=%d, want 2", count)
	}
	// Reject invalid supplemental op.
	if _, err := EncodeReply(Handle(5), []Record{{Op: "unknownRecord", ID: 1, Fact: Handle(1)}}); err == nil {
		t.Fatal("expected error for invalid supplemental record op")
	}
}
