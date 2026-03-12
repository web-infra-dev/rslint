package jsx_filename_extension

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxFilenameExtensionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFilenameExtensionRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = 1`,
			Tsx:  true,
		},
		{
			Code:    `const App = () => <div />`,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".tsx"}},
		},
		{
			// ignoreFilesWithoutCode: empty file with .tsx extension should not error
			Code:    ``,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".tsx"}, "allow": "as-needed", "ignoreFilesWithoutCode": true},
		},
		{
			// Default: .tsx is not in default extensions [.jsx], but no JSX in code = no error
			Code: `var x = 1`,
			Tsx:  true,
		},
		{
			// JSX in file with .tsx extension allowed
			Code:    `const App = () => <div><span /></div>`,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".tsx", ".jsx"}},
		},
		{
			// as-needed: no JSX in .tsx, but ignoreFilesWithoutCode with no statements
			Code:    ``,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".tsx"}, "allow": "as-needed", "ignoreFilesWithoutCode": true},
		},
		{
			// Fragment in allowed extension
			Code:    `const App = () => <></>`,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".tsx"}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `const x = 1;`,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".tsx"}, "allow": "as-needed"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "extensionOnlyForJSX"}},
		},
		{
			// JSX in .tsx but only .js allowed
			Code:    `const App = () => <div />`,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".js"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noJSXWithExtension"}},
		},
		{
			// Fragment in .tsx but only .jsx allowed
			Code:    `const App = () => <></>`,
			Tsx:     true,
			Options: map[string]interface{}{"extensions": []interface{}{".jsx"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noJSXWithExtension"}},
		},
	})
}
