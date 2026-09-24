// cspell:ignore ffoo
package no_absolute_path

import (
	"reflect"
	"runtime"
	"testing"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Options, AST adaptations, and fix boundaries beyond the pinned upstream suite.
func TestNoAbsolutePathExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &NoAbsolutePathRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "import \"/foo\"; export * from \"/foo\"; import(\"/foo\");",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"esmodule": false}},
			},
			{
				Code:     "require(\"/foo\");",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"commonjs": false}},
			},
			{
				Code:     "import \"/foo\"; require(\"/foo\"); define([\"/foo\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"esmodule": false, "commonjs": false, "amd": false}},
			},
			{
				Code:     "import \"/ignored/a\"; require(\"/ignored/a\"); import(\"/ignored/a\"); export * from \"/ignored/a\"; define([\"/ignored/a\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true, "ignore": []any{"^/ignored(?=/)"}}},
			},
			{
				Code:     "import \"/same/same\"; import \"/second\";",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"ignore": []any{"^/(same)/\\1$", "(?<=/)second$"}}},
			},
			{
				Code:     "require(); require(\"/foo\", 1); require(1); require(null); require(true); require(/foo/); require(`/foo`); require(\"/\" + name); import(`/foo`); import(name);",
				FileName: "/foo/bar/index.ts",
			},
			{
				Code:     "module.require(\"/foo\"); require.resolve(\"/foo\"); obj[\"require\"](\"/foo\"); obj?.require(\"/foo\"); (obj?.require)(\"/foo\"); new require(\"/foo\"); class C { #load() { return this.#load(\"/foo\"); } }",
				FileName: "/foo/bar/index.ts",
			},
			{
				Code:     "define([\"/foo\"]); define(\"named\", [\"/foo\"], cb); require([\"/foo\"], cb, extra); obj.define([\"/foo\"], cb); define(\"/foo\", cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true}},
			},
			{
				Code:     "define([, 1, null, true, `/foo`, \"/\" + name, ...[\"/foo\"], \"require\", \"exports\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true}},
			},
			{
				Code:     "import alias = require(\"/foo\"); type T = import(\"/foo\").T; (require as any)(\"/foo\"); require!(\"/foo\"); require(\"/foo\" as string); import(\"/foo\" as string); define(([\"/foo\"] as string[]), cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true}},
			},
			{
				Code:     "const value = 1; export {value}; const view = <a href=\"/foo\" />;",
				FileName: "/foo/bar/index.tsx",
			},
			{
				Code:     "import \"file:///foo\"; import \"node:path\"; import \"\"; import \"~/foo\";",
				FileName: "/foo/bar/index.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "import \"/foo\";",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{}},
				Output:   []string{"import \"..\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:     "import \"/foo\"; require(\"/foo\");",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"esmodule": true, "commonjs": true, "amd": false}},
				Output:   []string{"import \"..\"; require(\"..\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:     "import \"/foo\"; require(\"/foo\"); define([\"/foo\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"commonjs": false, "amd": true}},
				Output:   []string{"import \"..\"; require(\"/foo\"); define([\"..\"], cb);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:     "import \"/foo\"; require(\"/foo\"); define([\"/foo\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"esmodule": false, "amd": true}},
				Output:   []string{"import \"/foo\"; require(\"..\"); define([\"..\"], cb);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 30},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:     "define([\"/ignored/a\", \"/other\", \"/ignored/b\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"esmodule": false, "commonjs": false, "amd": true, "ignore": []any{"^/ignored/"}}},
				Output:   []string{"define([\"/ignored/a\", \"../../other\", \"/ignored/b\"], cb);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:     "import \"/Ignored/a\";",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"ignore": []any{"^/ignored/"}}},
				Output:   []string{"import \"../../Ignored/a\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:     "export {foo} from \"/foo\"; export * from \"/foo\"; export * as ns from \"/foo\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"export {foo} from \"..\"; export * from \"..\"; export * as ns from \"..\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 19, EndLine: 1, EndColumn: 25},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 47},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 69, EndLine: 1, EndColumn: 75},
				},
			},
			{
				Code:     "import type T from \"/foo\"; export type {T} from \"/foo\"; export type * from \"/foo\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import type T from \"..\"; export type {T} from \"..\"; export type * from \"..\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 49, EndLine: 1, EndColumn: 55},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 76, EndLine: 1, EndColumn: 82},
				},
			},
			{
				Code:     "import data from \"/foo\" with {type:\"json\"}; import(\"/foo\", {with:{type:\"json\"}});",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import data from \"..\" with {type:\"json\"}; import(\"..\", {with:{type:\"json\"}});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 52, EndLine: 1, EndColumn: 58},
				},
			},
			{
				Code:     "require?.((\"/foo\")); ((require))((\"/foo\")); require<string>(\"/foo\");",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"require?.((\"..\")); ((require))((\"..\")); require<string>(\"..\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 35, EndLine: 1, EndColumn: 41},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 61, EndLine: 1, EndColumn: 67},
				},
			},
			{
				Code:     "function load(require) { require(\"/foo\"); }",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"function load(require) { require(\"..\"); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 34, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:     "(define)(([\"require\", \"exports\", \"module\", (\"/foo\"), , `/skip`, ...[\"/skip\"], \"/foo/path\"]), false);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"(define)(([\"require\", \"exports\", \"module\", (\"..\"), , `/skip`, ...[\"/skip\"], \"../path\"]), false);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 45, EndLine: 1, EndColumn: 51},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 79, EndLine: 1, EndColumn: 90},
				},
			},
			{
				Code:     "define?.([\"/foo\"], cb); require([\"/foo\"], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true, "commonjs": false}},
				Output:   []string{"define?.([\"..\"], cb); require([\"..\"], cb);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 34, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:     "/** @type {any} */ (require)(/** @type {string} */ (\"/foo\")); define(/** @type {string[]} */ ([\"/foo\"]), cb);",
				FileName: "/foo/bar/index.js",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"/** @type {any} */ (require)(/** @type {string} */ (\"..\")); define(/** @type {string[]} */ ([\"..\"]), cb);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 53, EndLine: 1, EndColumn: 59},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 96, EndLine: 1, EndColumn: 102},
				},
			},
			{
				Code:     "const view = <span>{require(\"/foo\")}</span>;",
				FileName: "/foo/bar/index.tsx",
				Output:   []string{"const view = <span>{require(\"..\")}</span>;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:     "/*😀*/ import \"\\u002ffoo\";\nimport(\n  \"/foo/path\"\n);",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"/*😀*/ import \"..\";\nimport(\n  \"../path\"\n);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 26},
					{MessageId: "", Message: absolutePathMessage, Line: 3, Column: 3, EndLine: 3, EndColumn: 14},
				},
			},
			{
				Code:     "import \"/foo/ba\\\nr/baz\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./baz\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code:     "import \"/foo/bar\"; import \"/\"; import \"/foo/bar/.hidden\"; import \"/foo/bar/../bar//baz/\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\"; import \"../..\"; import \"./.hidden\"; import \"./baz\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 30},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 39, EndLine: 1, EndColumn: 57},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 66, EndLine: 1, EndColumn: 89},
				},
			},
			{
				Code:     "import \"/Foo/bar/baz\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"../../Foo/bar/baz\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:     "require('/foo/bar/a\\\\b\"c');",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"require(\"./a\\\\b\\\"c\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:     "import '/foo/bar/\\u0000\\n\\t';",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\\u0000\\n\\t\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:     "import \"/foo/bar/你好😀\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./你好😀\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 23},
				},
			},
			// JSON escaping keeps the same path value.
			{
				Code:     "import \"/foo/bar/<a>&\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\\u003ca\\u003e\\u0026\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 23},
				},
			},
			// Preserve the original code unit when fixing an unpaired surrogate.
			{
				Code:     "import \"/foo/bar/\\ud800\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\\ud800\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			// Additional Unicode and path-normalization boundaries.
			{
				Code:     "import \"/foo/bar/\\u2028\\u2029\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\\u2028\\u2029\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:     "import \"/foo/bar/\\ud83d\\ude00\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./😀\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:     "import \"/foo/bar/\\udc00\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\\udc00\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:     "import \"/foo/bar/\\ud800/../baz\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./baz\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 32},
				},
			},
			// Preserve module values and explicit relative paths in every visitor.
			{
				Code:     "require('/foo/bar/.hidden.cjs'); import('/foo/bar/..hidden.cjs'); export * from '/foo/bar/.config/index.js'; define(['/foo/bar/.d.ts'], cb);",
				FileName: "/foo/bar/index.ts",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"require(\"./.hidden.cjs\"); import(\"./..hidden.cjs\"); export * from \"./.config/index.js\"; define([\"./.d.ts\"], cb);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 64},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 81, EndLine: 1, EndColumn: 108},
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 118, EndLine: 1, EndColumn: 134},
				},
			},
			{
				Code:     "require('/foo/bar/a\\ud800\\\\ud800\"c');",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"require(\"./a\\ud800\\\\ud800\\\"c\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code:     "import \"/foo/bar/\\udc00\\ud800\\ud800\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./\\udc00\\ud800\\ud800\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:     "import \"/foo/bar/.\\ud800\";",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"import \"./.\\ud800\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:     "require('/foo/bar/.\\\\file');",
				FileName: "/foo/bar/index.ts",
				Output:   []string{"require(\"./.\\\\file\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 27},
				},
			},
		})
}

