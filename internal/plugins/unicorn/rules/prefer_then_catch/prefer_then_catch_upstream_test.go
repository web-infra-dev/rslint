// TestPreferThenCatchUpstream migrates the full valid/invalid suite from
// upstream test/prefer-then-catch.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live
// in prefer_then_catch_extras_test.go.
package prefer_then_catch_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	prefer_then_catch "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_then_catch"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID           = "prefer-then-catch"
	suggestionMessageID = "prefer-then-catch/suggestion"
)

const messageText = "Prefer `.then(…).catch(…)` over passing a rejection handler to `.then()`."
const suggestionText = "Move the rejection handler to `.catch()`."

func TestPreferThenCatchUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_then_catch.PreferThenCatchRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: argument count ----
			jsValid(`promise.then();`),
			jsValid(`promise.then(onFulfilled);`),
			jsValid(`promise.then(onFulfilled, onRejected, extraArgument);`),

			// ---- upstream valid: nullish handler ----
			jsValid(`promise.then(undefined, onRejected);`),
			jsValid(`promise.then(null, onRejected);`),
			jsValid(`promise.then(void 0, onRejected);`),
			jsValid(`promise.then(onFulfilled, undefined);`),
			jsValid(`promise.then(onFulfilled, null);`),
			jsValid(`promise.then(onFulfilled, void 0);`),

			// ---- upstream valid: spread arguments ----
			jsValid(`promise.then(...handlers, onRejected);`),
			jsValid(`promise.then(onFulfilled, ...handlers);`),

			// ---- upstream valid: member access shape ----
			jsValid(`promise["then"](onFulfilled, onRejected);`),
			jsValid(`promise?.then(onFulfilled, onRejected);`),
			jsValid(`promise.then?.(onFulfilled, onRejected);`),

			// ---- upstream valid: type-aware, non-native Promise receiver ----
			tsValid(`declare const object: {then(onFulfilled: () => void, onRejected: () => void): void};` + "\n" +
				`object.then(onFulfilled, onRejected);`),
			tsValid(`declare const promise: PromiseLike<string>;` + "\n" +
				`promise.then(onFulfilled, onRejected);`),
			tsValid(`declare const thenable: PromiseLike<string> & {catch(onRejected: () => void): void};` + "\n" +
				`thenable.then(onFulfilled, onRejected);`),
			tsValid(`declare const thenable: {` + "\n" +
				`	then(onFulfilled: (value: string) => void, onRejected: (reason: unknown) => void): {catch: string};` + "\n" +
				`};` + "\n" +
				`thenable.then(onFulfilled, onRejected);`),
			tsValid(`type Thenable = {then(onFulfilled: () => void, onRejected: () => void): void};` + "\n" +
				`declare const value: Promise<void> | Thenable;` + "\n" +
				`value.then(onFulfilled, onRejected);`),

			// ---- upstream valid: type-aware, user-shadows built-in Promise ----
			tsValid(`interface Promise<T> {` + "\n" +
				`	then(onFulfilled: (value: T) => void, onRejected: (reason: unknown) => void): void;` + "\n" +
				`}` + "\n" +
				`declare const promise: Promise<string>;` + "\n" +
				`promise.then(onFulfilled, onRejected);`),
			tsValid(`interface Promise<T> {` + "\n" +
				`	then(onFulfilled: (value: T) => void, onRejected: (reason: unknown) => void): Promise<void> | undefined;` + "\n" +
				`}` + "\n" +
				`declare const promise: Promise<string>;` + "\n" +
				`promise.then(onFulfilled, onRejected);`),

			// ---- upstream valid: type-asserted nullish handler ----
			tsValid(`promise.then(onFulfilled, undefined as (error: unknown) => void);`),
			tsValid(`promise.then(null!, onRejected);`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: simple form ----
			upstreamInvalid(`promise.then(onFulfilled, onRejected);`),
			upstreamInvalid(`promise.then(onFulfilled, function onRejected() {});`),

			// ---- upstream invalid: type-aware, `any` receiver reports by default ----
			upstreamInvalidTs(`declare const value: any;` + "\n" + `value.then(onFulfilled, onRejected);`),

			// ---- upstream invalid: shadowed undefined is not global, so isNullish is false ----
			upstreamInvalid(`function handlePromise(undefined) { promise.then(onFulfilled, undefined); }`),

			// ---- upstream invalid: non-method-call receiver shape ----
			upstreamInvalid(`Promise.resolve(value).then(onFulfilled, onRejected);`),

			// ---- upstream invalid: expression-statement wrappers ----
			upstreamInvalid(`void promise.then(onFulfilled, onRejected);`),

			// ---- upstream invalid: trailing comma ----
			upstreamInvalid(`promise.then(onFulfilled, onRejected,);`),

			// ---- upstream invalid: parentheses around rejection handler ----
			upstreamInvalid(`promise.then(onFulfilled, (onRejected));`),

			// ---- upstream invalid: multi-line call ----
			upstreamInvalidWithOutput("promise\n\t.then(\n\t\tonFulfilled,\n\t\tonRejected,\n\t)\n\t.then(next);",
				"promise\n\t.then(\n\t\tonFulfilled\n\t).catch(onRejected)\n\t.then(next);"),

			// ---- upstream invalid: type-asserted rejection handler ----
			upstreamInvalidTs(`promise.then(onFulfilled, onRejected as (error: unknown) => void);`),
			upstreamInvalidTs(`promise.then(onFulfilled, <(error: unknown) => void> onRejected);`),

			// ---- upstream invalid: rejection handler that is a function expression with comments ----
			// No suggestion because removing it would drop the comment inside its body.
			upstreamInvalidNoFix(`promise.then(onFulfilled, error => { /* Keep this comment. */ handle(error); });`),

			// ---- upstream invalid: rejection handlers not safe to move (no suggestion) ----
			upstreamInvalidNoFix(`promise.then(onFulfilled, createRejectionHandler());`),
			upstreamInvalidNoFix(`promise.then(onFulfilled, handlers.onRejected);`),
			upstreamInvalidNoFix("promise.then(onFulfilled, tag`handler`);"),
			upstreamInvalidNoFix(`promise.then(onFulfilled, [...handlers]);`),
			upstreamInvalidNoFix(`promise.then(onFulfilled, {...handlers});`),

			// ---- upstream invalid: comment arguments block the suggestion ----
			upstreamInvalidNoFix(`promise.then(onFulfilled, /* Do not move this comment. */ onRejected);`),
			upstreamInvalidNoFix(`promise.then(onFulfilled, onRejected /* Do not move this comment. */);`),
			upstreamInvalidNoFix(`promise.then(onFulfilled, onRejected, /* Do not move this comment. */);`),

			// ---- upstream invalid: native Promise (always has callable catch) ----
			upstreamInvalidTs(`declare const promise: Promise<string>;` + "\n" + `promise.then(onFulfilled, onRejected);`),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func upstreamInvalid(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, true, defaultSuggestionOutput(code)),
		},
	}
}

