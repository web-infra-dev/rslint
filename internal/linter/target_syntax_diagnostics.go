package linter

import (
	"context"

	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type targetSyntacticDiagnosticKey struct {
	path     string
	ruleName string
	pos      int
	end      int
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
		if coveredByProgramDiagnostics || i >= len(targetsByProgram) || len(targetsByProgram[i]) == 0 {
			continue
		}
		ctx := context.Background()
		for _, target := range targetsByProgram[i] {
			file := program.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range CollectFileSyntacticDiagnostics(ctx, program, file) {
				key := targetSyntacticDiagnosticKey{
					path:     diagnostic.FilePath,
					ruleName: diagnostic.RuleName,
					pos:      diagnostic.Range.Pos(),
					end:      diagnostic.Range.End(),
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				diagnostics = append(diagnostics, diagnostic)
			}
		}
	}
	return diagnostics
}
