package no_import_module_exports_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_import_module_exports"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportModuleExportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &no_import_module_exports.NoImportModuleExportsRule,
		[]rule_tester.ValidTestCase{
			{Code: `module.exports = 1;`},
			{Code: `import value from "value"; function f(module) { module.exports = value; }`},
			{Code: `import value from "value"; { const exports = {}; exports.value = value; }`},
			{Code: `import { exports } from "value"; exports.value = 1;`},
			{Code: `const module = {}; import value from "value"; module.value;`},
			{Code: `const exports = {}; import value from "value"; exports.value = value;`},
			{Code: `import value from "value"; (module as any).exports = value;`, FileName: "source.ts"},
			{Code: `import value from "value"; const element = <module.exports />;`, FileName: "source.tsx", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `module.exports = 1; import value from "value";`,
				Errors: importError(1, 21, 1, 47),
			},
			{
				Code: "import a from \"a\";\nmodule.exports = a;\nimport b from \"b\";",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: errorMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{MessageId: "", Message: errorMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 19},
				},
			},
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
			{
				Code:   "import value from 'value';\nmodule.exports = value;\nfunction unrelated(module) {}",
				Errors: importError(1, 1, 1, 27),
			},
			{Code: `import type { Value } from "value"; module.exports = 1;`, FileName: "source.ts", Errors: importError(1, 1, 1, 36)},
		},
	)
}

func TestNoImportModuleExportsSchema(t *testing.T) {
	valid := []any{map[string]any{"exceptions": []any{"**/*.js"}}}
	if err := no_import_module_exports.NoImportModuleExportsRule.Schema.Validate(valid); err != nil {
		t.Fatalf("rejected valid options: %v", err)
	}

	invalid := []any{map[string]any{"exceptions": []any{42}}}
	if err := no_import_module_exports.NoImportModuleExportsRule.Schema.Validate(invalid); err == nil {
		t.Fatalf("accepted non-string exception: %#v", invalid)
	}
}
