// TestPreferDestructuringExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch, Dimension 4 row, or real-user shape it
// covers so future refactors cannot silently regress it.
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
