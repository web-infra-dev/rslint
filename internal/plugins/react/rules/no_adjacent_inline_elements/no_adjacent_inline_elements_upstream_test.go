// TestNoAdjacentInlineElementsUpstream migrates the full valid/invalid suite
// from upstream tests/lib/rules/no-adjacent-inline-elements.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in the no_adjacent_inline_elements_extras_test.go file.
package no_adjacent_inline_elements

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func inlineElementError(line, column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "inlineElement",
		Message:   inlineElementMessage,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func TestNoAdjacentInlineElementsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAdjacentInlineElementsRule, []rule_tester.ValidTestCase{
		// ---- valid ----
		{Code: `<div />;`, Tsx: true},
		{Code: `<div><div></div><div></div></div>;`, Tsx: true},
		{Code: `<div><p></p><div></div></div>;`, Tsx: true},
		{Code: `<div><p></p><a></a></div>;`, Tsx: true},
		{Code: `<div><a></a>&nbsp;<a></a></div>;`, Tsx: true},
		{Code: `<div><a></a>&nbsp;some text &nbsp; <a></a></div>;`, Tsx: true},
		{Code: `<div><a></a>&nbsp;some text <a></a></div>;`, Tsx: true},
		{Code: `<div><a></a> <a></a></div>;`, Tsx: true},
		{Code: `<div><ul><li><a></a></li><li><a></a></li></ul></div>;`, Tsx: true},
		{Code: `<div><a></a> some text <a></a></div>;`, Tsx: true},
		{Code: `React.createElement("div", null, "some text");`, Tsx: true},
		{Code: `React.createElement("div", undefined, [React.createElement("a"), " some text ", React.createElement("a")]);`, Tsx: true},
		{Code: `React.createElement("div", undefined, [React.createElement("a"), " ", React.createElement("a")]);`, Tsx: true},
		{Code: `React.createElement(a, b);`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- invalid ----
		{Code: `<div><a></a><a></a></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 26)}},
		{Code: `<div><a></a><span></span></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 32)}},
		{Code: `React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 95)}},
	})
}
