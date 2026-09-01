// TestConsistentDateCloneUpstream migrates the complete valid/invalid suite
// from eslint-plugin-unicorn v74.0.0. rslint-specific edge-shape and branch
// lock-in cases live in consistent_date_clone_extras_test.go.
package consistent_date_clone_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	consistent_date_clone "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_date_clone"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "consistent-date-clone/error"
	message   = "Unnecessary `.getTime()` call."
)

func expectedError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
	searchFrom := 0
	start := -1
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("target not found in test input: " + target)
		}
		start = searchFrom + relative
		searchFrom = start + 1
	}

	lineStart := strings.LastIndex(code[:start], "\n") + 1
	line := strings.Count(code[:start], "\n") + 1
	column := start - lineStart + 1
	end := start + len(target)
	endLineStart := strings.LastIndex(code[:end], "\n") + 1
	endLine := strings.Count(code[:end], "\n") + 1
	endColumn := end - endLineStart + 1

	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestConsistentDateCloneUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_date_clone.ConsistentDateCloneRule,
		[]rule_tester.ValidTestCase{
			{Code: `new Date(date)`, FileName: "file.js"},
			{Code: `date.getTime()`, FileName: "file.js"},
			{Code: `new Date(...date.getTime())`, FileName: "file.js"},
			{Code: `new Date(getTime())`, FileName: "file.js"},
			{Code: `new Date(date.getTime(), extraArgument)`, FileName: "file.js"},
			{Code: `new Date(date.not_getTime())`, FileName: "file.js"},
			{Code: `new Date(date?.getTime())`, FileName: "file.js"},
			{Code: `new NotDate(date.getTime())`, FileName: "file.js"},
			{Code: `new Date(date[getTime]())`, FileName: "file.js"},
			{Code: `new Date(date.getTime(extraArgument))`, FileName: "file.js"},
			{Code: `Date(date.getTime())`, FileName: "file.js"},
			{
				Code: `new Date(
	date.getFullYear(),
	date.getMonth(),
	date.getDate(),
	date.getHours(),
	date.getMinutes(),
	date.getSeconds(),
	date.getMilliseconds(),
);`,
				FileName: "file.js",
			},
			{
				Code: `new Date(
	date.getFullYear(),
	date.getMonth(),
	date.getDate(),
	date.getHours(),
	date.getMinutes(),
	date.getSeconds(),
);`,
				FileName: "file.js",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `new Date(date.getTime())`,
				FileName: "file.js",
				Output:   []string{`new Date(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(date.getTime())`, `getTime()`, 0)},
			},
			{
				Code:     `new Date(date.getTime(),)`,
				FileName: "file.js",
				Output:   []string{`new Date(date,)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(date.getTime(),)`, `getTime()`, 0)},
			},
			{
				Code:     `new Date(new Date(date.getTime()).getTime())`,
				FileName: "file.js",
				Output:   []string{`new Date(new Date(date))`},
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(`new Date(new Date(date.getTime()).getTime())`, `getTime()`, 0),
					expectedError(`new Date(new Date(date.getTime()).getTime())`, `getTime()`, 1),
				},
			},
			{
				Code:     `new Date((0, date).getTime())`,
				FileName: "file.js",
				Output:   []string{`new Date((0, date))`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date((0, date).getTime())`, `getTime()`, 0)},
			},
			{
				Code:     `new Date(date.getTime(/* comment */))`,
				FileName: "file.js",
				Output:   []string{`new Date(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(date.getTime(/* comment */))`, `getTime(/* comment */)`, 0)},
			},
			{
				Code:     `new Date(date./* comment */getTime())`,
				FileName: "file.js",
				Output:   []string{`new Date(date)`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date(date./* comment */getTime())`, `getTime()`, 0)},
			},
			{
				Code:     `new Date((date as Date).getTime())`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`new Date((date as Date))`},
				Errors:   []rule_tester.InvalidTestCaseError{expectedError(`new Date((date as Date).getTime())`, `getTime()`, 0)},
			},
		},
	)
}
