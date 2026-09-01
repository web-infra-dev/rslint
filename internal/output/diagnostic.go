package output

import (
	"fmt"
	"math"
	"sort"

	"github.com/microsoft/typescript-go/shim/tspath"
)

type location struct {
	line   int
	column int
}

type diagnosticView struct {
	raw          Diagnostic
	relativePath string
	start        location
	end          location
}

func newDiagnosticView(
	diagnostic Diagnostic,
	paths tspath.ComparePathsOptions,
	requireSource bool,
) (diagnosticView, error) {
	if diagnostic.Start.Line < 0 || diagnostic.End.Line < diagnostic.Start.Line ||
		diagnostic.Start.Column < 0 || diagnostic.End.Column < 0 ||
		diagnostic.Start.Line == math.MaxInt || diagnostic.End.Line == math.MaxInt ||
		diagnostic.Start.Column == math.MaxInt || diagnostic.End.Column == math.MaxInt ||
		(diagnostic.Start.Line == diagnostic.End.Line && diagnostic.End.Column < diagnostic.Start.Column) {
		return diagnosticView{}, fmt.Errorf(
			"diagnostic %q for %q has invalid projected locations",
			diagnostic.RuleName,
			diagnostic.FilePath,
		)
	}
	if requireSource {
		if diagnostic.Source == nil {
			return diagnosticView{}, fmt.Errorf(
				"diagnostic %q for %q has no projected source",
				diagnostic.RuleName,
				diagnostic.FilePath,
			)
		}
		start, end := diagnostic.Range.Start, diagnostic.Range.End
		textLength := len(diagnostic.Source.text)
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
		if len(diagnostic.Source.lineStarts) == 0 ||
			diagnostic.Start.Line >= len(diagnostic.Source.lineStarts) ||
			diagnostic.End.Line >= len(diagnostic.Source.lineStarts) {
			return diagnosticView{}, fmt.Errorf(
				"diagnostic %q for %q has invalid projected source locations",
				diagnostic.RuleName,
				diagnostic.FilePath,
			)
		}
		if diagnostic.Start.Line != diagnostic.Source.lineAt(start) ||
			diagnostic.End.Line != diagnostic.Source.lineAt(end) {
			return diagnosticView{}, fmt.Errorf(
				"diagnostic %q for %q has projected lines inconsistent with its source range",
				diagnostic.RuleName,
				diagnostic.FilePath,
			)
		}
	}

	return diagnosticView{
		raw:          diagnostic,
		relativePath: tspath.ConvertToRelativePath(diagnostic.FilePath, paths),
		start:        location{line: diagnostic.Start.Line, column: diagnostic.Start.Column},
		end:          location{line: diagnostic.End.Line, column: diagnostic.End.Column},
	}, nil
}

func (source DiagnosticSource) lineAt(offset int) int {
	return sort.Search(len(source.lineStarts), func(index int) bool {
		return source.lineStarts[index] > offset
	}) - 1
}
