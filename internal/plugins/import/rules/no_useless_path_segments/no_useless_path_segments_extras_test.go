// cspell:ignore fdeep
package no_useless_path_segments_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_useless_path_segments"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUselessPathSegmentsExtras(t *testing.T) {
	both := map[string]any{"commonjs": true, "noUselessIndex": true}
	valid := []rule_tester.ValidTestCase{}
	for _, code := range []string{
		`import "./index"; require("./deep//a");`,
		`import "../missing"; import "./a..b"; import "./.hidden";`,
		`import "package//name"; import "/absolute//name"; import "";`,
		"import(`./deep//a`); require(`./deep//a`);",
		`require(); require('./deep//a', extra); require(variable); require.resolve('./deep//a');`,
		`object.require('./deep//a'); object['require']('./deep//a'); new require('./deep//a');`,
		`export { local }; const local = './deep//a';`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "files/foo.js", Options: map[string]any{}})
	}
	for _, code := range []string{
		"import(`./deep//a`); require(`./deep//a`);",
		`require(); require('./deep//a', extra); require(variable); require.resolve('./deep//a');`,
		`object.require('./deep//a'); object['require']('./deep//a'); new require('./deep//a');`,
		`import './unknown/index.ts';`, // Default extension set excludes .ts.
		`import '.alias';`,             // Bare dot-prefixed package resolves; ./.alias does not.
		`import './a/\n'; import './x/index.js\n'; import './x/index\u2028';`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "files/foo.js", Options: both})
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: `require('./deep//a'); import './index';`, FileName: "files/foo.js", Options: map[string]any{"commonjs": false, "noUselessIndex": false}},
		rule_tester.ValidTestCase{Code: `type T = import('./deep//a'); import value = require('./deep//a');`, FileName: "files/foo.ts", Options: both},
		rule_tester.ValidTestCase{Code: `require('./deep//a' as string); import('./deep//a'!); (require as any)('./deep//a'); (require!)('./deep//a'); import('./deep//a' satisfies string);`, FileName: "files/foo.ts", Options: both},
	)
	invalid := []rule_tester.InvalidTestCase{}
	for _, test := range []struct{ code, literal, imported, proposed, output string }{
		{`export * from './deep//a';`, `'./deep//a'`, "./deep//a", "./deep/a", `export * from "./deep/a";`},
		{`export { default as a } from './deep//a';`, `'./deep//a'`, "./deep//a", "./deep/a", `export { default as a } from "./deep/a";`},
		{`export * as ns from './deep//a';`, `'./deep//a'`, "./deep//a", "./deep/a", `export * as ns from "./deep/a";`},
		{`((require))(('./deep//a'));`, `'./deep//a'`, "./deep//a", "./deep/a", `((require))(("./deep/a"));`},
		{`require?.('./deep//a');`, `'./deep//a'`, "./deep//a", "./deep/a", `require?.("./deep/a");`},
		{`function f(require) { require('./deep//a'); }`, `'./deep//a'`, "./deep//a", "./deep/a", `function f(require) { require("./deep/a"); }`},
		{`import(('./deep//a'), { with: { type: 'json' } });`, `'./deep//a'`, "./deep//a", "./deep/a", `import(("./deep/a"), { with: { type: 'json' } });`},
		{`import './missing//file';`, `'./missing//file'`, "./missing//file", "./missing/file", `import "./missing/file";`},
		{`import '.hidden/file';`, `'.hidden/file'`, ".hidden/file", "./.hidden/file", `import "./.hidden/file";`},
		{`import '..hidden/file';`, `'..hidden/file'`, "..hidden/file", "./..hidden/file", `import "./..hidden/file";`},
		{`import './missing/../target';`, `'./missing/../target'`, "./missing/../target", "./target", `import "./target";`},
		{`import '../../no-useless-path-segments/files/deep/a.js';`, `'../../no-useless-path-segments/files/deep/a.js'`, "../../no-useless-path-segments/files/deep/a.js", "./deep/a.js", `import "./deep/a.js";`},
		{`import '../../no-useless-path-segments/files';`, `'../../no-useless-path-segments/files'`, "../../no-useless-path-segments/files", "./", `import "./";`},
		{`import './unknown/index.mjs';`, `'./unknown/index.mjs'`, "./unknown/index.mjs", "./unknown", `import "./unknown";`},
		{`import './unknown/index.cjs';`, `'./unknown/index.cjs'`, "./unknown/index.cjs", "./unknown", `import "./unknown";`},
		{`import '../index.cjs';`, `'../index.cjs'`, "../index.cjs", "..", `import "..";`},
		{`import './package/index.js';`, `'./package/index.js'`, "./package/index.js", "./package", `import "./package";`}, // Upstream follows package main after removing index.
		{"// 中文😀\nimport /* 😀 */ './深//度';", "'./深//度'", "./深//度", "./深/度", "// 中文😀\nimport /* 😀 */ \"./深/度\";"},
		{"import './deep/\\\n/a';", "'./deep/\\\n/a'", "./deep//a", "./deep/a", `import "./deep/a";`},
		{`import '.\u002fdeep//a';`, `'.\u002fdeep//a'`, "./deep//a", "./deep/a", `import "./deep/a";`},
		{`import './<>&//"x';`, `'./<>&//"x'`, `./<>&//"x`, `./<>&/"x`, `import "./<>&/\"x";`},
		{`import '.\\file';`, `'.\\file'`, `.\file`, `./.\file`, `import "./.\\file";`},
		{`import './a\\b//c';`, `'./a\\b//c'`, `./a\b//c`, `./a\b/c`, `import "./a\\b/c";`},
		{`import './\ud800//x';`, `'./\ud800//x'`, "./\xed\xa0\x80//x", "./\xed\xa0\x80/x", `import "./\ud800/x";`},
		{`import './\udc00//x';`, `'./\udc00//x'`, "./\xed\xb0\x80//x", "./\xed\xb0\x80/x", `import "./\udc00/x";`},
		{`import './\ud800\udc00//x';`, `'./\ud800\udc00//x'`, "./\U00010000//x", "./\U00010000/x", "import \"./\U00010000/x\";"},
		{`import './\u0085\u2028\u2029//x';`, `'./\u0085\u2028\u2029//x'`, "./\u0085\u2028\u2029//x", "./\u0085\u2028\u2029/x", "import \"./\u0085\u2028\u2029/x\";"},
		{`import './\x00\b\f\n\r\t//x';`, `'./\x00\b\f\n\r\t//x'`, "./\x00\b\f\n\r\t//x", "./\x00\b\f\n\r\t/x", `import "./\u0000\b\f\n\r\t/x";`},
	} {
		item := invalidPath(test.code, test.literal, test.imported, test.proposed, test.output, both)
		if test.proposed == "./" {
			item.Output = append(item.Output, `import ".";`)
		}
		invalid = append(invalid, item)
	}
	for _, code := range []string{
		`import type { T } from './deep//a';`, `export type { T } from './deep//a';`,
		`require<string>('./deep//a');`, `const element = <div>{import('./deep//a')}</div>;`,
	} {
		item := invalidPath(code, `'./deep//a'`, "./deep//a", "./deep/a", strings.ReplaceAll(code, `'./deep//a'`, `"./deep/a"`), both)
		item.FileName = "files/foo.tsx"
		invalid = append(invalid, item)
	}
	for _, settings := range []map[string]any{
		{"import/extensions": []any{".ts"}, "import/resolver": map[string]any{"node": map[string]any{"extensions": []any{".ts"}}}},
		{"import/parsers": map[string]any{"typescript": []any{".ts", ".ts"}}, "import/resolver": "typescript"},
	} {
		item := invalidPath(`import './typed/index.ts';`, `'./typed/index.ts'`, "./typed/index.ts", "./typed/", `import "./typed/";`, both)
		item.Settings = settings
		invalid = append(invalid, item)
	}
	item := invalidPath(`import './json/index.json';`, `'./json/index.json'`, "./json/index.json", "./json/", `import "./json/";`, both)
	item.Settings = map[string]any{"import/extensions": []any{".json"}}
	invalid = append(invalid, item)
	item = invalidPath(`import './typed-only//index';`, `'./typed-only//index'`, "./typed-only//index", "./typed-only/index", `import "./typed-only/index";`, both)
	item.Settings = map[string]any{"import/resolver": "typescript"}
	item.Output = append(item.Output, `import "./typed-only";`)
	invalid = append(invalid, item)
	item = invalidPath(`import './unknown/index.ts';`, `'./unknown/index.ts'`, "./unknown/index.ts", "./unknown", `import "./unknown";`, both)
	item.Settings = map[string]any{"import/extensions": []any{`.t(s|sx)`}}
	invalid = append(invalid, item)
	// Upstream compiles extension patterns without the Unicode flag: dots,
	// character classes, and quantifiers operate on UTF-16 code units.
	for _, tc := range []struct {
		extension, suffix string
		report            bool
	}{
		{`.(a|aa)+`, "aaa", true},
		{`.(.)`, "😀", false},
		{`.(..)`, "😀", true},
		{`.😀`, "😀", true},
		{`.[😀]`, "😀", false},
		{`.(😀)`, "😀", true},
		{`.😀+`, "😀😀", false},
		{`.\😀`, "😀", true},
		{`.\\😀`, "😀", false},
		{`.\ud83d\ude00`, "😀", true},
	} {
		imported := "./x/index." + tc.suffix
		literal := "'" + imported + "'"
		code := "import " + literal + ";"
		settings := map[string]any{"import/extensions": []any{tc.extension}}
		if tc.report {
			item := invalidPath(code, literal, imported, "./x", `import "./x";`, both)
			item.Settings = settings
			invalid = append(invalid, item)
		} else {
			valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "files/foo.js", Settings: settings, Options: both})
		}
	}
	item = invalidPath(`import './x/index.\ud800';`, `'./x/index.\ud800'`, "./x/index.\xed\xa0\x80", "./x", `import "./x";`, both)
	item.Settings = map[string]any{"import/extensions": []any{`.(.)`}}
	invalid = append(invalid, item)
	valid = append(valid, rule_tester.ValidTestCase{
		Code: `import './index.` + strings.Repeat("a", 200) + `b';`, FileName: "files/foo.js", Options: both,
		Settings: map[string]any{"import/extensions": []any{`.(a|aa)+`}},
	}) // A regexp timeout must never produce a diagnostic or a fix.
	// Settings participate in the upstream regex, so a match need not contain
	// a slash or end at index. Keep dirname safe for those matches too.
	for _, tc := range []struct{ imported, proposed string }{{".alias", "."}, {"./x/index.js/", "./x"}} {
		literal := `"` + tc.imported + `"`
		item := invalidPath("import "+literal+";", literal, tc.imported, tc.proposed, `import "`+tc.proposed+`";`, both)
		item.Settings = map[string]any{"import/extensions": []any{`.js)|[.]alias(`}, "import/core-modules": []any{tc.imported}}
		invalid = append(invalid, item)
	}
	rule_tester.RunRuleTester(pathRoot(t), "tsconfig.json", t, &target.NoUselessPathSegmentsRule, valid, invalid)
}

