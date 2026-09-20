package process_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations checked against eslint-plugin-n v18.3.0 and @typescript-eslint/parser v8.65.0.
func TestProcessExtras(t *testing.T) {
	runProcessTests(t,
		[]rule_tester.ValidTestCase{
			// shadowed loaders.
			{Code: "function f(require, process) { require('process'); process.getBuiltinModule('process'); }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// modified loaders.
			{Code: "require('process'); require = other; process.getBuiltinModule('process'); process = other;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// disabled loaders.
			{Code: "require('process'); process.getBuiltinModule('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Globals:         map[string]any{"require": "off", "process": "off"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// unrelated and dynamic loads.
			{Code: "require('buffer'); require.resolve('process'); import('process'); new require('process'); require(); require(...['process']);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// module names do not follow local constants.
			{Code: "const name = 'process'; require(name); process.getBuiltinModule(name);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// local process bindings.
			{Code: "function f(process) { process.exit(0); } class C { m(process) { return process; } }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// imported process.
			{Code: "import process from 'node:process'; process.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// ambient and namespace bindings.
			{Code: "declare const process: any; process.exit(); namespace globalThis { export const process = 1; } globalThis.process;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"}},
			// disabled process.
			{Code: "process.exit(0)",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Globals:         map[string]any{"process": "off"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// modified process.
			{Code: "process.exit(0); process = replacement;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// computed dynamic and private keys.
			{Code: "const key = 'process'; global[key]; class C { #process; m() { return global.#process; } }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// inline globals.
			{Code: "/* global process: off */ process.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// commonjs top-level declaration shadows process.
			{Code: "var process; process.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// script top-level declaration shadows process.
			{Code: "var process; process.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// self-referencing parameter shadows process.
			{Code: "function f(process = process) { return process.env; }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// type import shadows process.
			{Code: "import type process from 'other'; process.env;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"}},
			// loop assignment disables global reads.
			{Code: "for (process of values) {} process.env;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// nested functions without a process reference.
			{Code: "export function read(value) { return value.name; }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
			// unrelated ESM imports.
			{Code: "import { readFile } from 'node:fs'; export { join } from 'node:path';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json"},
		},
		[]rule_tester.InvalidTestCase{
			// loader assertions and import equals.
			{Code: "(require as any)('process'); import p = require('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 28)}},
			// require aliases and constant strings.
			{Code: "const load = require; load('pro' + 'cess'); const {require: other} = global; other(`node:process`);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 23, 1, 43), processError("preferGlobal", 1, 78, 1, 99)}},
			// builtin loader aliases.
			{Code: "const p = process; const {getBuiltinModule: load} = p; load('node:process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 56, 1, 76)}},
			// parenthesized and optional loaders.
			{Code: "(require)('process'); require?.('process'); process?.getBuiltinModule('process'); (process?.getBuiltinModule)('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 21), processError("preferGlobal", 1, 23, 1, 43), processError("preferGlobal", 1, 45, 1, 81), processError("preferGlobal", 1, 83, 1, 121)}},
			// global loader computed keys.
			{Code: "globalThis['require']('process'); global['process']['get' + 'BuiltinModule']('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 33), processError("preferGlobal", 1, 35, 1, 88)}},
			// ESM import shapes.
			{Code: "import p from 'process'; import * as ns from 'node:process'; import {exit} from 'process'; import 'process';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 25), processError("preferGlobal", 1, 26, 1, 61), processError("preferGlobal", 1, 62, 1, 91), processError("preferGlobal", 1, 92, 1, 109)}},
			// ESM export shapes.
			{Code: "export * from 'process'; export * as proc from 'node:process'; export {exit as quit} from 'process';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 25), processError("preferGlobal", 1, 26, 1, 63), processError("preferGlobal", 1, 64, 1, 101)}},
			// type-only imports and exports.
			{Code: "import type P from 'process'; export type {Process} from 'node:process';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 30), processError("preferGlobal", 1, 31, 1, 73)}},
			// multiline loader with UTF-16 prefix.
			{Code: "\"😀\"; require(\n  \"process\"\n);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 7, 3, 2)}},
			// JSDoc loader wrapper.
			{Code: "/** @type {any} */ (require)(\"process\")",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 20, 1, 40)}},
			// type-only references and property names.
			{Code: "type P = typeof process; interface T { process: string; } const o = { process: 1 }; o.process;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 17, 1, 24)}},
			// global member shapes.
			{Code: "global.process; globalThis['process']; global?.['pro' + 'cess'];",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 15), processError("preferModule", 1, 17, 1, 38), processError("preferModule", 1, 40, 1, 64)}},
			// global alias and destructuring.
			{Code: "const g = globalThis; const {process: p} = g; p.exit(0); const q = process; q.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 30, 1, 40), processError("preferModule", 1, 68, 1, 75)}},
			// parentheses optional and shorthand.
			{Code: "(process).exit?.(0); process?.env; const o = {process}; typeof process;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 2, 1, 9), processError("preferModule", 1, 22, 1, 29), processError("preferModule", 1, 47, 1, 54), processError("preferModule", 1, 64, 1, 71)}},
			// type assertions are global reads.
			{Code: "(process as any).exit(0); process!.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 2, 1, 9), processError("preferModule", 1, 27, 1, 34)}},
			// JSX tag and expression references.
			{Code: "const element = <process.Component value={process.env} />;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Tsx:             true,
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 18, 1, 25), processError("preferModule", 1, 43, 1, 50)}},
			// multiline global member with UTF-16 prefix.
			{Code: "\"😀\"; globalThis[\n  \"process\"\n];",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 7, 3, 2)}},
			// global member assignment is a read event.
			{Code: "global.process = replacement;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 15)}},
			// only alias seed is a read.
			{Code: "let p; p = process; p.exit(0);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 12, 1, 19)}},
			// file suppression and re-enable.
			{Code: "/* eslint-disable */ require('process'); /* eslint-enable */ require('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 62, 1, 80)}},
			// line suppression with later global read.
			{Code: "// eslint-disable-next-line\nprocess.env;\nprocess.env;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 3, 1, 3, 8)}},
			// escaped loader and module name.
			{Code: "\\u0072\\u0065\\u0071\\u0075\\u0069\\u0072\\u0065('pro\\u0063ess'); import p from 'node:pro\\u0063ess';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 59), processError("preferGlobal", 1, 61, 1, 95)}},
			// escaped global references.
			{Code: "pro\\u0063ess.env; globalThis['pro\\u0063ess'];",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 13), processError("preferModule", 1, 19, 1, 45)}},
			// conditional loader aliases retain duplicate reports.
			{Code: "const load = cond ? require : globalThis.require; load('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 51, 1, 66), processError("preferGlobal", 1, 51, 1, 66)}},
			// logical and comma loader values.
			{Code: "(0, require)('process'); (require || fallback)('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 24), processError("preferGlobal", 1, 26, 1, 58)}},
			// destructuring assignment loader.
			{Code: "let load; ({require: load} = globalThis); load('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 43, 1, 58)}},
			// parameter default loader.
			{Code: "function f(load = require) { load('process'); }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 30, 1, 45)}},
			// type alias shares process name.
			{Code: "type process = { value: number }; process.env;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 35, 1, 42)}},
			// satisfies and instantiation wrappers.
			{Code: "(require<string>)('process'); (process satisfies unknown).getBuiltinModule('process');",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferGlobal", 1, 1, 1, 29), processError("preferGlobal", 1, 31, 1, 86)}},
			// JSX member tag differs from expression reads.
			{Code: "const element = <globalThis.process value={process.env} />;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Tsx:             true,
				Errors:          []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 44, 1, 51)}},
			// block binding does not hide outer global.
			{Code: "process.env; { let process; process.env; }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 8)}},
			// global destructuring assignment.
			{Code: "({process} = globalThis); process.env;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 3, 1, 10)}},
			// configured browser root and disabled direct global.
			{Code: "window.process; globalThis.process; process.env;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Options:         []any{"never"},
				Globals:         map[string]any{"window": "readonly", "process": "off"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{processError("preferModule", 1, 1, 1, 15), processError("preferModule", 1, 17, 1, 35)}},
		},
	)
}
