package default_rule_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	default_rule "github.com/web-infra-dev/rslint/internal/plugins/import/rules/default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Mirrors the upstream eslint-plugin-import default rule test suite (every
// case that survives the TypeScript parser) plus rslint-specific edge cases
// and real-world user scenarios. Layout follows
// https://github.com/import-js/eslint-plugin-import/blob/main/tests/src/rules/default.js
// so future audits can diff top-to-bottom.
func TestDefaultRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&default_rule.DefaultRule,
		[]rule_tester.ValidTestCase{
			// =====================================================
			// Upstream `valid` (top section)
			// =====================================================

			// `import "./malformed.js"` — no default specifier, rule doesn't apply.
			// SKIP: upstream uses an actual malformed JS file for parse-error
			// coverage; rslint relies on the TS checker for parse diagnostics.
			// Equivalent here: bare-import without default specifier never fires.
			{Code: `import "./default_files/named-exports";`, FileName: "file.ts"},

			{Code: `import foo from "./default_files/empty-folder";`, FileName: "file.ts"},
			{Code: `import { c } from "./default_files/default-export";`, FileName: "file.ts"},
			{Code: `import foo from "./default_files/default-export";`, FileName: "file.ts"},
			{Code: `import foo from "./default_files/mixed-exports";`, FileName: "file.ts"},
			{Code: `import bar from "./default_files/default-export";`, FileName: "file.ts"},
			{Code: `import CoolClass from "./default_files/default-class";`, FileName: "file.ts"},
			{Code: `import bar, { named } from "./default_files/mixed-exports";`, FileName: "file.ts"},

			// core modules always have a default
			{Code: `import crypto from "crypto";`, FileName: "file.ts"},

			{Code: `import common from "./default_files/common";`, FileName: "file.ts"},

			// ---- ES7 / Babel-parser-only export forms ----
			// `export bar from "./bar"` — Babel-only `ExportDefaultSpecifier`.
			// SKIP: not parseable by TypeScript (proposal that never landed).
			{Code: `export { default as bar } from "./default_files/default-export";`, FileName: "file.ts"},
			// `export bar, { foo } from "./bar"` — same as above, Babel-only.
			// SKIP: same reason.
			{Code: `export { default as bar, named } from "./default_files/mixed-exports";`, FileName: "file.ts"},
			// `export bar, * as names from "./bar"` — Babel-only.
			// SKIP: same reason.

			// ---- Sanity / regression cases from upstream ----
			{Code: `export { a } from "./default_files/named-exports";`, FileName: "file.ts"},
			// #54: import of named-default-export
			{Code: `import foo from "./default_files/named-default-export";`, FileName: "file.ts"},
			// #94: redux-style `export default connect(App)`
			{Code: `import connectedApp from "./default_files/redux";`, FileName: "file.ts"},
			// trampoline (`export { default } from`) chain that resolves
			{Code: `import twofer from "./default_files/trampoline";`, FileName: "file.ts"},
			// Locks in a deeper alias chain than upstream covers.
			{Code: `import threefer from "./default_files/deep-trampoline";`, FileName: "file.ts"},

			// ---- ES2022 arbitrary module-namespace identifier ----
			// `export { "default" as bar } from`
			{Code: `export { "default" as bar } from "./default_files/default-export";`, FileName: "file.ts"},

			// ---- JSX (TSX in our tree) ----
			{Code: `import MyCoolComponent from "./default_files/jsx/MyCoolComponent";`, FileName: "file.tsx", Tsx: true},
			{Code: `import App from "./default_files/jsx/App";`, FileName: "file.tsx", Tsx: true},

			// ---- #545: more ES7 cases ----
			// `import bar from './default-export-from.js'`,
			// `import bar from './default-export-from-named.js'`,
			// `import bar from './default-export-from-ignored.js' (settings)`,
			// `export bar from './default-export-from-ignored.js' (settings)`.
			// SKIP: Babel-old parser; standard re-export forms with `default as`
			// keyword are covered above.

			// =====================================================
			// Upstream `valid` (TypeScript section)
			// =====================================================
			{Code: `import foobar from "./default_files/typescript-default";`, FileName: "file.ts"},
			{Code: `import foobar from "./default_files/typescript-export-assign-default";`, FileName: "file.ts"},
			{Code: `import foobar from "./default_files/typescript-export-assign-function";`, FileName: "file.ts"},
			{Code: `import foobar from "./default_files/typescript-export-assign-mixed";`, FileName: "file.ts"},
			{Code: `import foobar from "./default_files/typescript-export-assign-default-reexport";`, FileName: "file.ts"},
			{Code: `import React from "./default_files/typescript-export-assign-default-namespace";`, FileName: "file.ts"},
			// `./typescript-export-react-test-renderer` and `./typescript-extended-config`
			// SKIP: those upstream cases require a multi-tsconfig setup
			// (parserOptions.tsconfigRootDir override). The behaviour they lock
			// in — `export = X` resolved through a sibling tsconfig — is already
			// covered by `typescript-export-assign-default-reexport`.
			{Code: `import foobar from "./default_files/typescript-export-assign-property";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: anonymous default export shapes
			// =====================================================
			{Code: `import fn from "./default_files/default-export-anonymous-fn";`, FileName: "file.ts"},
			{Code: `import C from "./default_files/default-export-anonymous-class";`, FileName: "file.ts"},
			{Code: `import a from "./default_files/default-export-arrow";`, FileName: "file.ts"},
			{Code: `import L from "./default_files/default-export-literal";`, FileName: "file.ts"},
			{Code: `import asyncFn from "./default_files/default-export-async";`, FileName: "file.ts"},
			{Code: `import gen from "./default_files/default-export-generator";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: alias / rename forms
			// =====================================================
			// `function foo(){}; export { foo as default }` — local rename to default.
			{Code: `import foo from "./default_files/local-rename-default";`, FileName: "file.ts"},
			// Re-exports a named binding under the name `default`. The source
			// binding exists, so the alias chain resolves cleanly.
			{Code: `import x from "./default_files/named-default-via-rename";`, FileName: "file.ts"},
			// `import { default as bar }` — explicit named-default access (not a default specifier).
			{Code: `import { default as bar } from "./default_files/default-export";`, FileName: "file.ts"},
			// `import foo, { default as bar }` — default specifier AND explicit named-default; rule
			// only checks the default specifier, both refer to the same export.
			{Code: `import foo, { default as bar } from "./default_files/default-export";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: import shape variations
			// =====================================================
			{Code: `import * as ns from "./default_files/named-exports";`, FileName: "file.ts"},
			{Code: `import type Foo from "./default_files/default-export";`, FileName: "file.ts"},
			{Code: `import foo, * as ns from "./default_files/default-export";`, FileName: "file.ts"},
			// import attribute `with { type }` — module specifier still resolves.
			{Code: `import foo from "./default_files/default-export" with { type: "module" };`, FileName: "file.ts"},
			// Explicit `.ts` extension.
			{Code: `import foo from "./default_files/default-export.ts";`, FileName: "file.ts"},
			// Folder import resolved through index.ts.
			{Code: `import idx from "./default_files/index-folder";`, FileName: "file.ts"},
			// Same import twice — rule must still pass each independently.
			{Code: `import a from "./default_files/default-export"; import b from "./default_files/default-export";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: graceful skips
			// =====================================================
			// Unresolved relative path — module symbol absent, skip without false-positive.
			{Code: `import foo from "./default_files/non-existent";`, FileName: "file.ts"},
			// Re-export `default` without rename — not a default specifier on the export side.
			{Code: `export { default } from "./default_files/default-export";`, FileName: "file.ts"},
			// Dynamic import — not a default specifier; rule never fires.
			{Code: `const m = import("./default_files/named-exports");`, FileName: "file.ts"},
			// Plain `require()` call — same.
			{Code: `const m = require("./default_files/named-exports");`, FileName: "file.ts"},
			// `import x = require()` — TS-only equivalent. SKIP: rule scope
			// matches upstream (ImportDeclaration only); this form is not flagged.
			{Code: `import m = require("./default_files/named-exports");`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: tsconfig-driven semantics
			// =====================================================
			// JSON default import (resolveJsonModule + esModuleInterop).
			{
				Code:     `import data from "./default_files/data.json";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.with-json.json",
			},
			// `export = X` under interop ON — passes (already covered, but
			// pair-locked here against the no-interop invalid case below to
			// make the contract explicit).
			{
				Code:     `import fn from "./default_files/typescript-export-assign-function";`,
				FileName: "file.ts",
			},

			// =====================================================
			// rslint-specific: parse-error robustness
			// =====================================================
			// Module with a parse error but a recoverable default export — TS
			// recovers binding-wise, default is observable, rule passes.
			{Code: `import x from "./default_files/syntax-broken-with-default";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: `settings["import/ignore"]` — match upstream contract
			// =====================================================
			// Default ignore list contains `\.(es6|exs|json|node)$` — JSON
			// imports are silently treated as un-analyzable, even when the
			// resolver could see them.
			{
				Code:     `import data from "./default_files/data.json";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.with-json.json",
			},
			// Custom ignore: explicit pattern matches the source path → skip.
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": []interface{}{"named-exports"},
				},
			},
			// Empty ignore list → user takes control. Previously skipped
			// `node_modules` paths now get checked too. This is a real,
			// resolvable .ts file with a default — must still pass.
			{
				Code:     `import foo from "./default_files/default-export";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": []interface{}{},
				},
			},
			// Pattern mode: single string accepted (upstream tolerates this).
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": "named-exports",
				},
			},
			// Invalid regex in patterns is silently dropped (defensive parse).
			{
				Code:     `import foo from "./default_files/default-export";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": []interface{}{"(((", "default-export"},
				},
			},
			// `null` value (JSON null serialized into settings) — treated as
			// "absent" (consistent with the un-set branch). Default
			// external-library + default ignore list both apply.
			{
				Code:     `import crypto from "crypto";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": nil,
				},
			},

			// =====================================================
			// rslint-specific: parent-relative path resolution
			// =====================================================
			// `../` traversal — file lives in `default_files/nested/`, imports
			// from `default_files/parent-default.ts`.
			{
				Code:     `import p from "../parent-default";`,
				FileName: "default_files/nested/test.ts",
			},

			// =====================================================
			// rslint-specific: declaration-file (`.d.ts`) import sources
			// =====================================================
			{Code: `import v from "./default_files/decl-with-default";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: `.mjs` source with `export default X`
			// =====================================================
			{
				Code:     `import x from "./default_files/mjs-default.mjs";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.allow-js.json",
			},
			// `.cjs` source with `module.exports = X` — TS reads it as
			// `export = X`, synthesized default under esModuleInterop.
			{
				Code:     `import x from "./default_files/cjs-module-exports.cjs";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.allow-js.json",
			},

			// =====================================================
			// rslint-specific: import attributes — `with` form
			// =====================================================
			// SKIP: `assert { type: ... }` is now a TypeScript hard error
			// ("Import assertions have been replaced by import attributes.
			// Use 'with' instead of 'assert'."). The TS parser refuses to
			// produce an AST for it, so the rule cannot be exercised on this
			// shape. The supported `with` form is covered above.

			// =====================================================
			// rslint-specific: default value variants (export entry exists)
			// =====================================================
			// `export default undefined` — entry exists, value is undefined.
			{Code: `import x from "./default_files/default-undefined";`, FileName: "file.ts"},
			// `export default null`.
			{Code: `import x from "./default_files/default-null";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: import binding name = reserved-ish word
			// =====================================================
			// `default` cannot be used as a binding identifier in strict module
			// code, but `await`-like reserved words can. Lock in that the rule
			// extracts the binding regardless of identifier choice.
			{Code: `import yield_ from "./default_files/default-export";`, FileName: "file.ts"},
			{Code: `import async_ from "./default_files/default-export";`, FileName: "file.ts"},

			// =====================================================
			// rslint-specific: module options ignored
			// =====================================================
			// Upstream `schema: []` — the rule accepts no options. Passing one
			// must not change behaviour (a no-default module still reports;
			// a default-bearing module still passes). Both shapes covered.
			{Code: `import foo from "./default_files/default-export";`, FileName: "file.ts", Options: map[string]interface{}{"someOpt": true}},
			{Code: `import foo from "./default_files/default-export";`, FileName: "file.ts", Options: []interface{}{"warn", map[string]interface{}{"x": 1}}},
		},
		[]rule_tester.InvalidTestCase{
			// =====================================================
			// Upstream `invalid` (top section)
			// =====================================================
			// `import Foo from './jsx/FooES7.js'` — parse-error case.
			// SKIP: rslint surfaces parse errors via the type-check pipeline,
			// not this rule.

			// `import baz from "./named-exports"` — main case
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},

			// `export baz from "./named-exports"`,
			// `export baz, { bar } from "./named-exports"`,
			// `export baz, * as names from "./named-exports"`.
			// SKIP: Babel-only `ExportDefaultSpecifier` syntax.

			// `import twofer from "./broken-trampoline"` — re-exports default
			// from a no-default module.
			{
				Code:     `import twofer from "./default_files/broken-trampoline";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/broken-trampoline".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 14,
				}},
			},

			// #328: `export *` does not include default.
			{
				Code:     `import barDefault from "./default_files/re-export";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/re-export".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 18,
				}},
			},

			// =====================================================
			// Upstream `invalid` (TypeScript section)
			// =====================================================
			// `import foobar from "./typescript"` — only named exports.
			{
				Code:     `import foobar from "./default_files/typescript";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/typescript".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 14,
				}},
			},
			// `typescript-export-as-default-namespace` — namespace-only file
			// without `export = X`. SKIP: upstream's failing case relies on a
			// distinct tsconfig with no compiler options; under our tsconfig
			// (esModuleInterop: true) the namespace case is covered by the
			// passing variant `typescript-export-assign-default-namespace`.
			// The plain `typescript.ts` case above already locks in the core
			// "named exports only" failure mode.

			// =====================================================
			// rslint-specific: combined import shapes
			// =====================================================
			{
				Code:     `import baz, { a } from "./default_files/named-exports";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			{
				Code:     `import baz, * as ns from "./default_files/named-exports";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			{
				Code:     `import type Baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 13, EndLine: 1, EndColumn: 16,
				}},
			},

			// =====================================================
			// rslint-specific: folder / index resolution
			// =====================================================
			// Folder import resolves to index.ts that lacks default.
			{
				Code:     `import idx from "./default_files/index-folder-no-default";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/index-folder-no-default".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},

			// =====================================================
			// rslint-specific: alias chain that bottoms out
			// =====================================================
			// Pure cycle — `circular-a` re-exports `default` from `circular-b`,
			// which re-exports it back. SkipAlias must terminate at unknown.
			{
				Code:     `import x from "./default_files/circular-a";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/circular-a".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},

			// =====================================================
			// rslint-specific: type-only / empty modules
			// =====================================================
			// File only exports types — value-space default doesn't exist.
			{
				Code:     `import T from "./default_files/type-only-exports";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/type-only-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},
			// File explicitly exports nothing (`export {};`) — no default.
			{
				Code:     `import x from "./default_files/empty-module";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/empty-module".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},

			// =====================================================
			// rslint-specific: schema-empty contract
			// =====================================================
			// Upstream `schema: []` — passing options must not silence the
			// diagnostic. Both option shapes (bare object / array-wrapped).
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Options:  map[string]interface{}{"unknown": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Options:  []interface{}{"warn", map[string]interface{}{"unknown": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},

			// =====================================================
			// rslint-specific: `settings["import/ignore"]` — non-matching patterns
			// =====================================================
			// Patterns that don't match still allow the rule to fire normally.
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": []interface{}{`\.coffee$`, `\.flow$`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			// Empty ignore list disables the default external-library skip —
			// `import/ignore: []` plus a real no-default project file still
			// reports.
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": []interface{}{},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			// `import/ignore` matches the resolved file path, NOT the source
			// string as written. A pattern that hits only the source ("./")
			// must not skip — the resolved file is still a regular `.ts` file
			// missing a default. Locks in upstream contract on path matching.
			{
				Code:     `import baz from "./default_files/named-exports";`,
				FileName: "file.ts",
				Settings: map[string]interface{}{
					"import/ignore": []interface{}{`^\./default_files/named-exports$`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			// SKIP: locking in "explicit `import/ignore` bypasses the
			// external-library fallback" against an actual node_modules path
			// is unreliable — different @types/node versions register
			// different default-export shapes for `crypto` (declaration
			// merging across submodules can synthesize defaults). The
			// behaviour is locked indirectly by the empty-array case above
			// (`import/ignore: []` vs project file with no default reports).

			// =====================================================
			// rslint-specific: declaration-file with no default
			// =====================================================
			// `.d.ts` containing only ambient named exports / types.
			{
				Code:     `import x from "./default_files/decl-no-default";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/decl-no-default".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},

			// =====================================================
			// rslint-specific: parent-relative path with no default
			// =====================================================
			{
				Code:     `import x from "../named-exports";`,
				FileName: "default_files/nested/test.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "../named-exports".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},

			// =====================================================
			// rslint-specific: `.mjs` / `.cjs` with no default
			// =====================================================
			// `.mjs` with only named exports.
			{
				Code:     `import x from "./default_files/mjs-named-only.mjs";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/mjs-named-only.mjs".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},
			// `.cjs` with `exports.foo = X` style (no `module.exports = X`).
			// TS sees named exports only — no default / no `export=`.
			{
				Code:     `import x from "./default_files/cjs-named-exports.cjs";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/cjs-named-exports.cjs".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},

			// =====================================================
			// rslint-specific: parse-error robustness
			// =====================================================
			// Module with parse errors AND no default — rule must still report
			// without crashing.
			{
				Code:     `import x from "./default_files/syntax-broken-no-default";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/syntax-broken-no-default".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				}},
			},

			// =====================================================
			// rslint-specific: tsconfig-driven semantics
			// =====================================================
			// `export = X` under no-interop tsconfig — TS forbids
			// `import X from` of a CJS module, upstream rule flags it.
			{
				Code:     `import Foo from "./default_files/typescript-export-as-default-namespace";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/typescript-export-as-default-namespace".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			// Same `export = function` file — invalid under no-interop.
			{
				Code:     `import fn from "./default_files/typescript-export-assign-function";`,
				FileName: "file.ts",
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "default",
					Message:   `No default export found in imported module "./default_files/typescript-export-assign-function".`,
					Line:      1, Column: 8, EndLine: 1, EndColumn: 10,
				}},
			},

			// =====================================================
			// rslint-specific: multiple defaults missing in one file
			// =====================================================
			// Each ImportDeclaration with a default specifier reports
			// independently; locks in that the listener doesn't bail after
			// the first hit.
			{
				Code: `import a from "./default_files/named-exports";
import b from "./default_files/re-export";`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "default",
						Message:   `No default export found in imported module "./default_files/named-exports".`,
						Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
					},
					{
						MessageId: "default",
						Message:   `No default export found in imported module "./default_files/re-export".`,
						Line:      2, Column: 8, EndLine: 2, EndColumn: 9,
					},
				},
			},
		},
	)
}
