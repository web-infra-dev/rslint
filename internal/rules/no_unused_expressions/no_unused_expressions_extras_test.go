// TestNoUnusedExpressionsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
package no_unused_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedExpressionsExtras(t *testing.T) {
	allowShortCircuit := map[string]interface{}{"allowShortCircuit": true}
	allowShortCircuitAndTernary := map[string]interface{}{"allowShortCircuit": true, "allowTernary": true}
	allowShortCircuitArray := []interface{}{map[string]interface{}{"allowShortCircuit": true}}
	allowTernaryArray := []interface{}{map[string]interface{}{"allowTernary": true}}
	allowTaggedTemplates := map[string]interface{}{"allowTaggedTemplates": true}
	enforceForJSX := map[string]interface{}{"enforceForJSX": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedExpressionsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized side-effecting expression ----
			{Code: "((foo()));"},
			// ---- Dimension 4: parenthesized assignment expression ----
			{Code: "((foo = bar));"},
			// ---- Dimension 4: nested TS wrappers around a side-effecting call ----
			{Code: "((foo() as any)!);"},
			// ---- Dimension 4: TS `satisfies` wrapper ----
			// Locks in upstream Checker default: the core rule does not list
			// TSSatisfiesExpression, so it stays in the unknown/allowed bucket.
			{Code: "foo satisfies string;"},
			// ---- Dimension 4: optional-chain call is side-effecting ----
			{Code: "foo?.();"},
			// ---- Dimension 4: declaration/container forms are not expression statements ----
			{Code: "class C { field = foo; static other = bar; }"},
			// ---- Dimension 4: graceful degradation - rest binding pattern ----
			{Code: "const { ...rest } = obj;"},
			// ---- Dimension 4: graceful degradation - overload and declare body-absent forms ----
			{Code: "declare class C { overload(value: string): void; overload(value: number): void; }"},
			// ---- Real-user: eslint/eslint#7632 tagged template APIs are allowed when configured ----
			{Code: "injectGlobal`body { color: red; }`;", Options: allowTaggedTemplates},
			// ---- Real-user: eslint/eslint#12822 optional method call is side-effecting ----
			{Code: "foo?.bar();"},
			// ---- Real-user: eslint/eslint#13666 nullish coalescing with call on the right ----
			{Code: "foo ?? list.push(foo);", Options: allowShortCircuit},
			// Locks in upstream LogicalExpression arm: array-wrapped options use the same JSON path as JS tests.
			{Code: "foo && foo();", Options: allowShortCircuitArray},
			// Locks in upstream ConditionalExpression arm: array-wrapped allowTernary accepts both side-effecting branches.
			{Code: "foo ? bar() : baz();", Options: allowTernaryArray},
			// Locks in recursive unwrapping under allowed short-circuit right-hand sides.
			{Code: "foo && ((bar = baz) as any);", Options: allowShortCircuit},
			// Locks in recursive unwrapping under allowed ternary branches.
			{Code: "foo ? (bar() as any) : (baz()!);", Options: allowTernaryArray},
			// Locks in nested short-circuit + ternary recursion where every terminal branch has side effects.
			{Code: "foo && (bar ? baz() : qux());", Options: allowShortCircuitAndTernary},
			// Locks in upstream UnaryExpression branch: void/delete are accepted side effects.
			{Code: "foo ? void bar() : delete baz.qux;", Options: allowTernaryArray},
			// Locks in upstream LogicalExpression recursion through a void expression.
			{Code: "ready && void task();", Options: allowShortCircuit},
			// Locks in upstream UnaryExpression/update distinction: prefix update has side effects.
			{Code: "++i;"},
			// Locks in upstream AssignmentExpression unknown/allowed branch.
			{Code: "foo += 1;"},
			// Locks in upstream AssignmentExpression branch for logical assignment.
			{Code: "foo ||= bar;"},
			// Locks in the other logical assignment operators covered by ast.IsAssignmentOperator.
			{Code: "foo &&= bar;"},
			{Code: "foo ??= bar;"},
			// Locks in nested TS wrappers around satisfies: ESLint leaves TSSatisfiesExpression unknown/allowed.
			{Code: "(foo satisfies string) as any;"},
			// ---- Real-user: optional callback invocation is side-effecting ----
			{Code: "onDone?.(result);"},
			// ---- Real-user: styled-components tagged template APIs are allowed when configured ----
			{Code: "styled.div`color: red;`;", Options: allowTaggedTemplates},
			// N/A: autofix boundaries - this rule does not provide fixes.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized unused expression ----
			invalidUnusedExpression("((foo));", "((foo));"),
			// ---- Dimension 4: TS non-null assertion wrapper ----
			invalidUnusedExpression("foo!;", "foo!;"),
			// ---- Dimension 4: TS `as` type-expression wrapper ----
			invalidUnusedExpression("(foo as any);", "(foo as any);"),
			// ---- Dimension 4: TS type assertion wrapper ----
			invalidUnusedExpression("<any>foo;", "<any>foo;"),
			// ---- Dimension 4: nested TS wrappers around an unused expression ----
			invalidUnusedExpression("((foo as any)!);", "((foo as any)!);"),
			// ---- Dimension 4: type assertion around comma expression remains unused ----
			invalidUnusedExpression("<any>(foo(), bar);", "<any>(foo(), bar);"),
			// ---- Dimension 4: optional-chain member expression ----
			invalidUnusedExpression("foo?.bar;", "foo?.bar;"),
			// ---- Dimension 4: optional-chain member through nested TS wrappers ----
			invalidUnusedExpression("((foo?.bar as any)!);", "((foo?.bar as any)!);"),
			// ---- Dimension 4: element access string-literal key ----
			invalidUnusedExpression("foo['bar'];", "foo['bar'];"),
			// ---- Dimension 4: element access template-literal key ----
			invalidUnusedExpression("foo[`bar`];", "foo[`bar`];"),
			// ---- Dimension 4: element access numeric-literal key ----
			invalidUnusedExpression("foo[0];", "foo[0];"),
			// ---- Dimension 4: element access symbol key ----
			invalidUnusedExpression("foo[Symbol.iterator];", "foo[Symbol.iterator];"),
			// ---- Dimension 4: private identifier member expression ----
			invalidUnusedExpression("class C { #x; m() { this.#x; } }", "this.#x;"),
			// ---- Dimension 4: declaration/container form - class expression statement ----
			invalidUnusedExpression("(class {});", "(class {});"),
			// ---- Dimension 4: declaration/container form - function expression statement ----
			invalidUnusedExpression("(function () {});", "(function () {});"),
			// ---- Dimension 4: same-kind nesting/traversal boundary ----
			invalidUnusedExpression("function outer() { function inner() { foo; } }", "foo;"),
			// ---- Dimension 4: non-function block has no directive prologue ----
			invalidUnusedExpression("if (foo) { 'bar'; }", "'bar';"),
			// ---- Dimension 4: parenthesized string literal is not a directive prologue ----
			invalidUnusedExpression("(\"use strict\");", "(\"use strict\");"),
			// ---- Dimension 4: nested non-function block has no directive prologue after a real directive ----
			invalidUnusedExpression("function outer() { 'directive'; { 'not directive'; } }", "'not directive';"),
			// ---- Dimension 4: graceful degradation - object spread inside object literal ----
			invalidUnusedExpression("({ ...obj });", "({ ...obj });"),
			// ---- Dimension 4: graceful degradation - array spread inside array literal ----
			invalidUnusedExpression("[...items];", "[...items];"),
			// ---- Dimension 4: JSX expression statement with enforceForJSX through TS `as` wrapper ----
			invalidUnusedExpressionWithOptionsTsx("<div /> as any;", "<div /> as any;", enforceForJSX),
			// ---- Dimension 4: JSX fragment through TS wrapper with enforceForJSX ----
			invalidUnusedExpressionWithOptionsTsx("<></> as any;", "<></> as any;", enforceForJSX),
			// ---- Real-user: eslint/eslint#2102 / #14213 Chai getter chain remains an unused member expression ----
			invalidUnusedExpression("expect(value).to.be.true;", "expect(value).to.be.true;"),
			// ---- Real-user: Chai should-style property assertion remains an unused member expression ----
			invalidUnusedExpression("value.should.be.ok;", "value.should.be.ok;"),
			// Locks in upstream Checker.isDisallowed default path: call expressions are allowed, but member expressions are not.
			invalidUnusedExpression("expect(value).to.be;", "expect(value).to.be;"),
			// Locks in upstream LogicalExpression arm 1: no allowShortCircuit means always disallowed.
			invalidUnusedExpression("foo && bar();", "foo && bar();"),
			// Locks in upstream LogicalExpression arm 2: allowShortCircuit recurses into the right side.
			invalidUnusedExpressionWithOptions("foo && bar;", "foo && bar;", allowShortCircuit),
			// Locks in upstream LogicalExpression arm 2 for nullish coalescing.
			invalidUnusedExpressionWithOptions("foo ?? bar;", "foo ?? bar;", allowShortCircuit),
			// Locks in recursive TS wrapper unwrapping inside allowed short-circuit right-hand sides.
			invalidUnusedExpressionWithOptions("foo && ((bar as any)!);", "foo && ((bar as any)!);", allowShortCircuit),
			// Locks in upstream SequenceExpression branch under allowed short-circuit recursion.
			invalidUnusedExpressionWithOptions("foo && (bar(), baz());", "foo && (bar(), baz());", allowShortCircuit),
			// Locks in upstream ConditionalExpression arm 1: no allowTernary means always disallowed.
			invalidUnusedExpression("foo ? bar() : baz();", "foo ? bar() : baz();"),
			// Locks in upstream ConditionalExpression arm 2: consequent disallowed.
			invalidUnusedExpressionWithOptions("foo ? bar : baz();", "foo ? bar : baz();", map[string]interface{}{"allowTernary": true}),
			// Locks in upstream ConditionalExpression arm 2: alternate disallowed.
			invalidUnusedExpressionWithOptions("foo ? bar() : baz;", "foo ? bar() : baz;", map[string]interface{}{"allowTernary": true}),
			// Locks in nested short-circuit + ternary recursion when a terminal branch is unused.
			invalidUnusedExpressionWithOptions("foo && (bar ? baz() : qux);", "foo && (bar ? baz() : qux);", allowShortCircuitAndTernary),
			// Locks in ternary recursion through a valid short-circuit branch and an invalid alternate.
			invalidUnusedExpressionWithOptions("foo ? (bar && baz()) : qux;", "foo ? (bar && baz()) : qux;", allowShortCircuitAndTernary),
			// Locks in upstream SequenceExpression branch: comma expressions are disallowed even when the right side calls.
			invalidUnusedExpression("foo(), bar();", "foo(), bar();"),
			// Locks in upstream BinaryExpression branch for `in`.
			invalidUnusedExpression("\"key\" in object;", "\"key\" in object;"),
			// Locks in upstream BinaryExpression branch for `instanceof`.
			invalidUnusedExpression("value instanceof Type;", "value instanceof Type;"),
			// Locks in upstream TaggedTemplateExpression branch: disallowed by default.
			invalidUnusedExpression("injectGlobal`body { color: red; }`;", "injectGlobal`body { color: red; }`;"),
			// Locks in real-user styled-components tagged templates when allowTaggedTemplates is false.
			invalidUnusedExpression("styled.div`color: red;`;", "styled.div`color: red;`;"),
			// Locks in upstream UnaryExpression branch: typeof is disallowed.
			invalidUnusedExpression("typeof foo;", "typeof foo;"),
			// Locks in upstream UnaryExpression branch for other prefix operators without side effects.
			invalidUnusedExpression("~flags;", "~flags;"),
			// Locks in upstream MetaProperty branch.
			invalidUnusedExpression("function f() { new.target; }", "new.target;"),
			// Locks in upstream ThisExpression branch.
			invalidUnusedExpression("function f() { this; }", "this;"),
		},
	)
}
