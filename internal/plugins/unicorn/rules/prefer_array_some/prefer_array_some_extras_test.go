// TestPreferArraySomeExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
package prefer_array_some_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_some"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferArraySomeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_array_some.PreferArraySomeRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: Receiver wrappers ----

		// Optional-chain member access `foo?.find(fn)` — MatchDotMethodCall
		// rejects an optional member for the find branch (optionalMember:false).
		{Code: `if (foo?.find(fn)) {}`},
		// Optional call `foo.find?.(fn)` — optionalCall:false rejects it.
		{Code: `if (foo.find?.(fn)) {}`},
		// `.filter?.(fn).length > 0` — optional filter call, not matched.
		{Code: `array.filter?.(fn).length > 0`},
		// `array?.filter(fn).length > 0` — optional member on filter, not matched.
		{Code: `array?.filter(fn).length > 0`},

		// ---- Dimension 4: Access / key forms ----

		// Computed string key `foo["find"]` is not a dot call.
		{Code: `if (foo["find"](fn)) {}`},
		// Numeric-literal-keyed length access is not `.length`.
		{Code: `array.filter(fn)[0] > 0`},

		// ---- Dimension 4: findIndex arg-count boundary ----

		// findIndex with 0 args — ArgumentsLength:1 rejects it.
		{Code: `foo.findIndex() !== -1`},
		// findIndex with 2 args — rejected.
		{Code: `foo.findIndex(fn, extra) !== -1`},

		// ---- Dimension 4: RHS literal forms for findIndex ----

		// `!== 1` (not -1) must not match.
		{Code: `foo.findIndex(fn) !== 1`},
		// `> 0` (not -1, and `>` pairs with -1 only) must not match.
		{Code: `foo.findIndex(fn) > 0`},
		// `>= 1` (not 0) must not match.
		{Code: `foo.findIndex(fn) >= 1`},
		// `< 1` (not 0) must not match.
		{Code: `foo.findIndex(fn) < 1`},

		// N/A: autofix-comment interplay covered in invalid cases below.

		// ---- Locks in filter().length arm: `$`-prefixed receiver skipped ----
		{Code: `$foo.filter(fn).length > 0`},
		// Parenthesized `$`-prefixed receiver is still skipped.
		{Code: `($foo).filter(fn).length > 0`},

		// ---- Locks in filter().length arm: non-function first arg skipped ----
		{Code: `array.filter(1).length > 0`},
		{Code: `array.filter("x").length > 0`},
		{Code: `array.filter({}).length > 0`},
		{Code: `array.filter([]).length > 0`},
		{Code: `array.filter(new Foo()).length > 0`},
		// No first argument at all.
		{Code: `array.filter().length > 0`},

		// ---- Real-user: #1305 — `.filter(fn).length` with `.bind()` arg is a function ----
		// A `.bind()` call IS possibly a function, so it should still report
		// (covered in invalid). Here the negative: a non-bind call is skipped.
		{Code: `array.filter(getPredicate()).length > 0`},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized receiver in boolean position ----
		{
			Code:   `if ((foo).find(fn)) {}`,
			Errors: []rule_tester.InvalidTestCaseError{findErrorExtras(`if ((foo).some(fn)) {}`)},
		},
		// Double-parenthesized call in a control-flow test.
		{
			Code:   `if (((foo.find(fn)))) {}`,
			Errors: []rule_tester.InvalidTestCaseError{findErrorExtras(`if (((foo.some(fn)))) {}`)},
		},

		// ---- Locks in isCheckingUndefined arm: `!=` keeps positive ----
		{Code: `foo.find(fn) != undefined`, Errors: []rule_tester.InvalidTestCaseError{findErrorExtras(`foo.some(fn)`)}},

		// ---- Locks in findIndex arm: `>= 0` positive, `< 0` negated ----
		{Code: `foo.findIndex(fn) >= 0`, Output: []string{`foo.some(fn) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(fn) < 0`, Output: []string{`!foo.some(fn) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},

		// ---- Locks in filter().length arm: `.bind()` arg is a possible function ----
		// #1305-style: `array.filter(cb.bind(thisArg)).length > 0` reports.
		{Code: `array.filter(cb.bind(thisArg)).length > 0`, Output: []string{`array.some(cb.bind(thisArg))`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},

		// ---- Locks in filter().length arm: `!== 0` operator ----
		{Code: `array.filter(fn).length !== 0`, Output: []string{`array.some(fn)`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},

		// ---- Real-user: logical-expression boolean context (&& / ||) ----
		// `x && foo.find(fn)` where the whole thing is a control-flow test.
		{Code: `if (x && foo.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findErrorExtras(`if (x && foo.some(fn)) {}`)}},
	})
}

