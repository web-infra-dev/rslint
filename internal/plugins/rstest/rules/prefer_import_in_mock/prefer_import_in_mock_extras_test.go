// TestPreferImportInMockExtras covers what is specific to Rstest: the two
// names the utilities object is reachable under, the module mock APIs that do
// not accept a promise, the call shapes Rstest's mock transform does not
// rewrite, and the trivia the fix has to preserve.
package prefer_import_in_mock

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferImportInMockExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferImportInMockRule,
		[]rule_tester.ValidTestCase{
			// `mockRequire` and `doMockRequire` mock the CommonJS entry and
			// take a string module name only.
			{Code: `rs.mockRequire('./sum')`},
			{Code: `rs.doMockRequire('./sum')`},
			{Code: `rstest.mockRequire('./sum')`},
			{Code: `rstest.doMockRequire('./sum')`},

			// Every other module API takes a plain path.
			{Code: `rs.unmock('./sum')`},
			{Code: `rs.doUnmock('./sum')`},
			{Code: `rs.unmockRequire('./sum')`},
			{Code: `rs.doUnmockRequire('./sum')`},
			{Code: `await rs.importActual('./sum')`},
			{Code: `await rs.importMock('./sum')`},
			{Code: `rs.requireActual('./sum')`},
			{Code: `rs.requireMock('./sum')`},

			// A renamed binding never reaches the mock transform, so the call
			// is broken for a reason this rule does not fix.
			{Code: `import { rstest as vi } from '@rstest/core';
vi.mock('./sum', () => ({ sum: () => 0 }))`},
			{Code: `import { rs as mocker } from '@rstest/core';
mocker.doMock('./sum')`},
			{Code: `const { rstest: vi } = require('@rstest/core');
vi.mock('./sum')`},

			// `rs` from somewhere else, and a local binding that shadows the
			// global, are not the Rstest utilities object.
			{Code: `import { rs } from './helpers';
rs.mock('./sum')`},
			{Code: `const rs = createMocker();
rs.mock('./sum')`},
			{Code: `function register(rs) { rs.mock('./sum'); }`},

			// A namespace import reaches the utilities object through a
			// property, which the mock transform does not rewrite.
			{Code: `import * as core from '@rstest/core';
core.rs.mock('./sum')`},

			// Call shapes the mock transform does not rewrite.
			{Code: `rs?.mock('./sum')`},
			{Code: `rs.mock?.('./sum')`},
			{Code: `rs['mock']('./sum')`},

			// An explicit type argument already states the mocked shape.
			{Code: `rs.mock<{ sum: number }>('./sum', () => ({ sum: 1 }))`},
			{Code: `rstest.doMock<{ sum: number }>('./sum', () => ({ sum: 1 }))`},

			// A path that is not written as a quoted string.
			{Code: "rs.mock(`./sum`)"},
			{Code: `rs.mock(modulePath)`},
			{Code: `rs.mock('./' + name)`},
			{Code: `rs.mock()`},

			// Already an import.
			{Code: `rstest.mock(import('./sum'), { mock: true })`},
			{Code: `import { rs } from '@rstest/core';
rs.mock(import('./sum'))`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import { rs } from '@rstest/core';
rs.mock('./sum', () => ({ sum: () => 0 }))`,
				Output: []string{`import { rs } from '@rstest/core';
rs.mock(import('./sum'), () => ({ sum: () => 0 }))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      2,
					Column:    9,
					EndLine:   2,
					EndColumn: 16,
				}},
			},
			{
				Code: `import { rstest } from '@rstest/core';
rstest.doMock('./sum', { spy: true })`,
				Output: []string{`import { rstest } from '@rstest/core';
rstest.doMock(import('./sum'), { spy: true })`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 22,
				}},
			},
			{
				Code: `const { rs } = require('@rstest/core');
rs.mock('./sum')`,
				Output: []string{`const { rs } = require('@rstest/core');
rs.mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      2,
					Column:    9,
					EndLine:   2,
					EndColumn: 16,
				}},
			},
			// The utilities object is exported twice, so a binding renamed
			// from one of the two names onto the other still reaches the
			// transform.
			{
				Code: `import { rs as rstest } from '@rstest/core';
rstest.mock('./sum')`,
				Output: []string{`import { rs as rstest } from '@rstest/core';
rstest.mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 20,
				}},
			},
			// The path is wrapped as written, so a quote inside it and the
			// file's quote style both survive.
			{
				Code:   `rs.mock("./it's")`,
				Output: []string{`rs.mock(import("./it's"))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace "./it's" with import("./it's")`,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			{
				Code:   `rs.mock('./a')`,
				Output: []string{`rs.mock(import('./a'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   `Replace './a' with import('./a')`,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			// Each mocked module is reported on its own.
			{
				Code: `rs.mock('./sum');
rstest.doMock('./product');`,
				Output: []string{`rs.mock(import('./sum'));
rstest.doMock(import('./product'));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferImport",
						Message:   "Replace './sum' with import('./sum')",
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 16,
					},
					{
						MessageId: "preferImport",
						Message:   "Replace './product' with import('./product')",
						Line:      2,
						Column:    15,
						EndLine:   2,
						EndColumn: 26,
					},
				},
			},
		},
	)
}
