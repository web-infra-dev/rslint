package utils

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

type traversalCommentExpectation struct {
	text    string
	kind    ast.Kind
	newline bool
}

func parseCommentTraversalSource(source string, kind core.ScriptKind) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/comments.tsx",
		Path:     "/comments.tsx",
	}, source, kind)
}

func assertTraversalComments(t *testing.T, sourceFile *ast.SourceFile, node *ast.Node, expected []traversalCommentExpectation) {
	t.Helper()
	var want []ast.CommentRange
	for _, comment := range expected {
		pos := strings.Index(sourceFile.Text(), comment.text)
		if pos < 0 || comment.text == "" {
			t.Fatalf("expected comment %q is missing from the fixture", comment.text)
		}
		want = append(want, ast.CommentRange{
			TextRange:          core.NewTextRange(pos, pos+len(comment.text)),
			Kind:               comment.kind,
			HasTrailingNewLine: comment.newline,
		})
	}
	var got []ast.CommentRange
	ForEachComment(node, func(comment *ast.CommentRange) {
		got = append(got, *comment)
	}, sourceFile)
	// Compare raw callbacks, without sorting or deduplicating them.
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("raw comments = %#v, want %#v", got, want)
	}
}

func TestCommentTraversalTriviaAndSyntax(t *testing.T) {
	const (
		block = ast.KindMultiLineCommentTrivia
		line  = ast.KindSingleLineCommentTrivia
	)
	tests := []struct {
		name      string
		source    string
		want      []traversalCommentExpectation
		malformed bool
	}{
		{name: "empty"},
		{name: "whitespace only", source: " \t\v\f\r\n\u00a0\ufeff"},
		{name: "shebang only", source: "#!/usr/bin/env node"},
		{
			name:   "comments after shebang",
			source: "#!/usr/bin/env node\r\n/* first */\r\n// last",
			want:   []traversalCommentExpectation{{"/* first */", block, true}, {"// last", line, false}},
		},
		{
			name:   "comment only source",
			source: "/* first */\n// second",
			want:   []traversalCommentExpectation{{"/* first */", block, true}, {"// second", line, false}},
		},
		{
			name:      "unterminated comment",
			source:    "/* incomplete",
			want:      []traversalCommentExpectation{{"/* incomplete", block, false}},
			malformed: true,
		},
		{
			name:   "BOM and trailing CRLF",
			source: "\ufeff/* first */\nconst value = 1; // last\r\n",
			want:   []traversalCommentExpectation{{"/* first */", block, true}, {"// last", line, true}},
		},
		{
			name:   "Unicode line breaks",
			source: "// first\u2028/* second */\u2029const value = 1;",
			want:   []traversalCommentExpectation{{"// first", line, true}, {"/* second */", block, true}},
		},
		{
			name:   "trailing and leading ownership",
			source: "let first = 1; /* same line */\r\n/* next line */\r\nlet second = 2;",
			want:   []traversalCommentExpectation{{"/* same line */", block, false}, {"/* next line */", block, true}},
		},
		{
			name:   "string and regex markers",
			source: `const a = '/* string */'; const b = /[/*]/; const c = /https?:\/\//; /* real */`,
			want:   []traversalCommentExpectation{{"/* real */", block, false}},
		},
		{
			name:   "template expression and raw text",
			source: "const value = `/* raw */ ${1 /* expression */ + 2} // raw`; // end\n",
			want:   []traversalCommentExpectation{{"/* expression */", block, false}, {"// end", line, true}},
		},
		{
			name:      "missing initializer",
			source:    "const value = ; // recovered\n",
			want:      []traversalCommentExpectation{{"// recovered", line, true}},
			malformed: true,
		},
		{
			name:      "unterminated string",
			source:    "const value = 'unfinished\n/* recovered */ const next = 1;",
			want:      []traversalCommentExpectation{{"/* recovered */", block, false}},
			malformed: true,
		},
		{
			name:      "unterminated template expression",
			source:    "const value = `/* raw */ ${1 /* real */",
			want:      []traversalCommentExpectation{{"/* real */", block, false}},
			malformed: true,
		},
		{
			name:      "invalid UTF-8 before comment",
			source:    "const first = 1;\xff/* recovered */const next = 2;",
			malformed: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseCommentTraversalSource(test.source, core.ScriptKindTS)
			if test.malformed && len(sourceFile.Diagnostics()) == 0 {
				t.Fatal("fixture must exercise parser recovery")
			}
			assertTraversalComments(t, sourceFile, sourceFile.AsNode(), test.want)
		})
	}
}