func TestNoAbsolutePathEditDemand(t *testing.T) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/foo/bar/index.js", Path: "/foo/bar/index.js"}, `import value from "/foo/bar/.\ud800";`, core.ScriptKindJS)
	r := &NoAbsolutePathRule
	var all rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		})
		node := file.Statements.Nodes[0]
		r.Run(ctx, nil)[node.Kind](node)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
		}
		got := diagnostics[0]
		if demand == rule.EditDemandAll {
			all = got
		}
		if demand == rule.EditDemandAll || demand == rule.EditDemandAutofix {
			if got.FixesPtr == nil || len(*got.FixesPtr) != 1 || (*got.FixesPtr)[0].Text != `"./.\ud800"` || !reflect.DeepEqual(got.FixesPtr, all.FixesPtr) {
				t.Fatalf("demand %d: missing or incorrect fix", demand)
			}
		} else if got.FixesPtr != nil {
			t.Fatalf("demand %d: unexpected fix", demand)
		}
		if got.Suggestions != nil {
			t.Fatalf("demand %d: unexpected suggestions", demand)
		}
		want := all
		got.FixesPtr, want.FixesPtr = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("demand %d changed diagnostic identity", demand)
		}
	}
}

func TestNoAbsolutePathHostPaths(t *testing.T) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/host.js", Path: "/host.js"}, `import "C:/foo"; import "\\\\server\\share"; import "\\foo"; import "C:foo";`, core.ScriptKindJS)
	count := 0
	r := &NoAbsolutePathRule
	ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Demand: rule.EditDemandNone,
		Report: func(rule.RuleDiagnostic) { count++ },
	})
	listeners := r.Run(ctx, nil)
	for _, node := range file.Statements.Nodes {
		listeners[node.Kind](node)
	}
	want := 0
	if runtime.GOOS == "windows" {
		want = 3
	}
	if count != want {
		t.Fatalf("Windows spellings on %s: got %d diagnostics, want %d", runtime.GOOS, count, want)
	}
}

