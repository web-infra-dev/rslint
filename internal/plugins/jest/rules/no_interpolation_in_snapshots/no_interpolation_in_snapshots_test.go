package no_interpolation_in_snapshots_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_interpolation_in_snapshots"
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
			{Code: `expect.toHaveAssertions();`},
			{Code: "myObjectWants.toMatchInlineSnapshot({}, `${interpolated}`);"},
			{Code: "myObjectWants.toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);"},
			{Code: "toMatchInlineSnapshot({}, `${interpolated}`);"},
			{Code: "toMatchInlineSnapshot({}, `${interpolated1} ${interpolated2}`);"},
			{Code: `expect(something).toThrowErrorMatchingInlineSnapshot();`},
			{Code: "expect(something).toThrowErrorMatchingInlineSnapshot(`No interpolation`);"},
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
					{
						MessageId: "noInterpolation",
						Column:    53,
						EndColumn: 64,
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
		},
	)
}
