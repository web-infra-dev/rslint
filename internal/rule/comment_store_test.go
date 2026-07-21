package rule

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func parseCommentStoreSource(t *testing.T, source string) *ast.SourceFile {
	t.Helper()
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)
}

func TestCommentStoreCollectsOnceInSourceOrder(t *testing.T) {
	source := "const fake = '/* not a comment */';\n" +
		"const other = '// still not a comment'; // first\n" +
		"/* second */ const value = 1;\n"
	store := NewCommentStore(parseCommentStoreSource(t, source))
	if store.scanned {
		t.Fatal("new comment store was already scanned")
	}

	comments := store.All()
	if !store.scanned {
		t.Fatal("comment store did not record its scan")
	}
	var texts []string
	for _, comment := range comments {
		texts = append(texts, source[comment.Pos():comment.End()])
	}
	if want := []string{"// first", "/* second */"}; !reflect.DeepEqual(texts, want) {
		t.Fatalf("comments = %#v, want %#v", texts, want)
	}

	again := store.All()
	if len(again) != len(comments) || (len(again) > 0 && again[0] != comments[0]) {
		t.Fatal("second All call did not reuse the cached comment list")
	}
}

func TestCommentStoreWithoutCommentsFastPath(t *testing.T) {
	store := NewCommentStore(parseCommentStoreSource(t, "const quotient = 8 / 2;\nconst value = 'plain';\n"))
	if comments := store.All(); len(comments) != 0 {
		t.Fatalf("comments = %#v, want none", comments)
	}
	if !store.scanned {
		t.Fatal("fast path for a file without comments did not cache its result")
	}
}

func TestDisableManagerDefersOrdinaryComments(t *testing.T) {
	source := "// an ordinary comment\nalert('hello')\n"
	sourceFile := parseCommentStoreSource(t, source)
	store := NewCommentStore(sourceFile)
	manager := NewDisableManager(sourceFile, store)
	if store.scanned {
		t.Fatal("constructing DisableManager scanned comments")
	}

	if manager.IsRuleDisabled("no-alert", strings.Index(source, "alert")) {
		t.Fatal("ordinary comment disabled a rule")
	}
	if store.scanned {
		t.Fatal("ordinary comment caused the comment store to scan")
	}
}

func TestDisableManagerLazilyParsesDirective(t *testing.T) {
	source := "// eslint-disable-next-line no-alert\nalert('hello')\n"
	sourceFile := parseCommentStoreSource(t, source)
	store := NewCommentStore(sourceFile)
	manager := NewDisableManager(sourceFile, store)
	if store.scanned {
		t.Fatal("constructing DisableManager scanned comments")
	}

	if !manager.IsRuleDisabled("no-alert", strings.LastIndex(source, "alert")) {
		t.Fatal("disable-next-line directive was not applied")
	}
	if !store.scanned {
		t.Fatal("disable check did not lazily scan directive comments")
	}
}

func TestMayContainDisableDirective(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "rslint disable", text: "// rslint-disable no-alert", want: true},
		{name: "rslint enable", text: "/* rslint-enable */", want: true},
		{name: "eslint disable", text: "// eslint-disable-next-line no-alert", want: true},
		{name: "eslint enable", text: "/* eslint-enable no-alert */", want: true},
		{name: "unsupported prefix", text: "// lint-disable no-alert"},
		{name: "wrong case", text: "// ESLint-disable no-alert"},
		{name: "ordinary comment", text: "// lint rules are useful"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mayContainDisableDirective(test.text); got != test.want {
				t.Fatalf("mayContainDisableDirective(%q) = %v, want %v", test.text, got, test.want)
			}
		})
	}
}
