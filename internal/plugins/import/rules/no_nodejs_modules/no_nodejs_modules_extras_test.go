package no_nodejs_modules_test

import (
	"runtime"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_nodejs_modules"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoNodejsModulesResolution(t *testing.T) {
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-nodejs-modules")
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
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	for _, resolver := range []string{"node", "typescript"} {
		t.Run(resolver, func(t *testing.T) {
			settings := map[string]any{"import/core-modules": []any{"virtual"}, "import/resolver": resolver}
			rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_nodejs_modules.NoNodejsModulesRule,
				[]rule_tester.ValidTestCase{
					// A resolved subpath overrides lexical builtin classification.
					{Code: "import 'virtual/shim'; import 'fs/shim';", Settings: settings},
					// Removed stream internals may resolve to installed packages.
					{Code: "import '_stream_readable';", Settings: settings},
				},
				[]rule_tester.InvalidTestCase{
					// Exact core names take precedence over installed packages and TS paths.
					{Code: "import 'virtual';", Settings: settings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: `Do not import Node.js builtin module "virtual"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
					{Code: "import 'fs';", Settings: settings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: `Do not import Node.js builtin module "fs"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13}}},
					{Code: "import 'virtual/missing';", Settings: settings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: `Do not import Node.js builtin module "virtual/missing"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 26}}},
				},
			)
		})
	}
}

func TestNoNodejsModulesResolverErrors(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_nodejs_modules.NoNodejsModulesRule, nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "import 'ordinary';\nimport 'fs';", Settings: map[string]any{"import/resolver": 17},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Resolve error: invalid resolver config", Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
					{MessageId: "", Message: `Do not import Node.js builtin module "fs"`, Line: 2, Column: 1, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code: "import 'fs';", Settings: map[string]any{"import/resolver": "webpack"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: `Resolve error: unable to load resolver "webpack".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
					{MessageId: "", Message: `Do not import Node.js builtin module "fs"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
		},
	)
}

func TestNoNodejsModulesSchema(t *testing.T) {
	for _, options := range [][]any{
		{map[string]any{"allow": "fs"}},
		{map[string]any{"allow": []any{"fs", "fs"}}},
		{map[string]any{"allow": []any{17}}},
		{map[string]any{"unknown": true}},
		{map[string]any{}, map[string]any{}},
		{"fs"},
	} {
		if err := no_nodejs_modules.NoNodejsModulesRule.Schema.Validate(options); err == nil {
			t.Errorf("expected invalid options to be rejected: %#v", options)
		}
	}
}

func TestNoNodejsModulesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_nodejs_modules.NoNodejsModulesRule,
		[]rule_tester.ValidTestCase{
			// Removed builtins are also allowed when no matching package exists.
			{Code: "import '_stream_readable'; import 'node:_stream_readable';"},
			// Initializing the rule without module references must not resolve anything.
			{Code: "export const value = 1;", Settings: map[string]any{"import/resolver": 17}},
			// JSDoc imports are comments, not module declarations to check.
			{Code: "/** @import { Stats } from 'fs' */\n/** @type {import('fs').Stats} */\nlet stats;", FileName: "jsdoc.js", TSConfig: "tsconfig.allow-js.json"},
			// Authored satisfies expressions stay distinct from their string or callee.
			{Code: "require(('fs') satisfies string); (require satisfies any)('fs');"},
			{Code: "class C { #require; load() { this.#require('fs'); } } const node = <require.fs/>;", Tsx: true},
			// Disable directives are handled by the context, including after lazy setup.
			{Code: "// eslint-disable-next-line test\nimport 'fs';"},
			// Non-module syntax, member calls, AMD, and arguments other than literals are ignored.
			{Code: "const fs = 'fs'; export { fs }; export default fs; require(); require('fs', true); require(1); require(null); require(/fs/); require(`fs`); require('f' + 's'); import(`fs`); import(fs); require(...['fs']); obj.require('fs'); obj['require']('fs'); obj?.require('fs'); require.resolve('fs'); new require('fs'); define(['fs'], () => {}); require(['fs'], () => {});"},
			// TypeScript wrappers and type queries are not string literal module references.
			{Code: "import alias = require('fs'); type T = import('fs').Stats; require('fs' as string); (require as any)('fs'); require!('fs'); import('fs' as string);"},
			// Module names are case-sensitive and unknown node: names are not builtins.
			{Code: "import 'FS'; import 'node:FS'; import 'node:not-real'; import ''; import '@scope/fs'; import './fs';"},
			// The allow option matches decoded names exactly, including whitespace and empty strings.
			{Code: "import '\\u0066s'; require('node:fs'); import('fs/promises');",
				Options: []any{map[string]any{"allow": []any{"fs", "node:fs", "fs/promises", "", " fs "}}},
			},
			// Internal patterns use JavaScript regexps and take precedence over configured core modules.
			{Code: "import 'node:fs'; require('virtual');",
				Settings: map[string]any{"import/internal-regex": "(?<=node:)fs$|^virtual$", "import/core-modules": []any{"virtual"}},
			},
			// A configured full subpath does not make its package root builtin.
			{Code: "import 'virtual/subpath';",
				Settings: map[string]any{"import/core-modules": []any{"virtual/subpath"}},
			},
			// An absolute path and an empty specifier never classify as builtins.
			{Code: "import '/fs'; import '';",
				Settings: map[string]any{"import/core-modules": []any{"", "/fs"}},
			},
			// Allowed sources bypass resolution, even with an invalid resolver configuration.
			{Code: "import 'fs';",
				Options:  []any{map[string]any{"allow": []any{"fs"}}},
				Settings: map[string]any{"import/resolver": 17},
			},
		},
		[]rule_tester.InvalidTestCase{
			// An allowed first reference must not suppress a later disallowed module.
			{Code: "import 'fs';\nimport 'path';", Options: []any{map[string]any{"allow": []any{"fs"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "path"`, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
			}},
			// Unlike enforce-node-protocol-usage, upstream ignores import/node-version.
			{Code: "import 'node:fs';", Settings: map[string]any{"import/node-version": "0.0.0"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "node:fs"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
			}},
			{Code: "import 'node:fs';", Settings: map[string]any{"import/node-version": false}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "node:fs"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
			}},
			// The scanner already decodes escaped identifiers; generic calls remain calls.
			// cspell:ignore equire
			{Code: `\u0072equire('fs');`, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "fs"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			}},
			{Code: "require<string>('fs'); require?.<string>('path');", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "fs"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				{MessageId: "", Message: `Do not import Node.js builtin module "path"`, Line: 1, Column: 24, EndLine: 1, EndColumn: 49},
			}},
			// Module declarations do not hide nested import and export declarations.
			{Code: "declare module 'example' { import type { Stats } from 'fs'; export { Stats } from 'node:fs'; }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "fs"`, Line: 1, Column: 28, EndLine: 1, EndColumn: 60},
				{MessageId: "", Message: `Do not import Node.js builtin module "node:fs"`, Line: 1, Column: 61, EndLine: 1, EndColumn: 93},
			}},
			// Scoped core names can contain astral characters; ranges count UTF-16 units.
			{Code: "import '@😀/core';", Settings: map[string]any{"import/core-modules": []any{"@😀/core"}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "", Message: `Do not import Node.js builtin module "@😀/core"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			}},
			// An empty options object preserves the default.
			{Code: "import 'fs';",
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// An explicit empty allow list preserves the default.
			{Code: "require('fs');",
				Options: []any{map[string]any{"allow": []any{}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// Prefixes, module subpath names, case, and glob-like strings are distinct allow entries.
			{Code: "import 'node:fs'; import 'fs/promises'; import 'path';",
				Options: []any{map[string]any{"allow": []any{"fs", "PATH", "node:*"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs/promises\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 40},
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 1, Column: 41, EndLine: 1, EndColumn: 55},
				},
			},
			// Allowing node:fs does not allow the unprefixed fs specifier.
			{Code: "import 'fs';",
				Options: []any{map[string]any{"allow": []any{"node:fs"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// All static re-export shapes report the full declaration.
			{Code: "export { readFile } from 'fs';\nexport * from 'path';\nexport * as fs from 'node:fs';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 3, Column: 1, EndLine: 3, EndColumn: 31},
				},
			},
			// Explicit type imports and re-exports are checked, unlike import type queries.
			{Code: "import type FS from 'fs';\nimport { type Stats } from 'node:fs';\nexport type { Stats } from 'fs';\nexport type * from 'fs';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 2, Column: 1, EndLine: 2, EndColumn: 38},
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 3, Column: 1, EndLine: 3, EndColumn: 33},
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 4, Column: 1, EndLine: 4, EndColumn: 25},
				},
			},
			// Import attributes and dynamic-import options retain the complete report range.
			{Code: "import fs from 'fs' with { type: 'json' };\nimport('node:fs', { with: { type: 'json' } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 2, Column: 1, EndLine: 2, EndColumn: 46},
				},
			},
			// Parentheses and optional calls retain ESTree call and argument semantics.
			{Code: "(require)(('fs')); require?.('path'); ((require('node:fs'))); import(('fs'));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 1, Column: 20, EndLine: 1, EndColumn: 37},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 1, Column: 41, EndLine: 1, EndColumn: 59},
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 63, EndLine: 1, EndColumn: 77},
				},
			},
			// The rule checks require spelling rather than binding identity.
			{Code: "function load(require: any) { return require('fs'); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 38, EndLine: 1, EndColumn: 51},
				},
			},
			// JSX tag names are ignored while calls inside JSX expressions are visited.
			{Code: "const view = <require.fs value={require('fs')} />;",
				FileName: "virtual.tsx",
				Tsx:      true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 33, EndLine: 1, EndColumn: 46},
				},
			},
			// Unicode preceding a call and multiline declarations use UTF-16 columns.
			{Code: "const text = '😀'; require('fs');\nimport {\n  readFile\n} from 'node:fs';\nrequire('f\\\ns');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 20, EndLine: 1, EndColumn: 33},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 2, Column: 1, EndLine: 4, EndColumn: 18},
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 5, Column: 1, EndLine: 6, EndColumn: 4},
				},
			},
			// Escaped string literals are classified by their decoded value.
			{Code: "import '\\u0066s';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			// The package root determines unresolved subpath classification, including legacy builtins.
			{Code: "import 'fs/not-a-real-subpath'; import 'node:fs/extra'; import '_http_agent'; import 'node:_stream_readable'; import 'node:test/reporters';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs/not-a-real-subpath\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs/extra\"", Line: 1, Column: 33, EndLine: 1, EndColumn: 56},
					{MessageId: "", Message: "Do not import Node.js builtin module \"_http_agent\"", Line: 1, Column: 57, EndLine: 1, EndColumn: 78},
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:test/reporters\"", Line: 1, Column: 111, EndLine: 1, EndColumn: 140},
				},
			},
			// Configured core module roots include scoped names and unresolved relative roots.
			{Code: "import 'virtual'; import '@scope/core/subpath'; import '../missing';",
				Settings: map[string]any{"import/core-modules": []any{"virtual", "@scope/core", ".."}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"virtual\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: "Do not import Node.js builtin module \"@scope/core/subpath\"", Line: 1, Column: 19, EndLine: 1, EndColumn: 48},
					{MessageId: "", Message: "Do not import Node.js builtin module \"../missing\"", Line: 1, Column: 49, EndLine: 1, EndColumn: 69},
				},
			},
			// import/ignore does not filter this rule.
			{Code: "import 'fs';",
				Settings: map[string]any{"import/ignore": []any{".*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// A source that does not match the internal regexp must still be reported.
			{Code: "import 'node:fs';",
				Settings: map[string]any{"import/internal-regex": "^fs$"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
		},
	)
}

func TestNoNodejsModulesHostPaths(t *testing.T) {
	// Node path.isAbsolute is host-specific and does not classify URLs as paths.
	// TypeScript's portable path predicates deliberately have different semantics.
	settings := map[string]any{"import/core-modules": []any{"file:", "C:"}}
	valid := []rule_tester.ValidTestCase{}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "import 'file:///virtual';", Settings: settings, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "", Message: `Do not import Node.js builtin module "file:///virtual"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
		}},
	}
	if runtime.GOOS == "windows" {
		valid = append(valid, rule_tester.ValidTestCase{Code: "import 'C:/virtual';", Settings: settings})
	} else {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: "import 'C:/virtual';", Settings: settings, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "", Message: `Do not import Node.js builtin module "C:/virtual"`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_nodejs_modules.NoNodejsModulesRule, valid, invalid)
}
