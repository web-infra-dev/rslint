// TestPreferDestructuringUpstream migrates the full valid/invalid suite from
// ESLint v10.7.0 tests/lib/rules/prefer-destructuring.js 1:1. Position
// assertions cover line/column and endLine/endColumn for every invalid case.
// rslint-specific lock-in cases live in prefer_destructuring_extras_test.go.
package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func preferError(kind string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preferDestructuring",
		Message:   "Use " + kind + " destructuring.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestPreferDestructuringUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		[]rule_tester.ValidTestCase{
			// ---- Basic destructuring and uninitialized declarations ----
			{Code: "var [foo] = array;"},
			{Code: "var { foo } = object;"},
			{Code: "var foo;"},
			// ---- Renamed-property defaults and options ----
			{Code: "var foo = object.bar;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}}},
			{Code: "var foo = object.bar;", Options: []any{map[string]any{"object": true}}},
			{Code: "var foo = object.bar;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "var foo = object.bar;", Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "var foo = object['bar'];", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "var foo = object[bar];", Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "var { bar: foo } = object;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "var { bar: foo } = object;", Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "var { [bar]: foo } = object;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "var { [bar]: foo } = object;", Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			// ---- Per-kind enablement ----
			{Code: "var foo = array[0];", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": false}}}},
			{Code: "var foo = array[0];", Options: []any{map[string]any{"array": false}}},
			{Code: "var foo = object.foo;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": false}}}},
			{Code: "var foo = object['foo'];", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": false}}}},
			{Code: "({ foo } = object);"},
			// ---- Regression #8654: disabled array checks stay disabled ----
			{Code: "var foo = array[0];", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": false}}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "var foo = array[0];", Options: []any{map[string]any{"array": false}, map[string]any{"enforceForRenamedProperties": true}}},
			// ---- Assignment expressions and compound assignments ----
			{Code: "[foo] = array;"},
			{Code: "foo += array[0]"},
			{Code: "foo &&= array[0]"},
			{Code: "foo += bar.foo"},
			{Code: "foo ||= bar.foo"},
			{Code: "foo ??= bar['foo']"},
			// ---- Per-node-type assignment/declaration options ----
			{Code: "foo = object.foo;", Options: []any{map[string]any{"AssignmentExpression": map[string]any{"object": false}}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "foo = object.foo;", Options: []any{map[string]any{"AssignmentExpression": map[string]any{"object": false}}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "foo = array[0];", Options: []any{map[string]any{"AssignmentExpression": map[string]any{"array": false}}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "foo = array[0];", Options: []any{map[string]any{"AssignmentExpression": map[string]any{"array": false}}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "foo = array[0];", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": true}, "AssignmentExpression": map[string]any{"array": false}}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "var foo = array[0];", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": false}, "AssignmentExpression": map[string]any{"array": true}}, map[string]any{"enforceForRenamedProperties": false}}},
			{Code: "foo = object.foo;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}, "AssignmentExpression": map[string]any{"object": false}}}},
			{Code: "var foo = object.foo;", Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": false}, "AssignmentExpression": map[string]any{"object": true}}}},
			// ---- super, dynamic access, and already-destructured targets ----
			{Code: "class Foo extends Bar { static foo() {var foo = super.foo} }"},
			{Code: "foo = bar[foo];"},
			{Code: "var foo = bar[foo];"},
			{Code: "var {foo: {bar}} = object;", Options: []any{map[string]any{"object": true}}},
			{Code: "var {bar} = object.foo;", Options: []any{map[string]any{"object": true}}},
			// ---- Optional chaining ----
			{Code: "var foo = array?.[0];"},
			{Code: "var foo = object?.foo;"},
			// ---- Private identifiers ----
			{Code: "class C { #x; foo() { const x = this.#x; } }"},
			{Code: "class C { #x; foo() { x = this.#x; } }"},
			{Code: "class C { #x; foo(a) { x = a.#x; } }"},
			{Code: "class C { #x; foo() { const x = this.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "class C { #x; foo() { const y = this.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "class C { #x; foo() { x = this.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "class C { #x; foo() { y = this.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "class C { #x; foo(a) { x = a.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "class C { #x; foo(a) { y = a.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			{Code: "class C { #x; foo() { x = this.a.#x; } }", Options: []any{map[string]any{"array": true, "object": true}, map[string]any{"enforceForRenamedProperties": true}}},
			// ---- Explicit resource management ----
			{Code: "using foo = array[0];"},
			{Code: "using foo = object.foo;"},
			{Code: "await using foo = array[0];"},
			{Code: "await using foo = object.foo;"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Array and basic object access ----
			{
				Code:   "var foo = array[0];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("array", 1, 5, 1, 19)},
			},
			{
				Code:   "foo = array[0];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("array", 1, 1, 1, 15)},
			},
			{
				Code:   "var foo = object.foo;",
				Output: []string{"var {foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
			// ---- Autofix receiver precedence ----
			{
				Code:   "var foo = (a, b).foo;",
				Output: []string{"var {foo} = (a, b);"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
			{
				Code:   "var length = (() => {}).length;",
				Output: []string{"var {length} = () => {};"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 31)},
			},
			{
				Code:   "var foo = (a = b).foo;",
				Output: []string{"var {foo} = a = b;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 22)},
			},
			{
				Code:   "var foo = (a || b).foo;",
				Output: []string{"var {foo} = a || b;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 23)},
			},
			{
				Code:   "var foo = (f()).foo;",
				Output: []string{"var {foo} = f();"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 20)},
			},
			{
				Code:   "var foo = object.bar.foo;",
				Output: []string{"var {foo} = object.bar;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 25)},
			},
			// ---- Renamed and computed properties ----
			{
				Code:    "var foobar = object.bar;",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 24)},
			},
			{
				Code:    "var foobar = object.bar;",
				Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 24)},
			},
			{
				Code:    "var foo = object[bar];",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"object": true}}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 22)},
			},
			{
				Code:    "var foo = object[bar];",
				Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 22)},
			},
			{
				Code:    "var foo = object[foo];",
				Options: []any{map[string]any{"object": true}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 22)},
			},
			// ---- Same-name string access and assignments ----
			{
				Code:   "var foo = object['foo'];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 24)},
			},
			{
				Code:   "foo = object.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 1, 1, 17)},
			},
			{
				Code:   "foo = object['foo'];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 1, 1, 20)},
			},
			// ---- Per-kind and per-node-type options ----
			{
				Code:    "var foo = array[0];",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": true}}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("array", 1, 5, 1, 19)},
			},
			{
				Code:    "foo = array[0];",
				Options: []any{map[string]any{"AssignmentExpression": map[string]any{"array": true}}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("array", 1, 1, 1, 15)},
			},
			{
				Code:    "var foo = array[0];",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": true}, "AssignmentExpression": map[string]any{"array": false}}, map[string]any{"enforceForRenamedProperties": true}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("array", 1, 5, 1, 19)},
			},
			{
				Code:    "var foo = array[0];",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": true}, "AssignmentExpression": map[string]any{"array": false}}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("array", 1, 5, 1, 19)},
			},
			{
				Code:    "foo = array[0];",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": false}, "AssignmentExpression": map[string]any{"array": true}}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("array", 1, 1, 1, 15)},
			},
			{
				Code:    "foo = object.foo;",
				Options: []any{map[string]any{"VariableDeclarator": map[string]any{"array": true, "object": false}, "AssignmentExpression": map[string]any{"object": true}}},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 1, 1, 17)},
			},
			// ---- Nested super access ----
			{
				Code:   "class Foo extends Bar { static foo() {var bar = super.foo.bar} }",
				Output: []string{"class Foo extends Bar { static foo() {var {bar} = super.foo} }"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 43, 1, 62)},
			},
			// ---- Comments ----
			{
				Code:   "var /* comment */ foo = object.foo;",
				Output: []string{"var /* comment */ {foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 19, 1, 35)},
			},
			{
				Code:   "var a, /* comment */foo = object.foo;",
				Output: []string{"var a, /* comment */{foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 21, 1, 37)},
			},
			{
				Code:   "var foo /* comment */ = object.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 35)},
			},
			{
				Code:   "var a, foo /* comment */ = object.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 8, 1, 38)},
			},
			{
				Code:   "var foo /* comment */ = object.foo, a;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 35)},
			},
			{
				Code:   "var foo // comment\n = object.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 2, 14)},
			},
			{
				Code:   "var foo = /* comment */ object.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 35)},
			},
			{
				Code:   "var foo = // comment\n object.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 2, 12)},
			},
			{
				Code:   "var foo = (/* comment */ object).foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 37)},
			},
			{
				Code:   "var foo = (object /* comment */).foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 37)},
			},
			{
				Code:   "var foo = bar(/* comment */).foo;",
				Output: []string{"var {foo} = bar(/* comment */);"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 33)},
			},
			{
				Code:   "var foo = bar/* comment */.baz.foo;",
				Output: []string{"var {foo} = bar/* comment */.baz;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 35)},
			},
			{
				Code:   "var foo = bar[// comment\nbaz].foo;",
				Output: []string{"var {foo} = bar[// comment\nbaz];"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 2, 9)},
			},
			{
				Code:   "var foo // comment\n = bar(/* comment */).foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 2, 26)},
			},
			{
				Code:   "var foo = bar/* comment */.baz/* comment */.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 48)},
			},
			{
				Code:   "var foo = object// comment\n.foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 2, 5)},
			},
			{
				Code:   "var foo = object./* comment */foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 34)},
			},
			{
				Code:   "var foo = (/* comment */ object.foo);",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 37)},
			},
			{
				Code:   "var foo = (object.foo /* comment */);",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 37)},
			},
			{
				Code:   "var foo = object.foo/* comment */;",
				Output: []string{"var {foo} = object/* comment */;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
			{
				Code:   "var foo = object.foo// comment",
				Output: []string{"var {foo} = object// comment"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
			{
				Code:   "var foo = object.foo/* comment */, a;",
				Output: []string{"var {foo} = object/* comment */, a;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
			{
				Code:   "var foo = object.foo// comment\n, a;",
				Output: []string{"var {foo} = object// comment\n, a;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
			{
				Code:   "var foo = object.foo, /* comment */ a;",
				Output: []string{"var {foo} = object, /* comment */ a;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 5, 1, 21)},
			},
		},
	)
}
