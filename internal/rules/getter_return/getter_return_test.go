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
				Code:    `var foo = { get bar() {return;} };`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code:    `var foo = { get bar(){return true;} };`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code:    `var foo = { get bar(){if(bar) {return;} return true;} };`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code:    `class foo { get bar(){return true;} }`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code:    `class foo { get bar(){return;} }`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},

			// Throw statements as valid exit paths
			{Code: `var foo = { get bar(){ throw new Error("not implemented"); } };`},
			{Code: `class foo { get bar(){ if(baz) { throw new Error(); } return true; } }`},
			{Code: `class foo { get bar(){ if(baz) { return true; } else { throw new Error(); } } }`},
			{Code: `class foo { get bar(){ if(baz) { throw new Error(); } else { throw new Error(); } } }`},

			// Try/catch with return
			{Code: `class foo { get bar(){ try { return 1; } catch(e) { return 2; } } }`},
			{Code: `class foo { get bar(){ try { return 1; } catch(e) { throw e; } } }`},
			{Code: `class foo { get bar(){ try { throw new Error(); } catch(e) { return 1; } } }`},
			{Code: `class foo { get bar(){ try { return 1; } finally { } } }`},

			// Switch with return
			{Code: `class foo { get bar(){ switch(x) { case 1: return 1; default: return 2; } } }`},
			{Code: `class foo { get bar(){ switch(x) { case 1: return 1; case 2: return 2; default: throw new Error(); } } }`},

			// Object.defineProperty with throw
			{Code: `Object.defineProperty(foo, "bar", { get: function () { throw new Error("not implemented"); }});`},

			// Arrow functions with expression body are implicitly returning (not checked by ESLint)
			{Code: `Object.defineProperty(foo, "bar", { get: () => true });`},
			{Code: `Object.defineProperty(foo, "bar", { get: () => foo.bar });`},
			{Code: `Object.create(foo, { bar: { get: () => true } });`},
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

			// if-throw without else (not all paths covered)
			{
				Code: `class foo { get bar(){ if(baz) { throw new Error(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expected",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Switch without default (not all paths covered)
			{
				Code: `class foo { get bar(){ switch(x) { case 1: return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expectedAlways",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Try/catch where not all paths return
			{
				Code: `class foo { get bar(){ try { return 1; } catch(e) { } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expectedAlways",
						Line:      1,
						Column:    13,
					},
				},
			},

			// finally { return; } overrides try return, getter returns undefined
			{
				Code: `class foo { get bar(){ try { return 1; } finally { return; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "expectedAlways",
						Line:      1,
						Column:    13,
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
