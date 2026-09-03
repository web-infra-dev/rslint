// TestRequireTestTimeoutExtras locks in the Rstest-only augmentation the port
// spec asks for: every registration source, every registration shape, the
// Dimension 4 edge shapes each child-node access has to survive, and the two
// exemptions Rstest has that the Vitest rule never modeled — suite-level
// inheritance and the runtime config object.
package require_test_timeout_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_test_timeout"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireTestTimeoutExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_test_timeout.RequireTestTimeoutRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Reverse registration sources ----
			{Code: `import { test } from 'vitest'; test("a", () => {})`},
			{Code: `import { test } from '@jest/globals'; test("a", () => {})`},
			{Code: `import { it } from 'node:test'; it("a", () => {})`},
			{Code: `import { test } from './helpers'; test("a", () => {})`},
			// A local declaration shadows the Rstest global.
			{Code: `function test(name, fn) {}
test("a", () => {})`},
			{Code: `const it = (name, fn) => {};
it("a", () => {})`},

			// ---- B. Registrations this rule does not own ----
			{Code: `test.todo("a")`},
			{Code: `test.todo("a", () => {})`},
			{Code: `it.todo("a")`},
			{Code: `test.skip("a", () => {})`},
			{Code: `test.skip.each([1])("a %i", () => {})`},
			// A skip carried in through an alias is still a skip: the parser
			// records it as a semantic conclusion, not a call-site member.
			{Code: `const s = test.skip; s("a", () => {})`},
			// Suites and hooks are not tests.
			{Code: `describe("s", () => {})`},
			{Code: `beforeEach(() => {})`},
			{Code: `afterAll(() => {}, 500)`},
			// A registration with no callback has no body to time out;
			// reporting it belongs to rstest/prefer-todo.
			{Code: `test("a")`},
			{Code: `test()`},
			{Code: `test("a", ...rest)`},
			{Code: `test(...args)`},
			// The callback is imported, so the parser cannot see the function
			// it registers and the rule declines to guess.
			{Code: `import { handler } from './handler'; test("a", handler)`},

			// ---- C. Own timeout, numeric forms ----
			{Code: `test("a", () => {}, 0)`},
			{Code: `test("a", () => {}, -0)`},
			{Code: `test("a", () => {}, +500)`},
			{Code: `test("a", () => {}, 5e3)`},
			{Code: `test("a", () => {}, 0x1f)`},
			{Code: `test("a", () => {}, 5_000)`},
			{Code: `test("a", () => {}, 1.5)`},
			// A BigInt is not a number, so it is left unread rather than
			// reported.
			{Code: `test("a", () => {}, 500n)`},

			// ---- C. Own timeout, options forms ----
			{Code: `test("a", { timeout: 500 }, () => {})`},
			{Code: `test("a", { "timeout": 500 }, () => {})`},
			{Code: `test("a", { ["timeout"]: 500 }, () => {})`},
			{Code: "test(\"a\", { [`timeout`]: 500 }, () => {})"},
			{Code: `const timeout = 500; test("a", { timeout }, () => {})`},
			{Code: `test("a", { timeout: 500, retry: 2 }, () => {})`},
			// A computed key could be `timeout`, and a spread could carry it.
			{Code: `test("a", { [key]: 500 }, () => {})`},
			{Code: `test("a", { ...options }, () => {})`},
			// Later members overwrite earlier ones, so a spread or a computed
			// key after an explicit `timeout` can still supply the timeout.
			{Code: `test("a", { timeout: -1, ...options }, () => {})`},
			{Code: `test("a", { timeout: -1, [key]: 500 }, () => {})`},
			{Code: `test("a", { timeout: -1, timeout: 500 }, () => {})`},
			{Code: `test("a", { retry: 2, ...options }, () => {})`},
			{Code: `rs.setConfig({ testTimeout: -1, ...config }); test("a", () => {})`},
			// A `timeout` that is not a readable number leaves the options
			// object unread rather than reported.
			{Code: `test("a", { timeout: getTimeout() }, () => {})`},
			{Code: `test("a", { timeout: "500" }, () => {})`},
			{Code: `test("a", { timeout() { return 500 } }, () => {})`},

			// ---- C. Own timeout, resolved through a const ----
			{Code: `const t = 500; test("a", () => {}, t)`},
			{Code: `const options = { timeout: 500 }; test("a", options, () => {})`},
			{Code: `const options = { timeout: t }; test("a", options, () => {})`},
			{Code: `let t = 500; test("a", () => {}, t)`},
			{Code: `var t = 500; test("a", () => {}, t)`},
			{Code: `test("a", () => {}, t)`},

			// ---- D. Dimension 4 edge shapes ----
			// Parentheses and TypeScript assertions around the arguments.
			{Code: `test("a", (() => {}), (500))`},
			{Code: `test("a", ((() => {})), ((500)))`},
			{Code: `test("a", () => {}, 500 as number)`},
			{Code: `test("a", () => {}, 500!)`},
			{Code: `test("a", { timeout: 500 } satisfies Options, () => {})`},
			{Code: `test("a", { timeout: (500) }, () => {})`},
			{Code: `const t = (500); test("a", () => {}, t)`},
			// Optional call and optional member access on the registration.
			{Code: `test?.("a", () => {}, 500)`},
			{Code: `test?.only("a", () => {}, 500)`},
			// Member accessor spellings, including one split over lines.
			{Code: `test["concurrent"]("a", () => {}, 500)`},
			{Code: "test[`concurrent`](\"a\", () => {}, 500)"},
			{Code: `test
  // registers concurrently
  .concurrent("a", () => {}, 500)`},
			// N/A: private identifiers. This rule reads registration arguments
			// and object literal keys, neither of which can be a `#name`.

			// ---- E. Parameterized registrations ----
			{Code: `test.each([1, 2])("case %i", (n) => {}, 500)`},
			{Code: `test.for([1, 2])("case %s", (n) => {}, 500)`},
			{Code: `test.each([1, 2])("case %i", { timeout: 500 }, (n) => {})`},
			{Code: "test.each`\n  a\n  ${1}\n`(\"case $a\", ({ a }) => {}, 500)"},
			{Code: `test.runIf(ready)("a", () => {}, 500)`},
			{Code: `test.skipIf(broken)("a", () => {}, 500)`},
			{Code: `test.fails("a", () => {}, 500)`},
			{Code: `test.sequential("a", () => {}, 500)`},
			{Code: `test.extend({})("a", () => {}, 500)`},

			// ---- F. Registration sources that do report, shown timed ----
			{Code: `import { test } from '@rstest/core'; test("a", () => {}, 500)`},
			{Code: `import { it as check } from '@rstest/core'; check("a", () => {}, 500)`},
			{Code: `import * as core from '@rstest/core'; core.test("a", () => {}, 500)`},
			{Code: `const { test } = require('@rstest/core'); test("a", () => {}, 500)`},
			{Code: `import.meta.rstest.test("a", () => {}, 500)`},
			{Code: `import { test } from '@rstest/playwright'; test("a", () => {}, 500)`},

			// ---- G. Suite-level inheritance ----
			// Rstest applies suite options as defaults
			// (packages/core/src/runtime/runner/runtime.ts), so a test inside a
			// timed suite already has a definite timeout.
			{Code: `describe("s", { timeout: 5000 }, () => { test("a", () => {}) })`},
			{Code: `describe("s", () => { test("a", () => {}) }, 5000)`},
			{Code: `describe("outer", { timeout: 5000 }, () => { describe("inner", () => { test("a", () => {}) }) })`},
			{Code: `describe("s", { timeout: 5000 }, function () { test("a", () => {}) })`},
			{Code: `describe.each([1])("s %i", () => { test("a", () => {}) }, 5000)`},
			{Code: `const options = { timeout: 5000 }; describe("s", options, () => { test("a", () => {}) })`},
			// An options object the rule cannot read may still carry a timeout.
			{Code: `describe("s", options, () => { test("a", () => {}) })`},
			// The suite's own timeout still counts through an intermediate
			// callback, because the test is registered while the suite runs.
			{Code: `describe("s", { timeout: 5000 }, () => { [1, 2].forEach(() => { test("a", () => {}) }) })`},
			// A function this rule cannot tie back to a call argument could be
			// registered under any suite, so the walk gives up rather than
			// guess which one.
			{Code: `function suite() { test("a", () => {}) }
describe("s", { timeout: 5000 }, suite)`},
			{Code: `function suite() { test("a", () => {}) }
describe("s", suite)`},
			{Code: `(function () { test("a", () => {}) })()`},
			{Code: `class Suite { register() { test("a", () => {}) } }`},

			// ---- H. Runtime config through the utility object ----
			{Code: `rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `rstest.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `rs.setConfig({ testTimeout: 0 }); test("a", () => {})`},
			{Code: `rs["setConfig"]({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: "rs[`setConfig`]({ testTimeout: 5000 }); test(\"a\", () => {})"},
			{Code: `rs?.setConfig?.({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `import.meta.rstest.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `import { rs as runtime } from '@rstest/core'; runtime.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `import * as core from '@rstest/core'; core.rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `const core = require('@rstest/core'); core.rstest.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `const { rs } = require('@rstest/core'); rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `const config = { testTimeout: 5000 }; rs.setConfig(config); test("a", () => {})`},
			// A configuration object the rule cannot read may still set the
			// test timeout.
			{Code: `rs.setConfig(config); test("a", () => {})`},
			{Code: `rs.setConfig({ ...config }); test("a", () => {})`},
			{Code: `rs.setConfig({ [key]: 5000 }); test("a", () => {})`},
			{Code: `rs.setConfig({ testTimeout: getTimeout() }); test("a", () => {})`},
			{Code: `rs.setConfig({ testTimeout: 5000 }); rs.resetConfig(); rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `rs.resetConfig(); rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			// The configuration applies to every test registered after it.
			{Code: `rs.setConfig({ testTimeout: 5000 }); test("a", () => {}); test("b", () => {})`},
			// The exemption is positional, so a configuration written inside an
			// earlier test's callback covers the tests that follow it in the
			// source. Whether that callback has run by then is not something
			// the rule can see, and assuming it has not would report a test
			// whose timeout is configured.
			{Code: `test("a", () => { rs.setConfig({ testTimeout: 5000 }) }, 5); test("b", () => {})`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- A. Registration sources ----
			invalidCase(
				`import { test } from '@rstest/core'; test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`import { it as check } from '@rstest/core'; check("a", () => {})`,
				`check("a", () => {})`,
			),
			invalidCase(
				`import * as core from '@rstest/core'; core.test("a", () => {})`,
				`core.test("a", () => {})`,
			),
			invalidCase(
				`const { test } = require('@rstest/core'); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`import.meta.rstest.test("a", () => {})`,
				`import.meta.rstest.test("a", () => {})`,
			),
			invalidCase(
				`const { test } = import.meta.rstest; test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`import { test } from '@rstest/playwright'; test("a", () => {})`,
				`test("a", () => {})`,
			),
			// A plain same-file alias is followed through to the registration.
			invalidCase(`const t = test; t("a", () => {})`, `t("a", () => {})`),

			// ---- B. Registration shapes ----
			invalidCase(`test.each([1, 2])("case %i", (n) => {})`, `test.each([1, 2])("case %i", (n) => {})`),
			invalidCase(`test.for([1, 2])("case %s", (n) => {})`, `test.for([1, 2])("case %s", (n) => {})`),
			invalidCase(
				"test.each`\n  a\n  ${1}\n`(\"case $a\", ({ a }) => {})",
				"test.each`\n  a\n  ${1}\n`(\"case $a\", ({ a }) => {})",
			),
			invalidCase(`test.runIf(ready)("a", () => {})`, `test.runIf(ready)("a", () => {})`),
			invalidCase(`test.fails("a", () => {})`, `test.fails("a", () => {})`),
			invalidCase(`test.sequential("a", () => {})`, `test.sequential("a", () => {})`),
			invalidCase(`test["concurrent"]("a", () => {})`, `test["concurrent"]("a", () => {})`),
			invalidCase(`test?.("a", () => {})`, `test?.("a", () => {})`),
			invalidCase(
				`test
  // registers concurrently
  .concurrent("a", () => {})`,
				`test
  // registers concurrently
  .concurrent("a", () => {})`,
			),

			// ---- C. Callbacks named rather than inlined ----
			// Upstream only looks for an inline function and so never reports
			// these. A named callback needs a timeout just as much.
			invalidCase(
				`const handler = () => {}; test("a", handler)`,
				`test("a", handler)`,
			),
			invalidCase(
				`function handler() {} test("a", handler)`,
				`test("a", handler)`,
			),
			// A function named in the timeout slot is provably not a number.
			invalidCase(
				`function t() {} test("a", () => {}, t)`,
				`test("a", () => {}, t)`,
			),

			// ---- D. Options that are fully readable and carry no timeout ----
			invalidCase(`test("a", { retry: 2 }, () => {})`, `test("a", { retry: 2 }, () => {})`),
			invalidCase(`test("a", {}, () => {})`, `test("a", {}, () => {})`),
			invalidCase(`test("a", { 5: 1 }, () => {})`, `test("a", { 5: 1 }, () => {})`),
			invalidCase(
				`const options = { retry: 2 }; test("a", options, () => {})`,
				`test("a", options, () => {})`,
			),

			// ---- D. Negative timeouts ----
			// Rstest already spells "no limit" as `0`, so a negative value
			// leaves the test without a usable timeout.
			invalidCase(`test("a", () => {}, -1)`, `test("a", () => {}, -1)`),
			invalidCase(`test("a", { timeout: -1 }, () => {})`, `test("a", { timeout: -1 }, () => {})`),
			invalidCase(
				`const t = -1; test("a", () => {}, t)`,
				`test("a", () => {}, t)`,
			),
			// An explicit member after a spread has the last word.
			invalidCase(
				`test("a", { ...options, timeout: -1 }, () => {})`,
				`test("a", { ...options, timeout: -1 }, () => {})`,
			),
			invalidCase(
				`test("a", { timeout: 500, timeout: -1 }, () => {})`,
				`test("a", { timeout: 500, timeout: -1 }, () => {})`,
			),

			// ---- E. A suite timeout the rule refuses on a test ----
			// A negative suite timeout is not a timeout its tests can use.
			invalidCase(
				`describe("s", { timeout: -1 }, () => { test("a", () => {}) })`,
				`test("a", () => {})`,
			),
			invalidCase(
				`describe("s", () => { test("a", () => {}) }, -1)`,
				`test("a", () => {})`,
			),

			// ---- E. Suite inheritance that does not apply ----
			invalidCase(
				`describe("s", () => { test("a", () => {}) })`,
				`test("a", () => {})`,
			),
			invalidCase(
				`describe("s", { retry: 2 }, () => { test("a", () => {}) })`,
				`test("a", () => {})`,
			),
			// The suite is a Vitest one, so its options never reach an Rstest
			// test.
			invalidCase(
				`import { describe } from 'vitest'; describe("s", { timeout: 5000 }, () => { test("a", () => {}) })`,
				`test("a", () => {})`,
			),
			// Every test in the suite is reported, not just the first.
			invalidCase(
				`describe("s", () => { test("a", () => {}); test("b", () => {}) })`,
				`test("a", () => {})`,
				`test("b", () => {})`,
			),
			invalidCase(
				`[1, 2].forEach(() => { test("a", () => {}) })`,
				`test("a", () => {})`,
			),

			// ---- F. Runtime config that does not exempt ----
			invalidCase(`rs.setConfig({}); test("a", () => {})`, `test("a", () => {})`),
			invalidCase(`rs.setConfig({ retry: 2 }); test("a", () => {})`, `test("a", () => {})`),
			invalidCase(
				`rs.setConfig({ testTimeout: -1 }); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`rs.setConfig({ testTimeout: 5000 }); rs.resetConfig(); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`rs.setConfig({ testTimeout: 5000 }); rs["resetConfig"](); test("a", () => {})`,
				`test("a", () => {})`,
			),
			// `rs`, `rstest` and `import.meta.rstest` name one runtime
			// configuration, so a reset on any spelling cancels a `setConfig`
			// written on any other.
			invalidCase(
				`rs.setConfig({ testTimeout: 5000 }); rstest.resetConfig(); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`rstest.setConfig({ testTimeout: 5000 }); import.meta.rstest.resetConfig(); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`import * as core from '@rstest/core'; core.rs.setConfig({ testTimeout: 5000 }); rstest.resetConfig(); test("a", () => {})`,
				`test("a", () => {})`,
			),
			// The utility object is not Rstest's.
			invalidCase(`vi.setConfig({ testTimeout: 5000 }); test("a", () => {})`, `test("a", () => {})`),
			invalidCase(
				`import { rs } from 'vitest'; rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`const rs = { setConfig(config) {} }; rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`import * as other from 'vitest'; other.rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`,
				`test("a", () => {})`,
			),
		},
	)
}

