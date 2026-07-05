package no_loss_of_precision

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noLossOfPrecisionText = "This number literal will lose precision at runtime."

// TestNoLossOfPrecisionUpstream migrates the full valid/invalid suite from upstream tests/lib/rules/no-loss-of-precision.js 1:1. Position assertions cover line/column for every invalid case. rslint-specific lock-in cases live in the no_loss_of_precision_extras_test.go file.
func TestNoLossOfPrecisionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLossOfPrecisionRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint core valid cases ----
			{Code: `var x = 12345`},
			{Code: `var x = 123.456`},
			{Code: `var x = -123.456`},
			{Code: `var x = -123456`},
			{Code: `var x = 123e34`},
			{Code: `var x = 123.0e34`},
			{Code: `var x = 123e-34`},
			{Code: `var x = -123e34`},
			{Code: `var x = -123e-34`},
			{Code: `var x = 12.3e34`},
			{Code: `var x = 12.3e-34`},
			{Code: `var x = -12.3e34`},
			{Code: `var x = -12.3e-34`},
			{Code: `var x = 12300000000000000000000000`},
			{Code: `var x = -12300000000000000000000000`},
			{Code: `var x = 0.00000000000000000000000123`},
			{Code: `var x = -0.00000000000000000000000123`},
			{Code: `var x = 9007199254740991`},
			{Code: `var x = 0`},
			{Code: `var x = 0.0`},
			{Code: `var x = 0.000000000000000000000000000000000000000000000000000000000000000000000000000000`},
			{Code: `var x = -0`},
			{Code: `var x = 123.0000000000000000000000`},

			// ---- ESLint core valid cases: eslint/eslint#19957 ----
			{Code: `var x = 9.00e2`},
			{Code: `var x = 9.000e3`},
			{Code: `var x = 9.0000000000e10`},
			{Code: `var x = 9.00E2`},
			{Code: `var x = 9.000E3`},
			{Code: `var x = 9.100E3`},
			{Code: `var x = 9.0000000000E10`},
			{Code: `var x = 019.5`, Skip: true}, // SKIP: TypeScript parser rejects leading-zero decimals before rule execution.
			{Code: `var x = 0195`, Skip: true},  // SKIP: TypeScript parser rejects leading-zero decimals before rule execution.
			{Code: `var x = 00195`, Skip: true}, // SKIP: TypeScript parser rejects leading-zero decimals before rule execution.
			{Code: `var x = 0008`, Skip: true},  // SKIP: TypeScript parser rejects leading-zero decimals before rule execution.
			{Code: `var x = 0e5`},
			{Code: `var x = .42`},
			{Code: `var x = 42.`},

			// ---- ESLint core valid cases: numeric separators ----
			{Code: `var x = 12_34_56`},
			{Code: `var x = 12_3.4_56`},
			{Code: `var x = -12_3.4_56`},
			{Code: `var x = -12_34_56`},
			{Code: `var x = 12_3e3_4`},
			{Code: `var x = 123.0e3_4`},
			{Code: `var x = 12_3e-3_4`},
			{Code: `var x = 12_3.0e-3_4`},
			{Code: `var x = -1_23e-3_4`},
			{Code: `var x = -1_23.8e-3_4`},
			{Code: `var x = 1_230000000_00000000_00000_000`},
			{Code: `var x = -1_230000000_00000000_00000_000`},
			{Code: `var x = 0.0_00_000000000_000000000_00123`},
			{Code: `var x = -0.0_00_000000000_000000000_00123`},
			{Code: `var x = 0e5_3`},

			// ---- ESLint core valid cases: non-decimal literals ----
			{Code: `var x = 0b11111111111111111111111111111111111111111111111111111`},
			{Code: `var x = 0b111_111_111_111_1111_11111_111_11111_1111111111_11111111_111_111`},
			{Code: `var x = 0B11111111111111111111111111111111111111111111111111111`},
			{Code: `var x = 0B111_111_111_111_1111_11111_111_11111_1111111111_11111111_111_111`},
			{Code: `var x = 0o377777777777777777`},
			{Code: `var x = 0o3_77_777_777_777_777_777`},
			{Code: `var x = 0O377777777777777777`},
			{Code: `var x = 0377777777777777777`, Skip: true}, // SKIP: TypeScript parser rejects legacy octal literals before rule execution.
			{Code: `var x = 0x1FFFFFFFFFFFFF`},
			{Code: `var x = 0X1FFFFFFFFFFFFF`},

			// ---- ESLint core valid cases: non-number literals ----
			{Code: `var x = true`},
			{Code: `var x = 'abc'`},
			{Code: `var x = ''`},
			{Code: `var x = null`},
			{Code: `var x = undefined`},
			{Code: `var x = {}`},
			{Code: `var x = ['a', 'b']`},
			{Code: `var x = new Date()`},
			{Code: `var x = '9007199254740993'`},
			{Code: `var x = 0x1FFF_FFFF_FFF_FFF`},
			{Code: `var x = 0X1_FFF_FFFF_FFF_FFF`},

			// ---- TypeScript-ESLint deprecated wrapper valid cases ----
			{Code: `const x = 12345;`},
			{Code: `const x = 123.456;`},
			{Code: `const x = -123.456;`},
			{Code: `const x = 123_456;`},
			{Code: `const x = 123_00_000_000_000_000_000_000_000;`},
			{Code: `const x = 123.000_000_000_000_000_000_000_0;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint core invalid cases: decimal literals ----
			noLossInvalid(`var x = 9007199254740993`, `9007199254740993`),
			noLossInvalid(`var x = 9007199254740.993e3`, `9007199254740.993e3`),
			noLossInvalid(`var x = 9.007199254740993e15`, `9.007199254740993e15`),
			noLossInvalid(`var x = -9007199254740993`, `9007199254740993`),
			noLossInvalid(`var x = 900719.9254740994`, `900719.9254740994`),
			noLossInvalid(`var x = -900719.9254740994`, `900719.9254740994`),
			noLossInvalid(`var x = 900719925474099_3`, `900719925474099_3`),
			noLossInvalid(`var x = 90_0719925_4740.9_93e3`, `90_0719925_4740.9_93e3`),
			noLossInvalid(`var x = 9.0_0719925_474099_3e15`, `9.0_0719925_474099_3e15`),
			noLossInvalid(`var x = 90071992547409930e-1`, `90071992547409930e-1`),
			noLossInvalid(`var x = .9007199254740993e16`, `.9007199254740993e16`),
			noLossInvalid(`var x = 900719925474099.30e1`, `900719925474099.30e1`),
			noLossInvalid(`var x = -9_00719_9254_740993`, `9_00719_9254_740993`),
			noLossInvalid(`var x = 900_719.92_54740_994`, `900_719.92_54740_994`),
			noLossInvalid(`var x = -900_719.92_5474_0994`, `900_719.92_5474_0994`),
			noLossInvalid(`var x = 5123000000000000000000000000001`, `5123000000000000000000000000001`),
			noLossInvalid(`var x = -5123000000000000000000000000001`, `5123000000000000000000000000001`),
			noLossInvalid(`var x = 1230000000000000000000000.0`, `1230000000000000000000000.0`),
			noLossInvalid(`var x = 1.0000000000000000000000123`, `1.0000000000000000000000123`),
			noLossInvalid(`var x = 17498005798264095394980017816940970922825355447145699491406164851279623993595007385788105416184430592`, `17498005798264095394980017816940970922825355447145699491406164851279623993595007385788105416184430592`),
			noLossInvalid(`var x = 2e999`, `2e999`),
			noLossInvalid(`var x = .1230000000000000000000000`, `.1230000000000000000000000`),

			// ---- ESLint core invalid cases: non-decimal literals ----
			noLossInvalid(`var x = 0b100000000000000000000000000000000000000000000000000001`, `0b100000000000000000000000000000000000000000000000000001`),
			noLossInvalid(`var x = 0B100000000000000000000000000000000000000000000000000001`, `0B100000000000000000000000000000000000000000000000000001`),
			noLossInvalid(`var x = 0o400000000000000001`, `0o400000000000000001`),
			noLossInvalid(`var x = 0O400000000000000001`, `0O400000000000000001`),
			noLossSkippedInvalid(`var x = 0400000000000000001`, `0400000000000000001`), // SKIP: TypeScript parser rejects legacy octal literals before rule execution.
			noLossInvalid(`var x = 0x20000000000001`, `0x20000000000001`),
			noLossInvalid(`var x = 0X20000000000001`, `0X20000000000001`),
			noLossInvalid(`var x = 5123_00000000000000000000000000_1`, `5123_00000000000000000000000000_1`),
			noLossInvalid(`var x = -5_12300000000000000000000_0000001`, `5_12300000000000000000000_0000001`),
			noLossInvalid(`var x = 123_00000000000000000000_00.0_0`, `123_00000000000000000000_00.0_0`),
			noLossInvalid(`var x = 1.0_00000000000000000_0000123`, `1.0_00000000000000000_0000123`),
			noLossInvalid(`var x = 174_980057982_640953949800178169_409709228253554471456994_914061648512796239935950073857881054_1618443059_2`, `174_980057982_640953949800178169_409709228253554471456994_914061648512796239935950073857881054_1618443059_2`),
			noLossInvalid(`var x = 2e9_99`, `2e9_99`),
			noLossInvalid(`var x = .1_23000000000000_00000_0000_0`, `.1_23000000000000_00000_0000_0`),
			noLossInvalid(`var x = 0b1_0000000000000000000000000000000000000000000000000000_1`, `0b1_0000000000000000000000000000000000000000000000000000_1`),
			noLossInvalid(`var x = 0B10000000000_0000000000000000000000000000_000000000000001`, `0B10000000000_0000000000000000000000000000_000000000000001`),
			noLossInvalid(`var x = 0o4_00000000000000_001`, `0o4_00000000000000_001`),
			noLossInvalid(`var x = 0O4_0000000000000000_1`, `0O4_0000000000000000_1`),
			noLossInvalid(`var x = 0x2_0000000000001`, `0x2_0000000000001`),
			noLossInvalid(`var x = 0X200000_0000000_1`, `0X200000_0000000_1`),

			// ---- TypeScript-ESLint deprecated wrapper invalid cases ----
			noLossInvalid(`const x = 9007199254740993;`, `9007199254740993`),
			noLossInvalid(`const x = 9_007_199_254_740_993;`, `9_007_199_254_740_993`),
			noLossInvalid(`const x = 9_007_199_254_740.993e3;`, `9_007_199_254_740.993e3`),
			noLossInvalid(`const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;`, `0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001`),
		},
	)
}

func noLossInvalid(code string, target string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			noLossError(code, target),
		},
	}
}

func noLossInvalidWithTsx(code string, target string) rule_tester.InvalidTestCase {
	testCase := noLossInvalid(code, target)
	testCase.Tsx = true
	return testCase
}

func noLossSkippedInvalid(code string, target string) rule_tester.InvalidTestCase {
	testCase := noLossInvalid(code, target)
	testCase.Skip = true
	return testCase
}

func noLossInvalidWithTargets(code string, targets ...string) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(targets))
	for _, target := range targets {
		errors = append(errors, noLossError(code, target))
	}
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: errors,
	}
}

func noLossError(code string, target string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, target)
	if start < 0 {
		panic("target not found in no-loss-of-precision test: " + target)
	}
	end := start + len(target)
	line, column := noLossUTF16LineColumn(code, start)
	endLine, endColumn := noLossUTF16LineColumn(code, end)
	return rule_tester.InvalidTestCaseError{
		MessageId: "noLossOfPrecision",
		Message:   noLossOfPrecisionText,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func noLossUTF16LineColumn(text string, offset int) (int, int) {
	line := 1
	column := 1
	for i, r := range text {
		if i >= offset {
			break
		}
		if r == '\n' {
			line++
			column = 1
			continue
		}
		if r >= 0x10000 {
			column += 2
		} else {
			column++
		}
	}
	return line, column
}
