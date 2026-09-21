// cspell:ignore requ
package no_unpublished_require

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Compared with eslint-plugin-n v18.3.0, ESLint 10.2.1 and
// @typescript-eslint/parser 8.70.0, except the documented differences and
// native-only AST or missing-metadata cases marked below.
func TestNoUnpublishedRequireExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// No require references.
		{Code: "const answer = 42;", FileName: "public/src/input.js"},
		// Local bindings are not the global loader.
		{Code: "function f(require) { require('dev'); } { const require = other; require.resolve('dev'); }", FileName: "public/src/input.js"},
		// Hoisted declarations shadow the global.
		{Code: "require('dev'); var require;", FileName: "public/src/input.js"},
		// Authored globals can disable require.
		{Code: "require('dev');", FileName: "public/src/input.js", Globals: map[string]any{"require": "off"}},
		// A written global stops reference tracking.
		{Code: "require('dev'); require = other;", FileName: "public/src/input.js", Globals: map[string]any{"require": "writable"}},
		// Other calls and dynamic arguments are ignored.
		{Code: "obj.require('dev'); require.cache('dev'); new require('dev'); require.call(null, 'dev'); require(...['dev']); const name = 'dev'; require(name); require(`de${name}`);", FileName: "public/src/input.js"},
		// Builtins remain exempt with aliases and development declarations.
		{Code: "require('fs'); require('node:fs'); require('fs!loader');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"fs": "./hidden.js", "node:fs": "./hidden.js"}}}}},
		// Production dependency fields override devDependencies; missing packages are checked elsewhere.
		{Code: "require('runtime'); require('peer'); require('optional'); require('unknown'); require('null-version');", FileName: "public/src/input.js"},
		// Allowed package roots include scoped and virtual packages.
		{Code: "require('dev/sub'); require('@scope/dev/sub'); require('virtual:foo/sub');", FileName: "public/src/input.js", Options: []any{map[string]any{"allowModules": []any{"dev", "@scope/dev", "virtual:foo"}, "ignorePrivate": true}}},
		// Shared node settings supply the allow list.
		{Code: "require('dev');", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{"dev"}}}},
		// Legacy settings.n takes priority over settings.node.
		{Code: "require('dev');", FileName: "public/src/input.js", Settings: map[string]any{"n": map[string]any{"allowModules": []any{"dev"}}, "node": map[string]any{"allowModules": []any{}}}},
		// Invalid shared values fall through to node settings.
		{Code: "require('dev');", FileName: "public/src/input.js", Settings: map[string]any{"n": map[string]any{"allowModules": false}, "node": map[string]any{"allowModules": []any{"dev"}}}},
		// Unpublished sources are not checked.
		{Code: "require('dev');", FileName: "public/test/input.js"},
		// Private-package defaults also apply with an explicit option object.
		{Code: "require('dev');", FileName: "private/input.js", Options: []any{map[string]any{}}, Settings: map[string]any{"node": map[string]any{"ignorePrivate": false}}},
		// Root gitignore applies when there is no files list or npmignore.
		{Code: "require('dev');", FileName: "git/input.js"},
		// Extra lookup paths are relative to the working directory.
		{Code: "require('./tool');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"public/vendor"}}}},
		// Extra lookup paths also come from shared settings.
		{Code: "require('./tool');", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"public/vendor"}}}},
		// Custom extensions resolve published targets.
		{Code: "require('./odd');", FileName: "public/src/input.js", Options: []any{map[string]any{"tryExtensions": []any{".xyz"}}}},
		// Shared custom extensions work too.
		{Code: "require('./odd');", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".xyz"}}}},
		// Resolver aliases can redirect local targets.
		{Code: "require('./hidden.js');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"./hidden.js": "./public.js"}}}}},
		// CommonJS resolution honors explicit package entry fields.
		{Code: "require('./dir');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"mainFields": []any{"browser"}}}}},
		// Custom directory entry names select a published file.
		{Code: "require('./entry');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"mainFiles": []any{"alternate"}}}}},
		// Explicit conditions can select the import branch.
		{Code: "require('#choice');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"conditionNames": []any{"import"}}}}},
		// Fallback resolves a missing request to a published file.
		{Code: "require('./missing');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fallback": map[string]any{"./missing": "./public.js"}}}}},
		// Existing targets take precedence over fallback aliases.
		{Code: "require('./public.js');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fallback": map[string]any{"./public.js": "./hidden.js"}}}}},
		// Fully specified requests check directory paths without selecting their main entry.
		{Code: "require('./public.js'); require('./dir');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fullySpecified": true}}}},
		// Fallback redirects retain extension lookup with fullySpecified.
		{Code: "require('./missing');", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"fullySpecified": true, "fallback": map[string]any{"./missing": "./public"}}}}},
		// Conversions apply to local targets.
		{Code: "require('./hidden.js');", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/hidden.js": []any{"^src/", "dist/"}}}}},
		// Invalid shared conversion patterns skip checks; upstream throws.
		{Code: "require('dev');", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"**": []any{"[", "x"}}}}},
		// TypeScript declarations and JSX do not invoke require.
		{Code: "import value = require('dev'); type T = import('dev').T; const view = <require.resolve />;", FileName: "public/src/input.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Native AST coverage: ESLint rejects this private property access.
		{Code: "class Box { #resolve; run() { require.#resolve('dev'); } }", FileName: "public/src/input.ts"},
		// Ambient declarations bind a local require.
		{Code: "declare function require(name: string): unknown; require('dev');", FileName: "public/src/input.ts"},
		// Ordinary imports are outside the require rule.
		{Code: "import 'dev'; import('dev');", FileName: "public/src/input.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Native filesystem coverage: no package metadata means no publication policy.
		{Code: "require('../../outside.js');", FileName: "no-package/input.js"},
		// An empty converted target is the package root.
		{Code: "require('./hidden.js');", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/hidden.js": []any{".*", ""}}}}},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Package roots are reported even for subdirectories and loader suffixes.
		{Code: "require('dev/sub!loader'); require('@scope/dev/path'); require('self'); require('virtual:foo/sub');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 25}, {MessageId: "notPublished", Message: "\"@scope/dev\" is not published.", Line: 1, Column: 36, EndLine: 1, EndColumn: 53}, {MessageId: "notPublished", Message: "\"self\" is not published.", Line: 1, Column: 64, EndLine: 1, EndColumn: 70}, {MessageId: "notPublished", Message: "\"virtual:foo\" is not published.", Line: 1, Column: 81, EndLine: 1, EndColumn: 98}}},
		// Aliases, destructuring, computed methods and optional calls retain argument ranges.
		{Code: "const load = require; load('dev'); const {resolve: find} = require; find('dev'); require?.('dev'); require['re' + 'solve']('dev');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 74, EndLine: 1, EndColumn: 79}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 92, EndLine: 1, EndColumn: 97}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 124, EndLine: 1, EndColumn: 129}}},
		// Global objects and escaped identifier names are tracked.
		{Code: "globalThis['requ' + 'ire']('dev'); global.require.resolve('dev'); requ\\u0069re('dev');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 59, EndLine: 1, EndColumn: 64}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 80, EndLine: 1, EndColumn: 85}}},
		// Constants preserve JS conversion and number rounding.
		{Code: "require('d' + 'ev'); require(`d${'ev'}`); require(true); require(9007199254740993);", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 30, EndLine: 1, EndColumn: 40}, {MessageId: "notPublished", Message: "\"true\" is not published.", Line: 1, Column: 51, EndLine: 1, EndColumn: 55}, {MessageId: "notPublished", Message: "\"9007199254740992\" is not published.", Line: 1, Column: 66, EndLine: 1, EndColumn: 82}}},
		// Parentheses, comments, multiline arguments and UTF-16 columns.
		{Code: "const text = '😀'; (require)((/*keep*/'dev'));\nrequire(\n  'dev'\n);", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 39, EndLine: 1, EndColumn: 44}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 3, Column: 3, EndLine: 3, EndColumn: 8}}},
		// TypeScript wrappers around arguments preserve runtime values.
		{Code: "require(('dev' as string)); require('dev'!); require('dev' satisfies string);", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 10, EndLine: 1, EndColumn: 25}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 37, EndLine: 1, EndColumn: 43}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 54, EndLine: 1, EndColumn: 76}}},
		// Private packages can opt into checking.
		{Code: "require('dev');", FileName: "private/input.js", Options: []any{map[string]any{"ignorePrivate": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// A truthy private string is not private: true.
		{Code: "require('dev');", FileName: "string-private/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// Workspace production declarations do not exempt child development dependencies.
		{Code: "require('dev');", FileName: "workspace/child/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// An empty allow list overrides shared settings.
		{Code: "require('dev');", FileName: "public/src/input.js", Options: []any{map[string]any{"allowModules": []any{}}}, Settings: map[string]any{"node": map[string]any{"allowModules": []any{"dev"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// Uninstalled development dependencies are still reported.
		{Code: "require('empty-version');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"empty-version\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 24}}},
		// Repeated local references each report, including require.resolve.
		{Code: "require('./hidden.js'); require.resolve('./hidden.js');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22}, {MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 41, EndLine: 1, EndColumn: 54}}},
		// Allowing packages does not exempt local files.
		{Code: "require('./hidden.js');", FileName: "public/src/input.js", Options: []any{map[string]any{"allowModules": []any{"dev"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22}}},
		// Directory resolution checks the actual main and index files.
		{Code: "require('./dir'); require('./entry');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./dir\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}, {MessageId: "notPublished", Message: "\"./entry\" is not published.", Line: 1, Column: 27, EndLine: 1, EndColumn: 36}}},
		// CommonJS conditions exclude the import condition.
		{Code: "require('#choice');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"#choice\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
		// Fallback resolves a missing request to an unpublished file.
		{Code: "require('./missing');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fallback": map[string]any{"./missing": "./hidden.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./missing\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		// Fully specified requests disable extension guessing before publication checks.
		{Code: "require('./odd');", FileName: "public/src/input.js", Options: []any{map[string]any{"tryExtensions": []any{".xyz"}, "resolverConfig": map[string]any{"fullySpecified": true}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./odd\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// An empty extension list overrides shared extensions.
		{Code: "require('./odd');", FileName: "public/src/input.js", Options: []any{map[string]any{"tryExtensions": []any{}}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".xyz"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./odd\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// An empty lookup list overrides shared paths.
		{Code: "require('./tool');", FileName: "public/src/input.js", Options: []any{map[string]any{"resolvePaths": []any{}}}, Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"public/vendor"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./tool\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
		// TypeScript extension mapping and local aliases use target publication.
		{Code: "require('./target.js'); require('#alias/hidden'); require('@alias/hidden');", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./target.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22}, {MessageId: "notPublished", Message: "\"#alias/hidden\" is not published.", Line: 1, Column: 33, EndLine: 1, EndColumn: 48}, {MessageId: "notPublished", Message: "\"@alias/hidden\" is not published.", Line: 1, Column: 59, EndLine: 1, EndColumn: 74}}},
		// npmignore takes priority over gitignore.
		{Code: "require('dev');", FileName: "npm/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// The main entry remains published with an empty files list.
		{Code: "require('dev');", FileName: "main/index.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// An empty convertPath object overrides shared conversion.
		{Code: "require('./hidden.js');", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": map[string]any{}}}, Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"src/hidden.js": []any{"^src/", "dist/"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22}}},
		// The shared policy reads nested ignore files; upstream reads only the root.
		{Code: "require('./nested/ignored.js');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./nested/ignored.js\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 30}}},
		// Root README metadata stays published; upstream may skip it with files: [].
		{Code: "require('dev');", FileName: "metadata/README.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// Assignment aliases and optional resolve calls.
		{Code: "let load; (load = require)('dev'); load('dev'); require?.resolve?.('dev');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 41, EndLine: 1, EndColumn: 46}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 68, EndLine: 1, EndColumn: 73}}},
		// Computed destructuring follows the global loader.
		{Code: "const {['requ' + 'ire']: load} = globalThis; load('dev');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 51, EndLine: 1, EndColumn: 56}}},
		// JSDoc casts keep the runtime argument range.
		{Code: "require(/** @type {string} */ ('dev'));", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37}}},
		// Regular expression arguments use JavaScript string conversion.
		{Code: "require(/dev/); require(/^dev$/gi);", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"/dev/\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}, {MessageId: "notPublished", Message: "\"/^dev$/gi\" is not published.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34}}},
		// Unresolved non-package names use the process directory.
		{Code: "require(''); require('data:text/javascript,0'); require('https://example.com/a.js');", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11}, {MessageId: "notPublished", Message: "\"data:text/javascript,0\" is not published.", Line: 1, Column: 22, EndLine: 1, EndColumn: 46}, {MessageId: "notPublished", Message: "\"https://example.com/a.js\" is not published.", Line: 1, Column: 57, EndLine: 1, EndColumn: 83}}},
		// String operations preserve JavaScript UTF-16 semantics.
		{Code: "require('😀dev'.slice(2)); require(['d', 'ev'].join(''));", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 9, EndLine: 1, EndColumn: 25}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 36, EndLine: 1, EndColumn: 56}}},
		// TypeScript call type arguments do not change the target.
		{Code: "require<string>('dev'); class Box { value = require('dev'); }", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 53, EndLine: 1, EndColumn: 58}}},
	}
	for i := range valid {
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "commonjs"
		}
		if valid[i].Globals == nil {
			valid[i].Globals = map[string]any{"require": "readonly", "global": "readonly", "globalThis": "readonly"}
		}
	}
	for i := range invalid {
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "commonjs"
		}
		invalid[i].Globals = map[string]any{"require": "readonly", "global": "readonly", "globalThis": "readonly"}
	}
	rule_tester.RunRuleTester(unpublishedRequireRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &NoUnpublishedRequireRule, valid, invalid)
}

// Upstream's docs show convertPath: null, but its schema rejects that value.
func TestNoUnpublishedRequireSchema(t *testing.T) {
	for _, options := range []any{
		map[string]any{"convertPath": nil},
		map[string]any{"ignorePrivate": "true"},
		map[string]any{"ignoreTypeImport": true},
		map[string]any{"allowModules": []any{"dev", "dev"}},
		map[string]any{"tryExtensions": []any{"js"}},
		map[string]any{"unknown": true},
	} {
		if err := NoUnpublishedRequireRule.Schema.Validate([]any{options}); err == nil {
			t.Fatalf("expected invalid options: %#v", options)
		}
	}
}
