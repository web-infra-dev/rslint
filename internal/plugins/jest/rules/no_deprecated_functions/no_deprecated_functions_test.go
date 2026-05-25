package no_deprecated_functions_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_deprecated_functions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func jestSettings(version string) map[string]interface{} {
	return map[string]interface{}{
		"jest": map[string]interface{}{
			"version": version,
		},
	}
}

func TestNoDeprecatedFunctionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_deprecated_functions.NoDeprecatedFunctionsRule,
		[]rule_tester.ValidTestCase{
			{Code: `jest`, Settings: jestSettings("14")},
			{Code: `require("fs")`, Settings: jestSettings("14")},
			{Code: `jest.resetModuleRegistry`, Settings: jestSettings("14")},
			{Code: `require.requireActual`, Settings: jestSettings("17")},
			{Code: `jest.genMockFromModule`, Settings: jestSettings("25")},
			{Code: `jest.genMockFromModule`, Settings: jestSettings("25.1.1")},
			{Code: `require.requireActual`, Settings: jestSettings("17.2")},
		},
		[]rule_tester.InvalidTestCase{
			// jest.resetModuleRegistry -> jest.resetModules (Jest 21)
			{
				Code:     `jest.resetModuleRegistry()`,
				Output:   []string{`jest.resetModules()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `jest['resetModuleRegistry']()`,
				Output:   []string{`jest['resetModules']()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			// Double-call form: only the inner deprecated callee should be
			// reported and rewritten; the outer call's callee is itself a
			// CallExpression and must be ignored to preserve the chained-call
			// semantics.
			{
				Code:     `jest.resetModuleRegistry()()`,
				Output:   []string{`jest.resetModules()()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			// jest.addMatchers -> expect.extend (Jest 24)
			{
				Code:     `jest.addMatchers()`,
				Output:   []string{`expect.extend()`},
				Settings: jestSettings("24"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `jest['addMatchers']()`,
				Output:   []string{`expect['extend']()`},
				Settings: jestSettings("24"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			// require.requireMock -> jest.requireMock (Jest 21)
			{
				Code:     `require.requireMock()`,
				Output:   []string{`jest.requireMock()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `require['requireMock']()`,
				Output:   []string{`jest['requireMock']()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			// require.requireActual -> jest.requireActual (Jest 21)
			{
				Code:     `require.requireActual()`,
				Output:   []string{`jest.requireActual()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `require['requireActual']()`,
				Output:   []string{`jest['requireActual']()`},
				Settings: jestSettings("21"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			// jest.runTimersToTime -> jest.advanceTimersByTime (Jest 22)
			{
				Code:     `jest.runTimersToTime()`,
				Output:   []string{`jest.advanceTimersByTime()`},
				Settings: jestSettings("22"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `jest['runTimersToTime']()`,
				Output:   []string{`jest['advanceTimersByTime']()`},
				Settings: jestSettings("22"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			// jest.genMockFromModule -> jest.createMockFromModule (Jest 26)
			{
				Code:     `jest.genMockFromModule()`,
				Output:   []string{`jest.createMockFromModule()`},
				Settings: jestSettings("26"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `jest['genMockFromModule']()`,
				Output:   []string{`jest['createMockFromModule']()`},
				Settings: jestSettings("26"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `jest.genMockFromModule()`,
				Output:   []string{`jest.createMockFromModule()`},
				Settings: jestSettings("26.0.0-next.11"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
			{
				Code:     `jest['genMockFromModule']()`,
				Output:   []string{`jest['createMockFromModule']()`},
				Settings: jestSettings("26.0.0-next.11"),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecatedFunction"}},
			},
		},
	)
}
