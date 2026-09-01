// TestPreferImportInMockUpstream migrates the @vitest/eslint-plugin@v1.6.27
// prefer-import-in-mock suite. The cases that drive the vitest namespace
// through a renamed import are dropped rather than translated: a renamed
// `rstest` binding does not reach Rstest's mock transform at all. They are
// replaced by their Rstest counterpart in
// prefer_import_in_mock_extras_test.go, where the renamed binding is valid.
package prefer_import_in_mock

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferImportInMockUpstreamWithoutFix(t *testing.T) {
	unfixable := []any{map[string]any{"fixable": false}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferImportInMockRule,
		[]rule_tester.ValidTestCase{
			{Code: `rs.mock(import("foo"))`},
			{Code: `rs.mock(import("node:fs/promises"))`},
			{Code: `rs.mock(import("./foo.js"), () => ({ Foo: rs.fn() }))`},
			{Code: `rs.mock(import("./foo.js"), { spy: true });`},
			{Code: `rs.doMock(import("foo"))`},
			{Code: `rs.doMock(import("node:fs/promises"))`},
			{Code: `rs.doMock(import("./foo.js"), () => ({ Foo: rs.fn() }))`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `rs.mock('foo', () => {})`,
				Options: unfixable,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace 'foo' with import('foo')",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code:    `rs.mock("node:fs/promises")`,
				Options: unfixable,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "node:fs/promises" with import("node:fs/promises")`,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 27,
				}},
			},
			{
				Code:    `rs.mock("./foo.js", () => ({ Foo: rs.fn() }))`,
				Options: unfixable,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "./foo.js" with import("./foo.js")`,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code:    `rs.doMock('foo', () => {})`,
				Options: unfixable,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace 'foo' with import('foo')",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code:    `rs.doMock("node:fs/promises")`,
				Options: unfixable,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "node:fs/promises" with import("node:fs/promises")`,
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code:    `rs.doMock("./foo.js", () => ({ Foo: rs.fn() }))`,
				Options: unfixable,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "./foo.js" with import("./foo.js")`,
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
		},
	)
}

func TestPreferImportInMockUpstreamWithFix(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferImportInMockRule,
		[]rule_tester.ValidTestCase{
			{Code: `rs.mock(import("foo"))`},
			{Code: `rs.mock(import("node:fs/promises"))`},
			{Code: `rs.mock(import("./foo.js"), () => ({ Foo: rs.fn() }))`},
			{Code: `rs.mock(import("./foo.js"), { spy: true });`},
			{Code: `rs.doMock(import("foo"))`},
			{Code: `rs.doMock(import("node:fs/promises"))`},
			{Code: `rs.doMock(import("./foo.js"), () => ({ Foo: rs.fn() }))`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `rs.mock('foo', () => {})`,
				Output: []string{`rs.mock(import('foo'), () => {})`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace 'foo' with import('foo')",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code:   `rs.mock("node:fs/promises")`,
				Output: []string{`rs.mock(import("node:fs/promises"))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "node:fs/promises" with import("node:fs/promises")`,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 27,
				}},
			},
			{
				Code:   `rs.mock("./foo.js", () => ({ Foo: rs.fn() }))`,
				Output: []string{`rs.mock(import("./foo.js"), () => ({ Foo: rs.fn() }))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "./foo.js" with import("./foo.js")`,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code:   `rs.doMock('foo', () => {})`,
				Output: []string{`rs.doMock(import('foo'), () => {})`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace 'foo' with import('foo')",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code:   `rs.doMock("node:fs/promises")`,
				Output: []string{`rs.doMock(import("node:fs/promises"))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "node:fs/promises" with import("node:fs/promises")`,
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code:   `rs.doMock("./foo.js", () => ({ Foo: rs.fn() }))`,
				Output: []string{`rs.doMock(import("./foo.js"), () => ({ Foo: rs.fn() }))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "./foo.js" with import("./foo.js")`,
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
		},
	)
}
