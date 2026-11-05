package output

import (
	"fmt"
	"math"
)

func validateDiagnostic(
	diagnostic *Diagnostic,
	requireSource bool,
) error {
	if diagnostic.Start.Line < 0 || diagnostic.End.Line < diagnostic.Start.Line ||
		diagnostic.Start.Column < 0 || diagnostic.End.Column < 0 ||
		diagnostic.Start.Line == math.MaxInt || diagnostic.End.Line == math.MaxInt ||
		diagnostic.Start.Column == math.MaxInt || diagnostic.End.Column == math.MaxInt ||
		(diagnostic.Start.Line == diagnostic.End.Line && diagnostic.End.Column < diagnostic.Start.Column) {
		return fmt.Errorf(
			"diagnostic %q for %q has invalid projected locations",
			diagnostic.RuleName,
			diagnostic.FilePath,
		)
	}
	if requireSource {
		if diagnostic.Source == nil {
			return fmt.Errorf(
				"diagnostic %q for %q has no projected source",
				diagnostic.RuleName,
				diagnostic.FilePath,
			)
		}
		start, end := diagnostic.Range.Start, diagnostic.Range.End
		textLength := len(diagnostic.Source.text)
		if start < 0 || end < start || end > textLength {
			return fmt.Errorf(
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
			return fmt.Errorf(
				"diagnostic %q for %q has invalid projected source locations",
				diagnostic.RuleName,
				diagnostic.FilePath,
			)
		}
		if !diagnostic.Source.lineContains(diagnostic.Start.Line, start) ||
			!diagnostic.Source.lineContains(diagnostic.End.Line, end) {
			return fmt.Errorf(
				"diagnostic %q for %q has projected lines inconsistent with its source range",
				diagnostic.RuleName,
				diagnostic.FilePath,
			)
		}
	}

	return nil
}

func (source *DiagnosticSource) lineContains(line, offset int) bool {
	return source.lineStarts[line] <= offset &&
		(line+1 == len(source.lineStarts) || offset < source.lineStarts[line+1])
}
