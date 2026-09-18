package no_unneeded_async_expect_function

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnneededAsyncExpectFunctionExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnneededAsyncExpectFunctionRule,
		[]rule_tester.ValidTestCase{
			{Code: `const expect = createAssertionLibrary();
expect(async () => { await run(); }).resolves.toBe(1);`},
			{Code: `import { expect } from 'vitest';
expect(async () => { await run(); }).resolves.toBe(1);`},
			{Code: `import { expect } from '@jest/globals';
expect(async () => { await run(); }).resolves.toBe(1);`},
			{Code: `import { expect } from '@rstest/core';
function verify(expect: any) { expect(async () => { await run(); }).resolves.toBe(1); }`},
			{Code: `expect(async () => { await run(); }).toThrow();`},
			// Rstest invokes a function passed to .rejects. The wrapper can be
			// necessary to turn a synchronous throw into a rejected Promise.
			{Code: `declare function failsBeforeReturningPromise(): Promise<never>;
expect(async () => { await failsBeforeReturningPromise(); }).rejects.toThrow();`},
			{Code: `function throwsSynchronously(): Promise<never> {
  throw new Error('sync failure');
}
expect(async () => { await throwsSynchronously(); }).rejects.toThrow('sync failure');`},
			{Code: `expect.poll(async () => { await run(); }).resolves.toBe(1);`},
			{Code: `expect.element(async () => { await run(); }).rejects.toThrow();`},
			{Code: `expect(async function named() { await named(); }).rejects.toThrow();`},
			{Code: `expect(async function* () { await run(); }).rejects.toThrow();`},
			{Code: `expect(async function* () { await run(); }).resolves.toBe(1);`},
			{Code: `expect(async (value) => { await run(value); }).rejects.toThrow();`},
			{Code: `expect(async <Value>() => { await run<Value>(); }).rejects.toThrow();`},
			{Code: `expect(async function () { await this.run(); }).rejects.toThrow();`},
			{Code: `expect(async function () { await run(arguments); }).rejects.toThrow();`},
			{Code: `expect(async function () { await run(new.target); }).rejects.toThrow();`},
			{Code: `expect(async () => { return await run(); }).rejects.toThrow();`},
			{Code: `expect(async () => { await promise; }).rejects.toThrow();`},
			{Code: `expect(async () => { await run(); another(); }).rejects.toThrow();`},
			{Code: `expect(async () => [await run()]).rejects.toThrow();`},
			{Code: `expect(async () => { await run(); })[modifier].toThrow();`},
			{Code: `expect(async () => { await run(); }).resolves.rejects.toThrow();`},
		},
		[]rule_tester.InvalidTestCase{
			extrasInvalidNoFix(
				`function returnsPlainValue() { return 1; }
expect(async () => { await returnsPlainValue(); }).resolves.toBe(1);`,
				2, 8, 50,
			),
			extrasInvalidNoFix(
				`function throwsSynchronously(): never { throw new Error('sync failure'); }
expect(async () => { await throwsSynchronously(); }).resolves.toBe(1);`,
				2, 8, 52,
			),
			extrasInvalidNoFix(
				`expect((async () => { await run(); }) as any).resolves.toBe(1);`,
				1, 8, 45,
			),
			extrasInvalidNoFix(
				`expect(async () => { await run(); }).resolves.toBe(1);`,
				1, 8, 36,
			),
			extrasInvalidNoFix(
				`expect.soft(async () => { await run(); }).resolves.toBe(1);`,
				1, 13, 41,
			),
			extrasInvalidNoFix(
				`expect(async () => { await run(); }).resolves.to.equal(1);`,
				1, 8, 36,
			),
			extrasInvalidNoFix(
				`import { expect as check } from '@rstest/core';
check(async () => { await run(); }).resolves.toBe(1);`,
				2, 7, 35,
			),
			extrasInvalidNoFix(
				`import * as rstest from '@rstest/core';
rstest.expect(async () => { await run(); }).resolves.toBe(1);`,
				2, 15, 43,
			),
			extrasInvalidNoFix(
				`const { expect: check } = require('rstack/test');
check(async () => { await run(); }).resolves.toBe(1);`,
				2, 7, 35,
			),
			extrasInvalidNoFix(
				`const rstest = require('@rstest/core');
rstest.expect(async () => { await run(); }).resolves.toBe(1);`,
				2, 15, 43,
			),
			extrasInvalidNoFix(
				`import { expect as check } from '@rstest/playwright';
check(async () => { await run(); }).resolves.toBe(1);`,
				2, 7, 35,
			),
			extrasInvalidNoFix(
				`import.meta.rstest.expect(async () => { await run(); }).resolves.toBe(1);`,
				1, 27, 55,
			),
			extrasInvalidNoFix(
				`test('context', ({ expect }) => {
  expect(async () => { await run(); }).resolves.toBe(1);
});`,
				2, 10, 38,
			),
			extrasInvalidNoFix(
				`test('context', ctx => {
  ctx.expect(async () => { await run(); }).resolves.toBe(1);
});`,
				2, 14, 42,
			),
			extrasInvalidNoFix(
				`expect((async () => { await run(); })).resolves.toBe(1);`,
				1, 8, 38,
			),
			extrasInvalidNoFix(
				`expect(async () => await (run())).resolves.toBe(1);`,
				1, 8, 33,
			),
			extrasInvalidNoFix(
				`expect?.(async () => { await run(); }).resolves.toBe(1);`,
				1, 10, 38,
			),
			extrasInvalidNoFix(
				`expect(async () => { await run(); })['resolves'].toBe(1);`,
				1, 8, 36,
			),
			extrasInvalidNoFix(
				`expect(async function () { await client.arguments(); }).resolves.toBe(1);`,
				1, 8, 55,
			),
			extrasInvalidNoFix(
				`expect(async () => { /* preserve why */ await run(); }).resolves.toBe(1);`,
				1, 8, 55,
			),
			extrasInvalidNoFix(
				`expect(async () => { await run(); }, makeMessage()).resolves.toBe(1);`,
				1, 8, 36,
			),
		},
	)
}

