// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package numeric_separators_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/numeric_separators_style"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/numeric-separators-style.js
func TestNumericSeparatorsStyleUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Hexadecimal.
		{Code: "const foo = 0xAB_CD", FileName: "input.js"},
		{Code: "const foo = 0xAB", FileName: "input.js"},
		{Code: "const foo = 0xA", FileName: "input.js"},
		{Code: "const foo = 0xA_BC_DE_F0", FileName: "input.js"},
		{Code: "const foo = 0xab_e8_12", FileName: "input.js"},
		{Code: "const foo = 0xe", FileName: "input.js"},
		{Code: "const foo = 0Xab_e3_cd", FileName: "input.js"},
		// Octal.
		{Code: "const foo = 0o1234_5670", FileName: "input.js"},
		{Code: "const foo = 0o7777", FileName: "input.js"},
		{Code: "const foo = 0o01", FileName: "input.js"},
		{Code: "const foo = 0o12_7000_0000", FileName: "input.js"},
		{Code: "const foo = 0O1111_1111", FileName: "input.js"},
		// Binary.
		{Code: "const foo = 0b1010_0001_1000_0101", FileName: "input.js"},
		{Code: "const foo = 0b0000", FileName: "input.js"},
		{Code: "const foo = 0b10", FileName: "input.js"},
		{Code: "const foo = 0b1_0111_0101_0101", FileName: "input.js"},
		{Code: "const foo = 0B1010", FileName: "input.js"},
		// BigInt.
		{Code: "const foo = 0b1010n", FileName: "input.js"},
		{Code: "const foo = 0b1010_1010n", FileName: "input.js"},
		{Code: "const foo = 9_223_372_036_854_775_807n", FileName: "input.js"},
		{Code: "const foo = 807n", FileName: "input.js"},
		{Code: "const foo = 1n", FileName: "input.js"},
		{Code: "const foo = 9_372_854_807n", FileName: "input.js"},
		{Code: "const foo = 9807n", FileName: "input.js"},
		{Code: "const foo = 0n", FileName: "input.js"},
		// Numbers.
		{Code: "const foo = 12_345_678", FileName: "input.js"},
		{Code: "const foo = 123", FileName: "input.js"},
		{Code: "const foo = 1", FileName: "input.js"},
		{Code: "const foo = 1234", FileName: "input.js"},
		{Code: "const foo = 1_234", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, FileName: "input.js"},
		// Fractional parts.
		{Code: "const foo = 9807.123", FileName: "input.js"},
		{Code: "const foo = 3819.123432", FileName: "input.js"},
		{Code: "const foo = 138_789.12343242", FileName: "input.js"},
		{Code: "const foo = .0000001", FileName: "input.js"},
		{Code: "const foo = .00000", FileName: "input.js"},
		{Code: "const foo = 0.00000", FileName: "input.js"},
		{Code: "const foo = 0.55228474983", FileName: "input.js"},
		{Code: "const foo = 1.234_567", Options: map[string]any{"number": map[string]any{"fractionGroupLength": 3}}, FileName: "input.js"},
		{Code: "const foo = 149_597_870_700.31415_92653_58979", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3, "fractionGroupLength": 5}}, FileName: "input.js"},
		{Code: "const foo = 1.2345", Options: map[string]any{"number": map[string]any{"fractionGroupLength": 2}}, FileName: "input.js"},
		// Negative numbers.
		{Code: "const foo = -3000", FileName: "input.js"},
		{Code: "const foo = -10_000_000", FileName: "input.js"},
		// Exponential notation.
		{Code: "const foo = 1e10_000", FileName: "input.js"},
		{Code: "const foo = 39_804e1000", FileName: "input.js"},
		{Code: "const foo = -123_456e-100", FileName: "input.js"},
		{Code: "const foo = -100_000e-100_000", FileName: "input.js"},
		{Code: "const foo = -100_000e+100_000", FileName: "input.js"},
		{Code: "const foo = 3.6e12_000", FileName: "input.js"},
		{Code: "const foo = 3.6E12_000", FileName: "input.js"},
		{Code: "const foo = -1_200_000e5", FileName: "input.js"},
		// Other values.
		{Code: "const foo = -282_932 - (1938 / 10_000) * .1 + 18.1000002", FileName: "input.js"},
		{Code: "const foo = NaN", FileName: "input.js"},
		{Code: "const foo = Infinity", FileName: "input.js"},
		{Code: "const foo = \"1234567n\"", FileName: "input.js"},
		// Varying options.
		{Code: "const foo = 10000", Options: map[string]any{"number": map[string]any{"minimumDigits": 6}}, FileName: "input.js"},
		{Code: "const foo = 100_0000_0000", Options: map[string]any{"number": map[string]any{"groupLength": 4}}, FileName: "input.js"},
		{Code: "const foo = 0xA_B_C_D_E_1_2_3_4", Options: map[string]any{"hexadecimal": map[string]any{"groupLength": 1}}, FileName: "input.js"},
		{Code: "const foo = 0b111", Options: map[string]any{"number": map[string]any{"minimumDigits": 3, "groupLength": 1}}, FileName: "input.js"},
		{Code: `const binary = 0b10101010;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`, Options: map[string]any{"onlyIfContainsSeparator": true}, FileName: "input.js"},
		{Code: `const binary = 0b1010_1010;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`, Options: map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false}}, FileName: "input.js"},
		{Code: `const binary = 0b10_10_10_10;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`, Options: map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false, "groupLength": 2}}, FileName: "input.js"},
		{Code: `const binary = 0b10101010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`, Options: map[string]any{"binary": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 12345", Options: map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 12345678", Options: map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 12_345", Options: map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = -100_000e+100_000", Options: map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = -100000e+100000", Options: map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 0xA_B_C_D_E", Options: map[string]any{"hexadecimal": map[string]any{"onlyIfContainsSeparator": true, "groupLength": 1}}, FileName: "input.js"},
		{Code: "const foo = 0o7777", Options: map[string]any{"octal": map[string]any{"onlyIfContainsSeparator": true, "minimumDigits": 4}}, FileName: "input.js"},
		{Code: "const foo = 0xABCDEF012", Options: map[string]any{"hexadecimal": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 0o777777", Options: map[string]any{"octal": map[string]any{"onlyIfContainsSeparator": true, "minimumDigits": 3}}, FileName: "input.js"},
		{Code: "const foo = 0o777777", Options: map[string]any{"octal": map[string]any{"onlyIfContainsSeparator": true, "minimumDigits": 3, "groupLength": 2}}, FileName: "input.js"},
		{Code: "const foo = 0o777_777", Options: map[string]any{"octal": map[string]any{"onlyIfContainsSeparator": true, "minimumDigits": 2, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const foo = 0b01010101", Options: map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 0b0101_0101", Options: map[string]any{"onlyIfContainsSeparator": false, "binary": map[string]any{"onlyIfContainsSeparator": true}}, FileName: "input.js"},
		{Code: "const foo = 0b0101_0101", Options: map[string]any{"onlyIfContainsSeparator": false, "binary": map[string]any{"onlyIfContainsSeparator": false}}, FileName: "input.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Hexadecimal.
		invalidNumericCase("const foo = 0xA_B_CDE_F0", "const foo = 0xA_BC_DE_F0", nil, numericError(1, 13, 25)),
		invalidNumericCase("const foo = 0xABCDEF", "const foo = 0xAB_CD_EF", nil, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 0xA_B", "const foo = 0xAB", nil, numericError(1, 13, 18)),
		invalidNumericCase("const foo = 0XAB_C_D", "const foo = 0XAB_CD", nil, numericError(1, 13, 21)),
		// Octal.
		invalidNumericCase("const foo = 0o12_34_5670", "const foo = 0o1234_5670", nil, numericError(1, 13, 25)),
		invalidNumericCase("const foo = 0o7_7_77", "const foo = 0o7777", nil, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 0o010101010101", "const foo = 0o0101_0101_0101", nil, numericError(1, 13, 27)),
		invalidNumericCase("const foo = 0O010101010101", "const foo = 0O0101_0101_0101", nil, numericError(1, 13, 27)),
		// Binary.
		invalidNumericCase("const foo = 0b10_10_0001", "const foo = 0b1010_0001", nil, numericError(1, 13, 25)),
		invalidNumericCase("const foo = 0b0_00_0", "const foo = 0b0000", nil, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 0b10101010101010", "const foo = 0b10_1010_1010_1010", nil, numericError(1, 13, 29)),
		invalidNumericCase("const foo = 0B10101010101010", "const foo = 0B10_1010_1010_1010", nil, numericError(1, 13, 29)),
		// BigInt.
		invalidNumericCase("const foo = 0b10101010n", "const foo = 0b1010_1010n", nil, numericError(1, 13, 24)),
		invalidNumericCase("const foo = 1_9_223n", "const foo = 19_223n", nil, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 80_7n", "const foo = 807n", nil, numericError(1, 13, 18)),
		invalidNumericCase("const foo = 123456789_100n", "const foo = 123_456_789_100n", nil, numericError(1, 13, 27)),
		// Numbers.
		invalidNumericCase("const foo = 1_2_345_678", "const foo = 12_345_678", nil, numericError(1, 13, 24)),
		invalidNumericCase("const foo = 12_3", "const foo = 123", nil, numericError(1, 13, 17)),
		invalidNumericCase("const foo = 1234567890", "const foo = 1_234_567_890", nil, numericError(1, 13, 23)),
		// Fractional parts and numeric member access.
		invalidNumericCase("const foo = 0.552_284_749_83", "const foo = 0.55228474983", nil, numericError(1, 13, 29)),
		invalidNumericCase("const foo = 3819.123_4325", "const foo = 3819.1234325", nil, numericError(1, 13, 26)),
		invalidNumericCase("const foo = 138789.12343_2_42", "const foo = 138_789.12343242", nil, numericError(1, 13, 30)),
		invalidNumericCase("const foo = .000000_1", "const foo = .0000001", nil, numericError(1, 13, 22)),
		invalidNumericCase("const foo = 12345678..toString()", "const foo = 12_345_678..toString()", nil, numericError(1, 13, 22)),
		invalidNumericCase("const foo = 12345678 .toString()", "const foo = 12_345_678 .toString()", nil, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 1.234567", "const foo = 1.234_567", map[string]any{"number": map[string]any{"fractionGroupLength": 3}}, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 1789.123_432_42", "const foo = 1789.12343242", map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, numericError(1, 13, 28)),
		invalidNumericCase("const foo = -282_932 - (1938 / 10_000) * .1 + 18.100_000_2", "const foo = -282_932 - (1938 / 10_000) * .1 + 18.1000002", map[string]any{"number": map[string]any{"onlyIfContainsSeparator": true}}, numericError(1, 47, 59)),
		// Negative numbers.
		invalidNumericCase("const foo = -100000_1", "const foo = -1_000_001", nil, numericError(1, 14, 22)),
		// Exponential notation.
		invalidNumericCase("const foo = 1e10000", "const foo = 1e10_000", nil, numericError(1, 13, 20)),
		invalidNumericCase("const foo = 39804e10000", "const foo = 39_804e10_000", nil, numericError(1, 13, 24)),
		invalidNumericCase("const foo = -123456e100", "const foo = -123_456e100", nil, numericError(1, 14, 24)),
		invalidNumericCase("const foo = -100000e-10000", "const foo = -100_000e-10_000", nil, numericError(1, 14, 27)),
		invalidNumericCase("const foo = -1000e+10000", "const foo = -1000e+10_000", nil, numericError(1, 14, 25)),
		invalidNumericCase("const foo = -1000e+00010000", "const foo = -1000e+00_010_000", nil, numericError(1, 14, 28)),
		invalidNumericCase("const foo = 3.6e12000", "const foo = 3.6e12_000", nil, numericError(1, 13, 22)),
		invalidNumericCase("const foo = -1200000e5", "const foo = -1_200_000e5", nil, numericError(1, 14, 23)),
		invalidNumericCase("const foo = 3.65432E12000", "const foo = 3.65432E12_000", nil, numericError(1, 13, 26)),
		// Varying options.
		invalidNumericCase("const foo = 1000000", "const foo = 1_000_000", map[string]any{"number": map[string]any{"minimumDigits": 6}}, numericError(1, 13, 20)),
		invalidNumericCase("const foo = 10_000_000_000", "const foo = 100_0000_0000", map[string]any{"number": map[string]any{"groupLength": 4}}, numericError(1, 13, 27)),
		invalidNumericCase("const foo = 0xA_B_CD", "const foo = 0xA_B_C_D", map[string]any{"hexadecimal": map[string]any{"groupLength": 1}}, numericError(1, 13, 21)),
		invalidNumericCase("const foo = 0b1_11", "const foo = 0b111", map[string]any{"number": map[string]any{"minimumDigits": 3, "groupLength": 2}}, numericError(1, 13, 19)),
		invalidNumericCase("const foo = -100000e+100000", "const foo = -100_000e+100_000", map[string]any{"number": map[string]any{"onlyIfContainsSeparator": false}}, numericError(1, 14, 28)),
		invalidNumericCase(`const binary = 0b10_101010;
const octal = 0o76_543210;
const hexadecimal = 0xfe_dcba97;
const number = 12_345678.12345678e12345678;`, `const binary = 0b1010_1010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`, map[string]any{"onlyIfContainsSeparator": true}, numericError(1, 16, 27), numericError(2, 15, 26), numericError(3, 21, 32), numericError(4, 16, 43)),
		invalidNumericCase(`const binary = 0b10101010;
const octal = 0o76_543210;
const hexadecimal = 0xfe_dcba97;
const number = 12_345678.12345678e12345678;`, `const binary = 0b1010_1010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`, map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false}}, numericError(1, 16, 26), numericError(2, 15, 26), numericError(3, 21, 32), numericError(4, 16, 43)),
		invalidNumericCase(`const binary = 0b10101010;
const octal = 0o76_543210;
const hexadecimal = 0xfe_dcba97;
const number = 12_345678.12345678e12345678;`, `const binary = 0b10_10_10_10;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`, map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false, "groupLength": 2}}, numericError(1, 16, 26), numericError(2, 15, 26), numericError(3, 21, 32), numericError(4, 16, 43)),
		invalidNumericCase(`const binary = 0b10_101010;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`, `const binary = 0b1010_1010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`, map[string]any{"binary": map[string]any{"onlyIfContainsSeparator": true}}, numericError(1, 16, 27), numericError(2, 15, 25), numericError(3, 21, 31), numericError(4, 16, 42)),
		// Snapshot cases.
		invalidNumericCase("console.log(0XdeEdBeeFn)", "console.log(0Xde_Ed_Be_eFn)", nil, numericError(1, 13, 24)),
		invalidNumericCase("const foo = 12345678..toString()", "const foo = 12_345_678..toString()", nil, numericError(1, 13, 22)),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &numeric_separators_style.NumericSeparatorsStyleRule, valid, invalid)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/numeric-separators-style.md
func TestNumericSeparatorsStyleUpstreamDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "const foo = 1_234_444;", FileName: "input.js"},
		{Code: "const foo = 1234.56789;", FileName: "input.js"},
		{Code: "const foo = 0xAB_CD_EF;", FileName: "input.js"},
		{Code: "const foo = 0b1000_1111;", FileName: "input.js"},
		{Code: "const foo = 0o10_4421;", FileName: "input.js"},
		{Code: "const foo = 1_294_287_712n;", FileName: "input.js"},
		{Code: "const number = 100000; // Pass, this number does not contain separators", Options: map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false}}, FileName: "input.js"},
		{Code: "const groupedInteger = 12_345;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const fractional = 0.0000001;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const groupedFractional = 0.000_000_1;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3, "fractionGroupLength": 3}}, FileName: "input.js"},
		{Code: "const groupedLongFractional = 123.100_000_1;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3, "fractionGroupLength": 3}}, FileName: "input.js"},
		{Code: "const groupedBinary = 0b10_1010;", Options: map[string]any{"binary": map[string]any{"minimumDigits": 0, "groupLength": 4}}, FileName: "input.js"},
		{Code: "const groupedLongBinary = 0b1_0101_0001;", Options: map[string]any{"binary": map[string]any{"minimumDigits": 0, "groupLength": 4}}, FileName: "input.js"},
		{Code: "const groupedHexadecimal = 0xAB_CD_EF;", Options: map[string]any{"hexadecimal": map[string]any{"minimumDigits": 0, "groupLength": 2}}, FileName: "input.js"},
		{Code: "const smallNumber = 100;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const groupedNumber = 1_000;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const largeGroupedNumber = 1_000_000;", Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const foo = 1000;", Options: map[string]any{"number": map[string]any{"minimumDigits": 5, "groupLength": 3}}, FileName: "input.js"},
		{Code: "const octal = 0o7777;", Options: map[string]any{"octal": map[string]any{"minimumDigits": 0, "groupLength": 4}}, FileName: "input.js"},
		{Code: "const anotherOctal = 0o7777;", Options: map[string]any{"octal": map[string]any{"minimumDigits": 0, "groupLength": 4}}, FileName: "input.js"},
		{Code: "const groupedOctal = 0o12_7777;", Options: map[string]any{"octal": map[string]any{"minimumDigits": 0, "groupLength": 4}}, FileName: "input.js"},
		{Code: "const value = 1024;", Options: map[string]any{"number": map[string]any{"minimumDigits": 5}}, FileName: "input.js"},
		{Code: "const pi = 3.14159_26535_89793;", Options: map[string]any{"number": map[string]any{"fractionGroupLength": 5}}, FileName: "input.js"},
		{Code: "const values = [12_345, 0xAB_CD, 0b1010_1010, 0o1234_5670, 0.5522847498];", Options: map[string]any{"onlyIfContainsSeparator": false, "hexadecimal": map[string]any{"minimumDigits": 0, "groupLength": 2}, "binary": map[string]any{"minimumDigits": 0, "groupLength": 4}, "octal": map[string]any{"minimumDigits": 0, "groupLength": 4}, "number": map[string]any{"minimumDigits": 5, "groupLength": 3}}, FileName: "input.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		invalidNumericCase("const foo = 1_23_4444;", "const foo = 1_234_444;", nil, numericError(1, 13, 22)),
		invalidNumericCase("const foo = 1_234.567_89;", "const foo = 1234.56789;", nil, numericError(1, 13, 25)),
		invalidNumericCase("const foo = 0xAB_C_D_EF;", "const foo = 0xAB_CD_EF;", nil, numericError(1, 13, 24)),
		invalidNumericCase("const foo = 0b10_00_1111;", "const foo = 0b1000_1111;", nil, numericError(1, 13, 25)),
		invalidNumericCase("const foo = 0o1_0_44_21;", "const foo = 0o10_4421;", nil, numericError(1, 13, 24)),
		invalidNumericCase("const foo = 1_294_28771_2n;", "const foo = 1_294_287_712n;", nil, numericError(1, 13, 27)),
		invalidNumericCase("const binary = 0b101010001; // Fail, `binary` type don't require separators", "const binary = 0b1_0101_0001; // Fail, `binary` type don't require separators", map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false}}, numericError(1, 16, 27)),
		invalidNumericCase("const hexadecimal = 0xD_EED_BEE_F; // Fail, it contains separators and it's incorrectly grouped", "const hexadecimal = 0xDE_ED_BE_EF; // Fail, it contains separators and it's incorrectly grouped", map[string]any{"onlyIfContainsSeparator": true, "binary": map[string]any{"onlyIfContainsSeparator": false}}, numericError(1, 21, 34)),
		invalidNumericCase("const integer = 12345;", "const integer = 12_345;", map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3}}, numericError(1, 17, 22)),
		invalidNumericCase("const fractional = 0.000_0001;", "const fractional = 0.000_000_1;", map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3, "fractionGroupLength": 3}}, numericError(1, 20, 30)),
		invalidNumericCase("const longFractional = 123.1_000_001;", "const longFractional = 123.100_000_1;", map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 3, "fractionGroupLength": 3}}, numericError(1, 24, 37)),
		invalidNumericCase("const binary = 0b101010;", "const binary = 0b10_1010;", map[string]any{"binary": map[string]any{"minimumDigits": 0, "groupLength": 4}}, numericError(1, 16, 24)),
		invalidNumericCase("const longBinary = 0b1010_10001;", "const longBinary = 0b1_0101_0001;", map[string]any{"binary": map[string]any{"minimumDigits": 0, "groupLength": 4}}, numericError(1, 20, 32)),
		invalidNumericCase("const hexadecimal = 0xA_B_CD_EF;", "const hexadecimal = 0xAB_CD_EF;", map[string]any{"hexadecimal": map[string]any{"minimumDigits": 0, "groupLength": 2}}, numericError(1, 21, 32)),
		invalidNumericCase("const value = 1_024;", "const value = 1024;", map[string]any{"number": map[string]any{"minimumDigits": 5}}, numericError(1, 15, 20)),
		invalidNumericCase("const value = 0.5522847498;", "const value = 0.55228_47498;", map[string]any{"number": map[string]any{"fractionGroupLength": 5}}, numericError(1, 15, 27)),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &numeric_separators_style.NumericSeparatorsStyleRule, valid, invalid)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/numeric-separators-style.js
