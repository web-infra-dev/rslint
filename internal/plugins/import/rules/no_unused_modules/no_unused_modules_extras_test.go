package no_unused_modules

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const extrasTSConfig = `{
  "compilerOptions": {"allowJs": true, "jsx": "preserve", "module": "esnext", "moduleResolution": "bundler", "strict": true},
  "include": ["**/*"]
}`

// scenario is one isolated project under /extras/<name>/. files are relative
// to it; a tsconfig.json including every file is added unless given.
type scenario struct {
	name    string
	files   map[string]string
	valid   []rule_tester.ValidTestCase
	invalid []rule_tester.InvalidTestCase
}

func extrasPath(name, relative string) string {
	return "/extras/" + name + "/" + relative
}

func runScenario(t *testing.T, s scenario) {
	t.Helper()
	overlays := map[string]string{extrasPath(s.name, "tsconfig.json"): extrasTSConfig}
	for relative, text := range s.files {
		overlays[extrasPath(s.name, relative)] = text
	}
	t.Run(s.name, func(t *testing.T) {
		rule_tester.RunRuleTester(fixtureRoot(t, overlays), extrasPath(s.name, "tsconfig.json")[1:], t, &NoUnusedModulesRule, s.valid, s.invalid)
	})
}

func srcOptions(name string, extra map[string]any) map[string]any {
	options := map[string]any{"unusedExports": true, "src": []any{extrasPath(name, "src")}}
	for key, value := range extra {
		options[key] = value
	}
	return options
}

func TestNoUnusedModulesImportForms(t *testing.T) {
	const exporter = "export const a = 1; export const b = 2; export default 3;"
	allUnused := []rule_tester.InvalidTestCaseError{
		unusedRange("a", 1, 1, 1, 20), unusedRange("b", 1, 21, 1, 40), unusedRange("default", 1, 41, 1, 58),
	}
	cases := []struct {
		name     string
		importer string
		errors   []rule_tester.InvalidTestCaseError
	}{
		// An alias still uses the original export name.
		{"named-alias", `import { a as alias } from './a';`, allUnused[1:]},
		{"default", `import value from './a';`, allUnused[:2]},
		// The namespace binding need not be referenced, and covers default.
		{"namespace", `import * as ns from './a';`, nil},
		{"side-effect", `import './a';`, allUnused},
		{"empty-named", `import {} from './a';`, allUnused},
		{"dynamic-literal", `async function f() { await import('./a'); }`, nil},
		{"dynamic-template", "async function f() { await import(`./a`); }", allUnused},
		{"type-import", `import type { a } from './a';`, allUnused[1:]},
		{"inline-type-import", `import { type a, b } from './a';`, allUnused[2:]},
		// Upstream keys default imports apart from the name `default`, so it
		// misses this spelling of a default import.
		{"named-default", `import { default as value } from './a';`, allUnused[:2]},
		{"string-name", `import { "a" as value } from './a';`, allUnused[1:]},
		{"star-reexport", `export * from './a';`, allUnused[2:]},
		{"star-as-reexport", `export * as ns from './a';`, allUnused[2:]},
		{"named-reexport", `export { a, default as d } from './a';`, allUnused[1:2]},
		{"type-reexport", `export type { b } from './a';`, []rule_tester.InvalidTestCaseError{allUnused[0], allUnused[2]}},
		// TypeScript's CommonJS import binds the whole module.
		{"import-equals", `import a = require('./a');`, nil},
		// Static require, which upstream documents but does not count.
		{"require-destructure", `const { a, b: renamed } = require('./a');`, allUnused[2:]},
		{"require-member", `const value = require('./a').b;`, []rule_tester.InvalidTestCaseError{allUnused[0], allUnused[2]}},
		{"require-element", `const value = require('./a')['b'];`, []rule_tester.InvalidTestCaseError{allUnused[0], allUnused[2]}},
		{"require-default-member", `const value = require('./a').default;`, allUnused[:2]},
		{"require-binding", `const all = require('./a');`, nil},
		{"require-rest", `const { a, ...rest } = require('./a');`, nil},
		{"require-computed", `const key = 'a'; const { [key]: value } = require('./a');`, nil},
		{"require-argument", `use(require('./a'));`, nil},
		{"require-parenthesized", `const value = (require('./a')).a;`, allUnused[1:]},
		{"require-statement", `require('./a');`, allUnused},
		{"require-template", "const { a } = require(`./a`);", allUnused},
		// A local `require` may be anything; its calls are not module requests.
		{"require-shadowed-parameter", `function load(require) { return require('./a'); }`, allUnused},
		{"require-shadowed-declaration", `const require = () => ({}); const { a } = require('./a');`, allUnused},
	}
	for _, c := range cases {
		name := "form-" + c.name
		s := scenario{name: name, files: map[string]string{"src/use.ts": c.importer}}
		file := extrasPath(name, "src/a.ts")
		if len(c.errors) == 0 {
			s.valid = []rule_tester.ValidTestCase{{Options: srcOptions(name, nil), Code: exporter, FileName: file}}
		} else {
			s.invalid = []rule_tester.InvalidTestCase{{Options: srcOptions(name, nil), Code: exporter, FileName: file, Errors: c.errors}}
		}
		runScenario(t, s)
	}
}