func extrasInvalidNoFix(code string, line, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "noAsyncWrapperForExpectedPromise",
			Message:   "Avoid wrapping asynchronous expectations in an unnecessary async function.",
			Line:      line,
			Column:    column,
			EndLine:   line,
			EndColumn: endColumn,
		}},
	}
}

func TestNoUnneededAsyncExpectFunctionResolvesInSourceOnlyProgram(t *testing.T) {
	if NoUnneededAsyncExpectFunctionRule.RequiresTypeInfo {
		t.Fatal("rstest/no-unneeded-async-expect-function must not require type information")
	}
	code := `
expect(async () => { await globalCall(); }).resolves.toBe(1);
import { expect as check, test as rstestTest } from '@rstest/core';
import * as rstest from '@rstest/playwright';
const { expect: requiredExpect } = require('rstack/test');
check(async () => { await importedCall(); }).resolves.toBe(1);
rstest.expect(async () => { await browserCall(); }).resolves.toBe(1);
requiredExpect(async () => { await requiredCall(); }).resolves.toBe(1);
rstestTest('context', ({ expect: localExpect }) => {
  localExpect(async () => { await contextCall(); }).resolves.toBe(1);
});
import { expect as foreignExpect } from 'vitest';
foreignExpect(async () => { await foreignCall(); }).resolves.toBe(1);
function local(expect: any) { expect(async () => { await localCall(); }).resolves.toBe(1); }
`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "no-unneeded-async-expect-function-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName},
		Host:          host,
		CompilerOptions: &core.CompilerOptions{
			Module: core.ModuleKindESNext,
		},
		SingleThreaded: true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program with no TypeChecker")
	}

	var positions []int
	runNoUnneededAsyncExpectFunction(t, sourceProgram, fileName, rule.DiagnosticConsumer{
		Demand: rule.EditDemandNone,
		Report: func(diagnostic rule.RuleDiagnostic) {
			positions = append(positions, diagnostic.Range.Pos())
		},
	})
	sort.Ints(positions)
	if len(positions) != 5 {
		t.Fatalf("reported %d assertions, want 5 at %v", len(positions), positions)
	}
	wantPositions := make([]int, 0, 5)
	for _, marker := range []string{
		"async () => { await globalCall(); }",
		"async () => { await importedCall(); }",
		"async () => { await browserCall(); }",
		"async () => { await requiredCall(); }",
		"async () => { await contextCall(); }",
	} {
		position := strings.Index(code, marker)
		if position < 0 {
			t.Fatalf("missing source marker %q", marker)
		}
		wantPositions = append(wantPositions, position)
	}
	sort.Ints(wantPositions)
	if !reflect.DeepEqual(positions, wantPositions) {
		t.Fatalf("diagnostic positions = %v, want %v", positions, wantPositions)
	}
}

func TestNoUnneededAsyncExpectFunctionEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(async () => { await run(); }).resolves.toBe(1);`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		runNoUnneededAsyncExpectFunction(t, lintprogram.NewFromCompiler(program), sourceFile.FileName(), rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	want := withoutEdits(allEdits)
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone: diagnosticsOnly, rule.EditDemandAutofix: autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d diagnostic changed:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.FixesPtr != nil || autofixOnly.FixesPtr != nil ||
		suggestionOnly.FixesPtr != nil || allEdits.FixesPtr != nil {
		t.Fatal("rule unexpectedly materialized autofixes")
	}
	for _, diagnostic := range []rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		if diagnostic.Suggestions != nil {
			t.Fatal("rule unexpectedly materialized suggestions")
		}
	}
}

func runNoUnneededAsyncExpectFunction(
	t *testing.T,
	program *lintprogram.Program,
	fileName string,
	consumer rule.DiagnosticConsumer,
) {
	t.Helper()
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{program},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     NoUnneededAsyncExpectFunctionRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUnneededAsyncExpectFunctionRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer:       consumer,
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
}
