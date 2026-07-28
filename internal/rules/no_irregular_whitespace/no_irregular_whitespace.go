package no_irregular_whitespace

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-irregular-whitespace
var NoIrregularWhitespaceRule = rule.Rule{
	Name: "no-irregular-whitespace",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
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

// irregularWhitespaceSizeAt recognizes the UTF-8 encodings used by ESLint's
// IRREGULAR_WHITESPACE regex. Scanning bytes directly avoids decoding every
// ordinary source character into a rune. The second result distinguishes the
// two line terminators, which ESLint reports separately instead of grouping.
func irregularWhitespaceSizeAt(text string, pos int) (size int, lineTerminator bool) {
	switch text[pos] {
	case '\u000B', '\u000C':
		return 1, false
	case 0xC2:
		if pos+1 < len(text) && (text[pos+1] == 0x85 || text[pos+1] == 0xA0) {
			return 2, false
		}
	case 0xE1:
		if pos+2 < len(text) &&
			(text[pos+1] == 0x9A && text[pos+2] == 0x80 || // U+1680
				text[pos+1] == 0xA0 && text[pos+2] == 0x8E) { // U+180E
			return 3, false
		}
	case 0xE2:
		if pos+2 >= len(text) {
			return 0, false
		}
		switch text[pos+1] {
		case 0x80:
			last := text[pos+2]
			switch {
			case last >= 0x80 && last <= 0x8B:
				return 3, false // U+2000..U+200B
			case last == 0xA8 || last == 0xA9:
				return 3, true // U+2028, U+2029
			case last == 0xAF:
				return 3, false // U+202F
			}
		case 0x81:
			if text[pos+2] == 0x9F {
				return 3, false // U+205F
			}
		}
	case 0xE3:
		if pos+2 < len(text) && text[pos+1] == 0x80 && text[pos+2] == 0x80 {
			return 3, false // U+3000
		}
	case 0xEF:
		if pos+2 < len(text) && text[pos+1] == 0xBB && text[pos+2] == 0xBF {
			return 3, false // U+FEFF
		}
	}
	return 0, false
}

type errorInfo struct {
	pos int
	end int
}

var irregularWhitespaceFirstBytes = [256]bool{
	'\u000B': true,
	'\u000C': true,
	0xC2:     true,
	0xE1:     true,
	0xE2:     true,
	0xE3:     true,
	0xEF:     true,
}

func mayContainIrregularWhitespace(text string) bool {
	// IndexByte uses the standard library's optimized byte search. Most files
	// contain none of these possible UTF-8 lead bytes, so avoid even the cheap
	// byte-by-byte classifier below on the overwhelmingly common clean path.
	return strings.IndexByte(text, 0xE2) >= 0 ||
		strings.IndexByte(text, 0xEF) >= 0 ||
		strings.IndexByte(text, 0xC2) >= 0 ||
		strings.IndexByte(text, 0xE3) >= 0 ||
		strings.IndexByte(text, 0xE1) >= 0 ||
		strings.IndexByte(text, '\u000B') >= 0 ||
		strings.IndexByte(text, '\u000C') >= 0
}

func findIrregularWhitespace(text string) []errorInfo {
	if len(text) == 0 || !mayContainIrregularWhitespace(text) {
		return nil
	}

	var errors []errorInfo
	i := 0

	// Skip BOM at position 0 — matches ESLint behavior where BOM is
	// stripped from the source before rules see it.
	if len(text) >= 3 && text[0] == 0xEF && text[1] == 0xBB && text[2] == 0xBF {
		// UTF-8 BOM for U+FEFF
		i = 3
	}

	for i < len(text) {
		if !irregularWhitespaceFirstBytes[text[i]] {
			i++
			continue
		}
		size, lineTerminator := irregularWhitespaceSizeAt(text, i)
		if size == 0 {
			i++
			continue
		}

		// Group consecutive irregular whitespace (non-line-terminator) into one
		// error, matching ESLint's regex. U+2028 and U+2029 are always reported
		// individually.
		if lineTerminator {
			errors = append(errors, errorInfo{pos: i, end: i + size})
			i += size
			continue
		}

		start := i
		end := i + size
		for end < len(text) {
			nextSize, nextLineTerminator := irregularWhitespaceSizeAt(text, end)
			if nextSize == 0 || nextLineTerminator {
				break
			}
			end += nextSize
		}
		errors = append(errors, errorInfo{pos: start, end: end})
		i = end
	}
	return errors
}

func checkNoIrregularWhitespace(ctx rule.RuleContext, opts irregularWhitespaceOptions) {
	errors := findIrregularWhitespace(ctx.SourceFile.Text())
	if len(errors) == 0 {
		return
	}

	// Collect exempt ranges based on options.
	var exemptRanges []errorInfo
	collectExemptRanges(ctx, opts, errors, &exemptRanges)

	// Sort exempt ranges by position for efficient filtering.
	if len(exemptRanges) > 1 {
		sort.Slice(exemptRanges, func(i, j int) bool {
			return exemptRanges[i].pos < exemptRanges[j].pos
		})
	}

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

func rangeContainsError(start int, end int, errors []errorInfo) bool {
	idx := sort.Search(len(errors), func(i int) bool {
		return errors[i].end > start
	})
	return idx < len(errors) && errors[idx].pos < end
}

func collectExemptRanges(ctx rule.RuleContext, opts irregularWhitespaceOptions, errors []errorInfo, ranges *[]errorInfo) {
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
			// Errors are sorted and normally sparse. Prune every subtree that
			// cannot contain one instead of walking and allocating ranges for
			// every string/template/regexp in the file.
			start, end := node.Pos(), node.End()
			if start >= 0 && end >= start && !rangeContainsError(start, end, errors) {
				return false
			}

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
		for _, comment := range ctx.Comments.All() {
			if rangeContainsError(comment.Pos(), comment.End(), errors) {
				*ranges = append(*ranges, errorInfo{pos: comment.Pos(), end: comment.End()})
			}
		}
	}
}
