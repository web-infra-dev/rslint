package no_extraneous_require

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func extraneousRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-extraneous-require-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"nodenext","jsx":"preserve"},"include":["**/*"]}`,
	}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

// All 23 cases from eslint-plugin-n v18.3.0. Its documentation has options
// examples but no additional source examples. Complete ranges and messages
// were checked with ESLint 10.2.1 and the pinned upstream implementation.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-extraneous-require.js
func TestNoExtraneousRequireUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "$.require('bbb')", FileName: "dependencies/a.js"},
		{Code: "require('./bbb')", FileName: "dependencies/a.js"},
		{Code: "require(bbb)", FileName: "dependencies/a.js"},
		{Code: "require('aaa')", FileName: "dependencies/a.js"},
		{Code: "require('aaa/bbb')", FileName: "dependencies/a.js"},
		{Code: "require('@bbb/aaa')", FileName: "dependencies/a.js"},
		{Code: "require('@bbb/aaa/bbb')", FileName: "dependencies/a.js"},
		{Code: "require('aaa')", FileName: "devDependencies/a.js"},
		{Code: "require('aaa')", FileName: "peerDependencies/a.js"},
		{Code: "require('aaa')", FileName: "optionalDependencies/a.js"},
		{Code: "require('root-dep')", FileName: "workspace/packages/app/src/index.js", Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
		{Code: "require('root-dep')", FileName: "workspace-object/packages/app/src/index.js", Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
		{Code: "require('root-dep')", FileName: "workspace-nested/inner/packages/app/src/index.js", Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
		{Code: "require('ccc')", FileName: "dependencies/a.js"},
		{Code: "require('virtual:package-name');", FileName: "test.js"},
		{Code: "require('virtual:package-scope/name');", FileName: "test.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "require('bbb')", FileName: "dependencies/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		{Code: "require('bbb')", FileName: "devDependencies/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		{Code: "require('bbb')", FileName: "peerDependencies/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		{Code: "require('bbb')", FileName: "optionalDependencies/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		{Code: "require('root-dep')", FileName: "workspace-negated/packages/excluded/src/index.js", Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"root-dep\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}}},
		{Code: "require('outer-dep')", FileName: "workspace-nested/inner/packages/app/src/index.js", Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"outer-dep\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		{Code: "require('b'+'bb')", FileName: "dependencies/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
	}
	// Upstream's default tester enables CommonJS and Node globals.
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		valid[i].Globals = map[string]any{"require": "readonly", "global": "readonly"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		invalid[i].Globals = map[string]any{"require": "readonly", "global": "readonly"}
	}
	rule_tester.RunRuleTester(extraneousRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoExtraneousRequireRule, valid, invalid)
}
