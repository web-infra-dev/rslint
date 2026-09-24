package no_relative_parent_imports_test

import (
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRelativeParentImportsExtras(t *testing.T) {
	root := parentRoot(t)
	tsFile := "internal-modules/plugins/plugin2/index.ts"
	typescript := map[string]any{"import/resolver": "typescript"}
	absoluteParent := tspath.ResolvePath(root.Dir, "internal-modules/plugins/plugin.js")
	runParentTests(t, []rule_tester.ValidTestCase{
		{Code: `import '../missing.js'; import ''; import 'node:fs';`},
		{Code: `import '../plugin2/internal.js';`}, // Normalizes to a sibling.
		{Code: `require('../plugin.js'); define(['../plugin.js'], () => {});`},
		{Code: `require('../plugin.js'); define(['../plugin.js'], () => {});`, Options: map[string]any{"commonjs": false, "amd": false}},
		{Code: `import '../plugin.js'; export * from '../plugin.js'; import('../plugin.js');`, Options: map[string]any{"esmodule": false}},
		{Code: `import '../plugin.js'; require('../plugin.js'); define(['../plugin.js'], () => {});`, Options: map[string]any{"commonjs": true, "amd": true, "ignore": []any{`(?<=\.\./)plugin\.js$`}}},
		{Code: `import '../plugin.js';`, Settings: map[string]any{"import/core-modules": []any{"../plugin.js"}}},
		{Code: `import './../plugin.js';`, Settings: map[string]any{"import/internal-regex": `^\.\.[\\/]`}},
		// Non-literal sources, members and TypeScript wrappers are not ESTree string literals.
		{Code: "require(); require('../plugin.js', 1); require(1); require(`../plugin.js`); require('../' + name); require(...['../plugin.js']);\nimport(`../plugin.js`); import(name);\nmodule.require('../plugin.js'); require.resolve('../plugin.js'); obj?.require('../plugin.js'); obj['require']('../plugin.js'); new require('../plugin.js');", Options: map[string]any{"commonjs": true}},
		{Code: `class C { #require() {} load() { this.#require('../plugin.js'); } }`, Options: map[string]any{"commonjs": true}},
		{FileName: tsFile, Code: `import alias = require('../plugin.js'); type T = import('../plugin.js').Value; (require as any)('../plugin.js'); require('../plugin.js' as string); require!('../plugin.js'); import('../plugin.js' as string);`, Options: map[string]any{"commonjs": true}},
		{Code: "/** @import { Value } from '../plugin.js' */\n/** @type {import('../plugin.js').Value} */\nlet value;"},
		{Code: `const value = 1; export { value }; export default value;`},
		// AMD requires exactly two arguments and a direct require/define callee.
		{Code: "define('name', ['../plugin.js'], () => {}); define(['../plugin.js']); define('../plugin.js', () => {}); obj.define(['../plugin.js'], () => {});\nrequire(['require', 'exports', , 1, null, `../plugin.js`, name, ...paths], () => {});", Options: map[string]any{"amd": true}},
		{FileName: tsFile, Code: `define(['../plugin.js'] as string[], () => {}); define(['../plugin.js' as string], () => {});`, Options: map[string]any{"amd": true}},
		{Code: `define(['require', 'exports'], () => {});`, Options: map[string]any{"amd": true}, Settings: typescript},
		// Resolver defaults do not infer a TypeScript extension.
		{FileName: tsFile, Code: `import type { Value } from '../types';`},
		{Code: `import '@outside';`, Settings: typescript},
		{Code: `import '@vendor';`, Settings: map[string]any{"import/resolver": "typescript", "import/external-module-folders": []any{"vendor"}}},
		{Code: `import 'package';`, Settings: map[string]any{"import/external-module-folders": []any{""}}},
		// Option ignores run before resolution, even for a broken resolver.
		{Code: `import '../plugin.js';`, Options: map[string]any{"ignore": []any{"plugin"}}, Settings: map[string]any{"import/resolver": "missing-resolver"}},
	}, []rule_tester.InvalidTestCase{
		{Code: `import '../plugin.js';`, Options: map[string]any{}, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 22)},
		{Code: `import '../plugin.js';`, Options: map[string]any{"commonjs": false, "amd": false, "esmodule": true}, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 22)},
		{Code: `export * from '../plugin.js';`, Errors: parentError("index.js", "../plugin.js", 1, 15, 1, 29)},
		{Code: `export { value } from '../plugin.js';`, Errors: parentError("index.js", "../plugin.js", 1, 23, 1, 37)},
		{Code: `export * as plugin from '../plugin.js';`, Errors: parentError("index.js", "../plugin.js", 1, 25, 1, 39)},
		{Code: `import('../plugin.js', { with: { type: 'json' } });`, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 22)},
		{Code: `import(('../plugin.js'));`, Errors: parentError("index.js", "../plugin.js", 1, 9, 1, 23)},
		{Code: `import value from '../plugin.js' with { type: 'json' };`, Errors: parentError("index.js", "../plugin.js", 1, 19, 1, 33)},
		{Code: `(require)(('../plugin.js'));`, Options: map[string]any{"commonjs": true, "esmodule": false}, Errors: parentError("index.js", "../plugin.js", 1, 12, 1, 26)},
		// JSDoc casts are comments in ESTree, unlike authored TypeScript assertions.
		{Code: `require(/** @type {string} */ ('../plugin.js'));`, Options: map[string]any{"commonjs": true}, Errors: parentError("index.js", "../plugin.js", 1, 32, 1, 46)},
		{Code: `(/** @type {any} */ (require))('../plugin.js');`, Options: map[string]any{"commonjs": true}, Errors: parentError("index.js", "../plugin.js", 1, 32, 1, 46)},
		{Code: `define(/** @type {string[]} */ (['../plugin.js']), () => {});`, Options: map[string]any{"amd": true}, Errors: parentError("index.js", "../plugin.js", 1, 34, 1, 48)},
		{Code: `require?.('../plugin.js');`, Options: map[string]any{"commonjs": true}, Errors: parentError("index.js", "../plugin.js", 1, 11, 1, 25)},
		{Code: `function load(require) { require('../plugin.js'); }`, Options: map[string]any{"commonjs": true}, Errors: parentError("index.js", "../plugin.js", 1, 34, 1, 48)},
		{Code: `define(['../plugin.js'], () => {});`, Options: map[string]any{"amd": true, "esmodule": false}, Errors: parentError("index.js", "../plugin.js", 1, 9, 1, 23)},
		{Code: `(require)(([('../plugin.js')]), () => {});`, Options: map[string]any{"amd": true}, Errors: parentError("index.js", "../plugin.js", 1, 14, 1, 28)},
		{Code: `import '../plugin.js';`, Options: map[string]any{"ignore": []any{"^package$"}}, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 22)},
		// import/ignore controls export inspection, not this rule's ignore option.
		{Code: `import '../plugin.js';`, Settings: map[string]any{"import/ignore": []any{"plugin"}}, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 22)},
		{FileName: tsFile, Code: `import type { Value } from '../types';`, Settings: typescript, Errors: parentError("index.ts", "../types", 1, 28, 1, 38)},
		{FileName: tsFile, Code: `export type { Value } from '../types';`, Settings: typescript, Errors: parentError("index.ts", "../types", 1, 28, 1, 38)},
		{FileName: tsFile, Code: `import { type Value } from '../types';`, Settings: typescript, Errors: parentError("index.ts", "../types", 1, 28, 1, 38)},
		{FileName: "internal-modules/plugins/plugin2/index.d.ts", Code: `declare module 'virtual' { export { value } from '../plugin.js'; }`, Errors: parentError("index.d.ts", "../plugin.js", 1, 50, 1, 64)},
		{FileName: tsFile, Code: `namespace C { export const value = import('../plugin.js'); }`, Errors: parentError("index.ts", "../plugin.js", 1, 43, 1, 57)},
		{FileName: "internal-modules/plugins/plugin2/index.tsx", Code: `const view = <Widget value={require('../plugin.js')} />;`, Options: map[string]any{"commonjs": true}, Errors: parentError("index.tsx", "../plugin.js", 1, 37, 1, 51)},
		{Code: "// 😀\nimport foo\n  from '../plugin.js';", Errors: parentError("index.js", "../plugin.js", 3, 8, 3, 22)},
		{Code: "const face = '😀'; import('../😀.js');", Errors: parentError("index.js", "../😀.js", 1, 27, 1, 37)},
		{Code: "import '../plu\\\ngin.js';", Errors: parentError("index.js", "../plugin.js", 1, 8, 2, 8)},
		{Code: `import '\u002e./plugin.js';`, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 27)},
		{Code: `import '@api/service';`, Settings: typescript, Errors: parentError("index.js", "@api/service", 1, 8, 1, 22)},
		{Code: `import '@outside';`, Settings: map[string]any{"import/resolver": "typescript", "import/internal-regex": "^@outside$"}, Errors: parentError("index.js", "@outside", 1, 8, 1, 18)},
		{Code: `import '@vendor';`, Settings: map[string]any{"import/resolver": "typescript", "import/external-module-folders": []any{}}, Errors: parentError("index.js", "@vendor", 1, 8, 1, 17)},
		{Code: `import '../../node_modules/package/index.js';`, Errors: parentError("index.js", "../../node_modules/package/index.js", 1, 8, 1, 45)},
		{Code: `import '../../../outside.js';`, Errors: parentError("index.js", "../../../outside.js", 1, 8, 1, 29)},
		{Code: "import " + strconv.Quote(absoluteParent) + ";", Errors: parentError("index.js", absoluteParent, 1, 8, 1, 10+len(absoluteParent))},
		{Code: "import '../plugin.js';\nimport '../../api/service';", Settings: map[string]any{"import/resolver": "missing-resolver"}, Errors: []rule_tester.InvalidTestCaseError{{Message: `Resolve error: unable to load resolver "missing-resolver".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1}}},
	})
}
