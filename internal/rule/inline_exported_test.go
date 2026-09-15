package rule

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func parseInlineExportedForTest(t *testing.T, source string) (map[string]bool, []InlineExported) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	return ParseInlineExported(sourceFile, NewCommentStore(sourceFile))
}

func TestParseInlineExported_MetadataAndSourceFiltering(t *testing.T) {
	source := "/* exported foo, 'quoted' -- why, ignored */\n" +
		"/* exported foo, \"double\" */\n" +
		"// exported lineName\n" +
		"/* Exported upperName */\n" +
		"/* exported-value prefixName */\n" +
		"const fake = '/* exported stringName */';\n"

	names, exported := parseInlineExportedForTest(t, source)

	wantNames := map[string]bool{"foo": true, "quoted": true, "double": true}
	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("names = %#v, want %#v", names, wantNames)
	}

	want := []struct {
		name      string
		positions []int
	}{
		{name: "foo", positions: []int{12, 57}},
		{name: "quoted", positions: []int{18}},
		{name: "double", positions: []int{63}},
	}
	if len(exported) != len(want) {
		t.Fatalf("got %d inline exported entries, want %d: %#v", len(exported), len(want), exported)
	}
	for i, expected := range want {
		actual := exported[i]
		if actual.Name != expected.name {
			t.Fatalf("inline exported %d = %q, want %q", i, actual.Name, expected.name)
		}
		if len(actual.NameRanges) != len(expected.positions) {
			t.Fatalf("inline exported %q has %d ranges, want %d", actual.Name, len(actual.NameRanges), len(expected.positions))
		}
		for rangeIndex, textRange := range actual.NameRanges {
			wantPos := expected.positions[rangeIndex]
			if textRange.Pos() != wantPos || textRange.End() != wantPos+len(expected.name) {
				t.Errorf("inline exported %q range %d = %d:%d, want %d:%d", actual.Name, rangeIndex, textRange.Pos(), textRange.End(), wantPos, wantPos+len(expected.name))
			}
			if rangeText := source[textRange.Pos():textRange.End()]; rangeText != expected.name {
				t.Errorf("inline exported %q range %d contains %q", actual.Name, rangeIndex, rangeText)
			}
		}
	}
}

// TestParseInlineExported_ListSyntax locks in @eslint/plugin-kit's
// parseListConfig: commas alone separate entries, so whitespace inside one
// entry is part of the name.
func TestParseInlineExported_ListSyntax(t *testing.T) {
	tests := []struct {
		source string
		want   []string
	}{
		{source: `/* exported a b, c */`, want: []string{"a b", "c"}},
		{source: `/* exported  ,, a ,, */`, want: []string{"a"}},
		{source: `/* exported 'a', "b" */`, want: []string{"a", "b"}},
		{source: `/* exported 'c", '', ' */`, want: []string{`'c"`, `'`}},
		{source: `/*exported a*/`, want: []string{"a"}},
		{source: `/* exported */`, want: nil},
		{source: `/* exported -- a */`, want: nil},
		{source: `/* exported--a */`, want: nil},
		{source: `/* exported-a b */`, want: nil},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			_, exported := parseInlineExportedForTest(t, test.source)
			got := make([]string, 0, len(exported))
			for _, entry := range exported {
				got = append(got, entry.Name)
			}
			if len(got) == 0 {
				got = nil
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("names = %#v, want %#v", got, test.want)
			}
		})
	}
}
