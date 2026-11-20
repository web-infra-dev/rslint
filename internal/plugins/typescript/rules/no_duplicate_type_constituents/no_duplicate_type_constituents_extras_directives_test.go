package no_duplicate_type_constituents

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_explicit_any"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/no_undef"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoDuplicateTypeConstituentsTypeScriptDirectives(t *testing.T) {
	for _, directive := range []string{"@ts-expect-error", "@ts-ignore"} {
		for _, blank := range []string{"", "\n\n// intervening comment"} {
			for _, newline := range []string{"\n", "\r\n"} {
				for _, code := range []string{
					"export type T =\n  // {directive}\n  {x: Missing} | (|\n    // {directive}{blank}\n    {x: Missing});",
					"// {directive}{blank}\nexport type T = {x: Missing} | {x: Missing};",
				} {
					code = strings.ReplaceAll(code, "{directive}", directive)
					code = strings.ReplaceAll(code, "{blank}", blank)
					code = strings.ReplaceAll(code, "\n", newline)
					diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
					if len(diagnostics) != 1 || diagnostics[0].Message.Id != "duplicate" {
						t.Fatalf("expected one duplicate diagnostic: %v", diagnostics)
					}
					if len(diagnostics[0].Fixes()) != 0 {
						t.Fatal("fix would detach the TypeScript directive from its target")
					}
					if _, _, changed := linter.ApplyRuleFixes(code, diagnostics); changed {
						t.Fatal("unsafe TypeScript directive fix applied")
					}
				}
			}
		}
	}
	// A suppression above an unrelated declaration does not disable later fixes.
	code := "// @ts-expect-error\ntype Error = Missing;\ntype T = string | string;"
	diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
	if len(diagnostics) != 1 || len(diagnostics[0].Fixes()) == 0 {
		t.Fatalf("unrelated TypeScript directive prevented a safe fix: %v", diagnostics)
	}
	fixed, _, _ := linter.ApplyRuleFixes(code, diagnostics)
	lintSafetyFixture(t, fixed, rule.EditDemandNone)
}

func TestNoDuplicateTypeConstituentsNextLineTargets(t *testing.T) {
	for _, prefix := range []string{"eslint", "rslint"} {
		for _, newline := range []string{"\n", "\r\n", "\u2028", "\u2029"} {
			code := "export type T = {x: any} | (| /* " + prefix + "-disable-next-line @typescript-eslint/no-explicit-any */ {\n  x: any\n});\nexport type Outside = any;"
			code = strings.ReplaceAll(code, "\n", newline)
			diagnostics := lintDirectiveFixtureRules(t, code, []*rule.Rule{&NoDuplicateTypeConstituentsRule, &no_explicit_any.NoExplicitAnyRule})
			duplicate, outside := 0, 0
			for _, diagnostic := range diagnostics {
				if diagnostic.RuleName == NoDuplicateTypeConstituentsRule.Name {
					duplicate++
					if len(diagnostic.Fixes()) != 0 {
						t.Fatal("multiline removal would move the next-line target")
					}
				} else if diagnostic.Message.Id == "unexpectedAny" {
					outside++
				}
			}
			if duplicate != 1 || outside != 2 {
				t.Fatalf("expected the first and outside any to remain reported: %v", diagnostics)
			}
			if _, _, changed := linter.ApplyRuleFixes(code, diagnostics); changed {
				t.Fatal("unsafe next-line fix applied")
			}
		}
	}
	// A retained directive after the member can widen to cover the earlier
	// constituent when the multiline body is collapsed.
	trailing := "type T = {x: any} | {\n x: any\n}; // eslint-disable-line @typescript-eslint/no-explicit-any"
	found := false
	for _, diagnostic := range lintDirectiveFixtureRules(t, trailing, []*rule.Rule{&NoDuplicateTypeConstituentsRule, &no_explicit_any.NoExplicitAnyRule}) {
		if diagnostic.RuleName == NoDuplicateTypeConstituentsRule.Name {
			found = true
			if len(diagnostic.Fixes()) != 0 {
				t.Fatal("fix widened a trailing line directive")
			}
		}
	}
	if !found {
		t.Fatal("trailing-directive fixture did not exercise the duplicate rule")
	}

	// Inline line directives in the deleted body must also be retained.
	code := "type T = {x: any} | {x: any /* eslint-disable-line @typescript-eslint/no-explicit-any */};\ntype Outside = any;"
	diagnostics := lintDirectiveFixtureRules(t, code, []*rule.Rule{&NoDuplicateTypeConstituentsRule, &no_explicit_any.NoExplicitAnyRule})
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleName == NoDuplicateTypeConstituentsRule.Name && len(diagnostic.Fixes()) != 0 {
			t.Fatal("deleted body would remove a line directive")
		}
	}
}