// TestRequireTestTimeoutEditDemand locks in that the diagnostic never changes
// with edit demand. This rule offers neither a fix nor a suggestion: the
// timeout a test should get is the author's call, not the linter's.
func TestRequireTestTimeoutEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`test("a", () => {});
describe("s", () => { it("b", () => {}) });`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     require_test_timeout.RequireTestTimeoutRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return require_test_timeout.RequireTestTimeoutRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	allEdits := run(rule.EditDemandAll)
	for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
	} {
		for index := range allEdits {
			if !reflect.DeepEqual(diagnostics[index], allEdits[index]) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v",
					demand, index, diagnostics[index], allEdits[index])
			}
		}
	}
	for _, diagnostic := range allEdits {
		if diagnostic.FixesPtr != nil {
			t.Fatal("require-test-timeout unexpectedly materialized fixes")
		}
		if diagnostic.Suggestions != nil {
			t.Fatal("require-test-timeout unexpectedly materialized suggestions")
		}
	}
}

// `rstack/test` re-exports the Rstest core API, so registrations arrive through
// it and so does the `rs.setConfig` exemption that reads a project-wide
// timeout.
func TestRequireTestTimeoutRstackTestModule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_test_timeout.RequireTestTimeoutRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { test } from 'rstack/test'; test("a", () => {}, 500)`},
			{Code: `import { it as check } from 'rstack/test'; check("a", () => {}, 500)`},
			{Code: `import * as core from 'rstack/test'; core.test("a", () => {}, 500)`},
			{Code: `const { test } = require('rstack/test'); test("a", () => {}, 500)`},
			// The runtime config object sets the timeout for the whole file.
			{Code: `import { rs } from 'rstack/test'; rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `import { rs as runtime } from 'rstack/test'; runtime.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `import * as core from 'rstack/test'; core.rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `const core = require('rstack/test'); core.rstest.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
			{Code: `const { rs } = require('rstack/test'); rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase(
				`import { test } from 'rstack/test'; test("a", () => {})`,
				`test("a", () => {})`,
			),
			invalidCase(
				`import { it as check } from 'rstack/test'; check("a", () => {})`,
				`check("a", () => {})`,
			),
			invalidCase(
				`import * as core from 'rstack/test'; core.test("a", () => {})`,
				`core.test("a", () => {})`,
			),
			invalidCase(
				`const { test } = require('rstack/test'); test("a", () => {})`,
				`test("a", () => {})`,
			),
			// A sibling subpath exports no runtime config object, so it cannot
			// exempt the registration.
			invalidCase(
				`import { rs } from 'rstack/lib'; rs.setConfig({ testTimeout: 5000 }); test("a", () => {})`,
				`test("a", () => {})`,
			),
		},
	)
}
