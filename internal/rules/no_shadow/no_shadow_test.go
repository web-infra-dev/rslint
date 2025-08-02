package no_shadow

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoShadowRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoShadowRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
var a = 3;
function b() {
    var c = a;
}`,
			},
			{
				Code: `
function a() {}
function b() {
    var a = 10;
}`,
				Options: map[string]interface{}{
					"hoist": "never",
				},
			},
			{
				Code: `
var a = 3;
function b() {
    var a = 10;
}`,
				Options: map[string]interface{}{
					"allow": []interface{}{"a"},
				},
			},
			{
				Code: `
function foo() {
    var Object = 0;
}`,
				Options: map[string]interface{}{
					"builtinGlobals": false,
				},
			},
			{
				Code: `
type Foo = string;
function test() {
    const Foo = 1;
}`,
				Options: map[string]interface{}{
					"ignoreTypeValueShadow": true,
				},
			},
			{
				Code: `
interface Foo {}
class Bar {
    Foo: string;
}`,
				Options: map[string]interface{}{
					"ignoreTypeValueShadow": true,
				},
			},
			{
				Code: `let test = 1; type TestType = typeof test; type Func = (test: string) => typeof test;`,
				Options: map[string]interface{}{
					"ignoreFunctionTypeParameterNameValueShadow": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
var a = 3;
function b() {
    var a = 10;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
					},
				},
			},
			{
				Code: `
function foo() {
    var Object = 0;
}`,
				Options: map[string]interface{}{
					"builtinGlobals": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadowGlobal",
					},
				},
			},
			{
				Code: `
var a = 3;
var a = 10;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
					},
				},
			},
			{
				Code: `
var a = 3;
function b() {
    function a() {}
}`,
				Options: map[string]interface{}{
					"hoist": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
					},
				},
			},
			{
				Code: `
type Foo = string;
function test() {
    const Foo = 1;
}`,
				Options: map[string]interface{}{
					"ignoreTypeValueShadow": false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
					},
				},
			},
			{
				Code: `
function foo(a: number) {
    function bar() {
        function baz(a: string) {}
    }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
					},
				},
			},
			{
				Code: `
class A {
    method() {
        class A {}
    }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
					},
				},
			},
		},
	)
}