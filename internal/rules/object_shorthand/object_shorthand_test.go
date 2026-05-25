package object_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestObjectShorthandRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ObjectShorthandRule,
		[]rule_tester.ValidTestCase{
			// default — always
			{Code: `var x = {y() {}}`},
			{Code: `var x = {y}`},
			{Code: `var x = {a: b}`},
			{Code: `var x = {a: 'a'}`},
			{Code: `var x = {'a': 'a'}`},
			{Code: `var x = {'a': b}`},
			{Code: `var x = {y(x) {}}`},
			{Code: `var {x,y,z} = x`},
			{Code: `var {x: {y}} = z`},
			{Code: `var x = {*x() {}}`},
			{Code: `var x = {x: y}`},
			{Code: `var x = {x: y, y: z}`},
			{Code: `var x = {x() {}, y: z, l(){}}`},
			{Code: `var x = {[y]: y}`},
			{Code: `doSomething({x: y})`},
			{Code: `!{ a: function a(){} };`},
			// arrow functions allowed by default
			{Code: `var x = {y: (x)=>x}`},
			{Code: `doSomething({y: (x)=>x})`},
			// getters/setters allowed
			{Code: `var x = {get y() {}}`},
			{Code: `var x = {set y(z) {}}`},
			{Code: `var x = {get y() {}, set y(z) {}}`},

			// options: properties
			{Code: `var x = {[y]: y}`, Options: []any{"properties"}},
			{Code: `var x = {['y']: 'y'}`, Options: []any{"properties"}},
			{Code: `var x = {['y']: y}`, Options: []any{"properties"}},

			// options: methods
			{Code: `var x = {[y]() {}}`, Options: []any{"methods"}},
			{Code: `var x = {[y]: function x() {}}`, Options: []any{"methods"}},
			{Code: `var x = {[y]: y}`, Options: []any{"methods"}},
			{Code: `var x = {y() {}}`, Options: []any{"methods"}},
			{Code: `var x = {x, y() {}, a:b}`, Options: []any{"methods"}},

			// options: properties disables method shorthand enforcement
			{Code: `var x = {y}`, Options: []any{"properties"}},
			{Code: `var x = {y: {b}}`, Options: []any{"properties"}},

			// options: never
			{Code: `var x = {a: n, c: d, f: g}`, Options: []any{"never"}},
			{Code: `var x = {a: function(){}, b: {c: d}}`, Options: []any{"never"}},

			// ignoreConstructors
			{Code: `var x = {ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {_ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {$ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {__ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {_0ConstructorFunction: function(){}, a: b}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},
			{Code: `var x = {notConstructorFunction(){}, b: c}`, Options: []any{"always", map[string]any{"ignoreConstructors": true}}},

			// methodsIgnorePattern
			{Code: `var x = { foo: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: `var x = { 'foo': function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: `var x = { ['foo']: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^foo$"}}},
			{Code: `var x = { 123: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^123$"}}},
			{Code: `var x = { afoob: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "foo"}}},
			{Code: `var x = { afoob: function() {}  }`, Options: []any{"always", map[string]any{"methodsIgnorePattern": "^.foo.$"}}},

			// avoidQuotes
			{Code: `var x = {'a': function(){}}`, Options: []any{"always", map[string]any{"avoidQuotes": true}}},
			{Code: `var x = {['a']: function(){}}`, Options: []any{"methods", map[string]any{"avoidQuotes": true}}},
			{Code: `var x = {'y': y}`, Options: []any{"properties", map[string]any{"avoidQuotes": true}}},

			// consistent
			{Code: `var x = {a: a, b: b}`, Options: []any{"consistent"}},
			{Code: `var x = {a: b, c: d, f: g}`, Options: []any{"consistent"}},
			{Code: `var x = {a, b}`, Options: []any{"consistent"}},
			{Code: `var x = {a, b, get test() { return 1; }}`, Options: []any{"consistent"}},

			// consistent-as-needed
			{Code: `var x = {a, b}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {0: 'foo'}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {'key': 'baz'}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {foo: 'foo'}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {[foo]: foo}`, Options: []any{"consistent-as-needed"}},
			{Code: `var x = {foo: function foo() {}}`, Options: []any{"consistent-as-needed"}},

			// avoidExplicitReturnArrows
			{Code: `({ x: () => foo })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": false}}},
			{Code: `({ x: () => foo })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `({ x() { return; } })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `({ x: () => { this; } })`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
			{Code: `function foo() { ({ x: () => { arguments; } }) }`, Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var x = {x: x}`,
				Output: []string{`var x = {x}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `var x = {'x': x}`,
				Output: []string{`var x = {x}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `var x = {y: y, x: x}`,
				Output: []string{`var x = {y, x}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10},
					{MessageId: "expectedPropertyShorthand", Line: 1, Column: 16},
				},
			},
			{
				Code:   `var x = {y: function() {}}`,
				Output: []string{`var x = {y() {}}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `var x = {y: function*() {}}`,
				Output: []string{`var x = {*y() {}}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:   `doSomething({x: x})`,
				Output: []string{`doSomething({x})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 14}},
			},
			{
				Code:   `doSomething({y: function() {}})`,
				Output: []string{`doSomething({y() {}})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 14}},
			},
			{
				Code:   `doSomething({[y]: function() {}})`,
				Output: []string{`doSomething({[y]() {}})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 14}},
			},
			// `options: ["never"]`
			{
				Code:    `var x = {y() {}}`,
				Output:  []string{`var x = {y: function() {}}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {*y() {}}`,
				Output:  []string{`var x = {y: function*() {}}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodLongform", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {y}`,
				Output:  []string{`var x = {y: y}`},
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyLongform", Line: 1, Column: 10}},
			},
			// properties option
			{
				Code:    `var x = {x: x}`,
				Output:  []string{`var x = {x}`},
				Options: []any{"properties"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedPropertyShorthand", Line: 1, Column: 10}},
			},
			// methods option
			{
				Code:    `var x = {y: function() {}}`,
				Output:  []string{`var x = {y() {}}`},
				Options: []any{"methods"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			// avoidQuotes
			{
				Code:    `var x = {a: function(){}}`,
				Output:  []string{`var x = {a(){}}`},
				Options: []any{"methods", map[string]any{"avoidQuotes": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 10}},
			},
			{
				Code:    `var x = {'a'(){}}`,
				Output:  []string{`var x = {'a': function(){}}`},
				Options: []any{"always", map[string]any{"avoidQuotes": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedLiteralMethodLongform", Line: 1, Column: 10}},
			},
			// consistent
			{
				Code:    `var x = {a: a, b}`,
				Options: []any{"consistent"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
			// consistent-as-needed
			{
				Code:    `var x = {a: a, b: b}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedAllPropertiesShorthanded", Line: 1, Column: 9}},
			},
			{
				Code:    `var x = {a, z: function z(){}}`,
				Options: []any{"consistent-as-needed"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedMix", Line: 1, Column: 9}},
			},
			// avoidExplicitReturnArrows
			{
				Code:    `({ x: () => { return; } })`,
				Output:  []string{`({ x() { return; } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
			{
				Code:    `({ x: foo => { return; } })`,
				Output:  []string{`({ x(foo) { return; } })`},
				Options: []any{"always", map[string]any{"avoidExplicitReturnArrows": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expectedMethodShorthand", Line: 1, Column: 4}},
			},
		},
	)
}
