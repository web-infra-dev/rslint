// TestCapitalizedCommentsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension walk notes for capitalized-comments:
//   - Dimension 1 (AST node types): mostly N/A — the rule inspects raw
//     comment text and the surrounding real-token stream, not any AST node's
//     children, so parenthesization / literal-kind / computed-key shape
//     differences between tsgo and ESTree cannot affect it. The one place a
//     TS-specific token shape matters is the previous/next real-token lookup
//     that isInlineComment relies on; that is covered below.
//   - Dimension 2 (scoping & nesting): N/A — the rule performs no scope
//     lookup or ancestor walk; a comment's validity never depends on which
//     function/class/block contains it.
//   - Dimension 4 (access/key forms, declaration/container forms): N/A — the
//     rule never inspects a property, key, function, or class shape.
//   - Dimension 4 (nesting/traversal boundaries): N/A — comments cannot
//     nest, and the rule has no ancestor walk to bleed across a boundary.
//   - Dimension 4 (graceful degradation: SpreadAssignment/RestElement,
//     overload signatures/abstract/declare members): N/A — none of these
//     shapes can contain or affect a comment's own text. The applicable
//     graceful-degradation case here is "no real token exists on one or
//     both sides of the comment" (isInlineComment / isConsecutiveComment at
//     the start or end of a file), covered below.
package capitalized_comments

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCapitalizedCommentsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CapitalizedCommentsRule,
		[]rule_tester.ValidTestCase{
			// ---- Ignore-pattern timeout: an allow/ignore matcher fails open so
			// a pathological first alternative cannot create a false positive
			// while JavaScript continues to the matching second alternative. ----
			{
				Code:    "// aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!",
				Options: []any{"always", map[string]any{"ignorePattern": "(?:(a|aa)+$|a+!)"}},
			},

			// ---- Dimension 1: inline-comment detection across a TS-specific
			// token shape (generic type parameter list) ----
			{
				Code:    "function foo<T /* inline */>(x: T) {}",
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
			},

			// ---- Dimension 4: graceful degradation — a caseless letter
			// (CJK, general category Lo) is valid under BOTH "always" and
			// "never", since it is neither upper- nor lowercase ----
			// Locks in isCommentValid(): the fallthrough after step 7, when
			// neither the "always"+isLowercase nor "never"+isUppercase arm
			// fires.
			{Code: "// 丈 non-Latin letter", Options: []any{"always"}},
			{Code: "// 丈 non-Latin letter", Options: []any{"never"}},

			// ---- Dimension 4: graceful degradation — a combining mark
			// (general category Mn) as the first non-whitespace character is
			// not \p{L}, so it is valid regardless of mode ----
			// Locks in isCommentValid() step 7 (letter-pattern check).
			{Code: "// \u0301combining mark leads", Options: []any{"always"}},

			// ---- Branch: MAYBE_URL requires a character after "://" ----
			// Locks in isCommentValid() step 5: a comment that names a
			// scheme followed immediately by "://x" is exempt.
			{Code: "// http://x", Options: []any{"always"}},

			// ---- Branch: the default ignore pattern exempts this linter's
			// own directive comments, so `--fix` cannot rewrite one into a
			// spelling DisableManager no longer recognizes ----
			{Code: "// rslint-disable-next-line no-console\nconsole.log(1);", Options: []any{"always"}},
			{Code: "/* rslint-disable-next-line no-console */\nconsole.log(1);", Options: []any{"always"}},
			{Code: "/* rslint-enable no-console */", Options: []any{"always"}},

			// ---- Branch: caseless getNormalizedOptions()'s ignorePattern
			// fallback recognizes the "line"-only override for a Line
			// comment ----
			{Code: "// lowercase", Options: []any{"always", map[string]any{"line": map[string]any{"ignorePattern": "low"}}}},

			// ---- Contract: EndLine/EndColumn on a single-line report ----
			{Code: "// https://github.com/eslint/eslint", Options: []any{"always"}},

			// ---- Dimension 1: the preceding-token walk reaches tokens
			// that only appear inside a JSX or template-literal body, so a
			// comment sandwiched between them is inline ----
			{
				Code:     "var x = <div>a{/* lowercase */}b</div>;",
				FileName: "src/virtual.tsx",
				Options:  []any{"always", map[string]any{"ignoreInlineComments": true}},
			},
			{
				Code:    "`a${/* lowercase */ 1}b`;",
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
			},

			// ---- Dimension 1: a regular-expression literal is one token,
			// so the `//` inside it neither opens a comment nor moves the
			// preceding-token boundary past the earlier comment ----
			{
				Code:    "var re = /a\\/\\/b/; /* This Comment Is Fine. */ /* lowercase */",
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: graceful degradation — a comment alone in
			// the file has no real token on either side, so it is never
			// "inline" even when ignoreInlineComments is true ----
			// Locks in isInlineComment(): the missing-previousToken and
			// missing-nextToken early-return-false arms.
			{
				Code:    "/*lowercase*/",
				Output:  []string{"/*Lowercase*/"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Dimension 1: inline-comment detection stays disabled
			// across a TS generic-parameter boundary when
			// ignoreInlineComments is off ----
			{
				Code:   "function foo<T /* inline */>(x: T) {}",
				Output: []string{"function foo<T /* Inline */>(x: T) {}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 16}},
			},

			// ---- Real-user (eslint/eslint#13229): a trailing line comment
			// can never be "inline" — a Line comment always consumes the
			// rest of its line, so there is no token on the same line as
			// the comment's end, and ignoreInlineComments cannot exempt it.
			// This is surprising enough to users that it was filed as a bug
			// and closed as working-as-intended. ----
			{
				Code:    "const example = {\n  nonce: 'abc', // a numeric string\n}",
				Output:  []string{"const example = {\n  nonce: 'abc', // A numeric string\n}"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 2, Column: 17}},
			},

			// ---- Real-user (eslint/eslint#7901): a block comment that
			// leads its own line is not "inline" even though a token
			// follows it on the same line, because the *previous* token
			// sits on the line above — ignoreInlineComments only exempts a
			// comment sandwiched between two same-line tokens. ----
			{
				Code:    "trackEvent(\n    /* category */ 'foo',\n    /* action */ 'bar'\n);",
				Output:  []string{"trackEvent(\n    /* Category */ 'foo',\n    /* Action */ 'bar'\n);"},
				Options: []any{"always", map[string]any{"ignoreInlineComments": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLowercaseComment", Line: 2, Column: 5},
					{MessageId: "unexpectedLowercaseComment", Line: 3, Column: 5},
				},
			},

			// ---- Branch: getNormalizedOptions()'s `rawOptions[which] ||
			// rawOptions` fallback — a "line"-only override leaves Block
			// options at the bare defaults (the extra "line" key on the raw
			// object is inert), so a Block comment is unaffected by the
			// line-scoped ignorePattern. ----
			{
				Code:    "/* lowercase */",
				Output:  []string{"/* Lowercase */"},
				Options: []any{"always", map[string]any{"line": map[string]any{"ignorePattern": "low"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Branch: an empty-string ignorePattern is falsy upstream
			// (`if (ignorePatternStr)`), so no custom regexp is built and
			// the option behaves as if it were never set. ----
			{
				Code:    "// lowercase",
				Output:  []string{"// Lowercase"},
				Options: []any{"always", map[string]any{"ignorePattern": ""}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Branch: MAYBE_URL excludes "?" and "#" immediately after
			// "://", and requires at least one character there at all. ----
			{
				Code:    "// http://",
				Output:  []string{"// Http://"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},
			{
				Code:    "// http://?query",
				Output:  []string{"// Http://?query"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1}},
			},

			// ---- Branch: isConsecutiveComment() when the nearest
			// preceding item is a real token, not a comment — a second
			// comment separated from the first by code is not consecutive,
			// even though comments[index-1] exists in the file's comment
			// list. ----
			{
				Code:    "// This Comment Is Fine.\nfoo();\n// lowercase after code",
				Output:  []string{"// This Comment Is Fine.\nfoo();\n// Lowercase after code"},
				Options: []any{"always", map[string]any{"ignoreConsecutiveComments": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 3, Column: 1}},
			},

			// ---- Contract: EndLine/EndColumn on a single-line report,
			// including a multi-character replacement region check ----
			{
				Code:    "// lowercase",
				Output:  []string{"// Lowercase"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},

			// ---- Contract: EndLine/EndColumn on a multi-line block
			// comment report — the diagnostic spans the whole comment, not
			// just its first line. ----
			{
				Code:    "/*\nlowercase\n*/",
				Output:  []string{"/*\nLowercase\n*/"},
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedLowercaseComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 3}},
			},
		},
	)
}

// TestFollowingTokenStarts locks in the single-pass real-token boundary used
// by ignoreInlineComments. Computing this boundary per comment would scan
// all following trivia and make comment-only files quadratic.
func TestFollowingTokenStarts(t *testing.T) {
	t.Parallel()

	source := "first(/* Alpha */ 1); // Bravo\n/* Charlie */ last();"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	_, sourceFile, err := helper.CreateTestProgram(source, "following-token-starts.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	comments := rule.NewCommentStore(sourceFile).All()
	_, got := tokenBoundaries(sourceFile, comments)
	want := []int{strings.Index(source, "1"), strings.Index(source, "last"), strings.Index(source, "last")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("following token starts = %v, want %v", got, want)
	}
}

func BenchmarkTokenBoundaries(b *testing.B) {
	for _, commentCount := range []int{1_000, 2_000, 4_000} {
		b.Run(strconv.Itoa(commentCount), func(b *testing.B) {
			source := strings.Repeat("// Uppercase\n", commentCount) + "value;"
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			_, sourceFile, err := helper.CreateTestProgram(source, "token-boundaries.ts", "tsconfig.json")
			if err != nil {
				b.Fatal(err)
			}
			comments := rule.NewCommentStore(sourceFile).All()

			b.ResetTimer()
			for range b.N {
				preceding, following := tokenBoundaries(sourceFile, comments)
				if len(preceding) != commentCount || len(following) != commentCount {
					b.Fatalf("boundary lengths = %d/%d, want %d/%d", len(preceding), len(following), commentCount, commentCount)
				}
			}
		})
	}
}

// TestCapitalizedCommentsEditDemand exercises Dimension 3 (autofix
// boundaries): diagnostic count, message, and range must stay identical
// across every edit demand, and the fix must materialize only when autofix
// is requested. capitalized-comments has no suggestions, so the suggestion
// demand is expected to carry no artifacts either.
func TestCapitalizedCommentsEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"// lowercase one\n// lowercase two",
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     CapitalizedCommentsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return CapitalizedCommentsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got, want := withoutEdits(diagnostics[index]), withoutEdits(allEdits[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d changed diagnostic %d:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
		if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
			t.Fatalf("diagnostic %d: non-autofix demand materialized fixes", index)
		}
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		if fixes := *allEdits[index].FixesPtr; len(fixes) == 0 {
			t.Fatalf("diagnostic %d: all-edits demand produced no fixes", index)
		}
	}
}
