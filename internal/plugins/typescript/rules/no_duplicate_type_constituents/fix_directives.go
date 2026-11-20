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
	source    *ast.SourceFile
	comments  func() []*ast.CommentRange
	tsTargets []typeScriptDirectiveTarget
}

type typeScriptDirectiveTarget struct {
	directive core.TextRange
	target    core.TextRange
}

func (g *fixDirectiveGuard) safe(bodyRange core.TextRange, triviaStart int, fixes []rule.RuleFix) bool {
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
	if g.affectsTypeScriptDirective(fixes) {
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

// Targets follow TypeScript 6.0.3: directives are keyed by their ending
// line, and diagnostics look back through trimmed blank or // comment lines.
// A shared advancing cursor visits each intervening line only once per file.
func (g *fixDirectiveGuard) typeScriptTargets() []typeScriptDirectiveTarget {
	directives := g.source.CommentDirectives
	if len(directives) == 0 {
		return nil
	}
	if g.tsTargets != nil {
		return g.tsTargets
	}
	text := g.source.Text()
	lineStarts := scanner.GetECMALineStarts(g.source)
	g.tsTargets = make([]typeScriptDirectiveTarget, 0, len(directives))
	nextLine := 0
	for _, directive := range directives {
		directiveLine := scanner.GetECMALineOfPosition(g.source, directive.Loc.End())
		nextLine = max(nextLine, directiveLine+1)
		for nextLine < len(lineStarts) {
			lineEnd := len(text)
			if nextLine+1 < len(lineStarts) {
				lineEnd = int(lineStarts[nextLine+1])
			}
			lineText := ecmascript.StringTrim(text[int(lineStarts[nextLine]):lineEnd])
			if lineText != "" && !strings.HasPrefix(lineText, "//") {
				break
			}
			nextLine++
		}
		targetStart, targetEnd := len(text), len(text)
		if nextLine < len(lineStarts) {
			targetStart = int(lineStarts[nextLine])
			if nextLine+1 < len(lineStarts) {
				targetEnd = int(lineStarts[nextLine+1])
			}
		}
		g.tsTargets = append(g.tsTargets, typeScriptDirectiveTarget{directive: directive.Loc, target: core.NewTextRange(targetStart, targetEnd)})
	}
	return g.tsTargets
}

func (g *fixDirectiveGuard) affectsTypeScriptDirective(fixes []rule.RuleFix) bool {
	targets := g.typeScriptTargets()
	if len(targets) == 0 {
		return false
	}
	text := g.source.Text()
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		// Both sequences are monotonic: each target follows its directive.
		// Skip directives whose target ended before this edit, and those whose
		// comments start after it. Retained comments in gaps between separate
		// token edits are not treated as removed.
		first := sort.Search(len(targets), func(i int) bool { return targets[i].target.End() > start })
		last := sort.Search(len(targets), func(i int) bool { return targets[i].directive.Pos() >= end })
		for i := first; i < last; i++ {
			item := targets[i]
			if start < item.directive.End() && item.directive.Pos() < end {
				return true
			}
			if start < item.target.End() && item.target.Pos() < end {
				return true
			}
			gapStart, gapEnd := max(start, item.directive.End()), min(end, item.target.Pos())
			if gapStart < gapEnd && strings.ContainsAny(text[gapStart:gapEnd], ecmascript.LineTerminators) {
				return true
			}
		}
	}
	return false
}
