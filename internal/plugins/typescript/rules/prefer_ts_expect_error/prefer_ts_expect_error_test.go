package prefer_ts_expect_error

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTsExpectErrorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferTsExpectErrorRule, []rule_tester.ValidTestCase{
		{Code: `// @ts-nocheck`},
		{Code: `// @ts-check`},
		{Code: `// just a comment containing @ts-ignore somewhere`},
		{Code: `// @ts-ignorefoo`},
		{Code: `/* @ts-ignorefoo */`},
		{
			Code: `
{
  /*
        just a comment containing @ts-ignore somewhere in a block
      */
}
			`,
		},
		{Code: `// @ts-expect-error`},
		{
			Code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
			`,
		},
		{
			Code: `
/**
 * Explaining comment
 *
 * @ts-expect-error
 *
 * Not last line
 * */
			`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `// @ts-ignore`,
			Output: []string{
				`// @ts-expect-error`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
			},
		},
		{
			Code: `// @ts-ignore: Suppress next line`,
			Output: []string{
				`// @ts-expect-error: Suppress next line`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code: `///@ts-ignore: Suppress next line`,
			Output: []string{
				`///@ts-expect-error: Suppress next line`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
			`,
			Output: []string{
				`
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
			`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 3, Column: 3, EndLine: 3, EndColumn: 40},
			},
		},
		{
			Code: `/* @ts-ignore */`,
			Output: []string{
				`/* @ts-expect-error */`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
			},
		},
		{
			Code: `
/**
 * Explaining comment
 *
 * @ts-ignore */
			`,
			Output: []string{
				`
/**
 * Explaining comment
 *
 * @ts-expect-error */
			`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 2, Column: 1, EndLine: 5, EndColumn: 17},
			},
		},
		{
			Code: `/* @ts-ignore in a single block */`,
			Output: []string{
				`/* @ts-expect-error in a single block */`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
			},
		},
		{
			Code: `
/*
// @ts-ignore in a block with single line comments */
			`,
			Output: []string{
				`
/*
// @ts-expect-error in a block with single line comments */
			`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferExpectErrorComment", Line: 2, Column: 1, EndLine: 3, EndColumn: 54},
			},
		},
	})
}
