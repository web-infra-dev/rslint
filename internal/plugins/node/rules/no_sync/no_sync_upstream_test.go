package no_sync

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func syncRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-no-sync")
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

func syncError(name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "noSync", Message: "Unexpected sync method: '" + name + "'.", Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}
}

// Every test and executable documentation example from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-sync.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-sync.md
func TestNoSyncUpstream(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			{Code: "var foo = fs.foo.foo();", FileName: "input.js"},
			{Code: "fs.fooSync;", FileName: "input.js"},
			{Code: "fooSync;", FileName: "input.js"},
			{Code: "() => fooSync;", FileName: "input.js"},
			{Code: "var foo = fs.fooSync;", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			{Code: "var foo = fooSync;", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			{Code: "if (true) {fs.fooSync();}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			{Code: "if (true) {fooSync();}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			{Code: "fooSync();", FileName: "input.js", Options: []any{map[string]any{"ignores": []any{"fooSync"}}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "var foo = fs.fooSync();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 21)}},
			{Code: "var foo = fs.fooSync.apply();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 21)}},
			{Code: "var foo = fooSync();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 20)}},
			{Code: "var foo = fooSync.apply();", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 24)}},
			{Code: "var foo = fs.fooSync();", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": false}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 11, 1, 21)}},
			{Code: "if (true) {fs.fooSync();}", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 12, 1, 22)}},
			{Code: "function someFunction() {fs.fooSync();}", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 26, 1, 36)}},
			{Code: "function someFunction() {fs.fooSync();}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 26, 1, 36)}},
			{Code: "var a = function someFunction() {fs.fooSync();}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 34, 1, 44)}},
			{Code: "() => {fs.fooSync();}", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true, "ignores": []any{"barSync"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 1, 8, 1, 18)}},
		})
}

func TestNoSyncUpstreamTypes(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			{Code: "\ndeclare function fooSync(): void;\nfooSync();\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file"}}}}},
			{Code: "\ndeclare function fooSync(): void;\nfooSync();\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"fooSync"}}}}}},
			{Code: "\nconst stylesheet = new CSSStyleSheet();\nstylesheet.replaceSync(\"body { font-size: 1.4em; } p { color: red; }\");\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "lib", "name": []any{"CSSStyleSheet.replaceSync"}}}}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "\ndeclare function fooSync(): void;\nfooSync();\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "path": "**/bar.ts"}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 3, 1, 3, 10)}},
			{Code: "\ndeclare function fooSync(): void;\nfooSync();\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"barSync"}}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 3, 1, 3, 10)}},
			{Code: "\nconst stylesheet = new CSSStyleSheet();\nstylesheet.replaceSync(\"body { font-size: 1.4em; } p { color: red; }\");\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"CSSStyleSheet.replaceSync"}}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("CSSStyleSheet.replaceSync", 3, 1, 3, 23)}},
		})
}

func TestNoSyncUpstreamPackage(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			{Code: "\nimport { fooSync } from \"aaa\";\nfooSync();\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package", "package": "aaa", "name": []any{"fooSync"}}}}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "\nimport { fooSync } from \"aaa\";\nfooSync();\n", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "name": []any{"fooSync"}}}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("fooSync", 3, 1, 3, 10)}},
		})
}

func TestNoSyncDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(syncRoot(t), "tsconfig.json", t, &NoSyncRule,
		[]rule_tester.ValidTestCase{
			{Code: "obj.sync();\n\nasync(function() {\n    // ...\n});", FileName: "input.js"},
			{Code: "fs.readFileSync(somePath).toString();", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}},
			{Code: "fs.readFileSync(somePath);", FileName: "input.js", Options: []any{map[string]any{"ignores": []any{"readFileSync"}}}},
			{Code: "import { fooSync } from \"./foo\"\nfooSync()", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "file", "path": "./foo.ts"}}}}},
			{Code: "import { Effect } from \"effect\"\nconst value = Effect.runSync(Effect.succeed(42))", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "package", "package": "effect"}}}}},
			{Code: "const stylesheet = new CSSStyleSheet()\nstylesheet.replaceSync(\"body { font-size: 1.4em; } p { color: red; }\")", FileName: "input.ts", Options: []any{map[string]any{"ignores": []any{map[string]any{"from": "lib"}}}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "fs.existsSync(somePath);\n\nfunction foo() {\n  var contents = fs.readFileSync(somePath).toString();\n}", FileName: "input.js", Errors: []rule_tester.InvalidTestCaseError{syncError("existsSync", 1, 1, 1, 14), syncError("readFileSync", 4, 18, 4, 33)}},
			{Code: "function foo() {\n  var contents = fs.readFileSync(somePath).toString();\n}\n\nvar bar = baz => fs.readFileSync(qux);", FileName: "input.js", Options: []any{map[string]any{"allowAtRootLevel": true}}, Errors: []rule_tester.InvalidTestCaseError{syncError("readFileSync", 2, 18, 2, 33), syncError("readFileSync", 5, 18, 5, 33)}},
			{Code: "fs.readdirSync(somePath);", FileName: "input.js", Options: []any{map[string]any{"ignores": []any{"readFileSync"}}}, Errors: []rule_tester.InvalidTestCaseError{syncError("readdirSync", 1, 1, 1, 15)}},
		})
}

// These upstream tests mock JavaScript dependencies; native Go uses its own checker.
func TestNoSyncUpstreamMissingDependencies(t *testing.T) {
	for _, name := range []string{"ts-declaration-location", "TypeScript parser services", "both dependencies"} {
		t.Run(name, func(t *testing.T) { t.Skip("JavaScript dependency injection does not apply to the native rule") })
	}
}
