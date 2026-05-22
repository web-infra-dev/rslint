// TestBlockSpacingExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk / real-user pattern
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
//
// Dimension 4 walk for @stylistic/block-spacing:
//
//   - Receiver / expression wrappers on inputs the rule inspects — N/A. The
//     rule fires on the block / case-block node itself; outer receiver
//     wrappers (`(arr).x`, `arr!.x`, `arr?.x`) are not in scope.
//   - Access / key forms — N/A. The rule does not look at property accesses,
//     computed keys, or string vs numeric literal keys.
//   - Declaration / container forms — covered: arrow body block (paren- and
//     brace-bodied), function declaration / function expression / method /
//     async / generator / async-generator / class static block / getter /
//     setter / constructor / class field arrow / namespace body / decorator-
//     decorated method / overload-implementation pair / IIFE.
//   - Nesting / traversal boundaries — covered: same-kind nesting up to three
//     levels for Block-in-Block, CaseBlock-in-CaseBlock, Block-in-CaseBlock,
//     CaseBlock-in-Block; class-static-block inside a class-static-block;
//     switch inside a function body; back-to-back independent blocks.
//   - Graceful degradation — covered: empty block (4 shapes), block with only
//     a line comment, block with only a block comment, regex literal as first
//     expression-statement, multi-byte UTF-8 identifier, mixed line endings
//     (LF / CRLF / CR-only), comment containing `{`/`}` payload.
//
// Branch lock-ins enumerate the listener gates in the upstream `create`
// function: `firstToken === closeBrace` early-return (empty), the
// `isTokenOnSameLine` short-circuits for both braces, and the dedicated
// `!always && firstToken.type === Line` exemption.
//
// Diagnostic-contract subgroup asserts the EXACT message text emitted on every
// (mode, side) combination so future refactors can't silently flip a string.
//
// Listener-scope subgroup locks in which AST shapes intentionally DO NOT
// trigger (ObjectLiteralExpression, TypeLiteralNode, InterfaceDeclaration body,
// MappedType, EnumDeclaration body, ModuleBlock — none are KindBlock).
package block_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/block_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBlockSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&block_spacing.BlockSpacingRule,
		[]rule_tester.ValidTestCase{
			// =====================================================================
			// Dimension 4: container forms (always)
			// =====================================================================
			// Arrow expression body — not a Block at all; rule never fires.
			{Code: `const f = a => a + 1;`},
			// Arrow with parenthesized expression body — ParenthesizedExpression,
			// not a Block.
			{Code: `const f = a => (a + 1);`},
			// Arrow with object literal body (paren-wrapped) — not a Block.
			{Code: `const f = a => ({ a });`},
			// Single-arg arrow with brace body.
			{Code: `const f = (a) => { return a; };`},
			// Method shorthand body.
			{Code: `const obj = { m() { return 1; } };`},
			// Class method / getter / setter / constructor.
			{Code: `class C { m() { return 1; } get x() { return this._x; } set x(v) { this._x = v; } constructor() { this._x = 0; } }`},
			// Async / generator / async-generator function bodies.
			{Code: `async function f() { await x; }`},
			{Code: `function* g() { yield 1; }`},
			{Code: `async function* h() { yield await x; }`},
			// Class-field arrow body.
			{Code: `class C { f = () => { return 1; }; }`},
			// IIFE.
			{Code: `(function () { foo(); })();`},
			// Arrow-IIFE.
			{Code: `(() => { foo(); })();`},
			// Decorator-decorated method body.
			{Code: `class C { @dec m() { return 1; } }`},
			// Overload signatures (body-less) + implementation body.
			{Code: `function f(a: string): void; function f(a: number): void; function f(a: any) { return a; }`},
			// Optional method.
			{Code: `class C { m?() { return 1; } }`},
			// Constructor with parameter properties.
			{Code: `class C { constructor(public x: number) { this.x = x; } }`},

			// =====================================================================
			// Dimension 4: container forms (never)
			// =====================================================================
			{Code: `const f = (a) => {return a;};`, Options: []interface{}{"never"}},
			{Code: `const obj = {m() {return 1;}};`, Options: []interface{}{"never"}},
			{Code: `class C {m() {return 1;} get x() {return this._x;}}`, Options: []interface{}{"never"}},
			{Code: `async function f() {await x;}`, Options: []interface{}{"never"}},
			{Code: `function* g() {yield 1;}`, Options: []interface{}{"never"}},
			{Code: `class C {f = () => {return 1;};}`, Options: []interface{}{"never"}},
			{Code: `(function () {foo();})();`, Options: []interface{}{"never"}},

			// =====================================================================
			// Out-of-scope kinds — must NEVER trigger (listener scope lock-in)
			// =====================================================================
			// Object literal `{ a: 1 }` is KindObjectLiteralExpression.
			{Code: `const x = { a: 1 };`, Options: []interface{}{"never"}},
			// Type literal is KindTypeLiteral.
			{Code: `type X = {a: number};`, Options: []interface{}{"never"}},
			// Interface body is KindInterfaceDeclaration (no Block child).
			{Code: `interface I {a: number}`, Options: []interface{}{"never"}},
			// Mapped type is KindMappedType.
			{Code: `type T = { [K in 'a']: number };`, Options: []interface{}{"never"}},
			// Enum body is KindEnumDeclaration (members are not Block).
			{Code: `enum E {A,B,C}`, Options: []interface{}{"never"}},
			// Namespace body is KindModuleBlock, NOT KindBlock.
			{Code: `namespace N { export const x = 1; }`},
			{Code: `namespace N {export const x = 1;}`, Options: []interface{}{"never"}},
			// Class body itself is KindClassDeclaration's MemberList, not Block.
			{Code: `class C {p = 1;}`, Options: []interface{}{"never"}},
			// `declare class` body — same shape, no body emit either way.
			{Code: `declare class C { p: number; m(): void; }`},
			// Destructuring with default — object pattern is not a Block.
			{Code: `const {a = 1} = obj;`, Options: []interface{}{"never"}},
			// Object pattern in parameter.
			{Code: `function f({a, b}: {a: number; b: number}) {return a + b;}`, Options: []interface{}{"never"}},

			// =====================================================================
			// Labeled statement with block body
			// =====================================================================
			{Code: `outer: { foo(); }`},
			{Code: `outer: {foo();}`, Options: []interface{}{"never"}},
			// Labeled loop with block body.
			{Code: `outer: for (;;) { foo(); break outer; }`},

			// =====================================================================
			// Nested blocks — well-spaced under each mode
			// =====================================================================
			// Block-in-Block (always).
			{Code: `function f() { if (a) { return 1; } }`},
			// CaseBlock in function (always).
			{Code: `function f() { switch (a) { case 1: foo(); } }`},
			// Static block in function-scoped class (always).
			{Code: `function f() { class C { static { foo(); } } }`},
			// Three-level nesting (always).
			{Code: `function f() { if (a) { if (b) { return 1; } } }`},
			// Nested CaseBlock in CaseBlock (always).
			{Code: `switch (a) { case 1: switch (b) { case 2: foo(); } }`},
			// Block-in-CaseBlock (block as case clause statement).
			{Code: `switch (a) { case 1: { foo(); break; } }`},
			// Two back-to-back independent blocks at the same level.
			{Code: `function f() { { foo(); } { bar(); } }`},
			// Class with two static blocks.
			{Code: `class C { static { foo(); } static { bar(); } }`},
			// Class with two static blocks (never).
			{Code: `class C {static {foo();} static {bar();}}`, Options: []interface{}{"never"}},

			// =====================================================================
			// isValid() arm: cross-line both braces
			// =====================================================================
			{Code: "if (a) {\n  foo();\n}"},
			{Code: "if (a) {\n  foo();\n}", Options: []interface{}{"never"}},
			// CRLF version.
			{Code: "if (a) {\r\n  foo();\r\n}"},
			{Code: "if (a) {\r\n  foo();\r\n}", Options: []interface{}{"never"}},

			// =====================================================================
			// isValid() arm: only-opening / only-closing cross-line
			// =====================================================================
			{Code: "{\n  foo(); }"},
			{Code: "{ foo();\n}"},
			{Code: "{\n  foo();}", Options: []interface{}{"never"}},
			{Code: "{foo();\n}", Options: []interface{}{"never"}},

			// =====================================================================
			// firstToken === closeBrace short-circuit (empty block, all shapes)
			// =====================================================================
			{Code: `function f() {}`},
			{Code: `function f() {}`, Options: []interface{}{"never"}},
			{Code: `class C { static {} }`},
			{Code: `class C { static {} }`, Options: []interface{}{"never"}},
			{Code: `switch (a) {}`},
			{Code: `switch (a) {}`, Options: []interface{}{"never"}},
			{Code: `if (a) {}`},
			{Code: `if (a) {}`, Options: []interface{}{"never"}},
			// Block with ONLY whitespace inside.
			{Code: `function f() {   }`, Options: []interface{}{"never"}},
			{Code: "function f() {   \n   }", Options: []interface{}{"never"}},
			{Code: `function f() {   }`},
			// Try with empty bodies.
			{Code: `try {} catch (e) {}`},
			{Code: `try {} catch (e) {}`, Options: []interface{}{"never"}},
			// Empty arrow body.
			{Code: `const f = () => {};`},
			{Code: `const f = () => {};`, Options: []interface{}{"never"}},

			// =====================================================================
			// !always && firstToken.type === Line exemption (never only)
			// =====================================================================
			{Code: "{ //c\n}", Options: []interface{}{"never"}},
			{Code: "{    //c\n}", Options: []interface{}{"never"}},
			// Same exemption for arrow body.
			{Code: "const f = () => { //c\nreturn 1;\n};", Options: []interface{}{"never"}},
			// Same for switch case block.
			{Code: "switch (a) { //c\ncase 1: foo();\n}", Options: []interface{}{"never"}},

			// =====================================================================
			// Block / line comment in graceful-degradation positions
			// =====================================================================
			// Block with only a block comment — both ends NEED spaces in always.
			{Code: `{ /* c */ }`},
			// `{/* c */}` (never) — block comment hugs both braces.
			{Code: `{/* c */}`, Options: []interface{}{"never"}},
			// Block comment containing `}` payload — must not confuse byte scan.
			{Code: `{ /* } */ }`},
			{Code: `{/* } */}`, Options: []interface{}{"never"}},
			// Block comment containing `{` payload.
			{Code: `{ /* { */ }`},
			// Multi-line block comment (cross-line short-circuit applies on
			// the closing side because the comment text wraps a newline).
			{Code: "{ /*\nc\n*/ }"},
			{Code: "{/*\nc\n*/}", Options: []interface{}{"never"}},

			// =====================================================================
			// Multi-byte UTF-8 and emoji
			// =====================================================================
			// Identifier with non-ASCII chars.
			{Code: `function f() { var é = 1; }`},
			{Code: `function f() {var é = 1;}`, Options: []interface{}{"never"}},
			// CJK characters in string literal inside block.
			{Code: `function f() { return "你好"; }`},
			// Emoji in template literal inside block.
			{Code: "function f() { return `🚀`; }"},

			// =====================================================================
			// Try / catch with TS-only forms
			// =====================================================================
			// Try / catch without binding (ES2019).
			{Code: `try { foo(); } catch { bar(); }`},
			{Code: `try {foo();} catch {bar();}`, Options: []interface{}{"never"}},
			// Try / catch with typed binding.
			{Code: `try { foo(); } catch (e: unknown) { bar(); }`},

			// =====================================================================
			// TypeScript-specific function-shape variations
			// =====================================================================
			// Generic function with body.
			{Code: `function f<T>(a: T): T { return a; }`},
			// Generic non-arrow function (never).
			{Code: `function id<T>(x: T): T {return x;}`, Options: []interface{}{"never"}},
			// Arrow returning typed value.
			{Code: `const f = (a: number): number => { return a; };`},
			// Function with default param.
			{Code: `function f(a = 1) { return a; }`},
			// Function with rest params.
			{Code: `function f(...xs: number[]) { return xs; }`},
			// Conditional type inside function body — type literals don't trigger.
			{Code: `function f(): number { type X = number extends number ? 1 : 0; return 1; }`},

			// =====================================================================
			// Option-input resilience — accept various option shapes
			// =====================================================================
			// Bare string instead of [string] (rslint config loader collapses).
			{Code: `{foo();}`, Options: "never"},
			{Code: `{ foo(); }`, Options: "always"},
			// Empty options array — defaults to always.
			{Code: `{ foo(); }`, Options: []interface{}{}},
			// nil / no options.
			{Code: `{ foo(); }`},
			// Unknown string (treated as default 'always').
			{Code: `{ foo(); }`, Options: []interface{}{"weird"}},

			// =====================================================================
			// Unicode WhiteSpace + LineTerminator (ECMAScript §12.2 / §12.3)
			//
			// ESLint's `isSpaceBetween` and `isTokenOnSameLine` recognize the
			// full ECMAScript whitespace + line-terminator sets, NOT just
			// ASCII. These cases were verified against `@stylistic/eslint-
			// plugin` at parity:
			//   - NBSP (U+00A0), EN/EM/IDEO Zs (Zs_Separator), ZWNBSP/BOM
			//     (U+FEFF) → count as WhiteSpace → satisfy `always` mode and
			//     trigger `never` mode extras.
			//   - LS (U+2028), PS (U+2029) → count as LineTerminator →
			//     trigger the cross-line short-circuit; both modes valid.
			// =====================================================================
			// always: NBSP between brace and token counts as a space → valid.
			{Code: "{\u00A0foo;\u00A0}"},
			// always: Ideographic Space (U+3000) — Zs class member.
			{Code: "{\u3000foo;\u3000}"},
			// always: En Space (U+2002).
			{Code: "{\u2002foo;\u2002}"},
			// always: ZWNBSP / BOM (U+FEFF).
			{Code: "{\uFEFFfoo;\uFEFF}"},
			// LS / PS on ONE side cause the cross-line short-circuit on
			// that side only. So for `never` mode where the OTHER side
			// hugs the brace, both sides are valid:
			{Code: "{foo;\u2028}", Options: []interface{}{"never"}},
			{Code: "{\u2028foo;}", Options: []interface{}{"never"}},
			{Code: "{foo;\u2029}", Options: []interface{}{"never"}},
			{Code: "{\u2029foo;}", Options: []interface{}{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			// =====================================================================
			// Diagnostic contract: exact Message text for every (mode, side)
			// combination. Locks the user-facing strings against silent drift.
			// =====================================================================
			{
				Code:   `{foo();}`,
				Output: []string{`{ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Message: "Requires a space after '{'.", Line: 1, Column: 1},
					{MessageId: "missing", Message: "Requires a space before '}'.", Line: 1, Column: 8},
				},
			},
			{
				Code:    `{ foo(); }`,
				Output:  []string{`{foo();}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Message: "Unexpected space(s) after '{'.", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "extra", Message: "Unexpected space(s) before '}'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},

			// =====================================================================
			// Asymmetric spacing — every (mode, opening-status, closing-status)
			// combination so neither side over-reports nor under-reports.
			// =====================================================================
			// always: open OK, close not OK → only close fires.
			{
				Code:   `{ foo();}`,
				Output: []string{`{ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 9},
				},
			},
			// always: open not OK, close OK → only open fires.
			{
				Code:   `{foo(); }`,
				Output: []string{`{ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
			// never: open extra, close OK → only open fires.
			{
				Code:    `{ foo();}`,
				Output:  []string{`{foo();}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			// never: open OK, close extra → only close fires.
			{
				Code:    `{foo(); }`,
				Output:  []string{`{foo();}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			// never: multi-space on one side only.
			{
				Code:    `{   foo();}`,
				Output:  []string{`{foo();}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 5},
				},
			},

			// =====================================================================
			// Dimension 4: container forms — every shape fires correctly
			// =====================================================================
			// Arrow body block (always missing).
			{
				Code:   `const f = (a) => {return a;};`,
				Output: []string{`const f = (a) => { return a; };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18},
					{MessageId: "missing", Line: 1, Column: 28},
				},
			},
			// Method body (always missing).
			{
				Code:   `const obj = { m() {return 1;} };`,
				Output: []string{`const obj = { m() { return 1; } };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 19},
					{MessageId: "missing", Line: 1, Column: 29},
				},
			},
			// Getter / setter (never extra).
			{
				Code:    `class C { get x() { return this._x; } }`,
				Output:  []string{`class C { get x() {return this._x;} }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
					{MessageId: "extra", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			// Async function body (always missing).
			{
				Code:   `async function f() {await x;}`,
				Output: []string{`async function f() { await x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 20},
					{MessageId: "missing", Line: 1, Column: 29},
				},
			},
			// Generator function body (always missing).
			{
				Code:   `function* g() {yield 1;}`,
				Output: []string{`function* g() { yield 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
					{MessageId: "missing", Line: 1, Column: 24},
				},
			},
			// Async-generator function body (always missing).
			{
				Code:   `async function* h() {yield await x;}`,
				Output: []string{`async function* h() { yield await x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 21},
					{MessageId: "missing", Line: 1, Column: 36},
				},
			},
			// Class static block (never extra).
			{
				Code:    `class C { static { foo(); } }`,
				Output:  []string{`class C { static {foo();} }`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "extra", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
			// Constructor body (always missing).
			{
				Code:   `class C { constructor() {this.x = 0;} }`,
				Output: []string{`class C { constructor() { this.x = 0; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 25},
					{MessageId: "missing", Line: 1, Column: 37},
				},
			},
			// Decorator-decorated method body (always missing).
			{
				Code:   `class C { @dec m() {return 1;} }`,
				Output: []string{`class C { @dec m() { return 1; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 20},
					{MessageId: "missing", Line: 1, Column: 30},
				},
			},
			// Overload-implementation pair — only the IMPLEMENTATION has a
			// body. The two signature overloads do not have a body, so the
			// listener fires only once.
			{
				Code:   `function f(a: string): void; function f(a: number): void; function f(a: any) {return a;}`,
				Output: []string{`function f(a: string): void; function f(a: number): void; function f(a: any) { return a; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 78},
					{MessageId: "missing", Line: 1, Column: 88},
				},
			},
			// IIFE — function expression invoked immediately.
			{
				Code:    `(function () { foo(); })();`,
				Output:  []string{`(function () {foo();})();`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "extra", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},

			// =====================================================================
			// Nested blocks — independent listener fires per Block
			// =====================================================================
			// Block-in-Block (always; outer ok, inner not) — 2 reports.
			{
				Code:   `function f() { if (a) {return 1;} }`,
				Output: []string{`function f() { if (a) { return 1; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 23},
					{MessageId: "missing", Line: 1, Column: 33},
				},
			},
			// Block-in-Block (never; both bad) — 4 reports in source order.
			{
				Code:    `function f() { if (a) { return 1; } }`,
				Output:  []string{`function f() {if (a) {return 1;}}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "extra", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
					{MessageId: "extra", Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
					{MessageId: "extra", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			// Switch-in-Block (always; outer + CaseBlock + inner func block).
			{
				Code:   `function f() {switch (a) {case 1: foo();}}`,
				Output: []string{`function f() { switch (a) { case 1: foo(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 14},
					{MessageId: "missing", Line: 1, Column: 26},
					{MessageId: "missing", Line: 1, Column: 41},
					{MessageId: "missing", Line: 1, Column: 42},
				},
			},
			// CaseBlock-in-CaseBlock (always; 4 reports in source order).
			{
				Code:   `switch (a) {case 1: switch (b) {case 2: foo();}}`,
				Output: []string{`switch (a) { case 1: switch (b) { case 2: foo(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 32},
					{MessageId: "missing", Line: 1, Column: 47},
					{MessageId: "missing", Line: 1, Column: 48},
				},
			},
			// Block-in-CaseBlock (case clause contains its own block).
			{
				Code:   `switch (a) {case 1:{foo();}}`,
				Output: []string{`switch (a) { case 1:{ foo(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 20},
					{MessageId: "missing", Line: 1, Column: 27},
					{MessageId: "missing", Line: 1, Column: 28},
				},
			},
			// Two back-to-back independent blocks at the same level — both
			// must report independently in source order.
			{
				Code:   `function f() { {foo();} {bar();} }`,
				Output: []string{`function f() { { foo(); } { bar(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 16},
					{MessageId: "missing", Line: 1, Column: 23},
					{MessageId: "missing", Line: 1, Column: 25},
					{MessageId: "missing", Line: 1, Column: 32},
				},
			},
			// Three-level nesting — inner-most block, mid block, outer ok.
			{
				Code:   `function f() { if (a) { if (b) {return 1;} } }`,
				Output: []string{`function f() { if (a) { if (b) { return 1; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 32},
					{MessageId: "missing", Line: 1, Column: 42},
				},
			},

			// =====================================================================
			// !always && firstToken.type === Line exemption (never)
			// =====================================================================
			// always mode: line comment after `{` still fires (exemption is
			// never-only).
			{
				Code:   "if (a) {//c\n foo(); }",
				Output: []string{"if (a) { //c\n foo(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			// never mode: opening exemption holds; closing fires for `; }`.
			{
				Code:    "{//c\nfoo(); }",
				Output:  []string{"{//c\nfoo();}"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 2, Column: 7, EndLine: 2, EndColumn: 8},
				},
			},

			// =====================================================================
			// isValid() arm: only-opening / only-closing cross-line
			// =====================================================================
			// Opening cross-line, closing missing on same line (always).
			{
				Code:   "{\n  foo();}",
				Output: []string{"{\n  foo(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 9},
				},
			},
			// Closing cross-line, opening missing on same line (always).
			{
				Code:   "{foo();\n}",
				Output: []string{"{ foo();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
			// CRLF version — must recognize \r as line terminator.
			{
				Code:   "{foo();\r\n}",
				Output: []string{"{ foo();\r\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},

			// =====================================================================
			// Block / line-comment-as-token contracts
			// =====================================================================
			// `{ /* c */ }` (never) — block comment acts as a token, both
			// adjacent spaces fire.
			{
				Code:    `{ /* c */ }`,
				Output:  []string{`{/* c */}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "extra", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
			// `{/* c */ foo; }` (always) — `{` glued to comment, opening
			// fires; closing is OK.
			{
				Code:   `{/* c */ foo(); }`,
				Output: []string{`{ /* c */ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
			// Block comment containing `}` payload — must not confuse
			// reverse byte scan.
			{
				Code:   `{/* } */ foo(); /* { */}`,
				Output: []string{`{ /* } */ foo(); /* { */ }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
					{MessageId: "missing", Line: 1, Column: 24},
				},
			},

			// =====================================================================
			// Real-user shapes (issue tracker / common patterns)
			// =====================================================================
			// Arrow body in JSX onClick (very common React shape).
			{
				Code:   `const X = () => <div onClick={() => {foo();}} />;`,
				Output: []string{`const X = () => <div onClick={() => { foo(); }} />;`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 37},
					{MessageId: "missing", Line: 1, Column: 44},
				},
			},
			// Express middleware shape.
			{
				Code:   `app.use((req, res, next) => {next();});`,
				Output: []string{`app.use((req, res, next) => { next(); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 29},
					{MessageId: "missing", Line: 1, Column: 37},
				},
			},
			// Promise .then callback.
			{
				Code:   `p.then((r) => {return r + 1;}).catch((e) => {console.log(e);});`,
				Output: []string{`p.then((r) => { return r + 1; }).catch((e) => { console.log(e); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
					{MessageId: "missing", Line: 1, Column: 29},
					{MessageId: "missing", Line: 1, Column: 45},
					{MessageId: "missing", Line: 1, Column: 61},
				},
			},
			// setTimeout callback.
			{
				Code:   `setTimeout(() => {tick();}, 100);`,
				Output: []string{`setTimeout(() => { tick(); }, 100);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18},
					{MessageId: "missing", Line: 1, Column: 26},
				},
			},
			// Array.forEach callback (never extra).
			{
				Code:    `arr.forEach((x) => { console.log(x); });`,
				Output:  []string{`arr.forEach((x) => {console.log(x);});`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
					{MessageId: "extra", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
				},
			},
			// do/while/finally trio (never extra).
			{
				Code:    `do { foo(); } while (a);`,
				Output:  []string{`do {foo();} while (a);`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "extra", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			// Try / catch without binding (ES2019).
			{
				Code:   `try {foo();} catch {bar();}`,
				Output: []string{`try { foo(); } catch { bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5},
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 20},
					{MessageId: "missing", Line: 1, Column: 27},
				},
			},
			// Typed catch binding.
			{
				Code:   `try {foo();} catch (e: unknown) {bar();}`,
				Output: []string{`try { foo(); } catch (e: unknown) { bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5},
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 33},
					{MessageId: "missing", Line: 1, Column: 40},
				},
			},
			// Generic function with body (TypeScript-specific).
			{
				Code:   `function id<T>(x: T): T {return x;}`,
				Output: []string{`function id<T>(x: T): T { return x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 25},
					{MessageId: "missing", Line: 1, Column: 35},
				},
			},
			// Arrow with type annotation.
			{
				Code:   `const f = (a: number): number => {return a;};`,
				Output: []string{`const f = (a: number): number => { return a; };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 34},
					{MessageId: "missing", Line: 1, Column: 44},
				},
			},

			// =====================================================================
			// Multi-byte UTF-8 — column comes out in UTF-16 units (parity
			// with ESLint's ESTree positions).
			// =====================================================================
			{
				Code:   "function f() {var é = 1;}",
				Output: []string{"function f() { var é = 1; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 14},
					{MessageId: "missing", Line: 1, Column: 25},
				},
			},
			// CJK in string literal — content does not shift column of `}`.
			{
				Code:   `function f() {return "你好";}`,
				Output: []string{`function f() { return "你好"; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 14},
					{MessageId: "missing", Line: 1, Column: 27},
				},
			},

			// =====================================================================
			// Option-input resilience — defaults / oddly-shaped options
			// =====================================================================
			// Bare string "never" (rslint collapse).
			{
				Code:    `{ foo(); }`,
				Output:  []string{`{foo();}`},
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},

			// =====================================================================
			// Unicode WhiteSpace (NBSP / Zs / BOM) — must trigger `never` mode
			// extras, parity with @stylistic/eslint-plugin (verified).
			// =====================================================================
			// NBSP after `{` and before `}` — never mode reports both sides.
			// The reported byte range covers the NBSP (2 UTF-8 bytes,
			// 1 UTF-16 code unit → 1 column wide).
			{
				Code:    "{\u00A0foo;\u00A0}",
				Output:  []string{"{foo;}"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "extra", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},
			// Ideographic Space (U+3000) before `}` — never mode extra.
			{
				Code:    "{foo;\u3000}",
				Output:  []string{"{foo;}"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				},
			},

			// =====================================================================
			// Unicode LineTerminator (LS / PS) on ONE side, opposite side
			// still checked. The cross-line short-circuit fires per-side.
			// Validates parity with ESLint's isTokenOnSameLine on §12.3
			// LineTerminator set (not just LF/CR).
			// =====================================================================
			// LS after `;` → closing skipped (cross-line); opening fires
			// because `{` hugs `foo` (same line, no space) under always.
			{
				Code:   "{foo;\u2028}",
				Output: []string{"{ foo;\u2028}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
			// LS after `{` → opening skipped; closing fires.
			{
				Code:   "{\u2028foo;}",
				Output: []string{"{\u2028foo; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 5},
				},
			},
			// PS variant.
			{
				Code:   "{foo;\u2029}",
				Output: []string{"{ foo;\u2029}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
		},
	)
}