func TestNoUnusedModulesAnalyzedFiles(t *testing.T) {
	// A linted file outside src never adds usage. Upstream's result here
	// depends on whether outside/use.ts was linted first in the same process.
	runScenario(t, scenario{
		name: "outside-src-importer",
		files: map[string]string{
			"outside/use.ts": `import { a } from '../src/a';`,
		},
		invalid: []rule_tester.InvalidTestCase{{
			Options: srcOptions("outside-src-importer", nil), Code: `export const a = 1;`,
			FileName: extrasPath("outside-src-importer", "src/a.ts"),
			Errors:   []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1)},
		}},
		valid: []rule_tester.ValidTestCase{
			// The importer itself is linted outside src: nothing to check.
			{Options: srcOptions("outside-src-importer", nil), Code: `import { a } from '../src/a'; export const unused = 1;`, FileName: extrasPath("outside-src-importer", "outside/use.ts")},
		},
	})

	// ignoreExports exempts a file's exports, not its imports.
	runScenario(t, scenario{
		name:  "ignored-importer",
		files: map[string]string{"src/entry.ts": `import { a } from './a'; export const entry = 1;`},
		valid: []rule_tester.ValidTestCase{
			{Options: srcOptions("ignored-importer", map[string]any{"ignoreExports": []any{extrasPath("ignored-importer", "src/entry.ts")}}), Code: `export const a = 1;`, FileName: extrasPath("ignored-importer", "src/a.ts")},
			{Options: srcOptions("ignored-importer", map[string]any{"ignoreExports": []any{extrasPath("ignored-importer", "src/entry.ts")}}), Code: `import { a } from './a'; export const entry = 1;`, FileName: extrasPath("ignored-importer", "src/entry.ts")},
			// A relative glob resolves against the working directory.
			{Options: srcOptions("ignored-importer", map[string]any{"ignoreExports": []any{"extras/ignored-importer/src/entry.*"}}), Code: `import { a } from './a'; export const entry = 1;`, FileName: extrasPath("ignored-importer", "src/entry.ts")},
		},
		invalid: []rule_tester.InvalidTestCase{
			// ignoreExports only affects unusedExports. Upstream also skips
			// missingExports for these files, but only when unusedExports is on.
			{
				Options: map[string]any{"missingExports": true, "unusedExports": true, "src": []any{extrasPath("ignored-importer", "src")}, "ignoreExports": []any{extrasPath("ignored-importer", "src/empty.ts")}},
				Code:    `const local = 1;`, FileName: extrasPath("ignored-importer", "src/empty.ts"),
				Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 17)},
			},
		},
	})

	// import/ignore removes an importer's requests, as upstream builds no
	// export map for it.
	runScenario(t, scenario{
		name:  "import-ignore-importer",
		files: map[string]string{"src/use.ts": `import { a } from './a';`},
		invalid: []rule_tester.InvalidTestCase{{
			Options: srcOptions("import-ignore-importer", nil), Settings: map[string]any{"import/ignore": []any{"use\\.ts$"}},
			Code: `export const a = 1;`, FileName: extrasPath("import-ignore-importer", "src/a.ts"),
			Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1)},
		}},
	})

	// Without import/extensions, every JavaScript and TypeScript extension is
	// analyzed. Upstream analyzes only .js/.mjs/.cjs by default.
	runScenario(t, scenario{
		name: "extensions",
		files: map[string]string{
			"src/use.tsx":   `import { a } from './a';`,
			"src/use.mts":   `import { b } from './a';`,
			"src/use.cjs":   `const { c } = require('./a');`,
			"src/use.d.ts":  `import type { d } from './a';`,
			"src/notes.txt": `import { e } from './a';`,
		},
		valid: []rule_tester.ValidTestCase{
			{Options: srcOptions("extensions", nil), Code: `export const a = 1, b = 2, c = 3, d = 4;`, FileName: extrasPath("extensions", "src/a.ts")},
		},
		invalid: []rule_tester.InvalidTestCase{
			{Options: srcOptions("extensions", nil), Code: `export const e = 1;`, FileName: extrasPath("extensions", "src/a.ts"), Errors: []rule_tester.InvalidTestCaseError{unusedError("e", 1, 1)}},
			// An explicit list restricts the analyzed files.
			{
				Options: srcOptions("extensions", nil), Settings: map[string]any{"import/extensions": []any{".ts", ".cjs"}},
				Code: `export const a = 1, b = 2, c = 3, d = 4;`, FileName: extrasPath("extensions", "src/a.ts"),
				Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1), unusedError("b", 1, 1)},
			},
			// An exporter outside the list is not analyzed either.
			{
				Options: srcOptions("extensions", nil), Settings: map[string]any{"import/extensions": []any{".js"}},
				Code: `export const a = 1;`, FileName: extrasPath("extensions", "src/a.js"),
				Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1)},
			},
		},
	})

	// Globs select by pattern alone; directories by extension. Dotfiles and
	// node_modules are skipped as ESLint does by default.
	runScenario(t, scenario{
		name: "patterns",
		files: map[string]string{
			"src/lib/use.ts":              `import { a } from '../a';`,
			"src/.hidden/use.ts":          `import { b } from '../a';`,
			"src/node_modules/pkg/use.ts": `import { c } from '../../a';`,
			"src/.use.ts":                 `import { d } from './a';`,
		},
		invalid: []rule_tester.InvalidTestCase{
			{
				Options: srcOptions("patterns", nil), Code: `export const a = 1, b = 2, c = 3, d = 4;`, FileName: extrasPath("patterns", "src/a.ts"),
				Errors: []rule_tester.InvalidTestCaseError{unusedError("b", 1, 1), unusedError("c", 1, 1), unusedError("d", 1, 1)},
			},
			{
				Options: map[string]any{"unusedExports": true, "src": []any{extrasPath("patterns", "src/*.ts"), "extras/patterns/src/lib/*.ts"}},
				Code:    `export const a = 1, b = 2;`, FileName: extrasPath("patterns", "src/a.ts"),
				Errors: []rule_tester.InvalidTestCaseError{unusedError("b", 1, 1)},
			},
			{
				// A single-level glob does not reach src/lib.
				Options: map[string]any{"unusedExports": true, "src": []any{extrasPath("patterns", "src/*.ts")}},
				Code:    `export const a = 1;`, FileName: extrasPath("patterns", "src/a.ts"),
				Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1)},
			},
			{
				// Naming files: the exporter is analyzed, its importer is not.
				Options: map[string]any{"unusedExports": true, "src": []any{extrasPath("patterns", "src/a.ts"), extrasPath("patterns", "src/missing.ts")}},
				Code:    `export const a = 1;`, FileName: extrasPath("patterns", "src/a.ts"),
				Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1)},
			},
		},
	})

	// A file no Program owns is parsed alone and resolved without a project.
	runScenario(t, scenario{
		name: "unowned-importer",
		files: map[string]string{
			"tsconfig.json": `{"compilerOptions": {"module": "esnext", "moduleResolution": "bundler"}, "include": ["src/**/*.ts"]}`,
			"src/use.js":    `import { a } from './a';`,
		},
		valid: []rule_tester.ValidTestCase{
			{Options: srcOptions("unowned-importer", nil), Code: `export const a = 1;`, FileName: extrasPath("unowned-importer", "src/a.ts")},
		},
	})
}

