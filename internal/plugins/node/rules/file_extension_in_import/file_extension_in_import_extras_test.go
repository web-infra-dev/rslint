// cspell:ignore jsxdev
package file_extension_in_import

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Additional syntax and filesystem cases checked against eslint-plugin-n v18.3.0.
// Escaped-path removals preserve source spelling instead of upstream's unsafe offsets.
func TestFileExtensionInImportExtras(t *testing.T) {
	root := extensionRoot(t, "testdata/upstream.txtar", "testdata/extras.txtar")
	absolute := tspath.ResolvePath(root.Dir, "a")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule,
		[]rule_tester.ValidTestCase{
			// ignored import forms
			{Code: "import fs from 'node:fs'; import 'https://example.test/a'; import 'data:text/javascript,0'; import '#local'; import 'pkg/deep.js'; import '.prefix'; import(`./a`); import('./' + 'a'); require('./a'); new URL('./a', import.meta.url); export {};", FileName: "test.js"},
			// authored TS wrappers
			{Code: "import(('./a' as string)); import(('./a'!)); type T = import('./d').T; import alias = require('./a');", FileName: "test.ts"},
			{Code: "import(('./a' satisfies string)); import(0); import(null); import(true); import(1n);", FileName: "test.ts"},
			// missing JS targets
			{Code: "import './missing'; import './missing.xyz'; import './empty'; import './.env';", FileName: "test.js"},
			// empty extensions
			{Code: "import './a';", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{}}}},
			// directory override
			{Code: "import './my-folder'", FileName: "test.js", Options: []any{"always", map[string]any{".js": "never"}}},
			// allow explicit TS
			{Code: "import './lib.ts'", FileName: "allow/test.ts"},
			// explicit empty mapping
			{Code: "import './d'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{}}}},
			// query resource
			{Code: "import './a.js?raw'", FileName: "test.js"},
		},
		[]rule_tester.InvalidTestCase{
			// Absolute local paths use the same resolver and source-token range.
			{Code: "import " + strconv.Quote(absolute), FileName: "test.js", Output: []string{"import " + strconv.Quote(absolute+".js")}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: len("import "+strconv.Quote(absolute)) + 1}}},
			// exports
			{Code: "export * from './a'; export {default as value} from './b';", FileName: "test.js", Output: []string{"export * from './a.js'; export {default as value} from './b.json';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 20}, {MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 53, EndLine: 1, EndColumn: 58}}},
			// type declarations
			{Code: "import type { T } from './d'; export type { T } from './d'; import {type T as U} from './d';", FileName: "test.ts", Output: []string{"import type { T } from './d.js'; export type { T } from './d.js'; import {type T as U} from './d.js';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 29}, {MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 54, EndLine: 1, EndColumn: 59}, {MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 87, EndLine: 1, EndColumn: 92}}},
			// namespace export
			{Code: "export * as data from './a';", FileName: "test.js", Output: []string{"export * as data from './a.js';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28}}},
			// dynamic parentheses
			{Code: "import( /* before */ ('./a') /* after */ );", FileName: "test.js", Output: []string{"import( /* before */ ('./a.js') /* after */ );"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28}}},
			// JSDoc parentheses
			{Code: "import(/** @type {string} */ ('./a'));", FileName: "test.js", Output: []string{"import(/** @type {string} */ ('./a.js'));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36}}},
			// import attributes
			{Code: "import value from './b' with {type: 'json'};", FileName: "test.js", Output: []string{"import value from './b.json' with {type: 'json'};"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 24}}},
			// multiline and Unicode
			{Code: "const label = '😀';\nimport (\n './café'\n); import './😀';", FileName: "test.js", Output: []string{"const label = '😀';\nimport (\n './café.js'\n); import './😀.js';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 3, Column: 2, EndLine: 3, EndColumn: 10}, {MessageId: "requireExt", Message: "require file extension '.js'.", Line: 4, Column: 11, EndLine: 4, EndColumn: 17}}},
			// Unicode removal
			{Code: "const label = '😀'; import './café.js'; import './😀.js';", FileName: "test.js", Options: []any{"never"}, Output: []string{"const label = '😀'; import './café'; import './😀';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 39}, {MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 48, EndLine: 1, EndColumn: 57}}},
			// hidden basename
			{Code: "import './.hidden'", FileName: "test.js", Output: []string{"import './.hidden.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// missing TS fallback
			{Code: "import './missing'", FileName: "test.ts", Output: []string{"import './missing.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// nonexistent explicit path
			{Code: "import './missing.js'", FileName: "test.js", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
			// directory ambiguity
			{Code: "import './has-dir.js'", FileName: "test.js", Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
			// index fallback
			{Code: "import './coffee'", FileName: "test.js", Output: []string{"import './coffee/index.coffee'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.coffee'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// index preference
			{Code: "import './coffee/'", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".css", ".coffee"}}}, Output: []string{"import './coffee/index.css'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.css'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// index empty preferences
			{Code: "import './coffee'", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{}}}, Output: []string{"import './coffee/index.coffee'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.coffee'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// priority override
			{Code: "import './multi'", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".json", ".js"}}}, Output: []string{"import './multi.json'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// namespace precedence
			{Code: "import './multi'", FileName: "test.js", Settings: map[string]any{"n": map[string]any{"tryExtensions": []any{".json"}}, "node": map[string]any{"tryExtensions": []any{".js"}}}, Output: []string{"import './multi.json'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.json'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// resolved extension override
			{Code: "import './d.ts'", FileName: "test.ts", Options: []any{"never", map[string]any{".ts": "always"}}, Output: []string{"import './d.ts.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// mapped extension is not override key
			{Code: "import './d'", FileName: "test.ts", Options: []any{"always", map[string]any{".js": "never"}}, Output: []string{"import './d.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// explicit defaults
			{Code: "import './a'", FileName: "test.js", Options: []any{"always", map[string]any{}}, Output: []string{"import './a.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// mapped TS removal
			{Code: "import './d.js'", FileName: "test.ts", Options: []any{"never"}, Output: []string{"import './d'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// TS emitted module extensions
			{Code: "import './lib.mts'; import './lib.cts';", FileName: "test.ts", Output: []string{"import './lib.mts.mjs'; import './lib.cts.cjs';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}, {MessageId: "requireExt", Message: "require file extension '.cjs'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 39}}},
			// allow TS extensions
			{Code: "import './lib'; import './view';", FileName: "allow/test.ts", Output: []string{"import './lib.ts'; import './view';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.ts'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
			// inherited JSX emit
			{Code: "import './component'", FileName: "react/test.ts", Output: []string{"import './component.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// react native emit
			{Code: "import './component'", FileName: "native/test.ts", Output: []string{"import './component.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// custom duplicate map
			{Code: "import './d.ts'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{".ts", ".mjs"}, []any{".ts", ".cjs"}}}}, Output: []string{"import './d.ts.cjs'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.cjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// custom empty extension
			{Code: "import './missing'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{"", ".mjs"}}}}, Output: []string{"import './missing.mjs'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.mjs'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// JS ignores TS mapping
			{Code: "import './util.client'", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{".js", ".mjs"}}}}, Output: []string{"import './util.client.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
			// mapping namespace precedence
			{Code: "import './e.tsx'", FileName: "test.ts", Settings: map[string]any{"n": map[string]any{}, "node": map[string]any{"typescriptExtensionMap": "react"}}, Output: []string{"import './e.tsx.jsx'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.jsx'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// tsconfigPath selection
			{Code: "import './view.tsx'", FileName: "custom/test.ts", Settings: map[string]any{"node": map[string]any{"tsconfigPath": strings.ReplaceAll("{{root}}/custom/react.json", "{{root}}", root.Dir)}}, Output: []string{"import './view.tsx.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// config without JSX falls back
			{Code: "import './view.tsx'", FileName: "custom/test.ts", Settings: map[string]any{"node": map[string]any{"tsconfigPath": strings.ReplaceAll("{{root}}/custom/no-jsx.json", "{{root}}", root.Dir)}}, Output: []string{"import './view.tsx.jsx'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.jsx'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// resolvePaths selection
			{Code: "import './only'", FileName: "test.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{strings.ReplaceAll("{{root}}/search", "{{root}}", root.Dir)}}}, Output: []string{"import './only.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// escaped path safe removal
			{Code: "import './\\u0061.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './\\u0061'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// escaped extension safe removal
			{Code: "import './a.\\x6as'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// escaped path insertion
			{Code: "import './\\u0061'", FileName: "test.js", Output: []string{"import './\\u0061.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// ESTree also treats regular expressions as literals; do not rewrite them as strings.
			{Code: "import(/missing/)", FileName: "test.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// loader suffix
			{Code: "import './a.js!loader'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a!loader'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
			// preset preserve
			{Code: "import './e.tsx'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "preserve"}}, Output: []string{"import './e.tsx.jsx'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.jsx'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// preset react
			{Code: "import './e.tsx'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react"}}, Output: []string{"import './e.tsx.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// preset react-jsx
			{Code: "import './e.tsx'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react-jsx"}}, Output: []string{"import './e.tsx.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// preset react-jsxdev
			{Code: "import './e.tsx'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react-jsxdev"}}, Output: []string{"import './e.tsx.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// preset react-native
			{Code: "import './e.tsx'", FileName: "test.ts", Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": "react-native"}}, Output: []string{"import './e.tsx.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		},
	)
}

func TestFileExtensionInImportUnsafeExtension(t *testing.T) {
	root := extensionRoot(t, "testdata/upstream.txtar")
	for _, extension := range []string{".x'js", ".x\\js", ".x\njs", ".x\rjs"} {
		t.Run(extension, func(t *testing.T) {
			rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, nil, []rule_tester.InvalidTestCase{{
				Code: "import './d.ts'", FileName: "test.ts",
				Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{".ts", extension}}}},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '" + extension + "'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}},
			}})
		})
	}
	// A quote different from the literal's delimiter remains safe to insert.
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, nil, []rule_tester.InvalidTestCase{{
		Code: `import "./d.ts"`, FileName: "test.ts", Output: []string{`import "./d.ts.x'js"`},
		Settings: map[string]any{"node": map[string]any{"typescriptExtensionMap": []any{[]any{".ts", ".x'js"}}}},
		Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.x'js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}},
	}})
}

func TestFileExtensionInImportLiteralBackslash(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skip("Backslash is a literal filename character only on POSIX hosts.")
	}
	root := extensionRoot(t, "testdata/upstream.txtar")
	code := `import '.\\my-folder'`
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule,
		[]rule_tester.ValidTestCase{{Code: code, FileName: "test.js"}},
		[]rule_tester.InvalidTestCase{{
			Code: code, FileName: "test.ts",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: len(code) + 1}},
		}})
}

func TestFileExtensionInImportWindowsSeparators(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skip("Windows path separators require a Windows host.")
	}
	root := extensionRoot(t, "testdata/upstream.txtar")
	rooted := filepath.FromSlash(tspath.ResolvePath(root.Dir, "a.js"))
	rooted = strings.TrimPrefix(rooted, filepath.VolumeName(rooted))
	code := "import " + strconv.Quote(rooted)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, nil, []rule_tester.InvalidTestCase{{
		Code: `import '.\\\u0061.\x6as'`, FileName: "test.js", Options: []any{"never"}, Output: []string{`import '.\\\u0061'`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}},
	}, {
		Code: `import '.\\my-folder\\'`, FileName: "test.js", Output: []string{`import '.\\my-folder\\index.js'`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireExt", Message: "require file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}},
	}, {
		Code: code, FileName: "test.js", Options: []any{"never"}, Output: []string{"import " + strconv.Quote(strings.TrimSuffix(rooted, ".js"))},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: len(code) + 1}},
	}})
}

func TestFileExtensionInImportUnterminatedLiteral(t *testing.T) {
	root := extensionRoot(t, "testdata/upstream.txtar")
	code := `import './missing\'`
	filename := tspath.ResolvePath(root.Dir, "test.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{filename: code})
	// RuleTester rejects syntax errors; retain this partial editor input in a Program.
	p, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		Host: utils.CreateCompilerHost(root.Dir, fs), CompilerOptions: &core.CompilerOptions{},
		RootFileNames: []string{filename}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: p, File: filename,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: FileExtensionInImportRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners { return FileExtensionInImportRule.Run(ctx, nil) },
			}}
		},
		Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
	})
	if len(diagnostics) != 1 || diagnostics[0].Message.Id != "requireExt" || diagnostics[0].Message.Description != "require file extension '.js'." || diagnostics[0].Range != core.NewTextRange(7, len(code)) || diagnostics[0].FixesPtr != nil {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}

func TestFileExtensionInImportSymlinks(t *testing.T) {
	directory := t.TempDir()
	for name, content := range map[string]string{
		"tsconfig.json": `{"compilerOptions":{"allowJs":true}}`,
		"a.js":          "export {};",
		"target.js":     "export {};",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for name, target := range map[string]string{"only.txt": "missing", "linked.js": "target.js"} {
		if err := os.Symlink(target, filepath.Join(directory, name)); err != nil {
			if filepath.Separator == '\\' {
				t.Skipf("symlinks unavailable: %v", err)
			}
			t.Fatal(err)
		}
	}
	root := rule_tester.Root{Dir: tspath.NormalizePath(directory), FS: bundled.WrapFS(osvfs.FS())}
	t.Run("symlinks", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, nil, []rule_tester.InvalidTestCase{{
			Code: `import './only.txt'`, FileName: "test.js", Options: []any{"never"}, Output: []string{`import './only'`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.txt'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}},
		}, {
			Code: `import './linked.js'`, FileName: "test.js", Options: []any{"never"}, Output: []string{`import './linked'`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}},
		}})
	})
	t.Run("literal backslash directory", func(t *testing.T) {
		if filepath.Separator != '/' {
			t.Skip("Backslash is a literal filename character only on POSIX hosts.")
		}
		folder := filepath.Join(directory, `a\b`)
		if err := os.Mkdir(folder, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "index.js"), []byte("export {};"), 0o600); err != nil {
			t.Fatal(err)
		}
		// Documented difference: no directory expansion for literal backslashes.
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule,
			[]rule_tester.ValidTestCase{{Code: `import './a\\b'`, FileName: "test.js"}}, nil)
	})
	for _, ambiguous := range []bool{false, true} {
		output := []string{`import './a'`}
		if ambiguous {
			if err := os.Symlink("missing", filepath.Join(directory, "a.json")); err != nil {
				t.Fatal(err)
			}
			output = nil
		}
		// Finish each parallel tester group before changing the filesystem.
		// A fresh Program must see a newly added entry, even a broken symlink.
		t.Run(strconv.FormatBool(ambiguous), func(t *testing.T) {
			t.Run("new generation", func(t *testing.T) {
				rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, nil, []rule_tester.InvalidTestCase{{
					Code: `import './a.js'`, FileName: "test.js", Options: []any{"never"}, Output: output,
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}},
				}})
			})
		})
	}
}

func TestFileExtensionInImportEditDemand(t *testing.T) {
	for _, test := range []struct {
		code, style, message string
		wantFix              bool
	}{
		{`import './a'`, "always", "requireExt", true},
		{`import './my-folder'`, "always", "requireExt", true},
		{`import './a.js'`, "never", "forbidExt", true},
		{`import './multi.js'`, "never", "forbidExt", false},
		{`import './\u0061.js'`, "never", "forbidExt", true},
		{`import './a.\x6as'`, "never", "forbidExt", true},
	} {
		t.Run(test.code, func(t *testing.T) {
			root := extensionRoot(t, "testdata/upstream.txtar")
			compilerProgram, file, err := rule_tester.NewProgramHelper(root).CreateTestProgram(test.code, "input.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			p := lintprogram.NewFromCompiler(compilerProgram)
			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: p, File: file.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: FileExtensionInImportRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return FileExtensionInImportRule.Run(ctx, []any{test.style})
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				return diagnostics
			}
			all := run(rule.EditDemandAll)
			if len(all) != 1 || all[0].Message.Id != test.message || (all[0].FixesPtr != nil) != test.wantFix {
				t.Fatalf("unexpected diagnostics: %#v", all)
			}
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				got := run(demand)
				if len(got) != 1 {
					t.Fatalf("demand %d: %d diagnostics", demand, len(got))
				}
				if got[0].Range != all[0].Range || !reflect.DeepEqual(got[0].Message, all[0].Message) || got[0].Severity != all[0].Severity || got[0].RuleName != all[0].RuleName {
					t.Fatalf("demand %d changed diagnostic identity", demand)
				}
				if got[0].Suggestions != nil {
					t.Fatal("unexpected suggestion")
				}
				if demand&rule.EditDemandAutofix != 0 {
					if !reflect.DeepEqual(got[0].FixesPtr, all[0].FixesPtr) {
						t.Fatalf("demand %d changed fixes", demand)
					}
				} else if got[0].FixesPtr != nil {
					t.Fatalf("demand %d produced a fix", demand)
				}
			}
		})
	}
	// Source services are required; do not fall back to the host filesystem.
	if got := FileExtensionInImportRule.Run(rule.RuleContext{}, nil); got != nil {
		t.Fatal("expected no listeners without a Program")
	}
}

func TestFileExtensionInImportEscapedPaths(t *testing.T) {
	root := extensionRoot(t, "testdata/upstream.txtar", "testdata/extras.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, nil, []rule_tester.InvalidTestCase{
		{Code: "import '.\\/a.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import '.\\/a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		{Code: "import '\\u002e/a\\u002ejs'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import '\\u002e/a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
		{Code: "import './\\u{61}.\\u006a\\u0073'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './\\u{61}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 31}}},
		{Code: "import './a\\x2e\\x6a\\x73'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
		{Code: "import './\\uD83D\\uDE00.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './\\uD83D\\uDE00'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
		{Code: "import './\\u{1F600}.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './\\u{1F600}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		{Code: "import './😀.\\x6as'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './😀'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		{Code: "const label = '😀'; import './café.\\u006As';", FileName: "test.js", Options: []any{"never"}, Output: []string{"const label = '😀'; import './café';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 44}}},
		{Code: "import './a\\\n.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a\\\n'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 5}}},
		{Code: "import './a.\\\r\njs'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 4}}},
		{Code: "import './a.j\\\ns'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 3}}},
		{Code: "import './a.js\\\r\n'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a\\\r\n'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 2}}},
		{Code: "import './a.\\\u2028js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 4}}},
		{Code: "import './a.j\\\u2029s'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 3}}},
		{Code: "import './\\u0061.\\x6as!loader.js'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './\\u0061!loader.js'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 34}}},
		{Code: "import './a.\\x6as\\u0021loader'", FileName: "test.js", Options: []any{"never"}, Output: []string{"import './a\\u0021loader'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 31}}},
		{Code: "import( /* before */ ('./\\u0061.js') /* after */ );", FileName: "test.js", Options: []any{"never"}, Output: []string{"import( /* before */ ('./\\u0061') /* after */ );"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 36}}},
		{Code: "export * from \"./\\u0061.\\x6as\";", FileName: "test.js", Options: []any{"never"}, Output: []string{"export * from \"./\\u0061\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 31}}},
		{Code: "export type {T} from './\\u0064.js';", FileName: "test.ts", Options: []any{"never"}, Output: []string{"export type {T} from './\\u0064';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forbidExt", Message: "forbid file extension '.js'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 35}}},
	})
}

func TestFileExtensionInImportResourceSuffixes(t *testing.T) {
	root := extensionRoot(t, "testdata/upstream.txtar", "testdata/extras.txtar")
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, suffix := range []string{
		"!loader.js", "?raw", "#part", "?path=/other.js#part",
		"!loader?raw#part", "?raw!loader.js#part",
		`\u0021loader`, `\x3f` + "raw", `\u{23}part`,
	} {
		for _, test := range []struct {
			path, fixed, file, style string
		}{
			{"./a", "./a.js", "test.js", "always"},
			{"./a.js", "./a", "test.js", "never"},
			{"./my-folder", "./my-folder/index.js", "test.js", "always"},
			{"./my-folder/", "./my-folder/index.js", "test.js", "always"},
			{`./\u0061`, `./\u0061.js`, "test.js", "always"},
			{`./a.\x6as`, "./a", "test.js", "never"},
			{`./\ud83d\ude00`, `./\ud83d\ude00.js`, "test.js", "always"},
			{"./d", "./d.js", "test.ts", "always"},
			{"./d.js", "./d", "test.ts", "never"},
		} {
			code := "import '" + test.path + suffix + "'"
			fixed := "import '" + test.fixed + suffix + "'"
			messageID, message := "requireExt", "require file extension '.js'."
			if test.style == "never" {
				messageID, message = "forbidExt", "forbid file extension '.js'."
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code, FileName: test.file, Options: []any{test.style}, Output: []string{fixed},
				// All source spellings in this matrix are ASCII and single-line.
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 8, EndLine: 1, EndColumn: len(code) + 1}},
			})
			// Besides checking the single fix round, require a clean second lint.
			valid = append(valid, rule_tester.ValidTestCase{Code: fixed, FileName: test.file, Options: []any{test.style}})
		}
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &FileExtensionInImportRule, valid, invalid)
}
