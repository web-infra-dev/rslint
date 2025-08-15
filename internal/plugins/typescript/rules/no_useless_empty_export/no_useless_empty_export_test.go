package no_useless_empty_export

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessEmptyExportRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUselessEmptyExportRule, []rule_tester.ValidTestCase{
		{Code: "declare module '_'"},
		{Code: "import {} from '_';"},
		{Code: "import * as _ from '_';"},
		{Code: "export = {};"},
		{Code: "export = 3;"},
		{Code: "export const _ = {};"},
		{Code: `
const _ = {};
export default _;
`},
		{Code: `
export * from '_';
export = {};
`},
		{Code: `
export {};
`},
		// https://github.com/microsoft/TypeScript/issues/38592
		{
			Code: `
export type A = 1;
export {};
`,
		},
		{
			Code: `
export declare const a = 2;
export {};
`,
		},
		{
			Code: `
import type { A } from '_';
export {};
`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
export const _ = {};
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      3,
					Column:    1,
				},
			},
			Output: []string{`
export const _ = {};

`},
		},
		{
			Code: `
export * from '_';
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      3,
					Column:    1,
				},
			},
			Output: []string{`
export * from '_';

`},
		},
		{
			Code: `
export {};
export * from '_';
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      2,
					Column:    1,
				},
			},
			Output: []string{`

export * from '_';
`},
		},
		{
			Code: `
const _ = {};
export default _;
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      4,
					Column:    1,
				},
			},
			Output: []string{`
const _ = {};
export default _;

`},
		},
		{
			Code: `
export {};
const _ = {};
export default _;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      2,
					Column:    1,
				},
			},
			Output: []string{`

const _ = {};
export default _;
`},
		},
		{
			Code: `
const _ = {};
export { _ };
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      4,
					Column:    1,
				},
			},
			Output: []string{`
const _ = {};
export { _ };

`},
		},
		{
			Code: `
import _ = require('_');
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      3,
					Column:    1,
				},
			},
			Output: []string{`
import _ = require('_');

`},
		},
		{
			Code: `
import _ = require('_');
export {};
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      3,
					Column:    1,
				},
				{
					MessageId: "uselessExport",
					Line:      4,
					Column:    1,
				},
			},
			Output: []string{`
import _ = require('_');


`},
		},
		{
			Code: `
import { A } from '_';
export {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessExport",
					Line:      3,
					Column:    1,
				},
			},
			Output: []string{`
import { A } from '_';

`},
		},
	})
}
