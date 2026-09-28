// cspell:ignore jackspeak doesnotexist
package valid_mock_module_path

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// testFile stands in for upstream's `__filename`, beside its `fixtures/` tree.
const testFile = "src/rules/__tests__/valid-mock-module-path.test.ts"

func mockModuleRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "valid-mock-module-path-fixtures")
	archive := txtarfs.MustParseFile(t, "testdata/fixtures.txtar")
	names, err := archive.FileNames("")
	if err != nil || len(names) == 0 {
		t.Fatalf("valid-mock-module-path fixtures: %v", err)
	}
	files := map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"commonjs","jsx":"preserve"},"include":["**/*"]}`,
	}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	// Windows forbids `?` in file names, so this one stays out of the archive,
	// which the JavaScript suite writes to disk.
	files[tspath.ResolvePath(directory, "node_modules/s-star/lib/entry?x.js")] = "module.exports = {};\n"
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

func invalidAt(line int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{MessageId: "invalidMockModulePath", Line: line, Column: 1}}
}

// All upstream cases, in order.
// https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/__tests__/valid-mock-module-path.test.ts
//
// Upstream resolves package names from eslint-plugin-jest's own install
// location; here they resolve from the linted file, so the fixture archive
// provides the packages upstream's cases rely on.
//
// Not ported: upstream's final `describe` block mocks `path.resolve` to throw
// an unexpected error code and asserts the rule rethrows it. That tests error
// propagation out of Node's resolver, which has no counterpart here.
func TestValidMockModulePathUpstream(t *testing.T) {
	root := mockModuleRoot(t)
	rule_tester.RunRuleTester(
		root,
		"tsconfig.json",
		t,
		&ValidMockModulePathRule,
		[]rule_tester.ValidTestCase{
			{Code: `jest.mock("./fixtures/module")`, FileName: testFile},
			{Code: `jest.mock("./fixtures/module", () => {})`, FileName: testFile},
			{Code: `jest.mock()`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module", () => {})`, FileName: testFile},
			{Code: `describe("foo", () => {});`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module")`, FileName: testFile},
			{Code: `jest.mock("./fixtures/module/foo.ts")`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module/foo.ts")`, FileName: testFile},
			{Code: `jest.mock("./fixtures/module/foo.js")`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module/foo.js")`, FileName: testFile},
			{Code: `jest.mock("eslint")`, FileName: testFile},
			{Code: `jest.doMock("eslint")`, FileName: testFile},
			{Code: `jest.mock("child_process")`, FileName: testFile},
			{Code: `jest.mock(() => {})`, FileName: testFile},
			{Code: "const a = \"../module/does/not/exist\";\njest.mock(a);", FileName: testFile},
			{Code: `jest.mock("./fixtures/module/jsx/foo")`, FileName: testFile},
			{Code: `jest.mock("./fixtures/module/tsx/foo")`, FileName: testFile},
			{
				Code:     `jest.mock("./fixtures/module/tsx/foo")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{".jsx"}},
			},
			{
				Code:     `jest.mock("./fixtures/module/bar")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{".json"}},
			},
			{
				Code:     `jest.mock("./fixtures/module/bar")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{".css"}},
			},
			{Code: `jest.mock("./fixtures/module/tsx/foo", undefined, { virtual: false })`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: false })`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module/tsx/foo", undefined, { ...{} })`, FileName: testFile},
			{Code: `jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: [] })`, FileName: testFile},
			{Code: `jest.mock("../module/does/not/exist", undefined, { virtual: true })`, FileName: testFile},
			{Code: `jest.doMock("../module/does/not/exist", undefined, { virtual: true })`, FileName: testFile},
			{Code: `jest.doMock("../module/does/not/exist", undefined, { "virtual": true })`, FileName: testFile},
			{Code: `jest.doMock("../module/does/not/exist", undefined, { ["virtual"]: true })`, FileName: testFile},
			{Code: "jest.doMock(\"../module/does/not/exist\", undefined, { [`virtual`]: true })", FileName: testFile},
			// jest only cares if the value is truthy...
			{Code: `jest.doMock("../module/does/not/exist", undefined, { virtual: 1 })`, FileName: testFile},
			// jest only cares if the value is truthy...
			{Code: `jest.doMock("../module/does/not/exist", undefined, { virtual: "yes" })`, FileName: testFile},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `jest.mock('../module/does/not/exist')`,
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Message:   "Module path '../module/does/not/exist' does not exist or is not exported",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 38,
				}},
			},
			{
				Code:     `jest.mock("../file/does/not/exist.ts")`,
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Message:   `Module path "../file/does/not/exist.ts" does not exist or is not exported`,
					Line:      1, Column: 1,
				}},
			},
			{
				Code:     `jest.mock("./fixtures/module/foo.jsx")`,
				FileName: testFile,
				Options:  map[string]any{"moduleFileExtensions": []any{".tsx"}},
				Errors:   invalidAt(1),
			},
			// Upstream passes `{ moduleFileExtensions: undefined }`, which JSON
			// options cannot express; an omitted key is the same input.
			{
				Code:     `jest.mock("./fixtures/module/foo.jsx")`,
				FileName: testFile,
				Options:  map[string]any{},
				Errors:   invalidAt(1),
			},
			{
				Code:     `jest.mock("@doesnotexist/module")`,
				FileName: testFile,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidMockModulePath",
					Message:   `Module path "@doesnotexist/module" does not exist or is not exported`,
					Line:      1, Column: 1,
				}},
			},
			// the imported file does not exist, but since it's not in `exports`
			// a ERR_PACKAGE_PATH_NOT_EXPORTED error will be thrown instead
			{
				Code:     `jest.mock("jest-util/build/isInteractive")`,
				FileName: testFile,
				Errors:   invalidAt(1),
			},
			// the imported file does exist, but since it's not in `exports`
			// a ERR_PACKAGE_PATH_NOT_EXPORTED error will be thrown instead
			{
				Code:     `jest.mock("jackspeak/dist/commonjs/parse-args.js")`,
				FileName: testFile,
				Errors:   invalidAt(1),
			},
			{Code: `jest.mock('../module/does/not/exist', undefined, {})`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock('../module/does/not/exist', undefined, { assumeExists: true })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.mock('../module/does/not/exist', undefined, { virtual: false })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.doMock('../module/does/not/exist', undefined, { virtual: false })`, FileName: testFile, Errors: invalidAt(1)},
			{Code: `jest.doMock('../module/does/not/exist', undefined, { virtual: 0 })`, FileName: testFile, Errors: invalidAt(1)},
			{
				Code:     "const virtual = false;\n\njest.doMock('../module/does/not/exist', undefined, { virtual })",
				FileName: testFile,
				Errors:   invalidAt(3),
			},
			{
				Code:     "const virtual = true;\n\njest.doMock('../module/does/not/exist', undefined, { virtual })",
				FileName: testFile,
				Errors:   invalidAt(3),
			},
			{
				Code:     "const prop = 'virtual';\n\njest.doMock('../module/does/not/exist', undefined, { [prop]: true })",
				FileName: testFile,
				Errors:   invalidAt(3),
			},
			// we don't attempt to resolve the result of object spreads
			{
				Code:     `jest.doMock('../module/does/not/exist', undefined, { ...{ virtual: true } })`,
				FileName: testFile,
				Errors:   invalidAt(1),
			},
		},
	)
}
