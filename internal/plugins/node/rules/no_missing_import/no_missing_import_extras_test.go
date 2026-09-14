package no_missing_import

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMissingImportExtras(t *testing.T) {
	root := missingRoot(t, "testdata/extras.txtar")
	message := func(text string) string { return strings.ReplaceAll(text, "{{root}}", root.Dir) }
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoMissingImportRule,
		[]rule_tester.ValidTestCase{
			// self reference entry
			{Code: "import 'self';", FileName: "maps/input.js"},
			// imports #query
			{Code: "import '#query';", FileName: "maps/input.js"},
			// imports #fragment
			{Code: "import '#fragment';", FileName: "maps/input.js"},
			// imports #file-builtin
			{Code: "import '#file-builtin';", FileName: "maps/input.js"},
			// imports TS extension aliases
			{Code: "import '#file-alias';", FileName: "maps/input.ts"},
			{Code: "import '#files/entry';", FileName: "maps/input.js"},
			// export forms and import attributes
			{Code: "export {value} from './present.js'; export * from './present.js'; export * as ns from './present.js'; import './present.js' with { type: 'json' };", FileName: "src/input.js"},
			// expressions outside the upstream literal visitor
			{Code: "const path = 'missing'; import(path); import('missing' + '/part'); import(`missing`); require('missing'); obj?.import('missing');", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}},
			// authored TypeScript wrappers and import types
			{Code: "import('missing' as string); import('missing'!); type T = import('missing').T; import x = require('missing');", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}},
			// parenthesized dynamic import and escaped resource
			{Code: "import((('./present.js'))); import('./pr\\u0065sent.js?raw#part'); import('./present.js!loader?raw');", FileName: "src/input.js"},
			// Node replaces unpaired surrogates only when accessing the filesystem.
			{Code: `import './\ud800.js'; export * from './\udc00.js'; import('./\ud83d\ude00.js'); import 'unicode-target/\ud800';`, FileName: "src/input.js"},
			// default option objects
			{Code: "import 'pkg'; import '@scope/pkg';", FileName: "src/input.js", Options: []any{map[string]any{}}},
			// allow package roots and virtual modules
			{Code: "import 'allowed/subpath'; import '@allowed/pkg/subpath'; import 'virtual:module/entry';", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{"allowed", "@allowed/pkg", "virtual:module"}}}},
			// legacy settings take precedence
			{Code: "import 'allowed';", FileName: "src/input.js", Settings: map[string]any{"n": map[string]any{"allowModules": []any{"allowed"}}, "node": map[string]any{"allowModules": []any{}}}},
			// options override shared settings
			{Code: "import 'allowed';", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{"allowed"}}}, Settings: map[string]any{"node": map[string]any{"allowModules": []any{}}}},
			// explicit file ignores empty extension list
			{Code: "import './present.js'; import './entry.custom'; import './no-extension';", FileName: "src/input.js", Options: []any{map[string]any{"tryExtensions": []any{}}}},
			// custom extension and settings precedence
			{Code: "import './entry';", FileName: "src/input.js", Options: []any{map[string]any{"tryExtensions": []any{".custom"}}}, Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{}}}},
			// relative paths do not use module directories
			{Code: "import './present.js';", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}},
			// custom module directory string
			{Code: "import 'custom';", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": "custom_modules"}}}},
			// custom module directory shared list
			{Code: "import 'custom';", FileName: "src/input.js", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": []any{"custom_modules"}}}}},
			// ordered resolvePaths and cwd
			{Code: "import './remote.js';", FileName: "src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"missing", "other"}}}},
			// settings cwd affects additional paths
			{Code: "import './remote.js';", FileName: "src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"."}}}, Settings: map[string]any{"cwd": message("{{root}}/other")}},
			// extension aliases cover export and dynamic imports
			{Code: "export type {value} from './present.js'; export * from './present.js'; import('./present.js');", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}},
			// paths exact, wildcard and fallback targets
			{Code: "import '@exact'; import '@local/present.js'; import '@multi/present.js';", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}},
			// extension preset react
			{Code: "import \"./view.js\";", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react"}}, Settings: map[string]any{}},
			// extension preset react-jsx
			{Code: "import \"./view.js\";", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react-jsx"}}, Settings: map[string]any{}},
			// extension preset react-jsxdev
			{Code: "import \"./view.js\";", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react-jsxdev"}}, Settings: map[string]any{}},
			// extension preset react-native
			{Code: "import \"./view.js\";", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react-native"}}, Settings: map[string]any{}},
			// extension preset preserve
			{Code: "import \"./view.jsx\";", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "preserve"}}, Settings: map[string]any{}},
			// unconfigured explicit tsconfig falls through to settings
			{Code: "import './view.js';", FileName: "src/input.ts", Options: []any{map[string]any{"tsconfigPath": message("{{root}}/config/no-jsx.json")}}, Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react"}}},
			// mapping options override shared tsconfig
			{Code: "import './view.js';", FileName: "src/input.ts", Options: []any{map[string]any{"typescriptExtensionMap": "react"}}, Settings: map[string]any{"n": map[string]any{"tsconfigPath": message("{{root}}/config/explicit.json")}}},
			// allowImportingTsExtensions implicit defaults
			{Code: "import './present';", FileName: "direct/input.ts", Settings: map[string]any{}},
			// type-only package conditions
			{Code: "import type {Value} from 'types-only'; export type {Value} from 'types-only';", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}},
			// imports maps file, package, builtin
			{Code: "import '#entry'; import '#pkg'; import '#builtin';", FileName: "maps/input.js", Options: []any{}, Settings: map[string]any{}},
			// hash aliases from tsconfig
			{Code: "import '#local';", FileName: "maps/input.ts", Options: []any{}, Settings: map[string]any{}},
			// builtins and URL runtime imports
			{Code: "import '_http_agent'; import 'node:fs/promises'; import 'data:text/javascript,0'; import 'https://example.com/a';", FileName: "src/input.js"},
		},
		[]rule_tester.InvalidTestCase{
			// scalar conversion
			{Code: "import(1e999); import(0b1000_0000n); import(0); import(/foo/dimsuy);", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'Infinity' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 13}, {MessageId: "notFound", Message: message("Can't resolve '128' in '{{root}}/src'"), Line: 1, Column: 23, EndLine: 1, EndColumn: 35}, {MessageId: "notFound", Message: message("Can't resolve '0' in '{{root}}/src'"), Line: 1, Column: 45, EndLine: 1, EndColumn: 46}, {MessageId: "notFound", Message: message("Can't resolve '/foo/dimsuy' in '{{root}}/src'"), Line: 1, Column: 56, EndLine: 1, EndColumn: 67}}},
			// nested dynamic
			{Code: "import(import(\"missing\")); import((\"missing\"));", FileName: "src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 15, EndLine: 1, EndColumn: 24}, {MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 36, EndLine: 1, EndColumn: 45}}},
			// Documented difference: surrogates
			{Code: "import(\"\\ud800\"); import(\"\\udc00\"); import(\"\\u{1f600}\");", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '\xed\xa0\x80' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 16}, {MessageId: "notFound", Message: message("Can't resolve '\xed\xb0\x80' in '{{root}}/src'"), Line: 1, Column: 26, EndLine: 1, EndColumn: 34}, {MessageId: "notFound", Message: message("Can't resolve '😀' in '{{root}}/src'"), Line: 1, Column: 44, EndLine: 1, EndColumn: 55}}},
			// Normalizing a filesystem path must not enable directory fallback.
			{Code: `import './dir-\ud800';`, FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './dir-\xed\xa0\x80' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
			// self reference private and missing
			{Code: "import 'self/private'; import 'self/missing';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("\"./private\" is not exported under the conditions [\"node\",\"require\",\"import\"] from package {{root}}/maps (see exports field in {{root}}/maps/package.json)"), Line: 1, Column: 8, EndLine: 1, EndColumn: 22}, {MessageId: "notFound", Message: message("Package path ./missing is exported from package {{root}}/maps, but no valid target file was found (see exports field in {{root}}/maps/package.json)"), Line: 1, Column: 31, EndLine: 1, EndColumn: 45}}},
			// imports #null
			{Code: "import '#null';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Package import #null is not imported from package {{root}}/maps (see imports field in {{root}}/maps/package.json)"), Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
			// imports #types
			{Code: "import '#types';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Package import #types is not imported from package {{root}}/maps (see imports field in {{root}}/maps/package.json)"), Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// Documented difference: imports #array
			{Code: "import '#array';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '#array' in '{{root}}/maps'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// imports #array-fallback
			{Code: "import '#array-fallback';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '#array-fallback' in '{{root}}/maps'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// imports #pattern/
			{Code: "import '#pattern/';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: "Resolving to directories is not possible with the imports field (request was #pattern/)", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// Documented difference: imports #redirect
			{Code: "import '#redirect';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '#redirect' in '{{root}}/maps'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// imports #entry?raw
			{Code: "import '#entry?raw';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Package import #entry?raw is not imported from package {{root}}/maps (see imports field in {{root}}/maps/package.json)"), Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// imports #
			{Code: "import '#';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: "Request should have at least 2 characters", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}}},
			// inherited paths
			{Code: "import 'inherited/present.js';", FileName: "inherit/app/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'inherited/present.js' in '{{root}}/inherit/app'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 30}}},
			// baseUrl paths upstream behavior
			{Code: "import 'base/present.js';", FileName: "baseurl/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'base/present.js' in '{{root}}/baseurl'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// type and runtime conditions same resource
			{Code: "import type {Value} from 'types-only'; import 'types-only'; export type * from 'types-only';", FileName: "src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("\".\" is not exported under the conditions [\"node\",\"require\",\"import\"] from package {{root}}/node_modules/types-only (see exports field in {{root}}/node_modules/types-only/package.json)"), Line: 1, Column: 47, EndLine: 1, EndColumn: 59}}},
			// relative tsconfig option
			{Code: "import './alias-file.js';", FileName: "src/input.ts", Options: []any{map[string]any{"tsconfigPath": "./config/explicit.json"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './alias-file.js' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// Imports-map targets keep the import rule's no-directory behavior.
			{Code: "import '#directory';", FileName: "maps/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '#directory' in '{{root}}/maps'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// Upstream documents support for resolverConfig.modules only. Its
			// undocumented alias passthrough resolves this file; rslint ignores it.
			{Code: "import 'virtual';", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"virtual": "./present.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'virtual' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// empty allowed list overrides shared settings
			{Code: "import 'missing';", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{}}}, Settings: map[string]any{"node": map[string]any{"allowModules": []any{"missing"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// tryExtensions default and explicit empty list
			{Code: "import './present';", FileName: "src/input.js", Options: []any{map[string]any{"tryExtensions": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './present' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// empty module directories
			{Code: "import 'pkg';", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'pkg' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// last resolution failure uses source directory
			{Code: "import './missing';", FileName: "src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"other"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './missing' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// local directory main and index are rejected
			{Code: "import './directory'; import './package';", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './directory' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 21}, {MessageId: "notFound", Message: message("Can't resolve './package' in '{{root}}/src'"), Line: 1, Column: 30, EndLine: 1, EndColumn: 41}}},
			// missing alias target is still missing
			{Code: "import '@local/missing.js';", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '@local/missing.js' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
			// alias boundary must match a path segment
			{Code: "import '@exactly';", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '@exactly' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// empty extension map disables substitution
			{Code: "import './alias-file.js';", FileName: "src/input.ts", Options: []any{map[string]any{"typescriptExtensionMap": []any{}}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './alias-file.js' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// explicit TS config disables aliases
			{Code: "import './alias-file.js';", FileName: "src/input.ts", Options: []any{map[string]any{"tsconfigPath": message("{{root}}/config/explicit.json")}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve './alias-file.js' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// types field does not satisfy runtime lookup
			{Code: "import type {Value} from 'declaration-only';", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'declaration-only' in '{{root}}/src'"), Line: 1, Column: 26, EndLine: 1, EndColumn: 44}}},
			// ignoreTypeImport only skips whole import declarations
			{Code: "import type {Value} from 'missing'; export type {Value} from 'missing'; import {type Value as V} from 'missing';", FileName: "src/input.ts", Options: []any{map[string]any{"ignoreTypeImport": true}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 62, EndLine: 1, EndColumn: 71}, {MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 103, EndLine: 1, EndColumn: 112}}},
			// explicit false keeps type imports
			{Code: "import type {Value} from 'missing';", FileName: "src/input.ts", Options: []any{map[string]any{"ignoreTypeImport": false}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 26, EndLine: 1, EndColumn: 35}}},
			// type re-export all
			{Code: "export type * from 'missing';", FileName: "src/input.ts", Options: []any{map[string]any{"ignoreTypeImport": true}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 20, EndLine: 1, EndColumn: 29}}},
			// private and missing package exports
			{Code: "import 'pkg/private'; import 'pkg/unknown'; import 'pkg/missing';", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("\"./private\" is not exported under the conditions [\"node\",\"require\",\"import\"] from package {{root}}/node_modules/pkg (see exports field in {{root}}/node_modules/pkg/package.json)"), Line: 1, Column: 8, EndLine: 1, EndColumn: 21}, {MessageId: "notFound", Message: message("\"./unknown\" is not exported under the conditions [\"node\",\"require\",\"import\"] from package {{root}}/node_modules/pkg (see exports field in {{root}}/node_modules/pkg/package.json)"), Line: 1, Column: 30, EndLine: 1, EndColumn: 43}, {MessageId: "notFound", Message: message("Package path ./missing is exported from package {{root}}/node_modules/pkg, but no valid target file was found (see exports field in {{root}}/node_modules/pkg/package.json)"), Line: 1, Column: 52, EndLine: 1, EndColumn: 65}}},
			// missing imports map entry and target
			{Code: "import '#unknown'; import '#missing';", FileName: "maps/input.js", Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Package import #unknown is not imported from package {{root}}/maps (see imports field in {{root}}/maps/package.json)"), Line: 1, Column: 8, EndLine: 1, EndColumn: 18}, {MessageId: "notFound", Message: message("Can't resolve '#missing' in '{{root}}/maps'"), Line: 1, Column: 27, EndLine: 1, EndColumn: 37}}},
			// type-only URL is not a runtime URL exemption
			{Code: "import type X from 'https://example.com/a';", FileName: "src/input.ts", Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'https://example.com/a' in '{{root}}/src'"), Line: 1, Column: 20, EndLine: 1, EndColumn: 43}}},
			// dynamic scalar literals
			{Code: "import(null); import(false); import(true); import(1e3); import(0x10n); import(/a/mi);", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'null' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 12}, {MessageId: "notFound", Message: message("Can't resolve 'false' in '{{root}}/src'"), Line: 1, Column: 22, EndLine: 1, EndColumn: 27}, {MessageId: "notFound", Message: message("Can't resolve 'true' in '{{root}}/src'"), Line: 1, Column: 37, EndLine: 1, EndColumn: 41}, {MessageId: "notFound", Message: message("Can't resolve '1000' in '{{root}}/src'"), Line: 1, Column: 51, EndLine: 1, EndColumn: 54}, {MessageId: "notFound", Message: message("Can't resolve '16' in '{{root}}/src'"), Line: 1, Column: 64, EndLine: 1, EndColumn: 69}, {MessageId: "notFound", Message: message("Can't resolve '/a/im' in '{{root}}/src'"), Line: 1, Column: 79, EndLine: 1, EndColumn: 84}}},
			// non-ASCII and multiline ranges
			{Code: "const s = '😀'; import('缺少');\nexport {x}\nfrom 'missing';", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve '缺少' in '{{root}}/src'"), Line: 1, Column: 24, EndLine: 1, EndColumn: 28}, {MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 3, Column: 6, EndLine: 3, EndColumn: 15}}},
			// JSX import expression
			{Code: "const el = <div>{import('missing')}</div>;", FileName: "src/input.tsx", Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 25, EndLine: 1, EndColumn: 34}}},
			// duplicate imports retain individual diagnostics
			{Code: "import 'missing'; import('missing'); export * from 'missing';", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 17}, {MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 26, EndLine: 1, EndColumn: 35}, {MessageId: "notFound", Message: message("Can't resolve 'missing' in '{{root}}/src'"), Line: 1, Column: 52, EndLine: 1, EndColumn: 61}}},
			// invalid builtin name and file URL
			{Code: "import 'node:missing'; import 'file:///missing.js'; import ''; ", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: message("Can't resolve 'node:missing' in '{{root}}/src'"), Line: 1, Column: 8, EndLine: 1, EndColumn: 22}, {MessageId: "notFound", Message: message("Can't resolve 'file:///missing.js' in '{{root}}/src'"), Line: 1, Column: 31, EndLine: 1, EndColumn: 51}, {MessageId: "notFound", Message: message("Can't resolve '' in '{{root}}/src'"), Line: 1, Column: 60, EndLine: 1, EndColumn: 62}}},
		},
	)
}