func TestNoUnusedModulesExportShapes(t *testing.T) {
	name := "export-shapes"
	file := extrasPath(name, "src/a.ts")
	opts := srcOptions(name, nil)
	runScenario(t, scenario{
		name:  name,
		files: map[string]string{"src/use.ts": `import { used } from './a';`},
		valid: []rule_tester.ValidTestCase{
			{Options: opts, Code: `export const used = 1;`, FileName: file},
			// `export =` and `export import` are not checked, as upstream.
			{Options: opts, Code: `const value = 1; export = value;`, FileName: file},
			{Options: opts, Code: `namespace N { export const x = 1; } export import X = N.x;`, FileName: file},
			{Options: opts, Code: `export * from './b'; export * as ns from './b';`, FileName: file},
			{Options: opts, Code: `export {};`, FileName: file},
		},
		invalid: []rule_tester.InvalidTestCase{
			// Upstream crashes on the hole and reports `undefined` for the
			// default and nested elements.
			{Options: opts, Code: `export const [, hole, withDefault = 1, [nested], { key: renamed, ...rest }] = values;`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("hole", 1, 1, 1, 86), unusedError("withDefault", 1, 1), unusedError("nested", 1, 1), unusedError("renamed", 1, 1), unusedError("rest", 1, 1),
			}},
			{Options: opts, Code: "export let a = 1, b;\nexport var c;", FileName: file, Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("a", 1, 1, 1, 21), unusedRange("b", 1, 1, 1, 21), unusedRange("c", 2, 1, 2, 14),
			}},
			{Options: opts, Code: `export async function* gen() {} export abstract class Base {}`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("gen", 1, 1, 1, 32), unusedRange("Base", 1, 33, 1, 62),
			}},
			{Options: opts, Code: `export default class {}`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("default", 1, 1, 1, 24)}},
			{Options: opts, Code: `const value = 1; export { value as default };`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("default", 1, 27, 1, 43)}},
			// Only the namespace is a module export; upstream reports `a`
			// inside it instead.
			{Options: opts, Code: `export namespace N { export const a = 1; }`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("N", 1, 1, 1, 43)}},
			{Options: opts, Code: `export declare namespace Outer.Inner { const a: number; }`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedError("Outer", 1, 1)}},
			// Upstream misses ambient functions; overloads report once, on
			// the implementation.
			{Options: opts, Code: `export declare function f(): void;`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("f", 1, 1, 1, 35)}},
			{Options: opts, Code: "export declare function f(): void;\nexport declare function f(a: string): void;", FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("f", 1, 1, 1, 35)}},
			{Options: opts, Code: "export function f(): void;\nexport function f(a?: string) {}", FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("f", 2, 1, 2, 33)}},
			{Options: opts, Code: `export declare const a: number; export declare class C {} export const enum E { A }`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("a", 1, 1, 1, 32), unusedRange("C", 1, 33, 1, 58), unusedRange("E", 1, 59, 1, 84),
			}},
			// Multiline and non-ASCII positions count UTF-16 columns.
			{Options: opts, Code: "const ü = 1, 𝒳 = 2;\nexport {\n  ü,\n  𝒳 as wide,\n};", FileName: file, Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("ü", 3, 3, 3, 4), unusedRange("wide", 4, 3, 4, 13),
			}},
			{Options: opts, Code: `export { local } from './b'; export { "string name" as other } from './b';`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("local", 1, 10, 1, 15), unusedRange("other", 1, 39, 1, 61),
			}},
		},
	})
}

