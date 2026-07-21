// The tests in this file lock in focused branches, traversal contexts, and
// real-user scenarios that the upstream suite does not exercise. Larger
// configuration and cross-product matrices live in the sibling extras files.
package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuringExtras(t *testing.T) {
	enforceRenamedObject := []any{
		map[string]any{"object": true},
		map[string]any{"enforceForRenamedProperties": true},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		[]rule_tester.ValidTestCase{
			// Direct Rule.Run/API fallback: public config validation rejects an
			// empty first object because it matches both schema oneOf branches.
			// If validation is bypassed, it safely disables both check kinds.
			{Code: "const foo = object.foo;", Options: map[string]any{}},
			// Locks in upstream shouldCheck() arm 2: a missing per-node-type
			// assignment config disables the check.
			{Code: "foo = object.foo;", Options: map[string]any{"AssignmentExpression": map[string]any{}}},
			// ---- Dimension 4: a type wrapper around the whole inspected RHS is
			// not transparent; ESTree exposes TSAsExpression, not MemberExpression. ----
			{Code: "const foo = object.foo as string;"},
			// ---- Dimension 4: a non-null wrapper around the whole inspected RHS
			// is likewise not a member expression. ----
			{Code: "const foo = object.foo!;"},
			// ---- Dimension 4: optional property chain (real-user #12636) ----
			{Code: "const id = correlations?.id;"},
			// ---- Dimension 4: optional-call chain; all later links remain part
			// of the chain and must be ignored. ----
			{Code: "const foo = getObject?.().foo;"},
			// ---- Dimension 4: no-substitution template key is not ESTree
			// Literal and therefore does not satisfy the same-name arm. ----
			{Code: "const foo = object[`foo`];"},
			// ---- Dimension 4: negative numeric key is a UnaryExpression, not an
			// integer Literal, and does not match a target name by default. ----
			{Code: "const value = array[-1];"},
			// ---- Dimension 4: BigInt key is not an integer Number Literal. ----
			{Code: "const value = array[1n];"},
			// ---- Dimension 4: dynamic Symbol.iterator element key does not
			// satisfy the same-name arm. ----
			{Code: "const iterator = object[Symbol.iterator];"},
			// ---- Dimension 4: string-literal key with a renamed target stays
			// valid when renamed properties are not enforced. ----
			{Code: "const foo = object['bar'];"},
			// ---- Dimension 4: private identifier and '#foo' string are separate
			// key classes; neither is treated as the public name foo. ----
			{Code: "class C { #foo; method() { const foo = this.#foo; const other = this['#foo']; } }"},
			// ---- Real-user #9625: direct super receiver cannot be destructured. ----
			{Code: "class Node { static get inheritance() { return ['node']; } } class Element extends Node { static get inheritance() { const inheritance = super.inheritance; return inheritance; } }"},
			// Locks in upstream performCheck() arm 2: direct-super element access,
			// a branch the upstream suite covers only with dotted access.
			{Code: "class C extends B { method() { const value = super[0]; } }"},
			// ---- Dimension 4: graceful degradation — empty object binding
			// pattern has no identifier name and must not crash or match. ----
			{Code: "const {} = object.foo;"},
			// ---- Dimension 4: graceful degradation — rest binding pattern has
			// no single identifier name and stays valid without rename enforcement. ----
			{Code: "const {...rest} = object.foo;"},
			// ---- Real-user #11836: member assignment target does not satisfy the
			// default same-name check. ----
			{Code: "let obj1, obj2; obj1.property1 = obj2.property2;"},
			// ---- N/A: declaration/container variants — this rule targets
			// variable declarators and assignments, not function/class declarations. ----
			// ---- N/A: overload/abstract/declare body absence — no function or
			// class body is inspected by this rule. ----
			// ---- N/A: ancestor-container walks — the rule keeps no scope or
			// container state; nested listener independence is tested below. ----
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single parenthesized receiver ----
			{
				Code:   "const foo = (object).foo;",
				Output: []string{"const {foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 25)},
			},
			// ---- Dimension 4: multi-level parenthesized receiver ----
			{
				Code:   "const foo = ((object)).foo;",
				Output: []string{"const {foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 27)},
			},
			// ---- Dimension 4: TypeScript `as` receiver; unknown ESTree
			// precedence requires preserving one parenthesis layer in the fix. ----
			{
				Code:   "const foo = (object as any).foo;",
				Output: []string{"const {foo} = (object as any);"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 32)},
			},
			// ---- Dimension 4: TypeScript `satisfies` receiver ----
			{
				Code:   "const foo = (object satisfies Record<string, unknown>).foo;",
				Output: []string{"const {foo} = (object satisfies Record<string, unknown>);"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 59)},
			},
			// ---- Dimension 4: TypeScript non-null receiver ----
			{
				Code:   "const foo = object!.foo;",
				Output: []string{"const {foo} = (object!);"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 24)},
			},
			// Locks in upstream shouldFix() arm 1 on TypeScript annotations: core
			// reports and its fix removes the annotation.
			{
				Code:   "const foo: string = object.foo;",
				Output: []string{"const {foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 31)},
			},
			// ---- Dimension 4: parenthesized string-literal element key remains
			// a same-name Literal but computed access is never autofixed. ----
			{
				Code:   "const foo = object[(\"foo\")];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 28)},
			},
			// Locks in upstream performCheck() same-name arm 1: escaped source
			// spelling compares by cooked Literal value, matching foo.
			{
				Code:   "const foo = object['f\\u006fo'];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 31)},
			},
			// ---- Dimension 4: parenthesized integer element key ----
			{
				Code:   "const value = array[(0)];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("array", 1, 7, 1, 25)},
			},
			// Locks in upstream isArrayIndexAccess() arm 1: a non-decimal numeric
			// source form still has an integer Number value.
			{
				Code:   "const value = array[0x1];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("array", 1, 7, 1, 25)},
			},
			// ---- Dimension 4: template element key enters the object arm only
			// when renamed-property enforcement bypasses name matching. ----
			{
				Code:    "const foo = object[`bar`];",
				Options: enforceRenamedObject,
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 26)},
			},
			// ---- Dimension 4: Symbol.iterator dynamic element access ----
			{
				Code:    "const iterator = object[Symbol.iterator];",
				Options: enforceRenamedObject,
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 41)},
			},
			// Locks in upstream isArrayIndexAccess() arm 2: UnaryExpression is
			// not an integer Number value.
			{
				Code:    "const value = array[-1];",
				Options: enforceRenamedObject,
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 24)},
			},
			// Locks in upstream isArrayIndexAccess() arm 3: a non-integer Number
			// value enters the object path.
			{
				Code:    "const value = array[1.5];",
				Options: enforceRenamedObject,
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 25)},
			},
			// ---- Dimension 4: parentheses break an optional property chain, so
			// the outer ordinary access is checked and can be fixed. ----
			{
				Code:   "const foo = (object?.foo).foo;",
				Output: []string{"const {foo} = object?.foo;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 30)},
			},
			// ---- Dimension 4: parentheses break an optional-call chain too. ----
			{
				Code:   "const foo = (getObject?.()).foo;",
				Output: []string{"const {foo} = getObject?.();"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 32)},
			},
			// ---- Dimension 4: class static-block container form ----
			{
				Code:   "class C { static { const foo = object.foo; } }",
				Output: []string{"class C { static { const {foo} = object; } }"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 26, 1, 42)},
			},
			// ---- Dimension 4: same-kind nesting; both declarations are visited
			// independently and neither listener state bleeds across the function. ----
			{
				Code:   "const foo = outer.foo;\nfunction f() { const bar = inner.bar; }",
				Output: []string{"const {foo} = outer;\nfunction f() { const {bar} = inner; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 7, 1, 22),
					preferError("object", 2, 22, 2, 37),
				},
			},
			// ---- Dimension 4: graceful degradation — empty array binding
			// pattern still reaches the integer-index arm, which reports without
			// attempting an identifier-based fix. ----
			{
				Code:   "const [] = array[0];",
				Errors: []rule_tester.InvalidTestCaseError{preferError("array", 1, 7, 1, 20)},
			},
			// ---- Dimension 4: graceful degradation — object rest binding under
			// rename enforcement reports but is not autofixed. ----
			{
				Code:    "const {...rest} = object.foo;",
				Options: enforceRenamedObject,
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 29)},
			},
			// ---- Real-user #13678: a JSDoc cast comment inside the declaration
			// must suppress the otherwise-applicable fix. ----
			{
				Code: "const transactionData =\n      /** @type {Transaction} */\n      (data.transactionData);",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 7, 3, 29),
				},
			},
			// ---- Real-user #11836: rename enforcement intentionally reports a
			// member-expression assignment target, without offering a fix. ----
			{
				Code:    "let obj1, obj2;\nobj1.property1 = obj2.property2;",
				Options: enforceRenamedObject,
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 2, 1, 2, 32)},
			},
			// Locks in upstream fixIntoObjectDestructuring() arm 1: a comment
			// between retained receiver text and stripped parentheses blocks fix.
			{
				Code:   "const foo = ((/* c */ object)).foo;",
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 35)},
			},
			// ---- Dimension 3: multi-line whitespace is collapsed only inside
			// the replaced declarator; diagnostic end coordinates remain exact. ----
			{
				Code:   "const foo =\n  object\n    .foo;",
				Output: []string{"const {foo} = object;"},
				Errors: []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 3, 9)},
			},
			// Locks in upstream create() normalizedOptions arm 1 through the
			// bare-object JSON option path and same-name object check.
			{
				Code:    "const foo = source.foo;",
				Output:  []string{"const {foo} = source;"},
				Options: map[string]any{"object": true},
				Errors:  []rule_tester.InvalidTestCaseError{preferError("object", 1, 7, 1, 23)},
			},
		},
	)
}

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

