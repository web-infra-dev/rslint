package no_static_only_class_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_static_only_class"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All upstream `valid` / `invalid` cases (from `test/no-static-only-class.js`)
// are migrated below in declaration order. The upstream uses three test blocks:
//   - `test.snapshot` (JS) — see "JS snapshot block"
//   - `test.typescript` (TS-specific) — see "TS block"
//   - bare `test()` (JS) — see "JS bare block"
//
// Inline `Locks in upstream` comments mark cases we add for branches the
// upstream test suite itself does not exercise.
func TestNoStaticOnlyClass(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_static_only_class.NoStaticOnlyClassRule,
		[]rule_tester.ValidTestCase{
			// ---- JS snapshot block: valid ----

			// Empty class
			{Code: `class A {}`},
			{Code: `const A = class {}`},

			// `superClass`
			{Code: `class A extends B { static a() {}; }`},
			{Code: `const A = class extends B { static a() {}; }`},

			// Not static
			{Code: `class A { a() {} }`},
			{Code: `class A { constructor() {} }`},
			{Code: `class A { get a() {} }`},
			{Code: `class A { set a(value) {} }`},

			// `private` (PrivateIdentifier on key)
			{Code: `class A3 { static #a() {}; }`},
			{Code: `class A3 { static #a = 1; }`},
			{Code: `const A3 = class { static #a() {}; }`},
			{Code: `const A3 = class { static #a = 1; }`},

			// Static block — KindClassStaticBlockDeclaration is not a property
			// or method, so isStaticMember returns false and the class is not
			// reported.
			{Code: `class A2 { static {}; }`},

			// ---- TS block: valid ----

			// `private` repeated under TS-specific config (snapshot block already
			// covers these; upstream re-runs them under the TS suite).
			{Code: `class A { static #a() {}; }`},
			{Code: `class A { static #a = 1; }`},
			{Code: `const A = class { static #a() {}; }`},
			{Code: `const A = class { static #a = 1; }`},

			// TS class — TypeScript modifiers on a static field disqualify the
			// member from "static only" detection.
			{Code: `class A { static public a = 1; }`},
			{Code: `class A { static private a = 1; }`},
			{Code: `class A { static readonly a = 1; }`},
			{Code: `class A { static declare a = 1; }`},

			// Static block under TS
			{Code: `class A { static {}; }`},

			// ---- Lock-in cases (branches not directly covered upstream) ----

			// Locks in: a class containing both a static field AND a non-static
			// method must NOT be reported (the `body.some(!isStaticMember)`
			// branch returns early).
			{Code: `class A { static a = 1; b() {} }`},

			// Locks in: a static getter / setter alone (without other static
			// members) IS reported. The valid counter-test is the same shape
			// with a non-static getter — must not be reported.
			{Code: `class A { get a() {} static b() {} }`},

			// Locks in: ClassStaticBlockDeclaration mixed with valid static
			// members — the static block makes the rule skip the class.
			{Code: `class A { static {}; static a() {} }`},

			// Locks in: a static accessor / method with a `protected` modifier
			// is excluded (ModifierFlagsAccessibilityModifier covers protected).
			{Code: `class A { static protected a = 1; }`},

			// Locks in: a class-level decorator skips the rule even with a
			// static member.
			{Code: `@Foo class A { static a = 1; } function Foo(_:any){}`},

			// Locks in: a member-level decorator on a static field excludes the
			// member (ModifierFlagsDecorator).
			{Code: `class A { @bar static a = 1; } function bar(_:any,__:any){}`},

			// Locks in: an extends-only class with no static members is fine.
			{Code: `class A extends B {}`},

			// Locks in: outer class with a non-static method is NOT reported
			// even when its method body holds a *non-static-only* inner class.
			// Verifies the listener doesn't bleed past the class boundary.
			{Code: `class Outer { static helper() {} foo() { class Inner { instance() {} static a() {} } return Inner; } }`},

			// Locks in: TS abstract member modifier alone does NOT disqualify
			// a member. (Upstream's `isStaticMember` only checks accessibility,
			// readonly, declare, decorators, private — not abstract.) But an
			// abstract member is never `static`, so it is filtered by
			// `HasStaticModifier`. Class containing one abstract method has a
			// non-static-member → not reported.
			{Code: `abstract class A { abstract foo(): void; }`},

			// Locks in: a class containing only an index signature
			// (`[k: string]: any`). Index signatures are not
			// PropertyDeclaration / MethodDeclaration / accessor / constructor,
			// so they fail `isStaticMember` and the class is not reported.
			{Code: `class A { [k: string]: any; }`},

			// Locks in: a non-static field disqualifies the class even when
			// every method is static.
			{Code: `class A { static a() {} b = 1; }`},

			// Locks in: a class that extends through a parenthesized
			// expression. `extends (B)` — still an extends clause.
			{Code: `class A extends (B) { static a() {} }`},

			// Locks in: numeric / string / computed keys on static members
			// alone are still all-static and reported (covered in invalid)
			// — here, mixing one non-static numeric key blocks the rule.
			{Code: `class A { static 0() {} 1() {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- JS snapshot block: invalid ----

			// `class A { static a() {}; }` — head is `class A` (length 7).
			// Autofix: class → const, insert `= ` before `{`, append `;`.
			{
				Code:   `class A { static a() {}; }`,
				Output: []string{`const A = { a() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Message:   "Use an object instead of a class with only static members.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:   `class A { static a() {} }`,
				Output: []string{`const A = { a() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// Named class expression — autofix is suppressed upstream.
			{
				Code: `const A = class A { static a() {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 11, EndLine: 1, EndColumn: 18,
				}},
			},
			// Anonymous class expression — autofix drops `class`.
			{
				Code:   `const A = class { static a() {}; }`,
				Output: []string{`const A = { a() {}, }`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 11, EndLine: 1, EndColumn: 16,
				}},
			},
			// `static constructor()` — in tsgo this is a regular static method.
			{
				Code:   `class A { static constructor() {}; }`,
				Output: []string{`const A = { constructor() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// `export default class A` — modifiers are stripped from the head
			// range, so the report starts at `class` (column 16), matching
			// ESLint's ESTree where the ClassDeclaration child starts at `class`.
			// Named export-default → upstream skips autofix.
			{
				Code: `export default class A { static a() {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 16, EndLine: 1, EndColumn: 23,
				}},
			},
			// Anonymous `export default class` — autofix drops `class`.
			{
				Code:   `export default class { static a() {}; }`,
				Output: []string{`export default { a() {}, }`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 16, EndLine: 1, EndColumn: 21,
				}},
			},
			// `export class A` — head starts at `class` (column 8). Autofix
			// keeps `export` and rewrites `class A` → `const A = ...`.
			{
				Code:   `export class A { static a() {}; }`,
				Output: []string{`export const A = { a() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 8, EndLine: 1, EndColumn: 15,
				}},
			},

			// Multi-line class expression — autofix collapses `class\n\t{` →
			// `{`.
			{
				Code:   "function a() {\n\treturn class\n\t{\n\t\tstatic a() {}\n\t}\n}",
				Output: []string{"function a() {\n\treturn {\n\t\ta() {},\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      2, Column: 9, EndLine: 2, EndColumn: 14,
				}},
			},
			// Same with block-comment between `class` and `{` — special arm
			// replaces `class` with `{` and removes original `{`.
			{
				Code:   "function a() {\n\treturn class /* comment */\n\t{\n\t\tstatic a() {}\n\t}\n}",
				Output: []string{"function a() {\n\treturn { /* comment */\n\t\n\t\ta() {},\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      2, Column: 9, EndLine: 2, EndColumn: 14,
				}},
			},
			// Same with line-comment between `class` and `{`.
			{
				Code:   "function a() {\n\treturn class // comment\n\t{\n\t\tstatic a() {}\n\t}\n}",
				Output: []string{"function a() {\n\treturn { // comment\n\t\n\t\ta() {},\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      2, Column: 9, EndLine: 2, EndColumn: 14,
				}},
			},

			// Breaking edge cases — autofix produces a `const` with no
			// surrounding spaces (because the original had none).
			{
				Code:   "class A {static a(){}}\nclass B extends A {}",
				Output: []string{"const A = {a(){},};\nclass B extends A {}"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:   "class A {static a(){}}\nconsole.log(typeof A)",
				Output: []string{"const A = {a(){},};\nconsole.log(typeof A)"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			{
				Code:   "class A {static a(){}}\nconst a = new A;",
				Output: []string{"const A = {a(){},};\nconst a = new A;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// ---- TS block: invalid ----

			// Big multi-member class — single diagnostic + full autofix.
			{
				Code:   "class A {\n\tstatic a\n\tstatic b = 1\n\tstatic [c] = 2\n\tstatic [d]\n\tstatic e() {}\n\tstatic [f]() {}\n}",
				Output: []string{"const A = {\n\ta: undefined,\n\tb : 1,\n\t[c] : 2,\n\t[d]: undefined,\n\te() {},\n\t[f]() {},\n};"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// Same shape with explicit semicolons / extra parens.
			{
				Code:   "class A {\n\tstatic a;\n\tstatic b = 1;\n\tstatic [((c))] = ((2));\n\tstatic [d];\n\tstatic e() {};\n\tstatic [f]() {};\n}",
				Output: []string{"const A = {\n\ta: undefined,\n\tb : 1,\n\t[((c))] : ((2)),\n\t[d]: undefined,\n\te() {},\n\t[f]() {},\n};"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// Comments between every token — diagnostic location is the head;
			// fix preserves all interior comments verbatim, replacing only
			// `class`/`static`/`=`/`;` tokens.
			{
				Code:   "/* */\nclass /* */ A /* */ {\n\t/* */ static /* */ a /* */; /* */\n\t/* */ static /* */ b /* */ = /* */ 1 /* */; /* */\n\t/* */ static /* */ [ /* */ c /* */ ] /* */ = /* */ 2 /* */;  /* */\n\t/* */ static /* */ [/* */ d /* */] /* */;  /* */\n\t/* */ static /* */ /* */ e /* */ ( /* */ ) {/* */}/* */;  /* */\n\t/* */ static /* */ [/* */ f /* */ ] /* */ ( /* */ ) {/* */ }/* */ ;  /* */\n}\n/* */",
				Output: []string{"/* */\nconst /* */ A /* */ = {\n\t/* */ /* */ a /* */: undefined, /* */\n\t/* */ /* */ b /* */ : /* */ 1 /* */, /* */\n\t/* */ /* */ [ /* */ c /* */ ] /* */ : /* */ 2 /* */,  /* */\n\t/* */ /* */ [/* */ d /* */] /* */: undefined,  /* */\n\t/* */ /* */ /* */ e /* */ ( /* */ ) {/* */}/* */,  /* */\n\t/* */ /* */ [/* */ f /* */ ] /* */ ( /* */ ) {/* */ }/* */ ,  /* */\n};\n/* */"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      2, Column: 1,
				}},
			},

			// `this` in member value — still reported (no-fix case upstream).
			{
				Code: "class A {\n\tstatic a = 1;\n\tstatic b = this.a;\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// `this` in computed key — reported. Upstream still autofixes
			// because the value-text-includes-`this` check only inspects the
			// initializer's text, not the key.
			{
				Code:   `class A {static [this.a] = 1}`,
				Output: []string{`const A = {[this.a] : 1,};`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// String value containing 'this' — reported.
			{
				Code: "class A {\n\tstatic a = 1;\n\tstatic b = \"this\";\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// `declare class` — reported (no-fix upstream); class-level
			// `declare` modifier doesn't disqualify the class.
			{
				Code: `declare class A { static a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 9, EndLine: 1, EndColumn: 16,
				}},
			},
			// `abstract class` — reported.
			{
				Code: `abstract class A { static a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 10, EndLine: 1, EndColumn: 17,
				}},
			},
			// `implements` heritage — reported (only `extends` exempts).
			{
				Code: `class A implements B { static a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
				}},
			},
			// vscode-style realistic case — typed initialized fields, reported.
			{
				Code: "class NotebookKernelProviderAssociationRegistry {\n\tstatic extensionIds: (string | null)[] = [];\n\tstatic extensionDescriptions: string[] = [];\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 48,
				}},
			},

			// ---- JS bare block: invalid ----

			// Already covered above, but the bare block adds this exact case.
			// (Duplicate of the very first invalid case — kept to mirror upstream
			// layout faithfully.)
			{
				Code:   `class A { static a() {} }`,
				Output: []string{`const A = { a() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// ---- Lock-in cases (additional coverage) ----

			// Locks in: a class with only a static getter — head is reported,
			// fix preserves `get` keyword.
			{
				Code:   `class A { static get a() {} }`,
				Output: []string{`const A = { get a() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// Locks in: only a static setter — reported and fixed.
			{
				Code:   `class A { static set a(v) {} }`,
				Output: []string{`const A = { set a(v) {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// Locks in: nested static-only class expression inside a method
			// body of an OUTER (non-static-only) class. Only the inner class
			// is reported; the outer is not. Fix rewrites the inner.
			{
				Code:   `class Outer { foo() { return class { static a() {} }; } }`,
				Output: []string{`class Outer { foo() { return { a() {}, }; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 30, EndLine: 1, EndColumn: 35,
				}},
			},

			// Locks in: type parameters appear in the head range. Autofix is
			// suppressed (intentional rslint divergence — see rule .md).
			{
				Code: `class A<T> { static a(): T { return null!; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 11,
				}},
			},
			{
				Code: `class A<T extends B> { static a(): T { return null!; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
				}},
			},

			// Locks in: doubly-nested static-only class. Both inner classes
			// are independently invalid → 2 reports + 2 fixes applied.
			{
				Code:   "class A {\n\tstatic a() {}\n}\nclass B {\n\tstatic b() {}\n}",
				Output: []string{"const A = {\n\ta() {},\n};\nconst B = {\n\tb() {},\n};"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStaticOnlyClass", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "noStaticOnlyClass", Line: 4, Column: 1, EndLine: 4, EndColumn: 8},
				},
			},

			// Locks in: numeric / string keyed static members are still
			// reported and fixed.
			{
				Code:   `class A { static 0() {}; static "x"() {} }`,
				Output: []string{`const A = { 0() {}, "x"() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: two methods with no separators — verify per-method
			// `,` insertion order doesn't leave the second member without a
			// trailing comma.
			{
				Code:   `class A { static a() {} static b() {} }`,
				Output: []string{`const A = { a() {}, b() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: mixed property + method.
			{
				Code:   `class A { static a = 1; static b() {} }`,
				Output: []string{`const A = { a : 1, b() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: comment between value and trailing `;` in a property.
			// `findTrailingSemicolonInRange` must skip the comment when
			// locating `;`.
			{
				Code:   "class A { static a /* c */; }",
				Output: []string{"const A = { a /* c */: undefined, };"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: comment between `=` and value that contains `this`.
			// upstream's `getText(value)` excludes leading trivia, so the
			// fix proceeds despite the `this` token sitting in the comment.
			{
				Code:   "class A { static a = /* this */ 1 }",
				Output: []string{"const A = { a : /* this */ 1, };"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: computed-key static method — fix preserves `[k]`.
			{
				Code:   `class A { static [k]() {} }`,
				Output: []string{`const A = { [k]() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: generator method as a sole static member.
			{
				Code:   `class A { static *gen() { yield 1; } }`,
				Output: []string{`const A = { *gen() { yield 1; }, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: async method — `async` modifier sits between
			// `static` and the name; we only remove `static` and its
			// trailing whitespace, leaving `async` intact.
			{
				Code:   `class A { static async fn() {} }`,
				Output: []string{`const A = { async fn() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// ---- Go / tsgo-specific edge coverage ----

			// Locks in: async generator. `static async *fn() {}` should
			// preserve `async` and `*` after removing `static`. tsgo carries
			// `async` as a modifier and `*` as a separate AsteriskToken
			// (asteriskToken on MethodDeclaration), neither of which lives
			// in the slice we delete.
			{
				Code:   "class A { static async *fn() { yield 1; } }",
				Output: []string{"const A = { async *fn() { yield 1; }, };"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: arrow-function static field. The initializer text
			// is `() => 1` (no `this`) so the autofix proceeds. Verifies
			// the trivia-skipping `this` heuristic doesn't reach into the
			// arrow body's syntactic tokens.
			{
				Code:   `class A { static a = () => 1 }`,
				Output: []string{`const A = { a : () => 1, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: JSDoc on a class member is leading trivia in tsgo.
			// staticKeywordToken uses scanner.SkipTrivia from the static
			// modifier's Pos, so JSDoc is preserved when `static ` is
			// removed.
			{
				Code:   "class A {\n\t/** keep me */\n\tstatic a() {}\n}",
				Output: []string{"const A = {\n\t/** keep me */\n\ta() {},\n};"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: class not at line 1 — head range columns/lines must
			// reflect the class's actual position. The leading code path
			// goes through scanner.GetScannerForSourceFile with a non-zero
			// start.
			{
				Code: "// header line 1\n// header line 2\nclass A {\n\tstatic foo() {}\n}",
				Output: []string{
					"// header line 1\n// header line 2\nconst A = {\n\tfoo() {},\n};",
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      3, Column: 1, EndLine: 3, EndColumn: 8,
				}},
			},

			// Locks in: PropertyDeclaration with PostfixToken `!` (definite
			// assignment assertion). Autofix is suppressed because
			// `{ a! : 1, }` would be invalid TS (TS1255). See
			// "Differences from ESLint".
			{
				Code: `class A { static a! = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: PropertyDeclaration with PostfixToken `?` (optional).
			// Autofix is suppressed because `{ a?: undefined, }` would be
			// invalid TS (TS1162). See "Differences from ESLint".
			{
				Code: `class A { static a? = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: multiple consecutive `;` between methods. Each
			// stray SemicolonClassElement is its own member in tsgo;
			// upstream's ESTree absorbs them as filler and ESLint's fix
			// only redirects the first to a comma. rslint extends the
			// cleanup so that subsequent `;` characters are also erased
			// (their trivia/comments preserved), keeping the rewritten
			// output valid for `static a() {};; static b() {};` etc.
			{
				Code:   `class A { static a() {};; static b() {}; }`,
				Output: []string{`const A = { a() {}, b() {}, };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
			// Locks in: trivia between consecutive `;` is preserved when
			// each extra `;` token is erased (only the `;` char goes —
			// the comment between `;` and `;` survives).
			{
				Code:   `class A { static a() {} /*A*/ ; /*B*/ ; /*C*/ }`,
				Output: []string{`const A = { a() {} /*A*/ , /*B*/  /*C*/ };`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: TC39 `accessor` field. Autofix suppressed because
			// `accessor` has no object-literal analog. See "Differences
			// from ESLint".
			{
				Code: `class A { static accessor a = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},

			// Locks in: static overload signatures. The first two methods
			// have no body (Body() == nil); object-literal methods must
			// have bodies, so the autofix is suppressed. Diagnostic still
			// fires (the class is logically static-only).
			{
				Code: "class A {\n\tstatic a(x: number): void;\n\tstatic a(x: string): void;\n\tstatic a(x: any): void {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticOnlyClass",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
				}},
			},
		},
	)
}
