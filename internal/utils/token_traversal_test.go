package utils

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

const tokenTraversalSource = `/* leading */
const n = 0x1_f; // number
const text = '\u0061';
const big = 0x10n;
const pattern = /\/\* fake \*\//g;
` + "const template = `\\u0061${n /* expression */ + 1} tail`;\n" + `
const view = <Box<number> label="/* attribute */">{n}{/* child */}<span>text</span></Box>;
/* trailing */`

type tokenTraversalSnapshot struct {
	node, parent *ast.Node
	kind         ast.Kind
	pos, end     int
	flags        ast.NodeFlags
	literalFlags ast.TokenFlags
}

func snapshotTraversalToken(node *ast.Node) tokenTraversalSnapshot {
	var literalFlags ast.TokenFlags
	switch node.Kind {
	case ast.KindStringLiteral:
		literalFlags = node.AsStringLiteral().TokenFlags
	case ast.KindNumericLiteral:
		literalFlags = node.AsNumericLiteral().TokenFlags
	case ast.KindBigIntLiteral:
		literalFlags = node.AsBigIntLiteral().TokenFlags
	case ast.KindRegularExpressionLiteral:
		literalFlags = node.AsRegularExpressionLiteral().TokenFlags
	case ast.KindNoSubstitutionTemplateLiteral:
		literalFlags = node.AsNoSubstitutionTemplateLiteral().TemplateFlags
	case ast.KindTemplateHead:
		literalFlags = node.AsTemplateHead().TemplateFlags
	case ast.KindTemplateMiddle:
		literalFlags = node.AsTemplateMiddle().TemplateFlags
	case ast.KindTemplateTail:
		literalFlags = node.AsTemplateTail().TemplateFlags
	}
	return tokenTraversalSnapshot{node, node.Parent, node.Kind, node.Pos(), node.End(), node.Flags, literalFlags}
}

func parseTokenTraversalSource(source string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/traversal.tsx",
		Path:     "/traversal.tsx",
	}, source, core.ScriptKindTSX)
}

func collectTraversalTokens(sourceFile *ast.SourceFile) []tokenTraversalSnapshot {
	var tokens []tokenTraversalSnapshot
	ForEachToken(sourceFile.AsNode(), func(node *ast.Node) {
		tokens = append(tokens, snapshotTraversalToken(node))
	}, sourceFile)
	return tokens
}

func collectTraversalComments(sourceFile *ast.SourceFile) []ast.CommentRange {
	var comments []ast.CommentRange
	ForEachComment(sourceFile.AsNode(), func(comment *ast.CommentRange) {
		comments = append(comments, *comment)
	}, sourceFile)
	return comments
}

