package no_interpolation_in_snapshots_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_interpolation_in_snapshots"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInterpolationInSnapshotsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_interpolation_in_snapshots.NoInterpolationInSnapshotsRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect("something").toEqual("else");`},
			{Code: `expect(something).toMatchInlineSnapshot();`},
			{Code: "expect(something).toMatchInlineSnapshot(`No interpolation`);"},
			{Code: "expect(something).toMatchInlineSnapshot({}, `No interpolation`);"},
			{Code: `expect(something);`},
			{Code: `expect(something).not;`},
			{Code: `expect.hasAssertions();`},
			{Code: "myObjectWants.toMatchInlineSnapshot({}, `${interpolated}`);"},
			{Code: "myObjectWants.toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);"},
			{Code: "toMatchInlineSnapshot({}, `${interpolated}`);"},
			{Code: "toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);"},
			{Code: `expect(something).toThrowErrorMatchingInlineSnapshot();`},
			{Code: "expect(something).toThrowErrorMatchingInlineSnapshot(`No interpolation`);"},
			{Code: "expect(something).toThrowErrorMatchingInlineSnapshot(`No interpolation`, `case ${id}`);"},
			{Code: "expect(something).toMatchInlineSnapshot({}, `No interpolation`, `case ${id}`);"},
			{Code: "expect(something).toMatchInlineSnapshot(`No interpolation`, `case ${id}`);"},
			{Code: "const snapshot: string = `No interpolation`;\nexpect(something).toMatchInlineSnapshot(snapshot, `case ${id}`);"},

			// Snapshot matchers whose expected value lives outside the source
			// file: updating them never rewrites the argument, so interpolation
			// is legitimate.
			{Code: "expect(something).toMatchSnapshot(`${interpolated}`);"},
			{Code: "expect(something).matchSnapshot(`${interpolated}`);"},
			{Code: `expect(something).toThrowErrorMatchingSnapshot();`},
			// toMatchFileSnapshot is Rstest-only and takes a *path* as its first
			// argument — interpolating the file name is the intended usage. This
			// case pins down that the rule consults
			// RSTEST_INLINE_SNAPSHOT_MATCHERS and not RSTEST_SNAPSHOT_MATCHERS,
			// which does contain toMatchFileSnapshot.
			{Code: "test(\"case\", async () => {\n  await expect(data).toMatchFileSnapshot(`./__snapshots__/${name}.json`);\n});"},

			// Not a Rstest expect.
			{Code: "import { expect } from \"vitest\";\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);"},
			{Code: "import { expect } from \"@jest/globals\";\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);"},
			{Code: "const expect = createAssertionLibrary();\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);"},

			// A property-style Chai assertion has no arguments to interpolate into.
			{Code: `expect(something).to.be.ok;`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "expect(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    41,
						EndColumn: 58,
					},
				},
			},
			{
				Code: "expect(something).not.toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    45,
						EndColumn: 62,
					},
				},
			},
			{
				Code: "expect(something).toMatchInlineSnapshot({}, `${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    45,
						EndColumn: 62,
					},
				},
			},
			{
				Code: "expect(something).not.toMatchInlineSnapshot({}, `${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    49,
						EndColumn: 66,
					},
				},
			},
			{
				Code: "expect(something).toThrowErrorMatchingInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    54,
						EndColumn: 71,
					},
				},
			},
			{
				Code: "expect(something).not.toThrowErrorMatchingInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    58,
						EndColumn: 75,
					},
				},
			},
			{
				Code: "expect(something).toMatchInlineSnapshot(`${first}`, `${second}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    41,
						EndColumn: 51,
					},
				},
			},
			{
				Code: "expect(something).toMatchInlineSnapshot({}, `${snapshot}`, `case ${id}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    45,
						EndColumn: 58,
					},
				},
			},
			{
				Code: "const properties = {};\nexpect(something).toMatchInlineSnapshot(properties, `${snapshot}`, `case ${id}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Line:      2,
					},
				},
			},
			{
				Code: "expect(something).toThrowErrorMatchingInlineSnapshot(`${snapshot}`, `case ${id}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    54,
						EndColumn: 67,
					},
				},
			},
			{
				Code: "expect(something)[\"toMatchInlineSnapshot\"](`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    44,
						EndColumn: 61,
					},
				},
			},
			// Parenthesized argument: SkipParentheses still reaches the template.
			{
				Code: "expect(something).toMatchInlineSnapshot((`${interpolated}`));",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInterpolation",
						Column:    42,
						EndColumn: 59,
					},
				},
			},

			// Rstest-specific expect sources (AGENTS.md §2.4).
			{
				Code: "import { expect } from \"@rstest/core\";\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Line: 2, Column: 41},
				},
			},
			{
				Code: "import { expect as check } from \"@rstest/core\";\ncheck(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Line: 2, Column: 40},
				},
			},
			{
				Code: "const { expect } = require(\"@rstest/core\");\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Line: 2, Column: 41},
				},
			},
			{
				Code: "import * as rstest from \"@rstest/core\";\nrstest.expect(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Line: 2, Column: 48},
				},
			},
			{
				Code: "import.meta.rstest.expect(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 60},
				},
			},
			{
				Code: "if (import.meta.rstest) {\n  const api = import.meta.rstest;\n  api.expect(something).toMatchInlineSnapshot(`${interpolated}`);\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Line: 3},
				},
			},
			{
				Code: "import { expect } from \"@rstest/playwright\";\nexpect(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Line: 2, Column: 41},
				},
			},

			// The test context is the only expect source that requires
			// RstestCallAnalysis.Callbacks to discover the local root.
			{
				Code: "test(\"case\", ctx => ctx.expect(value).toMatchInlineSnapshot(`${interpolated}`));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 61},
				},
			},
			{
				Code: "test(\"case\", ({ expect }) => expect(value).toMatchInlineSnapshot(`${interpolated}`));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 66},
				},
			},
			{
				Code: "test(\"case\", ({ expect: check }) => check(value).toMatchInlineSnapshot(`${interpolated}`));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 72},
				},
			},

			// Rstest-specific chain shapes.
			{
				Code: "expect.soft(something).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 46},
				},
			},
			{
				// expect.poll omits the snapshot matchers at the type level, but
				// the rule judges interpolation, not API legality.
				Code: "expect.poll(getValue).toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 45},
				},
			},
			{
				// A forgotten expect(x) is valid-expect's business; the
				// interpolation is still reported.
				Code: "expect.toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 30},
				},
			},
			{
				// Chai allows several assertions per chain, so the matcher walk
				// must not stop at the first one. Reverting the rule to
				// parsed.Matcher makes exactly this case pass silently.
				Code: "expect(something).to.be.a(\"string\").and.toMatchInlineSnapshot(`${interpolated}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInterpolation", Column: 63},
				},
			},
		},
	)
}