func TestNoUselessPathSegmentsSchema(t *testing.T) {
	for _, options := range [][]any{{true}, {"always"}, {map[string]any{"commonjs": "true"}}, {map[string]any{"noUselessIndex": 1}}, {map[string]any{"unknown": true}}, {map[string]any{}, map[string]any{}}} {
		if err := target.NoUselessPathSegmentsRule.Schema.Validate(options); err == nil {
			t.Fatalf("accepted invalid options: %v", options)
		}
	}
}

func TestNoUselessPathSegmentsEditDemand(t *testing.T) {
	root := pathRoot(t)
	file := tspath.ResolvePath(root.Dir, "files/foo.js")
	r := &target.NoUselessPathSegmentsRule
	for _, tc := range []struct{ code, output string }{
		{`import './deep//a';`, `import "./deep/a";`},
		{`import './bar/index.js';`, `import "./bar/";`},
		{`import '../files/malformed';`, `import "./malformed";`},
	} {
		fs := utils.NewOverlayVFS(root.FS, map[string]string{file: tc.code})
		p, err := program.NewFromRoots(program.RootOptions{Host: utils.CreateCompilerHost(root.Dir, fs), CompilerOptions: program.SourceOnlyCompilerOptions(), RootFileNames: []string{file}, SingleThreaded: true})
		if err != nil {
			t.Fatal(err)
		}
		var all rule.RuleDiagnostic
		for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{SourceFile: p.GetSourceFile(file)}.WithProgram(p).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }})
			listeners := r.Run(ctx, rule_tester.ResolveTestCaseOptions(t, r, map[string]any{"noUselessIndex": true}))
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
			if demand&rule.EditDemandAutofix != 0 {
				if got.FixesPtr == nil || !reflect.DeepEqual(got.FixesPtr, all.FixesPtr) {
					t.Fatal("missing or changed fix")
				}
				output, _, fixed := linter.ApplyRuleFixes(tc.code, diagnostics)
				if !fixed || output != tc.output {
					t.Fatalf("got %q, want %q", output, tc.output)
				}
			} else if got.FixesPtr != nil {
				t.Fatal("unexpected fix")
			}
			want := all
			got.FixesPtr, want.FixesPtr = nil, nil
			if !reflect.DeepEqual(got, want) {
				t.Fatal("edit demand changed diagnostic")
			}
		}
		if len(p.SourceFiles()) != 1 {
			t.Fatal("resolution loaded dependency ASTs")
		}
	}
}

