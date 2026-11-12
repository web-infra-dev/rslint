package no_mismatched_map_key_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_mismatched_map_key"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoMismatchedMapKeyExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Unsupported guard and access shapes.
		{FileName: "case.js", Code: "if (!!map.has(a)) map.get(b); if (map.has(a, b)) map.get(c);"},
		{FileName: "case.js", Code: "if (map.has(a)) map.get(); if (map.has(a)) map.delete(b);"},
		{FileName: "case.js", Code: "if (map.has(a)) map.get(...b); if (map.has(a)) map.get?.(b);"},
		{FileName: "case.js", Code: "if (cache?.map.has(a)) cache.map.get(b); if (map.has(a)) (map?.get)(b);"},
		{FileName: "case.js", Code: "if (map.has(keys?.a)) map.get(keys.b); if (map.has(a)) map.get(keys[makeKey()]);"},
		{FileName: "case.ts", Code: "if ((ready && map.has(a)) as boolean) map.get(b); if ((map.has as typeof map.has)(a)) map.get(b);"},

		// Receiver writes and nested scopes.
		{FileName: "case.js", Code: "if (map.has(a)) { for (map in entries) map.get(b); }"},
		{FileName: "case.js", Code: "if (map.has(a)) use(({map} = next), map.get(b));"},
		{FileName: "case.js", Code: `if (map.has(a)) use(([(map)] = next), map.get(b));
if (map.has(a)) use(({p: (map)} = next), map.get(b));`},
		{FileName: "case.js", Code: `if (map.has(a)) use(([(map) = other] = next), map.get(b));
if (map.has(a)) { for ([(map)] of next) map.get(b); }`},
		{FileName: "case.js", Code: "if (cache.map.has(a)) { ([{value: cache}] = next, cache.map.get(b)); }"},
		{FileName: "case.js", Code: `if (holder.map.has("a")) ([holder.map] = [other], holder.map.get("b"));
if (holder.map.has("a")) ({value: holder.map = other} = next, holder.map.get("b"));
if (holder.map.has("a")) { for ([holder.map] of next) holder.map.get("b"); }`},
		{FileName: "case.js", Code: "if (map.has(a)) { if (other || map.has(b)) map.get(c); }"},
		{FileName: "case.js", Code: "if (map.has(a)) use(() => map.get(b), function () { map.get(c); });"},

		// A computed receiver can change through index shadowing or assignment.
		{FileName: "case.js", Code: `const cache = {a: new Map(), b: new Map()};
const k = "a";
if (cache[k].has("x")) { const k = "b", value = cache[k].get("y"); }`},
		{FileName: "case.js", Code: `let k = "a";
if (cache[k].has("x")) (k = "b", cache[k].get("y"));
if (cache[k].has("x")) ([k] = next, cache[k].get("y"));
if (cache[config.key].has("x")) (config.key = "b", cache[config.key].get("y"));`},
		{FileName: "case.js", Code: `let k = "a";
if (cache[k].has("x")) {
  for (const item of items) {
    if (cache[k].has("x")) k = "b";
    cache[k].get("y");
  }
}`},

		// Static keys with different syntax or bindings.
		{FileName: "case.js", Code: "if (map.has(0 / 0)) map.get(NaN); if (map.has(1n)) map.get(0x1n);"},
		{FileName: "case.js", Code: "const k1 = 3; const k2 = 3; if (map.has(k1)) map.get(k2);"},
		{FileName: "case.js", Code: "if (map.has(`a`)) map.get(\"a\");"},
		{FileName: "case.js", Code: `const key = {}, alias = key;
const map = new Map([[key, 1]]);
if (map.has(key)) map.get(alias);`},
		{FileName: "case.ts", Code: `const key = {}, alias = key as object, secondAlias = (alias!);
if (map.has(secondAlias)) map.get(key);`},

		// Source annotations also work through const aliases and local factories.
		{FileName: "case.ts", Code: `type Cache = { has(k: string): boolean; get(k: string): unknown };
function make(): Cache { throw 0; }
const cache = make();
if (cache.has(a)) cache.get(b);`},
		{FileName: "case.ts", Code: `type Cache = { has(k: string): boolean; get(k: string): unknown };
function f(cache: Cache) {
  const alias = (0, cache! satisfies Cache);
  if (alias.has(a)) alias.get(b);
}`},
		{FileName: "case.js", Code: `/** @type {{has: function(string): boolean, get: function(string): number}} */
const cache = getCache();
if (cache.has(a)) cache.get(b);`},
		// Rslint also recognizes inline casts and array annotations.
		{FileName: "case.js", Code: `const cache = /** @type {{has: (key: string) => boolean, get: (key: string) => number}} */ (loadCache());
if (cache.has("a")) cache.get("b");`},
		{FileName: "case.js", Code: `/** @type {*} */
const cache = /** @type {{has: (key: string) => boolean, get: (key: string) => number}} */ (loadCache());
if (cache.has("a")) cache.get("b");`},
		{FileName: "case.js", Code: `/** @type {string[]} */
const values = loadValues();
if (values.has("a")) values.get("b");`},
	}
	invalid := []rule_tester.InvalidTestCase{
		keyCase("case.js", "map.has(a) ?? map.get(b);", 1, 23, 1, 24),
		keyCase("case.js", "if (ready && (enabled && (map.has(a)))) map.get(b);", 1, 49, 1, 50),
		keyCase("case.js", "if (map.has(a)) { map.has(b) && map.get(c); }", 1, 41, 1, 42),
		keyCase("case.js", "if (map.has(a)) { if (otherMap.has(b)) map.get(c); }", 1, 48, 1, 49),
		keyCase("case.js", "if (map.has(a)) { (map ||= other, map.get(b)); } else map.get(c);", 1, 63, 1, 64),
		keyCase("case.js", "map.has(a) ? map.get(b) : (map = other, map.get(c));", 1, 22, 1, 23),
		{
			FileName: "case.js",
			Code:     "if (map.has(a)) { use(map.get(b), map.set(c, 1)); }",
			Errors:   []rule_tester.InvalidTestCaseError{keyError(1, 31, 1, 32), keyError(1, 43, 1, 44)},
		},
		// Computed method names run in the branch, unlike their function bodies.
		keyCase("case.js", "if (map.has(a)) { use({ [map.get(b)]() { map.get(c); } }); }", 1, 34, 1, 35),
		// The computed property name reads the receiver during destructuring.
		keyCase("case.js", "if (cache.map.has(a)) { ({[cache.map]: other} = next, cache.map.get(b)); }", 1, 69, 1, 70),
		keyCase("case.js", `if (cache.map.has(a)) cache["m" + "ap"].get(b);`, 1, 45, 1, 46),
		keyCase("case.js", "if (cache[k].has(a)) cache[k].get(b);", 1, 35, 1, 36),
		keyCase("case.js", `let k = "a";
if (cache[k].has("x")) {
  for (const item of items) {
    { let k = "b"; consume(k); }
    cache[k].get("y");
  }
}`, 5, 18, 5, 21),
		keyCase("case.js", "let key = {}; const alias = key; key = {}; if (map.has(key)) map.get(alias);", 1, 70, 1, 75),
		keyCase("case.js", "const key = {}, other = {}; if (map.has(key)) map.get(other);", 1, 55, 1, 60),
		keyCase("case.js", "if ((cache?.map).has(a)) cache?.map.get(b);", 1, 41, 1, 42),
		keyCase("case.js", "class Example { #map; run() { if (this.#map.has(a)) this.#map.get(b); } }", 1, 67, 1, 68),
		keyCase("case.js", `if (map.has(a)) map.get(keys[import("pkg")]);`, 1, 25, 1, 44),
		keyCase("case.js", "if (map.has(1n)) map.get(1);", 1, 26, 1, 27),
		keyCase("case.js", "if (map.has(null)) map.get(void 0);", 1, 28, 1, 34),
		keyCase("case.js", "if (map.has(Symbol.iterator)) map.get(Symbol.asyncIterator);", 1, 39, 1, 59),
		keyCase("case.js", `if (map.has("🔑")) map.get((/* comment */ "🗝️"));`, 1, 43, 1, 48),
		keyCase("case.js", `if (map.has("🔑")) {
  map.get((
    "🗝️" +
    "b"
  ));
}`, 3, 5, 4, 8),
		keyCase("case.jsx", "function App() { if (map.has(a)) return <div>{map.get(b)}</div>; }", 1, 55, 1, 56),
		// Authored TypeScript wrappers are compared transparently and reported in full.
		keyCase("case.ts", "if ((map.has(a) as boolean)) map.get(b!);", 1, 38, 1, 40),
		keyCase("case.ts", "if ((map.has(a))!) map.get((b satisfies string));", 1, 29, 1, 47),
		keyCase("case.ts", "if ((map as Map<string, number>).has(a)) map.get(<string>b);", 1, 50, 1, 59),
		keyCase("case.ts", "if (map.has(a)) map.get(b<string>);", 1, 25, 1, 34),
		keyCase("case.ts", "if (map.has(a as {[test()]: number})) map.get(b);", 1, 47, 1, 48),
		keyCase("case.ts", "interface Cache extends Map<string, number> {} function f(cache: Cache) { if (cache.has(a)) cache.get(b); }", 1, 103, 1, 104),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_mismatched_map_key.NoMismatchedMapKeyRule, valid, invalid)
}

func TestNoMismatchedMapKeySourceOnlyGenericAlias(t *testing.T) {
	t.Parallel()
	const code = `type Cache = { has(key: string): boolean; get(key: string): number };
type Identity<T> = T;
function read(cache: Identity<Cache>) {
  if (cache.has("a")) cache.get("b");
}
const map = new Map<string, number>();
if (map.has("first")) map.get("second");`
	root := fixtures.GetRootDir()
	name := tspath.ResolvePath(root.Dir, "generic-alias.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{name: code})
	source, err := program.NewFromRoots(program.RootOptions{
		RootFileNames:   []string{name},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: program.SourceOnlyCompilerOptions(),
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program: source,
		Files:   []string{name},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_mismatched_map_key.NoMismatchedMapKeyRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("expected a file without project type information")
					}
					return no_mismatched_map_key.NoMismatchedMapKeyRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic on the Map, got %+v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Message.Id != "no-mismatched-map-key" ||
		code[diagnostic.Range.Pos():diagnostic.Range.End()] != `"second"` {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
}
