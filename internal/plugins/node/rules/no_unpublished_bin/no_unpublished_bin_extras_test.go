package no_unpublished_bin

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func ignoredBin(name string, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "invalidIgnored",
		Message:   "npm ignores '" + name + "'. Check 'files' field of 'package.json' or '.npmignore'.",
		Line:      1, Column: 1, EndLine: endLine, EndColumn: endColumn,
	}}
}

func TestNoUnpublishedBinProgramRanges(t *testing.T) {
	// Program locations from ESLint 10.2.1 include all source trivia. No AST
	// expression inspection is involved, including in TypeScript and JSX.
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		code               string
		endLine, endColumn int
	}{
		{"", 1, 1},
		{"  ", 1, 3},
		{"// comment\n", 2, 1},
		{"\n/* leading */\nconsole.log(\"😀\");\n// trailing\n", 5, 1},
		{"#!/usr/bin/env node\nhello();\n", 3, 1},
		{"/*😀*/", 1, 7},
		{"\uFEFFhello();\r\n", 2, 1},
		{";;\n", 2, 1},
		{"'😀';\u2028'中';\u2029", 3, 1},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: test.code, FileName: "simple-files/a.js", Errors: ignoredBin("a.js", test.endLine, test.endColumn)})
	}
	rule_tester.RunRuleTester(unpublishedBinRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoUnpublishedBinRule, nil, invalid)
}

func TestNoUnpublishedBinDirectives(t *testing.T) {
	// Reporting Program starts at line 1, even when code follows a comment.
	// RuleTester registers the rule as "test". The same directives with the
	// public rule name were also compared through the native API and ESLint.
	rule_tester.RunRuleTester(unpublishedBinRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &NoUnpublishedBinRule,
		[]rule_tester.ValidTestCase{
			{Code: "/* eslint-disable test */\nhello();", FileName: "simple-files/a.js"},
			{Code: "hello(); // eslint-disable-line test", FileName: "simple-files/a.js"},
			{Code: "/* eslint-disable test */\n/* eslint-enable test */\nhello();", FileName: "simple-files/a.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "// eslint-disable-next-line test\nhello();", FileName: "simple-files/a.js", Errors: ignoredBin("a.js", 2, 9)},
			{Code: "\n/* eslint-disable test */\nhello();", FileName: "simple-files/a.js", Errors: ignoredBin("a.js", 3, 9)},
		})
}

func TestNoUnpublishedBinPublication(t *testing.T) {
	root := unpublishedBinRoot(t, "testdata/extras.txtar")
	valid := []rule_tester.ValidTestCase{}
	invalid := []rule_tester.InvalidTestCase{}
	for _, filename := range []string{
		"npm-precedence/cli.js", "files-precedence/cli.js", "reinclude/cli.js",
		"main/cli.js", "missing-bin/cli.js", "nested/inner/cli.js", "negated/bin/other.js",
		"unrelated.js",
		// Shared publication policy corrects upstream's cwd-dependent README
		// exemption, dot-prefixed names, and glob edge cases (see rule docs).
		"metadata/README.js", "metadata/LICENSE.js", "classes/bin/foo.js",
		"unclosed/[cli.js", "zero-step/lib/cli1.js",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: "hello();", FileName: filename})
	}
	// TypeScript's wildcard includes omit dot-prefixed files; select it explicitly.
	valid = append(valid, rule_tester.ValidTestCase{Code: "hello();", FileName: "metadata/..hidden.js", TSConfig: "dotfiles.tsconfig.json"})
	for _, test := range []struct{ filename, relative string }{
		{"git/cli.js", "cli.js"}, {"both/cli.js", "cli.js"},
		{"negated/bin/cli.js", "bin/cli.js"}, {"directory/bin/index.js", "bin/index.js"},
		{"extension/cli.js", "cli.js"}, {"invalid-bin/cli.js", "cli.js"},
		{"private/cli.js", "cli.js"}, {"fallback/inner/cli.js", "inner/cli.js"},
		{"whitespace/lib/foo.js", "lib/foo.js"},
		{"unicode/脚本😀.js", "脚本😀.js"},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: "hello();", FileName: test.filename, Errors: ignoredBin(test.relative, 1, 9)})
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoUnpublishedBinRule, valid, invalid)
}

