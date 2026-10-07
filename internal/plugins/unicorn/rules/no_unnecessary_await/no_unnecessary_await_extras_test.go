package no_unnecessary_await_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_await"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Upstream parses every snippet as a module, including JSX with top-level await.
func moduleRoot() rule_tester.Root {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.await.json"): `{"extends":"./tsconfig.json","compilerOptions":{"moduleDetection":"force"}}`,
	})
	return root
}

func TestNoUnnecessaryAwaitEditDemand(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ code, output string }{
		{"async function f() { return await // keep\n1; }", "async function f() { return ( // keep\n1); }"},
		{"foo()\nawait [];", "foo()\n;[];"},
		{"async function f() { await 1; run(); }", ""},
	} {
		t.Run(test.code, func(t *testing.T) {
			t.Parallel()
			helper := rule_tester.NewProgramHelper(moduleRoot())
			program, source, err := helper.CreateTestProgram(test.code, "input.mts", "tsconfig.await.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program),
					File:    source.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: no_unnecessary_await.NoUnnecessaryAwaitRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_unnecessary_await.NoUnnecessaryAwaitRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
						got = append(got, diagnostic)
					}},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			base := diagnostics[rule.EditDemandNone]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != base.Range || !reflect.DeepEqual(diagnostic.Message, base.Message) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := test.output != "" && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix || diagnostic.Suggestions != nil {
					t.Errorf("demand %d: unexpected edit artifacts", demand)
				}
			}
			if !reflect.DeepEqual(diagnostics[rule.EditDemandAutofix].FixesPtr, diagnostics[rule.EditDemandAll].FixesPtr) {
				t.Fatal("autofix and all demand produced different edits")
			}
			if test.output != "" {
				output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{diagnostics[rule.EditDemandAll]})
				if !fixed || output != test.output {
					t.Fatalf("got %q, want %q", output, test.output)
				}
			}
		})
	}
}

