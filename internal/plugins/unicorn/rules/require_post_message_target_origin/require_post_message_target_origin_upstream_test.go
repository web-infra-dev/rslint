// Upstream tests and documentation from eslint-plugin-unicorn v74.0.0 (MIT).
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/test/require-post-message-target-origin.js
package require_post_message_target_origin_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_post_message_target_origin"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const message = "Missing the `targetOrigin` argument."

func invalidCase(code, filename string, line, column, endLine, endColumn int, outputs ...string) rule_tester.InvalidTestCase {
	suggestions := make([]rule_tester.InvalidTestCaseSuggestion, 0, len(outputs))
	for _, output := range outputs {
		suggestions = append(suggestions, rule_tester.InvalidTestCaseSuggestion{MessageId: "suggestion", Output: output})
	}
	return rule_tester.InvalidTestCase{
		Code: code, FileName: filename,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "error", Message: message,
			Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
			Suggestions: suggestions,
		}},
	}
}

func TestRequirePostMessageTargetOriginUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&require_post_message_target_origin.RequirePostMessageTargetOriginRule,
		[]rule_tester.ValidTestCase{
			{Code: "window.postMessage(message, targetOrigin)", FileName: "file.js"},
			{Code: "postMessage(message)", FileName: "file.js"},
			{Code: "window.postMessage", FileName: "file.js"},
			{Code: "window.postMessage()", FileName: "file.js"},
			{Code: "window.postMessage(message, targetOrigin, transfer)", FileName: "file.js"},
			{Code: "window.postMessage(...message)", FileName: "file.js"},
			{Code: "window[postMessage](message)", FileName: "file.js"},
			{Code: "window[\"postMessage\"](message)", FileName: "file.js"},
			{Code: "window.notPostMessage(message)", FileName: "file.js"},
			{Code: "window.postMessage?.(message)", FileName: "file.js"},
			// Upstream documentation examples.
			{Code: "window.postMessage(sensitiveData, 'https://trusted-domain.com');", FileName: "file.js"},
			{Code: "window.postMessage({token: authToken}, 'https://api.example.com');", FileName: "file.js"},
			{Code: "window.postMessage({publicData: 'hello'}, '*');", FileName: "file.js"},
			{Code: "iframe.contentWindow.postMessage(data, 'https://expected-iframe-origin.com');", FileName: "file.js"},
		}, []rule_tester.InvalidTestCase{
			invalidCase("window.postMessage(message)", "file.js", 1, 27, 1, 28,
				"window.postMessage(message, window.location.origin)",
				"window.postMessage(message, '*')",
			),
			invalidCase("self.postMessage(message)", "file.js", 1, 25, 1, 26,
				"self.postMessage(message, self.location.origin)",
				"self.postMessage(message, '*')",
			),
			invalidCase("globalThis.postMessage(message)", "file.js", 1, 31, 1, 32,
				"globalThis.postMessage(message, globalThis.location.origin)",
				"globalThis.postMessage(message, '*')",
			),
			invalidCase("foo.postMessage(message )", "file.js", 1, 24, 1, 26,
				"foo.postMessage(message , foo.location.origin)",
				"foo.postMessage(message , self.location.origin)",
				"foo.postMessage(message , '*')",
			),
			invalidCase("foo?.postMessage(message )", "file.js", 1, 25, 1, 27,
				"foo?.postMessage(message , foo.location.origin)",
				"foo?.postMessage(message , self.location.origin)",
				"foo?.postMessage(message , '*')",
			),
			invalidCase("foo.postMessage( ((message)) )", "file.js", 1, 29, 1, 31,
				"foo.postMessage( ((message)) , foo.location.origin)",
				"foo.postMessage( ((message)) , self.location.origin)",
				"foo.postMessage( ((message)) , '*')",
			),
			invalidCase("foo.postMessage(message,)", "file.js", 1, 25, 1, 26,
				"foo.postMessage(message, foo.location.origin,)",
				"foo.postMessage(message, self.location.origin,)",
				"foo.postMessage(message, '*',)",
			),
			invalidCase("foo.postMessage(message , )", "file.js", 1, 26, 1, 28,
				"foo.postMessage(message ,  foo.location.origin,)",
				"foo.postMessage(message ,  self.location.origin,)",
				"foo.postMessage(message ,  '*',)",
			),
			invalidCase("foo.window.postMessage(message)", "file.js", 1, 31, 1, 32,
				"foo.window.postMessage(message, self.location.origin)",
				"foo.window.postMessage(message, '*')",
			),
			invalidCase("document.defaultView.postMessage(message)", "file.js", 1, 41, 1, 42,
				"document.defaultView.postMessage(message, self.location.origin)",
				"document.defaultView.postMessage(message, '*')",
			),
			invalidCase("getWindow().postMessage(message)", "file.js", 1, 32, 1, 33,
				"getWindow().postMessage(message, self.location.origin)",
				"getWindow().postMessage(message, '*')",
			),
			// Upstream documentation examples.
			invalidCase("window.postMessage(sensitiveData);", "file.js", 1, 33, 1, 34,
				"window.postMessage(sensitiveData, window.location.origin);",
				"window.postMessage(sensitiveData, '*');",
			),
			invalidCase("window.postMessage({token: authToken});", "file.js", 1, 38, 1, 39,
				"window.postMessage({token: authToken}, window.location.origin);",
				"window.postMessage({token: authToken}, '*');",
			),
			invalidCase("const iframe = document.querySelector('iframe');\niframe.contentWindow.postMessage(data);", "file.js", 2, 38, 2, 39,
				"const iframe = document.querySelector('iframe');\niframe.contentWindow.postMessage(data, self.location.origin);",
				"const iframe = document.querySelector('iframe');\niframe.contentWindow.postMessage(data, '*');",
			),
		})
}
