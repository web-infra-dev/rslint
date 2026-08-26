// TestSortVarsExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch, Dimension 4 row, or tsgo AST shape it covers, so future
// refactors cannot silently regress it. Upstream's migrated valid/invalid suite
// lives in sort_vars_upstream_test.go.
package sort_vars

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortVarsExtras(t *testing.T) {
	ignoreCase := map[string]any{"ignoreCase": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortVarsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: empty destructuring pattern has no sortable identifier ----
			{Code: "var {} = value;"},
			// ---- Dimension 4: empty array pattern has no sortable identifier ----
			{Code: "var [] = value;"},
			// ---- Dimension 4: one identifier needs no comparison ----
			{Code: "let only = value;"},
			// ---- Dimension 4: destructuring names are not sorted by this rule ----
			{Code: "const { z, a, nested: { y, b } } = value;"},
			// ---- Dimension 4: rest elements in binding patterns are ignored ----
			{Code: "const [z, ...a] = value;"},
			// ---- Dimension 4: separate declaration blocks never compare across statements ----
			{Code: "const b = 2; const a = 1;"},
			// ---- Real-user: eslint/eslint#2954 — a destructuring-only declaration does not crash under ignoreCase ----
			{Code: "var source = {m: 'M'}; var {m} = source;", Options: ignoreCase},
			// ---- Real-user: eslint/eslint#3474 — a leading empty pattern does not become an undefined comparison memo ----
			{Code: "var {}, a;", Options: ignoreCase},
			// ---- Locks in upstream filter(): patterns are skipped rather than treated as comparison boundaries ----
			{Code: "var a, {z} = source, b;"},
			// ---- N/A: access/key forms — only Identifier binding names participate ----
			// ---- N/A: class/function declaration forms — the target is a VariableDeclarationList ----
			// ---- N/A: overload/abstract members — they cannot contain variable declaration lists ----
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: a variable statement declaration list ----
			{
				Code:   "let b, a;",
				Output: []string{"let a, b;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortVars", Message: expectedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 9}},
			},
			// ---- Dimension 4: a for-loop header uses a bare VariableDeclarationList ----
			{
				Code:   "for (let b = 1, a = 0; a < b; a++) {}",
				Output: []string{"for (let a = 0, b = 1; a < b; a++) {}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortVars", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}},
			},
			// ---- Dimension 4: a TS type annotation belongs to the moved declarator ----
			{
				Code:   "let b: number = 2, a: number = 1;",
				Output: []string{"let a: number = 1, b: number = 2;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortVars", Line: 1, Column: 20}},
			},
			// ---- Dimension 4: parenthesized literal initializers remain fixable because ESTree erases the wrappers ----
			{
				Code:   "let b = ((2)), a = (1);",
				Output: []string{"let a = (1), b = ((2));"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 16)},
			},
			// ---- Dimension 4: TS non-null assertion initializer is non-literal and suppresses the fix ----
			{Code: "let b = 2, a = value!;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			// ---- Dimension 4: TS `as` initializer is non-literal and suppresses the fix ----
			{Code: "let b = 2, a = 1 as number;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			// ---- Dimension 4: TS `satisfies` initializer is non-literal and suppresses the fix ----
			{Code: "let b = 2, a = 1 satisfies number;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			// ---- Dimension 4: optional-chain initializer is non-literal and suppresses the fix ----
			{Code: "let b = 2, a = value?.x;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			// ---- Dimension 3: comments and multiline separators stay in their original layout slots ----
			{
				Code:   "let c = 3, /* first gap */\n    a = 1, // second gap\n    b = 2;",
				Output: []string{"let a = 1, /* first gap */\n    b = 2, // second gap\n    c = 3;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortVars", Line: 2, Column: 5},
					{MessageId: "sortVars", Line: 3, Column: 5},
				},
			},
			// ---- Dimension 3: a non-literal initializer anywhere in the participating list suppresses the whole fix ----
			{Code: "let c = 3, a = 1, b = sideEffect();", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12), sortVarsError(1, 19)}},
			// ---- Dimension 3: a non-literal initializer on an ignored pattern does not suppress the fix ----
			{
				Code:   "let b = 2, {x} = sideEffect(), a = 1;",
				Output: []string{"let a = 1, {x} = sideEffect(), b = 2;"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 32)},
			},
			// ---- Locks in upstream reduce() inversion arm: the previous accepted maximum is retained, producing two reports ----
			{
				Code:   "let d, c, a, b;",
				Output: []string{"let a, b, c, d;"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8), sortVarsError(1, 11), sortVarsError(1, 14)},
			},
			// ---- Locks in upstream fix() `fixed` arm: only the first of several reports carries the single list-wide edit ----
			{
				Code:   "let c, a, b;",
				Output: []string{"let a, b, c;"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8), sortVarsError(1, 11)},
			},
			// ---- Locks in upstream literal safety: every ESTree Literal family member permits a fix ----
			{
				Code:   "let z = /x/, y = 1n, x = null, w = false, v = 's';",
				Output: []string{"let v = 's', w = false, x = null, y = 1n, z = /x/;"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 14), sortVarsError(1, 22), sortVarsError(1, 32), sortVarsError(1, 43)},
			},
			// ---- Locks in upstream Literal distinction: a no-substitution template is still a TemplateLiteral and suppresses fixing ----
			{Code: "let b = 2, a = `text`;", Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 12)}},
			// ---- Locks in JavaScript UTF-16 relational ordering for non-BMP identifier names ----
			{
				Code:   "let ﾚ, 𐀀;",
				Output: []string{"let 𐀀, ﾚ;"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)},
			},
			// ---- Locks in JavaScript lowercasing for the Kelvin sign under ignoreCase ----
			{
				Code:    "let \u212A, j;",
				Output:  []string{"let j, \u212A;"},
				Options: ignoreCase,
				Errors:  []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)},
			},
			// ---- Locks in upstream sort comparator's equal-name arm: V8 reverses equal folded-name groups while fixing another inversion ----
			{
				Code:    "let a, A, z, b;",
				Output:  []string{"let A, a, b, z;"},
				Options: ignoreCase,
				Errors:  []rule_tester.InvalidTestCaseError{sortVarsError(1, 14)},
			},
			// ---- Schema default: an explicit empty option object is case-sensitive like no options ----
			{
				Code:    "let a, A;",
				Output:  []string{"let A, a;"},
				Options: map[string]any{},
				Errors:  []rule_tester.InvalidTestCaseError{sortVarsError(1, 8)},
			},
			// ---- Dimension 2: nested declaration lists are checked independently ----
			{
				Code:   "let b, a; function f() { let d, c; }",
				Output: []string{"let a, b; function f() { let c, d; }"},
				Errors: []rule_tester.InvalidTestCaseError{sortVarsError(1, 8), sortVarsError(1, 33)},
			},
		},
	)
}

func TestSortVarsEditDemand(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     "/edit-demand.ts",
	}, "let b = 2, a = 1;", core.ScriptKindTS)

	var declarationList *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclarationList {
			declarationList = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if declarationList == nil {
		t.Fatal("fixture has no variable declaration list")
	}

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(SortVarsRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})
		SortVarsRule.Run(ctx, nil)[ast.KindVariableDeclarationList](declarationList)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       diagnosticsOnly,
		rule.EditDemandAutofix:    autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got, want := withoutEdits(diagnostic), withoutEdits(allEdits); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.FixesPtr != nil || suggestionOnly.FixesPtr != nil {
		t.Fatal("non-autofix demand materialized fixes")
	}
	if autofixOnly.FixesPtr == nil || !reflect.DeepEqual(autofixOnly.FixesPtr, allEdits.FixesPtr) {
		t.Fatal("autofix and all-edits demands produced different fixes")
	}
	for _, diagnostic := range []rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		if diagnostic.Suggestions != nil {
			t.Fatal("autofix-only rule materialized suggestions")
		}
	}
}
