package lsp

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
)

func partialDocumentChange(
	startLine uint32,
	startCharacter uint32,
	endLine uint32,
	endCharacter uint32,
	text string,
) lsproto.TextDocumentContentChangePartialOrWholeDocument {
	return lsproto.TextDocumentContentChangePartialOrWholeDocument{
		Partial: &lsproto.TextDocumentContentChangePartial{
			Range: lsproto.Range{
				Start: lsproto.Position{Line: startLine, Character: startCharacter},
				End:   lsproto.Position{Line: endLine, Character: endCharacter},
			},
			Text: text,
		},
	}
}

func wholeDocumentChange(text string) lsproto.TextDocumentContentChangePartialOrWholeDocument {
	return lsproto.TextDocumentContentChangePartialOrWholeDocument{
		WholeDocument: &lsproto.TextDocumentContentChangeWholeDocument{Text: text},
	}
}

func TestApplyDocumentChanges(t *testing.T) {
	t.Parallel()

	rangeLength := uint32(999)
	tests := []struct {
		name    string
		content string
		changes []lsproto.TextDocumentContentChangePartialOrWholeDocument
		want    string
		wantErr bool
	}{
		{
			name:    "empty change list",
			content: "unchanged",
			want:    "unchanged",
		},
		{
			name:    "whole document fallback",
			content: "old",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{wholeDocumentChange("new")},
			want:    "new",
		},
		{
			name:    "ASCII insertion",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 1, 0, 1, "X")},
			want:    "aXbc",
		},
		{
			name:    "multiline deletion with CRLF",
			content: "a\r\nb",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 1, 1, 0, "")},
			want:    "ab",
		},
		{
			name:    "changes use preceding result",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				partialDocumentChange(0, 1, 0, 2, "XY"),
				partialDocumentChange(0, 3, 0, 4, "!"),
			},
			want: "aXY!",
		},
		{
			name:    "mixed whole and partial changes",
			content: "ignored",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				wholeDocumentChange("abc"),
				partialDocumentChange(0, 3, 0, 3, "d"),
			},
			want: "abcd",
		},
		{
			name:    "UTF-16 astral character",
			content: "a😀b",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 3, 0, 4, "c")},
			want:    "a😀c",
		},
		{
			name:    "UTF-16 astral insertion with sequential CRLF edit",
			content: "a😀b\r\nc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				partialDocumentChange(0, 3, 0, 3, "X\r\nY"),
				partialDocumentChange(1, 1, 1, 2, "B"),
			},
			want: "a😀X\r\nYB\r\nc",
		},
		{
			name:    "UTF-16 CJK character",
			content: "甲乙c",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 2, 0, 3, "d")},
			want:    "甲乙d",
		},
		{
			name:    "UTF-16 offset inside surrogate pair clamps before rune",
			content: "a😀b",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 2, 0, 2, "X")},
			want:    "aX😀b",
		},
		{
			name:    "LSP lines exclude Unicode line separator",
			content: "a\u2028b",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(0, 2, 0, 3, "c")},
			want:    "a\u2028c",
		},
		{
			name:    "out of bounds position clamps to document end",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{partialDocumentChange(99, 99, 99, 99, "d")},
			want:    "abcd",
		},
		{
			name:    "deprecated range length is ignored",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{{
				Partial: &lsproto.TextDocumentContentChangePartial{
					Range:       lsproto.Range{Start: lsproto.Position{Character: 1}, End: lsproto.Position{Character: 2}},
					RangeLength: &rangeLength,
					Text:        "X",
				},
			}},
			want: "aXc",
		},
		{
			name:    "reversed range is transactional",
			content: "a\nb",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{
				partialDocumentChange(0, 0, 0, 1, "A"),
				partialDocumentChange(1, 0, 0, 0, "invalid"),
			},
			want:    "a\nb",
			wantErr: true,
		},
		{
			name:    "missing change kind",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{{}},
			want:    "abc",
			wantErr: true,
		},
		{
			name:    "ambiguous change kind",
			content: "abc",
			changes: []lsproto.TextDocumentContentChangePartialOrWholeDocument{{
				Partial:       partialDocumentChange(0, 0, 0, 1, "A").Partial,
				WholeDocument: wholeDocumentChange("whole").WholeDocument,
			}},
			want:    "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := applyDocumentChanges(tt.content, tt.changes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("applyDocumentChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("applyDocumentChanges() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeLSPLineStarts(t *testing.T) {
	t.Parallel()

	content := "a\r\nb\rc\nd\u2028e"
	want := []int{0, 3, 5, 7}
	got := computeLSPLineStarts(content)
	if len(got) != len(want) {
		t.Fatalf("computeLSPLineStarts() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("computeLSPLineStarts() = %v, want %v", got, want)
		}
	}
}

func FuzzApplyDocumentChanges(f *testing.F) {
	f.Add("a😀b\r\nc", "X", uint32(0), uint32(1), uint32(0), uint32(3))
	f.Add("甲乙\nc", "", uint32(0), uint32(1), uint32(1), uint32(0))
	f.Add("", "content", uint32(99), uint32(99), uint32(99), uint32(99))

	f.Fuzz(func(
		t *testing.T,
		content string,
		replacement string,
		startLine uint32,
		startCharacter uint32,
		endLine uint32,
		endCharacter uint32,
	) {
		if !utf8.ValidString(content) || !utf8.ValidString(replacement) {
			t.Skip()
		}

		updated, err := applyDocumentChanges(content, []lsproto.TextDocumentContentChangePartialOrWholeDocument{
			partialDocumentChange(startLine, startCharacter, endLine, endCharacter, replacement),
		})
		if err != nil {
			if updated != content {
				t.Fatalf("failed change returned partial content %q, want original %q", updated, content)
			}
			return
		}
		if !utf8.ValidString(updated) {
			t.Fatalf("valid inputs produced invalid UTF-8: %q", updated)
		}
	})
}

var benchmarkDocumentContent string

func BenchmarkApplyDocumentChanges(b *testing.B) {
	for _, tt := range []struct {
		name string
		size int
	}{
		{name: "4KiB", size: 4 << 10},
		{name: "256KiB", size: 256 << 10},
		{name: "1MiB", size: 1 << 20},
	} {
		content := strings.Repeat("a", tt.size)
		change := []lsproto.TextDocumentContentChangePartialOrWholeDocument{
			partialDocumentChange(0, uint32(tt.size/2), 0, uint32(tt.size/2), "x"),
		}
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tt.size))
			for range b.N {
				updated, err := applyDocumentChanges(content, change)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkDocumentContent = updated
			}
		})
	}
}
