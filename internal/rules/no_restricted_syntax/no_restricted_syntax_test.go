package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedSyntaxRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Upstream ESLint suite — string format
			// ============================================================
			{Code: `doSomething();`},
			{
				Code:    `var foo = 42;`,
				Options: []interface{}{"ConditionalExpression"},
			},
			{
				Code:    `foo += 42;`,
				Options: []interface{}{"VariableDeclaration", "FunctionExpression"},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[name="bar"]`},
			},
			{
				Code:    `() => 5`,
				Options: []interface{}{"ArrowFunctionExpression > BlockStatement"},
			},
			{
				Code:    `({ foo: 1, bar: 2 })`,
				Options: []interface{}{"Property > Literal.key"},
			},
			{
				Code:    `A: for (;;) break;`,
				Options: []interface{}{"BreakStatement[label]"},
			},
			{
				Code:    `function foo(bar, baz) {}`,
				Options: []interface{}{"FunctionDeclaration[params.length>2]"},
			},

			// ============================================================
			// Upstream ESLint suite — object format
			// ============================================================
			{
				Code: `var foo = 42;`,
				Options: []interface{}{
					map[string]interface{}{"selector": "ConditionalExpression"},
				},
			},
			{
				Code: `({ foo: 1, bar: 2 })`,
				Options: []interface{}{
					map[string]interface{}{"selector": "Property > Literal.key"},
				},
			},
			{
				Code: `({ foo: 1, bar: 2 })`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": "FunctionDeclaration[params.length>2]",
						"message":  "custom error message.",
					},
				},
			},

			// ============================================================
			// Upstream ESLint suite — regex flag attribute
			// ============================================================
			{
				Code:    `console.log(/a/);`,
				Options: []interface{}{"Literal[regex.flags=/./]"},
			},

			// ============================================================
			// Selector boundary: empty / nil options should never match
			// ============================================================
			{Code: `with (x) {}`},
			{Code: `var a = 1;`, Options: nil},
			{Code: `var a = 1;`, Options: []interface{}{}},

			// ============================================================
			// Selector boundary: unknown ESTree name silently ignored
			// (every node in the file is a non-match)
			// ============================================================
			{Code: `var a = 1;`, Options: []interface{}{"NotARealNodeType"}},

			// ============================================================
			// Selector boundary: malformed selector silently dropped
			// ============================================================
			{Code: `var a = 1;`, Options: []interface{}{"["}},
			{Code: `var a = 1;`, Options: []interface{}{"BinaryExpression["}},
			{Code: `var a = 1;`, Options: []interface{}{":nth-child(abc)"}},

			// ============================================================
			// Selector boundary: object option without `selector` is dropped
			// ============================================================
			{
				Code: `var a = 1;`,
				Options: []interface{}{
					map[string]interface{}{"message": "no selector key"},
				},
			},

			// ============================================================
			// :not — non-matching head means whole selector is a no-op
			// ============================================================
			{
				Code:    `with (x) {}`,
				Options: []interface{}{"FunctionDeclaration:not([generator=true])"},
			},

			// ============================================================
			// :is / :matches — neither alternative matches
			// ============================================================
			{
				Code:    `var a = 1;`,
				Options: []interface{}{":is(WithStatement, ConditionalExpression)"},
			},
			{
				Code:    `var a = 1;`,
				Options: []interface{}{":matches(WithStatement, DoWhileStatement)"},
			},

			// ============================================================
			// Attribute presence on a node that lacks the attribute
			// ============================================================
			{
				Code:    `function foo() {}`,
				Options: []interface{}{"FunctionDeclaration[generator]"},
			},
			{
				Code:    `function foo() {}`,
				Options: []interface{}{"FunctionDeclaration[async]"},
			},

			// ============================================================
			// Attribute equality / inequality
			// ============================================================
			{
				Code:    `let a = 1;`,
				Options: []interface{}{"VariableDeclaration[kind='const']"},
			},
			{
				Code:    `const a = 1;`,
				Options: []interface{}{"VariableDeclaration[kind!='const']"},
			},
			{
				Code:    `function foo(a) {}`,
				Options: []interface{}{"FunctionDeclaration[params.length>=2]"},
			},
			{
				Code:    `function foo(a, b, c) {}`,
				Options: []interface{}{"FunctionDeclaration[params.length<=2]"},
			},
			{
				Code:    `function foo(a, b) {}`,
				Options: []interface{}{"FunctionDeclaration[params.length<2]"},
			},

			// ============================================================
			// Logical / Assignment / Sequence operator differentiation —
			// ESTree splits these from BinaryExpression. tsgo fuses them,
			// so a bare `BinaryExpression` selector should NOT match a
			// pure assignment or logical or comma.
			// ============================================================
			{
				Code:    `a = b;`,
				Options: []interface{}{"BinaryExpression"},
			},
			{
				Code:    `a && b;`,
				Options: []interface{}{"BinaryExpression"},
			},
			{
				Code:    `(a, b);`,
				Options: []interface{}{"BinaryExpression"},
			},
			{
				Code:    `a + b;`,
				Options: []interface{}{"AssignmentExpression"},
			},
			{
				Code:    `a = b;`,
				Options: []interface{}{"LogicalExpression"},
			},
			{
				Code:    `a + b;`,
				Options: []interface{}{"SequenceExpression"},
			},

			// ============================================================
			// Unary vs Update — `++x` and `+x` should not collide
			// ============================================================
			{
				Code:    `+x;`,
				Options: []interface{}{"UpdateExpression"},
			},
			{
				Code:    `x++;`,
				Options: []interface{}{"UnaryExpression"},
			},

			// ============================================================
			// Optional-chain link selectors do not match plain access
			// ============================================================
			{
				Code:    `a.b.c;`,
				Options: []interface{}{"[optional=true]"},
			},
			{
				Code:    `a.b.c;`,
				Options: []interface{}{"ChainExpression"},
			},

			// ============================================================
			// Real-world: forbid var, allow let/const
			// ============================================================
			{
				Code:    `let a = 1; const b = 2;`,
				Options: []interface{}{"VariableDeclaration[kind='var']"},
			},

			// ============================================================
			// Real-world: forbid console.* (use member-access form)
			// ============================================================
			{
				Code: `myLogger.log('safe');`,
				Options: []interface{}{
					`CallExpression[callee.object.name='console']`,
				},
			},

			// ============================================================
			// Real-world: forbid for-in, allow for-of
			// ============================================================
			{
				Code:    `for (const x of items) {}`,
				Options: []interface{}{"ForInStatement"},
			},

			// ============================================================
			// Real-world: forbid default exports
			// ============================================================
			{
				Code:    `export const x = 1;`,
				Options: []interface{}{"ExportDefaultDeclaration"},
			},

			// ============================================================
			// Real-world: forbid setTimeout via callee.name selector
			// ============================================================
			{
				Code:    `clearTimeout(id);`,
				Options: []interface{}{`CallExpression[callee.name='setTimeout']`},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Upstream ESLint suite — string format
			// ============================================================
			{
				Code:    `var foo = 41;`,
				Options: []interface{}{"VariableDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'VariableDeclaration' is not allowed.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `;function lol(a) { return 42; }`,
				Options: []interface{}{"EmptyStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'EmptyStatement' is not allowed.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `try { voila(); } catch (e) { oops(); }`,
				Options: []interface{}{"TryStatement", "CallExpression", "CatchClause"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'TryStatement' is not allowed.", Line: 1, Column: 1},
					{MessageId: "restrictedSyntax", Message: "Using 'CallExpression' is not allowed.", Line: 1, Column: 7},
					{MessageId: "restrictedSyntax", Message: "Using 'CatchClause' is not allowed.", Line: 1, Column: 18},
					{MessageId: "restrictedSyntax", Message: "Using 'CallExpression' is not allowed.", Line: 1, Column: 30},
				},
			},
			{
				Code:    `bar;`,
				Options: []interface{}{`Identifier[name="bar"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'Identifier[name="bar"]' is not allowed.`,
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `bar;`,
				Options: []interface{}{"Identifier", `Identifier[name="bar"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Identifier' is not allowed."},
					{MessageId: "restrictedSyntax", Message: `Using 'Identifier[name="bar"]' is not allowed.`},
				},
			},
			{
				Code:    `() => {}`,
				Options: []interface{}{"ArrowFunctionExpression > BlockStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'ArrowFunctionExpression > BlockStatement' is not allowed.",
					},
				},
			},
			{
				Code:    `({ foo: 1, 'bar': 2 })`,
				Options: []interface{}{"Property > Literal.key"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'Property > Literal.key' is not allowed.",
					},
				},
			},
			{
				Code:    `A: for (;;) break A;`,
				Options: []interface{}{"BreakStatement[label]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'BreakStatement[label]' is not allowed.",
					},
				},
			},
			{
				Code:    `function foo(bar, baz, qux) {}`,
				Options: []interface{}{"FunctionDeclaration[params.length>2]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'FunctionDeclaration[params.length>2]' is not allowed.",
					},
				},
			},

			// ============================================================
			// Upstream ESLint suite — object format
			// ============================================================
			{
				Code: `var foo = 41;`,
				Options: []interface{}{
					map[string]interface{}{"selector": "VariableDeclaration"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'VariableDeclaration' is not allowed.",
					},
				},
			},
			{
				Code: `function foo(bar, baz, qux) {}`,
				Options: []interface{}{
					map[string]interface{}{"selector": "FunctionDeclaration[params.length>2]"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'FunctionDeclaration[params.length>2]' is not allowed.",
					},
				},
			},
			{
				Code: `function foo(bar, baz, qux) {}`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": "FunctionDeclaration[params.length>2]",
						"message":  "custom error message.",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "custom error message."},
				},
			},
			{
				// the custom message may contain literal `{{selector}}` —
				// ESLint substitutes the message into a templated outer
				// message but the inner string is taken verbatim.
				Code: `function foo(bar, baz, qux) {}`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": "FunctionDeclaration[params.length>2]",
						"message":  "custom message with {{selector}}",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "custom message with {{selector}}"},
				},
			},

			// ============================================================
			// Upstream — regex flag attribute, optional chain, using/await
			// using, regex on import path, :is(...)
			// ============================================================
			{
				Code:    `console.log(/a/i);`,
				Options: []interface{}{"Literal[regex.flags=/./]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'Literal[regex.flags=/./]' is not allowed.",
					},
				},
			},
			{
				Code:    `var foo = foo?.bar?.();`,
				Options: []interface{}{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'ChainExpression' is not allowed.",
					},
				},
			},
			{
				Code:    `var foo = foo?.bar?.();`,
				Options: []interface{}{"[optional=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using '[optional=true]' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using '[optional=true]' is not allowed."},
				},
			},
			// Locks in upstream esquery issue #110 — `:nth-child` should
			// match nodes sitting in NodeList positions (here:
			// ExpressionStatement at body[0]) and nothing else, even when
			// the file body contains an optional chain.
			{
				Code:    `a?.b`,
				Options: []interface{}{":nth-child(1)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using ':nth-child(1)' is not allowed."},
				},
			},
			// Locks in upstream esquery behaviour: `* ~ *` only matches
			// nodes that have a preceding sibling within the same NodeList
			// field on the parent — the second `<div/>` in the array
			// literal qualifies; nothing else in the file does.
			{
				Code:    `const foo = [<div/>, <div/>]`,
				Tsx:     true,
				Options: []interface{}{"* ~ *"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using '* ~ *' is not allowed."},
				},
			},
			{
				Code:    `{ using x = foo(); }`,
				Options: []interface{}{"VariableDeclaration[kind='using']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'VariableDeclaration[kind='using']' is not allowed.",
					},
				},
			},
			{
				Code:    `async function f() { await using x = foo(); }`,
				Options: []interface{}{"VariableDeclaration[kind='await using']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'VariableDeclaration[kind='await using']' is not allowed.",
					},
				},
			},
			{
				Code:    `import values from 'some/path';`,
				Options: []interface{}{`ImportDeclaration[source.value=/^some\/path$/]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'ImportDeclaration[source.value=/^some\/path$/]' is not allowed.`,
					},
				},
			},
			{
				Code: `foo + bar + baz`,
				Options: []interface{}{
					":is(Identifier[name='foo'], Identifier[name='bar'], Identifier[name='baz'])",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using ':is(Identifier[name='foo'], Identifier[name='bar'], Identifier[name='baz'])' is not allowed.",
					},
					{
						MessageId: "restrictedSyntax",
						Message:   "Using ':is(Identifier[name='foo'], Identifier[name='bar'], Identifier[name='baz'])' is not allowed.",
					},
					{
						MessageId: "restrictedSyntax",
						Message:   "Using ':is(Identifier[name='foo'], Identifier[name='bar'], Identifier[name='baz'])' is not allowed.",
					},
				},
			},

			// ============================================================
			// Selector boundary: wildcard registers across many kinds and
			// reports on every visited node. Limit blast radius with a
			// known-narrow file.
			// ============================================================
			{
				Code:    `42`,
				Options: []interface{}{"Literal"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal' is not allowed."},
				},
			},

			// ============================================================
			// Selector boundary: descendant combinator (whitespace) — match
			// at any nesting depth.
			// ============================================================
			{
				Code:    `function f() { return function g() {}; }`,
				Options: []interface{}{"FunctionDeclaration FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'FunctionDeclaration FunctionExpression' is not allowed."},
				},
			},

			// ============================================================
			// Selector boundary: adjacent sibling combinator
			// ============================================================
			{
				Code:    `f(1, 2);`,
				Options: []interface{}{"Literal + Literal"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal + Literal' is not allowed."},
				},
			},

			// ============================================================
			// Selector boundary: :not pseudo
			// ============================================================
			{
				Code:    `class A { foo() {} }`,
				Options: []interface{}{"MethodDefinition:not([static=true])"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition:not([static=true])' is not allowed."},
				},
			},
			{
				Code:    `class A { static foo() {} bar() {} }`,
				Options: []interface{}{"MethodDefinition:not([static=true])"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition:not([static=true])' is not allowed."},
				},
			},

			// ============================================================
			// Selector boundary: attribute regex with escaped slash
			// ============================================================
			{
				Code:    `import foo from "lodash/get";`,
				Options: []interface{}{`ImportDeclaration[source.value=/^lodash\//]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ImportDeclaration[source.value=/^lodash\//]' is not allowed.`},
				},
			},

			// ============================================================
			// Selector boundary: attribute regex with case-insensitive flag
			// ============================================================
			{
				Code:    `import foo from "Lodash";`,
				Options: []interface{}{`ImportDeclaration[source.value=/lodash/i]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ImportDeclaration[source.value=/lodash/i]' is not allowed.`},
				},
			},

			// ============================================================
			// Selector boundary: attribute equality with single-quote string
			// ============================================================
			{
				Code:    `let foo;`,
				Options: []interface{}{`VariableDeclaration[kind='let']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'VariableDeclaration[kind='let']' is not allowed.`},
				},
			},

			// ============================================================
			// Selector boundary: attribute presence on optional `id`
			// (function expression with name vs anonymous)
			// ============================================================
			{
				Code:    `(function named() {});`,
				Options: []interface{}{"FunctionExpression[id]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'FunctionExpression[id]' is not allowed."},
				},
			},

			// ============================================================
			// AssignmentExpression vs BinaryExpression refinement
			// ============================================================
			{
				Code:    `a = b;`,
				Options: []interface{}{"AssignmentExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'AssignmentExpression' is not allowed."},
				},
			},
			{
				Code:    `a += b;`,
				Options: []interface{}{"AssignmentExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'AssignmentExpression' is not allowed."},
				},
			},

			// ============================================================
			// LogicalExpression refinement
			// ============================================================
			{
				Code:    `a && b;`,
				Options: []interface{}{"LogicalExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'LogicalExpression' is not allowed."},
				},
			},
			{
				Code:    `a ?? b;`,
				Options: []interface{}{"LogicalExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'LogicalExpression' is not allowed."},
				},
			},

			// ============================================================
			// SequenceExpression refinement (comma operator)
			// ============================================================
			{
				Code:    `var x = (a, b);`,
				Options: []interface{}{"SequenceExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'SequenceExpression' is not allowed."},
				},
			},

			// ============================================================
			// UnaryExpression vs UpdateExpression refinement
			// ============================================================
			{
				Code:    `+x;`,
				Options: []interface{}{"UnaryExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UnaryExpression' is not allowed."},
				},
			},
			{
				Code:    `typeof x;`,
				Options: []interface{}{"UnaryExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UnaryExpression' is not allowed."},
				},
			},
			{
				Code:    `++x;`,
				Options: []interface{}{"UpdateExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UpdateExpression' is not allowed."},
				},
			},
			{
				Code:    `x--;`,
				Options: []interface{}{"UpdateExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UpdateExpression' is not allowed."},
				},
			},

			// ============================================================
			// MemberExpression encompasses both dotted and bracket access
			// ============================================================
			{
				Code:    `a.b;`,
				Options: []interface{}{"MemberExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MemberExpression' is not allowed."},
				},
			},
			{
				Code:    `a['b'];`,
				Options: []interface{}{"MemberExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MemberExpression' is not allowed."},
				},
			},

			// ============================================================
			// computed-vs-not member access via `[computed=true]`
			// ============================================================
			{
				Code:    `a[b];`,
				Options: []interface{}{"MemberExpression[computed=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MemberExpression[computed=true]' is not allowed."},
				},
			},

			// ============================================================
			// Property covers PropertyAssignment + Shorthand + Method +
			// Get/Set; selecting `Property` should fire on each
			// ============================================================
			{
				Code:    `({ a, b: 1, get c() {}, set d(v) {}, e() {} });`,
				Options: []interface{}{"Property"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'Property' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'Property' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'Property' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'Property' is not allowed."},
				},
			},

			// ============================================================
			// Spread in array vs object — both should match SpreadElement
			// ============================================================
			{
				Code:    `[...a];`,
				Options: []interface{}{"SpreadElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'SpreadElement' is not allowed."},
				},
			},
			{
				Code:    `({...a});`,
				Options: []interface{}{"SpreadElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'SpreadElement' is not allowed."},
				},
			},

			// ============================================================
			// Real-world: forbid `with`
			// ============================================================
			{
				Code:    `with (x) { y; }`,
				Options: []interface{}{"WithStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'WithStatement' is not allowed."},
				},
			},

			// ============================================================
			// Real-world: forbid console.* via callee.object.name selector
			// ============================================================
			{
				Code:    `console.log('msg');`,
				Options: []interface{}{`CallExpression[callee.object.name='console']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'CallExpression[callee.object.name='console']' is not allowed.`,
					},
				},
			},

			// ============================================================
			// Real-world: forbid setTimeout via callee.name
			// ============================================================
			{
				Code:    `setTimeout(fn, 0);`,
				Options: []interface{}{`CallExpression[callee.name='setTimeout']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'CallExpression[callee.name='setTimeout']' is not allowed.`,
					},
				},
			},

			// ============================================================
			// Real-world: forbid for-in
			// ============================================================
			{
				Code:    `for (const k in obj) {}`,
				Options: []interface{}{"ForInStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ForInStatement' is not allowed."},
				},
			},

			// ============================================================
			// Real-world: forbid default exports
			// ============================================================
			{
				Code:    `export default 1;`,
				Options: []interface{}{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ExportDefaultDeclaration' is not allowed."},
				},
			},

			// ============================================================
			// Real-world: ban any `new Promise(...)` — encourage
			// `Promise.resolve` / `Promise.reject`
			// ============================================================
			{
				Code:    `new Promise(function (res, rej) { res(); });`,
				Options: []interface{}{`NewExpression[callee.name='Promise']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'NewExpression[callee.name='Promise']' is not allowed.`},
				},
			},

			// ============================================================
			// Real-world: forbid arrow fns whose body IS a BlockStatement
			// (i.e. require concise body). Exercises an attribute path
			// that resolves to a node and then reads the synthetic `type`
			// to compare against an ESTree name.
			// ============================================================
			{
				Code:    `const f = () => { return 5; };`,
				Options: []interface{}{"ArrowFunctionExpression[body.type='BlockStatement']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'ArrowFunctionExpression[body.type='BlockStatement']' is not allowed.",
					},
				},
			},

			// ============================================================
			// Real-world: ban async functions in a specific module
			// ============================================================
			{
				Code:    `async function f() {}`,
				Options: []interface{}{"FunctionDeclaration[async=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'FunctionDeclaration[async=true]' is not allowed."},
				},
			},

			// ============================================================
			// Real-world: ban generator functions
			// ============================================================
			{
				Code:    `function* g() {}`,
				Options: []interface{}{"FunctionDeclaration[generator=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'FunctionDeclaration[generator=true]' is not allowed."},
				},
			},

			// ============================================================
			// Real-world: object-format mixed with string-format options
			// ============================================================
			{
				Code: `var a = 1; with (x) {}`,
				Options: []interface{}{
					"WithStatement",
					map[string]interface{}{
						"selector": "VariableDeclaration",
						"message":  "Use let or const, not var.",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					// Order: selectors fire in registration order on each node;
					// nodes are visited in source order. VariableStatement
					// comes first, then WithStatement.
					{MessageId: "restrictedSyntax", Message: "Use let or const, not var."},
					{MessageId: "restrictedSyntax", Message: "Using 'WithStatement' is not allowed."},
				},
			},

			// ============================================================
			// Selector boundary: nested `:is(...)` inside a combinator —
			// match the outermost optional chain only
			// ============================================================
			{
				Code:    `a?.b?.c`,
				Options: []interface{}{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ChainExpression' is not allowed."},
				},
			},

			// ============================================================
			// Mixing nodes with shared fields: VariableDeclarator (the
			// inner declarator) vs VariableDeclaration (the statement)
			// ============================================================
			{
				Code:    `var a = 1, b = 2;`,
				Options: []interface{}{"VariableDeclarator"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'VariableDeclarator' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'VariableDeclarator' is not allowed."},
				},
			},
		},
	)
}
