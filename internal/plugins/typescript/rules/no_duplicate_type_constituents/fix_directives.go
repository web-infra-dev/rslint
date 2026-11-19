package no_duplicate_type_constituents

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// A comment's text is only part of its meaning: line directives also depend on
// the code they cover. This guard runs before reserving any member for removal.
type fixDirectiveGuard struct {
	source   *ast.SourceFile
	comments func() []*ast.CommentRange
}

func (g fixDirectiveGuard) safe(bodyRange core.TextRange, triviaStart int, fixes []rule.RuleFix) bool {
	text := g.source.Text()
	start, end := bodyRange.Pos(), bodyRange.End()
	removesLines := false
	for _, fix := range fixes {
		start = min(start, fix.Range.Pos())
		end = max(end, fix.Range.End())
		if strings.ContainsAny(text[fix.Range.Pos():fix.Range.End()], ecmascript.LineTerminators) {
			removesLines = true
		}
	}
	// Parser-owned metadata handles @ts-ignore/@ts-expect-error without
	// mistaking directive-like strings for comments or scanning unrelated ASTs.
	if g.affectsTypeScriptDirective(start, end, triviaStart) {
		return false
	}
	inspectStart := min(start, triviaStart)
	inspectEnd := end
	if removesLines {
		// A directive just before the edited line may cover a multiline body.
		line := scanner.GetECMALineOfPosition(g.source, inspectStart)
		lineStarts := scanner.GetECMALineStarts(g.source)
		inspectStart = int(lineStarts[max(0, line-1)])
		endLine := scanner.GetECMALineOfPosition(g.source, end)
		if endLine+1 < len(lineStarts) {
			inspectEnd = int(lineStarts[endLine+1])
		} else {
			inspectEnd = len(text)
		}
	}
	span := text[inspectStart:inspectEnd]
	if !strings.Contains(span, "//") && !strings.Contains(span, "/*") && !strings.Contains(span, "*/") {
		return true
	}
	// The cheap span check keeps the shared store lazy for string | string,
	// even if the file has a header comment and thousands of other declarations.
	comments := g.comments()
	first := sort.Search(len(comments), func(i int) bool { return comments[i].End() > inspectStart })
	for _, comment := range comments[first:] {
		if comment.Pos() >= inspectEnd {
			break
		}
		value := ecmascript.StringTrim(utils.CommentValue(text, comment))
		if !fixDirectiveComment(comment.Kind, value) {
			continue
		}
		if comment.Pos() < bodyRange.End() && bodyRange.Pos() < comment.End() {
			return false // deleting a global/block directive changes file semantics
		}
		if removesLines && fixLineDirective(value) {
			targetLine := scanner.GetECMALineOfPosition(g.source, comment.End())
			if strings.HasPrefix(value, "eslint-disable-next-line") || strings.HasPrefix(value, "rslint-disable-next-line") {
				targetLine++
			}
			if targetLine >= scanner.GetECMALineOfPosition(g.source, start) && targetLine <= scanner.GetECMALineOfPosition(g.source, end) {
				return false // retained directive must not slide onto another declaration
			}
		}
	}
	return true
}

// This rule protects the directives accepted by the runtime, rather than
// expanding IsDirectiveComment's upstream-compatible grammar for other rules.
func fixDirectiveComment(kind ast.Kind, value string) bool {
	if utils.IsDirectiveComment(kind, value) {
		return true
	}
	if kind != ast.KindMultiLineCommentTrivia {
		return false
	}
	for _, keyword := range []string{"global", "globals", "exported", "eslint"} {
		if rest, ok := strings.CutPrefix(value, keyword); ok {
			if rest == "" {
				return true
			}
			r, _ := utf8.DecodeRuneInString(rest)
			if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				return true
			}
		}
	}
	return false
}

func fixLineDirective(value string) bool {
	for _, prefix := range []string{"eslint-", "rslint-"} {
		if rest, ok := strings.CutPrefix(value, prefix); ok {
			for _, label := range []string{"disable-next-line", "disable-line"} {
				if suffix, ok := strings.CutPrefix(rest, label); ok {
					if suffix == "" {
						return true
					}
					r, _ := utf8.DecodeRuneInString(suffix)
					return ecmascript.IsWhiteSpaceOrLineTerminator(r)
				}
			}
		}
	}
	return false
}

func (g fixDirectiveGuard) affectsTypeScriptDirective(start, end, triviaStart int) bool {
	directives := g.source.CommentDirectives
	last := sort.Search(len(directives), func(i int) bool { return directives[i].Loc.Pos() >= end })
	if last == 0 {
		return false
	}
	directive := directives[last-1]
	if directive.Loc.End() > min(start, triviaStart) {
		return true // directive in the member or surrounding wrapper/trivia
	}
	// TypeScript searches preceding blank/comment-only lines for a suppression.
	// Keep a directive above the affected line bound to its original target.
	lineStarts := scanner.GetECMALineStarts(g.source)
	startLine := scanner.GetECMALineOfPosition(g.source, start)
	directiveLine := scanner.GetECMALineOfPosition(g.source, directive.Loc.Pos())
	if directiveLine >= startLine {
		return false
	}
	for line := directiveLine + 1; line < startLine; line++ {
		lineText := g.source.Text()[int(lineStarts[line]):int(lineStarts[line+1])]
		lineText = strings.TrimLeft(lineText, " \t")
		if !strings.HasPrefix(lineText, "//") && !strings.HasPrefix(lineText, "\r") && !strings.HasPrefix(lineText, "\n") {
			return false
		}
	}
	return true
}
