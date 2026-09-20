// cspell:ignore requ
package no_missing_require

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Checked against eslint-plugin-n v18.3.0, ESLint 10.2.1 and
// @typescript-eslint/parser 8.57.0, except the documented resolver differences.
func TestNoMissingRequireExtras(t *testing.T) {
	root := missingRequireRoot(t, "testdata/extras.txtar")
	valid := []rule_tester.ValidTestCase{
		// no require identifiers
		{Code: "const answer = 42;", FileName: "src/input.js"},
		// local and written roots
		{Code: "function f(require) { require('missing'); } { const require = other; require.resolve('missing'); } try {} catch(require) { require('missing'); }", FileName: "src/input.js"},
		// hoisted require binding
		{Code: "require('missing'); var require;", FileName: "src/input.js"},
		// written global root
		{Code: "require('missing'); require = other;", FileName: "src/input.js", Globals: map[string]any{"require": "writable"}},
		// untracked calls and identifiers
		{Code: "obj.require('missing'); require.cache('missing'); new require('missing'); require.call(null, 'missing'); require(...['missing']); const name = 'missing'; require(name); const key = 'resolve'; require[key]('missing');", FileName: "src/input.js"},
		// TypeScript declarations do not call require
		{Code: "import value = require('missing'); type T = import('missing').T;", FileName: "src/input.ts"},
		// Private access on require is invalid in ESLint; tsgo still supplies the AST.
		{Code: "class Box { #resolve; run() { require.#resolve('missing'); } }", FileName: "src/input.ts"},
		// builtins ignore resolver aliases
		{Code: "require('fs'); require('node:fs/promises'); require('_http_agent'); require('fs!loader');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"fs": "./absent.js", "node:fs/promises": "./absent.js", "_http_agent": "./absent.js"}}}}},
		// allowed roots and explicit defaults
		{Code: "require('allowed/sub'); require('@scope/allowed/sub'); require('virtual:module/entry'); require('allowed/sub');", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{"allowed", "@scope/allowed", "virtual:module"}, "resolvePaths": []any{}, "tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}}}},
		// shared allow list
		{Code: "require('allowed');", FileName: "src/input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{"allowed"}}}},
		// legacy settings precedence
		{Code: "require('allowed');", FileName: "src/input.js", Settings: map[string]any{"n": map[string]any{"allowModules": []any{"allowed"}}, "node": map[string]any{"allowModules": []any{}}}},
		// directory index, package main fallback and repeated targets
		{Code: "require('./dir'); require('./dir/'); require('pkg'); require('pkg'); require('self-package');", FileName: "src/input.js", Options: []any{map[string]any{}}},
		// loader parameters, queries and fragments
		{Code: "require('./present.js!loader'); require.resolve('./present.js?raw#part'); require('./pr\\u0065sent.js');", FileName: "src/input.js"},
		// TypeScript aliases resolve actual targets
		{Code: "require('@local/present.js'); require('./present.js');", FileName: "src/input.ts"},
		// TypeScript extensions are explicitly enabled
		{Code: "require('./present'); require('./present.ts');", FileName: "direct/input.ts"},
		// empty module directories still allow local and package imports
		{Code: "require('./present.js'); require('#entry'); require('#builtin'); require('self-package');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}},
		// custom module directory
		{Code: "require('custom');", FileName: "src/input.js", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": "custom_modules"}}}},
		// custom extensions override shared settings
		{Code: "require('./custom');", FileName: "src/input.js", Options: []any{map[string]any{"tryExtensions": []any{".ext"}}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js"}}}},
		// explicit filenames with empty extension search
		{Code: "require('./present.js'); require('./custom.ext');", FileName: "src/input.js", Options: []any{map[string]any{"tryExtensions": []any{}}}},
		// ordered extra lookup directories
		{Code: "require('./remote.js');", FileName: "src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"absent", "vendor"}}}},
		// configured cwd selects extra lookup directories
		{Code: "require('./remote.js');", FileName: "src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"."}}}, Settings: map[string]any{"cwd": tspath.ResolvePath(root.Dir, "vendor")}},
		// alias fallback and disabled alias
		{Code: "require('alias'); require('disabled');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"alias": []any{"./absent.js", "./present.js"}, "disabled": false}}}}},
		{Code: "require('virtual');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fallback": map[string]any{"virtual": "./present.js"}}}}},
		// explicit aliases override TypeScript paths
		{Code: "require('@broken/present');", FileName: "src/input.ts", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"@broken/present": "./present.js"}}}}},
		// custom export conditions
		{Code: "require('import-only');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"conditionNames": []any{"import"}}}}},
		// package entry fields
		{Code: "require('./dir');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"mainFields": []any{"browser"}}}}},
		// directory entry names
		{Code: "require('./entry');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"mainFiles": []any{"api"}}}}},
		// package alias field
		{Code: "require('alias-field');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"aliasFields": []any{"browser"}}}}},
		// explicit extension aliases
		{Code: "require('./custom.js');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"extensions": []any{}, "extensionAlias": map[string]any{".js": []any{".ext"}}}}}},
		// unpaired surrogates use Node filesystem replacement
		{Code: "require('./\\ud800.js');", FileName: "src/input.js"},
		// extension mapping react-jsx
		{Code: "require('./view.js');", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react-jsx"}}},
		// extension mapping react-jsxdev
		{Code: "require('./view.js');", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react-jsxdev"}}},
		// extension mapping react-native
		{Code: "require('./view.js');", FileName: "src/input.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react-native"}}},
		// shadowed and reassigned global objects
		{Code: "function f(globalThis) { globalThis.require('missing'); } global.require('missing'); global = other;", FileName: "src/input.js", Globals: map[string]any{"global": "writable"}},
		// ambient require is a local declaration
		{Code: "declare function require(name: string): unknown; require('missing');", FileName: "src/input.ts"},
		// locally created require is not the global API
		{Code: "import {createRequire} from 'node:module'; const require = createRequire(import.meta.url); require('missing');", FileName: "src/input.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// JSX member tags do not invoke require
		{Code: "const view = <require.resolve />;", FileName: "src/input.tsx"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// aliases, optional calls and computed property names
		{Code: "const load = require; const {resolve: lookup} = require; load('missing'); lookup?.('missing'); require?.['resolve']?.('missing'); globalThis['require']('missing');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 63, EndLine: 1, EndColumn: 72}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 84, EndLine: 1, EndColumn: 93}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 119, EndLine: 1, EndColumn: 128}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 153, EndLine: 1, EndColumn: 162}}},
		// parenthesized and escaped calls
		{Code: "(requ\\u0069re)((('missing'))); (require.resolve)('missing');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 18, EndLine: 1, EndColumn: 27}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 50, EndLine: 1, EndColumn: 59}}},
		// TypeScript argument and callee wrappers
		{Code: "(require as any)('missing'); require('missing' as string); require('missing'!); require('missing' satisfies string); require(<string>'missing');", FileName: "src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 18, EndLine: 1, EndColumn: 27}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 38, EndLine: 1, EndColumn: 57}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 68, EndLine: 1, EndColumn: 78}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 89, EndLine: 1, EndColumn: 115}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 126, EndLine: 1, EndColumn: 143}}},
		// JSX and class expressions
		{Code: "class Box { value = require('missing'); static { require.resolve('missing'); } } const el = <Box value={require('missing')}/>;", FileName: "src/input.tsx", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 29, EndLine: 1, EndColumn: 38}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 66, EndLine: 1, EndColumn: 75}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 113, EndLine: 1, EndColumn: 122}}},
		// constant expressions and scalar conversions
		{Code: "require('mis' + 'sing'); require(`mis${'sing'}`); require(true ? 'missing' : 'fs'); require(40n + 2n); require(null); require(false); require(/missing/g);", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 23}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 34, EndLine: 1, EndColumn: 48}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 59, EndLine: 1, EndColumn: 82}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '42' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 93, EndLine: 1, EndColumn: 101}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'null' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 112, EndLine: 1, EndColumn: 116}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'false' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 127, EndLine: 1, EndColumn: 132}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '/missing/g' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 143, EndLine: 1, EndColumn: 153}}},
		// multiline Unicode ranges
		{Code: "const label = '💡'; require('缺少');\r\nrequire(\r\n 'mis' +\r\n 'sing'\r\n);", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '缺少' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 29, EndLine: 1, EndColumn: 33}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 3, Column: 2, EndLine: 4, EndColumn: 8}}},
		// duplicate alias paths and repeated missing targets
		{Code: "const a = require; const b = require; const load = cond ? a : b; load('missing'); require('missing');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 71, EndLine: 1, EndColumn: 80}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 71, EndLine: 1, EndColumn: 80}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 91, EndLine: 1, EndColumn: 100}}},
		// nested calls, extra arguments and source order
		{Code: "const load = require; require(load('missing')); require('missing', {}); load('other-missing');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 36, EndLine: 1, EndColumn: 45}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 57, EndLine: 1, EndColumn: 66}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'other-missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 78, EndLine: 1, EndColumn: 93}}},
		// package exports distinguish require, import and types
		{Code: "require('import-only'); require('types-only'); require('pkg/private'); require('pkg/missing'); require('self-package/missing');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("\".\" is not exported under the conditions [\"node\",\"require\"] from package /__root__/node_modules/import-only (see exports field in /__root__/node_modules/import-only/package.json)", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 22}, {MessageId: "notFound", Message: strings.ReplaceAll("\".\" is not exported under the conditions [\"node\",\"require\"] from package /__root__/node_modules/types-only (see exports field in /__root__/node_modules/types-only/package.json)", "/__root__", root.Dir), Line: 1, Column: 33, EndLine: 1, EndColumn: 45}, {MessageId: "notFound", Message: strings.ReplaceAll("\"./private\" is not exported under the conditions [\"node\",\"require\"] from package /__root__/node_modules/pkg (see exports field in /__root__/node_modules/pkg/package.json)", "/__root__", root.Dir), Line: 1, Column: 56, EndLine: 1, EndColumn: 69}, {MessageId: "notFound", Message: strings.ReplaceAll("Package path ./missing is exported from package /__root__/node_modules/pkg, but no valid target file was found (see exports field in /__root__/node_modules/pkg/package.json)", "/__root__", root.Dir), Line: 1, Column: 80, EndLine: 1, EndColumn: 93}, {MessageId: "notFound", Message: strings.ReplaceAll("Package path ./missing is exported from package /__root__, but no valid target file was found (see exports field in /__root__/package.json)", "/__root__", root.Dir), Line: 1, Column: 104, EndLine: 1, EndColumn: 126}}},
		// unknown and missing imports mappings
		{Code: "require('#unknown'); require('#absent');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Package import #unknown is not imported from package /__root__ (see imports field in /__root__/package.json)", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 19}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '#absent' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 30, EndLine: 1, EndColumn: 39}}},
		// missing paths aliases must resolve
		{Code: "require('@broken/entry');", FileName: "src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '@broken/entry' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 24}}},
		// require URLs and invalid builtins are checked
		{Code: "require('https://example.com/a'); require('file:///missing.js'); require('node:missing'); require('');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'https://example.com/a' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 32}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'file:///missing.js' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 43, EndLine: 1, EndColumn: 63}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'node:missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 74, EndLine: 1, EndColumn: 88}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 99, EndLine: 1, EndColumn: 101}}},
		// declared but uninstalled dependencies and type stubs
		{Code: "require('missing'); require('absent');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'absent' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 29, EndLine: 1, EndColumn: 37}}},
		// empty allow list overrides shared settings
		{Code: "require('allowed');", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{}}}, Settings: map[string]any{"node": map[string]any{"allowModules": []any{"allowed"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'allowed' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// empty extension list overrides shared settings
		{Code: "require('./present');", FileName: "src/input.js", Options: []any{map[string]any{"tryExtensions": []any{}}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './present' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// empty lookup paths override shared settings
		{Code: "require('./remote.js');", FileName: "src/input.js", Options: []any{map[string]any{"resolvePaths": []any{}}}, Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"vendor"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './remote.js' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 22}}},
		// empty module directories disable package lookup
		{Code: "require('pkg');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'pkg' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// empty TypeScript extension mapping
		{Code: "require('./view.js');", FileName: "src/input.ts", Options: []any{map[string]any{"typescriptExtensionMap": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './view.js' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// empty resolver alias overrides tsconfig aliases
		{Code: "require('@local/present.js');", FileName: "src/input.ts", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '@local/present.js' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 28}}},
		// empty entry settings disable directory fallback
		{Code: "require('./dir');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"mainFiles": []any{}, "mainFields": []any{}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './dir' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// documented recursive alias diagnostic omits the upstream resolver stack
		{Code: "require('loop');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"loop": "other", "other": "loop"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: "Recursive alias while resolving 'loop'", Line: 1, Column: 9, EndLine: 1, EndColumn: 15}}},
		// disabled wildcard aliases leave unrelated requests checked
		{Code: "require('unrelated');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"pkg/*": false}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'unrelated' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// documented invalid imports target diagnostic
		{Code: "require('#invalid');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '#invalid' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 19}}},
		// documented object alias priority uses sorted keys
		{Code: "require('@app/special');", FileName: "src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"@app/special": "./present.js", "@app": "./absent"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '@app/special' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 23}}},
		// body declarations do not shadow parameter initializers
		{Code: "function load(value = require('missing')) { var require = other; }", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 31, EndLine: 1, EndColumn: 40}}},
		// computed global require and method aliases
		{Code: "const {['requ' + 'ire']: load} = globalThis; load('missing'); self['requ' + 'ire'].resolve('missing');", FileName: "src/input.js", Globals: map[string]any{"self": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 51, EndLine: 1, EndColumn: 60}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 92, EndLine: 1, EndColumn: 101}}},
		// JSDoc casts preserve argument ranges
		{Code: "require(/** @type {string} */ ('missing'));", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 32, EndLine: 1, EndColumn: 41}}},
		// empty allow list and loader parameters preserve repeated reports
		{Code: "require('missing!first'); require('missing!second'); require('node:fs!loader');", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 24}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 35, EndLine: 1, EndColumn: 51}}},
		// allowed package caching cannot exempt local paths or other packages
		{Code: "require('allowed!one'); require('allowed!two'); require('allowed/sub'); require('./allowed'); require('allowed-other');", FileName: "src/input.js", Options: []any{map[string]any{"allowModules": []any{"allowed"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './allowed' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 81, EndLine: 1, EndColumn: 92}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'allowed-other' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 103, EndLine: 1, EndColumn: 118}}},
		// null bytes in constant arguments are reported
		{Code: "require('bad\\0path');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'bad\u0000path' in '/__root__/src'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// line and block disables preserve the report after reenable (RuleTester registers the rule as test)
		{Code: "/* eslint-disable test */\nrequire('missing');\n/* eslint-enable test */\n// eslint-disable-next-line test\nrequire('missing');\nrequire('missing');", FileName: "src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'missing' in '/__root__/src'", "/__root__", root.Dir), Line: 6, Column: 9, EndLine: 6, EndColumn: 18}}},
	}
	for i := range valid {
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoMissingRequireRule, valid, invalid)
}
