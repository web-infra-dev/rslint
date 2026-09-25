package prefer_default_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/prefer_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/prefer-default-export.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/prefer-default-export.md
// Parser variants share cases; TypeScript parser comments use one representative name.
func TestPreferDefaultExportSingle(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_default_export.PreferDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "\n        export const foo = 'foo';\n        export const bar = 'bar';",
			},
			{
				Code: "\n        export default function bar() {};",
			},
			{
				Code: "\n        export const foo = 'foo';\n        export function bar() {};",
			},
			{
				Code: "\n        export const foo = 'foo';\n        export default bar;",
			},
			{
				Code: "\n        let foo, bar;\n        export { foo, bar }",
			},
			{
				Code: "\n        export const { foo, bar } = item;",
			},
			{
				Code: "\n        export const { foo, bar: baz } = item;",
			},
			{
				Code: "\n        export const { foo: { bar, baz } } = item;",
			},
			{
				Code: "\n        export const [a, b] = item;",
			},
			{
				Code: "\n        let item;\n        export const foo = item;\n        export { item };",
			},
			{
				Code: "\n        let foo;\n        export { foo as default }",
			},
			{
				Code: "\n        export * from './foo';",
			},
			// SKIP: Babel default re-export proposals are unsupported by the parser.
			{
				Code: "export Memory, { MemoryValue } from './Memory'",
				Skip: true,
			},
			{
				Code: "\n        import * as foo from './foo';",
			},
			{
				Code: "export type UserId = number;",
			},
			// SKIP: Babel default re-export proposals are unsupported by the parser.
			{
				Code: "export default from \"foo.js\"",
				Skip: true,
			},
			{
				Code: "export { a, b } from \"foo.js\"",
			},
			{
				Code: "\n        export const [CounterProvider,, withCounter] = func();;\n      ",
			},
			{
				Code: "let foo; export { foo as \"default\" };",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "\n        export function bar() {};",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 9, 2, 33)},
			},
			{
				Code:   "\n        export const foo = 'foo';",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 9, 2, 34)},
			},
			{
				Code:   "\n        const foo = 'foo';\n        export { foo };",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 3, 18, 3, 21)},
			},
			{
				Code:   "\n        export const { foo } = { foo: \"bar\" };",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 9, 2, 47)},
			},
			{
				Code:   "\n        export const { foo: { bar } } = { foo: { bar: \"baz\" } };",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 9, 2, 65)},
			},
			{
				Code:   "\n        export const [a] = [\"foo\"]",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 9, 2, 35)},
			},
		},
	)
}

func TestPreferDefaultExportAny(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_default_export.PreferDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "\n          export default function bar() {};",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "\n              export const foo = 'foo';\n              export const bar = 'bar';\n              export default 42;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "\n            export default a = 2;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "\n            export const a = 2;\n            export default function foo() {};",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "\n          export const a = 5;\n          export function bar(){};\n          let foo;\n          export { foo as default }",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "\n          export * from './foo';",
				Options: map[string]any{"target": "any"},
			},
			// SKIP: Babel default re-export proposals are unsupported by the parser.
			{
				Code:    "export Memory, { MemoryValue } from './Memory'",
				Skip:    true,
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "\n            import * as foo from './foo';",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "const a = 5;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export const a = 4; let foo; export { foo as \"default\" };",
				Options: map[string]any{"target": "any"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "\n        export const foo = 'foo';\n        export const bar = 'bar';",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 3, 9, 3, 34)},
			},
			{
				Code:    "\n        export const foo = 'foo';\n        export function bar() {};",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 3, 9, 3, 33)},
			},
			{
				Code:    "\n        let foo, bar;\n        export { foo, bar }",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 3, 23, 3, 26)},
			},
			{
				Code:    "\n        let item;\n        export const foo = item;\n        export { item };",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 4, 18, 4, 22)},
			},
			{
				Code:    "export { a, b } from \"foo.js\"",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 1, 13, 1, 14)},
			},
			{
				Code:    "\n        const foo = 'foo';\n        export { foo };",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 3, 18, 3, 21)},
			},
			{
				Code:    "\n        export const { foo } = { foo: \"bar\" };",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 2, 9, 2, 47)},
			},
			{
				Code:    "\n        export const { foo: { bar } } = { foo: { bar: \"baz\" } };",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 2, 9, 2, 65)},
			},
		},
	)
}

func TestPreferDefaultExportTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_default_export.PreferDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "\n            export type foo = string;\n            export type bar = number;\n            /* @typescript-eslint/parser */\n          ",
			},
			{
				Code: "\n            export type foo = string;\n            export type bar = number;\n            /* @typescript-eslint/parser */\n          ",
			},
			{
				Code: "export type foo = string /* @typescript-eslint/parser*/",
			},
			{
				Code: "export interface foo { bar: string; } /* @typescript-eslint/parser*/",
			},
			{
				Code: "export interface foo { bar: string; }; export function goo() {} /* @typescript-eslint/parser*/",
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

func TestPreferDefaultExportDocsSingle(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_default_export.PreferDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "// good1.js\n\n// There is a default export.\nexport const foo = 'foo';\nconst bar = 'bar';\nexport default bar;",
			},
			{
				Code: "// good2.js\n\n// There is more than one named export in the module.\nexport const foo = 'foo';\nexport const bar = 'bar';",
			},
			{
				Code: "// good3.js\n\n// There is more than one named export in the module\nconst foo = 'foo';\nconst bar = 'bar';\nexport { foo, bar }",
			},
			{
				Code: "// good4.js\n\n// There is a default export.\nconst foo = 'foo';\nexport { foo as default }",
			},
			{
				Code: "// export-star.js\n\n// Any batch export will disable this rule. The remote module is not inspected.\nexport * from './other-module'",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "// bad.js\n\n// There is only a single module export and it's a named export.\nexport const foo = 'foo';\n",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 4, 1, 4, 26)},
			},
		},
	)
}

func TestPreferDefaultExportDocsAny(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_default_export.PreferDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "// good1.js\n\n//has default export\nexport default function bar() {};",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "// good2.js\n\n// has default export\nlet foo;\nexport { foo as default }",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "// good3.js\n\n//contains multiple exports AND default export\nexport const a = 5;\nexport function bar(){};\nlet foo;\nexport { foo as default }",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "// good4.js\n\n// does not contain any exports => file is not checked by the rule\nimport * as foo from './foo';\ufeff",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "// export-star.js\n\n// Any batch export will disable this rule. The remote module is not inspected.\nexport * from './other-module'",
				Options: map[string]any{"target": "any"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "// bad1.js\n\n//has 2 named exports, but no default export\nexport const foo = 'foo';\nexport const bar = 'bar';",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 5, 1, 5, 26)},
			},
			{
				Code:    "// bad2.js\n\n// does not have default export\nlet foo, bar;\nexport { foo, bar }",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 5, 15, 5, 18)},
			},
			{
				Code:    "// bad3.js\n\n// does not have default export\nexport { a, b } from \"foo.js\"\ufeff",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 4, 13, 4, 14)},
			},
			{
				Code:    "// bad4.js\n\n// does not have default export\nlet item;\nexport const foo = item;\nexport { item };",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 6, 10, 6, 14)},
			},
		},
	)
}

const singleExportMessage = "Prefer default export on a file with single export."
const anyExportMessage = "Prefer default export to be present on every file that has export."

func preferDefaultExportError(message string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "",
		Message:   message,
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
