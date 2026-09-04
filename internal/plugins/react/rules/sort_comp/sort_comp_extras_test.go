// TestSortCompExtras covers tsgo edge shapes, real-user examples, and
// branch lock-ins that are not part of the upstream migration sibling.
package sort_comp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortCompExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortCompRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: non-Identifier keys do not match named groups. ----
		{Code: `class Foo extends React.Component { "render"() {} ["displayName"] = 1; }`, Tsx: true},
		{Code: "class Foo extends React.Component { [`render`]() {} render() {} }", Tsx: true},
		// ---- Dimension 4: nested non-React classes are independent. ----
		{Code: `class Outer extends React.Component { onClick() {} render() { class Helper { render() {} foo() {} } return null; } }`, Tsx: true},
		// ---- Dimension 4: empty containers and spread members are safe. ----
		{Code: `class Foo extends React.Component {}`, Tsx: true},
		{Code: `var Foo = createReactClass({ ...defaults, render() {} });`, Tsx: true},
		// ---- Dimension 4: wrappers around the createReactClass object. ----
		{Code: `var Foo = createReactClass(({ onClick() {}, render() {} }));`, Tsx: true},
		{Code: `class Foo extends React.Component { static #private = 1; render() {} }`, Tsx: true},
		// ---- Compatibility: @jsx changes the component pragma. ----
		{Code: `/** @jsx Preact */ class Foo extends React.Component { render() {} componentDidMount() {} }`, Tsx: true},
		// ---- Compatibility: non-plain class members are not variable/method groups. ----
		{Code: `abstract class Foo extends React.Component { render() {} abstract props: string; }`, Tsx: true, Options: sortCompOrder("instance-variables", "render")},
		{Code: `abstract class Foo extends React.Component { render() {} static abstract method(): void; }`, Tsx: true, Options: sortCompOrder("static-methods", "render")},
		{Code: `class Foo extends React.Component { render() {} accessor foo = 1; }`, Tsx: true, Options: sortCompOrder("instance-variables", "render")},
		{Code: `class Foo extends React.Component { render() {} accessor foo = () => {}; }`, Tsx: true, Options: sortCompOrder("instance-methods", "render")},
		{Code: `class Foo extends React.Component { render() {} static accessor foo = 1; }`, Tsx: true, Options: sortCompOrder("static-variables", "render")},
		{Code: `class Foo extends React.Component { render() {} get foo() {} }`, Tsx: true, Options: sortCompOrder("foo", "render")},
		{Code: `class Foo extends React.Component { render() {} set foo(value) {} }`, Tsx: true, Options: sortCompOrder("foo", "render")},
		// ---- Real-user: issue #2000 has an async arrow field referring to a prior field. ----
		{Code: `class Test extends React.PureComponent { fetch = async () => {}; lazyFetch = _.debounce(this.fetch, 1000); }`, Tsx: true},
		// ---- Branch lock-in: a named method can also match a custom regex group. ----
		{Code: `class Foo extends React.Component { render() {} onClick() {} }`, Tsx: true, Options: sortCompOrder("/on.*/", "render", "/.*Click/")},
		// ---- Branch lock-in: an unknown group without everything-else is Infinity. ----
		{Code: `class Foo extends React.Component { render() {} other() {} }`, Tsx: true, Options: sortCompOrder("render")},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 1/4: optional / computed names remain unmatched. ----
		{Code: `class Foo extends React.Component { render() {} displayName() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "displayName")}},
		// ---- Unmatched members retain JavaScript's Infinity distance when everything-else is absent. ----
		{Code: `class Foo extends React.Component { unknown() {} render() {} displayName() {} }`, Tsx: true, Options: sortCompOrder("lifecycle", "render"), Errors: []rule_tester.InvalidTestCaseError{
			sortCompError("unknown", "after", "displayName"),
			sortCompError("render", "after", "displayName"),
		}},
		{Code: `var Foo = createReactClass({ ...defaults, render() {}, displayName: "Foo" });`, Tsx: true, Options: sortCompOrder("lifecycle", "/^on.+$/", "render"), Errors: []rule_tester.InvalidTestCaseError{
			sortCompError("", "after", "displayName"),
			sortCompError("render", "after", "displayName"),
		}},
		// ---- Dimension 4: class expression is checked independently. ----
		{Code: `const Foo = class extends React.Component { render() {} componentDidMount() {} };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		// ---- Branch lock-in: custom group can be overridden while lifecycle remains. ----
		{Code: `class Foo extends React.Component { render() {} componentDidMount() {} }`, Tsx: true, Options: map[string]any{"order": []any{"lifecycle", "custom", "render"}, "groups": map[string]any{"custom": []any{"componentDidMount"}}}, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		{Code: `/** @jsx Preact */ class Foo extends Preact.Component { render() {} componentDidMount() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		{Code: `/** @jsx Preact */ const Foo = Preact.createReactClass({ render() {}, displayName() {} });`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "displayName")}},
		// ---- Real-user: fields with an explicit custom order are still classified. ----
		{Code: `class App extends React.Component { render() {} handler = () => {}; }`, Tsx: true, Options: sortCompOrder("instance-methods", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "handler")}},
		// ---- Real-user: issue #1814 demonstrates the interaction between
		// lifecycle fields and ordinary fields under the default order. ----
		{Code: `class App extends React.PureComponent { something = { whatever: true }; state = { count: 0 }; constructor() { super(); } componentDidMount() {} render() {} onWsMessage() {} onWsError() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			sortCompError("something", "after", "componentDidMount"),
			sortCompError("state", "after", "constructor"),
		}},
		// ---- Compatibility: computed identifiers are named keys in ESTree. ----
		{Code: `class Foo extends React.Component { render() {} [displayName]() {} }`, Tsx: true, Options: sortCompOrder("displayName", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "displayName")}},
		{Code: `class Foo extends React.Component { render() {} [foo.bar]() {} }`, Tsx: true, Options: sortCompOrder("/undefined/,", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "undefined")}},
		{Code: `var Foo = createReactClass({ render() {}, [foo.bar]() {} });`, Tsx: true, Options: sortCompOrder("/undefined/,", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "undefined")}},
		// ---- Compatibility: ESTree omits empty class elements. ----
		{Code: `class Foo extends React.Component { render() {} ; componentDidMount() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		// ---- Compatibility: literal keys preserve JavaScript's undefined text. ----
		{Code: `class Foo extends React.Component { render() {} "displayName"() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "undefined")}},
		// ---- Compatibility: upstream accepts a regex prefix before trailing text. ----
		{Code: `class Foo extends React.Component { render() {} onClick() {} }`, Tsx: true, Options: sortCompOrder("/on.*/,", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "onClick")}},
		{Code: `class Foo extends React.Component { get foo() {} render() {} }`, Tsx: true, Options: sortCompOrder("foo", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("getter functions", "after", "render")}},
		{Code: `class Foo extends React.Component { set foo(value) {} render() {} }`, Tsx: true, Options: sortCompOrder("foo", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("setter functions", "after", "render")}},
		// ---- Compatibility: abstract accessors retain getter/setter groups. ----
		{Code: `abstract class Foo extends React.Component { render() {} abstract static get value(): string; }`, Tsx: true, Options: sortCompOrder("getters", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "getter functions")}},
		{Code: `abstract class Foo extends React.Component { render() {} abstract static set value(next: string); }`, Tsx: true, Options: sortCompOrder("setters", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "setter functions")}},
		{Code: `class Foo extends React.Component { render() {} foo = (function() {}) }`, Tsx: true, Options: sortCompOrder("instance-methods", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "foo")}},
		{Code: `class Foo extends React.Component { render() {} foo = (() => {}) }`, Tsx: true, Options: sortCompOrder("instance-methods", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "foo")}},
		{Code: `class Foo extends React.Component { render() {} static foo(): void; static foo() {} }`, Tsx: true, Options: sortCompOrder("static-methods", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("foo", "before", "render")}},
		// ---- Compatibility: JSDoc-only React component declarations. ----
		{Code: `/** @extends React.Component */ class Foo { render() {} componentDidMount() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		{Code: `/** @augments React.PureComponent */ class Foo { render() {} componentDidMount() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		// ---- Compatibility: computed React component inheritance. ----
		{Code: `class Foo extends React[Component] { render() {} componentDidMount() {} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "componentDidMount")}},
		// ---- Compatibility: error deduplication uses the original entry snapshot. ----
		{Code: `class Foo extends React.Component { render() {} static foo() {} componentDidMount() {} onClick() {} displayName() {} }`, Tsx: true, Options: sortCompOrder("lifecycle", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("displayName", "before", "componentDidMount")}},
	})
}