func TestNoUnusedModulesTypeExports(t *testing.T) {
	name := "type-exports"
	file := extrasPath(name, "src/a.ts")
	ignoreTypes := srcOptions(name, map[string]any{"ignoreUnusedTypeExports": true})
	runScenario(t, scenario{
		name: name,
		valid: []rule_tester.ValidTestCase{
			{Options: ignoreTypes, Code: `export interface I {} export type T = 1; export enum E { A } export declare enum D { A }`, FileName: file},
			// Upstream exempts only direct declarations; an export list of
			// the same local types is exempt here too.
			{Options: ignoreTypes, Code: `interface I {} type T = 1; enum E { A } export { I, T, E };`, FileName: file},
			{Options: ignoreTypes, Code: `interface I {} export type { I }; export { type I as J };`, FileName: file},
			{Options: ignoreTypes, Code: `export type { Remote } from './b'; export { type Other } from './b';`, FileName: file},
		},
		invalid: []rule_tester.InvalidTestCase{
			{Options: ignoreTypes, Code: `const v = 1; export { v };`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedRange("v", 1, 23, 1, 24)}},
			// A name that is also a value, or an imported binding, may be a value.
			{Options: ignoreTypes, Code: `interface I {} const I = 1; export { I };`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedError("I", 1, 38)}},
			{Options: ignoreTypes, Code: `import { I } from './b'; export { I };`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedError("I", 1, 35)}},
			{Options: ignoreTypes, Code: `export { Remote } from './b';`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedError("Remote", 1, 10)}},
			{Options: ignoreTypes, Code: `export namespace N {} export default interface I {}`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedError("N", 1, 1), unusedError("default", 1, 23)}},
			// Without the option, types are reported like values.
			{Options: srcOptions(name, nil), Code: `interface I {} export type { I };`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{unusedError("I", 1, 30)}},
		},
	})
}

