package no_relative_packages_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_relative_packages"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoRelativePackagesExtras(t *testing.T) {
	file := "tests/files/package/index.js"
	common := map[string]any{"commonjs": true}
	both := map[string]any{"commonjs": true, "amd": true}
	valid := []rule_tester.ValidTestCase{
		{FileName: file, Code: `import '../missing'; import ''; import 'node:fs'; import 'package-named';`},
		{FileName: file, Code: `require('../package-named'); define(['../package-named'], () => {});`},
		{FileName: file, Code: `import '../package-named'; export * from '../package-named'; import('../package-named');`, Options: map[string]any{"esmodule": false}},
		{FileName: file, Code: `import '../package-named'; require('../package-named'); define(['../package-named'], () => {});`, Options: map[string]any{"commonjs": true, "amd": true, "ignore": []any{`(?<=/)package-named$`}}},
		{FileName: file, Code: `import '../package-named';`, Settings: map[string]any{"import/internal-regex": "package-named"}},
		{FileName: file, Code: `import '../package-named';`, Settings: map[string]any{"import/core-modules": []any{"../package-named"}}},
		{FileName: "tests/files/package-named/index.js", Code: `import '../same'; import './deep/a';`}, // Compare names, not package directories.
		{FileName: "tests/files/package-named/index.js", Code: `import '../array/index.js';`},          // Shared metadata reader rejects non-object manifests.
		{FileName: file, Code: `import '../bad';`}, // Invalid JSON stops lookup instead of inheriting the parent name.
		{FileName: "tests/files/bad/index.js", Code: `import '../package-named';`},
		{FileName: file, Code: "require(); require(1); require(name); require(`../package-named`); require('../package-named', 1); obj.require('../package-named'); require.resolve('../package-named'); new require('../package-named'); import(`../package-named`);", Options: both},
		{FileName: "tests/files/package/index.ts", Code: `import alias = require('../package-named'); type T = import('../package-named').Value; require('../package-named' as string); (require as any)('../package-named');`, Options: both},
		{FileName: file, Code: "define('name', ['../package-named'], () => {}); define(['../package-named']); define(['require', 'exports', , 1], () => {});", Options: both},
		{FileName: file, Code: `import '../package-named';`, Options: map[string]any{"ignore": []any{"package-named"}}, Settings: map[string]any{"import/resolver": "missing-resolver"}},
		{FileName: file, Code: `import '../package-named';`, Options: map[string]any{"esmodule": false, "commonjs": false, "amd": false}},
	}
	invalid := []rule_tester.InvalidTestCase{}
	for _, tc := range []struct {
		code, literal, imported, proposed string
		options                           any
	}{
		{`import '../package-named';`, `'../package-named'`, "../package-named", "package-named", map[string]any{}},
		{`export * from '../package-named';`, `'../package-named'`, "../package-named", "package-named", nil},
		{`export { value } from '../package-named';`, `'../package-named'`, "../package-named", "package-named", nil},
		{`export * as named from '../package-named';`, `'../package-named'`, "../package-named", "package-named", nil},
		{`import(('../package-named'), { with: { type: 'json' } });`, `'../package-named'`, "../package-named", "package-named", nil},
		{`(require)(('../package-named'));`, `'../package-named'`, "../package-named", "package-named", common},
		{`require?.('../package-named');`, `'../package-named'`, "../package-named", "package-named", common},
		{`function f(require) { require('../package-named'); }`, `'../package-named'`, "../package-named", "package-named", common},
		{`require(/** @type {string} */ ('../package-named'));`, `'../package-named'`, "../package-named", "package-named", common},
		{`define(['../package-named'], () => {});`, `'../package-named'`, "../package-named", "package-named", map[string]any{"amd": true, "esmodule": false}},
		{`require(['require', '../package-named', 'exports'], () => {});`, `'../package-named'`, "../package-named", "package-named", both},
		{`import '../package-named/index.js';`, `'../package-named/index.js'`, "../package-named/index.js", "package-named/index.js", nil},
		{`import '../package-named/deep/a';`, `'../package-named/deep/a'`, "../package-named/deep/a", "package-named/deep/a", nil},
		{`import '../package-named/';`, `'../package-named/'`, "../package-named/", "package-named", nil},
		{`import './../package-named';`, `'./../package-named'`, "./../package-named", "package-named", nil},
		{`import '\u002e./package-named';`, `'\u002e./package-named'`, "../package-named", "package-named", nil},
		{"// 中文😀\nimport foo\n from '../package-named';", `'../package-named'`, "../package-named", "package-named", nil},
		{`const emoji = '😀'; import('../package-named');`, `'../package-named'`, "../package-named", "package-named", nil},
	} {
		invalid = append(invalid, packageCase(file, tc.code, tc.literal, tc.imported, tc.proposed, tc.options))
	}
	for _, code := range []string{`import type { Value } from '../package-named';`, `export type { Value } from '../package-named';`, `const view = <Widget>{require('../package-named')}</Widget>;`} {
		invalid = append(invalid, packageCase("tests/files/package/index.tsx", code, `'../package-named'`, "../package-named", "package-named", common))
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{FileName: file, Code: "import '../package-named';\nimport '../package-scoped';", Settings: map[string]any{"import/resolver": "missing-resolver"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: `Resolve error: unable to load resolver "missing-resolver".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1}}})
	runPackages(t, valid, invalid)
}

func TestNoRelativePackagesUnverifiedEditDemand(t *testing.T) {
	root := packageRoot(t)
	file := tspath.ResolvePath(root.Dir, "tests/files/package/index.js")
	code := `import '../package-named';`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: code})
	p, err := program.NewFromRoots(program.RootOptions{Host: utils.CreateCompilerHost(root.Dir, fs), CompilerOptions: program.SourceOnlyCompilerOptions(), RootFileNames: []string{file}, SingleThreaded: true})
	if err != nil {
		t.Fatal(err)
	}
	r := &target.NoRelativePackagesRule
	var all rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{SourceFile: p.GetSourceFile(file)}.WithProgram(p).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }})
		listeners := r.Run(ctx, rule_tester.ResolveTestCaseOptions(t, r, nil))
		listeners[ast.KindImportDeclaration](ctx.SourceFile.Statements.Nodes[0])
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
		}
		got := diagnostics[0]
		if demand == rule.EditDemandAll {
			all = got
		}
		if got.Suggestions != nil {
			t.Fatal("unexpected suggestions")
		}
		if got.FixesPtr != nil {
			t.Fatal("unexpected autofix")
		}
		want := all
		got.FixesPtr, want.FixesPtr = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Fatal("edit demand changed diagnostic")
		}
	}
}

func TestNoRelativePackagesPackageMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, contextJSON, importJSON string
		reports                       bool
	}{
		{"no named context ancestor", `{}`, `{"name":"package-named"}`, false},
		{"no named import ancestor", `{"name":"package-named"}`, `{}`, false},
		{"null package stops lookup", `{"name":"outer"}`, `null`, false},
		{"empty name inherits ancestor", `{"name":"outer"}`, `{"name":""}`, false},
		{"BOM package name", `{"name":"outer"}`, "\ufeff{\"name\":\"package-named\"}", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := packageRoot(t)
			// Override the outer package and importer independently. The embedded
			// filesystem has no package.json above this isolated fixture root.
			root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
				tspath.ResolvePath(root.Dir, "package.json"):                           tc.contextJSON,
				tspath.ResolvePath(root.Dir, "tests/files/package/package.json"):       `{}`,
				tspath.ResolvePath(root.Dir, "tests/files/package-named/package.json"): tc.importJSON,
			})
			code := `import '../package-named/index.js';`
			if tc.name == "no named import ancestor" {
				root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "package.json"):                     `{}`,
					tspath.ResolvePath(root.Dir, "tests/files/package/package.json"): tc.contextJSON,
				})
			}
			if tc.reports {
				item := packageCase("tests/files/package/index.js", code, `'../package-named/index.js'`, "../package-named/index.js", "package-named/index.js", nil)
				rule_tester.RunRuleTester(root, "tsconfig.json", t, &target.NoRelativePackagesRule, nil, []rule_tester.InvalidTestCase{item})
			} else {
				rule_tester.RunRuleTester(root, "tsconfig.json", t, &target.NoRelativePackagesRule, []rule_tester.ValidTestCase{{FileName: "tests/files/package/index.js", Code: code}}, nil)
			}
		})
	}
}

func TestNoRelativePackagesSchema(t *testing.T) {
	for _, options := range [][]any{
		[]any{map[string]any{"unknown": true}},
		[]any{map[string]any{"commonjs": "true"}},
		[]any{map[string]any{"amd": 1}},
		[]any{map[string]any{"esmodule": nil}},
		[]any{map[string]any{"ignore": []any{}}},
		[]any{map[string]any{"ignore": []any{"x", "x"}}},
		[]any{map[string]any{"ignore": []any{1}}},
		[]any{map[string]any{"ignore": []any{"["}}},
		[]any{map[string]any{}, map[string]any{}},
	} {
		if err := target.NoRelativePackagesRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options %#v", options)
		}
	}
}
