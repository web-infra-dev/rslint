package no_import_module_exports_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_import_module_exports"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// cspell:ignore funs starwars
const errorMessage = "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')"

func upstreamRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-import-module-exports")
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
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

func importError(line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "", Message: errorMessage,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}}
}

// Every semantic case from eslint-plugin-import v2.32.0.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-import-module-exports.js
func TestNoImportModuleExportsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &no_import_module_exports.NoImportModuleExportsRule,
		[]rule_tester.ValidTestCase{
			{Code: "const thing = require('thing')\nmodule.exports = thing"},
			{Code: "import thing from 'otherthing'\nconsole.log(thing.module.exports)"},
			{Code: "import thing from 'other-thing'\nexport default thing"},
			{Code: "const thing = require('thing')\nexports.foo = bar"},
			{Code: "import { module } from 'qunit'\nmodule.skip('A test', function () {})"},
			{
				Code:     "import foo from 'path';\nmodule.exports = foo;",
				FileName: "index.js",
			},
			{
				Code:     "import foo from 'path';\nmodule.exports = foo;",
				FileName: "some/other/entry-point.js",
				Options:  []any{map[string]any{"exceptions": []any{"**/*/other/entry-point.js"}}},
			},
			{
				Code:     "import * as process from 'process';\nconsole.log(process.env);",
				FileName: "missing-entrypoint/cli.js",
			},
			{
				Code: `import fs from 'fs/promises';

const subscriptions = new Map();
export default async (client) => {
    const modules = await fs.readdir('./src/modules');
    await Promise.all(
        modules.map(async (moduleName) => {
            const module = await import(` + "`./modules/${moduleName}/module.js`" + `);
            if (module.enabled) {
                module.subscriptions.forEach((fun, event) => {
                    if (!subscriptions.has(event)) subscriptions.set(event, []);
                    subscriptions.get(event).push(fun);
                });
            }
        })
    );
    subscriptions.forEach((funs, event) => {
        client.on(event, (...args) => {
            funs.forEach(async (fun) => {
                try { await fun(client, ...args); } catch (e) { client.emit('error', e); }
            });
        });
    });
};`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "import { stuff } from 'starwars'\nmodule.exports = thing", Errors: importError(1, 1, 1, 33)},
			{Code: "import thing from 'starwars'\nconst baz = module.exports = thing\nconsole.log(baz)", Errors: importError(1, 1, 1, 29)},
			{Code: "import * as allThings from 'starwars'\nexports.bar = thing", Errors: importError(1, 1, 1, 38)},
			{Code: "import thing from 'other-thing'\nexports.foo = bar", Errors: importError(1, 1, 1, 32)},
			{
				Code:     "import foo from 'path';\nmodule.exports = foo;",
				FileName: "some/other/entry-point.js",
				Options:  []any{map[string]any{"exceptions": []any{"**/*/other/file.js"}}},
				Errors:   importError(1, 1, 1, 24),
			},
		},
	)
}

// All code examples from the pinned rule documentation.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-import-module-exports.md
func TestNoImportModuleExportsDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(t), "tsconfig.json", t, &no_import_module_exports.NoImportModuleExportsRule,
		[]rule_tester.ValidTestCase{
			{Code: "import thing from 'other-thing'\nexport default thing", FileName: "docs/source.js"},
			{Code: "const thing = require('thing')\nmodule.exports = thing", FileName: "docs/source.js"},
			{Code: "const thing = require('thing')\nexports.foo = bar", FileName: "docs/source.js"},
			{Code: "import thing from 'otherthing'\nconsole.log(thing.module.exports)", FileName: "docs/source.js"},
			{Code: "import foo from 'path';\nmodule.exports = foo;", FileName: "docs/lib/index.js"},
			{
				Code: "import foo from 'path';\nmodule.exports = foo;", FileName: "docs/some-file.js",
				Options: []any{map[string]any{"exceptions": []any{"**/*/some-file.js"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "import { stuff } from 'starwars'\nmodule.exports = thing", FileName: "docs/source.js", Errors: importError(1, 1, 1, 33)},
			{Code: "import * as allThings from 'starwars'\nexports.bar = thing", FileName: "docs/source.js", Errors: importError(1, 1, 1, 38)},
			{Code: "import thing from 'other-thing'\nexports.foo = bar", FileName: "docs/source.js", Errors: importError(1, 1, 1, 32)},
			{Code: "import thing from 'starwars'\nconst baz = module.exports = thing\nconsole.log(baz)", FileName: "docs/source.js", Errors: importError(1, 1, 1, 29)},
		},
	)
}
