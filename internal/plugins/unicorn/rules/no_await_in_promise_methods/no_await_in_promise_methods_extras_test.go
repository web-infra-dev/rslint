// TestNoAwaitInPromiseMethodsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / upstream issue it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package no_await_in_promise_methods_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_await_in_promise_methods "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_await_in_promise_methods"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAwaitInPromiseMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_await_in_promise_methods.NoAwaitInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional-chain roots remain unmatched ----
			tsValid(`Promise?.all([await promise])`),
			tsValid(`Promise.all?.([await promise])`),
			tsValid(`(Promise?.all)([await promise])`),

			// ---- Dimension 4: TypeScript receiver wrappers are not transparent in ESTree ----
			tsValid(`Promise!.all([await promise])`),
			tsValid(`(Promise as any).all([await promise])`),
			tsValid(`(Promise satisfies PromiseConstructor).all([await promise])`),

			// ---- Dimension 4: computed and non-identifier access keys do not match ----
			tsValid(`Promise["all"]([await promise])`),
			tsValid("Promise[`all`]([await promise])"),
			tsValid(`Promise[0]([await promise])`),
			tsValid(`Promise[Symbol.iterator]([await promise])`),

			// ---- Dimension 4: TypeScript wrappers around the array or element are not direct ESTree matches ----
			tsValid(`Promise.all([await promise] as const)`),
			tsValid(`Promise.all([(await promise)!])`),
			tsValid(`Promise.all([(await promise) as unknown])`),
			tsValid(`Promise.all([(await promise) satisfies unknown])`),

			// ---- Dimension 4: spreads, nested arrays, and empty containers degrade without masking siblings ----
			tsValid(`Promise.all([...(await promise)])`),
			tsValid(`Promise.all([[await promise]])`),
			tsValid(`Promise.all([])`),

			// ---- Dimension 4: calls/new expressions and non-array argument containers stay unmatched ----
			tsValid(`Promise.all({value: await promise})`),
			tsValid(`new Promise.all([await promise])`),

			// N/A: declaration/container variants for functions, classes, properties,
			// private names, overloads, abstract members, and destructuring are not
			// inspected by this CallExpression-and-array-element rule.

			// Locks in isPromiseMethodCallWithArrayExpression(): only direct array elements are examined.
			tsValid(`Promise.all([Promise.resolve(await promise)])`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: ESTree-transparent parentheses around receivers, callees, arguments, and elements ----
			tsInvalid(`((Promise)).all([await promise])`, `await promise`, "all", `((Promise)).all([promise])`),
			tsInvalid(`(Promise.all)([await promise])`, `await promise`, "all", `(Promise.all)([promise])`),
			tsInvalid(`Promise.all((([await promise])))`, `await promise`, "all", `Promise.all((([promise])))`),
			tsInvalid(`Promise.all([((await promise))])`, `await promise`, "all", `Promise.all([((promise))])`),

			// ---- Dimension 4: generic calls preserve the ordinary CallExpression match ----
			tsInvalid(`Promise.all<string>([await promise])`, `await promise`, "all", `Promise.all<string>([promise])`),

			// ---- Dimension 4: nested calls report only the matching direct element in each call ----
			tsInvalid(
				`Promise.all([Promise.any([await promise])])`,
				`await promise`,
				"any",
				`Promise.all([Promise.any([promise])])`,
			),

			// ---- Dimension 4: local bindings do not change upstream's syntactic Promise-name match ----
			tsInvalid(
				`const Promise = customPromise; Promise.all([await promise])`,
				`await promise`,
				"all",
				`const Promise = customPromise; Promise.all([promise])`,
			),

			// Locks in matchesNameConstraint(): escaped identifier spelling is compared by identifier value.
			tsInvalid(`Promise.\u0061ll([await promise])`, `await promise`, "all", `Promise.\u0061ll([promise])`),

			// Locks in removeSpacesAfter(): all consecutive ECMAScript whitespace after `await` is removed.
			tsInvalid(`Promise.all([await	promise])`, `await	promise`, "all", `Promise.all([promise])`),
			tsInvalid("Promise.all([await\u00A0promise])", "await\u00A0promise", "all", `Promise.all([promise])`),
			tsInvalid(`Promise.all([await(promise)])`, `await(promise)`, "all", `Promise.all([(promise)])`),

			// Locks in removeSpacesAfter(): comments stop whitespace removal and remain in place.
			tsInvalid(
				"Promise.all([await // keep this comment\n  promise])",
				"await // keep this comment\n  promise",
				"all",
				"Promise.all([// keep this comment\n  promise])",
			),

			// ---- Real-user: #2257 proposal shape with an awaited request among concurrent work ----
			tsInvalid(
				`async function load() { return Promise.race([await first(), second()]) }`,
				`await first()`,
				"race",
				`async function load() { return Promise.race([first(), second()]) }`,
			),

			// ---- Real-user: #2257 proposal shape using Promise.all with mixed awaited and pending work ----
			tsInvalid(
				`Promise.all([await promise, anotherPromise])`,
				`await promise`,
				"all",
				`Promise.all([promise, anotherPromise])`,
			),

			// Locks in upstream create() loop arms: holes and spreads are skipped, while a direct awaited sibling reports.
			tsInvalid(
				`Promise.all([, await /* preserve */ request, anotherRequest])`,
				`await /* preserve */ request`,
				"all",
				`Promise.all([, /* preserve */ request, anotherRequest])`,
			),

			// Locks in the full multi-line diagnostic range and suggestion output.
			tsInvalid(
				"Promise.all([\n  await first,\n  second,\n])",
				`await first`,
				"all",
				"Promise.all([\n  first,\n  second,\n])",
			),
		},
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mts"}
}

func tsInvalid(code, target, method, suggestionOutput string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mts",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, target, method, suggestionOutput),
		},
	}
}

// TestNoAwaitInPromiseMethodsEditDemand verifies that suggestions are built
// only when requested and diagnostic identity is independent of edit demand.
func TestNoAwaitInPromiseMethodsEditDemand(t *testing.T) {
	t.Parallel()

	const source = "Promise.all([await\n  promise])\n"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "edit-demand.mts", "tsconfig.json")
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     no_await_in_promise_methods.NoAwaitInPromiseMethodsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return no_await_in_promise_methods.NoAwaitInPromiseMethodsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					got = append(got, diagnostic)
				},
			},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}

	base := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		if diagnostic.Range != base.Range ||
			diagnostic.Message.Id != base.Message.Id ||
			diagnostic.Message.Description != base.Message.Description ||
			!reflect.DeepEqual(diagnostic.Message.Data, base.Message.Data) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d materialized an autofix for a suggestion-only rule", demand)
		}
	}

	if diagnostics[rule.EditDemandNone].Suggestions != nil ||
		diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions materialized without suggestion demand")
	}
	suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
	allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
	if suggestionOnly == nil || allSuggestions == nil ||
		!reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
		t.Fatal("suggestion artifacts differ between suggestion-only and all demand")
	}
	output, _, _ := linter.ApplyRuleFixes(source, []rule.RuleSuggestion{(*allSuggestions)[0]})
	if output != "Promise.all([promise])\n" {
		t.Fatalf("suggestion output = %q, want %q", output, "Promise.all([promise])\n")
	}
}
