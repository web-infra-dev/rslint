package exports_last_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/exports_last"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func exportsLastError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "", Message: "Export statements should appear at the end of the file",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/exports-last.js
func TestExportsLastUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &exports_last.ExportsLastRule,
		[]rule_tester.ValidTestCase{
			// Empty file.
			{
				Code: "// comment",
			},
			// No exports.
			{
				Code: "\n        const foo = 'bar'\n        const bar = 'baz'\n      ",
			},
			// Named export.
			{
				Code: "\n        const foo = 'bar'\n        export {foo}\n      ",
			},
			// Default export.
			{
				Code: "\n        const foo = 'bar'\n        export default foo\n      ",
			},
			// Only exports.
			{
				Code: "\n        export default foo\n        export const bar = true\n      ",
			},
			// Statements before exports.
			{
				Code: "\n        const foo = 'bar'\n        export default foo\n        export const bar = true\n      ",
			},
			// Multiline export.
			{
				Code: "\n        const foo = 'bar'\n        export default function bar () {\n          const very = 'multiline'\n        }\n        export const baz = true\n      ",
			},
			// Many exports.
			{
				Code: "\n        const foo = 'bar'\n        export default foo\n        export const so = 'many'\n        export const exports = ':)'\n        export const i = 'cant'\n        export const even = 'count'\n        export const how = 'many'\n      ",
			},
			// Export all.
			{
				Code: "\n        export * from './foo'\n      ",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Default export before a variable declaration.
			{
				Code:   "\n        export default 'bar'\n        const bar = true\n      ",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 9, 2, 29)},
			},
			// Named export before a variable declaration.
			{
				Code:   "\n        export const foo = 'bar'\n        const bar = true\n      ",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 9, 2, 33)},
			},
			// Export all before a variable declaration.
			{
				Code:   "\n        export * from './foo'\n        const bar = true\n      ",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 9, 2, 30)},
			},
			// Many exports around a variable declaration.
			{
				Code:   "\n        export default 'such foo many bar'\n        export const so = 'many'\n        const foo = 'bar'\n        export const exports = ':)'\n        export const i = 'cant'\n        export const even = 'count'\n        export const how = 'many'\n      ",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 9, 2, 43), exportsLastError(3, 9, 3, 33)},
			},
		},
	)
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/exports-last.md
func TestExportsLastDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &exports_last.ExportsLastRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "const arr = ['bar']\n\nexport const bool = true\n\nexport default bool\n\nexport function func() {\n  console.log('Hello World 🌍')\n}\n\nexport const str = 'foo'\n",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "\nconst bool = true\n\nexport default bool\n\nconst str = 'foo'\n\n",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(4, 1, 4, 20)},
			},
			{
				Code:   "\nexport const bool = true\n\nconst str = 'foo'\n\n",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 1, 2, 25)},
			},
		},
	)
}
