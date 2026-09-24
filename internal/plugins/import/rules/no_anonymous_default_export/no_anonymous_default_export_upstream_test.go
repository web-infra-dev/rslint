package no_anonymous_default_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_anonymous_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	arrayMessage    = "Assign array to a variable before exporting as module default"
	arrowMessage    = "Assign arrow function to a variable before exporting as module default"
	callMessage     = "Assign call result to a variable before exporting as module default"
	classMessage    = "Unexpected default export of anonymous class"
	functionMessage = "Unexpected default export of anonymous function"
	literalMessage  = "Assign literal to a variable before exporting as module default"
	objectMessage   = "Assign object to a variable before exporting as module default"
	newMessage      = "Assign instance to a variable before exporting as module default"
)

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-anonymous-default-export.js
// Includes examples from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-anonymous-default-export.md
func TestNoAnonymousDefaultExportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &no_anonymous_default_export.NoAnonymousDefaultExportRule,
		[]rule_tester.ValidTestCase{
			// Exports with identifiers. The docs typo `class MyClass() {}` is covered below as `class MyClass {}`.
			{
				Code: "const foo = 123\nexport default foo",
			},
			{
				Code: "export default function foo() {}",
			},
			{
				Code: "export default class MyClass {}",
			},
			// Allow each forbidden type with its option.
			{
				Code:    "export default []",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "export default () => {}",
				Options: map[string]any{"allowArrowFunction": true},
			},
			{
				Code:    "export default class {}",
				Options: map[string]any{"allowAnonymousClass": true},
			},
			{
				Code:    "export default function() {}",
				Options: map[string]any{"allowAnonymousFunction": true},
			},
			{
				Code:    "export default 123",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default 'foo'",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default `foo`",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default {}",
				Options: map[string]any{"allowObject": true},
			},
			{
				Code:    "export default foo(bar)",
				Options: map[string]any{"allowCallExpression": true},
			},
			{
				Code:    "export default new Foo()",
				Options: map[string]any{"allowNew": true},
			},
			// Multiple options.
			{
				Code:    "export default 123",
				Options: map[string]any{"allowLiteral": true, "allowObject": true},
			},
			{
				Code:    "export default {}",
				Options: map[string]any{"allowLiteral": true, "allowObject": true},
			},
			// Unrelated export syntaxes, including the ESLint >= 8.7 arbitrary export name case.
			{
				Code: "export * from 'foo'",
			},
			{
				Code: "const foo = 123\nexport { foo }",
			},
			{
				Code: "const foo = 123\nexport { foo as default }",
			},
			{
				Code: "const foo = 123\nexport { foo as \"default\" }",
			},
			// Call expressions are allowed by default for backwards compatibility.
			{
				Code: "export default foo(bar)",
			},
			// All 18 SYNTAX_CASES from v2.32.0/tests/src/utils.js; object rest uses the native parser.
			{
				Code: "for (let { foo, bar } of baz) {}",
			},
			{
				Code: "for (let [ foo, bar ] of baz) {}",
			},
			{
				Code: "const { x, y } = bar",
			},
			{
				Code: "const { x, y, ...z } = bar",
			},
			{
				Code: "let x; export { x }",
			},
			{
				Code: "let x; export { x as y }",
			},
			{
				Code: "export const x = null",
			},
			{
				Code: "export var x = null",
			},
			{
				Code: "export let x = null",
			},
			{
				Code: "export default x",
			},
			{
				Code: "export default class x {}",
			},
			{
				Code:     "import json from \"./data.json\"",
				Settings: map[string]any{"import/extensions": []any{".js"}},
			},
			{
				Code:     "import foo from \"./foobar.json\";",
				Settings: map[string]any{"import/extensions": []any{".js"}},
			},
			{
				Code:     "import foo from \"./foobar\";",
				Settings: map[string]any{"import/extensions": []any{".js"}},
			},
			{
				Code:     "import { foo } from \"./issue-370-commonjs-namespace/bar\"",
				Settings: map[string]any{"import/ignore": []any{"foo"}},
			},
			{
				Code:     "export * from \"./issue-370-commonjs-namespace/bar\"",
				Settings: map[string]any{"import/ignore": []any{"foo"}},
			},
			{
				Code: "import * as a from \"./commonjs-namespace/a\"; a.b",
			},
			{
				Code: "import { foo } from \"./ignore.invalid.extension\"",
			},
			// Documentation example with a space before the function parameters. Other docs examples match cases above.
			{
				Code:    "export default function () {}",
				Options: map[string]any{"allowAnonymousFunction": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Every upstream invalid case. Upstream uses message text without message IDs.
			{
				Code: "export default []",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrayMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "export default () => {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrowMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "export default class {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "export default function() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: functionMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "export default 123",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "export default 'foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "export default `foo`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "export default {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: objectMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export default foo(bar)",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "export default new Foo()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: newMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// An unrelated option does not suppress the diagnostic.
			{
				Code:    "export default 123",
				Options: map[string]any{"allowObject": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// Documentation example; the other failing examples are already covered above.
			{
				Code: "export default function () {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: functionMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
		},
	)
}
