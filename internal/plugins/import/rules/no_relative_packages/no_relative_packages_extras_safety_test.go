package no_relative_packages_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_relative_packages"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func safetyRoot(t *testing.T, id string, link, shadow bool) rule_tester.Root {
	t.Helper()
	root := rule_tester.Root{Dir: tspath.NormalizePath(t.TempDir()), FS: bundled.WrapFS(osvfs.FS())}
	archive := txtarfs.MustParseFile(t, "testdata/safety.txtar")
	names, err := archive.FileNames(id)
	if err != nil || len(names) == 0 {
		t.Fatalf("empty or invalid %s fixtures: %v", id, err)
	}
	for _, name := range names {
		data, err := archive.ReadFile(id + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		file := tspath.ResolvePath(root.Dir, name)
		if err := os.MkdirAll(tspath.GetDirectoryPath(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(tspath.ResolvePath(root.Dir, "tsconfig.json"), []byte(`{"compilerOptions":{"allowJs":true,"noLib":true,"jsx":"preserve"},"include":["packages/**/*"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if link {
		modules := tspath.ResolvePath(root.Dir, "node_modules")
		if err := os.MkdirAll(modules, 0o755); err != nil {
			t.Fatal(err)
		}
		destination := tspath.ResolvePath(root.Dir, "packages/b")
		if shadow {
			destination = tspath.ResolvePath(root.Dir, "installed-b")
		}
		if err := os.Symlink(destination, tspath.ResolvePath(modules, "b")); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func safePackageCase(file, code, imported, candidate string, options any) rule_tester.InvalidTestCase {
	literal := fmt.Sprintf("%q", imported)
	item := packageCase(file, code, literal, imported, candidate, options)
	if candidate != "" {
		item.Errors[0].Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "usePackageImport", Output: strings.Replace(code, literal, fmt.Sprintf("%q", candidate), 1),
		}}
	}
	return item
}

// These paired cases prove that the previous upstream fixes were unsafe and
// that suggestions retain the authored request only when Node resolves it to
// the same file. No case permits an autofix, including the positive controls.
func TestNoRelativePackagesCandidateSafety(t *testing.T) {
	for _, tc := range []struct {
		id, description, file, code, imported, candidate string
		link, shadow                                     bool
	}{
		{"C01", "root index control", "packages/a/source.cjs", "const value = require(\"../b\");", "../b", "b", true, false},
		{"C02", "explicit subfile control", "packages/a/source.cjs", "const value = require(\"../b/sub/tool.js\");", "../b/sub/tool.js", "b/sub/tool.js", true, false},
		{"C03", "extension inference control", "packages/a/source.cjs", "const value = require(\"../b/sub/tool\");", "../b/sub/tool", "b/sub/tool", true, false},
		{"D01", "directory index duplicates basename", "packages/a/source.cjs", "const value = require(\"../b/sub\");", "../b/sub", "b/sub", true, false},
		{"D02", "duplicate directory exists but is a different module", "packages/a/source.cjs", "const value = require(\"../b/sub\");", "../b/sub", "b/sub", true, false},
		{"D03", "unnamed subdirectory main duplicates basename", "packages/a/source.cjs", "const value = require(\"../b/sub\");", "../b/sub", "b/sub", true, false},
		{"D04", "explicit directory index is safe control", "packages/a/source.cjs", "const value = require(\"../b/sub/index.js\");", "../b/sub/index.js", "b/sub/index.js", true, false},
		{"M01", "package root main in dist but no dist index", "packages/a/source.cjs", "const value = require(\"../b\");", "../b", "b", true, false},
		{"M02", "package root main in dist chooses wrong dist index", "packages/a/source.cjs", "const value = require(\"../b\");", "../b", "b", true, false},
		{"M03", "explicit main file remains safe", "packages/a/source.cjs", "const value = require(\"../b/dist/entry.js\");", "../b/dist/entry.js", "b/dist/entry.js", true, false},
		{"M04", "root main filename can be non-index", "packages/a/source.cjs", "const value = require(\"../b\");", "../b", "b", true, false},
		{"N01", "file stem equals root folder basename", "packages/a/source.cjs", "const value = require(\"../b/b\");", "../b/b", "b/b", true, false},
		{"N02", "explicit extension avoids root-name collision", "packages/a/source.cjs", "const value = require(\"../b/b.js\");", "../b/b.js", "b/b.js", true, false},
		{"P01", "literal parent specifier turns into current directory", "packages/b/a/source.cjs", "const value = require(\"..\");", "..", "b", true, false},
		{"X01", "exports alias makes physical path private", "packages/a/source.cjs", "const value = require(\"../b/hidden.js\");", "../b/hidden.js", "", true, false},
		{"X02", "same exports alias with native ESM", "packages/a/source.mjs", "import value from \"../b/hidden.js\";", "../b/hidden.js", "", true, false},
		{"X03", "no public path for the file", "packages/a/source.cjs", "const value = require(\"../b/hidden.js\");", "../b/hidden.js", "", true, false},
		{"X04", "exports same spelling maps to another file", "packages/a/source.cjs", "const value = require(\"../b/hidden.js\");", "../b/hidden.js", "", true, false},
		{"X05", "explicit public file mapping is safe", "packages/a/source.cjs", "const value = require(\"../b/hidden.js\");", "../b/hidden.js", "b/hidden.js", true, false},
		{"X06", "require condition selects original file", "packages/a/source.cjs", "const value = require(\"../b/hidden.js\");", "../b/hidden.js", "", true, false},
		{"X07", "import condition selects different file", "packages/a/source.mjs", "import value from \"../b/hidden.js\";", "../b/hidden.js", "", true, false},
		{"X08", "exports pattern uses public alias", "packages/a/source.cjs", "const value = require(\"../b/lib/tool.js\");", "../b/lib/tool.js", "", true, false},
		{"L01", "package name is not linked or installed", "packages/a/source.cjs", "const value = require(\"../b/tool.js\");", "../b/tool.js", "", false, false},
		{"L02", "package name points to another installed version", "packages/a/source.cjs", "const value = require(\"../b/tool.js\");", "../b/tool.js", "", true, true},
		{"R01", "module field chosen by lint resolver differs from Node main", "packages/a/source.cjs", "const value = require(\"../b\");", "../b", "b", true, false},
		{"E01", "native ESM relative directory input already fails", "packages/a/source.mjs", "import value from \"../b/sub\";", "../b/sub", "", true, false},
		{"E02", "native ESM extensionless input already fails", "packages/a/source.mjs", "import value from \"../b/tool\";", "../b/tool", "", true, false},
		{"E03", "native ESM explicit file without exports is safe", "packages/a/source.mjs", "import value from \"../b/tool.js\";", "../b/tool.js", "b/tool.js", true, false},
	} {
		t.Run(tc.id+"/"+tc.description, func(t *testing.T) {
			root := safetyRoot(t, tc.id, tc.link, tc.shadow)
			options := map[string]any{"commonjs": true}
			item := safePackageCase(tc.file, tc.code, tc.imported, tc.candidate, options)
			rule_tester.RunRuleTester(root, "tsconfig.json", t, &target.NoRelativePackagesRule, nil, []rule_tester.InvalidTestCase{item})
		})
	}
}

func TestNoRelativePackagesSuggestionBoundaries(t *testing.T) {
	// These environments can still diagnose a cross-package import, but cannot
	// establish default Node runtime equivalence for an edit.
	for _, tc := range []struct {
		code, file, imported string
		settings             map[string]any
		options              any
	}{
		{`const value = require("../b");`, "source.cjs", "../b", map[string]any{"import/resolver": map[string]any{"node": map[string]any{"preserveSymlinks": true}}}, map[string]any{"commonjs": true}},
		{`const value = require("../b");`, "source.cjs", "../b", map[string]any{"import/resolver": []any{"node", "typescript"}}, map[string]any{"commonjs": true}},
		{`const value = require("../b");`, "source.cjs", "../b", map[string]any{"import/resolve": map[string]any{"extensions": []any{".js"}}}, map[string]any{"commonjs": true}},
		{`const value = require("../b");`, "source.ts", "../b", nil, map[string]any{"commonjs": true}},
		{`function load(require) { return require("../b"); }`, "source.cjs", "../b", nil, map[string]any{"commonjs": true}},
		{`define(["../b"], () => {});`, "source.js", "../b", nil, map[string]any{"amd": true}},
		{`import("../b/index.js", options);`, "source.mjs", "../b/index.js", nil, nil},
	} {
		t.Run(tc.code+fmt.Sprint(tc.settings)+tc.file, func(t *testing.T) {
			root := safetyRoot(t, "C01", true, false)
			item := safePackageCase("packages/a/"+tc.file, tc.code, tc.imported, "", tc.options)
			item.Settings = tc.settings
			rule_tester.RunRuleTester(root, "tsconfig.json", t, &target.NoRelativePackagesRule, nil, []rule_tester.InvalidTestCase{item})
		})
	}
}

func TestNoRelativePackagesVerifiedEditDemand(t *testing.T) {
	for _, tc := range []struct{ id, file, code, imported, candidate string }{
		{"C01", "packages/a/source.cjs", `const value = require("../b");`, "../b", "b"},
		{"E03", "packages/a/source.mjs", `import value from "../b/tool.js";`, "../b/tool.js", "b/tool.js"},
	} {
		root := safetyRoot(t, tc.id, true, false)
		file := tspath.ResolvePath(root.Dir, tc.file)
		p, err := program.NewFromRoots(program.RootOptions{Host: utils.CreateCompilerHost(root.Dir, root.FS), CompilerOptions: program.SourceOnlyCompilerOptions(), RootFileNames: []string{file}, SingleThreaded: true})
		if err != nil {
			t.Fatal(err)
		}
		r := &target.NoRelativePackagesRule
		var all rule.RuleDiagnostic
		for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{SourceFile: p.GetSourceFile(file)}.WithProgram(p).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }})
			listeners := r.Run(ctx, rule_tester.ResolveTestCaseOptions(t, r, map[string]any{"commonjs": true}))
			node := ctx.SourceFile.Statements.Nodes[0]
			if node.Kind == ast.KindVariableStatement {
				node = node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration().Initializer
			}
			listeners[node.Kind](node)
			if len(diagnostics) != 1 {
				t.Fatalf("demand %d: %d diagnostics", demand, len(diagnostics))
			}
			got := diagnostics[0]
			if demand == rule.EditDemandAll {
				all = got
			}
			if got.FixesPtr != nil {
				t.Fatal("unexpected autofix")
			}
			if demand&rule.EditDemandSuggestion != 0 {
				if got.Suggestions == nil || !reflect.DeepEqual(got.Suggestions, all.Suggestions) {
					t.Fatal("missing or changed suggestion")
				}
				suggestions := *got.Suggestions
				if len(suggestions) != 1 || suggestions[0].Message.Id != "usePackageImport" || suggestions[0].Message.Description != fmt.Sprintf("Replace with package import `%s`", tc.candidate) {
					t.Fatal("unexpected suggestion message")
				}
				output, _, changed := linter.ApplyRuleFixes(tc.code, suggestions)
				expected := strings.Replace(tc.code, fmt.Sprintf("%q", tc.imported), fmt.Sprintf("%q", tc.candidate), 1)
				if !changed || output != expected {
					t.Fatalf("suggestion output %q, want %q", output, expected)
				}
			} else if got.Suggestions != nil {
				t.Fatal("unexpected suggestion")
			}
			want := all
			got.Suggestions, want.Suggestions = nil, nil
			if !reflect.DeepEqual(got, want) {
				t.Fatal("edit demand changed diagnostic")
			}
		}
	}
}

func TestNoRelativePackagesScopedSuggestion(t *testing.T) {
	root := safetyRoot(t, "C01", true, false)
	if err := os.WriteFile(tspath.ResolvePath(root.Dir, "packages/b/package.json"), []byte(`{"name":"@scope/b"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	scope := tspath.ResolvePath(root.Dir, "node_modules/@scope")
	if err := os.MkdirAll(scope, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(tspath.ResolvePath(root.Dir, "packages/b"), tspath.ResolvePath(scope, "b")); err != nil {
		t.Fatal(err)
	}
	item := safePackageCase("packages/a/source.cjs", `const value = require("../b");`, "../b", "@scope/b", map[string]any{"commonjs": true})
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &target.NoRelativePackagesRule, nil, []rule_tester.InvalidTestCase{item})
}

func TestNoRelativePackagesUnicodeSuggestion(t *testing.T) {
	root := safetyRoot(t, "C01", true, false)
	if err := os.WriteFile(tspath.ResolvePath(root.Dir, "packages/b/工具.js"), []byte(`module.exports = 1;`), 0o644); err != nil {
		t.Fatal(err)
	}
	code := "const emoji = '😀';\nconst value = require(\"../b/工具.js\");"
	item := safePackageCase("packages/a/source.cjs", code, "../b/工具.js", "b/工具.js", map[string]any{"commonjs": true})
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &target.NoRelativePackagesRule, nil, []rule_tester.InvalidTestCase{item})
}
