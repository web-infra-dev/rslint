// TestPreferBlobReadingMethodsUpstream migrates the full valid/invalid suite
// from upstream test/prefer-blob-reading-methods.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in prefer_blob_reading_methods_extras_test.go.
package prefer_blob_reading_methods_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_blob_reading_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "error"

func expectedMessage(method string) string {
	replacement := "text"
	if method == "readAsArrayBuffer" {
		replacement = "arrayBuffer"
	}
	return fmt.Sprintf("Prefer `Blob#%s()` over `FileReader#%s(blob)`.", replacement, method)
}

func methodError(code, method string, occurrence int) rule_tester.InvalidTestCaseError {
	offset := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], method)
		if relative < 0 {
			panic("method not found in prefer-blob-reading-methods test: " + method)
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

	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   expectedMessage(method),
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(method),
	}
}

func TestPreferBlobReadingMethodsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_blob_reading_methods.PreferBlobReadingMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: `blob.arrayBuffer()`, FileName: "file.mjs"},
			{Code: `blob.text()`, FileName: "file.mjs"},
			{Code: `new Response(blob).arrayBuffer()`, FileName: "file.mjs"},
			{Code: `new Response(blob).text()`, FileName: "file.mjs"},
			{Code: `fileReader.readAsDataURL(blob)`, FileName: "file.mjs"},
			{Code: `fileReader.readAsBinaryString(blob)`, FileName: "file.mjs"},
			{Code: `fileReader.readAsText(blob, "ascii")`, FileName: "file.mjs"},
			{Code: `fileReader.readAsArrayBuffer(blob, extraArg)`, FileName: "file.mjs"},
			{Code: `fileReader?.readAsArrayBuffer(blob)`, FileName: "file.mjs"},
			{Code: `fileReader.readAsText?.(blob)`, FileName: "file.mjs"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `fileReader.readAsArrayBuffer(blob)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					methodError(`fileReader.readAsArrayBuffer(blob)`, "readAsArrayBuffer", 0),
				},
			},
			{
				Code:     `fileReader.readAsText(blob)`,
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					methodError(`fileReader.readAsText(blob)`, "readAsText", 0),
				},
			},
		},
	)
}
