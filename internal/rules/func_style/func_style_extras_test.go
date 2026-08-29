// TestFuncStyleExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension walk notes for func-style:
//   - Dimension 3 (autofix boundaries): N/A — the rule has no autofix or
//     suggestions.
//   - Dimension 4 access/key forms: N/A — the rule never inspects a property
//     or element access; it only classifies function-like declarations and
//     the variable declarator they may be assigned to.
package func_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFuncStyleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncStyleRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS type-expression wrappers on the assigned
			// value (`(X) as T`, `X satisfies T`, `X!`) put a node other than
			// VariableDeclaration directly above the function, so upstream's
			// `node.parent.type === "VariableDeclarator"` check — and this
			// port's matching Kind check — never matches; the function stays
			// unchecked regardless of style ----
			{Code: "var foo = (function(){}) as Fn;", Options: []any{"declaration"}},
			{Code: "var foo = function(){} satisfies Fn;", Options: []any{"declaration"}},
			{Code: "var foo = function(){}!;", Options: []any{"declaration"}},
			{Code: "var foo = (() => {}) as Fn;", Options: []any{"declaration"}},

			// ---- Dimension 4: a function value reached through a
			// destructuring pattern (object/array literal on the RHS, not the
			// declarator's own initializer) is never the declarator's direct
			// child, so it's never declaration-checked ----
			{Code: "var {foo} = {foo: function(){}};", Options: []any{"declaration"}},
			{Code: "var [foo] = [function(){}];", Options: []any{"declaration"}},

			// ---- Dimension 4 declaration/container forms: a class field
			// initialized with a function/arrow value is a PropertyDeclaration
			// initializer, never a VariableDeclarator — this rule (unlike
			// func-names) never inspects class bodies at all ----
			{Code: "class Foo { bar = () => {} }", Options: []any{"declaration"}},
			{Code: "class Foo { bar = function(){} }", Options: []any{"declaration"}},

			// ---- Graceful degradation: an ambient `declare function` (and,
			// equally, each individual overload signature) has a nil Body in
			// tsgo. ESLint parses these as TSDeclareFunction, a kind with no
			// listener attached, so upstream never visits them; the port
			// mirrors that by returning before any check once Body is nil ----
			{Code: "declare function foo(): void;", Options: []any{"expression"}},
			{Code: "declare function foo(): void;", Options: []any{"declaration"}},

			// ---- Locks in isOverloadedFunction's container branch: an
			// overload set declared inside a TS namespace body (ModuleBlock)
			// is found the same way as one at the top level (SourceFile) ----
			{Code: `
namespace NS {
  export function test(a: string): string;
  export function test(a: number): number;
  export function test(a: unknown) { return a; }
}
`, Options: []any{"expression"}},

			// ---- Locks in the FunctionDeclaration branch that checks
			// isDefaultExport before ever calling isOverloadedFunction: a
			// default-exported overload set is exempt for being a default
			// export, independent of the overload-detection branch ----
			{Code: `
export default function test(a: string): string;
export default function test(a: number): number;
export default function test(a: unknown) { return a; }
`, Options: []any{"expression"}},

			// ---- Real-user: a typed React-style functional component,
			// exported and assigned to a const with a generic type annotation
			// — the common "allowTypeAnnotation" use case in real codebases ----
			{Code: "export const Button: React.FC<Props> = (props) => { return null; };", Options: []any{"declaration", map[string]any{"allowTypeAnnotation": true}}},

			// ---- JSX: a `this` expression in a JSX child is traversed in the
			// enclosing arrow's frame, so ESLint preserves the arrow rather than
			// requiring a declaration. A nested callback's `this` has its own
			// frame and does not exempt the component arrow (tested below) ----
			{Code: "const Component = () => <div>{this.value}</div>;", Tsx: true, Options: []any{"declaration"}},

			// ---- JSDoc's transparent type cast does not change an arrow's
			// lexical `this` behavior: the cast is ignored, but the arrow is
			// still exempt because its body references `this` ----
			{Code: "const foo = /** @type {Function} */ (() => this.value);", FileName: "jsdoc-this.js", TSConfig: "tsconfig.allow-js.json", Options: []any{"declaration"}},

			// ---- Dimension 2 nesting: unlike a method/getter/setter/
			// constructor, a class static block is its own dedicated ESTree
			// node (StaticBlock), never a FunctionExpression — so upstream's
			// FunctionExpression listener never fires for it either, and its
			// `this` usage is expected to leak into the enclosing arrow's
			// frame exactly as it does here. Locks in that this port does
			// *not* push a frame for ast.KindClassStaticBlockDeclaration ----
			{Code: "var foo = () => { class C { static { this.x = 1; } } };", Options: []any{"declaration"}},

			// ---- tsgo represents `this` inside a type query as a plain
			// Identifier, while typescript-eslint exposes a ThisExpression. It
			// still marks the enclosing arrow as depending on `this` upstream ----
			{Code: "const foo = (): typeof this => ({});", Options: []any{"declaration"}},

			// ---- Dimension 2 nesting: a member's decorators and computed name
			// live inside the member node in tsgo, but hang off
			// MethodDefinition/Property — outside the FunctionExpression value
			// — in ESTree, so a `this`/`super` written there belongs to the
			// enclosing arrow's frame and keeps the arrow from being reported ----
			{Code: "class A { m() { var foo = () => ({ [this.k]() {} }); } }", Options: []any{"declaration"}},
			{Code: "class A extends B { m() { var foo = () => ({ [super.k]() {} }); } }", Options: []any{"declaration"}},
			{Code: "var foo = () => class { [this.k]() {} };", Options: []any{"declaration"}},
			{Code: "var foo = () => ({ get [this.k]() { return 1; } });", Options: []any{"declaration"}},
			{Code: "class A { m() { var foo = () => class { @this.dec n() {} }; } }", Options: []any{"declaration"}},

			// ---- A class decorator is not member metadata: ESTree keeps it on
			// the ClassDeclaration, which is not a frame either, so it stays
			// attributed to the enclosing arrow exactly as tsgo has it ----
			{Code: "var foo = () => { @this.dec class C {} };", Options: []any{"declaration"}},

			// ---- Locks in isOverloadedFunction's wrapper matching: upstream
			// pairs an exported implementation only with `export`ed signatures,
			// so a set whose export modifiers agree stays exempt ----
			{Code: "export function foo(a: string): void;\nexport function foo(a: any): void {}", Options: []any{"expression"}},
			{Code: "function foo(a: string): void;\nfunction foo(a: any): void {}", Options: []any{"expression"}},
			{Code: "export declare function foo(a: string): void;\nexport function foo(a: any): void {}", Options: []any{"expression"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: `(X).y`-style single/multi-level parenthesized
			// wrapper — tsgo represents each level as an explicit
			// ParenthesizedExpression, so the declarator has to be found by
			// walking back up through ast.WalkUpParenthesizedExpressions ----
			{
				Code:    "var foo = (function(){});",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "var foo = ((function(){}));",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "var foo = (() => {});",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "var foo = ((() => {}));",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- In checked JavaScript, tsgo materializes a JSDoc `@type`
			// cast as an AsExpression around the initializer. ESTree treats that
			// cast as transparent, so the arrow remains directly assigned to the
			// VariableDeclarator and must be declaration-checked ----
			{
				Code:     "const foo = /** @type {Function} */ (() => 1);",
				FileName: "jsdoc-cast.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration"},
				},
			},
			{
				Code:     "const foo = /** @type {Function} */ (function () { return 1; });",
				FileName: "jsdoc-function-cast.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []any{"declaration"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "declaration"}},
			},
			{
				Code:     "const foo = /** @type {Function} */ () => 1;",
				FileName: "jsdoc-direct-cast.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []any{"declaration"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "declaration"}},
			},
			// ---- A JSDoc type attached to the declaration is a comment in
			// ESTree, not a TypeScript type annotation. It is therefore still
			// checked even with allowTypeAnnotation enabled ----
			{
				Code:     "/** @type {() => number} */ const foo = () => 1;",
				FileName: "jsdoc-declaration.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  []any{"declaration", map[string]any{"allowTypeAnnotation": true}},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "declaration"}},
			},

			// ---- Dimension 2 nesting: a `for` loop's own VariableDeclarator
			// is still a VariableDeclaration whose parent is the loop, not a
			// VariableStatement — style-checked the same as a top-level var,
			// even though rewriting it as a declaration wouldn't parse ----
			{
				Code:    "for (var i = function(){};;) {}",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 10, EndLine: 1, EndColumn: 26},
				},
			},

			// ---- Locks in isNamedExportedDeclarator: a separate `export
			// {foo}` specifier clause never sets an export modifier on the
			// original `var foo = ...` statement, so upstream's
			// ExportNamedDeclaration-ancestor check (and this port's
			// VariableStatement-modifier check) treats it as a plain,
			// non-exported declarator — the "ignore" override for named
			// exports doesn't apply, and the base style still fires ----
			{
				Code:    "var foo = function(){}; export { foo };",
				Options: []any{"declaration", map[string]any{"overrides": map[string]any{"namedExports": "ignore"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Locks in upstream isOverloadedFunction's name-matching
			// requirement: three differently-named function declarations in a
			// row look structurally like an overload set, but only a
			// same-named, body-less sibling counts — the last one (the only
			// one with a body) has no such sibling, so it's reported ----
			{
				Code:    "\nfunction test1(a: string): string;\nfunction test2(a: number): number;\nfunction test3(a: unknown) {\n  return a;\n}\n",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 4, Column: 1, EndLine: 6, EndColumn: 2},
				},
			},

			// ---- Locks in isOverloadedFunction's Block-container branch: an
			// overload set nested inside another function's body is found via
			// the generic CanHaveStatements()/Statements() path, same as a
			// top-level SourceFile — but the enclosing `outer` itself is still
			// independently checked and, having no overload siblings of its
			// own, is reported ----
			{
				Code:    "\nfunction outer() {\n  function test(a: string): string;\n  function test(a: number): number;\n  function test(a: unknown) { return a; }\n}\n",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 2, Column: 1, EndLine: 6, EndColumn: 2},
				},
			},

			// ---- Locks in isOverloadedFunction's switch-case cross-clause
			// search returning false on a name mismatch: signatures split
			// across a `case`/`break`/`case` sequence are still searched
			// across every CaseClause of the enclosing CaseBlock, but a
			// differently-named implementation still finds no match ----
			{
				Code:    "\nswitch ($0) {\n  case $1:\n  function test1(a: string): string;\n  break;\n  case $2:\n  function test2(a: unknown) {\n  return a;\n  }\n}\n",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 7, Column: 3, EndLine: 9, EndColumn: 4},
				},
			},

			// ---- Real-user: the same typed React-style component as the
			// valid case above, but without allowTypeAnnotation — the type
			// annotation on its own doesn't exempt it ----
			{
				Code:    "export const Button: React.FC<Props> = (props) => { return null; };",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 14, EndLine: 1, EndColumn: 67},
				},
			},
			{
				Code:    "const Component = () => <div />;",
				Tsx:     true,
				Options: []any{"declaration"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "declaration"}},
			},
			{
				Code:    "const Component = () => <div onClick={() => this.handle()} />;",
				Tsx:     true,
				Options: []any{"declaration"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "declaration"}},
			},
			{
				Code:    "function Component() { return <div />; }",
				Tsx:     true,
				Options: []any{"expression"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expression"}},
			},

			// ---- Dimension 2 nesting: ESTree has no MethodDefinition/Property-
			// method node kind of its own — a class or object method's body is
			// a FunctionExpression value, so upstream's FunctionExpression
			// listener fires for it too and isolates its `this`/`super` from
			// an enclosing arrow. tsgo parses each of these as its own
			// distinct kind (never KindFunctionExpression), so the port has
			// to push/pop a stack frame for MethodDeclaration, GetAccessor,
			// SetAccessor, and Constructor as well, or a nested method's own
			// `this` would wrongly count as the enclosing arrow's ----
			{
				Code:    "var foo = () => { class C { bar(){ this; } } };",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:    "var foo = () => ({ get bar() { return this._bar; } });",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code:    "var foo = () => ({ set bar(v) { this._bar = v; } });",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code:    "var foo = () => (class { constructor(){ this.x = 1; } });",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 57},
				},
			},

			// ---- Only the innermost member's frame is hidden: the `this` here
			// belongs to the arrow inside the computed name, and the outer
			// arrow is still reported ----
			{
				Code:    "var foo = () => ({ [(() => this.x)()]() {} });",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 46},
				},
			},
			// ---- ... and a computed name nested one member deeper is hidden
			// from that member's frame, not from the arrow's ----
			{
				Code:    "var foo = () => class { m() { return class { [this.k]() {} }; } };",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 66},
				},
			},
			// ---- A parameter decorator hangs off the parameter, which ESTree
			// keeps inside the FunctionExpression value, so it stays within
			// the member's frame and the arrow is still reported ----
			{
				Code:    "var foo = () => { class C { m(@this.dec a) {} } };",
				Options: []any{"declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "declaration", Line: 1, Column: 5, EndLine: 1, EndColumn: 50},
				},
			},

			// ---- Locks in isOverloadedFunction's wrapper matching inside a
			// namespace body, where the container is a ModuleBlock rather than
			// the SourceFile ----
			{
				Code:    "namespace N { function foo(a: string): void; export function foo(a: any): void {} }",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 1, Column: 53, EndLine: 1, EndColumn: 82},
				},
			},

			// ---- Locks in isOverloadedFunction's wrapper matching: a
			// signature whose `export` modifier disagrees with the
			// implementation's is a different wrapper upstream
			// (`TSDeclareFunction` vs `ExportNamedDeclaration >
			// TSDeclareFunction`), so it never marks the implementation as
			// overloaded and the implementation is still reported ----
			{
				Code:    "function foo(a: string): void;\nexport function foo(a: any): void {}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 2, Column: 8, EndLine: 2, EndColumn: 37},
				},
			},
			{
				Code:    "export function foo(a: string): void;\nfunction foo(a: any): void {}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 2, Column: 1, EndLine: 2, EndColumn: 30},
				},
			},
			{
				Code:    "declare function foo(a: string): void;\nexport function foo(a: any): void {}",
				Options: []any{"expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expression", Line: 2, Column: 8, EndLine: 2, EndColumn: 37},
				},
			},
		},
	)
}