func TestNoUnpublishedBinConversion(t *testing.T) {
	root := unpublishedBinRoot(t, "testdata/extras.txtar")
	convert := map[string]any{"convertPath": []any{map[string]any{
		"include": []any{"src/**"}, "exclude": []any{"**/*.spec.ts"},
		"replace": []any{`^src/(.*)\.ts$`, "dist/$1.js"},
	}}}
	object := map[string]any{"convertPath": map[string]any{"src/**": []any{`^src/(.*)\.ts$`, "dist/$1.js"}}}
	nested := map[string]any{"convertPath": map[string]any{"src/**": []any{`^src/(.*)\.ts$`, "nested/$1.js"}}}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoUnpublishedBinRule,
		[]rule_tester.ValidTestCase{
			{Code: "", FileName: "conversion/src/cli.ts"},
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{}},
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{}}},
			{Code: "", FileName: "conversion/src/cli.spec.ts", Options: convert},
			{Code: "", FileName: "conversion/other/cli.ts", Options: convert},
			// Explicit options override n and node shared settings; n overrides node.
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{}}, Settings: map[string]any{"n": object}},
			{Code: "", FileName: "conversion/src/cli.ts", Settings: map[string]any{"n": map[string]any{"convertPath": map[string]any{}}, "node": object}},
			// Nested package metadata cannot exclude a file published by its owner.
			{Code: "", FileName: "conversion-published/src/cli.ts", Options: nested},
			// Invalid JS patterns are ignored instead of crashing the linter.
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{"[", "dist/cli.js"}}}},
			// Shared Node path resolution treats a leading slash as absolute.
			// Upstream joins it to the package then rejects it as an ignore path.
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{".*", "/dist/cli.js"}}}},
			// Object entries use lexical priority. Use an array for explicit order.
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{
				"src/**": []any{`^src/(.*)\.ts$`, "dist/$1.js"}, "**": []any{"^", ""},
			}}},
		}, []rule_tester.InvalidTestCase{
			// Nested package metadata cannot include a file excluded by its owner.
			{Code: "", FileName: "conversion/src/cli.ts", Options: nested, Errors: ignoredBin("nested/cli.js", 1, 1)},
			{Code: "const cli: string = 'cli';\n", FileName: "conversion/src/cli.ts", Options: convert, Errors: ignoredBin("dist/cli.js", 2, 1)},
			{Code: "", FileName: "conversion/src/cli.ts", Settings: map[string]any{"n": object}, Errors: ignoredBin("dist/cli.js", 1, 1)},
			{Code: "", FileName: "conversion/src/cli.ts", Settings: map[string]any{"n": map[string]any{"convertPath": nil}, "node": object}, Errors: ignoredBin("dist/cli.js", 1, 1)},
			{Code: "", FileName: "conversion/src/cli.ts", Options: object, Settings: map[string]any{"node": nested}, Errors: ignoredBin("dist/cli.js", 1, 1)},
			{Code: "const view = <main />;\n", FileName: "conversion/src/cli.tsx", Tsx: true, Options: map[string]any{"convertPath": map[string]any{"src/**": []any{`^src/(.*)\.tsx$`, "dist/$1.js"}}}, Errors: ignoredBin("dist/cli.js", 2, 1)},
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{".*", "../outside.js"}}}, Errors: ignoredBin("../outside.js", 1, 1)},
			// JavaScript lookahead and named replacements use the existing JS regexp helper.
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{`^(?=src/)(?<folder>src)/(.+)\.ts$`, "dist/$2.js"}}}, Errors: ignoredBin("dist/cli.js", 1, 1)},
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{`^src/(?<name>.+)\.ts$`, "dist/$<name>.js"}}}, Errors: ignoredBin("dist/cli.js", 1, 1)},
			{Code: "", FileName: "conversion/src/cli.ts", Options: map[string]any{"convertPath": []any{
				map[string]any{"include": []any{"src/**"}, "replace": []any{`^src/(.*)\.ts$`, "dist/$1.js"}},
				map[string]any{"include": []any{"**"}, "replace": []any{".*", "unrelated.js"}},
			}}, Errors: ignoredBin("dist/cli.js", 1, 1)},
		})
}

func TestNoUnpublishedBinSchema(t *testing.T) {
	// The upstream docs show convertPath:null as a placeholder, but the
	// upstream schema rejects it. Omit convertPath to use the default.
	for _, options := range []any{
		map[string]any{"convertPath": nil}, map[string]any{"unknown": true},
		map[string]any{"convertPath": []any{}},
		map[string]any{"convertPath": map[string]any{"src/**": []any{"pattern"}}},
		map[string]any{"convertPath": []any{map[string]any{"include": []any{}, "replace": []any{"x", "y"}}}},
		map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}}}},
	} {
		if err := NoUnpublishedBinRule.Schema.Validate(rule.NormalizeOptions(options)); err == nil {
			t.Errorf("accepted invalid options %#v", options)
		}
	}
	NoUnpublishedBinRule.Run(rule.RuleContext{}, nil)
}
