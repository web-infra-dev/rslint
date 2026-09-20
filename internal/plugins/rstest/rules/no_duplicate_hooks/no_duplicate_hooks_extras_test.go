// TestNoDuplicateHooksExtras covers Rstest provenance, Rstest-only suite APIs,
// source-only execution, and lexical-policy lock-ins. The complete migrated
// upstream suite lives in no_duplicate_hooks_upstream_test.go.
package no_duplicate_hooks_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_duplicate_hooks"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoDuplicateHooksExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_hooks.NoDuplicateHooksRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: foreign APIs and local shadows are ignored. ----
			{Code: `import { beforeEach } from 'vitest'; beforeEach(() => {}); beforeEach(() => {});`},
			{Code: `import { afterAll } from '@jest/globals'; afterAll(() => {}); afterAll(() => {});`},
			{Code: `function run(beforeEach) { beforeEach(() => {}); beforeEach(() => {}); }`},
			{Code: `const beforeAll = createHookRegistry(); beforeAll(() => {}); beforeAll(() => {});`},

			// ---- Dimension 2: factories alone and invalid hook chains do not register hooks. ----
			{Code: `describe.each([1]); describe.for([1]); describe.runIf(true); describe.skipIf(false);`},
			{Code: `beforeEach.skip(() => {}); beforeEach.skip(() => {});`},
			{Code: `import { test } from '@rstest/core'; test.beforeEach(() => {}); test.beforeEach(() => {});`},

			// ---- Dimension 4: dynamic members and unrelated receivers are ignored. ----
			{Code: `import * as rstest from '@rstest/core'; const hook = 'beforeEach'; rstest[hook](() => {}); rstest[hook](() => {});`},
			{Code: `subject.beforeEach(() => {}); subject.beforeEach(() => {});`},

			// Locks in upstream lexical policy: nested describe scopes are independent.
			{Code: `beforeEach(() => {}); describe('suite', () => { beforeEach(() => {}); });`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: named/renamed, namespace, CommonJS, re-export, and import.meta forms. ----
			{
				Code: `import { beforeEach as setup, beforeEach as prepare } from '@rstest/core';
setup(() => {});
prepare(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 1)},
			},
			{
				Code: `import * as rstest from '@rstest/core';
rstest.afterEach(() => {});
rstest.afterEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterEach", 3, 1)},
			},
			{
				Code: `const { afterAll: cleanup } = require('@rstest/core');
cleanup(() => {});
cleanup(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterAll", 3, 1)},
			},
			{
				Code: `import { beforeAll } from 'rstack/test';
beforeAll(() => {});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeAll", 3, 1)},
			},
			{
				Code: `import.meta.rstest.beforeAll(() => {});
import.meta.rstest.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeAll", 2, 1)},
			},

			// ---- Dimension 1: Playwright named and member hooks share semantic names. ----
			{
				Code: `import { beforeEach, test } from '@rstest/playwright';
beforeEach(() => {});
test.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 1)},
			},
			{
				Code: `import { test } from '@rstest/playwright';
const appTest = test.extend({});
appTest.afterEach(() => {});
appTest.afterEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterEach", 4, 1)},
			},

			// ---- Dimension 2: Rstest modifiers and describe.for open lexical scopes. ----
			{
				Code: `describe.runIf(enabled)('suite', () => {
  afterAll(() => {});
  afterAll(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterAll", 3, 3)},
			},
			{
				Code: `describe.for([1])('suite', () => {
  beforeEach(() => {});
  beforeEach(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 3)},
			},

			// Locks in upstream lexical policy: callback execution is not simulated.
			{
				Code: `describe('suite', () => {
  beforeEach(() => {});
  test('case', () => { beforeEach(() => {}); });
});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 24)},
			},
			{
				Code: `describe('suite', () => {
  beforeEach(() => {});
  function helper() { beforeEach(() => {}); }
});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 23)},
			},
			{
				Code: `beforeEach(() => {});
describe('suite', callback);
function callback() { beforeEach(() => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 23)},
			},
			{
				Code: `describe.each([])('suite', () => {
  beforeEach(() => {});
  beforeEach(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 3)},
			},
		},
	)
}

func TestNoDuplicateHooksSourceOnly(t *testing.T) {
	if no_duplicate_hooks.NoDuplicateHooksRule.RequiresTypeInfo {
		t.Fatal("rstest/no-duplicate-hooks must run without type information")
	}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "no-duplicate-hooks-source-only.ts")
	code := `import { beforeEach as setup } from '@rstest/core';
setup(() => {});
setup(() => {});
describe('suite', () => { afterAll(() => {}); afterAll(() => {}); });`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program")
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: no_duplicate_hooks.NoDuplicateHooksRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_duplicate_hooks.NoDuplicateHooksRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})
	if len(diagnostics) != 2 ||
		diagnostics[0].Message.Data["hook"] != "beforeEach" ||
		diagnostics[1].Message.Data["hook"] != "afterAll" {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}
