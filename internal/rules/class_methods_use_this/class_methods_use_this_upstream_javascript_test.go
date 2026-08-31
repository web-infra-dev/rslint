// TestClassMethodsUseThisUpstreamJavaScript migrates the full JavaScript-parser valid/invalid suite from
// upstream ESLint v10.9.0 tests/lib/rules/class-methods-use-this.js 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific
// branch and edge-shape lock-ins live in class_methods_use_this_extras_test.go.
package class_methods_use_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestClassMethodsUseThisUpstreamJavaScript(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ClassMethodsUseThisRule,
		[]rule_tester.ValidTestCase{
			{Code: "class A { constructor() {} }"},
			{Code: "class A { foo() {this} }"},
			{Code: "class A { foo() {this.bar = 'bar';} }"},
			{Code: "class A { foo() {bar(this);} }"},
			{Code: "class A extends B { foo() {super.foo();} }"},
			{Code: "class A { foo() { if(true) { return this; } } }"},
			{Code: "class A { static foo() {} }"},
			{Code: "({ a(){} });"},
			{Code: "class A { foo() { () => this; } }"},
			{Code: "({ a: function () {} });"},
			{Code: "class A { foo() {this} bar() {} }", Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"bar"}}}},
			{Code: "class A { \"foo\"() { } }", Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"foo"}}}},
			{Code: "class A { 42() { } }", Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"42"}}}},
			{Code: "class A { foo = function() {this} }"},
			{Code: "class A { foo = () => {this} }"},
			{Code: "class A { foo = () => {super.toString} }"},
			{Code: "class A { static foo = function() {} }"},
			{Code: "class A { static foo = () => {} }"},
			{Code: "class A { #bar() {} }", Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"#bar"}}}},
			{Code: "class A { foo = function () {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { foo = () => {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { foo() { return class { [this.foo] = 1 }; } }"},
			{Code: "class A { static {} }"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "class A { foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "class A { foo() {/**this**/} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "class A { foo() {var a = function () {this};} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "class A { foo() {var a = function () {var b = function(){this}};} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "class A { foo() {window.this} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "class A { foo() {that.this = 'this';} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "class A { foo() { () => undefined; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "class A { foo() {} bar() {} }",
				Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"bar"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "class A { foo() {} hasOwnProperty() {} }",
				Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'hasOwnProperty'.", Line: 1, Column: 20, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "class A { [foo]() {} }",
				Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "class A { #foo() { } foo() {} #bar() {} }",
				Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"#foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #bar.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: "class A { foo(){} 'bar'(){} 123(){} [`baz`](){} [a](){} [f(a)](){} get quux(){} set[a](b){} *quuux(){} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'bar'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 24},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method '123'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 32},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'baz'.", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method.", Line: 1, Column: 49, EndLine: 1, EndColumn: 52},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method.", Line: 1, Column: 57, EndLine: 1, EndColumn: 63},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'quux'.", Line: 1, Column: 68, EndLine: 1, EndColumn: 76},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter.", Line: 1, Column: 81, EndLine: 1, EndColumn: 87},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class generator method 'quuux'.", Line: 1, Column: 93, EndLine: 1, EndColumn: 99},
				},
			},
			{
				Code: "class A { foo = function() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: "class A { foo = () => {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: "class A { #foo = function() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #foo.", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: "class A { #foo = () => {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #foo.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "class A { #foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #foo.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: "class A { get #foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private getter #foo.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "class A { set #foo(x) {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private setter #foo.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "class A { foo () { return class { foo = this }; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: "class A { foo () { return function () { foo = this }; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: "class A { foo () { return class { static { this; } } } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
		},
	)
}
