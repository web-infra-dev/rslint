package es_builtins_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_unsupported_features/es_builtins"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Expected diagnostics were checked with eslint-plugin-n v18.3.0, ESLint 9.39.5
// and @typescript-eslint/parser 8.70.0. Extra cases stay in Go.
func TestESBuiltinsExtras(t *testing.T) {
	runESBuiltinsTests(t, []rule_tester.ValidTestCase{
		// The existing npm range utility ignores contradictory alternatives.
		// Upstream incorrectly reports Promise and Promise.any for the first
		// ordering. Keep the shared semantics documented for this rule.
		{Code: "Promise.any(xs)", Options: map[string]any{"version": ">=16 || >20 <16"}},
		{Code: "Promise.any(xs)", Options: map[string]any{"version": ">20 <16 || >=16"}},
		// An explicit version takes precedence over settings.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": ">=18"}, Settings: map[string]any{"n": map[string]any{"version": "^14"}}},
		// A fully supported union is accepted.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "^15 || >=16"}},
		// Multiple ignored APIs are independent.
		{Code: "Promise.any(items); Map.groupBy(items, key);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14", "ignores": []any{"Promise.any", "Map.groupBy"}}},
		// Computed variables and instance methods are outside the rule.
		{Code: "const name = \"any\"; Promise[name](xs); xs.at(0); Array.prototype.at; unknown.Promise.any;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Assignment to a global suppresses tracking for that global.
		{Code: "Promise = replacement; Promise.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Local bindings shadow builtin globals.
		{Code: "import Promise from \"promise\"; Promise.any(xs); function f(WeakRef) { return new WeakRef(x); }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Destructured binding shadows the global.
		{Code: "const {Promise} = local; Promise.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Disabled globals are not tracked.
		{Code: "Promise.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}, Globals: map[string]any{"Promise": "off"}},
		// A local global object shadows configured globals.
		{Code: "function f(globalThis) { globalThis.Promise.any(xs); }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Private keys do not match public API properties.
		{Code: "class C { #any; method() { return Promise.#any; } }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Ambient value declarations shadow global values.
		{Code: "declare const Promise: any; Promise.any(xs);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// JSDoc types do not read globals.
		{Code: "/** @type {WeakRef<object>} */ let value;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// Module references are not ES globals.
		{Code: "const {Promise} = require(\"other\"); Promise.any(xs); import {WeakRef} from \"other\"; new WeakRef(x);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}},
		// Returning an API from a function does not create a static alias.
		{Code: "function getPromise() { return Promise; } getPromise().any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// Modified global object is not traced.
		{Code: "globalThis = other; globalThis.Promise.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// Script var declarations shadow globals.
		{Code: "var Promise; Promise.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Options: map[string]any{"version": "14.0.0"}},
		// Type-only imports shadow global values as in upstream.
		{Code: "import type {Promise, WeakRef} from \"other\"; Promise.any(tasks); new WeakRef(object);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// Namespace value declarations shadow builtins.
		{Code: "namespace Promise { export const any = local; } Promise.any(tasks);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// JSDoc type queries remain comments.
		{Code: "/** @type {typeof WeakRef} */ let value;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// Catch bindings shadow global API roots.
		{Code: "try {} catch (Promise) { Promise.any(tasks); }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
		// Computed private-like strings are ordinary keys.
		{Code: "Promise[\"#any\"](tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}},
	}, []rule_tester.InvalidTestCase{
		// Omitted options use the Node 16 fallback.
		{Code: "Promise.any(items); Error.cause; Intl.supportedValuesOf(\"calendar\"); Map.groupBy(items, key);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Error.cause", "16.9.0", ">=16.0.0", 1, 21, 1, 32),
				unsupportedError("Intl.supportedValuesOf", "18.0.0", ">=16.0.0", 1, 34, 1, 56),
				unsupportedError("Map.groupBy", "21.0.0", ">=16.0.0", 1, 70, 1, 81),
			},
		},
		// Empty options keep the runtime default.
		{Code: "Error.cause;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Error.cause", "16.9.0", ">=16.0.0", 1, 1, 1, 12),
			},
		},
		// Explicit defaults match omitted options.
		{Code: "Error.cause;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": ">=16.0.0", "ignores": []any{}},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Error.cause", "16.9.0", ">=16.0.0", 1, 1, 1, 12),
			},
		},
		// Invalid range falls through to settings.n before settings.node.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "invalid"}, Settings: map[string]any{"n": map[string]any{"version": "^14"}, "node": map[string]any{"version": ">=18"}},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "^14", 1, 1, 1, 12),
			},
		},
		// Invalid and empty settings fall back.
		{Code: "Object.hasOwn(object, key);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": ""}, Settings: map[string]any{"n": map[string]any{"version": "invalid"}, "node": map[string]any{"version": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Object.hasOwn", "16.9.0", ">=16.0.0", 1, 1, 1, 14),
			},
		},
		// Both union alternatives must support the feature.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "^14 || >=16"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "^14 || >=16", 1, 1, 1, 12),
			},
		},
		// Prereleases are excluded from stable support ranges.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "15.0.0-rc.1"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise", "0.12.0", "15.0.0-rc.1", 1, 1, 1, 8),
				unsupportedError("Promise.any", "15.0.0", "15.0.0-rc.1", 1, 1, 1, 12),
			},
		},
		// Canonical null ranges retain npm subset behavior.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "<0.0.0-0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise", "0.12.0", "<0.0.0-0", 1, 1, 1, 8),
				unsupportedError("Promise.any", "15.0.0", "<0.0.0-0", 1, 1, 1, 12),
			},
		},
		// Range diagnostics use npm whitespace normalization.
		{Code: "Promise.any(items);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "  >= 14.0.0\t<15  "},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", ">= 14.0.0 <15", 1, 1, 1, 12),
			},
		},
		// Ignoring a root does not ignore its properties.
		{Code: "Atomics.add(array, 0, 1);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "8.0.0", "ignores": []any{"Atomics"}},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Atomics.add", "8.10.0", "8.0.0", 1, 1, 1, 12),
			},
		},
		// Ignoring a property does not ignore its root.
		{Code: "Atomics.add(array, 0, 1);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "8.0.0", "ignores": []any{"Atomics.add"}},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Atomics", "8.10.0", "8.0.0", 1, 1, 1, 8),
			},
		},
		// Version tables cover later static builtins.
		{Code: "Object.hasOwn(o, k); Map.groupBy(xs, f); Object.groupBy(xs, f); Intl.DisplayNames; Intl.Segmenter; Intl.Segments; Intl.supportedValuesOf(\"calendar\"); RegExp.hasIndices; Error.cause; Atomics.waitAsync;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "12.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Object.hasOwn", "16.9.0", "12.0.0", 1, 1, 1, 14),
				unsupportedError("Map.groupBy", "21.0.0", "12.0.0", 1, 22, 1, 33),
				unsupportedError("Object.groupBy", "21.0.0", "12.0.0", 1, 42, 1, 56),
				unsupportedError("Intl.DisplayNames", "14.0.0", "12.0.0", 1, 65, 1, 82),
				unsupportedError("Intl.Segmenter", "16.0.0", "12.0.0", 1, 84, 1, 98),
				unsupportedError("Intl.Segments", "16.0.0", "12.0.0", 1, 100, 1, 113),
				unsupportedError("Intl.supportedValuesOf", "18.0.0", "12.0.0", 1, 115, 1, 137),
				unsupportedError("RegExp.hasIndices", "16.0.0", "12.0.0", 1, 151, 1, 168),
				unsupportedError("Error.cause", "16.9.0", "12.0.0", 1, 170, 1, 181),
				unsupportedError("Atomics.waitAsync", "16.0.0", "12.0.0", 1, 183, 1, 200),
			},
		},
		// One expression can report the root and its member.
		{Code: "Reflect.ownKeys(o);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "5.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Reflect", "6.0.0", "5.0.0", 1, 1, 1, 8),
				unsupportedError("Reflect.ownKeys", "6.0.0", "5.0.0", 1, 1, 1, 16),
			},
		},
		// Computed static keys include strings and constant expressions.
		{Code: "Promise[\"any\"](xs); Promise[`a${\"ny\"}`](xs); Promise[\"an\" + \"y\"](xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 1, 1, 15),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 21, 1, 40),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 46, 1, 65),
			},
		},
		// Parentheses and optional accesses retain member ranges.
		{Code: "((Promise))?.[\"any\"]?.(xs); (Promise.any)(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 1, 1, 21),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 30, 1, 41),
			},
		},
		// Constructor reads report the callee instead of the new expression.
		{Code: "new (WeakRef)(object);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 6, 1, 13),
			},
		},
		// Aliases and defaults use canonical API names.
		{Code: "const P = Promise; P.any(xs); const {any: race = fallback} = P; race(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 20, 1, 25),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 38, 1, 58),
			},
		},
		// Assignment destructuring uses the complete property range.
		{Code: "let race; ({any: race = fallback} = Promise);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 13, 1, 33),
			},
		},
		// Global object references and aliases share the builtin trace.
		{Code: "globalThis.Promise.any(xs); global.Promise.any(xs); self.Promise.any(xs); window.Promise.any(xs); const g=globalThis; const {Promise:P}=g; P.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"}, Globals: map[string]any{"global": "readonly", "self": "readonly", "window": "readonly"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 1, 1, 23),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 29, 1, 47),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 53, 1, 69),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 75, 1, 93),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 140, 1, 145),
			},
		},
		// An unsupported globalThis read and its API both report.
		{Code: "globalThis.Promise.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "11"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("globalThis", "12.0.0", "11", 1, 1, 1, 11),
				unsupportedError("Promise.any", "15.0.0", "11", 1, 1, 1, 23),
			},
		},
		// Nested unsupported globals and members report in source range order.
		{Code: "globalThis.Map.groupBy(xs, key);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "0.1.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("globalThis", "12.0.0", "0.1.0", 1, 1, 1, 11),
				unsupportedError("Map", "0.12.0", "0.1.0", 1, 1, 1, 15),
				unsupportedError("Map.groupBy", "21.0.0", "0.1.0", 1, 1, 1, 23),
			},
		},
		// Two independent alias paths preserve duplicate reports.
		{Code: "const P = condition ? Promise : Promise; P.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 42, 1, 47),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 42, 1, 47),
			},
		},
		// A reassigned alias is followed like upstream.
		{Code: "let P = Promise; P = other; P.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 29, 1, 34),
			},
		},
		// Property writes and deletes are READ events.
		{Code: "Promise.any = polyfill; delete Promise.any;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 1, 1, 12),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 32, 1, 43),
			},
		},
		// Global directives enable global object tracking.
		{Code: "/* global window */ window.Promise.any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 21, 1, 39),
			},
		},
		// Class static blocks and computed keys evaluate runtime references.
		{Code: "class C { static { Promise.any(xs); } [Promise.any]() {} }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 20, 1, 31),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 40, 1, 51),
			},
		},
		// JSX member tags differ from expression containers.
		{Code: "const view = <Promise.any value={Promise.any(xs)} />;", FileName: "input.tsx", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 34, 1, 45),
			},
		},
		// TypeScript type-only references retain upstream scope behavior.
		{Code: "type A = typeof Promise.any; interface B extends WeakRef<object> {}", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14", 1, 50, 1, 57),
			},
		},
		// TypeScript expression wrappers match upstream reference boundaries.
		{Code: "(Promise as any).any(xs); Promise!.any(xs); (Promise satisfies unknown).any(xs);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 1, 1, 21),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 27, 1, 39),
				unsupportedError("Promise.any", "15.0.0", "14", 1, 45, 1, 76),
			},
		},
		// Type-only declarations do not shadow runtime values.
		{Code: "interface Promise {} type WeakRef = unknown; Promise.any(xs); new WeakRef(x);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 46, 1, 57),
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 67, 1, 74),
			},
		},
		// Class heritage reads runtime values.
		{Code: "class C extends WeakRef {}", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 17, 1, 24),
			},
		},
		// Decorators contain runtime expressions.
		{Code: "@Promise.any class C {}", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 2, 1, 13),
			},
		},
		// JSDoc casts preserve JavaScript expression tracking.
		{Code: "(/** @type {any} */ (Promise)).any(xs);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 1, 1, 35),
			},
		},
		// Unicode and multiline diagnostics use UTF-16 columns.
		{Code: "/* 🐱 */ Promise\n  .any(xs);\nconst {\n  any: café = fallback\n} = Promise;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14", 1, 10, 2, 7),
				unsupportedError("Promise.any", "15.0.0", "14", 4, 3, 4, 23),
			},
		},
		// Escaped identifiers and property names.
		{Code: "Prom\\u0069se.\\u0061ny(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 1, 1, 22),
			},
		},
		// Static computed destructuring keys.
		{Code: "const {[\"any\"]: race, [\"an\" + \"y\"]: other} = Promise;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 8, 1, 21),
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 23, 1, 42),
			},
		},
		// Rest bindings do not alias the source API.
		{Code: "const {any: race, ...rest} = Promise; rest.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 8, 1, 17),
			},
		},
		// Logical assignments keep the source API reference.
		{Code: "let P; P ||= Promise; P.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 23, 1, 28),
			},
		},
		// Compound assignments retain upstream alias tracking.
		{Code: "let P = 1; P += Promise; P.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 26, 1, 31),
			},
		},
		// Comma and coalescing expressions preserve the returned value.
		{Code: "const P = (0, Promise) ?? other; P.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 34, 1, 39),
			},
		},
		// An alias cycle terminates without losing the reference.
		{Code: "let P = Promise; let Q = P; P = Q; Q.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 36, 1, 41),
			},
		},
		// Disabled global does not disable a global object property.
		{Code: "Promise.any(tasks); globalThis.Promise.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"}, Globals: map[string]any{"Promise": "off"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 21, 1, 43),
			},
		},
		// With scopes follow upstream reference resolution.
		{Code: "with (context) { Promise.any(tasks); }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 18, 1, 29),
			},
		},
		// A runtime availability guard still reads the builtin.
		{Code: "if (typeof WeakRef !== \"undefined\") { new WeakRef(object); }", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 12, 1, 19),
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 43, 1, 50),
			},
		},
		// CRLF and non-BMP text preserve diagnostic columns.
		{Code: "// first\r\n/* 🐱 */ Promise.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 2, 10, 2, 21),
			},
		},
		// BOM and hashbang precede the member range.
		{Code: "\ufeff#!/usr/bin/env node\nPromise.any(tasks);", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 2, 1, 2, 12),
			},
		},
		// Type queries retain direct global reference reads.
		{Code: "type T = typeof WeakRef;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 17, 1, 24),
			},
		},
		// Instantiation expressions retain alias values.
		{Code: "const P = Promise<number>; P.any(tasks);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 28, 1, 33),
			},
		},
		// JSX reads a builtin constructor but not its member name.
		{Code: "const view = <WeakRef foo={Promise.any(tasks)} />;", FileName: "input.tsx", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 15, 1, 22),
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 28, 1, 39),
			},
		},
		// Declared type parameters do not shadow values.
		{Code: "function f<Promise>() { return Promise.any(tasks); }", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 32, 1, 43),
			},
		},
		// Default export identifiers read global values.
		{Code: "export default WeakRef;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("WeakRef", "14.6.0", "14.0.0", 1, 16, 1, 23),
			},
		},
		// Decorated auto accessors read runtime values.
		{Code: "class C { @Promise.any accessor value; }", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 12, 1, 23),
			},
		},
		// Tagged templates read the static method.
		{Code: "Promise.any`text`;", FileName: "extras.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: map[string]any{"version": "14.0.0"},
			Errors: []rule_tester.InvalidTestCaseError{
				unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 1, 1, 12),
			},
		},
	})
}

