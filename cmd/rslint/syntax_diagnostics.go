package main

import (
	"context"
	"fmt"

	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type syntacticDiagnosticKey struct {
	path string
	code int32
	pos  int
	end  int
}

func collectTargetSyntacticDiagnostics(
	programs []*lintprogram.Program,
	typeCheckPrograms []*lintprogram.Program,
	targetsByProgram [][]string,
	typeCheck bool,
	typeCheckOnly bool,
) []rule.RuleDiagnostic {
	if len(programs) == 0 || len(targetsByProgram) == 0 {
		return nil
	}

	seen := make(map[syntacticDiagnosticKey]struct{})
	var typeCheckProgramSet map[*lintprogram.Program]struct{}
	if typeCheck && typeCheckPrograms != nil {
		typeCheckProgramSet = make(map[*lintprogram.Program]struct{}, len(typeCheckPrograms))
		for _, program := range typeCheckPrograms {
			typeCheckProgramSet[program] = struct{}{}
		}
	}
	var diagnostics []rule.RuleDiagnostic
	for i, program := range programs {
		// When --type-check runs, tsconfig-backed Programs surface syntactic
		// diagnostics through the type-check phase only when they belong to its
		// catalog. Lint-only effective projects still report target syntax here.
		// We inspect every target so the lint-rule phase can skip malformed files,
		// matching ESLint.
		_, explicitlyCovered := typeCheckProgramSet[program]
		coveredByTypeCheck := typeCheck &&
			program.CanProvideProgramDiagnostics() &&
			(typeCheckPrograms == nil || explicitlyCovered)
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
				if coveredByTypeCheck || typeCheckOnly {
					continue
				}
				loc := diagnostic.Loc()
				key := syntacticDiagnosticKey{
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
