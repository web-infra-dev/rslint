package no_unresolved_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_unresolved"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/embedfs"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func unresolvedRoot(t *testing.T, insensitive bool) embedfs.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = "/no-unresolved"
	files := map[string]string{}
	for _, fixture := range []string{"upstream", "extras"} {
		archive := txtarfs.MustParseFile(t, "testdata/"+fixture+".txtar")
		names, err := archive.FileNames("")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			data, err := archive.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			files[tspath.ResolvePath(root.Dir, name)] = string(data)
		}
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	if insensitive {
		paths := map[string]string{}
		for file := range files {
			for file != "" && file != "/" {
				paths[strings.ToLower(file)] = file
				file = tspath.GetDirectoryPath(file)
			}
		}
		root.FS = caseInsensitiveFS{FS: root.FS, paths: paths}
	}
	return root
}

// Simulate both filesystem modes on every host, preserving directory spelling.
type caseInsensitiveFS struct {
	vfs.FS
	paths map[string]string
}

func (fs caseInsensitiveFS) actual(path string) string {
	if actual, ok := fs.paths[strings.ToLower(path)]; ok {
		return actual
	}
	return path
}
func (fs caseInsensitiveFS) UseCaseSensitiveFileNames() bool { return false }
func (fs caseInsensitiveFS) FileExists(path string) bool     { return fs.FS.FileExists(fs.actual(path)) }
func (fs caseInsensitiveFS) DirectoryExists(path string) bool {
	return fs.FS.DirectoryExists(fs.actual(path))
}
func (fs caseInsensitiveFS) ReadFile(path string) (string, bool) {
	return fs.FS.ReadFile(fs.actual(path))
}
func (fs caseInsensitiveFS) GetAccessibleEntries(path string) vfs.Entries {
	return fs.FS.GetAccessibleEntries(fs.actual(path))
}
func (fs caseInsensitiveFS) Realpath(path string) string { return fs.FS.Realpath(fs.actual(path)) }

func unresolvedErrors(code string, names ...string) []rule_tester.InvalidTestCaseError {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(names))
	from := 0
	for _, name := range names {
		start := strings.Index(code[from:], "'"+name+"'") + from
		if start < from {
			panic("missing expected literal: " + name)
		}
		prefix := code[:start]
		lineStart := strings.LastIndexByte(prefix, '\n') + 1
		line := strings.Count(prefix, "\n") + 1
		column := ecmascript.StringCodeUnitCount(prefix[lineStart:]) + 1
		errors = append(errors, rule_tester.InvalidTestCaseError{
			MessageId: "", Message: "Unable to resolve path to module '" + name + "'.",
			Line: line, Column: column, EndLine: line, EndColumn: column + ecmascript.StringCodeUnitCount(name) + 2,
		})
		from = start + len(name) + 2
	}
	return errors
}

