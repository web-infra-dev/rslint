// TestNoAwaitInPromiseMethodsUpstream migrates the full valid/invalid suite
// from upstream test/no-await-in-promise-methods.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in no_await_in_promise_methods_extras_test.go.
package no_await_in_promise_methods_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_await_in_promise_methods "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_await_in_promise_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageIDError      = "no-await-in-promise-methods/error"
	messageIDSuggestion = "no-await-in-promise-methods/suggestion"
)

func TestNoAwaitInPromiseMethodsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_await_in_promise_methods.NoAwaitInPromiseMethodsRule,
		[]rule_tester.ValidTestCase{
			jsValid(`Promise.all([promise1, promise2, promise3, promise4])`),
			jsValid(`Promise.allSettled([promise1, promise2, promise3, promise4])`),
			jsValid(`Promise.any([promise1, promise2, promise3, promise4])`),
			jsValid(`Promise.race([promise1, promise2, promise3, promise4])`),
			jsValid(`Promise.all(...[await promise])`),
			jsValid(`Promise.all([await promise], extraArguments)`),
			jsValid(`Promise.all()`),
			jsValid(`Promise.all(notArrayExpression)`),
			jsValid(`Promise.all([,])`),
			jsValid(`Promise[all]([await promise])`),
			jsValid(`Promise.all?.([await promise])`),
			jsValid(`Promise?.all([await promise])`),
			jsValid(`Promise.notListedMethod([await promise])`),
			jsValid(`NotPromise.all([await promise])`),
			jsValid(`Promise.all([(await promise, 0)])`),
			jsValid(`new Promise.all([await promise])`),

			// We are not checking these cases.
			jsValid(`globalThis.Promise.all([await promise])`),
			jsValid(`Promise["all"]([await promise])`),
		},
		[]rule_tester.InvalidTestCase{
			upstreamInvalid(
				`Promise.all([await promise])`,
				upstreamError(`Promise.all([await promise])`, `await promise`, "all", `Promise.all([promise])`),
			),
			upstreamInvalid(
				`Promise.allSettled([await promise])`,
				upstreamError(`Promise.allSettled([await promise])`, `await promise`, "allSettled", `Promise.allSettled([promise])`),
			),
			upstreamInvalid(
				`Promise.any([await promise])`,
				upstreamError(`Promise.any([await promise])`, `await promise`, "any", `Promise.any([promise])`),
			),
			upstreamInvalid(
				`Promise.race([await promise])`,
				upstreamError(`Promise.race([await promise])`, `await promise`, "race", `Promise.race([promise])`),
			),
			upstreamInvalid(
				`Promise.all([, await promise])`,
				upstreamError(`Promise.all([, await promise])`, `await promise`, "all", `Promise.all([, promise])`),
			),
			upstreamInvalid(
				`Promise.all([await promise,])`,
				upstreamError(`Promise.all([await promise,])`, `await promise`, "all", `Promise.all([promise,])`),
			),
			upstreamInvalid(
				`Promise.all([await promise],)`,
				upstreamError(`Promise.all([await promise],)`, `await promise`, "all", `Promise.all([promise],)`),
			),
			upstreamInvalid(
				`Promise.all([await (0, promise)],)`,
				upstreamError(`Promise.all([await (0, promise)],)`, `await (0, promise)`, "all", `Promise.all([(0, promise)],)`),
			),
			upstreamInvalid(
				`Promise.all([await (( promise ))])`,
				upstreamError(`Promise.all([await (( promise ))])`, `await (( promise ))`, "all", `Promise.all([(( promise ))])`),
			),
			upstreamInvalid(
				`Promise.all([await await promise])`,
				upstreamError(`Promise.all([await await promise])`, `await await promise`, "all", `Promise.all([await promise])`),
			),
			upstreamInvalid(
				`Promise.all([...foo, await promise1, await promise2])`,
				upstreamError(
					`Promise.all([...foo, await promise1, await promise2])`,
					`await promise1`,
					"all",
					`Promise.all([...foo, promise1, await promise2])`,
				),
				upstreamError(
					`Promise.all([...foo, await promise1, await promise2])`,
					`await promise2`,
					"all",
					`Promise.all([...foo, await promise1, promise2])`,
				),
			),
			// Multiple awaited elements without a spread.
			upstreamInvalid(
				`Promise.all([await promise1, await promise2])`,
				upstreamError(
					`Promise.all([await promise1, await promise2])`,
					`await promise1`,
					"all",
					`Promise.all([promise1, await promise2])`,
				),
				upstreamError(
					`Promise.all([await promise1, await promise2])`,
					`await promise2`,
					"all",
					`Promise.all([await promise1, promise2])`,
				),
			),
			upstreamInvalid(
				`Promise.any([await a, await b, await c])`,
				upstreamError(`Promise.any([await a, await b, await c])`, `await a`, "any", `Promise.any([a, await b, await c])`),
				upstreamError(`Promise.any([await a, await b, await c])`, `await b`, "any", `Promise.any([await a, b, await c])`),
				upstreamError(`Promise.any([await a, await b, await c])`, `await c`, "any", `Promise.any([await a, await b, c])`),
			),
			upstreamInvalid(
				`Promise.all([await /* comment*/ promise])`,
				upstreamError(
					`Promise.all([await /* comment*/ promise])`,
					`await /* comment*/ promise`,
					"all",
					`Promise.all([/* comment*/ promise])`,
				),
			),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

func upstreamInvalid(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors:   errors,
	}
}

func upstreamError(code, target, method, suggestionOutput string) rule_tester.InvalidTestCaseError {
	offset := strings.Index(code, target)
	if offset < 0 {
		panic("target not found in no-await-in-promise-methods test: " + target)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageIDError,
		Message:   fmt.Sprintf("Promise in `Promise.%s()` should not be awaited.", method),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: messageIDSuggestion,
			Output:    suggestionOutput,
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
