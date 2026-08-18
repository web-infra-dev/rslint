package utils

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	rslintutils "github.com/web-infra-dev/rslint/internal/utils"
)

func TestLineRangeWithComments(t *testing.T) {
	t.Parallel()

	const source = "  /* before */ import b from 'b'; /* block */ // line\r\n\timport a from 'a';\r\n"
	sourceFile, comments := parseCommentRangeFixture(t, source)
	lineStarts := sourceFile.ECMALineMap()
	first := sourceFile.Statements.Nodes[0]
	second := sourceFile.Statements.Nodes[1]

	start, end := LineRangeWithComments(source, first, lineStarts, comments)
	if start != 1 {
		t.Fatalf("first movable start = %d, want 1", start)
	}
	wantEnd := strings.Index(source, "\timport")
	if end != wantEnd {
		t.Fatalf("first movable end = %d, want %d", end, wantEnd)
	}
	if got := source[start:end]; got != " /* before */ import b from 'b'; /* block */ // line\r\n" {
		t.Fatalf("first movable text = %q", got)
	}

	secondStart, secondEnd := LineRangeWithComments(source, second, lineStarts, comments)
	if secondStart != wantEnd {
		t.Fatalf("second movable start = %d, want %d", secondStart, wantEnd)
	}
	if secondEnd != len(source) {
		t.Fatalf("second movable end = %d, want %d", secondEnd, len(source))
	}

	sameLineEnd := SameLineEndWithComments(source, first, lineStarts, comments)
	if want := strings.Index(source, "\r\n"); sameLineEnd != want {
		t.Fatalf("same-line end = %d, want %d", sameLineEnd, want)
	}
}

func TestCommentRangeCapsAdjacentComments(t *testing.T) {
	t.Parallel()

	const statement = "import b from 'b';"
	const comment = " /* kept */"
	source := statement + strings.Repeat(comment, 101) + "\n"
	sourceFile, comments := parseCommentRangeFixture(t, source)
	end := EndOfLineWithComments(source, sourceFile.Statements.Nodes[0], sourceFile.ECMALineMap(), comments)
	want := len(statement) + len(strings.Repeat(comment, 100)) + 1
	if end != want {
		t.Fatalf("movable end = %d, want %d", end, want)
	}
}

func TestStartOfLineWithCommentsUsesStatementEndLine(t *testing.T) {
	t.Parallel()

	const source = "/* detached */ import {\n  b,\n} from 'b';\n"
	sourceFile, comments := parseCommentRangeFixture(t, source)
	start := StartOfLineWithComments(source, sourceFile.Statements.Nodes[0], sourceFile.ECMALineMap(), comments)
	if want := len("/* detached */"); start != want {
		t.Fatalf("multiline import start = %d, want %d", start, want)
	}
}

func TestCommentRangeNilNode(t *testing.T) {
	t.Parallel()

	if start := StartOfLineWithComments("", nil, nil, nil); start != -1 {
		t.Fatalf("nil start = %d, want -1", start)
	}
	if end := EndOfLineWithComments("", nil, nil, nil); end != -1 {
		t.Fatalf("nil end = %d, want -1", end)
	}
	if end := SameLineEndWithComments("", nil, nil, nil); end != -1 {
		t.Fatalf("nil same-line end = %d, want -1", end)
	}
}

func parseCommentRangeFixture(t *testing.T, source string) (*ast.SourceFile, []*ast.CommentRange) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/comments.ts",
		Path:     "/comments.ts",
	}, source, core.ScriptKindTS)
	var comments []*ast.CommentRange
	rslintutils.ForEachComment(sourceFile.AsNode(), func(comment *ast.CommentRange) {
		comments = append(comments, comment)
	}, sourceFile)
	sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })
	return sourceFile, comments
}
