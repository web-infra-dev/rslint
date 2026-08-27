// TestNoHooksExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch, Dimension 4 row, or Rstest parser contract it covers, so
// future refactors can't silently regress them without breaking a named lock-in.
package no_hooks_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_hooks"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoHooksExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_hooks.NoHooksRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized receiver on an unrelated member call ----
			{Code: `(subject).beforeEach()`},
			// ---- Dimension 4: optional chain on an unrelated member call ----
			{Code: `subject?.beforeEach()`},
			// ---- Dimension 4: element access lookalike is not a parsed hook ----
			{Code: `subject["beforeEach"]()`},
			// ---- Dimension 4: parenthesized global identifier still resolves ----
			{Code: `(beforeEach).call(subject, () => {})`},
			// ---- N/A: numeric normalization ----
			// This rule does not compare literal values.
			// ---- N/A: computed-key fix invariance ----
			// This rule does not provide fixes or suggestions.
			// ---- N/A: edit-demand invariance ----
			// This rule does not provide fixes or suggestions.
			// ---- Real-user: execution-time lifecycle callbacks are not hooks ----
			{Code: `onTestFinished(() => {}); onTestFailed(() => {});`},
			{Code: `import { onTestFinished, onTestFailed } from '@rstest/core'; onTestFinished(() => {}); onTestFailed(() => {});`},
			// ---- Real-user: core member lookalikes and invalid chains stay ignored ----
			{Code: `import { test } from '@rstest/core'; test.beforeEach(() => {});`},
			{Code: `import { test } from '@rstest/playwright'; test.skip.beforeEach(() => {});`},
			{Code: `import { test } from '@rstest/playwright'; test.beforeEach.only(() => {});`},
			// ---- Locks in parser contract: non-identifier receivers never resolve ----
			// jest/no-hooks reports all four of these; ParseRstestFnCall does not.
			{Code: `new A().beforeEach()`},
			{Code: `arr[0].beforeEach()`},
			{Code: `this.beforeEach()`},
			{Code: `(<any>x).beforeEach()`},
			// ---- Locks in parser contract: a trailing chain is not a hook registration ----
			{Code: `beforeEach.skip(() => {})`},
			{Code: `beforeEach.each([1])(() => {})`},
			// ---- Locks in parser contract: non-plain-call positions ----
			{Code: `new beforeEach()`},
			{Code: "beforeEach`x`"},
			{Code: `(0, beforeEach)(() => {})`},
			{Code: `globalThis.beforeEach(() => {})`},
			// ---- Locks in upstream create() branch: non-hook CallExpression is ignored ----
			{Code: `notAHook(() => {})`},
			// ---- Locks in upstream create() branch: foreign imports are ignored ----
			{Code: `import { beforeAll } from 'vitest'; beforeAll(() => {});`},
			{Code: `import { beforeEach } from '@jest/globals'; beforeEach(() => {});`},
			// ---- Locks in upstream create() branch: local shadowing is ignored ----
			{Code: `const beforeAll = createHookRegistry(); beforeAll(() => {});`},
			{Code: `function run(beforeEach) { beforeEach(() => {}); }`},
			// ---- Locks in upstream allow.includes(name) true arm for semantic alias ----
			{
				Code: `import { beforeEach as teardown } from '@rstest/core'; teardown(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"beforeEach"}},
				},
			},
			// ---- Locks in upstream allow.includes(name) true arm for Playwright member hook ----
			{
				Code: `import { test } from '@rstest/playwright'; test.beforeEach(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"beforeEach"}},
				},
			},
			// ---- Locks in parser branch: import.meta source forms ----
			{
				Code: `const { afterAll } = import.meta.rstest; afterAll(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"afterAll"}},
				},
			},
			{
				Code: `const api = import.meta.rstest; api.beforeAll(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"beforeAll"}},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in parser branch: core globals and timeout overload ----
			{
				Code: `beforeAll(() => {}, 1000);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeAll' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 26,
				}},
			},
			// ---- Locks in parser branch: require destructuring ----
			{
				Code: `const { beforeEach } = require('@rstest/core'); beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    49,
				}},
			},
			// ---- Locks in parser branch: namespace import ----
			{
				Code: `import * as rstest from '@rstest/core'; rstest.afterEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterEach' hook",
					Line:      1,
					Column:    41,
				}},
			},
			// ---- Locks in parser branch: whole-module require ----
			{
				Code: `const rstest = require('@rstest/core'); rstest.afterAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterAll' hook",
					Line:      1,
					Column:    41,
				}},
			},
			// ---- Locks in parser branch: import.meta direct call ----
			{
				Code: `import.meta.rstest.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeAll' hook",
					Line:      1,
					Column:    1,
				}},
			},
			// ---- Locks in parser branch: same-file alias ----
			{
				Code: `const setup = beforeAll; setup(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeAll' hook",
					Line:      1,
					Column:    26,
				}},
			},
			// ---- Real-user: Playwright named hook export ----
			{
				Code: `import { beforeAll } from '@rstest/playwright'; beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeAll' hook",
					Line:      1,
					Column:    49,
				}},
			},
			// ---- Real-user: Playwright member hooks ----
			{
				Code: `import { test } from '@rstest/playwright'; test.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeAll' hook",
					Line:      1,
					Column:    44,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright'; test.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    44,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright'; test.afterEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterEach' hook",
					Line:      1,
					Column:    44,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright'; test.afterAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterAll' hook",
					Line:      1,
					Column:    44,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright'; test.extend({}).beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    44,
				}},
			},
			// ---- Locks in parser branch: element access on a resolved namespace ----
			{
				Code: `import * as rstest from '@rstest/core'; rstest["beforeEach"](() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    41,
					EndLine:   1,
					EndColumn: 71,
				}},
			},
			// ---- Locks in upstream create() branch: the report does not inspect arguments ----
			{
				Code: `beforeEach()`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `beforeEach?.(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 23,
				}},
			},
			{
				Code: `beforeEach(...fns)`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			// ---- Real-user: the shape every suite uses, a hook inside a describe callback ----
			{
				Code: `
import { describe, it, beforeEach } from '@rstest/core';

describe("suite", () => {
	beforeEach(() => {});

	it("works", () => {});
});
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      5,
					Column:    2,
					EndLine:   5,
					EndColumn: 22,
				}},
			},
			// ---- Locks in upstream create() branch: one file reports every hook, outermost first ----
			{
				Code: `beforeEach(() => { afterEach(() => {}) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedHook",
						Message:   "Unexpected 'beforeEach' hook",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 42,
					},
					{
						MessageId: "unexpectedHook",
						Message:   "Unexpected 'afterEach' hook",
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 39,
					},
				},
			},
			// ---- Locks in upstream allow.includes(name) false arm with semantic alias ----
			{
				Code: `import { beforeEach as teardown } from '@rstest/core'; teardown(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"afterEach"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    56,
				}},
			},
			// ---- Locks in upstream allow.includes(name) false arm with Playwright member hook ----
			{
				Code: `import { test } from '@rstest/playwright'; test.beforeEach(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"afterEach"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    44,
				}},
			},
		},
	)
}

