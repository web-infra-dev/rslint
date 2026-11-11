// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
package no_misrefactored_assignment_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_misrefactored_assignment"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-misrefactored-assignment.js
func TestNoMisrefactoredAssignmentUpstream(t *testing.T) {
	valid := jsCases(
		// No duplication.
		"a += b;", "a += b + c;", "a -= b - c;", "a *= b * c;",
		// Operator mismatch.
		"a += a - b;", "a -= a + b;", "a *= a + b;", "a /= a * b;",
		// Not a handled compound assignment.
		"a = a + b;", "a += a;",
		// Logical assignment operators.
		"a &&= a && b;", "a ||= a || b;", "a ??= a ?? b;",
		// Non-commutative operators with the target as the right operand.
		"a -= b - a;", "a /= b / a;", "a %= b % a;", "a **= b ** a;",
		"a <<= b << a;", "a >>= b >> a;", "a >>>= b >>> a;",
		// Different references.
		"a.x += a.y + b;", "a[0] += a[1] + b;", "foo += bar + baz;", "a.x += b.x + c;",
		// Computed key with a side effect.
		"a[foo()] += a[foo()] + b;",
		// Bitwise operator mismatch.
		"a |= a & b;", "a <<= a >> b;",
		// Target is not an immediate operand.
		"a += a + b + c;", "a += b + a + c;",
	)
	invalid := []rule_tester.InvalidTestCase{
		// Comments inside the kept operand.
		wholeAssignment("case.js", "a += a + getValue(/* keep */);", "a", 1, 30, "a += getValue(/* keep */);"),
		wholeAssignment("case.js", "a *= getValue(/* keep */) * a;", "a", 1, 30, "a *= getValue(/* keep */);"),
		wholeAssignment("case.js", "a += a + (condition ? /* keep */ first : second);", "a", 1, 49, "a += (condition ? /* keep */ first : second);"),
		wholeAssignment("case.js", "a += a + values[/* keep */ index];", "a", 1, 34, "a += values[/* keep */ index];"),

		// Comment boundaries around kept and discarded operands.
		wholeAssignment("case.js", "a += a + (/* keep */ condition ? first : second);", "a", 1, 49, "a += (/* keep */ condition ? first : second);"),
		wholeAssignment("case.js", "a += a + getValue(\r\n  // keep\r\n);", "a", 3, 2, "a += getValue(\r\n  // keep\r\n);"),
		wholeAssignment("case.js", "a += a + (/* keep */ first(), second);", "a", 1, 38, "a += (/* keep */ first(), second);"),
		wholeAssignment("case.js", "a *= a * (left /* keep */ + right);", "a", 1, 35, "a *= left /* keep */ + right;"),
		wholeAssignment("case.ts", "a += a + (value as /* keep */ number);", "a", 1, 38, "a += (value as /* keep */ number);"),
		wholeAssignment("case.js", "a += (/* keep */ a) + value;", "a", 1, 28, ""),
		wholeAssignment("case.js", "a += a /* keep */ + value;", "a", 1, 26, ""),
		wholeAssignment("case.js", "a += a + (/* keep */ value);", "a", 1, 28, ""),
		wholeAssignment("case.js", "a += a + (value /* keep */);", "a", 1, 28, ""),
		wholeAssignment("case.js", "a *= a * (/* keep */ left + right);", "a", 1, 35, ""),

		// The upstream runtime test checks that the kept call runs once with or
		// without the suggestion. Here, assert the edit and absence of an autofix.
		{
			FileName: "case.js",
			Code:     "let value = 2; let evaluations = 0; function getValue() { evaluations++; return 3; } value += value + getValue(/* keep */); JSON.stringify({value, evaluations});",
			Errors: []rule_tester.InvalidTestCaseError{
				assignmentError("value", 1, 86, 1, 123, "let value = 2; let evaluations = 0; function getValue() { evaluations++; return 3; } value += getValue(/* keep */); JSON.stringify({value, evaluations});"),
			},
		},

		// Snapshot suite: duplicated left operand for every operator.
		wholeAssignment("case.js", "a += a + b;", "a", 1, 11, "a += b;"),
		wholeAssignment("case.js", "a -= a - b;", "a", 1, 11, "a -= b;"),
		wholeAssignment("case.js", "a *= a * b;", "a", 1, 11, "a *= b;"),
		wholeAssignment("case.js", "a /= a / b;", "a", 1, 11, "a /= b;"),
		wholeAssignment("case.js", "a %= a % b;", "a", 1, 11, "a %= b;"),
		wholeAssignment("case.js", "a **= a ** b;", "a", 1, 13, "a **= b;"),
		wholeAssignment("case.js", "a &= a & b;", "a", 1, 11, "a &= b;"),
		wholeAssignment("case.js", "a |= a | b;", "a", 1, 11, "a |= b;"),
		wholeAssignment("case.js", "a ^= a ^ b;", "a", 1, 11, "a ^= b;"),
		wholeAssignment("case.js", "a <<= a << b;", "a", 1, 13, "a <<= b;"),
		wholeAssignment("case.js", "a >>= a >> b;", "a", 1, 13, "a >>= b;"),
		wholeAssignment("case.js", "a >>>= a >>> b;", "a", 1, 15, "a >>>= b;"),

		// Duplicated right operand for commutative operators.
		wholeAssignment("case.js", "a += b + a;", "a", 1, 11, "a += b;"),
		wholeAssignment("case.js", "a *= b * a;", "a", 1, 11, "a *= b;"),
		wholeAssignment("case.js", "a &= b & a;", "a", 1, 11, "a &= b;"),
		wholeAssignment("case.js", "a |= b | a;", "a", 1, 11, "a |= b;"),
		wholeAssignment("case.js", "a ^= b ^ a;", "a", 1, 11, "a ^= b;"),

		// Member, this, computed and private targets.
		wholeAssignment("case.js", "foo.bar += foo.bar + 1;", "foo.bar", 1, 23, "foo.bar += 1;"),
		wholeAssignment("case.js", "this.x += this.x + y;", "this.x", 1, 21, "this.x += y;"),
		wholeAssignment("case.js", "a.b[c] += a.b[c] + d;", "a.b[c]", 1, 21, "a.b[c] += d;"),
		wholeAssignment("case.js", `a.x += a["x"] + b;`, "a.x", 1, 18, "a.x += b;"),
		{
			FileName: "case.js",
			Code:     "class A { #x = 0; m() { this.#x += this.#x + b; } }",
			Errors: []rule_tester.InvalidTestCaseError{
				assignmentError("this.#x", 1, 25, 1, 47, "class A { #x = 0; m() { this.#x += b; } }"),
			},
		},

		// Parenthesized target operand, string concatenation and two matches.
		wholeAssignment("case.js", "a += (a) + b;", "a", 1, 13, "a += b;"),
		wholeAssignment("case.js", `s += s + "!";`, "s", 1, 13, `s += "!";`),
		wholeAssignment("case.js", "a += a + a;", "a", 1, 11, "a += a;"),
		{
			FileName: "case.js",
			Code:     "a += a + (b += b + c);",
			Errors: []rule_tester.InvalidTestCaseError{
				assignmentError("a", 1, 1, 1, 22, "a += (b += b + c);"),
				assignmentError("b", 1, 11, 1, 21, "a += a + (b += c);"),
			},
		},

		// The comment before the RHS is outside the replacement range.
		wholeAssignment("case.js", "a += a + /* keep */ b;", "a", 1, 22, ""),
		wholeAssignment("case.js", "a += /* lead */ a + b;", "a", 1, 22, "a += /* lead */ b;"),
		wholeAssignment("case.js", "a += a + (b + c);", "a", 1, 17, "a += b + c;"),
		wholeAssignment("case.js", "a += (a + b);", "a", 1, 13, "a += (b);"),

		// TypeScript wrappers on the duplicated operand.
		wholeAssignment("case.ts", "a += (a as number) + b;", "a", 1, 23, "a += b;"),
		wholeAssignment("case.ts", "a += a! + b;", "a", 1, 12, "a += b;"),
		wholeAssignment("case.ts", "a += (<number>a) + b;", "a", 1, 21, "a += b;"),

		// Grouping of the kept operand.
		wholeAssignment("case.js", "a += a + (b, c);", "a", 1, 16, "a += (b, c);"),
		wholeAssignment("case.js", "a += a + (b ? c : d);", "a", 1, 21, "a += (b ? c : d);"),
		wholeAssignment("case.js", "a += (b ? c : d) + a;", "a", 1, 21, "a += (b ? c : d);"),
		wholeAssignment("case.js", "a *= a * (b + c);", "a", 1, 17, "a *= b + c;"),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_misrefactored_assignment.NoMisrefactoredAssignmentRule, valid, invalid)
}

func TestNoMisrefactoredAssignmentDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_misrefactored_assignment.NoMisrefactoredAssignmentRule,
		jsCases("a += b;", "a *= b;", "a += b + c;", "a += a - b;", "a -= b - a;"),
		[]rule_tester.InvalidTestCase{
			wholeAssignment("case.js", "a += a + b;", "a", 1, 11, "a += b;"),
			wholeAssignment("case.js", "a *= b * a;", "a", 1, 11, "a *= b;"),
		},
	)
}

func jsCases(codes ...string) []rule_tester.ValidTestCase {
	cases := make([]rule_tester.ValidTestCase, len(codes))
	for i, code := range codes {
		cases[i] = rule_tester.ValidTestCase{FileName: "case.js", Code: code}
	}
	return cases
}

func wholeAssignment(file, code, target string, endLine, endColumn int, suggestion string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		FileName: file,
		Code:     code,
		Errors:   []rule_tester.InvalidTestCaseError{assignmentError(target, 1, 1, endLine, endColumn, suggestion)},
	}
}

func assignmentError(target string, line, column, endLine, endColumn int, suggestion string) rule_tester.InvalidTestCaseError {
	expected := rule_tester.InvalidTestCaseError{
		MessageId: "no-misrefactored-assignment/error",
		Message:   "`" + target + "` is duplicated in the right-hand side, which is likely a refactoring mistake.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
	if suggestion != "" {
		expected.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "no-misrefactored-assignment/suggestion",
			Output:    suggestion,
		}}
	}
	return expected
}
