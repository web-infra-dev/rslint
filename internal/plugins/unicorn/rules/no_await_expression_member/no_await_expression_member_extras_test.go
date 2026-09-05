// TestNoAwaitExpressionMemberExtras locks in upstream branches, AST edge
// shapes, and real-user examples. The complete upstream migration lives in
// no_await_expression_member_upstream_test.go.
package no_await_expression_member_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_await_expression_member"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAwaitExpressionMemberExtras(t *testing.T) {
	jsdoc := memberInvalid(`const value = /** @type {any} */ (await promise).value`, "value",
		`const {value} = /** @type {any} */ await promise`)
	jsdoc.FileName = "file.mjs"
	jsdocInitializer := memberInvalid(`const value = /** @type {any} */ ((await promise).value)`, "value",
		`const {value} = /** @type {any} */ (await promise)`)
	jsdocInitializer.FileName = "file.mjs"

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_await_expression_member.NoAwaitExpressionMemberRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream create() early return: the object must be await.
			{Code: `const value = (promise).value`},
			{Code: `const value = await promise.value`},
			{Code: `const value = object[await key]`},
			// ---- Dimension 4: authored TS wrappers are not transparent. ----
			{Code: `const value = ((await promise) as any).value`},
			{Code: `const value = ((await promise) satisfies any).value`},
			{Code: `const value = (<any>(await promise)).value`},
			{Code: `const value = (await promise)!.value`},
			// A sequence, unlike source parentheses, also changes the object.
			{Code: `const value = (0, await promise).value`},
			// ---- Dimension 4 N/A: missing function/class bodies do not affect
			// this member-expression rule. No options or name/key matching across
			// separate members is performed. ----
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: source parentheses are transparent on both the
			// receiver and initializer; only receiver parentheses are removed. ----
			// Locks in upstream array-fix arms for values 0 and 1, including
			// alternate numeric spellings and parenthesized computed keys.
			memberInvalid(`const x = (((await promise)))[((0))]`, "0", `const [x] = await promise`),
			memberInvalid(`const x = (((await promise)[1]))`, "1", `const [, x] = ((await promise))`),
			memberInvalid(`const x = (await promise)[0x1]`, "0x1", `const [, x] = await promise`),
			memberInvalid(`const x = (await promise)[0b0]`, "0b0", `const [x] = await promise`),
			memberInvalid(`const x = (await promise)[1.0]`, "1.0", `const [, x] = await promise`),
			memberInvalid(`const x = (await promise)[1e0]`, "1e0", `const [, x] = await promise`),
			// Locks in upstream object-fix arm: names match after decoding.
			memberInvalid(`const value = (((await promise))).value`, "value", `const {value} = await promise`),
			memberInvalid(`const v\u0061lue = (await promise).value`, "value", `const {v\u0061lue} = await promise`),
			memberInvalid(`const value = (await promise).v\u0061lue`, `v\u0061lue`, `const {value} = await promise`),
			jsdoc,
			jsdocInitializer,

			// ---- Dimension 4: computed/private keys all report, but cannot use
			// the identifier-property fix or numeric 0/1 fix. ----
			// Locks in upstream numeric-literal and property-Identifier gates.
			memberInvalid(`const x = (await promise)["0"]`, `"0"`),
			memberInvalid("const x = (await promise)[`x`]", "`x`"),
			memberInvalid(`const x = (await promise)[Symbol.iterator]`, "Symbol.iterator"),
			memberInvalid(`const x = (await promise)[0n]`, "0n"),
			memberInvalid(`const x = (await promise)[-0]`, "-0"),
			memberInvalid(`const x = (await promise)[+1]`, "+1"),
			memberInvalid(`const x = (await promise)[2]`, "2"),
			memberInvalid(`const x = (await promise)[1 as number]`, "1 as number"),
			memberInvalid(`const x = (await promise)[1 satisfies number]`, "1 satisfies number"),
			memberInvalid(`const x = (await promise)[1!]`, "1!"),
			memberInvalid(`class C { #value; async f() { const value = (await promise).#value; } }`, "#value"),

			// ---- Dimension 4: optional members report without a fix, whereas
			// an optional operation inside the awaited operand can be fixed. ----
			// Locks in upstream !memberExpression.optional in both fix branches.
			memberInvalid(`const x = (await promise)?.x`, "x"),
			memberInvalid(`const x = (await promise)?.[1]`, "1"),
			memberInvalid(`const x = (await promise?.()) .x`, "x", `const {x} = await promise?.()`),
			memberInvalid(`const x = (await promise.x)?.y?.()`, "y"),

			// Locks in upstream variable-declarator, identifier-binding,
			// initializer-identity, matching-name, and type-annotation gates.
			memberInvalid(`x = (await promise)[1]`, "1"),
			memberInvalid(`const renamed = (await promise).x`, "x"),
			memberInvalid(`const x: number = (await promise)[1]`, "1"),
			memberInvalid(`const x: number = (await promise).x`, "x"),
			memberInvalid(`const x = ((await promise).x) as number`, "x"),
			memberInvalid(`const x = ((await promise)[0])!`, "0"),
			memberInvalid(`for (const x of (await promise).x) {}`, "x"),
			// ---- Regression: unlike upstream v74.0.0, resource declarations
			// report without fixes because using cannot bind a pattern. ----
			memberInvalid(`using x = (await promise).x;`, "x"),
			memberInvalid(`using x = (await promise)[0];`, "0"),
			memberInvalid(`using x = (await promise)[1];`, "1"),
			memberInvalid(`await using x = (await promise).x;`, "x"),
			memberInvalid(`await using x = (await promise)[0];`, "0"),
			memberInvalid(`await using x = (await promise)[1];`, "1"),
			memberInvalid(`async function f() { using x = (((await promise).x)); }`, "x"),
			memberInvalid(`async function f() { await using x = (((await promise)[1])); }`, "1"),
			// ---- Dimension 4: empty/rest bindings and spread containers. ----
			memberInvalid(`const {} = (await promise).x`, "x"),
			memberInvalid(`const [...rest] = (await promise)[0]`, "0"),
			memberInvalid(`const x = {...(await promise).x}`, "x"),

			// ---- Dimension 4: comments, whitespace, and multiline ranges. ----
			memberInvalid(`const /* before */ x /* after */ = (/* inner */ await promise /* end */).x`, "x",
				`const /* before */ {x} /* after */ = /* inner */ await promise /* end */`),
			// Upstream removes comments in the deleted member suffix, too.
			memberInvalid(`const x = (await promise) /* member */ [/* key */ 0]`, "0", `const [x] = await promise`),
			memberInvalid("const x = (\n  await promise\n).x;", "x", "const {x} = \n  await promise\n;"),
			memberInvalid("const x = (await promise)[\n  1\n];", "1", "const [, x] = await promise;"),

			// ---- Dimension 4: nested members report independently, not once
			// per chain. The outer fix can leave another diagnostic behind. ----
			{
				Code: `const x = (await (await promise).y).x`, FileName: "file.mts",
				Errors: []rule_tester.InvalidTestCaseError{
					memberError(`const x = (await (await promise).y).x`, "x"),
					memberError(`const x = (await (await promise).y).x`, "y"),
				},
				Output: []string{`const {x} = await (await promise).y`},
			},
			// ---- Real-user: upstream #2557, dynamic import in a condition. ----
			memberInvalid(`if ((await import("electron-squirrel-startup")).default) { app.exit(); }`, "default"),
			// ---- Real-user: upstream #1458, filtering Promise.all results. ----
			memberInvalid(`const filtered = (await Promise.all(promises)).filter(Boolean);`, "filter"),
		},
	)
}

