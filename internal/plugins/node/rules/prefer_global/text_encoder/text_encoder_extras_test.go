package text_encoder_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional cases compared with eslint-plugin-n v18.3.0 and @typescript-eslint/parser v8.65.0.
func TestTextEncoderExtras(t *testing.T) {
	runTextEncoderTests(t,
		[]rule_tester.ValidTestCase{
			// Other modules and exports do not match.
			{
				Code: "require('other').TextEncoder; require('util').TextDecoder; require('util'); process.getBuiltinModule('url').TextEncoder;",
			},
			// Dynamic module names and properties are not followed.
			{
				Code: "require(moduleName).TextEncoder; require('util')[key]; process.getBuiltinModule(moduleName).TextEncoder;",
			},
			// Shadowed loaders are ignored.
			{
				Code: "function f(require, process) { require('util').TextEncoder; process.getBuiltinModule('util').TextEncoder; }",
			},
			// Disabled loaders are ignored.
			{
				Code:    "require('util').TextEncoder; process.getBuiltinModule('util').TextEncoder;",
				Globals: map[string]any{"require": "off", "process": "off"},
			},
			// Reassigned global loaders are ignored throughout the file.
			{
				Code: "require('util').TextEncoder; require = other; process.getBuiltinModule('util').TextEncoder; process = other;",
			},
			// Imported process and dynamic imports do not root tracking.
			{
				Code: "import process from 'node:process'; process.getBuiltinModule('util').TextEncoder; const util = await import('node:util'); util.TextEncoder;",
			},
			// Local TextEncoder bindings shadow the global.
			{
				Code:    "function f(TextEncoder) { return new TextEncoder(); } { const TextEncoder = Custom; new TextEncoder(); }",
				Options: []any{"never"},
			},
			// A disabled global is ignored.
			{
				Code:    "new TextEncoder();",
				Options: []any{"never"},
				Globals: map[string]any{"TextEncoder": "off"},
			},
			// A global write disables tracking throughout the file.
			{
				Code:    "new TextEncoder(); TextEncoder = Custom;",
				Options: []any{"never"},
			},
			// Rest bindings and private properties do not match.
			{
				Code: "const {...util} = require('util'); util.TextEncoder; class C { #TextEncoder; m() { require('util').#TextEncoder; } }",
			},
			// JSX tag names do not read tracked properties.
			{
				Code:     "import util from 'node:util'; const el = <util.TextEncoder />;",
				FileName: "input.tsx",
			},
			// Bare property keys are not global reads.
			{
				Code:    "const obj = {TextEncoder: 1}; obj.TextEncoder;",
				Options: []any{"never"},
			},
			// Computed property identifiers are not resolved from their declarations.
			{
				Code: "const key = 'Text' + 'Encoder'; const util = require('util'); const alias = util; const TE = alias[key]; new TE();",
			},
		},
		[]rule_tester.InvalidTestCase{
			// A value declaration does not shadow a reference in the type namespace.
			{
				Code:     "type Encoder = typeof TextEncoder; interface I { encoder: TextEncoder } declare const TextEncoder: any; new TextEncoder();",
				Options:  []any{"never"},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 59, 1, 70),
				},
			},
			// Parenthesized optional and computed module reads retain their ranges.
			{
				Code: "(require('util'))?.TextEncoder; require('node:util')[('TextEncoder')];",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 1, 1, 31),
					textEncoderAt("preferGlobal", 1, 33, 1, 70),
				},
			},
			// Destructuring defaults report the whole property.
			{
				Code: "const {TextEncoder: TE = fallback} = process.getBuiltinModule('node:util'); new TE();",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 8, 1, 34),
				},
			},
			// Destructuring assignment and parameter defaults are followed.
			{
				Code: "let TE; ({TextEncoder: TE} = require('util')); function f({TextEncoder: D} = require('util')) { return D; }",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 11, 1, 26),
					textEncoderAt("preferGlobal", 1, 60, 1, 74),
				},
			},
			// Named imports report the specifier including its alias.
			{
				Code: "import { TextEncoder as TE } from 'node:util'; new TE();",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 10, 1, 27),
				},
			},
			// Default and namespace imports expose the CommonJS exports.
			{
				Code: "import util from 'util'; import * as ns from 'node:util'; util.TextEncoder; ns.TextEncoder; ns.default.TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 59, 1, 75),
					textEncoderAt("preferGlobal", 1, 77, 1, 91),
					textEncoderAt("preferGlobal", 1, 93, 1, 115),
				},
			},
			// Re-exports report their specifier or whole export-all statement.
			{
				Code: "export {TextEncoder as Encoder} from 'util'; export * from 'node:util'; export * as util from 'util';",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 9, 1, 31),
					textEncoderAt("preferGlobal", 1, 46, 1, 72),
					textEncoderAt("preferGlobal", 1, 73, 1, 102),
				},
			},
			// Global object aliases and nested bindings are followed.
			{
				Code:    "const root = globalThis; root.TextEncoder; const {TextEncoder: TE} = global; new TE();",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 26, 1, 42),
					textEncoderAt("preferModule", 1, 51, 1, 66),
				},
			},
			// Global object properties are checked even when the bare global is disabled.
			{
				Code:    "globalThis.TextEncoder;",
				Options: []any{"never"},
				Globals: map[string]any{"TextEncoder": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 1, 1, 23),
				},
			},
			// Read reports stop at aliases of the constructor.
			{
				Code:    "const TE = TextEncoder; new TE(); TextEncoder;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 12, 1, 23),
					textEncoderAt("preferModule", 1, 35, 1, 46),
				},
			},
			// Global object writes are reference-tracker reads.
			{
				Code:    "global.TextEncoder = Custom; delete globalThis.TextEncoder;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 1, 1, 19),
					textEncoderAt("preferModule", 1, 37, 1, 59),
				},
			},
			// JSX component names and expression containers read the global.
			{
				Code:     "const el = <TextEncoder value={TextEncoder}>{new TextEncoder()}</TextEncoder>;",
				Options:  []any{"never"},
				FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 13, 1, 24),
					textEncoderAt("preferModule", 1, 32, 1, 43),
					textEncoderAt("preferModule", 1, 50, 1, 61),
					textEncoderAt("preferModule", 1, 66, 1, 77),
				},
			},
			// Type assertions are transparent to module reference tracking.
			{
				Code:     "(require('util') as any).TextEncoder; require('util')!.TextEncoder; (require('util') satisfies any).TextEncoder;",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 1, 1, 37),
					textEncoderAt("preferGlobal", 1, 39, 1, 67),
					textEncoderAt("preferGlobal", 1, 69, 1, 112),
				},
			},
			// JSDoc casts preserve the module reference.
			{
				Code: "(/** @type {any} */ (require('util'))).TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 1, 1, 51),
				},
			},
			// Unicode and multiline reads use UTF-16 columns.
			{
				Code:    "\"😀\"; require('util')\n  .TextEncoder;",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 7, 2, 15),
				},
			},
			// Global reads preserve multiline positions.
			{
				Code:    "\"😀\"; TextEncoder;\nglobalThis\n  ['TextEncoder'];",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 7, 1, 18),
					textEncoderAt("preferModule", 2, 1, 3, 18),
				},
			},
			// Configured window and self objects are followed.
			{
				Code:    "window.TextEncoder; self['TextEncoder'];",
				Options: []any{"never"},
				Globals: map[string]any{"window": "readonly", "self": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 1, 1, 19),
					textEncoderAt("preferModule", 1, 21, 1, 40),
				},
			},
			// Inline global declarations override disabled configuration.
			{
				Code:    "/* global TextEncoder */ new TextEncoder();",
				Options: []any{"never"},
				Globals: map[string]any{"TextEncoder": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 30, 1, 41),
				},
			},
			// Aliased loaders and static module names work.
			{
				Code: "const load = require; load('node:' + 'util').TextEncoder; const p = process; p.getBuiltinModule('util').TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 23, 1, 57),
					textEncoderAt("preferGlobal", 1, 78, 1, 116),
				},
			},
		},
	)
}

