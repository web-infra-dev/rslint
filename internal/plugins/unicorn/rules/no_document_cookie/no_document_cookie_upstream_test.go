// TestNoDocumentCookieUpstream migrates the complete valid/invalid suite from
// upstream test/no-document-cookie.js at v74.0.0. Position assertions cover
// every invalid case. rslint-specific cases live in
// no_document_cookie_extras_test.go.
package no_document_cookie_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_document_cookie"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID   = "no-document-cookie"
	messageText = "Do not use `document.cookie` directly."
)

func TestNoDocumentCookieUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_document_cookie.NoDocumentCookieRule,
		[]rule_tester.ValidTestCase{
			upstreamValid(`document.cookie`),
			upstreamValid(`const foo = document.cookie`),
			upstreamValid(`foo = document.cookie`),
			upstreamValid(`foo = document?.cookie`),
			upstreamValid(`foo = document.cookie + ";foo=bar"`),
			upstreamValid(`delete document.cookie`),
			upstreamValid(`if (document.cookie.includes("foo")){}`),
			upstreamValid(`Object.assign(document, {cookie: "foo=bar"})`),
			upstreamValid(`document[CONSTANTS_COOKIE] = "foo=bar"`),
			upstreamValid(`document[cookie] = "foo=bar"`),
			upstreamValid("const CONSTANTS_COOKIE = \"cookie\";\ndocument[CONSTANTS_COOKIE] = \"foo=bar\";"),
		},
		[]rule_tester.InvalidTestCase{
			upstreamInvalid(`document.cookie = "foo=bar"`, `document.cookie`),
			upstreamInvalid(`document.cookie += ";foo=bar"`, `document.cookie`),
			upstreamInvalid(`document.cookie = document.cookie + ";foo=bar"`, `document.cookie`),
			upstreamInvalid(`document.cookie &&= true`, `document.cookie`),
			upstreamInvalid(`document["coo" + "kie"] = "foo=bar"`, `document["coo" + "kie"]`),
			upstreamInvalid(`foo = document.cookie = "foo=bar"`, `document.cookie`),
			upstreamInvalid(`var doc = document; doc.cookie = "foo=bar"`, `doc.cookie`),
			upstreamInvalid(`let doc = document; doc.cookie = "foo=bar"`, `doc.cookie`),
			upstreamInvalid(`const doc = globalThis.document; doc.cookie = "foo=bar"`, `doc.cookie`),
			upstreamInvalid(`window.document.cookie = "foo=bar"`, `window.document.cookie`),
		},
	)
}

func upstreamValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Globals:  noDocumentCookieGlobals(),
	}
}

func upstreamInvalid(code, target string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Globals:  noDocumentCookieGlobals(),
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, 0),
		},
	}
}

func noDocumentCookieGlobals() map[string]any {
	return map[string]any{
		"document": "readonly",
		"global":   "readonly",
		"self":     "readonly",
		"window":   "readonly",
	}
}

func expectedError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
	offset := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("target not found in no-document-cookie test: " + target)
		}
		offset = searchFrom + relative
		searchFrom = offset + len(target)
	}

	line, column := lineColumn(code, offset)
	endLine, endColumn := lineColumn(code, offset+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   messageText,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func lineColumn(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index, character := range code {
		if index >= offset {
			break
		}
		if character == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
