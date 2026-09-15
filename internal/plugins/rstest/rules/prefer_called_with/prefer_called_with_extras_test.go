// Rstest parser and edit boundaries supplement prefer_called_with_upstream_test.go.
package prefer_called_with

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledWithExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `expect(fn).not.toHaveBeenCalledOnce();`},
		{Code: `expect.soft(fn).rejects.not.toHaveBeenCalledOnce();`},
		{Code: `expect(fn)['not'].toHaveBeenCalled();`},
		{Code: `expect(fn).toHaveBeenCalled;`},
		{Code: `expect(fn).resolves.rejects.toHaveBeenCalled();`},
		{Code: `expect(fn).not.not.toHaveBeenCalled();`},
		{Code: `expect(fn).unknown.toHaveBeenCalled();`},
		{Code: `expect.toHaveBeenCalled();`},
		{Code: `expect.toHaveBeenCalledOnce();`},
		{Code: `expect.assertions(1); expect.extend({}); expect.anything();`},
		{Code: `expect(fn).to.have.been.called;`},
		{Code: `expect(fn).calledOnce;`},
		// Only the first matcher is considered, even if a later matcher is call-style.
		{Code: `expect(fn).called.and.toHaveBeenCalled();`},
		{Code: `expect(fn).toBeTypeOf('function').and.toHaveBeenCalled();`},
		{Code: `expect(fn).toHaveBeenCalledWith(1).and.toHaveBeenCalledOnce();`},
		{Code: `expect(fn).not.toHaveBeenCalled().and.toHaveBeenCalledOnce();`},
		{Code: `expect(fn).toBeCalledOnce();`},
		{Code: `other(fn).toHaveBeenCalled();`},
		{Code: `custom.expect(fn).toHaveBeenCalled();`},
		{Code: `import { expect } from 'vitest'; expect(fn).toHaveBeenCalled();`},
		{Code: `import { expect } from '@jest/globals'; expect(fn).toHaveBeenCalled();`},
		{Code: `const expect = makeExpect(); expect(fn).toHaveBeenCalled();`},
		{Code: `import { expect } from '@rstest/core'; function check(expect: any) { expect(fn).toHaveBeenCalled(); }`},
		{Code: `import { expect as check } from '@rstest/core'; function run(check: any) { check(fn).toHaveBeenCalled(); }`},
		{Code: `import * as core from '@rstest/core'; function run(core: any) { core.expect(fn).toHaveBeenCalled(); }`},
		{Code: `function test(name: string, callback: any) {} test('local', ({ expect }) => expect(fn).toHaveBeenCalled());`},
		// Dynamic keys cannot identify a matcher, regardless of the identifier's spelling.
		{Code: `expect(fn)[toHaveBeenCalled]();`},
		{Code: `expect(fn)[('toHaveBeen' + 'Called')]();`},
		{Code: "expect(fn)[`toHaveBeen${suffix}`]();"},
		{Code: `expect(fn)[42]();`},
		// Mutable namespace receivers are not candidates in the existing shared analysis.
		{Code: `let core = require('@rstest/core'); core = replacement; core.expect(fn).toHaveBeenCalled();`},
		{Code: `let core = import.meta.rstest; core = replacement; core.expect(fn).toHaveBeenCalled();`},
		// each supplies row values only; for supplies TestContext as its second argument.
		{Code: `test.each([1])('sends request %s', (value, { expect }) => expect(fn).toHaveBeenCalled());`},
		// Type wrappers on the receiver are boundaries of the existing expect parser.
		{Code: `expect(fn)!.toHaveBeenCalled();`},
		{Code: `(expect(fn) as any).toHaveBeenCalled();`},
		{Code: `(expect(fn) satisfies Assertion).toHaveBeenCalled();`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, code := range []string{
		`expect(fn).rejects.toHaveBeenCalled();`,
		`expect.soft(fn).toHaveBeenCalled();`,
		`expect(fn, 'request sent').toHaveBeenCalled();`,
		`import { expect } from '@rstest/core'; expect(fn).toHaveBeenCalled();`,
		`import { expect as check } from '@rstest/core'; check(fn).toHaveBeenCalled();`,
		`const { expect: check } = require('@rstest/core'); check(fn).toHaveBeenCalled();`,
		`const core = require('@rstest/core'); core.expect(fn).toHaveBeenCalled();`,
		`import * as core from '@rstest/core'; core.expect(fn).toHaveBeenCalled();`,
		`import.meta.rstest.expect(fn).toHaveBeenCalled();`,
		`const { expect: check } = import.meta.rstest; check(fn).toHaveBeenCalled();`,
		`const core = import.meta.rstest; core.expect(fn).toHaveBeenCalled();`,
		`test('sends request', ({ expect }) => { expect(fn).toHaveBeenCalled(); });`,
		`test('sends request', ({ expect: check }) => { check(fn).toHaveBeenCalled(); });`,
		`import { test as check } from '@rstest/core'; check('sends request', ctx => { ctx.expect(fn).toHaveBeenCalled(); });`,
		`test.for([1])('sends request %s', (value, { expect }) => expect(fn).toHaveBeenCalled());`,
		// No Entry gate: poll permits mock matchers; element's narrower types remain the user's constraint.
		`expect.poll(() => fn).toHaveBeenCalled();`,
		`expect.element(locator).toHaveBeenCalled();`,
		// Dimension 4: parentheses, optional chains, generic calls and literal accessor forms.
		`((expect(fn))).toHaveBeenCalled();`,
		`((expect(fn).toHaveBeenCalled))();`,
		`expect?.(fn)?.toHaveBeenCalled?.();`,
		`(expect(fn)?.toHaveBeenCalled)?.();`,
		`expect(fn).toHaveBeenCalled<string>(value);`,
		`expect(fn)['toHaveBeenCalled']();`,
		`expect(fn)["toHaveBeenCalled"]();`,
		"expect(fn)[`toHaveBeenCalled`]();",
		`expect(fn)[(('toHaveBeenCalled'))]();`,
		`expect(fn)?.[ /* key */ 'toHaveBeenCalled' /* end */ ]?.(/* args */);`,
		`expect(fn).toHaveBeenCalled(sideEffect(), ...args);`,
		// Rstest parses the outermost chain once; Vitest reports this matcher twice.
		`expect(fn).toHaveBeenCalled()();`,
		`expect(fn).toHaveBeenCalled().and.not.toBeNull();`,
		`expect(fn).toHaveBeenCalled().and.toHaveBeenCalledOnce();`,
		`expect(fn).to.have.been.toHaveBeenCalled();`,
		`expect(fn).toHaveBeenCalled()[key]();`,
		`expect(fn). /* matcher */ toHaveBeenCalled /* call */ (/* argument */);`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: code, Output: []string{strings.Replace(code, "toHaveBeenCalled", "toHaveBeenCalledWith", 1)},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toHaveBeenCalledWith(/* expected args */)"}},
		})
	}
	invalid = append(invalid,
		// Vitest issues #910 and #917: once needs the real ExactlyOnceWith matcher, not OnceWith.
		rule_tester.InvalidTestCase{
			Code: `expect(fn).toHaveBeenCalledOnce()`, Output: []string{`expect(fn).toHaveBeenCalledExactlyOnceWith()`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toHaveBeenCalledExactlyOnceWith(/* expected args */)"}},
		},
		rule_tester.InvalidTestCase{
			Code: `expect(something).toHaveBeenCalledOnce();`, Output: []string{`expect(something).toHaveBeenCalledExactlyOnceWith();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toHaveBeenCalledExactlyOnceWith(/* expected args */)"}},
		},
		rule_tester.InvalidTestCase{
			Code: "expect(fn).\n  toHaveBeenCalled();", Output: []string{"expect(fn).\n  toHaveBeenCalledWith();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Line: 2, Column: 3, EndLine: 2, EndColumn: 19}},
		},
		rule_tester.InvalidTestCase{
			Code: `expect(fn)["toHaveBeenCalled"]();`, Output: []string{`expect(fn)["toHaveBeenCalledWith"]();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Line: 1, Column: 12, EndLine: 1, EndColumn: 30}},
		},
		rule_tester.InvalidTestCase{
			Code: "expect(fn)[\n  'toHaveBeenCalled'\n]();", Output: []string{"expect(fn)[\n  'toHaveBeenCalledWith'\n]();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Line: 2, Column: 3, EndLine: 2, EndColumn: 21}},
		},
		rule_tester.InvalidTestCase{
			Code: `expect(fn)['\x74oHaveBeenCalled']();`, Output: []string{`expect(fn)['toHaveBeenCalledWith']();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith"}},
		},
		rule_tester.InvalidTestCase{
			Code: `expect(fn).\u0074oHaveBeenCalled();`, Output: []string{`expect(fn).toHaveBeenCalledWith();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith"}},
		},
		rule_tester.InvalidTestCase{
			Code: `expect(expect(fn).toBeCalled()).toHaveBeenCalledOnce();`, Output: []string{`expect(expect(fn).toBeCalledWith()).toHaveBeenCalledExactlyOnceWith();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith"}, {MessageId: "preferCalledWith"}},
		},
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferCalledWithRule, valid, invalid)
}

func TestPreferCalledWithEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"expect(fn).toBeCalled();\nexpect.soft(fn)['toHaveBeenCalledOnce']();",
		"edit-demand.ts", "tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	var baseline []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(), HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: PreferCalledWithRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return PreferCalledWithRule.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		for index := range diagnostics {
			diagnostic := &diagnostics[index]
			if demand&rule.EditDemandAutofix == 0 {
				if diagnostic.FixesPtr != nil {
					t.Fatalf("demand %d materialized fixes", demand)
				}
			} else if diagnostic.FixesPtr == nil || len(*diagnostic.FixesPtr) != 1 {
				t.Fatalf("demand %d: expected one accessor fix", demand)
			}
			if diagnostic.Suggestions != nil {
				t.Fatal("unexpected suggestions")
			}
			diagnostic.FixesPtr = nil
		}
		if baseline == nil {
			baseline = diagnostics
		} else if !reflect.DeepEqual(diagnostics, baseline) {
			t.Fatalf("demand %d changed diagnostics", demand)
		}
	}
}
