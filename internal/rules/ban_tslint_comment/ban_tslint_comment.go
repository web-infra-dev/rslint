package ban_tslint_comment

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// tslint regex
// https://github.com/palantir/tslint/blob/95d9d958833fd9dc0002d18cbe34db20d0fbf437/src/enableDisableRules.ts#L32
var enableDisableRegex = regexp.MustCompile(`^\s*tslint:(enable|disable)(?:-(line|next-line))?(:|\s|$)`)

func toText(text string, isBlockComment bool) string {
	trimmed := strings.TrimSpace(text)
	if isBlockComment {
		return fmt.Sprintf("/* %s */", trimmed)
	}
	return fmt.Sprintf("// %s", trimmed)
}

func buildCommentDetectedMessage(text string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "commentDetected",
		Description: fmt.Sprintf("tslint comment detected: \"%s\"", text),
	}
}

var BanTslintCommentRule = rule.Rule{
	Name: "ban-tslint-comment",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceFile := ctx.SourceFile
		sourceText := string(sourceFile.Text())

		// Track processed positions to avoid duplicates
		processed := make(map[int]bool)

		// Process all tslint comments immediately
		pos := 0
		for {
			tslintPos := strings.Index(sourceText[pos:], "tslint:")
			if tslintPos == -1 {
				break
			}
			tslintPos += pos

			// Find the start of the comment
			commentStart := -1
			isBlockComment := false

			// Look for line comment start
			lineCommentStart := strings.LastIndex(sourceText[:tslintPos], "//")
			if lineCommentStart != -1 {
				// Verify there's no newline between // and tslint:
				if !strings.Contains(sourceText[lineCommentStart:tslintPos], "\n") {
					commentStart = lineCommentStart
				}
			}

			// Look for block comment start
			blockCommentStart := strings.LastIndex(sourceText[:tslintPos], "/*")
			if blockCommentStart != -1 {
				// Verify there's no */ between /* and tslint:
				if !strings.Contains(sourceText[blockCommentStart:tslintPos], "*/") {
					// Use block comment if it's closer than line comment
					if blockCommentStart > commentStart {
						commentStart = blockCommentStart
						isBlockComment = true
					}
				}
			}

			if commentStart == -1 || processed[commentStart] {
				pos = tslintPos + 7
				continue
			}

			// Mark this comment as processed
			processed[commentStart] = true

			// Find the end of the comment
			var commentEnd int
			if isBlockComment {
				endMarker := strings.Index(sourceText[commentStart:], "*/")
				if endMarker == -1 {
					pos = tslintPos + 7
					continue
				}
				commentEnd = commentStart + endMarker + 2
			} else {
				endMarker := strings.Index(sourceText[commentStart:], "\n")
				if endMarker == -1 {
					commentEnd = len(sourceText)
				} else {
					commentEnd = commentStart + endMarker
				}
			}

			// Extract the comment value
			var commentValue string
			if isBlockComment {
				// Remove /* and */ from block comments
				if commentEnd > commentStart+4 {
					commentValue = sourceText[commentStart+2 : commentEnd-2]
				}
			} else {
				// Remove // from line comments
				if commentEnd > commentStart+2 {
					commentValue = sourceText[commentStart+2 : commentEnd]
				}
			}

			// Test if this matches the tslint regex
			if enableDisableRegex.MatchString(commentValue) {
				// Calculate the proper range for removal
				removeStart := commentStart
				removeEnd := commentEnd

				// For line comments, check if we need to include preceding whitespace
				if !isBlockComment {
					// Check if there's code before the comment on the same line
					lineStart := strings.LastIndex(sourceText[:commentStart], "\n")
					if lineStart == -1 {
						lineStart = 0
					} else {
						lineStart++ // Move past the newline
					}

					// Check if there's non-whitespace content before the comment
					lineContent := sourceText[lineStart:commentStart]
					hasCodeBefore := strings.TrimSpace(lineContent) != ""

					if hasCodeBefore {
						// There's code before the comment, so we need to include preceding spaces
						// Find the start of whitespace before the comment
						spaceStart := commentStart
						for spaceStart > lineStart && sourceText[spaceStart-1] == ' ' {
							spaceStart--
						}
						removeStart = spaceStart
					} else {
						// No code before, remove the entire line including newline
						removeStart = lineStart
						if removeEnd < len(sourceText) && sourceText[removeEnd] == '\n' {
							removeEnd++
						}
					}
				}

				// Create text ranges - one for reporting position, one for fixing
				reportRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(commentStart).WithEnd(commentEnd)
				fixRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(removeStart).WithEnd(removeEnd)
				commentText := toText(strings.TrimSpace(commentValue), isBlockComment)

				ctx.ReportRangeWithSuggestions(reportRange, buildCommentDetectedMessage(commentText),
					rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:          "removeTslintComment",
							Description: "Remove the tslint comment",
						},
						FixesArr: []rule.RuleFix{
							rule.RuleFixRemoveRange(fixRange),
						},
					},
				)
			}

			pos = tslintPos + 7
		}

		// Return empty listeners since we've already processed everything
		return rule.RuleListeners{}
	},
}
