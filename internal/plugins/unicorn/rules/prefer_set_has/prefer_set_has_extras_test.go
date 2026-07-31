package prefer_set_has_test

// TestPreferSetHasExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_has"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferSetHasExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_set_has.PreferSetHasRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers on the includes call ----
			// Parenthesized receiver of a single includes lookup is still a
			// one-off lookup outside any loop → not reported.
			{Code: "const foo = [1, 2, 3];\nconst exists = (foo).includes(1);"},
			// N/A: TS non-null / as / satisfies on the receiver — the receiver
			// must be the plain declared identifier for isIncludesCall to match
			// (upstream compares `callee.object === node`), so wrapped receivers
			// never classify as an includes reference; they fall to the default
			// "unclassifiable" bail. Covered by the branch lock-in below.

			// ---- Dimension 4: optional chain on includes ----
			// `foo?.includes(1)` — optional call is excluded (upstream
			// optionalCall:false); single unclassifiable reference → bail.
			{Code: "const foo = [1, 2, 3];\nfor (let i = 0; i < 3; i++) { foo?.includes(1); }"},

			// ---- Dimension 4: computed / string-key includes ----
			// `foo['includes'](1)` is element access, not a dot method call →
			// unclassifiable → bail even inside a loop.
			{Code: "const foo = [1, 2, 3];\nfor (let i = 0; i < 3; i++) { foo['includes'](1); }"},

			// ---- Branch lock-in: getReferenceGroups default (unclassifiable) ----
			// A bare read of the array that is neither includes/length/extra
			// makes the whole rule bail, even with a valid includes call.
			{Code: "const foo = [1, 2, 3];\nsink(foo);\nfor (let i = 0; i < 3; i++) { foo.includes(1); }"},

			// ---- Branch lock-in: extra reference requires a unique literal ----
			// forEach (extra ref) present, but the initializer is an
			// Array.of() call (not a plain literal) → isKnownUniqueArrayExpression
			// is false → bail.
			{Code: "const foo = Array.of(1, 2, 3);\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) { foo.includes(1); }"},

			// ---- Branch lock-in: extra ref, literal with a non-comparable element ----
			// `[foo()]` element is not a statically comparable value → not a
			// known-unique array → bail when an extra ref is present.
			{Code: "const foo = [bar()];\nconst spread = [...foo];\nfunction has(v) { return foo.includes(v); }"},

			// ---- Branch lock-in: isMultipleCall false at top level ----
			// Single includes call not inside any loop/function boundary.
			{Code: "const foo = [1, 2, 3];\nfoo.includes(1);"},

			// ---- Real-user #3561: indexing in includes (foo[i]) ----
			// `bar.includes(foo[i])` — foo is the argument's object, not the
			// includes receiver, so it is an unclassifiable reference → bail.
			{Code: "const foo = [1, 2, 3];\nfor (const i of list) { bar.includes(foo[i]); }"},

			// ---- Real-user #2216: string false positive via slice/concat ----
			// `.slice()` on a const string identifier → not an array source.
			{Code: "const str = 'abc';\nconst foo = str.slice();\nfoo.includes('a') || foo.includes('b');"},
			{Code: "const str = 'abc';\nconst foo = str.concat('d');\nfoo.includes('a') || foo.includes('b');"},

			// ---- Branch lock-in: minimumItems with Array(negative/huge) ----
			// Array(-1) is not a valid array length → size unknown → below
			// minimumItems path bails.
			{Code: "const foo = Array(-1);\nfunction unicorn() { return foo.includes(1); }", Options: minimumItems(5)},
			// Array('x') → constructor size 1 < 5 → bail.
			{Code: "const foo = Array('x');\nfunction unicorn() { return foo.includes(1); }", Options: minimumItems(5)},
		},
		[]rule_tester.InvalidTestCase{
			// ---- PR #1479 review fix: parenthesized `.length` now classifies ----
			// `(foo).length` — a parenthesized `.length` read. isLengthRead now
			// unwraps parentheses (via WalkUpParenthesizedExpressions) like the
			// other classifiers, so the looped `includes` is reported and the third
			// line is rewritten to `(foo).size`. Previously this silently bailed.
			{
				Code:   "const foo = [1, 2, 3];\nfor (let i = 0; i < 3; i++) foo.includes(1);\nconst n = (foo).length;",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (let i = 0; i < 3; i++) foo.has(1);\nconst n = (foo).size;"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- PR #1479 review fix: NaN / undefined elements now modeled ----
			// evaluateStaticIdentifier now models `NaN` and `undefined` alongside
			// `Infinity`, so an array containing them is a known-unique literal, the
			// extra reference (forEach / spread) is allowed, and the rule reports +
			// autofixes to `new Set([...])` / `.has`. Previously these bailed.
			{
				Code:   "const foo = [NaN, 1, 2];\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) foo.includes(1);",
				Output: []string{"const foo = new Set([NaN, 1, 2]);\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) foo.has(1);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			{
				Code:   "const foo = [undefined, 1, 2];\nconst spread = [...foo];\nfor (let i = 0; i < 3; i++) foo.includes(1);",
				Output: []string{"const foo = new Set([undefined, 1, 2]);\nconst spread = [...foo];\nfor (let i = 0; i < 3; i++) foo.has(1);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Control for PR #1479 divergence: non-parenthesized `.length` ----
			// Contrast with the parenthesized `(foo).length` case above: the
			// non-parenthesized `.length` read has always classified correctly, so a
			// looped includes IS reported and both `.has`/`.size` rewrites apply.
			{
				Code:   "const foo = [1, 2, 3];\nfor (let i = 0; i < 3; i++) foo.includes(1);\nconst n = foo.length;",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (let i = 0; i < 3; i++) foo.has(1);\nconst n = foo.size;"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Control for PR #1479 divergence: Infinity element ----
			// Infinity has always been special-cased in evaluateStaticIdentifier;
			// kept alongside the NaN/undefined cases above to lock the three global
			// numeric identifiers together.
			{
				Code:   "const foo = [Infinity, 1, 2];\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) foo.includes(1);",
				Output: []string{"const foo = new Set([Infinity, 1, 2]);\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) foo.has(1);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Dimension 4: parenthesized receiver inside a loop reports ----
			// The `.includes` property is renamed even through the parens.
			{
				Code:   "const foo = [1, 2, 3];\nfor (let i = 0; i < 3; i++) { (foo).includes(1); }",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (let i = 0; i < 3; i++) { (foo).has(1); }"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- PR #1479 review fix: parenthesized extra-reference classifiers ----
			// isIterableUse / isArrayOrArgumentSpread / isAllowedForEachCall now
			// unwrap parentheses like isIncludesCall, so a parenthesized extra
			// reference is still recognized (rather than making the rule bail).
			// for-of iterable: `for (const x of (foo))`.
			{
				Code:   "const foo = [1, 2, 3];\nfor (const x of (foo)) log(x);\nfor (let i = 0; i < 3; i++) foo.includes(1);",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfor (const x of (foo)) log(x);\nfor (let i = 0; i < 3; i++) foo.has(1);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// spread argument: `[...(foo)]`.
			{
				Code:   "const foo = [1, 2, 3];\nconst copy = [...(foo)];\nfunction has(v) { return foo.includes(v); }",
				Output: []string{"const foo = new Set([1, 2, 3]);\nconst copy = [...(foo)];\nfunction has(v) { return foo.has(v); }"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
			// forEach receiver: `(foo).forEach(...)`.
			{
				Code:   "const foo = [1, 2, 3];\n(foo).forEach(x => log(x));\nfor (let i = 0; i < 3; i++) foo.includes(1);",
				Output: []string{"const foo = new Set([1, 2, 3]);\n(foo).forEach(x => log(x));\nfor (let i = 0; i < 3; i++) foo.has(1);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Branch lock-in: matchArrayStaticCall "from" arm ----
			// `Array.from(iterable)` with two includes calls → array source,
			// reported and wrapped.
			{
				Code:   "const foo = Array.from(bar);\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const foo = new Set(Array.from(bar));\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Branch lock-in: methodsReturnsArray via a listed method ----
			// `.toSorted()` is in methodsReturnsArray → array source.
			{
				Code:   "const foo = bar.toSorted();\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const foo = new Set(bar.toSorted());\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Branch lock-in: isIdentifierInitializedWithArray recursion ----
			// slice() on an identifier that is itself a const array literal →
			// treated as an array (the recursive const-initializer walk).
			{
				Code:   "const source = [1, 2, 3];\nconst foo = source.slice();\nfoo.includes(1) || foo.includes(2);",
				Output: []string{"const source = [1, 2, 3];\nconst foo = new Set(source.slice());\nfoo.has(1) || foo.has(2);"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 2, 7)},
			},

			// ---- Branch lock-in: known-unique literal enables extra refs ----
			// A unique primitive literal with a forEach extra ref + includes is
			// reported; the fix keeps forEach untouched and renames includes.
			{
				Code:   "const foo = [1, 2, 3];\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) { foo.includes(1); }",
				Output: []string{"const foo = new Set([1, 2, 3]);\nfoo.forEach(x => log(x));\nfor (let i = 0; i < 3; i++) { foo.has(1); }"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Branch lock-in: minimumItems boundary (exactly equal) ----
			// Size exactly equals minimumItems → reported.
			{
				Code:    "const foo = [1, 2, 3, 4, 5];\nfunction unicorn() { return foo.includes(1); }",
				Output:  []string{"const foo = new Set([1, 2, 3, 4, 5]);\nfunction unicorn() { return foo.has(1); }"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Branch lock-in: getObjectLength via Array.from({length}) ----
			// Static length object drives the size check at the boundary.
			{
				Code:    "const foo = Array.from({length: 6});\nfunction unicorn() { return foo.includes(1); }",
				Output:  []string{"const foo = new Set(Array.from({length: 6}));\nfunction unicorn() { return foo.has(1); }"},
				Options: minimumItems(5),
				Errors:  []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},

			// ---- Dimension 4: BigInt element uniqueness ----
			// BigInt literals are comparable static values; `[1n, 2n]` is a
			// known-unique array, so an extra spread ref does not block the report.
			{
				Code:   "const foo = [1n, 2n];\nconst spread = [...foo];\nfunction has(v) { return foo.includes(v); }",
				Output: []string{"const foo = new Set([1n, 2n]);\nconst spread = [...foo];\nfunction has(v) { return foo.has(v); }"},
				Errors: []rule_tester.InvalidTestCaseError{errorAt("foo", 1, 7)},
			},
		},
	)
}

// TestPreferSetHasEditDemand verifies the deferred fix/suggestion builders run
// only under their matching edit demand and never change the diagnostic
// identity. `foo = [1,2,3]` exposes an autofix; a non-rewritable type
// annotation (`[string, string]`) exposes a suggestion instead.
func TestPreferSetHasEditDemand(t *testing.T) {
	t.Parallel()

	configs := []struct {
		name           string
		fileName       string
		code           string
		wantFix        bool
		fixOutput      string
		wantSuggestion bool
		suggestOutput  string
	}{
		{
			name:      "autofix",
			fileName:  "edit-demand-autofix.ts",
			code:      "const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}\n",
			wantFix:   true,
			fixOutput: "const foo = new Set([1, 2, 3]);\nfunction unicorn() {\n\treturn foo.has(1);\n}\n",
		},
		{
			name:           "suggestion",
			fileName:       "edit-demand-suggestion.ts",
			code:           "const a: [string, string] = ['foo', 'bar'];\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(s)) {}\n}\n",
			wantSuggestion: true,
			suggestOutput:  "const a: [string, string] = new Set(['foo', 'bar']);\nfor (let i = 0; i < 3; i++) {\n\tif (a.has(s)) {}\n}\n",
		},
	}

	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createPreferSetHasProgram(t, config.fileName, config.code)
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone,
				rule.EditDemandAutofix,
				rule.EditDemandSuggestion,
				rule.EditDemandAll,
			} {
				got := lintPreferSetHasWithDemand(program, sourceFile, demand)
				if len(got) != 1 {
					t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}

			// Diagnostic identity (message + range) must be stable across demands.
			base := diagnostics[rule.EditDemandNone]
			for demand, diagnostic := range diagnostics {
				requireSamePreferSetHasDiagnostic(t, base, diagnostic, demand)
			}

			// None demand materializes no edits.
			if diagnostics[rule.EditDemandNone].FixesPtr != nil || diagnostics[rule.EditDemandNone].Suggestions != nil {
				t.Errorf("none demand unexpectedly materialized edits")
			}
			// Autofix-only never materializes suggestions; suggestion-only never fixes.
			if diagnostics[rule.EditDemandAutofix].Suggestions != nil {
				t.Errorf("autofix-only demand unexpectedly materialized suggestions")
			}
			if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
				t.Errorf("suggestion-only demand unexpectedly materialized autofixes")
			}

			autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
			allFixes := diagnostics[rule.EditDemandAll].FixesPtr
			if !config.wantFix {
				if autofixOnly != nil || allFixes != nil {
					t.Fatalf("unexpected autofixes: autofix=%#v all=%#v", autofixOnly, allFixes)
				}
			} else {
				if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
					t.Fatalf("autofix artifacts differ between autofix-only and all demand")
				}
				if applyFixes(config.code, *autofixOnly) != config.fixOutput {
					t.Fatalf("autofix output = %q, want %q", applyFixes(config.code, *autofixOnly), config.fixOutput)
				}
			}

			suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
			allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
			if !config.wantSuggestion {
				if suggestionOnly != nil || allSuggestions != nil {
					t.Fatalf("unexpected suggestions: suggestion=%#v all=%#v", suggestionOnly, allSuggestions)
				}
			} else {
				if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
					t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
				}
				if len(*suggestionOnly) != 1 {
					t.Fatalf("suggestions = %d, want 1", len(*suggestionOnly))
				}
				if applyFixes(config.code, (*suggestionOnly)[0].FixesArr) != config.suggestOutput {
					t.Fatalf("suggestion output = %q, want %q", applyFixes(config.code, (*suggestionOnly)[0].FixesArr), config.suggestOutput)
				}
			}
		})
	}
}

func lintPreferSetHasWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     prefer_set_has.PreferSetHasRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return prefer_set_has.PreferSetHasRule.Run(ctx, nil)
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

func requireSamePreferSetHasDiagnostic(t *testing.T, want, got rule.RuleDiagnostic, demand rule.EditDemand) {
	t.Helper()
	want.FixesPtr = nil
	want.Suggestions = nil
	got.FixesPtr = nil
	got.Suggestions = nil
	if !reflect.DeepEqual(got, want) {
		t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
	}
}

func createPreferSetHasProgram(t testing.TB, fileName, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

// applyFixes applies a set of non-overlapping RuleFix edits to code, in
// descending range order so earlier offsets stay valid.
func applyFixes(code string, fixes []rule.RuleFix) string {
	sorted := make([]rule.RuleFix, len(fixes))
	copy(sorted, fixes)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j-1].Range.Pos() < sorted[j].Range.Pos(); j-- {
			sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
		}
	}
	result := code
	for _, fix := range sorted {
		result = result[:fix.Range.Pos()] + fix.Text + result[fix.Range.End():]
	}
	return result
}
