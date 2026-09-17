package no_path_concat_test

import (
	"os"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Boundary cases checked against eslint-plugin-n v18.3.0 with eslint-utils 4.10.1.
func TestNoPathConcatExtras(t *testing.T) {
	runNoPathConcatTests(t, []rule_tester.ValidTestCase{
		// ordinary aliases and global-object paths are not direct global identifiers.
		{Code: "const dir = __dirname; dir + \"/x\"; globalThis.__dirname + \"/x\"; const {__filename: file} = global; file + \"/x\";"},
		// shadowed parameters and block bindings.
		{Code: "function f(__dirname) { return __dirname + \"/x\"; } { const __filename = \"x\"; __filename + \"/x\"; }"},
		// a global write suppresses all reads of that root.
		{Code: "__dirname + \"/x\"; __dirname = \"/tmp\";"},
		// destructuring and update writes suppress the roots.
		{Code: "__dirname + \"/x\"; ({__dirname} = obj); __filename + \"/x\"; __filename++;"},
		// declared global in script scope is modified.
		{Code: "var __dirname; __dirname + \"/x\";", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// disabled globals.
		{Code: "__dirname + \"/x\"; __filename + \"/x\";", Globals: map[string]any{"__dirname": "off", "__filename": "off"}},
		// undeclared globals.
		{Code: "__dirname + \"/x\"; __filename + \"/x\";", Globals: map[string]any{}},
		// authored TS assertions remain parents of the path root.
		{Code: "(__dirname as string) + \"/x\"; (__filename satisfies string) + \"/x\"; import.meta.dirname! + \"/x\";", FileName: "input.ts"},
		// script type alias defines the global name.
		{Code: "type __dirname = string; __dirname + \"/x\";", FileName: "input.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// only immediate concatenation is checked.
		{Code: "(__dirname + \"\") + \"/x\"; __dirname += \"/x\"; const paths = `${__filename}${\"\"}/x`;"},
		// computed import.meta paths.
		{Code: "import.meta[\"dir\" + \"name\"] + \"/x\"; import.meta[`file${\"name\"}`] + \"/x\"; const key = \"url\"; import.meta[key] + \"/x\";"},
		// optional import.meta paths are wrapped in a chain.
		{Code: "import.meta?.dirname + \"/x\"; `${import.meta?.url}/x`;"},
		// unrelated and private properties.
		{Code: "obj.dirname + \"/x\"; import.meta.other + \"/x\"; class C { #dirname; f() { return import.meta.#dirname + \"/x\"; } }"},
		// unknown interpolation does not skip ahead to later separators.
		{Code: "`${__dirname}${unknown}/x`; __dirname + (\"\" + \"/x\");"},
		// invalid cooked template element is unknown.
		{Code: "tag`${__dirname}\\unicode${\"/\"}`;"},
		// modified non-const suffix is unknown.
		{Code: "let suffix = \"/x\"; suffix = getSuffix(); __dirname + suffix;"},
		// empty strings and non-separator primitives.
		{Code: "__dirname + \"\"; __dirname + 1; __dirname + null; __dirname + {};"},
		// destructured and aliased sep values are not tracked reads.
		{Code: "const {sep} = require(\"path\"); const p = require(\"path\"); const alias = p.sep; __dirname + sep; __dirname + alias;"},
		// only the exact path module is tracked.
		{Code: "const p = require(\"node:path\"); __dirname + p.sep;"},
		// module names are evaluated without variable scope.
		{Code: "const name = \"path\"; const p = require(name); __dirname + p.sep;"},
		// require.resolve does not return the path module.
		{Code: "__dirname + require.resolve(\"path\").sep;"},
		// shadowed require is ignored.
		{Code: "function f(require) { const p = require(\"path\"); return __dirname + p.sep; }"},
		// modified require is ignored.
		{Code: "const p = require(\"path\"); require = other; __dirname + p.sep;"},
		// path methods and optional sep access are not separators.
		{Code: "const p = require(\"path\"); __dirname + p.join; __dirname + p?.sep;"},
		// named sep import and node prefix remain untracked upstream.
		{Code: "import {sep} from \"path\"; import p from \"node:path\"; __dirname + sep; __dirname + p.sep;"},
		// non-separator Unicode.
		{Code: "__dirname + \"\\ud800/x\"; __dirname + \"／x\"; __dirname + \"%2Fx\";"},
		// loop writes.
		{Code: "__dirname + \"/x\"; for (__dirname of paths) {} __filename + \"/x\"; for ({path: __filename} of paths) {}"},
		// inline disabled global.
		{Code: "/* global __dirname: off */\n__dirname + \"/x\";"},
		// type positions.
		{Code: "type Path = `${typeof __dirname}/x`;", FileName: "input.ts"},
		// asserted computed key.
		{Code: "import.meta[(\"dirname\" as const)] + \"/x\";", FileName: "input.ts"},
		// asserted member receiver.
		{Code: "(import.meta as any).dirname + \"/x\";", FileName: "input.ts"},
		// namespace shadows global.
		{Code: "namespace __dirname { export const name = \"dir\"; } __dirname + \"/x\";", FileName: "input.ts"},
		// empty cooked tail.
		{Code: "`${__dirname}`; `${import.meta.url}${\"\"}`;"},
		// Documented difference: upstream reports String.fromCodePoint(47), but rslint does not.
		{Code: "__dirname + String.fromCodePoint(47);"},
		// module rest binding.
		{Code: "const {...p} = require(\"path\"); __dirname + p.sep;"},
		// namespace rest assignment.
		{Code: "import * as ns from \"path\"; let p; ({...p} = ns); __dirname + p.default.sep;"},
		// array literal alias remains untracked.
		{Code: "const [load] = [require]; __dirname + load(\"path\").sep;"},
		// require alias becomes module.
		{Code: "let p = require; p = p(\"path\"); __dirname + p.sep;"},
		// namespace alias becomes default.
		{Code: "import * as ns from \"path\"; let p = ns; p = p.default; __dirname + p.sep;"},
		// resolve alias is not module.
		{Code: "const {resolve: load} = require; __dirname + load(\"path\").sep;"},
		// require and module in destructuring defaults.
		{Code: "let p; ({p = require} = source); p = p(\"path\"); __dirname + p.sep;"},
	}, []rule_tester.InvalidTestCase{
		// A shorthand require assignment still reaches the path module.
		{Code: "({require} = globalThis); __dirname + require('path').sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 27, 1, 58)}},
		// independent global writes.
		{Code: "__dirname + \"/x\"; __filename + \"/y\"; __dirname = other;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 19, 1, 36)}},
		// independent filename writes.
		{Code: "__dirname + \"/x\"; __filename + \"/y\"; __filename = other;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 17)}},
		// property key is not global write.
		{Code: "({__dirname: target} = obj); __dirname + \"/x\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 30, 1, 46)}},
		// escaped global identifier.
		// cspell:ignore dirn
		{Code: "__dirn\\u0061me + \"/x\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 22)}},
		// overwritten direct require alias.
		{Code: "let load = require; load = other; __dirname + load(\"path\").sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 35, 1, 63)}},
		// shadowed writes do not invalidate a global.
		{Code: "function f(__dirname) { __dirname = \"/tmp\"; } __dirname + \"/x\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 47, 1, 63)}},
		// parameter initializer is outside body bindings.
		{Code: "function f(value = __dirname + \"/x\") { var __dirname; }", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 20, 1, 36)}},
		// class static block and JSX expressions.
		{Code: "class C { static { __dirname + \"/x\"; } }; const view = <Panel value={import.meta.url + \"/x\"} />;", FileName: "input.tsx", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 20, 1, 36), concatErrorAt("useUrl", 1, 70, 1, 92)}},
		// parentheses preserve the binary range.
		{Code: "((__dirname)) + (\"/x\"); (import.meta)[(\"filename\")] + (\"/x\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 23), concatErrorAt("usePathFunctions", 1, 25, 1, 61)}},
		// JSDoc cast is transparent.
		{Code: "(/** @type {string} */ (__dirname)) + \"/x\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 43)}},
		// TypeScript suffix assertion uses upstream static evaluation.
		{Code: "__dirname + (\"/x\" as const);", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 28)}},
		// module type alias does not shadow a value.
		{Code: "export {}; type __dirname = string; __dirname + \"/x\";", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 37, 1, 53)}},
		// static string and template keys.
		{Code: "import.meta[`dirname`] + \"/x\"; `${(import.meta)[\"filename\"]}/x`; import.meta[\"u\\u0072l\"] + \"/x\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 30), concatErrorAt("usePathFunctions", 1, 32, 1, 64), concatErrorAt("useUrl", 1, 66, 1, 96)}},
		// multiline and UTF-16 ranges.
		{Code: "const emoji = \"😀\"; import.meta.url +\n  \"/x\";\n`start😀${__dirname}\n/x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("useUrl", 1, 21, 2, 7)}},
		// Multiline template with a UTF-16 start column.
		{Code: "const prefix = \"😀\"; `😀${__dirname}/\nfile`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 22, 2, 6)}},
		// Multiple roots report the complete template for each.
		{Code: "`${__dirname}/${__filename}/${import.meta.url}/x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("useUrl", 1, 1, 1, 50), concatErrorAt("usePathFunctions", 1, 1, 1, 50), concatErrorAt("usePathFunctions", 1, 1, 1, 50)}},
		// empty literal delegates to the next expression.
		{Code: "`${__dirname}${\"/\"}x`; `${import.meta.url}${unknown || \"/\"}x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 22), concatErrorAt("useUrl", 1, 24, 1, 62)}},
		// nested templates use their first element only.
		{Code: "__dirname + `${\"/\"}${unknown}`; __filename + `${\"\"}/x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 31)}},
		// tagged templates are checked.
		{Code: "tag`${__dirname}/x`; tag`${import.meta.url}/x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 4, 1, 20), concatErrorAt("useUrl", 1, 25, 1, 47)}},
		// sequence checks its final expression.
		{Code: "__dirname + (sideEffect(), \"/x\"); __dirname + (\"/x\", \".map\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 33)}},
		// assignments use their right-hand side.
		{Code: "__dirname + (part = \"/x\"); __filename + (part += \"/x\"); __dirname + (part ||= \"/x\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 26), concatErrorAt("usePathFunctions", 1, 28, 1, 55), concatErrorAt("usePathFunctions", 1, 57, 1, 84)}},
		// binary operators use their left operand.
		{Code: "__dirname + (\"/\" * unknown); __dirname + (\"x\" + \"/\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 28)}},
		// conditional and logical alternatives are conservatively checked.
		{Code: "__dirname + (false ? \"/x\" : \".map\"); __dirname + (\".map\" || \"/x\"); __filename + (false && \"/x\"); import.meta.dirname + (unknown ?? \"/x\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 36), concatErrorAt("usePathFunctions", 1, 38, 1, 66), concatErrorAt("usePathFunctions", 1, 68, 1, 96), concatErrorAt("usePathFunctions", 1, 98, 1, 137)}},
		// static local strings and computed properties.
		{Code: "const suffix = \"/x\"; const obj = {suffix}; __dirname + suffix; __filename + obj[\"suffix\"];", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 44, 1, 62), concatErrorAt("usePathFunctions", 1, 64, 1, 90)}},
		// constant coercion includes regexps and arrays.
		{Code: "__dirname + /x/; __dirname + [\"/x\"]; __dirname + String.raw`/x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 16), concatErrorAt("usePathFunctions", 1, 18, 1, 36), concatErrorAt("usePathFunctions", 1, 38, 1, 64)}},
		// CommonJS direct module property.
		{Code: "__dirname + require(\"path\").sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 32)}},
		// module object aliases and destructuring.
		{Code: "const {require: load} = global; const p = load(\"path\"); const alias = p; __dirname + alias[\"s\" + \"ep\"];", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 74, 1, 103)}},
		// constant module names and require aliases.
		{Code: "const load = require; const p = load(\"pa\" + \"th\"); __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 52, 1, 69)}},
		// parenthesized optional require result still exposes sep.
		{Code: "__dirname + (require?.(\"path\")).sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 36)}},
		// module assignments and overwritten aliases retain the upstream trace.
		{Code: "let p; p = require(\"path\"); p = other; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 40, 1, 57)}},
		// module defaults in parameters.
		{Code: "function f(p = require(\"path\")) { return __dirname + p.sep; }", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 42, 1, 59)}},
		// global require alias and logical module expressions.
		{Code: "const load = globalThis.require; const p = enabled ? load(\"path\") : other; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 76, 1, 93)}},
		// default ESM import exposes the module object.
		{Code: "import p from \"path\"; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 23, 1, 40)}},
		// named default import and module alias.
		{Code: "import {default as path} from \"path\"; const p = path; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 55, 1, 72)}},
		// strict namespace import exposes CommonJS under default.
		{Code: "import * as p from \"path\"; __dirname + p.sep; __dirname + p.default.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 47, 1, 72)}},
		// import.meta paths do not require CommonJS globals.
		{Code: "import.meta.dirname + \"/x\"; import.meta.filename + \"/x\"; import.meta.url + \"/x\";", Globals: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 27), concatErrorAt("usePathFunctions", 1, 29, 1, 56), concatErrorAt("useUrl", 1, 58, 1, 80)}},
		// explicit empty options.
		{Code: "__dirname + \"/x\";", Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 17)}},
		// escaped identifiers.
		{Code: "const p = r\\u0065quire(\"path\"); __dir\\u006e\\u0061me + p.s\\u0065p; import.meta.dir\\u006e\\u0061me + \"/x\";", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 33, 1, 65), concatErrorAt("usePathFunctions", 1, 67, 1, 103)}},
		// escaped path separators.
		{Code: "__dirname + \"\\u002fx\"; __filename + \"\\x2fx\"; `${__dirname}\\u002fx`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 22), concatErrorAt("usePathFunctions", 1, 24, 1, 44), concatErrorAt("usePathFunctions", 1, 46, 1, 67)}},
		// cooked line continuation.
		{Code: "__dirname + \"\\\n/x\"; `${__dirname}\\\n/x`;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 2, 4), concatErrorAt("usePathFunctions", 2, 6, 3, 4)}},
		// destructuring defaults and keys.
		{Code: "const {[__dirname + \"/x\"]: file = import.meta.dirname + \"/y\"} = object; [value = __filename + \"/z\"] = list;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 9, 1, 25), concatErrorAt("usePathFunctions", 1, 35, 1, 61), concatErrorAt("usePathFunctions", 1, 82, 1, 99)}},
		// inline globals.
		{Code: "/* global __dirname: readonly */\n__dirname + \"/x\";", Globals: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 2, 1, 2, 17)}},
		// JSDoc declarations are comments.
		{Code: "/** @typedef {string} __dirname */\n/** @import {default as p} from \"path\" */\n__dirname + \"/x\"; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 3, 1, 3, 17)}},
		// type-only path import.
		{Code: "import type p from \"path\"; __dirname + p.sep;", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 28, 1, 45)}},
		// type-only named default import.
		{Code: "import {type default as p} from \"path\"; __dirname + p.sep;", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 41, 1, 58)}},
		// static non-null suffix.
		{Code: "const suffix = \"/x\"; __dirname + suffix!;", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 22, 1, 41)}},
		// module object via destructured default.
		{Code: "import * as ns from \"path\"; const {default:p} = ns; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 53, 1, 70)}},
		// require module object through assignment pattern.
		{Code: "let p; ({p = require(\"path\")} = source); __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 42, 1, 59)}},
		// module object cycle.
		{Code: "let a = require(\"path\"); let b = a; a = b; __dirname + b.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 44, 1, 61)}},
		// optional static object.
		{Code: "const p = {sep:\"/\"}; __dirname + p?.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 22, 1, 40)}},
		// String.fromCharCode.
		{Code: "__dirname + String.fromCharCode(47);", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 36)}},
		// string slicing.
		{Code: "__dirname + \"x/dir\".slice(1);", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 29)}},
		// string concatenation method.
		{Code: "__dirname + \"/\".concat(\"dir\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 30)}},
		// array join.
		{Code: "__dirname + [\"\", \"dir\"].join(\"/\");", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 34)}},
		// template key in binding.
		{Code: "const {[`require`]: load} = global; __dirname + load(\"path\").sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 37, 1, 65)}},
		// template key in assignment default.
		{Code: "let load; ({[`require`]: load = fallback} = global); __dirname + load(\"path\").sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 54, 1, 82)}},
		// namespace default binding.
		{Code: "import * as ns from \"path\"; const {default: p = fallback} = ns; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 65, 1, 82)}},
		// mixed default and namespace import.
		{Code: "import p, * as ns from \"path\"; __dirname + p.sep; __filename + ns.default.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 32, 1, 49), concatErrorAt("usePathFunctions", 1, 51, 1, 78)}},
		// mixed default and named imports.
		{Code: "import p, {sep, default as other} from \"path\"; __dirname + p.sep; __dirname + sep; __filename + other.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 48, 1, 65), concatErrorAt("usePathFunctions", 1, 84, 1, 106)}},
		// string default import.
		{Code: "import {\"default\" as p} from \"path\"; __dirname + p.sep;", Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 38, 1, 55)}},
	})
}

func TestNoPathConcatPlatformSeparator(t *testing.T) {
	code := `__dirname + "\\file";`
	if os.PathSeparator == '\\' {
		runNoPathConcatTests(t, nil, []rule_tester.InvalidTestCase{
			{Code: code, Errors: []rule_tester.InvalidTestCaseError{concatErrorAt("usePathFunctions", 1, 1, 1, 21)}},
		})
	} else {
		runNoPathConcatTests(t, []rule_tester.ValidTestCase{{Code: code}}, nil)
	}
}
