package rule

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func parseExportedNamesForTest(t *testing.T, source string) map[string]bool {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.js",
		Path:     "/test.js",
	}, source, core.ScriptKindJS)
	return ParseExportedNames(sourceFile, NewCommentStore(sourceFile))
}

func TestParseExportedNames_ListSyntaxAndSourceFiltering(t *testing.T) {
	source := "const text = '/* exported stringName */';\n" +
		"const template = `/*exported templateName*/`;\n" +
		"const pattern = /\\/\\*exported regexName\\*\\//;\n" +
		"// exported lineName\n" +
		"/* Exported upperName */\n" +
		"/* exportedConfig prefixName */\n" +
		"/* exported foo, 'bar', \"baz\", two words, mismatched', '' -- reason, ignored */\n" +
		"/*exported foo, finalName*/\n"

	got := parseExportedNamesForTest(t, source)
	want := map[string]bool{
		"foo":         true,
		"bar":         true,
		"baz":         true,
		"two words":   true,
		"mismatched'": true,
		"finalName":   true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseExportedNames() = %#v, want %#v", got, want)
	}
}

func TestParseExportedNames_WhitespaceAndLabelBoundaries(t *testing.T) {
	zeroWidthSpace := string(rune(0x200B))
	source := "/*exported\u00A0unicodeName*/\n" +
		"/*exported\uFEFFbomName*/\n" +
		"/*exported\u0085notNextLine*/\n" +
		"/*exported" + zeroWidthSpace + "notZeroWidth*/\n" +
		"/*exported,notComma*/\n" +
		"/*exported*/\n"

	got := parseExportedNamesForTest(t, source)
	want := map[string]bool{"unicodeName": true, "bomName": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseExportedNames() = %#v, want %#v", got, want)
	}
}

func TestParseExportedNames_NoDirective(t *testing.T) {
	for _, source := range []string{
		"",
		"const text = '/*exported stringName*/';",
		"// exported lineName\nconst value = 1;",
	} {
		if got := parseExportedNamesForTest(t, source); got != nil {
			t.Errorf("ParseExportedNames(%q) = %#v, want nil", source, got)
		}
	}
}
