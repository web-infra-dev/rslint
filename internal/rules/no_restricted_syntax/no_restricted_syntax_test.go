package no_restricted_syntax

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const lazyMustBeWrappedSelector = `:matches(CallExpression[callee.object.name='React'][callee.property.name='lazy'], CallExpression[callee.name='lazy']):not([parent.callee.name='withSuspense'])`

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
				Options: []interface{}{"FunctionDeclaration[generator=true]"},
			},
			{
				Code:    `function foo() {}`,
				Options: []interface{}{"FunctionDeclaration[async=true]"},
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
			{
				Code: `
const A = withSuspense(lazy(() => import('./a')));
const B = withSuspense(React.lazy(() => import('./b')));
`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": lazyMustBeWrappedSelector,
						"message":  "lazy must be wrapped by withSuspense",
					},
				},
			},

			// ============================================================
			// JSDoc is skipped
			// ============================================================
			{
				Code:     "/** @param {number | null} exitCode */\nfunction finish(exitCode) {}",
				FileName: "file.mjs",
				Options:  []interface{}{`Literal[value=null]`},
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code:     "/** @param {(err?: NodeJS.ErrnoException | null) => void} done */\nfunction finish(done) {}",
				FileName: "file.cjs",
				Options:  []interface{}{`Literal[value=null]`},
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code:     "/** @param {Foo} x */\nfunction f(x) {}",
				FileName: "file.mjs",
				Options:  []interface{}{`FunctionDeclaration:has(Identifier[name='Foo'])`},
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code:     "/** @this {HTMLElement} */\nfunction f(value) {}",
				FileName: "file.mjs",
				Options:  []interface{}{`FunctionDeclaration[params.length=2]`},
				TSConfig: "tsconfig.allow-js.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "const config = /** @satisfies {{ value: number }} */ ({ value: 1 });",
				FileName: "file.mjs",
				Options:  []interface{}{"ObjectExpression"},
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
				},
			},
			// Control for the JSDoc :has() case above: the same recursive walk
			// still sees authored runtime children of the function.
			{
				Code:     "/** @param {Foo} x */\nfunction f(x) {}",
				FileName: "file.mjs",
				Options:  []interface{}{`FunctionDeclaration:has(Identifier[name='x'])`},
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'FunctionDeclaration:has(Identifier[name='x'])' is not allowed.",
					},
				},
			},

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
			{
				Code: `
const A = lazy(() => import('./a'));
const B = React.lazy(() => import('./b'));
`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": lazyMustBeWrappedSelector,
						"message":  "lazy must be wrapped by withSuspense",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "lazy must be wrapped by withSuspense"},
					{MessageId: "restrictedSyntax", Message: "lazy must be wrapped by withSuspense"},
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

