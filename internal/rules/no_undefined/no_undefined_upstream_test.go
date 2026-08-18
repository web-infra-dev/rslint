package no_undefined

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUndefinedUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-undefined.js (eslint v10.8.1) 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in no_undefined_extras_test.go.
func TestNoUndefinedUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefinedRule,
		[]rule_tester.ValidTestCase{
			{Code: `void 0`},
			{Code: `void!0`},
			{Code: `void-0`},
			{Code: `void+0`},
			{Code: `null`},
			{Code: `undefine`},
			{Code: `ndefined`},
			{Code: `a.undefined`},
			{Code: `this.undefined`},
			{Code: `global['undefined']`},

			// https://github.com/eslint/eslint/issues/7964
			{Code: `({ undefined: bar })`},
			{Code: `({ undefined: bar } = foo)`},
			{Code: `({ undefined() {} })`},
			{Code: `class Foo { undefined() {} }`},
			{Code: `(class { undefined() {} })`},
			{Code: `import { undefined as a } from 'foo'`},
			{Code: `export { undefined } from 'foo'`},
			{Code: `export { undefined as a } from 'foo'`},
			{Code: `export { a as undefined } from 'foo'`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `undefined`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 1}},
			},
			{
				Code:   `undefined.a`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 1}},
			},
			{
				Code:   `a[undefined]`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 3}},
			},
			{
				Code:   `undefined[0]`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 1}},
			},
			{
				Code:   `f(undefined)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 3}},
			},
			{
				Code:   `function f(undefined) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 12}},
			},
			{
				Code:   `function f() { var undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 20}},
			},
			{
				Code:   `function f() { undefined = true; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 16}},
			},
			{
				Code:   `var undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 5}},
			},
			{
				Code:   `try {} catch(undefined) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 14}},
			},
			{
				Code:   `function undefined() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 10}},
			},
			{
				Code:   `(function undefined(){}())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 11}},
			},
			{
				Code:   `var foo = function undefined() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 20}},
			},
			{
				Code:   `foo = function undefined() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 16}},
			},
			{
				Code:   `undefined = true`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 1}},
			},
			{
				Code:   `var undefined = true`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 5}},
			},
			{
				Code:   `({ undefined })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 4}},
			},
			{
				Code:   `({ [undefined]: foo })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 5}},
			},
			{
				Code:   `({ bar: undefined })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 9}},
			},
			{
				Code:   `({ bar: undefined } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 9}},
			},
			{
				Code:   `var { undefined } = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 7}},
			},
			{
				Code:   `var { bar: undefined } = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 12}},
			},
			{
				Code:   `({ undefined: function undefined() {} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 24}},
			},
			{
				Code:   `({ foo: function undefined() {} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 18}},
			},
			{
				Code:   `class Foo { [undefined]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 14}},
			},
			{
				Code:   `(class { [undefined]() {} })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 11}},
			},
			{
				Code: `var undefined = true; undefined = false;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 5},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 23},
				},
			},
			{
				Code:   `import undefined from 'foo'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 8}},
			},
			{
				Code:   `import * as undefined from 'foo'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 13}},
			},
			{
				Code:   `import { undefined } from 'foo'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 10}},
			},
			{
				Code:   `import { a as undefined } from 'foo'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 15}},
			},

			// NOTE: upstream also carries a commented-out case here —
			// `export { undefined }` (no `from`) — because acorn raises
			// "Export 'undefined' is not defined" at parse time for a local,
			// non-aliased export of a name with no local declaration. That is a
			// parser-level diagnostic outside this rule's scope, so it is not
			// migrated.

			{
				Code:   `let a = [b, ...undefined]`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 16}},
			},
			{
				Code:   `[a, ...undefined] = b`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 8}},
			},
			{
				Code:   `[a = undefined] = b`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 6}},
			},
		},
	)
}
