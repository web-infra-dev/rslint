package no_unpublished_require

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func unpublishedRequireRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-unpublished-require-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil || len(names) == 0 {
		t.Fatalf("unpublished require fixtures: %v", err)
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

// All 63 upstream cases, including two explained filename-less skips.
// The docs have configuration examples only; ignorePrivate is covered here,
// and their invalid convertPath: null example is covered by the schema test.
// Messages and complete ranges checked with ESLint 10.2.1 and eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unpublished-require.js
func TestNoUnpublishedRequireUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Upstream valid 1
		{Code: "require('fs');", FileName: "1/test.js"},
		// Upstream valid 2
		{Code: "require('aaa');", FileName: "1/test.js"},
		// Upstream valid 3
		{Code: "require('aaa/a/b/c');", FileName: "1/test.js"},
		// Upstream valid 4
		{Code: "require('./a');", FileName: "1/test.js"},
		// Upstream valid 5
		{Code: "require('./a.js');", FileName: "1/test.js"},
		// Upstream valid 6
		{Code: "require('./test');", FileName: "2/ignore1.js"},
		// Upstream valid 7
		{Code: "require('bbb');", FileName: "2/ignore1.js"},
		// Upstream valid 8
		{Code: "require('bbb/a/b/c');", FileName: "2/ignore1.js"},
		// Upstream valid 9
		{Code: "require('./ignore2');", FileName: "2/ignore1.js"},
		// Upstream valid 10
		{Code: "require('./pub/a');", FileName: "3/test.js"},
		// Upstream valid 11
		{Code: "require('./test2');", FileName: "3/test.js"},
		// Upstream valid 12
		{Code: "require('aaa');", FileName: "3/test.js"},
		// Upstream valid 13
		{Code: "require('bbb');", FileName: "3/test.js"},
		// Upstream valid 14
		{Code: "require('bbb');", FileName: "3/pub/ignore1.js"},
		// Upstream valid 15
		{Code: "require('../package.json');", FileName: "3/pub/test.js"},
		// Upstream valid 16
		{Code: "require('bbb');", FileName: "3/src/pub/test.js"},
		// Upstream valid 17
		{Code: "require('bbb!foo?a=b&c=d');", FileName: "3/src/pub/test.js"},
		// Upstream valid 18
		{Code: "require('./a');", FileName: "3/src/test.jsx", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"src/**/*.jsx": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}, "tryExtensions": []any{".js", ".jsx", ".json"}}}},
		// Upstream valid 19
		{Code: "require('./a');", FileName: "3/src/test.jsx", Options: []any{map[string]any{"convertPath": map[string]any{"src/**/*.jsx": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}, "tryExtensions": []any{".js", ".jsx", ".json"}}}},
		// Upstream valid 20
		{Code: "require('../test');", FileName: "3/src/test.jsx", Settings: map[string]any{"node": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**/*.jsx"}, "exclude": []any{"**/test.jsx"}, "replace": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}}},
		// Upstream valid 21
		{Code: "require('../test');", FileName: "3/src/test.jsx", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**/*.jsx"}, "exclude": []any{"**/test.jsx"}, "replace": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}}},
		// Upstream valid 22
		{Code: "require;", FileName: "1/test.js"},
		// Upstream valid 23
		{Code: "require('no-exist-package-0');", FileName: "1/test.js"},
		// Upstream valid 24: Upstream <input>: the native parser and lint API require a real filename.
		{Code: "require('no-exist-package-0');", Skip: true},
		// Upstream valid 25: Upstream <input>: the native parser and lint API require a real filename.
		{Code: "require('./b');", Skip: true},
		// Upstream valid 26
		{Code: "require();", FileName: "1/test.js"},
		// Upstream valid 27
		{Code: "require(foo);", FileName: "1/test.js"},
		// Upstream valid 28
		{Code: "require(777);", FileName: "1/test.js"},
		// Upstream valid 29
		{Code: "require(`foo${bar}`);", FileName: "1/test.js"},
		// Upstream valid 30
		{Code: "require('aaa');", FileName: "2/test.js"},
		// Upstream valid 31
		{Code: "require('./a');", FileName: "2/test.js"},
		// Upstream valid 32
		{Code: "require('.');", FileName: "issue48n/test.js"},
		// Upstream valid 33
		{Code: "require('./');", FileName: "issue48n/test.js"},
		// Upstream valid 34
		{Code: "require('..');", FileName: "issue48n/test/test.js"},
		// Upstream valid 35
		{Code: "require('./index.js');", FileName: "issue99/test/bin.js"},
		// Upstream valid 36
		{Code: "require('.');", FileName: "issue99/test/bin.js"},
		// Upstream valid 37
		{Code: "require('./src/helper.js');", FileName: "brace-extglob/index.js"},
		// Upstream valid 38
		{Code: "require('./Foo.js');", FileName: "case-insensitive/index.js"},
		// Upstream valid 39
		{Code: "require('./file1.js');", FileName: "brace-sequence/index.js"},
		// Upstream valid 40
		{Code: "require('electron');", FileName: "1/test.js", Options: []any{map[string]any{"allowModules": []any{"electron"}}}},
		// Upstream valid 41
		{Code: "require('virtual:package-name');", FileName: "test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-name"}}}},
		// Upstream valid 42
		{Code: "require('virtual:package-scope/name');", FileName: "test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-scope"}}}},
		// Upstream valid 43
		{Code: "require('bbb');", FileName: "3/src/readme.js"},
		// Upstream valid 44
		{Code: "require('bbb');", FileName: "negative-in-files/lib/__test__/index.js"},
		// Upstream valid 45
		{Code: "require('bbb');", FileName: "issue126/lib/test.js"},
		// Upstream valid 46
		{Code: "require('bbb');", FileName: "private-package/index.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Upstream invalid 1
		{Code: "require('./src/helper.test.js');", FileName: "brace-extglob/index.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./src/helper.test.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 31}}},
		// Upstream invalid 2
		{Code: "require('./src/test.js');", FileName: "extended-basename-exclusion/index.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./src/test.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 24}}},
		// Upstream invalid 3
		{Code: "require('./ignore1.js');", FileName: "2/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 23}}},
		// Upstream invalid 4
		{Code: "require('./ignore1');", FileName: "2/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// Upstream invalid 5
		{Code: "require('bbb');", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"bbb\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// Upstream invalid 6
		{Code: "require('./ignore1');", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// Upstream invalid 7
		{Code: "require('./abc');", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./abc\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// Upstream invalid 8
		{Code: "require('../test');", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid 9
		{Code: "require('../src/pub/a.js');", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../src/pub/a.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26}}},
		// Upstream invalid 10
		{Code: "require('../a.js');", FileName: "1/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../a.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid 11
		{Code: "require('../test');", FileName: "3/src/test.jsx", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"src/**/*.jsx": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid 12
		{Code: "require('../test');", FileName: "3/src/test.jsx", Options: []any{map[string]any{"convertPath": map[string]any{"src/**/*.jsx": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid 13
		{Code: "require('../test');", FileName: "3/src/test.jsx", Settings: map[string]any{"node": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**/*.jsx"}, "replace": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid 14
		{Code: "require('../test');", FileName: "3/src/test.jsx", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**/*.jsx"}, "replace": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid 15
		{Code: "require('./ignore1');", FileName: "2/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// Upstream invalid 16
		{Code: "require('../2/a.js');", FileName: "1/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../2/a.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// Upstream invalid 17
		{Code: "require('bbb');", FileName: "private-package/index.js", Options: []any{map[string]any{"ignorePrivate": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"bbb\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		valid[i].Globals = map[string]any{"require": "readonly", "global": "readonly"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		invalid[i].Globals = map[string]any{"require": "readonly", "global": "readonly"}
	}
	rule_tester.RunRuleTester(unpublishedRequireRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoUnpublishedRequireRule, valid, invalid)
}
