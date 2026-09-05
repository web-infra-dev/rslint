package no_empty_object_type

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoEmptyObjectTypeCrossFileDeclarationMergeSuggestions(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		declaration string
	}{
		{name: "interface", declaration: `interface Shared { value: string }`},
		{name: "class", declaration: `class Shared { value = "value" }`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root := fixtures.GetRootDir()
			interfacePath := tspath.ResolvePath(root.Dir, "merge-interface.ts")
			otherPath := tspath.ResolvePath(root.Dir, "merge-other.ts")
			fs := utils.NewOverlayVFS(root.FS, map[string]string{
				interfacePath: `interface Shared {}`,
				otherPath:     test.declaration,
			})
			host := utils.CreateCompilerHost(root.Dir, fs)
			program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
			if err != nil {
				t.Fatalf("failed to create program: %v", err)
			}
			sourceFile := program.GetSourceFile(interfacePath)
			if sourceFile == nil {
				t.Fatalf("source file %q not found", interfacePath)
			}

			var diagnostics []rule.RuleDiagnostic
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program:     lintprogram.NewFromCompiler(program),
				File:        sourceFile.FileName(),
				HasTypeInfo: true,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:     NoEmptyObjectTypeRule.Name,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return NoEmptyObjectTypeRule.Run(ctx, nil)
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{
					Demand: rule.EditDemandSuggestion,
					Report: func(diagnostic rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, diagnostic)
					},
				},
			})

			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
			}
			if diagnostics[0].Message.Id != "noEmptyInterface" {
				t.Fatalf("message id = %q, want noEmptyInterface", diagnostics[0].Message.Id)
			}
			if diagnostics[0].Suggestions != nil {
				t.Fatalf("suggestions = %d, want none for a cross-file declaration merge", len(*diagnostics[0].Suggestions))
			}
		})
	}
}
