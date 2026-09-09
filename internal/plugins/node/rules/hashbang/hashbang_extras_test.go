package hashbang

import (
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestHashbangSourceBoundaries(t *testing.T) {
	missing := func(endColumn int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: `This file needs shebang "#!/usr/bin/env node".`, Line: 1, Column: 1, EndLine: 1, EndColumn: endColumn}}
	}
	rule_tester.RunRuleTester(hashbangRoot(t), "tsconfig.json", t, &HashbangRule,
		[]rule_tester.ValidTestCase{
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env\u00a0node\nhello();"},
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env FOO=\"😀\" node\nhello();"},
			// Only source text and the configured extension matter, including TS/JSX syntax.
			{FileName: "no-bin-field/tool.ts", Code: "#!/usr/bin/env ts-node\nconst value: number = 1;", Options: map[string]any{"additionalExecutables": []any{"tool.ts"}, "executableMap": map[string]any{".ts": "ts-node"}}},
			{FileName: "no-bin-field/tool.tsx", Code: "#!/usr/bin/env tsx\nconst el = <div />;", Options: map[string]any{"additionalExecutables": []any{"tool.tsx"}, "executableMap": map[string]any{".tsx": "tsx"}}},
			// Scanner recognition alone must not remove these upstream-unrecognized prefixes.
			{FileName: "string-bin/lib/test.js", Code: "#!\nhello();"},
			{FileName: "string-bin/lib/test.js", Code: "#!/usr/bin/env node\u2028hello();"},
		}, []rule_tester.InvalidTestCase{
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env node", Output: []string{"#!/usr/bin/env node\n"}, Errors: missing(20)},
			{FileName: "string-bin/bin/test.js", Code: "#!\nhello();", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: missing(3)},
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env node\rhello();", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: missing(20)},
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env node\u2029hello();", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: missing(20)},
			{FileName: "string-bin/bin/test.js", Code: "\r\nhello();", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: missing(1)},
			// JS trim does not remove NEL, even though Go's TrimSpace does.
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env\u0085node\nhello();", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: missing(20)},
		})
}

