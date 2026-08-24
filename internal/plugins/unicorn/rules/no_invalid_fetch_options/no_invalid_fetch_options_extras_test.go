// TestNoInvalidFetchOptionsExtras locks in branches and edge shapes that the upstream test suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named lock-in.
package no_invalid_fetch_options_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_fetch_options"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidFetchOptionsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_invalid_fetch_options.NoInvalidFetchOptionsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Optional calls do not match direct fetch calls ----
			jsValid(`fetch?.(url, {body})`),
			jsValid(`(fetch)?.(url, {body})`),

			// ---- Dimension 4: Type wrappers around callees remain significant ----
			tsValid(`(fetch as any)(url, {body})`),
			tsValid(`fetch!(url, {body})`),
			tsValid(`(fetch satisfies any)(url, {body})`),
			tsValid(`new (Request as any)(url, {body})`),

			// ---- Dimension 4: Type wrappers around the options object remain significant ----
			tsValid(`fetch(url, ({body} as RequestInit))`),
			tsValid(`fetch(url, ({body} satisfies RequestInit))`),

			// ---- Dimension 4: Only non-computed identifier keys are recognized ----
			jsValid(`fetch(url, {"body": data})`),
			jsValid(`fetch(url, {0: data})`),
			jsValid(`fetch(url, {["body"]: data})`),
			jsValid(`fetch(url, {[` + "`body`" + `]: data})`),
			// N/A: PrivateIdentifier is not valid as an object-literal key.

			// ---- Dimension 4: Spread assignments degrade conservatively without hiding explicit methods ----
			jsValid(`fetch(url, {body, ...options})`),
			jsValid(`fetch(url, {...options, body, ...moreOptions})`),

			// ---- Dimension 4: Empty and non-object options degrade gracefully ----
			jsValid(`fetch()`),
			jsValid(`fetch(url, null)`),
			jsValid(`fetch(url, options)`),
			jsValid(`fetch(url, {...options})`),
			jsValid(`fetch(url, {method: "GET"})`),
			// N/A: RestElement belongs to binding patterns, not request option literals.
			// N/A: Body-absent overload/abstract/declare members cannot occur in object literals.

			// ---- Locks in upstream isObjectPropertyWithName(): method/getter/setter values are unknown ----
			jsValid(`fetch(url, {method() {}, body})`),
			jsValid(`fetch(url, {get method() { return "GET" }, body})`),
			jsValid(`fetch(url, {set method(value) {}, body})`),

			// ---- Locks in upstream body-absence branches: parentheses are transparent, TS assertions are not ----
			jsValid(`fetch(url, {body: (undefined)})`),
			jsValid(`fetch(url, {body: ((null))})`),
			jsValid(`fetch(url, {body: (void sideEffect())})`),
			jsValid(`function run(undefined) { fetch(url, {body: undefined}); }`),

			// ---- Locks in upstream getStaticValue(): unknown and non-string methods are ignored ----
			jsValid(`fetch(url, {method: getMethod(), body})`),
			jsValid(`fetch(url, {method: 1, body})`),
			jsValid(`fetch(url, {method: true, body})`),
			jsValid(`let method = "GET"; method = "POST"; fetch(url, {method, body})`),
			jsValid(`fetch(url, {method: "OPTIONS", body})`),
			jsValid(`fetch(url, {method: "TRACE", body})`),
			jsValid(`fetch(url, {method: "CONNECT", body})`),
			// ECMAScript does not trim U+0085 before numeric coercion, so the first
			// character becomes NUL rather than G.
			jsValid(`fetch(url, {method: String.fromCharCode("\u008571", 69, 84), body})`),

			// ---- Locks in upstream findLast(): the last matching member controls behavior ----
			jsValid(`fetch(url, {body() {}, body: undefined})`),
			jsValid(`fetch(url, {method: "HEAD", body, method() {}})`),

			// ---- Locks in upstream listener/callee gates ----
			jsValid(`globalThis.fetch(url, {body})`),
			jsValid(`new globalThis.Request(url, {body})`),
			jsValid(`fetch.call(globalThis, url, {body})`),
			jsValid(`new fetch(url, {body})`),
			jsValid(`Request(url, {body})`),

			// ---- Real-user: PR #2338 discussion confirms Fetch allows these methods with bodies ----
			jsValid(`fetch("/", {method: "OPTIONS", body: "foo"})`),
			jsValid(`fetch("/", {method: "TRACE", body: "foo"})`),

			// N/A: The rule has no autofix or suggestions.
			// N/A: The rule does not target function/class declarations or their async/generator variants.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: Parenthesized callees and options are transparent ----
			invalid(`(fetch)(url, {body})`, "body", "GET"),
			invalid(`((fetch))(url, (({body})))`, "body", "GET"),
			invalid(`new (Request)(url, ({body}))`, "body", "GET"),
			// Diagnostic columns count UTF-16 code units, as ESLint does.
			invalid(`fetch("😀", {body})`, "body", "GET"),

			// ---- Dimension 4: Type arguments do not wrap the callee ----
			tsInvalid(`fetch<string>(url, {body})`, "body", "GET"),

			// ---- Dimension 4: TS wrappers are transparent to static method evaluation ----
			tsInvalid(`fetch(url, {method: ("head" as const), body})`, "body", "HEAD"),
			tsInvalid(`fetch(url, {method: "get"!, body})`, "body", "GET"),
			tsInvalid(`fetch(url, {method: ("GET" satisfies string), body})`, "body", "GET"),

			// ---- Dimension 4: TS wrappers prevent the body absence shortcut ----
			tsInvalid(`fetch(url, {body: (undefined as any)})`, "body", "GET"),
			tsInvalid(`fetch(url, {body: ((void 0) as any)})`, "body", "GET"),

			// ---- Dimension 4: Object members cover shorthand, methods, getters, and setters ----
			invalid(`fetch(url, {body() {}})`, "body", "GET"),
			invalid(`fetch(url, {get body() { return data; }})`, "body", "GET"),
			invalid(`fetch(url, {set body(value) {}})`, "body", "GET"),
			invalid(`fetch(url, {body, "method": "GET"})`, "body", "GET"),
			invalid(`fetch(url, {body, ["method"]: "GET"})`, "body", "GET"),

			// ---- Dimension 4: Traversal reaches nested function, class, and object contexts ----
			invalid(`function run() { fetch(url, {body}); }`, "body", "GET"),
			invalid(`class Client { send() { return new Request(url, {body}); } }`, "body", "GET"),
			invalid(`fetch(url, {body: {body: null}})`, "body", "GET"),

			// ---- Locks in upstream missing-method branch: local same-name bindings still match ----
			invalid(`function fetch() {} fetch(url, {body})`, "body", "GET"),
			invalid(`const Request = class {}; new Request(url, {body})`, "body", "GET"),

			// ---- Locks in upstream spread branch: an explicit method is still authoritative ----
			invalid(`fetch(url, {method: "GET", ...options, body})`, "body", "GET"),
			invalid(`fetch(url, {...options, method: "HEAD", body})`, "body", "HEAD"),

			// ---- Locks in upstream getStaticValue(): string-producing expression branches ----
			invalid(`fetch(url, {method: `+"`head`"+`, body})`, "body", "HEAD"),
			invalid(`fetch(url, {method: "he" + "ad", body})`, "body", "HEAD"),
			invalid(`fetch(url, {method: true ? "GET" : "POST", body})`, "body", "GET"),
			invalid(`const method = "get"; fetch(url, {method, body})`, "body", "GET"),
			invalid(`fetch(url, {method: "get".toUpperCase(), body})`, "body", "GET"),
			invalid(`fetch(url, {method: "get".toUpperCase(1), body})`, "body", "GET"),
			invalid(`new Request(url, {method: String.fromCharCode(72, 69, 65, 68), body})`, "body", "HEAD"),
			// StringToNumber accepts non-decimal prefixes and trims U+FEFF exactly
			// as JavaScript does.
			invalid(`fetch(url, {method: String.fromCharCode("0x47", 69, 84), body})`, "body", "GET"),
			invalid(`fetch(url, {method: String.fromCharCode("\uFEFF71", 69, 84), body})`, "body", "GET"),
			invalid(`fetch(url, {method: Array.of("GET")[0], body})`, "body", "GET"),
			invalid(`fetch(url, {method: "xGETy".slice(1, 4), body})`, "body", "GET"),
			invalid(`fetch(url, {method: "xHEADy".substring(1, 5), body})`, "body", "HEAD"),
			invalid(`fetch(url, {method: "xGETy".substr(1, 3), body})`, "body", "GET"),
			invalid(`fetch(url, {method: "G".concat("ET"), body})`, "body", "GET"),
			invalid(`fetch(url, {method: "GET".charAt(0) + "ET", body})`, "body", "GET"),
			invalid(`fetch(url, {method: +"71" === 71 ? "GET" : "POST", body})`, "body", "GET"),
			invalid(`const S = String; fetch(url, {method: S.fromCharCode(71, 69, 84), body})`, "body", "GET"),
			invalid(`const A = Array; fetch(url, {method: A.of("HEAD")[0], body})`, "body", "HEAD"),

			// ---- Locks in upstream findLast(): last body/method wins across member kinds ----
			invalid(`fetch(url, {body: undefined, body() {}})`, "body", "GET", 2),
			invalid(`fetch(url, {method() {}, method: "HEAD", body})`, "body", "HEAD"),

			// ---- Real-user: issue #1989 request body without an explicit POST method ----
			invalid(`fetch("http://example.com", {body: "{}"})`, "body", "GET"),

			// ---- Real-user: PR #3299 hardening keeps side-effecting void bodies absent ----
			// This is valid upstream behavior; the paired invalid assertion ensures only
			// the final non-void body member is selected.
			invalid(`fetch(url, {body: void sideEffect(), body: data})`, "body", "GET", 2),
		},
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func tsInvalid(code string, target string, method string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.ts",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, method, occurrence...),
		},
	}
}