func upstreamInvalidTs(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.ts",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, true, defaultSuggestionOutput(code)),
		},
	}
}

func upstreamInvalidWithOutput(code, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, true, output),
		},
	}
}

func upstreamInvalidNoFix(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(code, false, ""),
		},
	}
}

func upstreamError(code string, withSuggestion bool, output string) rule_tester.InvalidTestCaseError {
	const target = "then"
	offset := strings.Index(code, target)
	if offset < 0 {
		panic("then not found in prefer-then-catch test: " + code)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	err := rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   messageText,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
	if withSuggestion {
		err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: suggestionMessageID,
			Output:    output,
		}}
	}
	return err
}

// defaultSuggestionOutput reconstructs upstream's suggestion for the common
// single-line shape `<receiver>.then(<first>, <second>)<tail>` — drop the
// rejection handler (and any whitespace after the first comma) up to and
// including the trailing comma, then append `.catch(<rejectionHandler)` after
// the closing paren of the rewritten `.then(...)`. Cases that don't fit this
// shape supply their own expected output via upstreamInvalidWithOutput.
func defaultSuggestionOutput(code string) string {
	const callMarker = "then("
	callOffset := strings.Index(code, callMarker)
	if callOffset < 0 {
		panic("then( not found in prefer-then-catch test: " + code)
	}
	afterCall := callOffset + len(callMarker)

	// Walk the argument list, tracking depth, to find the rejection handler's
	// leading and trailing comma boundaries. Both must be at top level (depth
	// 0) inside the call.
	depth := 0
	commaPositions := []int{}
	closeParen := -1
	for i := afterCall; i < len(code); i++ {
		switch code[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if code[i] == ')' && depth == 0 {
				closeParen = i
			} else {
				depth--
			}
		case ',':
			if depth == 0 {
				commaPositions = append(commaPositions, i)
			}
		}
		if closeParen >= 0 {
			break
		}
	}
	if closeParen < 0 || len(commaPositions) < 1 {
		panic("could not parse prefer-then-catch suggestion input: " + code)
	}

	firstComma := commaPositions[0]

	// Strip the trailing comma (and any whitespace between rejection handler
	// end and that comma). Whatever whitespace remained between the first
	// comma and the rejection handler is also trimmed, matching upstream's
	// removal range being measured from the first comma's end.
	dropEnd := firstComma + 1
	for dropEnd < len(code) {
		ch := code[dropEnd]
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
		dropEnd++
	}
	// The rejection handler ends just before the trailing comma (if any), or
	// at the closing paren of the call. A trailing comma in the args list must
	// not be carried into the `.catch(...)` rewrite.
	handlerEnd := closeParen
	if len(commaPositions) > 1 {
		handlerEnd = commaPositions[len(commaPositions)-1]
	}
	rejectionHandler := strings.TrimSpace(code[dropEnd:handlerEnd])

	return code[:firstComma] + ")" + ".catch(" + rejectionHandler + ")" + code[closeParen+1:]
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
