package ban_tslint_comment

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Regular expressions for matching TSLint directives
var (
	// Matches single-line comments: // tslint:disable or // tslint:enable
	singleLineTslintRegex = regexp.MustCompile(`^//\s*tslint:(disable|enable)`)

	// Matches multi-line comments: /* tslint:disable */ or /* tslint:enable */
	multiLineTslintRegex = regexp.MustCompile(`^/\*\s*tslint:(disable|enable)`)
)

// BanTslintCommentRule implements the ban-tslint-comment rule
// Bans // tslint:<rule-flag> comments
var BanTslintCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-tslint-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process the text to find tslint comments
	processComments(ctx, text)

	return rule.RuleListeners{}
}

// processComments scans the source text for tslint comments
func processComments(ctx rule.RuleContext, text string) {
	pos := 0
	length := len(text)
	lineStarts := calculateLineStarts(text)

	for pos < length {
		// Skip to next potential comment
		if pos+1 < length {
			if text[pos] == '/' && text[pos+1] == '/' {
				// Single-line comment
				commentStart := pos
				pos += 2
				lineEnd := pos
				for lineEnd < length && text[lineEnd] != '\n' && text[lineEnd] != '\r' {
					lineEnd++
				}
				commentText := text[commentStart:lineEnd]

				// Check if this is a tslint comment
				if singleLineTslintRegex.MatchString(commentText) {
					reportTslintComment(ctx, commentText, commentStart, lineEnd, lineStarts, text)
				}

				pos = lineEnd
			} else if text[pos] == '/' && text[pos+1] == '*' {
				// Multi-line comment
				commentStart := pos
				pos += 2
				commentEnd := pos
				for commentEnd+1 < length {
					if text[commentEnd] == '*' && text[commentEnd+1] == '/' {
						commentEnd += 2
						break
					}
					commentEnd++
				}
				commentText := text[commentStart:commentEnd]

				// Check if this is a tslint comment
				if multiLineTslintRegex.MatchString(commentText) {
					reportTslintComment(ctx, commentText, commentStart, commentEnd, lineStarts, text)
				}

				pos = commentEnd
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}

// calculateLineStarts returns the starting positions of each line
func calculateLineStarts(text string) []int {
	lineStarts := []int{0}
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}
	return lineStarts
}

// getLineAndColumn returns the line and column numbers for a given position
func getLineAndColumn(pos int, lineStarts []int) (line, column int) {
	for i := len(lineStarts) - 1; i >= 0; i-- {
		if pos >= lineStarts[i] {
			line = i + 1
			column = pos - lineStarts[i] + 1
			return
		}
	}
	return 1, 1
}

// reportTslintComment reports a tslint comment with autofix
func reportTslintComment(ctx rule.RuleContext, commentText string, start, end int, lineStarts []int, fullText string) {
	line, column := getLineAndColumn(start, lineStarts)

	// Create the fix
	fix := createFix(start, end, fullText)

	ctx.ReportRangeWithFixes(
		core.NewTextRange(start, end),
		rule.RuleMessage{
			Id:          "commentDetected",
			Description: "tslint is deprecated and you should stop using it",
		},
		*fix,
	)

	_ = line
	_ = column
}

// createFix creates a fix that removes the tslint comment
func createFix(start, end int, fullText string) *rule.RuleFix {
	// Check if we need to remove the entire line or just the comment

	// Look backwards to see if there's any non-whitespace before the comment
	hasContentBefore := false
	lineStart := start
	for lineStart > 0 && fullText[lineStart-1] != '\n' && fullText[lineStart-1] != '\r' {
		lineStart--
		if !isWhitespace(fullText[lineStart]) {
			hasContentBefore = true
		}
	}

	// Look forwards to see if there's any non-whitespace after the comment (on the same line)
	hasContentAfter := false
	lineEnd := end
	for lineEnd < len(fullText) && fullText[lineEnd] != '\n' && fullText[lineEnd] != '\r' {
		if !isWhitespace(fullText[lineEnd]) {
			hasContentAfter = true
			break
		}
		lineEnd++
	}

	// Skip the newline characters if removing the entire line
	if !hasContentBefore && !hasContentAfter {
		// Include the newline in the removal
		if lineEnd < len(fullText) && fullText[lineEnd] == '\r' {
			lineEnd++
		}
		if lineEnd < len(fullText) && fullText[lineEnd] == '\n' {
			lineEnd++
		}

		return &rule.RuleFix{
			Range: core.NewTextRange(lineStart, lineEnd),
			Text:  "",
		}
	}

	// If there's content before the comment (e.g., "someCode(); // tslint:disable-line")
	if hasContentBefore {
		// Remove just the comment, preserving whitespace before it but removing the comment
		// Find where the actual code ends
		codeEnd := start
		for codeEnd > lineStart && isWhitespace(fullText[codeEnd-1]) {
			codeEnd--
		}

		return &rule.RuleFix{
			Range: core.NewTextRange(codeEnd, end),
			Text:  "",
		}
	}

	// Otherwise, just remove the comment
	return &rule.RuleFix{
		Range: core.NewTextRange(start, end),
		Text:  "",
	}
}

// isWhitespace checks if a character is whitespace
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

// trimTrailingWhitespace removes trailing whitespace and newlines
func trimTrailingWhitespace(s string) string {
	return strings.TrimRight(s, " \t\r\n")
}
