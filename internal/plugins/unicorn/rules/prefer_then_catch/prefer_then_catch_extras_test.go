// TestPreferThenCatchExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / upstream issue it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package prefer_then_catch_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	prefer_then_catch "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_then_catch"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferThenCatchExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_then_catch.PreferThenCatchRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional-chain roots stay unmatched ----
			tsValid(`promise?.then(onFulfilled, onRejected);`),
			tsValid(`promise.then?.(onFulfilled, onRejected);`),
			tsValid(`(promise?.then)(onFulfilled, onRejected);`),

			// ---- Dimension 4: computed and non-identifier access keys do not match ----
			tsValid(`promise["then"](onFulfilled, onRejected);`),
			tsValid("promise[`then`](onFulfilled, onRejected);"),
			tsValid(`promise[0](onFulfilled, onRejected);`),
			tsValid(`promise[Symbol.iterator](onFulfilled, onRejected);`),

			// ---- Dimension 4: spread arguments disqualify the match ----
			// Locks in MatchDotMethodCall.RejectSpreadElement's spread rejection.
			tsValid(`promise.then(...handlers, onRejected);`),
			tsValid(`promise.then(onFulfilled, ...handlers);`),

			// ---- Dimension 4: argument-count boundaries ----
			tsValid(`promise.then();`),
			tsValid(`promise.then(onFulfilled);`),
			tsValid(`promise.then(onFulfilled, onRejected, extraArgument);`),

			// ---- Dimension 4: nullish handlers via TypeScript assertions stay nullish ----
			// `undefined` and `null` wrappers keep their semantic meaning after
			// unwrapTypeScriptExpression, matching upstream's isNullish pass.
			tsValid(`promise.then(onFulfilled, undefined as (error: unknown) => void);`),
			tsValid(`promise.then(null!, onRejected);`),
			tsValid(`promise.then(onFulfilled, void 0 as (error: unknown) => void);`),

			// N/A: declaration/container variants for functions, classes, properties,
			// private names, overloads, and abstract members are not relevant to a
			// CallExpression-and-argument-list rule.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: ESTree-transparent parentheses around the rejection handler ----
			extrasInvalid(`promise.then(onFulfilled, (onRejected));`),

			// ---- Dimension 4: trailing comma after rejection handler ----
			// Locks in trailingCommaEnd extending the removal past a trailing comma.
			extrasInvalid(`promise.then(onFulfilled, onRejected,);`),

			// ---- Dimension 4: type-asserted rejection handler ----
			extrasInvalidTs(`promise.then(onFulfilled, onRejected as (error: unknown) => void);`),

			// ---- Dimension 4: TS angle-bracket type cast on rejection handler ----
			extrasInvalidTs(`promise.then(onFulfilled, <(error: unknown) => void> onRejected);`),

			// ---- Dimension 4: multi-line call with trailing comma ----
			// Locks in that the removal covers both commas but preserves the leading
			// structure of the call.
			extrasInvalidWithCode(
				"promise\n\t.then(\n\t\tonFulfilled,\n\t\tonRejected,\n\t)\n\t.then(next);",
				"promise\n\t.then(\n\t\tonFulfilled\n\t).catch(onRejected)\n\t.then(next);",
			),

			// ---- Dimension 4: native Promise (callable catch) ----
			// Locks in canThenResultCatch: receiver is a built-in Promise and the
			// result type's `.catch` is callable, so the rule reports.
			extrasInvalidTs(`declare const promise: Promise<string>;` + "\n" +
				`promise.then(onFulfilled, onRejected);`),

			// ---- Dimension 4: comment between rejection handler and trailing paren ----
			// Locks in hasTrailingArgumentComment rejecting the suggestion (diagnostic
			// still fires, just without a fix).
			extrasInvalidNoFix(`promise.then(onFulfilled, onRejected /* Do not move this comment. */);`),

			// ---- Dimension 4: rejection handler as a call (not safe to move) ----
			// Locks in isRejectionHandlerSafeToMove rejecting call expressions.
			extrasInvalidNoFix(`promise.then(onFulfilled, createRejectionHandler());`),

			// ---- Rejection-handler comments move into `.catch(...)` ----
			// Locks in upstream behavior: handler source is reinserted verbatim, so
			// comments inside its body, expression, or parameters are preserved.
			extrasInvalidWithCode(
				`promise.then(a, x => { /* keep */ return x; });`,
				`promise.then(a).catch(x => { /* keep */ return x; });`,
			),
			extrasInvalidWithCode(
				`promise.then(a, x => /* keep */ x);`,
				`promise.then(a).catch(x => /* keep */ x);`,
			),
			extrasInvalidWithCode(
				`promise.then(a, function (/* keep */ x) { return x; });`,
				`promise.then(a).catch(function (/* keep */ x) { return x; });`,
			),

			// ---- Dimension 4: shadowed `undefined` parameter is NOT global ----
			// Locks in isGlobalUndefined using IsSymbolDeclaredInFile to detect a
			// local shadow that defeats the nullish-handler short-circuit.
			extrasInvalid(`function handlePromise(undefined) { promise.then(onFulfilled, undefined); }`),
		},
	)
}

// TestPreferThenCatchEditDemand verifies that suggestions are built only when
// requested and diagnostic identity is independent of edit demand.
func TestPreferThenCatchEditDemand(t *testing.T) {
	t.Parallel()

	const source = "promise.then(onFulfilled, onRejected);\n"
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
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     prefer_then_catch.PreferThenCatchRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_then_catch.PreferThenCatchRule.Run(ctx, nil)
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
	if output != "promise.then(onFulfilled).catch(onRejected);\n" {
		t.Fatalf("suggestion output = %q, want %q", output, "promise.then(onFulfilled).catch(onRejected);\n")
	}
}

func extrasInvalid(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, true, defaultSuggestionOutput(code)),
		},
	}
}

func extrasInvalidTs(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.ts",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, true, defaultSuggestionOutput(code)),
		},
	}
}

func extrasInvalidWithCode(code, expected string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, true, expected),
		},
	}
}

func extrasInvalidNoFix(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, false, ""),
		},
	}
}
