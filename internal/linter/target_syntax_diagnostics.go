package linter

import (
	"context"
	"fmt"

	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type targetSyntacticDiagnosticKey struct {
	path string
	code int32
	pos  int
	end  int
}

// CollectTargetSyntacticDiagnostics returns syntax diagnostics for the exact
// lint projection. When program diagnostics are included by the caller,
// project-backed Programs are skipped because their syntax diagnostics are
// emitted by that phase; source-only Programs remain covered here.
func CollectTargetSyntacticDiagnostics(
	programs []*lintprogram.Program,
	targetsByProgram [][]string,
	programDiagnosticsIncluded bool,
) []rule.RuleDiagnostic {
	if len(programs) == 0 || len(targetsByProgram) == 0 {
		return nil
	}

	seen := make(map[targetSyntacticDiagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for i, program := range programs {
		coveredByProgramDiagnostics := programDiagnosticsIncluded && program.CanProvideProgramDiagnostics()
		if i >= len(targetsByProgram) || len(targetsByProgram[i]) == 0 {
			continue
		}
		ctx := context.Background()
		for _, target := range targetsByProgram[i] {
			file := program.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range program.SyntacticDiagnostics(ctx, file) {
				if coveredByProgramDiagnostics {
					continue
				}
				loc := diagnostic.Loc()
				key := targetSyntacticDiagnosticKey{
					path: file.FileName(),
					code: diagnostic.Code(),
					pos:  loc.Pos(),
					end:  loc.End(),
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				diagnostics = append(diagnostics, rule.RuleDiagnostic{
					RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
					SourceFile:   file,
					FilePath:     file.FileName(),
					Range:        loc,
					Message:      rule.RuleMessage{Description: diagnostic.String()},
					Severity:     rule.SeverityError,
					Origin:       rule.DiagnosticOriginTypeScript,
					PreFormatted: true,
				})
			}
		}
	}
	return diagnostics
}
