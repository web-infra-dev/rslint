package console_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional cases compared with eslint-plugin-n v18.3.0 (ESLint 10.2.1).
func TestPreferGlobalConsoleExtras(t *testing.T) {
	runConsoleTests(t, []rule_tester.ValidTestCase{
		// Other modules and dynamic loads.
		{Code: "require('node:util'); require(name); require(); require(...['console']); process.getBuiltinModule(name); process.getBuiltinModule(); process.getBuiltinModule(...['console']); import('console');",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Shadowed loaders.
		{Code: "function f(require, process) { require('console'); process.getBuiltinModule('console'); }",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Reassigned loaders.
		{Code: "require('console'); require = custom; process.getBuiltinModule('console'); process = custom;",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Disabled loaders.
		{Code: "require('console'); process.getBuiltinModule('console');",
			Globals:  map[string]any{"require": "off", "process": "off"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Source variable is not a literal.
		{Code: "const name = 'console'; require(name); process.getBuiltinModule(name);",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Unknown method arguments do not produce constant module names.
		{Code: "require('console'.toString(unknown)); process.getBuiltinModule(name.toString());",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Local module bindings with never.
		{Code: "import console from 'console'; console.log(); export { Console } from 'console';",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Local console declarations.
		{Code: "const console = logger; console.log(); function f(console) { console.warn(); }",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Catch, class and block scopes.
		{Code: "try {} catch (console) { console.log(); } { let console; console.log(); } class Logger { method(console) { console.log(); } }",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Global writes disable direct tracking.
		{Code: "console.log(); console = logger; console.warn();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Read-write references disable tracking.
		{Code: "console++; console.log();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Disabled console.
		{Code: "console.log();",
			Options:  []any{"never"},
			Globals:  map[string]any{"console": "off"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Inline global override.
		{Code: "/* global console: off */ console.log();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Shadowed and modified global objects.
		{Code: "function f(globalThis) { globalThis.console.log(); } global.console.log(); global = custom;",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// A console property is not a global.
		{Code: "logger.console.log(); ({ console: logger }); class Logger { console() {} }",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Private keys are not global properties.
		{Code: "class Logger { #console; log() { return globalThis.#console; } }",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json"},
		// Type-only references and ambient declarations.
		{Code: "type Logger = typeof console; interface Options { logger: typeof console } declare const console: any; console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts"},
		// Import equals is not an ESM load.
		{Code: "import console = require('console'); console.log();",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts"},
	}, []rule_tester.InvalidTestCase{
		// Constant string methods resolve the same module names as literals.
		{Code: "require('console'.toString());\nprocess.getBuiltinModule('node:console'.toString());",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 30),
				consoleError("preferGlobal", 2, 1, 2, 52),
			}},
		// Constant source expressions.
		{Code: "require(`console`); require('con' + 'sole'); process.getBuiltinModule('node:' + 'console');",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 19),
				consoleError("preferGlobal", 1, 21, 1, 44),
				consoleError("preferGlobal", 1, 46, 1, 91),
			}},
		// Loader aliases.
		{Code: "const load = require; load('console'); const { getBuiltinModule: builtin } = process; builtin('node:console');",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 23, 1, 38),
				consoleError("preferGlobal", 1, 87, 1, 110),
			}},
		// Optional and parenthesized loaders.
		{Code: "(require)('console'); require?.('console'); process?.getBuiltinModule?.('console'); (process.getBuiltinModule)('console');",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 21),
				consoleError("preferGlobal", 1, 23, 1, 43),
				consoleError("preferGlobal", 1, 45, 1, 83),
				consoleError("preferGlobal", 1, 85, 1, 122),
			}},
		// Computed loader names.
		{Code: "process['get' + 'BuiltinModule']('console'); process[`getBuiltinModule`]('node:console'); process[key]('console');",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 44),
				consoleError("preferGlobal", 1, 46, 1, 89),
			}},
		// Only module loads report, not their aliases.
		{Code: "const logger = require('console'); const alias = logger; logger.log(); alias.warn(); const { log } = require('node:console');",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 16, 1, 34),
				consoleError("preferGlobal", 1, 102, 1, 125),
			}},
		// Multiline calls and UTF-16 positions.
		{Code: "/* 😀 */ require(\n 'console'\n);\nprocess.getBuiltinModule(\n 'node:console'\n);",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 10, 3, 2),
				consoleError("preferGlobal", 4, 1, 6, 2),
			}},
		// Default, namespace, named and side-effect imports.
		{Code: "import logger from 'console';\nimport * as namespace from 'node:console';\nimport { log, Console } from 'console';\nimport 'node:console';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 30),
				consoleError("preferGlobal", 2, 1, 2, 43),
				consoleError("preferGlobal", 3, 1, 3, 40),
				consoleError("preferGlobal", 4, 1, 4, 23),
			}},
		// Re-exports.
		{Code: "export { default as logger } from 'console';\nexport * from 'console';\nexport * as loggerNS from 'node:console';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 45),
				consoleError("preferGlobal", 2, 1, 2, 25),
				consoleError("preferGlobal", 3, 1, 3, 42),
			}},
		// Read and construct module values.
		{Code: "require('console').log(); new (require('console').Console)(sink);",
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 19),
				consoleError("preferGlobal", 1, 32, 1, 50),
			}},
		// Global objects retain console independently.
		{Code: "global.console.log(); globalThis.console.warn();",
			Options:  []any{"never"},
			Globals:  map[string]any{"console": "off"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 1, 1, 15),
				consoleError("preferModule", 1, 23, 1, 41),
			}},
		// Global object aliases and destructuring.
		{Code: "const root = globalThis; root['console'].log(); const { console: logger } = global; logger.log();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 26, 1, 41),
				consoleError("preferModule", 1, 57, 1, 72),
			}},
		// Configured browser global roots.
		{Code: "self.console.log(); window['console'].warn();",
			Options:  []any{"never"},
			Globals:  map[string]any{"self": "readonly", "window": "readonly"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 1, 1, 13),
				consoleError("preferModule", 1, 21, 1, 38),
			}},
		// Direct reads only once through aliases.
		{Code: "const logger = console; logger.log(); const { log } = console; log();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 16, 1, 23),
				consoleError("preferModule", 1, 55, 1, 62),
			}},
		// Computed and optional reads.
		{Code: "(console)?.log(); console['log'](); globalThis?.['con' + 'sole']?.warn();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 2, 1, 9),
				consoleError("preferModule", 1, 19, 1, 26),
				consoleError("preferModule", 1, 37, 1, 65),
			}},
		// Expression positions.
		{Code: "typeof console; void console; ({ console }); class Logger extends console.Console {}",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 8, 1, 15),
				consoleError("preferModule", 1, 22, 1, 29),
				consoleError("preferModule", 1, 34, 1, 41),
				consoleError("preferModule", 1, 67, 1, 74),
			}},
		// Multiline global member and UTF-16 positions.
		{Code: "/* 😀 */ console.log();\nglobalThis\n .console.log();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 10, 1, 17),
				consoleError("preferModule", 2, 1, 3, 10),
			}},
		// Type query without local binding.
		{Code: "type Logger = typeof console; interface Options { logger: typeof console }",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 22, 1, 29),
				consoleError("preferModule", 1, 66, 1, 73),
			}},
		// Type-only module declarations follow the same source preference.
		{Code: "import type { Console } from 'console';\nexport type { Console as Logger } from 'node:console';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 40),
				consoleError("preferGlobal", 2, 1, 2, 55),
			}},
		{Code: "import type ConsoleModule from 'console';\nimport type * as NodeConsole from 'node:console';\nexport type * from 'console';\nexport type * as Logger from 'node:console';",
			Options:         []any{"always"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 42),
				consoleError("preferGlobal", 2, 1, 2, 50),
				consoleError("preferGlobal", 3, 1, 3, 30),
				consoleError("preferGlobal", 4, 1, 4, 45),
			}},
		// Type-only and runtime declarations both reference the module.
		{Code: "import type { Console } from 'console';\nimport 'node:console';\nexport type { Console } from 'console';\nexport { log } from 'console';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 40),
				consoleError("preferGlobal", 2, 1, 2, 23),
				consoleError("preferGlobal", 3, 1, 3, 40),
				consoleError("preferGlobal", 4, 1, 4, 31),
			}},
		// Inline type specifiers retain module loading with verbatimModuleSyntax.
		{Code: "import { type Console } from 'console';\nexport { type Console as Logger } from 'node:console';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 40),
				consoleError("preferGlobal", 2, 1, 2, 55),
			}},
		// Mixed type and value specifiers still load the module.
		{Code: "import { type Console, log } from 'console';\nexport { type Console as Logger, log } from 'node:console';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 45),
				consoleError("preferGlobal", 2, 1, 2, 60),
			}},
		// Type assertions around global reads.
		{Code: "(console as any).log(); console!.warn(); (console satisfies unknown).error();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 2, 1, 9),
				consoleError("preferModule", 1, 25, 1, 32),
				consoleError("preferModule", 1, 43, 1, 50),
			}},
		// Type assertions around loaders.
		{Code: "(require as any)('console'); process!.getBuiltinModule('console');",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferGlobal", 1, 1, 1, 28),
				consoleError("preferGlobal", 1, 30, 1, 66),
			}},
		// JSX tag names versus expressions.
		{Code: "const element = <console.Log logger={console} />;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.tsx",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 18, 1, 25),
				consoleError("preferModule", 1, 38, 1, 45),
			}},
		// JSDoc types and cast expressions.
		{Code: "/** @type {typeof console} */ let logger; (/** @type {any} */ (console)).log();",
			Options:  []any{"never"},
			FileName: "input.js", TSConfig: "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{
				consoleError("preferModule", 1, 64, 1, 71),
			}},
	})
}