func TestNoDuplicateTypeConstituentsGlobalWhitespace(t *testing.T) {
	for _, keyword := range []string{"global", "globals"} {
		for _, whitespace := range []string{" ", "\t", "\n", "\r\n", "\u00a0", "\u2028", "\u2029", "\ufeff"} {
			code := "export type T = {x: string} | { /* " + keyword + whitespace + "reviewProbeGlobal */ x: string };\nreviewProbeGlobal();"
			diagnostics := lintDirectiveFixtureRules(t, code, []*rule.Rule{&NoDuplicateTypeConstituentsRule, &no_undef.NoUndefRule})
			if len(diagnostics) != 1 || diagnostics[0].Message.Id != "duplicate" {
				t.Fatalf("global directive did not suppress no-undef: %v", diagnostics)
			}
			if len(diagnostics[0].Fixes()) != 0 {
				t.Fatal("global directive would be deleted by the fix")
			}
		}
	}
	for _, value := range []string{"globalSomething x", "globalsThing x", "exportedThing x", "ordinary global x"} {
		if fixDirectiveComment(ast.KindMultiLineCommentTrivia, value) {
			t.Fatalf("ordinary comment matched a directive: %q", value)
		}
	}
	for _, whitespace := range []string{"\t", "\n", "\u2028"} {
		if !fixDirectiveComment(ast.KindMultiLineCommentTrivia, "exported"+whitespace+"entry") {
			t.Fatal("exported directive whitespace was not recognized")
		}
	}
}

func TestNoDuplicateTypeConstituentsCommentFastPath(t *testing.T) {
	code := commentBenchmarkSource()
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/file.ts", Path: "/file.ts"}, code, core.ScriptKindTS)
	last := source.Statements.Nodes[len(source.Statements.Nodes)-1].AsTypeAliasDeclaration().Type
	members := last.AsUnionTypeNode().Types.Nodes
	body := utils.TrimNodeTextRange(source, members[1])
	guard := fixDirectiveGuard{source: source, comments: func() []*ast.CommentRange {
		t.Fatal("a comment-free fix scanned comments from the whole file")
		return nil
	}}
	if !guard.safe(body, members[0].End(), []rule.RuleFix{rule.RuleFixRemoveRange(body)}) {
		t.Fatal("comment-free fix should remain available")
	}
	// Markers in a string trigger inspection but are not themselves directives.
	code = "type T = {x: '/* global\\nmissing */'} | {x: '/* global\\nmissing */'};"
	source, members = parseSyntaxMembers(code)
	comments := rule.NewCommentStore(source)
	guard = fixDirectiveGuard{source: source, comments: comments.All}
	body = utils.TrimNodeTextRange(source, members[1])
	if !guard.safe(body, members[1].Pos(), []rule.RuleFix{rule.RuleFixRemoveRange(body)}) {
		t.Fatal("directive-like string prevented a safe fix")
	}
}

func commentBenchmarkSource() string {
	var source strings.Builder
	source.WriteString("// Generated declarations.\n")
	for i := range 4000 {
		fmt.Fprintf(&source, "type Entry%d = {id: number; label: string};\n", i)
	}
	source.WriteString("type Tail = string | string;")
	return source.String()
}

// Reuse the Program and plan: rebuilding them would obscure a per-pass scan.
func BenchmarkNoDuplicateTypeConstituentsCommentPass(b *testing.B) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, source, err := helper.CreateTestProgram(commentBenchmarkSource(), "comment-benchmark.ts", "tsconfig.json")
	if err != nil {
		b.Fatal(err)
	}
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{lintprogram.NewFromCompiler(program)},
		TargetsByProgram: [][]string{{source.FileName()}}, SingleThreaded: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: NoDuplicateTypeConstituentsRule.Name, Environment: &rule.RuleEnvironment{}, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners { return NoDuplicateTypeConstituentsRule.Run(ctx, nil) }}}
		},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		count := 0
		_, err := linter.RunLinter(linter.RunLinterOptions{SingleThreaded: true, LintPlan: plan,
			Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) {
				count++
				if len(d.Fixes()) == 0 {
					b.Fatal("benchmark duplicate has no fix")
				}
			}}})
		if err != nil {
			b.Fatal(err)
		}
		if count != 1 {
			b.Fatalf("expected one diagnostic, got %d", count)
		}
	}
}

