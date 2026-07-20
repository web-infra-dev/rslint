// TestPreferDestructuringExtrasMatrix exercises cross-product cases that are
// impractical to read in the upstream mirror: numeric key boundaries, nested
// containers, TypeScript/JSX wrappers, and exact autofix preservation.
// The smaller branch-oriented cases live in prefer_destructuring_extras_test.go.
package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func oneLineMatrixError(
	code string,
	startColumn int,
	kind string,
	output string,
	options any,
	tsx bool,
) rule_tester.InvalidTestCase {
	// Matrix inputs are single-line ASCII snippets ending in a semicolon.
	// ESLint's exclusive end column is therefore exactly len(code).
	var outputs []string
	if output != "" {
		outputs = []string{output}
	}
	return rule_tester.InvalidTestCase{
		Code:    code,
		Output:  outputs,
		Options: options,
		Tsx:     tsx,
		Errors: []rule_tester.InvalidTestCaseError{
			preferError(kind, 1, startColumn, 1, len(code)),
		},
	}
}

func TestPreferDestructuringExtrasMatrix(t *testing.T) {
	enforceAll := []any{
		map[string]any{"array": true, "object": true},
		map[string]any{"enforceForRenamedProperties": true},
	}

	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: wrappers around the complete RHS remain visible to
		// @typescript-eslint/parser and therefore are not member expressions. ----
		{Code: "const foo = object.foo as unknown;"},
		{Code: "const foo = object.foo!;"},
		{Code: "const foo = object.foo satisfies unknown;"},

		// ---- Dimension 4: every link in an unbroken optional chain is ignored. ----
		{Code: "const foo = object?.foo!;"},
		{Code: "const foo = object?.[\"foo\"]!;"},
		{Code: "const foo = getObject?.()!.foo;"},

		// ---- Dimension 4: direct super/private/resource-management exclusions. ----
		{Code: "class C extends B { method() { const foo = super.foo; } }"},
		{Code: "class C { #foo; method() { const foo = this.#foo; } }"},
		{Code: "using foo = object.foo;"},
		{Code: "async function f() { await using value = array[0]; }"},

		// ---- Dimension 1: class fields are not VariableDeclarators, while an
		// assignment nested in a field initializer is covered below. ----
		{Code: "class C { foo = object.foo; }"},

		// ---- Dimension 4: non-identifier targets do not enter the default
		// same-name object branch. ----
		{Code: "object.foo = other.foo;"},
		{Code: "({ foo } = object.foo);"},
		{Code: "const { foo } = object.foo;"},
	}

	// ---- Dimension 4: non-integer and non-Number keys take the object path.
	// Without rename enforcement, none matches the identifier `value`. ----
	for _, key := range []string{
		"1.5",
		"5e-324",
		"1e309",
		"-1",
		"+1",
		"-0",
		"1n",
		"\"0\"",
		"true",
		"null",
		"/x/",
		"0 as number",
		"`0`",
		"dynamic",
		"Symbol.iterator",
	} {
		valid = append(valid, rule_tester.ValidTestCase{
			Code: "const value = array[" + key + "];",
		})
	}

	invalid := make([]rule_tester.InvalidTestCase, 0, 40)

	// ---- Dimension 4: every numeric-literal spelling whose JavaScript Number
	// value is an integer must enter the array branch. This includes overflow
	// rounding within Number and underflow to zero. ----
	for _, key := range []string{
		"0",
		"(0)",
		"0.0",
		"1e2",
		"0x10",
		"0o10",
		"0b10",
		"1_000",
		"9007199254740993",
		"1e20",
		"1e-400",
	} {
		code := "const value = array[" + key + "];"
		invalid = append(invalid, oneLineMatrixError(code, 7, "array", "", nil, false))
	}

	// ---- Dimension 4: the same non-integer/dynamic keys become object reports
	// when renamed properties are enforced; none may leak into the array arm. ----
	for _, key := range []string{
		"1.5",
		"5e-324",
		"1e309",
		"-1",
		"+1",
		"-0",
		"1n",
		"\"0\"",
		"true",
		"null",
		"/x/",
		"0 as number",
		"`0`",
		"dynamic",
		"Symbol.iterator",
	} {
		code := "const value = array[" + key + "];"
		invalid = append(invalid, oneLineMatrixError(code, 7, "object", "", enforceAll, false))
	}

	// ---- Dimension 4: a trailing non-null assertion belongs to ESLint's
	// ChainExpression only when it directly wraps the optional-chain link. ----
	invalid = append(invalid,
		oneLineMatrixError(
			"const foo = (object?.foo!).foo;",
			7,
			"object",
			"const {foo} = object?.foo!;",
			nil,
			false,
		),
		oneLineMatrixError(
			"const foo = (object?.[\"foo\"]!).foo;",
			7,
			"object",
			"const {foo} = object?.[\"foo\"]!;",
			nil,
			false,
		),
		oneLineMatrixError(
			"const foo = (getObject?.()!).foo;",
			7,
			"object",
			"const {foo} = getObject?.()!;",
			nil,
			false,
		),
		// Parentheses before `!` terminate the chain. The remaining
		// TSNonNullExpression is unknown to ESLint and stays parenthesized.
		oneLineMatrixError(
			"const foo = ((object?.foo)!).foo;",
			7,
			"object",
			"const {foo} = ((object?.foo)!);",
			nil,
			false,
		),
		// A TS `as` wrapper is also unknown to ESLint's precedence helper.
		oneLineMatrixError(
			"const foo = (object?.foo as any).foo;",
			7,
			"object",
			"const {foo} = (object?.foo as any);",
			nil,
			false,
		),
	)

	// ---- Dimension 4: JSX expression kinds are ordinary primary expressions
	// in ESLint, so stripping their source parentheses must not add new ones. ----
	invalid = append(invalid,
		oneLineMatrixError(
			"const foo = (<div />).foo;",
			7,
			"object",
			"const {foo} = <div />;",
			nil,
			true,
		),
		oneLineMatrixError(
			"const foo = (<div>text</div>).foo;",
			7,
			"object",
			"const {foo} = <div>text</div>;",
			nil,
			true,
		),
		oneLineMatrixError(
			"const foo = (<></>).foo;",
			7,
			"object",
			"const {foo} = <></>;",
			nil,
			true,
		),
	)

	// ---- Dimension 3: comment-like source text is not a parser comment and
	// must not suppress an otherwise safe fix. ----
	invalid = append(invalid,
		oneLineMatrixError(
			"const foo = \"/* not a comment */\".foo;",
			7,
			"object",
			"const {foo} = \"/* not a comment */\";",
			nil,
			false,
		),
		oneLineMatrixError(
			"const foo = (`// not a comment`).foo;",
			7,
			"object",
			"const {foo} = `// not a comment`;",
			nil,
			false,
		),
	)

	// ---- Dimension 2: nested assignment expressions are visited independently.
	// The VariableDeclarator reports first and fixes the outer access; the inner
	// assignment also reports but remains intentionally non-fixable. ----
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:   "const foo = (foo = object.foo).foo;",
		Output: []string{"const {foo} = foo = object.foo;"},
		Errors: []rule_tester.InvalidTestCaseError{
			preferError("object", 1, 7, 1, 35),
			preferError("object", 1, 14, 1, 30),
		},
	})

	// ---- Dimension 2: traversal remains independent through function, class,
	// method, and for-loop containers, mixing fixable object declarations with
	// a non-fixable array assignment. ----
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "function outer() {\n" +
			"  class C {\n" +
			"    method() {\n" +
			"      for (let foo = object.foo; foo; ) {\n" +
			"        value = array[0];\n" +
			"        const bar = source.bar;\n" +
			"      }\n" +
			"    }\n" +
			"  }\n" +
			"}",
		Output: []string{
			"function outer() {\n" +
				"  class C {\n" +
				"    method() {\n" +
				"      for (let {foo} = object; foo; ) {\n" +
				"        value = array[0];\n" +
				"        const {bar} = source;\n" +
				"      }\n" +
				"    }\n" +
				"  }\n" +
				"}",
		},
		Errors: []rule_tester.InvalidTestCaseError{
			preferError("object", 4, 16, 4, 32),
			preferError("array", 5, 9, 5, 25),
			preferError("object", 6, 15, 6, 31),
		},
	})

	// ---- Dimension 4: tsgo's assignment predicate must keep recognizing
	// legal TypeScript targets through the wrappers accepted by the parser.
	// Rename enforcement reports all of them, but assignments remain unfixable. ----
	for _, code := range []string{
		"((foo)) = object.foo;",
		"(foo as any) = object.foo;",
		"(<any>foo) = object.foo;",
		"foo! = object.foo;",
	} {
		invalid = append(invalid, oneLineMatrixError(
			code,
			1,
			"object",
			"",
			enforceAll,
			false,
		))
	}

	// ---- Dimension 4: array checks do not depend on an identifier LHS. ----
	invalid = append(invalid, oneLineMatrixError(
		"[value] = array[0];",
		1,
		"array",
		"",
		nil,
		false,
	))

	// ---- Dimension 1: assignments nested in otherwise-ignored class fields
	// are still visited through the BinaryExpression listener. ----
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: "class C { field = (foo = object.foo); }",
		Errors: []rule_tester.InvalidTestCaseError{
			preferError("object", 1, 20, 1, 36),
		},
	})

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		valid,
		invalid,
	)
}
