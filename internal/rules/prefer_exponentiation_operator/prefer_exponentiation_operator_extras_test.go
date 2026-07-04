package prefer_exponentiation_operator

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferExponentiationOperatorExtras locks in branches and edge shapes that the upstream test suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch, Dimension 4 row, or tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.
func TestPreferExponentiationOperatorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferExponentiationOperatorRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: dynamic element access keys stay unmatched ----
			{Code: "const key = 'pow'; Math[key](a, b);"},
			{Code: "Math[Symbol.for('pow')](a, b);"},
			{Code: "Math[0](a, b);"},
			{Code: "globalThis[getName()].pow(a, b);"},
			{Code: "Math[`p${suffix}`](a, b);"},
			{Code: "Math['po' + suffix](a, b);"},

			// ---- Dimension 4: value-shadowed receivers stay unmatched ----
			{Code: "let Math = other; (Math as any).pow(a, b);"},
			{Code: "const globalThis = other; (globalThis as any).Math.pow(a, b);"},
			{Code: "import { Math } from 'math-lib'; Math.pow(a, b);"},
			{Code: "enum Math { pow } Math.pow(a, b);"},
			{Code: "namespace Math { export const pow = f; } Math.pow(a, b);"},
			{Code: "const { Math } = ns; Math.pow(a, b);"},
			{Code: "try {} catch (Math) { Math.pow(a, b); }"},
			{Code: "for (let Math of items) { Math.pow(a, b); }"},

			// ---- Dimension 4: private names are a different key class from the string "pow" ----
			{Code: "class C { #pow; foo() { Math.#pow(a, b); } }"},

			// ---- Dimension 4: non-CALL uses of Math.pow stay unmatched ----
			{Code: "new (Math.pow)(a, b);"},
			{Code: "Math.pow.call(Math, a, b);"},
			{Code: "Reflect.apply(Math.pow, Math, [a, b]);"},

			// N/A: declaration/container forms do not affect this call-expression-only rule.
			// N/A: same-kind nesting is covered by shadowed receiver cases; the rule keeps no per-container traversal state.
			// N/A: overload signatures, abstract members, declare members, empty class/function bodies, and binding patterns are not inputs this rule inspects.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parentheses and TS wrappers around the receiver or callee are transparent ----
			invalidFixed("((Math)).pow(a, b)", "a**b"),
			invalidFixed("(Math as any).pow(a, b)", "a**b"),
			invalidFixed("Math!.pow(a, b)", "a**b"),
			invalidFixed("(Math satisfies any).pow(a, b)", "a**b"),
			invalidFixed("(globalThis as any).Math['pow'](a, b)", "a**b"),
			invalidFixed("(globalThis.Math as any).pow(a, b)", "a**b"),
			invalidFixed("((Math.pow as any))(a, b)", "a**b"),
			invalidFixed("(Math.pow!)(a, b)", "a**b"),
			invalidFixed("(Math['pow'] satisfies Function)(a, b)", "a**b"),
			invalidFixed("((globalThis['Math']).pow)(a, b)", "a**b"),

			// ---- Dimension 4: optional-chain receiver and optional-call wrappers still target global Math.pow ----
			invalidFixed("((Math?.pow))?.(a, b)", "a**b"),
			invalidFixed("globalThis?.Math?.['pow']?.(a, b)", "a**b"),

			// ---- Dimension 4: static element access keys include simple templates and string concatenation ----
			invalidFixed(`Math["pow"](a, b)`, "a**b"),
			invalidFixed("Math[`p${'ow'}`](a, b)", "a**b"),
			invalidFixed("Math[('pow' as const)](a, b)", "a**b"),
			invalidFixed("Math['p' + 'o' + 'w'](a, b)", "a**b"),
			invalidFixed("globalThis['Ma' + 'th'][`p${'ow'}`](a, b)", "a**b"),

			// ---- Dimension 4: type-only declarations do not shadow global Math/globalThis values ----
			invalidFixed("type Math = unknown; Math.pow(a, b);", "type Math = unknown; a**b;"),
			invalidFixed("interface Math { pow: unknown } Math.pow(a, b);", "interface Math { pow: unknown } a**b;"),
			invalidFixed("type globalThis = unknown; globalThis.Math.pow(a, b);", "type globalThis = unknown; a**b;"),

			// ---- Dimension 4: TS wrappers on inspected arguments and the whole replacement keep the resulting expression valid ----
			invalidFixed("Math.pow(a!, b!)", "a!**b!"),
			invalidFixed("Math.pow((a satisfies number), b)", "(a satisfies number)**b"),
			invalidFixed("const value = Math.pow(a, b) as number;", "const value = (a**b) as number;"),
			invalidFixed("const value = Math.pow(a, b)!;", "const value = (a**b)!;"),

			// Spread arguments are reported without an unsafe fix.
			invalidNoFix("Math.pow(...[], a)"),

			// globalThis path is independent from a local Math binding.
			invalidFixed("function f(Math) { return globalThis.Math.pow(a, b); }", "function f(Math) { return a**b; }"),

			// ---- Real-user: eslint/eslint#10482 original distance formula shape ----
			invalidFixed("const distance = Math.sqrt(Math.pow(dx, 2) + Math.pow(dy, 2));", "const distance = Math.sqrt(dx**2 + dy**2);"),

			// ---- Real-user: eslint/eslint#17173 TypeScript assertion around both operands ----
			invalidFixed("const v = Math.pow((value as number) + 1, exp as number);", "const v = ((value as number) + 1)**(exp as number);"),

			// ---- Real-user: eslint/eslint#20987 declaration-like base at expression-statement start ----
			invalidFixed("Math.pow(class Named { static x = 2 }.x, n) + scale;", "(class Named { static x = 2 }.x**n) + scale;"),

			// Locks in upstream doesBaseNeedParens() arm 1: lower-precedence base.
			invalidFixed("Math.pow(a || b, c)", "(a || b)**c"),

			// Locks in upstream doesBaseNeedParens() arm 2: unary base.
			invalidFixed("Math.pow(~a, b)", "(~a)**b"),

			// Locks in upstream doesBaseNeedParens() negative arm: update expressions are not unary expressions in ESTree.
			invalidFixed("Math.pow(++a, b)", "++a**b"),

			// Locks in upstream doesExponentNeedParens() arm 1: lower-precedence exponent.
			invalidFixed("Math.pow(a, b || c)", "a**(b || c)"),

			// Locks in upstream doesExponentNeedParens() negative arm: exponentiation is right-associative.
			invalidFixed("Math.pow(a, b ** c)", "a**b ** c"),

			// Locks in upstream doesBaseNeedParens() and doesExponentNeedParens() lower-precedence arms for conditional expressions.
			invalidFixed("const value = Math.pow(a ? b : c, d ? e : f);", "const value = (a ? b : c)**(d ? e : f);"),

			// Locks in upstream doesExponentiationExpressionNeedParens() call-argument exception.
			invalidFixed("callee(Math.pow(a, b)).prop", "callee(a**b).prop"),

			// Locks in upstream doesExponentiationExpressionNeedParens() computed-member-key exception.
			invalidFixed("object[Math.pow(a, b)].prop", "object[a**b].prop"),
			invalidFixed("const obj = { [Math.pow(a, b)]: value };", "const obj = { [a**b]: value };"),

			// Nested exponentiation fixes can require multiple fix passes to preserve associativity.
			invalidWithOutputs("const value = Math.pow(a, Math.pow(b, c));", "const value = a**Math.pow(b, c);", "const value = a**b**c;"),
			invalidWithOutputs("const value = Math.pow(Math.pow(a, b), c);", "const value = Math.pow(a, b)**c;", "const value = (a**b)**c;"),

			// Contexts where the whole replacement must be wrapped.
			invalidFixed("class C extends Math.pow(A, B) {}", "class C extends (A**B) {}"),
			invalidFixed("new (Math.pow(a, b).constructor)()", "new ((a**b).constructor)()"),

			// Fixes keep adjacent tokens from merging into a different token.
			invalidFixed("const x = a/Math.pow(/re/, b);", "const x = a/ /re/**b;"),
			invalidFixed("const x = a-Math.pow(--b, c);", "const x = a- --b**c;"),
			invalidFixed("const x = Math.pow(a, b)instanceof C;", "const x = a**b instanceof C;"),
			invalidFixed("const x = Math.pow(a, b)as number;", "const x = (a**b)as number;"),
			invalidFixed("const x = Math.pow(a, b)satisfies number;", "const x = (a**b)satisfies number;"),

			// ECMAScript line terminators beyond LF/CR still need ASI protection.
			invalidFixed("foo\u2028Math.pow({a:1}.a, 2)", "foo\u2028;({a:1}.a**2)"),
			invalidFixed("foo/*\n*/Math.pow([a, b].find(fn), c)", "foo/*\n*/;[a, b].find(fn)**c"),
			invalidFixed("foo// comment\nMath.pow(`template`, c)", "foo// comment\n;`template`**c"),

			// Locks in upstream report() no-fix arm: comments inside the call are preserved by skipping the fix.
			invalidNoFix("Math.pow(a /* base */, b)"),
		},
	)
}