func TestCommentTraversalJSXTriviaOwners(t *testing.T) {
	const block = ast.KindMultiLineCommentTrivia
	tests := []struct {
		name   string
		source string
		want   []traversalCommentExpectation
	}{
		{
			name:   "closing element keeps existing trailing behavior",
			source: "const value = <><A></A>/* text */</>; /* real */",
			// The existing closing-element case permits this trailing scan,
			// even though a closing fragment at the same depth does not.
			want: []traversalCommentExpectation{{"/* text */", block, false}, {"/* real */", block, false}},
		},
		{
			name:   "nested closing fragment excludes text",
			source: "const value = <><></>/* text */</>; /* real */",
			want:   []traversalCommentExpectation{{"/* real */", block, false}},
		},
		{
			name:   "self closing element and expression exclude text",
			source: "const value = <><A />/* first text */{1 /* expression */}/* second text */</>; /* real */",
			want:   []traversalCommentExpectation{{"/* expression */", block, false}, {"/* real */", block, false}},
		},
		{
			name:   "type arguments and attributes own trivia",
			source: "const value = <A<T /* type */> prop={1 /* attribute */}>/* text */</A>; /* real */",
			want:   []traversalCommentExpectation{{"/* type */", block, false}, {"/* attribute */", block, false}, {"/* real */", block, false}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseCommentTraversalSource(test.source, core.ScriptKindTSX)
			if len(sourceFile.Diagnostics()) != 0 {
				t.Fatal("JSX fixture must parse without errors")
			}
			assertTraversalComments(t, sourceFile, sourceFile.AsNode(), test.want)
		})
	}
}

func TestCommentTraversalLongTrivia(t *testing.T) {
	for _, length := range []int{7, 8, 9, 64} {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			padding := strings.Repeat(" ", length)
			source := padding + "/* leading */\r\nconst value = 1;" + padding + "/* same line */\r\n" + padding + "// next line\n"
			sourceFile := parseCommentTraversalSource(source, core.ScriptKindTS)
			assertTraversalComments(t, sourceFile, sourceFile.AsNode(), []traversalCommentExpectation{
				{"/* leading */", ast.KindMultiLineCommentTrivia, true},
				{"/* same line */", ast.KindMultiLineCommentTrivia, false},
				{"// next line", ast.KindSingleLineCommentTrivia, true},
			})
		})
	}
}

func TestCommentTraversalSkipsReparsedJSDoc(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		code string
	}{
		{"type", "/** @type {number[]} */", "const values = [1, 2];"},
		{"parameter", "/** @param {number[]} values */", "function use(values) { return values; }"},
		{"template", "/** @template T\n * @param {T} value\n */", "function identity(value) { return value; }"},
		{"overload", "/** @overload\n * @param {string} value\n * @returns {string}\n */", "function identity(value) { return value; }"},
		{"import", "/** @import { Widget } from 'pkg' */", "const value = 1;"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseCommentTraversalSource(test.doc+"\n"+test.code+" // after\n", core.ScriptKindJS)
			reparsed := 0
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Flags&ast.NodeFlagsReparsed != 0 {
					reparsed++
				}
				node.ForEachChild(visit)
				return false
			}
			visit(sourceFile.AsNode())
			if reparsed == 0 {
				t.Fatal("JS fixture must contain reparsed JSDoc nodes")
			}
			assertTraversalComments(t, sourceFile, sourceFile.AsNode(), []traversalCommentExpectation{
				{test.doc, ast.KindMultiLineCommentTrivia, true},
				{"// after", ast.KindSingleLineCommentTrivia, true},
			})
		})
	}
}

func TestCommentTraversalColdSubtrees(t *testing.T) {
	const source = "/* before first */\nfunction first() { /* inside first */ }\n" +
		"// before second\nfunction second() { /* inside second */ }\n/* after second */"
	for _, bodyOnly := range []bool{false, true} {
		name := "function"
		if bodyOnly {
			name = "body"
		}
		t.Run(name, func(t *testing.T) {
			// Parse separately for each subtree so no prior walk can populate
			// its token cache or accidentally supply file-wide comments.
			sourceFile := parseCommentTraversalSource(source, core.ScriptKindTS)
			if len(sourceFile.Statements.Nodes) != 2 {
				t.Fatal("fixture must contain two functions")
			}
			node := sourceFile.Statements.Nodes[1]
			want := []traversalCommentExpectation{{"// before second", ast.KindSingleLineCommentTrivia, true}}
			if bodyOnly {
				node = node.Body()
				want = nil
			}
			if node == nil {
				t.Fatal("subtree is missing")
			}
			want = append(want, traversalCommentExpectation{"/* inside second */", ast.KindMultiLineCommentTrivia, false})
			assertTraversalComments(t, sourceFile, node, want)
		})
	}
}

func TestCommentTraversalMissingTokenHasNoFileFallback(t *testing.T) {
	sourceFile := parseCommentTraversalSource("const value = ; /* outside */", core.ScriptKindTS)
	var missing *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node.Pos() == node.End() {
			missing = node
			return true
		}
		return node.ForEachChild(visit)
	}
	visit(sourceFile.AsNode())
	if missing == nil {
		t.Fatal("fixture must contain a zero-width missing identifier")
	}
	assertTraversalComments(t, sourceFile, missing, nil)
}
