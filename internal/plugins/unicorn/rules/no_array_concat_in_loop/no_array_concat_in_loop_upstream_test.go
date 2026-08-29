// TestNoArrayConcatInLoopUpstream migrates the full valid/invalid suite from
// upstream test/no-array-concat-in-loop.js at eslint-plugin-unicorn v73.0.0
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in no_array_concat_in_loop_extras_test.go.
package no_array_concat_in_loop_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_concat_in_loop"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	noArrayConcatInLoopMessageID = "no-array-concat-in-loop"
	noArrayConcatInLoopMessage   = "Do not use `Array#concat()` to accumulate an array in a loop."
)

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		FileName:        "file.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
	}
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		FileName:        "file.ts",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
	}
}

func concatError(code string, occurrence int) rule_tester.InvalidTestCaseError {
	const target = "concat"
	index := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("concat target not found in test case")
		}
		index = searchFrom + relative
		searchFrom = index + len(target)
	}
	line := 1 + strings.Count(code[:index], "\n")
	lineStart := strings.LastIndex(code[:index], "\n") + 1
	column := index - lineStart + 1
	return rule_tester.InvalidTestCaseError{
		MessageId: noArrayConcatInLoopMessageID,
		Message:   noArrayConcatInLoopMessage,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(target),
	}
}

func invalidConcat(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:            code,
		FileName:        "file.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Errors:          []rule_tester.InvalidTestCaseError{concatError(code, 0)},
	}
}

func invalidConcatTS(code string) rule_tester.InvalidTestCase {
	testCase := invalidConcat(code)
	testCase.FileName = "file.ts"
	testCase.LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	return testCase
}

func TestNoArrayConcatInLoopUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_concat_in_loop.NoArrayConcatInLoopRule,
		[]rule_tester.ValidTestCase{
			// ---- Upstream snapshot: valid ----
			jsValid(`let result = [];
result = result.concat(chunk);`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	result = other.concat(chunk);
}`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	other = result.concat(chunk);
}`),
			jsValid(`let result = [initial];
for (const chunk of chunks) {
	result = result.concat(chunk);
}`),
			jsValid(`let result;
for (const chunk of chunks) {
	result = result.concat(chunk);
}`),
			jsValid(`const result = [];
for (const chunk of chunks) {
	result = result.concat(chunk);
}`),
			jsValid(`const text = '';
for (const part of parts) {
	text.concat(part);
}`),
			jsValid(`let text = '';
for (const part of parts) {
	text = text.concat(part);
}`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	result = result?.concat(chunk);
}`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	result = result['concat'](chunk);
}`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	result = result.concat(chunk).filter(Boolean);
}`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	result = result.concat();
}`),
			jsValid(`for (const chunk of chunks) {
	let result = [];
	result = result.concat(chunk);
}`),
			jsValid(`let result = [];
for (const chunk of chunks) {
	function append() {
		result = result.concat(chunk);
	}
}`),
			jsValid(`let result = [];
const append = () => {
	result = result.concat(chunk);
};

for (const chunk of chunks) {
	append(chunk);
}`),
			jsValid(`this.result = [];
for (const chunk of chunks) {
	this.result = this.result.concat(chunk);
}`),
			jsValid(`const result = chunks.reduce((result, chunk) => result.concat(chunk), []);`),
			tsValid(`let result = ['initial'] as string[];
for (const chunk of chunks) {
	result = (result as string[]).concat(chunk);
}`),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Upstream snapshot: invalid ----
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	result = result.concat(chunk);
}`),
			invalidConcat(`let result = [];
for (let index = 0; index < chunks.length; index++) {
	result = result.concat(chunks[index]);
}`),
			invalidConcat(`let result = [];
for (const index in chunks) {
	result = result.concat(chunks[index]);
}`),
			invalidConcat(`let result = [];
while (chunks.length > 0) {
	result = result.concat(chunks.pop());
}`),
			invalidConcat(`let result = [];
do {
	result = result.concat(getChunk());
} while (hasMoreChunks());`),
			invalidConcat(`let result = [];
for (let index = 0; index < chunks.length; result = result.concat(chunks[index++])) {}`),
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	result = (result.concat(chunk));
}`),
			invalidConcat(`var result = [];
for (const chunk of chunks) {
	result = result.concat(chunk);
}`),
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	(result) = (result).concat(chunk);
}`),
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	result = result.concat(first, second);
}`),
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	result = result.concat(...chunkGroups);
}`),
			invalidConcatTS(`let result = [] as string[];
for (const chunk of chunks) {
	result = (result as string[]).concat(chunk);
}`),
			invalidConcatTS(`let result = [] satisfies string[];
for (const chunk of chunks) {
	result = result!.concat(chunk);
}`),
			invalidConcatTS(`let result = <string[]>[];
for (const chunk of chunks) {
	result = (<string[]>result).concat(chunk);
}`),
			invalidConcat(`for (let result = []; condition; result = result.concat(chunk)) {}`),
			invalidConcat(`for (let result = []; condition;) {
	result = result.concat(chunk);
}`),
		},
	)
}
