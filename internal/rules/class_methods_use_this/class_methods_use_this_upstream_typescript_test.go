// TestClassMethodsUseThisUpstreamTypeScript migrates the full TypeScript-parser valid/invalid suite from
// upstream ESLint v10.9.0 tests/lib/rules/class-methods-use-this.js 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific
// branch and edge-shape lock-ins live in class_methods_use_this_extras_test.go.
package class_methods_use_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestClassMethodsUseThisUpstreamTypeScript(t *testing.T) {
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
			{Code: "class A { accessor foo = function() {this} }"},
			{Code: "class A { accessor foo = () => {this} }"},
			{Code: "class A { accessor foo = 1; }"},
			{Code: "class A { foo = () => {super.toString} }"},
			{Code: "class A { static foo = function() {} }"},
			{Code: "class A { static foo = () => {} }"},
			{Code: "class A { static accessor foo = function() {} }"},
			{Code: "class A { static accessor foo = () => {} }"},
			{Code: "class A { #bar() {} }", Options: []interface{}{map[string]interface{}{"exceptMethods": []interface{}{"#bar"}}}},
			{Code: "class A { foo = function () {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { foo = () => {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { accessor foo = function () {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { accessor foo = () => {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { override foo = () => {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class Foo implements Bar { property = () => {} }", Options: []interface{}{map[string]interface{}{"enforceForClassFields": false}}},
			{Code: "class A { foo() { return class { [this.foo] = 1 }; } }"},
			{Code: "class A { static {} }"},
			{Code: "class Foo { override method() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { override [\"method\"]() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { private override method() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { protected override method() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { override accessor method = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { override get getter(): number {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { private override get getter(): number {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { protected override get getter(): number {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { override set setter(v: number) {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { private override set setter(v: number) {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { protected override set setter(v: number) {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo implements Bar { override method() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { private override method() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { protected override method() {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { override get getter(): number {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { private override get getter(): number {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { protected override get getter(): number {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { override set setter(v: number) {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { private override set setter(v: number) {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { protected override set setter(v: number) {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo { override property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { private override property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo { protected override property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true}}},
			{Code: "class Foo implements Bar { override property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { private override property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { protected override property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreOverrideMethods": true, "ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { method() {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { [\"method\"]() {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { accessor method = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { get getter() {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { set setter() {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "class Foo implements Bar { property = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "const Foo = class implements Bar { method() {} };", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "const Foo = class implements Bar { property = () => {} };", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "all"}}},
			{Code: "const Foo = class implements Bar { method() {} };", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}}},
			{Code: "class Foo implements Bar { [\"property\"] = () => {} }", Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "\n  class Foo {\n    method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 11},
				},
			},
			{
				Code: "\n  class Foo {\n    private method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 19},
				},
			},
			{
				Code: "\n  class Foo {\n    protected method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 21},
				},
			},
			{
				Code: "\n  class Foo {\n\taccessor method = function () {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class function.", Line: 3, Column: 20, EndLine: 3, EndColumn: 29},
				},
			},
			{
				Code: "\n  class Foo {\n    accessor method = () => {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class arrow function.", Line: 3, Column: 26, EndLine: 3, EndColumn: 28},
				},
			},
			{
				Code: "\n  class Foo {\n    private accessor method = () => {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class arrow function.", Line: 3, Column: 34, EndLine: 3, EndColumn: 36},
				},
			},
			{
				Code: "\n  class Foo {\n    protected accessor method = () => {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class arrow function.", Line: 3, Column: 36, EndLine: 3, EndColumn: 38},
				},
			},
			{
				Code: "\n\tclass A {\n\t\tfoo() {\n\t\t\treturn class {\n\t\t\t\taccessor bar = this;\n\t\t\t};\n\t\t}\n\t}\n\t\t\t",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Line: 3, Column: 3, EndLine: 3, EndColumn: 6},
				},
			},
			{
				Code: "\n  class Derived extends Base {\n    override method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 20},
				},
			},
			{
				Code: "\n  class Derived extends Base {\n    property = () => {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code: "\n  class Derived extends Base {\n    public property = () => {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code: "\n  class Derived extends Base {\n    override property = () => {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo {\n    #method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #method.", Line: 3, Column: 5, EndLine: 3, EndColumn: 12},
				},
			},
			{
				Code: "\n  class Foo {\n    get getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 15},
				},
			},
			{
				Code: "\n  class Foo {\n    private get getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code: "\n  class Foo {\n    protected get getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo {\n    get #getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private getter #getter.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code: "\n  class Foo {\n    private set setter(b: number) {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code: "\n  class Foo {\n    protected set setter(b: number) {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo {\n    set #setter(b: number) {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private setter #setter.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code: "\n  function fn() {\n    this.foo = 303;\n\n    class Foo {\n      method() {}\n    }\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 6, Column: 7, EndLine: 6, EndColumn: 13},
				},
			},
			{
				Code: "\n  class Foo {\n    override method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 20},
				},
			},
			{
				Code: "\n  class Foo {\n    override get getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code: "\n  class Foo {\n    override set setter(v: number) {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    override method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 20},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    override get getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    override set setter(v: number) {}\n  }\n            ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code: "\n  class Foo {\n    override property = () => {};\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    override property = () => {};\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 11},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    #method() {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #method.", Line: 3, Column: 5, EndLine: 3, EndColumn: 12},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    #method() {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #method.", Line: 3, Column: 5, EndLine: 3, EndColumn: 12},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    private method() {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 19},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    protected method() {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 21},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    get getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 15},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    get #getter(): number {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private getter #getter.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    get #getter(): number {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private getter #getter.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    private get getter(): number {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    protected get getter(): number {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'getter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    set setter(v: number) {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 15},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    set #setter(v: number) {}\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private setter #setter.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    set #setter(v: number) {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private setter #setter.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    private set setter(v: number) {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    protected set setter(v: number) {}\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter 'setter'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 25},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    property = () => {};\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code: "\n  class Foo implements Bar {\n    #property = () => {};\n  }\n        ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #property.", Line: 3, Column: 5, EndLine: 3, EndColumn: 17},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    #property = () => {};\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class private method #property.", Line: 3, Column: 5, EndLine: 3, EndColumn: 17},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    private property = () => {};\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code:    "\n  class Foo implements Bar {\n    protected property = () => {};\n  }\n        ",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'property'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 26},
				},
			},
			{
				Code: "const Foo = class implements Bar { method() {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    "const Foo = class implements Bar { private method() {} };",
				Options: []interface{}{map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'method'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 50},
				},
			},
		},
	)
}
