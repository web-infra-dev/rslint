package no_irregular_whitespace

import (
	"sort"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-irregular-whitespace
var NoIrregularWhitespaceRule = rule.Rule{
	Name: "no-irregular-whitespace",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		checkNoIrregularWhitespace(ctx, opts)
		return rule.RuleListeners{}
	},
}

type irregularWhitespaceOptions struct {
	skipStrings   bool
	skipComments  bool
	skipRegExps   bool
	skipTemplates bool
	skipJSXText   bool
}

func parseOptions(options any) irregularWhitespaceOptions {
	opts := irregularWhitespaceOptions{
		skipStrings:   true,
		skipComments:  false,
		skipRegExps:   false,
		skipTemplates: false,
		skipJSXText:   false,
	}

	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}

	if v, ok := optsMap["skipStrings"].(bool); ok {
		opts.skipStrings = v
	}
	if v, ok := optsMap["skipComments"].(bool); ok {
		opts.skipComments = v
	}
	if v, ok := optsMap["skipRegExps"].(bool); ok {
		opts.skipRegExps = v
	}
	if v, ok := optsMap["skipTemplates"].(bool); ok {
		opts.skipTemplates = v
	}
	if v, ok := optsMap["skipJSXText"].(bool); ok {
		opts.skipJSXText = v
	}

	return opts
}

// isIrregularWhitespace returns true if the rune is an irregular whitespace character
// (not a normal space or tab, but also not a standard line terminator \n, \r).
func isIrregularWhitespace(ch rune) bool {
	switch ch {
	case
		'\u000B', // verticalTab
		'\u000C', // formFeed
		'\u0085', // nextLine
		'\u00A0', // nonBreakingSpace
		'\u1680', // ogham
		'\u180E', // mongolianVowelSeparator
		'\u2000', // enQuad
		'\u2001', // emQuad
		'\u2002', // enSpace
		'\u2003', // emSpace
		'\u2004', // threePerEmSpace
		'\u2005', // fourPerEmSpace
		'\u2006', // sixPerEmSpace
		'\u2007', // figureSpace
		'\u2008', // punctuationSpace
		'\u2009', // thinSpace
		'\u200A', // hairSpace
		'\u200B', // zeroWidthSpace
		'\u202F', // narrowNoBreakSpace
		'\u205F', // mediumMathematicalSpace
		'\u3000', // ideographicSpace
		'\uFEFF', // byteOrderMark
		'\u2028', // lineSeparator
		'\u2029': // paragraphSeparator
		return true
	}
	return false
}

type errorInfo struct {
	pos int
	end int
}

func checkNoIrregularWhitespace(ctx rule.RuleContext, opts irregularWhitespaceOptions) {
	text := ctx.SourceFile.Text()
	if len(text) == 0 {
		return
	}

	// Collect all irregular whitespace positions.
	var errors []errorInfo
	i := 0

	// Skip BOM at position 0 — matches ESLint behavior where BOM is
	// stripped from the source before rules see it.
	if len(text) >= 3 && text[0] == 0xEF && text[1] == 0xBB && text[2] == 0xBF {
		// UTF-8 BOM for U+FEFF
		i = 3
	}

	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}

		if isIrregularWhitespace(r) {
			// Group consecutive irregular whitespace (non-line-terminator) into one error,
			// matching ESLint's IRREGULAR_WHITESPACE regex which matches runs of chars.
			// But line terminators (\u2028, \u2029) are always reported individually.
			if r == '\u2028' || r == '\u2029' {
				errors = append(errors, errorInfo{pos: i, end: i + size})
			} else {
				start := i
				end := i + size
				for end < len(text) {
					nextR, nextSize := utf8.DecodeRuneInString(text[end:])
					if isIrregularWhitespace(nextR) && nextR != '\u2028' && nextR != '\u2029' {
						end += nextSize
					} else {
						break
					}
				}
				errors = append(errors, errorInfo{pos: start, end: end})
				i = end
				continue
			}
		}

		i += size
	}

	if len(errors) == 0 {
		return
	}

	// Collect exempt ranges based on options.
	var exemptRanges []errorInfo
	collectExemptRanges(ctx, opts, &exemptRanges)

	// Sort exempt ranges by position for efficient filtering.
	sort.Slice(exemptRanges, func(i, j int) bool {
		return exemptRanges[i].pos < exemptRanges[j].pos
	})

	// Filter errors: remove those that fall entirely inside an exempt range.
	msg := rule.RuleMessage{
		Id:          "noIrregularWhitespace",
		Description: "Irregular whitespace not allowed.",
	}

	for _, err := range errors {
		if isInsideExemptRange(err, exemptRanges) {
			continue
		}
		ctx.ReportRange(core.NewTextRange(err.pos, err.end), msg)
	}
}

// isInsideExemptRange checks if the error falls entirely within any exempt range.
func isInsideExemptRange(err errorInfo, exemptRanges []errorInfo) bool {
	// Binary search for the first exempt range that could contain this error.
	idx := sort.Search(len(exemptRanges), func(i int) bool {
		return exemptRanges[i].end > err.pos
	})
	for i := idx; i < len(exemptRanges); i++ {
		ex := exemptRanges[i]
		if ex.pos > err.pos {
			break
		}
		if err.pos >= ex.pos && err.end <= ex.end {
			return true
		}
	}
	return false
}

func collectExemptRanges(ctx rule.RuleContext, opts irregularWhitespaceOptions, ranges *[]errorInfo) {
	sf := ctx.SourceFile

	// trimmedPos returns the start of the actual token, skipping leading trivia.
	// node.Pos() includes leading whitespace/comments; we need the real token start.
	trimmedPos := func(node *ast.Node) int {
		r := scanner.GetRangeOfTokenAtPosition(sf, node.Pos())
		return r.Pos()
	}

	// Walk the AST to collect exempt node ranges.
	if opts.skipStrings || opts.skipRegExps || opts.skipTemplates || opts.skipJSXText {
		var walk func(node *ast.Node) bool
		walk = func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindStringLiteral:
				if opts.skipStrings {
					*ranges = append(*ranges, errorInfo{pos: trimmedPos(node), end: node.End()})
				}
			case ast.KindRegularExpressionLiteral:
				if opts.skipRegExps {
					*ranges = append(*ranges, errorInfo{pos: trimmedPos(node), end: node.End()})
				}
			case ast.KindNoSubstitutionTemplateLiteral:
				if opts.skipTemplates {
					*ranges = append(*ranges, errorInfo{pos: trimmedPos(node), end: node.End()})
				}
			case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
				if opts.skipTemplates {
					*ranges = append(*ranges, errorInfo{pos: trimmedPos(node), end: node.End()})
				}
			case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
				if opts.skipJSXText {
					*ranges = append(*ranges, errorInfo{pos: node.Pos(), end: node.End()})
				}
			}
			node.ForEachChild(walk)
			return false
		}
		sf.AsNode().ForEachChild(walk)
	}

	// Collect comment ranges.
	if opts.skipComments {
		nodeFactory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
		text := sf.Text()
		// Iterate through all comments in the file using ForEachComment.
		utils.ForEachComment(sf.AsNode(), func(comment *ast.CommentRange) {
			*ranges = append(*ranges, errorInfo{pos: comment.Pos(), end: comment.End()})
		}, sf)
		// Also check for leading comments at position 0 which ForEachComment may miss.
		for comment := range scanner.GetLeadingCommentRanges(nodeFactory, text, 0) {
			*ranges = append(*ranges, errorInfo{pos: comment.Pos(), end: comment.End()})
		}
	}
}
