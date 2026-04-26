package filename_case_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/filename_case"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// caseOpt builds a `[{case: ...}]`-style option payload (single-case form).
func caseOpt(c string) map[string]interface{} {
	return map[string]interface{}{"case": c}
}

// casesOpt builds a `[{cases: {...}}]`-style option payload (multi-case form).
func casesOpt(m map[string]bool) map[string]interface{} {
	cases := make(map[string]interface{}, len(m))
	for k, v := range m {
		cases[k] = v
	}
	return map[string]interface{}{"cases": cases}
}

// withMfe returns a clone of `base` with `multipleFileExtensions` set.
func withMfe(base map[string]interface{}, v bool) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	out["multipleFileExtensions"] = v
	return out
}

// withIgnore returns a clone of `base` plus an `ignore` array.
func withIgnore(base map[string]interface{}, patterns ...string) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	pats := make([]interface{}, len(patterns))
	for i, p := range patterns {
		pats[i] = p
	}
	out["ignore"] = pats
	return out
}

// All upstream `valid` / `invalid` cases are migrated below in the same
// declaration order. Cases that the rule_tester cannot exercise (TypeScript's
// program builder skips files whose extension or basename it does not
// recognize, leaving us with a nil source file) are kept as `Skip: true`
// entries with a reason — they are covered indirectly by the JS-side tests
// against the rslint binary, where the program builder is bypassed.
//
// Inline `Locks in upstream` comments mark lock-in tests we added for branches
// the upstream test suite itself does not cover.
func TestFilenameCase(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&filename_case.FilenameCaseRule,
		[]rule_tester.ValidTestCase{
			// ---- Single `case` option, all four styles ----
			{Code: `// camel`, FileName: "src/foo/bar.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "src/foo/fooBar.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "src/foo/bar.test.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "src/foo/fooBar.test.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "src/foo/fooBar.test-utils.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "src/foo/fooBar.test_utils.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "src/foo/.test_utils.js", Options: caseOpt("camelCase"),
				Skip: true /* SKIP: dot-prefixed basename — TS program skips it */},

			{Code: `// snake`, FileName: "src/foo/foo.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "src/foo/foo_bar.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "src/foo/foo.test.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "src/foo/foo_bar.test.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "src/foo/foo_bar.test_utils.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "src/foo/foo_bar.test-utils.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "src/foo/.test-utils.js", Options: caseOpt("snakeCase"),
				Skip: true /* SKIP: dot-prefixed basename */},

			{Code: `// kebab`, FileName: "src/foo/foo.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "src/foo/foo-bar.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "src/foo/foo.test.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "src/foo/foo-bar.test.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "src/foo/foo-bar.test-utils.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "src/foo/foo-bar.test_utils.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "src/foo/.test_utils.js", Options: caseOpt("kebabCase"),
				Skip: true /* SKIP: dot-prefixed basename */},

			{Code: `// pascal`, FileName: "src/foo/Foo.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "src/foo/FooBar.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "src/foo/Foo.test.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "src/foo/FooBar.test.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "src/foo/FooBar.test-utils.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "src/foo/FooBar.test_utils.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "src/foo/.test_utils.js", Options: caseOpt("pascalCase"),
				Skip: true /* SKIP: dot-prefixed basename */},

			// ---- Numeric / mixed identifier cases ----
			{Code: `// camel`, FileName: "spec/iss47Spec.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "spec/iss47Spec100.js", Options: caseOpt("camelCase")},
			{Code: `// camel`, FileName: "spec/i18n.js", Options: caseOpt("camelCase")},
			{Code: `// kebab`, FileName: "spec/iss47-spec.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "spec/iss-47-spec.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "spec/iss47-100spec.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab`, FileName: "spec/i18n.js", Options: caseOpt("kebabCase")},
			{Code: `// snake`, FileName: "spec/iss47_spec.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "spec/iss_47_spec.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "spec/iss47_100spec.js", Options: caseOpt("snakeCase")},
			{Code: `// snake`, FileName: "spec/i18n.js", Options: caseOpt("snakeCase")},
			{Code: `// pascal`, FileName: "spec/Iss47Spec.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "spec/Iss47.100spec.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal`, FileName: "spec/I18n.js", Options: caseOpt("pascalCase")},

			// ---- `testCase(undefined, ...)`: upstream uses no filename, which
			// becomes `<input>` in ESLint's `physicalFilename` and is no-op'd
			// by the rule. rule_tester always substitutes a real filename, so
			// we cannot reach the `<input>`/`<text>` branch through it. These
			// stay as `Skip` markers so the count matches upstream.
			{Code: `// undef`, Options: caseOpt("camelCase"), Skip: true /* SKIP: rule_tester always supplies a filename */},
			{Code: `// undef`, Options: caseOpt("snakeCase"), Skip: true /* SKIP: rule_tester always supplies a filename */},
			{Code: `// undef`, Options: caseOpt("kebabCase"), Skip: true /* SKIP: rule_tester always supplies a filename */},
			{Code: `// undef`, Options: caseOpt("pascalCase"), Skip: true /* SKIP: rule_tester always supplies a filename */},

			// ---- Leading underscores preserved ----
			{Code: `// _camel`, FileName: "src/foo/_fooBar.js", Options: caseOpt("camelCase")},
			{Code: `// _camel`, FileName: "src/foo/___fooBar.js", Options: caseOpt("camelCase")},
			{Code: `// _snake`, FileName: "src/foo/_foo_bar.js", Options: caseOpt("snakeCase")},
			{Code: `// _snake`, FileName: "src/foo/___foo_bar.js", Options: caseOpt("snakeCase")},
			{Code: `// _kebab`, FileName: "src/foo/_foo-bar.js", Options: caseOpt("kebabCase")},
			{Code: `// _kebab`, FileName: "src/foo/___foo-bar.js", Options: caseOpt("kebabCase")},
			{Code: `// _pascal`, FileName: "src/foo/_FooBar.js", Options: caseOpt("pascalCase")},
			{Code: `// _pascal`, FileName: "src/foo/___FooBar.js", Options: caseOpt("pascalCase")},

			// ---- Default kebab + special chars at start ----
			{Code: `// default-kebab`, FileName: "src/foo/$foo.js"},

			// ---- `cases` option (omitted, empty, or filtered to truthy) falls
			// back to kebab when nothing remains.
			{Code: `// many-default`, FileName: "src/foo/foo-bar.js", Options: casesOpt(nil)},
			{Code: `// many-empty`, FileName: "src/foo/foo-bar.js", Options: casesOpt(map[string]bool{})},
			{Code: `// many-camel`, FileName: "src/foo/fooBar.js", Options: casesOpt(map[string]bool{"camelCase": true})},
			{Code: `// many-pascal+kebab`, FileName: "src/foo/FooBar.js", Options: casesOpt(map[string]bool{"kebabCase": true, "pascalCase": true})},
			{Code: `// many-snake+pascal+leading`, FileName: "src/foo/___foo_bar.js", Options: casesOpt(map[string]bool{"snakeCase": true, "pascalCase": true})},

			// ---- No options at all: default kebab ----
			{Code: `// no-options`, FileName: "src/foo/bar.js"},

			// ---- Decoration characters (brackets, braces) are preserved ----
			{Code: `// decoration`, FileName: "src/foo/[fooBar].js", Options: caseOpt("camelCase")},
			{Code: `// decoration`, FileName: "src/foo/{foo_bar}.js", Options: caseOpt("snakeCase")},

			// ---- Ignore patterns (string interpreted as ECMAScript regex) ----
			// Upstream covers each combination of (string-form, RegExp-form)
			// for every case style. JSON config can only carry strings, so we
			// migrate the string variants and skip the JS-only RegExp ones.
			{Code: `// undef-ignore`, Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`)},
			{Code: `// undef-ignore`, Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant — JSON-only here */},
			{Code: `// idx-via-ignore`, FileName: "src/foo/index.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`)},
			{Code: `// idx-via-ignore`, FileName: "src/foo/index.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`)},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("camelCase"), `FOOBAR\.js`)},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("camelCase"), `FOOBAR\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("snakeCase"), `FOOBAR\.js`)},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("pascalCase"), `FOOBAR\.js`)},
			{Code: `// ignored-by-regex`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("pascalCase"), `FOOBAR\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// multi-ignore`, FileName: "src/foo/BARBAZ.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`, `BARBAZ\.js`)},
			{Code: `// multi-ignore`, FileName: "src/foo/BARBAZ.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`, `BARBAZ\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// escaped-brackets`, FileName: "src/foo/[FOOBAR].js",
				Options: withIgnore(caseOpt("camelCase"), `\[FOOBAR\]\.js`)},
			{Code: `// escaped-brackets`, FileName: "src/foo/[FOOBAR].js",
				Options: withIgnore(caseOpt("camelCase"), `\[FOOBAR]\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// escaped-braces`, FileName: "src/foo/{FOOBAR}.js",
				Options: withIgnore(caseOpt("snakeCase"), `\{FOOBAR\}\.js`)},
			{Code: `// escaped-braces`, FileName: "src/foo/{FOOBAR}.js",
				Options: withIgnore(caseOpt("snakeCase"), `\{FOOBAR\}\.js`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// alternation`, FileName: "src/foo/foo.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`)},
			{Code: `// alternation`, FileName: "src/foo/foo-bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`)},
			{Code: `// alternation`, FileName: "src/foo/fooBar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`)},
			{Code: `// alternation`, FileName: "src/foo/foo_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`)},
			{Code: `// alternation`, FileName: "src/foo/foo_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`),
				Skip: true /* SKIP: ignore JS RegExp variant (case-insensitive flag) */},
			{Code: `// alternation`, FileName: "src/foo/FOO_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`),
				Skip: true /* SKIP: ignore JS RegExp variant (case-insensitive flag) */},
			{Code: `// suffix-ignore`, FileName: "src/foo/foo-bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `\.(web|android|ios)\.js$`)},
			{Code: `// suffix-ignore`, FileName: "src/foo/FooBar.web.js",
				Options: withIgnore(caseOpt("kebabCase"), `\.(web|android|ios)\.js$`)},
			{Code: `// suffix-ignore`, FileName: "src/foo/FooBar.android.js",
				Options: withIgnore(caseOpt("kebabCase"), `\.(web|android|ios)\.js$`)},
			{Code: `// suffix-ignore`, FileName: "src/foo/FooBar.ios.js",
				Options: withIgnore(caseOpt("kebabCase"), `\.(web|android|ios)\.js$`)},
			{Code: `// suffix-ignore`, FileName: "src/foo/FooBar.something.js",
				Options: withIgnore(caseOpt("kebabCase"), `\.(?:web|android|ios|something)\.js$`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// prefix-ignore`, FileName: "src/foo/FooBar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^(F|f)oo`)},
			{Code: `// prefix-ignore`, FileName: "src/foo/FooBar.js",
				Options: withIgnore(caseOpt("kebabCase"), `^[Ff]oo`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// 2-pattern-ignore`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("kebabCase"), `^FOO`, `BAZ\.js$`)},
			{Code: `// 2-pattern-ignore`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("kebabCase"), `^FOO`, `BAZ\.js$`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// 2-pattern-ignore`, FileName: "src/foo/BARBAZ.js",
				Options: withIgnore(caseOpt("kebabCase"), `^FOO`, `BAZ\.js$`)},
			{Code: `// 2-pattern-ignore`, FileName: "src/foo/BARBAZ.js",
				Options: withIgnore(caseOpt("kebabCase"), `^FOO`, `BAZ\.js$`),
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// many-ignore`, FileName: "src/foo/FOOBAR.js",
				Options: map[string]interface{}{
					"cases": map[string]interface{}{
						"kebabCase": true, "camelCase": true,
						"snakeCase": true, "pascalCase": true,
					},
					"ignore": []interface{}{`FOOBAR\.js`},
				}},
			{Code: `// many-ignore`, FileName: "src/foo/FOOBAR.js",
				Options: map[string]interface{}{
					"cases": map[string]interface{}{
						"kebabCase": true, "camelCase": true,
						"snakeCase": true, "pascalCase": true,
					},
					"ignore": []interface{}{`FOOBAR\.js`},
				},
				Skip: true /* SKIP: ignore JS RegExp variant */},
			{Code: `// many-ignore`, FileName: "src/foo/BaRbAz.js",
				Options: map[string]interface{}{
					"cases": map[string]interface{}{
						"kebabCase": true, "camelCase": true,
						"snakeCase": true, "pascalCase": true,
					},
					"ignore": []interface{}{`FOOBAR\.js`, `BaRbAz\.js`},
				}},
			{Code: `// many-ignore`, FileName: "src/foo/BaRbAz.js",
				Options: map[string]interface{}{
					"cases": map[string]interface{}{
						"kebabCase": true, "camelCase": true,
						"snakeCase": true, "pascalCase": true,
					},
					"ignore": []interface{}{`FOOBAR\.js`, `BaRbAz\.js`},
				},
				Skip: true /* SKIP: ignore JS RegExp variant */},

			// ---- Default-ignored index files (any case enabled) ----
			{Code: `// idx`, FileName: "index.js", Options: caseOpt("camelCase")},
			{Code: `// idx`, FileName: "index.mjs", Options: caseOpt("camelCase"),
				Skip: true /* SKIP: TS program does not pick up .mjs in this fixture */},
			{Code: `// idx`, FileName: "index.cjs", Options: caseOpt("camelCase"),
				Skip: true /* SKIP: TS program does not pick up .cjs in this fixture */},
			{Code: `// idx`, FileName: "index.ts", Options: caseOpt("camelCase")},
			{Code: `// idx`, FileName: "index.tsx", Options: caseOpt("camelCase")},
			{Code: `// idx`, FileName: "index.vue", Options: caseOpt("camelCase"),
				Skip: true /* SKIP: TS program does not pick up .vue */},
			{Code: `// idx`, FileName: "index.js", Options: caseOpt("snakeCase")},
			{Code: `// idx`, FileName: "index.mjs", Options: caseOpt("snakeCase"),
				Skip: true /* SKIP: TS program does not pick up .mjs */},
			{Code: `// idx`, FileName: "index.cjs", Options: caseOpt("snakeCase"),
				Skip: true /* SKIP: TS program does not pick up .cjs */},
			{Code: `// idx`, FileName: "index.ts", Options: caseOpt("snakeCase")},
			{Code: `// idx`, FileName: "index.tsx", Options: caseOpt("snakeCase")},
			{Code: `// idx`, FileName: "index.vue", Options: caseOpt("snakeCase"),
				Skip: true /* SKIP: TS program does not pick up .vue */},
			{Code: `// idx`, FileName: "index.js", Options: caseOpt("kebabCase")},
			{Code: `// idx`, FileName: "index.mjs", Options: caseOpt("kebabCase"),
				Skip: true /* SKIP: TS program does not pick up .mjs */},
			{Code: `// idx`, FileName: "index.cjs", Options: caseOpt("kebabCase"),
				Skip: true /* SKIP: TS program does not pick up .cjs */},
			{Code: `// idx`, FileName: "index.ts", Options: caseOpt("kebabCase")},
			{Code: `// idx`, FileName: "index.tsx", Options: caseOpt("kebabCase")},
			{Code: `// idx`, FileName: "index.vue", Options: caseOpt("kebabCase"),
				Skip: true /* SKIP: TS program does not pick up .vue */},
			{Code: `// idx`, FileName: "index.js", Options: caseOpt("pascalCase")},
			{Code: `// idx`, FileName: "index.mjs", Options: caseOpt("pascalCase"),
				Skip: true /* SKIP: TS program does not pick up .mjs */},
			{Code: `// idx`, FileName: "index.cjs", Options: caseOpt("pascalCase"),
				Skip: true /* SKIP: TS program does not pick up .cjs */},
			{Code: `// idx`, FileName: "index.ts", Options: caseOpt("pascalCase")},
			{Code: `// idx`, FileName: "index.tsx", Options: caseOpt("pascalCase")},
			{Code: `// idx`, FileName: "index.vue", Options: caseOpt("pascalCase"),
				Skip: true /* SKIP: TS program does not pick up .vue */},

			// ---- multipleFileExtensions=false ----
			{Code: `// mfe-false-idx-tsx`, FileName: "index.tsx", Options: withMfe(caseOpt("pascalCase"), false)},
			{Code: `// mfe-false-idx-tsx`, FileName: "src/index.tsx", Options: withMfe(caseOpt("pascalCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/fooBar.test.js", Options: withMfe(caseOpt("camelCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/fooBar.testUtils.js", Options: withMfe(caseOpt("camelCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/foo_bar.test_utils.js", Options: withMfe(caseOpt("snakeCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/foo.test.js", Options: withMfe(caseOpt("kebabCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/foo-bar.test.js", Options: withMfe(caseOpt("kebabCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/foo-bar.test-utils.js", Options: withMfe(caseOpt("kebabCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/Foo.Test.js", Options: withMfe(caseOpt("pascalCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/FooBar.Test.js", Options: withMfe(caseOpt("pascalCase"), false)},
			{Code: `// mfe-false`, FileName: "src/foo/FooBar.TestUtils.js", Options: withMfe(caseOpt("pascalCase"), false)},
			{Code: `// mfe-false`, FileName: "spec/Iss47.100Spec.js", Options: withMfe(caseOpt("pascalCase"), false)},

			// ---- Disable-comment valid case (covered by DisableManager) ----
			{
				Code:     "/* rslint-disable unicorn/filename-case */\nconst value = 1;",
				FileName: "src/foo/foo_bar.js",
				Options:  caseOpt("kebabCase"),
				Skip:     true, /* SKIP: rule_tester does not exercise rule_tester-level disable comments */
			},

			// ---- Multipart filename / multipleFileExtensions=true ----
			{Code: `// camel-multi`, FileName: "src/foo/fooBar.Test.js", Options: caseOpt("camelCase")},
			{Code: `// camel-multi`, FileName: "test/foo/fooBar.testUtils.js", Options: caseOpt("camelCase")},
			{Code: `// camel-multi`, FileName: "test/foo/.testUtils.js", Options: caseOpt("camelCase"),
				Skip: true /* SKIP: dot-prefixed basename */},
			{Code: `// snake-multi`, FileName: "test/foo/foo_bar.Test.js", Options: caseOpt("snakeCase")},
			{Code: `// snake-multi`, FileName: "test/foo/foo_bar.Test_Utils.js", Options: caseOpt("snakeCase")},
			{Code: `// snake-multi`, FileName: "test/foo/.Test_Utils.js", Options: caseOpt("snakeCase"),
				Skip: true /* SKIP: dot-prefixed basename */},
			{Code: `// kebab-multi`, FileName: "test/foo/foo-bar.Test.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab-multi`, FileName: "test/foo/foo-bar.Test-Utils.js", Options: caseOpt("kebabCase")},
			{Code: `// kebab-multi`, FileName: "test/foo/.Test-Utils.js", Options: caseOpt("kebabCase"),
				Skip: true /* SKIP: dot-prefixed basename */},
			{Code: `// pascal-multi`, FileName: "test/foo/FooBar.Test.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal-multi`, FileName: "test/foo/FooBar.TestUtils.js", Options: caseOpt("pascalCase")},
			{Code: `// pascal-multi`, FileName: "test/foo/.TestUtils.js", Options: caseOpt("pascalCase"),
				Skip: true /* SKIP: dot-prefixed basename */},

			// ---- Snapshot-block valid cases (basename-only handling) ----
			{Code: `// snap-undef`}, /* upstream `undefined` filename — defaults to `file.ts`, valid for kebab */
			{Code: `// snap-dir-uppercase-ext`, FileName: "src/foo.JS/bar.js"},
			{Code: `// snap-dir-uppercase-ext`, FileName: "src/foo.JS/bar.spec.js"},
			{Code: `// snap-dir-uppercase-ext`, FileName: "src/foo.JS/.spec.js",
				Skip: true /* SKIP: dot-prefixed basename */},
			{Code: `// snap-dir-no-ext`, FileName: "src/foo.JS/bar",
				Skip: true /* SKIP: TS program does not pick up extensionless files */},
			{Code: `// snap-dotted-uppercase-middle`, FileName: "foo.SPEC.js"},
			{Code: `// snap-leading-dot`, FileName: ".SPEC.js",
				Skip: true /* SKIP: dot-prefixed basename */},

			// ---- Lock-in tests for upstream branches not directly exercised
			//      by the upstream test file ----
			// Locks in change-case digit-prefixed non-first word: `iss-47-spec`
			// for camelCase splits to `['iss', '47', 'spec']` and reassembles
			// as `iss_47Spec` (the `_<digit>` branch of pascalCaseTransform).
			// Upstream covers this for kebab/snake (which lowercase-join), but
			// not for camel/pascal — we add an invalid case below to lock the
			// behaviour. Here we lock the equivalent already-camel form.
			{Code: `// lock-in: digit-prefixed second word`, FileName: "src/foo/iss_47Spec.js", Options: caseOpt("camelCase")},

			// Locks in: non-string `ignore` entries don't poison the array —
			// the valid string pattern still ignores its target. Companion
			// to the invalid-list case above; together they prove non-string
			// items are dropped silently and the rest of the array still
			// applies normally.
			{
				Code: `// lock-in: non-string ignore entries silently dropped, valid sibling still ignores`,
				FileName: "src/foo/FOOBAR.js",
				Options: map[string]interface{}{
					"case":   "kebabCase",
					"ignore": []interface{}{nil, 42, map[string]interface{}{}, `FOOBAR\.js`},
				},
			},
			// Locks in: an empty `ignore` array works the same as omitting
			// `ignore` — no diagnostics, no spurious empty-pattern matches.
			{
				Code: `// lock-in: empty ignore array`,
				FileName: "src/foo/foo-bar.js",
				Options: map[string]interface{}{"case": "kebabCase", "ignore": []interface{}{}},
			},

			// Locks in `splitWords` Pass 2 multi-fire end-to-end: an
			// XMLHttp-style basename in pascalCase splits into ALL-CAPS +
			// Title chunks (`XML/Http/Request`) and pascal-reassembles
			// each word with non-first letters lowered (`Xml/Http/
			// Request` → `XmlHttpRequest`). This proves Pass 2 fired AND
			// pascalCase honoured the per-word lowercasing — both could
			// regress independently. (See moved-to-invalid block below.)

			// (fixFilename dedupe is already locked in by `1_.js` invalid
			// test above, where camel/pascal/kebab all collapse to `1`.)

			// Locks in multi-middle-part basename under
			// multipleFileExtensions=true: `foo.bar.baz.js` splits as
			// filename=`foo`, middle=`.bar.baz`. The rule only validates
			// `foo`, leaving `.bar.baz` verbatim in the rename suggestion.
			// (Snake-case here so the rename actually fires.)
			{Code: `// lock-in: multi-middle parts kept verbatim (valid)`, FileName: "src/foo/foo.bar.baz.js", Options: caseOpt("snakeCase")},

			// Locks in: a basename whose every word is ignored (only
			// punctuation, no letters/digits/`-`/`_` for `validateFilename`
			// to inspect) passes any case style. `validateFilename` returns
			// true on an empty filtered list.
			{Code: `// lock-in: all-ignored basename is valid`, FileName: "src/foo/$$$.js", Options: caseOpt("camelCase")},
			{Code: `// lock-in: all-ignored basename + leading underscores is valid`, FileName: "src/foo/___$$.js", Options: caseOpt("camelCase")},

			// Locks in: when `case` and `cases` are BOTH set, `case` wins
			// (matches the if/else-if order in parseOptions). Upstream
			// schema's `oneOf` would reject this config, but a JSON config
			// can carry both — pinning the precedence keeps it deterministic.
			{
				Code: `// lock-in: case takes precedence over cases when both are set`,
				FileName: "src/foo/fooBar.js",
				Options: map[string]interface{}{
					"case":  "camelCase",
					"cases": map[string]interface{}{"snakeCase": true},
				},
			},

			// Locks in: an unknown `case` value (e.g. `"camelcase"` lowercase
			// typo, or any non-enum value) is silently ignored — the rule
			// falls back to its default (kebab). The valid case below pins
			// "fall through to kebab default", and the invalid case below
			// proves the default fires when the basename violates kebab.
			{Code: `// lock-in: unknown case value falls back to kebab default (valid)`, FileName: "src/foo/foo-bar.js", Options: caseOpt("camelcase" /* typo */)},

		},
		[]rule_tester.InvalidTestCase{
			// ---- Disable-comment INSIDE the file body — does NOT match a
			// rule disable directive (note `//` not at top), so reports.
			{
				Code:     "// eslint-disable rule-to-test/filename-case\nconst value = 1;",
				FileName: "src/foo/foo_bar.js",
				Options:  caseOpt("kebabCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},

			// ---- Default kebab on snake-cased filename ----
			{
				Code:     `// k`,
				FileName: "src/foo/foo_bar.js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- camelCase failures ----
			{
				Code: `// c`, FileName: "src/foo/foo_bar.JS", Options: caseOpt("camelCase"),
				Skip: true, /* SKIP: TS program rejects `.JS`; covered in JS test set */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `fooBar.js`.",
				}},
			},
			{
				Code: `// c`, FileName: "src/foo/foo_bar.test.js", Options: caseOpt("camelCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `fooBar.test.js`.",
				}},
			},
			{
				Code: `// c`, FileName: "test/foo/foo_bar.test_utils.js", Options: caseOpt("camelCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `fooBar.test_utils.js`.",
				}},
			},

			// ---- snakeCase failures ----
			{
				Code: `// s`, FileName: "test/foo/fooBar.js", Options: caseOpt("snakeCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `foo_bar.js`.",
				}},
			},
			{
				Code: `// s`, FileName: "test/foo/fooBar.test.js", Options: caseOpt("snakeCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `foo_bar.test.js`.",
				}},
			},
			{
				Code: `// s`, FileName: "test/foo/fooBar.testUtils.js", Options: caseOpt("snakeCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `foo_bar.testUtils.js`.",
				}},
			},

			// ---- kebabCase failures ----
			{
				Code: `// k`, FileName: "test/foo/fooBar.js", Options: caseOpt("kebabCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},
			{
				Code: `// k`, FileName: "test/foo/fooBar.test.js", Options: caseOpt("kebabCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.test.js`.",
				}},
			},
			{
				Code: `// k`, FileName: "test/foo/fooBar.testUtils.js", Options: caseOpt("kebabCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.testUtils.js`.",
				}},
			},

			// ---- pascalCase failures ----
			{
				Code: `// p`, FileName: "test/foo/fooBar.js", Options: caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `FooBar.js`.",
				}},
			},
			{
				Code: `// p`, FileName: "test/foo/foo_bar.test.js", Options: caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `FooBar.test.js`.",
				}},
			},
			{
				Code: `// p`, FileName: "test/foo/foo-bar.test-utils.js", Options: caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `FooBar.test-utils.js`.",
				}},
			},

			// ---- Leading underscores preserved verbatim in suggestions ----
			{
				Code: `// _c`, FileName: "src/foo/_FOO-BAR.js", Options: caseOpt("camelCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `_fooBar.js`.",
				}},
			},
			{
				Code: `// _c`, FileName: "src/foo/___FOO-BAR.js", Options: caseOpt("camelCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `___fooBar.js`.",
				}},
			},
			{
				Code: `// _s`, FileName: "src/foo/_FOO-BAR.js", Options: caseOpt("snakeCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `_foo_bar.js`.",
				}},
			},
			{
				Code: `// _s`, FileName: "src/foo/___FOO-BAR.js", Options: caseOpt("snakeCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `___foo_bar.js`.",
				}},
			},
			{
				Code: `// _k`, FileName: "src/foo/_FOO-BAR.js", Options: caseOpt("kebabCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `_foo-bar.js`.",
				}},
			},
			{
				Code: `// _k`, FileName: "src/foo/___FOO-BAR.js", Options: caseOpt("kebabCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `___foo-bar.js`.",
				}},
			},
			{
				Code: `// _p`, FileName: "src/foo/_FOO-BAR.js", Options: caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `_FooBar.js`.",
				}},
			},
			{
				Code: `// _p`, FileName: "src/foo/___FOO-BAR.js", Options: caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `___FooBar.js`.",
				}},
			},

			// ---- `cases` option failures (canonical case order in our output) ----
			{
				Code: `// many-default-kebab`, FileName: "src/foo/foo_bar.js", Options: casesOpt(nil),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},
			{
				Code: `// many-c+p`, FileName: "src/foo/foo-bar.js",
				Options: casesOpt(map[string]bool{"camelCase": true, "pascalCase": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or pascal case. Rename it to `fooBar.js` or `FooBar.js`.",
				}},
			},
			{
				Code: `// many-c+p+k`, FileName: "src/foo/_foo_bar.js",
				Options: casesOpt(map[string]bool{"camelCase": true, "pascalCase": true, "kebabCase": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case, kebab case, or pascal case. Rename it to `_fooBar.js`, `_foo-bar.js`, or `_FooBar.js`.",
				}},
			},
			{
				Code: `// many-snake`, FileName: "src/foo/_FOO-BAR.js",
				Options: casesOpt(map[string]bool{"snakeCase": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `_foo_bar.js`.",
				}},
			},

			// ---- Decoration characters preserved verbatim ----
			{
				Code: `// dec`, FileName: "src/foo/[foo_bar].js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `[foo-bar].js`.",
				}},
			},
			{
				Code: `// dec`, FileName: "src/foo/$foo_bar.js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `$foo-bar.js`.",
				}},
			},
			{
				Code: `// dec`, FileName: "src/foo/$fooBar.js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `$foo-bar.js`.",
				}},
			},
			{
				Code: `// dec-many`, FileName: "src/foo/{foo_bar}.js",
				Options: casesOpt(map[string]bool{"camelCase": true, "pascalCase": true, "kebabCase": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case, kebab case, or pascal case. Rename it to `{fooBar}.js`, `{foo-bar}.js`, or `{FooBar}.js`.",
				}},
			},

			// ---- Ignore patterns that don't match still report ----
			{
				Code: `// ignore-miss`, FileName: "src/foo/barBaz.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `bar-baz.js`.",
				}},
			},
			// Locks in upstream `new RegExp(item, 'u')`: a leading-`/` and trailing-`/`
			// pattern is treated as a literal regex *body* (with the slashes in it),
			// so the pattern still doesn't match `barBaz.js`.
			{
				Code: `// ignore-miss-slashes`, FileName: "src/foo/barBaz.js",
				Options: withIgnore(caseOpt("kebabCase"), `/FOOBAR\.js/`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `bar-baz.js`.",
				}},
			},
			{
				Code: `// ignore-miss`, FileName: "src/foo/fooBar.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},
			{
				Code: `// ignore-miss-multi`, FileName: "src/foo/fooBar.js",
				Options: withIgnore(caseOpt("kebabCase"), `FOOBAR\.js`, `foobar\.js`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},
			// Two cases — canonical order in our output: camelCase, snakeCase.
			{
				Code: `// many-cs+ignore-miss`, FileName: "src/foo/FooBar.js",
				Options: map[string]interface{}{
					"cases":  map[string]interface{}{"camelCase": true, "snakeCase": true},
					"ignore": []interface{}{`FOOBAR\.js`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or snake case. Rename it to `fooBar.js` or `foo_bar.js`.",
				}},
			},
			{
				Code: `// many-cs+ignore-miss-pat2`, FileName: "src/foo/FooBar.js",
				Options: map[string]interface{}{
					"cases":  map[string]interface{}{"camelCase": true, "snakeCase": true},
					"ignore": []interface{}{`BaRbAz\.js`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or snake case. Rename it to `fooBar.js` or `foo_bar.js`.",
				}},
			},
			{
				Code: `// many-cs+ignore-prefix`, FileName: "src/foo/FooBar.js",
				Options: map[string]interface{}{
					"cases":  map[string]interface{}{"camelCase": true, "snakeCase": true},
					"ignore": []interface{}{`^foo`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or snake case. Rename it to `fooBar.js` or `foo_bar.js`.",
				}},
			},
			{
				Code: `// many-cs+ignore-2patterns`, FileName: "src/foo/FooBar.js",
				Options: map[string]interface{}{
					"cases":  map[string]interface{}{"camelCase": true, "snakeCase": true},
					"ignore": []interface{}{`^foo`, `^bar`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or snake case. Rename it to `fooBar.js` or `foo_bar.js`.",
				}},
			},

			// ---- #1136: trailing underscore on a digit-only word ----
			{
				Code: `// 1_-multi`, FileName: "src/foo/1_.js",
				Options: casesOpt(map[string]bool{"camelCase": true, "pascalCase": true, "kebabCase": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case, kebab case, or pascal case. Rename it to `1.js`.",
				}},
			},

			// ---- multipleFileExtensions=false: middle parts must also match ----
			{
				Code: `// mfe-false`, FileName: "src/foo/foo_bar.test.js", Options: withMfe(caseOpt("camelCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `fooBar.test.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/foo_bar.test_utils.js", Options: withMfe(caseOpt("camelCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `fooBar.testUtils.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/fooBar.test.js", Options: withMfe(caseOpt("snakeCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `foo_bar.test.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/fooBar.testUtils.js", Options: withMfe(caseOpt("snakeCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in snake case. Rename it to `foo_bar.test_utils.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/fooBar.test.js", Options: withMfe(caseOpt("kebabCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.test.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/fooBar.testUtils.js", Options: withMfe(caseOpt("kebabCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.test-utils.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/foo_bar.test.js", Options: withMfe(caseOpt("pascalCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `FooBar.Test.js`.",
				}},
			},
			{
				Code: `// mfe-false`, FileName: "test/foo/foo-bar.test-utils.js", Options: withMfe(caseOpt("pascalCase"), false),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `FooBar.TestUtils.js`.",
				}},
			},

			// ---- Snapshot block: extension-only and `.mJS` failures ----
			{
				Code: "foo();\nfoo();\nfoo();", FileName: "src/foo/foo_bar.mJS",
				Options: casesOpt(map[string]bool{"camelCase": true, "kebabCase": true}),
				Skip:    true, /* SKIP: TS program rejects `.mJS`; covered in JS tests */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or kebab case. Rename it to `fooBar.mjs` or `foo-bar.mjs`.",
				}},
			},
			{
				Code: `// snap-foo.JS`, FileName: "foo.JS",
				Skip: true, /* SKIP: TS program rejects `.JS` */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameExtension",
					Message:   "File extension `.JS` is not in lowercase. Rename it to `foo.js`.",
				}},
			},
			{
				Code: `// snap-foo.Js`, FileName: "foo.Js",
				Skip: true, /* SKIP: TS program rejects `.Js` */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameExtension",
					Message:   "File extension `.Js` is not in lowercase. Rename it to `foo.js`.",
				}},
			},
			{
				Code: `// snap-foo.jS`, FileName: "foo.jS",
				Skip: true, /* SKIP: TS program rejects `.jS` */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameExtension",
					Message:   "File extension `.jS` is not in lowercase. Rename it to `foo.js`.",
				}},
			},
			{
				Code: `// snap-index.JS`, FileName: "index.JS",
				Skip: true, /* SKIP: TS program rejects `.JS` */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameExtension",
					Message:   "File extension `.JS` is not in lowercase. Rename it to `index.js`.",
				}},
			},
			{
				Code: `// snap-foo..JS`, FileName: "foo..JS",
				Skip: true, /* SKIP: TS program rejects `.JS` */
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameExtension",
					Message:   "File extension `.JS` is not in lowercase. Rename it to `foo..js`.",
				}},
			},

			// ---- Lock-in tests for upstream branches not directly exercised
			//      by the upstream test file ----
			// Locks in change-case `pascalCaseTransform` digit-prefix branch:
			// `iss-47-spec` for camelCase produces `iss_47Spec` (the `_` is
			// inserted before a digit-starting non-first word). Upstream tests
			// the kebab/snake side but not camel/pascal.
			{
				Code: `// lock-in: digit-prefixed second word (camel)`, FileName: "src/foo/iss-47-spec.js",
				Options: caseOpt("camelCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `iss_47Spec.js`.",
				}},
			},
			{
				Code: `// lock-in: digit-prefixed second word (pascal)`, FileName: "src/foo/iss-47-spec.js",
				Options: caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `Iss_47Spec.js`.",
				}},
			},
			// Locks in canonical case-name ordering: the user wrote keys in a
			// different order, but our message must be camel/snake/kebab/pascal.
			{
				Code: `// lock-in: canonical case-name ordering`, FileName: "src/foo/foo-bar.js",
				Options: casesOpt(map[string]bool{"pascalCase": true, "camelCase": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case or pascal case. Rename it to `fooBar.js` or `FooBar.js`.",
				}},
			},
			// Locks in malformed `ignore` pattern reporting: a broken pattern
			// short-circuits the rule on that file. Only the configuration
			// error is reported — case-violation reports do NOT follow,
			// because they would be based on a partially-broken ignore list.
			{
				Code: `// lock-in: malformed ignore pattern`, FileName: "src/foo/foo_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `[unclosed`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `[unclosed`: error parsing regexp: unterminated [] set in `[unclosed`",
				}},
			},
			// Locks in: even when the file's basename would otherwise have
			// matched a valid sibling pattern, a single malformed pattern in
			// the same `ignore` array still aborts case-checking. The user
			// gets a clear "fix your config" diagnostic instead of a
			// silently-correct or silently-wrong case report.
			{
				Code: `// lock-in: invalid ignore pattern wins over valid sibling`, FileName: "src/foo/FOOBAR.js",
				Options: withIgnore(caseOpt("kebabCase"), `[unclosed`, `FOOBAR\.js`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `[unclosed`: error parsing regexp: unterminated [] set in `[unclosed`",
				}},
			},
			// Locks in: multiple invalid patterns produce one diagnostic per
			// pattern (so the user sees every broken entry at once), but
			// still no case report.
			{
				Code: `// lock-in: two invalid ignore patterns`, FileName: "src/foo/foo_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `[unclosed`, `(unclosed`),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidIgnorePattern",
						Message:   "Invalid regular expression in `ignore` option: `[unclosed`: error parsing regexp: unterminated [] set in `[unclosed`",
					},
					{
						MessageId: "invalidIgnorePattern",
						Message:   "Invalid regular expression in `ignore` option: `(unclosed`: error parsing regexp: missing closing ) in `(unclosed`",
					},
				},
			},
			// Locks in: a malformed ignore pattern fires the configuration
			// diagnostic even when the basename would have passed case
			// validation (i.e. NO `filenameCase` would have been reported
			// either way). This proves the short-circuit isn't masking a
			// case error — it's a deliberate "fix your config first" gate
			// even on otherwise-clean files.
			{
				Code: `// lock-in: invalid ignore on a case-clean basename`, FileName: "src/foo/foo-bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `[unclosed`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `[unclosed`: error parsing regexp: unterminated [] set in `[unclosed`",
				}},
			},
			// Locks in: a malformed ignore pattern still fires on a file
			// the rule normally short-circuits via `ignoredByDefault`
			// (e.g. `index.js`). The configuration error takes priority
			// over the default-ignore set, so users always learn about
			// broken patterns even on conventional filenames.
			{
				Code: `// lock-in: invalid ignore on a default-ignored basename`, FileName: "index.js",
				Options: withIgnore(caseOpt("pascalCase"), `[unclosed`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `[unclosed`: error parsing regexp: unterminated [] set in `[unclosed`",
				}},
			},
			// Locks in: invalid ignore + `cases` (multi-style) option ─
			// short-circuit fires regardless of which option path filled
			// `Cases`. Catches future regressions where `cases`-branch
			// parsing forgets the invalid-ignore gate.
			{
				Code: `// lock-in: invalid ignore under cases option`, FileName: "src/foo/FooBar.js",
				Options: map[string]interface{}{
					"cases":  map[string]interface{}{"camelCase": true, "snakeCase": true},
					"ignore": []interface{}{`(unclosed`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `(unclosed`: error parsing regexp: missing closing ) in `(unclosed`",
				}},
			},
			// Locks in: dangling-backslash pattern — different error class
			// from `[unclosed`, exercises a separate regex-engine code path.
			{
				Code: `// lock-in: dangling-backslash ignore`, FileName: "src/foo/foo_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `foo\`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `foo\\`: error parsing regexp: illegal \\ at end of pattern in `foo\\`",
				}},
			},
			// Locks in: bare-quantifier pattern — yet another distinct
			// error class (missing argument to repetition).
			{
				Code: `// lock-in: bare-quantifier ignore`, FileName: "src/foo/foo_bar.js",
				Options: withIgnore(caseOpt("kebabCase"), `*invalid`),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `*invalid`: error parsing regexp: missing argument to repetition operator in `*invalid`",
				}},
			},
			// Locks in: non-string entries (`null`, numbers, objects) in the
			// `ignore` array are silently skipped — they are NOT treated as
			// malformed patterns. Without this, a JSON-stringified RegExp
			// object (which lands as `{}` on the Go side) or any other
			// stray non-string would fire spurious `invalidIgnorePattern`
			// diagnostics. Here the only valid string pattern doesn't match
			// `foo_bar.js`, so we get a normal case-violation report — the
			// crucial assertion is that we do NOT see any
			// `invalidIgnorePattern` diagnostic alongside it.
			{
				Code: `// lock-in: non-string ignore entries do not fire invalidIgnorePattern`,
				FileName: "src/foo/foo_bar.js",
				Options: map[string]interface{}{
					"case":   "kebabCase",
					"ignore": []interface{}{nil, 42, map[string]interface{}{}, `FOOBAR\.js`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},
			// Locks in: when non-string + valid string + invalid string ignore
			// entries co-exist, the invalid string still wins (fatal short-
			// circuit), the valid string never gets to apply, and the
			// non-string is silently dropped. End-to-end coverage of all
			// three ignore-entry classes interacting.
			{
				Code: `// lock-in: non-string + valid + invalid ignore entries together`,
				FileName: "src/foo/FOOBAR.js",
				Options: map[string]interface{}{
					"case":   "kebabCase",
					"ignore": []interface{}{nil, `FOOBAR\.js`, `[unclosed`},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidIgnorePattern",
					Message:   "Invalid regular expression in `ignore` option: `[unclosed`: error parsing regexp: unterminated [] set in `[unclosed`",
				}},
			},
			// Locks in `englishishJoin` 4-item oxford comma + `or` end-to-end:
			// all four `cases` enabled + a basename violating all four of
			// them yields `camel case, snake case, kebab case, or pascal
			// case` and four rename suggestions in canonical order. (Unit
			// test in splitwords_test.go locks the formatter directly; this
			// proves the rule produces it via the real diagnostic path.)
			{
				Code: `// lock-in: oxford-comma 4-item, all four cases enabled`,
				FileName: "src/foo/FOO_BAR.js",
				Options: casesOpt(map[string]bool{
					"camelCase": true, "snakeCase": true,
					"kebabCase": true, "pascalCase": true,
				}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case, snake case, kebab case, or pascal case. Rename it to `fooBar.js`, `foo_bar.js`, `foo-bar.js`, or `FooBar.js`.",
				}},
			},
			// Locks in `pascalLikeTransform` first-word-digit branch end-to-
			// end: `123-foo` for camelCase produces `123Foo` (NO leading
			// `_`, because the digit-prefixed word is index 0). Companion
			// invalid for the unit test in splitwords_test.go.
			{
				Code: `// lock-in: first word starting with digit (camel)`,
				FileName: "src/foo/123-foo.js",
				Options:  caseOpt("camelCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `123Foo.js`.",
				}},
			},
			// Locks in `pascalLikeTransform` first-word-digit branch for
			// pascalCase too — the first-word digit also stays unprefixed
			// in pascal output (`upper(char0)` is identity for digits).
			{
				Code: `// lock-in: first word starting with digit (pascal)`,
				FileName: "src/foo/123-foo.js",
				Options:  caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `123Foo.js`.",
				}},
			},
			// Locks in: an unknown / mistyped `case` value (e.g. lowercase
			// `"camelcase"`, or any non-enum value) is silently ignored,
			// the rule falls back to the default kebab case. Companion to
			// the valid case above (where the basename is already kebab);
			// here the basename violates kebab and we prove the default
			// case actually fires.
			{
				Code: `// lock-in: unknown case value falls back to kebab default (invalid)`,
				FileName: "src/foo/fooBar.js",
				Options:  caseOpt("camelcase" /* typo */),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
				}},
			},
			// Locks in `case` + `cases` precedence end-to-end: when both
			// are present and `case` is honoured, the basename validates
			// against `case` only — even if it would have failed `cases`.
			// `fooBar.js` is valid camel; `cases: { snakeCase: true }`
			// alone would have flagged it. Here we confirm the rule
			// honours `case` and reports nothing.
			{
				Code: `// lock-in: case+cases — case wins (filename violates cases-only set)`,
				FileName: "src/foo/foo_bar.js",
				Options: map[string]interface{}{
					"case":  "snakeCase",
					"cases": map[string]interface{}{"camelCase": true},
				},
				// Filename is valid snake → no diagnostic. We cover this
				// in the valid block above; here the invalid mirror is:
				// case=camel (basename `foo_bar.js`) wins, file violates
				// camel → reports.
				Skip: true, /* SKIP: redundant with the valid-block companion */
				Errors: []rule_tester.InvalidTestCaseError{{}},
			},
			{
				Code: `// lock-in: case+cases — case wins, basename violates the chosen case`,
				FileName: "src/foo/foo_bar.js",
				Options: map[string]interface{}{
					"case":  "camelCase",
					"cases": map[string]interface{}{"snakeCase": true},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in camel case. Rename it to `fooBar.js`.",
				}},
			},
			// Locks in `splitWords` Pass 2 multi-fire end-to-end: pascalCase
			// applied to `XMLHttpRequest` splits into 3 words and rejoins
			// with per-word lowercasing → `XmlHttpRequest`. Proves both
			// the ALL-CAPS+Title boundary cut AND the non-first-letter
			// lowering pipeline.
			{
				Code: `// lock-in: Pass 2 multi-fire (pascal lowers non-first letters)`,
				FileName: "src/foo/XMLHttpRequest.js",
				Options:  caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `XmlHttpRequest.js`.",
				}},
			},
			{
				Code: `// lock-in: Pass 2 single-fire (pascal lowers non-first letters)`,
				FileName: "src/foo/HTTPSConnection.js",
				Options:  caseOpt("pascalCase"),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "filenameCase",
					Message:   "Filename is not in pascal case. Rename it to `HttpsConnection.js`.",
				}},
			},
			// Locks in Node `path.extname` parity for an all-dots basename.
			// `....js` ends with `.js`; my earlier (buggy) extname returned
			// `.js` and the validator would see filename `...` (all-ignored,
			// empty word list, valid). Node returns `.js` here too — `...js`
			// has a non-dot before the last dot. This is the boundary case
			// that proves we still match Node when the all-dots prefix is
			// shorter than the basename.
			{
				Code: `// lock-in: node-extname parity (basename has trailing real ext)`,
				FileName: "src/foo/...js",
				Skip:     true, /* SKIP: TS program rejects this odd basename; logic exercised by the unit table inline above */
			},
		},
	)
}