func TestESBuiltinsPackageVersions(t *testing.T) {
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "es-builtins-versions")
	archive := txtarfs.MustParseFile(t, "testdata/versions.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("empty version fixtures")
	}
	files := map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"target":"esnext"},"include":["**/*.ts"]}`,
	}
	for _, name := range names {
		content, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(content)
	}
	root := rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &es_builtins.ESBuiltinsRule,
		[]rule_tester.ValidTestCase{
			{Code: "Promise.any(xs)", FileName: "engines/nested/input.ts"},
			{Code: "Promise.any(xs)", FileName: "engines/input.ts", Options: map[string]any{"version": ">=18"}},
			{Code: "Promise.any(xs)", FileName: "engines/input.ts", Settings: map[string]any{"node": map[string]any{"version": ">=18"}}},
		}, []rule_tester.InvalidTestCase{
			{Code: "Promise.any(xs)", FileName: "engines/input.ts", Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.any", "15.0.0", "^14", 1, 1, 1, 12)}},
			{Code: "Promise.any(xs)", FileName: "dev-engines/input.ts", Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.any", "15.0.0", "^14", 1, 1, 1, 12)}},
			{Code: "Promise.any(xs)", FileName: "dev-engines-array/input.ts", Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.any", "15.0.0", "^14", 1, 1, 1, 12)}},
			{Code: "Object.hasOwn(o, k)", FileName: "default/input.ts", Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.hasOwn", "16.9.0", ">=16.0.0", 1, 1, 1, 14)}},
		})
}

func TestESBuiltinsSchema(t *testing.T) {
	for _, options := range [][]any{
		nil, {map[string]any{}},
		{map[string]any{"version": "invalid", "ignores": []any{"Map.groupBy", "Atomics.add"}}},
	} {
		if err := es_builtins.ESBuiltinsRule.Schema.Validate(options); err != nil {
			t.Errorf("valid options %v: %v", options, err)
		}
	}
	for _, options := range [][]any{
		{map[string]any{"version": 18}},
		{map[string]any{"ignores": []any{"unknown"}}},
		{map[string]any{"ignores": []any{"Object"}}},
		{map[string]any{"ignores": []any{"Promise.any", "Promise.any"}}},
		{map[string]any{"ignores": "Promise.any"}},
		{map[string]any{"allowExperimental": true}},
		{map[string]any{}, map[string]any{}},
	} {
		if err := es_builtins.ESBuiltinsRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options %v", options)
		}
	}
}
