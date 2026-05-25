package no_useless_computed_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessComputedKeyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessComputedKeyRule,
		[]rule_tester.ValidTestCase{
			// ---- Object literals ----
			{Code: `({ 'a': 0, b(){} })`},
			{Code: `({ [x]: 0 });`},
			{Code: `({ a: 0, [b](){} })`},
			// Computed __proto__ in an object literal defines a property
			// named __proto__; non-computed __proto__ sets the prototype.
			{Code: `({ ['__proto__']: [] })`},

			// ---- Object destructuring ----
			{Code: `var { 'a': foo } = obj`},
			{Code: `var { [a]: b } = obj;`},
			{Code: `var { a } = obj;`},
			{Code: `var { a: a } = obj;`},
			{Code: `var { a: b } = obj;`},

			// ---- Class members with enforceForClassMembers: true ----
			{Code: `class Foo { a() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { 'a'() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { [x]() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { ['constructor']() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { static ['prototype']() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `(class { 'a'() {} })`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `(class { [x]() {} })`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `(class { ['constructor']() {} })`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `(class { static ['prototype']() {} })`, Options: map[string]interface{}{"enforceForClassMembers": true}},

			// ---- Class members with default options ----
			{Code: `class Foo { 'x'() {} }`},
			{Code: `(class { [x]() {} })`},
			{Code: `class Foo { static constructor() {} }`},
			{Code: `class Foo { prototype() {} }`},

			// ---- Class members with enforceForClassMembers: false ----
			{Code: `class Foo { ['x']() {} }`, Options: map[string]interface{}{"enforceForClassMembers": false}},
			{Code: `(class { ['x']() {} })`, Options: map[string]interface{}{"enforceForClassMembers": false}},
			{Code: `class Foo { static ['constructor']() {} }`, Options: map[string]interface{}{"enforceForClassMembers": false}},
			{Code: `class Foo { ['prototype']() {} }`, Options: map[string]interface{}{"enforceForClassMembers": false}},

			// ---- Class fields ----
			{Code: `class Foo { a }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { ['constructor'] }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { static ['constructor'] }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class Foo { static ['prototype'] }`, Options: map[string]interface{}{"enforceForClassMembers": true}},

			// BigInt literals: browsers throw on bigint property names, so
			// the rule deliberately leaves them alone.
			{Code: `({ [99999999999999999n]: 0 })`},

			// ---- Non-literal expressions are ineligible regardless of shape ----
			{Code: `const k = 'x'; const o = { [k]: 1 }`},
			{Code: `const x = { [Symbol()]: 1 }`},

			// ---- Namespace / module containing object literal is value position ----
			{Code: "namespace N { export const v = { ['__proto__']: [] } }"},

			// ---- TS 'as' / 'satisfies' wrappers inside computed brackets
			//      mean the key is no longer a plain literal; not reported. ----
			{Code: `const x = { [('x' as const)]: 1 }`},
			{Code: `const x = { [('x' satisfies string)]: 1 }`},

			// ---- Non-literal computed key kinds (structurally ineligible) ----
			// Template literals (even no-substitution) aren't flagged — matches
			// ESLint, which only inspects `Literal` nodes.
			{Code: "({ [`x`]: 0 })"},
			{Code: "class Foo { [`x`]() {} }"},
			// Unary-negative and other expressions are not literals.
			{Code: `({ [-1]: 0 })`},
			{Code: `({ [void 0]: 0 })`},
			// Regex literals: non-string/non-number.
			{Code: `({ [/x/]: 0 })`},

			// ---- Auto-accessor: typescript-eslint maps `accessor x` to
			//      `AccessorProperty`, which ESLint's core rule does not
			//      listen for. Match that (no report). ----
			{Code: `class Foo { accessor ['x'] = 1 }`},
			{Code: `class Foo { static accessor ['x'] = 1 }`},
			{Code: `class Foo { accessor ['constructor'] = 1 }`},
			{Code: `class Foo { static accessor ['prototype'] = 1 }`},

			// ---- Abstract class members surface as TSAbstract* nodes in
			//      typescript-eslint, which the core rule doesn't listen for.
			//      Match that (no report) for methods, getters, setters,
			//      and fields. ----
			{Code: `abstract class Foo { abstract ['x'](): void }`},
			{Code: `abstract class Foo { abstract ['x']: string }`},
			{Code: `abstract class Foo { abstract get ['x'](): number }`},
			{Code: `abstract class Foo { abstract set ['x'](v: number) }`},
			{Code: `abstract class Foo { abstract readonly ['x']: number }`},

			// ---- TypeScript-only containers must not be flagged ----
			{Code: `interface I { ['foo']: number }`},
			{Code: `type T = { ['foo']: number }`},
			{Code: `interface I { ['foo'](): void }`},
			// Signature-style members inside a type literal.
			{Code: `type T = { readonly ['foo']: number }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Object literal property keys ----
			{
				Code:   `({ ['0']: 0 })`,
				Output: []string{`({ '0': 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4, Message: "Unnecessarily computed property ['0'] found."},
				},
			},
			{
				Code:   `var { ['0']: a } = obj`,
				Output: []string{`var { '0': a } = obj`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 7, EndLine: 1, EndColumn: 15, Message: "Unnecessarily computed property ['0'] found."},
				},
			},
			{
				Code:   `({ ['0+1,234']: 0 })`,
				Output: []string{`({ '0+1,234': 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ [0]: 0 })`,
				Output: []string{`({ 0: 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4, Message: "Unnecessarily computed property [0] found."},
				},
			},
			{
				Code:   `var { [0]: a } = obj`,
				Output: []string{`var { 0: a } = obj`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 7},
				},
			},
			{
				Code:   `({ ['x']: 0 })`,
				Output: []string{`({ 'x': 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `var { ['x']: a } = obj`,
				Output: []string{`var { 'x': a } = obj`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 7},
				},
			},
			{
				Code:   `var { ['__proto__']: a } = obj`,
				Output: []string{`var { '__proto__': a } = obj`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 7},
				},
			},
			{
				Code:   `({ ['x']() {} })`,
				Output: []string{`({ 'x'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},

			// ---- Comments block auto-fix ----
			{
				Code: `({ [/* this comment prevents a fix */ 'x']: 0 })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code: `({ ['x' /* this comment also prevents a fix */]: 0 })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Parenthesized literals ----
			{
				Code:   `({ [('x')]: 0 })`,
				Output: []string{`({ 'x': 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `var { [('x')]: a } = obj`,
				Output: []string{`var { 'x': a } = obj`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 7},
				},
			},

			// ---- Generator / async object methods ----
			{
				Code:   `({ *['x']() {} })`,
				Output: []string{`({ *'x'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ async ['x']() {} })`,
				Output: []string{`({ async 'x'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},

			// ---- Adjacency between prefix keyword and numeric `.2` is safe;
			//      between keyword and digit-initial numeric requires a space.
			{
				Code:   `({ get[.2]() {} })`,
				Output: []string{`({ get.2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4, Message: "Unnecessarily computed property [.2] found."},
				},
			},
			{
				Code:   `({ set[.2](value) {} })`,
				Output: []string{`({ set.2(value) {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ async[.2]() {} })`,
				Output: []string{`({ async.2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ [2]() {} })`,
				Output: []string{`({ 2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ get [2]() {} })`,
				Output: []string{`({ get 2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ set [2](value) {} })`,
				Output: []string{`({ set 2(value) {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ async [2]() {} })`,
				Output: []string{`({ async 2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ get[2]() {} })`,
				Output: []string{`({ get 2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ set[2](value) {} })`,
				Output: []string{`({ set 2(value) {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ async[2]() {} })`,
				Output: []string{`({ async 2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ get['foo']() {} })`,
				Output: []string{`({ get'foo'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ *[2]() {} })`,
				Output: []string{`({ *2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ async*[2]() {} })`,
				Output: []string{`({ async*2() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},

			// ---- Object literal reserved-name keys that need no carve-out
			//      in value position (only __proto__ does).
			{
				Code:   `({ ['constructor']: 1 })`,
				Output: []string{`({ 'constructor': 1 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},
			{
				Code:   `({ ['prototype']: 1 })`,
				Output: []string{`({ 'prototype': 1 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},

			// ---- Class methods ----
			{
				Code:    `class Foo { ['0']() {} }`,
				Output:  []string{`class Foo { '0'() {} }`},
				Options: map[string]interface{}{"enforceForClassMembers": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:    `class Foo { ['0+1,234']() {} }`,
				Output:  []string{`class Foo { '0+1,234'() {} }`},
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { ['x']() {} }`,
				Output: []string{`class Foo { 'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code: `class Foo { [/* this comment prevents a fix */ 'x']() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code: `class Foo { ['x' /* this comment also prevents a fix */]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `class Foo { [('x')]() {} }`,
				Output: []string{`class Foo { 'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { *['x']() {} }`,
				Output: []string{`class Foo { *'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { async ['x']() {} }`,
				Output: []string{`class Foo { async 'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { get[.2]() {} }`,
				Output: []string{`class Foo { get.2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { set[.2](value) {} }`,
				Output: []string{`class Foo { set.2(value) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { async[.2]() {} }`,
				Output: []string{`class Foo { async.2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { [2]() {} }`,
				Output: []string{`class Foo { 2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { get [2]() {} }`,
				Output: []string{`class Foo { get 2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { set [2](value) {} }`,
				Output: []string{`class Foo { set 2(value) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { async [2]() {} }`,
				Output: []string{`class Foo { async 2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { get[2]() {} }`,
				Output: []string{`class Foo { get 2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { set[2](value) {} }`,
				Output: []string{`class Foo { set 2(value) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { async[2]() {} }`,
				Output: []string{`class Foo { async 2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { get['foo']() {} }`,
				Output: []string{`class Foo { get'foo'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { *[2]() {} }`,
				Output: []string{`class Foo { *2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { async*[2]() {} }`,
				Output: []string{`class Foo { async*2() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},

			// ---- Class reserved-name method carve-outs ----
			{
				Code:   `class Foo { static ['constructor']() {} }`,
				Output: []string{`class Foo { static 'constructor'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { ['prototype']() {} }`,
				Output: []string{`class Foo { 'prototype'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `(class { ['x']() {} })`,
				Output: []string{`(class { 'x'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:   `(class { ['__proto__']() {} })`,
				Output: []string{`(class { '__proto__'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:   `(class { static ['__proto__']() {} })`,
				Output: []string{`(class { static '__proto__'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:   `(class { static ['constructor']() {} })`,
				Output: []string{`(class { static 'constructor'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:   `(class { ['prototype']() {} })`,
				Output: []string{`(class { 'prototype'() {} })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},

			// ---- Class fields ----
			{
				Code:   `class Foo { ['0'] }`,
				Output: []string{`class Foo { '0' }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { ['0'] = 0 }`,
				Output: []string{`class Foo { '0' = 0 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { static[0] }`,
				Output: []string{`class Foo { static 0 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13},
				},
			},
			{
				Code:   `class Foo { ['#foo'] }`,
				Output: []string{`class Foo { '#foo' }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13, EndLine: 1, EndColumn: 21, Message: "Unnecessarily computed property ['#foo'] found."},
				},
			},
			{
				Code:   `(class { ['__proto__'] })`,
				Output: []string{`(class { '__proto__' })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:   `(class { static ['__proto__'] })`,
				Output: []string{`(class { static '__proto__' })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},
			{
				Code:   `(class { ['prototype'] })`,
				Output: []string{`(class { 'prototype' })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 10},
				},
			},

			// ---- Multi-line: lock EndLine / EndColumn ----
			{
				Code: `({
  ['x']: 0
})`,
				Output: []string{`({
  'x': 0
})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 2, Column: 3, EndLine: 2, EndColumn: 11, Message: "Unnecessarily computed property ['x'] found."},
				},
			},
			{
				Code: `class Foo {
  static ['x']() {
    return 1;
  }
}`,
				Output: []string{`class Foo {
  static 'x'() {
    return 1;
  }
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 2, Column: 3, EndLine: 4, EndColumn: 4, Message: "Unnecessarily computed property ['x'] found."},
				},
			},

			// ---- Numeric literal raw text is preserved (hex / exponent) ----
			{
				Code:   `({ [0x1]: 0 })`,
				Output: []string{`({ 0x1: 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4, Message: "Unnecessarily computed property [0x1] found."},
				},
			},
			{
				Code:   `({ [1e2]: 0 })`,
				Output: []string{`({ 1e2: 0 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4, Message: "Unnecessarily computed property [1e2] found."},
				},
			},

			// ---- PROBE: plain assignment destructuring (non-__proto__) ----
			{
				Code:   `({ ['y']: a } = b)`,
				Output: []string{`({ 'y': a } = b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Assignment destructuring: the __proto__ carve-out is for
			//      value-position literals only. tsgo reuses
			//      ObjectLiteralExpression for the LHS of `=` / for-in/of, so
			//      the rule must treat those as patterns. ----
			{
				Code:   `({ ['__proto__']: a } = b)`,
				Output: []string{`({ '__proto__': a } = b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4, Message: "Unnecessarily computed property ['__proto__'] found."},
				},
			},
			{
				Code:   `({ y: { ['__proto__']: a } } = b)`,
				Output: []string{`({ y: { '__proto__': a } } = b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 9},
				},
			},
			{
				Code:   `for ({ ['__proto__']: a } of arr) {}`,
				Output: []string{`for ({ '__proto__': a } of arr) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 8},
				},
			},
			{
				Code:   `for ({ ['__proto__']: a } in arr) {}`,
				Output: []string{`for ({ '__proto__': a } in arr) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 8},
				},
			},
			{
				Code:   `[{ ['__proto__']: a }] = [b]`,
				Output: []string{`[{ '__proto__': a }] = [b]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},

			// ---- Destructuring with default value ----
			{
				Code:   `var { ['x']: a = 1 } = obj`,
				Output: []string{`var { 'x': a = 1 } = obj`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 7},
				},
			},
			{
				Code:   `({ ['x']: a = 1 } = b)`,
				Output: []string{`({ 'x': a = 1 } = b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 4},
				},
			},

			// ---- `override` modifier is a plain MethodDefinition in TSESTree,
			//      so it should report. ----
			{
				Code:   `class Base { x() {} }; class Sub extends Base { override ['x']() {} }`,
				Output: []string{`class Base { x() {} }; class Sub extends Base { override 'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- `declare class` members surface as regular
			//      MethodDefinition/PropertyDefinition — should report. ----
			{
				Code:   `declare class Foo { ['x']: string }`,
				Output: []string{`declare class Foo { 'x': string }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 21},
				},
			},
			{
				Code:   `declare class Foo { ['x'](): void }`,
				Output: []string{`declare class Foo { 'x'(): void }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 21},
				},
			},

			// ---- Object literal wrapped in type assertion / as / satisfies ----
			{
				Code:   `const x = <const>{ ['x']: 1 }`,
				Output: []string{`const x = <const>{ 'x': 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { ['x']: 1 } as const`,
				Output: []string{`const x = { 'x': 1 } as const`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { ['x']: 1 } satisfies Record<string, number>`,
				Output: []string{`const x = { 'x': 1 } satisfies Record<string, number>`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- String-literal variants ----
			{
				Code:   `const x = { ["x"]: 1 }`,
				Output: []string{`const x = { "x": 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Message: `Unnecessarily computed property ["x"] found.`},
				},
			},
			{
				Code:   `const x = { [""]: 1 }`,
				Output: []string{`const x = { "": 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { ["\u00e9"]: 1 }`,
				Output: []string{`const x = { "\u00e9": 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { ["delete"]: 1 }`,
				Output: []string{`const x = { "delete": 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Numeric-literal variants (raw text preserved) ----
			{
				Code:   `const x = { [0b10]: 1 }`,
				Output: []string{`const x = { 0b10: 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Message: "Unnecessarily computed property [0b10] found."},
				},
			},
			{
				Code:   `const x = { [0o7]: 1 }`,
				Output: []string{`const x = { 0o7: 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { [1_000]: 1 }`,
				Output: []string{`const x = { 1_000: 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { [1e-2]: 1 }`,
				Output: []string{`const x = { 1e-2: 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Multiple parenthesis levels ----
			{
				Code:   `const x = { [(('x'))]: 1 }`,
				Output: []string{`const x = { 'x': 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Multi-modifier class members ----
			{
				Code:   `class Foo { public static ['x']() {} }`,
				Output: []string{`class Foo { public static 'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `class Foo { static readonly ['x'] = 1 }`,
				Output: []string{`class Foo { static readonly 'x' = 1 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `class Foo { static async *['x']() {} }`,
				Output: []string{`class Foo { static async *'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `class Foo { static get ['x']() { return 1 } }`,
				Output: []string{`class Foo { static get 'x'() { return 1 } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Position in expression contexts ----
			{
				Code:   `const arrow = () => ({ ['x']: 1 })`,
				Output: []string{`const arrow = () => ({ 'x': 1 })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const cond = true ? { ['x']: 1 } : { a: 2 }`,
				Output: []string{`const cond = true ? { 'x': 1 } : { a: 2 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `function f() { return { ['x']: 1 } }`,
				Output: []string{`function f() { return { 'x': 1 } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Nested & parameter destructuring ----
			{
				Code:   `function f({ ['x']: a } = {} as any) { return a }`,
				Output: []string{`function f({ 'x': a } = {} as any) { return a }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const { ['x']: a, ...rest } = { x: 1, y: 2 }`,
				Output: []string{`const { 'x': a, ...rest } = { x: 1, y: 2 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `try {} catch ({ ['x']: e }) {}`,
				Output: []string{`try {} catch ({ 'x': e }) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Multiple diagnostics inside a single object ----
			{
				Code:   `const x = { ['x']: 1, ['y']: 2, a: 3, ['z']: 4 }`,
				Output: []string{`const x = { 'x': 1, 'y': 2, a: 3, 'z': 4 }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
					{MessageId: "unnecessarilyComputedProperty"},
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Multi-line computed bracket span (newline is not a comment,
			//      so auto-fix still applies) ----
			{
				Code:   "const x = {\n  [\n    'x'\n  ]: 1,\n}",
				Output: []string{"const x = {\n  'x': 1,\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 2, Column: 3, EndLine: 4, EndColumn: 7},
				},
			},

			// ---- Class implementing interface with computed literal key ----
			{
				Code:   `interface I { x(): void } class C implements I { ['x']() {} }`,
				Output: []string{`interface I { x(): void } class C implements I { 'x'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Class extending generic with computed field ----
			{
				Code:   `class Box<T> { ['x']: T | undefined }`,
				Output: []string{`class Box<T> { 'x': T | undefined }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Namespace containing class with computed method ----
			{
				Code:   `namespace N { export class C { ['x']() {} } }`,
				Output: []string{`namespace N { export class C { 'x'() {} } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Multiple pattern destructuring in for-of ----
			{
				Code:   `for (const { a: { ['x']: b } } of [{ a: { x: 1 } }]) { void b }`,
				Output: []string{`for (const { a: { 'x': b } } of [{ a: { x: 1 } }]) { void b }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Fix produces valid syntax even when key value is a
			//      context-sensitive keyword (async/get/set). ----
			{
				Code:   `const x = { async ['async']() {} }`,
				Output: []string{`const x = { async 'async'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { get ['get']() { return 1 } }`,
				Output: []string{`const x = { get 'get'() { return 1 } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},
			{
				Code:   `const x = { set ['set'](v) {} }`,
				Output: []string{`const x = { set 'set'(v) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Non-static class method named 'prototype' should report
			//      (only non-static 'constructor' is carved out). ----
			{
				Code:   `class Foo { ['prototype']() {} }`,
				Output: []string{`class Foo { 'prototype'() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- Object with same literal key repeated as get + set ----
			{
				Code:   `class Foo { get ['k']() { return 1 } set ['k'](v) {} }`,
				Output: []string{`class Foo { get 'k'() { return 1 } set 'k'(v) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty"},
					{MessageId: "unnecessarilyComputedProperty"},
				},
			},

			// ---- TypeScript type-modifying wrappers inside the bracket do
			//      NOT make the key a string literal. `('x' as const)` is an
			//      AsExpression, not a Literal — should NOT report. ----
			// (Valid — covered in valid section below.)

			// ---- Class getter/setter with computed literal key ----
			{
				Code:   `class Foo { get ['x']() { return 1 } }`,
				Output: []string{`class Foo { get 'x'() { return 1 } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13, Message: "Unnecessarily computed property ['x'] found."},
				},
			},
			{
				Code:   `class Foo { set ['x'](v) {} }`,
				Output: []string{`class Foo { set 'x'(v) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessarilyComputedProperty", Line: 1, Column: 13, Message: "Unnecessarily computed property ['x'] found."},
				},
			},
		},
	)
}
