// TestPreferAddEventListenerOptionsUpstream migrates the complete v74.0.0
// upstream test/prefer-add-event-listener-options.js suite. rslint-specific
// edge shapes and branch lock-ins live in the sibling extras test file.
package prefer_add_event_listener_options_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	prefer_add_event_listener_options "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_add_event_listener_options"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "prefer-add-event-listener-options"

func expectedMessage(value string) string {
	return "Prefer `{capture: " + value + "}` over `" + value + "`."
}

func upstreamError(value string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   expectedMessage(value),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestPreferAddEventListenerOptionsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_add_event_listener_options.PreferAddEventListenerOptionsRule,
		[]rule_tester.ValidTestCase{
			{Code: `window.addEventListener("click", listener)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, {capture: true})`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, {capture: false})`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, {passive: true})`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, {once: true})`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, {signal})`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, options)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, capture)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, Boolean(value))`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, condition ? true : false)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, undefined)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, null)`, FileName: "file.js"},
			{Code: `window["addEventListener"]("click", listener, true)`, FileName: "file.js"},
			{Code: `window?.addEventListener("click", listener, true)`, FileName: "file.js"},
			{Code: `window.addEventListener?.("click", listener, true)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", ...arguments_, true)`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `window.addEventListener("click", listener, true)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 1, 44, 1, 48)},
			},
			{
				Code:     `window.addEventListener("click", listener, false)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", listener, {capture: false})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("false", 1, 44, 1, 49)},
			},
			{
				Code:     `window.addEventListener("click", listener, (true))`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", listener, ({capture: true}))`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 1, 45, 1, 49)},
			},
			{
				Code:     `window.addEventListener("click", () => {}, true)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", () => {}, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 1, 44, 1, 48)},
			},
			{
				Code:     `window.addEventListener("click", function () {}, false)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", function () {}, {capture: false})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("false", 1, 50, 1, 55)},
			},
			{
				Code:     `document.body.addEventListener("click", listener, true)`,
				FileName: "file.js",
				Output:   []string{`document.body.addEventListener("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 1, 51, 1, 55)},
			},
			{
				Code:     `(window).addEventListener("click", listener, false)`,
				FileName: "file.js",
				Output:   []string{`(window).addEventListener("click", listener, {capture: false})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("false", 1, 46, 1, 51)},
			},
			{
				Code:     `window.addEventListener("click", listener, /* useCapture */ true)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", listener, /* useCapture */ {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 1, 61, 1, 65)},
			},
			{
				Code:     `window.addEventListener("click", listener, true /* useCapture */)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", listener, {capture: true} /* useCapture */)`},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 1, 44, 1, 48)},
			},
			{
				Code:     "window.addEventListener(\n\t\"click\",\n\tlistener,\n\ttrue\n)",
				FileName: "file.js",
				Output:   []string{"window.addEventListener(\n\t\"click\",\n\tlistener,\n\t{capture: true}\n)"},
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError("true", 4, 2, 4, 6)},
			},
		},
	)
}
