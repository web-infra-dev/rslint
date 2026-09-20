package no_unpublished_import

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"testing"
)

func unpublishedImportRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-unpublished-import-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"nodenext","jsx":"preserve"},"include":["**/*"]}`}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

// All 48 upstream cases; the two filename-less cases remain explained skips.
// Includes the documentation's type-import and private-package examples.
// Exact messages and ranges verified against ESLint 10.2.1 and eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unpublished-import.js
func TestNoUnpublishedImportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(unpublishedImportRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoUnpublishedImportRule,
		[]rule_tester.ValidTestCase{
			// Published and ignored importing files.
			{Code: "import fs from 'fs';", FileName: "1/test.js"},
			{Code: "import aaa from 'aaa'; aaa();", FileName: "1/test.js"},
			{Code: "import c from 'aaa/a/b/c';", FileName: "1/test.js"},
			{Code: "import a from './a';", FileName: "1/test.js"},
			{Code: "import a from './a.js';", FileName: "1/test.js"},
			{Code: "import test from './test';", FileName: "2/ignore1.js"},
			{Code: "import bbb from 'bbb';", FileName: "2/ignore1.js"},
			{Code: "import c from 'bbb/a/b/c';", FileName: "2/ignore1.js"},
			{Code: "import ignore2 from './ignore2';", FileName: "2/ignore1.js"},
			{Code: "import a from './pub/a';", FileName: "3/test.js"},
			{Code: "import test2 from './test2';", FileName: "3/test.js"},
			{Code: "import aaa from 'aaa';", FileName: "3/test.js"},
			{Code: "import bbb from 'bbb';", FileName: "3/test.js"},
			{Code: "import bbb from 'bbb';", FileName: "3/pub/ignore1.js"},
			{Code: "import p from '../package.json';", FileName: "3/pub/test.js"},
			{Code: "import bbb from 'bbb';", FileName: "3/src/pub/test.js"},
			{Code: "import bbb from 'bbb!foo?a=b&c=d';", FileName: "3/src/pub/test.js"},
			// Upstream <input>: the native parser and lint API require a real filename.
			{Code: "import noExistPackage0 from 'no-exist-package-0';", Skip: true},
			// Upstream <input>: the native parser and lint API require a real filename.
			{Code: "import b from './b';", Skip: true},
			// Relative filenames.
			{Code: "import aaa from 'aaa';", FileName: "2/test.js"},
			{Code: "import a from './a';", FileName: "2/test.js"},
			// Allowed modules, including virtual package roots.
			{Code: "import electron from 'electron';", FileName: "1/test.js", Options: []any{map[string]any{"allowModules": []any{"electron"}}}},
			{Code: "import a from 'virtual:package-name';", FileName: "test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-name"}}}},
			{Code: "import a from 'virtual:package-scope/name';", FileName: "test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-scope"}}}},
			// Extension configuration.
			{Code: "import abc from './abc';", FileName: "4/index.jsx", Options: []any{map[string]any{"tryExtensions": []any{".jsx"}}}},
			// Automatic publication applies only at the package root.
			{Code: "import bbb from 'bbb';", FileName: "3/src/readme.js"},
			// Negative files patterns.
			{Code: "import bbb from 'bbb';", FileName: "negative-in-files/lib/__test__/index.js"},
			// Private packages.
			{Code: "import bbb from 'bbb';", FileName: "private-package/index.js"},
			// Whole type imports can be ignored (upstream issue 78).
			{Code: "import type foo from 'foo';", FileName: "1/test.ts", Options: []any{map[string]any{"ignoreTypeImport": true}}},
			// TypeScript aliases and allowed packages (upstream issue 421).
			{Code: "import foo from '@test/dev'", FileName: "tsconfig-paths-wildcard/index.ts", Options: []any{map[string]any{"allowModules": []any{"@test/dev"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// Unpublished local paths and development dependencies.
			{Code: "import ignore1 from './ignore1.js';", FileName: "2/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1.js\" is not published.", Line: 1, Column: 21, EndLine: 1, EndColumn: 35}}},
			{Code: "import bbb from 'bbb';", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"bbb\" is not published.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}}},
			{Code: "import ignore1 from './ignore1.js';", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1.js\" is not published.", Line: 1, Column: 21, EndLine: 1, EndColumn: 35}}},
			{Code: "import abc from './abc.json';", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./abc.json\" is not published.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29}}},
			{Code: "import test from '../test';", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test\" is not published.", Line: 1, Column: 18, EndLine: 1, EndColumn: 27}}},
			{Code: "import a from '../src/pub/a.js';", FileName: "3/pub/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../src/pub/a.js\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 32}}},
			{Code: "import a from '../a.js';", FileName: "1/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../a.js\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}}},
			// Relative filename.
			{Code: "import ignore1 from './ignore1.js';", FileName: "2/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1.js\" is not published.", Line: 1, Column: 21, EndLine: 1, EndColumn: 35}}},
			// Object and array path conversion, through settings and options.
			{Code: "import a from '../test.jsx';", FileName: "3/src/test.jsx", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"src/**/*.jsx": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test.jsx\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
			{Code: "import a from '../test.jsx';", FileName: "3/src/test.jsx", Options: []any{map[string]any{"convertPath": map[string]any{"src/**/*.jsx": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test.jsx\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
			{Code: "import a from '../test.jsx';", FileName: "3/src/test.jsx", Settings: map[string]any{"node": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**/*.jsx"}, "replace": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test.jsx\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
			{Code: "import a from '../test.jsx';", FileName: "3/src/test.jsx", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**/*.jsx"}, "replace": []any{"src/(.+?)\\.jsx", "pub/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../test.jsx\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
			// Default extensions do not include JSX.
			{Code: "import abc from './abc';", FileName: "4/index.jsx", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./abc\" is not published.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24}}},
			// Paths outside the package.
			{Code: "import a from '../2/a.js';", FileName: "1/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"../2/a.js\" is not published.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26}}},
			// Dynamic imports.
			{Code: "function f() { import('./ignore1.js') }", FileName: "2/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./ignore1.js\" is not published.", Line: 1, Column: 23, EndLine: 1, EndColumn: 37}}},
			// Type imports are checked by default (upstream issue 78).
			{Code: "import type foo from 'foo';", FileName: "1/test.ts", Options: []any{map[string]any{"ignoreTypeImport": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"foo\" is not published.", Line: 1, Column: 22, EndLine: 1, EndColumn: 27}}},
			{Code: "import type foo from 'foo';", FileName: "1/test.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"foo\" is not published.", Line: 1, Column: 22, EndLine: 1, EndColumn: 27}}},
			// Private packages can opt into checking.
			{Code: "import bbb from 'bbb';", FileName: "private-package/index.js", Options: []any{map[string]any{"ignorePrivate": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"bbb\" is not published.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}}},
		},
	)
}