// Further cases compared with eslint-plugin-n v18.3.0. Type-only exports
// retain the documented behavior; all other expectations match upstream.
func TestPreferGlobalConsoleReferenceEdges(t *testing.T) {
	runConsoleTests(t, []rule_tester.ValidTestCase{
		// Intrinsic JSX tag.
		{Code: "const element = <console />;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.tsx"},
		// Global object JSX members.
		{Code: "const element = <globalThis.console.Log />;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.tsx"},
		// New and tagged loaders.
		{Code: "new require('console'); require`console`;",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json"},
		// Global member JSDoc types.
		{Code: "/** @type {typeof globalThis.console} */ let logger;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json"},
		// Uninitialized script global declaration.
		{Code: "var console; console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json"},
		// Script declaration with initializer.
		{Code: "var console = logger; console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json"},
		// Namespace declaration shadows console.
		{Code: "namespace console { export type T = number; } console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts"},
		// Global member in type query.
		{Code: "type Logger = typeof globalThis.console;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts"},
		// TypeScript rejects exporting a global declaration (TS2661).
		{Code: "export type { console as Logger };",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts"},
		// Same line suppression.
		{Code: "console.log(); // eslint-disable-line",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json"},
	}, []rule_tester.InvalidTestCase{
		// Opening and closing JSX members.
		{Code: "const element = <console.Log></console.Log>;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.tsx",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 18, 1, 25),
				consoleError("preferModule", 1, 32, 1, 39)}},
		// JSX expression references.
		{Code: "const element = <div>{console.log(\"hi\")}</div>;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.tsx",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 23, 1, 30)}},
		// Duplicate module loader paths.
		{Code: "const load = flag ? require : globalThis.require; load('console');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 51, 1, 66),
				consoleError("preferGlobal", 1, 51, 1, 66)}},
		// Alias cycles.
		{Code: "let load = require; load = load; load('console');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 34, 1, 49)}},
		// Duplicate global object paths.
		{Code: "const root = flag ? global : globalThis; root.console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 42, 1, 54),
				consoleError("preferModule", 1, 42, 1, 54)}},
		// Conditional and comma loaders.
		{Code: "(flag ? require : fallback)('console'); (0, require)('console'); (require, other)('console');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 1, 1, 39),
				consoleError("preferGlobal", 1, 41, 1, 64)}},
		// Assignment and default loader aliases.
		{Code: "let load; (load = require)('console'); function f(load = require) { load('console'); }",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 11, 1, 38),
				consoleError("preferGlobal", 1, 69, 1, 84)}},
		// Global property writes are reads.
		{Code: "globalThis.console = custom; delete globalThis.console;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 1, 1, 19),
				consoleError("preferModule", 1, 37, 1, 55)}},
		// Destructured global default.
		{Code: "const {console: logger = custom} = globalThis;",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 8, 1, 32)}},
		// Destructuring assignment and rest.
		{Code: "let logger; ({console: logger} = globalThis); const {...root} = globalThis; root.console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 15, 1, 30)}},
		// Global aliases are read at the source.
		{Code: "let logger; (logger = console).log(); function f(logger = console) { logger.log(); }",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 23, 1, 30),
				consoleError("preferModule", 1, 59, 1, 66)}},
		// Extra argument still loads a module.
		{Code: "require('console', ignored); process.getBuiltinModule('console', ignored);",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 1, 1, 28),
				consoleError("preferGlobal", 1, 30, 1, 74)}},
		// Parenthesized optional chain boundary.
		{Code: "(process?.getBuiltinModule)('console'); (globalThis?.require)('console');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 1, 1, 39),
				consoleError("preferGlobal", 1, 41, 1, 73)}},
		// Overload parameter does not shadow global.
		{Code: "declare function log(console: unknown): void; console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 47, 1, 54)}},
		// Type declaration shares console name.
		{Code: "type console = string; console.log();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 24, 1, 31)}},
		// Computed names and extends expressions.
		{Code: "class Logger extends (console.Console) { [console.log()]() {} }",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors: []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 23, 1, 30),
				consoleError("preferModule", 1, 43, 1, 50)}},
		// Import attributes range.
		{Code: "import logger from 'node:console' with {type: 'json'};",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 1, 1, 55)}},
		// Export without source.
		{Code: "export { console as logger };",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			FileName:        "input.ts",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 1, 10, 1, 17)}},
		// Line suppression.
		{Code: "// eslint-disable-next-line\nconsole.log();\nconsole.warn();",
			Options:         []any{"never"},
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferModule", 3, 1, 3, 8)}},
		// Block suppression and enable.
		{Code: "/* eslint-disable */ require('console'); /* eslint-enable */ require('console');",
			LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			FileName:        "input.js",
			TSConfig:        "tsconfig.allowJs.json",
			Errors:          []rule_tester.InvalidTestCaseError{consoleError("preferGlobal", 1, 62, 1, 80)}},
	})
}
