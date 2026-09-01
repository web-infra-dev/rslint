package utils_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// hookParseProbe reports every parsed registration call with its kind and
// resolved name, so invalid test cases assert exact parse results and valid
// test cases assert that the parser returned nil.
var hookParseProbe = rule.Rule{
	Name:             "rstest/hook-parse-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("parsedFn", fmt.Sprintf(
					"kind=%s name=%s", parsed.Kind, parsed.Name,
				)))
			},
		}
	},
}

func parsedHookError(kind string, name string) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "parsedFn",
		Message:   fmt.Sprintf("kind=%s name=%s", kind, name),
	}}
}

func TestParseRstestFnCallHooks(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &hookParseProbe,
		[]rule_tester.ValidTestCase{
			// onTestFinished / onTestFailed are execution-time APIs, not hooks.
			{Code: `onTestFinished(() => {});`},
			{Code: `onTestFailed(() => {});`},
			{Code: `import { onTestFinished } from '@rstest/core'; onTestFinished(() => {});`},
			{Code: `import { onTestFailed } from '@rstest/core'; onTestFailed(() => {});`},
			// Hooks accept no chained members.
			{Code: `beforeAll.skip(() => {});`},
			{Code: `beforeEach.each([1])(() => {});`},
			// Hooks on the test object are Playwright-only, and only while the
			// receiver is still a PlaywrightTest.
			{Code: `import { test } from '@rstest/core'; test.beforeEach(() => {});`},
			{Code: `import { test } from '@rstest/playwright'; test.skip.beforeEach(() => {});`},
			{Code: `import { test } from '@rstest/playwright'; test.beforeEach.only(() => {});`},
			// Foreign frameworks and shadowed locals must not resolve.
			{Code: `import { beforeAll } from 'vitest'; beforeAll(() => {});`},
			{Code: `import { beforeEach } from '@jest/globals'; beforeEach(() => {});`},
			{Code: `const beforeAll = createHookRegistry(); beforeAll(() => {});`},
			// Only comma expressions are transparent callee wrappers.
			{Code: `(flag && test)("case", () => {});`},
		},
		[]rule_tester.InvalidTestCase{
			// All four hooks as globals.
			{Code: `beforeAll(() => {});`, Errors: parsedHookError("hook", "beforeAll")},
			{Code: `beforeEach(() => {});`, Errors: parsedHookError("hook", "beforeEach")},
			{Code: `afterEach(() => {});`, Errors: parsedHookError("hook", "afterEach")},
			{Code: `afterAll(() => {});`, Errors: parsedHookError("hook", "afterAll")},
			// With the optional timeout argument.
			{Code: `beforeAll(() => {}, 1000);`, Errors: parsedHookError("hook", "beforeAll")},
			// Source forms.
			{
				Code:   `import { beforeAll } from '@rstest/core'; beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `import { beforeEach as setup } from '@rstest/core'; setup(() => {});`,
				Errors: parsedHookError("hook", "beforeEach"),
			},
			{
				Code:   `import * as rstest from '@rstest/core'; rstest.afterEach(() => {});`,
				Errors: parsedHookError("hook", "afterEach"),
			},
			{
				Code:   `import.meta.rstest.beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `const { afterAll } = import.meta.rstest; afterAll(() => {});`,
				Errors: parsedHookError("hook", "afterAll"),
			},
			{
				Code:   `const setup = beforeAll; setup(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `import { beforeAll } from '@rstest/playwright'; beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			// PlaywrightTest exposes the four hooks as members, and extend()
			// returns another PlaywrightTest.
			{
				Code:   `import { test } from '@rstest/playwright'; test.beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `import { test } from '@rstest/playwright'; test.beforeEach(() => {});`,
				Errors: parsedHookError("hook", "beforeEach"),
			},
			{
				Code:   `import { test } from '@rstest/playwright'; test.afterEach(() => {});`,
				Errors: parsedHookError("hook", "afterEach"),
			},
			{
				Code:   `import { test } from '@rstest/playwright'; test.afterAll(() => {});`,
				Errors: parsedHookError("hook", "afterAll"),
			},
			{
				Code:   `import { test } from '@rstest/playwright'; test.extend({}).beforeEach(() => {});`,
				Errors: parsedHookError("hook", "beforeEach"),
			},
			// Test and describe parsing is unchanged.
			{Code: `test("case", () => {});`, Errors: parsedHookError("test", "test")},
			{Code: `describe("suite", () => {});`, Errors: parsedHookError("describe", "describe")},
			// A comma expression evaluates to its right operand before the call.
			{Code: `(0, test)("case", () => {});`, Errors: parsedHookError("test", "test")},
			{Code: `(sideEffect(), describe.only)("suite", () => {});`, Errors: parsedHookError("describe", "describe")},
			{Code: `(0, beforeEach)(() => {});`, Errors: parsedHookError("hook", "beforeEach")},
			{Code: `(0, import.meta.rstest.test)("case", () => {});`, Errors: parsedHookError("test", "test")},
			{Code: `const wrapped = (0, test); wrapped("case", () => {});`, Errors: parsedHookError("test", "test")},
		},
	)
}

// hookKindProbe reports the semantic kind returned by the shared call analysis.
var hookKindProbe = rule.Rule{
	Name:             "rstest/hook-kind-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("kinds", fmt.Sprintf(
					"kind=%s",
					parsed.Kind,
				)))
			},
		}
	},
}

func TestRstestCallAnalysisFnKinds(t *testing.T) {
	kindsError := func(message string) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{{MessageId: "kinds", Message: message}}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &hookKindProbe,
		[]rule_tester.ValidTestCase{
			{Code: `notAFnCall(() => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `beforeEach(() => {});`,
				Errors: kindsError("kind=hook"),
			},
			{
				Code:   `import { test } from '@rstest/playwright'; test.beforeEach(() => {});`,
				Errors: kindsError("kind=hook"),
			},
			{
				Code:   `test("case", () => {});`,
				Errors: kindsError("kind=test"),
			},
			{
				Code:   `describe("suite", () => {});`,
				Errors: kindsError("kind=describe"),
			},
		},
	)
}
