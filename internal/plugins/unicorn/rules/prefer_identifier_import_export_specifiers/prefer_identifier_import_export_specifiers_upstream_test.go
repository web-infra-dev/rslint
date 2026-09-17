// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-identifier-import-export-specifiers.js
package prefer_identifier_import_export_specifiers_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_identifier_import_export_specifiers"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferIdentifierImportExportSpecifiersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_identifier_import_export_specifiers.PreferIdentifierImportExportSpecifiersRule, []rule_tester.ValidTestCase{
		{Code: "import \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import foo from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import * as foo from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {foo as bar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {\"a string\" as aString} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {\"foo-bar\" as fooBar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {\"\" as empty} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = 1;\nexport {foo};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = 1;\nexport {foo as bar};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export {foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export {foo as bar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = 1;\nexport {foo as \"a string\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = 1;\nexport {foo as \"foo-bar\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export {\"a string\" as aString} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export {\"foo-bar\" as fooBar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export {\"\" as empty} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export * from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export * as foo from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export * as \"a string\" from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import foo from \"foo\" with {type: \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import foo from \"foo\" with {\"foo-bar\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import foo from \"foo\" with {\"\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import foo from \"foo\" with {\"0\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export {foo} from \"foo\" with {\"a string\": \"x\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nimport {foo as foo} from 'foo';\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = 1;\n// ✅\nexport {foo as bar};\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nexport {foo as bar} from 'foo';\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nimport foo from 'foo' with {type: 'json'};\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nimport {'a string' as aString} from 'foo';\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "import {\"foo\" as foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {foo as foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"default\" as defaultExport} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {default as defaultExport} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `default` over string literal `\"default\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"foo\" as foo, \"bar\" as bar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {foo as foo, bar as bar} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"foo\" as foo} from \"foo\" with {type: \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {foo as foo} from \"foo\" with {type: \"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"foo\"as foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {foo as foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"\\u0066oo\" as foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {foo as foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"\\u0066oo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1;\nexport {foo as \"bar\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = 1;\nexport {foo as bar};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 2, Column: 16, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1;\nexport {foo as \"default\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = 1;\nexport {foo as default};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `default` over string literal `\"default\"`.", Line: 2, Column: 16, EndLine: 2, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1, baz = 2;\nexport {foo as \"bar\", baz as \"qux\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = 1, baz = 2;\nexport {foo as bar, baz as qux};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 2, Column: 16, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `qux` over string literal `\"qux\"`.", Line: 2, Column: 30, EndLine: 2, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1;\nexport {foo as\"bar\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = 1;\nexport {foo as bar};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 2, Column: 15, EndLine: 2, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1;\nexport {foo as \"\\u0062ar\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = 1;\nexport {foo as bar};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"\\u0062ar\"`.", Line: 2, Column: 16, EndLine: 2, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\" as bar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo as bar} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"default\" as defaultExport} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {default as defaultExport} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `default` over string literal `\"default\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\" as bar, \"baz\" as qux} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo as bar, baz as qux} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `baz` over string literal `\"baz\"`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\"} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\" as \"bar\"} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo as bar} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\"as bar} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo as bar} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\"as\"bar\"} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo as bar} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import type {\"foo\" as Foo} from \"foo\";", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import type {foo as Foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export type {\"foo\" as Foo} from \"foo\";", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export type {foo as Foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export * as \"foo\" from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export * as foo from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export * as\"foo\" from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export * as foo from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import foo from \"foo\" with {\"type\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import foo from \"foo\" with {type: \"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {foo} from \"foo\" with {\"type\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo} from \"foo\" with {type: \"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import foo from \"foo\" with{\"type\":\"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import foo from \"foo\" with{type:\"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import foo from \"foo\" with {\"type\": \"json\", \"other\": \"x\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import foo from \"foo\" with {type: \"json\", other: \"x\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `other` over string literal `\"other\"`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import foo from \"foo\" with {\"type\": \"json\"};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import foo from \"foo\" with {type: \"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"if\" as foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {if as foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `if` over string literal `\"if\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"yield\" as foo} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {yield as foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `yield` over string literal `\"yield\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import foo from \"foo\" with {\"default\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import foo from \"foo\" with {default: \"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `default` over string literal `\"default\"`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"foo\" as \"foo\"} from \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {foo as foo} from \"foo\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nimport {'foo' as foo} from 'foo';\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nimport {foo as foo} from 'foo';\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `'foo'`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1;\n// ❌\nexport {foo as 'bar'};\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = 1;\n// ❌\nexport {foo as bar};\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `'bar'`.", Line: 3, Column: 16, EndLine: 3, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nexport {'foo' as bar} from 'foo';\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nexport {foo as bar} from 'foo';\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `'foo'`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nimport foo from 'foo' with {'type': 'json'};\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nimport foo from 'foo' with {type: 'json'};\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `'type'`.", Line: 2, Column: 29, EndLine: 2, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
