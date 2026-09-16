package no_unpublished_bin

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func unpublishedBinRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-unpublished-bin-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"nodenext","jsx":"preserve"},"include":["**/*"]}`,
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

// All 39 cases from eslint-plugin-n v18.3.0, with diagnostic ranges checked
// against ESLint 10.2.1. No upstream cases are skipped.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unpublished-bin.js
func TestNoUnpublishedBinUpstream(t *testing.T) {
	rule_tester.RunRuleTester(unpublishedBinRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoUnpublishedBinRule,
		[]rule_tester.ValidTestCase{
			// Package bin entries and publication patterns.
			{Code: "'simple-ok/a.js'", FileName: "simple-ok/a.js"},
			{Code: "'multi-ok/a.js'", FileName: "multi-ok/a.js"},
			{Code: "'multi-ok/b.js'", FileName: "multi-ok/b.js"},
			{Code: "'simple-files/x.js'", FileName: "simple-files/x.js"},
			{Code: "'multi-files/x.js'", FileName: "multi-files/x.js"},
			{Code: "'simple-files/lib/a.js'", FileName: "simple-files/lib/a.js"},
			{Code: "'multi-files/lib/a.js'", FileName: "multi-files/lib/a.js"},
			{Code: "'simple-npmignore/x.js'", FileName: "simple-npmignore/x.js"},
			{Code: "'multi-npmignore/x.js'", FileName: "multi-npmignore/x.js"},
			{Code: "'simple-npmignore/lib/a.js'", FileName: "simple-npmignore/lib/a.js"},
			{Code: "'multi-npmignore/lib/a.js'", FileName: "multi-npmignore/lib/a.js"},
			{Code: "'issue115/lib/a.js'", FileName: "issue115/lib/a.js"},
			{Code: "'brace-extglob/bin/foo.js'", FileName: "brace-extglob/bin/foo.js"},
			// Unnamed input (represented by a virtual file without package metadata).
			{Code: "'stdin'", FileName: "input.js"},
			// convertPath option.
			{Code: "'simple-files/a.js'", FileName: "simple-files/a.js", Options: []any{map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			{Code: "'multi-files/a.js'", FileName: "multi-files/a.js", Options: []any{map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			{Code: "'simple-npmignore/a.js'", FileName: "simple-npmignore/a.js", Options: []any{map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			{Code: "'multi-npmignore/a.js'", FileName: "multi-npmignore/a.js", Options: []any{map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			// convertPath shared setting.
			{Code: "'simple-files/a.js'", FileName: "simple-files/a.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			{Code: "'multi-files/a.js'", FileName: "multi-files/a.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			{Code: "'simple-npmignore/a.js'", FileName: "simple-npmignore/a.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			{Code: "'multi-npmignore/a.js'", FileName: "multi-npmignore/a.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"a.js": []any{"a.js", "lib/a.js"}}}}},
			// The documentation package example has no publication exclusions.
			{Code: "", FileName: "documentation/bin/index.js"},
		},
		[]rule_tester.InvalidTestCase{
			// files field of package.json.
			{Code: "'simple-files/a.js'", FileName: "simple-files/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{Code: "'multi-files/a.js'", FileName: "multi-files/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			{Code: "'multi-files/b.js'", FileName: "multi-files/b.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'b.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			// .npmignore.
			{Code: "'simple-npmignore/a.js'", FileName: "simple-npmignore/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}},
			{Code: "'multi-npmignore/a.js'", FileName: "multi-npmignore/a.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			// files field with convertPath.
			{Code: "'simple-files/x.js'", FileName: "simple-files/x.js", Options: []any{map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{Code: "'multi-files/x.js'", FileName: "multi-files/x.js", Options: []any{map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			{Code: "'multi-files/x.js'", FileName: "multi-files/x.js", Options: []any{map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "b.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'b.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			// .npmignore with object and array convertPath.
			{Code: "'simple-npmignore/x.js'", FileName: "simple-npmignore/x.js", Options: []any{map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}},
			{Code: "'multi-npmignore/x.js'", FileName: "multi-npmignore/x.js", Options: []any{map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			{Code: "'simple-npmignore/x.js'", FileName: "simple-npmignore/x.js", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"x.js"}, "replace": []any{"x.js", "a.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}},
			{Code: "'multi-npmignore/x.js'", FileName: "multi-npmignore/x.js", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"x.js"}, "replace": []any{"x.js", "a.js"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
			// files field with shared convertPath.
			{Code: "'simple-files/x.js'", FileName: "simple-files/x.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{Code: "'multi-files/x.js'", FileName: "multi-files/x.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			{Code: "'multi-files/x.js'", FileName: "multi-files/x.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "b.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'b.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
			// .npmignore with shared convertPath.
			{Code: "'simple-npmignore/x.js'", FileName: "simple-npmignore/x.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}},
			{Code: "'multi-npmignore/x.js'", FileName: "multi-npmignore/x.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"x.js": []any{"x.js", "a.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidIgnored", Message: "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
		},
	)
}
