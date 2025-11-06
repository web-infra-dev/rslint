package getter_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestGetterReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&GetterReturnRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Object getters with return
			{Code: `var foo = { get bar(){return true;} };`},

			// Class getters with return
			{Code: `class foo { get bar(){return true;} }`},
			{Code: `class foo { get bar(){if(baz){return true;} else {return false;} } }`},
			{Code: `class foo { get(){return true;} }`},

			// Object.defineProperty with getter returning value
			{Code: `Object.defineProperty(foo, "bar", { get: function () {return true;}});`},
			{Code: `Object.defineProperty(foo, "bar", { get: function () { ~function (){ return true; }();return true;}});`},

			// Object.defineProperties
			{Code: `Object.defineProperties(foo, { bar: { get: function () {return true;}} });`},

			// Reflect.defineProperty
			{Code: `Reflect.defineProperty(foo, "bar", { get: function () {return true;}});`},

			// Object.create
			{Code: `Object.create(foo, { bar: { get() {return true;} } });`},
			{Code: `Object.create(foo, { bar: { get: function () {return true;} } });`},

			// Non-getter functions
			{Code: `var get = function(){};`},
			{Code: `var foo = { bar(){} };`},
			{Code: `var foo = { get: function () {} }`},

			// With allowImplicit option
			{
				Code: `var foo = { get bar() {return;} };`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `var foo = { get bar(){return true;} };`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `var foo = { get bar(){if(bar) {return;} return true;} };`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `class foo { get bar(){return true;} }`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `class foo { get bar(){return;} }`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Object getters without return
			{
				Code: `var foo = { get bar() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `var foo = { get bar(){if(baz) {return true;}} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expectedAlways",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `var foo = { get bar() { return; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Class getters without return
			{
				Code: `class foo { get bar(){} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `class foo { get bar(){ if (baz) { return true; }}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expectedAlways",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Object.defineProperty without return
			{
				Code: `Object.defineProperty(foo, 'bar', { get: function (){}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    37,
					},
				},
			},
			{
				Code: `Object.defineProperty(foo, 'bar', { get(){} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    37,
					},
				},
			},

			// Optional chaining (ES2020)
			{
				Code: `Object?.defineProperty(foo, 'bar', { get: function (){} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    38,
					},
				},
			},
			{
				Code: `(Object?.defineProperty)(foo, 'bar', { get: function (){} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    40,
					},
				},
			},
		},
	)
}
