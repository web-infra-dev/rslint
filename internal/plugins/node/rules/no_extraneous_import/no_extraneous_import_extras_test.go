package no_extraneous_import

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations checked against eslint-plugin-n v18.3.0 with ESLint 10.2.1.
func TestNoExtraneousImportExtras(t *testing.T) {
	root := extraneousRoot(t, "testdata/extras.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoExtraneousImportRule, []rule_tester.ValidTestCase{
		// exports array fallback
		{Code: "import 'array-export';", FileName: "input.js"},

		// broken condition does not fall through
		{Code: "import 'broken-condition';", FileName: "input.js"},

		// Documented difference: enhanced-resolve's undocumented alias override
		// resolves this local file and reports "virtual"; only modules is supported here.
		{Code: "import 'virtual';", FileName: "input.js", Options: map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"virtual": "./local.js"}}}},
		// non npm sources
		{Code: "import './local'; import '/absolute'; import 'node:fs'; import 'fs'; import 'data:text/javascript,0'; import 'https://example.com/a.js'; import '#internal'; import 'virtual:thing';", FileName: "input.js"},
		// declared and self
		{Code: "import 'declared'; export * from 'optional/sub'; import 'app/sub';", FileName: "input.js"},
		// missing runtime
		{Code: "import 'missing'; import 'only'; import 'only-ts';", FileName: "input.js"},
		// ignored import shapes
		{Code: "require('runtime'); const obj = {import() {}}; obj.import('runtime'); import(`runtime`); import('run' + 'time');", FileName: "input.js"},
		// TS import types and equals ignored
		{Code: "type Value = import('runtime').Value; import value = require('runtime');", FileName: "input.ts"},
		// JSX attributes ignored
		{Code: "const view = <Panel source='runtime'/>;", FileName: "input.tsx"},
		// type declarations allowed
		{Code: "import type { Value } from 'typed'; export type * from 'typed'; import type { Value as Other } from '@scope/typed/sub';", FileName: "input.ts"},
		// exports non runtime
		{Code: "import 'types-export'; import 'exports/private.js'; import 'exports/blocked';", FileName: "input.js"},
		// allow package and scoped subdirectories
		{Code: "import 'runtime/part'; import '@scope/runtime';", FileName: "input.js", Options: []any{map[string]any{"allowModules": []any{"runtime", "@scope/runtime"}}}},
		// shared allow fallback
		{Code: "import 'runtime';", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{"runtime"}}}},
		// legacy shared namespace
		{Code: "import 'runtime';", FileName: "input.js", Settings: map[string]any{"n": map[string]any{"allowModules": []any{"runtime"}}, "node": map[string]any{"allowModules": []any{}}}},
		// invalid settings fall through
		{Code: "import 'runtime';", FileName: "input.js", Settings: map[string]any{"n": map[string]any{"allowModules": false}, "node": map[string]any{"allowModules": []any{"runtime"}}}},
		// unconfigured lookup
		{Code: "import 'external'; import 'bower';", FileName: "input.js"},
		// empty extra lookup overrides
		{Code: "import 'external';", FileName: "input.js", Options: []any{map[string]any{"resolvePaths": []any{}}}, Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"extra"}}}},
		// custom modules overrides
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{"bower_components"}}}}},
		// empty modules
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}},
		// empty extensions
		{Code: "import 'runtime';", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{}}}},
		// TS implicit extensions
		{Code: "import 'only-ts';", FileName: "input.ts"},
		// TS alias exemption
		{Code: "import 'alias/part'; import 'alias-other';", FileName: "aliases/input.ts"},
		// malformed package stops resolution
		{Code: "import 'runtime';", FileName: "malformed/input.js"},
	}, []rule_tester.InvalidTestCase{
		// An unmatched nested condition may fall through before target selection.
		{Code: "import 'nested-conditions'; import 'array-conditions';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "extraneous", Message: `"nested-conditions" is extraneous.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 27},
			{MessageId: "extraneous", Message: `"array-conditions" is extraneous.`, Line: 1, Column: 36, EndLine: 1, EndColumn: 54},
		}},
		// typesVersions is not runtime metadata
		{Code: "import 'versioned';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"versioned\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},

		// explicit tsconfig mapping
		{Code: "import 'tsx-only/index.js';", FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"tsconfigPath": tspath.ResolvePath(root.Dir, "jsx-options.json")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"tsx-only\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},

		// TS preserve preset
		{Code: "import 'tsx-only/index.jsx';", FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "preserve"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"tsx-only\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},

		// TS react preset
		{Code: "import 'tsx-only/index.js';", FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"tsx-only\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},

		// custom TS extension mapping
		{Code: "import 'custom/index.js';", FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{".custom", ".js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"custom\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},

		// all reporting shapes
		{Code: "import 'runtime'; import value from '@scope/runtime/part'; export * from 'runtime'; export { value } from 'runtime'; export * as values from 'runtime'; import('runtime');", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}, {MessageId: "extraneous", Message: "\"@scope/runtime\" is extraneous.", Line: 1, Column: 37, EndLine: 1, EndColumn: 58}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 74, EndLine: 1, EndColumn: 83}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 107, EndLine: 1, EndColumn: 116}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 142, EndLine: 1, EndColumn: 151}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 160, EndLine: 1, EndColumn: 169}}},
		// unicode multiline positions
		{Code: "const text = '😀';\nimport(/*位置*/\n ('runt\\u0069me')\n);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 3, Column: 3, EndLine: 3, EndColumn: 17}}},
		// attributes
		{Code: "import value from 'runtime' with { type: 'json' }; import('runtime', {with:{type:'json'}});", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 19, EndLine: 1, EndColumn: 28}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 59, EndLine: 1, EndColumn: 68}}},
		// inline types stay value imports
		{Code: "import { type Value } from 'typed'; export { type Value } from 'typed';", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"typed\" is extraneous.", Line: 1, Column: 28, EndLine: 1, EndColumn: 35}, {MessageId: "extraneous", Message: "\"typed\" is extraneous.", Line: 1, Column: 64, EndLine: 1, EndColumn: 71}}},
		// undeclared type imports
		{Code: "import type { Value } from 'runtime'; export type { Value } from 'runtime'; export type * from 'runtime';", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 66, EndLine: 1, EndColumn: 75}, {MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 96, EndLine: 1, EndColumn: 105}}},
		// TS parenthesized dynamic
		{Code: "import((('runtime'))); import('runtime' as string);", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19}}},
		// literal dynamic coercions
		{Code: "import(42); import(false); import(null); import(42n);", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 10}, {MessageId: "extraneous", Message: "\"false\" is extraneous.", Line: 1, Column: 20, EndLine: 1, EndColumn: 25}, {MessageId: "extraneous", Message: "\"null\" is extraneous.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39}, {MessageId: "extraneous", Message: "\"42\" is extraneous.", Line: 1, Column: 49, EndLine: 1, EndColumn: 52}}},
		// explicit defaults
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// default runtime extensions
		{Code: "import 'only-mjs'; import 'only-cjs'; import 'only-json'; import 'only-node'; import 'extensionless';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"only-mjs\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}, {MessageId: "extraneous", Message: "\"only-cjs\" is extraneous.", Line: 1, Column: 27, EndLine: 1, EndColumn: 37}, {MessageId: "extraneous", Message: "\"only-json\" is extraneous.", Line: 1, Column: 46, EndLine: 1, EndColumn: 57}, {MessageId: "extraneous", Message: "\"only-node\" is extraneous.", Line: 1, Column: 66, EndLine: 1, EndColumn: 77}, {MessageId: "extraneous", Message: "\"extensionless\" is extraneous.", Line: 1, Column: 86, EndLine: 1, EndColumn: 101}}},
		// main retains explicit extension
		{Code: "import 'entry';", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"entry\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
		// main directory
		{Code: "import 'directory-main';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"directory-main\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		// declaration main and conditions
		{Code: "import type {Value} from 'decl-only'; import type {Value as T} from 'types-export'; import 'conditional';", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"decl-only\" is extraneous.", Line: 1, Column: 26, EndLine: 1, EndColumn: 37}, {MessageId: "extraneous", Message: "\"types-export\" is extraneous.", Line: 1, Column: 69, EndLine: 1, EndColumn: 83}, {MessageId: "extraneous", Message: "\"conditional\" is extraneous.", Line: 1, Column: 92, EndLine: 1, EndColumn: 105}}},
		// exports and wildcard
		{Code: "import 'exports'; import 'exports/parts/part';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}, {MessageId: "extraneous", Message: "\"exports\" is extraneous.", Line: 1, Column: 26, EndLine: 1, EndColumn: 46}}},
		// loader params
		{Code: "import 'runtime!loader';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		// module query preserved in message
		{Code: "import 'runtime?raw';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime?raw\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
		// options override shared
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"allowModules": []any{}}}, Settings: map[string]any{"node": map[string]any{"allowModules": []any{"runtime"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// allow is exact not prefix
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"allowModules": []any{"run"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// extra lookup option
		{Code: "import 'external';", FileName: "input.js", Options: []any{map[string]any{"resolvePaths": []any{"extra"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"external\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
		// extra lookup setting
		{Code: "import 'external';", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"extra"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"external\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
		// custom modules option
		{Code: "import 'bower';", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{"node_modules", "bower_components"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bower\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
		// custom modules shared
		{Code: "import 'bower';", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": []any{"bower_components"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"bower\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
		// empty resolver overrides settings
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{}}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": []any{}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// custom extensions
		{Code: "import 'custom';", FileName: "input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".custom"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"custom\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		// TS explicit extensions
		{Code: "import 'only-ts';", FileName: "explicit/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"only-ts\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// TS extension alias
		{Code: "import 'only-ts/index.js';", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"only-ts\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
		// JS does not exempt TS alias
		{Code: "import 'alias/part';", FileName: "aliases/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"alias\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		// invalid nested package falls back
		{Code: "import 'runtime';", FileName: "nested/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// invalid dependency fields
		{Code: "import 'runtime';", FileName: "invalid-fields/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// convertPath object inert
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"convertPath": map[string]any{"**": []any{"^.*$", "missing"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// convertPath array inert
		{Code: "import 'runtime';", FileName: "input.js", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"**"}, "exclude": []any{"skip/**"}, "replace": []any{"^.*$", "missing"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"runtime\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
	})
}

// Runtime resolution boundaries checked against the pinned upstream resolver.
func TestNoExtraneousImportResolutionExtras(t *testing.T) {
	root := extraneousRoot(t, "testdata/resolution.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoExtraneousImportRule, []rule_tester.ValidTestCase{
		// Preserve the final export-path component for tsgo's validation.
		{Code: "import 'exports-last-node-modules';", FileName: "input.js"},
		{Code: "import 'exports-last-dot';", FileName: "input.js"},
		{Code: "import 'exports-last-parent';", FileName: "input.js"},
		{Code: "import 'exports-array-last-node-modules';", FileName: "input.js"},
		{Code: "import 'exports-array-last-dot';", FileName: "input.js"},
		{Code: "import 'exports-array-last-parent';", FileName: "input.js"},
		// A trailing directory separator must not enable a later target.
		{Code: "import 'exports-array-directory-slash';", FileName: "input.js"},
		{Code: "import 'exports-condition-directory-slash';", FileName: "input.js"},
		// exports-no-condition/js
		{Code: "import 'exports-no-condition';", FileName: "input.js"},
		// exports-null-default/js
		{Code: "import 'exports-null-default';", FileName: "input.js"},
		// main-ts/ts
		{Code: "import 'main-ts';", FileName: "input.ts"},
		// exports-ts/ts
		{Code: "import 'exports-ts';", FileName: "input.ts"},
		// main-extensionless/ts
		{Code: "import 'main-extensionless';", FileName: "input.ts"},
		// exports-directory-slash/js
		{Code: "import 'exports-directory-slash';", FileName: "input.js"},
		// exports-null/js
		{Code: "import 'exports-null';", FileName: "input.js"},
		// exports-number/js
		{Code: "import 'exports-number';", FileName: "input.js"},
		// exports-pattern/ts
		{Code: "import 'exports-pattern/file';", FileName: "input.ts"},
		// exports-parent/js
		{Code: "import 'exports-parent';", FileName: "input.js"},
		// exports-main/js
		{Code: "import 'exports-main';", FileName: "input.js"},
		// subpath-js/ts
		{Code: "import 'subpath-js/index.js';", FileName: "input.ts"},
		// exports-late-null/js
		{Code: "import 'exports-late-null';", FileName: "input.js"},
		// exports-nested-array/js
		{Code: "import 'exports-nested-array';", FileName: "input.js"},
	}, []rule_tester.InvalidTestCase{
		// main-js/ts
		{Code: "import 'main-js';", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"main-js\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		// exports-js/ts
		{Code: "import 'exports-js';", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-js\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		// exports-extensionless/js
		{Code: "import 'exports-extensionless';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-extensionless\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 31}}},
		// exports-directory/js
		{Code: "import 'exports-directory';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-directory\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
		// exports-invalid/js
		{Code: "import 'exports-invalid';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-invalid\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
		// exports-bad-condition/js
		{Code: "import 'exports-bad-condition';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-bad-condition\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 31}}},
		// main-parent/js
		{Code: "import 'main-parent';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"main-parent\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
		// main-double-extension/js
		{Code: "import 'main-double-extension';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"main-double-extension\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 31}}},
		// dot.name/js
		{Code: "import 'dot.name';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"dot.name\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
		// ext-appended/js
		{Code: "import 'ext-appended/sub.js';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"ext-appended\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},
		// invalid-map-tsconfig
		{Code: "import 'tsx/index.js';", FileName: "input.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "bad-preset", "tsconfigPath": tspath.ResolvePath(root.Dir, "react.json")}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"tsx\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
		// exports-null-condition/js
		{Code: "import 'exports-null-condition';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-null-condition\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 32}}},
		// exports-parent-array/js
		{Code: "import 'exports-parent-array';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-parent-array\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 30}}},
		// exports-directory-main/js
		{Code: "import 'exports-directory-main';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-directory-main\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 32}}},
		// exports-directory-cycle/js
		{Code: "import 'exports-directory-cycle';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-directory-cycle\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 33}}},
		// exports-directory-broken/js
		{Code: "import 'exports-directory-broken';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-directory-broken\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 34}}},
		// exports-query/js
		{Code: "import 'exports-query';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-query\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// main-query/js
		{Code: "import 'main-query';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"main-query\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		// exports-false/js
		{Code: "import 'exports-false';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-false\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// exports-empty/js
		{Code: "import 'exports-empty';", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"exports-empty\" is extraneous.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// mixed-type-and-runtime
		{Code: "import type {Value} from 'main-js'; import 'main-js'; export type * from 'main-js'; import('main-js');", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "extraneous", Message: "\"main-js\" is extraneous.", Line: 1, Column: 44, EndLine: 1, EndColumn: 53}, {MessageId: "extraneous", Message: "\"main-js\" is extraneous.", Line: 1, Column: 92, EndLine: 1, EndColumn: 101}}},
	})
}
