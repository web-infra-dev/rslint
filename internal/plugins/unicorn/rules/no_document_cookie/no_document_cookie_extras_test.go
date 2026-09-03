// TestNoDocumentCookieExtras covers ReferenceTracker branches, tsgo edge
// shapes, and real-user scenarios absent from the upstream suite. The exact
// upstream migration lives in no_document_cookie_upstream_test.go.
package no_document_cookie_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_document_cookie"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDocumentCookieExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_document_cookie.NoDocumentCookieRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream filter(): reads and deletes are not assignments.
			extraValid(`console.log(document.cookie)`),
			extraValid(`delete document.cookie`),

			// Locks in ReferenceTracker's dynamic-key rejection branch.
			extraValid(`const key = "cookie"; document[key] = "foo=bar"`),
			extraValid(`document[0] = "foo=bar"`),
			extraValid(`document[Symbol.cookie] = "foo=bar"`),

			// Locks in ReferenceTracker's modified-global root guard.
			extraValid(`document = replacement; document.cookie = "foo=bar"`),
			extraValid(`document.cookie = "foo=bar"; document = replacement`),
			extraValid(`window = replacement; window.document.cookie = "foo=bar"`),

			// Locks in global-reference checks for local shadowing and disabled globals.
			extraValid(`function set(document) { document.cookie = "foo=bar" }`),
			{
				Code:     `document.cookie = "foo=bar"`,
				FileName: "file.mjs",
				Globals:  map[string]any{"document": "off"},
			},

			// Locks in SequenceExpression's non-final branch.
			extraValid(`const doc = (document, replacement); doc.cookie = "foo=bar"`),

			// ---- Dimension 4: optional access is a read and cannot be an assignment target ----
			extraValid(`const value = document?.cookie`),

			// ---- Dimension 4: full-expression TS wrappers fail upstream's direct-parent filter ----
			extraValid(`(document.cookie as string) = "foo=bar"`),
			extraValid(`(document.cookie)! = "foo=bar"`),

			// ---- Dimension 4: empty/rest binding patterns degrade gracefully ----
			extraValid(`const {} = globalThis`),
			extraValid(`const {...rest} = globalThis; rest.cookie = "foo=bar"`),

			// ---- Real-user: issue #1542 reads document.cookie in logging code ----
			extraValid(`console.log("cookie", document.cookie)`),

			// N/A: private keys; private identifiers cannot be used on document.
			// N/A: function/class container variants; the rule targets assignment expressions.
			// N/A: body-absent declarations; no declaration body is inspected.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver and full-expression wrappers ----
			extraInvalid(`(document).cookie = "foo=bar"`, `(document).cookie`),
			extraInvalid(`((document)).cookie = "foo=bar"`, `((document)).cookie`),
			extraInvalid(`(document.cookie) = "foo=bar"`, `document.cookie`),

			// ---- Dimension 4: TypeScript receiver wrappers are transparent ----
			extraInvalid(`document!.cookie = "foo=bar"`, `document!.cookie`),
			extraInvalid(`(document as Document).cookie = "foo=bar"`, `(document as Document).cookie`),
			extraInvalid(`(document satisfies Document).cookie = "foo=bar"`, `(document satisfies Document).cookie`),

			// ---- Dimension 4: static element-access key forms are tracked ----
			extraInvalid(`document["cookie"] = "foo=bar"`, `document["cookie"]`),
			extraInvalid("document[`cookie`] = \"foo=bar\"", "document[`cookie`]"),

			// Locks in ReferenceTracker's assignment-alias branch.
			extraInvalid(`let doc; doc = document; doc.cookie = "foo=bar"`, `doc.cookie`),

			// Locks in ReferenceTracker's object-pattern alias branch.
			extraInvalid(`const {document: doc} = globalThis; doc.cookie = "foo=bar"`, `doc.cookie`),
			extraInvalid(`let doc; ({document: doc} = globalThis); doc.cookie = "foo=bar"`, `doc.cookie`),

			// Locks in ReferenceTracker's parameter/default AssignmentPattern branch.
			extraInvalid(`function set(doc = document) { doc.cookie = "foo=bar" }`, `doc.cookie`),

			// Locks in ReferenceTracker's conditional/logical/pass-through branches.
			extraInvalid(`const doc = flag ? document : replacement; doc.cookie = "foo=bar"`, `doc.cookie`),
			extraInvalid(`const doc = document || replacement; doc.cookie = "foo=bar"`, `doc.cookie`),
			extraInvalid(`const doc = (0, document); doc.cookie = "foo=bar"`, `doc.cookie`),

			// Locks in chained local aliases and all supported global-object roots.
			extraInvalid(`const first = document; const second = first; second.cookie = "foo=bar"`, `second.cookie`),
			extraInvalid(`self.document.cookie = "foo=bar"`, `self.document.cookie`),
			extraInvalid(`global.document.cookie = "foo=bar"`, `global.document.cookie`),

			// Locks in ReferenceTracker's per-root traversal: converging roots report twice.
			extraInvalidCount(
				`const doc = flag ? document : window.document; doc.cookie = "foo=bar"`,
				`doc.cookie`,
				2,
			),

			// ---- Dimension 4: same-kind nesting keeps each alias scope independent ----
			extraInvalid(
				"const doc = document;\nfunction set() {\n\tconst nested = doc;\n\tnested.cookie = \"foo=bar\";\n}",
				`nested.cookie`,
			),

			// ---- Real-user: issue #2301 library-generated cookie strings still assign directly ----
			extraInvalid(`document.cookie = setCookie(MyCookie, {foo: "baz"})`, `document.cookie`),

			// ---- Real-user: PR #1833 added window.document tracking ----
			extraInvalid(`window["document"]["cookie"] = "foo=bar"`, `window["document"]["cookie"]`),
		},
	)
}

func extraValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:     code,
		FileName: extraFileName(code),
		Globals:  noDocumentCookieGlobals(),
	}
}

func extraInvalid(code, target string) rule_tester.InvalidTestCase {
	return extraInvalidCount(code, target, 1)
}

func extraInvalidCount(code, target string, count int) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, count)
	for index := range count {
		errors[index] = expectedError(code, target, 0)
	}
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: extraFileName(code),
		Globals:  noDocumentCookieGlobals(),
		Errors:   errors,
	}
}

func extraFileName(code string) string {
	for _, marker := range []string{" as ", " satisfies ", "!.", ")!"} {
		if strings.Contains(code, marker) {
			return "file.ts"
		}
	}
	return "file.mjs"
}
