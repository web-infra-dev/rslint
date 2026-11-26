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
	// coveredLines memoizes typeScriptCovers for runs of blank or // lines.
	coveredLines map[int]bool
	lineStarts   []core.TextPos
	// 0 until checked, then 1 if the file can contain a TypeScript directive
	// and -1 otherwise.
	typeScriptDirectiveText int8
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
	if g.editsTypeScriptTarget(fixes) {
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
		removed := comment.Pos() < bodyRange.End() && bodyRange.Pos() < comment.End()
		if removed && g.typeScriptDirective(comment) {
			return false // its target would lose its suppression
		}
		value := ecmascript.StringTrim(utils.CommentValue(text, comment))
		if !fixDirectiveComment(comment.Kind, value) {
			continue
		}
		if removed {
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

// A @ts-expect-error or @ts-ignore directive applies to every diagnostic
// whose line is reached by looking back through blank or //-prefixed lines.
// Those lines are classified textually, so they include template and JSX
// text. An edit on any line a directive reaches may change what it
// suppresses, and joining lines moves the edit onto such a line, so both
// are covered by checking each line an edit touches.
//
// The search walks back from each edit, so files whose directives all follow
// the edits, or end before the preceding code line, cost one line check.
func (g *fixDirectiveGuard) editsTypeScriptTarget(fixes []rule.RuleFix) bool {
	if g.typeScriptDirectiveText == 0 {
		text := g.source.Text()
		g.typeScriptDirectiveText = -1
		if len(g.source.CommentDirectives) != 0 || strings.Contains(text, "@ts-expect-error") || strings.Contains(text, "@ts-ignore") {
			g.typeScriptDirectiveText = 1
		}
	}
	if g.typeScriptDirectiveText < 0 {
		return false
	}
	if g.lineStarts == nil {
		g.lineStarts = scanner.GetECMALineStarts(g.source)
	}
	// The edits of one fix are close together: search once, then step.
	start := fixes[0].Range.Pos()
	for _, fix := range fixes[1:] {
		start = min(start, fix.Range.Pos())
	}
	anchor := scanner.ComputeLineOfPosition(g.lineStarts, start)
	lineOf := func(pos int) int {
		line := anchor
		for line+1 < len(g.lineStarts) && int(g.lineStarts[line+1]) <= pos {
			line++
		}
		return line
	}
	checked := -1 // token edits usually share a line
	for _, fix := range fixes {
		last := lineOf(fix.Range.End())
		for line := lineOf(fix.Range.Pos()); line <= last; line++ {
			if line != checked && g.typeScriptCovers(line) {
				return true
			}
		}
		checked = last
	}
	return false
}

func (g *fixDirectiveGuard) typeScriptCovers(line int) bool {
	if covered, ok := g.coveredLines[line]; ok {
		return covered
	}
	covered := false
	previous := line - 1
	for ; previous >= 0; previous-- {
		if g.typeScriptDirectiveOnLine(previous) {
			covered = true
			break
		}
		if !g.blankOrCommentLine(previous) {
			break
		}
		// A skipped line continues the same search.
		if known, ok := g.coveredLines[previous]; ok {
			covered = known
			break
		}
	}
	if previous < line-1 {
		if g.coveredLines == nil {
			g.coveredLines = make(map[int]bool)
		}
		for skipped := max(previous+1, 0); skipped <= line; skipped++ {
			g.coveredLines[skipped] = covered
		}
	}
	return covered
}

func (g *fixDirectiveGuard) lineBounds(line int) (int, int) {
	end := len(g.source.Text())
	if line+1 < len(g.lineStarts) {
		end = int(g.lineStarts[line+1])
	}
	return int(g.lineStarts[line]), end
}

// Mirrors TypeScript 6.0.3's String.prototype.trim and /^\s*\/\/.*$/ test,
// a superset of tsgo's space/tab-only check.
func (g *fixDirectiveGuard) blankOrCommentLine(line int) bool {
	start, end := g.lineBounds(line)
	lineText := g.source.Text()[start:end]
	for i := 0; i < len(lineText); i++ {
		switch b := lineText[i]; {
		case b == ' ' || b == '\t' || b == '\v' || b == '\f':
			continue
		case b < utf8.RuneSelf:
			return b == '\n' || b == '\r' || strings.HasPrefix(lineText[i:], "//")
		}
		lineText = ecmascript.StringTrim(lineText[i:])
		return lineText == "" || strings.HasPrefix(lineText, "//")
	}
	return true
}

// Directives are keyed by the line on which their comment ends.
func (g *fixDirectiveGuard) typeScriptDirectiveOnLine(line int) bool {
	start, end := g.lineBounds(line)
	if end == len(g.source.Text()) {
		end++ // a comment may end at EOF
	}
	// Both TypeScript 6.0.3 and tsgo require the name on the last line.
	lineText := g.source.Text()[start:min(end, len(g.source.Text()))]
	if !strings.Contains(lineText, "@ts-") || !strings.Contains(lineText, "@ts-expect-error") && !strings.Contains(lineText, "@ts-ignore") {
		return false
	}
	directives := g.source.CommentDirectives
	i := sort.Search(len(directives), func(i int) bool { return directives[i].Loc.End() >= start })
	if i < len(directives) && directives[i].Loc.End() < end {
		return true
	}
	comments := g.comments()
	for j := sort.Search(len(comments), func(j int) bool { return comments[j].End() >= start }); j < len(comments) && comments[j].End() < end; j++ {
		if typeScript6Directive(g.source.Text(), comments[j]) {
			return true
		}
	}
	return false
}

// The fix must stay safe under both checkers users run: tsgo, whose
// CommentDirectives the parser records, and TypeScript 6.0.3. Neither
// grammar contains the other: tsgo accepts any number of extra slashes, while
// TypeScript 6.0.3 accepts any ECMAScript whitespace before the @.
func (g *fixDirectiveGuard) typeScriptDirective(comment *ast.CommentRange) bool {
	directives := g.source.CommentDirectives
	i := sort.Search(len(directives), func(i int) bool { return directives[i].Loc.End() >= comment.End() })
	if i < len(directives) && directives[i].Loc.End() == comment.End() {
		return true
	}
	return typeScript6Directive(g.source.Text(), comment)
}

// TypeScript 6.0.3 tests a line comment with /^\/\/\/?\s*@(ts-expect-error|ts-ignore)/
// and the trimmed last line of a block comment with
// /^(?:\/|\*)*\s*@(ts-expect-error|ts-ignore)/.
func typeScript6Directive(text string, comment *ast.CommentRange) bool {
	value := text[comment.Pos():comment.End()]
	trimStart := func(s string) string { return strings.TrimLeftFunc(s, ecmascript.IsWhiteSpaceOrLineTerminator) }
	if comment.Kind == ast.KindSingleLineCommentTrivia {
		value = strings.TrimPrefix(strings.TrimPrefix(value, "//"), "/")
	} else {
		if i := strings.LastIndexAny(value, ecmascript.LineTerminators); i >= 0 {
			_, size := utf8.DecodeRuneInString(value[i:])
			value = value[i+size:]
		}
		value = strings.TrimLeft(trimStart(value), "/*")
	}
	value = trimStart(value)
	return strings.HasPrefix(value, "@ts-expect-error") || strings.HasPrefix(value, "@ts-ignore")
}
