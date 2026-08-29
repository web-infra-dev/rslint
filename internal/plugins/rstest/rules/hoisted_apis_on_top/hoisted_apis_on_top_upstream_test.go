// TestHoistedApisOnTopUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 hoisted-apis-on-top suite
// (tests/hoisted-apis-on-top.test.ts) 1:1, with the receiver rewritten to the
// Rstest utilities object. Two upstream cases change: the alias cases
// (`import { vi as v }`) become valid, because Rstest matches the receiver by
// the name written at the call site and leaves a renamed binding untouched;
// and `unmock` gains the non-hoisted-counterpart suggestion, since Rstest
// exposes `doUnmock`. Position assertions cover line/column for every invalid
// case. Rstest API surface, receiver shapes, fix boundaries and edit-demand
// coverage live in hoisted_apis_on_top_extras_test.go.
package hoisted_apis_on_top

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHoistedApisOnTopUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&HoistedApisOnTopRule,
		[]rule_tester.ValidTestCase{
			// ---- already at the top of the file ----
			{Code: `rs.mock();`},
			{Code: `
rs.hoisted();
import foo from 'bar';
`},
			{Code: `
import foo from 'bar';
rs.unmock(baz);
`},
			{Code: `const foo = await rs.hoisted(async () => {});`},
			{Code: `
import { rstest } from '@rstest/core';
rstest.mock('./foo');
`},

			// ---- the receiver's name is what the build matches ----
			// DIVERGENCE: upstream reports this, tracking `r` as an alias of
			// the utilities object. Rstest never rewrites a renamed binding —
			// the call throws where it is written instead of being lifted — so
			// it is an ordinary runtime call wherever it sits.
			{Code: `
import { rs as r } from '@rstest/core';
if (foo) {
  r.mock('./foo');
}
`},
		},
		[]rule_tester.InvalidTestCase{
			// DIVERGENCE: upstream treats a `rs` bound by an import from
			// another package as unrelated and stays silent. Rstest lifts the
			// call anyway, because the rewrite matches the receiver by the name
			// written at the call site.
			{
				Code: `
import { rs } from 'some-other-module';
if (foo) {
  rs.mock('./foo');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
import { rs } from 'some-other-module';
rs.mock('./foo');
if (foo) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
import { rs } from 'some-other-module';
if (foo) {
  rs.doMock('./foo');
}
`,
						},
					},
				}},
			},

			// ---- a block is a runtime location ----
			{
				Code: `
if (foo) {
  rs.mock('foo', () => {});
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Message:   "Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `rs.mock('foo', () => {});

if (foo) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
if (foo) {
  rs.doMock('foo', () => {});
}
`,
						},
					},
				}},
			},
			{
				Code: `
import foo from 'bar';

if (foo) {
  rs.hoisted();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestMoveHoistedApiToTop",
						Output: `
import foo from 'bar';
rs.hoisted();

if (foo) {
  ;
}
`,
					}},
				}},
			},
			{
				Code: `
import foo from 'bar';

if (foo) {
  rs.unmock();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
import foo from 'bar';
rs.unmock();

if (foo) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
import foo from 'bar';

if (foo) {
  rs.doUnmock();
}
`,
						},
					},
				}},
			},
			{
				Code: `
import foo from 'bar';

if (foo) {
  rs.mock();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
import foo from 'bar';
rs.mock();

if (foo) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
import foo from 'bar';

if (foo) {
  rs.doMock();
}
`,
						},
					},
				}},
			},

			// ---- the insertion point is the last import in the file ----
			{
				Code: `
if (shouldMock) {
  rs.mock(import('something'), () => bar);
}

import something from 'something';
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
if (shouldMock) {
  ;
}

import something from 'something';
rs.mock(import('something'), () => bar);
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
if (shouldMock) {
  rs.doMock(import('something'), () => bar);
}

import something from 'something';
`,
						},
					},
				}},
			},

			// ---- the other spelling of the utilities object ----
			{
				Code: `
import { rstest } from '@rstest/core';

if (condition) {
  rstest.mock('./foo');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
import { rstest } from '@rstest/core';
rstest.mock('./foo');

if (condition) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
import { rstest } from '@rstest/core';

if (condition) {
  rstest.doMock('./foo');
}
`,
						},
					},
				}},
			},
			{
				Code: `
import { rs } from '@rstest/core';

if (condition) {
  rs.hoisted(() => {});
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestMoveHoistedApiToTop",
						Output: `
import { rs } from '@rstest/core';
rs.hoisted(() => {});

if (condition) {
  ;
}
`,
					}},
				}},
			},
			{
				Code: `
import { rstest } from '@rstest/core';

if (condition) {
  rstest.unmock('./foo');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
import { rstest } from '@rstest/core';
rstest.unmock('./foo');

if (condition) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
import { rstest } from '@rstest/core';

if (condition) {
  rstest.doUnmock('./foo');
}
`,
						},
					},
				}},
			},
			{
				Code: `
import { rs } from '@rstest/core';

if (condition) {
  rs.mock('./foo');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hoistedApisOnTop",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestMoveHoistedApiToTop",
							Output: `
import { rs } from '@rstest/core';
rs.mock('./foo');

if (condition) {
  ;
}
`,
						},
						{
							MessageId: "suggestUseNonHoistedApi",
							Output: `
import { rs } from '@rstest/core';

if (condition) {
  rs.doMock('./foo');
}
`,
						},
					},
				}},
			},
		},
	)
}
