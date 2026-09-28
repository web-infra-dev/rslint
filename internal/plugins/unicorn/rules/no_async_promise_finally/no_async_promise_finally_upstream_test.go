// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-async-promise-finally.js
package no_async_promise_finally_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_async_promise_finally"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noAsyncPromiseFinallyMessage = "Do not pass an async function to `Promise#finally()`."

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		FileName:        "case.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
	}
}

func validTS(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		FileName:        "file.ts",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
	}
}

func invalid(code, reportText string) rule_tester.InvalidTestCase {
	return invalidFile(code, reportText, "case.js")
}

func invalidTS(code, reportText string) rule_tester.InvalidTestCase {
	return invalidFile(code, reportText, "file.ts")
}

func invalidFile(code, reportText, fileName string) rule_tester.InvalidTestCase {
	start := strings.LastIndex(code, reportText)
	if start < 0 {
		panic("report text not found in no-async-promise-finally fixture")
	}
	prefix := code[:start]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := start + 1
	if lastNewline >= 0 {
		column = start - lastNewline
	}

	endPrefix := code[:start+len(reportText)]
	endLine := strings.Count(endPrefix, "\n") + 1
	endLastNewline := strings.LastIndex(endPrefix, "\n")
	endColumn := start + len(reportText) + 1
	if endLastNewline >= 0 {
		endColumn = start + len(reportText) - endLastNewline
	}

	return rule_tester.InvalidTestCase{
		Code:            code,
		FileName:        fileName,
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Output:          []string{},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId:   "no-async-promise-finally",
			Message:     noAsyncPromiseFinallyMessage,
			Line:        line,
			Column:      column,
			EndLine:     endLine,
			EndColumn:   endColumn,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
		}},
	}
}

func TestNoAsyncPromiseFinallyUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		valid("promise.finally(() => {})"),
		valid("promise.finally(() => cleanup())"),
		valid("promise.finally(function () {})"),
		valid("promise.finally(async function * () {})"),
		valid("promise.finally()"),
		valid("promise.finally(undefined)"),
		valid("promise.finally(cleanup)"),
		valid("promise.finally(object.cleanup)"),
		valid("promise.then(async () => {})"),
		valid("promise.catch(async () => {})"),
		valid("finalizer(async () => {})"),
		valid("promise.notFinally(async () => {})"),
		valid("promise[method](async () => {})"),
		valid("const cleanup = () => {}; promise.finally(cleanup);"),
		valid("let cleanup = async () => {}; promise.finally(cleanup);"),
		valid("async function * cleanup() {} promise.finally(cleanup);"),
		valid("const cleanup = async function * () {}; promise.finally(cleanup);"),
		valid("const cleanup = async () => {}; promise.finally(...[cleanup]);"),
		valid(`import {cleanup} from "./cleanup.js"; promise.finally(cleanup);`),
		validTS("function foo(object: {finally(handler: () => Promise<void>): void}) { object.finally(async () => {}); }"),
		validTS("function foo(object: {finally(handler: () => Promise<void>): void}) { const cleanup = async () => {}; object.finally(cleanup); }"),

		// Documentation examples.
		valid("promise.finally(() => {\n\tcleanupSynchronously();\n});"),
		valid("function cleanup() {}\n\npromise.finally(cleanup);"),
		valid("promise.finally(() => {\n\treturn cleanup().catch(error => {\n\t\tlogCleanupError(error);\n\t});\n});"),
	}

	invalidCases := []rule_tester.InvalidTestCase{
		invalid("promise.finally(async () => {})", "async () => {}"),
		invalid("promise.finally(async () => cleanup())", "async () => cleanup()"),
		invalid("promise.finally(async () => { await cleanup(); })", "async () => { await cleanup(); }"),
		invalid("promise.finally(async function () { await cleanup(); })", "async function () { await cleanup(); }"),
		invalid("Promise.resolve(value).finally(async () => cleanup())", "async () => cleanup()"),
		invalid("new Promise(resolve => resolve()).finally(async () => cleanup())", "async () => cleanup()"),
		invalid("promise?.finally(async () => {})", "async () => {}"),
		invalid("promise.finally?.(async () => {})", "async () => {}"),
		invalid(`promise["finally"](async () => {})`, "async () => {}"),
		invalid("promise[`finally`](async () => {})", "async () => {}"),
		invalid(`const method = "finally"; promise[method](async () => {});`, "async () => {}"),
		invalid("async function cleanup() {} promise.finally(cleanup);", "cleanup"),
		invalid("const cleanup = async () => {}; promise.finally(cleanup);", "cleanup"),
		invalid("const cleanup = async function () {}; promise.finally(cleanup);", "cleanup"),
		invalidTS("type Callback = () => void; promise.finally((async () => {}) as Callback);", "(async () => {}) as Callback"),
		invalidTS("type Callback = () => void; const cleanup = (async () => {}) as Callback; promise.finally(cleanup);", "cleanup"),
		invalidTS("function foo(promise: Promise<string>) { promise.finally(async () => {}); }", "async () => {}"),

		// Documentation examples.
		invalid("promise.finally(async () => {\n\tawait cleanup();\n});", "async () => {\n\tawait cleanup();\n}"),
		invalid("async function cleanup() {}\n\npromise.finally(cleanup);", "cleanup"),
		invalid(
			"const original = new Error('original');\nconst cleanup = new Error('cleanup');\n\ntry {\n\tawait Promise.reject(original).finally(async () => {\n\t\tawait Promise.reject(cleanup);\n\t});\n} catch (error) {\n\tconsole.log(error.message);\n\t//=> 'cleanup'\n}",
			"async () => {\n\t\tawait Promise.reject(cleanup);\n\t}",
		),
	}

	if len(validCases) != 24 || len(invalidCases) != 20 {
		t.Fatalf("coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_async_promise_finally.NoAsyncPromiseFinallyRule,
		validCases,
		invalidCases,
	)
}