func TestNoHooksAllowSchema(t *testing.T) {
	valid := []any{map[string]any{"allow": []any{"beforeEach", "afterAll"}}}
	if err := no_hooks.NoHooksRule.Schema.Validate(valid); err != nil {
		t.Errorf("expected hook names to pass schema validation, got: %v", err)
	}

	invalidCases := []struct {
		name   string
		config []any
	}{
		{
			name:   "typo rejected",
			config: []any{map[string]any{"allow": []any{"beforeeach"}}}, // cspell:ignore beforeeach
		},
		{
			name:   "additional property rejected",
			config: []any{map[string]any{"allow": []any{"beforeEach"}, "extra": true}},
		},
		{
			name:   "second option rejected",
			config: []any{map[string]any{"allow": []any{"beforeEach"}}, map[string]any{}},
		},
	}
	for _, tc := range invalidCases {
		if err := no_hooks.NoHooksRule.Schema.Validate(tc.config); err == nil {
			t.Errorf("%s: expected schema validation to fail", tc.name)
		}
	}
}

func TestNoHooksSkipsSourceOnlyPrograms(t *testing.T) {
	if !no_hooks.NoHooksRule.RequiresTypeInfo {
		t.Fatal("rstest/no-hooks must require type information")
	}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "source-only.ts")
	code := `
const beforeEach = createHookRegistry();
beforeEach(() => {});
import { beforeEach as vitestBeforeEach } from 'vitest';
vitestBeforeEach(() => {});
import { beforeEach as setup } from '@rstest/core';
setup(() => {});
`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}

	diagnosticCount := 0
	programs := []*lintprogram.Program{program}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             no_hooks.NoHooksRule.Name,
				Severity:         rule.SeverityError,
				RequiresTypeInfo: no_hooks.NoHooksRule.RequiresTypeInfo,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_hooks.NoHooksRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Report: func(rule.RuleDiagnostic) { diagnosticCount++ },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if diagnosticCount != 0 {
		t.Fatalf("source-only run produced %d diagnostics", diagnosticCount)
	}
	if _, ran := result.ExecutedRules[no_hooks.NoHooksRule.Name]; ran {
		t.Fatal("type-aware rule ran for a source-only Program")
	}
}
