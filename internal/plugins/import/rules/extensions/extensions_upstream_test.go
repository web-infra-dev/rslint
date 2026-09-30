package extensions_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/extensions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// cspell:ignore barjson barhbs exceljs barjs barnone rootverse
// Every semantic case from eslint-plugin-import v2.32.0. Parser variants use
// the native parser; the webpack-only case remains an explained skip.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/extensions.js
func TestExtensionsUpstream(t *testing.T) {
	t.Run("extensions", func(t *testing.T) {
		rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule,
			[]rule_tester.ValidTestCase{
				{Code: "import a from \"@/a\"", FileName: "files/input.ts"},
				{Code: "import a from \"a\"", FileName: "files/input.ts"},
				{Code: "import dot from \"./file.with.dot\"", FileName: "files/input.ts"},
				{Code: "import a from \"a/index.js\"", FileName: "files/input.ts", Options: []any{"always"}},
				{Code: "import dot from \"./file.with.dot.js\"", FileName: "files/input.ts", Options: []any{"always"}},
				{Code: "import a from \"a\"\nimport packageConfig from \"./package.json\"", FileName: "files/input.ts", Options: []any{map[string]any{"json": "always", "js": "never"}}},
				{Code: "import lib from \"./bar\"\nimport component from \"./bar.jsx\"\nimport data from \"./bar.json\"", FileName: "files/input.ts", Options: []any{"never"}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx", ".json"}}}},
				{Code: "import bar from \"./bar\"\nimport barjson from \"./bar.json\"\nimport barhbs from \"./bar.hbs\"", FileName: "files/input.ts", Options: []any{"always", map[string]any{"js": "never", "jsx": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx", ".json", ".hbs"}}}},
				{Code: "import bar from \"./bar.js\"\nimport pack from \"./package\"", FileName: "files/input.ts", Options: []any{"never", map[string]any{"js": "always", "json": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".json"}}}},
				{Code: "import path from \"path\"", FileName: "files/input.ts"},
				{Code: "import path from \"path\"", FileName: "files/input.ts", Options: []any{"never"}},
				{Code: "import path from \"path\"", FileName: "files/input.ts", Options: []any{"always"}},
				{Code: "import thing from \"./fake-file.js\"", FileName: "files/input.ts", Options: []any{"always"}},
				{Code: "import thing from \"non-package\"", FileName: "files/input.ts", Options: []any{"never"}},
				{Code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n        import express from 'express'\n      ", FileName: "files/input.ts", Options: []any{"ignorePackages"}},
				{Code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n        import express from 'express'\n      ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ignorePackages": true}}},
				{Code: "\n        import foo from './foo'\n        import bar from './bar'\n        import Component from './Component'\n        import express from 'express'\n      ", FileName: "files/input.ts", Options: []any{"never", map[string]any{"ignorePackages": true}}},
				{Code: "import exceljs from \"exceljs\"", FileName: "files/internal-modules/plugins/plugin.js", Options: []any{"always", map[string]any{"js": "never", "jsx": "never"}}, Settings: map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": []any{".js", ".jsx", ".json"}}, "webpack": map[string]any{"config": "webpack.empty.config.js"}}}},
				{Code: "export { foo } from \"./foo.js\"\nlet bar; export { bar }", FileName: "files/input.ts", Options: []any{"always"}},
				{Code: "export { foo } from \"./foo\"\nlet bar; export { bar }", FileName: "files/input.ts", Options: []any{"never"}},
				{Code: "import lib from \"pkg.js\"\nimport lib2 from \"pgk/package\"\nimport lib3 from \"@name/pkg.js\"", FileName: "files/input.ts", Options: []any{"never"}},
				{Code: "import bare from \"./foo?a=True.ext\"", FileName: "files/input.ts", Options: []any{"never"}},
				{Code: "import bare from \"./foo.js?a=True\"", FileName: "files/input.ts", Options: []any{"always"}},
				{Code: "import lib from \"pkg\"\nimport lib2 from \"pgk/package.js\"\nimport lib3 from \"@name/pkg\"", FileName: "files/input.ts", Options: []any{"always"}},
			}, []rule_tester.InvalidTestCase{
				{Code: "import a from \"a/index.js\"", FileName: "files/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"a/index.js\"", Line: 1, Column: 15, EndLine: 1, EndColumn: 27}}},
				{Code: "import dot from \"./file.with.dot\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension \"js\" for \"./file.with.dot\"", Line: 1, Column: 17, EndLine: 1, EndColumn: 34}}},
				{Code: "import a from \"a/index.js\"\nimport packageConfig from \"./package\"", FileName: "files/input.ts", Options: []any{map[string]any{"json": "always", "js": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".json"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"a/index.js\"", Line: 1, Column: 15, EndLine: 1, EndColumn: 27}, {MessageId: "", Message: "Missing file extension \"json\" for \"./package\"", Line: 2, Column: 27, EndLine: 2, EndColumn: 38}}},
				{Code: "import lib from \"./bar.js\"\nimport component from \"./bar.jsx\"\nimport data from \"./bar.json\"", FileName: "files/input.ts", Options: []any{"never"}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx", ".json"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./bar.js\"", Line: 1, Column: 17, EndLine: 1, EndColumn: 27}}},
				{Code: "import lib from \"./bar.js\"\nimport component from \"./bar.jsx\"\nimport data from \"./bar.json\"", FileName: "files/input.ts", Options: []any{map[string]any{"json": "always", "js": "never", "jsx": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx", ".json"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./bar.js\"", Line: 1, Column: 17, EndLine: 1, EndColumn: 27}}},
				{Code: "import component from \"./bar.jsx\"\nimport data from \"./bar.json\"", FileName: "files/input.ts", Options: []any{map[string]any{"json": "always", "js": "never", "jsx": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".jsx", ".json", ".js"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"jsx\" for \"./bar.jsx\"", Line: 1, Column: 23, EndLine: 1, EndColumn: 34}}},
				{Code: "import \"./bar.coffee\"", FileName: "files/input.ts", Options: []any{"never", map[string]any{"js": "always", "jsx": "always"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".coffee", ".js"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"coffee\" for \"./bar.coffee\"", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
				{Code: "import barjs from \"./bar.js\"\nimport barjson from \"./bar.json\"\nimport barnone from \"./bar\"", FileName: "files/input.ts", Options: []any{"always", map[string]any{"json": "always", "js": "never", "jsx": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx", ".json"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./bar.js\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
				{Code: "import barjs from \".\"\nimport barjs2 from \"..\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension \"js\" for \".\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 22}, {MessageId: "", Message: "Missing file extension \"js\" for \"..\"", Line: 2, Column: 20, EndLine: 2, EndColumn: 24}}},
				{Code: "import barjs from \"./bar.js\"\nimport barjson from \"./bar.json\"\nimport barnone from \"./bar\"", FileName: "files/input.ts", Options: []any{"never", map[string]any{"json": "always", "js": "never", "jsx": "never"}}, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx", ".json"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./bar.js\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
				{Code: "import thing from \"./fake-file.js\"", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./fake-file.js\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 35}}},
				{Code: "import thing from \"non-package/test\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"non-package/test\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 37}}},
				{Code: "import thing from \"@name/pkg/test\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"@name/pkg/test\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 35}}},
				{Code: "import thing from \"@name/pkg/test.js\"", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"@name/pkg/test.js\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 38}}},
				{Code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component'\n        import baz from 'foo/baz'\n        import baw from '@scoped/baw/import'\n        import chart from '@/configs/chart'\n        import express from 'express'\n      ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ignorePackages": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./Component\"", Line: 4, Column: 31, EndLine: 4, EndColumn: 44}, {MessageId: "", Message: "Missing file extension for \"@/configs/chart\"", Line: 7, Column: 27, EndLine: 7, EndColumn: 44}}},
				{Code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component'\n        import baz from 'foo/baz'\n        import baw from '@scoped/baw/import'\n        import chart from '@/configs/chart'\n        import express from 'express'\n      ", FileName: "files/input.ts", Options: []any{"ignorePackages"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./Component\"", Line: 4, Column: 31, EndLine: 4, EndColumn: 44}, {MessageId: "", Message: "Missing file extension for \"@/configs/chart\"", Line: 7, Column: 27, EndLine: 7, EndColumn: 44}}},
				{Code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n        import express from 'express'\n      ", FileName: "files/input.ts", Options: []any{"never", map[string]any{"ignorePackages": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./foo.js\"", Line: 2, Column: 25, EndLine: 2, EndColumn: 35}, {MessageId: "", Message: "Unexpected use of file extension \"jsx\" for \"./Component.jsx\"", Line: 4, Column: 31, EndLine: 4, EndColumn: 48}}},
				{Code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n      ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"pattern": map[string]any{"jsx": "never"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"jsx\" for \"./Component.jsx\"", Line: 4, Column: 31, EndLine: 4, EndColumn: 48}}},
				{Code: "export { foo } from \"./foo\"\nlet bar; export { bar }", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./foo\"", Line: 1, Column: 21, EndLine: 1, EndColumn: 28}}},
				{Code: "export { foo } from \"./foo.js\"\nlet bar; export { bar }", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./foo.js\"", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}}},
				{Code: "import withExtension from \"./foo.js?a=True\"", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./foo.js?a=True\"", Line: 1, Column: 27, EndLine: 1, EndColumn: 44}}},
				{Code: "import withoutExtension from \"./foo?a=True.ext\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./foo?a=True.ext\"", Line: 1, Column: 30, EndLine: 1, EndColumn: 48}}},
				{Code: "const { foo } = require(\"./foo\")\nexport { foo }", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./foo\"", Line: 1, Column: 25, EndLine: 1, EndColumn: 32}}},
				{Code: "const { foo } = require(\"./foo.js\")\nexport { foo }", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./foo.js\"", Line: 1, Column: 25, EndLine: 1, EndColumn: 35}}},
				{Code: "export { foo } from \"./foo\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./foo\"", Line: 1, Column: 21, EndLine: 1, EndColumn: 28}}},
				{Code: "\n        import foo from \"@/ImNotAScopedModule\";\n        import chart from '@/configs/chart';\n      ", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"@/ImNotAScopedModule\"", Line: 2, Column: 25, EndLine: 2, EndColumn: 47}, {MessageId: "", Message: "Missing file extension for \"@/configs/chart\"", Line: 3, Column: 27, EndLine: 3, EndColumn: 44}}},
				{Code: "export { foo } from \"./foo.js\"", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./foo.js\"", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}}},
				{Code: "export * from \"./foo\"", FileName: "files/input.ts", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./foo\"", Line: 1, Column: 15, EndLine: 1, EndColumn: 22}}},
				{Code: "export * from \"./foo.js\"", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"./foo.js\"", Line: 1, Column: 15, EndLine: 1, EndColumn: 25}}},
				{Code: "import foo from \"@/ImNotAScopedModule.js\"", FileName: "files/input.ts", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"@/ImNotAScopedModule.js\"", Line: 1, Column: 17, EndLine: 1, EndColumn: 42}}},
				{Code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module/index.js';\n\n        import bar from './bar';\n      ", FileName: "files/input.ts", Options: []any{"never"}, Settings: map[string]any{"import/resolver": "webpack", "import/external-module-folders": []any{"node_modules", "symlinked-module"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unexpected use of file extension \"js\" for \"@test-scope/some-module/index.js\"", Line: 3, Column: 23, EndLine: 3, EndColumn: 57}}, Skip: true}, // Webpack JavaScript resolver plugins cannot run in the native runtime.
				{Code: "import * as test from \".\"", FileName: "files/internal-modules/test.js", Options: []any{"ignorePackages"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \".\"", Line: 1, Column: 23, EndLine: 1, EndColumn: 26}}},
				{Code: "import * as test from \"..\"", FileName: "files/internal-modules/plugins/plugin.js", Options: []any{"ignorePackages"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"..\"", Line: 1, Column: 23, EndLine: 1, EndColumn: 27}}},
			})
	})
	t.Run("typescript: extensions ignore type-only", func(t *testing.T) {
		rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule,
			[]rule_tester.ValidTestCase{
				{Code: "import type T from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ts": "never", "tsx": "never", "js": "never", "jsx": "never"}}},
				{Code: "export type { MyType } from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ts": "never", "tsx": "never", "js": "never", "jsx": "never"}}},
			}, []rule_tester.InvalidTestCase{
				{Code: "import T from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ts": "never", "tsx": "never", "js": "never", "jsx": "never"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./typescript-declare\"", Line: 1, Column: 15, EndLine: 1, EndColumn: 37}}},
				{Code: "export { MyType } from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ts": "never", "tsx": "never", "js": "never", "jsx": "never"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./typescript-declare\"", Line: 1, Column: 24, EndLine: 1, EndColumn: 46}}},
				{Code: "import type T from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ts": "never", "tsx": "never", "js": "never", "jsx": "never", "checkTypeImports": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./typescript-declare\"", Line: 1, Column: 20, EndLine: 1, EndColumn: 42}}},
				{Code: "export type { MyType } from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ts": "never", "tsx": "never", "js": "never", "jsx": "never", "checkTypeImports": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"./typescript-declare\"", Line: 1, Column: 29, EndLine: 1, EndColumn: 51}}},
			})
	})
	t.Run("typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set", func(t *testing.T) {
		rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule,
			[]rule_tester.ValidTestCase{
				{Code: "import type { MyType } from \"./typescript-declare.ts\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"checkTypeImports": true}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}},
				{Code: "export type { MyType } from \"./typescript-declare.ts\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"checkTypeImports": true}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}},
				{Code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src.ts';\n              import { $exists } from 'rootverse+bfe:src/symbols.ts';\n\n              import type { Entries } from 'type-fest';\n            ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ignorePackages": true, "checkTypeImports": true, "pathGroupOverrides": []any{map[string]any{"pattern": "multiverse{*,*/**}", "action": "enforce"}}}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}},
				{Code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src.ts';\n              import { $exists } from 'rootverse+bfe:src/symbols.ts';\n\n              import type { Entries } from 'type-fest';\n            ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ignorePackages": true, "checkTypeImports": true, "pathGroupOverrides": []any{map[string]any{"pattern": "rootverse{*,*/**}", "action": "enforce"}}}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}},
				{Code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src';\n              import { $exists } from 'rootverse+bfe:src/symbols';\n\n              import type { Entries } from 'type-fest';\n            ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ignorePackages": true, "checkTypeImports": true, "pathGroupOverrides": []any{map[string]any{"pattern": "multiverse{*,*/**}", "action": "enforce"}, map[string]any{"pattern": "rootverse{*,*/**}", "action": "ignore"}}}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}},
			}, []rule_tester.InvalidTestCase{
				{Code: "import type { MyType } from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"checkTypeImports": true}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension \"ts\" for \"./typescript-declare\"", Line: 1, Column: 29, EndLine: 1, EndColumn: 51}}},
				{Code: "export type { MyType } from \"./typescript-declare\";", FileName: "files/input.ts", Options: []any{"always", map[string]any{"checkTypeImports": true}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension \"ts\" for \"./typescript-declare\"", Line: 1, Column: 29, EndLine: 1, EndColumn: 51}}},
				{Code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src';\n              import { $exists } from 'rootverse+bfe:src/symbols';\n\n              import type { Entries } from 'type-fest';\n            ", FileName: "files/input.ts", Options: []any{"always", map[string]any{"ignorePackages": true, "checkTypeImports": true, "pathGroupOverrides": []any{map[string]any{"pattern": "rootverse{*,*/**}", "action": "enforce"}, map[string]any{"pattern": "universe{*,*/**}", "action": "ignore"}}}}, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"alwaysTryTypes": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Missing file extension for \"rootverse+debug:src\"", Line: 4, Column: 42, EndLine: 4, EndColumn: 63}, {MessageId: "", Message: "Missing file extension for \"rootverse+bfe:src/symbols\"", Line: 5, Column: 39, EndLine: 5, EndColumn: 66}}},
			})
	})
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/extensions.md
// Examples without a specified filesystem are checked as unresolved imports.
// The two ignorePackages examples containing @/foo are errors in the pinned
// implementation (and its tests), despite being labelled valid in the docs.
func TestExtensionsDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `import bar from './foo/bar.json';`, Options: []any{"never"}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, group := range []struct {
		options        []any
		valid, invalid []string
	}{
		{[]any{"never"}, []string{"./foo", "./bar", "./Component", "express/index", "path"}, []string{"./foo.js", "./bar.json", "./Component.jsx", "express/index.js"}},
		{[]any{"always"}, []string{"./foo.js", "./bar.json", "./Component.jsx", "path", "@/foo.js"}, []string{"./foo", "./bar", "./Component", "@/foo"}},
		{[]any{"ignorePackages"}, []string{"./foo.js", "./bar.json", "./Component.jsx", "express"}, []string{"./foo", "./bar", "./Component", "@/foo"}},
		{[]any{"always", map[string]any{"ignorePackages": true}}, []string{"./Component.jsx", "foo/baz.js", "express"}, []string{"@/foo"}},
	} {
		for _, name := range group.valid {
			valid = append(valid, rule_tester.ValidTestCase{Code: "import value from '" + name + "';", Options: group.options})
		}
		for _, name := range group.invalid {
			code := "import value from '" + name + "';"
			message := "Missing file extension for \"" + name + "\""
			if group.options[0] == "never" {
				extension := name[strings.LastIndexByte(name, '.')+1:]
				message = "Unexpected use of file extension \"" + extension + "\" for \"" + name + "\""
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: group.options, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, "'"+name+"'", message)}})
		}
	}
	for _, mode := range []string{"always", "never"} {
		name, message := "./foo", `Missing file extension for "./foo"`
		if mode == "never" {
			name, message = "./foo.ts", `Unexpected use of file extension "ts" for "./foo.ts"`
		}
		for _, declaration := range []string{"import type { Foo }", "export type { Foo }"} {
			code := declaration + " from '" + name + "';"
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: []any{mode, map[string]any{"checkTypeImports": true}}, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, "'"+name+"'", message)}})
		}
	}
	root := extensionsRoot(t)
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "docs/foo/bar.js"):   "export default 1;",
		tspath.ResolvePath(root.Dir, "docs/foo/bar.json"): "{}",
	})
	for i := range valid {
		valid[i].FileName = "docs/input.ts"
	}
	for i := range invalid {
		invalid[i].FileName = "docs/input.ts"
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &extensions.ExtensionsRule, valid, invalid)
}
