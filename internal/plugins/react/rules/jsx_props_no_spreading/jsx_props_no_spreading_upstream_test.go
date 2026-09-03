// TestJsxPropsNoSpreadingUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/jsx-props-no-spreading.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in the jsx_props_no_spreading_extras_test.go file.
package jsx_props_no_spreading

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noSpreadingMessage = "Prop spreading is forbidden"

func noSpreadingError(line, column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noSpreading",
		Message:   noSpreadingMessage,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func TestJsxPropsNoSpreadingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPropsNoSpreadingRule, []rule_tester.ValidTestCase{
		// ---- valid ----
		{Code: `<App one_prop={one_prop} two_prop={two_prop} />`, Tsx: true},
		{Code: `<div one_prop={one_prop} two_prop={two_prop}></div>`, Tsx: true},
		{Code: `const newProps = {...props}; <App one_prop={newProps.one_prop} two_prop={newProps.two_prop} style={{...styles}} />`, Tsx: true},
		{Code: `<App><Image {...props} /><img {...props} /></App>`, Tsx: true, Options: map[string]any{"exceptions": []any{"Image", "img"}}},
		{Code: `<App><Image {...props} /><img src={src} alt={alt} /></App>`, Tsx: true, Options: map[string]any{"custom": "ignore"}},
		{Code: `<App><Image {...props} /><img {...props} /></App>`, Tsx: true, Options: map[string]any{"custom": "enforce", "html": "ignore", "exceptions": []any{"Image"}}},
		{Code: `<App><img {...props} /><Image src={src} alt={alt} /><div {...someOtherProps} /></App>`, Tsx: true, Options: map[string]any{"html": "ignore"}},
		{Code: `<App><Foo {...{ prop1, prop2, prop3 }} /></App>`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},
		{Code: `<App><components.Group {...props} /><Nav.Item {...props} /></App>`, Tsx: true, Options: map[string]any{"exceptions": []any{"components.Group", "Nav.Item"}}},
		{Code: `<App><components.Group {...props} /><Nav.Item {...props} /></App>`, Tsx: true, Options: map[string]any{"custom": "ignore"}},
		{Code: `<App><components.Group {...props} /><Nav.Item {...props} /></App>`, Tsx: true, Options: map[string]any{"custom": "enforce", "html": "ignore", "exceptions": []any{"components.Group", "Nav.Item"}}},
	}, []rule_tester.InvalidTestCase{
		// ---- invalid ----
		{Code: `<App {...props} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 16)}},
		{Code: `<div {...props}></div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 16)}},
		{Code: `<App {...props} some_other_prop={some_other_prop} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 16)}},
		{Code: `<App><Image {...props} /><span {...props} /></App>`, Tsx: true, Options: map[string]any{"exceptions": []any{"Image", "img"}}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 32, 42)}},
		{Code: `<App><Image {...props} /><img {...props} /></App>`, Tsx: true, Options: map[string]any{"custom": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 31, 41)}},
		{Code: `<App><Image {...props} /><img {...props} /></App>`, Tsx: true, Options: map[string]any{"html": "ignore", "exceptions": []any{"Image", "img"}}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 31, 41)}},
		{Code: `<App><Image {...props} /><img {...props} /><div {...props} /></App>`, Tsx: true, Options: map[string]any{"custom": "ignore", "html": "ignore", "exceptions": []any{"Image", "img"}}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 13, 23), noSpreadingError(1, 31, 41)}},
		{Code: `<App><img {...props} /><Image {...props} /></App>`, Tsx: true, Options: map[string]any{"html": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 31, 41)}},
		{Code: `<App><Foo {...{ prop1, prop2, prop3 }} /></App>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 11, 39)}},
		{Code: `<App><Foo {...{ prop1, ...rest }} /></App>`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 11, 34)}},
		{Code: `<App><Foo {...{ ...props }} /></App>`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 11, 28)}},
		{Code: `<App><Foo {...props} /></App>`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 11, 21)}},
		{Code: `<App><components.Group {...props} /><Nav.Item {...props} /></App>`, Tsx: true, Options: map[string]any{"exceptions": []any{"components.DropdownIndicator", "Nav.Item"}}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 24, 34)}},
	})
}
