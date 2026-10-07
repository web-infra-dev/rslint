// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-accidental-bitwise-operator.js
package no_accidental_bitwise_operator_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_accidental_bitwise_operator"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	errorMessageID      = "no-accidental-bitwise-operator/error"
	suggestionMessageID = "no-accidental-bitwise-operator/suggestion"
)

func valid(code, filename string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: filename}
}

func operatorIndex(code, operator string, occurrence int) int {
	searchFrom := 0
	index := -1
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], operator)
		if relative < 0 {
			panic("operator not found in no-accidental-bitwise-operator fixture: " + operator)
		}
		index = searchFrom + relative
		searchFrom = index + len(operator)
	}
	return index
}

func invalidAt(code, operator, logicalOperator, filename string, occurrence int) rule_tester.InvalidTestCase {
	index := operatorIndex(code, operator, occurrence)
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: filename,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: errorMessageID,
			Message: "Unexpected bitwise operator `" + operator + "`. Did you mean the logical operator `" + logicalOperator + "`?",
			Line: 1, Column: index + 1, EndLine: 1, EndColumn: index + len(operator) + 1,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: suggestionMessageID,
				Output:    code[:index] + logicalOperator + code[index+len(operator):],
			}},
		}},
	}
}

func invalid(code, operator, logicalOperator, filename string) rule_tester.InvalidTestCase {
	return invalidAt(code, operator, logicalOperator, filename, 0)
}

func TestNoAccidentalBitwiseOperatorUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		valid(`a | b;`, "file.js"),
		valid(`a & b;`, "file.js"),
		valid(`flags & MASK;`, "file.js"),
		valid(`x | 0;`, "file.js"),
		valid(`x | 1;`, "file.js"),
		valid(`options | someVariable;`, "file.js"),
		valid(`a ^ b;`, "file.js"),
		valid(`a << b;`, "file.js"),
		valid(`a >> b;`, "file.js"),
		valid(`a >>> b;`, "file.js"),
		valid(`~x;`, "file.js"),
		valid(`x | (a + b);`, "file.js"),
		valid(`foo() | bar();`, "file.js"),
		valid(`x | null;`, "file.js"),
		valid(`x | 1n;`, "file.js"),
		valid(`x | /regex/;`, "file.js"),
		valid(`a && b;`, "file.js"),
		valid(`a || b;`, "file.js"),
		valid(`obj && obj.prop;`, "file.js"),
		valid(`obj1 & obj2.a;`, "file.js"),
		valid(`obj.a & obj.b;`, "file.js"),
		valid(`obj & obj.a.b;`, "file.js"),
		valid(`a & b.c;`, "file.js"),
		valid(`this & this.a;`, "file.js"),
		valid(`obj & obj;`, "file.js"),
		valid(`obj & obj.prop();`, "file.js"),
		valid(`obj & obj?.a;`, "file.js"),
		valid(`x &= {};`, "file.js"),
		valid(`x &= 1;`, "file.js"),
		valid(`x |= 1;`, "file.js"),
		valid(`x |= y;`, "file.js"),
		valid(`options | ({} as Foo);`, "file.ts"),
	}
	invalidCases := []rule_tester.InvalidTestCase{
		invalid(`obj & obj.a;`, `&`, `&&`, "file.js"),
		invalid(`if (obj & obj.prop) {}`, `&`, `&&`, "file.js"),
		invalid(`obj & obj[key];`, `&`, `&&`, "file.js"),
		invalid(`(obj) & obj.a;`, `&`, `&&`, "file.js"),
		invalid(`obj /* comment */ & obj.a;`, `&`, `&&`, "file.js"),
		invalid(`options | {};`, `|`, `||`, "file.js"),
		invalid(`options | '';`, `|`, `||`, "file.js"),
		invalid(`options | true;`, `|`, `||`, "file.js"),
		invalid(`options | [];`, `|`, `||`, "file.js"),
		invalid("options | `template`;", `|`, `||`, "file.js"),
		invalid(`x | function () {};`, `|`, `||`, "file.js"),
		invalid(`x | (() => {});`, `|`, `||`, "file.js"),
		invalid(`x | class {};`, `|`, `||`, "file.js"),
		invalid(`foo() | {};`, `|`, `||`, "file.js"),
		invalid(`a.b | {};`, `|`, `||`, "file.js"),
		invalidAt(`a | b | {};`, `|`, `||`, "file.js", 1),
		invalid(`input |= '';`, `|=`, `||=`, "file.js"),
		invalid(`input |= {};`, `|=`, `||=`, "file.js"),
		invalid(`input |= false;`, `|=`, `||=`, "file.js"),
		invalid(`obj & obj.a;`, `&`, `&&`, "file.ts"),
		invalid(`options | {};`, `|`, `||`, "file.ts"),
	}
	if len(validCases) != 32 || len(invalidCases) != 21 {
		t.Fatalf("upstream coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_accidental_bitwise_operator.NoAccidentalBitwiseOperatorRule,
		validCases, invalidCases,
	)
}

func TestNoAccidentalBitwiseOperatorDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_accidental_bitwise_operator.NoAccidentalBitwiseOperatorRule,
		[]rule_tester.ValidTestCase{
			valid(`if (object && object.property) {}`, "file.js"),
			valid(`options = options || {};`, "file.js"),
			valid(`input ||= '';`, "file.js"),
			valid("const masked = flags & MASK;\nconst truncated = value | 0;", "file.js"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`if (object & object.property) {}`, `&`, `&&`, "file.js"),
			invalid(`options = options | {};`, `|`, `||`, "file.js"),
			invalid(`input |= '';`, `|=`, `||=`, "file.js"),
		},
	)
}
