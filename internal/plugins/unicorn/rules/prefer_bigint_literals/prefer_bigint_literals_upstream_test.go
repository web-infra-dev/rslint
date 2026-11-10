// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-bigint-literals.js
package prefer_bigint_literals_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_bigint_literals"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferBigIntLiteralsUpstream(t *testing.T) {
	hundredNines := strings.Repeat("9", 100)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_bigint_literals.PreferBigIntLiteralsRule,
		[]rule_tester.ValidTestCase{
			// Numbers and strings.
			{Code: "1n", FileName: "case.js"},
			{Code: "BigInt()", FileName: "case.js"},
			{Code: "BigInt(1, 1)", FileName: "case.js"},
			{Code: "BigInt(...[1])", FileName: "case.js"},
			{Code: "BigInt(true)", FileName: "case.js"},
			{Code: "BigInt(null)", FileName: "case.js"},
			{Code: "new BigInt(1)", FileName: "case.js"},
			{Code: "Not_BigInt(1)", FileName: "case.js"},
			{Code: `BigInt("1.0")`, FileName: "case.js"},
			{Code: `BigInt("1.1")`, FileName: "case.js"},
			{Code: `BigInt("1e3")`, FileName: "case.js"},
			{Code: `BigInt("1" + "2")`, FileName: "case.js"},
			{Code: "BigInt?.(1)", FileName: "case.js"},
			{Code: "BigInt(1.1)", FileName: "case.js"},
			{Code: "typeof BigInt", FileName: "case.js"},
			{Code: "BigInt(1n)", FileName: "case.js"},
			{Code: `BigInt("not-number")`, FileName: "case.js"},
			{Code: `BigInt("1_2")`, FileName: "case.js"},
			{Code: "BigInt(!1)", FileName: "case.js"},
			{Code: "BigInt(~1)", FileName: "case.js"},
			{Code: `BigInt("++1")`, FileName: "case.js"},
			{Code: `BigInt("+ 1")`, FileName: "case.js"},
			{Code: "BigInt(void 0)", FileName: "case.js"},
			{Code: "BigInt(NaN)", FileName: "case.js"},
			{Code: "BigInt(Infinity)", FileName: "case.js"},
			{Code: "BigInt(-Infinity)", FileName: "case.js"},
			{Code: "BigInt(BigInt)", FileName: "case.js"},
			{Code: "BigInt(globalThis.x)", FileName: "case.js"},
			{Code: "BigInt(x)", FileName: "case.js"},
			{Code: `BigInt?.("1")`, FileName: "case.js"},
			// Templates and escapes.
			{Code: "BigInt(`1.1`)", FileName: "case.js"},
			{Code: "BigInt(`not-number`)", FileName: "case.js"},
			{Code: "BigInt(tag`1`)", FileName: "case.js"},
			{Code: "BigInt(`1${value}`)", FileName: "case.js"},
			// Documentation examples.
			{Code: "const bigint = 1n;", FileName: "case.js"},
			{Code: "const large = 9007199254740991n;", FileName: "case.js"},
			{Code: "const hex = 0xFFn;", FileName: "case.js"},
			{Code: "const computed = BigInt(getSomeNumber());", FileName: "case.js"},
			{Code: "const astronomicalNumber = 999999999999999999999999n;", FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			// Numbers and strings.
			fixed(`BigInt("0")`, "0n", at(1, 1, 1, 12)),
			fixed(`BigInt("  0  ")`, "0n", at(1, 1, 1, 16)),
			fixed(`BigInt("9007199254740993")`, "9007199254740993n", at(1, 1, 1, 27)),
			fixed(`BigInt("0B11")`, "0B11n", at(1, 1, 1, 15)),
			fixed(`BigInt("0O777")`, "0O777n", at(1, 1, 1, 16)),
			fixed(`BigInt("0XFe")`, "0XFen", at(1, 1, 1, 15)),
			fixed(`BigInt("`+hundredNines+`")`, hundredNines+"n", at(1, 1, 1, 111)),
			fixed("BigInt(0)", "0n", at(1, 1, 1, 10)),
			fixed("BigInt(0B11_11)", "0B11_11n", at(1, 1, 1, 16)),
			fixed("BigInt(0O777_777)", "0O777_777n", at(1, 1, 1, 18)),
			fixed("BigInt(0XFe_fE)", "0XFe_fEn", at(1, 1, 1, 16)),
			suggested("BigInt(9007199254740993)", "9007199254740992n", at(1, 1, 1, 25)),
			suggested("BigInt(0x20000000000001)", "9007199254740992n", at(1, 1, 1, 25)),
			suggested("BigInt(9_007_199_254_740_993)", "9007199254740992n", at(1, 1, 1, 30)),
			suggested("BigInt(0x20_00_00_00_00_00_01)", "9007199254740992n", at(1, 1, 1, 31)),
			suggested("BigInt(1.0)", "1n", at(1, 1, 1, 12)),
			suggested("BigInt(1e2)", "100n", at(1, 1, 1, 12)),
			suggested("BigInt(/* comment */1)", "1n", at(1, 1, 1, 23)),
			suggested("BigInt("+hundredNines+")",
				"10000000000000000159028911097599180468360808563945281389781327557747838772170381060813469985856815104n",
				at(1, 1, 1, 109)),
			fixed(`BigInt("-1")`, "(-1n)", at(1, 1, 1, 13)),
			suggested(`BigInt("+1")`, "1n", at(1, 1, 1, 13)),
			fixed("BigInt(-1)", "(-1n)", at(1, 1, 1, 11)),
			suggested("BigInt(+1)", "1n", at(1, 1, 1, 11)),
			fixed("-BigInt(-1)", "-(-1n)", at(1, 2, 1, 12)),
			fixed(`-BigInt("-1")`, "-(-1n)", at(1, 2, 1, 14)),
			fixed("-BigInt(1)", "-1n", at(1, 2, 1, 11)),
			fixed(`-BigInt("1")`, "-1n", at(1, 2, 1, 13)),
			suggested(`BigInt(" +1 ")`, "1n", at(1, 1, 1, 15)),
			fixed(`BigInt(" -1 ")`, "(-1n)", at(1, 1, 1, 15)),
			fixed("\n\t\tfoo\n\t\tBigInt(\"-1\")\n\t\t", "\n\t\tfoo\n\t\t;(-1n)\n\t\t", at(3, 3, 3, 15)),
			suggested(`-BigInt("+1")`, "-1n", at(1, 2, 1, 14)),
			suggested("-BigInt(/* comment */-1)", "-(-1n)", at(1, 2, 1, 25)),
			fixed("-(BigInt(-1))", "-(-1n)", at(1, 3, 1, 13)),
			fixed(`2n - BigInt("-1")`, "2n - (-1n)", at(1, 6, 1, 18)),
			fixed(`2n -BigInt("-1")`, "2n -(-1n)", at(1, 5, 1, 17)),
			fixed("BigInt(-1).toString()", "(-1n).toString()", at(1, 1, 1, 11)),
			fixed("functionCall(-BigInt(1))", "functionCall(-1n)", at(1, 15, 1, 24)),
			fixed("obj[BigInt(1)]", "obj[1n]", at(1, 5, 1, 14)),
			fixed("arr[BigInt(0)]", "arr[0n]", at(1, 5, 1, 14)),
			fixed("void BigInt(1)", "void 1n", at(1, 6, 1, 15)),
			fixed("typeof BigInt(1)", "typeof 1n", at(1, 8, 1, 17)),
			fixed("BigInt(1) + BigInt(2)", "1n + 2n", at(1, 1, 1, 10), at(1, 13, 1, 22)),
			fixed("BigInt(1) ** 2n", "1n ** 2n", at(1, 1, 1, 10)),
			fixed("condition ? BigInt(1) : BigInt(2)", "condition ? 1n : 2n", at(1, 13, 1, 22), at(1, 25, 1, 34)),
			fixed("+BigInt(1)", "+1n", at(1, 2, 1, 11)),
			fixed("+BigInt(-1)", "+(-1n)", at(1, 2, 1, 12)),
			fixed(`+BigInt("1")`, "+1n", at(1, 2, 1, 13)),
			suggested("BigInt(+0)", "0n", at(1, 1, 1, 11)),
			fixed("BigInt(-0)", "0n", at(1, 1, 1, 11)),
			suggested(`BigInt("+0")`, "0n", at(1, 1, 1, 13)),
			fixed(`-BigInt("-0")`, "-0n", at(1, 2, 1, 14)),
			fixed(`1n - (( BigInt("-1") ))`, "1n - (( -1n ))", at(1, 9, 1, 21)),
			suggested("BigInt(-(+(-10)))", "10n", at(1, 1, 1, 18)),
			// Templates and escapes.
			fixed("BigInt(`1`)", "1n", at(1, 1, 1, 12)),
			fixed("BigInt(``)", "0n", at(1, 1, 1, 11)),
			fixed("BigInt(`  00001  `)", "1n", at(1, 1, 1, 20)),
			fixed("BigInt(`0XFe`)", "0XFen", at(1, 1, 1, 15)),
			fixed("BigInt(`-1`).toString()", "(-1n).toString()", at(1, 1, 1, 13)),
			suggested("BigInt(`+1`)", "1n", at(1, 1, 1, 13)),
			suggested("BigInt(/* comment */ `1`)", "1n", at(1, 1, 1, 26)),
			fixed("BigInt(\"1\\\n2\")", "12n", at(1, 1, 2, 4)),
			fixed("BigInt(`\\u{31}`)", "1n", at(1, 1, 1, 17)),
			fixed(`BigInt("\u{31}")`, "1n", at(1, 1, 1, 17)),
			// Leading zeros and exponent notation.
			fixed(`BigInt("00001");`, "1n;", at(1, 1, 1, 16)),
			fixed(`BigInt("007");`, "7n;", at(1, 1, 1, 14)),
			fixed(`BigInt("-00001");`, "(-1n);", at(1, 1, 1, 17)),
			fixed(`BigInt("  00001  ");`, "1n;", at(1, 1, 1, 20)),
			fixed(`BigInt("-0");`, "0n;", at(1, 1, 1, 13)),
			fixed(`BigInt("08");`, "8n;", at(1, 1, 1, 13)),
			fixed(`BigInt("09");`, "9n;", at(1, 1, 1, 13)),
			fixed(`BigInt("080");`, "80n;", at(1, 1, 1, 14)),
			fixed(`BigInt("0000008");`, "8n;", at(1, 1, 1, 18)),
			suggested("BigInt(1e+21);", "1000000000000000000000n;", at(1, 1, 1, 14)),
			// Documentation examples.
			fixed("const bigint = BigInt(1);", "const bigint = 1n;", at(1, 16, 1, 25)),
			fixed("const large = BigInt('9007199254740991');", "const large = 9007199254740991n;", at(1, 15, 1, 41)),
			fixed("const hex = BigInt('0xFF');", "const hex = 0xFFn;", at(1, 13, 1, 27)),
		},
	)
}

func TestPreferBigIntLiteralsUpstreamLegacyOctal(t *testing.T) {
	// tsgo emits parse diagnostics for these valid script-mode literals. Use
	// the lenient lint entry point so both upstream cases still run in Go.
	checkLegacySuggestion(t, "BigInt(0777)", "511n")
	checkLegacySuggestion(t, "BigInt(0888)", "888n")
}

func at(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "prefer-bigint-literals/error",
		Message:   "Prefer using bigint literal over `BigInt(…)`.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func fixed(code, output string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "case.js", Output: []string{output}, Errors: errors}
}

func suggested(code, output string, diagnostic rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
		{MessageId: "prefer-bigint-literals/suggestion", Output: output},
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{diagnostic}}
}