func TestNumericSeparatorsStyleUpstreamOtherLanguages(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// TOML requires an ESLint language plugin.
		{Code: "value = 1_000_000", FileName: "example.toml", Skip: true},
		{Code: "value = [-1234, +1234, -12_345, +12_345]", FileName: "example.toml", Skip: true},
		{Code: "value = [inf, +inf, -inf, nan, +nan, -nan]", FileName: "example.toml", Skip: true},
		{Code: "value = [\"1234567\", 1979-05-27T07:32:00Z, 1979-05-27, 07:32:00]", FileName: "example.toml", Skip: true},
		{Code: "value = [0xAB_CD, 0o1234_5670, 0b1010_1010]", FileName: "example.toml", Skip: true},
		{Code: "value = [1234567, 0xABCDEF]", Options: map[string]any{"onlyIfContainsSeparator": true}, FileName: "example.toml", Skip: true},
		{Code: "value = [-12345, +12345]", Options: map[string]any{"number": map[string]any{"minimumDigits": 6}}, FileName: "example.toml", Skip: true},
		// SOML requires an ESLint language plugin.
		{Code: "value: [1_000_000, 0xAB_CD, 0b1010_1010, 0o1234_5670]", FileName: "input.soml", Skip: true},
		{Code: "value: [infinity, -infinity]", FileName: "input.soml", Skip: true},
		{Code: "value: ['1234567', 2026-09-19T14:00:00Z, 12345s]", FileName: "input.soml", Skip: true},
		{Code: "value: 1234567", Options: map[string]any{"onlyIfContainsSeparator": true}, FileName: "input.soml", Skip: true},
	}
	invalid := []rule_tester.InvalidTestCase{
		// TOML requires an ESLint language plugin.
		{Code: "value = 1000000 # Preserve this comment.", FileName: "example.toml", Skip: true, Output: []string{"value = 1_000_000 # Preserve this comment."}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 9, 16)}},
		{Code: "value = [-123456, +123456]", FileName: "example.toml", Skip: true, Output: []string{"value = [-123_456, +123_456]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 10, 17), numericError(1, 19, 26)}},
		{Code: "value = 9223372036854775807", FileName: "example.toml", Skip: true, Output: []string{"value = 9_223_372_036_854_775_807"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 9, 28)}},
		{Code: "value = [0xABCDEF, 0o12345670, 0b10101010]", FileName: "example.toml", Skip: true, Output: []string{"value = [0xAB_CD_EF, 0o1234_5670, 0b1010_1010]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 10, 18), numericError(1, 20, 30), numericError(1, 32, 42)}},
		{Code: "value = {nested = [12345, 1_2_345, 0.123_456]}", FileName: "example.toml", Skip: true, Output: []string{"value = {nested = [12_345, 12_345, 0.123456]}"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 20, 25), numericError(1, 27, 34), numericError(1, 36, 45)}},
		{Code: "value = [-12345.123456E+10000, +12345.123456e-10000]", FileName: "example.toml", Skip: true, Output: []string{"value = [-12_345.123456E+10_000, +12_345.123456e-10_000]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 10, 30), numericError(1, 32, 52)}},
		{Code: "value = -1234567.12345678", FileName: "example.toml", Skip: true, Options: map[string]any{"number": map[string]any{"groupLength": 4, "fractionGroupLength": 3}}, Output: []string{"value = -123_4567.123_456_78"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 9, 26)}},
		{Code: "value = [+123, -123]", FileName: "example.toml", Skip: true, Options: map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 1}}, Output: []string{"value = [+1_2_3, -1_2_3]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 10, 14), numericError(1, 16, 20)}},
		{Code: "value = [1234567, 12_34567, 0xABCDEF]", FileName: "example.toml", Skip: true, Options: map[string]any{"onlyIfContainsSeparator": true, "hexadecimal": map[string]any{"onlyIfContainsSeparator": false, "groupLength": 3}}, Output: []string{"value = [1234567, 1_234_567, 0xABC_DEF]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 19, 27), numericError(1, 29, 37)}},
		// SOML requires an ESLint language plugin.
		{Code: "value: 1000000 # keep", FileName: "input.soml", Skip: true, Output: []string{"value: 1_000_000 # keep"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 8, 15)}},
		{Code: "value: [-123456, 9223372036854775807]", FileName: "input.soml", Skip: true, Output: []string{"value: [-123_456, 9_223_372_036_854_775_807]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 9, 16), numericError(1, 18, 37)}},
		{Code: "value: [0xABCDEF, 0b10101010, 0o12345670]", FileName: "input.soml", Skip: true, Output: []string{"value: [0xAB_CD_EF, 0b1010_1010, 0o1234_5670]"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 9, 17), numericError(1, 19, 29), numericError(1, 31, 41)}},
		{Code: "value: 1_2_345.123_456e-12", FileName: "input.soml", Skip: true, Output: []string{"value: 12_345.123456e-12"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 8, 27)}},
		{Code: "value: -1234567.12345678", FileName: "input.soml", Skip: true, Options: map[string]any{"number": map[string]any{"groupLength": 4, "fractionGroupLength": 3}}, Output: []string{"value: -123_4567.123_456_78"}, Errors: []rule_tester.InvalidTestCaseError{numericError(1, 8, 25)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &numeric_separators_style.NumericSeparatorsStyleRule, valid, invalid)
}

func TestNumericSeparatorsStyleUpstreamLegacyIntegers(t *testing.T) {
	t.Parallel()
	// tsgo reports parse errors for these sloppy-mode inputs. Run the
	// listener on the recovered literal to check that it stays silent.
	for _, code := range []string{
		"const foo = 0777777",
		"var foo = 0999999",
		"let foo = 0111222",
	} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			if diagnostics := lintNumericLiterals(t, code, rule.EditDemandAll); len(diagnostics) != 0 {
				t.Fatalf("unexpected numeric separator diagnostics: %#v", diagnostics)
			}
		})
	}
}

func invalidNumericCase(code, output string, options any, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "input.js", Output: []string{output}, Options: options, Errors: errors}
}

func numericError(line, column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "numeric-separators-style", Message: "Invalid group length in numeric value.",
		Line: line, Column: column, EndLine: line, EndColumn: endColumn,
	}
}
