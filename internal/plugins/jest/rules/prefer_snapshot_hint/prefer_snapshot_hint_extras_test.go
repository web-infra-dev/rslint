// Rule-specific branch lock-ins, real-user cases, and AST/provenance extras.
// The complete upstream suites live in prefer_snapshot_hint_upstream_test.go.
// N/A: private matcher names, JSX containers, fixes and suggestions.
package prefer_snapshot_hint

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferSnapshotHintExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferSnapshotHintRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: ESTree method values are function expressions ----
		{Code: "const checks = { first() { expect(value).toMatchSnapshot(); }, second() { expect(other).toMatchSnapshot(); } };", Options: []any{"multi"}},
		{Code: "class Checks { constructor() { expect(value).toMatchSnapshot(); } second() { expect(other).toMatchSnapshot(); } }", Options: []any{"multi"}},
		{Code: "const checks = { get value() { expect(value).toMatchSnapshot(); }, set value(next) { expect(next).toMatchSnapshot(); } };", Options: []any{"multi"}},

		// Locks in missingHint: throw one argument; match two arguments
		{Code: "expect(run).toThrowErrorMatchingSnapshot(hint); expect(value).toMatchSnapshot(properties, hint);", Options: []any{"always"}},
		// Locks in missingHint: string and static template
		{Code: "expect(value).toMatchSnapshot(\"\"); expect(value).toMatchSnapshot(`snapshot`);", Options: []any{"always"}},
		// Locks in missingHint: excess arguments with first string
		{Code: "expect(value).toMatchSnapshot(\"hint\", extra, more);", Options: []any{"always"}},
		// Locks in create: non-assertion, non-snapshot, dynamic accessor, shadow
		{Code: "other(value).toMatchSnapshot();", Options: []any{"always"}},
		{Code: "expect(value);", Options: []any{"always"}},
		{Code: "expect(value).toMatchInlineSnapshot(); expect(run).toThrowErrorMatchingInlineSnapshot();", Options: []any{"always"}},

		{Code: "expect(value).toMatchSnapshot;", Options: []any{"always"}},
		{Code: "expect(value)[matcher]();", Options: []any{"always"}},
		{Code: "expect(value)[`to${name}`]();", Options: []any{"always"}},
		{Code: "expect(value)[123]();", Options: []any{"always"}},
		{Code: "function helper(expect) { expect(value).toMatchSnapshot(); }", Options: []any{"always"}},
		// ---- Dimension 4: parenthesized hint ----
		{Code: "expect(value).toMatchSnapshot((\"hint\"));", Options: []any{"always"}},
		// ---- Dimension 4: TS assertion-chain wrappers remain parser boundaries ----
		{Code: "expect(value)!.toMatchSnapshot();", Options: []any{"always"}},
		{Code: "(expect(value) as Assertion).toMatchSnapshot();", Options: []any{"always"}},
		{Code: "(expect(value) satisfies Assertion).toMatchSnapshot();", Options: []any{"always"}},
		// Locks in multi: single snapshot
		{Code: "expect(value).toMatchSnapshot();"},
		// Locks in multi: separate function expressions
		{Code: "const first = function() { expect(value).toMatchSnapshot(); }; const second = () => { expect(other).toMatchSnapshot(); };", Options: []any{"multi"}},
		// ---- Real-user: jest-community/eslint-plugin-jest#1068, primitive and array hints ----
		{Code: "test(\"compiler errors\", async () => { const errors = await compile(); expect(errors.length).toMatchSnapshot(\"error count\"); expect(errors).toMatchSnapshot(\"errors\"); });", Options: []any{"multi"}},
		// ---- Real-user: jest-community/eslint-plugin-jest#1074, sibling test groups ----
		{Code: "describe(\"compiler\", () => { it(\"count\", () => { expect(count).toMatchSnapshot(); }); it(\"details\", () => { expect(errors).toMatchSnapshot(\"errors\"); expect(warnings).toMatchSnapshot(\"warnings\"); }); });", Options: []any{"multi"}},
		// Locks in registration: no callback keeps enter/exit balanced
		{Code: "test.todo(\"later\"); test(\"single\", () => { expect(value).toMatchSnapshot(); });", Options: []any{"multi"}},
		// Renamed registration resets scope
		{Code: "import { test as check } from \"@jest/globals\"; describe(\"suite\", () => { check(\"one\", () => { expect(value).toMatchSnapshot(); }); check(\"two\", () => { expect(value).toMatchSnapshot(); }); });", Options: []any{"multi"}},
		// ---- Real-user: named callbacks registered as distinct tests keep distinct groups ----
		{Code: "const register = () => { const first = () => expect('first').toMatchSnapshot(); test('first', first); const second = () => expect('second').toMatchSnapshot(); test('second', second); }; describe('suite', register);", Options: []any{"multi"}},
		{Code: "const callback = () => expect('test').toMatchSnapshot(); test('case', callback); function helper() { expect('helper').toMatchSnapshot(); }", Options: []any{"multi"}},
		{Code: "test('case', callback); function callback() { expect('test').toMatchSnapshot(); } function helper() { expect('helper').toMatchSnapshot(); }", Options: []any{"multi"}},
	}, []rule_tester.InvalidTestCase{
		// The upstream parser treats this static-looking call as an expect matcher.
		{Code: "expect.toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingHint", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// Locks in missingHint: non-string single argument
		{Code: "expect(value).toMatchSnapshot({});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(hint);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(null);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(undefined);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(123);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(true);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(/hint/);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value).toMatchSnapshot(...hints);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		// Locks in missingHint: zero or excess arguments
		{Code: "expect(value).toThrowErrorMatchingSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 43},
		}},
		{Code: "expect(value).toThrowErrorMatchingSnapshot(\"hint\", true);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 43},
		}},
		{Code: "expect(value).toMatchSnapshot({}, \"hint\", 3);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		// ---- Dimension 4: parentheses, optional call, computed key, types, trivia, Unicode ----
		{Code: "(expect(value)).toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 17, EndLine: 1, EndColumn: 32},
		}},
		{Code: "expect?.(value)?.toMatchSnapshot?.();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 18, EndLine: 1, EndColumn: 33},
		}},
		{Code: "expect(value)[\"toMatchSnapshot\"]();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
		}},
		{Code: "expect(value)[`toMatchSnapshot`]();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
		}},
		{Code: "expect(value).toMatchSnapshot<Type>();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		{Code: "expect(value) /* comment */ .toMatchSnapshot(/* properties */ {});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 30, EndLine: 1, EndColumn: 45},
		}},
		{Code: "// 中文\nexpect(\"用户\").toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 14, EndLine: 2, EndColumn: 29},
		}},
		{Code: "expect(value).\n  toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 3, EndLine: 2, EndColumn: 18},
		}},
		// ---- Dimension 4: TS hint wrapper is not a string literal ----
		{Code: "expect(value).toMatchSnapshot(\"hint\" as const);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		// ---- Dimension 4: default mode and source-file flush ----
		{Code: "expect(a).toMatchSnapshot(); expect(b).toMatchSnapshot();", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 40, EndLine: 1, EndColumn: 55},
		}},
		// Locks in multi: nested helpers contribute to enclosing expression
		{Code: "const body = () => { const inner = function() { expect(value).toMatchSnapshot(); }; expect(other).toMatchSnapshot(); };", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 63, EndLine: 1, EndColumn: 78},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 99, EndLine: 1, EndColumn: 114},
		}},
		// Locks in multi: function declarations do not create expression boundaries
		{Code: "function first() { expect(value).toMatchSnapshot(); } function second() { expect(other).toMatchSnapshot(); }", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 34, EndLine: 1, EndColumn: 49},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 89, EndLine: 1, EndColumn: 104},
		}},
		// Registration isolation preserves the enclosing function's own group.
		{Code: "const helper = () => {\n  expect('before').toMatchSnapshot();\n  test('inner', () => {\n    expect('inner').toMatchSnapshot();\n  });\n  expect('after').toMatchSnapshot();\n};", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 20, EndLine: 2, EndColumn: 35},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 6, Column: 19, EndLine: 6, EndColumn: 34},
		}},
		// Locks in registration: parameterized tests
		{Code: "test.each([1, 2])(\"row\", value => { expect(value).toMatchSnapshot(); expect(value).toThrowErrorMatchingSnapshot(); });", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 51, EndLine: 1, EndColumn: 66},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 84, EndLine: 1, EndColumn: 112},
		}},
		// Locks in registration: tagged template
		{Code: "test.each`value\n${1}`(\"row\", value => { expect(value).toMatchSnapshot(); expect(value).toMatchSnapshot(); });", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 39, EndLine: 2, EndColumn: 54},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 72, EndLine: 2, EndColumn: 87},
		}},
		// Named and renamed expect imports
		{Code: "import { expect as check } from \"@jest/globals\"; check(value).toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 63, EndLine: 1, EndColumn: 78},
		}},
		// Locks in upstream: interpolated hint is not a static string
		{Code: "expect(value).toMatchSnapshot(`state ${name}`);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
		}},
		// Locks in upstream: not is still checked
		{Code: "expect(value).not.toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 19, EndLine: 1, EndColumn: 34},
		}},
	})
}

