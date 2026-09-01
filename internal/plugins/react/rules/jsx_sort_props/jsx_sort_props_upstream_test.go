// TestJsxSortPropsUpstream migrates the complete eslint-plugin-react v7.37.5
// suite from tests/lib/rules/jsx-sort-props.js. Rslint-specific edge cases live
// in the sibling jsx_sort_props_extras_test.go file.
package jsx_sort_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func jsxSortError(id, message string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: id, Message: message, Line: line, Column: column}
}

func TestJsxSortPropsUpstream(t *testing.T) {
	raw, err := rule_tester.LoadESLintTestSuiteFromJSON("testdata/jsx_sort_props_upstream.json")
	if err != nil {
		t.Fatal(err)
	}
	suite := rule_tester.ConvertESLintTestSuite(raw)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxSortPropsRule, suite.Valid, suite.Invalid)
}
