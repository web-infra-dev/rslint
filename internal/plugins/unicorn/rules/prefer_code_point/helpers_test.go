// Test helpers shared by the upstream and extras test files. The reported
// node is always the property identifier (`charCodeAt` or `fromCharCode`), so
// the position arithmetic lives here once instead of being repeated per case.
package prefer_code_point_test

import (
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageIDCharCodeAt   = "error/charCodeAt"
	messageIDFromCharCode = "error/fromCharCode"

	messageTextCharCodeAt   = "Prefer `String#codePointAt()` over `String#charCodeAt()`."
	messageTextFromCharCode = "Prefer `String.fromCodePoint()` over `String.fromCharCode()`."

	suggestionIDCodePointAt   = "suggestion/codePointAt"
	suggestionIDFromCodePoint = "suggestion/fromCodePoint"
)

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

func validTS(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

// invalid reports the first occurrence of `token` in `code`, asserting the
// diagnostic range covers that token. When suggest is true the single
// suggestion must rewrite the token to replacement.
func invalid(code, token, replacement string, suggest bool) rule_tester.InvalidTestCase {
	return invalidIn(code, "file.mjs", token, replacement, suggest)
}

func invalidTS(code, token, replacement string, suggest bool) rule_tester.InvalidTestCase {
	return invalidIn(code, "file.ts", token, replacement, suggest)
}

func invalidIn(code, fileName, token, replacement string, suggest bool) rule_tester.InvalidTestCase {
	offset := strings.Index(code, token)
	if offset < 0 {
		panic("token not found in prefer-code-point test: " + token + " in " + code)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(token))

	testCase := rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageIDFor(token),
			Message:   messageTextFor(token),
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		}},
	}

	if suggest {
		testCase.Errors[0].Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: suggestionIDFor(token),
			Output:    code[:offset] + replacement + code[offset+len(token):],
		}}
	}

	return testCase
}

func messageIDFor(token string) string {
	if token == "charCodeAt" {
		return messageIDCharCodeAt
	}
	return messageIDFromCharCode
}

func messageTextFor(token string) string {
	if token == "charCodeAt" {
		return messageTextCharCodeAt
	}
	return messageTextFromCharCode
}

func suggestionIDFor(token string) string {
	if token == "charCodeAt" {
		return suggestionIDCodePointAt
	}
	return suggestionIDFromCodePoint
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
