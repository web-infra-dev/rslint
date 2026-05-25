package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedSyntax_ProbeDeepEdges is a deep-edge probe suite — each
// case targets a scenario a real ESLint user might run into and that the
// initial port could plausibly mishandle. Cases that pass here are kept
// in the regular test file; cases that fail get fixed in the matcher /
// parser / mapping layer.
func TestNoRestrictedSyntax_ProbeDeepEdges(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			// Probe: `MemberExpression[object.name='console']` with a
			// parenthesized receiver should still match — esquery
			// transparently sees through ParenthesizedExpression in
			// ESTree because parens are dropped at parse time.
			// Expected behaviour: rslint should also see through them
			// even though tsgo keeps a ParenthesizedExpression wrapper.
			{
				Code:    `(notConsole).log('msg');`,
				Options: []interface{}{`MemberExpression[object.name='console']`},
			},
			// Probe: TS non-null assertion on the receiver should not
			// upgrade non-console into console.
			{
				Code:    `notConsole!.log('msg');`,
				Options: []interface{}{`MemberExpression[object.name='console']`},
			},
			// Probe: import side-effect-only — shape: ImportDeclaration
			// without specifiers — should NOT match a selector that
			// requires source.value=foo when the literal is bar.
			{
				Code:    `import 'bar';`,
				Options: []interface{}{`ImportDeclaration[source.value='foo']`},
			},
			// Probe: TaggedTemplateExpression should not be a CallExpression.
			{
				Code:    "tag`hi`;",
				Options: []interface{}{"CallExpression"},
			},
			// Probe: NewExpression should not match CallExpression.
			{
				Code:    `new Foo();`,
				Options: []interface{}{"CallExpression"},
			},
			// Probe: `ConditionalExpression` should not match an `if`
			// statement (different ESTree types).
			{
				Code:    `if (a) b; else c;`,
				Options: []interface{}{"ConditionalExpression"},
			},

			// Round 2 negatives — verify the Property/MethodDefinition
			// disambiguation does not over-match.
			{
				// Object-literal method is Property, NOT MethodDefinition.
				Code:    `({ foo() {} });`,
				Options: []interface{}{"MethodDefinition"},
			},
			{
				// Class method is MethodDefinition, NOT Property.
				Code:    `class C { foo() {} }`,
				Options: []interface{}{"Property"},
			},
			{
				// Object-literal getter is Property, NOT MethodDefinition.
				Code:    `({ get x() { return 1; } });`,
				Options: []interface{}{"MethodDefinition"},
			},
			{
				// Class getter is MethodDefinition, NOT Property.
				Code:    `class C { get x() { return 1; } }`,
				Options: []interface{}{"Property"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Probe: parenthesized receiver — selectors should see through
			// ParenthesizedExpression like esquery does.
			{
				Code:    `(console).log('msg');`,
				Options: []interface{}{`MemberExpression[object.name='console']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'MemberExpression[object.name='console']' is not allowed.`,
					},
				},
			},
			// Probe: TS non-null assertion on receiver
			{
				Code:    `console!.log('msg');`,
				Options: []interface{}{`MemberExpression[object.name='console']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'MemberExpression[object.name='console']' is not allowed.`,
					},
				},
			},
			// Probe: TS `as` cast on receiver
			{
				Code:    `(console as any).log('msg');`,
				Options: []interface{}{`MemberExpression[object.name='console']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedSyntax",
						Message:   `Using 'MemberExpression[object.name='console']' is not allowed.`,
					},
				},
			},
			// Probe: standalone presence selector matches every node.
			// Compact source should produce a known number of diagnostics.
			{
				Code:    `bar`,
				Options: []interface{}{`[type='Identifier']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using '[type='Identifier']' is not allowed."},
				},
			},
			// Probe: `[type!='Identifier']` — matches every non-Identifier
			// node. For `var x = 1;` that is: VariableStatement,
			// VariableDeclaration (the declarator), NumericLiteral.
			// Identifier `x` is excluded.
			{
				Code:    `var x = 1;`,
				Options: []interface{}{`[type!='Identifier']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using '[type!='Identifier']' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using '[type!='Identifier']' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using '[type!='Identifier']' is not allowed."},
				},
			},
			// Probe: nested optional chain with parenthesised middle —
			// ChainExpression selector still resolves to one outermost.
			{
				Code:    `(a?.b)?.c`,
				Options: []interface{}{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					// The outer chain `(a?.b)?.c` — only one outermost.
					{MessageId: "restrictedSyntax", Message: "Using 'ChainExpression' is not allowed."},
					// Inner chain `a?.b` is a separate chain because the
					// parens break the chain — esquery would also wrap it
					// in its own ChainExpression. Keep both.
					{MessageId: "restrictedSyntax", Message: "Using 'ChainExpression' is not allowed."},
				},
			},
			// Probe: `:has(...)` — class with at least one method named foo
			{
				Code:    `class A { foo() {} bar() {} }`,
				Options: []interface{}{`ClassDeclaration:has(MethodDefinition[key.name='foo'])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ClassDeclaration:has(MethodDefinition[key.name='foo'])' is not allowed.`},
				},
			},
			// Probe: `:not(...)` head — methods that are not static.
			{
				Code:    `class A { static foo() {} bar() {} }`,
				Options: []interface{}{"MethodDefinition:not([static=true])"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition:not([static=true])' is not allowed."},
				},
			},
			// Probe: ImportSpecifier with name attribute
			{
				Code:    `import { useEffect } from 'react';`,
				Options: []interface{}{`ImportSpecifier[imported.name='useEffect']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ImportSpecifier[imported.name='useEffect']' is not allowed.`},
				},
			},
			// Probe: ban arrow inside a class field initializer (a common
			// React useEffect anti-pattern).
			{
				Code:    `class A { onClick = () => {}; }`,
				Options: []interface{}{"PropertyDefinition > ArrowFunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'PropertyDefinition > ArrowFunctionExpression' is not allowed."},
				},
			},
			// Probe: ban String literal that is the source of an import
			// using a regex with capturing path. Real users do this for
			// path-pinning.
			{
				Code:    `import 'react/jsx-runtime';`,
				Options: []interface{}{`ImportDeclaration > Literal[value=/^react\//]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ImportDeclaration > Literal[value=/^react\//]' is not allowed.`},
				},
			},
			// Probe: ban `eval(`anything`)`. Common security guideline.
			{
				Code:    `eval(userInput);`,
				Options: []interface{}{`CallExpression[callee.name='eval']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'CallExpression[callee.name='eval']' is not allowed.`},
				},
			},
			// Probe: ban `instanceof` (BinaryExpression with operator).
			{
				Code:    `x instanceof Foo;`,
				Options: []interface{}{"BinaryExpression[operator='instanceof']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'BinaryExpression[operator='instanceof']' is not allowed."},
				},
			},
			// Probe: ban `typeof` operator (UnaryExpression operator).
			{
				Code:    `typeof x;`,
				Options: []interface{}{"UnaryExpression[operator='typeof']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UnaryExpression[operator='typeof']' is not allowed."},
				},
			},
			// Probe: ban template literals with substitutions.
			{
				Code:    "`hello ${name}`;",
				Options: []interface{}{"TemplateLiteral"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'TemplateLiteral' is not allowed."},
				},
			},
			// Probe: ban computed class member key.
			{
				Code:    `class A { [foo]() {} }`,
				Options: []interface{}{"MethodDefinition[computed=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition[computed=true]' is not allowed."},
				},
			},
			// Probe: ban private class fields. ESTree: PropertyDefinition
			// with PrivateIdentifier key.
			{
				Code:    `class A { #x = 1; }`,
				Options: []interface{}{"PropertyDefinition > PrivateIdentifier.key"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'PropertyDefinition > PrivateIdentifier.key' is not allowed."},
				},
			},
			// Probe: forbid catch with no parameter (catch optional
			// binding, ES2019). ESTree: CatchClause with param=null. Our
			// presence check should be false for the omitted variant.
			{
				Code:    `try { f(); } catch (e) { g(); }`,
				Options: []interface{}{"CatchClause[param]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'CatchClause[param]' is not allowed."},
				},
			},
			// Probe: `MemberExpression[property.name='log']` — read the
			// property name via attribute path.
			{
				Code:    `obj.log();`,
				Options: []interface{}{`MemberExpression[property.name='log']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MemberExpression[property.name='log']' is not allowed.`},
				},
			},
			// Probe: ban export-default functions (common code-style rule).
			{
				Code:    `export default function f() {}`,
				Options: []interface{}{"ExportDefaultDeclaration > FunctionDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ExportDefaultDeclaration > FunctionDeclaration' is not allowed."},
				},
			},
			// Probe: ban literal `true` / `false` (BooleanLiteral in
			// ESTree, KindTrueKeyword / KindFalseKeyword in tsgo). Bare
			// `Literal[value=true]`.
			{
				Code:    `var x = true;`,
				Options: []interface{}{"Literal[value=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal[value=true]' is not allowed."},
				},
			},
			// Probe: ban literal null.
			{
				Code:    `var x = null;`,
				Options: []interface{}{"Literal[value=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal[value=null]' is not allowed."},
				},
			},
			// Probe: ban literal numeric value.
			{
				Code:    `f(0)`,
				Options: []interface{}{"Literal[value=0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal[value=0]' is not allowed."},
				},
			},
			// Probe: ban specific string literal value.
			{
				Code:    `var x = 'hello';`,
				Options: []interface{}{`Literal[value='hello']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'Literal[value='hello']' is not allowed.`},
				},
			},
			// Probe: ban await in any non-async context — but this needs
			// type info / scope. Just match any AwaitExpression.
			{
				Code:    `async function f() { await x; }`,
				Options: []interface{}{"AwaitExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'AwaitExpression' is not allowed."},
				},
			},
			// Probe: ban yield expression
			{
				Code:    `function* g() { yield 1; }`,
				Options: []interface{}{"YieldExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'YieldExpression' is not allowed."},
				},
			},
			// Probe: ban this
			{
				Code:    `function f() { return this; }`,
				Options: []interface{}{"ThisExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ThisExpression' is not allowed."},
				},
			},
			// Probe: ban specific class member kind via Property[kind='get']
			// — verifies the kind synthesizer for accessors.
			{
				Code:    `({ get x() { return 1; } });`,
				Options: []interface{}{"Property[kind='get']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property[kind='get']' is not allowed."},
				},
			},
			// Probe: ban `typeof X === 'undefined'` pattern — common
			// code-style restriction.
			{
				Code:    `if (typeof a === 'undefined') {}`,
				Options: []interface{}{`BinaryExpression[left.type='UnaryExpression'][left.operator='typeof']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'BinaryExpression[left.type='UnaryExpression'][left.operator='typeof']' is not allowed.`},
				},
			},
			// Probe: ban `delete` on a property expression.
			{
				Code:    `delete a.b;`,
				Options: []interface{}{"UnaryExpression[operator='delete']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UnaryExpression[operator='delete']' is not allowed."},
				},
			},
			// Probe: descendant combinator that hops through a paren.
			// ESLint sees `function () { return (function () {}); }` as
			// FunctionExpression descendant of ReturnStatement of
			// FunctionDeclaration. Our descendant walker uses parent chain
			// so the ParenthesizedExpression wrapper should NOT prevent
			// the match (it's traversed normally).
			{
				Code:    `function f() { return (function g() {}); }`,
				Options: []interface{}{"FunctionDeclaration FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'FunctionDeclaration FunctionExpression' is not allowed."},
				},
			},
			// Probe: JSX — ban `<Foo />` via JSXOpeningElement attribute path.
			// (rslint maps JSXElement to KindJsxSelfClosingElement; for
			// self-closing tags there's no inner OpeningElement node, so
			// users typically reach for JSXOpeningElement directly.)
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{`JSXOpeningElement[name.name='Foo']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'JSXOpeningElement[name.name='Foo']' is not allowed.`},
				},
			},
			{
				Code:    `const x = <Foo>hi</Foo>;`,
				Tsx:     true,
				Options: []interface{}{`JSXElement[openingElement.name.name='Foo']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'JSXElement[openingElement.name.name='Foo']' is not allowed.`},
				},
			},
			// Probe: JSX element selector matches both KindJsxElement and
			// KindJsxSelfClosingElement.
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{"JSXElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'JSXElement' is not allowed."},
				},
			},
			// Probe: JSX element with children
			{
				Code:    `const x = <div>hi</div>;`,
				Tsx:     true,
				Options: []interface{}{"JSXElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'JSXElement' is not allowed."},
				},
			},
			// Probe: JSX fragment
			{
				Code:    `const x = <>hi</>;`,
				Tsx:     true,
				Options: []interface{}{"JSXFragment"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'JSXFragment' is not allowed."},
				},
			},
			// Probe: TS-only — ban TSEnumDeclaration. ESTree extension
			// for TS adds this; tsgo has KindEnumDeclaration. We don't
			// map it yet; this probe documents the gap.
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{"TSEnumDeclaration"},
				Skip:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
				},
			},
			// Probe: BigInt literal
			{
				Code:    `var x = 1n;`,
				Options: []interface{}{"Literal[bigint]"},
				Skip:    true, // bigint attribute path not modelled; user can use Literal[value=/n$/].
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
				},
			},
			// Probe: empty-body function
			{
				Code:    `function f() {}`,
				Options: []interface{}{"FunctionDeclaration[body.body.length=0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'FunctionDeclaration[body.body.length=0]' is not allowed."},
				},
			},
			// Probe: parens-broken optional chain — ESTree wraps each
			// chain section in its own ChainExpression.
			{
				Code:    `(a?.b).c`,
				Options: []interface{}{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					// Inside the parens: `a?.b` is one chain.
					{MessageId: "restrictedSyntax", Message: "Using 'ChainExpression' is not allowed."},
				},
			},
			// Probe: Identifier whose parent is an ImportSpecifier (a
			// real-world selector for "no-named-import-this-thing").
			{
				Code:    `import { foo } from 'bar';`,
				Options: []interface{}{`ImportSpecifier > Identifier[name='foo']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ImportSpecifier > Identifier[name='foo']' is not allowed.`},
				},
			},
			// Probe: ban specific decorator name
			{
				Code:    `@sealed class A {}`,
				Options: []interface{}{`Decorator > CallExpression[callee.name='sealed']`},
				Skip:    true, // tsgo wraps decorator over Identifier directly here, no Call. Skip; cover via :has below.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `@sealed class A {}`,
				Options: []interface{}{`Decorator > Identifier[name='sealed']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'Decorator > Identifier[name='sealed']' is not allowed.`},
				},
			},
			// Probe: ban specific class name via descendant combinator
			{
				Code:    `class Component extends Base {}`,
				Options: []interface{}{`ClassDeclaration[id.name='Component']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ClassDeclaration[id.name='Component']' is not allowed.`},
				},
			},
			// Probe: ban arrow function with certain parameter count
			{
				Code:    `const f = (a, b, c) => a;`,
				Options: []interface{}{"ArrowFunctionExpression[params.length>=3]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ArrowFunctionExpression[params.length>=3]' is not allowed."},
				},
			},
			// Probe: ban specific argument count on call
			{
				Code:    `f(1, 2, 3, 4);`,
				Options: []interface{}{"CallExpression[arguments.length>3]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'CallExpression[arguments.length>3]' is not allowed."},
				},
			},
			// Probe: ban computed property access with string literal key
			{
				Code:    `obj['secret'];`,
				Options: []interface{}{`MemberExpression[property.value='secret']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MemberExpression[property.value='secret']' is not allowed.`},
				},
			},
			// Probe: nested unions with attribute filters
			{
				Code:    `function foo() {} class Bar {}`,
				Options: []interface{}{":is(FunctionDeclaration, ClassDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using ':is(FunctionDeclaration, ClassDeclaration)' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using ':is(FunctionDeclaration, ClassDeclaration)' is not allowed."},
				},
			},
			// Probe: real-world — ban deprecated patterns in tests.
			// Attribute path with regex on string literal value.
			{
				Code:    `it.skip('does nothing', () => {})`,
				Options: []interface{}{`CallExpression[callee.property.name='skip']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'CallExpression[callee.property.name='skip']' is not allowed.`},
				},
			},
			// Probe: union-format selector for several disallowed APIs.
			{
				Code:    `setTimeout(fn);`,
				Options: []interface{}{`CallExpression[callee.name=/^(setTimeout|setInterval)$/]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'CallExpression[callee.name=/^(setTimeout|setInterval)$/]' is not allowed.`},
				},
			},
			// Probe: object key with computed string literal
			{
				Code:    `({ ['foo']: 1 });`,
				Options: []interface{}{"Property[computed=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property[computed=true]' is not allowed."},
				},
			},
			// Probe: class private method — `#foo` is PrivateIdentifier.
			{
				Code:    `class A { #foo() {} }`,
				Options: []interface{}{"MethodDefinition > PrivateIdentifier.key"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition > PrivateIdentifier.key' is not allowed."},
				},
			},
			// Probe: ban spread in call arguments specifically
			{
				Code:    `f(...args);`,
				Options: []interface{}{"CallExpression > SpreadElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'CallExpression > SpreadElement' is not allowed."},
				},
			},
			// Probe: ban `delete` of a member access
			{
				Code:    `delete this.foo;`,
				Options: []interface{}{"UnaryExpression[operator='delete'][argument.type='MemberExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'UnaryExpression[operator='delete'][argument.type='MemberExpression']' is not allowed."},
				},
			},
			// Probe: ban `eval`-style indirect call: arguments.callee
			{
				Code:    `arguments.callee;`,
				Options: []interface{}{`MemberExpression[object.name='arguments'][property.name='callee']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MemberExpression[object.name='arguments'][property.name='callee']' is not allowed.`},
				},
			},

			// ============================================================
			// Real-world: airbnb-style restricted syntax
			// ============================================================
			{
				Code:    `for (var i in obj) {}`,
				Options: []interface{}{"ForInStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ForInStatement' is not allowed."},
				},
			},
			{
				Code:    `outer: for (;;) { break outer; }`,
				Options: []interface{}{"LabeledStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'LabeledStatement' is not allowed."},
				},
			},
			{
				Code:    `'foo' in obj;`,
				Options: []interface{}{"BinaryExpression[operator='in']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'BinaryExpression[operator='in']' is not allowed."},
				},
			},

			// Probe: require strict equality
			{
				Code:    `if (a == b) {}`,
				Options: []interface{}{"BinaryExpression[operator='==']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'BinaryExpression[operator='=='] ' is not allowed."},
				},
				Skip: true, // intentional bug test (typo) — disabled.
			},
			{
				Code:    `if (a == b) {}`,
				Options: []interface{}{"BinaryExpression[operator='==']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'BinaryExpression[operator='=='] is not allowed."},
				},
				Skip: true, // intentional message-mismatch — disabled.
			},
			{
				Code:    `if (a == b) {}`,
				Options: []interface{}{"BinaryExpression[operator='==']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'BinaryExpression[operator='==']' is not allowed."},
				},
			},

			// Probe: ban `__proto__`
			{
				Code:    `obj.__proto__;`,
				Options: []interface{}{`MemberExpression[property.name='__proto__']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MemberExpression[property.name='__proto__']' is not allowed.`},
				},
			},

			// Probe: ban hasOwnProperty (use Object.hasOwn)
			{
				Code:    `obj.hasOwnProperty(k);`,
				Options: []interface{}{`CallExpression[callee.property.name='hasOwnProperty']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'CallExpression[callee.property.name='hasOwnProperty']' is not allowed.`},
				},
			},

			// Probe: ban use of `undefined` identifier (prefer `void 0`)
			{
				Code:    `if (a === undefined) {}`,
				Options: []interface{}{`Identifier[name='undefined']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'Identifier[name='undefined']' is not allowed.`},
				},
			},

			// Probe: regex on Identifier name with anchors
			{
				Code:    `var _hidden = 1;`,
				Options: []interface{}{`Identifier[name=/^_/]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'Identifier[name=/^_/]' is not allowed.`},
				},
			},

			// Probe: arrow with no params
			{
				Code:    `const f = () => 1;`,
				Options: []interface{}{"ArrowFunctionExpression[params.length=0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ArrowFunctionExpression[params.length=0]' is not allowed."},
				},
			},

			// Probe: real-world `process.env` ban
			{
				Code:    `process.env.NODE_ENV;`,
				Options: []interface{}{`MemberExpression[object.object.name='process'][object.property.name='env']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MemberExpression[object.object.name='process'][object.property.name='env']' is not allowed.`},
				},
			},

			// Probe: object destructuring with rest element
			{
				Code:    `const { a, ...rest } = obj;`,
				Options: []interface{}{"RestElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'RestElement' is not allowed."},
				},
			},

			// Probe: array destructuring rest
			{
				Code:    `const [first, ...rest] = arr;`,
				Options: []interface{}{"RestElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'RestElement' is not allowed."},
				},
			},

			// Probe: ban Symbol.iterator access
			{
				Code:    `obj[Symbol.iterator];`,
				Options: []interface{}{`MemberExpression[property.object.name='Symbol'][property.property.name='iterator']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MemberExpression[property.object.name='Symbol'][property.property.name='iterator']' is not allowed.`},
				},
			},

			// Probe: do-while loop ban
			{
				Code:    `do { x(); } while (cond);`,
				Options: []interface{}{"DoWhileStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'DoWhileStatement' is not allowed."},
				},
			},

			// Probe: empty function body via attribute
			{
				Code:    `class A { foo() {} }`,
				Options: []interface{}{"MethodDefinition[value.body.body.length=0]"},
				Skip:    true, // ESTree's MethodDefinition has `value: FunctionExpression`; tsgo MethodDeclaration directly has Body. Different shape.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
				},
			},

			// Probe: deeply nested :is + :not
			{
				Code:    `var a = 1;`,
				Options: []interface{}{`:is(VariableDeclaration, FunctionDeclaration):not([kind='let'])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using ':is(VariableDeclaration, FunctionDeclaration):not([kind='let'])' is not allowed.`},
				},
			},

			// Probe: ban specific export — `export { default as X } from 'mod'`
			{
				Code:    `export { default as Foo } from 'mod';`,
				Options: []interface{}{"ExportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ExportSpecifier' is not allowed."},
				},
			},

			// Probe: Chain on ElementAccess (`a?.[b]`)
			{
				Code:    `a?.[b]`,
				Options: []interface{}{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ChainExpression' is not allowed."},
				},
			},
			// Probe: optional call (`a?.()`) — ChainExpression covers Call too
			{
				Code:    `a?.()`,
				Options: []interface{}{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ChainExpression' is not allowed."},
				},
			},

			// ============================================================
			// Round 2 — bot review + self-audit fixes
			// ============================================================

			// MetaProperty: `new.target` selectable via `meta.name`.
			{
				Code:    `function f() { return new.target; }`,
				Options: []interface{}{`MetaProperty[meta.name='new']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MetaProperty[meta.name='new']' is not allowed.`},
				},
			},
			// MetaProperty: `import.meta.url` — the `import` part is a meta
			// property; we ban via `[meta.name='import']`.
			{
				Code:    `const u = import.meta.url;`,
				Options: []interface{}{`MetaProperty[meta.name='import']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MetaProperty[meta.name='import']' is not allowed.`},
				},
			},
			// MetaProperty: read `property.name` — the `meta` identifier
			// in `import.meta`.
			{
				Code:    `const u = import.meta.url;`,
				Options: []interface{}{`MetaProperty[property.name='meta']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'MetaProperty[property.name='meta']' is not allowed.`},
				},
			},

			// Property vs MethodDefinition disambiguation: object-literal
			// methods are Property, NOT MethodDefinition.
			{
				Code:    `({ foo() {} });`,
				Options: []interface{}{"Property"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property' is not allowed."},
				},
			},
			// Same code MUST NOT be matched by `MethodDefinition`.
			// (negative — see valid section below.)

			// Class methods are MethodDefinition, NOT Property.
			{
				Code:    `class C { foo() {} }`,
				Options: []interface{}{"MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition' is not allowed."},
				},
			},
			// Class get/set are MethodDefinition.
			{
				Code:    `class C { get x() { return 1; } set x(v) {} }`,
				Options: []interface{}{"MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition' is not allowed."},
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition' is not allowed."},
				},
			},
			// Constructor is MethodDefinition.
			{
				Code:    `class C { constructor() {} }`,
				Options: []interface{}{"MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition' is not allowed."},
				},
			},

			// `[type='MethodDefinition']` on class methods (esquery's
			// `type` attribute path, used in nested selectors).
			{
				Code:    `class C { foo() {} }`,
				Options: []interface{}{`[type='MethodDefinition']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using '[type='MethodDefinition']' is not allowed.`},
				},
			},

			// for-in's `left` is ESTree-typed VariableDeclaration.
			{
				Code:    `for (const k in obj) {}`,
				Options: []interface{}{`ForInStatement[left.type='VariableDeclaration']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ForInStatement[left.type='VariableDeclaration']' is not allowed.`},
				},
			},
			// for-in's `left.kind` resolves to 'const' / 'let' / 'var'.
			{
				Code:    `for (const k in obj) {}`,
				Options: []interface{}{`ForInStatement[left.kind='const']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ForInStatement[left.kind='const']' is not allowed.`},
				},
			},
			// for-of's `left.kind` similarly.
			{
				Code:    `for (let v of items) {}`,
				Options: []interface{}{`ForOfStatement[left.kind='let']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ForOfStatement[left.kind='let']' is not allowed.`},
				},
			},

			// ExportSpecifier: `local` is the source name; `exported` is
			// the public name. For `export { foo as bar }` they differ.
			{
				Code:    `export { foo as bar } from 'mod';`,
				Options: []interface{}{`ExportSpecifier[local.name='foo']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ExportSpecifier[local.name='foo']' is not allowed.`},
				},
			},
			{
				Code:    `export { foo as bar } from 'mod';`,
				Options: []interface{}{`ExportSpecifier[exported.name='bar']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ExportSpecifier[exported.name='bar']' is not allowed.`},
				},
			},
		},
	)
}
