package no_named_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_export"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-named-export.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-export.md
func TestNoNamedExportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_named_export.NoNamedExportRule,
		[]rule_tester.ValidTestCase{
			// Upstream tests.
			{
				Code:            "module.export.foo = function () {}",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code: "module.export.foo = function () {}",
			},
			{
				Code: "export default function bar() {};",
			},
			{
				Code: "let foo; export { foo as default }",
			},
			// SKIP: Babel default re-export proposals are not supported by the parser.
			{
				Code: "export default from \"foo.js\"",
				Skip: true,
			},
			{
				Code: "import * as foo from './foo';",
			},
			{
				Code: "import foo from './foo';",
			},
			{
				Code: "import {default as foo} from './foo';",
			},
			{
				Code: "let foo; export { foo as \"default\" }",
			},
			// Upstream documentation.
			{
				Code: "// good1.js\n\n// There is only a single module export and it's a default export.\nexport default 'bar';",
			},
			{
				Code: "// good2.js\n\n// There is only a single module export and it's a default export.\nconst foo = 'foo';\nexport { foo as default }",
			},
			// SKIP: Babel default re-export proposals are not supported by the parser.
			{
				Code: "// good3.js\n\n// There is only a single module export and it's a default export.\nexport default from './other-module';",
				Skip: true,
			},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream tests.
			{
				Code:   "\n        export const foo = 'foo';\n        export const bar = 'bar';\n      ",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(2, 9, 2, 34), noNamedExportError(3, 9, 3, 34)},
			},
			{
				Code:   "\n        export const foo = 'foo';\n        export default bar;",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(2, 9, 2, 34)},
			},
			{
				Code:   "\n        export const foo = 'foo';\n        export function bar() {};\n      ",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(2, 9, 2, 34), noNamedExportError(3, 9, 3, 33)},
			},
			{
				Code:   "export const foo = 'foo';",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 26)},
			},
			{
				Code:   "\n        const foo = 'foo';\n        export { foo };\n      ",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(3, 9, 3, 24)},
			},
			{
				Code:   "let foo, bar; export { foo, bar }",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 15, 1, 34)},
			},
			{
				Code:   "export const { foo, bar } = item;",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 34)},
			},
			{
				Code:   "export const { foo, bar: baz } = item;",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 39)},
			},
			{
				Code:   "export const { foo: { bar, baz } } = item;",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 43)},
			},
			{
				Code:   "\n        let item;\n        export const foo = item;\n        export { item };\n      ",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(3, 9, 3, 33), noNamedExportError(4, 9, 4, 25)},
			},
			{
				Code:   "export * from './foo';",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 23)},
			},
			{
				Code:   "export const { foo } = { foo: \"bar\" };",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 39)},
			},
			{
				Code:   "export const { foo: { bar } } = { foo: { bar: \"baz\" } };",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 57)},
			},
			{
				Code:   "export { a, b } from \"foo.js\"",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 30)},
			},
			{
				Code:   "export type UserId = number;",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 29)},
			},
			// SKIP: Babel default re-export proposals are not supported by the parser.
			{
				Code:   "export foo from \"foo.js\"",
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 0, 0)},
			},
			// SKIP: Babel default re-export proposals are not supported by the parser.
			{
				Code:   "export Memory, { MemoryValue } from './Memory'",
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 0, 0)},
			},
			// Upstream documentation.
			{
				Code:   "// bad1.js\n\n// There is only a single module export and it's a named export.\nexport const foo = 'foo';",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(4, 1, 4, 26)},
			},
			{
				Code:   "// bad2.js\n\n// There is more than one named export in the module.\nexport const foo = 'foo';\nexport const bar = 'bar';",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(4, 1, 4, 26), noNamedExportError(5, 1, 5, 26)},
			},
			{
				Code:   "// bad3.js\n\n// There is more than one named export in the module.\nconst foo = 'foo';\nconst bar = 'bar';\nexport { foo, bar }",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(6, 1, 6, 20)},
			},
			{
				Code:   "// bad4.js\n\n// There is more than one named export in the module.\nexport * from './other-module'",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(4, 1, 4, 31)},
			},
			{
				Code:   "// bad5.js\n\n// There is a default and a named export.\nexport const foo = 'foo';\nconst bar = 'bar';\nexport default 'bar';",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(4, 1, 4, 26)},
			},
		},
	)
}

func noNamedExportError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "",
		Message:   "Named exports are not allowed.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
