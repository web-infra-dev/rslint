// Test helpers shared by the upstream and extras test files. Kept in a
// separate _test.go file so each suite can be compiled and run independently
// while still agreeing on a single source of truth for the operator-token
// position arithmetic and the InvalidTestCase builder.
package no_xor_as_exponentiation_test

import (
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageIDError      = "no-xor-as-exponentiation/error"
	messageIDSuggestion = "no-xor-as-exponentiation/suggestion"

	messageErrorText      = "Unexpected bitwise XOR operator `^`. Did you mean the exponentiation operator `**`?"
	messageSuggestionText = "Replace `^` with `**`."
)

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

// xorInvalid reports the operator-token position of the first `^` in `code`
// and asserts the suggestion rewrites it to `**`.
func xorInvalid(code string) rule_tester.InvalidTestCase {
	offset := strings.Index(code, "^")
	if offset < 0 {
		panic("no `^` found in no-xor-as-exponentiation test: " + code)
	}
	return xorAtInvalid(code, offset)
}

// xorAtInvalid reports the operator-token position of the `^` at
// `indexOfCaret` (0-based byte offset) in `code`, asserting the suggestion
// replaces that one specific token with `**`. Use this for cases where the
// code contains multiple `^` and the rule fires on a non-first one.
func xorAtInvalid(code string, indexOfCaret int) rule_tester.InvalidTestCase {
	if indexOfCaret < 0 || indexOfCaret >= len(code) || code[indexOfCaret] != '^' {
		panic("indexOfCaret does not point at `^` in: " + code)
	}

	line, column := lineColumnForOffset(code, indexOfCaret)
	endLine, endColumn := lineColumnForOffset(code, indexOfCaret+len("^"))

	replaced := code[:indexOfCaret] + "**" + code[indexOfCaret+len("^"):]

	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageIDError,
			Message:   messageErrorText,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: messageIDSuggestion,
				Output:    replaced,
			}},
		}},
	}
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index := 0; index < offset; {
		r, size := utf8.DecodeRuneInString(code[index:])
		if r == '\n' {
			line++
			column = 1
		} else if r > 0xFFFF {
			column += 2
		} else {
			column++
		}
		index += size
	}
	return line, column
}
