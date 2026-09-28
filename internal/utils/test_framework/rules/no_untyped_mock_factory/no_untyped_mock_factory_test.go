package no_untyped_mock_factory

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFixRequiresSignatureAcceptingOneTypeArgument(t *testing.T) {
	testRule := NewRule(Config{
		Name: "test/no-untyped-mock-factory",
		Candidates: func(rule.RuleContext) func(*ast.Node) bool {
			return func(*ast.Node) bool { return true }
		},
		CanFixCallee: func(rule.RuleContext, *ast.Node) bool { return true },
	})
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&testRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code:   `declare const api: { mock<T, U>(path: string, factory: () => T): void }; api.mock('./async-mock-factories', () => ({}));`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}},
			},
			{
				Code:   `declare const api: { mock<T, U = unknown>(path: string, factory: () => T): void }; api.mock('./async-mock-factories', () => ({}));`,
				Output: []string{`declare const api: { mock<T, U = unknown>(path: string, factory: () => T): void }; api.mock<typeof import('./async-mock-factories')>('./async-mock-factories', () => ({}));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./async-mock-factories')`"}},
			},
			{
				Code:   `declare const api: { mock<T = unknown, U = unknown>(path: string, factory: () => T): void }; api.mock('./async-mock-factories', () => ({}));`,
				Output: []string{`declare const api: { mock<T = unknown, U = unknown>(path: string, factory: () => T): void }; api.mock<typeof import('./async-mock-factories')>('./async-mock-factories', () => ({}));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}},
			},
		},
	)
}
