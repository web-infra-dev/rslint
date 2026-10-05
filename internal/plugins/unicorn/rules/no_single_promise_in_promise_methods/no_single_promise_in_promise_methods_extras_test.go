package no_single_promise_in_promise_methods_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_single_promise_in_promise_methods"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSinglePromiseInPromiseMethodsEditDemand(t *testing.T) {
	const code = "await Promise.race([promise]); Promise.any([other]); Promise.all([last]);"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(code, "edit-demand.mts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	r := no_single_promise_in_promise_methods.NoSinglePromiseInPromiseMethodsRule
	diagnostics := make(map[rule.EditDemand][]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: r.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return r.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics[demand] = append(diagnostics[demand], d)
			}},
		})
		if len(diagnostics[demand]) != 3 {
			t.Fatalf("demand %d: expected three diagnostics, got %d", demand, len(diagnostics[demand]))
		}
	}
	all := diagnostics[rule.EditDemandAll]
	if len(all[0].Fixes()) != 1 || all[1].Suggestions == nil || len(*all[1].Suggestions) != 2 ||
		all[2].FixesPtr != nil || all[2].Suggestions != nil {
		t.Fatal("expected an autofix, two suggestions, and a report without edits")
	}
	for demand, got := range diagnostics {
		for i, diagnostic := range got {
			if diagnostic.Range != all[i].Range || !reflect.DeepEqual(diagnostic.Message, all[i].Message) {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			var fixes *[]rule.RuleFix
			if demand&rule.EditDemandAutofix != 0 {
				fixes = all[i].FixesPtr
			}
			var suggestions *[]rule.RuleSuggestion
			if demand&rule.EditDemandSuggestion != 0 {
				suggestions = all[i].Suggestions
			}
			if !reflect.DeepEqual(diagnostic.FixesPtr, fixes) || !reflect.DeepEqual(diagnostic.Suggestions, suggestions) {
				t.Errorf("demand %d returned unexpected edits for diagnostic %d", demand, i)
			}
		}
	}
}

func TestNoSinglePromiseInPromiseMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_single_promise_in_promise_methods.NoSinglePromiseInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "Promise.race([]); Promise.any([,,]);", FileName: "case.mjs"},
			{Code: "(Promise?.race)([promise]); (Promise.race)?.([promise]);", FileName: "case.mjs"},
			{Code: "Promise!.race([promise]); (Promise as any).race([promise]); (Promise satisfies PromiseConstructor).race([promise]);", FileName: "case.ts"},
			{Code: "Promise.race!([promise]); Promise.race([promise] as const); Promise.race([promise] satisfies unknown);", FileName: "case.ts"},
			{Code: "(<PromiseConstructor>Promise).race([promise]);", FileName: "case.ts"},
			{Code: "class Container { static #race() {} static run() { Promise.#race([promise]); } }", FileName: "case.mjs"},
		},
		[]rule_tester.InvalidTestCase{
			// An outer autofix exposes the nested call to await on the next pass.
			{
				Code:     "await Promise.race([Promise.any([promise])])",
				FileName: "case.mjs",
				Output:   []string{"await Promise.any([promise])", "await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "no-single-promise-in-promise-methods/error",
						Message:   "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line:      1, Column: 20, EndLine: 1, EndColumn: 44,
					},
					{
						MessageId: "no-single-promise-in-promise-methods/error",
						Message:   "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line:      1, Column: 33, EndLine: 1, EndColumn: 42,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "await Promise.race([promise])"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "await Promise.race([Promise.resolve(promise)])"},
						},
					},
				},
			},
			// Parenthesized assignment targets remain identifiers in ESTree.
			{
				Code:     "[(foo)] = await Promise.all([promise])",
				FileName: "case.mjs",
				Output:   []string{"foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "no-single-promise-in-promise-methods/error",
					Message:   "Wrapping single-element array with `Promise.all()` is unnecessary.",
					Line:      1, Column: 29, EndLine: 1, EndColumn: 38,
				}},
			},
			// Parentheses around receiver, callee, array, call, and value
			{
				Code:     "await (((Promise).race)(([(promise)])))",
				FileName: "case.mjs",
				Output:   []string{"await ((promise))"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 26, EndLine: 1, EndColumn: 37,
					},
				},
			},
			// Local binding still matches the syntactic name
			{
				Code:     "const Promise = custom; Promise.any([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 37, EndLine: 1, EndColumn: 46,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "const Promise = custom; promise"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "const Promise = custom; Promise.resolve(promise)"},
						},
					},
				},
			},
			// Escaped names use identifier values
			{
				Code:     "Promise.r\\u0061ce([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 19, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "promise"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(promise)"},
						},
					},
				},
			},
			// JavaScript JSDoc casts are transparent to matching
			{
				Code:     "(/** @type {any} */ (Promise)).race([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 37, EndLine: 1, EndColumn: 46,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "(/** @type {any} */ (Promise)).resolve(promise)"},
						},
					},
				},
			},
			// JSDoc cast on the array preserves comments in the resolve suggestion
			{
				Code:     "Promise.race(/** @type {any[]} */ ([promise]))",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 36, EndLine: 1, EndColumn: 45,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(/** @type {any[]} */ (promise))"},
						},
					},
				},
			},
			// JSDoc cast around the call remains when unwrapping
			{
				Code:     "await /** @type {Promise<void>} */ (Promise.race([promise]))",
				FileName: "case.mjs",
				Output:   []string{"await /** @type {Promise<void>} */ (promise)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 50, EndLine: 1, EndColumn: 59,
					},
				},
			},
			// Generic calls and TS as precedence
			{
				Code:     "await Promise.race<Value>([promise as Value])",
				FileName: "case.ts",
				Output:   []string{"await (promise as Value)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 27, EndLine: 1, EndColumn: 45,
					},
				},
			},
			// TS satisfies precedence
			{
				Code:     "await Promise.any([promise satisfies Value])",
				FileName: "case.ts",
				Output:   []string{"await (promise satisfies Value)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 19, EndLine: 1, EndColumn: 44,
					},
				},
			},
			// Awaited assertions and non-null values
			{
				Code:     "await Promise.race([<Value>promise]); await Promise.any([promise!]);",
				FileName: "case.ts",
				Output:   []string{"await <Value>promise; await promise!;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 20, EndLine: 1, EndColumn: 36,
					},
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 57, EndLine: 1, EndColumn: 67,
					},
				},
			},
			// Optional-chain value needs parentheses when unwrapped without await
			{
				Code:     "foo()\nPromise.any([object?.promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 2, Column: 13, EndLine: 2, EndColumn: 30,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "foo()\n;(object?.promise)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "foo()\nPromise.resolve(object?.promise)"},
						},
					},
				},
			},
			// JSX expression value
			{
				Code:     "Promise.race([<div />])",
				FileName: "case.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "(<div />)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(<div />)"},
						},
					},
				},
			},
			// Preserve nested parentheses in destructuring fixes
			{
				Code:     "const [foo,] = (await (Promise.all([(promise)])))",
				FileName: "case.mjs",
				Output:   []string{"const foo = (await ((promise)))"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 36, EndLine: 1, EndColumn: 47,
					},
				},
			},
			// Assignment destructuring through parentheses
			{
				Code:     "([foo,] = (await (Promise.all([promise]))))",
				FileName: "case.mjs",
				Output:   []string{"(foo = (await (promise)))"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 31, EndLine: 1, EndColumn: 40,
					},
				},
			},
			// A member assignment target is not a single identifier
			{
				Code:     "[object.foo] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 34, EndLine: 1, EndColumn: 43,
					},
				},
			},
			// Nested pattern does not unwrap
			{
				Code:     "const [[foo]] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 35, EndLine: 1, EndColumn: 44,
					},
				},
			},
			// Multiple bindings do not unwrap
			{
				Code:     "const [foo, bar] = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 38, EndLine: 1, EndColumn: 47,
					},
				},
			},
			// Object binding does not unwrap
			{
				Code:     "const {0: foo} = await Promise.all([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 36, EndLine: 1, EndColumn: 45,
					},
				},
			},
			// Loop iterable is not an initializer
			{
				Code:     "for (const [foo] of await Promise.all([promise])) {}",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 39, EndLine: 1, EndColumn: 48,
					},
				},
			},
			// Zero written in hexadecimal with extra parentheses
			{
				Code:     "const foo = ((await Promise.all([promise]))[(0x0)])",
				FileName: "case.mjs",
				Output:   []string{"const foo = (await promise)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 33, EndLine: 1, EndColumn: 42,
					},
				},
			},
			// Numeric separators in a zero index
			{
				Code:     "const foo = (await Promise.all([promise]))[0.0_0]",
				FileName: "case.ts",
				Output:   []string{"const foo = await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			// String and bigint indices do not match the numeric literal
			{
				Code:     "const a = (await Promise.all([promise]))[\"0\"]; const b = (await Promise.all([promise]))[0n]",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 30, EndLine: 1, EndColumn: 39,
					},
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 77, EndLine: 1, EndColumn: 86,
					},
				},
			},
			// Optional index access is wrapped in a ChainExpression
			{
				Code:     "const foo = (await Promise.all([promise]))?.[0]",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 32, EndLine: 1, EndColumn: 41,
					},
				},
			},
			// Compound assignments are safe replacement positions
			{
				Code:     "foo += (await Promise.all([promise]))[0]",
				FileName: "case.mjs",
				Output:   []string{"foo += await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 27, EndLine: 1, EndColumn: 36,
					},
				},
			},
			// Returned index access is not a safe replacement position
			{
				Code:     "async function load() { return (await Promise.all([promise]))[0] }",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 51, EndLine: 1, EndColumn: 60,
					},
				},
			},
			// Await wrapper assertions stop the direct-parent match
			{
				Code:     "const [foo] = (await Promise.all([promise])) as [Value]",
				FileName: "case.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 34, EndLine: 1, EndColumn: 43,
					},
				},
			},
			// Comments in the receiver prevent unwrapping
			{
				Code:     "Promise /* receiver */ .race([promise])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 30, EndLine: 1, EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise /* receiver */ .resolve(promise)"},
						},
					},
				},
			},
			// Keep all comments and trailing commas in resolve suggestions
			{
				Code:     "Promise.any(/* argument */ ([/* element */ (promise), /* trailing */]),)",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 29, EndLine: 1, EndColumn: 70,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(/* argument */ (/* element */ (promise) /* trailing */),)"},
						},
					},
				},
			},
			// Comments in the await-to-call gap survive
			{
				Code:     "const foo = (await /* keep */ Promise.all([promise]))[0]",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 43, EndLine: 1, EndColumn: 52,
					},
				},
			},
			// An unbraced control-flow body does not require an ASI semicolon
			{
				Code:     "if (ok) Promise.race([1])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 22, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "if (ok) (1)"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "if (ok) Promise.resolve(1)"},
						},
					},
				},
			},
			// A previous explicit semicolon prevents an ASI hazard
			{
				Code:     "foo();\nPromise.race([getPromise()])",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 2, Column: 14, EndLine: 2, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "foo();\n(getPromise())"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "foo();\nPromise.resolve(getPromise())"},
						},
					},
				},
			},
			// UTF-16 and multiline diagnostic range
			{
				Code:     "const text = \"😀\"; await Promise.any([\n  promise\n])",
				FileName: "case.mjs",
				Output:   []string{"const text = \"😀\"; await promise"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 38, EndLine: 3, EndColumn: 2,
					},
				},
			},

			// A TypeScript instantiation retains its type arguments when suggested
			{
				Code:     "const result = Promise.any([make<Value>,]);",
				FileName: "case.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 28, EndLine: 1, EndColumn: 42,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "const result = (make<Value>);"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "const result = Promise.resolve(make<Value>);"},
						},
					},
				},
			},

			// An awaited TypeScript instantiation has primary precedence
			{
				Code:     "const result = await Promise.race([make<Value>]);",
				FileName: "case.ts",
				Output:   []string{"const result = await make<Value>;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 35, EndLine: 1, EndColumn: 48,
					},
				},
			},

			// JSX text commas are not array separators
			{
				Code:     "const result = Promise.any([<div>,</div>,]);",
				FileName: "case.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.any()` is unnecessary.",
						Line: 1, Column: 28, EndLine: 1, EndColumn: 43,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/unwrap", Output: "const result = (<div>,</div>);"},
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "const result = Promise.resolve(<div>,</div>);"},
						},
					},
				},
			},

			// Comments and a parenthesized sequence before the trailing comma
			{
				Code:     "Promise.race([(first, promise) /* before */, // after\n]);",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 2, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve((first, promise) /* before */ // after\n);"},
						},
					},
				},
			},

			// A comma inside a regular expression is not a trailing comma
			{
				Code:     "Promise.race([/[,]/ // after\n]);",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 2, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(/[,]/ // after\n);"},
						},
					},
				},
			},

			// Template tokens and comments before the trailing comma
			{
				Code:     "Promise.race([tag`template,${promise}` /* before */, // after\n]);",
				FileName: "case.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 14, EndLine: 2, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "no-single-promise-in-promise-methods/use-promise-resolve", Output: "Promise.resolve(tag`template,${promise}` /* before */ // after\n);"},
						},
					},
				},
			},

			// Yield expressions need parentheses when moved under await
			{
				Code:     "async function* load() { await Promise.race([yield promise]); }",
				FileName: "case.mjs",
				Output:   []string{"async function* load() { await (yield promise); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.race()` is unnecessary.",
						Line: 1, Column: 45, EndLine: 1, EndColumn: 60,
					},
				},
			},

			// A JavaScript JSDoc type does not block destructuring fixes
			{
				Code:     "/** @type {[unknown]} */ const [result] = await Promise.all([promise]);",
				FileName: "case.mjs",
				Output:   []string{"/** @type {[unknown]} */ const result = await promise;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 61, EndLine: 1, EndColumn: 70,
					},
				},
			},

			// TypeScript assertions on assignment targets do not unwrap
			{
				Code:     "[result as Value] = await Promise.all([promise]);",
				FileName: "case.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-single-promise-in-promise-methods/error", Message: "Wrapping single-element array with `Promise.all()` is unnecessary.",
						Line: 1, Column: 39, EndLine: 1, EndColumn: 48,
					},
				},
			},
		})
}
