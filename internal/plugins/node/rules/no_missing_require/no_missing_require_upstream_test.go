// cspell:ignore jquery
package no_missing_require

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func missingRequireRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-missing-require-fixtures")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil || len(names) == 0 {
		t.Fatalf("missing require fixtures: %v", err)
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

// All 79 upstream cases, followed by the documentation's source examples.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-missing-require.js
func TestNoMissingRequireUpstream(t *testing.T) {
	root := missingRequireRoot(t, "testdata/upstream.txtar")
	valid := []rule_tester.ValidTestCase{
		// upstream valid 1
		{Code: "require('fs');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 2
		{Code: "require('node:fs');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 3
		{Code: "require('node:test');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 4
		{Code: "require('eslint');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 5
		{Code: "require('rimraf/package.json');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 6
		{Code: "require('./a');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 7
		{Code: "require('./a.js');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 8
		{Code: "require('./a.config');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 9
		{Code: "require('./a.config.js');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 10
		{Code: "require('./b');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 11
		{Code: "require('./b.json');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 12
		{Code: "require('./c.coffee');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 13
		{Code: "require('mocha');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 14
		{Code: "require(`eslint`);", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 15
		{Code: "require('mocha!foo?a=b&c=d');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 16
		{Code: "require('./c');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"tryExtensions": []any{".coffee"}}}},
		// upstream valid 17
		{Code: "require('./c');", FileName: "tests/fixtures/no-missing/test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".coffee"}}}},
		// upstream valid 18
		{Code: "require('./fixtures/no-missing/a');", FileName: "tests/fixtures/no-missing/test.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{tspath.ResolvePath(root.Dir, "tests")}}}},
		// upstream valid 19
		{Code: "require('./fixtures/no-missing/a');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolvePaths": []any{tspath.ResolvePath(root.Dir, "tests")}}}},
		// upstream valid 20
		{Code: "require('./fixtures/no-missing/a');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolvePaths": []any{"tests"}}}},
		// upstream valid 21
		{Code: "require('./fixtures/no-missing/a');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolvePaths": []any{"scripts", "tests"}}}},
		// upstream valid 22
		{Code: "require('./a');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolvePaths": []any{"tests"}}}},
		// upstream valid 23
		{Code: "require('a');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing")}}}}},
		// upstream valid 24
		{Code: "require('my-module');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"modules": []any{tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing/my_modules")}}}}},
		// upstream valid 25
		{Code: "require;", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 26
		{Code: "require('no-exist-package-0');", FileName: "tests/fixtures/no-missing/test.js", Globals: map[string]any{"require": "off"}},
		// upstream valid 27: Unknown <input> filenames cannot be represented by the native parser or lint API.
		{Code: "require('no-exist-package-0');", Skip: true},
		// upstream valid 28: Unknown <input> filenames cannot be represented by the native parser or lint API.
		{Code: "require('./b');", Skip: true},
		// upstream valid 29
		{Code: "require();", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 30
		{Code: "require(foo);", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 31
		{Code: "require(`foo${bar}`);", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 32
		{Code: "require('eslint');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 33
		{Code: "require('./a');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 34
		{Code: "require('.');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 35
		{Code: "require('./');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 36
		{Code: "require('./foo');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 37
		{Code: "require('./foo/');", FileName: "tests/fixtures/no-missing/test.js"},
		// upstream valid 38
		{Code: "require('electron');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"electron"}}}},
		// upstream valid 39
		{Code: "require('jquery.cookie');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"jquery.cookie"}}}},
		// upstream valid 40
		{Code: "require('virtual:package-name');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-name"}}}},
		// upstream valid 41
		{Code: "require('virtual:package-scope/name');", FileName: "tests/fixtures/no-missing/test.js", Options: []any{map[string]any{"allowModules": []any{"virtual:package-scope"}}}},
		// upstream valid 42
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 43
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 44
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 45
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 46
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 47
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/test.ts", Options: []any{map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 48
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 49
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/test.ts", Options: []any{map[string]any{"typescriptExtensionMap": []any{[]any{"", ".js"}, []any{".ts", ".js"}, []any{".cts", ".cjs"}, []any{".mts", ".mjs"}, []any{".tsx", ".js"}}}}},
		// upstream valid 50
		{Code: "require('./e.jsx');", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "preserve"}}},
		// upstream valid 51
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/test.tsx", Options: []any{map[string]any{"typescriptExtensionMap": "react"}}},
		// upstream valid 52
		{Code: "require('./e.jsx');", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "preserve"}}},
		// upstream valid 53
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/test.tsx", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react"}}},
		// upstream valid 54
		{Code: "require('./e.jsx');", FileName: "tests/fixtures/no-missing/ts-react/test.tsx", Options: []any{map[string]any{"tsconfigPath": tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing/ts-preserve/tsconfig.json")}}},
		// upstream valid 55
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/ts-preserve/test.tsx", Options: []any{map[string]any{"tsconfigPath": tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing/ts-react/tsconfig.json")}}},
		// upstream valid 56
		{Code: "require('./e.jsx');", FileName: "tests/fixtures/no-missing/ts-react/test.tsx", Settings: map[string]any{"node": map[string]any{"tsconfigPath": tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing/ts-preserve/tsconfig.json")}}},
		// upstream valid 57
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/ts-preserve/test.tsx", Settings: map[string]any{"node": map[string]any{"tsconfigPath": tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing/ts-react/tsconfig.json")}}},
		// upstream valid 58
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/ts-react/test.tsx"},
		// upstream valid 59
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/ts-react/test.ts"},
		// upstream valid 60
		{Code: "require('./e.jsx');", FileName: "tests/fixtures/no-missing/ts-preserve/test.tsx"},
		// upstream valid 61
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/ts-preserve/test.ts"},
		// upstream valid 62
		{Code: "require('./e.js');", FileName: "tests/fixtures/no-missing/ts-extends/test.tsx"},
		// upstream valid 63
		{Code: "require('./d.js');", FileName: "tests/fixtures/no-missing/ts-extends/test.ts"},
		// upstream valid 64
		{Code: "require.resolve('eslint');", FileName: "tests/fixtures/no-missing/test.js"},
		// documentation: existing and dynamic targets
		{Code: "var existingFile = require(\"./existing-file\");\nvar existingModule = require(\"existing-module\");\nvar foo = require(FOO_NAME);", FileName: "tests/fixtures/no-missing/test.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// upstream invalid 1
		{Code: "require('no-exist-package-0');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'no-exist-package-0' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 29}}},
		// upstream invalid 2
		{Code: "require('@mysticatea/test');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve '@mysticatea/test' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 27}}},
		// upstream invalid 3
		{Code: "require('./c');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './c' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// upstream invalid 4
		{Code: "require('./d');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './d' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// upstream invalid 5
		{Code: "require('./a.json');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './a.json' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 19}}},
		// upstream invalid 6
		{Code: "require('no-exist-package-0');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'no-exist-package-0' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 29}}},
		// upstream invalid 7
		{Code: "require('./c');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './c' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// upstream invalid 8
		{Code: "require('./bar');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './bar' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		// upstream invalid 9
		{Code: "require('./bar/');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './bar/' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
		// upstream invalid 10
		{Code: "require('./A');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './A' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
		// upstream invalid 11
		{Code: "require.resolve('no-exist-package-0');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'no-exist-package-0' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 17, EndLine: 1, EndColumn: 37}}},
		// upstream invalid 12
		{Code: "require('virtual:package-name');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'virtual:package-name' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 31}}},
		// upstream invalid 13
		{Code: "require('virtual:package-scope/name');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'virtual:package-scope/name' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 37}}},
		// upstream invalid 14
		{Code: "require('data:text/javascript,const x = 123;');", FileName: "tests/fixtures/no-missing/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'data:text/javascript,const x = 123;' in '/__root__/tests/fixtures/no-missing'", "/__root__", root.Dir), Line: 1, Column: 9, EndLine: 1, EndColumn: 46}}},
		// documentation: runtime error
		{Code: "const foo = require(\"./foo\");", FileName: "documentation/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './foo' in '/__root__/documentation'", "/__root__", root.Dir), Line: 1, Column: 21, EndLine: 1, EndColumn: 28}}},
		// documentation: missing file and package
		{Code: "var typoFile = require(\"./typo-file\");\nvar typoModule = require(\"typo-module\");", FileName: "documentation/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve './typo-file' in '/__root__/documentation'", "/__root__", root.Dir), Line: 1, Column: 24, EndLine: 1, EndColumn: 37}, {MessageId: "notFound", Message: strings.ReplaceAll("Can't resolve 'typo-module' in '/__root__/documentation'", "/__root__", root.Dir), Line: 2, Column: 26, EndLine: 2, EndColumn: 39}}},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoMissingRequireRule, valid, invalid)
}

// The same absolute fixture is linted with its directory as the working root.
func TestNoMissingRequireUpstreamWorkingDirectory(t *testing.T) {
	root := missingRequireRoot(t, "testdata/upstream.txtar")
	root.Dir = tspath.ResolvePath(root.Dir, "tests/fixtures/no-missing")
	rule_tester.RunRuleTester(root, "../../../tsconfig.json", t, &NoMissingRequireRule,
		[]rule_tester.ValidTestCase{{Code: "require('../../lib/rules/no-missing-require');", FileName: "test.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}}}, nil)
}