// Further reference and AST cases compared with eslint-plugin-n v18.3.0.
func TestTextEncoderReferenceEdges(t *testing.T) {
	runTextEncoderTests(t,
		[]rule_tester.ValidTestCase{
			// Empty and spread module arguments are not statically known.
			{
				Code: "require(); process.getBuiltinModule(); require(...['util']).TextEncoder; process.getBuiltinModule(...['util']).TextEncoder;",
			},
			// Import-equals and type namespace members do not create ESM reads.
			{
				Code:     "import util = require('node:util'); util.TextEncoder; import type * as types from 'util'; type D = typeof types.TextEncoder;",
				FileName: "review.ts",
			},
			// Script declarations shadow the configured global.
			{
				Code:            "var TextEncoder; new TextEncoder();",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Comma and logical expressions preserve only the possible module values.
			{
				Code: "(0, require('util')).TextEncoder; (require('util'), other).TextEncoder; (flag && require('util')).TextEncoder; (flag || require('util')).TextEncoder; (flag ?? require('util')).TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 1, 1, 33),
					textEncoderAt("preferGlobal", 1, 73, 1, 110),
					textEncoderAt("preferGlobal", 1, 112, 1, 149),
					textEncoderAt("preferGlobal", 1, 151, 1, 188),
				},
			},
			// Interface extends and class implements expose the same member reads as class extends.
			{
				Code:     "import util from 'util'; class C extends util.TextEncoder {} interface I extends util.TextEncoder {} class D implements util.TextEncoder {}",
				FileName: "edge.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 42, 1, 58),
					textEncoderAt("preferGlobal", 1, 82, 1, 98),
					textEncoderAt("preferGlobal", 1, 121, 1, 137),
				},
			},
			// A body declaration does not shadow a parameter initializer.
			{
				Code:    "function f(value = TextEncoder) { var TextEncoder; }",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 20, 1, 31),
				},
			},
			// Static-block bindings do not shadow reads in class methods.
			{
				Code:    "class C { static { const TextEncoder = Other; new TextEncoder(); } method() { return new TextEncoder(); } }",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 90, 1, 101),
				},
			},
			// Shorthand values and computed property keys are global reads.
			{
				Code:    "const value = {TextEncoder, [TextEncoder]: TextEncoder};",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 16, 1, 27),
					textEncoderAt("preferModule", 1, 30, 1, 41),
					textEncoderAt("preferModule", 1, 44, 1, 55),
				},
			},
			// Line disable directives suppress only the next read.
			{
				Code: "/* eslint-disable-next-line test */\nrequire('util').TextEncoder;\nrequire('util').TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 3, 1, 3, 28),
				},
			},
			// Block disable and enable directives preserve later reads.
			{
				Code:    "/* eslint-disable test */\nnew TextEncoder();\n/* eslint-enable test */\nnew TextEncoder();",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 4, 5, 4, 16),
				},
			},
			// Computed names can reach the builtin loader without an identifier named process.
			{
				Code: "const module = globalThis['pro' + 'cess']['getBuiltinModule']('node:util'); module['Text' + 'Encoder'];",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 77, 1, 103),
				},
			},
			// Cyclic aliases terminate without losing the tracked read.
			{
				Code: "let a = require('util'); let b = a; a = b; b.TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 44, 1, 57),
				},
			},
			// Independent conditional paths retain duplicate reports.
			{
				Code: "const util = flag ? require('util') : require('node:util'); util.TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 61, 1, 77),
					textEncoderAt("preferGlobal", 1, 61, 1, 77),
				},
			},
			// Static computed keys and escaped names retain their authored ranges.
			{
				Code: "const util = require('util'); const alias = util;\nalias['Text' + 'Encoder'];\nconst {[`TextEncoder`]: D} = util;\nutil['TextEncode\\u0072'];",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 2, 1, 2, 26),
					textEncoderAt("preferGlobal", 3, 8, 3, 26),
					textEncoderAt("preferGlobal", 4, 1, 4, 25),
				},
			},
			// String-named imports and re-exports select the builtin export.
			{
				Code: "import { 'TextEncoder' as D } from 'util'; export { 'TextEncoder' as Encoder } from 'node:util';",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 10, 1, 28),
					textEncoderAt("preferGlobal", 1, 53, 1, 77),
				},
			},
			// Type-only imports and re-exports still read the named export.
			{
				Code:     "import type { TextEncoder } from 'util'; export type { TextEncoder as D } from 'node:util'; export type * from 'util';",
				FileName: "review.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 15, 1, 26),
					textEncoderAt("preferGlobal", 1, 56, 1, 72),
					textEncoderAt("preferGlobal", 1, 93, 1, 119),
				},
			},
			// Escaped global identifiers and property names match decoded names.
			{
				Code:    "new TextEncode\\u0072(); globalThis.TextEncode\\u0072;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 5, 1, 21),
					textEncoderAt("preferModule", 1, 25, 1, 52),
				},
			},
			// Direct type references read the global but qualified type queries do not read a property.
			{
				Code:     "type D = TextEncoder; type T = typeof TextEncoder; type G = typeof globalThis.TextEncoder; interface I extends TextEncoder {}",
				Options:  []any{"never"},
				FileName: "review.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 10, 1, 21),
					textEncoderAt("preferModule", 1, 39, 1, 50),
					textEncoderAt("preferModule", 1, 112, 1, 123),
				},
			},
			// JavaScript JSX references the opening tag and expression containers.
			{
				Code:     "const el = <TextEncoder value={TextEncoder}>{new TextEncoder()}</TextEncoder>;",
				Options:  []any{"never"},
				FileName: "review.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 13, 1, 24),
					textEncoderAt("preferModule", 1, 32, 1, 43),
					textEncoderAt("preferModule", 1, 50, 1, 61),
				},
			},
			// Optional builtin loading and optional global access remain tracked.
			{
				Code: "process?.getBuiltinModule?.('node:util')?.TextEncoder; globalThis.require?.('util')?.TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 1, 1, 54),
					textEncoderAt("preferGlobal", 1, 56, 1, 97),
				},
			},
			// Hoisted local variables shadow only their own function.
			{
				Code:    "function f() { new TextEncoder(); var TextEncoder; } new TextEncoder();",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 58, 1, 69),
				},
			},
			// Inline disabling of the bare global does not disable global object properties.
			{
				Code:    "/* global TextEncoder: off */ new TextEncoder(); globalThis?.['TextEncoder'];",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 50, 1, 77),
				},
			},
			// Module property assignment and deletion still read the export.
			{
				Code: "require('util').TextEncoder = Other; delete require('util').TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 1, 1, 28),
					textEncoderAt("preferGlobal", 1, 45, 1, 72),
				},
			},
			// Local alias reassignment does not erase its tracked origin.
			{
				Code: "let util = require('util'); util = other; util.TextEncoder;",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 43, 1, 59),
				},
			},
			// Global object shadowing is local to its function.
			{
				Code:    "function f(globalThis) { return globalThis.TextEncoder; } globalThis.TextEncoder;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferModule", 1, 59, 1, 81),
				},
			},
			// Nested destructuring through a global loader preserves both bindings.
			{
				Code: "const {require: load} = globalThis; const {TextEncoder: D = Other} = load('util');",
				Errors: []rule_tester.InvalidTestCaseError{
					textEncoderAt("preferGlobal", 1, 44, 1, 66),
				},
			},
		},
	)
}