func TestNoUnusedModulesMissingExports(t *testing.T) {
	file := fixturePath("missing.ts")
	opts := map[string]any{"missingExports": true}
	runGroup(t, "missing exports extras", nil, []rule_tester.ValidTestCase{
		// Upstream reports each of these: it does not count `export *`,
		// namespaces, ambient functions or TypeScript-only export forms.
		{Options: opts, Code: `export * from './a';`, FileName: file},
		{Options: opts, Code: `export * as ns from './a';`, FileName: file},
		{Options: opts, Code: `export namespace N { export const a = 1; }`, FileName: file},
		{Options: opts, Code: `export declare function f(): void;`, FileName: file},
		{Options: opts, Code: `const value = 1; export = value;`, FileName: file},
		{Options: opts, Code: `namespace N { export const x = 1; } export import X = N.x;`, FileName: file},
		{Options: opts, Code: `export interface I {}`, FileName: file},
		{Options: opts, Code: `export const [, a] = [0, 1];`, FileName: file},
		{Options: opts, Code: `export default interface I {}`, FileName: file},
		// No options and unusedExports-only configurations never check this.
		{Code: `const a = 1;`, FileName: file},
		{Options: map[string]any{"unusedExports": true, "src": []any{fixturePath("missing.ts")}}, Code: `const a = 1;`, FileName: file},
	}, []rule_tester.InvalidTestCase{
		{Options: opts, Code: `export {}`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 10)}},
		{Options: opts, Code: `export as namespace X;`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 23)}},
		{Options: opts, Code: `"use strict"; const a = 1`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 14)}},
		{Options: opts, Code: `import './side-effect';`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 24)}},
		// An `export` inside a namespace does not export from the module.
		{Options: opts, Code: `namespace N { export const a = 1; }`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 36)}},
		{Options: opts, Code: `module.exports = {};`, FileName: fixturePath("missing.js"), Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 21)}},
		{Options: opts, Code: ``, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 1)}},
		{Options: opts, Code: "\n\n// only a comment\n", FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(4, 1, 4, 1)}},
		{Options: opts, Code: "#!/usr/bin/env node\nconst a = 1;", FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(2, 1, 2, 13)}},
		// Both options report independently.
		{
			Options: map[string]any{"missingExports": true, "unusedExports": true, "src": []any{fixturePath("missing.ts")}},
			Code:    `const a = 1;`, FileName: file, Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 13)},
		},
	})
}

func TestNoUnusedModulesPackageEntries(t *testing.T) {
	entry := func(name, pkg string, valid []string, invalid []string) {
		s := scenario{name: name, files: map[string]string{"package.json": pkg}}
		opts := map[string]any{"unusedExports": true, "src": []any{extrasPath(name, "")}}
		for _, file := range valid {
			s.valid = append(s.valid, rule_tester.ValidTestCase{Options: opts, Code: `export const a = 1;`, FileName: extrasPath(name, file)})
		}
		for _, file := range invalid {
			s.invalid = append(s.invalid, rule_tester.InvalidTestCase{Options: opts, Code: `export const a = 1;`, FileName: extrasPath(name, file), Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 1)}})
		}
		runScenario(t, s)
	}
	// Upstream exempts bin, browser and main only; this port also exempts
	// every published `exports` target.
	entry("exports-string", `{"exports": "./lib/index.js"}`, []string{"lib/index.js"}, []string{"lib/other.js"})
	entry("exports-conditions", `{"exports": {".": {"types": "./lib/index.d.ts", "import": "./lib/index.mjs", "require": ["./lib/index.cjs"]}, "./feature": {"node": "./lib/feature.js", "default": null}}}`,
		[]string{"lib/index.mjs", "lib/index.cjs", "lib/feature.js"}, []string{"lib/other.mjs"})
	entry("exports-pattern", `{"exports": {"./features/*": "./src/features/*.js", "./internal/*": null}}`,
		[]string{"src/features/a.js", "src/features/deep/b.js"}, []string{"src/features/a.ts", "src/other.js"})
	entry("exports-private", `{"private": true, "exports": "./index.js", "main": "./index.js"}`, nil, []string{"index.js"})
	// `module` is a bundler convention, not a published Node entry point.
	entry("module-field", `{"module": "./index.mjs", "main": "./index.cjs"}`, []string{"index.cjs"}, []string{"index.mjs"})
	// A main path names one exact file, without extension resolution.
	entry("main-exact", `{"main": "./index"}`, nil, []string{"index.js"})
}