func TestNoUnnecessaryAwaitExtras(t *testing.T) {
	rule_tester.RunRuleTester(moduleRoot(), "tsconfig.await.json", t, &no_unnecessary_await.NoUnnecessaryAwaitRule,
		[]rule_tester.ValidTestCase{
			// Expression kinds and TypeScript boundaries
			{Code: "await undefined; await NaN; await Infinity;", FileName: "input.ts"},
			{Code: "await (value = 1); await (value += 1); await (value &&= 1);", FileName: "input.ts"},
			{Code: "await (a && b); await (a || b); await (a ?? b);", FileName: "input.ts"},
			{Code: "await (condition ? 1 : 2);", FileName: "input.ts"},
			{Code: "await obj.then; await obj[\"then\"]; await obj?.then; await obj?.();", FileName: "input.ts"},
			{Code: "class C { #promise; async f() { return await this.#promise; } }", FileName: "input.ts"},
			{Code: "await (0, (1 as number));", FileName: "input.ts"},
			{Code: "await (0, (value as Promise<number>));", FileName: "input.ts"},
			{Code: "await (promise satisfies Promise<number>);", FileName: "input.ts"},
			{Code: "await (<Promise<number>>promise);", FileName: "input.ts"},
			{Code: "await (0, 1!);", FileName: "input.ts"},
			{Code: "await ((() => 1)<number>);", FileName: "input.ts"},
		}, []rule_tester.InvalidTestCase{
			// Expression kinds and TypeScript boundaries
			// Parentheses avoid the documented JS parser limitation for `await !`.
			{Code: "await (!Promise.resolve());", FileName: "input.mjs",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
				Output: []string{"(!Promise.resolve());"},
			},
			{Code: "await /x/;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"/x/;"},
			},
			{Code: "await `value`;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"`value`;"},
			},
			{Code: "await (a + b);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(a + b);"},
			},
			{Code: "await (key in object);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(key in object);"},
			},
			{Code: "await typeof value;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"typeof value;"},
			},
			{Code: "await delete object.key;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"delete object.key;"},
			},
			{Code: "await (promise, (1, 2));", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(promise, (1, 2));"},
			},
			{Code: "await (promise, function () {});", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(promise, function () {});"},
			},
			{Code: "await ((1 as number)!);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"((1 as number)!);"},
			},
			{Code: "await (<number>1);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(<number>1);"},
			},
			{Code: "await (<div promise={promise} />);", FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(<div promise={promise} />);"},
			},
			{Code: "class C { #p; async f(other) { return await (#p in other); } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 39, EndLine: 1, EndColumn: 44},
				},
				Output: []string{"class C { #p; async f(other) { return (#p in other); } }"},
			},
			{Code: "await (function () {} as unknown);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(function () {} as unknown);"},
			},
			{Code: "await (class {} satisfies Object);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"(class {} satisfies Object);"},
			},
			{Code: "async function f() { return (await 1) as number; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
			},
			{Code: "async function f() { return (await 1)!; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
			},
			// Tail evaluation boundaries
			{Code: "async function f() { if (q) await []; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { if (q) []; }"},
			},
			{Code: "async function f() { if (q) run(); else await []; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 41, EndLine: 1, EndColumn: 46},
				},
				Output: []string{"async function f() { if (q) run(); else []; }"},
			},
			{Code: "async function f() { if (await true) run(); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 26, EndLine: 1, EndColumn: 31},
				},
			},
			{Code: "async function f() { { await 1; } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 24, EndLine: 1, EndColumn: 29},
				},
				Output: []string{"async function f() { { 1; } }"},
			},
			{Code: "async function f() { if (q) { await 1; } run(); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36},
				},
			},
			{Code: "async function f() { const x = await 1, y = 2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37},
				},
			},
			{Code: "async function f() { const x = 1, y = await 2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 39, EndLine: 1, EndColumn: 44},
				},
				Output: []string{"async function f() { const x = 1, y = 2; }"},
			},
			{Code: "async function f() { const x = await 1; run(); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37},
				},
			},
			{Code: "async function f() { while (q) { await 1; } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 34, EndLine: 1, EndColumn: 39},
				},
			},
			{Code: "async function f() { switch (q) { case 0: await 1; } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 43, EndLine: 1, EndColumn: 48},
				},
			},
			{Code: "async function f() { try { run(); } catch { await 1; } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 45, EndLine: 1, EndColumn: 50},
				},
			},
			{Code: "async function f() { label: { await 1; } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36},
				},
			},
			{Code: "async function f() { return [await 1]; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
			},
			{Code: "async function f() { value = await 1; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
			},
			{Code: "async function f() { return await 1; unreachable(); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
			},
			{Code: "export const value = await 1;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
				},
			},
			{Code: "export default await 1;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21},
				},
			},
			{Code: "export default async () => await 1;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33},
				},
				Output: []string{"export default async () => 1;"},
			},
			{Code: "class C { async f() { return await 1; } }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
				Output: []string{"class C { async f() { return 1; } }"},
			},
			{Code: "const object = { async f() { return await 1; } };", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 37, EndLine: 1, EndColumn: 42},
				},
				Output: []string{"const object = { async f() { return 1; } };"},
			},
			{Code: "const f = async function () { return await 1; }; run();", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 38, EndLine: 1, EndColumn: 43},
				},
				Output: []string{"const f = async function () { return 1; }; run();"},
			},
			{Code: "async function* f() { return await 1; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
				Output: []string{"async function* f() { return 1; }"},
			},
			{Code: "const f = async () => (await 1);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 24, EndLine: 1, EndColumn: 29},
				},
				Output: []string{"const f = async () => (1);"},
			},
			// Fix spacing, comments, parentheses and ASI
			{Code: "async function f() { return await\n1; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return ( 1); }"},
			},
			{Code: "async function f() { throw await\n1 }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33},
				},
				Output: []string{"async function f() { throw ( 1) }"},
			},
			{Code: "async function f() { return await (\n1); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return ( (\n1)); }"},
			},
			{Code: "async function f() { return (await\n// keep\n1); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
				},
				Output: []string{"async function f() { return (// keep\n1); }"},
			},
			{Code: "async function f() { return await /* keep */\n1 /* end */; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return ( /* keep */\n1 /* end */); }"},
			},
			{Code: "async function f() { return await // keep\n(1); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return ( // keep\n(1)); }"},
			},
			{Code: "await  1;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"1;"},
			},
			{Code: "\"😀\"; await 1;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 7, EndLine: 1, EndColumn: 12},
				},
				Output: []string{"\"😀\"; 1;"},
			},
			{Code: "foo()\nawait /x/;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"foo()\n;/x/;"},
			},
			{Code: "const x = {}\nawait `x`;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"const x = {}\n;`x`;"},
			},
			{Code: "foo()\nawait /* keep */ [];", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"foo()\n/* keep */ ;[];"},
			},
			{Code: "function f() {}\nawait [];", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"function f() {}\n[];"},
			},
			{Code: "class C {}\nawait [];", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"class C {}\n[];"},
			},
			{Code: "await ((await promise));", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
				Output: []string{"((await promise));"},
			},
			{Code: "await await 1;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 7, EndLine: 1, EndColumn: 12},
				},
				Output: []string{"await 1;", "1;"},
			},
			{Code: "async function f() { return await /** @type {number} */ (1); }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return /** @type {number} */ (1); }"},
			},
			{Code: "async function f() { return /** @type {number} */ (await 1); }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 52, EndLine: 1, EndColumn: 57},
				},
				Output: []string{"async function f() { return /** @type {number} */ (1); }"},
			},
			{Code: "await /** @type {Function} */ (function () {});", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// JSDoc casts, line endings, binding patterns and token boundaries.
			{Code: "async function f() { return await !Promise.resolve(); }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return !Promise.resolve(); }"},
			},
			{Code: "async function f() { return /** @type {number} */ (await // keep\n1); }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 52, EndLine: 1, EndColumn: 57},
				},
				Output: []string{"async function f() { return /** @type {number} */ (// keep\n1); }"},
			},
			{Code: "async function f() { return /** @satisfies {number} */ (await // keep\n1); }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 57, EndLine: 1, EndColumn: 62},
				},
				Output: []string{"async function f() { return /** @satisfies {number} */ (// keep\n1); }"},
			},
			{Code: "foo()\nawait /** @type {number} */ (.5);", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"foo()\n/** @type {number} */ ;(.5);"},
			},
			{Code: "async function f() { throw await /* keep */\r\n1 /* end */; }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33},
				},
				Output: []string{"async function f() { throw ( /* keep */\r\n1 /* end */); }"},
			},
			{Code: "async function f() { return await 1; }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
				Output: []string{"async function f() { return ( 1); }"},
			},
			{Code: "async function f() { const { x = await 1 } = value; }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 34, EndLine: 1, EndColumn: 39},
				},
			},
			{Code: "async function f() { const { [await 1]: x } = value; }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36},
				},
			},
			{Code: "async function f() { const [x] = await []; }", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 34, EndLine: 1, EndColumn: 39},
				},
				Output: []string{"async function f() { const [x] = []; }"},
			},
			{Code: "await using x = await [];", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22},
				},
				Output: []string{"await using x = [];"},
			},
			{Code: "foo()\nawait .5;", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"foo()\n;.5;"},
			},
			{Code: "foo()\nawait \"[a]\";", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-unnecessary-await", Message: "Do not `await` non-promise value.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
				Output: []string{"foo()\n\"[a]\";"},
			},
		})
}
