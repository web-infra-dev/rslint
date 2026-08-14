package main

import (
	"context"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func typeScriptRuleDiagnostic(file *ast.SourceFile, diagnostic *ast.Diagnostic) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
		SourceFile:   file,
		FilePath:     file.FileName(),
		Range:        diagnostic.Loc(),
		Message:      rule.RuleMessage{Description: diagnostic.String()},
		Severity:     rule.SeverityError,
		Origin:       rule.DiagnosticOriginTypeScript,
		PreFormatted: true,
	}
}

func buildGapPrograms(
	groups [][]resolvedLintTarget,
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) ([]*lintprogram.Program, []rule.RuleDiagnostic, error) {
	if len(groups) == 0 {
		return nil, nil, nil
	}

	programs := make([]*lintprogram.Program, len(groups))
	var diagnostics []rule.RuleDiagnostic
	for groupIndex, group := range groups {
		rootFileNames := make([]string, len(group))
		for targetIndex, target := range group {
			rootFileNames[targetIndex] = target.Path
		}
		gapProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   rootFileNames,
			Host:            buildContext.NewTransientCompilerHost(currentDirectory),
			CompilerOptions: fallbackCompilerOptions(),
			SingleThreaded:  singleThreaded,
		})
		if err != nil {
			return nil, nil, err
		}
		programs[groupIndex] = gapProgram
		for _, file := range gapProgram.SourceFiles() {
			fileDiagnostics := gapProgram.SyntacticDiagnostics(context.Background(), file)
			for _, diagnostic := range fileDiagnostics {
				diagnostics = append(diagnostics, typeScriptRuleDiagnostic(file, diagnostic))
			}
		}
	}
	return programs, diagnostics, nil
}

// appendGapPrograms preserves CLI binding order while appending already-built
// gap generations to the rslint Program sequence. Target selection remains run
// metadata rather than becoming Program state; source generations without
// program-wide diagnostics simply return no diagnostics through the facade.
func appendGapPrograms(
	boundPrograms []*lintprogram.Program,
	gapPrograms []*lintprogram.Program,
	targetFiles [][]string,
) ([]*lintprogram.Program, [][]string) {
	programs := append([]*lintprogram.Program(nil), boundPrograms...)
	programs = append(programs, gapPrograms...)
	if len(gapPrograms) == 0 {
		return programs, targetFiles
	}

	combinedTargets := make([][]string, len(boundPrograms), len(programs))
	copy(combinedTargets, targetFiles)
	for _, gapProgram := range gapPrograms {
		files := gapProgram.SourceFiles()
		targets := make([]string, len(files))
		for fileIndex, file := range files {
			targets[fileIndex] = file.FileName()
		}
		combinedTargets = append(combinedTargets, targets)
	}

	return programs, combinedTargets
}
