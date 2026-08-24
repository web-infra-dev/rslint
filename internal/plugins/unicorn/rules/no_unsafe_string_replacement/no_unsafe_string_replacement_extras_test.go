// TestNoUnsafeStringReplacementExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / real-user scenario it
// covers, so future refactors cannot silently regress it.
package no_unsafe_string_replacement_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unsafe_string_replacement"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeStringReplacementExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unsafe_string_replacement.NoUnsafeStringReplacementRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: replacement expression wrappers are transparent ----
			tsValid(`template.replace("x", (("safe")))`),
			tsValid(`template.replace("x", ("safe" as string)!)`),
			tsValid(`template.replace("x", "safe" satisfies string)`),
			tsValid(`template.replace("x", async () => value)`),
			tsValid(`template.replace("x", async function * () { yield value; })`),

			// ---- Dimension 4: only dot-property replace calls match ----
			jsValid(`template["replaceAll"]("x", replacement)`),
			jsValid("template[`replace`](\"x\", replacement)"),
			jsValid(`template[0]("x", replacement)`),
			jsValid(`template[Symbol.replace]("x", replacement)`),

			// ---- Dimension 4: plain object key equivalence classes ----
			jsValid(`template.replace("x", {safe: value})`),
			jsValid(`template.replace("x", {"safe": value})`),
			jsValid(`template.replace("x", {0: value})`),
			jsValid(`template.replace("x", {1n: value})`),
			jsValid(`template.replace("x", {true: value})`),
			jsValid(`template.replace("x", {false: value})`),
			jsValid(`template.replace("x", {null: value})`),
			jsValid(`template.replace("x", {safe})`),
			jsValid(`template.replace("x", {})`),

			// ---- Dimension 4: private and wrapped non-member callees do not match ----
			jsValid(`class Box { #replace(value, replacement) {} run(replacement) { this.#replace("x", replacement) } }`),
			jsValid(`(template?.replace)("x", replacement)`),
			tsValid(`(template.replace as unknown)("x", replacement)`),

			// ---- Dimension 4: empty and spread argument shapes do not crash ----
			jsValid(`template.replace()`),
			jsValid(`template.replace("x", ...[])`),
			jsValid(`template.replace(...[], replacement)`),

			// ---- Dimension 4: body-absent declarations are unrelated containers ----
			tsValid(`declare function replace(value: string, replacement: unknown): string;`),
			tsValid(`declare abstract class Box { abstract replace(value: string, replacement: unknown): string }`),

			// ---- Real-user: #3437 Next.js router.replace object options ----
			jsValid(`const router = useRouter(); router.replace("/about", {locale: "en"});`),

			// ---- Real-user: #3437 typed non-string router receiver ----
			tsValid(`interface Router { replace(href: string, options: unknown): void }
declare const router: Router;
declare const options: unknown;
router.replace("/about", options);`),
			// Type information applies to JavaScript too when the Program provides it.
			jsValid(`/** @type {{replace(href: string, options: unknown): void}} */
const router = {};
router.replace("/about", options);`),

			// Locks in upstream isAllowedReplacement() static String.raw arm.
			jsValid("template.replace(\"x\", ((String.raw))`safe`)"),
			// Locks in ESLint's value-reference semantics: type-only String declarations do not shadow.
			tsValid(`interface String {} template.replace("x", String.raw` + "`safe`" + `)`),
			// Locks in upstream isPlainObjectReplacement() const-initializer arm.
			jsValid(`const options = {locale: value}; template.replace("x", options)`),
			// Locks in upstream union combination: an all-non-string union is skipped.
			tsValid(`declare const value: number | boolean; value.replace("x", replacement)`),
			// Boolean literals have intrinsic names upstream and are known non-string.
			tsValid(`declare const value: {flag: true}; value.flag.replace("x", replacement)`),
			tsValid(`declare function getFlag(): true; getFlag().replace("x", replacement)`),
			// Locks in upstream type-parameter constraint recursion.
			tsValid(`function run<T extends number>(value: T) { value.replace("x", replacement); }`),
			// Locks in upstream class-heritage recursion: String wrapper subclasses are non-string receivers.
			tsValid(`class Text extends String {} declare const value: Text; value.replace("x", replacement)`),

			// ---- Provably non-string receivers stay silent, whatever route reaches the type ----
			// The rule reports on undecided types, so numeric, bigint and tuple
			// types must be decided rather than left unknown. Upstream reaches
			// the same verdict for annotated receivers through its syntax
			// classifier; see the deliberate divergences in the invalid list for
			// the receivers where it cannot.
			tsValid(`declare const value: 1; value.replace("x", replacement)`),
			tsValid(`declare const value: 1n; value.replace("x", replacement)`),
			tsValid(`type N = 1; declare const value: N; value.replace("x", replacement)`),
			tsValid(`function run<T extends 1>(value: T) { value.replace("x", replacement); }`),
			tsValid(`declare const value: 1 | 2; value.replace("x", replacement)`),
			tsValid(`declare const value: 1 & {brand: true}; value.replace("x", replacement)`),
			tsValid(`declare const value: [string]; value.replace("x", replacement)`),
			tsValid(`declare const value: readonly [string]; value.replace("x", replacement)`),
			tsValid(`declare const value: [a: string]; value.replace("x", replacement)`),
			tsValid(`declare const value: [string?]; value.replace("x", replacement)`),
			tsValid(`type Pair = [string, number]; declare const value: Pair; value.replace("x", replacement)`),
			tsValid(`function run<T extends [string]>(value: T) { value.replace("x", replacement); }`),
			tsValid(`declare const value: [string] | 1; value.replace("x", replacement)`),
			// Deliberate divergence from upstream: an inferred literal type is the
			// same type as an annotated one, so a type-information-only classifier
			// cannot report the first while skipping the second. Upstream reports
			// these because its syntax classifier sees no annotation and its type
			// classifier leaves symbol-less literal types unknown.
			tsValid(`const value = 1; value.replace("x", replacement)`),
			tsValid(`(1).replace("x", replacement)`),
			tsValid(`(1n).replace("x", replacement)`),
			jsValid(`(1).replace("x", replacement)`),
			// Same divergence after control-flow narrowing: the narrowed type is 1 | 2.
			tsValid(`function run(value: 1 | 2 | string) { if (typeof value === "number") { value.replace("x", replacement) } }`),

			// N/A: private object-literal keys are invalid JavaScript syntax.
			// N/A: autofix boundaries do not apply; this rule has no fix or suggestion.
			// N/A: class/function nesting state does not apply; the listener is stateless.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: dynamic replacements under wrappers still report ----
			tsInvalid(`template.replace("x", ((replacement)))`, `replacement`, "replace"),
			tsInvalid(`template.replace("x", (replacement as string)!)`, `(replacement as string)!`, "replace"),
			tsInvalid(`template.replace("x", replacement satisfies string)`, `replacement satisfies string`, "replace"),

			// ---- Dimension 4: receiver wrappers and optional chains remain method calls ----
			tsInvalid(`((template)).replace("x", replacement)`, `replacement`, "replace"),
			tsInvalid(`(template as string).replace("x", replacement)`, `replacement`, "replace"),
			tsInvalid(`template!.replace("x", replacement)`, `replacement`, "replace"),
			tsInvalid(`template?.replace?.("x", replacement)`, `replacement`, "replace"),
			invalid(`(template.replace)("x", replacement)`, `replacement`, "replace"),

			// ---- Dimension 4: computed/method/accessor/spread object members are not plain ----
			invalid(`template.replace("x", {["safe"]: value})`, `{["safe"]: value}`, "replace"),
			invalid(`template.replace("x", {safe() {}})`, `{safe() {}}`, "replace"),
			invalid(`template.replace("x", {get safe() { return value; }})`, `{get safe() { return value; }}`, "replace"),
			invalid(`template.replace("x", {...value})`, `{...value}`, "replace"),

			// ---- Dimension 4: all object coercion property spellings are unsafe ----
			invalid(`template.replace("x", {"toString": value})`, `{"toString": value}`, "replace"),
			invalid(`template.replace("x", {valueOf})`, `{valueOf}`, "replace"),
			invalid(`template.replace("x", {__proto__: null})`, `{__proto__: null}`, "replace"),

			// ---- Dimension 4: sibling calls in nested containers report independently ----
			{
				Code:     `function run() { template.replace("x", outer); return () => template.replaceAll("x", inner); }`,
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(`function run() { template.replace("x", outer); return () => template.replaceAll("x", inner); }`, `outer`, "replace"),
					expectedError(`function run() { template.replace("x", outer); return () => template.replaceAll("x", inner); }`, `inner`, "replaceAll"),
				},
			},

			// ---- Real-user: #2309 HTML-escaped dynamic value still expands $` ----
			invalid(`function imageLink(url) {
	const template = '<img src="{url}">';
	return template.replace('{url}', htmlEscape(url));
}`, `htmlEscape(url)`, "replace"),

			// ---- Real-user: #2309 unknown identifier must remain conservative ----
			invalid(`IMAGE_TEMPLATE.replace('{url}', safeReplacement)`, `safeReplacement`, "replace"),

			// Locks in upstream getConstVariableInitializer(): aliases are not followed.
			invalid(`const options = {}; const alias = options; template.replace("x", alias)`, `alias`, "replace", 2),
			// Locks in upstream getConstVariableInitializer(): let bindings are not plain-object exemptions.
			invalid(`let options = {}; template.replace("x", options)`, `options`, "replace", 2),
			// Locks in upstream isStaticStringRawTaggedTemplate(): computed String.raw is rejected.
			invalid("template.replace(\"x\", String[\"raw\"]`safe`)", "String[\"raw\"]`safe`", "replace"),
			// Locks in upstream global-reference check for String.raw.
			invalid("function run(String) { template.replace(\"x\", String.raw`safe`); }", "String.raw`safe`", "replace"),
			// Locks in ESLint's value-reference semantics: a namespace is a runtime String binding.
			tsInvalid(`namespace String {} template.replace("x", String.raw`+"`safe`"+`)`, "String.raw`safe`", "replace"),
			// Locks in upstream union combination: mixed string/non-string is unknown and reports.
			tsInvalid(`declare const value: string | number; value.replace("x", replacement)`, `replacement`, "replace"),
			// Locks in upstream isKnownNonString() target arm: known strings remain checked.
			tsInvalid(`declare const value: string; value.replace("x", replacement)`, `replacement`, "replace"),
			// Locks in upstream type-parameter target-constraint recursion.
			tsInvalid(`function run<T extends string>(value: T) { value.replace("x", replacement); }`, `replacement`, "replace"),
			// Locks in upstream intersection combination: one string constituent is a target.
			tsInvalid(`declare const value: string & {brand: true}; value.replace("x", replacement)`, `replacement`, "replace"),
			// Locks in upstream union combination: nullish plus string is conservatively unknown.
			tsInvalid(`declare const value: string | null; value.replace("x", replacement)`, `replacement`, "replace"),
			// Locks in upstream unknown type arm.
			tsInvalid(`declare const value: unknown; value.replace("x", replacement)`, `replacement`, "replace"),
			// A union that mixes a string with a non-string is still undecided.
			tsInvalid(`declare const value: 1 | string; value.replace("x", replacement)`, `replacement`, "replace"),
			// Template literal types with substitutions are not TSLiteralType and
			// remain string targets.
			tsInvalid("declare const value: `a${string}`; value.replace(\"x\", replacement)", `replacement`, "replace"),
			// Deliberate divergence from upstream: a no-substitution template
			// literal type is a string, so the receiver is a target here even
			// though upstream's isStringTypeAnnotation() rejects it. Every route
			// to the literal type behaves the same.
			tsInvalid("declare const value: `x`; value.replace(\"x\", replacement)", `replacement`, "replace"),
			tsInvalid("declare const value: `x` | 1; value.replace(\"x\", replacement)", `replacement`, "replace"),
			tsInvalid("declare const value: `x` & {}; value.replace(\"x\", replacement)", `replacement`, "replace"),
			tsInvalid("type T = `x`; declare const value: T; value.replace(\"x\", replacement)", `replacement`, "replace"),
			tsInvalid("function run<T extends `x`>(value: T) { value.replace(\"x\", replacement); }", `replacement`, "replace"),
			// A merged type/value name has multiple definitions and must remain unknown.
			tsInvalid(`type Sep = number; const Sep = "-"; Sep.replace("x", replacement)`, `replacement`, "replace"),
			tsInvalid(`type T = {a: 1}; const T = "x"; T.replace("x", replacement)`, `replacement`, "replace"),
		},
	)
}
