package prefer_number_properties_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferNumberPropertiesExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / upstream issue it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestPreferNumberPropertiesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_number_properties.PreferNumberPropertiesRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: dynamic global-object element access is not a static global reference ----
			{Code: `globalThis[name];`},
			{Code: "globalThis[`parse${kind}`](value);"},

			// ---- Dimension 4: global-object writes and direct deletes are not reads ----
			{Code: `globalThis.NaN = value;`},
			{Code: `globalThis["parseFloat"] ??= value;`},
			{Code: `delete globalThis.NaN;`},
			{Code: `delete (globalThis).NaN;`},
			{Code: `delete (globalThis as any).NaN;`},
			{Code: `({foo: globalThis.NaN} = {});`},
			{Code: `((globalThis).NaN) = value;`},
			{Code: `globalThis["NaN" as const] = value;`},

			// ---- Dimension 4: local global-object aliases shadow the real global object ----
			{Code: `const globalThis = {}; globalThis.NaN;`},
			{Code: `function f(window) { return window.parseFloat(value); }`},
			{Code: `function f(globalThis) { return (globalThis as any).NaN; }`},

			// ---- Dimension 4: value-space declarations shadow globals; type-only declarations do not ----
			{Code: `namespace NaN {}; NaN;`},
			{Code: `enum NaN {}; NaN;`},
			{Code: `declare const NaN: number; NaN;`},
			{Code: `import type {NaN} from "foo"; NaN;`},

			// ---- Dimension 4: checkInfinity defaults to false for bare and namespaced reads ----
			{Code: `globalThis.Infinity;`},
			{Code: `-globalThis.Infinity;`},
			{Code: `Infinity;`, Options: map[string]interface{}{"checkInfinity": false}},

			// ---- Dimension 4: option parsing accepts array-wrapped object options ----
			{Code: `NaN;`, Options: []interface{}{map[string]interface{}{"checkNaN": false}}},
			{Code: `globalThis.NaN;`, Options: map[string]interface{}{"checkNaN": false}},

			// ---- Real-user: #2311 Infinity remains allowed by default ----
			{Code: `const isPositiveZero = value => value === 0 && 1 / value === Infinity;`},

			// N/A: declaration/container function forms do not create extra traversal boundaries; every identifier read is checked independently.
			// N/A: private identifiers are declaration/property names, not global references to `NaN` or `Infinity`.
			// N/A: spread/rest shapes are only relevant when the identifier inside them is a value read; upstream already covers destructuring write positions.
			// N/A: overload/abstract/declare body-absent forms do not contain value reads beyond declaration names.

			// Config `off` un-declares the builtin or the replacement `Number` — no report.
			{Code: `parseInt("10");`, Globals: map[string]bool{"parseInt": false}},
			{Code: `parseInt("10");`, Globals: map[string]bool{"Number": false}},
			{Code: `globalThis.parseInt("10");`, Globals: map[string]bool{"parseInt": false}},
			{Code: `globalThis.parseInt("10");`, Globals: map[string]bool{"globalThis": false}},
			{Code: `globalThis.parseInt("10");`, Globals: map[string]bool{"Number": false}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized references are transparent ----
			fixed(`(parseFloat)("10.5");`, `(Number.parseFloat)("10.5");`, "parseFloat", "parseFloat", 1, 2, 1, 12),
			fixed(`((NaN));`, `((Number.NaN));`, "NaN", "NaN", 1, 3, 1, 6),

			// ---- Dimension 4: TS assertion wrappers around references are transparent ----
			fixed(`(NaN as number);`, `(Number.NaN as number);`, "NaN", "NaN", 1, 2, 1, 5),
			fixed(`NaN!;`, `Number.NaN!;`, "NaN", "NaN", 1, 1, 1, 4),
			fixed(`(NaN satisfies number);`, `(Number.NaN satisfies number);`, "NaN", "NaN", 1, 2, 1, 5),
			fixed(`(<number>NaN);`, `(<number>Number.NaN);`, "NaN", "NaN", 1, 10, 1, 13),

			// ---- Dimension 4: type-only declarations do not shadow value-space globals ----
			fixed(`type NaN = number; NaN;`, `type NaN = number; Number.NaN;`, "NaN", "NaN", 1, 20, 1, 23),
			fixed(`interface NaN {}; NaN;`, `interface NaN {}; Number.NaN;`, "NaN", "NaN", 1, 19, 1, 22),
			fixed(`type globalThis = {}; globalThis.NaN;`, `type globalThis = {}; Number.NaN;`, "NaN", "NaN", 1, 23, 1, 37),

			// ---- Dimension 4: static element access on global objects matches a global reference ----
			fixed(`globalThis["NaN"];`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 18),
			fixed("globalThis[`NaN`];", `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 18),
			fixed(`globalThis[("NaN")];`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 20),
			fixed(`globalThis["NaN" as const];`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 27),
			fixed(`window["parseFloat"](foo);`, `Number.parseFloat(foo);`, "parseFloat", "parseFloat", 1, 1, 1, 21),
			fixed(`globalThis[("parseFloat")](value);`, `Number.parseFloat(value);`, "parseFloat", "parseFloat", 1, 1, 1, 27),
			fixedWithOptions(`globalThis["Infinity"];`, `Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 1, 1, 23),

			// ---- Dimension 4: parenthesized and TS-wrapped global objects are transparent ----
			fixed(`(globalThis).NaN;`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 17),
			fixed(`((globalThis)).parseFloat(value);`, `Number.parseFloat(value);`, "parseFloat", "parseFloat", 1, 1, 1, 26),
			fixed(`(globalThis as any).NaN;`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 24),
			fixed(`globalThis!.NaN;`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 16),
			fixed(`(globalThis satisfies typeof globalThis).NaN;`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 45),

			// ---- Dimension 4: optional-chain global-object reads still match ----
			fixed(`globalThis?.NaN;`, `Number.NaN;`, "NaN", "NaN", 1, 1, 1, 16),
			suggested(`globalThis?.isNaN(foo);`, "isNaN", "isNaN", `Number.isNaN(foo);`, 1, 1, 1, 18),
			fixed(`parseFloat?.("1");`, `Number.parseFloat?.("1");`, "parseFloat", "parseFloat", 1, 1, 1, 11),
			fixedWithOptions(`globalThis?.Infinity;`, `Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 1, 1, 21),
			fixedWithOptions(`-globalThis?.Infinity;`, `-Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 2, 1, 22),
			fixed(`delete globalThis?.NaN;`, `delete Number.NaN;`, "NaN", "NaN", 1, 8, 1, 23),
			fixed(`delete (globalThis?.["NaN"]);`, `delete (Number.NaN);`, "NaN", "NaN", 1, 9, 1, 28),
			suggested(`(globalThis as any)?.isFinite?.(value);`, "isFinite", "isFinite", `Number.isFinite?.(value);`, 1, 1, 1, 30),

			// ---- Dimension 4: computed class/object keys are value reads ----
			fixed(`const foo = {"NaN": NaN};`, `const foo = {"NaN": Number.NaN};`, "NaN", "NaN", 1, 21, 1, 24),
			fixed(`class Foo { [NaN]() {} }`, `class Foo { [Number.NaN]() {} }`, "NaN", "NaN", 1, 14, 1, 17),
			fixed(`globalThis[NaN];`, `globalThis[Number.NaN];`, "NaN", "NaN", 1, 12, 1, 15),
			fixedWithOptions(`globalThis[-Infinity];`, `globalThis[Number.NEGATIVE_INFINITY];`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 12, 1, 21),

			// ---- Dimension 4: nesting and traversal boundaries keep each reference independent ----
			fixed(`class C { static { Number.isNaN(NaN); } }`, `class C { static { Number.isNaN(Number.NaN); } }`, "NaN", "NaN", 1, 33, 1, 36),
			fixed(`const x = {...extra, value: NaN};`, `const x = {...extra, value: Number.NaN};`, "NaN", "NaN", 1, 29, 1, 32),
			fixed(`const { value = NaN, ...rest } = obj;`, `const { value = Number.NaN, ...rest } = obj;`, "NaN", "NaN", 1, 17, 1, 20),
			fixed(`function f(parseInt) { return function g() { return parseInt(value) + globalThis.parseInt(value); }; }`, `function f(parseInt) { return function g() { return parseInt(value) + Number.parseInt(value); }; }`, "parseInt", "parseInt", 1, 71, 1, 90),

			// ---- Real-user: #1439 global object references are detected ----
			fixed(`globalThis.parseFloat(cssLength);`, `Number.parseFloat(cssLength);`, "parseFloat", "parseFloat", 1, 1, 1, 22),
			fixedWithOptions(`globalThis.Infinity;`, `Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 1, 1, 20),

			// ---- Real-user: #2192 v64 still offers a suggestion for known-number arguments ----
			suggested(`isNaN(foo - 1);`, "isNaN", "isNaN", `Number.isNaN(foo - 1);`, 1, 1, 1, 6),

			// ---- Real-user: #863 checkInfinity enables negative Infinity fixes ----
			fixedWithOptions(`delete -Infinity;`, `delete Number.NEGATIVE_INFINITY;`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 8, 1, 17),

			// Locks in upstream checkProperty() arm 1: shorthand property values expand to key/value pairs.
			fixed(`const foo = {NaN};`, `const foo = {NaN: Number.NaN};`, "NaN", "NaN", 1, 14, 1, 17),

			// Locks in upstream checkProperty() arm 2: unsafe globals produce suggestions, not autofixes.
			suggested(`isFinite(value);`, "isFinite", "isFinite", `Number.isFinite(value);`, 1, 1, 1, 9),

			// Locks in upstream checkProperty() arm 3: positive Infinity maps to POSITIVE_INFINITY.
			fixedWithOptions(`globalThis.Infinity;`, `Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", []interface{}{map[string]interface{}{"checkInfinity": true}}, 1, 1, 1, 20),
			fixedWithOptions(`NaN; Infinity;`, `NaN; Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true, "checkNaN": false}, 1, 6, 1, 14),

			// Locks in upstream checkProperty() arm 4: negative Infinity reports on the unary expression.
			fixedWithOptions(`-globalThis.Infinity`, `Number.NEGATIVE_INFINITY`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 1, 1, 21),
			fixedWithOptions(`(-globalThis.Infinity).toString();`, `(Number.NEGATIVE_INFINITY).toString();`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 2, 1, 22),
			fixedWithOptions(`function f(){throw-Infinity}`, `function f(){throw Number.NEGATIVE_INFINITY}`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 19, 1, 28),

			// Locks in upstream defaultOptions merge: `{}` keeps checkNaN enabled.
			fixedWithOptions(`NaN;`, `Number.NaN;`, "NaN", "NaN", map[string]interface{}{}, 1, 1, 1, 4),

			// ---- Real-user: config objects often carry parser helpers and sentinel values together ----
			{
				Code:   `const config = {parse: globalThis.parseFloat, missing: (globalThis as any)["NaN" as const]};`,
				Output: []string{`const config = {parse: Number.parseFloat, missing: Number.NaN};`},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("parseFloat", "parseFloat", 1, 24, 1, 45),
					expected("NaN", "NaN", 1, 56, 1, 91),
				},
			},
		},
	)
}
