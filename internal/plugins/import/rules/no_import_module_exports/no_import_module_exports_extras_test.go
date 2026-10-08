package no_import_module_exports_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_import_module_exports"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportModuleExportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &no_import_module_exports.NoImportModuleExportsRule,
		[]rule_tester.ValidTestCase{
			{Code: `import value from "value"; function f(module) { module.exports = value; }`},
			{Code: `import value from "value"; { const exports = {}; exports.value = value; }`},
			{Code: `import { exports } from "value"; exports.value = 1;`},
			{Code: `import value from "value"; (module as any).exports = value;`, FileName: "source.ts"},
			{Code: `import value from "value"; const element = <module.exports />;`, FileName: "source.tsx", Tsx: true},
			// Upstream marks the file as reported even when the first CommonJS
			// member occurs before any import declaration.
			{Code: `module.exports = 1; import value from "value";`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "import a from \"a\";\nimport b from \"b\";\nmodule.value;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: errorMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{MessageId: "", Message: errorMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 19},
				},
			},
			{Code: `import value from "value"; exports["value"];`, Errors: importError(1, 1, 1, 27)},
			{Code: `import value from "value"; module?.exports;`, Errors: importError(1, 1, 1, 27)},
			{
				Code:     `import value from "value"; module.exports = value;`,
				FileName: "missing-entrypoint/cli.js",
				Errors:   importError(1, 1, 1, 27),
			},
			// A top-level declaration lives in eslint-scope's module scope and
			// therefore does not count as a shadowing non-module scope.
			{Code: `const module = {}; import value from "value"; module.value;`, Errors: importError(1, 20, 1, 46)},
			{Code: `import type { Value } from "value"; module.exports = 1;`, FileName: "source.ts", Errors: importError(1, 1, 1, 36)},
		},
	)
}
