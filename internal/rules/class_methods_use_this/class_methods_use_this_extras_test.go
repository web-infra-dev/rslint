// TestClassMethodsUseThisExtras locks in branches and edge shapes that the
// upstream ESLint test suite does not exercise. Each case carries an inline
// comment naming the source branch, Dimension 4 row, or real-user issue it
// covers; the complete upstream migration lives in the sibling
// class_methods_use_this_upstream_*_test.go files.
//
// Dimension walk notes:
//   - Dimension 3 (autofix boundaries): N/A — this rule has no autofix or
//     suggestions.
//   - Dimension 4 (element access): N/A — the rule observes ThisKeyword and
//     SuperKeyword directly and does not interpret member-access receivers.
//   - Dimension 4 (SpreadAssignment / RestElement): N/A — object members and
//     binding patterns are outside the rule's class-member/function boundary.
package class_methods_use_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func coreOptions(options map[string]interface{}) []interface{} {
	return []interface{}{options}
}

func missingThisAt(message string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "missingThis",
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestClassMethodsUseThisExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ClassMethodsUseThisRule,
		[]rule_tester.ValidTestCase{
			// ---- Review regression: TypeScript type-query `this` counts as use ----
			{Code: `class C { foo(value: typeof this) {} }`},
			{Code: `class C { foo(): typeof this {} }`},
			{Code: `class C { foo<T extends typeof this>() {} }`},
			{Code: `class C { foo(value: Array<typeof this>) {} }`},
			{Code: `class C { foo = (value: typeof this) => {}; }`},
			{Code: `class C { foo(value: typeof this.value) {} }`},
			{Code: `class C { foo(value: typeof this.value.deep) {} }`},

			// A field decorator runs before the field value frame. Its `this`
			// therefore belongs to the enclosing method, as it does in ESTree.
			{
				Code: `class C { foo() { return class { @dec(this) field = () => {}; }; } }`,
				Options: coreOptions(map[string]interface{}{
					"enforceForClassFields": false,
				}),
			},

			// ---- Dimension 4: receiver wrappers — parenthesized this ----
			{Code: `class C { foo() { return (this).value; } }`},
			{Code: `class C { foo() { return ((this)).value; } }`},

			// ---- Dimension 4: receiver wrappers — TypeScript wrappers ----
			{Code: `class C { foo() { return this!.value; } }`},
			{Code: `class C { foo() { return (this as C).value; } }`},
			{Code: `class C { foo() { return this satisfies C; } }`},

			// ---- Dimension 4: receiver wrappers — optional chain ----
			{Code: `class C { foo() { return this?.value; } }`},
			{Code: `class C { foo() { return this?.method(); } }`},

			// ---- Dimension 4: access/key forms — static exception classes ----
			{
				Code:    `class C { foo() {} "bar"() {} 0() {} #baz() {} }`,
				Options: coreOptions(map[string]interface{}{"exceptMethods": []interface{}{"foo", "bar", "0", "#baz"}}),
			},

			// ---- Dimension 4: declaration/container forms — class expressions ----
			{Code: `const C = class { foo() { return this.value; } };`},
			{Code: `const C = class Named { foo() { return this.value; } };`},

			// ---- Dimension 4: declaration/container forms — async/generator variants ----
			{Code: `class C { async foo() { return this.value; } }`},
			{Code: `class C { *foo() { yield this.value; } }`},
			{Code: `class C { async *foo() { yield this.value; } }`},

			// ---- Dimension 4: nesting boundary — arrows inherit this ----
			{Code: `class C { foo() { return () => this.value; } }`},
			{Code: `class C { foo() { return () => () => this.value; } }`},

			// ---- Dimension 4: nesting boundary — nested class remains isolated ----
			{Code: `class C { foo() { this.value; return class { bar() { this.value; } }; } }`},

			// ---- Dimension 4: graceful degradation — empty/bodyless forms ----
			{Code: `class C {}`},
			{Code: `const C = class {};`},
			{Code: `abstract class C { abstract foo(): void; }`},
			{Code: `declare class C { foo(): void; field: number; }`},
			{Code: `class C { foo(value: string): void; foo(value: number): void; foo(value: unknown) { this.value = value; } }`},
			{Code: `class Outer { outer() { return class Inner { method(x: typeof this.x): void; }; } }`},
			{Code: `class Outer { outer() { return class Inner { static method(x: typeof this.x): void; }; } }`},
			{Code: `class Outer { outer() { abstract class Inner { abstract get value(): typeof this.x; } return Inner; } }`},
			{Code: `class Outer { outer() { function inner(x: typeof this.x): void; return inner; } }`},
			{Code: `class Outer { outer() { abstract class Inner { abstract value: typeof this.x; } return Inner; } }`},
			{Code: `class Outer { outer() { abstract class Inner { abstract [key]: typeof this.x; } return Inner; } }`},
			{Code: `class Outer { outer() { abstract class Inner { abstract accessor value: typeof this.x; } return Inner; } }`},

			// ---- Dimension 4: class-field function wrappers ----
			{Code: `class C { foo = (() => { this.value; }); }`},
			{Code: `class C { foo = (function () { this.value; }); }`},
			// Authored TypeScript wrappers remain visible in ESTree and therefore
			// keep the function from being a direct class-field value.
			{Code: `class C { foo = ((() => {}) as unknown); }`},
			{Code: `class C { foo = ((() => {}) satisfies unknown); }`},

			// ---- Real-user: eslint/eslint#17976 — explicit override opt-out ----
			{
				Code:    `class B extends A { override foo() {} }`,
				Options: coreOptions(map[string]interface{}{"ignoreOverrideMethods": true}),
			},

			// ---- Real-user: eslint/eslint#19507 — implements opt-out ----
			{
				Code:    `class C implements I { foo() {} field = () => {}; }`,
				Options: coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "all"}),
			},

			// Locks in isIncludedInstanceMethod() static and constructor arms.
			{Code: `class C { constructor() {} static foo() {} static get value() {} static set value(v) {} static { this.value = 1; } }`},

			// Locks in isIncludedInstanceMethod() enforceForClassFields=false arm.
			{
				Code:    `class C { foo = () => {}; accessor bar = function () {}; }`,
				Options: coreOptions(map[string]interface{}{"enforceForClassFields": false}),
			},

			// Locks in isIncludedInstanceMethod() exceptMethods identifier/literal/private arms.
			{
				Code:    `class C { foo() {} "bar"() {} 16() {} #baz() {} }`,
				Options: coreOptions(map[string]interface{}{"exceptMethods": []interface{}{"foo", "bar", "16", "#baz"}}),
			},

			// Locks in hasImplements() ClassExpression arm and "all" mode.
			{
				Code:    `const C = class implements I { foo() {} };`,
				Options: coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "all"}),
			},

			// Locks in public-fields: public method skipped, private one uses this.
			{
				Code:    `class C implements I { foo() {} private bar() { this.value; } }`,
				Options: coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}),
			},

			// Locks in exitFunction() usesThis and markThisUsed() Super arm.
			{Code: `class C extends B { foo() { super.foo(); } }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Review regression: JSDoc casts are transparent around class-field functions ----
			{
				Code:     `class A { field = /** @type {() => void} */ (() => {}); }`,
				FileName: "jsdoc-cast-arrow.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 46,
				)},
			},
			{
				Code:     `class A { field = /** @type {() => void} */ (function () {}); }`,
				FileName: "jsdoc-cast-function.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 55,
				)},
			},
			{
				Code:     `class A { field = /** @satisfies {() => void} */ (() => {}); }`,
				FileName: "jsdoc-satisfies-arrow.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 51,
				)},
			},
			{
				Code:     `class A { #field = /** @type {Function} */ (async () => {}); }`,
				FileName: "jsdoc-private-async-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class private async method #field.",
					1, 11, 1, 51,
				)},
			},
			{
				Code:     `class A { field = /** @type {Function} */ (/** @satisfies {Function} */ (() => {})); }`,
				FileName: "nested-jsdoc-casts.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 74,
				)},
			},
			{
				Code:     `class A { field = /** @type {(value: unknown) => void} */ (value => {}); }`,
				FileName: "jsdoc-single-param-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 59,
				)},
			},
			{
				Code:     `class A { field = /** @satisfies {(value: unknown) => void} */ (value => {}); }`,
				FileName: "jsdoc-satisfies-single-param-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 64,
				)},
			},
			{
				Code:     `class A { field = /** @type {(value: unknown) => void} */ ((value => {})); }`,
				FileName: "jsdoc-parenthesized-single-param-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 60,
				)},
			},
			{
				Code:     `class A { field = /** @type {(value: unknown) => Promise<void>} */ (async value => {}); }`,
				FileName: "jsdoc-async-single-param-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class async method 'field'.",
					1, 11, 1, 75,
				)},
			},
			{
				Code:     `class A { @dec field = /** @type {Function} */ (value => {}); }`,
				FileName: "decorated-jsdoc-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 48,
				)},
			},
			{
				Code:     `class A { @dec field = /** @type {Function} */ (() => {}); }`,
				FileName: "decorated-jsdoc-zero-param-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 49,
				)},
			},
			{
				Code:     `class A { @dec field = /** @type {Function} */ (function () {}); }`,
				FileName: "decorated-jsdoc-function-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class method 'field'.",
					1, 11, 1, 58,
				)},
			},
			{
				Code:     `class A { accessor field = /** @type {Function} */ (value => {}); }`,
				FileName: "jsdoc-auto-accessor.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt(
					"Expected 'this' to be used by class arrow function.",
					1, 59, 1, 61,
				)},
			},

			// Abstract properties do not create a frame, but the following
			// concrete method still does and reports independently.
			{
				Code:   `class Outer { outer() { abstract class Inner { abstract value: typeof this.x; method() {} } return Inner; } }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 1, 79, 1, 85)},
			},

			// A bodyless `declare` field remains a PropertyDefinition upstream,
			// so its type annotation stays isolated from the enclosing method.
			{
				Code:   `class Outer { outer() { return class Inner { declare value: typeof this.x; }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'outer'.", 1, 15, 1, 20)},
			},

			// ---- Review regression: JSDoc must not synthesize ESTree modifiers ----
			{
				Code:     "class Example {\n  /** @override */\n  method() {}\n}",
				FileName: "jsdoc-override.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  coreOptions(map[string]interface{}{"ignoreOverrideMethods": true}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 3, 3, 3, 9)},
			},
			{
				Code:     "class Example {\n  /** @override */\n  method() {}\n}",
				FileName: "jsdoc-override-checked.js",
				TSConfig: "tsconfig.allowJs.json",
				Options:  coreOptions(map[string]interface{}{"ignoreOverrideMethods": true}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 3, 3, 3, 9)},
			},
			{
				Code:     "class Example {\n  /** @override */\n  field = () => {};\n}",
				FileName: "jsdoc-override-field.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  coreOptions(map[string]interface{}{"ignoreOverrideMethods": true}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'field'.", 3, 3, 3, 11)},
			},
			{
				Code:     "class Example {\n  /** @override */\n  get value() {}\n}",
				FileName: "jsdoc-override-getter.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  coreOptions(map[string]interface{}{"ignoreOverrideMethods": true}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class getter 'value'.", 3, 3, 3, 12)},
			},

			// ---- Review regression: JSDoc must not synthesize ESTree heritage ----
			{
				Code:     "/** @implements {Contract} */\nclass Example {\n  method() {}\n}",
				FileName: "jsdoc-implements-all.mjs",
				TSConfig: "tsconfig.allow-js.json",
				Options:  coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "all"}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 3, 3, 3, 9)},
			},
			{
				Code:     "/** @implements {Contract} */\nclass Example {\n  method() {}\n}",
				FileName: "jsdoc-implements-all-checked.js",
				TSConfig: "tsconfig.allowJs.json",
				Options:  coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "all"}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 3, 3, 3, 9)},
			},
			{
				Code:     "/** @implements {Contract} */\nclass Example {\n  method() {}\n}",
				FileName: "jsdoc-implements-public-fields.mjs",
				TSConfig: "tsconfig.allow-js.json",
				Options:  coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 3, 3, 3, 9)},
			},
			{
				Code:     "/** @implements {Contract} */\nconst Example = class { method() {} };",
				FileName: "jsdoc-implements-class-expression.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "all"}),
				Errors:   []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'method'.", 2, 25, 2, 31)},
			},

			// A plain `this` type is TSThisType upstream, not the ThisExpression
			// operand of a type query, so it does not count as runtime `this` use.
			{
				Code:   `class C { foo(value: this) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 14)},
			},

			// ---- Review regression: decorators execute outside member frames ----
			{
				Code:   `class C { @dec(this) foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 25)},
			},
			{
				Code:   `class C extends B { @dec(super.value) foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 21, 1, 42)},
			},
			{
				Code:   `class C { @dec(() => this.value) get foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class getter 'foo'.", 1, 11, 1, 41)},
			},
			{
				Code:   `class C { @dec(this) set foo(value) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class setter 'foo'.", 1, 11, 1, 29)},
			},

			// Decorated field heads include the decorator in ESLint's range.
			{
				Code:   `class C { @dec(this) field = () => {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'field'.", 1, 11, 1, 30)},
			},

			// ---- Options contract: omitted options and [{}] have identical defaults ----
			{
				Code:   `class C { foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 14)},
			},
			{
				Code:    `class C { foo() {} }`,
				Options: coreOptions(map[string]interface{}{}),
				Errors:  []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 14)},
			},

			// ---- Dimension 4: declaration/container forms — multiline class expression ----
			{
				Code: `const C = class {
  foo() {}
};`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 2, 3, 2, 6)},
			},

			// ---- Dimension 4: access/key forms — computed names never match exceptions ----
			{
				Code:    `class C { [foo]() {} }`,
				Options: coreOptions(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors:  []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method.", 1, 11, 1, 16)},
			},
			{
				Code:    `class C { ["foo"]() {} }`,
				Options: coreOptions(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors:  []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 18)},
			},

			// ---- Dimension 4: async/generator/async-generator diagnostics ----
			{
				Code: `class C {
  async foo() {}
  *bar() {}
  async *baz() {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingThisAt("Expected 'this' to be used by class async method 'foo'.", 2, 3, 2, 12),
					missingThisAt("Expected 'this' to be used by class generator method 'bar'.", 3, 3, 3, 7),
					missingThisAt("Expected 'this' to be used by class async generator method 'baz'.", 4, 3, 4, 13),
				},
			},

			// ---- Dimension 4: nesting boundary — regular function owns this ----
			// ---- Real-user: eslint/eslint#13527 ----
			{
				Code: `class C {
  foo() {
    function nested() { return this.value; }
    nested();
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 2, 3, 2, 6)},
			},

			// ---- Dimension 4: nesting boundary — object method owns this ----
			{
				Code:   `class C { foo() { return { bar() { return this.value; } }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 14)},
			},

			// ---- Dimension 4: nesting boundary — static block owns this ----
			{
				Code:   `class C { foo() { return class { static { this.value; } }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 14)},
			},

			// ---- Dimension 4: same-kind nesting reports independently ----
			{
				Code: `class C { foo() { return class { bar() {} }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingThisAt("Expected 'this' to be used by class method 'bar'.", 1, 34, 1, 37),
					missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 14),
				},
			},

			// Locks in ignoreOverrideMethods=false and true matrix.
			{
				Code:    `class C extends B { override foo() {} }`,
				Options: coreOptions(map[string]interface{}{"ignoreOverrideMethods": false}),
				Errors:  []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 21, 1, 33)},
			},

			// Locks in ignoreClassesWithImplements=public-fields private/protected/# arms.
			{
				Code: `class C implements I {
  private foo() {}
  protected bar() {}
  #baz() {}
}`,
				Options: coreOptions(map[string]interface{}{"ignoreClassesWithImplements": "public-fields"}),
				Errors: []rule_tester.InvalidTestCaseError{
					missingThisAt("Expected 'this' to be used by class method 'foo'.", 2, 3, 2, 14),
					missingThisAt("Expected 'this' to be used by class method 'bar'.", 3, 3, 3, 16),
					missingThisAt("Expected 'this' to be used by class private method #baz.", 4, 3, 4, 7),
				},
			},

			// Locks in enforceForClassFields=true explicit boolean arm.
			{
				Code:    `class C { foo = () => {}; }`,
				Options: coreOptions(map[string]interface{}{"enforceForClassFields": true}),
				Errors:  []rule_tester.InvalidTestCaseError{missingThisAt("Expected 'this' to be used by class method 'foo'.", 1, 11, 1, 17)},
			},

			// Locks in private/static/dynamic name equivalence classes.
			{
				Code:    `class C { #foo() {} [foo]() {} }`,
				Options: coreOptions(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					missingThisAt("Expected 'this' to be used by class private method #foo.", 1, 11, 1, 15),
					missingThisAt("Expected 'this' to be used by class method.", 1, 21, 1, 26),
				},
			},
		},
	)
}