// Expected values follow Node's path.posix.relative, including Windows strings
// interpreted as POSIX components rather than as drives or separators.
func TestRelativeImportPath(t *testing.T) {
	for _, tc := range []struct{ from, to, want string }{
		{"/", "/foo", "./foo"},
		{"/foo/bar", "/foo/bar", "./"},
		{"/foo/bar", "///foo//bar/baz", "./baz"},
		{"/foo/bar", "/foo/bar/.hidden", "./.hidden"},
		{"/foo/bar", "/foo/bar/..hidden/file", "./..hidden/file"},
		{"/foo/bar", `/foo/bar/.\file`, `./.\file`},
		{"/foo/bar", "/foo/.hidden", "../.hidden"},
		{"C:/foo/bar", "C:/foo/baz", "../baz"},
		{"C:/foo/bar", "D:/foo/baz", "../../../D:/foo/baz"},
		{"C:/foo/bar", `\foo`, `../../../\foo`},
		{`C:\foo\bar`, `C:\foo\baz`, `../C:\foo\baz`},
		{"//host/share/dir", "//host/share/file", "../file"},
	} {
		if got := relativeImportPath(posixPathComponents(tc.from, "/work"), tc.to, "/work"); got != tc.want {
			t.Errorf("relative path from %q to %q = %q, want %q", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestQuoteModulePath(t *testing.T) {
	for _, units := range [][]uint16{
		{},
		{0xD800},
		{0xDBFF},
		{0xDC00},
		{0xDFFF},
		{0xD83D, 0xDE00},
		{0xDC00, 0xD800},
		{0xD800, 0xD800, 0xDC00, 0xDFFF},
		{0xD800, '\\', 'u', 'd', '8', '0', '0', '"', '\n', 0, 0x2028, '<', '&', 0xDFFF},
	} {
		text, ok := quoteModulePath(ecmascript.StringFromCodeUnits(units))
		if !ok || !utf8.ValidString(text) {
			t.Fatalf("code units %x produced an invalid source string %q", units, text)
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/quoted.js", Path: "/quoted.js"}, text, core.ScriptKindJS)
		value := file.Statements.Nodes[0].AsExpressionStatement().Expression.Text()
		if got := ecmascript.StringCodeUnits(value); !reflect.DeepEqual(got, units) {
			t.Errorf("quoted %x as %s, parsed back as %x", units, text, got)
		}
	}
	for _, value := range []string{"\xff", "\xed\xa0", "\xed\xa0\x80\xff"} {
		if _, ok := quoteModulePath(value); ok {
			t.Errorf("invalid bytes %x must not produce a lossy fix", value)
		}
	}
}
