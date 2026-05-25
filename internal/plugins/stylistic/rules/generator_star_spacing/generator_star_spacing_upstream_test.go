// TestGeneratorStarSpacingUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/generator-star-spacing/generator-star-spacing.test.ts
// 1:1. Position assertions cover line/column for every invalid case (all
// upstream cases are single-line; the * report span is one byte, so endColumn
// is always column+1). rslint-specific lock-in cases (tsgo AST edge shapes,
// branch lock-ins, real-user shapes, multi-line + comment fixtures) live in
// generator_star_spacing_extras_test.go.
package generator_star_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/generator_star_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// ---- option helpers (mirror upstream shorthand string vs object forms) ----

func optStr(s string) []any { return []any{s} }

func optObj(m map[string]any) []any { return []any{m} }

// optBA shorthand — `{ before, after }` only.
func optBA(b, a bool) []any {
	return []any{map[string]any{"before": b, "after": a}}
}

// ---- error helpers (Line is always 1; EndColumn is always Column+1) ----

func errMB(col int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "missingBefore", Line: 1, Column: col}
}
func errMA(col int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "missingAfter", Line: 1, Column: col}
}
func errUB(col int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "unexpectedBefore", Line: 1, Column: col}
}
func errUA(col int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "unexpectedAfter", Line: 1, Column: col}
}

func TestGeneratorStarSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&generator_star_spacing.GeneratorStarSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- Default ("before") ----
			{Code: `function foo(){}`},
			{Code: `function *foo(){}`},
			{Code: `function *foo(arg1, arg2){}`},
			{Code: `var foo = function *foo(){};`},
			{Code: `var foo = function *(){};`},
			{Code: `var foo = { *foo(){} };`},
			{Code: `var foo = {*foo(){} };`},
			{Code: `class Foo { *foo(){} }`},
			{Code: `class Foo {*foo(){} }`},
			{Code: `class Foo { static *foo(){} }`},
			{Code: `var foo = {*[ foo ](){} };`},
			{Code: `class Foo {*[ foo ](){} }`},

			// ---- "before" ----
			{Code: `function foo(){}`, Options: optStr("before")},
			{Code: `function *foo(){}`, Options: optStr("before")},
			{Code: `function *foo(arg1, arg2){}`, Options: optStr("before")},
			{Code: `var foo = function *foo(){};`, Options: optStr("before")},
			{Code: `var foo = function *(){};`, Options: optStr("before")},
			{Code: `var foo = { *foo(){} };`, Options: optStr("before")},
			{Code: `var foo = {*foo(){} };`, Options: optStr("before")},
			{Code: `class Foo { *foo(){} }`, Options: optStr("before")},
			{Code: `class Foo {*foo(){} }`, Options: optStr("before")},
			{Code: `class Foo { static *foo(){} }`, Options: optStr("before")},
			{Code: `class Foo {*[ foo ](){} }`, Options: optStr("before")},
			{Code: `var foo = {*[ foo ](){} };`, Options: optStr("before")},

			// ---- "after" ----
			{Code: `function foo(){}`, Options: optStr("after")},
			{Code: `function* foo(){}`, Options: optStr("after")},
			{Code: `function* foo(arg1, arg2){}`, Options: optStr("after")},
			{Code: `var foo = function* foo(){};`, Options: optStr("after")},
			{Code: `var foo = function* (){};`, Options: optStr("after")},
			{Code: `var foo = {* foo(){} };`, Options: optStr("after")},
			{Code: `var foo = { * foo(){} };`, Options: optStr("after")},
			{Code: `class Foo {* foo(){} }`, Options: optStr("after")},
			{Code: `class Foo { * foo(){} }`, Options: optStr("after")},
			{Code: `class Foo { static* foo(){} }`, Options: optStr("after")},
			{Code: `var foo = {* [foo](){} };`, Options: optStr("after")},
			{Code: `class Foo {* [foo](){} }`, Options: optStr("after")},

			// ---- "both" ----
			{Code: `function foo(){}`, Options: optStr("both")},
			{Code: `function * foo(){}`, Options: optStr("both")},
			{Code: `function * foo(arg1, arg2){}`, Options: optStr("both")},
			{Code: `var foo = function * foo(){};`, Options: optStr("both")},
			{Code: `var foo = function * (){};`, Options: optStr("both")},
			{Code: `var foo = { * foo(){} };`, Options: optStr("both")},
			{Code: `var foo = {* foo(){} };`, Options: optStr("both")},
			{Code: `class Foo { * foo(){} }`, Options: optStr("both")},
			{Code: `class Foo {* foo(){} }`, Options: optStr("both")},
			{Code: `class Foo { static * foo(){} }`, Options: optStr("both")},
			{Code: `var foo = {* [foo](){} };`, Options: optStr("both")},
			{Code: `class Foo {* [foo](){} }`, Options: optStr("both")},

			// ---- "neither" ----
			{Code: `function foo(){}`, Options: optStr("neither")},
			{Code: `function*foo(){}`, Options: optStr("neither")},
			{Code: `function*foo(arg1, arg2){}`, Options: optStr("neither")},
			{Code: `var foo = function*foo(){};`, Options: optStr("neither")},
			{Code: `var foo = function*(){};`, Options: optStr("neither")},
			{Code: `var foo = {*foo(){} };`, Options: optStr("neither")},
			{Code: `var foo = { *foo(){} };`, Options: optStr("neither")},
			{Code: `class Foo {*foo(){} }`, Options: optStr("neither")},
			{Code: `class Foo { *foo(){} }`, Options: optStr("neither")},
			{Code: `class Foo { static*foo(){} }`, Options: optStr("neither")},
			{Code: `var foo = {*[ foo ](){} };`, Options: optStr("neither")},
			{Code: `class Foo {*[ foo ](){} }`, Options: optStr("neither")},

			// ---- { "before": true, "after": false } ----
			{Code: `function foo(){}`, Options: optBA(true, false)},
			{Code: `function *foo(){}`, Options: optBA(true, false)},
			{Code: `function *foo(arg1, arg2){}`, Options: optBA(true, false)},
			{Code: `var foo = function *foo(){};`, Options: optBA(true, false)},
			{Code: `var foo = function *(){};`, Options: optBA(true, false)},
			{Code: `var foo = { *foo(){} };`, Options: optBA(true, false)},
			{Code: `var foo = {*foo(){} };`, Options: optBA(true, false)},
			{Code: `class Foo { *foo(){} }`, Options: optBA(true, false)},
			{Code: `class Foo {*foo(){} }`, Options: optBA(true, false)},
			{Code: `class Foo { static *foo(){} }`, Options: optBA(true, false)},

			// ---- { "before": false, "after": true } ----
			{Code: `function foo(){}`, Options: optBA(false, true)},
			{Code: `function* foo(){}`, Options: optBA(false, true)},
			{Code: `function* foo(arg1, arg2){}`, Options: optBA(false, true)},
			{Code: `var foo = function* foo(){};`, Options: optBA(false, true)},
			{Code: `var foo = function* (){};`, Options: optBA(false, true)},
			{Code: `var foo = {* foo(){} };`, Options: optBA(false, true)},
			{Code: `var foo = { * foo(){} };`, Options: optBA(false, true)},
			{Code: `class Foo {* foo(){} }`, Options: optBA(false, true)},
			{Code: `class Foo { * foo(){} }`, Options: optBA(false, true)},
			{Code: `class Foo { static* foo(){} }`, Options: optBA(false, true)},

			// ---- { "before": true, "after": true } ----
			{Code: `function foo(){}`, Options: optBA(true, true)},
			{Code: `function * foo(){}`, Options: optBA(true, true)},
			{Code: `function * foo(arg1, arg2){}`, Options: optBA(true, true)},
			{Code: `var foo = function * foo(){};`, Options: optBA(true, true)},
			{Code: `var foo = function * (){};`, Options: optBA(true, true)},
			{Code: `var foo = { * foo(){} };`, Options: optBA(true, true)},
			{Code: `var foo = {* foo(){} };`, Options: optBA(true, true)},
			{Code: `class Foo { * foo(){} }`, Options: optBA(true, true)},
			{Code: `class Foo {* foo(){} }`, Options: optBA(true, true)},
			{Code: `class Foo { static * foo(){} }`, Options: optBA(true, true)},

			// ---- { "before": false, "after": false } ----
			{Code: `function foo(){}`, Options: optBA(false, false)},
			{Code: `function*foo(){}`, Options: optBA(false, false)},
			{Code: `function*foo(arg1, arg2){}`, Options: optBA(false, false)},
			{Code: `var foo = function*foo(){};`, Options: optBA(false, false)},
			{Code: `var foo = function*(){};`, Options: optBA(false, false)},
			{Code: `var foo = {*foo(){} };`, Options: optBA(false, false)},
			{Code: `var foo = { *foo(){} };`, Options: optBA(false, false)},
			{Code: `class Foo {*foo(){} }`, Options: optBA(false, false)},
			{Code: `class Foo { *foo(){} }`, Options: optBA(false, false)},
			{Code: `class Foo { static*foo(){} }`, Options: optBA(false, false)},

			// ---- full configurability ----
			{Code: `function * foo(){}`, Options: optObj(map[string]any{"before": false, "after": false, "named": "both"})},
			{Code: `var foo = function * (){};`, Options: optObj(map[string]any{"before": false, "after": false, "anonymous": "both"})},
			{Code: `class Foo { * foo(){} }`, Options: optObj(map[string]any{"before": false, "after": false, "method": "both"})},
			{Code: `var foo = { * foo(){} }`, Options: optObj(map[string]any{"before": false, "after": false, "method": "both"})},
			{Code: `var foo = { bar: function * () {} }`, Options: optObj(map[string]any{"before": false, "after": false, "anonymous": "both"})},
			{Code: `class Foo { static * foo(){} }`, Options: optObj(map[string]any{"before": false, "after": false, "method": "both"})},
			{Code: `var foo = { * foo(){} }`, Options: optObj(map[string]any{"before": false, "after": false, "shorthand": "both"})},

			// ---- default to top level "before" ----
			{Code: `function *foo(){}`, Options: optObj(map[string]any{"method": "both"})},

			// ---- don't apply unrelated override ----
			{Code: `function*foo(){}`, Options: optObj(map[string]any{"before": false, "after": false, "method": "both"})},

			// ---- ensure using object-type override works ----
			{Code: `function * foo(){}`, Options: optObj(map[string]any{"before": false, "after": false, "named": map[string]any{"before": true, "after": true}})},

			// ---- unspecified option uses default ----
			{Code: `function *foo(){}`, Options: optObj(map[string]any{"before": false, "after": false, "named": map[string]any{"before": true}})},

			// ---- async / arrow (ecmaVersion 8) - generator-star-spacing must not flag non-generators ----
			{Code: `async function foo() { }`},
			{Code: `(async function() { })`},
			{Code: `async () => { }`},
			{Code: `({async foo() { }})`},
			{Code: `class A {async foo() { }}`},
			{Code: `(class {async foo() { }})`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Default ("before") ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function *foo(){}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingBefore", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:   `function* foo(arg1, arg2){}`,
				Output: []string{`function *foo(arg1, arg2){}`},
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errUA(9)},
			},
			{
				Code:   `var foo = function*foo(){};`,
				Output: []string{`var foo = function *foo(){};`},
				Errors: []rule_tester.InvalidTestCaseError{errMB(19)},
			},
			{
				Code:   `var foo = function* (){};`,
				Output: []string{`var foo = function *(){};`},
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errUA(19)},
			},
			{
				Code:   `var foo = {* foo(){} };`,
				Output: []string{`var foo = {*foo(){} };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `class Foo {* foo(){} }`,
				Output: []string{`class Foo {*foo(){} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `class Foo { static* foo(){} }`,
				Output: []string{`class Foo { static *foo(){} }`},
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errUA(19)},
			},

			// ---- "before" ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function *foo(){}`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9)},
			},
			{
				Code:   `function* foo(arg1, arg2){}`,
				Output: []string{`function *foo(arg1, arg2){}`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errUA(9)},
			},
			{
				Code:   `var foo = function*foo(){};`,
				Output: []string{`var foo = function *foo(){};`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19)},
			},
			{
				Code:   `var foo = function* (){};`,
				Output: []string{`var foo = function *(){};`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errUA(19)},
			},
			{
				Code:   `var foo = {* foo(){} };`,
				Output: []string{`var foo = {*foo(){} };`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(12)},
			},
			{
				Code:   `class Foo {* foo(){} }`,
				Output: []string{`class Foo {*foo(){} }`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(12)},
			},
			{
				Code:   `var foo = {* [ foo ](){} };`,
				Output: []string{`var foo = {*[ foo ](){} };`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(12)},
			},
			{
				Code:   `class Foo {* [ foo ](){} }`,
				Output: []string{`class Foo {*[ foo ](){} }`},
				Options: optStr("before"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(12)},
			},

			// ---- "after" ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function* foo(){}`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(9)},
			},
			{
				Code:   `function *foo(arg1, arg2){}`,
				Output: []string{`function* foo(arg1, arg2){}`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errMA(10)},
			},
			{
				Code:   `var foo = function *foo(){};`,
				Output: []string{`var foo = function* foo(){};`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errMA(20)},
			},
			{
				Code:   `var foo = function *(){};`,
				Output: []string{`var foo = function* (){};`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errMA(20)},
			},
			{
				Code:   `var foo = { *foo(){} };`,
				Output: []string{`var foo = { * foo(){} };`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `class Foo { *foo(){} }`,
				Output: []string{`class Foo { * foo(){} }`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `class Foo { static *foo(){} }`,
				Output: []string{`class Foo { static* foo(){} }`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errMA(20)},
			},
			{
				Code:   `var foo = { *[foo](){} };`,
				Output: []string{`var foo = { * [foo](){} };`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `class Foo { *[foo](){} }`,
				Output: []string{`class Foo { * [foo](){} }`},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},

			// ---- "both" ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function * foo(){}`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},
			{
				Code:   `function*foo(arg1, arg2){}`,
				Output: []string{`function * foo(arg1, arg2){}`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},
			{
				Code:   `var foo = function*foo(){};`,
				Output: []string{`var foo = function * foo(){};`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `var foo = function*(){};`,
				Output: []string{`var foo = function * (){};`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `var foo = {*foo(){} };`,
				Output: []string{`var foo = {* foo(){} };`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(12)},
			},
			{
				Code:   `class Foo {*foo(){} }`,
				Output: []string{`class Foo {* foo(){} }`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(12)},
			},
			{
				Code:   `class Foo { static*foo(){} }`,
				Output: []string{`class Foo { static * foo(){} }`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `var foo = {*[foo](){} };`,
				Output: []string{`var foo = {* [foo](){} };`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(12)},
			},
			{
				Code:   `class Foo {*[foo](){} }`,
				Output: []string{`class Foo {* [foo](){} }`},
				Options: optStr("both"),
				Errors: []rule_tester.InvalidTestCaseError{errMA(12)},
			},

			// ---- "neither" ----
			{
				Code:   `function * foo(){}`,
				Output: []string{`function*foo(){}`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},
			{
				Code:   `function * foo(arg1, arg2){}`,
				Output: []string{`function*foo(arg1, arg2){}`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},
			{
				Code:   `var foo = function * foo(){};`,
				Output: []string{`var foo = function*foo(){};`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errUA(20)},
			},
			{
				Code:   `var foo = function * (){};`,
				Output: []string{`var foo = function*(){};`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errUA(20)},
			},
			{
				Code:   `var foo = { * foo(){} };`,
				Output: []string{`var foo = { *foo(){} };`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(13)},
			},
			{
				Code:   `class Foo { * foo(){} }`,
				Output: []string{`class Foo { *foo(){} }`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(13)},
			},
			{
				Code:   `class Foo { static * foo(){} }`,
				Output: []string{`class Foo { static*foo(){} }`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errUA(20)},
			},
			{
				Code:   `var foo = { * [ foo ](){} };`,
				Output: []string{`var foo = { *[ foo ](){} };`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(13)},
			},
			{
				Code:   `class Foo { * [ foo ](){} }`,
				Output: []string{`class Foo { *[ foo ](){} }`},
				Options: optStr("neither"),
				Errors: []rule_tester.InvalidTestCaseError{errUA(13)},
			},

			// ---- { "before": true, "after": false } ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function *foo(){}`},
				Options: optBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9)},
			},
			{
				Code:   `function* foo(arg1, arg2){}`,
				Output: []string{`function *foo(arg1, arg2){}`},
				Options: optBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errUA(9)},
			},
			{
				Code:   `var foo = function*foo(){};`,
				Output: []string{`var foo = function *foo(){};`},
				Options: optBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19)},
			},
			{
				Code:   `var foo = function* (){};`,
				Output: []string{`var foo = function *(){};`},
				Options: optBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errUA(19)},
			},
			{
				Code:   `var foo = {* foo(){} };`,
				Output: []string{`var foo = {*foo(){} };`},
				Options: optBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{errUA(12)},
			},
			{
				Code:   `class Foo {* foo(){} }`,
				Output: []string{`class Foo {*foo(){} }`},
				Options: optBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{errUA(12)},
			},

			// ---- { "before": false, "after": true } ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function* foo(){}`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errMA(9)},
			},
			{
				Code:   `function *foo(arg1, arg2){}`,
				Output: []string{`function* foo(arg1, arg2){}`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errMA(10)},
			},
			{
				Code:   `var foo = function *foo(){};`,
				Output: []string{`var foo = function* foo(){};`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errMA(20)},
			},
			{
				Code:   `var foo = function *(){};`,
				Output: []string{`var foo = function* (){};`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errMA(20)},
			},
			{
				Code:   `var foo = { *foo(){} };`,
				Output: []string{`var foo = { * foo(){} };`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `class Foo { *foo(){} }`,
				Output: []string{`class Foo { * foo(){} }`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `class Foo { static *foo(){} }`,
				Output: []string{`class Foo { static* foo(){} }`},
				Options: optBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errMA(20)},
			},

			// ---- { "before": true, "after": true } ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function * foo(){}`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},
			{
				Code:   `function*foo(arg1, arg2){}`,
				Output: []string{`function * foo(arg1, arg2){}`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},
			{
				Code:   `var foo = function*foo(){};`,
				Output: []string{`var foo = function * foo(){};`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `var foo = function*(){};`,
				Output: []string{`var foo = function * (){};`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `var foo = {*foo(){} };`,
				Output: []string{`var foo = {* foo(){} };`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMA(12)},
			},
			{
				Code:   `class Foo {*foo(){} }`,
				Output: []string{`class Foo {* foo(){} }`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMA(12)},
			},
			{
				Code:   `class Foo { static*foo(){} }`,
				Output: []string{`class Foo { static * foo(){} }`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},

			// ---- { "before": false, "after": false } ----
			{
				Code:   `function * foo(){}`,
				Output: []string{`function*foo(){}`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},
			{
				Code:   `function * foo(arg1, arg2){}`,
				Output: []string{`function*foo(arg1, arg2){}`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},
			{
				Code:   `var foo = function * foo(){};`,
				Output: []string{`var foo = function*foo(){};`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errUA(20)},
			},
			{
				Code:   `var foo = function * (){};`,
				Output: []string{`var foo = function*(){};`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errUA(20)},
			},
			{
				Code:   `var foo = { * foo(){} };`,
				Output: []string{`var foo = { *foo(){} };`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUA(13)},
			},
			{
				Code:   `class Foo { * foo(){} }`,
				Output: []string{`class Foo { *foo(){} }`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUA(13)},
			},
			{
				Code:   `class Foo { static * foo(){} }`,
				Output: []string{`class Foo { static*foo(){} }`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(20), errUA(20)},
			},

			// ---- full configurability ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function * foo(){}`},
				Options: optObj(map[string]any{"before": false, "after": false, "named": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},
			{
				Code:   `var foo = function*(){};`,
				Output: []string{`var foo = function * (){};`},
				Options: optObj(map[string]any{"before": false, "after": false, "anonymous": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `class Foo { *foo(){} }`,
				Output: []string{`class Foo { * foo(){} }`},
				Options: optObj(map[string]any{"before": false, "after": false, "method": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `var foo = { *foo(){} }`,
				Output: []string{`var foo = { * foo(){} }`},
				Options: optObj(map[string]any{"before": false, "after": false, "method": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},
			{
				Code:   `var foo = { bar: function*() {} }`,
				Output: []string{`var foo = { bar: function * () {} }`},
				Options: optObj(map[string]any{"before": false, "after": false, "anonymous": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(26), errMA(26)},
			},
			{
				Code:   `class Foo { static*foo(){} }`,
				Output: []string{`class Foo { static * foo(){} }`},
				Options: optObj(map[string]any{"before": false, "after": false, "method": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},
			{
				Code:   `var foo = { *foo(){} }`,
				Output: []string{`var foo = { * foo(){} }`},
				Options: optObj(map[string]any{"before": false, "after": false, "shorthand": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMA(13)},
			},

			// ---- default to top level "before" ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function *foo(){}`},
				Options: optObj(map[string]any{"method": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9)},
			},

			// ---- don't apply unrelated override ----
			{
				Code:   `function * foo(){}`,
				Output: []string{`function*foo(){}`},
				Options: optObj(map[string]any{"before": false, "after": false, "method": "both"}),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},

			// ---- ensure using object-type override works ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function * foo(){}`},
				Options: optObj(map[string]any{"before": false, "after": false, "named": map[string]any{"before": true, "after": true}}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},

			// ---- unspecified option uses default ----
			{
				Code:   `function*foo(){}`,
				Output: []string{`function *foo(){}`},
				Options: optObj(map[string]any{"before": false, "after": false, "named": map[string]any{"before": true}}),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9)},
			},

			// ---- async generators ----
			{
				Code:   `({ async * foo(){} })`,
				Output: []string{`({ async*foo(){} })`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},
			{
				Code:   `({ async*foo(){} })`,
				Output: []string{`({ async * foo(){} })`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(9), errMA(9)},
			},
			{
				Code:   `class Foo { async * foo(){} }`,
				Output: []string{`class Foo { async*foo(){} }`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(19), errUA(19)},
			},
			{
				Code:   `class Foo { async*foo(){} }`,
				Output: []string{`class Foo { async * foo(){} }`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(18), errMA(18)},
			},
			{
				Code:   `class Foo { static async * foo(){} }`,
				Output: []string{`class Foo { static async*foo(){} }`},
				Options: optBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{errUB(26), errUA(26)},
			},
			{
				Code:   `class Foo { static async*foo(){} }`,
				Output: []string{`class Foo { static async * foo(){} }`},
				Options: optBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{errMB(25), errMA(25)},
			},
		},
	)
}
