package linter

import (
	"context"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func newTypeScriptDiagnostic(sourceFile *ast.SourceFile, diagnostic *ast.Diagnostic, description string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
		SourceFile:   sourceFile,
		FilePath:     sourceFile.FileName(),
		Range:        diagnostic.Loc(),
		Message:      rule.RuleMessage{Description: description},
		Severity:     rule.SeverityError,
		Origin:       rule.DiagnosticOriginTypeScript,
		PreFormatted: true,
	}
}

// CollectFileSyntacticDiagnostics projects ts-go syntax diagnostics for one
// source file into rslint's shared diagnostic model.
func CollectFileSyntacticDiagnostics(
	ctx context.Context,
	program *lintprogram.Program,
	sourceFile *ast.SourceFile,
) []rule.RuleDiagnostic {
	if program == nil || sourceFile == nil {
		return nil
	}

	typeScriptDiagnostics := program.SyntacticDiagnostics(ctx, sourceFile)
	diagnostics := make([]rule.RuleDiagnostic, 0, len(typeScriptDiagnostics))
	for _, diagnostic := range typeScriptDiagnostics {
		diagnostics = append(diagnostics, newTypeScriptDiagnostic(sourceFile, diagnostic, diagnostic.String()))
	}
	return diagnostics
}