func assertTraversalTokens(t *testing.T, got, want []tokenTraversalSnapshot) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		// Comparing structs directly checks pointer identity as well as fields;
		// deep equality alone could accept a different node with the same fields.
		if got[i] != want[i] {
			t.Fatalf("token %d changed: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestTokenTraversalPreservesCanonicalNodesWithComments(t *testing.T) {
	for _, warmTokens := range []bool{false, true} {
		t.Run(fmt.Sprintf("warm_tokens_%t", warmTokens), func(t *testing.T) {
			sourceFile := parseTokenTraversalSource(tokenTraversalSource)
			var want []tokenTraversalSnapshot
			if warmTokens {
				want = collectTraversalTokens(sourceFile)
			}
			callbacks := 0
			ForEachComment(sourceFile.AsNode(), func(_ *ast.CommentRange) {
				callbacks++
				got := collectTraversalTokens(sourceFile)
				if want == nil {
					want = got
				}
				assertTraversalTokens(t, got, want)
				// Exercise GetChildren directly on nodes whose keywords and
				// punctuation are materialized from scanner-filled gaps.
				for _, child := range GetChildren(sourceFile.AsNode(), sourceFile) {
					if ast.IsTokenKind(child.Kind) {
						continue
					}
					left := GetChildren(child, sourceFile)
					right := GetChildren(child, sourceFile)
					if !slices.Equal(left, right) {
						t.Fatal("GetChildren did not reuse canonical nodes")
					}
				}
			}, sourceFile)
			if callbacks == 0 || len(want) == 0 {
				t.Fatal("fixture must exercise both comments and tokens")
			}
			assertTraversalTokens(t, collectTraversalTokens(sourceFile), want)
			assertTraversalTokens(t, collectTraversalTokens(sourceFile), want)
			if !slices.ContainsFunc(want, func(token tokenTraversalSnapshot) bool {
				return token.literalFlags != ast.TokenFlagsNone
			}) {
				t.Fatal("fixture must contain literals with nonzero token flags")
			}
		})
	}
}

func TestTokenTraversalNestedCallbacks(t *testing.T) {
	sourceFile := parseTokenTraversalSource(tokenTraversalSource)
	otherFile := parseTokenTraversalSource("/* other */ const x = 1; // end\n")
	want := collectTraversalComments(parseTokenTraversalSource(tokenTraversalSource))
	wantOther := collectTraversalComments(parseTokenTraversalSource(otherFile.Text()))
	if len(want) < 2 || len(wantOther) == 0 {
		t.Fatal("fixtures must contain comments")
	}
	var retained []*ast.CommentRange
	var values []ast.CommentRange
	ForEachComment(sourceFile.AsNode(), func(comment *ast.CommentRange) {
		retained = append(retained, comment)
		values = append(values, *comment)
		if len(retained) != 1 {
			return
		}
		// Nest comment walks inside a token callback, itself called inside a
		// comment callback. Every invocation must own its scanner and stack.
		nested := false
		ForEachToken(sourceFile.AsNode(), func(_ *ast.Node) {
			if nested {
				return
			}
			nested = true
			if got := collectTraversalComments(sourceFile); !slices.Equal(got, want) {
				t.Fatalf("nested same-file comments = %v, want %v", got, want)
			}
			if got := collectTraversalComments(otherFile); !slices.Equal(got, wantOther) {
				t.Fatalf("nested other-file comments = %v, want %v", got, wantOther)
			}
			if len(collectTraversalTokens(otherFile)) == 0 {
				t.Fatal("other-file token walk was empty")
			}
		}, sourceFile)
		if !nested {
			t.Fatal("nested callback was not exercised")
		}
	}, sourceFile)
	if !slices.Equal(values, want) {
		t.Fatalf("outer comments = %v, want %v", values, want)
	}
	seen := make(map[*ast.CommentRange]bool)
	for i, comment := range retained {
		if seen[comment] {
			t.Fatalf("comment %d reused a callback pointer", i)
		}
		seen[comment] = true
		if *comment != values[i] {
			t.Fatalf("retained comment %d changed after its callback", i)
		}
	}
}

func TestTokenTraversalStopsAtEveryToken(t *testing.T) {
	sourceFile := parseTokenTraversalSource(tokenTraversalSource)
	want := collectTraversalTokens(sourceFile)
	wantComments := collectTraversalComments(sourceFile)
	if len(want) == 0 || len(wantComments) == 0 {
		t.Fatal("fixture must contain tokens and comments")
	}
	for stop := range len(want) + 1 {
		t.Run(fmt.Sprintf("stop_%d", stop), func(t *testing.T) {
			var got []tokenTraversalSnapshot
			stopped := forEachToken(sourceFile.AsNode(), func(node *ast.Node) bool {
				got = append(got, snapshotTraversalToken(node))
				if len(got)-1 != stop {
					return false
				}
				// Nested walks at the stopping callback must not resume or
				// extend the outer traversal after it returns true.
				if gotComments := collectTraversalComments(sourceFile); !slices.Equal(gotComments, wantComments) {
					t.Fatal("comments changed in stopping callback")
				}
				assertTraversalTokens(t, collectTraversalTokens(sourceFile), want)
				return true
			}, sourceFile)
			if stopped != (stop < len(want)) {
				t.Fatalf("stopped = %t for index %d", stopped, stop)
			}
			assertTraversalTokens(t, got, want[:min(stop+1, len(want))])
		})
	}
}

func TestTokenTraversalConcurrentReadOnlySource(t *testing.T) {
	sourceFile := parseTokenTraversalSource(tokenTraversalSource)
	wantComments := collectTraversalComments(parseTokenTraversalSource(tokenTraversalSource))
	if len(wantComments) == 0 {
		t.Fatal("fixture must contain comments")
	}
	const workers = 8
	start := make(chan struct{})
	results := make(chan []tokenTraversalSnapshot, workers)
	for range workers {
		go func() {
			<-start
			var first []tokenTraversalSnapshot
			for range 4 {
				if got := collectTraversalComments(sourceFile); !slices.Equal(got, wantComments) {
					t.Errorf("concurrent comments = %v, want %v", got, wantComments)
				}
				got := collectTraversalTokens(sourceFile)
				if first == nil {
					first = got
				} else if !slices.Equal(got, first) {
					t.Error("concurrent token walk changed canonical nodes")
				}
			}
			results <- first
		}()
	}
	close(start)
	var got [][]tokenTraversalSnapshot
	for range workers {
		got = append(got, <-results)
	}
	wantTokens := collectTraversalTokens(sourceFile)
	if len(wantTokens) == 0 {
		t.Fatal("fixture must contain tokens")
	}
	for _, tokens := range got {
		assertTraversalTokens(t, tokens, wantTokens)
	}
}

func TestTokenTraversalDeepAndWide(t *testing.T) {
	const depth, width = 512, 1024
	tests := []struct {
		name, source string
		tokens       int
	}{
		{"deep", "const value = " + strings.Repeat("(", depth) + "0" + strings.Repeat(")", depth) + ";", 2*depth + 6},
		{"wide", "const values = [" + strings.Repeat("0,", width) + "];", 2*width + 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseTokenTraversalSource("/* first */ " + test.source + " /* last */")
			var comments []string
			ForEachComment(sourceFile.AsNode(), func(comment *ast.CommentRange) {
				comments = append(comments, sourceFile.Text()[comment.Pos():comment.End()])
			}, sourceFile)
			if !slices.Equal(comments, []string{"/* first */", "/* last */"}) {
				t.Fatalf("comments = %v", comments)
			}
			tokens := collectTraversalTokens(sourceFile)
			if len(tokens) != test.tokens {
				t.Fatalf("token count = %d, want %d", len(tokens), test.tokens)
			}
			assertTraversalTokens(t, collectTraversalTokens(sourceFile), tokens)
		})
	}
}
