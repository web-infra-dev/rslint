package enforce_node_protocol_usage_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/enforce_node_protocol_usage"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEnforceNodeProtocolUsageSchema(t *testing.T) {
	for _, options := range [][]any{nil, {}, {"sometimes"}, {true}, {map[string]any{}}, {"always", "never"}} {
		if err := target.EnforceNodeProtocolUsageRule.Schema.Validate(options); err == nil {
			t.Fatalf("accepted invalid options: %v", options)
		}
	}
}

func TestEnforceNodeProtocolUsageEditDemand(t *testing.T) {
	r := &target.EnforceNodeProtocolUsageRule
	for _, tc := range []struct{ mode, code, output string }{
		{"always", `import "fs";`, `import "node:fs";`},
		{"never", `import "node:fs";`, `import "fs";`},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/edit-demand.ts", Path: "/edit-demand.ts"}, tc.code, core.ScriptKindTS)
			options := rule_tester.ResolveTestCaseOptions(t, r, []any{tc.mode})
			var all rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
				})
				r.Run(ctx, options)[ast.KindImportDeclaration](file.Statements.Nodes[0])
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
				}
				got := diagnostics[0]
				if got.Message.Id != "" {
					t.Fatalf("unexpected message ID: %q", got.Message.Id)
				}
				if demand == rule.EditDemandAll {
					all = got
				}
				if got.Suggestions != nil {
					t.Fatal("unexpected suggestions")
				}
				if demand&rule.EditDemandAutofix != 0 {
					if got.FixesPtr == nil || !reflect.DeepEqual(got.FixesPtr, all.FixesPtr) {
						t.Fatalf("demand %d: missing or changed autofix", demand)
					}
					output, _, fixed := linter.ApplyRuleFixes(tc.code, diagnostics)
					if !fixed || output != tc.output {
						t.Fatalf("demand %d: got %q, want %q", demand, output, tc.output)
					}
				} else if got.FixesPtr != nil {
					t.Fatalf("demand %d: unexpected autofix", demand)
				}
				want := all
				got.FixesPtr, want.FixesPtr = nil, nil
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("demand %d changed the diagnostic", demand)
				}
			}
		})
	}
}

func TestEnforceNodeProtocolUsageInvalidVersion(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, mode := range []string{"always", "never"} {
		for _, version := range []any{"bad", "16", 16, nil, "4294967296.0.0"} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code:     "require(variable);\nimport 'fs';\nexport { readFile } from 'fs';\nrequire('node:path');\nimport('stream');",
				Options:  []any{mode},
				Settings: map[string]any{"import/node-version": version},
				Errors: []rule_tester.InvalidTestCaseError{{
					Message: "`import/node-version` setting must be a string in the format \"10.23.45\" (a semver version, with no leading zero)",
					Line:    2, Column: 8, EndLine: 2, EndColumn: 12,
				}},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &target.EnforceNodeProtocolUsageRule, nil, invalid)
}