func TestNoUnresolvedExtras(t *testing.T) {
	root := unresolvedRoot(t, false)
	valid := []rule_tester.ValidTestCase{
		{Code: `import './bar';`, Options: map[string]any{}},
		{Code: `import './bar';`, Options: map[string]any{"esmodule": true, "commonjs": false, "amd": false, "caseSensitive": true, "caseSensitiveStrict": false}},
		{Code: `import 'missing'; export * from 'missing'; import('missing');`, Options: map[string]any{"esmodule": false}},
		{Code: "require(`missing`); import(`missing`); import(`${name}`); require('missing' as string); require!( 'missing' ); (require as any)('missing');", Options: map[string]any{"commonjs": true}},
		{Code: `obj.require('missing'); obj?.require('missing'); require.resolve('missing'); new require('missing'); import value = require('missing'); type T = import('missing').T;`, Options: map[string]any{"commonjs": true}},
		{Code: `export type * from 'missing'; export type * as ns from 'missing'; import type T from 'missing';`},
		{FileName: "script.js", Code: "/** @import { Value } from 'missing' */\n/** @type {import('missing').Value} */\nlet value;"},
		{Code: `define('named', ['missing'], callback); define(['missing']); define([...list, 0, , null], callback);`, Options: map[string]any{"amd": true}},
		{Code: `import 'virtual/generated';`, Options: map[string]any{"ignore": []any{`(?<=virtual/)generated$`}}},
		{Code: `require('ignored'); define(['ignored'], callback); import('ignored');`, Options: map[string]any{"commonjs": true, "amd": true, "ignore": []any{"^ignored$"}}},
		{Code: `import 'node:fs/promises'; import '_http_agent'; import 'module';`},
		{Code: `import 'exact/subpath';`, Settings: map[string]any{"import/core-modules": []any{"exact/subpath"}, "import/resolver": "missing"}},
		{Code: `import 'entry-fallback'; import 'legacy-exports'; import './runtime.node'; import './unparsed.css';`},
		{Code: `import './only-types';`, Settings: map[string]any{"import/resolver": "typescript"}},
		// The native TypeScript adapter uses the bound project, not resolver-plugin options.
		{Code: `import './only-types';`, Settings: map[string]any{"import/resolver": map[string]any{"typescript": map[string]any{"project": "missing.json", "alwaysTryTypes": false}}}},
		{Code: `import './bar';`, Settings: map[string]any{"import/resolver": []any{"typescript", "node"}}},
		{Code: `import './bar';`, Settings: map[string]any{"import/resolver": []any{nil, "node"}}},
		{Code: `import './bar';`, Settings: map[string]any{"import/resolver": []any{map[string]any{"node": map[string]any{"extensions": []any{}}}, "missing-resolver", "node"}}},
		{Code: `import 'jsx-module/foo';`, Settings: map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": []any{".jsx"}}}}},
		{Code: `import 'src-bar';`, Settings: map[string]any{"import/resolver": map[string]any{"node": map[string]any{"moduleDirectory": "src-root"}}}},
		{Code: `import './bar.js';`, Settings: map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": []any{}}}}},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: `import '\u006dissing';`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unable to resolve path to module 'missing'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
		{Code: "import 'miss\\\ning';", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Unable to resolve path to module 'missing'.", Line: 1, Column: 8, EndLine: 2, EndColumn: 5}}},
	}
	add := func(code string, options any, settings map[string]any, names ...string) {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: options, Settings: settings, Errors: unresolvedErrors(code, names...)})
	}
	add(`(require)(('missing')); require?.('optional');`, map[string]any{"commonjs": true}, nil, "missing", "optional")
	add(`/** @type {any} */ (require)('cast-callee'); require(/** @type {string} */ ('cast-source'));`, map[string]any{"commonjs": true}, nil, "cast-callee", "cast-source")
	invalid[len(invalid)-1].FileName = "script.js"
	add(`define(/** @type {string[]} */ (['cast-array']), callback);`, map[string]any{"amd": true}, nil, "cast-array")
	invalid[len(invalid)-1].FileName = "script.js"
	add(`function f(require) { require('shadowed'); }`, map[string]any{"commonjs": true}, nil, "shadowed")
	add(`(define)((['missing', 'require', 'exports', 'module']), callback);`, map[string]any{"amd": true}, nil, "missing")
	add(`import { type T } from 'inline'; export { type T } from 'inline-export';`, nil, nil, "inline", "inline-export")
	add(`import(('dynamic')); import('attributes', { with: { type: 'json' } });`, nil, nil, "dynamic", "attributes")
	add(`import ''; export {} from ''; import('');`, nil, nil, "", "", "")
	add("const emoji = '😀'; import 'missing';\nexport * from\n  '另一个';", nil, nil, "missing", "另一个")
	add(`const el = <C value={import('jsx')} />;`, nil, nil, "jsx")
	invalid[len(invalid)-1].FileName = "element.tsx"
	add(`import 'ignored-by-setting';`, nil, map[string]any{"import/ignore": []any{".*"}}, "ignored-by-setting")
	add(`import './only-types'; import 'only-types-package'; import 'fs/not-a-builtin';`, nil, nil, "./only-types", "only-types-package", "fs/not-a-builtin")
	add(`import 'exact/subpath';`, nil, map[string]any{"import/core-modules": []any{"exact"}}, "exact/subpath")
	add(`import './bar';`, nil, map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": []any{}}}}, "./bar")
	add(`import './bar';`, nil, map[string]any{"import/resolver": []any{"node", map[string]any{"node": map[string]any{"extensions": []any{}}}}}, "./bar")
	add(`import './bar.js?raw';`, nil, nil, "./bar.js?raw")
	add(`import './jsx/MyUncoolComponent.jsx';`, map[string]any{"caseSensitive": false}, nil, "./jsx/MyUncoolComponent.jsx")
	add(`import 'missing'; require('checked');`, map[string]any{"esmodule": false, "commonjs": true}, nil, "checked")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule, valid, invalid)
}

func TestNoUnresolvedIgnoreValidation(t *testing.T) {
	if err := no_unresolved.NoUnresolvedRule.Schema.Validate([]any{map[string]any{"ignore": []any{"["}}}); err == nil {
		t.Fatal("an invalid ignore expression must fail configuration validation")
	}
}

func TestNoUnresolvedCaseCache(t *testing.T) {
	root := unresolvedRoot(t, true)
	file := tspath.ResolvePath(root.Dir, "input.js")
	code := `import '/NO-UNRESOLVED/bar.js';`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: code})
	p, err := program.NewFromRoots(program.RootOptions{
		Host: utils.CreateCompilerHost(root.Dir, fs), CompilerOptions: program.SourceOnlyCompilerOptions(),
		RootFileNames: []string{file}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Reuse the same Program: strictness and the process cwd must not share
	// cached results, and a failed check must not poison a later relaxed one.
	for _, test := range []struct {
		cwd    string
		strict bool
		want   int
	}{
		{root.Dir, false, 0}, {root.Dir, true, 1}, {"/", false, 1}, {root.Dir, false, 0},
	} {
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{SourceFile: p.GetSourceFile(file)}.WithProgram(p).
			WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory(test.cwd)).
			WithReporter(no_unresolved.NoUnresolvedRule.Name, rule.SeverityError, func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) })
		no_unresolved.NoUnresolvedRule.Run(ctx, []any{map[string]any{"caseSensitiveStrict": test.strict}})
		if len(diagnostics) != test.want {
			t.Fatalf("cwd=%q strict=%v: got %#v, want %d diagnostics", test.cwd, test.strict, diagnostics, test.want)
		}
		if test.want > 0 {
			d := diagnostics[0]
			if d.Message.Id != "" || d.Message.Description != "Casing of /NO-UNRESOLVED/bar.js does not match the underlying filesystem." || d.Range != core.NewTextRange(7, len(code)-1) || d.FixesPtr != nil || d.Suggestions != nil {
				t.Fatalf("unexpected casing diagnostic: %#v", d)
			}
		}
	}
}

