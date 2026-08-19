// TestNoInvalidFetchOptionsUpstream migrates the full valid/invalid suite from upstream
// v73.0.0 test/no-invalid-fetch-options.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// no_invalid_fetch_options_extras_test.go file.
package no_invalid_fetch_options_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_fetch_options"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-invalid-fetch-options"

func invalidBodyMessage(method string) string {
	return "\"body\" is not allowed when method is \"" + method + "\"."
}

func TestNoInvalidFetchOptionsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_invalid_fetch_options.NoInvalidFetchOptionsRule,
		[]rule_tester.ValidTestCase{
			jsValid(`fetch(url, {method: "POST", body})`),
			jsValid(`new Request(url, {method: "POST", body})`),
			jsValid(`fetch(url, {})`),
			jsValid(`new Request(url, {})`),
			jsValid(`fetch(url)`),
			jsValid(`new Request(url)`),
			jsValid(`fetch(url, {method: "UNKNOWN", body})`),
			jsValid(`new Request(url, {method: "UNKNOWN", body})`),
			jsValid(`fetch(url, {body: undefined})`),
			jsValid(`new Request(url, {body: undefined})`),
			jsValid(`fetch(url, {body: null})`),
			jsValid(`new Request(url, {body: null})`),
			// `void` always evaluates to `undefined`, so the body is effectively absent.
			jsValid(`fetch(url, {body: void 0})`),
			jsValid(`new Request(url, {method: "GET", body: void 0})`),
			jsValid(`fetch(url, {...options, body})`),
			jsValid(`new Request(url, {...options, body})`),
			jsValid(`new fetch(url, {body})`),
			jsValid(`Request(url, {body})`),
			jsValid(`not_fetch(url, {body})`),
			jsValid(`new not_Request(url, {body})`),
			jsValid(`fetch({body}, url)`),
			jsValid(`new Request({body}, url)`),
			jsValid(`fetch(url, {[body]: "foo=bar"})`),
			jsValid(`new Request(url, {[body]: "foo=bar"})`),
			jsValid("fetch(url, {\n\tbody: 'foo=bar',\n\tbody: undefined,\n});"),
			jsValid("new Request(url, {\n\tbody: 'foo=bar',\n\tbody: undefined,\n});"),
			jsValid("fetch(url, {\n\tmethod: 'HEAD',\n\tbody: 'foo=bar',\n\tmethod: 'post',\n});"),
			jsValid("new Request(url, {\n\tmethod: 'HEAD',\n\tbody: 'foo=bar',\n\tmethod: 'post',\n});"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`fetch(url, {body})`, "body", "GET"),
			invalid(`new Request(url, {body})`, "body", "GET"),
			invalid(`fetch(url, {method: "GET", body})`, "body", "GET"),
			invalid(`new Request(url, {method: "GET", body})`, "body", "GET"),
			invalid(`fetch(url, {method: "HEAD", body})`, "body", "HEAD"),
			invalid(`new Request(url, {method: "HEAD", body})`, "body", "HEAD"),
			invalid(`fetch(url, {method: "head", body})`, "body", "HEAD"),
			invalid(`new Request(url, {method: "head", body})`, "body", "HEAD"),
			invalid(`const method = "head"; new Request(url, {method, body: "foo=bar"})`, "body", "HEAD"),
			invalid(`const method = "head"; fetch(url, {method, body: "foo=bar"})`, "body", "HEAD"),
			invalid(`fetch(url, {body}, extraArgument)`, "body", "GET"),
			invalid(`new Request(url, {body}, extraArgument)`, "body", "GET"),
			invalid("fetch(url, {\n\tbody: undefined,\n\tbody: 'foo=bar',\n});", "body", "GET", 2),
			invalid("new Request(url, {\n\tbody: undefined,\n\tbody: 'foo=bar',\n});", "body", "GET", 2),
			invalid("fetch(url, {\n\tmethod: 'post',\n\tbody: 'foo=bar',\n\tmethod: 'HEAD',\n});", "body", "HEAD"),
			invalid("new Request(url, {\n\tmethod: 'post',\n\tbody: 'foo=bar',\n\tmethod: 'HEAD',\n});", "body", "HEAD"),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func invalid(code string, target string, method string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, method, occurrence...),
		},
	}
}

func expectedError(code string, target string, method string, occurrence ...int) rule_tester.InvalidTestCaseError {
	nth := 1
	if len(occurrence) > 0 {
		nth = occurrence[0]
	}

	offset := nthIndex(code, target, nth)
	if offset < 0 {
		panic("target not found in no-invalid-fetch-options test: " + target + " in " + code)
	}
	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   invalidBodyMessage(method),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nthIndex(value string, target string, nth int) int {
	searchStart := 0
	for i := 1; i <= nth; i++ {
		offset := strings.Index(value[searchStart:], target)
		if offset < 0 {
			return -1
		}
		searchStart += offset
		if i == nth {
			return searchStart
		}
		searchStart += len(target)
	}
	return -1
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for i := range offset {
		if code[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
