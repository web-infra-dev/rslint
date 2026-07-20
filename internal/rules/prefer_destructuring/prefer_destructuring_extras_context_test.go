// TestPreferDestructuringExtrasContext locks in traversal behavior through
// nested containers and distinguishes simple assignments from similar syntax.
package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuringExtrasContext(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// These initializer-like forms are neither VariableDeclarators nor
		// AssignmentExpressions in ESTree.
		{Code: "function f(foo = object.foo) { return foo; }"},
		{Code: "const f = (foo = object.foo) => foo;"},
		{Code: "class C { foo = object.foo; }"},
		{Code: "enum E { Value = object.foo }"},
		{Code: "const { foo = object.foo } = source;"},
		{Code: "const objectValue = { foo: object.foo };"},
		// For-in/of right-hand expressions are not assignments, and the loop
		// declaration has no ordinary declarator initializer.
		{Code: "for (const foo of object.foo) consume(foo);"},
		{Code: "for (foo of object.foo) consume(foo);"},
		{Code: "for (const foo in object.foo) consume(foo);"},
		// A TypeScript expression around the whole RHS prevents it from being
		// the MemberExpression inspected by the core rule.
		{Code: "const foo = object.foo as unknown;"},
		{Code: "const foo = object.foo satisfies unknown;"},
		{Code: "const foo = object.foo!;"},
	}

	// Only the plain `=` operator creates the AssignmentExpression that
	// upstream checks. Cover every compound assignment token accepted by
	// modern JavaScript so tsgo predicate changes cannot widen the listener.
	for _, operator := range []string{
		"+=",
		"-=",
		"*=",
		"/=",
		"%=",
		"**=",
		"<<=",
		">>=",
		">>>=",
		"&=",
		"^=",
		"|=",
		"&&=",
		"||=",
		"??=",
	} {
		valid = append(valid,
			rule_tester.ValidTestCase{Code: "foo " + operator + " object.foo;"},
			rule_tester.ValidTestCase{Code: "value " + operator + " array[0];"},
		)
	}

	enforceObject := []any{
		map[string]any{"object": true},
		map[string]any{"enforceForRenamedProperties": true},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			Code:   "export const foo = object.foo;",
			Output: []string{"export const {foo} = object;"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 14, 1, 30),
			},
		},
		{
			Code:   "for (let foo = object.foo; foo; ) consume(foo);",
			Output: []string{"for (let {foo} = object; foo; ) consume(foo);"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 10, 1, 26),
			},
		},
		{
			Code: "for (foo = object.foo; foo; ) consume(foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 6, 1, 22),
			},
		},
		{
			Code: "export default (foo = object.foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 17, 1, 33),
			},
		},
		{
			Code: "const callback = () => (foo = object.foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 25, 1, 41),
			},
		},
		// In chained and nested assignments, only nodes whose own RHS is an
		// access expression report. Traversal visits the outer node first.
		{
			Code: "foo = bar = object.bar;",
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 7, 1, 23),
			},
		},
		{
			Code: "foo = (bar = object.bar).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 1, 1, 29),
				preferError("object", 1, 8, 1, 24),
			},
		},
		// Integer-index reports do not depend on the shape of the legal
		// assignment target.
		oneLineMatrixError(
			"target.foo = array[0];",
			1,
			"array",
			"",
			nil,
			false,
		),
		oneLineMatrixError(
			"target[\"foo\"] = array[0];",
			1,
			"array",
			"",
			nil,
			false,
		),
		oneLineMatrixError(
			"this.foo = array[0];",
			1,
			"array",
			"",
			nil,
			false,
		),
		oneLineMatrixError(
			"[...values] = array[0];",
			1,
			"array",
			"",
			nil,
			false,
		),
		{
			Code: "({ value } = array[0]);",
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("array", 1, 2, 1, 22),
			},
		},
		{
			Code:    "({ remote: local } = object.remote);",
			Options: enforceObject,
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 2, 1, 35),
			},
		},
		{
			Code:   "namespace N { export const foo = object.foo; }",
			Output: []string{"namespace N { export const {foo} = object; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 28, 1, 44),
			},
		},
		{
			Code:   "try { const foo = object.foo; } catch { value = array[0]; }",
			Output: []string{"try { const {foo} = object; } catch { value = array[0]; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 13, 1, 29),
				preferError("array", 1, 41, 1, 57),
			},
		},
		{
			Code:   "switch (kind) { case 0: { const foo = object.foo; break; } }",
			Output: []string{"switch (kind) { case 0: { const {foo} = object; break; } }"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 33, 1, 49),
			},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		valid,
		invalid,
	)
}
