package require_dynamic_import_entry

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const reviewDynamicImportMessage = "TypeScript does not collect this import() dependency because its path is not a string literal or a template literal without substitutions. Review every module Rspack may load, including module side effects. Add source files whose runtime code may be removed to Rslim entries. Once all targets are safe, suppress this call with rslint-disable-next-line rslim/require-dynamic-import-entry and a reason."

const reviewRequireMessage = "TypeScript does not collect this require() dependency for Rslim. Review every module Rspack may load, including module side effects. Add source files whose runtime code may be removed to Rslim entries. Once all targets are safe, suppress this call with rslint-disable-next-line rslim/require-dynamic-import-entry and a reason."

func TestRequireDynamicImportEntry(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &RequireDynamicImportEntryRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { value } from './module';`},
			{Code: `import('./module');`},
			{Code: "import(`./module`);"},
			{Code: `await import('./module', { with: { type: 'json' } });`},
			{Code: `import('./module', options);`},
			{Code: `import('./module');`, FileName: "input.js"},
			{Code: `type Module = typeof import('./module');`},
			{Code: `import.meta.resolve('./module');`},
			{Code: `require.resolve('./module');`},
			{Code: `loader.require('./module');`},
			{Code: `require('./module');`},
			{Code: "require(`./module`);"},
			{Code: `require('./module');`, FileName: "input.js"},
			{Code: "require(`./module`);", FileName: "input.js"},
			{Code: `require('./module');`, FileName: "input.tsx"},
			{Code: "require(`./module`);", FileName: "input.tsx"},
			{Code: `require('./module');`, FileName: "input.jsx"},
			{Code: "require(`./module`);", FileName: "input.jsx"},
			{Code: `import dep = require('./module');`},
			{Code: `import type dep = require('./module');`},
			{Code: "// rslint-disable-next-line test -- Checked all target modules and Rslim entries.\nconst mod = await import(\n  path\n);"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import(path);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "reviewDynamicImport", Message: reviewDynamicImportMessage,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 13,
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
				Code: `await import(path, { with: { type: 'json' } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "reviewDynamicImport", Line: 1, Column: 7, EndLine: 1, EndColumn: 47,
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
			{Code: `require(path);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `require();`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: "require(`./${name}/module`);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `const path = './module'; require(path);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `require('./' + 'module');`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `require('./module', options);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `require(path);`, FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire", Message: reviewRequireMessage}}},
			{Code: `require();`, FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `require('./module', options);`, FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewRequire"}}},
			{Code: `const path = './module'; import(path);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewDynamicImport"}}},
			{Code: `import('./' + 'module');`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewDynamicImport"}}},
			{Code: `import(path);`, FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "reviewDynamicImport"}}},
		},
	)
}