func TestEnforceNodeProtocolUsageInvalidVersionWithDirectives(t *testing.T) {
	r := &target.EnforceNodeProtocolUsageRule
	for _, mode := range []string{"always", "never"} {
		for _, tc := range []struct {
			name, code string
			line       int
		}{
			{"next line", "// eslint-disable-next-line import/enforce-node-protocol-usage\nimport 'fs';\nimport 'fs';\nimport 'path';", 3},
			{"same line", "import 'fs'; // eslint-disable-line import/enforce-node-protocol-usage\nimport 'fs';", 2},
			{"re-enabled", "/* eslint-disable import/enforce-node-protocol-usage */\nimport 'fs';\n/* eslint-enable import/enforce-node-protocol-usage */\nimport 'fs';", 4},
			{"rslint prefix", "// rslint-disable-next-line import/enforce-node-protocol-usage\nimport 'fs';\nimport 'fs';", 3},
			{"unrelated rule", "// eslint-disable-next-line no-console\nimport 'fs';\nimport 'path';", 2},
			{"literal position", "import\n// eslint-disable-next-line import/enforce-node-protocol-usage\n'fs';\nimport 'fs';", 4},
			{"only disabled reference", "// eslint-disable-next-line import/enforce-node-protocol-usage\nimport 'fs';", 0},
			{"whole file disabled", "/* eslint-disable */\nimport 'fs';\nimport 'path';", 0},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/directives.ts", Path: "/directives.ts"}, tc.code, core.ScriptKindTS)
				comments := rule.NewCommentStore(file)
				var diagnostics []rule.RuleDiagnostic
				// Bind the actual name: RunRuleTester registers its rule as "test".
				ctx := (rule.RuleContext{
					SourceFile: file, Settings: map[string]any{"import/node-version": "bad"},
					DisableManager: rule.NewDisableManager(file, comments),
				}).WithReporter(r.Name, rule.SeverityError, func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				})
				listener := r.Run(ctx, []any{mode})[ast.KindImportDeclaration]
				for _, node := range file.Statements.Nodes {
					listener(node)
				}
				if tc.line == 0 {
					if len(diagnostics) != 0 {
						t.Fatalf("disabled references reported %d diagnostics", len(diagnostics))
					}
					return
				}
				if len(diagnostics) != 1 {
					t.Fatalf("got %d diagnostics, want one at the first enabled reference", len(diagnostics))
				}
				got := diagnostics[0]
				line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(file, got.Range.Pos())
				endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(file, got.Range.End())
				if line+1 != tc.line || column+1 != 8 || endLine+1 != tc.line || endColumn+1 != 12 {
					t.Fatalf("range = %d:%d-%d:%d, want %d:8-%d:12", line+1, column+1, endLine+1, endColumn+1, tc.line, tc.line)
				}
				if got.Message.Id != "" || got.Message.Description != "`import/node-version` setting must be a string in the format \"10.23.45\" (a semver version, with no leading zero)" || got.FixesPtr != nil || got.Suggestions != nil {
					t.Fatalf("unexpected configuration diagnostic: %+v", got)
				}
			})
		}
	}
}

