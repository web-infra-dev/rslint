// TestNoArrayFromFillUpstream migrates the complete valid/invalid suite from
// upstream test/no-array-from-fill.js at v74.0.0. Position assertions cover
// every invalid case. rslint-specific cases live in
// no_array_from_fill_extras_test.go.
package no_array_from_fill_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_from_fill"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID   = "no-array-from-fill"
	messageText = "Use the `Array.from(…, mapFunction)` argument instead of chaining `.fill()`."
)

func TestNoArrayFromFillUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_from_fill.NoArrayFromFillRule,
		[]rule_tester.ValidTestCase{
			upstreamValid(`Array.from({length: 3})`),
			upstreamValid(`Array.from({length: 3}, (_, index) => index)`),
			upstreamValid(`Array.from({length: 3}).map((_, index) => index)`),
			upstreamValid(`Array.from(items).fill(0)`),
			upstreamValid(`Array.from({length: 3, 0: "value"}).fill(0)`),
			upstreamValid(`Array.from({...length}).fill(0)`),
			upstreamValid(`Array.from({["length"]: 3}).fill(0)`),
			upstreamValid(`Array.from({length: 3}).fill(0, 1)`),
			upstreamValid(`Array.from({length: 3}).fill(0, 1, 2)`),
			upstreamValid(`Array.from({length: 3}).fill(...value)`),
			upstreamValid(`Array.from?.({length: 3}).fill(0)`),
			upstreamValid(`Array.from({length: 3})?.fill(0)`),
			upstreamValid(`Array.from({length: 3}).fill?.(0)`),
			upstreamValid(`NotArray.from({length: 3}).fill(0)`),
			upstreamValid(`Array.notFrom({length: 3}).fill(0)`),
			upstreamValid(`Array.from({length: 3}).slice().fill(0)`),
			upstreamValid(`const Array = {from() { return {fill() { return {map() {}}; }}; }}; Array.from({length: 3}).fill().map((_, index) => index)`),
			upstreamValid(`function unicorn(Array) { return Array.from({length: 3}).fill(0); }`),
		},
		[]rule_tester.InvalidTestCase{
			upstreamInvalid(`Array.from({length: 3}).fill(0)`),
			upstreamInvalid(`Array.from({length: 3}).fill()`),
			upstreamInvalid(`Array.from({length}).fill(null)`),
			upstreamInvalid(`Array.from({"length": 3}).fill(0)`),
			upstreamInvalid(`Array.from({length: 3}).fill({})`),
			upstreamInvalid(`Array.from({length: 3}).fill(0).map((_, index) => index)`),
			upstreamInvalid(`Array.from({length: 3}).fill().map((value, index) => index)`),
			upstreamInvalid(`Array.from({length: 3}).fill(0).flatMap((_, index) => [index])`),
			upstreamInvalid(`Array.from({length: 3}).fill().flatMap(value => [value])`),
			upstreamInvalid(`Array.from({length: 3}).fill(0).filter(Boolean)`),
			upstreamInvalid("Array.from({length: 3})\n\t.fill(0)\n\t.map((_, index) => index);"),
			upstreamInvalid("Array.from(\n\t{length: 3}\n)\n\t.fill(0)\n\t.map((_, index) => index);"),
		},
	)
}

func upstreamValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func upstreamInvalid(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, "fill", 0),
		},
	}
}

func expectedError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
	offset := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("target not found in no-array-from-fill test: " + target)
		}
		offset = searchFrom + relative
		searchFrom = offset + len(target)
	}

	line, column := lineColumn(code, offset)
	endLine, endColumn := lineColumn(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   messageText,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func lineColumn(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index, character := range code {
		if index >= offset {
			break
		}
		if character == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
