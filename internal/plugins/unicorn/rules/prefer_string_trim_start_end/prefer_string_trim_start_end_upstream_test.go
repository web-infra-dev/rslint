// TestPreferStringTrimStartEndUpstream migrates the complete v74.0.0
// upstream test/prefer-string-trim-start-end.js suite. rslint-specific edge
// shapes and branch lock-ins live in the sibling extras test file.
package prefer_string_trim_start_end_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_trim_start_end"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "prefer-string-trim-start-end"

func expectedMessage(method, replacement string) string {
	return "Prefer `String#" + replacement + "()` over `String#" + method + "()`."
}

func trimInvalid(code, fileName, method, replacement string, occurrence int) rule_tester.InvalidTestCase {
	offset := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], method)
		if relative < 0 {
			panic("method not found in prefer-string-trim-start-end test: " + method)
		}
		offset = searchFrom + relative
		searchFrom = offset + len(method)
	}

	prefix := code[:offset]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := offset + 1
	if lastNewline >= 0 {
		column = offset - lastNewline
	}

	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Output:   []string{code[:offset] + replacement + code[offset+len(method):]},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   expectedMessage(method, replacement),
			Line:      line,
			Column:    column,
			EndLine:   line,
			EndColumn: column + len(method),
		}},
	}
}

func TestPreferStringTrimStartEndUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_string_trim_start_end.PreferStringTrimStartEndRule,
		[]rule_tester.ValidTestCase{
			{Code: `function f(foo: number[]) { foo.trimLeft(); }`, FileName: "file.ts"},
			{Code: `foo.trimStart()`, FileName: "file.js"},
			{Code: `foo.trimStart?.()`, FileName: "file.js"},
			{Code: `foo.trimEnd()`, FileName: "file.js"},
			{Code: `new foo.trimLeft();`, FileName: "file.js"},
			{Code: `trimLeft();`, FileName: "file.js"},
			{Code: `foo['trimLeft']();`, FileName: "file.js"},
			{Code: `foo[trimLeft]();`, FileName: "file.js"},
			{Code: `foo.bar();`, FileName: "file.js"},
			{Code: `foo.trimLeft(extra);`, FileName: "file.js"},
			{Code: `foo.trimLeft(...argumentsArray)`, FileName: "file.js"},
			{Code: `foo.bar(trimLeft)`, FileName: "file.js"},
			{Code: `foo.bar(foo.trimLeft)`, FileName: "file.js"},
			{Code: `trimLeft.foo()`, FileName: "file.js"},
			{Code: `foo.trimLeft.bar()`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			trimInvalid(`function f(foo: string) { foo.trimLeft(); }`, "file.ts", "trimLeft", "trimStart", 0),
			trimInvalid(`foo.trimLeft()`, "file.js", "trimLeft", "trimStart", 0),
			trimInvalid(`foo.trimRight()`, "file.js", "trimRight", "trimEnd", 0),
			trimInvalid(`trimLeft.trimRight()`, "file.js", "trimRight", "trimEnd", 0),
			trimInvalid(`foo.trimLeft.trimRight()`, "file.js", "trimRight", "trimEnd", 0),
			trimInvalid(`"foo".trimLeft()`, "file.js", "trimLeft", "trimStart", 0),
			trimInvalid("foo\n\t// comment\n\t.trimRight/* comment */(\n\t\t/* comment */\n\t)", "file.js", "trimRight", "trimEnd", 0),
			trimInvalid(`foo?.trimLeft()`, "file.js", "trimLeft", "trimStart", 0),
		},
	)
}
