package server

import (
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type lintDiagnosticProjection struct {
	diagnostics         []api.Diagnostic
	errorCount          int
	warningCount        int
	fixableErrorCount   int
	fixableWarningCount int
}

func projectLintDiagnostics(diagnostics []rule.RuleDiagnostic) lintDiagnosticProjection {
	projection := lintDiagnosticProjection{
		diagnostics: make([]api.Diagnostic, 0, len(diagnostics)),
	}
	for _, diagnostic := range diagnostics {
		projection.diagnostics = append(projection.diagnostics, projectLintDiagnostic(diagnostic))

		hasFix := diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) > 0
		switch diagnostic.Severity {
		case rule.SeverityError:
			projection.errorCount++
			if hasFix {
				projection.fixableErrorCount++
			}
		case rule.SeverityWarning:
			projection.warningCount++
			if hasFix {
				projection.fixableWarningCount++
			}
		}
	}
	return projection
}

func projectLintDiagnostic(diagnostic rule.RuleDiagnostic) api.Diagnostic {
	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(
		diagnostic.SourceFile,
		diagnostic.Range.Pos(),
	)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(
		diagnostic.SourceFile,
		diagnostic.Range.End(),
	)
	projected := api.Diagnostic{
		RuleName:  diagnostic.RuleName,
		MessageId: diagnostic.Message.Id,
		Message:   diagnostic.Message.Description,
		FilePath:  diagnostic.FilePath,
		Range: api.Range{
			Start: api.Position{Line: startLine + 1, Column: int(startColumn) + 1},
			End:   api.Position{Line: endLine + 1, Column: int(endColumn) + 1},
		},
		Severity: diagnostic.Severity.String(),
	}

	text := diagnostic.SourceFile.Text()
	if diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) > 0 {
		projected.Fixes = projectLintFixes(text, *diagnostic.FixesPtr)
	}
	if diagnostic.Suggestions != nil && len(*diagnostic.Suggestions) > 0 {
		projected.Suggestions = make([]api.Suggestion, 0, len(*diagnostic.Suggestions))
		for _, suggestion := range *diagnostic.Suggestions {
			projected.Suggestions = append(projected.Suggestions, api.Suggestion{
				MessageId: suggestion.Message.Id,
				Message:   suggestion.Message.Description,
				Data:      suggestion.Message.Data,
				Fixes:     projectLintFixes(text, suggestion.FixesArr),
			})
		}
	}
	return projected
}

func projectLintFixes(text string, fixes []rule.RuleFix) []api.Fix {
	if len(fixes) == 0 {
		return nil
	}
	projected := make([]api.Fix, 0, len(fixes))
	for _, fix := range fixes {
		projected = append(projected, api.Fix{
			Text:     fix.Text,
			StartPos: byteOffsetToUTF16(text, fix.Range.Pos()),
			EndPos:   byteOffsetToUTF16(text, fix.Range.End()),
		})
	}
	return projected
}

// byteOffsetToUTF16 converts a byte offset within text to a flat UTF-16 code
// unit offset, the unit ESLint uses for fix and suggestion ranges. Diagnostic
// line and column positions instead count UTF-16 units from the line start.
func byteOffsetToUTF16(text string, byteOffset int) int {
	if byteOffset < 0 {
		// ESLint uses [-1, 0] to remove a byte order mark. The negative offset
		// is meaningful as-is and has no prefix text to measure.
		return byteOffset
	}
	if byteOffset == 0 {
		return 0
	}
	if byteOffset >= len(text) {
		return int(core.UTF16Len(text))
	}
	return int(core.UTF16Len(text[:byteOffset]))
}
