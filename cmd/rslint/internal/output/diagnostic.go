package output

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/hostpath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type location struct {
	line   int
	column int
}

type diagnosticView struct {
	raw          rule.RuleDiagnostic
	relativePath string
	start        location
	end          location
}

type fileWarningView struct {
	raw          FileWarning
	relativePath string
}

func newFileWarningView(warning FileWarning, paths tspath.ComparePathsOptions) fileWarningView {
	return fileWarningView{
		raw: warning,
		relativePath: hostpath.ConvertToRelativePath(
			warning.FilePath,
			paths.CurrentDirectory,
			paths.UseCaseSensitiveFileNames,
		),
	}
}

func newDiagnosticView(diagnostic rule.RuleDiagnostic, paths tspath.ComparePathsOptions) (diagnosticView, error) {
	if diagnostic.SourceFile == nil {
		return diagnosticView{}, fmt.Errorf("diagnostic %q for %q has no source file", diagnostic.RuleName, diagnostic.FilePath)
	}

	start, end := diagnostic.Range.Pos(), diagnostic.Range.End()
	textLength := len(diagnostic.SourceFile.Text())
	if start < 0 || end < start || end > textLength {
		return diagnosticView{}, fmt.Errorf(
			"diagnostic %q for %q has invalid range [%d,%d) for source length %d",
			diagnostic.RuleName,
			diagnostic.FilePath,
			start,
			end,
			textLength,
		)
	}

	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, start)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, end)

	return diagnosticView{
		raw: diagnostic,
		relativePath: hostpath.ConvertToRelativePath(
			diagnostic.FilePath,
			paths.CurrentDirectory,
			paths.UseCaseSensitiveFileNames,
		),
		start: location{line: startLine, column: int(startColumn)},
		end:   location{line: endLine, column: int(endColumn)},
	}, nil
}
