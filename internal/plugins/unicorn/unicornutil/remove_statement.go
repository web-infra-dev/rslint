package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// StatementRemovalRange extends a statement's range to its otherwise empty
// line, leaving leading and trailing comments outside the range.
func StatementRemovalRange(sourceFile *ast.SourceFile, statement core.TextRange) core.TextRange {
	text := sourceFile.Text()
	start, end := statement.Pos(), statement.End()
	lines := sourceFile.ECMALineMap()
	lineStart := int(lines[scanner.ComputeLineOfPosition(lines, start)])
	endLine := scanner.ComputeLineOfPosition(lines, end)
	lineEnd := len(text)
	if endLine+1 < len(lines) {
		lineEnd = utils.LineContentEnd(text, int(lines[endLine+1]))
	}
	if ecmascript.SkipLeadingWhitespace(text, lineStart, start) != start ||
		ecmascript.SkipLeadingWhitespace(text, end, lineEnd) != lineEnd {
		return statement
	}

	end = lineEnd
	if lineStart == 0 {
		start = 0
		end += len(ecmascript.LineTerminatorSequenceAt(text, end))
	} else {
		start = utils.LineContentEnd(text, lineStart)
	}
	return core.NewTextRange(start, end)
}
