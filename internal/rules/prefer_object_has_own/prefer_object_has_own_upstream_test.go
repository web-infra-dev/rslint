package prefer_object_has_own

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferObjectHasOwnUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/prefer-object-has-own.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// prefer_object_has_own_extras_test.go file.
func TestPreferObjectHasOwnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferObjectHasOwnRule,
		[]rule_tester.ValidTestCase{
			// ---- `Object` receiver ----
			{Code: `Object`},
			{Code: `Object(obj, prop)`},
			{Code: `Object.hasOwnProperty`},
			{Code: `Object.hasOwnProperty(prop)`},
			{Code: `hasOwnProperty(obj, prop)`},
			{Code: `foo.hasOwnProperty(prop)`},
			{Code: `foo.hasOwnProperty(obj, prop)`},
			{Code: `Object.hasOwnProperty.call`},
			{Code: `foo.Object.hasOwnProperty.call(obj, prop)`},
			{Code: `foo.hasOwnProperty.call(obj, prop)`},
			{Code: `foo.call(Object.prototype.hasOwnProperty, Object.prototype.hasOwnProperty.call)`},
			{Code: `Object.foo.call(obj, prop)`},
			{Code: `Object.hasOwnProperty.foo(obj, prop)`},
			{Code: `Object.hasOwnProperty.call.foo(obj, prop)`},
			{Code: `Object[hasOwnProperty].call(obj, prop)`},
			{Code: `Object.hasOwnProperty[call](obj, prop)`},
			{Code: `class C { #hasOwnProperty; foo() { Object.#hasOwnProperty.call(obj, prop) } }`},
			{Code: `class C { #call; foo() { Object.hasOwnProperty.#call(obj, prop) } }`},
			{Code: `(Object) => Object.hasOwnProperty.call(obj, prop)`}, // not global Object
			// ---- `Object.prototype` receiver ----
			{Code: `Object.prototype`},
			{Code: `Object.prototype(obj, prop)`},
			{Code: `Object.prototype.hasOwnProperty`},
			{Code: `Object.prototype.hasOwnProperty(obj, prop)`},
			{Code: `Object.prototype.hasOwnProperty.call`},
			{Code: `foo.Object.prototype.hasOwnProperty.call(obj, prop)`},
			{Code: `foo.prototype.hasOwnProperty.call(obj, prop)`},
			{Code: `Object.foo.hasOwnProperty.call(obj, prop)`},
			{Code: `Object.prototype.foo.call(obj, prop)`},
			{Code: `Object.prototype.hasOwnProperty.foo(obj, prop)`},
			{Code: `Object.prototype.hasOwnProperty.call.foo(obj, prop)`},
			{Code: `Object.prototype.prototype.hasOwnProperty.call(a, b);`},
			{Code: `Object.hasOwnProperty.prototype.hasOwnProperty.call(a, b);`},
			{Code: `Object.prototype[hasOwnProperty].call(obj, prop)`},
			{Code: `Object.prototype.hasOwnProperty[call](obj, prop)`},
			{Code: `class C { #hasOwnProperty; foo() { Object.prototype.#hasOwnProperty.call(obj, prop) } }`},
			{Code: `class C { #call; foo() { Object.prototype.hasOwnProperty.#call(obj, prop) } }`},
			{Code: `Object[prototype].hasOwnProperty.call(obj, prop)`},
			{Code: `class C { #prototype; foo() { Object.#prototype.hasOwnProperty.call(obj, prop) } }`},
			{Code: `(Object) => Object.prototype.hasOwnProperty.call(obj, prop)`}, // not global Object
			// ---- object-literal receiver ----
			{Code: `({})`},
			{Code: `({}(obj, prop))`},
			{Code: `({}.hasOwnProperty)`},
			{Code: `({}.hasOwnProperty(prop))`},
			{Code: `({}.hasOwnProperty(obj, prop))`},
			{Code: `({}.hasOwnProperty.call)`},
			{Code: `({}).prototype.hasOwnProperty.call(a, b);`},
			{Code: `({}.foo.call(obj, prop))`},
			{Code: `({}.hasOwnProperty.foo(obj, prop))`},
			{Code: `({}[hasOwnProperty].call(obj, prop))`},
			{Code: `({}.hasOwnProperty[call](obj, prop))`},
			{Code: `({}).hasOwnProperty[call](object, property)`},
			{Code: `({})[hasOwnProperty].call(object, property)`},
			{Code: `class C { #hasOwnProperty; foo() { ({}.#hasOwnProperty.call(obj, prop)) } }`},
			{Code: `class C { #call; foo() { ({}.hasOwnProperty.#call(obj, prop)) } }`},
			{Code: `({ foo }.hasOwnProperty.call(obj, prop))`},        // object literal should be empty
			{Code: `(Object) => ({}).hasOwnProperty.call(obj, prop)`}, // Object is shadowed, so Object.hasOwn cannot be used here
			// ---- already using `Object.hasOwn` ----
			{Code: `
        let obj = {};
        Object.hasOwn(obj,"");
        `},
			{Code: `const hasProperty = Object.hasOwn(object, property);`},
			{Code: `/* global Object: off */
        ({}).hasOwnProperty.call(a, b);`}},
		[]rule_tester.InvalidTestCase{
			// ---- dotted receivers ----
			{
				Code:   `Object.hasOwnProperty.call(obj, 'foo')`,
				Output: []string{`Object.hasOwn(obj, 'foo')`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 39,
					},
				},
			},
			{
				Code:   `Object.hasOwnProperty.call(obj, property)`,
				Output: []string{`Object.hasOwn(obj, property)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 42,
					},
				},
			},
			{
				Code:   `Object.prototype.hasOwnProperty.call(obj, 'foo')`,
				Output: []string{`Object.hasOwn(obj, 'foo')`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 49,
					},
				},
			},
			{
				Code:   `({}).hasOwnProperty.call(obj, 'foo')`,
				Output: []string{`Object.hasOwn(obj, 'foo')`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 37,
					},
				},
			},
			// prevent autofixing if there are any comments
			{
				Code: `Object/* comment */.prototype.hasOwnProperty.call(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 56,
					},
				},
			},
			// ---- parenthesized receivers ----
			{
				Code:   `const hasProperty = Object.prototype.hasOwnProperty.call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 75,
					},
				},
			},
			{
				Code:   `const hasProperty = (( Object.prototype.hasOwnProperty.call(object, property) ));`,
				Output: []string{`const hasProperty = (( Object.hasOwn(object, property) ));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 78,
					},
				},
			},
			{
				Code:   `const hasProperty = (( Object.prototype.hasOwnProperty.call ))(object, property);`,
				Output: []string{`const hasProperty = (( Object.hasOwn ))(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 81,
					},
				},
			},
			{
				Code:   `const hasProperty = (( Object.prototype.hasOwnProperty )).call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 81,
					},
				},
			},
			{
				Code:   `const hasProperty = (( Object.prototype )).hasOwnProperty.call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 81,
					},
				},
			},
			{
				Code:   `const hasProperty = (( Object )).prototype.hasOwnProperty.call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 81,
					},
				},
			},
			{
				Code:   `const hasProperty = {}.hasOwnProperty.call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 61,
					},
				},
			},
			{
				Code:   `const hasProperty={}.hasOwnProperty.call(object, property);`,
				Output: []string{`const hasProperty=Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    19,
						EndLine:   1,
						EndColumn: 59,
					},
				},
			},
			{
				Code:   `const hasProperty = (( {}.hasOwnProperty.call(object, property) ));`,
				Output: []string{`const hasProperty = (( Object.hasOwn(object, property) ));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 64,
					},
				},
			},
			{
				Code:   `const hasProperty = (( {}.hasOwnProperty.call ))(object, property);`,
				Output: []string{`const hasProperty = (( Object.hasOwn ))(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 67,
					},
				},
			},
			{
				Code:   `const hasProperty = (( {}.hasOwnProperty )).call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 67,
					},
				},
			},
			{
				Code:   `const hasProperty = (( {} )).hasOwnProperty.call(object, property);`,
				Output: []string{`const hasProperty = Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 67,
					},
				},
			},
			{
				Code:   `function foo(){return {}.hasOwnProperty.call(object, property)}`,
				Output: []string{`function foo(){return Object.hasOwn(object, property)}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    23,
						EndLine:   1,
						EndColumn: 63,
					},
				},
			},
			// ---- the replacement must not fuse with the token in front of it ----
			{
				Code:   `function foo(){return{}.hasOwnProperty.call(object, property)}`,
				Output: []string{`function foo(){return Object.hasOwn(object, property)}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 62,
					},
				},
			},
			{
				Code:   `function foo(){return/*comment*/{}.hasOwnProperty.call(object, property)}`,
				Output: []string{`function foo(){return/*comment*/Object.hasOwn(object, property)}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    33,
						EndLine:   1,
						EndColumn: 73,
					},
				},
			},
			{
				Code:   `async function foo(){return await{}.hasOwnProperty.call(object, property)}`,
				Output: []string{`async function foo(){return await Object.hasOwn(object, property)}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    34,
						EndLine:   1,
						EndColumn: 74,
					},
				},
			},
			{
				Code:   `async function foo(){return await/*comment*/{}.hasOwnProperty.call(object, property)}`,
				Output: []string{`async function foo(){return await/*comment*/Object.hasOwn(object, property)}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    45,
						EndLine:   1,
						EndColumn: 85,
					},
				},
			},
			{
				Code:   `for (const x of{}.hasOwnProperty.call(object, property).toString());`,
				Output: []string{`for (const x of Object.hasOwn(object, property).toString());`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 56,
					},
				},
			},
			{
				Code:   `for (const x of/*comment*/{}.hasOwnProperty.call(object, property).toString());`,
				Output: []string{`for (const x of/*comment*/Object.hasOwn(object, property).toString());`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 67,
					},
				},
			},
			{
				Code:   `for (const x in{}.hasOwnProperty.call(object, property).toString());`,
				Output: []string{`for (const x in Object.hasOwn(object, property).toString());`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 56,
					},
				},
			},
			{
				Code:   `for (const x in/*comment*/{}.hasOwnProperty.call(object, property).toString());`,
				Output: []string{`for (const x in/*comment*/Object.hasOwn(object, property).toString());`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 67,
					},
				},
			},
			{
				Code:   `function foo(){return({}.hasOwnProperty.call)(object, property)}`,
				Output: []string{`function foo(){return(Object.hasOwn)(object, property)}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 64,
					},
				},
			},
			// ---- computed receivers ----
			{
				Code:   `Object['prototype']['hasOwnProperty']['call'](object, property);`,
				Output: []string{`Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 64,
					},
				},
			},
			{
				Code:   "Object[`prototype`][`hasOwnProperty`][`call`](object, property);",
				Output: []string{`Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 64,
					},
				},
			},
			{
				Code:   `Object['hasOwnProperty']['call'](object, property);`,
				Output: []string{`Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 51,
					},
				},
			},
			{
				Code:   "Object[`hasOwnProperty`][`call`](object, property);",
				Output: []string{`Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 51,
					},
				},
			},
			{
				Code:   `({})['hasOwnProperty']['call'](object, property);`,
				Output: []string{`Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 49,
					},
				},
			},
			{
				Code:   "({})[`hasOwnProperty`][`call`](object, property);",
				Output: []string{`Object.hasOwn(object, property);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 49,
					},
				},
			}},
	)
}
