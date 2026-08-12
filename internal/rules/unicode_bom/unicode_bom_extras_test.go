package unicode_bom

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

// TestUnicodeBomExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row it covers, so future refactors can't
// silently regress them without breaking a named lock-in.
//
// Dimension 4 rows that do not apply to this rule, and why:
//   - N/A: receiver / expression wrappers (parens, `!`, `as`, `satisfies`,
//     optional chain) — the rule never reads an expression; it tests one
//     property of the file.
//   - N/A: access / key forms (identifier vs string vs numeric vs private vs
//     computed) — the rule inspects no property names.
//   - N/A: declaration / container forms (class vs class expression, function
//     vs arrow vs method, async / generator) — the rule targets no declaration.
//   - N/A: nesting / traversal boundaries — the rule runs once per file and
//     registers no listeners, so there is no traversal to bleed past.
//   - N/A: SpreadAssignment / RestElement graceful degradation — no object or
//     binding pattern is ever visited.
//
// The rows that DO apply are the text-shape analogues of "graceful
// degradation": files with no content for a mark to precede, and a U+FEFF that
// is present but not in the first position.
func TestUnicodeBomExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&UnicodeBomRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream's `defaultOptions: ["never"]`: with no options
			// at all a mark-free file is silent.
			{Code: "var a = 123;"},
			// Same default, reached through an explicitly empty option array.
			{Code: "var a = 123;", Options: []interface{}{}},

			// ---- Dimension 4: graceful degradation — empty file ----
			// Nothing to precede, and `never` is satisfied by the absence of a mark.
			{Code: "", Options: []interface{}{"never"}},

			// ---- Dimension 4: U+FEFF present but not in the first position ----
			// Mid-file: ordinary text, not a mark.
			{Code: "var a = 123; \uFEFF var b = 1;", Options: []interface{}{"never"}},
			// Inside a string literal on line 1.
			{Code: "var a = '\uFEFF';", Options: []interface{}{"never"}},
			// After a leading newline — position 0 is the newline, not the mark.
			{Code: "\n\uFEFFvar a = 123;", Options: []interface{}{"never"}},

			// ---- Dimension 4: a shebang, with no mark in play ----
			{Code: "#!/usr/bin/env node\nvar a = 123;", Options: []interface{}{"never"}},

			// ---- Options via the JSON path: bare string rather than an array ----
			// The CLI hands a single positional option through as `["never"]`;
			// this asserts the rule survives options that are not an array.
			{Code: "var a = 123;", Options: "never"},

			// ---- `always`: the mark is present, which is all the mode asks ----
			{Code: "\uFEFFvar a = 123;", Options: []interface{}{"always"}},
			// A file that is nothing but the mark still carries one.
			{Code: "\uFEFF", Options: []interface{}{"always"}},
			// Doubled: the first is the file's mark, the second is text behind
			// it, and the mode is satisfied by the first.
			{Code: "\uFEFF\uFEFFvar a = 123;", Options: []interface{}{"always"}},
			// A mark ahead of a shebang, the shape Windows editors write.
			{Code: "\uFEFF#!/usr/bin/env node\nvar a = 123;", Options: []interface{}{"always"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: graceful degradation — file is nothing but a mark ----
			{
				Code:    "\uFEFF",
				Options: []interface{}{"never"},
				Output:  []string{""},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected",
					Message:   "Unexpected Unicode BOM (Byte Order Mark).",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- Dimension 4: a doubled leading U+FEFF ----
			// Only the first is the file's mark; the second is ordinary text at
			// the front of it, and one report covers the file either way.
			{
				Code:    "\uFEFF\uFEFFvar a = 123;",
				Options: []interface{}{"never"},
				// Two passes: removing the file's mark promotes the second
				// U+FEFF to first position and makes it the mark in turn.
				Output: []string{"\uFEFFvar a = 123;", "var a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- Real-user: eslint#6580 / #4878 — a mark ahead of a shebang ----
			// Files written by Windows editors. The mark is not part of the
			// text, so `#!` is still the first thing in it and the file parses.
			{
				Code:    "\uFEFF#!/usr/bin/env node\nvar a = 123;",
				Options: []interface{}{"never"},
				Output:  []string{"#!/usr/bin/env node\nvar a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- Dimension 4: CRLF line endings ----
			{
				Code:    "\uFEFFvar a = 123;\r\nvar b = 1;\r\n",
				Options: []interface{}{"never"},
				Output:  []string{"var a = 123;\r\nvar b = 1;\r\n"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- Dimension 4: multibyte content after the mark ----
			// Column math is UTF-16 based; the report must stay at column 1.
			{
				Code:    "\uFEFFvar 日本 = '𝟘';",
				Options: []interface{}{"never"},
				Output:  []string{"var 日本 = '𝟘';"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- Dimension 4: a mark immediately followed by a line break ----
			// The report stays on line 1 even though line 1 holds nothing else.
			{
				Code:    "\uFEFF\nvar a = 123;",
				Options: []interface{}{"never"},
				Output:  []string{"\nvar a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- Dimension 4: .tsx source ----
			{
				Code:    "\uFEFFconst a = <div />;\n",
				Tsx:     true,
				Options: []interface{}{"never"},
				Output:  []string{"const a = <div />;\n"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// Locks in the option default reached with no options configured at
			// all, and again through an explicitly empty option array.
			{
				Code:   "\uFEFFvar a = 123;",
				Output: []string{"var a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},
			{
				Code:    "\uFEFFvar a = 123;",
				Options: []interface{}{},
				Output:  []string{"var a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- `always`: graceful degradation \u2014 an empty file ----
			// There is nothing for the mark to precede, and writing it in is
			// the whole of the fixed file.
			{
				Code:    "",
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFF"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected",
					Message:   "Expected Unicode BOM (Byte Order Mark).",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- `always`: a U+FEFF present but not in the first position ----
			// Mid-file it is ordinary text, so the file still has no mark.
			{
				Code:    "var a = 123; \uFEFFvar b = 1;",
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFFvar a = 123; \uFEFFvar b = 1;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},
			// After a leading newline \u2014 position 0 is the newline, not a mark.
			{
				Code:    "\n\uFEFFvar a = 123;",
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFF\n\uFEFFvar a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- `always`: the mark goes ahead of a shebang ----
			{
				Code:    "#!/usr/bin/env node\nvar a = 123;",
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFF#!/usr/bin/env node\nvar a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- `always`: CRLF line endings ----
			{
				Code:    "var a = 123;\r\nvar b = 1;\r\n",
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFFvar a = 123;\r\nvar b = 1;\r\n"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- `always`: multibyte content, and .tsx source ----
			{
				Code:    "var 日本 = '𝟘';",
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFFvar 日本 = '𝟘';"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},
			{
				Code:    "const a = <div />;\n",
				Tsx:     true,
				Options: []interface{}{"always"},
				Output:  []string{"\uFEFFconst a = <div />;\n"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},

			// ---- `always` via the JSON path: bare string rather than array ----
			{
				Code:    "var a = 123;",
				Options: "always",
				Output:  []string{"\uFEFFvar a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expected", Line: 1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},
		},
	)
}

// TestUnicodeBomSchema locks in the option surface: both modes are accepted,
// and anything else is refused where every other unusable configuration is
// refused — schema validation, before the rule runs — rather than accepted and
// then silently ignored.
func TestUnicodeBomSchema(t *testing.T) {
	t.Parallel()

	assert.NilError(t, UnicodeBomRule.Schema.Validate([]any{"always"}))
	assert.NilError(t, UnicodeBomRule.Schema.Validate([]any{"never"}))
	assert.NilError(t, UnicodeBomRule.Schema.Validate([]any{}))
	assert.Assert(t, UnicodeBomRule.Schema.Validate([]any{"sometimes"}) != nil,
		"the schema must reject a mode that is neither always nor never")
	assert.Assert(t, UnicodeBomRule.Schema.Validate([]any{"never", "never"}) != nil,
		"the schema must reject a second option")
}

// TestUnicodeBomTextNeverCarriesTheMark locks in the contract the rule rests
// on: whatever the source came from, the text a rule sees starts past the mark,
// so every offset a rule reports is measured against the same thing.
func TestUnicodeBomTextNeverCarriesTheMark(t *testing.T) {
	t.Parallel()

	diagnostics := runUnicodeBom(t, "\uFEFFvar a = 123;", rule.EditDemandAll)
	assert.Equal(t, 1, len(diagnostics))
	assert.Equal(t, "var a = 123;", diagnostics[0].SourceFile.Text())
}

// TestUnicodeBomEditDemand asserts that the diagnostic identity (count,
// messageId, message, range) is identical under every edit demand, and that
// the autofix materializes only when EditDemandAutofix is requested. The rule
// emits no suggestions, so EditDemandSuggestion must behave like
// EditDemandNone.
func TestUnicodeBomEditDemand(t *testing.T) {
	t.Parallel()

	const code = "\uFEFFvar a = 123;"
	const withoutBOM = "var a = 123;"

	for _, demand := range []struct {
		name    string
		demand  rule.EditDemand
		wantFix bool
	}{
		{"none", rule.EditDemandNone, false},
		{"autofix", rule.EditDemandAutofix, true},
		{"suggestion", rule.EditDemandSuggestion, false},
		{"all", rule.EditDemandAll, true},
	} {
		t.Run(demand.name, func(t *testing.T) {
			t.Parallel()

			diagnostics := runUnicodeBom(t, code, demand.demand)

			assert.Equal(t, 1, len(diagnostics), "diagnostic count must not depend on edit demand")
			d := diagnostics[0]
			assert.Equal(t, "Unexpected Unicode BOM (Byte Order Mark).", d.Message.Description)
			assert.Equal(t, 0, d.Range.Pos(), "report range must not depend on edit demand")
			assert.Equal(t, 0, d.Range.End(), "report range must not depend on edit demand")
			assert.Assert(t, d.Suggestions == nil, "the rule emits no suggestions")

			fixes := d.Fixes()
			if demand.wantFix {
				assert.Equal(t, 1, len(fixes))
				assert.Equal(t, -1, fixes[0].Range.Pos(), "the mark sits one position ahead of the text")
				assert.Equal(t, 0, fixes[0].Range.End())
			} else {
				assert.Equal(t, 0, len(fixes))
			}

			output, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
			assert.Equal(t, demand.wantFix, fixed)
			if demand.wantFix {
				assert.Equal(t, withoutBOM, output, "the fix removes the mark")
			} else {
				assert.Equal(t, code, output, "no fix, no change")
			}
		})
	}
}

// runUnicodeBom lints one snippet with the rule bound to an explicit edit
// demand, which RunRuleTester cannot express (it always requests every edit
// category).
func runUnicodeBom(t *testing.T, code string, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()

	root := fixtures.GetRootDir()
	fileName := "file.ts"
	fs := utils.NewOverlayVFS(root.FS, map[string]string{tspath.ResolvePath(root.Dir, fileName): code})
	host := utils.CreateCompilerHost(root.Dir, fs)

	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	assert.NilError(t, err)

	var diagnostics []rule.RuleDiagnostic
	_, err = linter.RunLinter(linter.RunLinterOptions{
		Programs:       []*lintprogram.Program{lintprogram.NewTypeScript(program)},
		SingleThreaded: true,
		Scope:          linter.FileScope{Files: []string{program.GetSourceFile(fileName).FileName()}},
		ExcludePaths:   []string{},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     UnicodeBomRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return UnicodeBomRule.Run(ctx, []any{"never"})
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
		},
	})
	assert.NilError(t, err)

	return diagnostics
}
