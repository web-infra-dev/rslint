// TestClassMethodsUseThisUpstream migrates the full valid/invalid suite from
// upstream typescript-eslint's
//   packages/eslint-plugin/tests/rules/class-methods-use-this/class-methods-use-this.test.ts
//   packages/eslint-plugin/tests/rules/class-methods-use-this/class-methods-use-this-core.test.ts
// 1:1. Position assertions cover line/column for every invalid case (the
// upstream typescript-eslint suite only asserts messageId on most cases,
// so this layer adds line/column to satisfy the rslint requirement that
// invalid cases pin position; the core suite already locks them down).
//
// rslint-specific lock-in cases (Dimension 4 edge shapes, branch lock-ins,
// real-user issue shapes) live in class_methods_use_this_extras_test.go.
package class_methods_use_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// objectOption is the array-wrapped single-option shape that matches
// rule_tester's JSON path through utils.GetOptionsMap — the typed-struct
// shortcut would silently bypass it. See PORT_RULE.md Phase 2 Step 4.
func objectOption(opts map[string]interface{}) []interface{} {
	return []interface{}{opts}
}

func TestClassMethodsUseThisUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ClassMethodsUseThisRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// typescript-eslint specific: ignoreClassesThatImplementAnInterface / ignoreOverrideMethods
			// ============================================================
			{
				Code: `
class Foo implements Bar {
  method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},
			{
				Code: `
class Foo implements Bar {
  accessor method = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},
			{
				Code: `
class Foo implements Bar {
  get getter() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},
			{
				Code: `
class Foo implements Bar {
  set setter() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},
			{
				Code: `
class Foo {
  override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  private override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  protected override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  override accessor method = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  private override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  protected override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  private override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  protected override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo implements Bar {
  override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": true,
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  private override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  protected override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": true,
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  private override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  protected override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": true,
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  private override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  protected override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},
			{
				Code: `
class Foo {
  override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  private override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo {
  protected override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
			},
			{
				Code: `
class Foo implements Bar {
  override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": true,
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{
					"enforceForClassFields":                 false,
					"ignoreClassesThatImplementAnInterface": false,
				}),
			},
			{
				Code: `
class Foo {
  override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{
					"enforceForClassFields": false,
					"ignoreOverrideMethods": false,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  private override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo implements Bar {
  protected override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 true,
				}),
			},
			{
				Code: `
class Foo {
  accessor method = () => {
    this;
  };
}
      `,
			},
			{
				Code: `
class Foo {
  accessor method = function () {
    this;
  };
}
      `,
			},

			// ============================================================
			// Core ESLint parity (forked into typescript-eslint core test file)
			// ============================================================
			{Code: `class A { constructor() {} }`},
			{Code: `class A { foo() {this} }`},
			{Code: `class A { foo() {this.bar = 'bar';} }`},
			{Code: `class A { foo() {bar(this);} }`},
			{Code: `class A extends B { foo() {super.foo();} }`},
			{Code: `class A { foo() { if(true) { return this; } } }`},
			{Code: `class A { static foo() {} }`},
			{Code: `({ a(){} });`},
			{Code: `class A { foo() { () => this; } }`},
			{Code: `({ a: function () {} });`},
			{
				Code:    `class A { foo() {this} bar() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"bar"}}),
			},
			{
				Code:    `class A { "foo"() { } }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
			},
			{
				Code:    `class A { 42() { } }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"42"}}),
			},
			{Code: `class A { foo = function() {this} }`},
			{Code: `class A { foo = () => {this} }`},
			{Code: `class A { foo = () => {super.toString} }`},
			{Code: `class A { static foo = function() {} }`},
			{Code: `class A { static foo = () => {} }`},
			{
				Code:    `class A { #bar() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"#bar"}}),
			},
			{
				Code:    `class A { foo = function () {} }`,
				Options: objectOption(map[string]interface{}{"enforceForClassFields": false}),
			},
			{
				Code:    `class A { foo = () => {} }`,
				Options: objectOption(map[string]interface{}{"enforceForClassFields": false}),
			},
			{Code: `class A { foo() { return class { [this.foo] = 1 }; } }`},
			{Code: `class A { static {} }`},
		},

		[]rule_tester.InvalidTestCase{
			// ============================================================
			// typescript-eslint specific: ignoreClassesThatImplementAnInterface / ignoreOverrideMethods
			// ============================================================
			{
				Code: `
class Foo {
  method() {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code: `
class Foo {
  private method() {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 17},
				},
			},
			{
				Code: `
class Foo {
  protected method() {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 19},
				},
			},
			{
				Code: `
class Foo {
  accessor method = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  private accessor method = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  protected accessor method = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  #method() {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code: `
class Foo {
  get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 13},
				},
			},
			{
				Code: `
class Foo {
  private get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 21},
				},
			},
			{
				Code: `
class Foo {
  protected get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code: `
class Foo {
  get #getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 14},
				},
			},
			{
				Code: `
class Foo {
  set setter(b: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": false,
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 13},
				},
			},
			{
				Code: `
class Foo {
  private set setter(b: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 21},
				},
			},
			{
				Code: `
class Foo {
  protected set setter(b: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code: `
class Foo {
  set #setter(b: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3, EndLine: 3, EndColumn: 14},
				},
			},
			{
				Code: `
class Foo implements Bar {
  method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  #method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  private method() {}
}
      `,
				// interface cannot have private/protected modifier on members;
				// "public-fields" only ignores public members.
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  protected method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  get #getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  private get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  protected get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  set #setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  private set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  protected set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  override method() {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": false,
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  override get getter(): number {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": false,
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  override set setter(v: number) {}
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": false,
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  #property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo {
  override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": false}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  override property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": false,
					"ignoreOverrideMethods":                 false,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  private property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
class Foo implements Bar {
  protected property = () => {};
}
      `,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 3, Column: 3},
				},
			},
			{
				Code: `
function fn() {
  this.foo = 303;

  class Foo {
    method() {}
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Line: 6, Column: 5},
				},
			},

			// ============================================================
			// Core ESLint parity (forked into typescript-eslint core test file)
			// ============================================================
			{
				Code: `class A { foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class A { foo() {/**this**/} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class A { foo() {var a = function () {this};} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class A { foo() {var a = function () {var b = function(){this}};} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class A { foo() {window.this} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class A { foo() {that.this = 'this';} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class A { foo() { () => undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code:    `class A { foo() {} bar() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"bar"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code:    `class A { foo() {} hasOwnProperty() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'hasOwnProperty'.",
						Line:      1, Column: 20,
					},
				},
			},
			{
				Code:    `class A { [foo]() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code:    `class A { #foo() { } foo() {} #bar() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"#foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 22,
					},
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private method '#bar'.",
						Line:      1, Column: 31,
					},
				},
			},
			{
				Code: "class A { foo(){} 'bar'(){} 123(){} [`baz`](){} [a](){} [f(a)](){} get quux(){} set[a](b){} *quuux(){} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'foo'.", Column: 11},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'bar'.", Column: 19},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method '123'.", Column: 29},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method 'baz'.", Column: 37},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method.", Column: 49},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class method.", Column: 57},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class getter 'quux'.", Column: 68},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class setter.", Column: 81},
					{MessageId: "missingThis", Message: "Expected 'this' to be used by class generator method 'quuux'.", Column: 93},
				},
			},
			{
				Code: `class A { foo = function() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 25,
					},
				},
			},
			{
				Code: `class A { foo = () => {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 17,
					},
				},
			},
			{
				Code: `class A { #foo = function() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private method '#foo'.",
						Line:      1, Column: 11, EndColumn: 26,
					},
				},
			},
			{
				Code: `class A { #foo = () => {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private method '#foo'.",
						Line:      1, Column: 11, EndColumn: 18,
					},
				},
			},
			{
				Code: `class A { #foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private method '#foo'.",
						Line:      1, Column: 11, EndColumn: 15,
					},
				},
			},
			{
				Code: `class A { get #foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private getter '#foo'.",
						Line:      1, Column: 11, EndColumn: 19,
					},
				},
			},
			{
				Code: `class A { set #foo(x) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private setter '#foo'.",
						Line:      1, Column: 11, EndColumn: 19,
					},
				},
			},
			{
				Code: `class A { foo () { return class { foo = this }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 15,
					},
				},
			},
			{
				Code: `class A { foo () { return function () { foo = this }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 15,
					},
				},
			},
			{
				Code: `class A { foo () { return class { static { this; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 15,
					},
				},
			},
		},
	)
}
