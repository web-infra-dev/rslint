package no_invalid_remove_event_listener_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_invalid_remove_event_listener "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_remove_event_listener"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "no-invalid-remove-event-listener"
	message   = "The listener argument should be a function reference."
)

// TestNoInvalidRemoveEventListenerUpstream migrates the full valid/invalid
// suite from upstream test/no-invalid-remove-event-listener.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in no_invalid_remove_event_listener_extras_test.go.
func TestNoInvalidRemoveEventListenerUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_invalid_remove_event_listener.NoInvalidRemoveEventListenerRule,
		[]rule_tester.ValidTestCase{
			// ---- CallExpression ----
			jsValid(`new el.removeEventListener("click", () => {})`),
			jsValid(`el.removeEventListener?.("click", () => {})`),
			jsValid(`el.notRemoveEventListener("click", () => {})`),
			jsValid(`el[removeEventListener]("click", () => {})`),
			jsValid(`el["removeEventListener"]("click", () => {})`),

			// ---- Arguments ----
			jsValid(`el.removeEventListener("click")`),
			jsValid(`el.removeEventListener()`),
			jsValid(`el.removeEventListener(() => {})`),
			jsValid(`el.removeEventListener(...["click", () => {}], () => {})`),
			jsValid(`el.removeEventListener(...args, () => {})`),
			jsValid(`el.removeEventListener(() => {}, "click")`),
			jsValid(`window.removeEventListener("click", bind())`),
			jsValid(`window.removeEventListener("click", handler.notBind())`),
			jsValid(`window.removeEventListener("click", handler[bind]())`),
			jsValid(`window.removeEventListener("click", handler.bind?.())`),
			jsValid(`window.removeEventListener("click", handler?.bind())`),

			jsValid(`window.removeEventListener(handler)`),
			jsValid(`class MyComponent {
	handler() {}
	disconnectedCallback() {
		this.removeEventListener('click', this.handler);
	}
}`),
			jsValid(`this.removeEventListener("click", getListener())`),
			jsValid(`el.removeEventListener("scroll", handler)`),
			jsValid(`el.removeEventListener("keydown", obj.listener)`),
			jsValid(`removeEventListener("keyup", () => {})`),
			jsValid(`removeEventListener("keydown", function () {})`),
		},
		[]rule_tester.InvalidTestCase{
			bindInvalid(`window.removeEventListener("scroll", handler.bind(abc))`),
			bindInvalid(`window.removeEventListener("scroll", this.handler.bind(abc))`),
			arrowInvalid(`window.removeEventListener("click", () => {})`),
			functionInvalid(`window.removeEventListener("keydown", function () {})`, `function `),
			// Named function expression and async arrow are still inline functions.
			functionInvalid(`el.removeEventListener("click", function handleClick() {})`, `function handleClick`),
			arrowInvalid(`el.removeEventListener("click", async () => {})`),
			arrowInvalid(`el.removeEventListener("click", (e) => { e.preventDefault(); })`),
			bindInvalid(`el.removeEventListener("mouseover", fn.bind(abc))`),
			bindInvalid(`el?.removeEventListener("mouseover", fn.bind(abc))`),
			functionInvalid(`el.removeEventListener("mouseout", function (e) {})`, `function `),
			functionInvalid(`el?.removeEventListener("mouseout", function (e) {})`, `function `),
			functionInvalid(`el.removeEventListener("mouseout", function (e) {}, true)`, `function `),
			functionInvalid(`el.removeEventListener("click", function (e) {}, ...moreArguments)`, `function `),
			invalid(`el.removeEventListener(() => {}, () => {}, () => {})`, `=>`, 2),
		},
	)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func bindInvalid(code string) rule_tester.InvalidTestCase {
	return invalid(code, "bind", 1)
}

func arrowInvalid(code string) rule_tester.InvalidTestCase {
	return invalid(code, "=>", 1)
}

func functionInvalid(code string, target string) rule_tester.InvalidTestCase {
	return invalid(code, target, 1)
}

func invalid(code string, target string, occurrence int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, occurrence),
		},
	}
}

func expectedError(code string, target string, occurrence int) rule_tester.InvalidTestCaseError {
	offset := nthIndex(code, target, occurrence)
	if offset < 0 {
		panic("target not found in no-invalid-remove-event-listener test: " + target + " in " + code)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nthIndex(text string, target string, occurrence int) int {
	from := 0
	for i := range occurrence {
		index := strings.Index(text[from:], target)
		if index < 0 {
			return -1
		}
		from += index
		if i+1 < occurrence {
			from += len(target)
		}
	}
	return from
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
