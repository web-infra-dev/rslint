package no_top_level_await_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_top_level_await"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics were checked against eslint-plugin-n v18.3.0 with
// @typescript-eslint/parser 8.57.0, including exact UTF-16 ranges.
func TestNoTopLevelAwaitExtras(t *testing.T) {
	rule_tester.RunRuleTester(awaitRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_top_level_await.NoTopLevelAwaitRule,
		[]rule_tester.ValidTestCase{
			// async function expression and object generator.
			{Code: "const fn = async function () { await load(); }; const o = { async *method() { for await (const x of stream) { yield x; } } };", FileName: "published/index.ts"},
			// private method and static block callbacks.
			{Code: "class C { async #method() { await load(); } static { (async () => { await load(); })(); } get value() { return async () => await load(); } set value(v) { void (async () => await v)(); } }", FileName: "published/index.ts"},
			// computed names inside an outer function.
			{Code: "async function outer() { class C { [await key()]() {} } const o = { [await key()]() {} }; }", FileName: "published/index.ts"},
			// ordinary using does not await.
			{Code: "using value = open();", FileName: "published/index.ts"},
			// await using inside methods and arrows.
			{Code: "const f = async () => { await using value = open(); }; class C { async method() { await using value = open(); } }", FileName: "published/index.ts"},
			// bin alias is ignored after conversion.
			{Code: "await load();", FileName: "files/src/cli.ts", Options: []any{map[string]any{"ignoreBin": true, "convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}},
			// bin index alias is ignored.
			{Code: "await load();", FileName: "files/lib/cli/index.js", Options: []any{map[string]any{"ignoreBin": true}}},
			// conversion arrays support exclusions.
			{Code: "await load();", FileName: "files/src/test/index.ts", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}},
			// converted path still obeys npmignore.
			{Code: "await load();", FileName: "files/src/ignored.ts", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}},
			// explicit conversion overrides shared settings.
			{Code: "await load();", FileName: "files/src/index.ts", Options: []any{map[string]any{"convertPath": map[string]any{}}}, Settings: map[string]any{"node": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}},
			// any env hashbang ignores a bin.
			{Code: "#!/usr/bin/env custom\nawait load();", FileName: "published/index.ts", Options: []any{map[string]any{"ignoreBin": true}}},
			// BOM is stripped before testing the hashbang.
			{Code: "\uFEFF#!/usr/bin/env node\nawait load();", FileName: "published/index.ts", Options: []any{map[string]any{"ignoreBin": true}}},
			// gitignore fallback.
			{Code: "await load();", FileName: "gitignore/ignored.js"},
			// non-main file is unpublished.
			{Code: "await load();", FileName: "main/other.js"},
		},
		[]rule_tester.InvalidTestCase{
			// top-level block.
			{Code: "if (ready) { await load(); }", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 14, 1, 26)}},
			// parentheses, optional call and type assertion.
			{Code: "export const data = (await (load?.())) as Data;", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 22, 1, 38)}},
			// class heritage is evaluated outside its methods.
			{Code: "class C extends (await base()) { async method() { await local(); } }", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 18, 1, 30)}},
			// computed object method and accessor names.
			{Code: "const object = { async [await key()]() { await local(); }, get [await otherKey()]() { return 1; }, set [await setterKey()](value) {} };", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 25, 1, 36), forbiddenAt(1, 65, 1, 81), forbiddenAt(1, 105, 1, 122)}},
			// computed class method, accessor and field names.
			{Code: "class C { [await key()]() {} get [await getKey()]() { return 1; } set [await setKey()](v) {} [await fieldKey()] = 1; static [await staticKey()] = 2; }", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 12, 1, 23), forbiddenAt(1, 35, 1, 49), forbiddenAt(1, 72, 1, 86), forbiddenAt(1, 95, 1, 111), forbiddenAt(1, 126, 1, 143)}},
			// class and method decorators.
			{Code: "@decorate(await ready())\nclass C { @decorate(await methodReady()) method() {} }", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 11, 1, 24), forbiddenAt(2, 21, 2, 40)}},
			// JSX expression.
			{Code: "const view = <Widget value={await load()} />;", FileName: "published/view.tsx", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 29, 1, 41)}},
			// bodyless declarations do not affect the next await.
			{Code: "declare function load(): Promise<void>;\nawait load();", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 1, 2, 13)}},
			// await using includes multiple declarators and the semicolon.
			{Code: "await using first = open(), second = open();", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 45)}},
			// multiline await using range.
			{Code: "{\n  /* 🐱 */ await using café = (\n    open()\n  );\n}", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 12, 4, 5)}},
			// await using in a loop header.
			{Code: "for (await using item of items) {}", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 6, 1, 22)}},
			// all four reporting sites in one loop.
			{Code: "for await (await using item of await items) { await item.read(); }", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 67), forbiddenAt(1, 12, 1, 28), forbiddenAt(1, 32, 1, 43), forbiddenAt(1, 47, 1, 64)}},
			// Unicode and comments before the awaited expression.
			{Code: "/* 🐱 */ const café = await café?.(); // trailing", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 23, 1, 37)}},
			// parenthesized await.
			{Code: "(await (load()));", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 2, 1, 16)}},
			// nested await expressions.
			{Code: "await (await load());", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 21), forbiddenAt(1, 8, 1, 20)}},
			// multiline for await range.
			{Code: "for await (\n const item of stream\n) {\n consume(item);\n}", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 5, 2)}},
			// await following an async function.
			{Code: "const task = async () => await local();\nawait outer();", FileName: "published/index.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 1, 2, 14)}},
			// explicit empty options.
			{Code: "await load();", FileName: "published/index.ts", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// explicit default checks bins.
			{Code: "await load();", FileName: "files/lib/cli.js", Options: []any{map[string]any{"ignoreBin": false}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// conversion array selects a published file.
			{Code: "await load();", FileName: "files/src/index.ts", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// settings.node conversion.
			{Code: "await load();", FileName: "files/src/index.ts", Settings: map[string]any{"node": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// legacy settings.n conversion.
			{Code: "await load();", FileName: "files/src/index.ts", Settings: map[string]any{"n": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// settings.n takes precedence.
			{Code: "await load();", FileName: "files/src/index.ts", Settings: map[string]any{"n": map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/test/**"}, "replace": []any{"^src/(.*)\\.ts$", "lib/$1.js"}}}}, "node": map[string]any{"convertPath": map[string]any{}}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// direct interpreter is not the env prefix.
			{Code: "#!/usr/bin/node\nawait load();", FileName: "published/index.ts", Options: []any{map[string]any{"ignoreBin": true}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 1, 2, 13)}},
			// npmignore replaces gitignore.
			{Code: "await load();", FileName: "npmignore/included.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// files list disables gitignore fallback.
			{Code: "await load();", FileName: "files-gitignore/lib/index.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			// main remains published despite an empty files list.
			{Code: "await load();", FileName: "main/index.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
		},
	)
}

// These cases retain nodeutil's established, documented differences from upstream.
func TestNoTopLevelAwaitSharedPathPolicy(t *testing.T) {
	rule_tester.RunRuleTester(awaitRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_top_level_await.NoTopLevelAwaitRule,
		[]rule_tester.ValidTestCase{
			// Object mappings have lexical priority; use arrays to control it.
			{Code: "await load();", FileName: "files/src/index.ts", Options: map[string]any{"convertPath": map[string]any{
				"src/**": []any{`^src/(.*)\.ts$`, "lib/$1.js"},
				"**":     []any{`^src/(.*)\.ts$`, "test/$1.js"},
			}}},
			// An invalid conversion regexp skips the rule instead of throwing.
			{Code: "await load();", FileName: "published/index.ts", Options: map[string]any{"convertPath": map[string]any{"**": []any{"[", "lib/index.js"}}}},
			// Publication is relative to the converted target's package.
			{Code: "await load();", FileName: "files/src/index.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{".*", "lib/nested/private.js"}}}},
			{Code: "await load();", FileName: "files/src/index.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{".*", "../outside.js"}}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "await load();", FileName: "files/src/index.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{".*", "lib/nested/published.js"}}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			{Code: "await load();", FileName: "metadata/README.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			{Code: "await load();", FileName: "published/..hidden.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			{Code: "await load();", FileName: "malformed-bin/index.js", Options: map[string]any{"ignoreBin": true}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
		},
	)
}

func TestNoTopLevelAwaitSchema(t *testing.T) {
	for _, options := range [][]any{
		{map[string]any{"ignoreBin": "true"}},
		{map[string]any{"unknown": true}},
		{map[string]any{"convertPath": nil}}, // The upstream docs show null, but its schema rejects it.
		{map[string]any{"convertPath": []any{}}},
		{map[string]any{"convertPath": map[string]any{"src/**": []any{"src", "lib", "extra"}}}},
		{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}}}}},
		{map[string]any{}, map[string]any{}},
	} {
		if err := no_top_level_await.NoTopLevelAwaitRule.Schema.Validate(options); err == nil {
			t.Errorf("expected invalid options to be rejected: %#v", options)
		}
	}
}

// ESTree's function boundaries differ from this-containers and TypeScript emit
// boundaries: parameter decorators are inside functions; type-only computed
// names, namespaces and enums still expose await expressions to this rule.
func TestNoTopLevelAwaitASTBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(awaitRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_top_level_await.NoTopLevelAwaitRule,
		[]rule_tester.ValidTestCase{
			// Parameter decorator belongs to the method function.
			{Code: "class C { method(@decorate(await load()) value: unknown) {} }", FileName: "published/input.ts"},
			// Constructor parameter decorator belongs to the function.
			{Code: "class C { constructor(@decorate(await load()) value: unknown) {} }", FileName: "published/input.ts"},
			// Method decorator evaluated in containing function.
			{Code: "async function outer() { class C { @decorate(await load()) method() {} } }", FileName: "published/input.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// JSDoc cast around await.
			{Code: "const value = /** @type {unknown} */ (await load());", FileName: "published/input.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 39, 1, 51)}},
			// Comments within await using.
			{Code: "await /* comment */ using value = open();", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 42)}},
			// Loop with an empty body.
			{Code: "for await (const item of stream);", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 34)}},
			// CRLF loop range.
			{Code: "for await (const item of stream) {\r\n  await consume(item);\r\n}", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 3, 2), forbiddenAt(2, 3, 2, 22)}},
			// Await using declaration and awaited initializer.
			{Code: "await using value = await open();", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 34), forbiddenAt(1, 21, 1, 33)}},
			// Interface computed method name.
			{Code: "interface I { [await key()](): void; }", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 16, 1, 27)}},
			// Type literal computed method name.
			{Code: "type T = { [await key()](): void };", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 24)}},
			// Abstract computed method name.
			{Code: "abstract class C { abstract [await key()](): void; }", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 30, 1, 41)}},
			// Namespace top-level await.
			{Code: "namespace N { await load(); }", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 15, 1, 27)}},
			// Enum member initializer.
			{Code: "enum E { Value = await load() }", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 18, 1, 30)}},
			// Computed member after abstract and overloaded methods.
			{Code: "abstract class C { abstract method(): void; method2(value: string): void; method2(value: unknown) {} [await key()]() {} }", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 103, 1, 114)}},
			// tsgo currently parses await(...) as a call inside computed names:
			// parsePropertyName restores statementHasAwaitIdentifier, preventing
			// the module await reparse. Keep the upstream expectation until the
			// parser supports it; do not reconstruct expressions in this rule.
			{Code: "const object = { [await (async () => { await load(); return key(); })()]() {} };", FileName: "published/input.ts", Skip: true, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 19, 1, 72)}},
			// Both documented rewrites expose an unambiguous await expression.
			{Code: "const object = { [await keyPromise]() {} };", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 19, 1, 35)}},
			{Code: "const key = await (keyPromise); const object = { [key]() {} };", FileName: "published/input.ts", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 31)}},
		},
	)
}