// Unicode suppression semantics are pinned to TypeScript 6.0.3. The native
// checker currently treats some of these blank lines differently, so this
// test uses the production lint dispatcher; the native API differential also
// checks the unchanged source with the pinned TypeScript compiler.
func TestNoDuplicateTypeConstituentsUnicodeDirectiveTargets(t *testing.T) {
	var cases []rule_tester.InvalidTestCase
	for _, whitespace := range []string{"\u00a0", "\ufeff", "\u2003", "\t", "\v", "\f"} {
		for _, intervening := range []string{whitespace, whitespace + "// comment", whitespace + "\n" + whitespace + "// comment"} {
			for _, newline := range []string{"\n", "\r\n"} {
				code := "type AsString<T extends string> = string;\n// @ts-expect-error\n" + intervening + "\nexport type T = string | AsString<123>;"
				code = strings.ReplaceAll(code, "\n", newline)
				column := len("export type T = string | ") + 1
				line := strings.Count(code, "\n") + 1
				cases = append(cases, rule_tester.InvalidTestCase{
					Code:   code,
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: line, Column: column, EndLine: line, EndColumn: column + len("AsString<123>")}},
				})
			}
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, nil, cases)
}

func TestNoDuplicateTypeConstituentsInlineDirectiveQuickFix(t *testing.T) {
	for _, directive := range []string{"@ts-expect-error", "@ts-ignore"} {
		for _, newline := range []string{"\n", "\r\n"} {
			code := "export type T = string | /* " + directive + " */ string;" + newline + "const bad: number = 'x';"
			diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
			if len(diagnostics) != 1 || diagnostics[0].Message.Id != "duplicate" {
				t.Fatalf("expected a duplicate with a quick fix: %v", diagnostics)
			}
			fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
			want := "export type T = string  /* " + directive + " */ ;" + newline + "const bad: number = 'x';"
			if !changed || fixed != want {
				t.Fatalf("safe inline fix = %q, want %q", fixed, want)
			}
			if after := lintSafetyFixture(t, fixed, rule.EditDemandAll); len(after) != 0 {
				t.Fatalf("inline fix did not converge: %v", after)
			}
		}
	}
	// Earlier and later directives cannot hide one that actually intersects
	// an edit. Protect the directive inside the body, with an unaffected target.
	code := "// @ts-expect-error\nconst first: number = 'x';\nexport type T = {x: string} | { /* @ts-ignore */ x: string };\nconst second: number = 'x';"
	diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
	if len(diagnostics) != 1 || len(diagnostics[0].Fixes()) != 0 {
		t.Fatalf("deleting an internal directive should not have a fix: %v", diagnostics)
	}
}

func typeScriptBlankBenchmarkSource(blankLines, duplicateTypes int) string {
	var code strings.Builder
	code.WriteString("// @ts-expect-error\n")
	code.WriteString(strings.Repeat("\n", blankLines))
	code.WriteString("const bad: number = 'x';\n")
	for i := range duplicateTypes {
		fmt.Fprintf(&code, "type T%d = string | string;\n", i)
	}
	return code.String()
}

func BenchmarkNoDuplicateTypeConstituentsDirectiveTargets(b *testing.B) {
	for _, blankLines := range []int{0, 2000} {
		for _, duplicateTypes := range []int{1, 1000} {
			b.Run(fmt.Sprintf("blank_%d/types_%d", blankLines, duplicateTypes), func(b *testing.B) {
				helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
				program, source, err := helper.CreateTestProgram(typeScriptBlankBenchmarkSource(blankLines, duplicateTypes), "ts-directive-benchmark.ts", "tsconfig.json")
				if err != nil {
					b.Fatal(err)
				}
				plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
					Programs:         []*lintprogram.Program{lintprogram.NewFromCompiler(program)},
					TargetsByProgram: [][]string{{source.FileName()}}, SingleThreaded: true,
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: NoDuplicateTypeConstituentsRule.Name, Environment: &rule.RuleEnvironment{}, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners { return NoDuplicateTypeConstituentsRule.Run(ctx, nil) }}}
					},
				})
				if err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					count := 0
					_, err := linter.RunLinter(linter.RunLinterOptions{SingleThreaded: true, LintPlan: plan,
						Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) {
							count++
							if len(d.Fixes()) == 0 {
								b.Fatal("unrelated directive prevented a quick fix")
							}
						}}})
					if err != nil {
						b.Fatal(err)
					}
					if count != duplicateTypes {
						b.Fatalf("expected %d duplicates, got %d", duplicateTypes, count)
					}
				}
			})
		}
	}
}
