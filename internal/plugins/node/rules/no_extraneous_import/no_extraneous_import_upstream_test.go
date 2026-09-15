package no_extraneous_import

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func extraneousRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-extraneous-fixtures")
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

// All 30 cases from eslint-plugin-n v18.3.0, including dynamic import and
// TypeScript groups. The rule documentation contains configuration examples
// but no additional source examples. Locations below were checked with ESLint.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-extraneous-import.js
func TestNoExtraneousImportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(extraneousRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoExtraneousImportRule,
		[]rule_tester.ValidTestCase{
			{Code: "import bbb from './bbb'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import aaa from 'aaa'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import bbb from 'aaa/bbb'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import aaa from '@bbb/aaa'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import bbb from '@bbb/aaa/bbb'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import aaa from 'aaa'", FileName: "devDependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import aaa from 'aaa'", FileName: "peerDependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import aaa from 'aaa'", FileName: "optionalDependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import '#b'", FileName: "import-map/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import rootDep from 'root-dep'", FileName: "workspace/packages/app/src/index.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
			{Code: "import rootDevDep from 'root-dev-dep'", FileName: "workspace/packages/app/src/index.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
			{Code: "import rootDep from 'root-dep'", FileName: "workspace-object/packages/app/src/index.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
			{Code: "import rootDep from 'root-dep'", FileName: "workspace-nested/inner/packages/app/src/index.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
			{Code: "import ccc from 'ccc'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import foo from '@configurations/foo'", FileName: "tsconfig-paths/index.ts", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import foo from '~configurations/foo'", FileName: "tsconfig-paths/index.ts", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import foo from '#configurations/foo'", FileName: "tsconfig-paths/index.ts", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import a from 'virtual:package-name';", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
			{Code: "import a from 'virtual:package-scope/name';", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}},
		}, []rule_tester.InvalidTestCase{
			{Code: "import bbb from 'bbb'", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}}},
			{Code: "import bbb from 'bbb'", FileName: "devDependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}}},
			{Code: "import bbb from 'bbb'", FileName: "peerDependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}}},
			{Code: "import bbb from 'bbb'", FileName: "optionalDependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}}},
			{Code: "import rootDep from 'root-dep'", FileName: "workspace-negated/packages/excluded/src/index.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"root-dep\" is extraneous.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}}},
			{Code: "import outerDep from 'outer-dep'", FileName: "workspace-nested/inner/packages/app/src/index.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"outer-dep\" is extraneous.", Line: 1, Column: 22, EndLine: 1, EndColumn: 33}}},
			{Code: "function f() { import('bbb') }", FileName: "dependencies/a.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ts"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bbb\" is extraneous.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28}}},
		})
}
func TestNoExtraneousImportUpstreamTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(extraneousRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoExtraneousImportRule,
		[]rule_tester.ValidTestCase{
			{Code: "import type { Glob } from 'picomatch'", FileName: "typesOnly/a.ts"},
			{Code: "export type { Glob } from 'picomatch'", FileName: "typesOnly/a.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "import { scan } from 'picomatch'", FileName: "typesOnly/a.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"picomatch\" is extraneous.", Line: 1, Column: 22, EndLine: 1, EndColumn: 33}}},
			{Code: "export { scan } from 'picomatch'", FileName: "typesOnly/a.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"picomatch\" is extraneous.", Line: 1, Column: 22, EndLine: 1, EndColumn: 33}}},
		})
}
