package new_for_builtins

import (
	"reflect"
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestDeferredReportsPreserveDiagnostics(t *testing.T) {
	source := `Array();
		new String();
		new Symbol();
		Date();
		Date(1);
		WebAssembly();`

	withoutEdits := runRuleForTest(t, source, rule.EditDemandNone)
	withEdits := runRuleForTest(t, source, rule.EditDemandAll)
	if len(withoutEdits) != len(withEdits) {
		t.Fatalf("diagnostic count without edits = %d, with edits = %d", len(withoutEdits), len(withEdits))
	}
	if len(withEdits) != 6 {
		t.Fatalf("diagnostic count = %d, want 6", len(withEdits))
	}

	for index := range withEdits {
		withoutEdit := withoutEdits[index]
		withEdit := withEdits[index]
		if withoutEdit.Range != withEdit.Range ||
			withoutEdit.Message.Id != withEdit.Message.Id ||
			withoutEdit.Message.Description != withEdit.Message.Description ||
			!reflect.DeepEqual(withoutEdit.Message.Data, withEdit.Message.Data) {
			t.Fatalf("diagnostic %d changed with edit demand", index)
		}
		if withoutEdit.FixesPtr != nil || withoutEdit.Suggestions != nil {
			t.Fatalf("diagnostic %d materialized edits with EditDemandNone", index)
		}
	}

	wantAutofix := []bool{true, false, true, true, false, false}
	wantSuggestion := []bool{false, false, false, false, true, false}
	for index, diagnostic := range withEdits {
		if got := diagnostic.FixesPtr != nil; got != wantAutofix[index] {
			t.Errorf("diagnostic %d autofix presence = %t, want %t", index, got, wantAutofix[index])
		}
		if got := diagnostic.Suggestions != nil; got != wantSuggestion[index] {
			t.Errorf("diagnostic %d suggestion presence = %t, want %t", index, got, wantSuggestion[index])
		}
	}

	disabled := runRuleForTest(
		t,
		"/* eslint-disable unicorn/new-for-builtins */\nArray();",
		rule.EditDemandAll,
	)
	if len(disabled) != 0 {
		t.Fatalf("disabled diagnostic count = %d, want 0", len(disabled))
	}
}

func TestReferenceResolutionWithAndWithoutRefStore(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "global builtin",
			source: "Array();",
			want:   []string{"Array"},
		},
		{
			name:   "local variable",
			source: "const Array = function() {}; Array();",
			want:   []string{""},
		},
		{
			name:   "exported local variable",
			source: "export const Array = function() {}; Array();",
			want:   []string{""},
		},
		{
			name:   "exported alias",
			source: "export const A = Array; A();",
			want:   []string{"Array"},
		},
		{
			name:   "hoisted variable",
			source: "function run() { Array(); var Array = function() {}; }",
			want:   []string{""},
		},
		{
			name:   "body var does not shadow parameter initializer",
			source: "function run(value = Array()) { var Array = function() {}; }",
			want:   []string{"Array"},
		},
		{
			name:   "destructured parameter",
			source: "function run({Array}) { Array(); }",
			want:   []string{""},
		},
		{
			name:   "destructured catch binding",
			source: "try {} catch ({Array}) { Array(); }",
			want:   []string{""},
		},
		{
			name:   "switch clauses share lexical scope",
			source: "switch (value) { case 0: Array(); break; case 1: const Array = function() {}; }",
			want:   []string{""},
		},
		{
			name:   "local global object",
			source: "const globalThis = {Array() {}}; globalThis.Array();",
			want:   []string{""},
		},
		{
			name:   "destructured alias",
			source: "const {Array: A} = globalThis; A();",
			want:   []string{"Array"},
		},
		{
			name:   "type-only declaration does not block alias source",
			source: "const A = Array; interface Array {} A();",
			want:   []string{"Array"},
		},
		{
			name:   "later local global object blocks alias source",
			source: "function run() { const {Array: A} = globalThis; const globalThis = {}; A(); }",
			want:   []string{""},
		},
		{
			name: "inner shadow of alias",
			source: `const {Array: A} = globalThis;
				{ const A = function() {}; A(); }
				A();`,
			want: []string{"", "Array"},
		},
		{
			name:   "type-only declaration",
			source: "interface Map {} Map();",
			want:   []string{"Map"},
		},
		{
			name:   "runtime namespace declaration",
			source: "namespace Intl { export const local = 1; } Intl.DateTimeFormat();",
			want:   []string{""},
		},
		{
			name:   "runtime enum declaration",
			source: "enum Map {} Map();",
			want:   []string{""},
		},
		{
			name:   "alias named like global object",
			source: "const {Array: window} = globalThis; window();",
			want:   []string{"Array"},
		},
		{
			name: "same alias spelling in disjoint scopes",
			source: `{ const {Array: A} = globalThis; A(); }
				{ const {Map: A} = globalThis; A(); }`,
			want: []string{"Array", "Map"},
		},
		{
			name:   "later local global object",
			source: "function run() { globalThis.Array(); const globalThis = {}; }",
			want:   []string{""},
		},
	}

	for _, testCase := range tests {
		for _, withRefs := range []bool{false, true} {
			mode := "fallback"
			if withRefs {
				mode = "refs"
			}
			t.Run(testCase.name+"/"+mode, func(t *testing.T) {
				got := resolveCallReferences(t, testCase.source, withRefs)
				if !slices.Equal(got, testCase.want) {
					t.Fatalf("resolved references = %q, want %q", got, testCase.want)
				}
			})
		}
	}
}

func resolveCallReferences(t *testing.T, source string, withRefs bool) []string {
	t.Helper()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/new-for-builtins-reference-test.ts",
		Path:     tspath.Path("/new-for-builtins-reference-test.ts"),
	}, source, core.ScriptKindTS)
	binder.BindSourceFile(sourceFile)

	ctx := rule.RuleContext{SourceFile: sourceFile}
	if withRefs {
		options := core.CompilerOptions{}
		ctx.Refs = rule.NewRefStore(sourceFile, &options, nil)
	}
	state := newRuleState(ctx)

	references := make([]string, 0)
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			if ref, ok := state.referenceFromExpression(call.Expression); ok {
				references = append(references, ref.name)
			} else {
				references = append(references, "")
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return references
}

func runRuleForTest(t *testing.T, source string, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/new-for-builtins-deferred-report-test.ts",
		Path:     tspath.Path("/new-for-builtins-deferred-report-test.ts"),
	}, source, core.ScriptKindTS)
	binder.BindSourceFile(sourceFile)

	comments := rule.NewCommentStore(sourceFile)
	diagnostics := make([]rule.RuleDiagnostic, 0)
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(NewForBuiltinsRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Demand: demand,
		Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	options := core.CompilerOptions{}
	ctx.Refs = rule.NewRefStore(sourceFile, &options, nil)

	listeners := NewForBuiltinsRule.Run(ctx, nil)
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return diagnostics
}
