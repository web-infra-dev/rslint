// cspell:ignore yargs
package file_extension_in_import

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func extensionRoot(t *testing.T, archives ...string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-file-extension")
	files := map[string]string{tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"nodenext","jsx":"preserve"},"include":["**/*"]}`}
	for _, archiveName := range archives {
		archive := txtarfs.MustParseFile(t, archiveName)
		names, err := archive.FileNames("")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			data, err := archive.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			files[tspath.ResolvePath(directory, name)] = string(data)
		}
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

// All 62 upstream cases and six documentation examples from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/file-extension-in-import.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/file-extension-in-import.md
func TestFileExtensionInImportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(extensionRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &FileExtensionInImportRule,
		[]rule_tester.ValidTestCase{
			// upstream valid 1
			{Code: "import 'eslint'", FileName: "test.js"},
			// upstream valid 2
			{Code: "import '@typescript-eslint/parser'", FileName: "test.js"},
			// upstream valid 3
			{Code: "import '@typescript-eslint\\parser'", FileName: "test.js"},
			// upstream valid 4
			{Code: "import 'punycode/'", FileName: "test.js"},
			// upstream valid 5
			{Code: "import 'xxx'", FileName: "test.js"},
			// upstream valid 6
			{Code: "import './a.js'", FileName: "test.js"},
			// upstream valid 7
			{Code: "import './b.json'", FileName: "test.js"},
			// upstream valid 8
			{Code: "import './c.mjs'", FileName: "test.js"},
			// upstream valid 9
			{Code: "import './d.js'", FileName: "test.js"},
			// upstream valid 10
			{Code: "import './a.js'", FileName: "test.ts"},
			// upstream valid 11
			{Code: "import './d.js'", FileName: "test.ts"},
			// upstream valid 12
			{Code: "import './a.js'", FileName: "test.js", Options: []any{"always"}},
			// upstream valid 13
			{Code: "import './b.json'", FileName: "test.js", Options: []any{"always"}},
			// upstream valid 14
			{Code: "import './c.mjs'", FileName: "test.js", Options: []any{"always"}},
			// upstream valid 15
			{Code: "import './d.jsx'", FileName: "test.tsx", Options: []any{"always"}},
			// upstream valid 16
			{Code: "import './a'", FileName: "test.js", Options: []any{"never"}},
			// upstream valid 17
			{Code: "import './b'", FileName: "test.js", Options: []any{"never"}},
			// upstream valid 18
			{Code: "import './c'", FileName: "test.js", Options: []any{"never"}},
			// upstream valid 19
			{Code: "import './a'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}},
			// upstream valid 20
			{Code: "import './b.json'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}},
			// upstream valid 21
			{Code: "import './c.mjs'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}},
			// upstream valid 22
			{Code: "import './a'", FileName: "test.js", Options: []any{"never", map[string]any{".json": "always"}}},
			// upstream valid 23
			{Code: "import './b.json'", FileName: "test.js", Options: []any{"never", map[string]any{".json": "always"}}},
			// upstream valid 24
			{Code: "import './c'", FileName: "test.js", Options: []any{"never", map[string]any{".json": "always"}}},
			// upstream valid 25
			{Code: "import '@apollo/client/core'", FileName: "test.js", Options: []any{"always"}},
			// upstream valid 26
			{Code: "import 'yargs/helpers'", FileName: "test.js", Options: []any{"always"}},
			// upstream valid 27
			{Code: "import 'firebase-functions/v1/auth'", FileName: "test.js", Options: []any{"always"}},
			// upstream valid 28
			{Code: "require('./d.js');", FileName: "test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 29
			{Code: "require('./e.js');", FileName: "test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 30
			{Code: "require('./d.js');", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 31
			{Code: "require('./e.js');", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
			// upstream valid 32
			{Code: "require('./file.js');", FileName: "ts-allow-extension/test.ts"},
			// upstream valid 33
			{Code: "require('./file.ts');", FileName: "ts-allow-extension/test.ts"},
			// documentation always correct
			{Code: "import eslint from \"eslint\"\nimport foo from \"./path/to/a/file.js\"", FileName: "test.js", Options: []any{"always"}},
			// documentation never correct
			{Code: "import eslint from \"eslint\"\nimport foo from \"./path/to/a/file\"", FileName: "test.js", Options: []any{"never"}},
			// documentation overrides
			{Code: "import eslint from \"eslint\"\nimport script from \"./script\"\nimport styles from \"./styles.css\"\nimport logo from \"./logo.png\"", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}},
		},
		[]rule_tester.InvalidTestCase{
			// upstream invalid 1
			{Code: "import './a'", FileName: "test.js", Output: []string{"import './a.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 2
			{Code: "import './a'", FileName: "test.ts", Output: []string{"import './a.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 3
			{Code: "import './d'", FileName: "test.ts", Output: []string{"import './d.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 4
			{Code: "import { util } from './my-folder'", FileName: "test.js", Output: []string{"import { util } from './my-folder/index.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 35}}},
			// upstream invalid 5
			{Code: "import { util } from './my-folder/'", FileName: "test.js", Output: []string{"import { util } from './my-folder/index.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 36}}},
			// upstream invalid 6
			{Code: "import './b'", FileName: "test.js", Output: []string{"import './b.json'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 7
			{Code: "import './c'", FileName: "test.js", Output: []string{"import './c.mjs'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 8
			{Code: "import './a'", FileName: "test.js", Options: []any{"always"}, Output: []string{"import './a.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 9
			{Code: "import './b'", FileName: "test.js", Options: []any{"always"}, Output: []string{"import './b.json'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 10
			{Code: "import './c'", FileName: "test.js", Options: []any{"always"}, Output: []string{"import './c.mjs'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 11
			{Code: "import './a.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// upstream invalid 12
			{Code: "import './b.json'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './b'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 13
			{Code: "import './c.mjs'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './c'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 14
			{Code: "import './a.js'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// upstream invalid 15
			{Code: "import './b'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}, Output: []string{"import './b.json'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 16
			{Code: "import './c'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}, Output: []string{"import './c.mjs'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 17
			{Code: "import './a.js'", FileName: "test.js", Options: []any{"never", map[string]any{".json": "always"}}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// upstream invalid 18
			{Code: "import './b'", FileName: "test.js", Options: []any{"never", map[string]any{".json": "always"}}, Output: []string{"import './b.json'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 19
			{Code: "import './c.mjs'", FileName: "test.js", Options: []any{"never", map[string]any{".json": "always"}}, Output: []string{"import './c'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 20
			{Code: "import './multi'", FileName: "test.js", Options: []any{"always"}, Output: []string{"import './multi.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 21
			{Code: "import './multi.js'", FileName: "test.js", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 22
			{Code: "import './multi.json'", FileName: "test.js", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
			// upstream invalid 23
			{Code: "import './utils.client'", FileName: "test.ts", Output: []string{"import './utils.client.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
			// upstream invalid 24
			{Code: "import './util.client'", FileName: "test.ts", Output: []string{"import './util.client.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
			// upstream invalid 25
			{Code: "import './util.client'", FileName: "test.js", Output: []string{"import './util.client.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
			// upstream invalid 26
			{Code: "import './my-things.client'", FileName: "test.ts", Output: []string{"import './my-things.client/index.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
			// upstream invalid 27
			{Code: "import './my-things.client'", FileName: "test.js", Output: []string{"import './my-things.client/index.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
			// upstream invalid 28
			{Code: "function f() { import('./a') }", FileName: "test.js", Output: []string{"function f() { import('./a.js') }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28}}},
			// upstream invalid 29
			{Code: "function f() { import('./a.js') }", FileName: "test.js", Options: []any{"never"}, Output: []string{"function f() { import('./a') }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31}}},
			// documentation introduction
			{Code: "import foo from \"./path/to/a/file\"\nexport * from \"./path/to/a/file\"", FileName: "test.js", Output: []string{"import foo from \"./path/to/a/file.js\"\nexport * from \"./path/to/a/file.js\""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 35}, {MessageId: "requireExt", Message: "require file extension '.js'.", Line: 2, Column: 15, EndLine: 2, EndColumn: 33}}},
			// documentation always incorrect
			{Code: "import foo from \"./path/to/a/file\"", FileName: "test.js", Options: []any{"always"}, Output: []string{"import foo from \"./path/to/a/file.js\""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 35}}},
			// documentation never incorrect
			{Code: "import foo from \"./path/to/a/file.js\"", FileName: "test.js", Options: []any{"never"}, Output: []string{"import foo from \"./path/to/a/file\""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 38}}},
		},
	)
}
