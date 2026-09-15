package no_missing_import

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func missingRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-missing-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{tspath.ResolvePath(directory, "tsconfig.json"): "{\"compilerOptions\":{\"allowJs\":true,\"noEmit\":true,\"target\":\"esnext\",\"module\":\"nodenext\",\"jsx\":\"preserve\"},\"include\":[\"**/*\"]}"}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

// All 79 cases and source examples from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-missing-import.js
func TestNoMissingImportUpstream(t *testing.T) {
	root := missingRoot(t, "testdata/upstream.txtar")
	message := func(text string) string { return strings.ReplaceAll(text, "{{root}}", root.Dir) }
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoMissingImportRule,
		[]rule_tester.ValidTestCase{
			// Upstream <input> cases: tsgo requires absolute filenames, so these
			// cannot be represented by the native parser or lint API.
			{Code: "import abc from 'no-exist-package-0';", Skip: true},
			{Code: "import b from './b';", Skip: true},
			// upstream valid 1
			{Code: "import eslint from 'eslint';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 2
			{Code: "import fs from 'fs';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 3
			{Code: "import fs from 'node:fs';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 4
			{Code: "import eslint from 'eslint'", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 5
			{Code: "import a from './a.js';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 6
			{Code: "import a from './d.js';", FileName: "tests/fixtures/no-missing/test.ts"},
			// upstream valid 7
			{Code: "import aConfig from './a.config.js';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 8
			{Code: "import b from './b.json';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 9
			{Code: "import c from './c.coffee';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 10
			{Code: "import mocha from 'mocha';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 11
			{Code: "import something from 'cjs-module-with-no-main';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 12
			{Code: "import something from 'esm-module';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 13
			{Code: "import something from 'esm-module/sub';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 14
			{Code: "import mocha from 'mocha!foo?a=b&c=d';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 15
			{Code: "import a from './e.jsx';", FileName: "tests/fixtures/no-missing/test.tsx"},
			// upstream valid 16
			{Code: "import 'misconfigured-default';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 17
			{Code: "import './c';", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"tryExtensions": []any{".coffee"}}}},
			// upstream valid 18
			{Code: "import './c';", FileName: "tests/fixtures/no-missing/test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".coffee"}}}},
			// upstream valid 21
			{Code: "const foo=0, bar=1; export {foo, bar};", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 22
			{Code: "import eslint from 'eslint'", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 23
			{Code: "import a from './a.js';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 24
			{Code: "import electron from 'electron';", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"electron"}}}},
			// upstream valid 25
			{Code: "import a from 'virtual:package-name';", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-name"}}}},
			// upstream valid 26
			{Code: "import a from 'virtual:package-scope/name';", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-scope"}}}},
			// upstream valid 27
			{Code: "import a from './fixtures/no-missing/a.js';", FileName: "tests/fixtures/no-missing/test.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"tests"}}}},
			// upstream valid 28
			{Code: "import a from './fixtures/no-missing/a.js';", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolvePaths": []any{"tests"}}}},
			// upstream valid 29
			{Code: "import a from './fixtures/no-missing/a.js';", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolvePaths": []any{"scripts", "tests"}}}},
			// upstream valid 30
			{Code: "import d from './d.js';", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 31
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 32
			{Code: "import d from './d.js';", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 33
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 34
			{Code: "import e from './e.jsx';", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "preserve"}}},
			// upstream valid 35
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react"}}},
			// upstream valid 36
			{Code: "import e from './e.jsx';", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "preserve"}}},
			// upstream valid 37
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react"}}},
			// upstream valid 38
			{Code: "import e from './e.jsx';", FileName: "tests/fixtures/no-missing/ts-react/test.tsx", Options: []any{map[string]any{"tsconfigPath": "tests/fixtures/no-missing/ts-preserve/tsconfig.json"}}},
			// upstream valid 39
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/ts-preserve/test.tsx", Options: []any{map[string]any{"tsconfigPath": "tests/fixtures/no-missing/ts-react/tsconfig.json"}}},
			// upstream valid 40
			{Code: "import e from './e.jsx';", FileName: "tests/fixtures/no-missing/ts-react/test.tsx", Settings: map[string]any{"node": map[string]any{"tsconfigPath": "tests/fixtures/no-missing/ts-preserve/tsconfig.json"}}},
			// upstream valid 41
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/ts-preserve/test.tsx", Settings: map[string]any{"node": map[string]any{"tsconfigPath": "tests/fixtures/no-missing/ts-react/tsconfig.json"}}},
			// upstream valid 42
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/ts-react/test.tsx"},
			// upstream valid 43
			{Code: "import d from './d.js';", FileName: "tests/fixtures/no-missing/ts-react/test.ts"},
			// upstream valid 44
			{Code: "import e from './e.jsx';", FileName: "tests/fixtures/no-missing/ts-preserve/test.tsx"},
			// upstream valid 45
			{Code: "import d from './d.js';", FileName: "tests/fixtures/no-missing/ts-preserve/test.ts"},
			// upstream valid 46
			{Code: "import e from './e.js';", FileName: "tests/fixtures/no-missing/ts-extends/test.tsx"},
			// upstream valid 47
			{Code: "import d from './d.js';", FileName: "tests/fixtures/no-missing/ts-extends/test.ts"},
			// upstream valid 48
			{Code: "import before from '@direct';", FileName: "tests/fixtures/no-missing/ts-paths/test.ts"},
			// upstream valid 49
			{Code: "import before from '@wild/where.js';", FileName: "tests/fixtures/no-missing/ts-paths/test.ts"},
			// upstream valid 50
			{Code: "import('@module');", FileName: "tests/fixtures/no-missing/issue-314/src/example.ts"},
			// upstream valid 51
			{Code: "import type d from 'types-only';", FileName: "tests/fixtures/no-missing/test.ts"},
			// upstream valid 52
			{Code: "import './file.ts';", FileName: "tests/fixtures/no-missing/ts-allow-extension/test.ts"},
			// upstream valid 53
			{Code: "import plugin from 'eslint-plugin-n';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 54
			{Code: "import isIp from '#is-ip';", FileName: "tests/fixtures/no-missing/issue-285/test.js"},
			// upstream valid 55
			{Code: "import type missing from '@type/this-does-not-exists';", FileName: "tests/fixtures/no-missing/test.ts", Options: []any{map[string]any{"ignoreTypeImport": true}}},
			// upstream valid 56
			{Code: "import 'data:text/javascript,const x = 123;';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 57
			{Code: "import 'https://example.com/module.js';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 58
			{Code: "import 'http://localhost/module.js';", FileName: "tests/fixtures/no-missing/test.js"},
			// upstream valid 59
			{Code: "function f() { import(foo) }", FileName: "tests/fixtures/no-missing/test.js"},
			// documentation: existing file and package
			{Code: "import existingFile from \"./existing-file\";\nimport existingModule from \"existing-module\";", FileName: "tests/fixtures/no-missing/test.js"},
			// documentation: ignoreTypeImport
			{Code: "import type { TypeOnly } from \"@types/only-types\";", FileName: "tests/fixtures/no-missing/test.ts", Options: []any{map[string]any{"ignoreTypeImport": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// upstream invalid 1
			{Code: "import abc from 'no-exist-package-0';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'no-exist-package-0' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 17, EndLine: 1, EndColumn: 37}}},
			// upstream invalid 2
			{Code: "import abcdef from 'esm-module/sub.mjs';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("\"./sub.mjs\" is not exported under the conditions [\"node\",\"require\",\"import\"] from package {{root}}/tests/fixtures/no-missing/node_modules/esm-module (see exports field in {{root}}/tests/fixtures/no-missing/node_modules/esm-module/package.json)"), Line: 1, Column: 20, EndLine: 1, EndColumn: 40}}},
			// upstream invalid 3
			{Code: "import test from '@mysticatea/test';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '@mysticatea/test' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 18, EndLine: 1, EndColumn: 36}}},
			// upstream invalid 4
			{Code: "import c from './c';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './c' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 5
			{Code: "import d from './d';", FileName: "tests/fixtures/no-missing/test.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './d' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 6
			{Code: "import d from './d';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './d' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 7
			{Code: "import a from './a.json';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './a.json' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 25}}},
			// upstream invalid 8
			{Code: "import './file.js';", FileName: "tests/fixtures/no-missing/ts-allow-extension/test.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './file.js' in '{{root}}/tests/fixtures/no-missing/ts-allow-extension'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// upstream invalid 9
			{Code: "import eslint from 'no-exist-package-0';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'no-exist-package-0' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 20, EndLine: 1, EndColumn: 40}}},
			// upstream invalid 10
			{Code: "import c from './c';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './c' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 11
			{Code: "import a from './bar';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './bar' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 22}}},
			// upstream invalid 12
			{Code: "import a from './bar/';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './bar/' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 23}}},
			// upstream invalid 13
			{Code: "import a from '.';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '.' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 14
			{Code: "import a from './';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 19}}},
			// upstream invalid 15
			{Code: "import a from './foo';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './foo' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 22}}},
			// upstream invalid 16
			{Code: "import a from './foo/';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './foo/' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 23}}},
			// upstream invalid 17
			{Code: "import a from './A.js';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './A.js' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 23}}},
			// upstream invalid 18
			{Code: "function f() { import('no-exist-package-0') }", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'no-exist-package-0' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 23, EndLine: 1, EndColumn: 43}}},
			// upstream invalid 19
			{Code: "import a from 'virtual:package-name';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'virtual:package-name' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 37}}},
			// upstream invalid 20
			{Code: "import a from 'virtual:package-scope/name';", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'virtual:package-scope/name' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 43}}},
			// documentation: missing file and package
			{Code: "import typoFile from \"./typo-file\";\nimport typoModule from \"typo-module\";", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './typo-file' in '{{root}}/tests/fixtures/no-missing'"), Line: 1, Column: 22, EndLine: 1, EndColumn: 35}, {MessageId: "notFound", Message: message("Can't resolve 'typo-module' in '{{root}}/tests/fixtures/no-missing'"), Line: 2, Column: 24, EndLine: 2, EndColumn: 37}}},
		},
	)
}
