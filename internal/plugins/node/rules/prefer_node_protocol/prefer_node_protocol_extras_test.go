package prefer_node_protocol_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_node_protocol"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNodeProtocolEditDemand(t *testing.T) {
	program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(
		`import "fs"; require("path"); process.getBuiltinModule("util");`, "input.js", "tsconfig.allowJs.json")
	if err != nil {
		t.Fatal(err)
	}
	r := prefer_node_protocol.PreferNodeProtocolRule
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: r.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		return diagnostics
	}
	all := run(rule.EditDemandAll)
	if len(all) != 3 {
		t.Fatalf("got %d diagnostics, want 3", len(all))
	}
	for _, d := range all {
		if d.FixesPtr == nil || len(*d.FixesPtr) != 1 {
			t.Fatal("expected one insertion per diagnostic")
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		got := run(demand)
		if len(got) != len(all) {
			t.Fatalf("demand %d: wrong diagnostic count", demand)
		}
		for i, d := range got {
			if d.Range != all[i].Range || !reflect.DeepEqual(d.Message, all[i].Message) {
				t.Fatal("edit demand changed diagnostic")
			}
			if d.Suggestions != nil {
				t.Fatal("unexpected suggestions")
			}
			if demand&rule.EditDemandAutofix != 0 {
				if !reflect.DeepEqual(d.FixesPtr, all[i].FixesPtr) {
					t.Fatal("edit demand changed fix")
				}
			} else if d.FixesPtr != nil {
				t.Fatal("fix computed without autofix demand")
			}
		}
	}
}

func TestPreferNodeProtocolSchema(t *testing.T) {
	for _, options := range [][]any{{true}, {"16"}, {map[string]any{"version": 16}}, {map[string]any{"unknown": true}}, {map[string]any{}, map[string]any{}}} {
		if err := prefer_node_protocol.PreferNodeProtocolRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options: %#v", options)
		}
	}
}

func TestPreferNodeProtocolOversizedVersion(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		version  string
		esm, cjs bool
	}{
		{"<=4294967296", false, false},
		{">=12 <4294967296", false, false},
		{"~12.4294967295.0", true, false},
		{"^0.0.4294967295", false, false},
		{">4294967295", true, true},
		{">=16 <4294967296", true, true},
		{"14.13.4294967296", true, false},
		{"9007199254740991.0.0", true, true},
	} {
		output := `import "fs";` + "\n" + `require("fs");` + "\n" + `process.getBuiltinModule("node:fs");`
		var errors []rule_tester.InvalidTestCaseError
		if test.esm {
			output = strings.Replace(output, `import "fs"`, `import "node:fs"`, 1)
			errors = append(errors, rule_tester.InvalidTestCaseError{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12})
		}
		if test.cjs {
			output = strings.Replace(output, `require("fs")`, `require("node:fs")`, 1)
			errors = append(errors, rule_tester.InvalidTestCaseError{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 13})
		}
		errors = append(errors, rule_tester.InvalidTestCaseError{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 3, Column: 26, EndLine: 3, EndColumn: 30})
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:     `import "fs";` + "\n" + `require("fs");` + "\n" + `process.getBuiltinModule("fs");`,
			Options:  map[string]any{"version": test.version},
			Settings: map[string]any{"node": map[string]any{"version": ">=16"}},
			Output:   []string{output}, Errors: errors,
		})
	}
	runProtocolTests(t, nil, invalid)
}

