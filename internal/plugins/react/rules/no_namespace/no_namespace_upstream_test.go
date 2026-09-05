package no_namespace

// TestNoNamespaceUpstream migrates eslint-plugin-react's
// tests/lib/rules/no-namespace.js from v7.37.5. Position assertions cover the
// complete reported JSX opening element or createElement call. rslint-specific
// cases live in no_namespace_extras_test.go.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamespaceUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `<testcomponent />`, Tsx: true},
		{Code: `React.createElement("testcomponent")`, Tsx: true},
		{Code: `<testComponent />`, Tsx: true},
		{Code: `React.createElement("testComponent")`, Tsx: true},
		{Code: `<test_component />`, Tsx: true},
		{Code: `React.createElement("test_component")`, Tsx: true},
		{Code: `<TestComponent />`, Tsx: true},
		{Code: `React.createElement("TestComponent")`, Tsx: true},
		{Code: `<object.testcomponent />`, Tsx: true},
		{Code: `React.createElement("object.testcomponent")`, Tsx: true},
		{Code: `<object.testComponent />`, Tsx: true},
		{Code: `React.createElement("object.testComponent")`, Tsx: true},
		{Code: `<object.test_component />`, Tsx: true},
		{Code: `React.createElement("object.test_component")`, Tsx: true},
		{Code: `<object.TestComponent />`, Tsx: true},
		{Code: `React.createElement("object.TestComponent")`, Tsx: true},
		{Code: `<Object.testcomponent />`, Tsx: true},
		{Code: `React.createElement("Object.testcomponent")`, Tsx: true},
		{Code: `<Object.testComponent />`, Tsx: true},
		{Code: `React.createElement("Object.testComponent")`, Tsx: true},
		{Code: `<Object.test_component />`, Tsx: true},
		{Code: `React.createElement("Object.test_component")`, Tsx: true},
		{Code: `<Object.TestComponent />`, Tsx: true},
		{Code: `React.createElement("Object.TestComponent")`, Tsx: true},
		{Code: `React.createElement(null)`, Tsx: true},
		{Code: `React.createElement(true)`, Tsx: true},
		{Code: `React.createElement({})`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		{
			Code: `<ns:testcomponent />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:testcomponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		{
			Code: `React.createElement("ns:testcomponent")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:testcomponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 40,
			}},
		},
		{
			Code: `<ns:testComponent />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:testComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		{
			Code: `React.createElement("ns:testComponent")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:testComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 40,
			}},
		},
		{
			Code: `<ns:test_component />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:test_component must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 22,
			}},
		},
		{
			Code: `React.createElement("ns:test_component")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:test_component must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 41,
			}},
		},
		{
			Code: `<ns:TestComponent />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:TestComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		{
			Code: `React.createElement("ns:TestComponent")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:TestComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 40,
			}},
		},
		{
			Code: `<Ns:testcomponent />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:testcomponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		{
			Code: `React.createElement("Ns:testcomponent")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:testcomponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 40,
			}},
		},
		{
			Code: `<Ns:testComponent />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:testComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		{
			Code: `React.createElement("Ns:testComponent")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:testComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 40,
			}},
		},
		{
			Code: `<Ns:test_component />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:test_component must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 22,
			}},
		},
		{
			Code: `React.createElement("Ns:test_component")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:test_component must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 41,
			}},
		},
		{
			Code: `<Ns:TestComponent />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:TestComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		{
			Code: `React.createElement("Ns:TestComponent")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:TestComponent must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 40,
			}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNamespaceRule, valid, invalid)
}
