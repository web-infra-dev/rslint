package no_empty_file_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_empty_file"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// The shared RuleTester binds the rule under "test". Use its real name here
// to check rule-specific suppression and rslint's native directive aliases.
func TestNoEmptyFileDirectiveSuppression(t *testing.T) {
	for _, prefix := range []string{"eslint", "rslint"} {
		for _, tc := range []struct {
			code string
			want int
		}{
			{"\n  /* PREFIX-disable unicorn/no-empty-file */", 0},
			{"\n  // PREFIX-disable-line unicorn/no-empty-file", 0},
			{"// PREFIX-disable-next-line unicorn/no-empty-file", 1},
			{"/* PREFIX-enable */", 1},
			{"// PREFIX-enable", 0},
			{"// PREFIX-disable no-console", 0},
			{"/* PREFIX-disable no-console */", 1},
			{"/* PREFIX-disable no-console */\n// explanation", 0},
		} {
			code := strings.ReplaceAll(tc.code, "PREFIX", prefix)
			t.Run(code, func(t *testing.T) {
				sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/example.js", Path: "/example.js"}, code, core.ScriptKindJS)
				comments := rule.NewCommentStore(sf)
				var diagnostics []rule.RuleDiagnostic
				ctx := rule.RuleContext{SourceFile: sf, Comments: comments, DisableManager: rule.NewDisableManager(sf, comments)}.
					WithReporter(no_empty_file.NoEmptyFileRule.Name, rule.SeverityError, func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) })
				no_empty_file.NoEmptyFileRule.Run(ctx, []any{map[string]any{"allowComments": true}})
				if len(diagnostics) != tc.want {
					t.Fatalf("got %d diagnostics, want %d", len(diagnostics), tc.want)
				}
				if tc.want == 1 {
					d := diagnostics[0]
					if d.Message.Id != "no-empty-file" || d.Message.Description != "Empty files are not allowed." ||
						d.Range.Pos() != 0 || d.Range.End() != len(code) || d.FixesPtr != nil || d.Suggestions != nil {
						t.Fatalf("unexpected diagnostic: %+v", d)
					}
				}
			})
		}
	}
}

// Covers directive boundaries, TypeScript/JSX content, triple-slash spelling,
// options, UTF-16 ranges, and disable/enable comment handling.
// Expected diagnostics were checked against eslint-plugin-unicorn v77.0.0.
func TestNoEmptyFileExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_file.NoEmptyFileRule,
		[]rule_tester.ValidTestCase{
			{Code: "'use client';\nexport {};", FileName: "example.js"},
			{Code: "{}; 'not a directive';", FileName: "example.js"},
			{Code: "'use strict'; ('still content');", FileName: "example.js"},
			{Code: "'use strict'; `content`;", FileName: "example.js"},
			{Code: "(\"use strict\" as const)", FileName: "example.ts"},
			{Code: "\"use strict\" as string;", FileName: "example.ts"},
			{Code: "\"use strict\"!;", FileName: "example.ts"},
			{Code: "\"use strict\" satisfies string;", FileName: "example.ts"},
			{Code: "interface Foo {}", FileName: "example.ts"},
			{Code: "declare module 'virtual' {}", FileName: "example.d.ts"},
			{Code: "export type Foo = {};", FileName: "example.ts"},
			{Code: "import type { Foo } from 'foo';", FileName: "example.ts"},
			{Code: "<div />", FileName: "example.tsx"},
			{Code: "label: {}", FileName: "example.js"},
			{Code: "if (false) {}", FileName: "example.js"},
			{Code: "function empty() {}", FileName: "example.js"},
			{Code: ";\n/// not a reference", FileName: "example.js"},
			{Code: "////", FileName: "example.js"},
			{Code: "/* block */\n// ordinary", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* global foo */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* eslint no-console: off */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* eslint-enable-not-a-directive */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* eslint-enable: */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/*\u0085eslint-enable */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* eslint-enable\u0085 */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* eslint-disable-line no-console\n */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* eslint-disable-next-line no-console */\n// explanation", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "\n  /* eslint-disable */", FileName: "example.js"},
			{Code: "\n  // eslint-disable-line", FileName: "example.js"},
		}, []rule_tester.InvalidTestCase{
			// typescript-eslint's Program starts at the first token (or EOF).
			{Code: "\n  /* 注释😀 */\n", FileName: "example.ts", Errors: emptyFileError(3, 1, 3, 1)},
			{Code: "\n  \"use strict\"; // trailing\n", FileName: "example.ts", Errors: emptyFileError(2, 3, 3, 1)},
			{Code: " \r\n ", FileName: "example.ts", Errors: emptyFileError(2, 2, 2, 2)},
			{Code: "#!/usr/bin/env node\n", FileName: "example.ts", Errors: emptyFileError(2, 1, 2, 1)},
			{Code: "\"use client\"; \"use strict\";", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 28)},
			{Code: "\"use strict\"; ; { ; {} };", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 26)},
			{Code: "  ; // trailing\n", FileName: "example.js", Errors: emptyFileError(1, 1, 2, 1)},
			{Code: "\n// leading\n{}\n", FileName: "example.js", Errors: emptyFileError(1, 1, 4, 1)},
			{Code: "/* 注释😀 */\r\n", FileName: "example.js", Errors: emptyFileError(1, 1, 2, 1)},
			{Code: " \u2028\u2029", FileName: "example.js", Errors: emptyFileError(1, 1, 3, 1)},
			{Code: "// / not a triple slash", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 24)},
			{Code: "/* /// not a line comment */", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 29)},
			{Code: "// comment", FileName: "example.js", Options: map[string]any{}, Errors: emptyFileError(1, 1, 1, 11)},
			{Code: "// comment", FileName: "example.js", Options: map[string]any{"allowComments": false}, Errors: emptyFileError(1, 1, 1, 11)},
			{Code: "", FileName: "example.js", Options: map[string]any{"allowComments": false}, Errors: emptyFileError(1, 1, 1, 1)},
			{Code: "\"use client\"; // comment", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 25)},
			{Code: ";\n/* comment */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 2, 14)},
			{Code: "\n/* eslint-enable */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(2, 1, 2, 20)},
			{Code: "/*\uFEFFeslint-enable */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 20)},
			{Code: "/* eslint-enable\uFEFFno-console */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 31)},
			{Code: "\n/* eslint-enable -- reason */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(2, 1, 2, 30)},
			{Code: "/* eslint-disable-next-line\n no-console */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 2, 15)},
			{Code: "\n/* eslint-disable no-console */", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(2, 1, 2, 32)},
			{Code: "// eslint-disable-next-line", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 28)},
			{Code: "/* 😀 */\n  /* eslint-enable */\n{}", FileName: "example.js", Errors: emptyFileError(2, 3, 2, 22)},
			{Code: "/* eslint-enable */\n/* eslint-disable */", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 20)},
			{Code: "\n// eslint-disable-next-line no-console", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(2, 1, 2, 39)},
		})
}
