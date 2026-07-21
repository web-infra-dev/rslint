package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

// ForEachComment now reuses one *ast.NodeFactory across every token of a file
// instead of allocating two per token. This test locks the behavior that
// matters for every caller (DisableManager + the comment-reading rules):
// each token's leading/trailing comments must still be reported independently
// and completely — i.e. the shared factory must not smear or drop anything.
func TestForEachComment_ReuseFactoryReportsAllCommentsPerToken(t *testing.T) {
	// Comments deliberately attached to several distinct tokens, mixing line
	// and block, leading and trailing, plus a shebang. If factory reuse
	// leaked state across tokens, the collected set would differ from this.
	src := "#!/usr/bin/env node\n" +
		"// leadingA\n" +
		"const a = 1; // trailingA\n" +
		"/* leadingB */ const b = 2;\n" +
		"function f() {} /* trailingF */\n"

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "f.ts")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := CreateCompilerHost(tmpDir, fs)
	prog, err := CreateProgramFromOptions(true, &core.CompilerOptions{}, []string{file}, host)
	if err != nil {
		t.Fatal(err)
	}
	var sf *ast.SourceFile
	for _, f := range prog.GetSourceFiles() {
		if f.FileName() == file || filepath.Base(f.FileName()) == "f.ts" {
			sf = f
		}
	}
	if sf == nil {
		t.Fatal("source file not found")
	}

	type cmt struct {
		text string
		line bool
	}
	var got []cmt
	ForEachComment(sf.AsNode(), func(c *ast.CommentRange) {
		got = append(got, cmt{
			text: src[c.Pos():c.End()],
			line: c.Kind == ast.KindSingleLineCommentTrivia,
		})
	}, sf)

	// Every comment in the file must appear exactly once with correct text
	// and kind. Order is by token walk; we assert as a set/count to stay
	// robust to walk order while still catching any drop/dup/smear.
	want := map[string]bool{ // text -> isLine
		"// leadingA":     true,
		"// trailingA":    true,
		"/* leadingB */":  false,
		"/* trailingF */": false,
	}
	// Shebang is reported by the scanner as a SingleLineCommentTrivia-shaped
	// range at pos 0; accept it if present but don't require a specific kind.
	seen := map[string]int{}
	for _, c := range got {
		seen[c.text]++
		if isLine, ok := want[c.text]; ok && isLine != c.line {
			t.Errorf("comment %q kind mismatch: gotLine=%v wantLine=%v", c.text, c.line, isLine)
		}
	}
	for text := range want {
		if seen[text] != 1 {
			t.Errorf("comment %q expected exactly once, got %d (factory reuse smeared/dropped?)", text, seen[text])
		}
	}
}

func TestForEachComment_CommentOnlySourceFile(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "mixed comments",
			source: "/* first */\n// second",
			want:   []string{"/* first */", "// second"},
		},
		{
			name:   "comments after shebang",
			source: "#!/usr/bin/env node\n/* first */\n// second",
			want:   []string{"/* first */", "// second"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, test.source, core.ScriptKindTS)

			var got []string
			ForEachComment(sourceFile.AsNode(), func(comment *ast.CommentRange) {
				got = append(got, test.source[comment.Pos():comment.End()])
			}, sourceFile)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("comments = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestHasCommentInSpan(t *testing.T) {
	src := `const a = "/* not a comment */"; const b = 1 /* real */ + 2;`

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "f.ts")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := CreateCompilerHost(tmpDir, fs)
	prog, err := CreateProgramFromOptions(true, &core.CompilerOptions{}, []string{file}, host)
	if err != nil {
		t.Fatal(err)
	}
	sf := prog.GetSourceFile(file)
	if sf == nil {
		t.Fatal("source file not found")
	}

	// HasCommentInSpan takes the file's pre-collected, sorted comment list
	// (what ctx.Comments.All() returns in production) rather than the source file
	// itself — mirrors how linter.go builds it once per file.
	var comments []*ast.CommentRange
	ForEachComment(sf.AsNode(), func(comment *ast.CommentRange) {
		comments = append(comments, comment)
	}, sf)
	sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })

	commentGapStart := strings.Index(src, "1")
	commentGapEnd := strings.Index(src, "+")
	if !HasCommentInSpan(comments, commentGapStart, commentGapEnd) {
		t.Fatal("expected real block comment in numeric-expression gap")
	}

	stringStart := strings.Index(src, `"/*`)
	stringEnd := strings.Index(src, `*/"`) + len(`*/"`)
	if HasCommentInSpan(comments, stringStart, stringEnd) {
		t.Fatal("string literal comment markers must not count as comments")
	}

	if HasCommentInSpan(comments, commentGapStart, commentGapStart) {
		t.Fatal("empty span must not contain comments")
	}
}
