package prefer_to_have_been_called

import (
	"reflect"
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

func TestPreferToHaveBeenCalledExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	for _, code := range []string{
		`expect(fn).toHaveBeenCalledTimes;`, `expect(fn).toHaveBeenCalledTimes();`,
		`expect(fn).toHaveBeenCalledTimes()(0);`, `expect.toHaveBeenCalledTimes(0);`,
		`expect(fn).toHaveReturnedTimes(0);`, `expect(fn).callCount(0);`,
		`expect(fn).not.toHaveBeenCalled();`, `expect(fn).not.not.toHaveBeenCalledTimes(0);`,
		`expect(fn).resolves.rejects.toHaveBeenCalledTimes(0);`,
		`expect.element(locator).toHaveBeenCalledTimes(0);`,
		`expect(fn)[toHaveBeenCalledTimes](0);`, `expect(fn)[42](0);`,
		`expect(fn)["toHaveBeen" + "CalledTimes"](0);`,
		`expect(fn)!.toHaveBeenCalledTimes(0);`, `(expect(fn) as any).toHaveBeenCalledTimes(0);`,
		`(expect(fn) satisfies Assertion).toHaveBeenCalledTimes(0);`,
		`import { expect } from 'vitest'; expect(fn).toHaveBeenCalledTimes(0);`,
		`import { expect } from '@jest/globals'; expect(fn).toHaveBeenCalledTimes(0);`,
		`import { expect } from '@playwright/test'; expect(fn).toHaveBeenCalledTimes(0);`,
		`import type { expect as check } from '@rstest/core'; check(fn).toHaveBeenCalledTimes(0);`,
		`const expect = custom; expect(fn).toHaveBeenCalledTimes(0);`,
		`import { expect } from '@rstest/core'; function run(expect: any) { expect(fn).toHaveBeenCalledTimes(0); }`,
		`let core = require('@rstest/core'); core = other; core.expect(fn).toHaveBeenCalledTimes(0);`,
		`test.each([1])('row', (row, { expect }) => expect(fn).toHaveBeenCalledTimes(0));`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	for _, count := range []string{"1", "-0", "+0", "0n", "'0'", "false", "null", "count", "...counts", "0 satisfies number", "0!", "1 - 1"} {
		valid = append(valid, rule_tester.ValidTestCase{Code: "expect(fn).toHaveBeenCalledTimes(" + count + ");"})
	}
	var invalid []rule_tester.InvalidTestCase
	add := func(code, output string) {
		t.Helper()
		start := strings.Index(code, "toHaveBeenCalledTimes")
		name := "toHaveBeenCalledTimes"
		if start == -1 {
			name = "toBeCalledTimes"
			start = strings.Index(code, name)
		}
		line := strings.Count(code[:start], "\n") + 1
		column := start - strings.LastIndex(code[:start], "\n")
		if start > 0 && (code[start-1] == '\'' || code[start-1] == '"' || code[start-1] == '`') {
			column--
			name += "xx"
		}
		item := rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: line, Column: column, EndLine: line, EndColumn: column + len(name)}}}
		if output != "" {
			item.Output = []string{output}
		}
		invalid = append(invalid, item)
	}
	// Dimension 4: accessors, optional boundaries, wrappers, and numeric spellings.
	for _, pair := range [][2]string{
		{`expect(fn)['toBeCalledTimes'](0);`, `expect(fn).not['toHaveBeenCalled']();`},
		{"expect(fn)[`toHaveBeenCalledTimes`](0);", "expect(fn).not[`toHaveBeenCalled`]();"},
		{`expect(fn)[("toHaveBeenCalledTimes")](0);`, `expect(fn).not[("toHaveBeenCalled")]();`},
		{`expect(fn)?.toHaveBeenCalledTimes(0);`, `expect(fn)?.not.toHaveBeenCalled();`},
		{`expect(fn)?.['toHaveBeenCalledTimes'](0);`, `expect(fn)?.not['toHaveBeenCalled']();`},
		{`expect(fn).toHaveBeenCalledTimes?.(0);`, `expect(fn).not.toHaveBeenCalled?.();`},
		{`expect(fn)?.not.toHaveBeenCalledTimes(0);`, `expect(fn)?.toHaveBeenCalled();`},
		{`expect(fn)['not']?.['toHaveBeenCalledTimes'](0);`, `expect(fn)?.['toHaveBeenCalled']();`},
		{`((expect(fn))).toHaveBeenCalledTimes(((0)));`, `((expect(fn))).not.toHaveBeenCalled();`},
		{`(expect(fn).toHaveBeenCalledTimes)(0);`, `(expect(fn).not.toHaveBeenCalled)();`},
		{`expect(fn).toHaveBeenCalledTimes<number>(0,);`, `expect(fn).not.toHaveBeenCalled();`},
		{`expect(fn).not.toHaveBeenCalledTimes?.<number>(0);`, `expect(fn).toHaveBeenCalled?.();`},
		{`expect(fn). /* keep */ toHaveBeenCalledTimes(0);`, `expect(fn).not. /* keep */ toHaveBeenCalled();`},
		{`expect(fn)./* keep */not.toHaveBeenCalledTimes(0);`, `expect(fn)/* keep */.toHaveBeenCalled();`},
		{`expect(fn).toHaveBeenCalledTimes /* keep */ (0);`, `expect(fn).not.toHaveBeenCalled /* keep */ ();`},
	} {
		add(pair[0], pair[1])
	}
	for _, count := range []string{"0.0", "0x0", "0b0", "0o0", "0e2", "0.0_0", "0 as const", "<number>0"} {
		add("expect(fn).toHaveBeenCalledTimes("+count+");", "expect(fn).not.toHaveBeenCalled();")
	}
	for _, factory := range []string{"expect", "expect.soft"} {
		add(factory+`(fn).toHaveBeenCalledTimes(0);`, factory+`(fn).not.toHaveBeenCalled();`)
	}
	for _, pair := range [][2]string{
		{`import { expect as check } from '@rstest/core';`, "check"},
		{`import { expect } from 'rstack/test';`, "expect"},
		{`import * as core from '@rstest/core';`, "core.expect"},
		{`const { expect: check } = require('@rstest/core');`, "check"},
		{`const core = require('rstack/test');`, "core.expect"},
		{`import { expect } from '@rstest/playwright';`, "expect"},
		{`const { expect: check } = import.meta.rstest;`, "check"},
		{"", "import.meta.rstest.expect"},
	} {
		add(pair[0]+"\n"+pair[1]+`(fn).toHaveBeenCalledTimes(0);`, pair[0]+"\n"+pair[1]+`(fn).not.toHaveBeenCalled();`)
	}
	// Real-user: local concurrent expect and parameterized fixture contexts.
	for _, code := range []string{
		`test.concurrent('does not notify', ({ expect }) => expect(listener).toHaveBeenCalledTimes(0));`,
		`test.for([1])('does not retry', (row, ctx) => ctx.expect(retry).toHaveBeenCalledTimes(0));`,
		"test.each`value\n${1}`('does not retry', () => expect(retry).toHaveBeenCalledTimes(0));",
		`test.extend({})('does not retry', ({ expect: check }) => check(retry).toHaveBeenCalledTimes(0));`,
	} {
		add(code, "")
	}
	for _, code := range []string{
		`(expect(fn).toHaveBeenCalledTimes(0));`,
		`await expect(Promise.resolve(fn)).resolves.toHaveBeenCalledTimes(0);`,
		`test('does not notify', ({ expect }) => { expect(listener).toHaveBeenCalledTimes(0); });`,
	} {
		add(code, strings.ReplaceAll(code, "toHaveBeenCalledTimes(0)", "not.toHaveBeenCalled()"))
	}
	// Poll binds its callback to the assertion whose negation the fix would change.
	for _, code := range []string{
		`await expect.poll(function () { this.toBeNull(); return mock; }, { timeout: 10 }).toHaveBeenCalledTimes(0);`,
		`await expect.poll(function () { observed = this; return mock; }).not.toHaveBeenCalledTimes(0);`,
		`await expect.poll(callback).toHaveBeenCalledTimes(0);`,
		`(await expect.poll(() => fn).toHaveBeenCalledTimes(0));`,
		`test('poll', async ({ expect }) => { await expect.poll(callback).toHaveBeenCalledTimes(0); });`,
	} {
		add(code, "")
	}
	for _, write := range []string{"check = replacement", "check ||= replacement", "[check] = replacements", "({ expect: check } = replacement)", "check++", "for (check of replacements) {}"} {
		valid = append(valid, rule_tester.ValidTestCase{Code: `let { expect: check } = require('@rstest/core'); ` + write + `; check(fn).toHaveBeenCalledTimes(0);`})
	}
	for _, write := range []string{
		"expect = replacement",
		"expect ||= replacement",
		"expect++",
		"[expect] = replacements",
		"({ expect } = replacement)",
		"for (expect of replacements) {}",
		"for (expect in replacements) {}",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: write + `; expect(fn).toHaveBeenCalledTimes(0);`})
	}
	for _, code := range []string{
		`let { expect: check } = require('@rstest/core'); check = () => ({ toHaveBeenCalledTimes() {} }); check(null).toHaveBeenCalledTimes(0);`,
		`let { expect: check } = import.meta.rstest; check = replacement; check(fn).toHaveBeenCalledTimes(0);`,
		`test('context', ({ expect: check }) => { check = replacement; check(fn).toHaveBeenCalledTimes(0); });`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	add(`let { expect: check } = require('@rstest/core'); check(fn).toHaveBeenCalledTimes(0);`, `let { expect: check } = require('@rstest/core'); check(fn).not.toHaveBeenCalled();`)
	// Declaration names are absent from the reference index, but var initializers still overwrite bindings.
	for _, declaration := range []string{
		`var check = () => ({ toHaveBeenCalledTimes() {} });`,
		`var { expect: check } = other;`,
		`var [check] = others;`,
		`var { nested: { check } } = other;`,
		`for (var check of others) {}`,
		`for (var check in other) {}`,
		`for (var { expect: check } of others) {}`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: `var { expect: check } = require('@rstest/core'); ` + declaration + ` check(null).toHaveBeenCalledTimes(0);`})
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: `var { expect: check } = import.meta.rstest; var check = replacement; check(fn).toHaveBeenCalledTimes(0);`},
		rule_tester.ValidTestCase{Code: `test('context', ({ expect: check }) => { var check = replacement; check(fn).toHaveBeenCalledTimes(0); });`},
	)
	for _, code := range []string{
		`var { expect: check } = require('@rstest/core'); var check; check(fn).toHaveBeenCalledTimes(0);`,
		`var check; var { expect: check } = require('@rstest/core'); check(fn).toHaveBeenCalledTimes(0);`,
		`var { expect: check } = require('@rstest/core'); function shadow() { var check = other; } check(fn).toHaveBeenCalledTimes(0);`,
	} {
		add(code, strings.ReplaceAll(code, "toHaveBeenCalledTimes(0)", "not.toHaveBeenCalled()"))
	}
	// Returned Chai assertions retain negation even when reused outside the original chain.
	for _, code := range []string{
		`const assertion = expect(fn).toHaveBeenCalledTimes(0); assertion.toHaveBeenCalledTimes(0);`,
		`const assertion = expect(fn).not.toHaveBeenCalledTimes(0); assertion.toHaveBeenCalledTimes(1);`,
		`assertion = expect(fn).toHaveBeenCalledTimes(0); assertion.toHaveBeenCalledTimes(0);`,
		`function assertion() { return expect(fn).toHaveBeenCalledTimes(0); } assertion().toHaveBeenCalledTimes(0);`,
		`const assertion = () => expect(fn).toHaveBeenCalledTimes(0); assertion().toHaveBeenCalledTimes(0);`,
		`consume(expect(fn).toHaveBeenCalledTimes(0)); function consume(assertion) { assertion.toHaveBeenCalledTimes(0); }`,
		`const assertions = [expect(fn).toHaveBeenCalledTimes(0)]; assertions[0].toHaveBeenCalledTimes(0);`,
		`const result = { assertion: expect(fn).toHaveBeenCalledTimes(0) }; result.assertion.toHaveBeenCalledTimes(0);`,
		`const assertion = (expect(fn).toHaveBeenCalledTimes(0));`,
		`const assertion = expect(fn).toHaveBeenCalledTimes(0) as any;`,
		`const assertion = await expect(Promise.resolve(fn)).resolves.toHaveBeenCalledTimes(0);`,
		`async function assertion() { return await expect.poll(() => fn).toHaveBeenCalledTimes(0); }`,
		`consume(condition ? expect(fn).toHaveBeenCalledTimes(0) : other);`,
		`consume((sideEffect(), expect(fn).toHaveBeenCalledTimes(0)));`,
	} {
		add(code, "")
	}
	// Removing additional arguments, comments, or observable chain state is not safe.
	for _, code := range []string{
		`expect(fn).toHaveBeenCalledTimes(0, notify());`,
		`expect(fn).toHaveBeenCalledTimes(0, ...counts);`,
		`expect(fn).toHaveBeenCalledTimes(/* keep */ 0);`,
		`expect(fn).toHaveBeenCalledTimes(0 /* keep */);`,
		`expect(fn).toHaveBeenCalledTimes</* keep */ number>(0);`,
		`expect(fn).toHaveBeenCalledTimes(0).and.called;`,
		`expect(fn).toBeTypeOf('function').and.toHaveBeenCalledTimes(0);`,
		`expect(fn).toHaveBeenCalledTimes(0).message;`,
		`expect(fn).toHaveBeenCalledTimes(0)();`,
	} {
		add(code, "")
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:   `expect(fn).toHaveBeenCalledTimes(0).and.toHaveBeenCalledTimes(0);`,
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 12}, {MessageId: "preferMatcher", Line: 1, Column: 41}},
	})
	// N/A: declaration/body variants are not inspected by this matcher-only rule.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferToHaveBeenCalledRule, valid, invalid)
}

