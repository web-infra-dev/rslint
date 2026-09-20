package buffer_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_global/buffer"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional reference-tracking and AST cases verified against eslint-plugin-n v18.3.0.
func TestPreferGlobalBufferExtras(t *testing.T) {
	runBufferTests(t, []rule_tester.ValidTestCase{
		// Other modules and exports do not match.
		{
			Code:     "require('safe-buffer').Buffer; require('buffer/').Buffer; require('node:safe-buffer').Buffer; require('buffer').SlowBuffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// Local loaders shadow the configured globals.
		{
			Code:     "function f(require, process) { require('buffer').Buffer; process.getBuiltinModule('buffer').Buffer; }",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// Writing a global loader disables its tracking throughout the file.
		{
			Code:     "require('buffer').Buffer; require = other; process.getBuiltinModule('buffer').Buffer; process = other;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// Dynamic module names and computed properties are not resolved through variables.
		{
			Code:     "const name = 'buffer', key = 'Buffer'; require(name).Buffer; require('buffer')[key]; require(); process.getBuiltinModule(); require(...['buffer']).Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// Side effects, unused namespaces and dynamic imports do not read Buffer.
		{
			Code:            "import 'buffer'; import * as b from 'buffer'; import('buffer').then(b => b.Buffer);",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// TypeScript import-equals and type queries are outside runtime reference tracking.
		{
			Code:            "import b = require('buffer'); type T = typeof b.Buffer; b.Buffer.alloc(1);",
			FileName:        "input.ts",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Rest and array patterns are not followed.
		{
			Code:     "const { ...b } = require('buffer'); b.Buffer; const [B] = require('buffer'); B;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// Private members are not public Buffer properties.
		{
			Code:     "class C { #Buffer; m() { require('buffer').#Buffer; } }",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// JSX member tags are not ordinary property accesses.
		{
			Code:     "const b = require('buffer'); const el = <b.Buffer />;",
			FileName: "input.tsx",
		},
		// Disabled loaders are ignored.
		{
			Code:     "require('buffer').Buffer; process.getBuiltinModule('buffer').Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Globals:  map[string]any{"require": "off", "process": "off"},
		},
		// Local Buffer declarations and imports are allowed in never mode.
		{
			Code:            "import { Buffer } from 'node:buffer'; Buffer.alloc(1); function f(Buffer) { return Buffer; }",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Writes disable tracking of the global Buffer binding.
		{
			Code:     "Buffer.alloc(1); Buffer = Other;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
		},
		// Disabled global Buffer is ignored.
		{
			Code:     "Buffer.alloc(1);",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Globals:  map[string]any{"Buffer": "off"},
		},
		// TypeScript type queries and authored ambient declarations do not read a global.
		{
			Code:            "type B = typeof Buffer; declare const Buffer: any; Buffer.alloc(1);",
			FileName:        "input.ts",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Constructors and tagged templates are not module loader calls.
		{
			Code:     "(new require('buffer')).Buffer; require`buffer`.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
		},
		// Script declarations shadow configured globals.
		{
			Code:            "var Buffer; Buffer.alloc(1);",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
		},
		// Namespace and ambient declarations shadow configured globals.
		{
			Code:            "namespace Buffer { export const alloc = other; } Buffer.alloc(1); declare const process: any; process.getBuiltinModule('buffer').Buffer;",
			FileName:        "input.ts",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
	}, []rule_tester.InvalidTestCase{
		// Direct CommonJS reads include optional and parenthesized access.
		{
			Code:     "require('buffer').Buffer; (require('node:buffer'))?.['Buffer'];",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 1, 1, 25),
				preferBufferAt("preferGlobal", 1, 27, 1, 63),
			},
		},
		// Computed constant module names and keys use JavaScript values.
		{
			Code:     "require(\"buf\" + \"fer\")[`Buf${\"fer\"}`];",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 1, 1, 38),
			},
		},
		// Module aliases are followed through assignments and default destructuring.
		{
			Code:     "const b = require('buffer'); let alias; alias = b; const { Buffer: B = Other } = alias; alias.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 60, 1, 77),
				preferBufferAt("preferGlobal", 1, 89, 1, 101),
			},
		},
		// Destructuring assignment reports the property including its default.
		{
			Code:     "let B; ({ Buffer: B = Other } = require('buffer'));",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 11, 1, 28),
			},
		},
		// Loader aliases and global-object loaders retain their module identity.
		{
			Code:     "const load = require; load('buffer').Buffer; const { getBuiltinModule: get } = process; get('node:buffer').Buffer; global.require('buffer').Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 23, 1, 44),
				preferBufferAt("preferGlobal", 1, 89, 1, 114),
				preferBufferAt("preferGlobal", 1, 116, 1, 147),
			},
		},
		// ESM named, default and namespace imports use legacy CommonJS exports.
		{
			Code:            "import { Buffer as B } from 'buffer'; import b from 'node:buffer'; import * as ns from 'buffer'; b.Buffer; ns.Buffer; ns.default.Buffer;",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 10, 1, 21),
				preferBufferAt("preferGlobal", 1, 98, 1, 106),
				preferBufferAt("preferGlobal", 1, 108, 1, 117),
				preferBufferAt("preferGlobal", 1, 119, 1, 136),
			},
		},
		// ESM re-exports report specifiers and export-all declarations.
		{
			Code:            "export { Buffer as B } from 'buffer'; export * from 'node:buffer'; export * as b from 'buffer';",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 10, 1, 21),
				preferBufferAt("preferGlobal", 1, 39, 1, 67),
				preferBufferAt("preferGlobal", 1, 68, 1, 96),
			},
		},
		// TypeScript assertions on module expressions remain tracked.
		{
			Code:            "const b = (require('buffer') as any); b!.Buffer; (require('buffer') satisfies object).Buffer;",
			FileName:        "input.ts",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 39, 1, 48),
				preferBufferAt("preferGlobal", 1, 50, 1, 93),
			},
		},
		// Type-only named imports follow the upstream import-specifier behavior.
		{
			Code:            "import type { Buffer } from 'buffer'; export type { Buffer as B } from 'node:buffer';",
			FileName:        "input.ts",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 15, 1, 21),
				preferBufferAt("preferGlobal", 1, 53, 1, 64),
			},
		},
		// JSX expression containers read Buffer properties.
		{
			Code:     "const b = require('buffer'); const el = <b.Buffer value={b.Buffer} />;",
			FileName: "input.tsx",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 58, 1, 66),
			},
		},
		// Ranges use UTF-16 columns and cover multiline member access.
		{
			Code:     "\"😀\"; require('buffer')\n  .Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 7, 2, 10),
			},
		},
		// Global Buffer calls, construction, typeof and shorthand are reads.
		{
			Code:     "Buffer.alloc(1); new Buffer(1); typeof Buffer; const obj = { Buffer };",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 1, 1, 7),
				preferBufferAt("preferModule", 1, 22, 1, 28),
				preferBufferAt("preferModule", 1, 40, 1, 46),
				preferBufferAt("preferModule", 1, 62, 1, 68),
			},
		},
		// Global-object properties are read even when global Buffer is disabled.
		{
			Code:     "global.Buffer; globalThis['Buffer']; window.Buffer; self.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Globals:  map[string]any{"Buffer": "off", "window": "readonly", "self": "readonly"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 1, 1, 14),
				preferBufferAt("preferModule", 1, 16, 1, 36),
				preferBufferAt("preferModule", 1, 38, 1, 51),
				preferBufferAt("preferModule", 1, 53, 1, 64),
			},
		},
		// Destructuring and aliases of global objects are followed.
		{
			Code:     "const g = globalThis; const { Buffer: B } = g; B.alloc(1); const C = Buffer; C.alloc(1);",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 31, 1, 40),
				preferBufferAt("preferModule", 1, 70, 1, 76),
			},
		},
		// Member writes are reference reads while direct binding writes disable tracking.
		{
			Code:     "global.Buffer = Other; delete globalThis.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 1, 1, 14),
				preferBufferAt("preferModule", 1, 31, 1, 48),
			},
		},
		// Local shadowing is restricted to its own scope.
		{
			Code:     "function f(Buffer) { return Buffer; } Buffer.alloc(1);",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 39, 1, 45),
			},
		},
		// Inline globals override configuration.
		{
			Code:     "/* global Buffer: readonly */ Buffer.alloc(1);",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Globals:  map[string]any{"Buffer": "off"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 31, 1, 37),
			},
		},
		// Parentheses and optional chains preserve the original reporting ranges.
		{
			Code:     "((Buffer))?.alloc(1); (globalThis)?.[(\"Buffer\")];",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 3, 1, 9),
				preferBufferAt("preferModule", 1, 23, 1, 49),
			},
		},
		// JavaScript JSDoc casts keep runtime module references.
		{
			Code:     "const b = /** @type {any} */ (require('buffer')); b.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 51, 1, 59),
			},
		},
		// JSX Buffer names and expression containers respect scope references.
		{
			Code:     "const el = <Buffer value={Buffer} />;",
			FileName: "input.tsx",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 13, 1, 19),
				preferBufferAt("preferModule", 1, 27, 1, 33),
			},
		},
		// Conditional module origins preserve duplicate reports.
		{
			Code:     "const b = flag ? require('buffer') : require('node:buffer'); b.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 62, 1, 70),
				preferBufferAt("preferGlobal", 1, 62, 1, 70),
			},
		},
		// Logical and sequence expressions preserve only flowing values.
		{
			Code:     "(other || require('buffer')).Buffer; (other ?? require('buffer')).Buffer; (0, require('buffer')).Buffer; (require('buffer'), other).Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 1, 1, 36),
				preferBufferAt("preferGlobal", 1, 38, 1, 73),
				preferBufferAt("preferGlobal", 1, 75, 1, 104),
			},
		},
		// Cyclic aliases stop without dropping reachable module reads.
		{
			Code:     "let a = require('buffer'); let b = a; a = b; b.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 46, 1, 54),
			},
		},
		// Parameter defaults retain module origins in nested scopes.
		{
			Code:     "function f(b = require('buffer')) { return () => b.Buffer; }",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 50, 1, 58),
			},
		},
		// Module property writes report the original Buffer access.
		{
			Code:     "const b = require('buffer'); b.Buffer = Other; delete b.Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 30, 1, 38),
				preferBufferAt("preferGlobal", 1, 55, 1, 63),
			},
		},
		// Escaped global identifiers and computed keys retain their source ranges.
		{
			Code:     "Buff\\u0065r.alloc(1); globalThis[\"Buff\\u0065r\"];",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 1, 1, 12),
				preferBufferAt("preferModule", 1, 23, 1, 48),
			},
		},
		// Type-only names and class heritage retain upstream scope semantics.
		{
			Code:            "type B = typeof Buffer; type I = Buffer; interface X extends Buffer {} class C implements Buffer {} class D extends Buffer {}",
			FileName:        "input.ts",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 17, 1, 23),
				preferBufferAt("preferModule", 1, 34, 1, 40),
				preferBufferAt("preferModule", 1, 62, 1, 68),
				preferBufferAt("preferModule", 1, 91, 1, 97),
				preferBufferAt("preferModule", 1, 117, 1, 123),
			},
		},
		// Type parameters do not hide value-space globals.
		{
			Code:            "function f<Buffer>() { return Buffer.alloc(1); }",
			FileName:        "input.ts",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 31, 1, 37),
			},
		},
		// Script module destructuring still reports the imported property.
		{
			Code:            "var {Buffer} = require('buffer'); Buffer.alloc(1);",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 6, 1, 12),
			},
		},
		// Loop and catch bindings stay local to their scopes.
		{
			Code:     "for (const Buffer of values) Buffer.alloc(1); try {} catch (Buffer) { Buffer.alloc(1); } Buffer.alloc(1);",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 90, 1, 96),
			},
		},
		// Function-local bindings do not hide global-object properties.
		{
			Code:     "function f(Buffer) { return globalThis.Buffer; }",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferModule", 1, 29, 1, 46),
			},
		},
		// Computed destructuring keys and string-named imports keep full ranges.
		{
			Code:            "import { 'Buffer' as B } from 'node:buffer'; const { ['Buf' + 'fer']: C } = require('buffer');",
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 1, 10, 1, 23),
				preferBufferAt("preferGlobal", 1, 54, 1, 72),
			},
		},
		// Line and block directives suppress only selected statements.
		{
			Code:     "// eslint-disable-next-line\nrequire('buffer').Buffer;\n/* eslint-disable */ require('buffer').Buffer;\n/* eslint-enable */ require('buffer').Buffer;",
			FileName: "input.js",
			TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				preferBufferAt("preferGlobal", 4, 21, 4, 45),
			},
		},
	})
}

func TestPreferGlobalBufferSchema(t *testing.T) {
	for _, options := range [][]any{{"sometimes"}, {true}, {nil}, {map[string]any{}}, {"always", "never"}} {
		if err := buffer.PreferGlobalBufferRule.Schema.Validate(options); err == nil {
			t.Errorf("expected invalid options: %#v", options)
		}
	}
}
