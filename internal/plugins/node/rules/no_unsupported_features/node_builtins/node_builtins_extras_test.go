// cspell:ignore Tlsa
package node_builtins

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Additional reference-tracking and option cases checked against eslint-plugin-n v18.3.0.
func TestNodeBuiltinsExtras(t *testing.T) {
	runBuiltinTests(t,
		[]rule_tester.ValidTestCase{
			// unknown computed names.
			{Code: "require(mod).rm; require('fs')[key];",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// shadowed loaders.
			{Code: "function f(require, process) { require('fs').rm; process.getBuiltinModule('fs').rm; }",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// disabled globals.
			{Code: "require('fs').rm; Buffer.alloc(1);",
				Options: []any{map[string]any{"version": "0.12.0"}},
				Globals: map[string]any{"require": "off", "Buffer": "off"},
			},
			// written loader.
			{Code: "require = other; require('fs').rm;",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// rest destructuring is not inferred.
			{Code: "const {...fs} = require('fs'); fs.rm('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// recursive sea namespace.
			{Code: "require('node:sea').sea.sea.getAsset;",
				Options: []any{map[string]any{"version": "20.12.0", "allowExperimental": true}},
			},
			// prefix supports backport.
			{Code: "import fs from 'node:fs';",
				Options: []any{map[string]any{"version": "12.20.0"}},
			},
			// prefix-only modules.
			{Code: "require('test'); require('sea'); require('sqlite');",
				Options: []any{map[string]any{"version": "20"}},
			},
			// dynamic import matches upstream limitation.
			{Code: "const fs = await import('fs'); fs.rm('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// new target is not import meta.
			{Code: "function C() { return new.target.dirname; }",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// ignore affects both module spellings.
			{Code: "require('fs').rm; require('node:fs').rm;",
				Options: []any{map[string]any{"version": "14.14.0", "ignores": []any{"fs.rm"}}},
			},
			// ignore entire prefix load.
			{Code: "require('node:fs');",
				Options: []any{map[string]any{"version": "8", "ignores": []any{"fs"}}},
			},
			// experimental supported union.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": "^16.15.0 || >=17.5.0", "allowExperimental": true}},
			},
			// option precedence.
			{Code: "fetch('x');",
				Options:  []any{map[string]any{"version": "21"}},
				Settings: map[string]any{"n": map[string]any{"version": "20"}},
			},
			// empty range handling.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": ">24 <20"}},
			},
			// multiple backport branches.
			{Code: "require('module').builtinModules;",
				Options: []any{map[string]any{"version": "^6.13.0 || ^8.10.0 || >=9.3.0"}},
			},
			// JSX member tags are not member reads.
			{Code: "import fs from 'fs'; const el = <fs.rm/>;",
				Options:  []any{map[string]any{"version": "8"}},
				FileName: "input.jsx",
				Tsx:      true,
			},
			// private property is not a builtin API.
			{Code: "class C { #rm; m() { return require('fs').#rm; } }",
				Options: []any{map[string]any{"version": "8"}},
			},
			// JSDoc type is not runtime read.
			{Code: "/** @type {Buffer.alloc} */ let b;",
				Options: []any{map[string]any{"version": "0.12.0"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// parentheses and optional access.
			{Code: "(require('fs'))?.['rm']?.('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// constant property expressions.
			{Code: "require('fs')['r' + 'm']; require(`fs`).rm;",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 43},
				},
			},
			// alias through assignment.
			{Code: "let fs; fs = require('fs'); fs.rm('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
			},
			// aliases and destructuring.
			{Code: "const fs = require('fs'); const other = fs; const {rm: remove} = other; remove('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 52, EndLine: 1, EndColumn: 62},
				},
			},
			// global object aliases.
			{Code: "const {Buffer: B} = globalThis; B.alloc(1);",
				Options: []any{map[string]any{"version": "0.12.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer.alloc' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '0.12.0'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 40},
				},
			},
			// getBuiltinModule aliases.
			{Code: "const {getBuiltinModule: load} = process; load('fs').rm;",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.getBuiltinModule' is still an experimental feature and is not supported until Node.js 22.3.0 (backported: ^20.16.0). The configured version range is '8.0.0'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 30},
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 43, EndLine: 1, EndColumn: 56},
				},
			},
			// recursive Module namespace.
			{Code: "require('module').Module.Module.createRequire;",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'module.Module.Module.createRequire' is still an experimental feature and is not supported until Node.js 12.2.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
				},
			},
			// recursive namespace aliases.
			{Code: "const {Module: m} = require('module'); const a = m.Module; a.Module.createRequire;",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'module.Module.Module.Module.createRequire' is still an experimental feature and is not supported until Node.js 12.2.0. The configured version range is '8.0.0'.", Line: 1, Column: 60, EndLine: 1, EndColumn: 82},
				},
			},
			// prefix has its own introduction.
			{Code: "require('fs'); require('node:fs');",
				Options: []any{map[string]any{"version": "12.19.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs' is still an experimental feature and is not supported until Node.js 14.13.1 (backported: ^12.20.0). The configured version range is '12.19.0'.", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			// prefix unsupported release gap.
			{Code: "import fs from 'node:fs';",
				Options: []any{map[string]any{"version": "13.14.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs' is still an experimental feature and is not supported until Node.js 14.13.1 (backported: ^12.20.0). The configured version range is '13.14.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// unknown prefixed legacy module.
			{Code: "require('node:constants'); require('node:domain');",
				Options: []any{map[string]any{"version": "0.0.1"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'require' is still an experimental feature and is not supported until Node.js 0.1.13. The configured version range is '0.0.1'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "not-supported-till", Message: "The 'require' is still an experimental feature and is not supported until Node.js 0.1.13. The configured version range is '0.0.1'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 35},
				},
			},
			// known prefix-only test.
			{Code: "import test from 'node:test';",
				Options: []any{map[string]any{"version": "16.16.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'test' is still an experimental feature and is not supported until Node.js 20.0.0. The configured version range is '16.16.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			// default import member.
			{Code: "import fs from 'fs'; fs.rm('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
				},
			},
			// namespace default alias.
			{Code: "import * as fs from 'fs'; fs.default.rm('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 40},
				},
			},
			// re-export property.
			{Code: "export { rm as remove } from 'fs';",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// export all stable ordering.
			{Code: "export * from 'dns';",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "not-supported-till", Message: "The 'dns.resolveCaa' is still an experimental feature and is not supported until Node.js 15.0.0 (backported: ^14.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "not-supported-till", Message: "The 'dns.resolveTlsa' is still an experimental feature and is not supported until Node.js 23.9.0 (backported: ^22.15.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "not-supported-till", Message: "The 'dns.setDefaultResultOrder' is still an experimental feature and is not supported until Node.js 16.4.0 (backported: ^14.18.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "not-supported-till", Message: "The 'dns.getDefaultResultOrder' is still an experimental feature and is not supported until Node.js 20.1.0 (backported: ^18.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "not-supported-till", Message: "The 'dns.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// namespace re-export.
			{Code: "export * as dns from 'dns';",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "not-supported-till", Message: "The 'dns.resolveCaa' is still an experimental feature and is not supported until Node.js 15.0.0 (backported: ^14.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "not-supported-till", Message: "The 'dns.resolveTlsa' is still an experimental feature and is not supported until Node.js 23.9.0 (backported: ^22.15.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "not-supported-till", Message: "The 'dns.setDefaultResultOrder' is still an experimental feature and is not supported until Node.js 16.4.0 (backported: ^14.18.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "not-supported-till", Message: "The 'dns.getDefaultResultOrder' is still an experimental feature and is not supported until Node.js 20.1.0 (backported: ^18.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "not-supported-till", Message: "The 'dns.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// import meta aliases and destructuring.
			{Code: "const meta = import.meta; const {dirname: dir} = meta;",
				Options: []any{map[string]any{"version": "20.10.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '20.10.0'.", Line: 1, Column: 34, EndLine: 1, EndColumn: 46},
				},
			},
			// import meta optional access.
			{Code: "import.meta?.['resolve']('x');",
				Options: []any{map[string]any{"version": "18.18.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '18.18.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// ignores are exact names.
			{Code: "require('fs').rm;",
				Options: []any{map[string]any{"version": "8", "ignores": []any{"fs"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// experimental not yet introduced.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": "14", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'fetch' is not an experimental feature until Node.js 17.5.0 (backported: ^16.15.0). The configured version range is '14'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// experimental does not enable stable-only features.
			{Code: "require('fs').rm;",
				Options: []any{map[string]any{"version": "8", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// experimental release gap.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": "^16.15.0 || 17.4.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'fetch' is not an experimental feature until Node.js 17.5.0 (backported: ^16.15.0). The configured version range is '^16.15.0 || 17.4.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// unsupported feature without stable release.
			{Code: "import.meta.main;",
				Options: []any{map[string]any{"version": "24.2.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'import.meta.main' is still an experimental feature The configured version range is '24.2.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// fallback version.
			{Code: "fetch('x');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '>=16.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// explicit fallback and empty ignores.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": ">=16.0.0", "allowExperimental": false, "ignores": []any{}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '>=16.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// empty options.
			{Code: "fetch('x');",
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '>=16.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// invalid option falls through settings.
			{Code: "fetch('x');",
				Options:  []any{map[string]any{"version": "invalid"}},
				Settings: map[string]any{"node": map[string]any{"version": "20"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '20'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// settings precedence.
			{Code: "fetch('x');",
				Settings: map[string]any{"n": map[string]any{"version": "20"}, "node": map[string]any{"version": "21"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '20'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// raw range preserves operator spacing.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": "  >=  16.0.0\n"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '>= 16.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// prerelease excluded from stable.
			{Code: "fetch('x');",
				Options: []any{map[string]any{"version": "21.0.0-rc.1"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '21.0.0-rc.1'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// backport gap.
			{Code: "require('module').builtinModules;",
				Options: []any{map[string]any{"version": ">=6.13.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'module.builtinModules' is still an experimental feature and is not supported until Node.js 9.3.0 (backported: ^8.10.0, ^6.13.0). The configured version range is '>=6.13.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			// UTF-16 multiline positions.
			{Code: "const text = '😀';\n  require('fs')\n    .rm('x');",
				Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 2, Column: 3, EndLine: 3, EndColumn: 8},
				},
			},
			// typescript assertion wrapper.
			{Code: "(require('fs') as any).rm;",
				Options:  []any{map[string]any{"version": "8"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// typescript non-null wrapper.
			{Code: "require('fs')!.rm;",
				Options:  []any{map[string]any{"version": "8"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			// typescript import type.
			{Code: "import type {rm} from 'fs'; type T = typeof rm;",
				Options:  []any{map[string]any{"version": "8"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16},
				},
			},
			// typescript runtime heritage.
			{Code: "import { AsyncLocalStorage } from 'async_hooks'; class Store extends AsyncLocalStorage {}",
				Options:  []any{map[string]any{"version": "10"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '10'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
					{MessageId: "not-supported-till", Message: "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '10'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
				},
			},
		},
	)
}

func TestNodeBuiltinsVersionMetadata(t *testing.T) {
	base := fixtures.GetRootDir()
	dir := tspath.ResolvePath(base.Dir, "builtins-version-fixtures")
	archive := txtarfs.MustParseFile(t, "testdata/versions.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(dir, name)] = string(data)
	}
	root := rule_tester.Root{Dir: dir, FS: utils.NewOverlayVFS(base.FS, files)}
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		file, version string
		options       any
		settings      map[string]any
	}{
		{file: "input.js", version: ">=16.0.0"},
		{file: "engines/input.js", version: "^18.0.0"},
		{file: "dev/input.js", version: "20"},
		{file: "array/input.js", version: "18"},
		{file: "empty/input.js", version: ""},
		{file: "engines/nearest/input.js", version: ">=16.0.0"},
		{file: "engines/option.js", version: "19", options: []any{map[string]any{"version": "19"}}, settings: map[string]any{"node": map[string]any{"version": "20"}}},
		{file: "engines/settings.js", version: "20", settings: map[string]any{"node": map[string]any{"version": "20"}}},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: "fetch('x');", FileName: test.file, Globals: builtinTestGlobals(nil), Options: test.options, Settings: test.settings,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '" + test.version + "'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
		})
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NodeBuiltinsRule, nil, invalid)
}

func TestNodeBuiltinsSchema(t *testing.T) {
	for _, options := range [][]any{
		{map[string]any{"version": 18}},
		{map[string]any{"allowExperimental": "true"}},
		{map[string]any{"ignores": []any{"node:fs.rm"}}},
		{map[string]any{"ignores": []any{"unknown"}}},
		{map[string]any{"ignores": []any{"fs.rm", "fs.rm"}}},
		{map[string]any{"extra": true}},
		{map[string]any{}, map[string]any{}},
	} {
		if err := NodeBuiltinsRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options: %#v", options)
		}
	}
}

func TestNodeBuiltinsInvocationMetadata(t *testing.T) {
	runBuiltinTests(t,
		[]rule_tester.ValidTestCase{
			// constructor without a call restriction.
			{Code: "new (require('http').Server)();", Options: []any{map[string]any{"version": "24"}}},
			// ignore canonical read name also ignores calls.
			{Code: "require('http').Server();", Options: []any{map[string]any{"version": "24", "ignores": []any{"http.Server"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// constructor introduction differs from read.
			{Code: "new Buffer(0);", Options: []any{map[string]any{"version": "0.1.89"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer' is still an experimental feature and is not supported until Node.js 0.1.90. The configured version range is '0.1.89'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
					{MessageId: "not-supported-till", Message: "The 'Buffer' is still an experimental feature and is not supported until Node.js 0.1.103. The configured version range is '0.1.89'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
				},
			},
			// constructor stable while read unsupported.
			{Code: "new Buffer(0);", Options: []any{map[string]any{"version": "0.1.90"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer' is still an experimental feature and is not supported until Node.js 0.1.103. The configured version range is '0.1.90'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
				},
			},
			// constructors with no stable support.
			{Code: "const {Stats} = require('fs'); new Stats();", Options: []any{map[string]any{"version": "24"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'fs.Stats' is still an experimental feature The configured version range is '24'.", Line: 1, Column: 32, EndLine: 1, EndColumn: 43},
				},
			},
			// call metadata is independent of read.
			{Code: "require('http').Server();", Options: []any{map[string]any{"version": "24"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'http.Server' is still an experimental feature The configured version range is '24'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// read call and construct on same feature.
			{Code: "const {Hash} = require('crypto'); Hash('sha256'); new Hash('sha256');", Options: []any{map[string]any{"version": "24"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'crypto.Hash' is still an experimental feature The configured version range is '24'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 49},
					{MessageId: "not-supported-yet", Message: "The 'crypto.Hash' is still an experimental feature The configured version range is '24'.", Line: 1, Column: 51, EndLine: 1, EndColumn: 69},
				},
			},
			// call ignores follow upstream canonical path.
			{Code: "require('http').Server();", Options: []any{map[string]any{"version": "24", "ignores": []any{"http.Server()"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'http.Server' is still an experimental feature The configured version range is '24'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
		},
	)
}

// eslint-plugin-n v18.3.0 differential probes for additional TS/JSDoc and binding shapes.
func TestNodeBuiltinsLanguageBoundaries(t *testing.T) {
	runBuiltinTests(t,
		[]rule_tester.ValidTestCase{
			// parameterTDZ.
			{Code: "function f(x = require('fs').rm, require) {}", Options: []any{map[string]any{"version": "8.0.0"}}},
			// JSDocFetch.
			{Code: "/** @type {typeof fetch} */ let request;", Options: []any{map[string]any{"version": "8.0.0"}}},
			// JSDocDOM.
			{Code: "/** @type {typeof DOMException} */ let ErrorType;", Options: []any{map[string]any{"version": "8.0.0"}}},
			// constantKey.
			{Code: "const key = 'rm'; require('fs')[key];", Options: []any{map[string]any{"version": "8.0.0"}}},
			// constantModule.
			{Code: "const mod = 'fs'; require(mod).rm;", Options: []any{map[string]any{"version": "8.0.0"}}},
			// disableDirective.
			{Code: "// eslint-disable-next-line\nrequire('fs').rm;", Options: []any{map[string]any{"version": "8.0.0"}}},
			// importEquals.
			{Code: "import fs = require('fs'); fs.rm;", Options: []any{map[string]any{"version": "8.0.0"}},
				FileName: "input.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			// assignmentPattern.
			{Code: "let rm; ({rm} = require('fs')); rm('x');", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
				},
			},
			// nestedBindingDefault.
			{Code: "const {promises: {rm: remove} = {}} = require('fs'); remove('x');", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '8.0.0'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 35},
					{MessageId: "not-supported-till", Message: "The 'fs.promises.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29},
				},
			},
			// aliasDefault.
			{Code: "function f(fs = require('fs')) { fs.rm('x'); }", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 34, EndLine: 1, EndColumn: 39},
				},
			},
			// compoundAssignment.
			{Code: "let fs; fs ||= require('fs'); fs.rm;", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36},
				},
			},
			// memberWrite.
			{Code: "require('fs').rm = alternative;", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// JSDocCast.
			{Code: "(/** @type {any} */ (require('fs'))).rm;", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			// escapedIdentifier.
			{Code: "f\\u0065tch('x');", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			// optionalRequire.
			{Code: "require?.('fs').rm;", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// unicodeAndBOM.
			{Code: "\ufeff/*😀*/ require('fs').rm;", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24},
				},
			},
			// commentsInRange.
			{Code: "(/*x*/ require('fs') /*y*/).rm;", Options: []any{map[string]any{"version": "8.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			// typeQuery.
			{Code: "type T = typeof fetch;", Options: []any{map[string]any{"version": "8.0.0"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '8.0.0'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22},
				},
			},
			// typeExport.
			{Code: "export type {rm} from 'fs';", Options: []any{map[string]any{"version": "8.0.0"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16},
				},
			},
			// inlineTypeImport.
			{Code: "import {type rm} from 'fs';", Options: []any{map[string]any{"version": "8.0.0"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '8.0.0'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16},
				},
			},
			// decorators.
			{Code: "@decorate(fetch) class C {}", Options: []any{map[string]any{"version": "8.0.0"}},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '8.0.0'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
		},
	)
}

func TestNodeBuiltinsRepeatedImportMeta(t *testing.T) {
	const code = "const meta = import.meta; meta.dirname; import.meta.filename; meta.dirname; import.meta.dirname;"
	message := func(name string, column, endColumn int) rule_tester.InvalidTestCaseError {
		return rule_tester.InvalidTestCaseError{
			MessageId: "not-supported-till",
			Message:   "The 'import.meta." + name + "' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '20.0.0'.",
			Line:      1, Column: column, EndLine: 1, EndColumn: endColumn,
		}
	}
	runBuiltinTests(t, nil, []rule_tester.InvalidTestCase{
		{Code: code, Options: []any{map[string]any{"version": "20.0.0"}}, Errors: []rule_tester.InvalidTestCaseError{
			message("dirname", 27, 39), message("filename", 41, 61), message("dirname", 63, 75), message("dirname", 77, 96),
		}},
		{Code: code, Options: []any{map[string]any{"version": "20.0.0", "ignores": []any{"import.meta.dirname"}}}, Errors: []rule_tester.InvalidTestCaseError{
			message("filename", 41, 61),
		}},
	})
}

// Release metadata contains a v-prefixed backport; compare and format it as npm semver does.
func TestNodeBuiltinsPrefixedReleaseVersions(t *testing.T) {
	runBuiltinTests(t,
		[]rule_tester.ValidTestCase{
			{Code: "performance.clearResourceTimings();", Options: []any{map[string]any{"version": "16.17.0"}}},
			{Code: "const {performance: p} = require('node:perf_hooks'); p.clearResourceTimings();", Options: []any{map[string]any{"version": "16.17.0"}}},
			{Code: "performance.clearResourceTimings();", Options: []any{map[string]any{"version": "18.2.0"}}},
			{Code: "const {performance: p} = require('node:perf_hooks'); p.clearResourceTimings();", Options: []any{map[string]any{"version": "18.2.0"}}},
			{Code: "performance.clearResourceTimings();", Options: []any{map[string]any{"version": "^16.17.0 || >=18.2.0"}}},
			{Code: "const {performance: p} = require('node:perf_hooks'); p.clearResourceTimings();", Options: []any{map[string]any{"version": "^16.17.0 || >=18.2.0"}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "performance.clearResourceTimings();", Options: []any{map[string]any{"version": "16.16.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'performance.clearResourceTimings' is still an experimental feature and is not supported until Node.js 18.2.0 (backported: ^v16.17.0). The configured version range is '16.16.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{Code: "const {performance: p} = require('node:perf_hooks'); p.clearResourceTimings();", Options: []any{map[string]any{"version": "16.16.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'perf_hooks.performance.clearResourceTimings' is still an experimental feature and is not supported until Node.js 18.2.0 (backported: ^v16.17.0). The configured version range is '16.16.0'.", Line: 1, Column: 54, EndLine: 1, EndColumn: 76},
				},
			},
			{Code: "performance.clearResourceTimings();", Options: []any{map[string]any{"version": "17.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'performance.clearResourceTimings' is still an experimental feature and is not supported until Node.js 18.2.0 (backported: ^v16.17.0). The configured version range is '17.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{Code: "const {performance: p} = require('node:perf_hooks'); p.clearResourceTimings();", Options: []any{map[string]any{"version": "17.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'perf_hooks.performance.clearResourceTimings' is still an experimental feature and is not supported until Node.js 18.2.0 (backported: ^v16.17.0). The configured version range is '17.0.0'.", Line: 1, Column: 54, EndLine: 1, EndColumn: 76},
				},
			},
			{Code: "performance.clearResourceTimings();", Options: []any{map[string]any{"version": "18.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'performance.clearResourceTimings' is still an experimental feature and is not supported until Node.js 18.2.0 (backported: ^v16.17.0). The configured version range is '18.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{Code: "const {performance: p} = require('node:perf_hooks'); p.clearResourceTimings();", Options: []any{map[string]any{"version": "18.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'perf_hooks.performance.clearResourceTimings' is still an experimental feature and is not supported until Node.js 18.2.0 (backported: ^v16.17.0). The configured version range is '18.1.0'.", Line: 1, Column: 54, EndLine: 1, EndColumn: 76},
				},
			},
		},
	)
}
