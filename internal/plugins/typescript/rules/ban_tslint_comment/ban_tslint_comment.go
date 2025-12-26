package ban_tslint_comment

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Regular expression to match TSLint directive comments
// Matches patterns like:
// - tslint:disable
// - tslint:enable
// - tslint:disable-line
// - tslint:disable-next-line
// - tslint:enable-line
var tslintCommentRegex = regexp.MustCompile(`^\s*tslint:(enable|disable)(?:-(line|next-line))?(?::|s|$)`)

// BanTslintCommentRule implements the ban-tslint-comment rule
// Bans // tslint:<rule flag> comments
var BanTslintCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-tslint-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process the text to find TSLint comments
	processComments(ctx, text)

	return rule.RuleListeners{}
}

// processComments scans the source text for comments and checks for TSLint directives
func processComments(ctx rule.RuleContext, text string) {
	pos := 0
	length := len(text)

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
				checkComment(ctx, commentText, commentStart, false)
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
				checkComment(ctx, commentText, commentStart, true)
				pos = commentEnd
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}

// checkComment checks a single comment for TSLint directives
func checkComment(ctx rule.RuleContext, commentText string, commentStart int, isMultiLine bool) {
	var contentToCheck string

	if isMultiLine {
		// For multi-line comments, remove /* and */ and check the content
		contentToCheck = commentText
		if strings.HasPrefix(contentToCheck, "/*") {
			contentToCheck = contentToCheck[2:]
		}
		if strings.HasSuffix(contentToCheck, "*/") {
			contentToCheck = contentToCheck[:len(contentToCheck)-2]
		}
	} else {
		// For single-line comments, remove // and check the content
		contentToCheck = commentText
		if strings.HasPrefix(contentToCheck, "//") {
			contentToCheck = contentToCheck[2:]
		}
	}

	// Check if the content matches TSLint directive pattern
	if tslintCommentRegex.MatchString(contentToCheck) {
		// Report the TSLint directive
		ctx.ReportRange(
			core.NewTextRange(commentStart, commentStart+len(commentText)),
			rule.RuleMessage{
				Id:          "commentDetected",
				Description: "tslint comment detected: \"" + strings.TrimSpace(commentText) + "\"",
			},
		)
	}
}
