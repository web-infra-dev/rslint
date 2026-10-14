package no_internal_modules_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_internal_modules"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"testing"
)

func internalRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-internal-modules")
	archive := txtarfs.MustParseFile(t, "testdata/modules.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("fixture archive is empty")
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

func runInternalTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "internal-modules/plugins/plugin.js"
		}
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = "internal-modules/plugins/plugin.js"
		}
	}
	rule_tester.RunRuleTester(internalRoot(t), "tsconfig.json", t, &no_internal_modules.NoInternalModulesRule, valid, invalid)
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-internal-modules.js
// All upstream cases; TS parser variants share the native parser. The webpack
// resolver case is retained as a skip: native rules cannot execute JS resolvers.
func TestNoInternalModulesUpstream(t *testing.T) {
	runInternalTests(t, []rule_tester.ValidTestCase{
		// Imports.
		{Code: "import a from \"./plugin2\"", FileName: "internal-modules/plugins/plugin.js"},
		{Code: "const a = require(\"./plugin2\")", FileName: "internal-modules/plugins/plugin.js"},
		{Code: "const a = require(\"./plugin2/\")", FileName: "internal-modules/plugins/plugin.js"},
		{Code: "const dynamic = \"./plugin2/\"; const a = require(dynamic)", FileName: "internal-modules/plugins/plugin.js"},
		{Code: "import b from \"./internal.js\"", FileName: "internal-modules/plugins/plugin2/index.js"},
		{Code: "import get from \"lodash.get\"", FileName: "internal-modules/plugins/plugin2/index.js"},
		{Code: "import b from \"@org/package\"", FileName: "internal-modules/plugins/plugin2/internal.js"},
		{Code: "import b from \"../../api/service\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"allow": []any{"**/api/*"}}}},
		{Code: "import \"jquery/dist/jquery\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"allow": []any{"jquery/dist/*"}}}},
		{Code: "import \"./app/index.js\";\nimport \"./app/index\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"allow": []any{"**/index{.js,}"}}}},
		{Code: "import a from \"./plugin2/thing\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"forbid": []any{"**/api/*"}}}},
		{Code: "const a = require(\"./plugin2/thing\")", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"forbid": []any{"**/api/*"}}}},
		{Code: "import b from \"app/a\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"app/**/**"}}}},
		{Code: "import b from \"@org/package\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"@org/package/*"}}}},
		// Exports.
		{Code: "export {a} from \"./internal.js\"", FileName: "internal-modules/plugins/plugin2/index.js"},
		{Code: "export * from \"lodash.get\"", FileName: "internal-modules/plugins/plugin2/index.js"},
		{Code: "export {b} from \"@org/package\"", FileName: "internal-modules/plugins/plugin2/internal.js"},
		{Code: "export {b} from \"../../api/service\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"allow": []any{"**/api/*"}}}},
		{Code: "export * from \"jquery/dist/jquery\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"allow": []any{"jquery/dist/*"}}}},
		{Code: "export * from \"./app/index.js\";\nexport * from \"./app/index\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"allow": []any{"**/index{.js,}"}}}},
		{Code: "\n        export class AuthHelper {\n\n          static checkAuth(auth) {\n          }\n        }\n      "},
		{FileName: "internal-modules/plugins/plugin.ts", Code: "\n          export class AuthHelper {\n\n            public static checkAuth(auth?: string): boolean {\n            }\n          }\n        "},
		{Code: "export * from \"./plugin2/thing\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"forbid": []any{"**/api/*"}}}},
		{Code: "export * from \"app/a\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"app/**/**"}}}},
		{Code: "export { b } from \"@org/package\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"@org/package/*"}}}},
		{Code: "export * from \"./app/index.js\";\nexport * from \"./app/index\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"**/index.ts"}}}},
	}, []rule_tester.InvalidTestCase{
		// Imports.
		{Code: "import \"./plugin2/index.js\";\nimport \"./plugin2/app/index\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"allow": []any{"*/index.js"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/app/index\" is not allowed.", Line: 2, Column: 8, EndLine: 2, EndColumn: 29}}},
		{Code: "import \"./app/index.js\"", FileName: "internal-modules/plugins/plugin2/internal.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./app/index.js\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		{Code: "import b from \"./plugin2/internal\"", FileName: "internal-modules/plugins/plugin.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 35}}},
		{Code: "import a from \"../api/service/index\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"allow": []any{"**/internal-modules/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"../api/service/index\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 37}}},
		{Code: "import b from \"@org/package/internal\"", FileName: "internal-modules/plugins/plugin2/internal.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"@org/package/internal\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 38}}},
		{Code: "import get from \"debug/node\"", FileName: "internal-modules/plugins/plugin.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"debug/node\" is not allowed.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29}}},
		{Code: "import \"./app/index.js\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"*/app/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./app/index.js\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		{Code: "import b from \"@org/package\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"@org/**"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"@org/package\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 29}}},
		{Code: "import b from \"app/a/b\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"app/**/**"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"app/a/b\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}}},
		{Code: "import get from \"lodash.get\"", FileName: "internal-modules/plugins/plugin2/index.js", Options: []any{map[string]any{"forbid": []any{"lodash.*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"lodash.get\" is not allowed.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29}}},
		{Code: "import \"./app/index.js\";\nimport \"./app/index\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"**/index{.js,}"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./app/index.js\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}, {Message: "Reaching to \"./app/index\" is not allowed.", Line: 2, Column: 8, EndLine: 2, EndColumn: 21}}},
		{Code: "import \"@/api/service\";", Options: []any{map[string]any{"forbid": []any{"**/api/*"}}}, Settings: map[string]any{"import/resolver": map[string]any{"webpack": map[string]any{"config": map[string]any{"resolve": map[string]any{"alias": map[string]any{"@": "internal-modules"}}}}}}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"@/api/service\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// Exports.
		{Code: "export * from \"./plugin2/index.js\";\nexport * from \"./plugin2/app/index\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"allow": []any{"*/index.js"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/app/index\" is not allowed.", Line: 2, Column: 15, EndLine: 2, EndColumn: 36}}},
		{Code: "export * from \"./app/index.js\"", FileName: "internal-modules/plugins/plugin2/internal.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./app/index.js\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 31}}},
		{Code: "export {b} from \"./plugin2/internal\"", FileName: "internal-modules/plugins/plugin.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 17, EndLine: 1, EndColumn: 37}}},
		{Code: "export {a} from \"../api/service/index\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"allow": []any{"**/internal-modules/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"../api/service/index\" is not allowed.", Line: 1, Column: 17, EndLine: 1, EndColumn: 39}}},
		{Code: "export {b} from \"@org/package/internal\"", FileName: "internal-modules/plugins/plugin2/internal.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"@org/package/internal\" is not allowed.", Line: 1, Column: 17, EndLine: 1, EndColumn: 40}}},
		{Code: "export {get} from \"debug/node\"", FileName: "internal-modules/plugins/plugin.js", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"debug/node\" is not allowed.", Line: 1, Column: 19, EndLine: 1, EndColumn: 31}}},
		{Code: "export * from \"./plugin2/thing\"", FileName: "internal-modules/plugins/plugin.js", Options: []any{map[string]any{"forbid": []any{"**/plugin2/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/thing\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 32}}},
		{Code: "export * from \"app/a\"", FileName: "internal-modules/plugins/plugin2/internal.js", Options: []any{map[string]any{"forbid": []any{"**"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"app/a\" is not allowed.", Line: 1, Column: 15, EndLine: 1, EndColumn: 22}}},
	})
}

// All executable examples from the pinned documentation. The proposal-only
// export-default-from example is retained as a skip; native named reexports
// cover its module check in the extras suite.
func TestNoInternalModulesUpstreamDocs(t *testing.T) {
	runInternalTests(t, []rule_tester.ValidTestCase{{FileName: "my-project/entry.js", Code: "import 'source-map-support/register';\nimport { settings } from '../app';\nimport getUser from '../actions/getUser';\nexport * from 'source-map-support/register';\nexport { settings } from '../app';", Options: []any{map[string]any{"allow": []any{"**/actions/*", "source-map-support/*"}}}},
		{FileName: "my-project/entry.js", Code: "import 'source-map-support';\nimport { getUser } from '../actions';\nexport * from 'source-map-support';\nexport { getUser } from '../actions';", Options: []any{map[string]any{"forbid": []any{"**/actions/*", "source-map-support/*"}}}}}, []rule_tester.InvalidTestCase{{FileName: "my-project/entry.js", Code: "import { settings } from './app/index';\nimport userReducer from './reducer/user';\nimport configureStore from './redux/configureStore';\nexport { settings } from './app/index';\nexport * from './reducer/user';", Options: []any{map[string]any{"allow": []any{"**/actions/*", "source-map-support/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./app/index\" is not allowed.", Line: 1, Column: 26, EndLine: 1, EndColumn: 39}, {Message: "Reaching to \"./reducer/user\" is not allowed.", Line: 2, Column: 25, EndLine: 2, EndColumn: 41}, {Message: "Reaching to \"./redux/configureStore\" is not allowed.", Line: 3, Column: 28, EndLine: 3, EndColumn: 52}, {Message: "Reaching to \"./app/index\" is not allowed.", Line: 4, Column: 26, EndLine: 4, EndColumn: 39}, {Message: "Reaching to \"./reducer/user\" is not allowed.", Line: 5, Column: 15, EndLine: 5, EndColumn: 31}}},
		{FileName: "my-project/entry.js", Code: "import 'source-map-support/register';\nimport getUser from '../actions/getUser';\nexport * from 'source-map-support/register';", Options: []any{map[string]any{"forbid": []any{"**/actions/*", "source-map-support/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"source-map-support/register\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 37}, {Message: "Reaching to \"../actions/getUser\" is not allowed.", Line: 2, Column: 21, EndLine: 2, EndColumn: 41}, {Message: "Reaching to \"source-map-support/register\" is not allowed.", Line: 3, Column: 15, EndLine: 3, EndColumn: 44}}},
		{FileName: "my-project/entry.js", Code: "export getUser from '../actions/getUser';", Options: []any{map[string]any{"forbid": []any{"**/actions/*", "source-map-support/*"}}}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"../actions/getUser\" is not allowed.", Line: 1, Column: 21, EndLine: 1, EndColumn: 41}}}})
}
