package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestOutermostParenthesizedExpression(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "consume((((value)))); consume(plain);", core.ScriptKindTS)

	value := findNodeWithText(t, sourceFile, "value")
	outer := OutermostParenthesizedExpression(value)
	if got := TrimmedNodeText(sourceFile, outer); got != "(((value)))" {
		t.Fatalf("OutermostParenthesizedExpression(value) = %q, want %q", got, "(((value)))")
	}
	if outer.Parent == nil || !ast.IsCallExpression(outer.Parent) {
		t.Fatal("outermost parentheses should remain the direct call argument")
	}

	plain := findNodeWithText(t, sourceFile, "plain")
	if got := OutermostParenthesizedExpression(plain); got != plain {
		t.Fatal("unparenthesized node should be returned unchanged")
	}
	if got := OutermostParenthesizedExpression(nil); got != nil {
		t.Fatal("nil input should return nil")
	}
}

func TestESTreeCallCallee(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		scriptKind core.ScriptKind
		wantText   string
		wantNil    bool
	}{
		{name: "plain member", code: `P.shape({})`, scriptKind: core.ScriptKindTS, wantText: "P.shape"},
		{name: "optional call", code: `P.shape?.({})`, scriptKind: core.ScriptKindTS, wantText: "P.shape"},
		{name: "optional receiver", code: `P?.shape({})`, scriptKind: core.ScriptKindTS, wantText: "P?.shape"},
		{name: "parenthesized member optional call", code: `(P.shape)?.({})`, scriptKind: core.ScriptKindTS, wantText: "P.shape"},
		{name: "parenthesized optional receiver", code: `(P?.shape)({})`, scriptKind: core.ScriptKindTS, wantNil: true},
		{name: "parenthesized optional receiver and call", code: `(P?.shape)?.({})`, scriptKind: core.ScriptKindTS, wantNil: true},
		{name: "authored TypeScript assertion", code: `(P.shape as any)({})`, scriptKind: core.ScriptKindTS, wantText: "P.shape as any"},
		{name: "JSDoc type member", code: `/** @type {any} */ (P.shape)({})`, scriptKind: core.ScriptKindJS, wantText: "P.shape"},
		{name: "JSDoc type optional receiver", code: `/** @type {any} */ (P?.shape)({})`, scriptKind: core.ScriptKindJS, wantNil: true},
		{name: "bare JSDoc satisfies optional receiver", code: `/** @satisfies {any} */ P?.shape({})`, scriptKind: core.ScriptKindJS, wantText: "P?.shape"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/case.ts",
				Path:     "/case.ts",
			}, test.code, test.scriptKind)
			var call *ast.Node
			VisitDescendants(sourceFile.AsNode(), func(node *ast.Node) bool {
				if call == nil && node.Kind == ast.KindCallExpression {
					call = node
				}
				return true
			})
			if call == nil {
				t.Fatal("expected a call expression")
			}

			callee := ESTreeCallCallee(call.AsCallExpression().Expression)
			if test.wantNil {
				if callee != nil {
					t.Fatalf("ESTreeCallCallee() = %q, want nil", TrimmedNodeText(sourceFile, callee))
				}
				return
			}
			if callee == nil {
				t.Fatal("ESTreeCallCallee() = nil")
			}
			if got := TrimmedNodeText(sourceFile, callee); got != test.wantText {
				t.Fatalf("ESTreeCallCallee() = %q, want %q", got, test.wantText)
			}
		})
	}

	if got := ESTreeCallCallee(nil); got != nil {
		t.Fatal("nil input should return nil")
	}
}

func TestESTreeMembers(t *testing.T) {
	t.Parallel()

	first := &ast.Node{Kind: ast.KindMethodDeclaration}
	second := &ast.Node{Kind: ast.KindPropertyDeclaration}
	empty := &ast.Node{Kind: ast.KindSemicolonClassElement}

	withoutEmpty := []*ast.Node{first, second}
	if got := ESTreeMembers(withoutEmpty); &got[0] != &withoutEmpty[0] {
		t.Fatal("member slice without empty elements should be returned unchanged")
	}

	withEmpty := []*ast.Node{empty, first, empty, second, empty}
	got := ESTreeMembers(withEmpty)
	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("ESTreeMembers() = %v, want [%p %p]", got, first, second)
	}
	if len(withEmpty) != 5 || withEmpty[0] != empty || withEmpty[2] != empty || withEmpty[4] != empty {
		t.Fatal("filtering members mutated the AST-owned input slice")
	}
}
