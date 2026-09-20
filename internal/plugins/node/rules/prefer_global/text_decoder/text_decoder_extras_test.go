package text_decoder_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional cases compared with eslint-plugin-n v18.3.0 and @typescript-eslint/parser v8.65.0.
func TestTextDecoderExtras(t *testing.T) {
	runTextDecoderTests(t,
		[]rule_tester.ValidTestCase{
			// Other modules and exports do not match.
			{
				Code: "require('other').TextDecoder; require('util').TextEncoder; require('util'); process.getBuiltinModule('url').TextDecoder;",
			},
			// Dynamic module names and properties are not followed.
			{
				Code: "require(moduleName).TextDecoder; require('util')[key]; process.getBuiltinModule(moduleName).TextDecoder;",
			},
			// Shadowed loaders are ignored.
			{
				Code: "function f(require, process) { require('util').TextDecoder; process.getBuiltinModule('util').TextDecoder; }",
			},
			// Disabled loaders are ignored.
			{
				Code:    "require('util').TextDecoder; process.getBuiltinModule('util').TextDecoder;",
				Globals: map[string]any{"require": "off", "process": "off"},
			},
			// Reassigned global loaders are ignored throughout the file.
			{
				Code: "require('util').TextDecoder; require = other; process.getBuiltinModule('util').TextDecoder; process = other;",
			},
			// Imported process and dynamic imports do not root tracking.
			{
				Code: "import process from 'node:process'; process.getBuiltinModule('util').TextDecoder; const util = await import('node:util'); util.TextDecoder;",
			},
			// Local TextDecoder bindings shadow the global.
			{
				Code:    "function f(TextDecoder) { return new TextDecoder(); } { const TextDecoder = Custom; new TextDecoder(); }",
				Options: []any{"never"},
			},
			// A disabled global is ignored.
			{
				Code:    "new TextDecoder();",
				Options: []any{"never"},
				Globals: map[string]any{"TextDecoder": "off"},
			},
			// A global write disables tracking throughout the file.
			{
				Code:    "new TextDecoder(); TextDecoder = Custom;",
				Options: []any{"never"},
			},
			// Rest bindings and private properties do not match.
			{
				Code: "const {...util} = require('util'); util.TextDecoder; class C { #TextDecoder; m() { require('util').#TextDecoder; } }",
			},
			// JSX tag names do not read tracked properties.
			{
				Code:     "import util from 'node:util'; const el = <util.TextDecoder />;",
				FileName: "input.tsx",
			},
			// Bare property keys are not global reads.
			{
				Code:    "const obj = {TextDecoder: 1}; obj.TextDecoder;",
				Options: []any{"never"},
			},
			// Computed property identifiers are not resolved from their declarations.
			{
				Code: "const key = 'Text' + 'Decoder'; const util = require('util'); const alias = util; const TD = alias[key]; new TD();",
			},
		},
		[]rule_tester.InvalidTestCase{
			// A value declaration does not shadow a reference in the type namespace.
			{
				Code:     "type Decoder = typeof TextDecoder; interface I { decoder: TextDecoder } declare const TextDecoder: any; new TextDecoder();",
				Options:  []any{"never"},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 59, 1, 70),
				},
			},
			// Parenthesized optional and computed module reads retain their ranges.
			{
				Code: "(require('util'))?.TextDecoder; require('node:util')[('TextDecoder')];",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 1, 1, 31),
					textDecoderAt("preferGlobal", 1, 33, 1, 70),
				},
			},
			// Destructuring defaults report the whole property.
			{
				Code: "const {TextDecoder: TD = fallback} = process.getBuiltinModule('node:util'); new TD();",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 8, 1, 34),
				},
			},
			// Destructuring assignment and parameter defaults are followed.
			{
				Code: "let TD; ({TextDecoder: TD} = require('util')); function f({TextDecoder: D} = require('util')) { return D; }",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 11, 1, 26),
					textDecoderAt("preferGlobal", 1, 60, 1, 74),
				},
			},
			// Named imports report the specifier including its alias.
			{
				Code: "import { TextDecoder as TD } from 'node:util'; new TD();",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 10, 1, 27),
				},
			},
			// Default and namespace imports expose the CommonJS exports.
			{
				Code: "import util from 'util'; import * as ns from 'node:util'; util.TextDecoder; ns.TextDecoder; ns.default.TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 59, 1, 75),
					textDecoderAt("preferGlobal", 1, 77, 1, 91),
					textDecoderAt("preferGlobal", 1, 93, 1, 115),
				},
			},
			// Re-exports report their specifier or whole export-all statement.
			{
				Code: "export {TextDecoder as Decoder} from 'util'; export * from 'node:util'; export * as util from 'util';",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 9, 1, 31),
					textDecoderAt("preferGlobal", 1, 46, 1, 72),
					textDecoderAt("preferGlobal", 1, 73, 1, 102),
				},
			},
			// Global object aliases and nested bindings are followed.
			{
				Code:    "const root = globalThis; root.TextDecoder; const {TextDecoder: TD} = global; new TD();",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 26, 1, 42),
					textDecoderAt("preferModule", 1, 51, 1, 66),
				},
			},
			// Global object properties are checked even when the bare global is disabled.
			{
				Code:    "globalThis.TextDecoder;",
				Options: []any{"never"},
				Globals: map[string]any{"TextDecoder": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 1, 1, 23),
				},
			},
			// Read reports stop at aliases of the constructor.
			{
				Code:    "const TD = TextDecoder; new TD(); TextDecoder;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 12, 1, 23),
					textDecoderAt("preferModule", 1, 35, 1, 46),
				},
			},
			// Global object writes are reference-tracker reads.
			{
				Code:    "global.TextDecoder = Custom; delete globalThis.TextDecoder;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 1, 1, 19),
					textDecoderAt("preferModule", 1, 37, 1, 59),
				},
			},
			// JSX component names and expression containers read the global.
			{
				Code:     "const el = <TextDecoder value={TextDecoder}>{new TextDecoder()}</TextDecoder>;",
				Options:  []any{"never"},
				FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 13, 1, 24),
					textDecoderAt("preferModule", 1, 32, 1, 43),
					textDecoderAt("preferModule", 1, 50, 1, 61),
					textDecoderAt("preferModule", 1, 66, 1, 77),
				},
			},
			// Type assertions are transparent to module reference tracking.
			{
				Code:     "(require('util') as any).TextDecoder; require('util')!.TextDecoder; (require('util') satisfies any).TextDecoder;",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 1, 1, 37),
					textDecoderAt("preferGlobal", 1, 39, 1, 67),
					textDecoderAt("preferGlobal", 1, 69, 1, 112),
				},
			},
			// JSDoc casts preserve the module reference.
			{
				Code: "(/** @type {any} */ (require('util'))).TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 1, 1, 51),
				},
			},
			// Unicode and multiline reads use UTF-16 columns.
			{
				Code:    "\"😀\"; require('util')\n  .TextDecoder;",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 7, 2, 15),
				},
			},
			// Global reads preserve multiline positions.
			{
				Code:    "\"😀\"; TextDecoder;\nglobalThis\n  ['TextDecoder'];",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 7, 1, 18),
					textDecoderAt("preferModule", 2, 1, 3, 18),
				},
			},
			// Configured window and self objects are followed.
			{
				Code:    "window.TextDecoder; self['TextDecoder'];",
				Options: []any{"never"},
				Globals: map[string]any{"window": "readonly", "self": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 1, 1, 19),
					textDecoderAt("preferModule", 1, 21, 1, 40),
				},
			},
			// Inline global declarations override disabled configuration.
			{
				Code:    "/* global TextDecoder */ new TextDecoder();",
				Options: []any{"never"},
				Globals: map[string]any{"TextDecoder": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 30, 1, 41),
				},
			},
			// Aliased loaders and static module names work.
			{
				Code: "const load = require; load('node:' + 'util').TextDecoder; const p = process; p.getBuiltinModule('util').TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 23, 1, 57),
					textDecoderAt("preferGlobal", 1, 78, 1, 116),
				},
			},
		},
	)
}

