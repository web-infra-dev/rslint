// Additional cases compared with eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-deprecated-api.js
// cspell:ignore cipheriv decipheriv fips freelist lchmod linklist prng unenroll
package no_deprecated_api

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoDeprecatedAPIExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Type assertions on assignment targets do not create tracked aliases
		{Code: "let B; (B as any) = Buffer; new B();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Type queries are not runtime property reads
		{Code: "type F = typeof process.binding; type B = typeof Buffer; interface T {x: typeof process.binding}",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Namespace declarations shadow configured globals
		{Code: "namespace process { export type T = number; } process.binding(); namespace Buffer {export type T = number} new Buffer();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Authored ambient variables shadow configured globals
		{Code: "declare const Buffer: any; new Buffer(); declare const process: any; process.binding();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// JSDoc type expressions are not runtime reads
		{Code: "/** @type {typeof process.binding} */ let foo; /** @typedef {typeof Buffer} B */ const b = 1;",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "deprecated-input.js", TSConfig: "tsconfig.allowJs.json",
		},
		// Import-equals declarations are not ESM imports
		{Code: "import b = require('buffer'); new b.Buffer();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Spread arguments are not static module names
		{Code: "require(...['fs']).exists; process.getBuiltinModule(...['fs']).exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "deprecated-input.js", TSConfig: "tsconfig.allowJs.json",
		},
		// Only real Node builtins have node: aliases, including removed APIs.
		{Code: "require('node:safe-buffer').Buffer(); require('node:_linklist'); require('node:freelist');"},
		// Shadowed and modified global roots
		{Code: "function f(require, process, Buffer) { require('fs').exists; process.binding(); new Buffer(); }",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		},
		// A write disables global tracking throughout the file
		{Code: "Buffer(1); Buffer = Other; global.Buffer(2); global = other; require('fs').exists; require = other; process.binding(); process = other;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		},
		// Unresolved assignment does not introduce a binding
		{Code: "missing = require('buffer').Buffer; new missing();",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		},
		// Destructuring rest is not traced
		{Code: "const {...b} = require('buffer'); new b.Buffer();",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		},
		// Array patterns are not traced
		{Code: "const [b] = require('buffer'); new b();",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		},
		// Undeclared and disabled globals
		{Code: "Buffer(); process.binding(); require('fs').exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Globals:         map[string]any{"Buffer": "off", "process": "off", "require": "off"},
		},
		// Private property is not a public deprecated API
		{Code: "class C { #exists; f() { require('fs').#exists; } }",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		},
		// Dynamic import and resolve do not read APIs
		{Code: "import('domain'); require.resolve('domain'); const fs = await import('fs'); fs.exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Imported process does not root getBuiltinModule tracking
		{Code: "import process from 'process'; process.getBuiltinModule('fs').exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// TS heritage is not a call
		{Code: "import {Buffer} from 'buffer'; class C extends Buffer {}",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
	}
	invalid := []rule_tester.InvalidTestCase{
		// tsgo unwraps template literal property names in binding patterns.
		{Code: "const {[`Buffer`]: B} = require('buffer'); new B(); const {[`require`]: load} = global; load('fs').exists;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 89, EndLine: 1, EndColumn: 106},
			},
		},
		// Object parameter defaults track module APIs
		{Code: "function f({Buffer: B} = require('buffer')) { new B(); } function g({exists: e} = require('fs')) {}",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 70, EndLine: 1, EndColumn: 79},
			},
		},
		// Quoted import and export names
		{Code: "import {'Buffer' as B} from 'buffer'; new B(); export {'exists' as 'check'} from 'fs';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 39, EndLine: 1, EndColumn: 46},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 56, EndLine: 1, EndColumn: 75},
			},
		},
		// Merged type and value names retain the value binding
		{Code: "type B = string; let B; B = Buffer; new B();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
			},
		},
		// JSDoc declarations do not shadow configured globals
		{Code: "/** @typedef {{x:number}} Buffer */ Buffer(); /** @namespace process */ process.binding();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "deprecated-input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 45},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 73, EndLine: 1, EndColumn: 88},
			},
		},
		// Logical assignment aliases
		{Code: "let B; B ||= Buffer; new B(); let C; C &&= Buffer; new C(); let D; D ??= Buffer; new D();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "deprecated-input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 82, EndLine: 1, EndColumn: 89},
			},
		},
		// Class static blocks retain lexical shadowing
		{Code: "class C {static {const B = Buffer; new B();} f(){const Buffer = other; new Buffer();}}",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "deprecated-input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
			},
		},
		// Decorator expressions read runtime APIs
		{Code: "class C { @process.binding accessor x; }",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 12, EndLine: 1, EndColumn: 27},
			},
		},
		// Only direct references inside a namespace are traced
		{Code: "namespace N { export const B = Buffer; new B(); } new N.B();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 40, EndLine: 1, EndColumn: 47},
			},
		},
		// Using declarations retain module aliases
		{Code: "using b = require('buffer'); new b.Buffer();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 44},
			},
		},
		// Optional access, calls and parenthesized constructors
		{Code: "require?.('fs')?.exists?.('x'); (Buffer)?.(1); new (Buffer)(2); (process?.binding)('x');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 33, EndLine: 1, EndColumn: 46},
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 48, EndLine: 1, EndColumn: 63},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 66, EndLine: 1, EndColumn: 82},
			},
		},
		// Computed literal expressions and dynamic keys
		{Code: "require('f' + 's')[`${'exists'}`]; require('fs')[key]; const key = 'exists';",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			},
		},
		// Nested assignment destructuring
		{Code: "let found; ({exists: found} = require('fs')); let legacy; ({env: {NODE_REPL_HISTORY_FILE: legacy = ''}} = process);",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 14, EndLine: 1, EndColumn: 27},
				{MessageId: "deprecated", Message: "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.", Line: 1, Column: 67, EndLine: 1, EndColumn: 102},
			},
		},
		// Defaults and assignment expressions
		{Code: "let B; function f(B = require('buffer').Buffer) { new B(); } ({B = require('buffer').Buffer} = {}); new B();",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 51, EndLine: 1, EndColumn: 58},
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 101, EndLine: 1, EndColumn: 108},
			},
		},
		// Logical, conditional and sequence aliases
		{Code: "const b = ok ? require('buffer') : other; new b.Buffer(); const c = (unused, require('fs')); c.exists; const d = other ?? require('fs'); d.exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 43, EndLine: 1, EndColumn: 57},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 94, EndLine: 1, EndColumn: 102},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 138, EndLine: 1, EndColumn: 146},
			},
		},
		// Read callbacks stop at aliases
		{Code: "const exists = require('fs').exists; exists; exists(); const p = process; p.binding;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 16, EndLine: 1, EndColumn: 36},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 75, EndLine: 1, EndColumn: 84},
			},
		},
		// Cycles do not recurse
		{Code: "let a = require('buffer'); let b = a; a = b; new b.Buffer();",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 46, EndLine: 1, EndColumn: 60},
			},
		},
		// Configured assignment alias
		{Code: "saved = require('buffer').Buffer; new saved();",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Globals:         map[string]any{"saved": "writable"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 35, EndLine: 1, EndColumn: 46},
			},
		},
		// Global object aliases
		{Code: "const {Buffer: B} = globalThis; new B(); window.process.binding; self.require('fs').exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Globals:         map[string]any{"globalThis": "readonly", "window": "readonly", "self": "readonly"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 33, EndLine: 1, EndColumn: 40},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 42, EndLine: 1, EndColumn: 64},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 66, EndLine: 1, EndColumn: 91},
			},
		},
		// Import default aliases
		{Code: "import {default as fs} from 'fs'; fs.exists; import * as ns from 'buffer'; const {default:b} = ns; new b.Buffer();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 35, EndLine: 1, EndColumn: 44},
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 100, EndLine: 1, EndColumn: 114},
			},
		},
		// Whole-module imports report the declaration
		{Code: "import 'domain'; import * as x from 'node:sys'; import {default as y} from 'punycode';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				{MessageId: "deprecated", Message: "'sys' module was deprecated since v0.3.0. Use 'util' module instead.", Line: 1, Column: 18, EndLine: 1, EndColumn: 48},
				{MessageId: "deprecated", Message: "'punycode' module was deprecated since v7.0.0. Use 'https://www.npmjs.com/package/punycode' instead.", Line: 1, Column: 49, EndLine: 1, EndColumn: 87},
			},
		},
		// Named and namespace reexports
		{Code: "export {exists as old, lchmod} from 'fs'; export * from 'sys'; export * as legacy from 'buffer';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22},
				{MessageId: "deprecated", Message: "'fs.lchmod' was deprecated since v0.4.0.", Line: 1, Column: 24, EndLine: 1, EndColumn: 30},
				{MessageId: "deprecated", Message: "'sys' module was deprecated since v0.3.0. Use 'util' module instead.", Line: 1, Column: 43, EndLine: 1, EndColumn: 63},
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 64, EndLine: 1, EndColumn: 97},
			},
		},
		// getBuiltinModule aliases and constant arguments
		{Code: "const {getBuiltinModule: get} = process; get('f'+'s').exists; get(); get(name); const req = require; req('sys');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 42, EndLine: 1, EndColumn: 61},
				{MessageId: "deprecated", Message: "'sys' module was deprecated since v0.3.0. Use 'util' module instead.", Line: 1, Column: 102, EndLine: 1, EndColumn: 112},
			},
		},
		// Global read versus module ignore scopes
		{Code: "process.binding(); require('process').binding();",
			Options:         []any{map[string]any{"ignoreGlobalItems": []any{"process.binding"}}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46},
			},
		},
		// Module read versus global ignore scopes
		{Code: "process.binding(); require('process').binding();",
			Options:         []any{map[string]any{"ignoreModuleItems": []any{"process.binding"}}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
			},
		},
		// Multiline and UTF-16 ranges
		{Code: "const text = '😀';\nconst {\n  exists: found\n} = require('fs');\nrequire('fs')\n .exists;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 3, Column: 3, EndLine: 3, EndColumn: 16},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 5, Column: 1, EndLine: 6, EndColumn: 9},
			},
		},
		// TypeScript expressions and aliases
		{Code: "const B = require('buffer').Buffer as any; new B(); const fs = require('fs') satisfies object; fs!.exists; (Buffer as any)(1); (Buffer<number>)(2);",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 96, EndLine: 1, EndColumn: 106},
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 108, EndLine: 1, EndColumn: 126},
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 128, EndLine: 1, EndColumn: 147},
			},
		},
		// Type-only imports still report reads
		{Code: "import type {SlowBuffer} from 'buffer'; export type {exists} from 'fs'; type B = typeof Buffer;",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 54, EndLine: 1, EndColumn: 60},
			},
		},
		// Type-only declarations do not shadow value references
		{Code: "interface Buffer {} new Buffer(); type process = {}; process.binding();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 21, EndLine: 1, EndColumn: 33},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 54, EndLine: 1, EndColumn: 69},
			},
		},
		// JSX member tags are not runtime member expressions
		{Code: "const fs = require('fs'); const el = <fs.exists value={process.binding()} />;",
			Tsx:             true,
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 1, Column: 56, EndLine: 1, EndColumn: 71},
			},
		},
		// Explicit defaults and deprecated option
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=16.0.0", "ignoreModuleItems": []any{}, "ignoreGlobalItems": []any{}, "ignoreIndirectDependencies": false}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Options version precedes both settings
		{Code: "Buffer();",
			Options:         []any{map[string]any{"version": "4"}},
			Settings:        map[string]any{"n": map[string]any{"version": "6"}, "node": map[string]any{"version": "8"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Invalid options version falls through to settings.n
		{Code: "Buffer();",
			Options:         []any{map[string]any{"version": "invalid"}},
			Settings:        map[string]any{"n": map[string]any{"version": "4"}, "node": map[string]any{"version": "8"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Invalid n setting falls through to node
		{Code: "Buffer();",
			Settings:        map[string]any{"n": map[string]any{"version": "invalid"}, "node": map[string]any{"version": "5.10"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Non-string setting is converted
		{Code: "Buffer();",
			Settings:        map[string]any{"node": map[string]any{"version": float64(4)}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Absent version and empty options use fallback
		{Code: "require('fs').exists;",
			Options:         []any{map[string]any{}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Version range *
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "*"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ""}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range invalid
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "invalid"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >=0.0.1
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=0.0.1"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >=0.0.2 <0.11.15
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=0.0.2 <0.11.15"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >=0.11.15
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=0.11.15"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range ^5.9.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "^5.9.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range ^5.10.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "^5.10.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range ~5.10
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "~5.10"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range 5.10.x
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "5.10.x"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >= 5.10.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">= 5.10.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range v5.10.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "v5.10.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range 5.10.0 || 4.0.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "5.10.0 || 4.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range 5.10.0 - 6
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "5.10.0 - 6"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >5.9.9
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">5.9.9"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range <0.0.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "<0.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >=10 <5
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=10 <5"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range 5.10.0-beta.1
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "5.10.0-beta.1"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >=5.10.0-beta.1
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=5.10.0-beta.1"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range ^0.0.1
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "^0.0.1"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range ~>5.10.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "~>5.10.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >=5.10.0+build
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">=5.10.0+build"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range 6 || *
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": "6 || *"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Version range >5.10.0 <5.10.0
		{Code: "Buffer(); require('fs').exists;",
			Options:         []any{map[string]any{"version": ">5.10.0 <5.10.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
			},
		},
		// Remaining static APIs from upstream metadata
		{Code: "require(\"_stream_wrap\");\nrequire(\"crypto\")._toBuf;\nrequire(\"crypto\").DEFAULT_ENCODING;\nrequire(\"crypto\").createCipher;\nrequire(\"crypto\").createDecipher;\nrequire(\"crypto\").fips;\nrequire(\"crypto\").prng;\nrequire(\"crypto\").pseudoRandomBytes;\nrequire(\"crypto\").rng;\nrequire(\"module\").Module.createRequireFromPath;\nrequire(\"net\")._setSimultaneousAccepts;\nrequire(\"process\").assert;\nrequire(\"process\").binding;\nrequire(\"process\").report.triggerReport;\nrequire(\"repl\").REPLServer;\nrequire(\"repl\").Recoverable;\nrequire(\"repl\").REPL_MODE_MAGIC;\nrequire(\"repl\").builtinModules;\nnew (require(\"safe-buffer\").Buffer)();\nrequire(\"safe-buffer\").Buffer();\nrequire(\"safe-buffer\").SlowBuffer;\nrequire(\"timers\").enroll;\nrequire(\"timers\").unenroll;\nrequire(\"tls\").convertNPNProtocols;\nrequire(\"url\").parse;\nrequire(\"url\").resolve;\nrequire(\"zlib\").BrotliCompress();\nrequire(\"zlib\").BrotliDecompress();\nrequire(\"zlib\").Deflate();\nrequire(\"zlib\").DeflateRaw();\nrequire(\"zlib\").Gunzip();\nrequire(\"zlib\").Gzip();\nrequire(\"zlib\").Inflate();\nrequire(\"zlib\").InflateRaw();\nrequire(\"zlib\").Unzip();\nCOUNTER_NET_SERVER_CONNECTION;\nCOUNTER_NET_SERVER_CONNECTION_CLOSE;\nCOUNTER_HTTP_SERVER_REQUEST;\nCOUNTER_HTTP_SERVER_RESPONSE;\nCOUNTER_HTTP_CLIENT_REQUEST;\nCOUNTER_HTTP_CLIENT_RESPONSE;\nprocess.assert;\nprocess.binding;\nprocess.report.triggerReport;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Globals:         map[string]any{"COUNTER_NET_SERVER_CONNECTION": "readonly", "COUNTER_NET_SERVER_CONNECTION_CLOSE": "readonly", "COUNTER_HTTP_SERVER_REQUEST": "readonly", "COUNTER_HTTP_SERVER_RESPONSE": "readonly", "COUNTER_HTTP_CLIENT_REQUEST": "readonly", "COUNTER_HTTP_CLIENT_RESPONSE": "readonly", "process": "readonly"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'_stream_wrap' module was deprecated since v12.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				{MessageId: "deprecated", Message: "'crypto._toBuf' was deprecated since v11.0.0.", Line: 2, Column: 1, EndLine: 2, EndColumn: 25},
				{MessageId: "deprecated", Message: "'crypto.DEFAULT_ENCODING' was deprecated since v10.0.0.", Line: 3, Column: 1, EndLine: 3, EndColumn: 35},
				{MessageId: "deprecated", Message: "'crypto.createCipher' was deprecated since v10.0.0. Use 'crypto.createCipheriv()' instead.", Line: 4, Column: 1, EndLine: 4, EndColumn: 31},
				{MessageId: "deprecated", Message: "'crypto.createDecipher' was deprecated since v10.0.0. Use 'crypto.createDecipheriv()' instead.", Line: 5, Column: 1, EndLine: 5, EndColumn: 33},
				{MessageId: "deprecated", Message: "'crypto.fips' was deprecated since v10.0.0. Use 'crypto.getFips()' and 'crypto.setFips()' instead.", Line: 6, Column: 1, EndLine: 6, EndColumn: 23},
				{MessageId: "deprecated", Message: "'crypto.prng' was deprecated since v11.0.0. Use 'crypto.randomBytes()' instead.", Line: 7, Column: 1, EndLine: 7, EndColumn: 23},
				{MessageId: "deprecated", Message: "'crypto.pseudoRandomBytes' was deprecated since v11.0.0. Use 'crypto.randomBytes()' instead.", Line: 8, Column: 1, EndLine: 8, EndColumn: 36},
				{MessageId: "deprecated", Message: "'crypto.rng' was deprecated since v11.0.0. Use 'crypto.randomBytes()' instead.", Line: 9, Column: 1, EndLine: 9, EndColumn: 22},
				{MessageId: "deprecated", Message: "'module.Module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.", Line: 10, Column: 1, EndLine: 10, EndColumn: 47},
				{MessageId: "deprecated", Message: "'net._setSimultaneousAccepts' was deprecated since v12.0.0.", Line: 11, Column: 1, EndLine: 11, EndColumn: 39},
				{MessageId: "deprecated", Message: "'process.assert' was deprecated since v10.0.0. Use 'require(\"assert\")' instead.", Line: 12, Column: 1, EndLine: 12, EndColumn: 26},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 13, Column: 1, EndLine: 13, EndColumn: 27},
				{MessageId: "deprecated", Message: "'process.report.triggerReport' was deprecated since v11.12.0. Use 'process.report.writeReport()' instead.", Line: 14, Column: 1, EndLine: 14, EndColumn: 40},
				{MessageId: "deprecated", Message: "'repl.REPLServer' was deprecated since v22.9.0. Use new repl.REPLServer() instead.", Line: 15, Column: 1, EndLine: 15, EndColumn: 27},
				{MessageId: "deprecated", Message: "'repl.Recoverable' was deprecated since v22.9.0. Use new repl.Recoverable() instead.", Line: 16, Column: 1, EndLine: 16, EndColumn: 28},
				{MessageId: "deprecated", Message: "'repl.REPL_MODE_MAGIC' was deprecated since v8.0.0.", Line: 17, Column: 1, EndLine: 17, EndColumn: 32},
				{MessageId: "deprecated", Message: "'repl.builtinModules' was deprecated since v22.16.0. Use module.builtinModules instead.", Line: 18, Column: 1, EndLine: 18, EndColumn: 31},
				{MessageId: "deprecated", Message: "'new safe-buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 19, Column: 1, EndLine: 19, EndColumn: 38},
				{MessageId: "deprecated", Message: "'safe-buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 20, Column: 1, EndLine: 20, EndColumn: 32},
				{MessageId: "deprecated", Message: "'safe-buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 21, Column: 1, EndLine: 21, EndColumn: 34},
				{MessageId: "deprecated", Message: "'timers.enroll' was deprecated since v10.0.0. Use 'setTimeout()' or 'setInterval()' instead.", Line: 22, Column: 1, EndLine: 22, EndColumn: 25},
				{MessageId: "deprecated", Message: "'timers.unenroll' was deprecated since v10.0.0. Use 'clearTimeout()' or 'clearInterval()' instead.", Line: 23, Column: 1, EndLine: 23, EndColumn: 27},
				{MessageId: "deprecated", Message: "'tls.convertNPNProtocols' was deprecated since v10.0.0.", Line: 24, Column: 1, EndLine: 24, EndColumn: 35},
				{MessageId: "deprecated", Message: "'url.parse' was deprecated since v11.0.0. Use 'url.URL' constructor instead.", Line: 25, Column: 1, EndLine: 25, EndColumn: 21},
				{MessageId: "deprecated", Message: "'url.resolve' was deprecated since v11.0.0. Use 'url.URL' constructor instead.", Line: 26, Column: 1, EndLine: 26, EndColumn: 23},
				{MessageId: "deprecated", Message: "'zlib.BrotliCompress()' was deprecated since v22.9.0. Use new zlib.BrotliCompress() instead.", Line: 27, Column: 1, EndLine: 27, EndColumn: 33},
				{MessageId: "deprecated", Message: "'zlib.BrotliDecompress()' was deprecated since v22.9.0. Use new zlib.BrotliDecompress() instead.", Line: 28, Column: 1, EndLine: 28, EndColumn: 35},
				{MessageId: "deprecated", Message: "'zlib.Deflate()' was deprecated since v22.9.0. Use new zlib.Deflate() instead.", Line: 29, Column: 1, EndLine: 29, EndColumn: 26},
				{MessageId: "deprecated", Message: "'zlib.DeflateRaw()' was deprecated since v22.9.0. Use new zlib.DeflateRaw() instead.", Line: 30, Column: 1, EndLine: 30, EndColumn: 29},
				{MessageId: "deprecated", Message: "'zlib.Gunzip()' was deprecated since v22.9.0. Use new zlib.Gunzip() instead.", Line: 31, Column: 1, EndLine: 31, EndColumn: 25},
				{MessageId: "deprecated", Message: "'zlib.Gzip()' was deprecated since v22.9.0. Use new zlib.Gzip() instead.", Line: 32, Column: 1, EndLine: 32, EndColumn: 23},
				{MessageId: "deprecated", Message: "'zlib.Inflate()' was deprecated since v22.9.0. Use new zlib.Inflate() instead.", Line: 33, Column: 1, EndLine: 33, EndColumn: 26},
				{MessageId: "deprecated", Message: "'zlib.InflateRaw()' was deprecated since v22.9.0. Use new zlib.InflateRaw() instead.", Line: 34, Column: 1, EndLine: 34, EndColumn: 29},
				{MessageId: "deprecated", Message: "'zlib.Unzip()' was deprecated since v22.9.0. Use new zlib.Unzip() instead.", Line: 35, Column: 1, EndLine: 35, EndColumn: 24},
				{MessageId: "deprecated", Message: "'COUNTER_NET_SERVER_CONNECTION' was deprecated since v11.0.0.", Line: 36, Column: 1, EndLine: 36, EndColumn: 30},
				{MessageId: "deprecated", Message: "'COUNTER_NET_SERVER_CONNECTION_CLOSE' was deprecated since v11.0.0.", Line: 37, Column: 1, EndLine: 37, EndColumn: 36},
				{MessageId: "deprecated", Message: "'COUNTER_HTTP_SERVER_REQUEST' was deprecated since v11.0.0.", Line: 38, Column: 1, EndLine: 38, EndColumn: 28},
				{MessageId: "deprecated", Message: "'COUNTER_HTTP_SERVER_RESPONSE' was deprecated since v11.0.0.", Line: 39, Column: 1, EndLine: 39, EndColumn: 29},
				{MessageId: "deprecated", Message: "'COUNTER_HTTP_CLIENT_REQUEST' was deprecated since v11.0.0.", Line: 40, Column: 1, EndLine: 40, EndColumn: 28},
				{MessageId: "deprecated", Message: "'COUNTER_HTTP_CLIENT_RESPONSE' was deprecated since v11.0.0.", Line: 41, Column: 1, EndLine: 41, EndColumn: 29},
				{MessageId: "deprecated", Message: "'process.assert' was deprecated since v10.0.0. Use 'require(\"assert\")' instead.", Line: 42, Column: 1, EndLine: 42, EndColumn: 15},
				{MessageId: "deprecated", Message: "'process.binding' was deprecated since v10.9.0.", Line: 43, Column: 1, EndLine: 43, EndColumn: 16},
				{MessageId: "deprecated", Message: "'process.report.triggerReport' was deprecated since v11.12.0. Use 'process.report.writeReport()' instead.", Line: 44, Column: 1, EndLine: 44, EndColumn: 29},
			},
		},
	}

	for i := range valid {
		valid[i].Globals = deprecatedTestGlobals(valid[i].Globals)
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "commonjs"
		}
	}
	for i := range invalid {
		invalid[i].Globals = deprecatedTestGlobals(invalid[i].Globals)
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "commonjs"
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedAPIRule, valid, invalid)
}

func TestNoDeprecatedAPIVersionMetadata(t *testing.T) {
	base := fixtures.GetRootDir()
	dir := tspath.ResolvePath(base.Dir, "deprecated-version-fixtures")
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
	for _, tc := range []struct {
		file     string
		replace  bool
		options  any
		settings map[string]any
	}{
		{file: "input.js", replace: true},
		{file: "old/input.js"},
		{file: "engines-first/input.js", replace: true},
		{file: "dev-object/input.js"},
		{file: "dev-array/input.js"},
		{file: "dev-invalid-first/input.js", replace: true},
		{file: "empty-range/input.js"},
		{file: "invalid/input.js", replace: true},
		{file: "malformed/input.js", replace: true},
		{file: "old/nearest/input.js", replace: true},
		{file: "old/option.js", replace: true, options: []any{map[string]any{"version": "6"}}},
		{file: "old/settings.js", replace: true, settings: map[string]any{"node": map[string]any{"version": "6"}}},
		{file: "old/zero-setting.js", settings: map[string]any{"node": map[string]any{"version": float64(0)}}},
		{file: "old/invalid-option.js", options: []any{map[string]any{"version": "invalid"}}},
		{file: "input.js", options: []any{map[string]any{"version": ">=5.10.0-1beta"}}},
		{file: "input.js", options: []any{map[string]any{"version": "<=4294967296"}}},
		{file: "input.js", options: []any{map[string]any{"version": "^0.0.4294967295"}}},
		{file: "input.js", replace: true, options: []any{map[string]any{"version": ">4294967295"}}},
		{file: "input.js", replace: true, options: []any{map[string]any{"version": "5.x.1"}}},
	} {
		message := "'Buffer()' was deprecated since v6.0.0"
		if tc.replace {
			message += ". Use 'Buffer.alloc()' or 'Buffer.from()' instead"
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: "Buffer();", FileName: tc.file, Globals: deprecatedTestGlobals(nil), Options: tc.options, Settings: tc.settings,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: message + ".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}},
		})
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoDeprecatedAPIRule, nil, invalid)
}

func TestNoDeprecatedAPISchema(t *testing.T) {
	for _, options := range [][]any{
		[]any{map[string]any{"version": 6}},
		[]any{map[string]any{"ignoreModuleItems": []any{"node:fs.exists"}}},
		[]any{map[string]any{"ignoreGlobalItems": []any{"Buffer"}}},
		[]any{map[string]any{"ignoreGlobalItems": []any{"Buffer()", "Buffer()"}}},
		[]any{map[string]any{"unknown": true}},
		[]any{map[string]any{}, map[string]any{}},
	} {
		if err := NoDeprecatedAPIRule.Schema.Validate(options); err == nil {
			t.Errorf("expected invalid options: %#v", options)
		}
	}
}

func TestNoDeprecatedAPIExportAllOrder(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedAPIRule, nil, []rule_tester.InvalidTestCase{{
		Code: "export * from 'crypto';", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "deprecated", Message: "'crypto._toBuf' was deprecated since v11.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.Credentials' was deprecated since v0.12.0. Use 'tls.SecureContext' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.DEFAULT_ENCODING' was deprecated since v10.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.createCipher' was deprecated since v10.0.0. Use 'crypto.createCipheriv()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.createCredentials' was deprecated since v0.12.0. Use 'tls.createSecureContext()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.createDecipher' was deprecated since v10.0.0. Use 'crypto.createDecipheriv()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.fips' was deprecated since v10.0.0. Use 'crypto.getFips()' and 'crypto.setFips()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.prng' was deprecated since v11.0.0. Use 'crypto.randomBytes()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.pseudoRandomBytes' was deprecated since v11.0.0. Use 'crypto.randomBytes()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			{MessageId: "deprecated", Message: "'crypto.rng' was deprecated since v11.0.0. Use 'crypto.randomBytes()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		},
	}})
}
