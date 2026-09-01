// TestJsxSortPropsUpstream migrates representative valid and invalid cases
// from eslint-plugin-react v7.37.5. Rslint-specific edge cases live in the
// sibling jsx_sort_props_extras_test.go file.
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxSortPropsRule, []rule_tester.ValidTestCase{
		{Code: `<App />`, Tsx: true},
		{Code: `<App {...this.props} a b c />`, Tsx: true},
		{Code: `<App c {...this.props} a b />`, Tsx: true},
		{Code: `<App A a />`, Tsx: true},
		{Code: `<App a b="b" />`, Tsx: true, Options: map[string]any{"shorthandFirst": true}},
		{Code: `<App a="a" b="b" x y />`, Tsx: true, Options: map[string]any{"shorthandLast": true}},
		{Code: `<App a z onBar onFoo />`, Tsx: true, Options: map[string]any{"callbacksLast": true}},
		{Code: `<App a A />`, Tsx: true, Options: map[string]any{"ignoreCase": true}},
		{Code: `<App b a />`, Tsx: true, Options: map[string]any{"noSortAlphabetically": true}},
		{Code: `<App children={<App />} key={0} ref="r" a b />`, Tsx: true, Options: map[string]any{"reservedFirst": true}},
		{Code: `<div dangerouslySetInnerHTML={{ __html: "x" }} key={0} a />`, Tsx: true, Options: map[string]any{"reservedFirst": true}},
	}, []rule_tester.InvalidTestCase{
		{Code: `<App b a />`, Tsx: true, Output: []string{`<App a b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 8)}},
		{Code: `<App aa aB />`, Tsx: true, Output: []string{`<App aB aa />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 9)}},
		{Code: `<App B a />`, Tsx: true, Options: map[string]any{"ignoreCase": true}, Output: []string{`<App a B />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 8)}},
		{Code: `<App {...props} b a />`, Tsx: true, Output: []string{`<App {...props} a b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 19)}},
		{Code: `<App a onBar onFoo z />`, Tsx: true, Options: map[string]any{"callbacksLast": true}, Output: []string{`<App a z onBar onFoo />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listCallbacksLast", "Callbacks must be listed after all other props", 1, 14)}},
		{Code: `<App a="a" b />`, Tsx: true, Options: map[string]any{"shorthandFirst": true}, Output: []string{`<App b a="a" />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listShorthandFirst", "Shorthand props must be listed before all other props", 1, 12)}},
		{Code: `<App b a="a" />`, Tsx: true, Options: map[string]any{"shorthandLast": true}, Output: []string{`<App a="a" b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listShorthandLast", "Shorthand props must be listed after all other props", 1, 6)}},
		{Code: `<App a key={1} />`, Tsx: true, Options: map[string]any{"reservedFirst": true}, Output: []string{`<App key={1} a />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listReservedPropsFirst", "Reserved props must be listed before all other props", 1, 8)}},
		{Code: `<div a dangerouslySetInnerHTML={{ __html: "x" }} />`, Tsx: true, Options: map[string]any{"reservedFirst": true}, Output: []string{`<div dangerouslySetInnerHTML={{ __html: "x" }} a />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listReservedPropsFirst", "Reserved props must be listed before all other props", 1, 8)}},
		{Code: `<App key={4} />`, Tsx: true, Options: map[string]any{"reservedFirst": []any{}}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listIsEmpty", "A customized reserved first list must not be empty", 1, 6)}},
		{Code: `<App key={5} />`, Tsx: true, Options: map[string]any{"reservedFirst": []any{"notReserved"}}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("noUnreservedProps", "A customized reserved first list must only contain a subset of React reserved props. Remove: notReserved", 1, 6)}},
		{Code: "<App\n  a\n  b={{\n    bB: 1,\n  }}\n/>", Tsx: true, Options: map[string]any{"multiline": "first"}, Output: []string{"<App\n  b={{\n    bB: 1,\n  }}\n  a\n/>"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listMultilineFirst", "Multiline props must be listed before all other props", 3, 3)}},
		{Code: "<App\n  a={{\n    aA: 1,\n  }}\n  b\n/>", Tsx: true, Options: map[string]any{"multiline": "last"}, Output: []string{"<App\n  b\n  a={{\n    aA: 1,\n  }}\n/>"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listMultilineLast", "Multiline props must be listed after all other props", 2, 3)}},
	})
}
