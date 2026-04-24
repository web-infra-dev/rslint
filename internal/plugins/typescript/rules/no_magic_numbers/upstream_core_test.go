package no_magic_numbers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestUpstreamCoreESLintValid mirrors valid test cases from ESLint core's
// no-magic-numbers test file that are not already covered in the main test.
func TestUpstreamCoreESLintValid(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// ---- Binary / octal literal array indexes ----
		{Code: `foo[0b110]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[0o71]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
	}, nil)
}

// TestUpstreamCoreESLintInvalid mirrors invalid test cases from ESLint core's
// no-magic-numbers test file that are not already covered in the main test.
func TestUpstreamCoreESLintInvalid(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, nil, []rule_tester.InvalidTestCase{
		// ---- Negative non-integer array indexes (ignoreArrayIndexes doesn't help) ----
		{
			Code:    `foo[-0.1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0.1."}},
		},
		// ---- Negative binary array index ----
		{
			Code:    `foo[-0b110]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0b110."}},
		},
		// ---- Negative octal array index ----
		{
			Code:    `foo[-0o71]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0o71."}},
		},
		// ---- Negative hex array index ----
		{
			Code:    `foo[-0x12]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x12."}},
		},
		// ---- Non-integer from exponent (0.12e1 = 1.2) ----
		{
			Code:    `foo[0.12e1]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 0.12e1."}},
		},
		// ---- Non-integer from exponent (1.678e2 = 167.8) ----
		{
			Code:    `foo[1.678e2]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.678e2."}},
		},
		// ---- Non-integer (100.9) ----
		{
			Code:    `foo[100.9]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 100.9."}},
		},
		// ---- Above max index (1e300 coerces to "1e+300") ----
		{
			Code:    `foo[1e300]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e300."}},
		},
		// ---- Infinity (1e310) ----
		{
			Code:    `foo[1e310]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1e310."}},
		},
		// ---- Negative Infinity (-1e310) ----
		{
			Code:    `foo[-1e310]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1e310."}},
		},
		// ---- Negative hex BigInt index ----
		{
			Code:    `foo[-0x12n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x12n."}},
		},

		// ---- Default values: no options at all ----
		{
			Code:   `const { param = 123 } = sourceObject;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123."}},
		},
		// ---- Default values: empty options object ----
		{
			Code:    `const { param = 123 } = sourceObject;`,
			Options: map[string]interface{}{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 123."}},
		},
		// ---- Array destructuring default: ignoreDefaultValues: false ----
		{
			Code:    `const [one = 1, two = 2] = []`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1."},
				{MessageId: "noMagic", Message: "No magic number: 2."},
			},
		},
		// ---- Destructuring assignment default: ignoreDefaultValues: false ----
		{
			Code:    `var one, two; [one = 1, two = 2] = []`,
			Options: map[string]interface{}{"ignoreDefaultValues": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1."},
				{MessageId: "noMagic", Message: "No magic number: 2."},
			},
		},

		// ---- Array index: no options ----
		{
			Code:   `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3."}},
		},
		// ---- Array index: empty options ----
		{
			Code:    `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Options: map[string]interface{}{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3."}},
		},
		// ---- Array index: ignoreArrayIndexes explicitly false ----
		{
			Code:    `var data = ['foo', 'bar', 'baz']; var third = data[3];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 3."}},
		},

		// ---- ignoreClassFieldInitialValues: static ----
		{
			Code:    `class C { static foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 24}},
		},
		// ---- ignoreClassFieldInitialValues: private ----
		{
			Code:    `class C { #foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 18}},
		},
		// ---- ignoreClassFieldInitialValues: static private ----
		{
			Code:    `class C { static #foo = 2; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 25}},
		},
		// ---- ignoreClassFieldInitialValues: empty options ----
		{
			Code:    `class C { foo = 2; }`,
			Options: map[string]interface{}{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 17}},
		},

		// ---- Double negation BigInt (- -1n) ----
		{
			Code:    `foo[- -1n]`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1n."}},
		},

		// ---- Hex numbers in expressions ----
		{
			Code: `console.log(0x1A + 0x02);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 0x1A."},
				{MessageId: "noMagic", Message: "No magic number: 0x02."},
			},
		},
	})
}