func TestHashbangExtras(t *testing.T) {
	nodeError := rule_tester.InvalidTestCaseError{MessageId: "expectedHashbangNode", Message: `This file needs shebang "#!/usr/bin/env node".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1}
	unicodeError := nodeError
	unicodeError.EndColumn = 21
	noHashbang := rule_tester.InvalidTestCaseError{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}
	rule_tester.RunRuleTester(hashbangRoot(t), "tsconfig.json", t, &HashbangRule,
		[]rule_tester.ValidTestCase{
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env -u FOO -P=/bin FOO=\"hello world\" node\nhello();"},
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env -ivS node\nhello();", Options: map[string]any{}},
			{FileName: "string-bin/bin/test.js", Code: "#!/usr/bin/env node\nhello();", Options: map[string]any{"ignoreUnpublished": false, "additionalExecutables": []any{}, "executableMap": map[string]any{".js": "node"}, "convertPath": map[string]any{}}},
			{FileName: "string-bin/lib/test.js", Code: "#!/usr/bin/env node"}, // Upstream only recognizes hashbangs ending in LF.
			{FileName: "unpublished/something.test.js", Code: "#!/usr/bin/env node\nhello();", Options: map[string]any{"additionalExecutables": []any{"*.TEST.js"}, "ignoreUnpublished": true}},
			// JavaScript /i does not fold the Kelvin sign to the Latin letter K.
			{FileName: "no-bin-field/K.js", Code: "hello();", Options: map[string]any{"additionalExecutables": []any{"K.js"}}},
			// Two wildcards select the two UTF-16 code units in an emoji filename.
			{FileName: "no-bin-field/😀.js", Code: "#!/usr/bin/env node\nhello();", Options: map[string]any{"additionalExecutables": []any{"??.js"}}},
			{FileName: "string-bin/src/bin/test.js", Code: "#!/usr/bin/env node\nhello();", Settings: map[string]any{"n": map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/", ""}}}, "node": map[string]any{"convertPath": map[string]any{}}}},
			{FileName: "string-bin/src/bin/test.js", Code: "hello();", Options: map[string]any{"convertPath": map[string]any{}}, Settings: map[string]any{"n": map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/", ""}}}}},
			{FileName: "string-bin/src/bin/test.js", Code: "#!/usr/bin/env node\nhello();", Options: map[string]any{"convertPath": []any{map[string]any{"include": []any{"src/**"}, "exclude": []any{"src/lib/**"}, "replace": []any{"^src/", ""}}}}},
			// An explicit pattern array does not split embedded newlines.
			{FileName: "no-bin-field/b.js", Code: "hello();", Options: map[string]any{"additionalExecutables": []any{"*.js\n!a.js"}}},
			// A convertPath backslash is literal, so this pattern must not
			// accidentally turn the ordinary source into the bin entry.
			{FileName: "string-bin/src/a.js", Code: "hello();", Options: map[string]any{"convertPath": map[string]any{`src/\*.js`: []any{".*", "bin/test.js"}}}},
			// The adapter skips an invalid conversion regexp rather than crash or apply a fix to the wrong path.
			{FileName: "string-bin/bin/test.js", Code: "hello();", Options: map[string]any{"convertPath": map[string]any{"**": []any{"[", ""}}}},
		}, []rule_tester.InvalidTestCase{
			// The second lexical capture selects bin/test.js even though the
			// first capture is named. This conversion must require a hashbang.
			{FileName: "string-bin/src/bin/test.js", Code: "", Options: map[string]any{"convertPath": map[string]any{"src/**": []any{`(?<dir>src)/(.*)`, "$2"}}}, Output: []string{"#!/usr/bin/env node\n"}, Errors: []rule_tester.InvalidTestCaseError{nodeError}},
			{FileName: "string-bin/bin/test.js", Code: "", Output: []string{"#!/usr/bin/env node\n"}, Errors: []rule_tester.InvalidTestCaseError{nodeError}},
			{FileName: "string-bin/bin/test.js", Code: "\n\nhello();", Output: []string{"#!/usr/bin/env node\n\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{nodeError}},
			{FileName: "string-bin/bin/test.js", Code: "console.log('😀中文');", Output: []string{"#!/usr/bin/env node\nconsole.log('😀中文');"}, Errors: []rule_tester.InvalidTestCaseError{unicodeError}},
			{FileName: "string-bin/lib/test.js", Code: "\uFEFF#!/usr/bin/env node\nhello();", Output: []string{"\uFEFFhello();"}, Errors: []rule_tester.InvalidTestCaseError{noHashbang}},
			{FileName: "unpublished/something.test.js", Code: "#!/usr/bin/env node\nhello();", Options: map[string]any{"additionalExecutables": []any{"*.js", "!*.test.js"}}, Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{noHashbang}},
		})
}

func TestHashbangPaths(t *testing.T) {
	for name, want := range map[string]string{".hidden": "", "cli": "", "cli.ts": ".ts", ".cli.js": ".js", "cli.": "."} {
		if got := fileExtension(name); got != want {
			t.Errorf("extension of %q = %q, want %q", name, got, want)
		}
	}
	for _, text := range []string{"#!/usr/bin/env node\t--flag\n", "#!/usr/bin/env bun\n", "#!/usr/bin/node\n"} {
		if isNodeShebang(text, "node") {
			t.Errorf("wrong interpreter matched %q", text)
		}
	}
	for _, options := range []any{map[string]any{"convertPath": nil}, map[string]any{"ignoreUnpublished": "yes"}, map[string]any{"additionalExecutables": []any{42}}, map[string]any{"executableMap": map[string]any{"ts": "node"}}, map[string]any{"executableMap": map[string]any{".ts": "ts node"}}} {
		if err := HashbangRule.Schema.Validate(rule.NormalizeOptions(options)); err == nil {
			t.Errorf("accepted invalid options %#v", options)
		}
	}
}

func TestHashbangUnpublished(t *testing.T) {
	root := hashbangRoot(t)
	archive := txtarfs.MustParseFile(t, "testdata/publishing.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	for _, test := range []struct {
		file string
		want bool
	}{
		{"git/secret.js", true}, {"git/open.js", false},
		{"class-exclusion/lib/foo.js", false},
		{"newline-files/lib/foo.js", true},
		{"npm/secret.js", false}, {"npm/private.js", true},
		{"files/lib/keep.js", false}, {"files/lib/drop.js", true}, {"files/outside.js", true},
		{"files/src/cli.js", false}, {"files/package.json", false},
		{"extended/lib/a.js", false}, {"extended/lib/b.js", false}, {"extended/lib/c.js", true},
		{"extended/bin/cli1.js", false}, {"extended/bin/cli3.js", false}, {"extended/bin/cli4.js", true},
		{"extended/lib/private/a.js", true},
		{"empty/a.js", false},
		// Two wildcards match the two UTF-16 units in a supplementary character.
		{"unicode/lib/😀.js", false},
		// The shared brace expander treats a zero step as one. Upstream's
		// publication matcher leaves the sequence unexpanded instead.
		{"zero-step/lib/cli1.js", false},
		// Removing ! must preserve the leading slash for both matcher paths.
		{"rooted-exclusion/lib/foo.js", false},
		{"rooted-exclusion/lib/nested/foo.js", false},
		{"unrooted-exclusion/lib/foo.js", true},
		{"unrooted-exclusion/lib/nested/foo.js", true},
		{"rooted-extended-exclusion/lib/foo.js", false},
		{"rooted-extended-exclusion/lib/nested/foo.js", false},
		{"unrooted-extended-exclusion/lib/foo.js", true},
		{"unrooted-extended-exclusion/lib/nested/foo.js", true},
	} {
		t.Run(test.file, func(t *testing.T) {
			program, _, err := rule_tester.NewProgramHelper(root).CreateTestProgram("hello();", "empty/probe.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			p := lintprogram.NewFromCompiler(program)
			absolute := tspath.ResolvePath(root.Dir, test.file)
			pkg := nodeutil.FindPackage(p, absolute)
			if pkg == nil {
				t.Fatal("package not found")
			}
			relative := tspath.GetRelativePathFromDirectory(pkg.Directory(), absolute, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
			if got := nodeutil.IsUnpublished(p, absolute, relative); got != test.want {
				t.Errorf("unpublished = %v, want %v", got, test.want)
			}
		})
	}
	// Publication filtering must preserve executable diagnostics and fixes,
	// including when an unrelated root exclusion or Unicode pattern is present.
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &HashbangRule,
		[]rule_tester.ValidTestCase{
			{FileName: "newline-files/lib/foo.js", Code: "hello();", Options: map[string]any{"ignoreUnpublished": true}},
			{FileName: "unrooted-exclusion/lib/foo.js", Code: "hello();", Options: map[string]any{"ignoreUnpublished": true}},
			{FileName: "unrooted-extended-exclusion/lib/foo.js", Code: "hello();", Options: map[string]any{"ignoreUnpublished": true}},
		},
		[]rule_tester.InvalidTestCase{
			{
				FileName: "class-exclusion/lib/foo.js",
				Code:     "hello();",
				Options:  map[string]any{"ignoreUnpublished": true},
				Output:   []string{"#!/usr/bin/env node\nhello();"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}},
			},
			{
				FileName: "rooted-exclusion/lib/foo.js",
				Code:     "hello();",
				Options:  map[string]any{"ignoreUnpublished": true},
				Output:   []string{"#!/usr/bin/env node\nhello();"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedHashbangNode",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				FileName: "rooted-extended-exclusion/lib/foo.js",
				Code:     "hello();",
				Options:  map[string]any{"ignoreUnpublished": true},
				Output:   []string{"#!/usr/bin/env node\nhello();"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedHashbangNode",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 9,
				}},
			},
			{
				FileName: "unicode/lib/😀.js",
				Code:     "#!/usr/bin/env node\nhello();",
				Options:  map[string]any{"ignoreUnpublished": true},
				Output:   []string{"hello();"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedHashbang", Message: "This file needs no shebang.",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 20,
				}},
			},
		})
}

func TestHashbangEditDemand(t *testing.T) {
	for _, test := range []struct {
		file, code string
		count, end int
	}{
		{"string-bin/bin/test.js", "hello();", 1, 8},
		{"string-bin/lib/test.js", "#!/usr/bin/env node\nhello();", 1, 19},
		{"string-bin/bin/test.js", "\uFEFF#!/usr/bin/env node\r\nhello();", 2, 19},
	} {
		root := hashbangRoot(t)
		program, file, err := rule_tester.NewProgramHelper(root).CreateTestProgram(test.code, test.file, "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		p := lintprogram.NewFromCompiler(program)
		run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
			var diags []rule.RuleDiagnostic
			ctx := rule.RuleContext{SourceFile: file, BOM: rule.NewSourceBOM(p.FS(), file.FileName())}.WithProgram(p).WithDiagnosticConsumer(HashbangRule.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) }})
			HashbangRule.Run(ctx, nil)
			return diags
		}
		all := run(rule.EditDemandAll)
		if len(all) != test.count {
			t.Fatalf("got %d diagnostics", len(all))
		}
		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
			got := run(demand)
			if len(got) != len(all) {
				t.Fatal("diagnostic count depends on edit demand")
			}
			for i, d := range got {
				if d.Range != all[i].Range || !reflect.DeepEqual(d.Message, all[i].Message) || d.Range != core.NewTextRange(0, test.end) {
					t.Errorf("diagnostic identity changed: %#v", d)
				}
				if d.Suggestions != nil {
					t.Fatal("unexpected suggestions")
				}
				if demand&rule.EditDemandAutofix != 0 {
					if d.FixesPtr == nil || !reflect.DeepEqual(d.FixesPtr, all[i].FixesPtr) {
						t.Fatal("fix missing or changed")
					}
				} else if d.FixesPtr != nil {
					t.Fatal("unexpected fix")
				}
			}
		}
	}
	// A context without source services does not attempt host filesystem IO.
	HashbangRule.Run(rule.RuleContext{}, nil)
}
