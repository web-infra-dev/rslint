// TestNoConfusingNonNullAssertionExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package no_confusing_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConfusingNonNullAssertionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingNonNullAssertionRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: Receiver / expression wrappers — parenthesized non-null on left ----
		// tsgo wraps in ParenthesizedExpression; left.End() lands after `)`, not `!`. Must NOT report.
		{Code: `(a!) == b;`},
		{Code: `(a!) === b;`},
		{Code: `(a!) = b;`},
		{Code: `(a!) in b;`},
		{Code: `(a!) instanceof b;`},
		// ---- Dimension 4: Multi-level parens on left ----
		{Code: `((a!)) == b;`},
		{Code: `((((a!)))) == b;`},
		// ---- Tokenizer: greedy `!=` / `!==` match — without intervening space, `!` is consumed into the operator, not into a non-null assertion ----
		{Code: `a!=b;`},
		{Code: `a!==b;`},
		{Code: `a != b;`},
		{Code: `a !== b;`},
		// ---- Real-user: type-guard idiom — the very pattern users type instead of `value! == null` — must never be flagged ----
		{Code: `const value: string | null = null; value! != null;`},
		{Code: `const value: string | undefined = undefined; value! !== undefined;`},
		// ---- ForInStatement / ForOfStatement: `in` / `of` here are statement keywords, NOT BinaryExpression operators — listener doesn't fire ----
		{Code: `for (const k in obj) {}`},
		{Code: `for (const k of arr) {}`},
		// ---- Locks in upstream isConfusingOperator() arm "false": non-confusing operators are ignored ----
		// `!=` and `!==` are the very operators upstream wants users to type instead — they must stay valid.
		{Code: `a! != b;`},
		{Code: `a! !== b;`},
		// Other comparison / arithmetic / logical / nullish / compound-assignment ops are not in the watched set.
		{Code: `a! < b;`},
		{Code: `a! > b;`},
		{Code: `a! <= b;`},
		{Code: `a! >= b;`},
		{Code: `a! + b;`},
		{Code: `a! - b;`},
		{Code: `a! * b;`},
		{Code: `a! / b;`},
		{Code: `a! && b;`},
		{Code: `a! || b;`},
		{Code: `a! ?? b;`},
		{Code: `a! += b;`},
		{Code: `a! -= b;`},
		{Code: `a! ??= b;`},
		{Code: `a! ||= b;`},
		{Code: `a! &&= b;`},
		{Code: `a! , b;`},
		// ---- Locks in upstream leftHandFinalToken arm "no !": left side doesn't end in `!` ----
		{Code: `a == b;`},
		{Code: `a === b;`},
		{Code: `a = b;`},
		{Code: `a in b;`},
		{Code: `a instanceof b;`},
		// ---- Locks in upstream check that only LEFT-side ! triggers (RHS ! is fine) ----
		{Code: `a == b!;`},
		{Code: `a === b!;`},
		{Code: `a in b!;`},
		{Code: `a instanceof b!;`},
		// ---- Dimension 4: Receiver wrappers — ! on the wrapper position ----
		// (a + b!) plus then operator: parens make it valid.
		{Code: `(a || b!) == c;`},
		{Code: `(a && b!) === c;`},
		{Code: `(a + b!) instanceof c;`},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: Double non-null (a!!) on LHS ----
		// left = NonNullExpression(NonNullExpression(a)); only outer-most `!` is removed by the suggestion.
		{
			Code: `a!! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a! == b;`},
					},
				},
			},
		},
		// ---- Dimension 4: Element-access receiver — x[0]! == y ----
		{
			Code: `arr[0]! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `arr[0] == b;`},
					},
				},
			},
		},
		// ---- Dimension 4: Optional-chain receiver — foo?.bar! == y ----
		// In tsgo the optional chain is a flag on PropertyAccess, not a wrapper node.
		// The NonNullExpression sits above it, so left.Kind is NonNullExpression as usual.
		{
			Code: `foo?.bar! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `foo?.bar == b;`},
					},
				},
			},
		},
		// ---- Dimension 4: Type-assertion receiver — (a as B)! == c ----
		{
			Code: `(a as number)! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `(a as number) == b;`},
					},
				},
			},
		},
		// ---- Locks in upstream switch operator==='===' arm separately from '==' ----
		{
			Code: `a + b! === c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `(a + b!) === c;`},
					},
				},
			},
		},
		// ---- Locks in upstream switch operator==='in' wrapUpLeft suggestion when left.type != TSNonNullExpression ----
		{
			Code: `a + b! in c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `(a + b!) in c;`},
					},
				},
			},
		},
		// ---- Locks in upstream switch operator==='instanceof' wrapUpLeft suggestion when left.type != TSNonNullExpression ----
		// `+` (precedence 12) binds tighter than `instanceof` (11), so the LHS is BinaryExpression(a + NonNullExpression(b)),
		// not a NonNullExpression — exercises the wrapUpLeft-only branch.
		{
			Code: `a + b! instanceof c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `(a + b!) instanceof c;`},
					},
				},
			},
		},
		// N/A: upstream's `=` × wrapUpLeft branch (left isn't TSNonNullExpression) is unreachable in
		// parseable TS — assignment requires an assignable LHS, which a `+`/`||`-style BinaryExpression
		// is not. Upstream lists the suggestion defensively but has no test for it; rslint inherits
		// that gap (cannot fixture the case without producing a syntax error).
		// ---- Real-user: array-element non-null before `instanceof` (closed-issue style) ----
		{
			Code: `items[0]! instanceof Foo;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `items[0] instanceof Foo;`},
						{MessageId: "wrapUpLeft", Output: `(items[0]!) instanceof Foo;`},
					},
				},
			},
		},
		// ---- Real-user: `this`-qualified field non-null before `===` ----
		{
			Code: `this.value! === null;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `this.value === null;`},
					},
				},
			},
		},
		// ---- Real-user: call-expression result non-null before `in` ----
		{
			Code: `getKey()! in container;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `getKey() in container;`},
						{MessageId: "wrapUpLeft", Output: `(getKey()!) in container;`},
					},
				},
			},
		},
		// ---- Locks in nested BinaryExpression: outer flagged, inner skipped ----
		// Both BinaryExpressions visit the listener; only the outer should fire.
		{
			Code: `a! == (b! == c);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a == (b! == c);`},
					},
				},
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a! == (b == c);`},
					},
				},
			},
		},
		// ---- Confirms message-text exact format for `in` carries the operator data ----
		{
			Code: `a! in b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Message:   "Confusing combination of non-null assertion and `in` operator like `a! in b`, which might be misinterpreted as `!(a in b)`.",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `a in b;`},
						{MessageId: "wrapUpLeft", Output: `(a!) in b;`},
					},
				},
			},
		},
		// ---- Confirms message-text exact format for `instanceof` carries the operator data ----
		{
			Code: `a! instanceof b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Message:   "Confusing combination of non-null assertion and `instanceof` operator like `a! instanceof b`, which might be misinterpreted as `!(a instanceof b)`.",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `a instanceof b;`},
						{MessageId: "wrapUpLeft", Output: `(a!) instanceof b;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: NewExpression — new Foo()! == bar ----
		{
			Code: `new Foo()! == bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `new Foo() == bar;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: plain CallExpression — foo()! == bar ----
		{
			Code: `foo()! == bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `foo() == bar;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: generic call — foo<string>()! == bar (TypeArguments on CallExpression) ----
		{
			Code: `declare function foo<T>(): T; foo<string>()! == bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `declare function foo<T>(): T; foo<string>() == bar;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: TaggedTemplateExpression — tag`x`! == bar ----
		{
			Code: "foo`x`! == bar;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: "foo`x` == bar;"},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: deep chained PropertyAccess — a.b.c.d! == e ----
		{
			Code: `a.b.c.d! == e;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a.b.c.d == e;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: chained non-null assertions — a!.b!.c! == d ----
		// Only the outer-most `!` is removed; the inner ones stay (each ! is a distinct NonNullExpression node).
		{
			Code: `a!.b!.c! == d;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a!.b!.c == d;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: `as const` wrapper ----
		{
			Code: `(a as const)! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `(a as const) == b;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: `satisfies` expression (TS-only, distinct AST kind from `as`) ----
		{
			Code: `(a satisfies number)! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `(a satisfies number) == b;`},
					},
				},
			},
		},
		// ---- Dimension 4 receiver: `this` keyword as receiver — this! == null ----
		{
			Code: `class C { f() { this! == null; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `class C { f() { this == null; } }`},
					},
				},
			},
		},
		// ---- wrapUpLeft-only branch with TypeOfExpression on LHS (different non-NonNull AST kind) ----
		{
			Code: `typeof a! == 'string';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `(typeof a!) == 'string';`},
					},
				},
			},
		},
		// ---- wrapUpLeft-only branch with VoidExpression on LHS ----
		{
			Code: `void a! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `(void a!) == b;`},
					},
				},
			},
		},
		// ---- wrapUpLeft-only branch with AwaitExpression on LHS (inside async function) ----
		{
			Code: `async function f() { await x! == y; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `async function f() { (await x!) == y; }`},
					},
				},
			},
		},
		// ---- Whitespace tolerance: extra spaces between `!` and `==` are still flagged (`a !` is still NonNullExpression(a)) ----
		{
			Code: `a!  ==  b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a  ==  b;`},
					},
				},
			},
		},
		// ---- Whitespace tolerance: multi-line gap between `!` and the operator ----
		{
			Code: "foo()!\n  == bar;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: "foo()\n  == bar;"},
					},
				},
			},
		},
		// ---- Context: BinaryExpression as a call-argument — listener fires on the nested BinaryExpression ----
		{
			Code: `foo(a! == b);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `foo(a == b);`},
					},
				},
			},
		},
		// ---- Context: BinaryExpression as a ForStatement condition ----
		{
			Code: `for (let i = 0; i! == n; i++) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `for (let i = 0; i == n; i++) {}`},
					},
				},
			},
		},
		// ---- Context: BinaryExpression as object-literal property value ----
		{
			Code: `const o = { x: a! == b };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `const o = { x: a == b };`},
					},
				},
			},
		},
		// ---- Context: BinaryExpression as return expression in arrow body ----
		{
			Code: `const f = () => a! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `const f = () => a == b;`},
					},
				},
			},
		},
		// ---- Real-user: optional-chain-into-property non-null before `===` (common React-state pattern) ----
		{
			Code: `state?.value! === 'pending';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `state?.value === 'pending';`},
					},
				},
			},
		},
		// ---- Real-user: array element instanceof guard (post-find pattern) ----
		{
			Code: `(arr.find(x => x.id === id))! instanceof Item;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `(arr.find(x => x.id === id)) instanceof Item;`},
						{MessageId: "wrapUpLeft", Output: `((arr.find(x => x.id === id))!) instanceof Item;`},
					},
				},
			},
		},
	})
}
