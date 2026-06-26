package prefer_spy_on_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_spy_on"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSpyOnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_spy_on.PreferSpyOnRule,
		[]rule_tester.ValidTestCase{
			{Code: `Date.now = () => 10`},
			{Code: `window.fetch = jest.fn`},
			{Code: `Date.now = fn()`},
			{Code: `obj.mock = jest.something()`},
			{Code: `const mock = jest.fn()`},
			{Code: `mock = jest.fn()`},
			{Code: `const mockObj = { mock: jest.fn() }`},
			{Code: `mockObj = { mock: jest.fn() }`},
			{Code: "window[`${name}`] = jest[`fn${expression}`]()"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `obj.a = jest.fn(); const test = 10;`,
				Output: []string{`jest.spyOn(obj, 'a').mockImplementation(); const test = 10;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   `Date['now'] = jest['fn']()`,
				Output: []string{`jest.spyOn(Date, 'now').mockImplementation()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:   "window[`${name}`] = jest[`fn`]()",
				Output: []string{"jest.spyOn(window, `${name}`).mockImplementation()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:   `obj['prop' + 1] = jest['fn']()`,
				Output: []string{`jest.spyOn(obj, 'prop' + 1).mockImplementation()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:   `obj.one.two = jest.fn(); const test = 10;`,
				Output: []string{`jest.spyOn(obj.one, 'two').mockImplementation(); const test = 10;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:   `obj.a = jest.fn(() => 10,)`,
				Output: []string{`jest.spyOn(obj, 'a').mockImplementation(() => 10)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:   `obj.a.b = jest.fn(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();`,
				Output: []string{`jest.spyOn(obj.a, 'b').mockImplementation(() => ({})).mockReturnValue('default').mockReturnValueOnce('first call'); test();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 91},
				},
			},
			{
				Code:   `window.fetch = jest.fn(() => ({})).one.two().three().four`,
				Output: []string{`jest.spyOn(window, 'fetch').mockImplementation(() => ({})).one.two().three().four`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
				},
			},
			{
				Code:   `foo[bar] = jest.fn().mockReturnValue(undefined)`,
				Output: []string{`jest.spyOn(foo, bar).mockImplementation().mockReturnValue(undefined)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
				},
			},
			{
				Code: `
        foo.bar = jest.fn().mockImplementation(baz => baz)
        foo.bar = jest.fn(a => b).mockImplementation(baz => baz)
      `,
				Output: []string{`
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
        jest.spyOn(foo, 'bar').mockImplementation(baz => baz)
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 2, Column: 9, EndLine: 2, EndColumn: 59},
					{MessageId: "useJestSpyOn", Line: 3, Column: 9, EndLine: 3, EndColumn: 65},
				},
			},
			{
				Code:   `foo.bar = (jest.fn()).mockImplementation(baz => baz)`,
				Output: []string{`jest.spyOn(foo, 'bar').mockImplementation(baz => baz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code:   `foo.bar = (jest.fn(a => b)).mockImplementation(baz => baz)`,
				Output: []string{`jest.spyOn(foo, 'bar').mockImplementation(baz => baz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 59},
				},
			},
			{
				Code:   `obj.a = (jest.fn())`,
				Output: []string{`jest.spyOn(obj, 'a').mockImplementation()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   `obj.a = ((jest.fn()))`,
				Output: []string{`jest.spyOn(obj, 'a').mockImplementation()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:   `obj.a = (jest.fn(() => 10))`,
				Output: []string{`jest.spyOn(obj, 'a').mockImplementation(() => 10)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:   `obj.a = (jest.fn()).one.two()`,
				Output: []string{`jest.spyOn(obj, 'a').mockImplementation().one.two()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useJestSpyOn", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
		},
	)
}
