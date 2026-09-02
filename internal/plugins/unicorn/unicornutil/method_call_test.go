package unicornutil

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
)

func parseMethodCallTestSource(fileName, code string, scriptKind core.ScriptKind) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.Path(fileName),
	}, code, scriptKind)
}

func firstMethodCallTestCall(t *testing.T, sourceFile *ast.SourceFile) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if ast.IsCallExpression(node) {
			found = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatal("expected a call expression")
	}
	return found
}

func TestMatchDotMethodCallESTreeCalleeWrappers(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		code       string
		scriptKind core.ScriptKind
		want       bool
	}{
		{
			name:       "plain method",
			fileName:   "/test.js",
			code:       `object.method("value")`,
			scriptKind: core.ScriptKindJS,
			want:       true,
		},
		{
			name:       "parenthesized callee",
			fileName:   "/test.js",
			code:       `(object.method)("value")`,
			scriptKind: core.ScriptKindJS,
			want:       true,
		},
		{
			name:       "JSDoc type cast callee",
			fileName:   "/test.js",
			code:       `/** @type {any} */ (object.method)("value")`,
			scriptKind: core.ScriptKindJS,
			want:       true,
		},
		{
			name:       "JSDoc satisfies cast callee",
			fileName:   "/test.js",
			code:       `/** @satisfies {any} */ (object.method)("value")`,
			scriptKind: core.ScriptKindJS,
			want:       true,
		},
		{
			name:       "JSDoc cast does not hide optional member",
			fileName:   "/test.js",
			code:       `/** @type {any} */ (object?.method)("value")`,
			scriptKind: core.ScriptKindJS,
		},
		{
			name:       "authored TypeScript as expression",
			fileName:   "/test.ts",
			code:       `(object.method as any)("value")`,
			scriptKind: core.ScriptKindTS,
		},
		{
			name:       "authored TypeScript satisfies expression",
			fileName:   "/test.ts",
			code:       `(object.method satisfies any)("value")`,
			scriptKind: core.ScriptKindTS,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parseMethodCallTestSource(test.fileName, test.code, test.scriptKind)
			call := firstMethodCallTestCall(t, sourceFile)
			_, got := MatchDotMethodCall(call, DotMethodCallOptions{Method: "method"})
			if got != test.want {
				t.Fatalf("MatchDotMethodCall(%q) = %v, want %v", test.code, got, test.want)
			}
		})
	}
}