func TestNoRestrictedSyntaxEsqueryParity(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `import('./static');`,
				Options: []interface{}{"CallExpression"},
			},
			{
				Code:    `import(source);`,
				Options: []interface{}{`ImportExpression[callee]`},
			},
			{
				Code:    `import(source);`,
				Options: []interface{}{`ImportExpression[arguments]`},
			},
			{
				Code:    `import('./static');`,
				Options: []interface{}{`ImportExpression[source.type!='Literal']`},
			},
			{
				Code:    `require('./static');`,
				Options: []interface{}{`CallExpression[callee.name='require'][arguments.0.type!='Literal']`},
			},
			{
				Code:    `const A = withSuspense((lazy(() => import('./a'))));`,
				Options: []interface{}{lazyMustBeWrappedSelector},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[missing='value']`},
			},
			{
				// On JS files Identifier has no `optional` property (in TS,
				// @typescript-eslint/parser sets Identifier.optional = false).
				Code:     `foo;`,
				FileName: "identifier-optional.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []interface{}{`Identifier[optional]`},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[raw]`},
			},
			{
				Code:    `class C { async method() {} }`,
				Options: []interface{}{`MethodDefinition[async]`},
			},
			{
				Code:    `const fn = function () {};`,
				Options: []interface{}{`FunctionExpression[id]`},
			},
			{
				Code:    `const fn = function () {};`,
				Options: []interface{}{`FunctionExpression[id.name='fn']`},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[name='foo'][name='bar']`},
			},
			{
				Code:    `function outer() { foo; }`,
				Options: []interface{}{`BlockStatement:has(Program Identifier[name='foo'])`},
			},
			{
				Code: `/* eslint-disable test */
foo;`,
				Options: []interface{}{`Identifier[name='foo']`},
			},
			{
				Code: `// eslint-disable-next-line test
foo;`,
				Options: []interface{}{`Identifier[name='foo']`},
			},
			{
				Code:    `foo; // eslint-disable-line test`,
				Options: []interface{}{`Identifier[name='foo']`},
			},
			{
				Code:    `const x = <Foo.Bar />;`,
				Options: []interface{}{`MemberExpression`},
				Tsx:     true,
			},
			{
				// ESLint's SourceCodeTraverser dispatches top-level identifier selectors
				// by exact case, so lowercase `identifier` matches nothing at top level.
				Code:    `foo;`,
				Options: []interface{}{`identifier[name='foo']`},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `foo();`,
				Options: []interface{}{"CallExpression"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `for (let index = 0; index < 1; index++) {}`,
				Options: []interface{}{"VariableDeclaration"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `for (let index = 0; index < 1; index++) {}`,
				Options: []interface{}{"ForStatement > VariableDeclaration"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `import(source);`,
				Options: []interface{}{`ImportExpression[source.type!='Literal']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `import data from "./data.json" with { type: "json" };`,
				Options: []interface{}{`ImportAttribute[key.name='type'][value.value='json']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `require(source);`,
				Options: []interface{}{`CallExpression[callee.name='require'][arguments.0.type!='Literal']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `object.property;`,
				Options: []interface{}{"MemberExpression[computed]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo();`,
				Options: []interface{}{"CallExpression[optional]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `'text';`,
				Options: []interface{}{"Literal[raw]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `function regular() {}`,
				Options: []interface{}{"FunctionDeclaration[async]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const arrow = () => {};`,
				Options: []interface{}{"ArrowFunctionExpression[generator]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[missing!='value']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo();`,
				Options: []interface{}{`CallExpression[callee.property.name='undefined']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const fn = function () {};`,
				Options: []interface{}{`FunctionExpression[id=null]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const fn = function () {};`,
				Options: []interface{}{`FunctionExpression[id.name=null]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const fn = function () {};`,
				Options: []interface{}{`FunctionExpression[id.name=type(object)]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const fn = function () {};`,
				Options: []interface{}{`FunctionExpression[id<1]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `function noArgs() {}`,
				Options: []interface{}{`FunctionDeclaration[params='']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `function noArgs() {}`,
				Options: []interface{}{`FunctionDeclaration[params<1]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const empty = "";`,
				Options: []interface{}{`Literal[value<1]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[name<'z']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[name='foo'][name='foo']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code: `foo;`,
				Options: []interface{}{
					map[string]interface{}{"selector": `[name='foo']`, "message": "universe first"},
					map[string]interface{}{"selector": `Identifier[name='foo']`, "message": "specific second"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "universe first"},
					{MessageId: "restrictedSyntax", Message: "specific second"},
				},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[name=type(string)]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`#Identifier[name='foo']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `function f() {}`,
				Options: []interface{}{":function"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const value = 42;`,
				Options: []interface{}{`.body.declarations.init`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `x;`,
				Options: []interface{}{`Identifier:has(Identifier[name='x'])`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `if (x) { y(); }`,
				Options: []interface{}{`IfStatement:has(> Identifier[name='x'])`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `if ((x)) { y(); }`,
				Options: []interface{}{`IfStatement:has(> Identifier[name='x'])`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `(foo);`,
				Options: []interface{}{"*"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier[name=/\u{66}oo/u]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const value = "\n";`,
				Options: []interface{}{`Literal[value='\n']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const slash = "/";`,
				Options: []interface{}{`Literal[value=/[/]/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const value = .5;`,
				Options: []interface{}{`Literal[value=.5]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code: `foo;`,
				Options: []interface{}{
					map[string]interface{}{"selector": "Identifier", "message": "first"},
					map[string]interface{}{"selector": "Identifier", "message": "last"},
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Message: "last"}},
			},
			{
				Code:    `const A = withSuspense(memo(lazy(() => import('./a'))));`,
				Options: []interface{}{lazyMustBeWrappedSelector},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo(); bar(); baz();`,
				Options: []interface{}{`ExpressionStatement + !ExpressionStatement`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `foo;`,
				Options: []interface{}{`Identifier:exit`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Message:   "Using 'Identifier:exit' is not allowed.",
				}},
			},
			{
				Code:    `class A { foo() {} x = 1; static {} }`,
				Options: []interface{}{`ClassBody`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 38,
				}},
			},
			{
				Code:    `class A { foo() {} x = 1; static {} }`,
				Options: []interface{}{`ClassDeclaration > ClassBody[type='ClassBody'][body.length=3]`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 38,
				}},
			},
			{
				Code:    `class A { foo() {} x = 1; static {} }`,
				Options: []interface{}{`ClassBody > MethodDefinition`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`ClassBody[parent.type='ClassDeclaration']`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`:is(ClassDeclaration, ClassBody)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = <Foo.Bar baz={value}>{/* hi */}{children}</Foo.Bar>;`,
				Options: []interface{}{`JSXMemberExpression`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = <Foo.Bar baz={value}>{/* hi */}{children}</Foo.Bar>;`,
				Options: []interface{}{`JSXIdentifier`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = <Foo.Bar baz={value}>{/* hi */}{children}</Foo.Bar>;`,
				Options: []interface{}{`Identifier`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = <div>{/* hi */}</div>;`,
				Options: []interface{}{`JSXExpressionContainer > JSXEmptyExpression`},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <div>{/* hi */}</div>;`,
				Options: []interface{}{`JSXEmptyExpression[parent.type='JSXExpressionContainer']`},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <div>{/* hi */}</div>;`,
				Options: []interface{}{`:is(JSXExpressionContainer, JSXEmptyExpression)`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`[type='ClassBody']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`ClassDeclaration ClassBody`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`ClassDeclaration:has(> ClassBody)`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C {}`,
				Options: []interface{}{`ClassBody[id]`},
			},
			{
				Code:    `class C {}`,
				Options: []interface{}{`ClassBody:not([id])`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <div>{/* hi */}</div>;`,
				Options: []interface{}{`[type='JSXEmptyExpression']`},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <div>{/* hi */}</div>;`,
				Options: []interface{}{`JSXExpressionContainer JSXEmptyExpression`},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <div>{/* hi */}</div>;`,
				Options: []interface{}{`JSXExpressionContainer:has(> JSXEmptyExpression)`},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo.Bar />;`,
				Options: []interface{}{`JSXMemberExpression[type='JSXMemberExpression']`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = <Foo.Bar />;`,
				Options: []interface{}{`JSXMemberExpression[object.type='JSXIdentifier']`},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo />;`,
				Options: []interface{}{`[type='JSXElement']`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`MethodDefinition[value.parent.type='MethodDefinition']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`MethodDefinition > FunctionExpression`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`FunctionExpression > BlockStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m<T>() {} }`,
				Options: []interface{}{`FunctionExpression`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code:    `class C { async m<T extends X>() {} }`,
				Options: []interface{}{`FunctionExpression`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
			{
				Code:    `class C { m<T extends U>() {} }`,
				Options: []interface{}{`FunctionExpression`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 30,
				}},
			},
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{`TSEnumDeclaration[id.name='E']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{`TSEnumDeclaration[body.type='TSEnumBody']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{`TSEnumBody[members.length=2]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{`TSEnumBody > TSEnumMember`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = 0x1n + 1_000n;`,
				Options: []interface{}{`Literal[bigint='1']`, `Literal[bigint='1000']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `class C { m() {} /* } */ }`,
				Options: []interface{}{`ClassBody`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1, Column: 9, EndLine: 1, EndColumn: 27,
				}},
			},
			{
				Code:    `@sealed() class A {}`,
				Options: []interface{}{`Decorator[expression.type='CallExpression']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo();`,
				Options: []interface{}{`CallExpression:exit`, `Identifier`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Identifier' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'CallExpression:exit' is not allowed."},
				},
			},
		},
	)
}

func TestNoRestrictedSyntaxOptimizedDispatch(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `class C {}`,
				Options: []interface{}{
					"ClassDeclaration[superClass.name='Base']",
					"ClassDeclaration[superClass.name='Other']",
				},
			},
			{
				Code: `let value;`,
				Options: []interface{}{
					"VariableDeclarator[init.name='foo']",
					"VariableDeclarator[init.name='bar']",
				},
			},
			{
				Code: `value instanceof CustomEvent;`,
				Options: []interface{}{
					`BinaryExpression[operator='instanceof'][right.name=/^HTML\w+/]`,
					`BinaryExpression[operator='instanceof'][right.name=/^SVG\w+/]`,
				},
			},
			{
				Code: `document['body'];`,
				Options: []interface{}{
					`MemberExpression[property.name='body']`,
					`MemberExpression[property.name='head']`,
				},
			},
			{
				Code:    `new Promise(resolve => resolve());`,
				Options: []interface{}{`NewExpression[callee.object.name='Intl']`},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `foo;`,
				Options: []interface{}{
					map[string]interface{}{"selector": `Identifier[name=/^f/]`, "message": "regex start"},
					map[string]interface{}{"selector": `Identifier[name='foo']`, "message": "exact"},
					map[string]interface{}{"selector": `Identifier[name=/o$/]`, "message": "regex end"},
					map[string]interface{}{"selector": `Identifier[name='bar']`, "message": "other exact"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "exact"},
					{MessageId: "restrictedSyntax", Message: "regex start"},
					{MessageId: "restrictedSyntax", Message: "regex end"},
				},
			},
			{
				Code: `foo();`,
				Options: []interface{}{
					`CallExpression[callee.property.name='query']`,
					map[string]interface{}{"selector": `CallExpression[callee.name=/^foo$/]`, "message": "fallback"},
					`CallExpression[callee.property.name='find']`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "fallback"},
				},
			},
			{
				Code: `'foo';`,
				Options: []interface{}{
					`Literal[value='foo']`,
					`Literal[value='bar']`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal[value='foo']' is not allowed."},
				},
			},
			{
				Code: `document.body;`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": `MemberExpression[object.name='document'][property.name='body']`,
						"message":  "first duplicate",
					},
					map[string]interface{}{
						"selector": `MemberExpression[object.name='document'][property.name='body']`,
						"message":  "second duplicate",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "second duplicate"},
				},
			},
			{
				Code: `value instanceof HTMLDivElement;`,
				Options: []interface{}{
					`BinaryExpression[operator='instanceof'][right.name=/^HTML\w+/]`,
					`BinaryExpression[operator='instanceof'][right.name=/^SVG\w+/]`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'BinaryExpression[operator='instanceof'][right.name=/^HTML\\w+/]' is not allowed.",
					},
				},
			},
			{
				Code:    `new Intl.DateTimeFormat();`,
				Options: []interface{}{`NewExpression[callee.object.name='Intl']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   "Using 'NewExpression[callee.object.name='Intl']' is not allowed.",
					},
				},
			},
		},
	)
}

func TestNoRestrictedSyntaxDispatchPlanning(t *testing.T) {
	binaryEntries := []ruleEntry{
		buildEntryFromString(`BinaryExpression[operator='instanceof'][right.name='MouseEvent']`),
		buildEntryFromString(`BinaryExpression[operator='instanceof'][right.name=/^HTML\w+/]`),
		buildEntryFromString(`BinaryExpression[operator='instanceof'][right.name=/^SVG\w+/]`),
		buildEntryFromString(`BinaryExpression[operator='instanceof'][right.name='KeyboardEvent']`),
		buildEntryFromString(`BinaryExpression[operator='instanceof'][right.name='PointerEvent']`),
		buildEntryFromString(`BinaryExpression[operator='instanceof'][right.name='DragEvent']`),
	}
	binaryBucket := buildRuleBucket(binaryEntries)
	if binaryBucket.dispatch == nil || len(binaryBucket.dispatch.path) != 1 || binaryBucket.dispatch.path[0].raw != "operator" {
		t.Fatalf("binary dispatch path is %v, want operator", binaryBucket.dispatch)
	}
	if binaryBucket.dispatch.byValue != nil || binaryBucket.dispatch.singleValue != "instanceof" {
		t.Fatalf("single-value dispatch = (%q, byValue=%v), want singleValue=instanceof", binaryBucket.dispatch.singleValue, binaryBucket.dispatch.byValue)
	}

	singleBucket := buildRuleBucket([]ruleEntry{
		buildEntryFromString(`NewExpression[callee.object.name='Intl']`),
	})
	if singleBucket.dispatch != nil {
		t.Fatalf("single-selector bucket unexpectedly built dispatch %v", singleBucket.dispatch)
	}

	multiBucket := buildRuleBucket([]ruleEntry{
		buildEntryFromString(`NewExpression[callee.object.name='Intl']`),
		buildEntryFromString(`NewExpression[callee.object.name='Date']`),
	})
	if multiBucket.dispatch == nil || len(multiBucket.dispatch.path) != 3 ||
		multiBucket.dispatch.path[0].raw != "callee" ||
		multiBucket.dispatch.path[1].raw != "object" ||
		multiBucket.dispatch.path[2].raw != "name" {
		t.Fatalf("multi-selector dispatch path is %v, want callee.object.name", multiBucket.dispatch)
	}

	valueBucket := buildRuleBucket([]ruleEntry{
		buildEntryFromString(`Literal[value='text']`),
		buildEntryFromString(`Literal[value='other']`),
	})
	if valueBucket.dispatch == nil || len(valueBucket.dispatch.path) != 1 || valueBucket.dispatch.path[0].raw != "value" {
		t.Fatalf("Literal.value dispatch path is %v, want value", valueBucket.dispatch)
	}

	chainedBucket := buildRuleBucket([]ruleEntry{
		buildEntryFromString(`CallExpression[callee.name='eval']`),
		buildEntryFromString(`CallExpression[callee.name='Function']`),
		buildEntryFromString(`CallExpression[callee.property.name='log']`),
		buildEntryFromString(`CallExpression[callee.property.name='warn']`),
	})
	if chainedBucket.dispatch == nil || chainedBucket.dispatch.next == nil {
		t.Fatalf("expected chained dispatch for mixed CallExpression paths, got %+v", chainedBucket.dispatch)
	}

	unsupportedBucket := buildRuleBucket([]ruleEntry{
		buildEntryFromString(`CallExpression[arguments.length='1']`),
		buildEntryFromString(`CallExpression[arguments.length='2']`),
	})
	if unsupportedBucket.dispatch != nil {
		t.Fatalf("unsupported path unexpectedly built dispatch %v", unsupportedBucket.dispatch)
	}
}

func TestNoRestrictedSyntaxDispatchMatchesFullMatcher(t *testing.T) {
	entries := parseRuleOptions([]interface{}{
		`Identifier[name='foo']`,
		`Identifier[name='bar']`,
		`Identifier[name=/^f/]`,
		`Identifier[name='foo'][name='foo']`,
		`Identifier[name='foo'][name='bar']`,
		`Identifier[name='foo'][type='Identifier']`,
		`Identifier[missing='undefined']`,
		`CallExpression[callee.name=undefined]`,
		`CallExpression[callee.name='eval']`,
		`FunctionExpression[id.name=null]`,
		`FunctionExpression[id.name='named']`,
	})
	bucket := buildRuleBucket(entries)
	if bucket.dispatch == nil {
		t.Fatal("expected a string dispatch plan")
	}

	sourceFile := tsparser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/dispatch-attack.ts",
		Path:     "/dispatch-attack.ts",
	}, `foo; bar; function foo() {} object.foo(); eval('x'); const f = function() {}; const g = function named() {};`, core.ScriptKindTS)

	var builder estree.Builder
	if err := builder.WithTree(sourceFile, estree.TreeOptions{SourceType: "module"}, func(tree *estree.Tree) {
		var attack func(estree.NodeID)
		attack = func(id estree.NodeID) {
			full := make([]bool, len(entries))
			optimized := make([]bool, len(entries))
			for index := range entries {
				full[index] = matchesSelector(entries[index].compiled, tree, id, estree.NoNode, nil)
			}

			for cur := bucket.dispatch; cur != nil; cur = cur.next {
				for _, index := range cur.lookup(tree, id) {
					optimized[index] = matchesSelector(entries[index].compiled, tree, id, estree.NoNode, nil)
				}
				for _, index := range cur.fallback {
					optimized[index] = matchesSelector(entries[index].compiled, tree, id, estree.NoNode, nil)
				}
			}
			for index := range entries {
				if optimized[index] != full[index] {
					t.Errorf("node %s selector %q: optimized=%v full=%v", tree.Type(id), entries[index].selector, optimized[index], full[index])
				}
			}
			tree.ForEachChild(id, func(child estree.NodeID) bool {
				attack(child)
				return false
			})
		}
		attack(tree.Root())
	}); err != nil {
		t.Fatal(err)
	}
}

func TestNoRestrictedSyntaxRulePlanCache(t *testing.T) {
	options := []any{
		"Identifier[name='target']",
		"Identifier[name='other']",
	}
	want := cachedRulePlan(options)

	// In-place slice mutation must not return a stale cached plan.
	mutated := []any{"UnknownNodeType"}
	if p := cachedRulePlan(mutated); !p.empty {
		t.Fatal("unknown node type should produce empty plan")
	}
	mutated[0] = "WithStatement"
	if p := cachedRulePlan(mutated); p.empty {
		t.Fatal("in-place mutated slice hit stale empty plan in cache")
	}

	var waitGroup sync.WaitGroup
	for range 32 {
		waitGroup.Go(func() {
			if got := cachedRulePlan(options); got != want {
				t.Errorf("concurrent cache lookup returned plan %p, want %p", got, want)
			}
		})
	}
	waitGroup.Wait()

	for index := range maxCachedRulePlans + 16 {
		cachedRulePlan([]any{"Identifier[name='entry" + strconv.Itoa(index) + "']"})
	}

	rulePlanCache.RLock()
	defer rulePlanCache.RUnlock()
	if len(rulePlanCache.entries) > maxCachedRulePlans {
		t.Fatalf("rule plan cache has %d entries, want at most %d", len(rulePlanCache.entries), maxCachedRulePlans)
	}
	if len(rulePlanCache.order) != len(rulePlanCache.entries) {
		t.Fatalf("cache order has %d entries for %d plans", len(rulePlanCache.order), len(rulePlanCache.entries))
	}
}

// cspell:ignore FFFD blockstatement
func TestNoRestrictedSyntaxAdversarialRegressions(t *testing.T) {
	// Parser depth guard rejects deeply nested selectors cleanly without stack overflow.
	deepNot := strings.Repeat(":not(", 200) + "Identifier" + strings.Repeat(")", 200)
	if _, err := parseSelector(deepNot); err == nil {
		t.Fatal("expected depth limit error for 200-deep nested :not()")
	}

	highSurrogate := ecmascript.StringFromCodeUnits([]uint16{0xD83D})
	lowSurrogate := ecmascript.StringFromCodeUnits([]uint16{0xDE00})

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			// `.parent` is not a valid child field path in class selectors.
			{
				Code:    `const x = 1;`,
				Options: []any{`VariableDeclaration.parent.body`},
			},
			// Regex negation with top-level alternation must not false-positive on second branch.
			{
				Code:    `bar;`,
				Options: []any{`Identifier[name!=/^foo|bar/]`},
			},
			// Deeply nested `:has` completes instantly via memoization.
			{
				Code:    `;`,
				Options: []any{`*` + strings.Repeat(`:has(`, 28) + `WithStatement` + strings.Repeat(`)`, 28)},
			},
			// ES2025 duplicate capture group name in separate alternatives is a valid RegExp (not null).
			{
				Code:    `const r = /(?<k>a)|(?<k>b)/;`,
				Options: []any{`Literal[value=null]`},
			},
			// Indexing a supplementary character must yield its UTF-16 surrogate, never U+FFFD.
			{
				Code:    `const x = '😀';`,
				Options: []any{"Literal[value.0=\"\uFFFD\"]"},
			},
			{
				Code:    `const x = '😀';`,
				Options: []any{"Literal[value.1=\"\uFFFD\"]"},
			},
			// Indexing a pre-existing lone surrogate yields 1 UTF-16 code unit, not 3 replacement bytes.
			{
				Code:    `const x = '\uD83D';`,
				Options: []any{"Literal[value.0=\"\uFFFD\"]"},
			},
			{
				Code:    `const x = '\uD83D';`,
				Options: []any{`Literal[value.1]`},
			},
			// `import(Name)` and `import(null)` set TSImportType.source to undefined.
			{
				Code:    "type Name = 'pkg';\ntype T = import(Name);",
				Options: []any{`TSImportType[source]`},
			},
			{
				Code:    "type D = import(null);",
				Options: []any{`TSImportType[source]`},
			},
			{
				Code:    "type D = import(null);",
				Options: []any{`Literal[value=null]`},
			},
			// Regex matching on WTF-8 lone surrogates must not decode them as U+FFFD.
			{
				Code:    `const x = '😀';`,
				Options: []any{`Literal[value.0=/\uFFFD/]`},
			},
			{
				Code:    `const x = '\uD83D';`,
				Options: []any{`Literal[value=/\uFFFD/]`},
			},
			// Exact-case top-level selector dispatch and `#*` literal identifier (not wildcard).
			{
				Code:    `foo;`,
				Options: []any{`identifier`},
			},
			{
				Code:    `foo;`,
				Options: []any{`#*`},
			},
			{
				Code:    `foo;`,
				Options: []any{`Identifier#*`},
			},
			// Invalid attribute syntax (`type(...)` with `<`/`>`, empty regex class `/[]/`) is rejected.
			{
				Code:    `const x = 1;`,
				Options: []any{`:not([value > type(number)])`},
			},
			{
				Code:    `const x = 'a';`,
				Options: []any{`:not(Literal[value=/[]/])`},
			},
			// Annex B legacy regex literals remain valid RegExp objects (not null).
			{
				Code:    `const r = /\p{Letter}/;`,
				Options: []any{`Literal[value=null]`},
			},
			{
				Code:    `const r = /\01/;`,
				Options: []any{`Literal[value=null]`},
			},
			// Signed zero / signed digits are not canonical array or string indices.
			{
				Code:    `const a = [10, 20];`,
				Options: []any{`ArrayExpression[elements.-0]`},
			},
			{
				Code:    `const a = [10, 20];`,
				Options: []any{`ArrayExpression[elements.+0]`},
			},
			{
				Code:    `const s = "xyz";`,
				Options: []any{`Literal[value.-0="x"]`},
			},
			{
				Code:    `const s = "xyz";`,
				Options: []any{`Literal[value.-00="x"]`},
			},
			// Documented difference: `loc` and `range` are not projected on Tree nodes.
			{
				Code:    `foo;`,
				Options: []any{`Identifier[loc.start.line=1]`},
			},
			{
				Code:    `foo;`,
				Options: []any{`Identifier[range.0=0]`},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid-type `import(Name)` and unsupported JSDoc types must not hide later statements.
			{
				Code:    "type Name = 'pkg';\ntype T = import(Name);\ndebugger;",
				Options: []any{`DebuggerStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 3, Column: 1, EndLine: 3, EndColumn: 10}},
			},
			{
				Code:    "type Name = 'pkg';\ntype T = import(Name);",
				Options: []any{`TSImportType`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 2, Column: 10, EndLine: 2, EndColumn: 22}},
			},
			{
				Code:    "type T = ?number;\ndebugger;",
				Options: []any{`DebuggerStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 10}},
			},
			// Reverse sibling subjects with class pseudo-selectors match ESLint 10 nodeTypes semantics.
			{
				Code:    "debugger;\nfunction f() {}",
				Options: []any{`!DebuggerStatement ~ :declaration`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
				},
			},
			{
				Code:    "debugger;\nfunction f() {}",
				Options: []any{`DebuggerStatement + !:declaration`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
				},
			},
			{
				Code:    "debugger;\nfunction f() {}",
				Options: []any{`!DebuggerStatement ~ :statement`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
				},
			},
			{
				// `:function` narrows nodeTypes to function nodes in ESLint 10, so only FunctionDeclaration matches.
				Code:    "debugger;\nfunction f() {}",
				Options: []any{`!DebuggerStatement ~ :function`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
				},
			},
			// ES2025 duplicate capture groups in separate alternatives expose RegExp properties.
			{
				Code:    `const r = /(?<k>a)|(?<k>b)/;`,
				Options: []any{`Literal[value.source="(?<k>a)|(?<k>b)"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 28}},
			},
			// UTF-16 code-unit indexing preserves high/low surrogates and genuine U+FFFD.
			{
				Code:    `const x = '😀';`,
				Options: []any{`Literal[value.0="` + highSurrogate + `"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `const x = '😀';`,
				Options: []any{`Literal[value.1="` + lowSurrogate + `"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `const x = '\uD83D';`,
				Options: []any{`Literal[value.0="` + highSurrogate + `"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}},
			},
			{
				Code:    "const x = '\uFFFD';",
				Options: []any{"Literal[value.0=\"\uFFFD\"]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 14}},
			},
			// anchoredLiteralPrefix with optional quantifier and alternation.
			{
				Code:    `const u = "http://example.com";`,
				Options: []any{`Literal[value=/^https?:\/\//]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 31}},
			},
			{
				Code:    `useLayoutEffect(fn);`,
				Options: []any{`CallExpression[callee.name=/^useEffect|useLayoutEffect/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}},
			},
			{
				Code:    `a;`,
				Options: []any{`Identifier[name=/^ab?$/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 2}},
			},
			// Compound atoms in arbitrary order.
			{
				Code:    `const x = 1;`,
				Options: []any{`[kind="const"]VariableDeclaration`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    `a.b;`,
				Options: []any{`:not([computed=true])MemberExpression`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo();`,
				Options: []any{`.callee#Identifier`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			// String code-unit indexing (`[name.0="_"]`).
			{
				Code:    `const _a = 1;`,
				Options: []any{`Identifier[name.0="_"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 7, EndLine: 1, EndColumn: 9}},
			},
			// RegExp.lastIndex own property.
			{
				Code:    `const r = /foo/;`,
				Options: []any{`Literal[value.lastIndex=0]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 16}},
			},
			// Invalid same-alternative duplicate capture group yields null value.
			{
				Code:    `const r = /(?<x>a)(?<x>b)/;`,
				Options: []any{`Literal[value=null]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 27}},
			},
			// Multi-segment class selector `.body.declarations` (O(len(path))).
			{
				Code:    `const x = 1;`,
				Options: []any{`VariableDeclarator.body.declarations`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 7, EndLine: 1, EndColumn: 12}},
			},
			// stringAttrDispatch with undefined / null coerced keys.
			{
				Code:    `obj.bar();`,
				Options: []any{`CallExpression[callee.name=undefined]`, `CallExpression[callee.name='eval']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
			},
			{
				Code:    `const f = function() {};`,
				Options: []any{`FunctionExpression[id.name=null]`, `FunctionExpression[id.name='foo']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 24}},
			},
			// Regex matching on WTF-8 lone surrogates and valid astral characters.
			{
				Code:    `const x = '😀';`,
				Options: []any{`Literal[value.0=/^.$/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `const x = '😀';`,
				Options: []any{`Literal[value=/😀/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `const x = '\uD83D';`,
				Options: []any{`Literal[value=/^\uD83D$/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}},
			},
			// Case-insensitive matching still works in non-top-level positions & `:FUNCTION` does not narrow.
			{
				Code:    `function f() { return 1; }`,
				Options: []any{`blockstatement > ReturnStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 25}},
			},
			{
				Code:    `foo;`,
				Options: []any{`:not(:not(identifier))`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    "debugger;\nfunction f() {}",
				Options: []any{`!DebuggerStatement ~ :FUNCTION`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
				},
			},
			// Unquoted slash values & ordering operators in attribute values (esquery 1.7.0).
			{
				Code:    `a / b;`,
				Options: []any{`BinaryExpression[operator=/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{
				Code:    `const x = "//";`,
				Options: []any{`Literal[value=//]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `const x = "/a/";`,
				Options: []any{`Literal[value < /foo/]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 16}},
			},
			// `:first-child` adjacent to a type identifier without whitespace.
			{
				Code:    `foo;`,
				Options: []any{`:first-childExpressionStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}},
			},
			// JS JSDoc `@private` on constructor does not drop constructor key.
			{
				Code:     `class C { /** @private */ constructor() {} }`,
				FileName: "jsdoc-ctor.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []any{`MethodDefinition[key.name='constructor']`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 27, EndLine: 1, EndColumn: 43}},
			},
			// TS template literal preserves CRLF in TemplateElement.value.raw.
			{
				Code:    "const s = `a\r\nb`;",
				Options: []any{`TemplateElement[value.raw="a\r\nb"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 2, EndColumn: 3}},
			},
			// Empty `[]` inside a later attribute/branch backtracks regex parsing cleanly for `[operator=/]`.
			{
				Code:    `a / b; const s = "x] []";`,
				Options: []any{`BinaryExpression[operator=/], Literal[value="x] []"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `"x] []" / b;`,
				Options: []any{`BinaryExpression[operator=/][left.value="x] []"]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
			},
			// Template-literal argument inside `import(...)` degrades locally without hiding later statements.
			{
				Code:    "type T = import(`a${string}b`);\ndebugger;",
				Options: []any{`DebuggerStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 10}},
			},
			// Leading JSDoc `@typedef` in JS does not cancel directive prologue on `"use strict"`.
			{
				Code:     "/** @typedef {number} T */\n\"use strict\";",
				FileName: "jsdoc-directive.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []any{`ExpressionStatement[directive="use strict"]`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 14}},
			},
			// Selector string preserves backslash before raw line terminators.
			{
				Code:    `const s = "a\\\nb";`,
				Options: []any{"Literal[value=\"a\\\nb\"]"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}},
			},
		},
	)
}