func TestNoUselessPathSegmentsResolverError(t *testing.T) {
	// Unknown JavaScript resolvers cannot run natively. Report once per file,
	// while keeping upstream's comparison of two unresolved paths.
	rule_tester.RunRuleTester(pathRoot(t), "tsconfig.json", t, &target.NoUselessPathSegmentsRule, nil, []rule_tester.InvalidTestCase{{
		Code: `import './missing//one'; import '../missing';`, FileName: "files/foo.js",
		Settings: map[string]any{"import/resolver": "webpack"},
		Output:   []string{`import "./missing/one"; import '../missing';`},
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "", Message: `Resolve error: unable to load resolver "webpack".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
			{MessageId: "", Message: `Useless path segments for "./missing//one", should be "./missing/one"`, Line: 1, Column: 8, EndLine: 1, EndColumn: 24},
		},
	}})
}

func TestNoUselessPathSegmentsRequireResolutionMode(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, code := range []string{`(require)('./typed-only//index');`, `((require))(('./typed-only//index'));`} {
		item := invalidPath(code, `'./typed-only//index'`, "./typed-only//index", "./typed-only/index", strings.ReplaceAll(code, `'./typed-only//index'`, `"./typed-only/index"`), map[string]any{"commonjs": true})
		item.FileName = "files/foo.mts"
		item.Settings = map[string]any{"import/resolver": "typescript"}
		invalid = append(invalid, item)
	}
	rule_tester.RunRuleTester(pathRoot(t), "tsconfig.nodenext.json", t, &target.NoUselessPathSegmentsRule, nil, invalid)
}