// Binding resolution must work without a TypeChecker as well as in the main suite.
func TestPreferSnapshotHintSourceOnly(t *testing.T) {
	for _, code := range []string{
		`expect(value).toMatchSnapshot();`,
		`import { expect as check } from "@jest/globals"; check(value).toMatchSnapshot();`,
		`expect(value).toMatchSnapshot(); function helper(expect) { expect(other).toMatchSnapshot(); }`,
	} {
		t.Run(code, func(t *testing.T) {
			runPreferSnapshotHintSourceOnly(t, code, []any{"always"}, 1)
		})
	}
	t.Run("registered named callbacks have separate groups", func(t *testing.T) {
		runPreferSnapshotHintSourceOnly(t, `
const register = () => {
  const first = () => expect('first').toMatchSnapshot();
  test('first', first);
  const second = () => expect('second').toMatchSnapshot();
  test('second', second);
};
describe('suite', register);`, []any{"multi"}, 0)
	})
	t.Run("hoisted registered callback has its own group", func(t *testing.T) {
		runPreferSnapshotHintSourceOnly(t, `
test('case', callback);
function callback() { expect('test').toMatchSnapshot(); }
function helper() { expect('helper').toMatchSnapshot(); }`, []any{"multi"}, 0)
	})
}

func runPreferSnapshotHintSourceOnly(t *testing.T, code string, options []any, want int) {
	t.Helper()
	root := fixtures.GetRootDir()
	name := tspath.ResolvePath(root.Dir, "snapshot-hint-source-only.ts")
	host := utils.CreateCompilerHost(root.Dir, utils.NewOverlayVFS(root.FS, map[string]string{name: code}))
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{RootFileNames: []string{name}, Host: host, CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext}, SingleThreaded: true})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only program")
	}
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs: []*lintprogram.Program{program}, TargetsByProgram: [][]string{{name}}, SingleThreaded: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: PreferSnapshotHintRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return PreferSnapshotHintRule.Run(ctx, options) }}}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	_, err = linter.RunLinter(linter.RunLinterOptions{LintPlan: plan, SingleThreaded: true, Consumer: rule.DiagnosticConsumer{Report: func(d rule.RuleDiagnostic) {
		count++
		if got := code[d.Range.Pos():d.Range.End()]; got != "toMatchSnapshot" {
			t.Errorf("unexpected diagnostic range: %q", got)
		}
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("got %d diagnostics, want %d", count, want)
	}
}