func TestPreferNodeProtocolBindings(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `function f(require, process, globalThis) { require("fs"); process.getBuiltinModule("path"); globalThis.process.getBuiltinModule("util"); }`},
		{Code: `function require(value) { return value; } require("fs");`},
		{Code: `{ require("fs"); let require = custom; }`},
		{Code: `try {} catch (process) { process.getBuiltinModule("fs"); }`},
		{Code: `const {require, process} = custom; require("fs"); process.getBuiltinModule("fs");`},
		{Code: `import require from "custom"; import process from "custom"; require("fs"); process.getBuiltinModule("fs");`},
		{Code: `import {createRequire} from "custom"; const require = createRequire(import.meta.url); require("fs");`},
		{Code: `import {createRequire} from "node:module"; function f(createRequire) { const require = createRequire(import.meta.url); require("fs"); }`},
		{Code: `import {createRequire} from "node:module"; let require = createRequire(import.meta.url); require = custom; require("fs");`},
		{Code: `const require = custom; const process = require("node:process"); process.getBuiltinModule("fs");`},
		{Code: `let process = require("node:process"); process = custom; process.getBuiltinModule("fs");`},
		{Code: `import type process from "node:process"; process.getBuiltinModule("fs");`, FileName: "input.ts"},
		{Code: `import type {createRequire} from "node:module"; const require = createRequire(import.meta.url); require("fs");`, FileName: "input.ts"},
		{Code: `declare const process: any; declare const require: any; process.getBuiltinModule("fs"); require("fs");`, FileName: "input.ts"},
		{Code: `function f(require) { require("fs"); }`, LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, FileName: "input.cjs"},
		// Module specifiers are exact names, not host filesystem paths.
		{Code: `import process from "./process"; process.getBuiltinModule("fs");`},
		{Code: `import {createRequire} from "C:/module"; const require = createRequire(import.meta.url); require("fs");`},
		{Code: `require("./fs"); require("C:\\fs"); process.getBuiltinModule("\\\\server\\fs"); require("fs\\promises"); import("file:///fs");`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, prefix := range []string{
		`import {createRequire} from "node:module"; const require = createRequire(import.meta.url);`,
		`import {createRequire as makeRequire} from "node:module"; const require = makeRequire(import.meta.url);`,
		`import module from "node:module"; const require = module.createRequire(import.meta.url);`,
		`import * as module from "node:module"; const require = module["createRequire"](import.meta.url);`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: prefix + "\n" + `require("fs");`, Output: []string{prefix + "\n" + `require("node:fs");`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 13}},
		})
	}
	for _, prefix := range []string{
		`import process from "node:process";`,
		`import * as process from "node:process";`,
		`import {default as process} from "node:process";`,
		`const process = require("node:process");`,
		`interface process { custom: string }`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: prefix + "\n" + `process.getBuiltinModule("fs");`, FileName: "input.ts",
			Output: []string{prefix + "\n" + `process.getBuiltinModule("node:fs");`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 2, Column: 26, EndLine: 2, EndColumn: 30}},
		})
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: `require("fs");`, FileName: "input.cjs", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		Output: []string{`require("node:fs");`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}},
	})
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: `const process = require("node:process");` + "\n" + `process.getBuiltinModule("fs");`, FileName: "input.cjs", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
		Output: []string{`const process = require("node:process");` + "\n" + `process.getBuiltinModule("node:fs");`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 2, Column: 26, EndLine: 2, EndColumn: 30}},
	})
	runProtocolTests(t, valid, invalid)
}

func TestPreferNodeProtocolEmptyAlternatives(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, version := range []string{">=16 || >20 <16", ">20 <16 || >=16"} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: `import "fs"; require("fs");`, Options: map[string]any{"version": version},
			Output: []string{`import "node:fs"; require("node:fs");`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12},
				{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
			},
		})
	}
	runProtocolTests(t, nil, invalid)
}

