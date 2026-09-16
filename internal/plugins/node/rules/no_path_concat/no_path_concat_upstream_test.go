package no_path_concat_test

import (
	"os"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_path_concat"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func concatErrorAt(id string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Use path.join() or path.resolve() instead of string concatenation."
	if id == "useUrl" {
		message = "Use new URL() instead of string concatenation."
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: id, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runNoPathConcatTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	globals := map[string]any{"__dirname": "readonly", "__filename": "readonly", "require": "readonly", "global": "readonly"}
	for i := range valid {
		if valid[i].Globals == nil {
			valid[i].Globals = globals
		}
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
		}
	}
	for i := range invalid {
		if invalid[i].Globals == nil {
			invalid[i].Globals = globals
		}
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.js"
		}
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &no_path_concat.NoPathConcatRule, valid, invalid)
}

// All cases from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-path-concat.js
func TestNoPathConcatUpstream(t *testing.T) {
	escapedSeparator := "\\" + string(os.PathSeparator)
	runNoPathConcatTests(t, []rule_tester.ValidTestCase{
		{Code: "var fullPath = dirname + \"foo.js\";"},
		{Code: "var fullPath = __dirname == \"foo.js\";"},
		{Code: "if (fullPath === __dirname) {}"},
		{Code: "if (__dirname === fullPath) {}"},
		{Code: "var fullPath = \"/foo.js\" + __filename;"},
		{Code: "var fullPath = \"/foo.js\" + __dirname;"},
		{Code: "var fullPath = __filename + \".map\";"},
		{Code: "var fullPath = `${__filename}.map`;"},
		{Code: "var fullPath = __filename + (test ? \".js\" : \".ts\");"},
		{Code: "var fullPath = __filename + (ext || \".js\");"},
		{Code: "var fullPath = import.meta.dirname + \".map\";"},
		{Code: "var fullPath = import.meta.filename + \".map\";"},
		{Code: "var fullUrl = import.meta.url + \".map\";"},
	}, []rule_tester.InvalidTestCase{
		{Code: "var fullPath = __dirname + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 37)}},
		{Code: "var fullPath = __filename + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 38)}},
		{Code: "var fullPath = `${__dirname}/foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 37)}},
		{Code: "var fullPath = `${__filename}/foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 38)}},
		{Code: "var path = require(\"path\"); var fullPath = `${__dirname}${path.sep}foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 44, 1, 75)}},
		{Code: "var path = require(\"path\"); var fullPath = `${__filename}${path.sep}foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 44, 1, 76)}},
		{Code: "var path = require(\"path\"); var fullPath = __dirname + path.sep + `foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 44, 1, 64)}},
		{Code: "var fullPath = __dirname + \"/\" + \"foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 31)}},
		{Code: "var fullPath = __dirname + (\"/\" + \"foo.js\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 44)}},
		{Code: "var fullPath = __dirname + (test ? \"/foo.js\" : \"/bar.js\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 58)}},
		{Code: "var fullPath = __dirname + (extraPath || \"/default.js\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 56)}},
		{Code: "var fullPath = __dirname + \"" + escapedSeparator + "foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 38)}},
		{Code: "var fullPath = __filename + \"" + escapedSeparator + "foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 39)}},
		{Code: "var fullPath = `${__dirname}" + escapedSeparator + "foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 38)}},
		{Code: "var fullPath = `${__filename}" + escapedSeparator + "foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 39)}},
		{Code: "var fullPath = import.meta.dirname + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 47)}},
		{Code: "var fullPath = import.meta.filename + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 48)}},
		{Code: "var fullUrl = import.meta.url + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("useUrl", 1, 15, 1, 42)}},
		{Code: "var fullUrl = import.meta[\"url\"] + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("useUrl", 1, 15, 1, 45)}},
		{Code: "var fullUrl = `${import.meta[`url`]}/foo.js`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("useUrl", 1, 15, 1, 45)}},
	})
}

// All cases from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-path-concat.md
func TestNoPathConcatDocumentation(t *testing.T) {
	runNoPathConcatTests(t, []rule_tester.ValidTestCase{
		{Code: "var fullPath = path.join(__dirname, \"foo.js\");"},
		{Code: "var fullPath = path.resolve(__dirname, \"foo.js\");"},
		{Code: "const url = new URL(\"./foo.js\", import.meta.url)"},
		{Code: "const fullPath1 = path.join(__dirname, \"foo.js\");\nconst fullPath2 = path.join(__filename, \"foo.js\");\nconst fullPath3 = __dirname + \".js\";\nconst fullPath4 = __filename + \".map\";\nconst fullPath5 = `${__dirname}_foo.js`;\nconst fullPath6 = `${__filename}.test.js`;\nconst fullPath7 = path.join(import.meta.dirname, \"foo.js\");\nconst fullPath8 = path.join(import.meta.filename, \"foo.js\");\nconst fullUrl = new URL(\"./foo.js\", import.meta.url);"},
	}, []rule_tester.InvalidTestCase{
		{Code: "var fullPath = __dirname + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 16, 1, 37)}},
		{Code: "const fullPath1 = __dirname + \"/foo.js\";\nconst fullPath2 = __filename + \"/foo.js\";\nconst fullPath3 = `${__dirname}/foo.js`;\nconst fullPath4 = `${__filename}/foo.js`;\nconst fullPath5 = import.meta.dirname + \"/foo.js\";\nconst fullPath6 = import.meta.filename + \"/foo.js\";\nconst fullUrl = import.meta.url + \"/foo.js\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 19, 1, 40), concatErrorAt("usePathFunctions", 2, 19, 2, 41), concatErrorAt("usePathFunctions", 3, 19, 3, 40), concatErrorAt("usePathFunctions", 4, 19, 4, 41), concatErrorAt("usePathFunctions", 5, 19, 5, 50), concatErrorAt("usePathFunctions", 6, 19, 6, 51), concatErrorAt("useUrl", 7, 17, 7, 44)}},
	})
}
