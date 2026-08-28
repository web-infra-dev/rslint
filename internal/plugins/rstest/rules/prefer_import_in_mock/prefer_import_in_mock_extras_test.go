// TestPreferImportInMockExtras covers what is specific to Rstest: the two
// names the utilities object is reachable under, the module mock APIs that do
// not accept a promise, the call shapes Rstest's mock transform does and does
// not rewrite, and the trivia the fix has to preserve.
//
// Every shape asserted here was checked against rstest 0.11.8 by mocking a
// real module and observing whether the mock took effect, whether the call
// threw "mock() was not transformed by Rstest", or whether the build failed.
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

			// A name that is neither `rs` nor `rstest` is never rewritten by
			// the transform, however it was bound.
			{Code: `import { rstest as vi } from '@rstest/core';
vi.mock('./sum', () => ({ sum: () => 0 }))`},
			{Code: `import { rs as mocker } from '@rstest/core';
mocker.doMock('./sum')`},
			{Code: `const { rstest: vi } = require('@rstest/core');
vi.mock('./sum')`},

			// Receivers the transform does not rewrite.
			{Code: `import * as core from '@rstest/core';
core.rs.mock('./sum')`},
			{Code: `rs?.mock('./sum')`},
			{Code: `rs.mock?.('./sum')`},
			{Code: `rs['mock']('./sum')`},

			// Positions the call cannot be hoisted out of. Each of these
			// either throws at run time or breaks the build, and wrapping the
			// path repairs neither.
			{Code: `consume(rs.mock('./sum'))`},
			{Code: `const mocked = rs.mock('./sum')`},
			{Code: `rs.mock('./sum'), 0;`},
			{Code: `const register = () => rs.mock('./sum')`},
			{Code: `await rs.mock('./sum')`},

			// Argument lists the transform gives up on.
			{Code: `rs.mock(...args)`},
			{Code: `rs.mock('./sum', ...rest)`},
			{Code: `rs.mock('./sum', () => ({ sum: 1 }), extra)`},

			// An explicit type argument already states the mocked shape.
			{Code: `rs.mock<{ sum: number }>('./sum', () => ({ sum: 1 }))`},
			{Code: `rstest.doMock<{ sum: number }>('./sum', () => ({ sum: 1 }))`},

			// A path that is not written as a quoted string. A template
			// literal fails Rstest's build even without a substitution.
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
			// The transform reads the receiver's spelling, not its origin, so
			// a local binding and a parameter named `rs` are rewritten too.
			{
				Code: `const rs = getMocker();
rs.mock('./sum')`,
				Output: []string{`const rs = getMocker();
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
			{
				Code: `function register(rs) {
  rs.mock('./sum');
}`,
				Output: []string{`function register(rs) {
  rs.mock(import('./sum'));
}`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      2,
					Column:    11,
					EndLine:   2,
					EndColumn: 18,
				}},
			},
			{
				Code: `import { rs } from './helpers';
rs.mock('./sum')`,
				Output: []string{`import { rs } from './helpers';
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
			// Parentheses around the call, the callee, the receiver and the
			// path are all transparent to the transform.
			{
				Code:   `(rs.mock)('./sum')`,
				Output: []string{`(rs.mock)(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code:   `(rs).mock('./sum')`,
				Output: []string{`(rs).mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code:   `rs.mock(('./sum'))`,
				Output: []string{`rs.mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			// TypeScript's type-only syntax is erased before the transform
			// runs, so an asserted receiver, callee and path are all rewritten
			// like the bare form. The assertion on the path goes with the
			// replacement, since it no longer describes a promise.
			{
				Code:   `(rs as any).mock('./sum')`,
				Output: []string{`(rs as any).mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 25,
				}},
			},
			{
				Code:   `rs!.mock('./sum')`,
				Output: []string{`rs!.mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			{
				Code:   `(rs.mock as any)('./sum')`,
				Output: []string{`(rs.mock as any)(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 25,
				}},
			},
			{
				Code:   `rs.mock('./sum' as string)`,
				Output: []string{`rs.mock(import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 26,
				}},
			},
			{
				Code:   `(rs.mock('./sum'));`,
				Output: []string{`(rs.mock(import('./sum')));`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			// A statement nested in a block or a branch is still hoistable.
			{
				Code: `if (isCI) {
  rstest.doMock('./sum');
}`,
				Output: []string{`if (isCI) {
  rstest.doMock(import('./sum'));
}`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferImport",
					Message:   "Replace './sum' with import('./sum')",
					Line:      2,
					Column:    17,
					EndLine:   2,
					EndColumn: 24,
				}},
			},
			// A comment in the space the whole-argument replacement would
			// remove withholds the fix; the diagnostic still stands.
			{
				Code:   `rs.mock((/* keep */ './sum'))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImport", Line: 1, Column: 9}},
			},
			{
				Code:   `rs.mock('./sum' /* keep */ as string)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImport", Line: 1, Column: 9}},
			},
			// A comment outside the replaced span is untouched, so the fix
			// still applies.
			{
				Code:   `rs.mock(/* keep */ './sum')`,
				Output: []string{`rs.mock(/* keep */ import('./sum'))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImport", Line: 1, Column: 20}},
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
