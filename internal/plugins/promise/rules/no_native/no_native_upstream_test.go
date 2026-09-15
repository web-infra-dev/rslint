// Tests migrated from eslint-plugin-promise v7.3.0 __tests__/no-native.js
// and docs/rules/no-native.md. Additional coverage lives in no_native_extras_test.go.
package no_native_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_native"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const noNativeMessage = `"Promise" is not defined.`

func noNativeTestRoot() rule_tester.Root {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.allowJs.json"): `{
			"extends": "./tsconfig.json",
			"compilerOptions": { "allowJs": true }
		}`,
	})
	return root
}

func TestNoNativeUpstream(t *testing.T) {
	for _, filename := range []string{"file.js", "file.ts"} {
		t.Run(filename, func(t *testing.T) {
			rule_tester.RunRuleTester(noNativeTestRoot(), "tsconfig.allowJs.json", t, &no_native.NoNativeRule,
				[]rule_tester.ValidTestCase{
					// local variable
					{
						Code:     `var Promise = null; function x() { return Promise.resolve("hi"); }`,
						FileName: filename,
					},
					// polyfill fallback
					{
						Code:     `var Promise = window.Promise || require("bluebird"); var x = Promise.reject();`,
						FileName: filename,
					},
					// import
					{
						Code:     `import Promise from "bluebird"; var x = Promise.reject();`,
						FileName: filename,
					},
					// documentation valid
					{
						Code: `const Promise = require('bluebird')
const x = Promise.resolve('good')`,
						FileName: filename,
					},
				},
				[]rule_tester.InvalidTestCase{
					// constructor
					{
						Code:     `new Promise(function(reject, resolve) { })`,
						FileName: filename,
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 12}},
					},
					// static method
					{
						Code:     `Promise.resolve()`,
						FileName: filename,
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
					},
					// browser environment
					{
						Code:     `new Promise(function(reject, resolve) { })`,
						FileName: filename,
						Globals:  map[string]any{"Promise": false, "window": false},
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 12}},
					},
					// node environment
					{
						Code:     `new Promise(function(reject, resolve) { })`,
						FileName: filename,
						Globals:  map[string]any{"Promise": false, "global": false},
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 12}},
					},
					// es6 environment
					{
						Code:            `Promise.resolve()`,
						FileName:        filename,
						LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
						Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
					},
					// writable global
					{
						Code:     `Promise.resolve()`,
						FileName: filename,
						Globals:  map[string]any{"Promise": true},
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
					},
					// disabled global
					{
						Code:     `Promise.resolve()`,
						FileName: filename,
						Globals:  map[string]any{"Promise": "off"},
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
					},
					// documentation invalid
					{
						Code:     `const x = Promise.resolve('bad')`,
						FileName: filename,
						Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "name", Message: noNativeMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 18}},
					},
				},
			)
		})
	}
}
