package no_extraneous_require

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// CommonJS calls and aliases checked against eslint-plugin-n v18.3.0,
// ESLint 10.2.1, eslint-utils 4.10.1 and typescript-eslint/parser 8.57.0.
func TestNoExtraneousRequireExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// upstream cycle guard also applies when the tracked value changes
		{Code: "/* global holder:writable */ let load = globalThis; load = load.require; load('runtime'); holder = globalThis; holder = holder.require; holder('runtime');", FileName: "input.js"},
		// untracked global alias
		{Code: "load = require; load('runtime');", FileName: "input.js"},
		// write in destructuring invalidates root
		{Code: "({require} = other); require('runtime');", FileName: "input.js"},
		// untracked literal forms
		{Code: "require(/runtime/); require(''); require(...['runtime']);", FileName: "input.js"},
		// builtins, paths and unknown modules
		{Code: "require('node:fs'); require('fs'); require('_http_agent'); require('./local'); require('/absolute'); require('#internal'); require('https://example.com/a'); require('data:text/javascript,0'); require('missing');", FileName: "input.js"},
		// all dependency fields and self
		{Code: "require('declared/sub'); require('dev'); require('peer'); require('optional'); require('app/sub');", FileName: "input.js"},
		// untracked syntactic forms
		{Code: "$.require('runtime'); require.cache('runtime'); new require('runtime'); require.call(null,'runtime'); import('runtime'); require(); require(...['runtime']);", FileName: "input.js"},
		// constant names do not resolve identifiers
		{Code: "const name = 'runtime'; require(name); const method = 'resolve'; require[method]('runtime');", FileName: "input.js"},
		// shadowed bindings
		{Code: "function f(require) { require('runtime'); } { const require = factory(); require.resolve('runtime'); } try {} catch(require) { require('runtime'); }", FileName: "input.js"},
		// hoisted binding
		{Code: "require('runtime'); var require;", FileName: "input.js"},
		// written global root
		{Code: "require('runtime'); require = factory;", FileName: "input.js", Globals: map[string]any{"require": "writable"}},
		// update invalidates root
		{Code: "require++; require('runtime');", FileName: "input.js", Globals: map[string]any{"require": "writable"}},
		// disabled require global
		{Code: "require('runtime');", FileName: "input.js", Globals: map[string]any{"require": "off"}},
		// shadowed global object
		{Code: "function f(global) { global.require('runtime'); }", FileName: "input.js"},
		// written global object
		{Code: "global.require('runtime'); global = other;", FileName: "input.js", Globals: map[string]any{"global": "writable"}},
		// no scope for globals in constant evaluation
		{Code: "require(undefined); require(String('runtime')); require(String.raw`runtime`);", FileName: "input.js"},
		// untracked rest bindings
		{Code: "const {...copy} = require; copy.resolve('runtime'); const [load] = [require]; load('runtime');", FileName: "input.js"},
		// TS import equals and type imports are ignored
		{Code: "import value = require('runtime'); type Value = import('runtime').Value;", FileName: "input.ts"},
		// type-only global prevents tracking
		{Code: "type require = {}; require('runtime');", FileName: "input.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// private keys ignored
		{Code: "class C { #resolve; run() { require.#resolve('runtime'); } }", FileName: "input.ts"},
		// allow scoped subdirectories
		{Code: "require('@scope/runtime/sub');", FileName: "input.js", Options: []any{map[string]any{"allowModules": []any{"@scope/runtime"}}}},
		// allow setting
		{Code: "require('runtime');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{"runtime"}}}},
		// legacy settings precedence
		{Code: "require('runtime');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{}}, "n": map[string]any{"allowModules": []any{"runtime"}}}},
		// empty resolvePaths overrides setting
		{Code: "require('external');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"extra"}}}, Options: []any{map[string]any{"resolvePaths": []any{}}}},
		// empty extensions overrides setting
		{Code: "require('runtime');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js"}}}, Options: []any{map[string]any{"tryExtensions": []any{}}}},
		// no declaration or implicit TS fallback
		{Code: "require('type-only'); require('ts-only');", FileName: "input.js"},
		// TS alias exemption
		{Code: "require('runtime/sub');", FileName: "paths/input.ts"},
		// empty modules disables lookup
		{Code: "require('runtime');", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}},
		// import-only and broken require condition
		{Code: "require('import-only'); require('blocked');", FileName: "input.js"},
		// malformed nearest package ignored
		{Code: "require('runtime');", FileName: "broken/input.js"},
		// workspace object permits dev dependency
		{Code: "require('workspace-dep');", FileName: "workspace-positive/packages/app/input.js"},
		// documented resolver alias difference
		{Code: "require('virtual');", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"virtual": "./local.js"}}}}},
		// documented workspace range difference
		{Code: "require('workspace-dep');", FileName: "workspace-range/packages/2/input.js"},
		// documented compound BigInt difference
		{Code: "require(40n + 2n);", FileName: "input.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// escaped require
		{Code: "r\\u0065quire('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 14, EndLine: 1, EndColumn: 23}}},
		// computed require through escaped global
		{Code: "glo\\u0062alThis[''+'require']('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 31, EndLine: 1, EndColumn: 40}}},
		// computed global binding
		{Code: "const {[''+'require']: load} = globalThis; load('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 49, EndLine: 1, EndColumn: 58}}},
		// computed property through global alias
		{Code: "const root = window; root[''+'require']('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 41, EndLine: 1, EndColumn: 50}}},
		// literal coercion in TS wrappers
		{Code: "require(0x2an as bigint); require(<bigint>0b101010n); require((0o52n satisfies bigint)); require(4_2);", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 24}, {MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 35, EndLine: 1, EndColumn: 52}, {MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 64, EndLine: 1, EndColumn: 86}, {MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 98, EndLine: 1, EndColumn: 101}}},
		// numeric precision and BigInt bases
		{Code: "require(9007199254740993); require(0x20000000000001); require(9_007_199_254_740_993); require(0x10000000000000000n);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"9007199254740992\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 25}, {MessageId: "extraneous", Message: "\"9007199254740992\" is extraneous.", Line: 1, Column: 36, EndLine: 1, EndColumn: 52}, {MessageId: "extraneous", Message: "\"9007199254740992\" is extraneous.", Line: 1, Column: 63, EndLine: 1, EndColumn: 84}, {MessageId: "extraneous", Message: "\"18446744073709551616\" is extraneous.", Line: 1, Column: 95, EndLine: 1, EndColumn: 115}}},
		// global property assignment keeps tracking
		{Code: "globalThis.require = other; globalThis.require('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 48, EndLine: 1, EndColumn: 57}}},
		// logical assignment alias
		{Code: "let load; load ||= require; load('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 34, EndLine: 1, EndColumn: 43}}},
		// nested default binding
		{Code: "function f({x = require} = {}) { x('runtime'); }", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 36, EndLine: 1, EndColumn: 45}}},
		// direct and resolved subdirectories
		{Code: "require('runtime'); require.resolve('@scope/runtime/sub');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "extraneous", Message: "\"@scope/runtime\" is extraneous.", Line: 1, Column: 37, EndLine: 1, EndColumn: 57}}},
		// optional computed calls
		{Code: "require?.('runtime'); require?.resolve?.('runtime'); require['re'+'solve']('runtime'); (require?.resolve)('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 42, EndLine: 1, EndColumn: 51}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 76, EndLine: 1, EndColumn: 85}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 107, EndLine: 1, EndColumn: 116}}},
		// parentheses and sequence
		{Code: "((require))((('runtime'))); (0, require)('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 42, EndLine: 1, EndColumn: 51}}},
		// constant expression targets
		{Code: "require(`runtime`); require('run'+'time'); require(`run${'time'}`); require(true ? 'runtime' : 'missing'); require(['runtime'][0]); require(({name:'runtime'}).name);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 29, EndLine: 1, EndColumn: 41}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 52, EndLine: 1, EndColumn: 66}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 77, EndLine: 1, EndColumn: 105}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 116, EndLine: 1, EndColumn: 130}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 141, EndLine: 1, EndColumn: 164}}},
		// primitive coercions
		{Code: "require(42); require(42n); require(false); require(null); require(true); require(void 0);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11}, {MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25}, {MessageId: "extraneous", Message: "\"false\" is extraneous.", Line: 1, Column: 36, EndLine: 1, EndColumn: 41}, {MessageId: "extraneous", Message: "\"null\" is extraneous.", Line: 1, Column: 52, EndLine: 1, EndColumn: 56}, {MessageId: "extraneous", Message: "\"true\" is extraneous.", Line: 1, Column: 67, EndLine: 1, EndColumn: 71}, {MessageId: "extraneous", Message: "\"undefined\" is extraneous.", Line: 1, Column: 82, EndLine: 1, EndColumn: 88}}},
		// aliases and overwritten aliases
		{Code: "const load = require; load('runtime'); let again = load; again = factory; again.resolve('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 89, EndLine: 1, EndColumn: 98}}},
		// assignments and cycles
		{Code: "let load, again; load = require; again = load; load = again; again('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 68, EndLine: 1, EndColumn: 77}}},
		// object binding
		{Code: "const {resolve: locate} = require; locate('runtime'); const {require:{resolve}} = global; resolve('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 43, EndLine: 1, EndColumn: 52}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 99, EndLine: 1, EndColumn: 108}}},
		// object assignment and defaults
		{Code: "let load, locate; ({require:load} = global); ({resolve:locate} = load); locate('runtime'); function f(read = require) { read('runtime'); }", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 80, EndLine: 1, EndColumn: 89}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 126, EndLine: 1, EndColumn: 135}}},
		// shorthand assignment
		{Code: "let resolve; ({resolve} = require); resolve('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 45, EndLine: 1, EndColumn: 54}}},
		// binding default
		{Code: "const {load = require} = obj; load('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 36, EndLine: 1, EndColumn: 45}}},
		// assignment default
		{Code: "let load; ({load = require} = obj); load('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 42, EndLine: 1, EndColumn: 51}}},
		// global objects and alias
		{Code: "global.require('runtime'); globalThis.require.resolve('runtime'); self.require('runtime'); window['require']('runtime'); const root = global; root.require('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 55, EndLine: 1, EndColumn: 64}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 80, EndLine: 1, EndColumn: 89}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 110, EndLine: 1, EndColumn: 119}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 156, EndLine: 1, EndColumn: 165}}},
		// global object independent from require setting
		{Code: "global.require('runtime'); require('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25}}, Globals: map[string]any{"require": "off"}},
		// logical and conditional aliases preserve duplicates
		{Code: "const load = flag ? require : require; load('runtime'); const fallback = other || require; fallback('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 45, EndLine: 1, EndColumn: 54}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 45, EndLine: 1, EndColumn: 54}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 101, EndLine: 1, EndColumn: 110}}},
		// extra argument and loader parameters
		{Code: "require('runtime!loader', ignored); require.resolve('runtime?raw');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 25}, {MessageId: "extraneous", Message: "\"runtime?raw\" is extraneous.", Line: 1, Column: 53, EndLine: 1, EndColumn: 66}}},
		// Unicode and multiline ranges
		{Code: "const text = '😀'; require('runtime');\nrequire(/*位置*/\n ('runt\\u0069me')\n);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 3, Column: 3, EndLine: 3, EndColumn: 17}}},
		// JSX expression
		{Code: "const view = <Panel value={require('runtime')}/>;", FileName: "input.tsx", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 36, EndLine: 1, EndColumn: 45}}},
		// TS wrappers and generic call
		{Code: "(require as any)('runtime' as string); require!('runtime'); require<string>('runtime'); (require satisfies Function)('runtime');", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 18, EndLine: 1, EndColumn: 37}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 49, EndLine: 1, EndColumn: 58}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 77, EndLine: 1, EndColumn: 86}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 118, EndLine: 1, EndColumn: 127}}},
		// type-only local shadows in module
		{Code: "export {}; type require = {}; require('runtime');", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 39, EndLine: 1, EndColumn: 48}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// JS JSDoc casts
		{Code: "/** @typedef {any} require */\n(/** @type {any} */ (require))('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 2, Column: 32, EndLine: 2, EndColumn: 41}}},
		// function default before body binding
		{Code: "function f(value = require('runtime')) { var require; }", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37}}},
		// require inside class static block
		{Code: "class C { static { require('runtime'); } }", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37}}},
		// empty options and explicit defaults
		{Code: "require('runtime');", FileName: "input.js", Options: []any{map[string]any{"allowModules": []any{}, "resolvePaths": []any{}, "tryExtensions": []any{".js", ".json", ".node", ".mjs", ".cjs"}, "resolverConfig": map[string]any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// empty option overrides allow setting
		{Code: "require('runtime');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{"runtime"}}}, Options: []any{map[string]any{"allowModules": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// allow matches exact package
		{Code: "require('runtime');", FileName: "input.js", Options: []any{map[string]any{"allowModules": []any{"run"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// resolvePaths option
		{Code: "require('external');", FileName: "input.js", Options: []any{map[string]any{"resolvePaths": []any{"extra"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"external\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}}},
		// resolvePaths setting
		{Code: "require('external');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"extra"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"external\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}}},
		// extension option
		{Code: "require('custom');", FileName: "input.js", Options: []any{map[string]any{"tryExtensions": []any{".custom"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"custom\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
		// extension setting
		{Code: "require('custom');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".custom"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"custom\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
		// explicit entry extension
		{Code: "require('entry');", FileName: "input.js", Options: []any{map[string]any{"tryExtensions": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"entry\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// all default runtime extensions
		{Code: "require('only-mjs'); require('only-cjs'); require('only-json'); require('only-node');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"only-mjs\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}, {MessageId: "extraneous", Message: "\"only-cjs\" is extraneous.", Line: 1, Column: 30, EndLine: 1, EndColumn: 40}, {MessageId: "extraneous", Message: "\"only-json\" is extraneous.", Line: 1, Column: 51, EndLine: 1, EndColumn: 62}, {MessageId: "extraneous", Message: "\"only-node\" is extraneous.", Line: 1, Column: 73, EndLine: 1, EndColumn: 84}}},
		// TS explicit extension alias
		{Code: "require('ts-only/index.js');", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"ts-only\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 27}}},
		// TS opt-in extensions
		{Code: "require('ts-only');", FileName: "explicit/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"ts-only\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// module directory array
		{Code: "require('bower');", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{"bower_components"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bower\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// module directory scalar
		{Code: "require('bower');", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": "bower_components"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bower\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// module directory shared
		{Code: "require('bower');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": "bower_components"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bower\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// empty resolver option overrides shared
		{Code: "require('runtime');", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}, Options: []any{map[string]any{"resolverConfig": map[string]any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// require conditions
		{Code: "require('require-only'); require('conditional');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"require-only\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 23}, {MessageId: "extraneous", Message: "\"conditional\" is extraneous.", Line: 1, Column: 34, EndLine: 1, EndColumn: 47}}},
		// convertPath object does not affect check
		{Code: "require('runtime');", FileName: "input.js", Options: []any{map[string]any{"convertPath": map[string]any{"**/*.js": []any{"^(.*)$", "missing/$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// convertPath array does not affect check
		{Code: "require('runtime');", FileName: "input.js", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"**/*.js"}, "exclude": []any{"vendor/**"}, "replace": []any{"^(.*)$", "missing/$1"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// documented workspace caret difference
		{Code: "require('workspace-dep');", FileName: "workspace-caret/packages/app/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"workspace-dep\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 24}}},
		// CommonJS file extension
		{Code: "require('runtime');", FileName: "input.cjs", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// CommonJS TypeScript extension
		{Code: "require('runtime');", FileName: "input.cts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// module source with Node globals
		{Code: "require('runtime');", FileName: "input.mjs", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// types dependency cannot supply runtime
		{Code: "require('typed');", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"typed\" is extraneous.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// inline require global
		{Code: "/* global require:readonly */ require('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 39, EndLine: 1, EndColumn: 48}}, Globals: map[string]any{"require": "off"}},
	}
	globals := func(overrides map[string]any) map[string]any {
		values := map[string]any{"require": "readonly", "global": "readonly", "globalThis": "readonly", "self": "readonly", "window": "readonly"}
		for key, value := range overrides {
			values[key] = value
		}
		return values
	}
	for i := range valid {
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "commonjs"
		}
		valid[i].Globals = globals(valid[i].Globals)
	}
	for i := range invalid {
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "commonjs"
		}
		invalid[i].Globals = globals(invalid[i].Globals)
	}
	rule_tester.RunRuleTester(extraneousRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &NoExtraneousRequireRule, valid, invalid)
}

func TestNoExtraneousRequireWithoutPackage(t *testing.T) {
	base := fixtures.GetRootDir()
	root := rule_tester.Root{Dir: base.Dir, FS: utils.NewOverlayVFS(base.FS, map[string]string{
		tspath.ResolvePath(base.Dir, "node_modules/runtime/index.js"): "module.exports = {};",
	})}
	rule_tester.RunRuleTester(root, "tsconfig.allowJs.json", t, &NoExtraneousRequireRule,
		[]rule_tester.ValidTestCase{{Code: "require('runtime');", FileName: "package-less.js", Globals: map[string]any{"require": "readonly"}}}, nil)
}
