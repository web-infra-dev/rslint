// cspell:ignore requ lobal
package no_restricted_require

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func restrictedRequireRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-restricted-require")
	archive := txtarfs.MustParseFile(t, "testdata/resolution.txtar")
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
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

func runRestrictedRequireTests(t *testing.T, root rule_tester.Root, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		}
		if valid[i].Globals == nil {
			valid[i].Globals = map[string]any{"require": "readonly", "global": "readonly"}
		}
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.js"
		}
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
		}
		if invalid[i].Globals == nil {
			invalid[i].Globals = map[string]any{"require": "readonly", "global": "readonly"}
		}
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoRestrictedRequireRule, valid, invalid)
}

// Expectations verified against eslint-plugin-n v18.3.0 with the TypeScript parser.
func TestNoRestrictedRequireExtras(t *testing.T) {
	root := restrictedRequireRoot(t)
	runRestrictedRequireTests(t, root, []rule_tester.ValidTestCase{
		// empty restrictions
		{Code: "require('fs'); require.resolve('fs');", Options: []any{[]any{}}},
		// empty and negative-only groups
		{Code: "require('fs');", Options: []any{[]any{map[string]any{"name": []any{}}, map[string]any{"name": []any{"!fs", "!!fs"}}, map[string]any{"name": []any{"path", "!fs"}}}}},
		// ordered exclusion
		{Code: "require('foo/bar');", Options: []any{[]any{map[string]any{"name": []any{"foo/*", "!foo/bar"}}}}},
		// exact builtin and package names
		{Code: "require('node:fs'); require('foo/bar');", Options: []any{[]any{"fs", "foo"}}},
		// literal glob punctuation
		{Code: "require('foo/a'); require('foo/b');", Options: []any{[]any{"foo/?", "foo/[ab]", "foo/{a,b}"}}},
		// query remains part of the module name
		{Code: "require('fs?raw'); require('fs#part');", Options: []any{[]any{"fs"}}},
		// jsx names and types are not calls
		{Code: "type T = typeof require; type M = import('fs'); const node = <require.resolve />;", Options: []any{[]any{"fs"}}, FileName: "input.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// declarations shadow require
		{Code: "function run(require) { require('fs'); } { const require = other; require.resolve('fs'); }", Options: []any{[]any{"fs"}}},
		// hoisted require shadows before declaration
		{Code: "require('fs'); var require;", Options: []any{[]any{"fs"}}, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// assignment disables global tracking
		{Code: "require('fs'); require = other;", Options: []any{[]any{"fs"}}},
		// global disabled in module mode
		{Code: "require('fs'); global.require('fs');", Options: []any{[]any{"fs"}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"require": "off", "global": "off"}},
		// non-call and unsupported arguments
		{Code: "new require('fs'); require(); require(...['fs']); require(name); const name2='fs'; require(name2); obj.require('fs');", Options: []any{[]any{"fs"}}},
		// imports do not count as require calls
		{Code: "import fs from 'fs'; export * from 'fs'; import('fs'); import m = require('fs');", Options: []any{[]any{"fs"}}, FileName: "input.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// import condition excluded
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/esm.js")}}},
		// unresolved packages and blocked exports have no path
		{Code: "require('missing'); require('pkg/hidden'); require('fs'); require('https://example.com/a');", Options: []any{[]any{filepath.Join(root.Dir, "**")}}},
		// module names and absolute exclusions
		{Code: "require('./server');", Options: []any{[]any{map[string]any{"name": []any{"./server", "!" + filepath.Join(root.Dir, "server/index.js")}}}}},
		// absolute match and name exclusion
		{Code: "require('./server');", Options: []any{[]any{map[string]any{"name": []any{filepath.Join(root.Dir, "server/**"), "!./server"}}}}},
		// empty extensions preserve lexical path
		{Code: "require('./server/api');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{}}}},
		// scope-dependent computed key
		{Code: "const method = 'require'; globalThis[method]('fs');", Options: []any{[]any{"fs"}}},
		// private property is not resolve
		{Code: "class Box { #resolve; check() { require.#resolve('fs'); } }", Options: []any{[]any{"fs"}}},
		// TypeScript ambient function declaration
		{Code: "declare function require(name: string): unknown; require('fs');", Options: []any{[]any{"fs"}}, FileName: "input.ts"},
		// upstream also treats a TypeScript type-only declaration as shadowing
		{Code: "type require = string; require('fs');", Options: []any{[]any{"fs"}}, FileName: "input.ts"},
		// line terminators in names stay literal
		{Code: "require('fs\\n'); require('fs\\r'); require('fs\\u2028');", Options: []any{[]any{"fs"}}},
		// file without require references
		{Code: "export const value = 1;", Options: []any{[]any{"fs"}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// BigInt mixed types and invalid operations
		{Code: "require(`${40n + 2}`); require(`${1n / 0n}`); require(`${2n ** -1n}`); require(`${42n >>> 1n}`); require('fs'.charAt(0n));", Options: []any{[]any{"*"}}},
		// alias removes original restriction
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/index.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"pkg": filepath.Join(root.Dir, "server/api.js")}}}}},
		// disabled alias removes original restriction
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/index.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"pkg": false}}}}},
		// exact alias does not match subpath
		{Code: "require('exact/api');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"exact$": filepath.Join(root.Dir, "server")}}}}},
		// empty alias overrides TypeScript paths
		{Code: "require('alias/api');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{}}}}},
		// explicit alias overrides TypeScript paths
		{Code: "require('alias/api');", Options: []any{[]any{filepath.Join(root.Dir, "server/index.js")}}, FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"alias/api": filepath.Join(root.Dir, "server/index.js")}}}}},
		// empty extension aliases override TypeScript defaults
		{Code: "require('./server/mapped.js');", Options: []any{[]any{filepath.Join(root.Dir, "server/mapped.ts")}}, FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"extensionAlias": map[string]any{}}}}},
		// require condition no longer matches
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/index.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"conditionNames": []any{"import"}}}}},
		// empty conditions disable conditional entries
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/index.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"conditionNames": []any{}}}}},
	}, []rule_tester.InvalidTestCase{
		// later positive restores a match
		{Code: "require('foo/bar');", Options: []any{[]any{map[string]any{"name": []any{"foo/*", "!foo/bar", "foo/bar"}, "message": "Use public API."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used. Use public API.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// first matching restriction and independent groups
		{Code: "require('fs');", Options: []any{[]any{map[string]any{"name": []any{"*", "!fs"}}, map[string]any{"name": "fs", "message": "first"}, map[string]any{"name": "fs", "message": "second"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used. first", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}}},
		// prefixes are explicitly restricted
		{Code: "require('node:fs'); require('_http_agent');", Options: []any{[]any{"node:fs", "_http_agent"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'node:fs' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "restricted", Message: "'_http_agent' module is restricted from being used.", Line: 1, Column: 29, EndLine: 1, EndColumn: 42}}},
		// literal extended groups and unicode glob
		{Code: "require('(fs)'); require('foo/雪');", Options: []any{[]any{"(fs)", "foo/*"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'(fs)' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15}, {MessageId: "restricted", Message: "'foo/雪' module is restricted from being used.", Line: 1, Column: 26, EndLine: 1, EndColumn: 33}}},
		// loader stripping keeps query and fragment
		{Code: "require('fs!loader'); require('foo?raw#part!loader'); require('!loader');", Options: []any{[]any{"fs", "foo?raw#part", ""}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}, {MessageId: "restricted", Message: "'foo?raw#part' module is restricted from being used.", Line: 1, Column: 31, EndLine: 1, EndColumn: 52}, {MessageId: "restricted", Message: "'' module is restricted from being used.", Line: 1, Column: 63, EndLine: 1, EndColumn: 72}}},
		// multiline custom message and unicode columns
		{Code: "const emoji = '😀'; require('文件');\nrequire('文' +\n  '件');", Options: []any{[]any{map[string]any{"name": "文件", "message": "换用 public API.\nPlease migrate."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'文件' module is restricted from being used. 换用 public API.\nPlease migrate.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33}, {MessageId: "restricted", Message: "'文件' module is restricted from being used. 换用 public API.\nPlease migrate.", Line: 2, Column: 9, EndLine: 3, EndColumn: 6}}},
		// resolve methods and aliases
		{Code: "require.resolve('fs');\nconst load = require; load('fs');\nconst {resolve: locate} = require; locate('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 2, Column: 28, EndLine: 2, EndColumn: 32}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 3, Column: 43, EndLine: 3, EndColumn: 47}}},
		// computed properties and global object aliases
		{Code: "global['requ' + 'ire']('fs');\nconst root = globalThis; root.require['resolve']('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 2, Column: 50, EndLine: 2, EndColumn: 54}}},
		// optional calls and properties
		{Code: "require?.('fs'); require?.resolve?.('fs'); (require?.resolve)('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 37, EndLine: 1, EndColumn: 41}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 63, EndLine: 1, EndColumn: 67}}},
		// parentheses and JSDoc casts
		{Code: "(require)(('fs')); require(/** @type {string} */ ('fs'));", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 51, EndLine: 1, EndColumn: 55}}},
		// TypeScript argument and callee wrappers
		{Code: "(require as any)('fs'); require('fs' as string); require('fs'!); require('fs' satisfies string); require(<string>'fs');", Options: []any{[]any{"fs"}}, FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 18, EndLine: 1, EndColumn: 22}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 33, EndLine: 1, EndColumn: 47}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 58, EndLine: 1, EndColumn: 63}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 74, EndLine: 1, EndColumn: 95}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 106, EndLine: 1, EndColumn: 118}}},
		// JSX expression call
		{Code: "const node = <Box value={require('fs')} />;", Options: []any{[]any{"fs"}}, FileName: "input.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 34, EndLine: 1, EndColumn: 38}}},
		// ESM with explicit Node globals
		{Code: "export const fs = require('fs');", Options: []any{[]any{"fs"}}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}}},
		// string expression folding
		{Code: "require('f' + 's'); require(`f${'s'}`); require(true ? 'fs' : 'path'); require(['f','s'].join(''));", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 49, EndLine: 1, EndColumn: 69}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 80, EndLine: 1, EndColumn: 98}}},
		// JavaScript scalar conversion
		{Code: "require(2); require(true); require(null); require(/fs/g); require(42n);", Options: []any{[]any{"2", "true", "null", "/fs/g", "42"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'2' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10}, {MessageId: "restricted", Message: "'true' module is restricted from being used.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25}, {MessageId: "restricted", Message: "'null' module is restricted from being used.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40}, {MessageId: "restricted", Message: "'/fs/g' module is restricted from being used.", Line: 1, Column: 51, EndLine: 1, EndColumn: 56}, {MessageId: "restricted", Message: "'42' module is restricted from being used.", Line: 1, Column: 67, EndLine: 1, EndColumn: 70}}},
		// extra arguments do not change target
		{Code: "require('fs', {flag: true});", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}}},
		// independent alias paths retain upstream duplicates
		{Code: "const a = require; const b = require; const load = cond ? a : b; load('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 71, EndLine: 1, EndColumn: 75}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 71, EndLine: 1, EndColumn: 75}}},
		// nested calls and source order
		{Code: "const load = require; require(load('fs')); load('path'); require('fs');", Options: []any{[]any{"fs", "path"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40}, {MessageId: "restricted", Message: "'path' module is restricted from being used.", Line: 1, Column: 49, EndLine: 1, EndColumn: 55}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 66, EndLine: 1, EndColumn: 70}}},
		// commonjs directory index and package main
		{Code: "require('./server'); require('./server/main');", Options: []any{[]any{filepath.Join(root.Dir, "server/index.js"), filepath.Join(root.Dir, "server/main/entry.js")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}, {MessageId: "restricted", Message: "'./server/main' module is restricted from being used.", Line: 1, Column: 30, EndLine: 1, EndColumn: 45}}},
		// require export condition
		{Code: "require('pkg'); require.resolve('pkg/sub');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/index.js"), filepath.Join(root.Dir, "node_modules/pkg/private/*")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}, {MessageId: "restricted", Message: "'pkg/sub' module is restricted from being used.", Line: 1, Column: 33, EndLine: 1, EndColumn: 42}}},
		// missing local target uses lexical path
		{Code: "require('./server/missing'); require.resolve('./server/../missing');", Options: []any{[]any{filepath.Join(root.Dir, "server/missing"), filepath.Join(root.Dir, "missing")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/missing' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 27}, {MessageId: "restricted", Message: "'./server/../missing' module is restricted from being used.", Line: 1, Column: 46, EndLine: 1, EndColumn: 67}}},
		// failed path then name match
		{Code: "require('missing');", Options: []any{[]any{map[string]any{"name": []any{filepath.Join(root.Dir, "**"), "missing"}, "message": ""}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'missing' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// package imports map
		{Code: "require('#entry'); require('#directory');", Options: []any{[]any{filepath.Join(root.Dir, "server/**")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'#entry' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}, {MessageId: "restricted", Message: "'#directory' module is restricted from being used.", Line: 1, Column: 28, EndLine: 1, EndColumn: 40}}},
		// query and fragment are retained on resolved paths
		{Code: "require('./server/api.js?raw#part');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js?raw#part")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/api.js?raw#part' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 35}}},
		// TypeScript aliases and emitted extension
		{Code: "require('alias/api.js'); require('./server/mapped.js');", Options: []any{[]any{filepath.Join(root.Dir, "server/**")}}, FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/mapped.js' module is restricted from being used.", Line: 1, Column: 34, EndLine: 1, EndColumn: 54}}},
		// custom module directory
		{Code: "require('custom');", Options: []any{[]any{filepath.Join(root.Dir, "custom_modules/**")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": []any{"custom_modules"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'custom' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
		// additional lookup path
		{Code: "require('extra');", Options: []any{[]any{filepath.Join(root.Dir, "vendor/**")}}, Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{filepath.Join(root.Dir, "vendor")}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'extra' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// custom extension lookup
		{Code: "require('./server/custom');", Options: []any{[]any{filepath.Join(root.Dir, "server/custom.ext")}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ext"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/custom' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26}}},
		// legacy settings take priority
		{Code: "require('./server/custom');", Options: []any{[]any{filepath.Join(root.Dir, "server/custom.ext")}}, Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".ext"}}, "node": map[string]any{"tryExtensions": []any{".js"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/custom' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26}}},
		// resolver alias restriction
		{Code: "require('virtual');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"virtual": filepath.Join(root.Dir, "server/api.js")}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'virtual' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// compound BigInt conversion
		{Code: "require(40n + 2n);", Options: []any{[]any{"42"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'42' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
		// escaped global identifier
		{Code: "requ\\u0069re('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18}}},
		// computed require on configured global objects
		{Code: "window['require']('fs'); self['require']('fs');", Options: []any{[]any{"fs"}}, Globals: map[string]any{"window": "readonly", "self": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 19, EndLine: 1, EndColumn: 23}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 42, EndLine: 1, EndColumn: 46}}},
		// global object destructuring
		{Code: "const { require: load } = globalThis; load('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 44, EndLine: 1, EndColumn: 48}}},
		// sequence calls and indirect invocation
		{Code: "(0, require)('fs'); require.call(null, 'fs'); require.apply(null, ['fs']);", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18}}},
		// nested class field and static block
		{Code: "class Box { value = require('fs'); static { require.resolve('fs'); } }", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 61, EndLine: 1, EndColumn: 65}}},
		// destructuring defaults and assignment aliases
		{Code: "const { resolve = fallback } = require; resolve('fs'); let load; load = require; load('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 49, EndLine: 1, EndColumn: 53}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 87, EndLine: 1, EndColumn: 91}}},
		// non-ASCII module and CRLF columns
		{Code: "const label = '💡'; require('💡');\r\nrequire(\r\n 'f' +\r\n 's'\r\n);", Options: []any{[]any{"💡", "fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'💡' module is restricted from being used.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 3, Column: 2, EndLine: 4, EndColumn: 5}}},
		// JavaScript numeric rounding and negative zero
		{Code: "require(0x1000000000000000); require(-0); require(1e21);", Options: []any{[]any{"1152921504606847000", "0", "1e+21"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'1152921504606847000' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 27}, {MessageId: "restricted", Message: "'0' module is restricted from being used.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40}, {MessageId: "restricted", Message: "'1e+21' module is restricted from being used.", Line: 1, Column: 51, EndLine: 1, EndColumn: 55}}},
		// successful TypeScript path alias
		{Code: "require('alias/mapped.js');", Options: []any{[]any{filepath.Join(root.Dir, "server/mapped.ts")}}, FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'alias/mapped.js' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26}}},
		// trailing directory separator
		{Code: "require('./server/');", Options: []any{[]any{filepath.Join(root.Dir, "server/index.js")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// escaped global object identifier
		{Code: "\\u0067lobalThis['require']('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 28, EndLine: 1, EndColumn: 32}}},
		// disabled line and scoped directives
		{Code: "// eslint-disable-next-line test\nrequire('fs');\n/* eslint-disable test */\nrequire.resolve('fs');\n/* eslint-enable test */\nrequire('fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 6, Column: 9, EndLine: 6, EndColumn: 13}}},
		// BigInt strings and branch selection
		{Code: "require('pkg' + (40n + 2n)); require(`${42n}`); require(0n ? 'path' : 'fs'); require(1n === 1n ? 'fs' : 'path');", Options: []any{[]any{"pkg42", "42", "fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg42' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 27}, {MessageId: "restricted", Message: "'42' module is restricted from being used.", Line: 1, Column: 38, EndLine: 1, EndColumn: 46}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 57, EndLine: 1, EndColumn: 75}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 86, EndLine: 1, EndColumn: 111}}},
		// BigInt signs division remainder and shifts
		{Code: "require('' + (-43n / 2n)); require('' + (-43n % 2n)); require('' + (42n << -1n));", Options: []any{[]any{"-21", "-1", "21"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'-21' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 25}, {MessageId: "restricted", Message: "'-1' module is restricted from being used.", Line: 1, Column: 36, EndLine: 1, EndColumn: 52}, {MessageId: "restricted", Message: "'21' module is restricted from being used.", Line: 1, Column: 63, EndLine: 1, EndColumn: 80}}},
		// BigInt exact mixed comparisons
		{Code: "require(9007199254740993n > 9007199254740992 ? 'fs' : 'path'); require(1n === 1 ? 'path' : 'fs');", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 61}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 72, EndLine: 1, EndColumn: 96}}},
		// alias replaces installed module
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"pkg": filepath.Join(root.Dir, "server/api.js")}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// aliases exact prefix wildcard and fallback
		{Code: "require('exact'); require('prefix/api'); require('wild-api'); require('fallback');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"exact$": filepath.Join(root.Dir, "server/api.js"), "prefix": filepath.Join(root.Dir, "server"), "wild-*": filepath.Join(root.Dir, "server/*.js"), "fallback": []any{"./absent", filepath.Join(root.Dir, "server/api.js")}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'exact' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}, {MessageId: "restricted", Message: "'prefix/api' module is restricted from being used.", Line: 1, Column: 27, EndLine: 1, EndColumn: 39}, {MessageId: "restricted", Message: "'wild-api' module is restricted from being used.", Line: 1, Column: 50, EndLine: 1, EndColumn: 60}, {MessageId: "restricted", Message: "'fallback' module is restricted from being used.", Line: 1, Column: 71, EndLine: 1, EndColumn: 81}}},
		// alias chains and cycles
		{Code: "require('a'); require('cycle');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"a": "b", "b": filepath.Join(root.Dir, "server/api.js"), "cycle": "other", "other": "cycle"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'a' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// ordered overlapping aliases
		{Code: "require('virtual/api');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": []any{map[string]any{"name": "virtual/api", "onlyModule": true, "alias": filepath.Join(root.Dir, "server/api.js")}, map[string]any{"name": "virtual", "alias": "missing"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'virtual/api' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22}}},
		// extension override replaces shared extensions
		{Code: "require('./server/custom');", Options: []any{[]any{filepath.Join(root.Dir, "server/custom.ext")}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".js"}, "resolverConfig": map[string]any{"extensions": []any{".ext"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/custom' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26}}},
		// extension alias in JavaScript
		{Code: "require('./server/mapped.js');", Options: []any{[]any{filepath.Join(root.Dir, "server/mapped.ts")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"extensionAlias": map[string]any{".js": []any{".ts"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./server/mapped.js' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29}}},
		// explicit conditions replace require
		{Code: "require('pkg');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/esm.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"conditionNames": []any{"import"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
	})
}

// Upstream also ignores the unrelated package root for this wildcard alias.
// A disabled subpath must not bypass an absolute restriction on that root.
func TestNoRestrictedRequireResolverEdges(t *testing.T) {
	root := restrictedRequireRoot(t)
	runRestrictedRequireTests(t, root, []rule_tester.ValidTestCase{
		// Documented limits: upstream reports each of these calls.
		{Code: "require((1n << 65536n) ? 'fs' : 'path');", Options: []any{[]any{"fs"}}},
		{Code: "require('./server');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"mainFiles": []any{"api"}}}}},
		// The equivalent upstream object puts virtual/api before virtual.
		// Use an alias array when declaration order must be preserved.
		{Code: "require('virtual/api');", Options: []any{[]any{filepath.Join(root.Dir, "server/api.js")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"virtual/api": filepath.Join(root.Dir, "server/api.js"), "virtual": "missing"}}}}},
	}, []rule_tester.InvalidTestCase{
		{Code: "require('pkg'); require('pkg/sub');", Options: []any{[]any{filepath.Join(root.Dir, "node_modules/pkg/**")}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"pkg/*": false}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
	})
}

func TestNoRestrictedRequireSchema(t *testing.T) {
	for _, tc := range []struct {
		options []any
		valid   bool
	}{
		{nil, true},
		{[]any{[]any{}}, true},
		{[]any{[]any{"fs", map[string]any{"name": []any{}, "message": ""}}}, true},
		{[]any{"fs"}, false},
		{[]any{[]any{"fs"}, []any{}}, false},
		{[]any{[]any{false}}, false},
		{[]any{[]any{map[string]any{}}}, false},
		{[]any{[]any{map[string]any{"name": []any{42}}}}, false},
		{[]any{[]any{map[string]any{"name": "fs", "message": false}}}, false},
		{[]any{[]any{map[string]any{"name": "fs", "extra": true}}}, false},
	} {
		err := NoRestrictedRequireRule.Schema.Validate(tc.options)
		if (err == nil) != tc.valid {
			t.Errorf("options %#v: valid = %v, error = %v", tc.options, tc.valid, err)
		}
	}
}
