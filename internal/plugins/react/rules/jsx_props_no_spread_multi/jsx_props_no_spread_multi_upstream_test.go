// TestJsxPropsNoSpreadMultiUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/jsx-props-no-spread-multi.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in the jsx_props_no_spread_multi_extras_test.go file.
package jsx_props_no_spread_multi

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noMultiSpreadingMessage = "Spreading the same expression multiple times is forbidden"

func noMultiSpreadingError(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noMultiSpreading",
		Line:      line,
		Column:    column,
	}
}

func TestJsxPropsNoSpreadMultiUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPropsNoSpreadMultiRule, []rule_tester.ValidTestCase{
		// ---- valid ----
		{Code: `const a = {}; <App {...a} />`, Tsx: true},
		{Code: `const a = {}; const b = {}; <App {...a} {...b} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- invalid ----
		{
			Code: `const props = {}; <App {...props} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 35),
			},
		},
		{
			Code: `const props = {}; <div {...props} a="a" {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 41),
			},
		},
		{
			Code: `const props = {}; <div {...props} {...props} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 35),
				noMultiSpreadingError(1, 46),
			},
		},
	})
}
