// TestIframeMissingSandboxUpstream migrates every valid and invalid case from
// eslint-plugin-react v7.37.5's tests/lib/rules/iframe-missing-sandbox.js.
// The documentation examples are also retained below. tsgo-specific coverage
// lives in iframe_missing_sandbox_extras_test.go.
package iframe_missing_sandbox

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIframeMissingSandboxUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &IframeMissingSandboxRule, []rule_tester.ValidTestCase{
		{Code: `<div sandbox="__unknown__" />;`, Tsx: true},

		{Code: `<iframe sandbox="" />;`, Tsx: true},
		{Code: `<iframe sandbox={""} />`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "" });`, Tsx: true},

		{Code: `<iframe src="foo.htm" sandbox></iframe>`, Tsx: true},
		{Code: `React.createElement("iframe", { src: "foo.htm", sandbox: true })`, Tsx: true},

		{Code: `<iframe src="foo.htm" sandbox sandbox></iframe>`, Tsx: true},

		{Code: `<iframe sandbox="allow-forms"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-modals"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-orientation-lock"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-pointer-lock"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-popups"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-popups-to-escape-sandbox"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-presentation"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-same-origin"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-scripts"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-top-navigation"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-top-navigation-by-user-activation"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-forms allow-modals"></iframe>`, Tsx: true},
		{Code: `<iframe sandbox="allow-popups allow-popups-to-escape-sandbox allow-pointer-lock allow-same-origin allow-top-navigation"></iframe>`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-forms" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-modals" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-orientation-lock" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-pointer-lock" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-popups" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-popups-to-escape-sandbox" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-presentation" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-same-origin" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-scripts" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-top-navigation" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-top-navigation-by-user-activation" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-forms allow-modals" })`, Tsx: true},
		{Code: `React.createElement("iframe", { sandbox: "allow-popups allow-popups-to-escape-sandbox allow-pointer-lock allow-same-origin allow-top-navigation" })`, Tsx: true},

		// Documentation's non-warning examples.
		{Code: `var React = require('react'); var Frame = <iframe sandbox="allow-popups"/>;`, Tsx: true},
		{Code: `var React = require('react'); var Frame = () => (<div><iframe sandbox="allow-popups"></iframe>{React.createElement('iframe', { sandbox: "allow-popups" })}</div>);`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `<iframe></iframe>;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 9}},
		},
		{
			Code: `<iframe/>;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
		},
		{
			Code: `React.createElement("iframe");`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 30}},
		},
		{
			Code: `React.createElement("iframe", {});`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 34}},
		},
		{
			Code: `React.createElement("iframe", null);`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 36}},
		},

		// Documentation's warning examples.
		{
			Code: `var React = require('react'); var Frame = () => (<div><iframe></iframe>{React.createElement('iframe')}</div>);`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 55, EndLine: 1, EndColumn: 63},
				{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 73, EndLine: 1, EndColumn: 102},
			},
		},
		{
			Code: `<iframe sandbox="__unknown__"></iframe>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 31}},
		},
		{
			Code: `React.createElement("iframe", { sandbox: "__unknown__" })`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 58}},
		},
		{
			Code: `<iframe sandbox="allow-popups __unknown__"/>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 45}},
		},
		{
			Code: `<iframe sandbox="__unknown__ allow-popups"/>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 45}},
		},
		{
			Code: `<iframe sandbox=" allow-forms __unknown__ allow-popups __unknown__  "/>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 72},
				{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 72},
			},
		},
		{
			Code: `<iframe sandbox="allow-scripts allow-same-origin"></iframe>;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidCombination", Message: msgInvalidCombination, Line: 1, Column: 1, EndLine: 1, EndColumn: 51}},
		},
		{
			Code: `<iframe sandbox="allow-same-origin allow-scripts"/>;`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidCombination", Message: msgInvalidCombination, Line: 1, Column: 1, EndLine: 1, EndColumn: 52}},
		},
	})
}