func TestPreferDestructuringExtrasRealUser(t *testing.T) {
	assignmentsDisabled := []any{map[string]any{
		"AssignmentExpression": map[string]any{"array": false, "object": false},
	}}
	enforceObject := []any{
		map[string]any{"object": true},
		map[string]any{"enforceForRenamedProperties": true},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		[]rule_tester.ValidTestCase{
			// A dynamic collection index is not an integer literal and cannot be
			// represented by positional array destructuring.
			{Code: "const item = rows[index];"},
			// Renamed environment variables remain allowed by default.
			{Code: "const environment = process.env.ENVIRONMENT;"},
			// ESLint #16514: projects can explicitly keep control-flow
			// assignments when only declarations should be checked.
			{
				Code:    "let value;\nif (condition) {\n  value = source.value;\n}",
				Options: assignmentsDisabled,
			},
			// Member assignments are not identifier assignments and therefore
			// do not enter the default same-name object check.
			{Code: "cache.value = response.value;"},
			// Already-destructured application code is left untouched.
			{Code: "const { data } = await client.get(url);"},
		},
		[]rule_tester.InvalidTestCase{
			// ESLint #14918: build-time process.env access intentionally follows
			// the same same-name object rule as every other member expression.
			oneLineMatrixError(
				"const ENVIRONMENT = process.env.ENVIRONMENT;",
				7,
				"object",
				"const {ENVIRONMENT} = process.env;",
				nil,
				false,
			),
			{
				Code:   "const API_KEY = process.env.API_KEY; const DOMAIN = process.env.DOMAIN;",
				Output: []string{"const {API_KEY} = process.env; const {DOMAIN} = process.env;"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 7, 1, 36),
					preferError("object", 1, 44, 1, 71),
				},
			},
			// ESLint #16043: every integer Number literal is an array access,
			// including indexes beyond zero.
			oneLineMatrixError(
				"const second = \"some string\".split(\" \")[1];",
				7,
				"array",
				"",
				nil,
				false,
			),
			// A hook result is another common tuple-like access; array reports
			// are deliberately not autofixed because preceding slots matter.
			oneLineMatrixError(
				"const state = useState(initial)[0];",
				7,
				"array",
				"",
				nil,
				false,
			),
			{
				Code: "async function loadAll(tasks) { const results = await Promise.all(tasks); " +
					"const first = results[0]; return first; }",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("array", 1, 81, 1, 99),
				},
			},
			// ESLint #16514: assignment expressions are checked by default even
			// when the variable was declared earlier or assignment is nested in
			// control flow. Assignment diagnostics never carry an autofix.
			{
				Code: "let value;\nif (condition) {\n  value = source.value;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 3, 3, 3, 23),
				},
			},
			// Await, calls, and stripped receiver parentheses retain an exact
			// expression when converted to a declaration destructuring fix.
			{
				Code:   "async function load() { const data = (await client.get(url)).data; }",
				Output: []string{"async function load() { const {data} = await client.get(url); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 31, 1, 66),
				},
			},
			oneLineMatrixError(
				"const value = event.target.value;",
				7,
				"object",
				"const {value} = event.target;",
				nil,
				false,
			),
			oneLineMatrixError(
				"const current = ref.current;",
				7,
				"object",
				"const {current} = ref;",
				nil,
				false,
			),
			oneLineMatrixError(
				"const id = response.data.user.id;",
				7,
				"object",
				"const {id} = response.data.user;",
				nil,
				false,
			),
			oneLineMatrixError(
				"const join = require(\"path\").join;",
				7,
				"object",
				"const {join} = require(\"path\");",
				nil,
				false,
			),
			// Type arguments belong to the retained call receiver.
			oneLineMatrixError(
				"const data = client.get<Result>().data;",
				7,
				"object",
				"const {data} = client.get<Result>();",
				nil,
				false,
			),
			{
				Code:   "function Component(props) { const title = props.title; return title; }",
				Output: []string{"function Component(props) { const {title} = props; return title; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 35, 1, 54),
				},
			},
			{
				Code:   "function reducer(state) { const user = state.auth.user; return user; }",
				Output: []string{"function reducer(state) { const {user} = state.auth; return user; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 33, 1, 55),
				},
			},
			{
				Code: "async function route(request) { const id = request.params.id; " +
					"const token = request.headers.token; return id + token; }",
				Output: []string{
					"async function route(request) { const {id} = request.params; " +
						"const {token} = request.headers; return id + token; }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 39, 1, 61),
					preferError("object", 1, 69, 1, 98),
				},
			},
			{
				Code: "const zero = rows[0], one = rows[1], dynamic = rows[index];",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("array", 1, 7, 1, 21),
					preferError("array", 1, 23, 1, 36),
				},
			},
			// Renamed-property enforcement applies to real environment aliases
			// but still cannot offer the declaration-only same-name autofix.
			oneLineMatrixError(
				"const environment = process.env.ENVIRONMENT;",
				7,
				"object",
				"",
				enforceObject,
				false,
			),
			// UTF-16 columns and CRLF line boundaries must match ESLint while
			// the fix preserves all text outside the selected declarator.
			{
				Code:   "const before = \"😀\";\r\nconst value = event.target.value;",
				Output: []string{"const before = \"😀\";\r\nconst {value} = event.target;"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 2, 7, 2, 33),
				},
			},
		},
	)
}
