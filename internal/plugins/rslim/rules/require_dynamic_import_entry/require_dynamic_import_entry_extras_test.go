package require_dynamic_import_entry

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const reviewDynamicImportMessage = "Review this dynamic import for Rslim: identify every module Rspack may load, including module side effects. Add source files whose runtime code may be removed to Rslim entries. Once all targets are safe, suppress this call with rslint-disable-next-line rslim/require-dynamic-import-entry and a reason."

func TestRequireDynamicImportEntry(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireDynamicImportEntryRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { value } from './module';`},
			{Code: `type Module = typeof import('./module');`},
			{Code: `import.meta.resolve('./module');`},
			{Code: "// rslint-disable-next-line test -- Checked all target modules and Rslim entries.\nimport('./module');"},
			{Code: "// rslint-disable-next-line test -- Checked all target modules and Rslim entries.\nconst mod = await import(\n  path\n);"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import('./module');`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "reviewDynamicImport", Message: reviewDynamicImportMessage,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 19,
				}},
			},
			{
				Code: `import(path); import(` + "`./${name}/module`" + `);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reviewDynamicImport", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
					{MessageId: "reviewDynamicImport", Line: 1, Column: 15, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `await import('./module', { with: { type: 'json' } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "reviewDynamicImport", Line: 1, Column: 7, EndLine: 1, EndColumn: 53,
				}},
			},
			{
				Code: "const mod = await import(\n  path\n);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "reviewDynamicImport", Line: 1, Column: 19, EndLine: 3, EndColumn: 2,
				}},
			},
			{
				Code: `import();`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "reviewDynamicImport", Line: 1, Column: 1, EndLine: 1, EndColumn: 9,
				}},
			},
		},
	)
}