// Further reference and AST cases compared with eslint-plugin-n v18.3.0.
func TestTextDecoderReferenceEdges(t *testing.T) {
	runTextDecoderTests(t,
		[]rule_tester.ValidTestCase{
			// Empty and spread module arguments are not statically known.
			{
				Code: "require(); process.getBuiltinModule(); require(...['util']).TextDecoder; process.getBuiltinModule(...['util']).TextDecoder;",
			},
			// Import-equals and type namespace members do not create ESM reads.
			{
				Code:     "import util = require('node:util'); util.TextDecoder; import type * as types from 'util'; type D = typeof types.TextDecoder;",
				FileName: "review.ts",
			},
			// Script declarations shadow the configured global.
			{
				Code:            "var TextDecoder; new TextDecoder();",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Cyclic aliases terminate without losing the tracked read.
			{
				Code: "let a = require('util'); let b = a; a = b; b.TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 44, 1, 57),
				},
			},
			// Independent conditional paths retain duplicate reports.
			{
				Code: "const util = flag ? require('util') : require('node:util'); util.TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 61, 1, 77),
					textDecoderAt("preferGlobal", 1, 61, 1, 77),
				},
			},
			// Static computed keys and escaped names retain their authored ranges.
			{
				Code: "const util = require('util'); const alias = util;\nalias['Text' + 'Decoder'];\nconst {[`TextDecoder`]: D} = util;\nutil['TextDecode\\u0072'];",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 2, 1, 2, 26),
					textDecoderAt("preferGlobal", 3, 8, 3, 26),
					textDecoderAt("preferGlobal", 4, 1, 4, 25),
				},
			},
			// String-named imports and re-exports select the builtin export.
			{
				Code: "import { 'TextDecoder' as D } from 'util'; export { 'TextDecoder' as Decoder } from 'node:util';",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 10, 1, 28),
					textDecoderAt("preferGlobal", 1, 53, 1, 77),
				},
			},
			// Type-only imports and re-exports still read the named export.
			{
				Code:     "import type { TextDecoder } from 'util'; export type { TextDecoder as D } from 'node:util'; export type * from 'util';",
				FileName: "review.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 15, 1, 26),
					textDecoderAt("preferGlobal", 1, 56, 1, 72),
					textDecoderAt("preferGlobal", 1, 93, 1, 119),
				},
			},
			// Escaped global identifiers and property names match decoded names.
			{
				Code:    "new TextDecode\\u0072(); globalThis.TextDecode\\u0072;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 5, 1, 21),
					textDecoderAt("preferModule", 1, 25, 1, 52),
				},
			},
			// Direct type references read the global but qualified type queries do not read a property.
			{
				Code:     "type D = TextDecoder; type T = typeof TextDecoder; type G = typeof globalThis.TextDecoder; interface I extends TextDecoder {}",
				Options:  []any{"never"},
				FileName: "review.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 10, 1, 21),
					textDecoderAt("preferModule", 1, 39, 1, 50),
					textDecoderAt("preferModule", 1, 112, 1, 123),
				},
			},
			// JavaScript JSX references the opening tag and expression containers.
			{
				Code:     "const el = <TextDecoder value={TextDecoder}>{new TextDecoder()}</TextDecoder>;",
				Options:  []any{"never"},
				FileName: "review.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 13, 1, 24),
					textDecoderAt("preferModule", 1, 32, 1, 43),
					textDecoderAt("preferModule", 1, 50, 1, 61),
				},
			},
			// Optional builtin loading and optional global access remain tracked.
			{
				Code: "process?.getBuiltinModule?.('node:util')?.TextDecoder; globalThis.require?.('util')?.TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 1, 1, 54),
					textDecoderAt("preferGlobal", 1, 56, 1, 97),
				},
			},
			// Hoisted local variables shadow only their own function.
			{
				Code:    "function f() { new TextDecoder(); var TextDecoder; } new TextDecoder();",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 58, 1, 69),
				},
			},
			// Inline disabling of the bare global does not disable global object properties.
			{
				Code:    "/* global TextDecoder: off */ new TextDecoder(); globalThis?.['TextDecoder'];",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 50, 1, 77),
				},
			},
			// Module property assignment and deletion still read the export.
			{
				Code: "require('util').TextDecoder = Other; delete require('util').TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 1, 1, 28),
					textDecoderAt("preferGlobal", 1, 45, 1, 72),
				},
			},
			// Local alias reassignment does not erase its tracked origin.
			{
				Code: "let util = require('util'); util = other; util.TextDecoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 43, 1, 59),
				},
			},
			// Global object shadowing is local to its function.
			{
				Code:    "function f(globalThis) { return globalThis.TextDecoder; } globalThis.TextDecoder;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferModule", 1, 59, 1, 81),
				},
			},
			// Nested destructuring through a global loader preserves both bindings.
			{
				Code: "const {require: load} = globalThis; const {TextDecoder: D = Other} = load('util');",
				Errors: []rule_tester.InvalidTestCaseError{
					textDecoderAt("preferGlobal", 1, 44, 1, 66),
				},
			},
		},
	)
}
