// Expected reduce fixes also remove sparse-array holes.
// Additional cases verified against eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-object-from-entries.js
package prefer_object_from_entries_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_object_from_entries"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferObjectFromEntriesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_object_from_entries.PreferObjectFromEntriesRule,
		[]rule_tester.ValidTestCase{
			// AST, scope, options and edit boundaries; expected diagnostics and edits checked against v79.0.0.
			// computed/private methods.
			{Code: "pairs[\"reduce\"](object => ({...object, key}), {});", FileName: "case.js"},
			// computed/private methods.
			{Code: "class C { #reduce; run() { return this.#reduce(object => ({...object, key}), {}); } }", FileName: "case.js"},
			// optional/typed callees.
			{Code: "(pairs?.reduce)(object => ({...object, key}), {});", FileName: "case.js"},
			// optional/typed callees.
			{Code: "(pairs.reduce as any)(object => ({...object, key}), {});", FileName: "case.ts", Options: []any{map[string]any{}}},
			// typed wrappers.
			{Code: "pairs.reduce((object => ({...object, key})) as any, {});", FileName: "case.ts", Options: []any{map[string]any{}}},
			// typed wrappers.
			{Code: "pairs.reduce(object => ({...object, key}), {} as any);", FileName: "case.ts", Options: []any{map[string]any{}}},
			// typed wrappers.
			{Code: "pairs.reduce(object => ({...object, key} as any), {});", FileName: "case.ts", Options: []any{map[string]any{}}},
			// callback shape.
			{Code: "pairs.reduce(() => ({...object, key}), {});", FileName: "case.js"},
			// accumulator references.
			{Code: "pairs.reduce(object => ({...object, [object]: value}), {});", FileName: "case.js"},
			// block callback.
			{Code: "pairs.reduce(object => { return {...object, key}; }, {});", FileName: "case.js"},
			// computed/optional functions.
			{Code: "_[\"fromPairs\"](pairs); _.fromPairs?.(pairs); (_.fromPairs)(...pairs);", FileName: "case.js"},
			// optional boundary.
			{Code: "(_?.fromPairs)(pairs);", FileName: "case.js"},
			// private function path.
			{Code: "class C { #fromPairs; run() { return this.#fromPairs(pairs); } }", FileName: "case.js", Options: []any{map[string]any{"functions": []any{"this.fromPairs"}}}},
			// function arity.
			{Code: "_.fromPairs(pairs, extra); foo(pairs);", FileName: "case.js"},
			// statement boundaries.
			{Code: "export const result = {}; for (const [key, value] of pairs) result[key] = value;", FileName: "case.js"},
			// statement boundaries.
			{Code: "const result = {}; ; for (const [key, value] of pairs) result[key] = value;", FileName: "case.js"},
			// binding patterns.
			{Code: "const {result} = {}; for (const [key, value] of pairs) result[key] = value;", FileName: "case.js"},
			// binding patterns.
			{Code: "const result = {}; for ([key, value] of pairs) result[key] = value;", FileName: "case.js"},
			// binding patterns.
			{Code: "const result = {}; for (const [, value] of pairs) result[key] = value;", FileName: "case.js"},
			// binding patterns.
			{Code: "const result = {}; for (const [key, value, extra] of pairs) result[key] = value;", FileName: "case.js"},
			// binding patterns.
			{Code: "const result = {}; for (const [[key], value] of pairs) result[key] = value;", FileName: "case.js"},
			// binding patterns.
			{Code: "const result = {}; for (const [key, result] of pairs) result[key] = result;", FileName: "case.js"},
			// loop shape.
			{Code: "const result = {}; for (const key in pairs) result[key] = value;", FileName: "case.js"},
			// loop shape.
			{Code: "const result = {}; for (const [key, value] of pairs) {}", FileName: "case.js"},
			// loop shape.
			{Code: "const result = {}; for (const [key, value] of pairs) { { result[key] = value; } }", FileName: "case.js"},
			// loop shape.
			{Code: "const result = {}; for (const [key, value] of pairs) { if (ok) result[key] = value; }", FileName: "case.js"},
			// loop references.
			{Code: "const result = {}; for (const [key, value] of source(result)) result[key] = value;", FileName: "case.js"},
			// TS assignment wrappers.
			{Code: "const result = {}; for (const [key, value] of pairs) (result as any)[key] = value;", FileName: "case.ts", Options: []any{map[string]any{}}},
			// TS assignment wrappers.
			{Code: "const result = {}; for (const [key, value] of pairs) result[key] = value as any;", FileName: "case.ts", Options: []any{map[string]any{}}},
			// typed factories.
			{Code: "const values = BigUint64Array.of(...input); values.reduce((o, v) => ({...o, [v]: v}), {});", FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			// The rewrite skips holes and preserves callback indices.
			{
				Code: `const records = new Array(3);
records[2] = {id: 'b', value: 2};
const result = records.reduce((acc, record, index) => ({...acc, [record.id]: record.value + index}), {});`,
				FileName: "case.js",
				Output: []string{`const records = new Array(3);
records[2] = {id: 'b', value: 2};
const result = Object.fromEntries(records.flatMap((record, index) => [[record.id, record.value + index]]));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 3, Column: 24, EndLine: 3, EndColumn: 30},
				},
			},
			// The original return annotation would reject the entry pair.
			{
				Code: `const pairs: number[] = [1, 2];
const result = pairs.reduce((o, v): Record<string, number> => ({...o, [v]: v}), {});`,
				FileName: "case.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 2, Column: 22, EndLine: 2, EndColumn: 28},
				},
			},
			// AST, scope, options and edit boundaries; expected diagnostics and edits checked against v79.0.0.
			// computed sequences.
			{
				Code: "pairs.reduce(object => ({...object, [(first(), key)]: (next(), value)}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[(first(), key), (next(), value)]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// prototype spelling.
			{
				Code: "pairs.reduce(object => ({...object, [\"__proto__\"]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[\"__proto__\", value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// nested shadowing.
			{
				Code: "pairs.reduce(object => ({...object, key: (object => object)(value)}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', (object => object)(value)]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// JSX value.
			{
				Code: "pairs.reduce((object, value) => ({...object, [value.id]: <Value item={value} />}), {});", FileName: "case.tsx",
				Options: []any{map[string]any{}},
				Output:  []string{"Object.fromEntries(pairs.flatMap((value) => [[value.id, <Value item={value} />]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// typed parameters.
			{
				Code: "pairs.reduce((object: Record<string, number>, value: number) => ({...object, [value]: value}), {});", FileName: "case.ts",
				Options: []any{map[string]any{}},
				Output:  []string{"Object.fromEntries(pairs.flatMap((value: number) => [[value, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// generic callback.
			{
				Code: "pairs.reduce(<T,>(object: Record<string, T>, value: T) => ({...object, key: value}), {});", FileName: "case.ts",
				Options: []any{map[string]any{}},
				Output:  []string{"Object.fromEntries(pairs.flatMap(<T,>(value: T) => [['key', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// JS JSDoc callback.
			{
				Code: "pairs.reduce(object => /** @type {any} */ ({...object, key}), {});", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// JS JSDoc receiver.
			{
				Code: "(/** @type {any} */ (pairs)).reduce(object => ({...object, key}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries((/** @type {any} */ (pairs)).flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
				},
			},
			// JS JSDoc initial value.
			{
				Code: "pairs.reduce(object => ({...object, key}), /** @type {any} */ ({}));", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// default options.
			{
				Code: "_.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{}},
				Output:  []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			// explicit default.
			{
				Code: "lodash.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{}}},
				Output:  []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `lodash.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// trimmed options.
			{
				Code: "utils.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{" utils.fromPairs\ufeff"}}},
				Output:  []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `utils.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// custom paths.
			{
				Code: "this.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"this.fromPairs"}}},
				Output:  []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `this.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			// custom meta path.
			{
				Code: "import.meta.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"import.meta.fromPairs"}}},
				Output:  []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `import.meta.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// duplicate builtin.
			{
				Code: "_.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"_.fromPairs"}}},
				Output:  []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			// parenthesized function.
			{
				Code: "((_.fromPairs))(pairs);", FileName: "case.js",
				Output: []string{"((Object.fromEntries))(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 14},
				},
			},
			// multiline/UTF16 ranges.
			{
				Code: "const emoji = \"💩\";\n\t((lodash\n .fromPairs))(pairs);", FileName: "case.js",
				Output: []string{"const emoji = \"💩\";\n\t((Object.fromEntries))(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `lodash.fromPairs()`.", Line: 2, Column: 4, EndLine: 3, EndColumn: 12},
				},
			},
			// UTF16 reduce range.
			{
				Code: "const emoji = \"💩\"; pairs.reduce(o => ({...o, key}), {});", FileName: "case.js",
				Output: []string{"const emoji = \"💩\"; Object.fromEntries(pairs.flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 33},
				},
			},
			// loop var declaration.
			{
				Code: "var result = {}; for (let [key, value] of pairs) (result[(key)] = (value));", FileName: "case.js",
				Output: []string{"var result = Object.fromEntries(pairs); "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 18, EndLine: 1, EndColumn: 76},
				},
			},
			// loop scope.
			{
				Code: "switch (mode) { case 0: const result = {}; for (const [key, value] of pairs) result[key] = value; break; }", FileName: "case.js",
				Output: []string{"switch (mode) { case 0: const result = Object.fromEntries(pairs);  break; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 44, EndLine: 1, EndColumn: 98},
				},
			},
			// static block scope.
			{
				Code: "class C { static { const result = {}; for (const [key, value] of pairs) result[key] = value; } }", FileName: "case.js",
				Output: []string{"class C { static { const result = Object.fromEntries(pairs);  } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 39, EndLine: 1, EndColumn: 93},
				},
			},
			// iterable shadowing.
			{
				Code: "const result = {}; for (const [key, value] of (() => { const result = pairs; return result; })()) result[key] = value;", FileName: "case.js",
				Output: []string{"const result = Object.fromEntries((() => { const result = pairs; return result; })()); "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 20, EndLine: 1, EndColumn: 119},
				},
			},
			// iterable property name.
			{
				Code: "const result = {}; for (const [key, value] of source.result) result[key] = value;", FileName: "case.js",
				Output: []string{"const result = Object.fromEntries(source.result); "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 20, EndLine: 1, EndColumn: 82},
				},
			},
			// iterable sequence.
			{
				Code: "const result = {}; for (const [key, value] of (first, pairs)) result[key] = value;", FileName: "case.js",
				Output: []string{"const result = Object.fromEntries((first, pairs)); "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 20, EndLine: 1, EndColumn: 83},
				},
			},
			// loop CRLF.
			{
				Code: "const result = {};\r\nfor (const [key, value] of pairs) {\r\n result[key] = value;\r\n}\r\nuse(result);", FileName: "case.js",
				Output: []string{"const result = Object.fromEntries(pairs);\r\nuse(result);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// trailing comment boundary.
			{
				Code: "const result = {}; for (const [key, value] of pairs) result[key] = value; /* keep */ use(result);", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 20, EndLine: 1, EndColumn: 74},
				},
			},
			// function type arguments.
			{
				Code: "_.fromPairs<Record<string, number>>(pairs);", FileName: "case.ts",
				Options: []any{map[string]any{}},
				Output:  []string{"Object.fromEntries<Record<string, number>>(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
		},
	)
}

func TestPreferObjectFromEntriesEditDemand(t *testing.T) {
	for _, test := range []struct{ name, code, output, reported string }{
		{"reduce", "pairs.reduce(o => ({...o, key}), {});", "Object.fromEntries(pairs.flatMap(() => [['key', key]]));", "reduce"},
		{"function", "_.fromPairs(pairs);", "Object.fromEntries(pairs);", "_.fromPairs"},
		{"loop", "const o = {}; for (const [k, v] of pairs) o[k] = v;", "const o = Object.fromEntries(pairs); ", "for (const [k, v] of pairs) o[k] = v;"},
		{"comment", "pairs.reduce(o => ({...o, /* keep */ key}), {});", "", "reduce"},
		{"type-arguments", "pairs.reduce<Result>(o => ({...o, key}), {});", "", "reduce"},
	} {
		t.Run(test.name, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			var baseline *rule.RuleDiagnostic
			var autofixes []rule.RuleFix
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: prefer_object_from_entries.PreferObjectFromEntriesRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return prefer_object_from_entries.PreferObjectFromEntriesRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				start := strings.Index(test.code, test.reported)
				if diagnostic.Range.Pos() != start || diagnostic.Range.End() != start+len(test.reported) {
					t.Fatalf("demand %d: unexpected range: %+v", demand, diagnostic.Range)
				}
				wantFix := test.output != "" && demand&rule.EditDemandAutofix != 0
				if (diagnostic.FixesPtr != nil) != wantFix || diagnostic.Suggestions != nil {
					t.Fatalf("demand %d: unexpected edits: %+v", demand, diagnostic)
				}
				if wantFix {
					output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
					if !fixed || output != test.output {
						t.Fatalf("demand %d: fixed to %q, want %q", demand, output, test.output)
					}
					if autofixes != nil && !reflect.DeepEqual(autofixes, *diagnostic.FixesPtr) {
						t.Fatalf("demand %d changed fixes", demand)
					}
					autofixes = append([]rule.RuleFix(nil), (*diagnostic.FixesPtr)...)
				}
				diagnostic.FixesPtr = nil
				if baseline == nil {
					baseline = &diagnostic
				} else if !reflect.DeepEqual(*baseline, diagnostic) {
					t.Fatalf("demand %d changed diagnostic", demand)
				}
			}
		})
	}
}

// Additional JSX, TypeScript, source-text and statement-boundary cases.
// Expectations were verified with eslint-plugin-unicorn v79.0.0.
func TestPreferObjectFromEntriesASTBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_object_from_entries.PreferObjectFromEntriesRule,
		[]rule_tester.ValidTestCase{
			// JSX member accumulator.
			{Code: "pairs.reduce(o => ({...o, value: <o.Component />}), {});", FileName: "case.tsx"},
			// JSX component accumulator.
			{Code: "pairs.reduce(O => ({...O, value: <O />}), {});", FileName: "case.tsx"},
			// type query accumulator.
			{Code: "pairs.reduce(o => ({...o, key: value as typeof o}), {});", FileName: "case.ts"},
			// escaped prototype.
			{Code: "pairs.reduce(o => ({...o, \"\\u005f_proto__\": value}), {});", FileName: "case.js"},
			// unbraced parent boundary.
			{Code: "const o = {}; if (ready) for (const [k, v] of pairs) o[k] = v;", FileName: "case.js"},
			// label parent boundary.
			{Code: "const o = {}; label: for (const [k, v] of pairs) o[k] = v;", FileName: "case.js"},
			// different block boundary.
			{Code: "const o = {}; { for (const [k, v] of pairs) o[k] = v; }", FileName: "case.js"},
			// prior non-declaration.
			{Code: "use({}); for (const [k, v] of pairs) o[k] = v;", FileName: "case.js"},
			// empty configured path.
			{Code: "_.unknown(pairs);", FileName: "case.js", Options: []any{map[string]any{"functions": []any{"", "utils..fromPairs", " . "}}}},
			// instantiated callee.
			{Code: "(fromPairs<Record<string, number>>)(pairs);", FileName: "case.ts", Options: []any{map[string]any{"functions": []any{"fromPairs"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// JSX intrinsic name.
			{
				Code: "pairs.reduce(o => ({...o, value: <o />}), {});", FileName: "case.tsx",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['value', <o />]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// JSX property name.
			{
				Code: "pairs.reduce(o => ({...o, value: <UI.o />}), {});", FileName: "case.tsx",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['value', <UI.o />]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// shadowed type parameter.
			{
				Code: "pairs.reduce(o => ({...o, key: <o,>(value: o) => value}), {});", FileName: "case.tsx",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', <o,>(value: o) => value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// escaped identifiers.
			{
				Code: "pairs.reduce(\\u006f => ({...o, \\u006bey: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['\\u006bey', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// escaped function property.
			{
				Code: "_.fr\\u006fmPairs(pairs);", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// property name is not a reference.
			{
				Code: "pairs.reduce(o => Object.assign(o, {o: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['o', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// comma before initial value.
			{
				Code: "pairs.reduce(o => ({...o, key}) /* retained */, {} /* retained */);", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]] /* retained */ /* retained */));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// comment outside callback.
			{
				Code: "pairs.reduce((/* retained */ o => ({...o, key}) /* retained */), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap((/* retained */ () => [['key', key]] /* retained */)));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// CRLF callback comments.
			{
				Code: "pairs.reduce(o => ({...o, // retained\r\n key}), {});", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// namespace statements.
			{
				Code: "namespace N { const o = {}; for (const [k, v] of pairs) o[k] = v; }", FileName: "case.ts",
				Output: []string{"namespace N { const o = Object.fromEntries(pairs);  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 29, EndLine: 1, EndColumn: 66},
				},
			},
			// default clause statements.
			{
				Code: "switch (x) { default: const o = {}; for (const [k, v] of pairs) o[k] = v; }", FileName: "case.js",
				Output: []string{"switch (x) { default: const o = Object.fromEntries(pairs);  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 37, EndLine: 1, EndColumn: 74},
				},
			},
			// typed iterable.
			{
				Code: "const o: Record<string, number> = {}; for (const [k, v] of pairs as [string, number][]) o[k] = v;", FileName: "case.ts",
				Output: []string{"const o: Record<string, number> = Object.fromEntries(pairs as [string, number][]); "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 39, EndLine: 1, EndColumn: 98},
				},
			},
			// conditional iterable.
			{
				Code: "const o = {}; for (const [k, v] of (ready ? a : b)) o[k] = v;", FileName: "case.js",
				Output: []string{"const o = Object.fromEntries((ready ? a : b)); "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 1, Column: 15, EndLine: 1, EndColumn: 62},
				},
			},
			// no semicolon in nested scope.
			{
				Code: "{ const o = {}\nfor (const [k, v] of pairs) o[k] = v; }", FileName: "case.js",
				Output: []string{"{ const o = Object.fromEntries(pairs)\n; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 2, EndColumn: 38},
				},
			},
			// super configured path.
			{
				Code: "class C extends Parent { f() { return super.fromPairs(pairs); } }", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"super.fromPairs"}}},
				Output:  []string{"class C extends Parent { f() { return Object.fromEntries(pairs); } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `super.fromPairs()`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 54},
				},
			},
			// new.target configured path.
			{
				Code: "function C() { return new.target.fromPairs(pairs); }", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"new.target.fromPairs"}}},
				Output:  []string{"function C() { return Object.fromEntries(pairs); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `new.target.fromPairs()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 43},
				},
			},
		},
	)
}
