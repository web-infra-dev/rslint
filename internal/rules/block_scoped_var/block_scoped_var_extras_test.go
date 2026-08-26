// Edge-shape augmentation (Dimensions 1-4) and upstream-branch lock-ins for
// block-scoped-var, beyond what tests/lib/rules/block-scoped-var.js covers.
// See block_scoped_var_upstream_test.go for the migrated upstream suite.
package block_scoped_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func outOfScopeTSX(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	testCase := outOfScope(code, errors...)
	testCase.Tsx = true
	return testCase
}

func outOfScopeJSX(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	testCase := outOfScope(code, errors...)
	testCase.FileName = "block-scoped-var-invalid.jsx"
	testCase.TSConfig = "tsconfig.allowJs.json"
	return testCase
}

func scopeErrWithEnd(name string, defLine, defColumn, line, column, endColumn int) rule_tester.InvalidTestCaseError {
	testError := scopeErr(name, defLine, defColumn, line, column)
	testError.EndColumn = endColumn
	return testError
}

func TestBlockScopedVarExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&BlockScopedVarRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: TS-specific wrappers on an in-scope read ----
			{Code: "function f() { if (true) { var a = 1; (a).toString(); } }"},
			{Code: "function f() { if (true) { var a = 1; a!.toString(); } }"},
			{Code: "function f() { if (true) { var a: any = 1; (a as string).length; } }"},
			{Code: "function f() { if (true) { var a = 1; a?.toString(); } }"},

			// ---- Dimension 2: 3+ levels of nesting, only the boundary that matters flags ----
			{Code: "function f() { if (true) { if (true) { if (true) { var a = 1; a; } } } }"},
			{Code: "function f() { var a = 1; if (true) { if (true) { if (true) { a; } } } }"},

			// ---- Dimension 4: graceful degradation on binding-pattern shapes ----
			{Code: "function f() { if (true) { var { a, ...rest } = { a: 1 }; a; rest; } }"},
			{Code: "function f() { if (true) { var [a, ...rest] = [1]; a; rest; } }"},
			{Code: "function f() { if (true) { var {} = {}; } }"},
			{Code: "function f() { if (true) { var [] = []; } }"},

			// ---- Dimension 4: TS type annotation on an in-scope var ----
			{Code: "function f() { if (true) { var a: number = 1; a; } }"},

			// ---- Dimension 4: ambient `declare var` does not crash or false-positive ----
			{Code: "declare var a: number; a;"},
			{Code: "function f() { declare var a: number; a; }"},

			// ---- Real-user: catch-clause param shadows an unrelated same-named outer var ----
			{Code: "function f() { var e = 1; try {} catch (e) { e; } e; }"},

			// ---- Locks in: a bare `var` declarator (no initializer) never
			// ---- creates a self-reference, so two sibling bare declarations
			// ---- of the same name never cross-flag each other. ----
			{Code: "if (true) { var a; } else { var a; }"},
			{Code: "function f() { for (var a;;) {} for (var a;;) {} }"},

			// ---- Locks in: a for-of/for-in binding's implicit self-reference
			// ---- does not extend to an unrelated later bare `var` of the
			// ---- same name — it is still gated on that later declarator
			// ---- having its own initializer.
			{Code: "function f() { for (var a of []) {} var a; }"},

			// ---- Real-user: class static blocks are independent var scopes,
			// ---- so the same name declared in two different static blocks
			// ---- of one class never cross-flags, with or without an
			// ---- initializer.
			{Code: "class C { static { var a; a; } static { var a; a; } }"},
			{Code: "class C { static { var a = 1; } static { var a = 2; } }"},

			// ---- Real-user: a nested function's `var` never cross-flags an
			// ---- outer function's same-named `var` — they are different
			// ---- symbols in different var scopes, even with matching
			// ---- initializer shapes.
			{Code: "function outer() { if (true) { var a = 1; } function inner() { if (true) { var a = 2; } } }"},

			// ---- Scope-manager parity: declaration writes resolve in their
			// ---- lexical scope, independently from the hoisted `var` symbol.
			{Code: "function f(){try{}catch(e){if(x){var e=1;}else{var e=2;}}}"},
			{Code: "export{};function f(){if(x){var A;}export type {A};}"},
			{Code: "if(q){var A;}else{var A;}([x=A] += rhs)"},
			{Code: "if(x){var A;}interface A{}type X<T>=A extends infer A?A:never"},
			{Code: "export{};if(x){var Map;}type X=Map<string,string>"},
			{Code: "if(q){var undefined}else{var undefined}type X=undefined"},
			{Code: "if(q){var A,B}else{var A,B}class C implements A.B{}"},

			// ---- A function header cannot resolve to a body-only declaration,
			// ---- even when tsgo lowers another parameter through the body.
			{Code: "const a=9;function f(x=(a=1),y=z?.q){if(q){var a;}}"},

			// ---- Reopened TypeScript namespaces have distinct scope-manager
			// ---- variables even though TypeScript merges their exported symbols.
			{Code: "namespace N{export var a;if(q){var a=1;}}namespace N{export var a;if(r){var a=2;}}"},

			// ---- Method parameter decorators run outside the method function
			// ---- scope, but keep scopes nested inside the decorator expression.
			{Code: "if(q){var a;}else{var a;}class C{m(@(()=>{let a=1;return a})() x:any){}}"},

			// ---- TSX pragma lookup is definition-order-sensitive for parameters.
			{Code: "function f(x=<div/>,React){if(q){var React;}}", Tsx: true},
			{Code: "if(q){var React;}else{var React;}class C{m(@dec React:any=<div/>){}}", Tsx: true},

			// ---- Function-type defaults are parser-accepted but ignored by the
			// ---- TypeScript scope manager's type visitor.
			{Code: "if(q){var React;}else{var React;}type F=(x=<div/>)=>void", Tsx: true},
			{Code: "if(q){var A;}else{var A;}type F=(x=A)=>void"},

			// ---- Named class self-references belong to the class-local variable,
			// ---- not an outer var merged by TypeScript's binder.
			{Code: "if(x){var A;}class A{m(){return A}}"},

			// ---- Import-type qualifiers are not visited by TypeVisitor; type
			// ---- arguments remain references and are covered by invalid cases.
			{Code: "export{};if(x){var A;}interface A{}type X=import(\"m\").A"},
			{Code: "export{};if(x){var A;}interface A{}type X=typeof import(\"m\").A"},

			// ---- JSX in JS/JSX follows Espree: no React pseudo-reference and
			// ---- no references for namespaced tag pieces or lowercase Unicode.
			{Code: "if(x){var React;}else{var React;}<div/>", FileName: "block-scoped-var-valid.jsx", TSConfig: "tsconfig.allowJs.json"},
			{Code: "if(x){var foo;}else{var foo;}<foo:bar/>", FileName: "block-scoped-var-valid.jsx", TSConfig: "tsconfig.allowJs.json"},
			{Code: "if(x){var é;}else{var é;}<é/>", Tsx: true},
			{Code: "function f(){if(x){var ƛ;}else{var ƛ;}<ƛ></ƛ>}", Tsx: true},
			{Code: "function f(){if(x){var ɤ;}else{var ɤ;}<ɤ></ɤ>}", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: TS wrappers around an out-of-scope read still flag the identifier ----
			outOfScope("function f() { if (true) { var a = 1; } (a).toString(); }",
				scopeErr("a", 1, 32, 1, 42)),
			outOfScope("function f() { if (true) { var a = 1; } ((a)).toString(); }",
				scopeErr("a", 1, 32, 1, 43)),
			outOfScope("function f() { if (true) { var a = 1; } a!.toString(); }",
				scopeErr("a", 1, 32, 1, 41)),
			outOfScope("function f() { if (true) { var a: any = 1; } (a as string).length; }",
				scopeErr("a", 1, 32, 1, 47)),
			outOfScope("function f() { if (true) { var a = 1; } a?.toString(); }",
				scopeErr("a", 1, 32, 1, 41)),
			outOfScope("function f() { if (true) { var a = function(){}; } a?.(); }",
				scopeErr("a", 1, 32, 1, 52)),

			// ---- Dimension 4: computed key and template-literal reads of an out-of-scope var ----
			outOfScope("function f() { if (true) { var a = 1; } var o = { [a]: 1 }; }",
				scopeErr("a", 1, 32, 1, 52)),
			outOfScope("function f() { if (true) { var a = 1; } `${a}`; }",
				scopeErr("a", 1, 32, 1, 44)),

			// ---- Dimension 4: destructured rest binding used outside its block ----
			outOfScope("function f() { if (true) { var { a, ...rest } = { a: 1 }; } rest; }",
				scopeErr("rest", 1, 40, 1, 61)),
			outOfScope("function f() { if (true) { var [a, ...rest] = [1]; } rest; }",
				scopeErr("rest", 1, 39, 1, 54)),

			// ---- Dimension 2: 3-level nesting, innermost declaration leaks past its own block only ----
			outOfScope("function f() { if (true) { if (true) { if (true) { var a = 1; } a; } } }",
				scopeErr("a", 1, 56, 1, 65)),

			// ---- Real-user: `var` assigned only inside a conditional branch, read after the block ----
			outOfScope("function connect(cb) { if (cb) { var onDone = function() { cb(); }; } return onDone; }",
				scopeErr("onDone", 1, 38, 1, 78)),

			// ---- Real-user: try/catch fallback assignment read after both blocks (classic gotcha) ----
			// Locks in that sibling `var` declarations of the same name cross-flag
			// each other's own identifier, not just later reads (see the if/else
			// and duplicate-for-loop cases in the upstream suite).
			outOfScope("function parse(input) { try { var result = input.a; } catch (e) { var result = null; } return result; }",
				scopeErr("result", 1, 71, 1, 35),
				scopeErr("result", 1, 35, 1, 71),
				scopeErr("result", 1, 35, 1, 95),
				scopeErr("result", 1, 71, 1, 95)),

			// ---- Locks in: an initialized `var` declarator's self-reference
			// ---- only fires against a sibling declarator that itself has an
			// ---- initializer (or is a for-in/for-of binding) — a bare
			// ---- sibling declarator is never flagged and never flags back.
			outOfScope("function f() { try { var a = compute(); } catch (e) { var a; } return a; }",
				scopeErr("a", 1, 59, 1, 26),
				scopeErr("a", 1, 26, 1, 71),
				scopeErr("a", 1, 59, 1, 71)),
			outOfScope("function f() { for (var a in {}) {} for (var a in {}) {} }",
				scopeErr("a", 1, 46, 1, 25),
				scopeErr("a", 1, 25, 1, 46)),

			// ---- Locks in: `hasInitializer` is checked per declarator, not
			// ---- per declaration list — a bare declarator sitting next to an
			// ---- initialized one in the same `var` statement keeps its own
			// ---- gating.
			outOfScope("if (true) { var a, b = 1; } else { var a = 2, b; }",
				scopeErr("b", 1, 47, 1, 20),
				scopeErr("a", 1, 17, 1, 40)),

			// ---- Locks in: a for-of binding's implicit self-reference does
			// ---- cross-flag against a *later* declarator that has its own
			// ---- initializer.
			outOfScope("function f() { for (var a of []) {} var a = 1; }",
				scopeErr("a", 1, 25, 1, 41)),

			// ---- Real-user: a fully bare branch among initialized ones is
			// ---- still checked as an occurrence (against reads and against
			// ---- initialized siblings) even though it never joins the
			// ---- sibling list itself.
			outOfScope("if (x) { var a = 1; } else if (y) { var a; } else { var a = 3; }",
				scopeErr("a", 1, 41, 1, 14),
				scopeErr("a", 1, 57, 1, 14),
				scopeErr("a", 1, 14, 1, 57),
				scopeErr("a", 1, 41, 1, 57)),

			// ---- eslint-scope records both the declaration initialization and
			// ---- the binding default as writes to each destructured name.
			outOfScope("if (true) { var { a = 1 } = {}; } else { var { a = 1 } = {}; }",
				scopeErr("a", 1, 48, 1, 19),
				scopeErr("a", 1, 48, 1, 19),
				scopeErr("a", 1, 19, 1, 48),
				scopeErr("a", 1, 19, 1, 48)),

			// ---- Canonical var-scope identity is shared with function
			// ---- declarations, while each declaration write keeps its own
			// ---- lexical resolution.
			outOfScope("function f(){function a(){}if(x){var a=1;}else{var a=2;}}",
				scopeErr("a", 1, 52, 1, 38),
				scopeErr("a", 1, 38, 1, 52)),

			// ---- A duplicate declarator contributes one definition occurrence,
			// ---- so the trailing reference is reported exactly once.
			outOfScope("function f(){if(x){var a,a;}a;}",
				scopeErr("a", 1, 24, 1, 29)),

			// ---- Catch parameters intercept declaration writes, but references
			// ---- to the function-scoped variable still use the canonical group.
			outOfScope("function f(){try{}catch(e){if(x){var e=1;}}if(y){var e=2;}e;}",
				scopeErr("e", 1, 38, 1, 54),
				scopeErr("e", 1, 38, 1, 59),
				scopeErr("e", 1, 54, 1, 59)),

			// ---- Parameter defaults are declaration writes on the merged
			// ---- function variable, including detached duplicate bindings.
			outOfScope("function f(a:number=1){if(x){var a;}}",
				scopeErrWithEnd("a", 1, 34, 1, 12, 20)),
			outOfScope("function f([a,a=1]=[]){if(x){var a;}}",
				scopeErr("a", 1, 34, 1, 13),
				scopeErr("a", 1, 34, 1, 15),
				scopeErr("a", 1, 34, 1, 15)),
			outOfScope("function f(a=0){function a(){}if(q){var a;}}",
				scopeErr("a", 1, 41, 1, 12)),

			// ---- TSESTree declaration identifiers retain their type annotation
			// ---- in the reported range.
			outOfScope("if(x){var a:number=1}else{var a:string=2}",
				scopeErrWithEnd("a", 1, 31, 1, 11, 19),
				scopeErrWithEnd("a", 1, 11, 1, 31, 39)),

			// ---- Parser-accepted non-var writes and namespace type references
			// ---- participate in the same TSESTree scope variable.
			outOfScope("function f(){if(x){var a;}const a=1;}",
				scopeErr("a", 1, 24, 1, 33)),
			outOfScope("export{};if(x){var A;}namespace A{}type X=A",
				scopeErr("A", 1, 20, 1, 43)),
			outOfScope("export namespace N{if(x){var A;}export namespace A{}type X=A}",
				scopeErr("A", 1, 30, 1, 60)),
			outOfScope("export{};if(x){var A;}export interface A{}type X=A",
				scopeErr("A", 1, 20, 1, 50)),
			outOfScope("export{};if(x){var A;}export class A{}type X=A",
				scopeErr("A", 1, 20, 1, 46)),
			outOfScope("export{};if(x){var A;}interface A{}type X=import(\"m\").X<A>",
				scopeErr("A", 1, 20, 1, 57)),
			outOfScope("export{};if(q){var A;}else{var A;}export as namespace A;",
				scopeErr("A", 1, 20, 1, 55),
				scopeErr("A", 1, 32, 1, 55)),
			outOfScope("if(q){var x;}else{var x;}interface F{[x:string]:typeof x}",
				scopeErr("x", 1, 11, 1, 56),
				scopeErr("x", 1, 23, 1, 56)),
			outOfScope("if(x){var PropertyKey;}type X=PropertyKey",
				scopeErr("PropertyKey", 1, 11, 1, 31)),
			outOfScope("if(q){var Infinity}else{var Infinity}type X=Infinity",
				scopeErr("Infinity", 1, 11, 1, 45),
				scopeErr("Infinity", 1, 29, 1, 45)),
			outOfScope("/*global Custom*/if(q){var Custom}else{var Custom}type X=Custom",
				scopeErr("Custom", 1, 28, 1, 58),
				scopeErr("Custom", 1, 44, 1, 58)),
			{
				Code:     "if(x){var PropertyKey;}type X=PropertyKey",
				TSConfig: "tsconfig.noLib.json",
				Errors: []rule_tester.InvalidTestCaseError{
					scopeErr("PropertyKey", 1, 11, 1, 31),
				},
			},
			outOfScope("if(x){var A;}type X<T>=typeof A extends infer A?typeof A:typeof A",
				scopeErr("A", 1, 11, 1, 31),
				scopeErr("A", 1, 11, 1, 56),
				scopeErr("A", 1, 11, 1, 65)),
			outOfScope("function f<T extends ((a:any)=>typeof a)>(a=1){if(q){var a;}}",
				scopeErr("a", 1, 58, 1, 43)),

			// ---- Function headers reject body-only targets, but allow an early
			// ---- type definition merged with the body var.
			outOfScope("if(q){var A;}else{var A;}function f(x=(()=>A)){function A(){}}",
				scopeErr("A", 1, 11, 1, 44),
				scopeErr("A", 1, 23, 1, 44)),
			outOfScope("const A=0;function f<A>(x:typeof A){if(q){var A;}}",
				scopeErr("A", 1, 47, 1, 34)),
			outOfScope("export{};if(x){var A;}interface A{}function f(){let A=0;class C implements A{}}",
				scopeErr("A", 1, 20, 1, 76)),

			// ---- A named class expression is not in scope in its class-level
			// ---- decorator; method parameter decorators are outside the method
			// ---- function scope but preserve scopes nested in the decorator.
			outOfScope("function f(){if(x){var A;}else{var A;}const X=@dec(A) class A{}}",
				scopeErr("A", 1, 24, 1, 52),
				scopeErr("A", 1, 36, 1, 52)),
			outOfScope("if(q){var a;}else{var a;}class C{m(@a a:any){if(z){var a;}}}",
				scopeErr("a", 1, 11, 1, 37),
				scopeErr("a", 1, 23, 1, 37)),
			outOfScope("if(q){var a;}else{var a;}class C{m(@(()=>{if(z){var a=1;}return a})() x:any){}}",
				scopeErr("a", 1, 53, 1, 65)),

			// ---- TypeScript's JSX scope manager emits one definition-based
			// ---- React pseudo-reference, with definition-order-sensitive scopes.
			outOfScopeTSX("if(x){var React;}else{var React;}<div/>",
				scopeErr("React", 1, 27, 1, 11)),
			{
				Code:     "if(x){var preact;}else{var preact;}<></>",
				Tsx:      true,
				TSConfig: "tsconfig.block-scoped-var-jsx.json",
				Errors: []rule_tester.InvalidTestCaseError{
					scopeErr("preact", 1, 28, 1, 11),
					scopeErr("preact", 1, 28, 1, 11),
				},
			},
			outOfScopeTSX("<div/>;if(a){var React;}<span/>;if(b){var React;}",
				scopeErr("React", 1, 43, 1, 18)),
			outOfScopeTSX("function f({x=<div/>,React}){if(q){var React;}}",
				scopeErr("React", 1, 40, 1, 22)),
			outOfScopeTSX("function f<React>(){<div/>;if(q){var React;}else{var React;}}",
				scopeErr("React", 1, 38, 1, 12),
				scopeErr("React", 1, 54, 1, 12)),
			outOfScopeTSX("if(x){var React;}else{var React;}interface I<React>{[<div/>]:string}",
				scopeErr("React", 1, 11, 1, 46),
				scopeErr("React", 1, 27, 1, 46)),
			outOfScopeTSX("if(x){var React;}else{var React;}class C{static{<div/>;let React;}}",
				scopeErr("React", 1, 27, 1, 11)),

			// ---- JSX inside a named class-expression decorator resolves outside
			// ---- the class self-name for both pseudo and physical references.
			outOfScopeTSX("function f(){if(x){var React;}const C=@dec(<React/>) class React{}}",
				scopeErr("React", 1, 24, 1, 45)),

			// ---- TSESTree treats both pieces of opening and closing namespaced
			// ---- tags as physical value references.
			outOfScopeTSX("function f(){if(x){var foo;var bar;}else{var foo;var bar;}<foo:bar></foo:bar>;}",
				scopeErr("foo", 1, 24, 1, 60),
				scopeErr("foo", 1, 46, 1, 60),
				scopeErr("bar", 1, 32, 1, 64),
				scopeErr("bar", 1, 54, 1, 64),
				scopeErr("foo", 1, 24, 1, 70),
				scopeErr("foo", 1, 46, 1, 70),
				scopeErr("bar", 1, 32, 1, 74),
				scopeErr("bar", 1, 54, 1, 74)),

			// ---- JSX visitor ordering follows typescript-eslint rather than
			// ---- tsgo's generic child order.
			outOfScopeTSX("if(x){var React;}else{var React;}declare function call<T>(x:any):any;call<{[<div/>]:1}>(()=>{interface React{};return <div/>})",
				scopeErr("React", 1, 11, 1, 104),
				scopeErr("React", 1, 27, 1, 104)),
			outOfScopeTSX("if(x){var React;}else{var React;}((({[(()=>{interface React{};return <div/>})()]:a} as {[<div/>]:1}) as any)=o)",
				scopeErr("React", 1, 11, 1, 55),
				scopeErr("React", 1, 27, 1, 55)),
			outOfScopeTSX("if(x){var React;}else{var React;}for((()=>{interface React{};<span/>;return o})()[<div/>] in xs){}",
				scopeErr("React", 1, 27, 1, 11)),

			// ---- JS/JSX follows Espree physical component references without
			// ---- adding the TypeScript React pseudo-reference.
			outOfScopeJSX("if(x){var Foo;}else{var Foo;}<Foo/>",
				scopeErr("Foo", 1, 11, 1, 31),
				scopeErr("Foo", 1, 25, 1, 31)),
			outOfScopeJSX("if(x){var Foo;}else{var Foo;}<Foo></Foo>",
				scopeErr("Foo", 1, 11, 1, 31),
				scopeErr("Foo", 1, 25, 1, 31)),
		},
	)
}
