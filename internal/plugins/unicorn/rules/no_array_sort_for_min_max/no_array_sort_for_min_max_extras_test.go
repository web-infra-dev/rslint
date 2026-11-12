package no_array_sort_for_min_max

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArraySortForMinMaxExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Literal indexes and exact arity.
		{Code: `array.sort((a, b) => a - b)["0"];
array.sort((a, b) => a - b)[0n];
array.sort((a, b) => a - b)[-0];
array.sort((a, b) => a - b).at(+0);
array.sort((a, b) => a - b).at(-1n);
array.sort((a, b) => a - b).at(...[0]);
array.sort((a, b) => a - b).at(0, 1);
array.sort((a, b) => a - b)['at'](0);
array.sort((a, b) => a - b)?.at(0);
array.sort((a, b) => a - b).at?.(0);
array.sort(...comparators)[0];
array.sort((a, b) => a - b, compare)[0];`},
		// Plain numeric comparators.
		{Code: `array.sort((a = 0, b) => a - b)[0];
array.sort((a, ...b) => a - b)[0];
array.sort(([a], b) => a - b)[0];
array.sort((a, b, c) => a - b)[0];
array.sort((a, b) => a + b)[0];
array.sort((a, b) => c - b)[0];
array.sort((a, b) => a - a)[0];
array.sort((a, b) => { a - b; })[0];
array.sort((a, b) => { return; })[0];`},
		// Assignment targets.
		{Code: `([array.sort((a, b) => a - b)[0]] = values);
({x: array.sort((a, b) => a - b)[0] = value} = object);
for ([array.sort((a, b) => a - b)[0]] of values) {}
++(array.sort((a, b) => a - b)[0]);
(array.sort((a, b) => a - b)[0]) += value;
delete (array.sort((a, b) => a - b).at(-1));`},
		// Optional-chain and TS wrapper boundaries.
		{Code: `(object?.array).sort((a, b) => a - b)[0];
(object?.getArray<number>).toSorted((a, b) => a - b)[0];
(array.sort((a, b) => a - b) as number[])[0];
array.sort(((a, b) => a - b) as Comparator)[0];
delete (array.sort((a, b) => a - b).at(0) as number);
class Collection { #sort(fn: any) {} f() { return this.#sort((a, b) => a - b)[0]; } }`, FileName: "input.ts"},
		// BigInt receiver evidence.
		{Code: `const view = BigInt64Array.of(1n, 2n);
view.sort((a, b) => a - b)[0];
(flag ? new Int8Array() : new BigUint64Array()).sort((a, b) => a - b).at(-1);
/** @param {BigInt64Array} input */
function f(input) { return input.sort((a, b) => a - b)[0]; }
/** @returns {BigUint64Array} */
function make() { return new BigUint64Array(); }
make().toSorted((a, b) => a - b)[0];`},
		// BigInt aliases and intersections.
		{Code: `type View = (BigUint64Array & { tag: string }) | Int8Array;
function f(input: View) { return input.toSorted((a, b) => a - b)[0]; }
function g<T extends BigInt64Array>(input: T) { return input.sort((a, b) => a - b).at(0); }`, FileName: "input.ts"},
	}
	invalid := []minMaxCase{
		// Numeric literal spelling and parentheses.
		{
			code:       "const value = ((array).sort(((a, b) => (a) - (b))))[(0x0)];",
			method:     "min",
			position:   [4]int{1, 15, 1, 59},
			suggestion: "const value = Math.min(...array);",
		},
		// Negative index spelling.
		{
			code:       "const value = array.toSorted((a, b) => b - a).at((-(0b1)));",
			method:     "min",
			position:   [4]int{1, 15, 1, 59},
			suggestion: "const value = Math.min(...array);",
		},
		// Typed-array factory with numeric elements.
		{
			code:       "const value = Float64Array.of(3, 2).sort((a, b) => a - b)[0];",
			method:     "min",
			position:   [4]int{1, 15, 1, 61},
			suggestion: "const value = Math.min(...Float64Array.of(3, 2));",
		},
		// Access is read inside an assignment target.
		{code: "array.sort((a, b) => a - b)[0].property = value;", method: "min", position: [4]int{1, 1, 1, 31}, suggestion: "Math.min(...array).property = value;"},
		// Default initializer is read inside a pattern.
		{code: "[value = array.sort((a, b) => a - b)[0]] = items;", method: "min", position: [4]int{1, 10, 1, 40}, suggestion: "[value = Math.min(...array)] = items;"},
		// Optional chain in an argument.
		{
			code:       "const value = getArray(key?.id).toSorted((a, b) => a - b).at(-1);",
			method:     "max",
			position:   [4]int{1, 15, 1, 65},
			suggestion: "const value = Math.max(...getArray(key?.id));",
		},
		// Receiver with a sequence expression.
		{
			code:       "const value = (recordAccess(), array).sort((a, b) => a - b)[0];",
			method:     "min",
			position:   [4]int{1, 15, 1, 63},
			suggestion: "const value = Math.min(...(recordAccess(), array));",
		},
		// Constructor element key stays unwrapped.
		{code: "new constructors[array.sort((a, b) => a - b)[0]]();", method: "min", position: [4]int{1, 18, 1, 48}, suggestion: "new constructors[Math.min(...array)]();"},
		// Comment after the report stays in place.
		{code: "array.sort((a, b) => a - b)[0] /* keep */;", method: "min", position: [4]int{1, 1, 1, 31}, suggestion: "Math.min(...array) /* keep */;"},
		// TypeScript non-null receiver and comparator operands.
		{
			code:       "const value = array!.sort((a: number, b: number) => (a as number) - b!)[0];",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 15, 1, 75},
			suggestion: "const value = Math.min(...array!);",
		},
		// UTF-16 positions and JSX expression.
		{
			code:       "const view = <p>🧮{array.toSorted((a, b) => b - a).at(0)}</p>;",
			fileName:   "input.jsx",
			method:     "max",
			position:   [4]int{1, 20, 1, 57},
			suggestion: "const view = <p>🧮{Math.max(...array)}</p>;",
		},
		// The suggestion would call Math.min on the parameter passed as null.
		{
			code:     "function f(Math) { return [3, 1, 2].toSorted((a, b) => a - b)[0]; }\nf(null);",
			method:   "min",
			position: [4]int{1, 27, 1, 65},
		},
		{
			code:     "const max = values.sort((a, b) => a - b).at(-1);",
			globals:  map[string]any{"Math": "off"},
			method:   "max",
			position: [4]int{1, 13, 1, 48},
		},
	}
	runMinMaxCases(t, valid, invalid)
}

