// TestRequireHookExtras covers what the migrated upstream suite does not: the
// Rstest utilities-object contract, the two @vitest/eslint-plugin cases Rstest
// reverses, ts-go statement shapes, and the suite forms Rstest adds. The
// complete migrated upstream suite lives in require_hook_upstream_test.go.
package require_hook_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_hook"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRequireHookExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&require_hook.RequireHookRule,
		[]rule_tester.ValidTestCase{
			// ---- A. The utilities object, in every spelling that reaches it ----
			{Code: `rs.mock('./m');`},
			{Code: `rstest.mock('./m');`},
			{Code: `rs['mock']('./m');`},
			{Code: `rs.fn();`},
			{Code: `rstest.clearAllMocks();`},
			{Code: `import { rs } from '@rstest/core'; rs.spyOn(console, 'log');`},
			{Code: `import { rs as r } from '@rstest/core'; r.spyOn(console, 'log');`},
			{Code: `import { rstest as helpers } from '@rstest/core'; helpers.resetAllMocks();`},
			{Code: `import * as core from '@rstest/core'; core.rs.fn();`},
			{Code: `import { rstest } from 'rstack/test'; rstest.restoreAllMocks();`},
			{Code: `import.meta.rstest.rs.mock('./m');`},
			// A member chain rooted at the utilities object stays exempt, which
			// is how mock configuration is written.
			{Code: `rs.spyOn(console, 'log').mockReturnValue(undefined);`},
			{Code: `rs.fn().mockName('handler');`},
			// The build rewrites a plugin-managed member by the receiver as
			// written, so it runs even where the file declares its own `rs`.
			{Code: `const rs = createHelpers(); rs.mock('./m');`},

			// ---- B. Assertions are the neighbouring rules' business ----
			{Code: `expect(value).toBe(1);`},
			{Code: `expect.hasAssertions();`},
			{Code: `expect.soft(value).toBe(1);`},

			// ---- C. Statement shapes that carry no reportable initializer ----
			{Code: `setup?.();`},
			{Code: `foo?.bar();`},
			{Code: `import('x');`},
			{Code: `export let value = setup();`},
			{Code: `const value = setup();`},
			{Code: `let value;`},
			{Code: `describe('a test', () => setup());`},
			{Code: `describe.only('suite', () => {
  beforeEach(() => setup());
});`},
			// A callback passed by name is not descended into: the same
			// function can be registered more than once and called elsewhere.
			{Code: `function body() {
  setup();
}

describe('suite', body);`},

			// ---- D. allowedFunctionCalls matches the whole callee chain ----
			{
				Code:    `helper.setup();`,
				Options: allowedFunctionCalls("helper.setup"),
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- E. The @vitest/eslint-plugin `startsWith('vi')` defect is
			// not inherited. Spelling the same exemption as an `rs` prefix
			// would swallow most Rstest setup helpers. ----
			{
				Code:   `video.play();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `response.json();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `resetDatabase();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `restoreServer().then(done);`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `rsHelpers.setup();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			// A renamed binding that is not the utilities object stays
			// reportable even though it is spelled `rs`.
			{
				Code:   `import { rs } from './helpers'; rs.setup();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 33)},
			},

			// ---- F. Reversed from @vitest/eslint-plugin: Rstest has no such
			// API, so the call is ordinary setup code. ----
			{
				Code: `describe('scoped', () => {
  test.scoped({ example: 'value' });
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code:   `expectTypeOf(value).toBeString();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},

			// ---- G. Rstest APIs that are neither registrations nor hooks ----
			{
				Code:   `assert(value);`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `onTestFinished(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `onTestFailed(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},

			// ---- H. ts-go statement shapes ----
			{
				Code:   `(setup());`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `(foo?.bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `using value = setup();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `await using value = setup();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code: `new NodeExtensionTester()
  .shouldMatch()
  .runTests();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},

			// ---- I. Suite forms whose body still runs during collection ----
			// A skipped suite still executes its callback while the file is
			// collected, so setup written there runs even though no test does.
			{
				Code: `describe.skip('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code: `describe.each([1, 2])('%s', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code: `describe.for([{ a: 1 }])('$a', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code: `describe.runIf(isCI)('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code: `import { describe } from '@rstest/core';

describe('suite', function () {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(4, 3)},
			},
			{
				Code: `import.meta.rstest.describe('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code: `function register() {
  describe('suite', () => {
    setup();
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(3, 5)},
			},
			{
				Code: `if (condition) {
  describe('suite', () => {
    setup();
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(3, 5)},
			},
			// A describe nested in another call is reported as the outer
			// statement only, matching upstream: the outer call is itself
			// unrecognized setup code.
			{
				Code: `foo(describe('suite', () => {
  setup();
}));`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code: `const suite = describe('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			// A foreign framework's describe is not an Rstest registration, so
			// the call itself is the reported setup statement and its body is
			// not descended into.
			{
				Code: `import { describe } from 'vitest';

describe('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(3, 1)},
			},

			// ---- J. allowedFunctionCalls is matched, not prefixed ----
			{
				Code:    `helper.setup();`,
				Options: allowedFunctionCalls("setup"),
				Errors:  []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
		},
	)
}

func TestRequireHookSourceOnly(t *testing.T) {
	if require_hook.RequireHookRule.RequiresTypeInfo {
		t.Fatal("rstest/require-hook must run without type information")
	}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "require-hook-source-only.ts")
	code := `import { rs as helpers } from '@rstest/core';
helpers.mock('./m');
setup();
describe('suite', () => { teardown(); });`
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
				Name: require_hook.RequireHookRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return require_hook.RequireHookRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})
	if len(diagnostics) != 2 {
		t.Fatalf("expected the renamed utilities import to stay exempt without a TypeChecker, got: %#v", diagnostics)
	}
}
