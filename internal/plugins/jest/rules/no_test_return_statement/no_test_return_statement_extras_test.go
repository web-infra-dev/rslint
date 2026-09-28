package no_test_return_statement_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_test_return_statement"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestNoTestReturnStatementExtras locks in branches and edge shapes the
// upstream suite does not exercise. The upstream-mirrored cases live in
// no_test_return_statement_upstream_test.go.
func TestNoTestReturnStatementExtras(t *testing.T) {
	runNoTestReturnStatement(
		t,
		[]rule_tester.ValidTestCase{
			// Only statements directly in the callback body are checked; an early
			// guard nested in another statement ends the test deliberately.
			{Code: `test('guarded', () => {
  if (!supported) return;
  expect(1).toBe(1);
});`},
			{Code: `test('loop', () => {
  for (const item of items) { return; }
});`},
			// A nested function's return belongs to that function.
			{Code: `test('nested', () => {
  const load = () => { return 1; };
  expect(load()).toBe(1);
});`},
			// Hooks and suites are not tests.
			{Code: `beforeEach(() => { return setup(); });`},
			{Code: `describe('suite', () => { return; });`},
			// A named callback that is also called directly may need its return
			// value there, so it is not reported.
			{Code: `it('one', myTest);
function myTest() { return expect(1).toBe(1); }
const value = myTest();`},
			// Exported callbacks may be used as helpers elsewhere.
			{Code: `export function myTest() { return expect(1).toBe(1); }
it('one', myTest);`},
			{Code: `function myTest() { return expect(1).toBe(1); }
it('one', myTest);
export { myTest };`},
			{Code: `export const myTest = () => { return expect(1).toBe(1); };
it('one', myTest);`},
			// A reassigned binding may hold a different function by the time the
			// test runs.
			{Code: `let myTest = () => { return 1; };
myTest = () => {};
it('one', myTest);`},
			// A parameter shadowing the function name is a different binding.
			{Code: `function myTest() { return 1; }
function register(myTest) { it('one', myTest); }`},
			// A local function named `test` is not Jest's test.
			{Code: `const test = (name, fn) => fn();
test('local', () => { return 1; });`},
			// Upstream only accepts a function literal as the callback, so a
			// TypeScript assertion around it is not looked through.
			{Code: `test('asserted', (() => { return 1; }) as () => number);`},
			// A named callback that resolves to no local function is skipped.
			{Code: `import { myTest } from './helpers';
it('one', myTest);`},
		},
		[]rule_tester.InvalidTestCase{
			// A bare return is still a return statement.
			{
				Code:   `test('bare', () => { return; });`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(1, 22, 29)},
			},
			// async callbacks are checked the same way.
			{
				Code:   `test('async', async () => { return load(); });`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(1, 29, 43)},
			},
			// Only the first direct return is reported.
			{
				Code: `test('twice', () => {
  return 1;
  return 2;
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 12)},
			},
			// Parentheses around the callback are looked through.
			{
				Code:   `test('parens', ((() => { return 1; })));`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(1, 26, 35)},
			},
			// A const-bound function is resolved like a function declaration.
			{
				Code: `const myTest = () => { return expect(1).toBe(1); };
it('one', myTest);`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(1, 24, 49)},
			},
			// A callback shared by several tests is reported once, including when
			// it is declared after its uses.
			{
				Code: `it('one', myTest);
it.each([1])('two', myTest);
function myTest() {
  return expect(1).toBe(1);
}`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(4, 3, 28)},
			},
			// Renamed imports from @jest/globals are recognized.
			{
				Code: `import { test as check } from '@jest/globals';
check('renamed', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 26, 35)},
			},
			// Non-ASCII text before the return keeps UTF-16 columns.
			{
				Code:   `test('日本', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(1, 20, 29)},
			},
		},
	)
}

// BenchmarkNoTestReturnStatementSharedCallback registers one named callback
// many times. Every registration resolves the same binding, and the callback's
// ownership check walks every reference to it, so the time per registration
// must stay flat as registrations grow.
func BenchmarkNoTestReturnStatementSharedCallback(b *testing.B) {
	root := fixtures.GetRootDir()
	for _, registrations := range []int{2000, 8000} {
		b.Run(fmt.Sprintf("registrations=%d", registrations), func(b *testing.B) {
			var code strings.Builder
			code.WriteString("function callback() { return 1; }\n")
			for index := range registrations {
				fmt.Fprintf(&code, "test('case %d', callback);\n", index)
			}
			fileName := tspath.ResolvePath(root.Dir, fmt.Sprintf("no-test-return-statement-bench-%d.ts", registrations))
			fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code.String()})
			program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
				RootFileNames:   []string{fileName},
				Host:            utils.CreateCompilerHost(root.Dir, fs),
				CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
				SingleThreaded:  true,
			})
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				count := 0
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: program, File: fileName,
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     no_test_return_statement.NoTestReturnStatementRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_test_return_statement.NoTestReturnStatementRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) { count++ }},
				})
				if count != 1 {
					b.Fatalf("got %d diagnostics, want 1", count)
				}
			}
		})
	}
}
