// TestPreferSpyOnUpstream migrates the @vitest/eslint-plugin v1.6.27
// prefer-spy-on suite, with the namespace rewritten to Rstest's `rs`. Every
// case is carried over.
//
// Where the factory is called without an implementation, the expected fix
// installs `() => {}` instead of an argument-less
// `.mockImplementation()`: an Rstest spy with no implementation calls through
// to the original method, while `rs.fn()` returned `undefined`.
package prefer_spy_on

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func useSpyOn(line, column, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "useRsSpyOn",
		Message:   "Use `rs.spyOn` instead",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}}
}

func TestPreferSpyOnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferSpyOnRule,
		[]rule_tester.ValidTestCase{
			{Code: `Date.now = () => 10`},
			{Code: `window.fetch = rs.fn`},
			{Code: `Date.now = fn()`},
			{Code: `obj.mock = rs.something()`},
			{Code: `const mock = rs.fn()`},
			{Code: `mock = rs.fn()`},
			{Code: `const mockObj = { mock: rs.fn() }`},
			{Code: `mockObj = { mock: rs.fn() }`},
			{Code: "window[`${name}`] = rs[`fn${expression}`]()"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `obj.a = rs.fn(); const test = 10;`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => {}); const test = 10;`},
				Errors: useSpyOn(1, 1, 16),
			},
			{
				Code:   `Date['now'] = rs['fn']()`,
				Output: []string{`rs.spyOn(Date, 'now').mockImplementation(() => {})`},
				Errors: useSpyOn(1, 1, 25),
			},
			{
				Code:   "window[`${name}`] = rs[`fn`]()",
				Output: []string{"rs.spyOn(window, `${name}`).mockImplementation(() => {})"},
				Errors: useSpyOn(1, 1, 31),
			},
			{
				Code:   `obj['prop' + 1] = rs['fn']()`,
				Output: []string{`rs.spyOn(obj, 'prop' + 1).mockImplementation(() => {})`},
				Errors: useSpyOn(1, 1, 29),
			},
			{
				Code:   `obj.one.two = rs.fn(); const test = 10;`,
				Output: []string{`rs.spyOn(obj.one, 'two').mockImplementation(() => {}); const test = 10;`},
				Errors: useSpyOn(1, 1, 22),
			},
			{
				Code:   `obj.a = rs.fn(() => 10,)`,
				Output: []string{`rs.spyOn(obj, 'a').mockImplementation(() => 10)`},
				Errors: useSpyOn(1, 1, 25),
			},
			{
				Code:   `obj.a.b = rs.fn(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();`,
				Output: []string{`rs.spyOn(obj.a, 'b').mockImplementation(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();`},
				Errors: useSpyOn(1, 1, 89),
			},
			{
				Code:   `window.fetch = rs.fn(() => ({})).one.two().three().four`,
				Output: []string{`rs.spyOn(window, 'fetch').mockImplementation(() => ({})).one.two().three().four`},
				Errors: useSpyOn(1, 1, 56),
			},
			{
				Code:   `foo[bar] = rs.fn().mockReturnValue(undefined)`,
				Output: []string{`rs.spyOn(foo, bar).mockImplementation(() => {}).mockReturnValue(undefined)`},
				Errors: useSpyOn(1, 1, 46),
			},
			{
				Code: `
        foo.bar = rs.fn().mockImplementation(baz => baz)
        foo.bar = rs.fn(a => b).mockImplementation(baz => baz)
      `,
				Output: []string{`
        rs.spyOn(foo, 'bar').mockImplementation(baz => baz)
        rs.spyOn(foo, 'bar').mockImplementation(baz => baz)
      `},
				Errors: append(useSpyOn(2, 9, 57), useSpyOn(3, 9, 63)...),
			},
		},
	)
}
