// TestNoAdjacentInlineElementsExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / real-user shape
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
// N/A: key forms, declaration containers, function containers, rest/spread
// bindings, and overload/abstract/declare members are outside this child-list
// rule's syntax surface.
package no_adjacent_inline_elements

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAdjacentInlineElementsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAdjacentInlineElementsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: paired and self-closing inline elements ----
		{Code: `<div><Foo /><a /></div>;`, Tsx: true},
		// ---- Dimension 4: non-simple tag names are not inline matches ----
		{Code: `<div><Components.A /><Components.B /></div>;`, Tsx: true},
		{Code: `<div><svg:a /><svg:b /></div>;`, Tsx: true},
		// ---- Dimension 4: literal boundary whitespace ----
		{Code: `React.createElement("div", null, [React.createElement("a"), "\u00a0", React.createElement("a")]);`, Tsx: true},
		// ---- Dimension 4: nested traversal boundary ----
		{Code: `<><a /><a /></>;`, Tsx: true},
		{Code: `<div><a />{/* comment */}<a /></div>;`, Tsx: true},
		// ---- Dimension 4: parenthesized and optional-chain pragma ----
		{Code: `(React).createElement("div", null, [React.createElement("a"), " ", React.createElement("a")]);`, Tsx: true},
		{Code: `React?.createElement("div", null, [React.createElement("a"), " ", React.createElement("a")]);`, Tsx: true},
		{Code: `(React as any).createElement("div", null, [React.createElement("a"), React.createElement("a")]);`, Tsx: true},
		// N/A: receiver type wrappers are deliberately not recognized by the
		// upstream isCreateElement predicate; this case locks that behavior.
		// ---- Real-user: issue #2620, one React.createElement child ----
		{Code: `React.createElement("div", null, "Hello");`, Tsx: true},
		// ---- Real-user: issue #2575, React.createElement without children ----
		{Code: `React.createElement(a, b);`, Tsx: true},
		// Locks in upstream isInline() fallback branch: a call with no arguments
		// is not an inline child and must not crash the linter.
		{Code: `React.createElement("div", null, [foo(), React.createElement("a")]);`, Tsx: true},
		// Locks in upstream isInline() JSX branch: a block-level tag separates
		// two inline tags.
		{Code: `<div><a /><div /><span /></div>;`, Tsx: true},
		// Locks in the non-array third-argument branch: a single child value is
		// not treated as an array of children.
		{Code: `React.createElement("div", null, React.createElement("a"));`, Tsx: true},
		// Locks in upstream validate() early return: an empty JSX child list is
		// not inspected.
		{Code: `<div></div>;`, Tsx: true},
		// Locks in upstream CallExpression argument-count guard: fewer than three
		// arguments do not enter array-child validation.
		{Code: `React.createElement("div", undefined);`, Tsx: true},
		// Locks in upstream validate() first-pair return: only one diagnostic is
		// emitted for a container with three adjacent inline children.
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: multi-line JSX container range ----
		{Code: "<div>\n  <a /><span />\n</div>;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "inlineElement", Message: inlineElementMessage, Line: 1, Column: 1, EndLine: 3, EndColumn: 7}}},
		// Locks in upstream isCreateElement() destructured-import branch.
		{Code: "import { createElement } from \"react\";\ncreateElement(\"div\", null, [createElement(\"a\"), createElement(\"span\")]);", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "inlineElement", Message: inlineElementMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 72}}},
		// Locks in upstream's computed-identifier callee branch.
		{Code: `React[createElement]("div", null, [React.createElement("a"), React.createElement("span")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 91)}},
		{Code: `React?.[createElement]("div", null, [React.createElement("a"), React.createElement("span")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 93)}},
		{Code: `<div><img /><input /></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 28)}},
		{Code: `<div><img></img><input></input></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 38)}},
		{Code: `<div><a /><span /></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 25)}},
		{Code: `React.createElement("div", null, [React.createElement("a"), "", React.createElement("a")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 91)}},
		{Code: `<div><span><a /><a /></span></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "inlineElement", Message: inlineElementMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 29}}},
		{Code: "<div>\n  <a />\n  <span />\n</div>;", Tsx: true},
		// ---- Literal children in createElement arrays remain inline after
		// JavaScript-to-TypeScript AST kind mapping ----
		{Code: `React.createElement("div", null, [React.createElement("a"), 0, React.createElement("a")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 90)}},
		{Code: `React.createElement("div", null, [React.createElement("a"), true, React.createElement("a")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 93)}},
		{Code: `React.createElement("div", null, [React.createElement("a"), null, React.createElement("a")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 93)}},
		{Code: `React.createElement("div", null, [React.createElement("a"), /x/, React.createElement("a")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 92)}},
		{Code: `React.createElement("div", null, [React.createElement("a"), 1n, React.createElement("a")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 91)}},
		// ---- Dimension 4: first adjacent pair is reported once ----
		{Code: `<div><a /><span /><b /></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 30)}},
		// Locks in upstream isInline() call-expression branch: the first string
		// argument names the rendered element.
		{Code: `React.createElement("div", null, [foo("a"), foo("span")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 58)}},
		// ---- Real-user: adjacent React.createElement children ----
		{Code: `React.createElement("div", undefined, [React.createElement("button"), React.createElement("input")]);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 101)}},
		// ---- Real-user: inline links in a list ----
		{Code: `<nav><a href="/one" /><a href="/two" /></nav>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{inlineElementError(1, 1, 46)}},
	})
}
