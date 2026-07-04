package prefer_numeric_literals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferNumericLiteralsExtras locks in branches and edge shapes that the upstream test suite doesn't exercise. Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.
func TestPreferNumericLiteralsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferNumericLiteralsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS-only wrappers around the callee do not match ESLint's identifier/member checks ----
			{Code: `(parseInt as Function)("11", 2);`},
			{Code: `parseInt!("11", 2);`},
			{Code: `(Number as any).parseInt("11", 2);`},
			{Code: `Number!.parseInt("11", 2);`},
			{Code: `parseInt("11" as string, 2);`},
			{Code: `parseInt("11", 2 as number);`},

			// ---- Dimension 4: dynamic and non-matching element keys are not Number.parseInt ----
			{Code: `Number[parseInt]("11", 2);`},
			{Code: `Number[0]("11", 2);`},
			{Code: `Number["notParseInt"]("11", 2);`},
			{Code: `Number["parseInt" as string]("11", 2);`},

			// ---- Dimension 4: shadowed globals remain valid across nesting boundaries ----
			{Code: `function f(parseInt) { parseInt("11", 2); }`},
			{Code: `function f(Number) { Number.parseInt("11", 2); }`},
			{Code: `const parseInt = Number.parseInt; parseInt("11", 2);`},
			{Code: `function f(Number) { Number?.parseInt("11", 2); }`},

			// ---- Dimension 4: graceful degradation for spread and missing argument shapes ----
			{Code: `parseInt(...args);`},
			{Code: `parseInt("11", ...radix);`},
			{Code: `parseInt();`},

			// Locks in upstream create() condition arm: radix number is outside the tracked map.
			{Code: `parseInt("11", 10);`},
			{Code: `Number.parseInt("11", 36);`},

			// Locks in upstream create() condition arm: first argument is static-looking but not a string/template literal.
			{Code: `parseInt(11n, 2);`},

			// N/A: object/property declaration key forms are not inspected by this call-expression rule.
			// N/A: function/class declaration containers do not create rule-owned traversal state.
			// N/A: body-absent TS declarations are not parseInt call expressions.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: deeply parenthesized callee ----
			preferNumericInvalidFixed(`((parseInt))("11", 2);`, `((parseInt))("11", 2)`, `0b11;`, "binary", "parseInt"),

			// ---- Dimension 4: parenthesized arguments are transparent, matching ESTree's flattened parentheses ----
			preferNumericInvalidFixed(`parseInt(("767"), (8));`, `parseInt(("767"), (8))`, `0o767;`, "octal", "parseInt"),

			// ---- Dimension 4: static element keys for Number.parseInt match ESLint's isSpecificMemberAccess ----
			preferNumericInvalidFixed(`Number["parseInt"]("11", 2);`, `Number["parseInt"]("11", 2)`, `0b11;`, "binary", `Number["parseInt"]`),
			preferNumericInvalidFixed("Number[`parseInt`](`767`, 8);", "Number[`parseInt`](`767`, 8)", `0o767;`, "octal", "Number[`parseInt`]"),
			preferNumericInvalidFixed(`Number?.["parseInt"]("11", 2);`, `Number?.["parseInt"]("11", 2)`, `0b11;`, "binary", `Number?.["parseInt"]`),

			// ---- Dimension 4: non-decimal numeric radix literals normalize to 2/8/16 ----
			preferNumericInvalidFixed(`parseInt("11", 0b10);`, `parseInt("11", 0b10)`, `0b11;`, "binary", "parseInt"),
			preferNumericInvalidFixed(`parseInt("767", 0o10);`, `parseInt("767", 0o10)`, `0o767;`, "octal", "parseInt"),
			preferNumericInvalidFixed(`parseInt("1F7", 0x10);`, `parseInt("1F7", 0x10)`, `0x1F7;`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`parseInt("11", 2.0);`, `parseInt("11", 2.0)`, `0b11;`, "binary", "parseInt"),
			preferNumericInvalidFixed(`parseInt("767", 8e0);`, `parseInt("767", 8e0)`, `0o767;`, "octal", "parseInt"),

			// ---- Dimension 4: multi-line report range and fix replacement ----
			preferNumericInvalidFixed("const value = Number.parseInt(\n  \"1F7\",\n  16\n);", "Number.parseInt(\n  \"1F7\",\n  16\n)", `const value = 0x1F7;`, "hexadecimal", "Number.parseInt"),

			// ---- Dimension 4: nested expression containers and multiple reports in the same file ----
			preferNumericInvalidFixed(`const value = () => parseInt("11", 2);`, `parseInt("11", 2)`, `const value = () => 0b11;`, "binary", "parseInt"),
			preferNumericInvalidFixed(`class C { field = Number.parseInt("1F", 16); }`, `Number.parseInt("1F", 16)`, `class C { field = 0x1F; }`, "hexadecimal", "Number.parseInt"),
			preferNumericInvalidFixed(`const obj = { [parseInt("11", 2)]: true };`, `parseInt("11", 2)`, `const obj = { [0b11]: true };`, "binary", "parseInt"),
			preferNumericInvalidFixed(`class C { [Number.parseInt("1F", 16)]() {} }`, `Number.parseInt("1F", 16)`, `class C { [0x1F]() {} }`, "hexadecimal", "Number.parseInt"),
			preferNumericInvalidFixedTwo(
				`foo(parseInt("11", 2), Number.parseInt("1F", 16));`,
				`parseInt("11", 2)`,
				"binary",
				"parseInt",
				`Number.parseInt("1F", 16)`,
				"hexadecimal",
				"Number.parseInt",
				`foo(0b11, 0x1F);`,
			),

			// ---- Real-user: eslint/eslint#13045 no-substitution template literals should report ----
			preferNumericInvalidFixed("Number.parseInt(`111110111`, 2) === 503;", "Number.parseInt(`111110111`, 2)", `0b111110111 === 503;`, "binary", "Number.parseInt"),

			// ---- Real-user: eslint/eslint#13568 numeric separators should report but not autofix ----
			preferNumericInvalidNoFix("Number.parseInt(`5_000`, 8);", "Number.parseInt(`5_000`, 8)", "octal", "Number.parseInt"),

			// Locks in upstream isParseInt() identifier arm: only the outer call that is not shadowed reports.
			preferNumericInvalidFixed(`parseInt("11", 2); function f(parseInt) { parseInt("11", 2); }`, `parseInt("11", 2)`, `0b11; function f(parseInt) { parseInt("11", 2); }`, "binary", "parseInt"),

			// Locks in upstream fix() prefix arm: keyword before replacement needs whitespace.
			preferNumericInvalidFixed(`function *f(){ yield((parseInt))("11", 2) }`, `((parseInt))("11", 2)`, `function *f(){ yield 0b11 }`, "binary", "parseInt"),

			// Locks in upstream fix() suffix arm: keyword after replacement needs whitespace.
			preferNumericInvalidFixed(`const ok = parseInt("11", 2)in foo;`, `parseInt("11", 2)`, `const ok = 0b11 in foo;`, "binary", "parseInt"),
		},
	)
}