// findErrorExtras mirrors findError from the upstream test file (same package,
// but that helper lives in the other _test.go; redeclare a local one to keep
// the files independent).
func findErrorExtras(output string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageSome,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: messageSuggest, Output: output},
		},
	}
}

// TestPreferArraySomeEditDemand verifies the deferred fix / suggestion builders
// only materialize when their artifact category is demanded, and that the
// diagnostic identity (message + range) is invariant across demands. The rule
// emits a suggestion for `.find()` and an autofix for `.findIndex(...) !== -1`.
func TestPreferArraySomeEditDemand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		source         string
		wantFix        bool
		wantSuggestion bool
	}{
		{
			name:           "find suggestion",
			source:         "if (foo.find(fn)) {}\n",
			wantSuggestion: true,
		},
		{
			name:    "findIndex autofix",
			source:  "foo.findIndex(fn) !== -1;\n",
			wantFix: true,
		},
		{
			name:    "filter length autofix",
			source:  "array.filter(fn).length > 0;\n",
			wantFix: true,
		},
	}

	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createPreferArraySomeProgram(
				t,
				preferArraySomeFileName(index),
				tc.source,
			)
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone,
				rule.EditDemandAutofix,
				rule.EditDemandSuggestion,
				rule.EditDemandAll,
			} {
				got := lintPreferArraySomeWithDemand(program, sourceFile, demand)
				if len(got) != 1 {
					t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}

			// Diagnostic identity (range + message) is invariant across demands.
			base := diagnostics[rule.EditDemandNone]
			if base.FixesPtr != nil || base.Suggestions != nil {
				t.Errorf("EditDemandNone unexpectedly materialized edits")
			}
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != base.Range {
					t.Errorf("demand %d changed diagnostic range", demand)
				}
				if diagnostic.Message.Id != base.Message.Id ||
					diagnostic.Message.Description != base.Message.Description {
					t.Errorf("demand %d changed diagnostic message", demand)
				}
			}

			// Autofix appears only under Autofix / All.
			if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
				t.Errorf("suggestion-only demand materialized autofixes")
			}
			autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
			allFixes := diagnostics[rule.EditDemandAll].FixesPtr
			if tc.wantFix {
				if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
					t.Fatalf("autofix artifacts differ between autofix-only and all demand")
				}
			} else if autofixOnly != nil {
				t.Errorf("unexpected autofix: %#v", *autofixOnly)
			}

			// Suggestions appear only under Suggestion / All.
			if diagnostics[rule.EditDemandAutofix].Suggestions != nil {
				t.Errorf("autofix-only demand materialized suggestions")
			}
			suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
			allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
			if tc.wantSuggestion {
				if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
					t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
				}
			} else if suggestionOnly != nil {
				t.Errorf("unexpected suggestions: %#v", *suggestionOnly)
			}
		})
	}
}

func preferArraySomeFileName(index int) string {
	return "edit-demand-" + string(rune('a'+index)) + ".ts"
}

func createPreferArraySomeProgram(t testing.TB, fileName, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{
		tspath.ResolvePath(rootDir.Dir, fileName): code,
	})
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func lintPreferArraySomeWithDemand(program *compiler.Program, sourceFile *ast.SourceFile, demand rule.EditDemand) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     prefer_array_some.PreferArraySomeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return prefer_array_some.PreferArraySomeRule.Run(ctx, nil)
				},
			}}
		},
		ExcludePaths: []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}
