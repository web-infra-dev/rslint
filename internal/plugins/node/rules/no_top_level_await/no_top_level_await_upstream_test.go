package no_top_level_await_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_top_level_await"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func awaitRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "no-top-level-await-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"nodenext","jsx":"preserve"},"include":["**/*","**/.*"]}`,
	}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

func forbiddenAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "forbidden",
		Message:   "Top-level `await` is forbidden in published modules.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// All 22 valid and 14 invalid cases from:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-top-level-await.js
// Unnamed ESLint input is represented by a virtual file outside any package.
func TestNoTopLevelAwaitUpstream(t *testing.T) {
	rule_tester.RunRuleTester(awaitRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_top_level_await.NoTopLevelAwaitRule,
		[]rule_tester.ValidTestCase{
			// Published synchronous code.
			{Code: "import * as foo from 'foo'", FileName: "simple-files/lib/a.js"},
			{Code: "for (const e of iterate()) { /* ... */ }", FileName: "simple-files/lib/a.js"},
			// Non-top-level await.
			{Code: "async function fn () { const foo = await import('foo') }", FileName: "simple-bin/lib/a.js"},
			{Code: "async function fn () { for await (const e of asyncIterate()) { /* ... */ } }", FileName: "simple-bin/lib/a.js"},
			{Code: "const fn = async () => await import('foo')", FileName: "simple-bin/lib/a.js"},
			{Code: "const fn = async () => { for await (const e of asyncIterate()) { /* ... */ } }", FileName: "simple-bin/lib/a.js"},
			// Unpublished files.
			{Code: "const foo = await import('foo')", FileName: "simple-files/src/a.js"},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-files/src/a.js"},
			{Code: "const foo = await import('foo')", FileName: "dot-slash-files/src/a.js"},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "dot-slash-files/src/a.js"},
			{Code: "const foo = await import('foo')", FileName: "slash-files/src/a.js"},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "slash-files/src/a.js"},
			{Code: "const foo = await import('foo')", FileName: "simple-npmignore/src/a.js"},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-npmignore/src/a.js"},
			// ignoreBin.
			{Code: "const foo = await import('foo')", FileName: "simple-bin/a.js", Options: []any{map[string]any{"ignoreBin": true}}},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-bin/a.js", Options: []any{map[string]any{"ignoreBin": true}}},
			{Code: "#!/usr/bin/env node\nconst foo = await import('foo')", FileName: "simple-files/lib/a.js", Options: []any{map[string]any{"ignoreBin": true}}},
			{Code: "#!/usr/bin/env node\nfor await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-files/lib/a.js", Options: []any{map[string]any{"ignoreBin": true}}},
			// await using.
			{Code: "async function f() { await using foo = x }", FileName: "simple-files/lib/a.js"},
			// convertPath.
			{Code: "const foo = await import('foo')", FileName: "simple-files/test/a.ts", Options: []any{map[string]any{"convertPath": map[string]any{"src/**/*": []any{"src/(.+).ts", "lib/$1.js"}}}}},
			// Unknown files.
			{Code: "const foo = await import('foo')"},
			{Code: "const foo = await import('foo')", FileName: "unknown.js"},
		},
		[]rule_tester.InvalidTestCase{
			// Published files, including executables by default.
			{Code: "const foo = await import('foo')", FileName: "simple-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32)}},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 52)}},
			{Code: "const foo = await import('foo')", FileName: "dot-slash-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32)}},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "dot-slash-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 52)}},
			{Code: "const foo = await import('foo')", FileName: "slash-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32)}},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "slash-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 52)}},
			{Code: "const foo = await import('foo')", FileName: "simple-npmignore/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32)}},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-npmignore/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 52)}},
			{Code: "const foo = await import('foo')", FileName: "simple-bin/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32)}},
			{Code: "for await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-bin/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 52)}},
			{Code: "#!/usr/bin/env node\nconst foo = await import('foo')", FileName: "simple-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 13, 2, 32)}},
			{Code: "#!/usr/bin/env node\nfor await (const e of asyncIterate()) { /* ... */ }", FileName: "simple-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 1, 2, 52)}},
			// convertPath.
			{Code: "const foo = await import('foo')", FileName: "simple-files/src/a.ts", Options: []any{map[string]any{"convertPath": map[string]any{"src/**/*": []any{"src/(.+).ts", "lib/$1.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32)}},
			// await using.
			{Code: "await using foo = x", FileName: "simple-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 20)}},
		},
	)
}

// Both JavaScript examples from:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-top-level-await.md
// The tester supplies the rule-enabling comments as configuration.
func TestNoTopLevelAwaitDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(awaitRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_top_level_await.NoTopLevelAwaitRule,
		[]rule_tester.ValidTestCase{
			{Code: "#!/usr/bin/env node\nconst foo = await import('foo');\nfor await (const e of asyncIterate()) {\n    // ...\n}", FileName: "simple-files/lib/a.js", Options: []any{map[string]any{"ignoreBin": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const foo = await import('foo');\nfor await (const e of asyncIterate()) {\n    // ...\n}", FileName: "simple-files/lib/a.js", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 32), forbiddenAt(2, 1, 4, 2)}},
		},
	)
}