func TestPreferToHaveBeenCalledSourceOnlyEditDemand(t *testing.T) {
	code := `import { expect as check, test } from '@rstest/core';
check(fn).toHaveBeenCalledTimes(0);
test('context', ({ expect }) => { expect(fn).not.toHaveBeenCalledTimes(0); });
function shadow(check: any) { check(fn).toHaveBeenCalledTimes(0); }
let { expect: changed } = require('@rstest/core');
changed = replacement;
changed(fn).toHaveBeenCalledTimes(0);
expect = () => ({ toHaveBeenCalledTimes() {} });
expect(null).toHaveBeenCalledTimes(0);
await check.poll(function () { this.toBeNull(); return fn; }).toHaveBeenCalledTimes(0);`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "called-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only program")
	}
	var baseline []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program, File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: PreferToHaveBeenCalledRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return PreferToHaveBeenCalledRule.Run(ctx, nil) }}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
		}
		for i := range diagnostics {
			wantFix := (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll) && diagnostics[i].Range.Pos() < strings.Index(code, "await check.poll")
			if (diagnostics[i].FixesPtr != nil) != wantFix {
				t.Fatalf("demand %d: unexpected fixes", demand)
			}
			diagnostics[i].FixesPtr = nil
		}
		if baseline == nil {
			baseline = diagnostics
		} else if !reflect.DeepEqual(baseline, diagnostics) {
			t.Fatalf("demand %d changed diagnostics", demand)
		}
	}
}
