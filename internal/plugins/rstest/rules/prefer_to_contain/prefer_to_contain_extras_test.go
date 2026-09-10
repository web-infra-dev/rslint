// TestPreferToContainExtras covers Rstest expect sources, runtime-semantic
// exclusions, tsgo edge shapes, fix safety and shared-engine branch lock-ins.
// The complete upstream baseline is in prefer_to_contain_upstream_test.go.
package prefer_to_contain

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToContainExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToContainRule,
		[]rule_tester.ValidTestCase{
			// ---- Rstest runtime boundary: poll retries its callback ----
			{Code: `await expect.poll(() => list.includes(item)).toBe(true);`},
			// ---- Rstest runtime boundary: promise modifiers assert a settled value ----
			{Code: `await expect(Promise.resolve(list.includes(item))).resolves.toBe(true);`},
			{Code: `await expect(Promise.reject(list.includes(item))).rejects.toEqual(true);`},
			// ---- Rstest runtime boundary: browser element expect has another matcher set ----
			{Code: `expect.element(locator).toEqual(true);`},
			// ---- Rstest runtime boundary: Array#includes and toContain disagree on NaN ----
			{Code: `expect([NaN].includes(NaN)).toBe(true);`},
			{Code: `expect(values.includes(NaN as number)).toEqual(false);`},
			{Code: `expect(values.includes(Number.NaN)).toBe(true);`},
			{Code: `expect(values.includes(globalThis['NaN'])).toBe(false);`},
			{Code: `expect('abc'.includes(/a/ as any)).toBe(false);`},
			{Code: "expect(`a${value}`.includes(/a/ as any)).toBe(false);"},
			// ---- Rstest provenance: foreign and locally shadowed expect values ----
			{Code: `import { expect } from 'vitest'; expect(list.includes(item)).toBe(true);`},
			{Code: `import { expect } from '@jest/globals'; expect(list.includes(item)).toBe(true);`},
			{Code: `const expect = createExpect(); expect(list.includes(item)).toBe(true);`},
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(list.includes(item)).toBe(true); }`},
			// ---- Dimension 4: unsupported and dynamic accessors ----
			{Code: `expect(list[includes](item)).toBe(true);`},
			{Code: `const key = "incl" + suffix; expect(list[key](item)).toBe(true);`},
			{Code: `expect(list.includes?.(item)).toBe(true);`},
			{Code: `expect(list?.includes(item)).toBe(true);`},
			{Code: `expect(list.includes(item))[matcher](true);`},
			{Code: `expect(list.includes(item))[0](true);`},
			// ---- Dimension 4: calls and arguments that do not match the rule grammar ----
			{Code: `expect(list.includes()).toBe(true);`},
			{Code: `expect(list.includes(item, 0)).toBe(true);`},
			{Code: `expect(list.includes(...items)).toBe(true);`},
			{Code: `expect(list.includes(item)).toBe();`},
			{Code: `expect(list.includes(item)).toBe(true, false);`},
			{Code: `expect(list.includes(item)).toBe(...values);`},
			{Code: `expect(list.includes(item)).toBeTruthy();`},
			{Code: `expect(list.includes(item)).to.equal(true);`},
			{Code: `(expect(list.includes(item)) as any).toBe(true);`},
			{Code: `expect(list.includes(item))!.toBe(true);`},
			// N/A: declarations, functions, class members, object keys and binding
			// patterns are not nodes inspected by this assertion-call rule.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parentheses and assertion wrappers ----
			{
				Code:   `expect(((list.includes(item)))).toEqual((true as boolean));`,
				Output: []string{`expect(((list))).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Message: "Use toContain() instead", Line: 1, Column: 33}},
			},
			{
				Code:   `expect(list["includes"]((item as Item))).toStrictEqual(false as boolean);`,
				Output: []string{`expect(list).not.toContain((item as Item));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 42}},
			},
			// ---- Dimension 4: authored matcher accessor spelling survives ----
			{
				Code:   `expect(list.includes(item))["toEqual"](true);`,
				Output: []string{`expect(list)["toContain"](item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 29}},
			},
			{
				Code:   `expect(list.includes(item)).toEqual<boolean>(true);`,
				Output: []string{`expect(list).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 29}},
			},
			{
				Code:   `expect(list.includes(item)).toEqual</* keep */ boolean>(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 29}},
			},
			{
				Code:   "expect(list[`includes`](item))[`toBe`](false);",
				Output: []string{"expect(list).not[`toContain`](item);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 32}},
			},
			// ---- Dimension 3: comments outside replaced expressions survive ----
			{
				Code:   `expect(list.includes(item)). /* matcher */ toBe(true);`,
				Output: []string{`expect(list). /* matcher */ toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 44}},
			},
			// ---- Dimension 3: comments inside replaced expressions withhold fixes ----
			{
				Code:   `expect(list.includes(/* keep */ item)).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 40}},
			},
			{
				Code:   `expect(list.includes(item)).toBe(/* keep */ true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 29}},
			},
			// Moving the item after expect(message()) would reorder evaluation.
			{
				Code:   `expect(list.includes(nextItem()), message()).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 46}},
			},
			{
				Code:   `expect(list.includes(nextItem())).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 35}},
			},
			{
				Code:   `expect(list.includes(item.value)).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 35}},
			},
			// Unary numeric coercion can invoke user code, so moving it past
			// expect() would change the observable evaluation order.
			{
				Code:   `expect(list.includes(+item)).toBe(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 30}},
			},
			// ---- Rstest source resolution ----
			{
				Code: `import { expect as check } from '@rstest/core';
check(list.includes(item)).toBe(true);`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(list).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 2, Column: 28}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(list.includes(item)).not.toEqual(false);`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(list).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 2, Column: 38}},
			},
			{
				Code: `const { expect } = require('rstack/test');
expect(list.includes(item)).toStrictEqual(false);`,
				Output: []string{`const { expect } = require('rstack/test');
expect(list).not.toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 2, Column: 29}},
			},
			{
				Code:   `import.meta.rstest.expect(list.includes(item)).toBe(true);`,
				Output: []string{`import.meta.rstest.expect(list).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 48}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(list.includes(item)).toEqual(true);`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(list).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 2, Column: 29}},
			},
			{
				Code:   `test('contains', ({ expect: check }) => check(list.includes(item)).toBe(true));`,
				Output: []string{`test('contains', ({ expect: check }) => check(list).toContain(item));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 68}},
			},
			// ---- Rstest factory: soft assertions preserve soft behavior ----
			{
				Code:   `expect.soft(list.includes(item)).not.toBe(false);`,
				Output: []string{`expect.soft(list).toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 38}},
			},
			// Locks in equality matcher, boolean literal and negation truth table.
			{
				Code:   `expect(list.includes(item)).not.toStrictEqual(true);`,
				Output: []string{`expect(list).not.toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Message: "Use toContain() instead", Line: 1, Column: 33, EndLine: 1, EndColumn: 46}},
			},
			{
				Code: `expect(list.includes(item))
  .toEqual(true);`,
				Output: []string{`expect(list)
  .toContain(item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Message: "Use toContain() instead", Line: 2, Column: 4, EndLine: 2, EndColumn: 11}},
			},
			// Array.from() materializes sparse holes before Rstest's matcher checks
			// containment, so this remains equivalent to includes(undefined).
			{
				Code:   `expect([,].includes(undefined)).toBe(true);`,
				Output: []string{`expect([,]).toContain(undefined);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 33}},
			},
			{
				Code:   `expect('123'.includes(2 as any)).toBe(true);`,
				Output: []string{`expect('123').toContain(2 as any);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 34}},
			},
			// ---- Real-user: jest#1009 bracket accessor regression ----
			{
				Code:   `expect(list['includes'](item))['not']['toEqual'](false);`,
				Output: []string{`expect(list)['toContain'](item);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 39}},
			},
			// ---- Real-user: jest#1282 trailing-comma regression ----
			{
				Code:   `expect(list.includes(item,),).toBe(false,);`,
				Output: []string{`expect(list,).not.toContain(item,);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 31}},
			},
			// A trailing call reaches the matcher through two call nodes but reports once.
			{
				Code:   `expect(list.includes(item)).toBe(true).then(done);`,
				Output: []string{`expect(list).toContain(item).then(done);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 29}},
			},
		},
	)
}

func TestPreferToContainEditDemand(t *testing.T) {
	t.Parallel()
	help := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := help.CreateTestProgram(
		`expect(list.includes(item)).not.toEqual(false);`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(), HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: PreferToContainRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return PreferToContainRule.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
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
		rule.EditDemandNone: diagnosticsOnly, rule.EditDemandAutofix: autofixOnly, rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d diagnostic changed:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.FixesPtr != nil || suggestionOnly.FixesPtr != nil {
		t.Fatal("autofixes were materialized without being requested")
	}
	if autofixOnly.FixesPtr == nil || allEdits.FixesPtr == nil || !reflect.DeepEqual(*autofixOnly.FixesPtr, *allEdits.FixesPtr) {
		t.Fatal("requested autofixes do not match all-edits output")
	}
	for _, diagnostic := range []rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		if diagnostic.Suggestions != nil {
			t.Fatal("prefer-to-contain unexpectedly materialized suggestions")
		}
	}
}