func TestNoUnresolvedSourceOnly(t *testing.T) {
	root := unresolvedRoot(t, false)
	file := tspath.ResolvePath(root.Dir, "input.js")
	code := `import './malformed.js'; import './bar.json'; import 'missing';`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: code})
	p, err := program.NewFromRoots(program.RootOptions{
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{AllowJs: core.TSTrue},
		RootFileNames:   []string{file}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{SourceFile: p.GetSourceFile(file)}.WithProgram(p).WithReporter(
		no_unresolved.NoUnresolvedRule.Name, rule.SeverityError,
		func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
	)
	no_unresolved.NoUnresolvedRule.Run(ctx, nil)
	if len(diagnostics) != 1 || diagnostics[0].Message.Description != "Unable to resolve path to module 'missing'." {
		t.Fatalf("source-only diagnostics = %#v", diagnostics)
	}
	diagnostic := diagnostics[0]
	start := strings.Index(code, "'missing'")
	if diagnostic.Message.Id != "" || diagnostic.Range != core.NewTextRange(start, start+len("'missing'")) || diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
		t.Fatalf("source-only diagnostic range or artifacts = %#v", diagnostic)
	}
	if len(p.SourceFiles()) != 1 {
		t.Fatal("runtime resolution loaded dependency ASTs into the source-only Program")
	}
}

func TestNoUnresolvedSymlinkImporter(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/extras.txtar")
	root := tspath.NormalizePath(archive.Materialize(t, "symlink"))
	project := tspath.ResolvePath(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(tspath.ResolvePath(root, "real/package"), tspath.ResolvePath(project, "source")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation is unavailable: %v", err)
		}
		t.Fatal(err)
	}
	file := tspath.ResolvePath(project, "source/input.js")
	p, err := program.NewFromRoots(program.RootOptions{
		Host: utils.CreateCompilerHost(root, osvfs.FS()), CompilerOptions: program.SourceOnlyCompilerOptions(),
		RootFileNames: []string{file}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, preserve := range []bool{true, false} {
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile: p.GetSourceFile(file),
			Settings:   map[string]any{"import/resolver": map[string]any{"node": map[string]any{"preserveSymlinks": preserve}}},
		}.WithProgram(p).WithReporter(no_unresolved.NoUnresolvedRule.Name, rule.SeverityError, func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) })
		no_unresolved.NoUnresolvedRule.Run(ctx, nil)
		if preserve {
			if len(diagnostics) != 1 || diagnostics[0].Message.Description != "Unable to resolve path to module 'dependency'." || diagnostics[0].Message.Id != "" || diagnostics[0].Range != core.NewTextRange(7, 19) || diagnostics[0].FixesPtr != nil || diagnostics[0].Suggestions != nil {
				t.Fatalf("preserved symlink diagnostics = %#v", diagnostics)
			}
		} else if len(diagnostics) != 0 {
			t.Fatalf("real importer directory must find dependency: %#v", diagnostics)
		}
	}
}
