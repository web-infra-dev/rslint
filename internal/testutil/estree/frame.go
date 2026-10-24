// Package estree reads source snapshots for plugin integration test doubles.
package estree

import (
	"encoding/binary"
	"testing"
)

// SourceText extracts the source column from the current private ESTree frame.
func SourceText(t testing.TB, frame []byte) string {
	t.Helper()
	if len(frame) < 40 {
		t.Fatal("missing ESTree frame")
	}
	count := func(offset int) int { return int(binary.LittleEndian.Uint32(frame[offset:])) }
	nodes, tokens := count(8), count(32)
	offset := (40 + nodes*2 + 3) &^ 3
	offset += nodes*12 + count(12)*4 + count(16)*4 + count(24) + count(28)*8
	offset = (offset + tokens + 3) &^ 3
	offset += tokens*8 + count(36)*12
	if offset > len(frame) {
		t.Fatal("incomplete ESTree frame")
	}
	return string(frame[offset:])
}
