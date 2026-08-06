package no_octal

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoOctalParsedLiterals(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		wantReports int
		wantText    string
	}{
		// ---- ESLint upstream invalid cases (legacy octal + leading-zero decimal) ----
		{name: "01234", code: "01234", wantReports: 1, wantText: "01234"},
		{name: "07", code: "07", wantReports: 1, wantText: "07"},
		{name: "00", code: "00", wantReports: 1, wantText: "00"},
		{name: "08 (leading-zero decimal)", code: "08", wantReports: 1, wantText: "08"},
		{name: "09.1", code: "09.1", wantReports: 1, wantText: "09.1"},
		{name: "09e1", code: "09e1", wantReports: 1, wantText: "09e1"},
		{name: "09.1e1", code: "09.1e1", wantReports: 1, wantText: "09.1e1"},
		{name: "018", code: "018", wantReports: 1, wantText: "018"},
		{name: "019.1", code: "019.1", wantReports: 1, wantText: "019.1"},
		{name: "019e1", code: "019e1", wantReports: 1, wantText: "019e1"},
		{name: "019.1e1", code: "019.1e1", wantReports: 1, wantText: "019.1e1"},
		{name: "separator after legacy octal token", code: "01_2", wantReports: 1, wantText: "01"},
		{name: "separator after leading-zero token", code: "08_9", wantReports: 1, wantText: "08"},

		// ---- ESLint upstream valid cases ----
		{name: "0", code: "0"},
		{name: "0.1", code: "0.1"},
		{name: "0.5e1", code: "0.5e1"},
		{name: "0x1234", code: "0x1234"},
		{name: "0X5", code: "0X5"},

		// ---- Modern literal forms (correctly excluded) ----
		{name: "0o17", code: "0o17"},
		{name: "0O17", code: "0O17"},
		{name: "0b101", code: "0b101"},
		{name: "0B101", code: "0B101"},
		{name: "invalid separator after zero", code: "0_1"},

		// ---- Plain decimals ----
		{name: "1", code: "1"},
		{name: "123", code: "123"},
		{name: "1.5", code: "1.5"},
		{name: "1e5", code: "1e5"},
		{name: ".5", code: ".5"},

		// ---- Report range excludes leading trivia and unary operators ----
		{name: "leading trivia", code: "/* 077 */\n\t01234", wantReports: 1, wantText: "01234"},
		{name: "unary minus", code: "-077", wantReports: 1, wantText: "077"},

		// ---- Suppression still keys off the trimmed literal start ----
		{name: "disabled next line", code: "/* eslint-disable-next-line no-octal */\n077"},
		{name: "other rule disabled", code: "/* eslint-disable-next-line other-rule */\n077", wantReports: 1, wantText: "077"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/literal.js",
				Path:     "/literal.js",
			}, tt.code, core.ScriptKindJS)
			diagnostics := runNoOctalOnParsedSource(sourceFile)
			if len(diagnostics) != tt.wantReports {
				t.Fatalf("diagnostics = %d, want %d", len(diagnostics), tt.wantReports)
			}
			if tt.wantReports != 0 {
				diagnostic := diagnostics[0]
				gotText := sourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotText != tt.wantText {
					t.Errorf("diagnostic range text = %q, want %q", gotText, tt.wantText)
				}
				if diagnostic.Message.Id != "noOctal" {
					t.Errorf("message ID = %q, want noOctal", diagnostic.Message.Id)
				}
			}
		})
	}
}

func runNoOctalOnParsedSource(sourceFile *ast.SourceFile) []rule.RuleDiagnostic {
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(NoOctalRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	listeners := NoOctalRule.Run(ctx, nil)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	return diagnostics
}

func TestNoOctalRule(t *testing.T) {
	// Invalid cases cannot be exercised via the tsconfig-bound rule_tester because
	// the TypeScript parser rejects octal literals (TS1121) and leading-zero
	// decimals (TS1489) as syntactic errors, preventing program creation.
	// Detection logic is covered by TestIsOctalLiteralRaw; production files use the
	// lenient fallback program, where the listener fires normally.
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoOctalRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `var a = 'hello world';`},
			{Code: `0x1234`},
			{Code: `0X5;`},
			{Code: `a = 0;`},
			{Code: `0.1`},
			{Code: `0.5e1`},

			// ---- Modern literal forms that share a leading zero ----
			{Code: `0o17`},
			{Code: `0O17`},
			{Code: `0b101`},
			{Code: `0B101`},
			{Code: `0n`},

			// ---- Plain decimals ----
			{Code: `123`},
			{Code: `1.5`},
			{Code: `1e5`},
			{Code: `.5`},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
