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
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestFnCallWithOfficialExtensions(node, ctx)
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
			// Foreign frameworks and shadowed locals must not resolve.
			{Code: `import { beforeAll } from 'vitest'; beforeAll(() => {});`},
			{Code: `import { beforeEach } from '@jest/globals'; beforeEach(() => {});`},
			{Code: `const beforeAll = createHookRegistry(); beforeAll(() => {});`},
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
			// Test and describe parsing is unchanged.
			{Code: `test("case", () => {});`, Errors: parsedHookError("test", "test")},
			{Code: `describe("suite", () => {});`, Errors: parsedHookError("describe", "describe")},
		},
	)
}

// hookKindProbe reports IsTypeOfRstestFnCall results for every parsed
// registration call: single-kind, multi-kind variadic, and the zero-kind
// degenerate form.
var hookKindProbe = rule.Rule{
	Name:             "rstest/hook-kind-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if rstestUtils.ParseRstestFnCallWithOfficialExtensions(node, ctx) == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("kinds", fmt.Sprintf(
					"hook=%t test=%t testOrHook=%t zeroKinds=%t",
					rstestUtils.IsTypeOfRstestFnCall(node, ctx, rstestUtils.RstestFnTypeHook),
					rstestUtils.IsTypeOfRstestFnCall(node, ctx, rstestUtils.RstestFnTypeTest),
					rstestUtils.IsTypeOfRstestFnCall(node, ctx,
						rstestUtils.RstestFnTypeTest, rstestUtils.RstestFnTypeHook),
					rstestUtils.IsTypeOfRstestFnCall(node, ctx),
				)))
			},
		}
	},
}

func TestIsTypeOfRstestFnCall(t *testing.T) {
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
				Errors: kindsError("hook=true test=false testOrHook=true zeroKinds=false"),
			},
			{
				Code:   `test("case", () => {});`,
				Errors: kindsError("hook=false test=true testOrHook=true zeroKinds=false"),
			},
			{
				Code:   `describe("suite", () => {});`,
				Errors: kindsError("hook=false test=false testOrHook=false zeroKinds=false"),
			},
		},
	)
}
