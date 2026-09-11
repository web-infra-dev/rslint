package iframe_missing_sandbox

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIframeMissingSandboxExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &IframeMissingSandboxRule, []rule_tester.ValidTestCase{
		// Tokens supported upstream but not exercised by its v7.37.5 suite.
		{Code: `<iframe sandbox="allow-downloads allow-downloads-without-user-activation allow-storage-access-by-user-activation" />`, Tsx: true},
		// Direct JSX attributes decode HTML entities before the rule sees their
		// ESTree Literal value; entities can form both a token and a delimiter.
		{Code: `<iframe sandbox="allow&#x2D;forms" />`, Tsx: true},
		{Code: `<iframe sandbox="allow-forms&#x20;allow-modals" />`, Tsx: true},
		// Upstream validates only direct Literal values; expressions remain dynamic.
		{Code: `<iframe sandbox={"__unknown__"} />`, Tsx: true},
		{Code: `React.createElement("iframe", {sandbox: value})`, Tsx: true},
		// The first matching property wins, matching Array.prototype.find upstream.
		{Code: `React.createElement("iframe", {sandbox: "allow-forms", sandbox: "__unknown__"})`, Tsx: true},
		// A destructured React factory is supported by upstream's isCreateElement.
		{Code: `import { createElement } from "react"; createElement("iframe", {sandbox: "allow-forms"})`, Tsx: true},
		// Upstream reads MemberExpression.property.name, so an identifier in a
		// computed access is still recognized as createElement.
		{Code: `React[createElement]("iframe", {sandbox: "allow-forms"})`, Tsx: true},
		{Code: `React?.[createElement]("iframe", {sandbox: "allow-forms"})`, Tsx: true},
		// Configured pragma support comes from the shared upstream-equivalent matcher.
		{Code: `h.createElement("iframe", {sandbox: "allow-forms"})`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `React.createElement("iframe", {"sandbox": "allow-forms"})`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 58}},
		},
		{
			Code: `React.createElement("iframe", {[sandbox]: "__unknown__"})`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidValue", Message: `An iframe element defines a sandbox attribute with invalid value "__unknown__"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 58}},
		},
		{
			Code: `import { createElement } from "react"; createElement("iframe")`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 40, EndLine: 1, EndColumn: 63}},
		},
		{
			Code: `React[createElement]("iframe")`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 31}},
		},
		{
			Code: `React?.[createElement]("iframe")`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 33}},
		},
		{
			Code: `h.createElement("iframe")`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "attributeMissing", Message: msgAttributeMissing, Line: 1, Column: 1, EndLine: 1, EndColumn: 26}},
		},
	})
}