func TestNoAwaitExpressionMemberEditDemand(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ code, output string }{
		{`const value = (await promise).value`, `const {value} = await promise`},
		{`const value = (await promise)[1]`, `const [, value] = await promise`},
		{`const value = (await promise)?.value`, ""},
		{`using value = (await promise).value`, ""},
		{`await using value = (await promise)[1]`, ""},
	} {
		t.Run(test.code, func(t *testing.T) {
			t.Parallel()
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/edit-demand.mts", Path: "/edit-demand.mts",
			}, test.code, core.ScriptKindTS)
			var baseline rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll,
			} {
				var diagnostics []rule.RuleDiagnostic
				comments := rule.NewCommentStore(sourceFile)
				ctx := rule.RuleContext{
					SourceFile: sourceFile, Comments: comments,
					DisableManager: rule.NewDisableManager(sourceFile, comments),
				}.WithDiagnosticConsumer(no_await_expression_member.NoAwaitExpressionMemberRule.Name,
					rule.SeverityError, rule.DiagnosticConsumer{
						Demand: demand,
						Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
					})
				listeners := no_await_expression_member.NoAwaitExpressionMemberRule.Run(ctx, nil)
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if listener := listeners[node.Kind]; listener != nil {
						listener(node)
					}
					return node.ForEachChild(visit)
				}
				visit(sourceFile.AsNode())
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				if demand == rule.EditDemandNone {
					baseline = diagnostic
				}
				withoutFixes := diagnostic
				withoutFixes.FixesPtr = nil
				if !reflect.DeepEqual(withoutFixes, baseline) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := demand&rule.EditDemandAutofix != 0 && test.output != ""
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Fatalf("demand %d: unexpected fixes: %+v", demand, diagnostic.FixesPtr)
				}
				if wantFix {
					output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
					if !fixed || output != test.output {
						t.Errorf("demand %d: fixed = %v, output = %q, want %q", demand, fixed, output, test.output)
					}
				}
			}
		})
	}
}