// Expectations checked with eslint-plugin-import v2.32.0 and is-core-module v2.17.0.
// Covers parentheses, optional/computed/private access, JSX, TypeScript, JSDoc,
// raw literal text, UTF-16 ranges and version-specific builtin availability.
func TestEnforceNodeProtocolUsageExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &target.EnforceNodeProtocolUsageRule,
		[]rule_tester.ValidTestCase{
			// Like upstream, version validation is deferred until a source is checked.
			{Code: "export const value = 1;", Options: []any{"always"}, Settings: map[string]any{"import/node-version": "invalid"}},
			{
				Code: "export * from 'fs';", Options: []any{"always"},
			},
			{
				Code: "export * as ns from 'fs';", Options: []any{"always"},
			},
			{
				Code: "import fs = require('fs');", Options: []any{"always"},
			},
			{
				Code: "type FS = typeof import('fs');", Options: []any{"always"},
			},
			{
				Code: "require?.('fs');", Options: []any{"always"},
			},
			{
				Code: "require!('fs');", Options: []any{"always"},
			},
			{
				Code: "require('fs' as string);", Options: []any{"always"},
			},
			{
				Code: "obj['require']('fs');", Options: []any{"always"},
			},
			{
				Code: "class C { #require() {} method() { this.#require('fs'); } }", Options: []any{"always"},
			},
			{
				Code: "export const value = 'fs'; export { value };", Options: []any{"always"},
			},
			{
				Code: "process.getBuiltinModule('fs');", Options: []any{"always"},
			},
			{
				Code: "/** @import { Stats } from 'fs' */\n/** @type {import('fs').Stats} */\nlet stats;", Options: []any{"always"},
				FileName: "case.js",
			},
			{
				Code: "export * from 'node:fs';", Options: []any{"never"},
			},
			{
				Code: "export * as ns from 'node:fs';", Options: []any{"never"},
			},
			{
				Code: "import fs = require('node:fs');", Options: []any{"never"},
			},
			{
				Code: "type FS = typeof import('node:fs');", Options: []any{"never"},
			},
			{
				Code: "require?.('node:fs');", Options: []any{"never"},
			},
			{
				Code: "require!('node:fs');", Options: []any{"never"},
			},
			{
				Code: "require('node:fs' as string);", Options: []any{"never"},
			},
			{
				Code: "obj['require']('node:fs');", Options: []any{"never"},
			},
			{
				Code: "class C { #require() {} method() { this.#require('node:fs'); } }", Options: []any{"never"},
			},
			{
				Code: "export const value = 'node:fs'; export { value };", Options: []any{"never"},
			},
			{
				Code: "process.getBuiltinModule('node:fs');", Options: []any{"never"},
			},
			{
				Code: "/** @import { Stats } from 'node:fs' */\n/** @type {import('node:fs').Stats} */\nlet stats;", Options: []any{"never"},
				FileName: "case.js",
			},
			{
				Code: "import \"\";", Options: []any{"always"},
			},
			{
				Code: "import \"Node:fs\";", Options: []any{"always"},
			},
			{
				Code: "import \"node:unknown\";", Options: []any{"never"},
			},
			{
				Code: "import \"node:fs/not-a-builtin\";", Options: []any{"never"},
			},
			{
				Code: "import \"test\";", Options: []any{"always"},
			},
			{
				Code: "import \"node:test\";", Options: []any{"never"},
			},
			{
				Code: "import \"custom-core\";", Options: []any{"always"},
				Settings: map[string]any{"import/core-modules": []any{"custom-core"}},
			},
			{
				Code: "import \"node:custom-core\";", Options: []any{"never"},
				Settings: map[string]any{"import/core-modules": []any{"custom-core"}},
			},
			{
				Code: "import 'fs';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "14.17.0"},
			},
			{
				Code: "import 'fs';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "15.9.0"},
			},
			{
				Code: "import 'fs/promises';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "10.0.0"},
			},
			{
				Code: "import 'node:fs/promises';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "10.1.0"},
			},
			{
				Code: "import 'assert/strict';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "14.18.0"},
			},
			{
				Code: "import 'stream/web';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "16.4.0"},
			},
			{
				Code: "import 'node:stream/web';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "16.4.0"},
			},
			{
				Code: "import 'node:test/reporters';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "20.2.0"},
			},
			{
				Code: "import '_stream_wrap';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "26.0.0"},
			},
			{
				Code: "import 'node:_stream_wrap';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "26.0.0"},
			},
			{
				Code: "import '_debugger';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "7.9.0"},
			},
			{
				Code: "import 'inspector/promises';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "18.0.0"},
			},
			{
				Code: "import 'wasi';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "19.0.0"},
			},
		}, []rule_tester.InvalidTestCase{
			// Upstream removes five raw characters even when the prefix is escaped.
			// Preserve the builtin target rather than leaving part of the raw prefix.
			// cspell:ignore eode
			{
				Code: `import "\x6eode:fs";`, Options: []any{"never"}, Output: []string{`import "fs";`},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 1, 20)},
			},
			{
				Code: `import "n\u006fde:fs";`, Options: []any{"never"}, Output: []string{`import "fs";`},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 1, 22)},
			},
			{
				Code: "(require)(('fs'));", Options: []any{"always"},
				Output: []string{"(require)(('node:fs'));"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 12, 1, 16)},
			},
			{
				Code: "import( /* source */ ('fs'));", Options: []any{"always"},
				Output: []string{"import( /* source */ ('node:fs'));"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 23, 1, 27)},
			},
			{
				Code: "import('fs', { with: { type: 'json' } });", Options: []any{"always"},
				Output: []string{"import('node:fs', { with: { type: 'json' } });"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 8, 1, 12)},
			},
			{
				Code: "require<string>('fs');", Options: []any{"always"},
				Output: []string{"require<string>('node:fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 17, 1, 21)},
			},
			{
				Code: "const require = (x) => x; require('fs');", Options: []any{"always"},
				Output: []string{"const require = (x) => x; require('node:fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 35, 1, 39)},
			},
			{
				Code: "require('fs')?.readFile;", Options: []any{"always"},
				Output: []string{"require('node:fs')?.readFile;"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 9, 1, 13)},
			},
			{
				Code: "import type { Stats } from 'fs';", Options: []any{"always"},
				Output: []string{"import type { Stats } from 'node:fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 28, 1, 32)},
			},
			{
				Code: "export type { Stats } from 'fs';", Options: []any{"always"},
				Output: []string{"export type { Stats } from 'node:fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 28, 1, 32)},
			},
			{
				Code: "export {} from 'fs';", Options: []any{"always"},
				Output: []string{"export {} from 'node:fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 16, 1, 20)},
			},
			{
				Code: "const face = '😀'; require('fs');", Options: []any{"always"},
				Output: []string{"const face = '😀'; require('node:fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 28, 1, 32)},
			},
			{
				Code: "import {\n  readFile,\n} from /* source */ 'fs';", Options: []any{"always"},
				Output: []string{"import {\n  readFile,\n} from /* source */ 'node:fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 3, 21, 3, 25)},
			},
			{
				Code: "const element = <Box value={require('fs')} />;", Options: []any{"always"},
				FileName: "case.tsx",
				Output:   []string{"const element = <Box value={require('node:fs')} />;"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 37, 1, 41)},
			},
			{
				Code: "require(/** @type {string} */ ('fs'));", Options: []any{"always"},
				FileName: "case.js",
				Output:   []string{"require(/** @type {string} */ ('node:fs'));"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 32, 1, 36)},
			},
			{
				Code: "(require)(('node:fs'));", Options: []any{"never"},
				Output: []string{"(require)(('fs'));"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 12, 1, 21)},
			},
			{
				Code: "import( /* source */ ('node:fs'));", Options: []any{"never"},
				Output: []string{"import( /* source */ ('fs'));"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 23, 1, 32)},
			},
			{
				Code: "import('node:fs', { with: { type: 'json' } });", Options: []any{"never"},
				Output: []string{"import('fs', { with: { type: 'json' } });"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 1, 17)},
			},
			{
				Code: "require<string>('node:fs');", Options: []any{"never"},
				Output: []string{"require<string>('fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 17, 1, 26)},
			},
			{
				Code: "const require = (x) => x; require('node:fs');", Options: []any{"never"},
				Output: []string{"const require = (x) => x; require('fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 35, 1, 44)},
			},
			{
				Code: "require('node:fs')?.readFile;", Options: []any{"never"},
				Output: []string{"require('fs')?.readFile;"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 9, 1, 18)},
			},
			{
				Code: "import type { Stats } from 'node:fs';", Options: []any{"never"},
				Output: []string{"import type { Stats } from 'fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 28, 1, 37)},
			},
			{
				Code: "export type { Stats } from 'node:fs';", Options: []any{"never"},
				Output: []string{"export type { Stats } from 'fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 28, 1, 37)},
			},
			{
				Code: "export {} from 'node:fs';", Options: []any{"never"},
				Output: []string{"export {} from 'fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 16, 1, 25)},
			},
			{
				Code: "const face = '😀'; require('node:fs');", Options: []any{"never"},
				Output: []string{"const face = '😀'; require('fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 28, 1, 37)},
			},
			{
				Code: "import {\n  readFile,\n} from /* source */ 'node:fs';", Options: []any{"never"},
				Output: []string{"import {\n  readFile,\n} from /* source */ 'fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 3, 21, 3, 30)},
			},
			{
				Code: "const element = <Box value={require('node:fs')} />;", Options: []any{"never"},
				FileName: "case.tsx",
				Output:   []string{"const element = <Box value={require('fs')} />;"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 37, 1, 46)},
			},
			{
				Code: "require(/** @type {string} */ ('node:fs'));", Options: []any{"never"},
				FileName: "case.js",
				Output:   []string{"require(/** @type {string} */ ('fs'));"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 32, 1, 41)},
			},
			{
				Code: "import \"\\x66s\";", Options: []any{"always"},
				Output: []string{"import \"node:\\x66s\";"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 8, 1, 15)},
			},
			{
				Code: "import \"node:\\x66s\";", Options: []any{"never"},
				Output: []string{"import \"\\x66s\";"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 1, 20)},
			},
			{
				Code: "import 'f\\\ns';", Options: []any{"always"},
				Output: []string{"import 'node:f\\\ns';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 8, 2, 3)},
			},
			{
				Code: "import 'node:f\\\ns';", Options: []any{"never"},
				Output: []string{"import 'f\\\ns';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 2, 3)},
			},
			{
				Code: "import 'fs';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "14.18.0"},
				Output:   []string{"import 'node:fs';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 8, 1, 12)},
			},
			{
				Code: "import 'fs';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "16.0.0"},
				Output:   []string{"import 'node:fs';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 8, 1, 12)},
			},
			{
				Code: "import 'node:fs/promises';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "10.0.0"},
				Output:   []string{"import 'fs/promises';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("never", "fs/promises", 1, 8, 1, 26)},
			},
			{
				Code: "import 'node:assert/strict';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "15.0.0"},
				Output:   []string{"import 'assert/strict';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("never", "assert/strict", 1, 8, 1, 28)},
			},
			{
				Code: "import 'stream/web';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "16.5.0"},
				Output:   []string{"import 'node:stream/web';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "stream/web", 1, 8, 1, 20)},
			},
			{
				Code: "import 'test/reporters';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "19.9.0"},
				Output:   []string{"import 'node:test/reporters';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "test/reporters", 1, 8, 1, 24)},
			},
			{
				Code: "import 'node:test/reporters';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "19.9.0"},
				Output:   []string{"import 'test/reporters';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("never", "test/reporters", 1, 8, 1, 29)},
			},
			{
				Code: "import '_http_agent';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "22.0.0"},
				Output:   []string{"import 'node:_http_agent';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "_http_agent", 1, 8, 1, 21)},
			},
			{
				Code: "import 'node:_debugger';", Options: []any{"never"},
				Settings: map[string]any{"import/node-version": "7.9.0"},
				Output:   []string{"import '_debugger';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("never", "_debugger", 1, 8, 1, 24)},
			},
			{
				Code: "import 'fs';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "014.018.000"},
				Output:   []string{"import 'node:fs';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 8, 1, 12)},
			},
			{
				Code: "import 'inspector/promises';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "19.0.0"},
				Output:   []string{"import 'node:inspector/promises';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "inspector/promises", 1, 8, 1, 28)},
			},
			{
				Code: "import 'wasi';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "18.17.0"},
				Output:   []string{"import 'node:wasi';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "wasi", 1, 8, 1, 14)},
			},
			{
				Code: "import 'wasi';", Options: []any{"always"},
				Settings: map[string]any{"import/node-version": "20.0.0"},
				Output:   []string{"import 'node:wasi';"},
				Errors:   []rule_tester.InvalidTestCaseError{protocolError("always", "wasi", 1, 8, 1, 14)},
			},
		})
}

// Additional AST and literal cases checked against eslint-plugin-import v2.32.0.
func TestEnforceNodeProtocolUsageASTBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &target.EnforceNodeProtocolUsageRule,
		[]rule_tester.ValidTestCase{
			{Code: "new require('fs');", Options: []any{"always"}},
			{Code: "import(('fs' as const));", Options: []any{"always"}},
			{Code: "import('fs'!);", Options: []any{"always"}},
			{Code: "(require as Function)('fs');", Options: []any{"always"}},
			{Code: "(require satisfies Function)('fs');", Options: []any{"always"}},
			{Code: "export type * from 'fs';", Options: []any{"always"}},
			{Code: "import.defer('fs');", Options: []any{"always"}},
			{Code: "new require('node:fs');", Options: []any{"never"}},
			{Code: "import(('node:fs' as const));", Options: []any{"never"}},
			{Code: "import('node:fs'!);", Options: []any{"never"}},
			{Code: "(require as Function)('node:fs');", Options: []any{"never"}},
			{Code: "(require satisfies Function)('node:fs');", Options: []any{"never"}},
			{Code: "export type * from 'node:fs';", Options: []any{"never"}},
			{Code: "import.defer('node:fs');", Options: []any{"never"}},
			{Code: "import \"node:node:node:fs\";", Options: []any{"never"}},
		}, []rule_tester.InvalidTestCase{
			{Code: "import defer * as fs from 'fs';", Options: []any{"always"},
				Output: []string{"import defer * as fs from 'node:fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 27, 1, 31)},
			},
			{Code: "r\\u0065quire('fs');", Options: []any{"always"},
				Output: []string{"r\\u0065quire('node:fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 14, 1, 18)},
			},
			{Code: "require /* callee */ (( /* argument */ 'fs'));", Options: []any{"always"},
				Output: []string{"require /* callee */ (( /* argument */ 'node:fs'));"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 40, 1, 44)},
			},
			{Code: "const face = '😀';\r\nrequire(\r\n  'fs'\r\n);", Options: []any{"always"},
				Output: []string{"const face = '😀';\r\nrequire(\r\n  'node:fs'\r\n);"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 3, 3, 3, 7)},
			},
			{Code: "class C { #fs = require('fs'); }", Options: []any{"always"},
				Output: []string{"class C { #fs = require('node:fs'); }"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 25, 1, 29)},
			},
			{Code: "import fs from 'fs' with { type: 'json' };", Options: []any{"always"},
				Output: []string{"import fs from 'node:fs' with { type: 'json' };"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 16, 1, 20)},
			},
			{Code: "function call(require) { return require('fs'); }", Options: []any{"always"},
				Output: []string{"function call(require) { return require('node:fs'); }"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("always", "fs", 1, 41, 1, 45)},
			},
			{Code: "import defer * as fs from 'node:fs';", Options: []any{"never"},
				Output: []string{"import defer * as fs from 'fs';"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 27, 1, 36)},
			},
			{Code: "r\\u0065quire('node:fs');", Options: []any{"never"},
				Output: []string{"r\\u0065quire('fs');"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 14, 1, 23)},
			},
			{Code: "require /* callee */ (( /* argument */ 'node:fs'));", Options: []any{"never"},
				Output: []string{"require /* callee */ (( /* argument */ 'fs'));"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 40, 1, 49)},
			},
			{Code: "const face = '😀';\r\nrequire(\r\n  'node:fs'\r\n);", Options: []any{"never"},
				Output: []string{"const face = '😀';\r\nrequire(\r\n  'fs'\r\n);"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 3, 3, 3, 12)},
			},
			{Code: "class C { #fs = require('node:fs'); }", Options: []any{"never"},
				Output: []string{"class C { #fs = require('fs'); }"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 25, 1, 34)},
			},
			{Code: "import fs from 'node:fs' with { type: 'json' };", Options: []any{"never"},
				Output: []string{"import fs from 'fs' with { type: 'json' };"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 16, 1, 25)},
			},
			{Code: "function call(require) { return require('node:fs'); }", Options: []any{"never"},
				Output: []string{"function call(require) { return require('fs'); }"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 41, 1, 50)},
			},
			{Code: "import \"node:node:fs\";", Options: []any{"never"},
				Output: []string{"import \"node:fs\";", "import \"fs\";"},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "node:fs", 1, 8, 1, 22)},
			},
			// Escapes in the prefix must preserve the intended module after fixing.
			{Code: `import 'node\u{3a}fs';`, Options: []any{"never"}, Output: []string{`import 'fs';`},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 1, 22)},
			},
			{Code: `import 'no\
de:fs';`, Options: []any{"never"}, Output: []string{`import 'fs';`},
				Errors: []rule_tester.InvalidTestCaseError{protocolError("never", "fs", 1, 8, 2, 7)},
			},
		})
}
