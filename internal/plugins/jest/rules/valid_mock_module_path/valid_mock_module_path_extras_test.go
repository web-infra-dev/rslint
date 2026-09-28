// cspell:ignore jackspeak doesnotexist
package valid_mock_module_path

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidMockModulePathExtras(t *testing.T) {
	root := mockModuleRoot(t)
	fixturesFile := "src/rules/__tests__/fixtures/nested.test.ts"
	rootFile := "root.test.ts"
	jsFile := "src/rules/__tests__/valid-mock-module-path.test.js"

	rule_tester.RunRuleTester(
		root,
		"tsconfig.json",
		t,
		&ValidMockModulePathRule,
		[]rule_tester.ValidTestCase{
			// ---- Local paths: any file or directory at the joined path ----
			{Code: `jest.mock("./fixtures/module/")`, FileName: testFile},
			{Code: `jest.mock(".")`, FileName: testFile},
			{Code: `jest.mock("..")`, FileName: testFile},
			{Code: `jest.mock("./fixtures/../fixtures/module/bar.css")`, FileName: testFile},
			// Any specifier starting with `.` is a local path, including a dotfile.
			{Code: `jest.mock(".hidden")`, FileName: fixturesFile},
			// Extensions are appended verbatim, without adding a dot.
			{
				Code:     `jest.mock("./fixtures/module/Data.")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{"custom"}},
			},
			// An empty list still accepts the exact path.
			{
				Code:     `jest.mock("./fixtures/module/foo.ts")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{}},
			},

			// ---- Packages, resolved like require.resolve from the linted file ----
			{Code: `jest.mock("legacy-main")`, FileName: testFile},
			{Code: `jest.mock("legacy-main/lib/deep/file.json")`, FileName: testFile},
			{Code: `jest.mock("legacy-main/lib/deep/file")`, FileName: testFile},
			{Code: `jest.mock("index-only")`, FileName: testFile},
			{Code: `jest.mock("@scope/pkg")`, FileName: testFile},
			{Code: `jest.mock("ts-source-only")`, FileName: testFile},
			{Code: `jest.mock("eslint/package.json")`, FileName: testFile},
			{Code: `jest.mock("jackspeak")`, FileName: testFile},
			{Code: `jest.mock("nested-only")`, FileName: testFile},
			{Code: `jest.mock("/valid-mock-module-path-fixtures/node_modules/legacy-main/lib/entry")`, FileName: testFile},
			{Code: `jest.mock("node:fs")`, FileName: testFile},
			{Code: `jest.mock("fs/promises")`, FileName: testFile},
			{Code: `jest.mock("node:test")`, FileName: testFile},
			// Subpath imports come from the package that owns the linted file.
			{Code: `jest.mock("#fixture-module")`, FileName: testFile},
			// A fallback array skips every entry that is not a valid target, as Node does.
			{Code: `jest.mock("x-null-fallback")`, FileName: testFile},
			{Code: `jest.mock("x-false-fallback")`, FileName: testFile},
			{Code: `jest.mock("x-nested-null")`, FileName: testFile},
			{Code: `jest.mock("x-invalid-fallback")`, FileName: testFile},
			{Code: `jest.mock("x-condition-fallback")`, FileName: testFile},
			{Code: `jest.mock("x-require-null")`, FileName: testFile},
			{Code: `jest.mock("x-nested-array")`, FileName: testFile},
			{Code: `jest.mock("x-pattern/sub/x.js")`, FileName: testFile},
			{Code: `jest.mock("x-pattern-ext/ext/x")`, FileName: testFile},

			// ---- Only a string literal names the module ----
			{Code: "jest.mock(`./missing`)", FileName: testFile},
			{Code: `jest.mock(...["./missing"])`, FileName: testFile},
			{Code: `jest.mock("./missing" as string)`, FileName: testFile},

			// ---- Only jest.mock and jest.doMock are checked ----
			{Code: `jest.unmock("./missing")`, FileName: testFile},
			{Code: `jest.requireActual("./missing")`, FileName: testFile},
			{Code: `jest.dontMock("./missing")`, FileName: testFile},
			{Code: `foo.mock("./missing")`, FileName: testFile},
			{Code: `mock("./missing")`, FileName: testFile},
			{Code: `jest[method]("./missing")`, FileName: testFile},
			{Code: `(jest.mock as any)("./missing")`, FileName: testFile},
			{Code: `jest.mock!("./missing")`, FileName: testFile},
			{Code: "const jest = { mock() {} };\njest.mock('./missing');", FileName: testFile},
			// A call that is itself called, passed as an argument or accessed is
			// not parsed as a Jest call, so only the outermost call of a chain is checked.
			{Code: `jest.mock("./missing").mock("./fixtures/module")`, FileName: testFile},
			{Code: `jest.mock("./missing")?.mock("./fixtures/module")`, FileName: testFile},
			{Code: `(jest.mock("./missing")).foo`, FileName: testFile},
			{Code: `foo(jest.mock("./missing"))`, FileName: testFile},
			{Code: `a[jest.mock("./missing")]`, FileName: testFile},

			// ---- virtual ----
			// Like upstream, a computed identifier key counts by its name.
			{Code: `jest.mock("./missing", undefined, { [virtual]: true })`, FileName: testFile},
			{Code: `jest.mock("./missing", undefined, { virtual: false, virtual: true })`, FileName: testFile},
			{Code: `jest.mock("./missing", undefined, { virtual: 1n })`, FileName: testFile},
			{Code: `jest.mock("./missing", undefined, { virtual: 0x1 })`, FileName: testFile},
			{Code: `jest.mock("./missing", undefined, { virtual: /re/ })`, FileName: testFile},
			{Code: `jest.mock("./missing", undefined, { virtual: (true) })`, FileName: testFile},
			{Code: `jest.mock("./missing", undefined, ({ virtual: true }))`, FileName: testFile},
			{Code: `jest.mock("missing-package", () => ({}), { virtual: true })`, FileName: testFile},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Local paths ----
			{Code: `jest.mock(".missing")`, FileName: fixturesFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./fixtures/module/Data.")`, FileName: testFile, Errors: invalidAt(1)},
			// Directories are not searched for an index file with an extension.
			{
				Code:     `jest.mock("./fixtures/module/foo")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{}},
				Errors:   invalidAt(1),
			},
			// Extensions outside the list are not tried.
			{
				Code:     `jest.mock("./fixtures/module/bar")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{".js"}},
				Errors:   invalidAt(1),
			},

			// ---- Packages ----
			{Code: `jest.mock("legacy-main/lib/deep/missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("@scope/missing")`, FileName: testFile, Errors: invalidAt(1)},
			// No `require` or `default` export condition.
			{Code: `jest.mock("esm-only")`, FileName: testFile, Errors: invalidAt(1)},
			// A declaration file is not a runtime module.
			{Code: `jest.mock("types-only")`, FileName: testFile, Errors: invalidAt(1)},
			// The subpath exports only a `types` condition.
			{Code: `jest.mock("eslint/rules")`, FileName: testFile, Errors: invalidAt(1)},
			// Lookup starts from the linted file's directory.
			{Code: `jest.mock("nested-only")`, FileName: rootFile, Errors: invalidAt(1)},
			{Code: `jest.mock("/valid-mock-module-path-fixtures/node_modules/legacy-main/lib/missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("node:not-a-builtin")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("#absent")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("#undefined-import")`, FileName: testFile, Errors: invalidAt(1)},
			// An `exports` or `imports` target must name an existing file exactly.
			{Code: `jest.mock("x-extensionless")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("x-pattern/sub/x")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("x-directory")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("x-directory-bare")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("#extensionless")`, FileName: testFile, Errors: invalidAt(1)},
			// The first valid target is final even when its file is missing.
			{Code: `jest.mock("x-missing-first")`, FileName: testFile, Errors: invalidAt(1)},
			// `?` and `#` are part of the path, not a query or fragment.
			{Code: `jest.mock("fs?raw")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("node:fs#x")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("eslint?x")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./fixtures/module/foo?x")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./fixtures/module/foo.ts#x")`, FileName: testFile, Errors: invalidAt(1)},
			// A POSIX backslash is a filename character, not a separator.
			{Code: `jest.mock(".\\fixtures\\module\\foo")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./fixtures\\module")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("legacy-main\\lib\\deep\\file.json")`, FileName: testFile, Errors: invalidAt(1)},
			// A directory with no index or package entry is not a module.
			{Code: `jest.mock("/")`, FileName: testFile, Errors: invalidAt(1)},

			// ---- Call shapes ----
			{Code: `jest['mock']("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: "jest[`doMock`](\"./missing\")", FileName: testFile, Errors: invalidAt(1)},
			// Like upstream, a computed identifier property counts by its name.
			{Code: `jest[mock]("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock?.("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest?.mock("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `(jest.mock)("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./fixtures/module").mock("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", () => ({}))`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.foo.mock("./missing")`, FileName: testFile, Errors: invalidAt(1)},
			{Code: "jest.mock(\"./missing\")`x`", FileName: testFile, Errors: invalidAt(1)},
			{
				Code:     `const mocked = jest.mock("./missing");`,
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Line:      1, Column: 16, EndLine: 1, EndColumn: 38,
				}},
			},
			{
				Code:     "import { jest } from '@jest/globals';\njest.mock('./missing');",
				FileName: testFile,
				Errors:   invalidAt(2),
			},
			{
				Code:     "import { jest as j } from '@jest/globals';\nj.doMock('./missing');",
				FileName: testFile,
				Errors:   invalidAt(2),
			},
			// The raw literal is reported without its parentheses.
			{
				Code:     `jest.mock(("./missing"))`,
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Message:   `Module path "./missing" does not exist or is not exported`,
					Line:      1, Column: 1, EndLine: 1, EndColumn: 25,
				}},
			},
			{
				Code:     `jest.mock('./escaped')`,
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Message:   `Module path './escaped' does not exist or is not exported`,
					Line:      1, Column: 1,
				}},
			},
			{
				Code:     `jest.mock(/** @type {string} */ ("./missing"))`,
				FileName: jsFile,
				Errors:   invalidAt(1),
			},
			{
				Code:     "jest.mock(\n  './missing',\n);",
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Line:      1, Column: 1, EndLine: 3, EndColumn: 2,
				}},
			},
			{
				Code:     "/* é */ jest.mock('./ü');",
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Message:   "Module path './ü' does not exist or is not exported",
					Line:      1, Column: 9, EndLine: 1, EndColumn: 25,
				}},
			},

			// ---- virtual: only a truthy literal in the third argument ----
			{Code: `jest.mock("./missing", { virtual: true })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual: 0n })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual: 0x0 })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual: "" })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual: null })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: "jest.mock(\"./missing\", undefined, { virtual: `yes` })", FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual: !0 })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual: true as boolean })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { virtual() { return true; } })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { get virtual() { return true; } })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock("./missing", undefined, { 1: true })`, FileName: testFile, Errors: invalidAt(1)},
		},
	)
}
