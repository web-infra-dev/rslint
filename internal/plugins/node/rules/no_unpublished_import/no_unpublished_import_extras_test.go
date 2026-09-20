package no_unpublished_import

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnpublishedImportExtras(t *testing.T) {
	rule_tester.RunRuleTester(unpublishedImportRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &NoUnpublishedImportRule,
		[]rule_tester.ValidTestCase{
			// Production dependency fields override development declarations; unknown packages belong to no-missing-import.
			{Code: "import 'runtime'; import 'peer'; import 'optional'; import 'unknown';", FileName: "public/src/input.js"},
			// Shared allowModules accepts node settings.
			{Code: "import 'dev/part';", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"allowModules": []any{"dev"}}}},
			// Legacy n settings take precedence over node settings.
			{Code: "import 'dev';", FileName: "public/src/input.js", Settings: map[string]any{"n": map[string]any{"allowModules": []any{"dev"}}, "node": map[string]any{"allowModules": []any{}}}},
			// AllowModules exempts virtual package roots.
			{Code: "import 'virtual:foo/entry';", FileName: "public/src/input.js", Options: []any{map[string]any{"allowModules": []any{"virtual:foo"}}}},
			// Imports with non-literal arguments, TypeScript wrappers, import types and require are outside this rule.
			{Code: "const name = 'dev'; import(name); import(`dev`); import('de' + 'v'); import('dev' as string); import('dev'!); type T = import('dev').T; import x = require('dev'); obj?.import('dev'); require('dev');", FileName: "public/src/input.ts"},
			// Private packages are ignored with explicit defaults.
			{Code: "import 'dev';", FileName: "private/input.js", Options: []any{map[string]any{"ignorePrivate": true}}},
			// Unpublished importers are not checked.
			{Code: "import 'dev'; import './hidden.js';", FileName: "public/src/hidden.js"},
			// Gitignore is used when npmignore and files are absent.
			{Code: "import 'dev';", FileName: "git/test.js"},
			// Target package metadata does not override the importing package publication list.
			{Code: "import './child/index.js';", FileName: "public/src/input.js"},
			// Package metadata stays published despite ignore rules.
			{Code: "import '../package.json';", FileName: "public/src/input.js"},
			// Local files use resolution and TypeScript extension substitution.
			{Code: "import './target.js';", FileName: "public/src/input.ts"},
			// A resolver alias changes a local target before publication checks.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"alias": map[string]any{"./hidden.js": "./public.js"}}}}},
			// Fallback redirects missing local targets before publication checks.
			{Code: "import './hidden-missing.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fallback": map[string]any{"./hidden-missing.js": "./public.js"}}}}},
			// Additional resolution paths are tried before the source directory.
			{Code: "import './hidden';", FileName: "public/src/input.js", Options: []any{map[string]any{"resolvePaths": []any{"public/dist"}}}},
			// Shared resolvePaths use the same resolution.
			{Code: "import './hidden';", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{"public/dist"}}}},
			// Shared tryExtensions can locate an included custom extension.
			{Code: "import './odd';", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".xyz"}}}},
			// Target conversion can include an otherwise unpublished file.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/hidden.js": []any{"^src/", "dist/"}}}}},
			// A conversion to the package directory leaves no target to check.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/hidden.js": []any{".*", ""}}}}},
			// Names starting with two dots are ordinary files, not parent paths.
			{Code: "import './..hidden.js';", FileName: "dot/src/index.js"},
			// Invalid shared conversion patterns skip checking instead of throwing.
			{Code: "import 'dev';", FileName: "public/src/input.js", Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"**": []any{"[", "x"}}}}}},
		[]rule_tester.InvalidTestCase{
			// Literal coercion preserves JS number rounding, BigInt values and regexp spelling.
			{Code: "import(9007199254740993); import(9007199254740993n); import(true); import(false); import(null); import(/hidden/);", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notPublished", Message: `"9007199254740992" is not published.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 24},
				{MessageId: "notPublished", Message: `"9007199254740993" is not published.`, Line: 1, Column: 34, EndLine: 1, EndColumn: 51},
				{MessageId: "notPublished", Message: `"true" is not published.`, Line: 1, Column: 61, EndLine: 1, EndColumn: 65},
				{MessageId: "notPublished", Message: `"false" is not published.`, Line: 1, Column: 75, EndLine: 1, EndColumn: 80},
				{MessageId: "notPublished", Message: `"null" is not published.`, Line: 1, Column: 90, EndLine: 1, EndColumn: 94},
				{MessageId: "notPublished", Message: `"/hidden/" is not published.`, Line: 1, Column: 104, EndLine: 1, EndColumn: 112},
			}},
			// Module bodies and JSX expressions still contain import/export sources.
			{Code: "declare module 'nested' { export {Value} from 'dev'; }", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: `"dev" is not published.`, Line: 1, Column: 47, EndLine: 1, EndColumn: 52}}},
			{Code: "const View = () => <div>{import('dev')}</div>;", FileName: "public/src/input.tsx", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: `"dev" is not published.`, Line: 1, Column: 33, EndLine: 1, EndColumn: 38}}},
			// Membership, not the truthiness of a dependency version, determines publication.
			{Code: "import 'null-version'; import 'empty-version';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: `"empty-version" is not published.`, Line: 1, Column: 31, EndLine: 1, EndColumn: 46}}},
			// A development-only dependency is checked even when it is not installed.
			{Code: "import 'dev';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// Package paths report the package root, including scoped and loader imports.
			{Code: "import '@scope/dev/internal'; import 'dev/part!raw';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"@scope/dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 38, EndLine: 1, EndColumn: 52}}},
			// The package name and workspace dependencies are not publication exemptions.
			{Code: "import 'self';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"self\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
			// Workspace parents do not supply published dependencies.
			{Code: "import 'dev';", FileName: "workspace/child/index.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// An explicit empty allowModules overrides shared values.
			{Code: "import 'dev';", FileName: "public/src/input.js", Options: []any{map[string]any{"allowModules": []any{}}}, Settings: map[string]any{"node": map[string]any{"allowModules": []any{"dev"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// Allowing a package does not allow an unrelated scoped package.
			{Code: "import '@scope/dev';", FileName: "public/src/input.js", Options: []any{map[string]any{"allowModules": []any{"dev"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"@scope/dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// Type import defaults and re-export semantics.
			{Code: "import type {Value} from 'dev'; export type {Value} from 'dev'; export type * from 'dev'; import {type Value as V} from 'dev';", FileName: "public/src/input.ts", Options: []any{map[string]any{"ignoreTypeImport": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 58, EndLine: 1, EndColumn: 63}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 84, EndLine: 1, EndColumn: 89}, {MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 121, EndLine: 1, EndColumn: 126}}},
			// Empty options preserve both runtime defaults.
			{Code: "import type {Value} from 'dev';", FileName: "public/src/input.ts", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 26, EndLine: 1, EndColumn: 31}}},
			// Named, star, namespace and dynamic imports report source literals.
			{Code: "export {value} from './hidden.js'; export * from './hidden.js'; export * as ns from './hidden.js'; import((('./hidden.js')));", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 21, EndLine: 1, EndColumn: 34}, {MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 50, EndLine: 1, EndColumn: 63}, {MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 85, EndLine: 1, EndColumn: 98}, {MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 109, EndLine: 1, EndColumn: 122}}},
			// Import attributes and an escaped literal retain the full literal range.
			{Code: "import './hidden\\u002Ejs' with {type: 'json'};", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			// UTF-16 columns and multiline dynamic imports.
			{Code: "const emoji = '😀'; import('dev');\nimport(\n  './hidden.js'\n);", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33}, {MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 3, Column: 3, EndLine: 3, EndColumn: 16}}},
			// Builtins, side-effect-free exports and JSDoc-wrapped dynamic imports.
			{Code: "import 'node:fs'; import '_http_agent'; const x=1; export {x}; import(/** @type {string} */ ('dev'));", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 94, EndLine: 1, EndColumn: 99}}},
			// Private packages can opt into checking.
			{Code: "import 'dev';", FileName: "private/input.js", Options: []any{map[string]any{"ignorePrivate": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// Only the boolean true marks a package private.
			{Code: "import 'dev';", FileName: "string-private/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// Npmignore takes precedence over gitignore.
			{Code: "import 'dev';", FileName: "npm/test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// The main entry is published even with an empty files list.
			{Code: "import 'dev';", FileName: "main/index.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// Missing relative targets retain their lexical path.
			{Code: "import './hidden-missing.js';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden-missing.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},
			// Fallback cannot replace an existing unpublished target.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"resolverConfig": map[string]any{"fallback": map[string]any{"./hidden.js": "./public.js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// Package import maps resolve to the publication target.
			{Code: "import '#public'; import '#hidden';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"#hidden\" is not published.", Line: 1, Column: 26, EndLine: 1, EndColumn: 35}}},
			// Type-only imports activate the types export condition.
			{Code: "import type {Value} from '#types'; import '#types';", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"#types\" is not published.", Line: 1, Column: 26, EndLine: 1, EndColumn: 34}}},
			// TypeScript path aliases do not exempt development dependencies.
			{Code: "import '@alias/hidden';", FileName: "public/src/input.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"@alias/hidden\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
			// Explicit resolver extensions override tryExtensions.
			{Code: "import './hidden';", FileName: "public/src/input.js", Options: []any{map[string]any{"tryExtensions": []any{".xyz"}, "resolverConfig": map[string]any{"extensions": []any{".js"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// URL targets are subject to the upstream publication check.
			{Code: "import 'https://example.com/module.js'; import 'data:text/javascript,export default 1';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"https://example.com/module.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 39}, {MessageId: "notPublished", Message: "\"data:text/javascript,export default 1\" is not published.", Line: 1, Column: 48, EndLine: 1, EndColumn: 87}}},
			// Unresolved import maps use the working directory as their fallback.
			{Code: "import '#missing';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"#missing\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// Source conversion can make an excluded importer published.
			{Code: "import 'dev';", FileName: "public/input.js", Options: []any{map[string]any{"convertPath": map[string]any{"input.js": []any{"^", "dist/"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// Conversion excludes preserve the target path.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"**/hidden.js"}, "replace": []any{"^src/", "dist/"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// The first matching conversion wins.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "replace": []any{"^src/", "src/"}}, map[string]any{"include": []any{"src/**"}, "replace": []any{"^src/", "dist/"}}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// Rule conversion takes precedence over settings.
			{Code: "import './hidden.js';", FileName: "public/src/input.js", Options: []any{map[string]any{"convertPath": map[string]any{}}}, Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"src/hidden.js": []any{"^src/", "dist/"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./hidden.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// Shared publication policy respects nested ignore files; upstream ignores them.
			{Code: "import './nested/ignored.js';", FileName: "public/src/input.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"./nested/ignored.js\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},
			// Root README metadata is published; upstream can misclassify it.
			{Code: "import 'dev';", FileName: "metadata/README.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notPublished", Message: "\"dev\" is not published.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}}},
	)
}

func TestNoUnpublishedImportProcessDirectory(t *testing.T) {
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "process-directory")
	fileName := tspath.ResolvePath(directory, "input.ts")
	const code = "import '#missing';"
	fs := utils.NewOverlayVFS(base.FS, map[string]string{
		fileName: code,
		tspath.ResolvePath(directory, "package.json"): `{}`,
	})
	p, err := program.NewFromRoots(program.RootOptions{
		Host: utils.CreateCompilerHost(directory, fs), CompilerOptions: &core.CompilerOptions{},
		RootFileNames: []string{fileName}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, cwd string
		want      int
	}{
		{"context absent uses Program directory", "", 0},
		{"process owns package", directory, 0},
		{"process outside package", base.Dir, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			var diagnostics []rule.RuleDiagnostic
			ctx := (rule.RuleContext{SourceFile: p.GetSourceFile(fileName), Settings: map[string]any{"cwd": directory}}).
				WithProgram(p).WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory(test.cwd)).
				WithReporter(NoUnpublishedImportRule.Name, rule.SeverityError, func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) })
			listeners := NoUnpublishedImportRule.Run(ctx, nil)
			listeners[rule.ListenerOnExit(ast.KindEndOfFile)](ctx.SourceFile.AsNode())
			if len(diagnostics) != test.want {
				t.Fatalf("got %d diagnostics, want %d", len(diagnostics), test.want)
			}
			if test.want != 0 {
				diagnostic := diagnostics[0]
				if diagnostic.Message.Id != "notPublished" || diagnostic.Message.Description != `"#missing" is not published.` || diagnostic.Range.Pos() != 7 || diagnostic.Range.End() != 17 {
					t.Fatalf("unexpected diagnostic: %#v", diagnostic)
				}
			}
		})
	}
}

// The upstream docs show convertPath: null, but the upstream schema rejects it.
func TestNoUnpublishedImportSchema(t *testing.T) {
	for _, options := range []any{map[string]any{"convertPath": nil}, map[string]any{"ignorePrivate": "true"}, map[string]any{"ignoreTypeImport": 1}, map[string]any{"allowModules": []any{"dev", "dev"}}, map[string]any{"tryExtensions": []any{"js"}}, map[string]any{"unknown": true}} {
		if err := NoUnpublishedImportRule.Schema.Validate([]any{options}); err == nil {
			t.Fatalf("expected invalid options: %#v", options)
		}
	}
}