func TestNoArraySortForMinMaxEditDemand(t *testing.T) {
	t.Parallel()
	const code = `
const first = array.sort((a, b) => a - b)[0];
const last = array.toSorted((a, b) => a - b).at(/* keep */ -1);
`
	program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(code, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	options := linter.LintSingleFileOptions{
		Program: lintprogram.NewFromCompiler(program),
		File:    sourceFile.FileName(),
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: NoArraySortForMinMaxRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoArraySortForMinMaxRule.Run(ctx, nil)
			}}}
		},
	}
	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		options.Consumer = rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }}
		linter.LintSingleFile(options)
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
			if all[0].Suggestions == nil || len(*all[0].Suggestions) != 1 ||
				(*all[0].Suggestions)[0].Message.Description != "Replace with `Math.min(...array)`." {
				t.Fatal("expected a Math.min suggestion on the first diagnostic")
			}
		}
		for i, diagnostic := range diagnostics {
			if diagnostic.Range != all[i].Range || !reflect.DeepEqual(diagnostic.Message, all[i].Message) {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d produced an automatic fix", demand)
			}
			if i == 0 && demand&rule.EditDemandSuggestion != 0 {
				if !reflect.DeepEqual(diagnostic.Suggestions, all[i].Suggestions) {
					t.Errorf("demand %d changed the suggestion", demand)
				}
			} else if diagnostic.Suggestions != nil {
				t.Errorf("demand %d produced an unexpected suggestion for diagnostic %d", demand, i)
			}
		}
	}
}