// Expectations compared with eslint-plugin-n v18.3.0, including complete ranges and fixes.
func TestPreferNodeProtocolAst(t *testing.T) {
	runProtocolTests(t, []rule_tester.ValidTestCase{
		{Code: "(require!)(\"fs\"); require((\"fs\" satisfies string));", FileName: "input.ts"},
		{Code: "process[String(\"getBuiltinModule\")](\"fs\");"},
		{Code: "process[(() => \"getBuiltinModule\")()](\"fs\");"},
		{Code: "new require(\"fs\"); new process.getBuiltinModule(\"fs\"); require.call(null, \"fs\");"},
		{Code: "require?.(\"fs\"); const r = require; r(\"fs\"); require.resolve(\"fs\");"},
		{Code: "(process?.getBuiltinModule)(\"fs\"); (globalThis?.process).getBuiltinModule(\"fs\");"},
		{Code: "const key = \"getBuiltinModule\"; process[key](\"fs\");"},
		{Code: "global.process.getBuiltinModule(\"fs\"); foo.process.getBuiltinModule(\"fs\"); process.foo.getBuiltinModule(\"fs\");"},
		{Code: "process.getBuiltinModule(); process.getBuiltinModule(...[\"fs\"]); process.getBuiltinModule(`fs`);"},
		{Code: "import fs = require(\"fs\"); type T = import(\"fs\").Stats; type U = typeof import(\"path\");", FileName: "input.ts"},
		{Code: "(process as any).getBuiltinModule(\"fs\"); process!.getBuiltinModule(\"fs\");", FileName: "input.ts"},
		{Code: "class X { #getBuiltinModule() {} f() { this.#getBuiltinModule(\"fs\"); } }"},
		{Code: "export const value = 1; export {value};"},
		{Code: "import \"fs!loader\"; require(\"fs!loader\"); process.getBuiltinModule(\"fs!loader\");"},
		{Code: "require(\"fs\" + \"\"); import(`fs`); process.getBuiltinModule(12);"},
	}, []rule_tester.InvalidTestCase{
		{Code: "require<string>(\"fs\"); (require<string>)(\"fs\");", FileName: "input.ts", Output: []string{"require<string>(\"node:fs\"); (require<string>)(\"fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21}}},
		{Code: "process.getBuiltinModule<string>(\"fs\"); (process.getBuiltinModule<string>)(\"fs\");", FileName: "input.ts", Output: []string{"process.getBuiltinModule<string>(\"node:fs\"); (process.getBuiltinModule<string>)(\"fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 38}}},
		{Code: "process[(\"getBuiltinModule\" satisfies string)](\"fs\");", FileName: "input.ts", Output: []string{"process[(\"getBuiltinModule\" satisfies string)](\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 52}}},
		{Code: "process[[\"getBuiltinModule\"]](\"fs\");", Output: []string{"process[[\"getBuiltinModule\"]](\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35}}},
		{Code: "process[[\"getBuiltin\", \"Module\"].join(\"\")](\"fs\");", Output: []string{"process[[\"getBuiltin\", \"Module\"].join(\"\")](\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 48}}},
		{Code: "/* 😀 */ require(\"\\u0066s\");", Output: []string{"/* 😀 */ require(\"node:\\u0066s\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 27}}},
		{Code: "require(\"f\\\r\ns\");", Output: []string{"require(\"node:f\\\r\ns\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 9, EndLine: 2, EndColumn: 3}}},
		{Code: "require(/** @satisfies {string} */ (\"fs\"));", Output: []string{"require(/** @satisfies {string} */ (\"node:fs\"));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 41}}},
		{Code: "(/** @type {any} */ (process)).getBuiltinModule(\"fs\");", Output: []string{"(/** @type {any} */ (process)).getBuiltinModule(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 49, EndLine: 1, EndColumn: 53}}},
		{Code: "req\\u0075ire(\"fs\"); proc\\u0065ss.getBuiltinModule(\"path\");", Output: []string{"req\\u0075ire(\"node:fs\"); proc\\u0065ss.getBuiltinModule(\"node:path\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 51, EndLine: 1, EndColumn: 57}}},
		{Code: "import \"node:fs\"; require(\"pkg\"); import \"fs\"; require(\"path\");", Output: []string{"import \"node:fs\"; require(\"pkg\"); import \"node:fs\"; require(\"node:path\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 42, EndLine: 1, EndColumn: 46}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 56, EndLine: 1, EndColumn: 62}}},
		{Code: "process.getBuiltinModule(\"fs\"); import \"path\"; require(\"util\");", Options: map[string]any{"version": ">=10"}, Output: []string{"process.getBuiltinModule(\"node:fs\"); import \"path\"; require(\"util\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30}}},
		{Code: "process.getBuiltinModule(\"fs\"); import \"path\"; require(\"util\");", Options: map[string]any{"version": "12.20.0"}, Output: []string{"process.getBuiltinModule(\"node:fs\"); import \"node:path\"; require(\"util\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 46}}},
		{Code: "require(\"fs\"); import \"path\"; process.getBuiltinModule(\"util\");", Options: map[string]any{"version": ">=14.18.0"}, Output: []string{"require(\"fs\"); import \"node:path\"; process.getBuiltinModule(\"node:util\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:util` over `util`.", Line: 1, Column: 56, EndLine: 1, EndColumn: 62}}},
		{Code: "require((\"fs\")); (require)(\"path\");", Output: []string{"require((\"node:fs\")); (require)(\"node:path\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 14}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 34}}},
		{Code: "process?.getBuiltinModule?.(\"fs\"); globalThis?.process?.getBuiltinModule?.(\"path\", extra);", Output: []string{"process?.getBuiltinModule?.(\"node:fs\"); globalThis?.process?.getBuiltinModule?.(\"node:path\", extra);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 76, EndLine: 1, EndColumn: 82}}},
		{Code: "(globalThis.process).getBuiltinModule((\"fs\")); ((process)).getBuiltinModule(\"path\");", Output: []string{"(globalThis.process).getBuiltinModule((\"node:fs\")); ((process)).getBuiltinModule(\"node:path\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 44}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 77, EndLine: 1, EndColumn: 83}}},
		{Code: "process[\"getBuiltin\" + \"Module\"](\"fs\"); globalThis[`process`][`getBuiltinModule`](\"path\");", Output: []string{"process[\"getBuiltin\" + \"Module\"](\"node:fs\"); globalThis[`process`][`getBuiltinModule`](\"node:path\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 38}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 83, EndLine: 1, EndColumn: 89}}},
		{Code: "require(\"fs\" as string); (require as Function)(\"fs\"); process[\"getBuiltinModule\" as const](\"fs\");", FileName: "input.ts", Output: []string{"require(\"fs\" as string); (require as Function)(\"fs\"); process[\"getBuiltinModule\" as const](\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 92, EndLine: 1, EndColumn: 96}}},
		{Code: "export * from \"fs\"; export * as path from \"path\";", Output: []string{"export * from \"node:fs\"; export * as path from \"node:path\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 49}}},
		{Code: "import type {Stats} from \"fs\"; export type {Stats} from \"fs\"; export type * from \"path\";", FileName: "input.ts", Output: []string{"import type {Stats} from \"node:fs\"; export type {Stats} from \"node:fs\"; export type * from \"node:path\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 57, EndLine: 1, EndColumn: 61}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 82, EndLine: 1, EndColumn: 88}}},
		{Code: "import fs from \"fs\" with {type:\"json\"}; import(\"path\", {with:{type:\"json\"}});", Output: []string{"import fs from \"node:fs\" with {type:\"json\"}; import(\"node:path\", {with:{type:\"json\"}});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 54}}},
		{Code: "const element = <div>{require(\"fs\")}</div>;", FileName: "input.tsx", Output: []string{"const element = <div>{require(\"node:fs\")}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35}}},
		{Code: "/** @type {import(\"fs\").Stats} */ const stats = {};\n/** @import {Stats} from \"fs\" */\nimport \"path\";", Output: []string{"/** @type {import(\"fs\").Stats} */ const stats = {};\n/** @import {Stats} from \"fs\" */\nimport \"node:path\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 3, Column: 8, EndLine: 3, EndColumn: 14}}},
		{Code: "require(/** @type {string} */ (\"fs\"));", Output: []string{"require(/** @type {string} */ (\"node:fs\"));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36}}},
		{Code: "/* 😀 */ require(\n /* retain */ \"\\x66s\");", Output: []string{"/* 😀 */ require(\n /* retain */ \"node:\\x66s\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 2, Column: 15, EndLine: 2, EndColumn: 22}}},
	})
}

// Expectations compared with eslint-plugin-n v18.3.0, including complete ranges and fixes.
func TestPreferNodeProtocolOptions(t *testing.T) {
	runProtocolTests(t, []rule_tester.ValidTestCase{
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": ">=12.20.0"}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": ">=10"}, Settings: map[string]any{"node": map[string]any{"version": ">=16"}}},
		{Code: "import \"fs\"; require(\"fs\");", Settings: map[string]any{"node": map[string]any{"version": ">=10"}}},
		{Code: "import \"fs\"; require(\"fs\");", Settings: map[string]any{"n": map[string]any{"version": ">=10"}, "node": map[string]any{"version": ">=16"}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": "invalid"}, Settings: map[string]any{"n": map[string]any{"version": "invalid"}, "node": map[string]any{"version": ">=10"}}},
		{Code: "import \"fs\"; require(\"fs\");", Settings: map[string]any{"node": map[string]any{"version": float64(12)}}},
		{Code: "import \"fs\"; require(\"fs\");", Settings: map[string]any{"node": map[string]any{"version": []any{}}}},
	}, []rule_tester.InvalidTestCase{
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{}, Output: []string{"import \"node:fs\"; require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": ">=16.0.0"}, Output: []string{"import \"node:fs\"; require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": "^12.20.0 || ^14.18.0 || >=16"}, Output: []string{"import \"node:fs\"; require(\"fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		{Code: "import \"fs\"; require(\"fs\"); process.getBuiltinModule(\"fs\");", Options: map[string]any{"version": "*"}, Output: []string{"import \"fs\"; require(\"fs\"); process.getBuiltinModule(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 58}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": ">=14.18.0"}, Output: []string{"import \"node:fs\"; require(\"fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": "^14.18.0 || >=16.0.0"}, Output: []string{"import \"node:fs\"; require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": "invalid"}, Output: []string{"import \"node:fs\"; require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26}}},
		{Code: "import \"fs\"; require(\"fs\");", Options: map[string]any{"version": ""}, Output: []string{"import \"node:fs\"; require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}, {MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26}}},
	})
}

// Expectations compared with eslint-plugin-n v18.3.0, including complete ranges and fixes.
func TestPreferNodeProtocolBuiltins(t *testing.T) {
	runProtocolTests(t, []rule_tester.ValidTestCase{
		{Code: "import \"domain\";"},
		{Code: "import \"punycode\";"},
		{Code: "import \"test\";"},
		{Code: "import \"test/reporters\";"},
		{Code: "import \"sea\";"},
		{Code: "import \"sqlite\";"},
		{Code: "import \"quic\";"},
		{Code: "import \"constants\";"},
		{Code: "import \"sys\";"},
		{Code: "import \"_http_agent\";"},
		{Code: "import \"_stream_wrap\";"},
		{Code: "import \"stream/iter\";"},
		{Code: "import \"zlib/iter\";"},
		{Code: "import \"node:fs\";"},
	}, []rule_tester.InvalidTestCase{
		{Code: "import \"assert\";", Output: []string{"import \"node:assert\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:assert` over `assert`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"assert/strict\";", Output: []string{"import \"node:assert/strict\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:assert/strict` over `assert/strict`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		{Code: "import \"async_hooks\";", Output: []string{"import \"node:async_hooks\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:async_hooks` over `async_hooks`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
		{Code: "import \"buffer\";", Output: []string{"import \"node:buffer\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:buffer` over `buffer`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"child_process\";", Output: []string{"import \"node:child_process\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:child_process` over `child_process`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		{Code: "import \"cluster\";", Output: []string{"import \"node:cluster\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:cluster` over `cluster`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		{Code: "import \"console\";", Output: []string{"import \"node:console\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:console` over `console`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		{Code: "import \"crypto\";", Output: []string{"import \"node:crypto\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:crypto` over `crypto`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"dgram\";", Output: []string{"import \"node:dgram\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:dgram` over `dgram`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
		{Code: "import \"diagnostics_channel\";", Output: []string{"import \"node:diagnostics_channel\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:diagnostics_channel` over `diagnostics_channel`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},
		{Code: "import \"dns\";", Output: []string{"import \"node:dns\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:dns` over `dns`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
		{Code: "import \"dns/promises\";", Output: []string{"import \"node:dns/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:dns/promises` over `dns/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
		{Code: "import \"events\";", Output: []string{"import \"node:events\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:events` over `events`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"fs\";", Output: []string{"import \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		{Code: "import \"fs/promises\";", Output: []string{"import \"node:fs/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
		{Code: "import \"http2\";", Output: []string{"import \"node:http2\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:http2` over `http2`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
		{Code: "import \"http\";", Output: []string{"import \"node:http\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:http` over `http`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
		{Code: "import \"https\";", Output: []string{"import \"node:https\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:https` over `https`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
		{Code: "import \"inspector\";", Output: []string{"import \"node:inspector\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:inspector` over `inspector`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
		{Code: "import \"inspector/promises\";", Output: []string{"import \"node:inspector/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:inspector/promises` over `inspector/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		{Code: "import \"module\";", Output: []string{"import \"node:module\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:module` over `module`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"net\";", Output: []string{"import \"node:net\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:net` over `net`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
		{Code: "import \"os\";", Output: []string{"import \"node:os\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:os` over `os`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		{Code: "import \"path\";", Output: []string{"import \"node:path\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:path` over `path`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
		{Code: "import \"path/posix\";", Output: []string{"import \"node:path/posix\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:path/posix` over `path/posix`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		{Code: "import \"path/win32\";", Output: []string{"import \"node:path/win32\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:path/win32` over `path/win32`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		{Code: "import \"perf_hooks\";", Output: []string{"import \"node:perf_hooks\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:perf_hooks` over `perf_hooks`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		{Code: "import \"process\";", Output: []string{"import \"node:process\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:process` over `process`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		{Code: "import \"querystring\";", Output: []string{"import \"node:querystring\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:querystring` over `querystring`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
		{Code: "import \"readline\";", Output: []string{"import \"node:readline\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:readline` over `readline`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
		{Code: "import \"readline/promises\";", Output: []string{"import \"node:readline/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:readline/promises` over `readline/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
		{Code: "import \"repl\";", Output: []string{"import \"node:repl\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:repl` over `repl`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
		{Code: "import \"stream\";", Output: []string{"import \"node:stream\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:stream` over `stream`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"stream/promises\";", Output: []string{"import \"node:stream/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:stream/promises` over `stream/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
		{Code: "import \"stream/web\";", Output: []string{"import \"node:stream/web\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:stream/web` over `stream/web`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		{Code: "import \"stream/consumers\";", Output: []string{"import \"node:stream/consumers\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:stream/consumers` over `stream/consumers`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
		{Code: "import \"string_decoder\";", Output: []string{"import \"node:string_decoder\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:string_decoder` over `string_decoder`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		{Code: "import \"timers\";", Output: []string{"import \"node:timers\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:timers` over `timers`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"timers/promises\";", Output: []string{"import \"node:timers/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:timers/promises` over `timers/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
		{Code: "import \"tls\";", Output: []string{"import \"node:tls\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:tls` over `tls`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
		{Code: "import \"trace_events\";", Output: []string{"import \"node:trace_events\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:trace_events` over `trace_events`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
		{Code: "import \"tty\";", Output: []string{"import \"node:tty\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:tty` over `tty`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
		{Code: "import \"url\";", Output: []string{"import \"node:url\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:url` over `url`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
		{Code: "import \"util\";", Output: []string{"import \"node:util\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:util` over `util`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
		{Code: "import \"util/types\";", Output: []string{"import \"node:util/types\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:util/types` over `util/types`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
		{Code: "import \"v8\";", Output: []string{"import \"node:v8\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:v8` over `v8`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		{Code: "import \"vm\";", Output: []string{"import \"node:vm\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:vm` over `vm`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		{Code: "import \"wasi\";", Output: []string{"import \"node:wasi\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:wasi` over `wasi`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
		{Code: "import \"worker_threads\";", Output: []string{"import \"node:worker_threads\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:worker_threads` over `worker_threads`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		{Code: "import \"zlib\";", Output: []string{"import \"node:zlib\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:zlib` over `zlib`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14}}},
	})
}
