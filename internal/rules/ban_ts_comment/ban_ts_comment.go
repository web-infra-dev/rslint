package ban_ts_comment

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

const defaultMinimumDescriptionLength = 3

type DirectiveConfig interface{}

type OptionsShape struct {
	MinimumDescriptionLength *int                       `json:"minimumDescriptionLength"`
	TsCheck                  DirectiveConfig            `json:"ts-check"`
	TsExpectError            DirectiveConfig            `json:"ts-expect-error"`
	TsIgnore                 DirectiveConfig            `json:"ts-ignore"`
	TsNocheck                DirectiveConfig            `json:"ts-nocheck"`
}

type MatchedTSDirective struct {
	Description string
	Directive   string
}

func buildReplaceTsIgnoreWithTsExpectErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceTsIgnoreWithTsExpectError",
		Description: "Replace \"@ts-ignore\" with \"@ts-expect-error\".",
	}
}

func buildTsDirectiveCommentMessage(directive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveComment",
		Description: fmt.Sprintf("Do not use \"@ts-%s\" because it alters compilation errors.", directive),
	}
}

func buildTsDirectiveCommentDescriptionNotMatchPatternMessage(directive, format string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
		Description: fmt.Sprintf("The description for the \"@ts-%s\" directive must match the %s format.", directive, format),
	}
}

func buildTsDirectiveCommentRequiresDescriptionMessage(directive string, minimumDescriptionLength int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentRequiresDescription",
		Description: fmt.Sprintf("Include a description after the \"@ts-%s\" directive to explain why the @ts-%s is necessary. The description must be %d characters or longer.", directive, directive, minimumDescriptionLength),
	}
}

func buildTsIgnoreInsteadOfExpectErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsIgnoreInsteadOfExpectError",
		Description: "Use \"@ts-expect-error\" instead of \"@ts-ignore\", as \"@ts-ignore\" will do nothing if the following line is error-free.",
	}
}

func parseDirectiveConfig(config DirectiveConfig) (bool, string, string) {
	if config == nil {
		return false, "", ""
	}

	switch v := config.(type) {
	case bool:
		return v, "", ""
	case string:
		if v == "allow-with-description" {
			return true, "allow-with-description", ""
		}
		return true, "", ""
	case map[string]interface{}:
		if descFormat, ok := v["descriptionFormat"].(string); ok {
			return true, "description-format", descFormat
		}
		return true, "", ""
	}
	return false, "", ""
}

func getStringLength(s string) int {
	// Count meaningful characters for description length
	// Include basic characters and some emoji/unicode but be more restrictive
	count := 0
	runes := []rune(s)
	
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		
		// Count ASCII letters, numbers, and meaningful punctuation
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == ' ' || r == '.' || r == ',' || r == ':' || r == ';' || r == '!' || r == '?' ||
			r == '-' || r == '_' || r == '(' || r == ')' || r == '[' || r == ']' ||
			r == '{' || r == '}' || r == '\'' || r == '"' || r == '`' || r == '/' || r == '\\' ||
			r == '+' || r == '=' || r == '@' || r == '#' || r == '$' || r == '%' || r == '&' ||
			r == '*' || r == '<' || r == '>' || r == '|' {
			count++
		} else if r >= 0x1F000 {
			// Count emoji and other high Unicode characters, but handle multi-codepoint sequences
			// Family emoji like 👨‍👩‍👧‍👦 are composed of multiple codepoints joined by ZWJ (U+200D)
			count++
			// Skip over zero-width joiners and variation selectors that are part of emoji sequences
			for i+1 < len(runes) && (runes[i+1] == 0x200D || (runes[i+1] >= 0xFE00 && runes[i+1] <= 0xFE0F)) {
				i++
				if i+1 < len(runes) && runes[i+1] >= 0x1F000 {
					i++ // Skip the next emoji component
				}
			}
		}
	}
	return count
}

func execDirectiveRegEx(regex *regexp.Regexp, str string) *MatchedTSDirective {
	match := regex.FindStringSubmatch(str)
	if match == nil {
		return nil
	}

	subexpNames := regex.SubexpNames()
	var directive, description string
	for i, name := range subexpNames {
		if i > 0 && i < len(match) {
			switch name {
			case "directive":
				directive = match[i]
			case "description":
				description = match[i]
			}
		}
	}

	if directive == "" {
		return nil
	}

	return &MatchedTSDirective{
		Description: description,
		Directive:   directive,
	}
}

var (
	// Compile regex patterns once at package level
	singleLinePragmaRegEx              = regexp.MustCompile(`^\/\/+\s*@ts-(?P<directive>check|nocheck)(?P<description>[\s\S]*)$`)
	commentDirectiveRegExSingleLine    = regexp.MustCompile(`^\s*@ts-(?P<directive>expect-error|ignore)(?P<description>[\s\S]*)`)
	commentDirectiveRegExMultiLine     = regexp.MustCompile(`^\s*(?:\/|\*)*\s*@ts-(?P<directive>expect-error|ignore)(?P<description>[\s\S]*)`)
	singleLinePragmaRegExForMultiLine  = regexp.MustCompile(`^\s*(?:\/|\*)*\s*@ts-(?P<directive>check|nocheck)(?P<description>[\s\S]*)`)
)

func findDirectiveInComment(commentRange ast.CommentRange, sourceText string) *MatchedTSDirective {
	// Ensure positions are within bounds
	startPos := commentRange.Pos()
	endPos := commentRange.End()
	
	if startPos < 0 || startPos >= len(sourceText) || endPos <= startPos || endPos > len(sourceText) {
		return nil
	}
	
	commentText := sourceText[startPos:endPos]
	
	if commentRange.Kind == ast.KindSingleLineCommentTrivia {
		// For single line comments, strip the leading slashes and check for directives
		if strings.HasPrefix(commentText, "//") {
			// Remove all leading slashes
			commentValue := commentText
			for strings.HasPrefix(commentValue, "/") {
				commentValue = commentValue[1:]
			}
			
			// First check for pragma comments (check/nocheck)
			// Pragma comments should only have 2-3 slashes, not more
			originalSlashCount := len(commentText) - len(strings.TrimLeft(commentText, "/"))
			if originalSlashCount <= 3 {
				if matchedPragma := execDirectiveRegEx(singleLinePragmaRegExForMultiLine, commentValue); matchedPragma != nil {
					return matchedPragma
				}
			}
			
			// Then check for directive comments (expect-error/ignore)
			return execDirectiveRegEx(commentDirectiveRegExSingleLine, commentValue)
		}
		// If no "//" prefix, try matching the whole text
		return execDirectiveRegEx(commentDirectiveRegExSingleLine, commentText)
	}

	// Multi-line comments
	commentValue := commentText
	
	// Check if the comment text includes the delimiters
	if strings.HasPrefix(commentText, "/*") && strings.HasSuffix(commentText, "*/") {
		// Remove "/*" and "*/" if present
		if len(commentText) >= 4 {
			commentValue = commentText[2 : len(commentText)-2]
		}
	}
	
	// For single-line block comments like /* @ts-check */, check the entire content
	if !strings.Contains(commentValue, "\n") {
		// Single line block comment - check the entire content
		trimmedValue := strings.TrimSpace(commentValue)
		
		// Check for pragma comments first
		if matchedPragma := execDirectiveRegEx(singleLinePragmaRegExForMultiLine, trimmedValue); matchedPragma != nil {
			return matchedPragma
		}
		
		// Check for directive comments
		return execDirectiveRegEx(commentDirectiveRegExMultiLine, trimmedValue)
	}
	
	// Multi-line block comments - check only the last line
	commentLines := strings.Split(commentValue, "\n")
	if len(commentLines) == 0 {
		return nil
	}
	
	// Check only the last line for directives
	lastLine := commentLines[len(commentLines)-1]
	trimmedLine := strings.TrimSpace(lastLine)
	
	// Check for pragma comments
	if matchedPragma := execDirectiveRegEx(singleLinePragmaRegExForMultiLine, trimmedLine); matchedPragma != nil {
		return matchedPragma
	}
	
	// Check for directive comments
	if matched := execDirectiveRegEx(commentDirectiveRegExMultiLine, trimmedLine); matched != nil {
		return matched
	}
	
	return nil
}

var BanTsCommentRule = rule.Rule{
	Name: "ban-ts-comment",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := OptionsShape{
			MinimumDescriptionLength: utils.Ref(defaultMinimumDescriptionLength),
			TsCheck:                  false,
			TsExpectError:            "allow-with-description",
			TsIgnore:                 true,
			TsNocheck:                true,
		}

		if options != nil {
			if optMap, ok := options.(map[string]interface{}); ok {
				if minLen, exists := optMap["minimumDescriptionLength"]; exists {
					if minLenInt, ok := minLen.(float64); ok {
						opts.MinimumDescriptionLength = utils.Ref(int(minLenInt))
					} else if minLenInt, ok := minLen.(int); ok {
						opts.MinimumDescriptionLength = utils.Ref(minLenInt)
					}
				}
				if tsCheck, exists := optMap["ts-check"]; exists {
					opts.TsCheck = tsCheck
				}
				if tsExpectError, exists := optMap["ts-expect-error"]; exists {
					opts.TsExpectError = tsExpectError
				}
				if tsIgnore, exists := optMap["ts-ignore"]; exists {
					opts.TsIgnore = tsIgnore
				}
				if tsNocheck, exists := optMap["ts-nocheck"]; exists {
					opts.TsNocheck = tsNocheck
				}
			}
		}

		descriptionFormats := make(map[string]*regexp.Regexp)

		directiveConfigs := map[string]DirectiveConfig{
			"ts-expect-error": opts.TsExpectError,
			"ts-ignore":       opts.TsIgnore,
			"ts-nocheck":      opts.TsNocheck,
			"ts-check":        opts.TsCheck,
		}

		for directive, config := range directiveConfigs {
			_, mode, format := parseDirectiveConfig(config)
			if mode == "description-format" && format != "" {
				if regex, err := regexp.Compile(format); err == nil {
					descriptionFormats[directive] = regex
				}
			}
		}

		// Process a comment and report issues if needed
		processComment := func(commentRange ast.CommentRange, sourceFile *ast.SourceFile, sourceText string, firstStatement *ast.Node) {
			directive := findDirectiveInComment(commentRange, sourceText)
			if directive == nil {
				return
			}

			// Special handling for ts-nocheck
			if directive.Directive == "nocheck" {
				// Get the comment text to check if it's a block comment
				commentText := sourceText[commentRange.Pos():commentRange.End()]
				isBlockComment := strings.HasPrefix(commentText, "/*") || strings.HasPrefix(commentText, "/**")
				
				if isBlockComment {
					// Block comments with ts-nocheck are always allowed
					return
				}
				
				// For line comments, allow if they appear before the first statement
				if firstStatement == nil {
					// No statements in file, allow ts-nocheck
					return
				}
				// Allow ts-nocheck at the top of the file (before any statements)
				if commentRange.Pos() < firstStatement.Pos() {
					return
				}
			}

			directiveName := "ts-" + directive.Directive
			config, exists := directiveConfigs[directiveName]
			if !exists {
				return
			}

			enabled, mode, descFormat := parseDirectiveConfig(config)

			// Handle when directive is disabled (false)
			if !enabled {
				return
			}

			// Special handling for ts-check directive
			if directive.Directive == "check" {
				// For ts-check, when enabled=true and mode="", allow block comments but ban line comments
				if enabled && mode == "" {
					// Check if this is a block comment by looking at the comment text
					commentText := sourceText[commentRange.Pos():commentRange.End()]
					isBlockComment := strings.HasPrefix(commentText, "/*") || strings.HasPrefix(commentText, "/**")
					
					// For ts-check, only report error for line comments when enabled=true
					if !isBlockComment {
						ctx.ReportRange(commentRange.TextRange, buildTsDirectiveCommentMessage(directive.Directive))
					}
					return
				}
			} else {
				// Handle other directives when enabled (true) - should report error
				if enabled && mode == "" {
					// Special handling for ts-ignore
					if directive.Directive == "ignore" {
						// Get the comment text for the fix
						commentStart := commentRange.Pos()
						commentEnd := commentRange.End()
						commentText := sourceText[commentStart:commentEnd]


						// Create replacement text
						replacementText := strings.Replace(commentText, "@ts-ignore", "@ts-expect-error", 1)

						// Create suggestion to replace ts-ignore with ts-expect-error
						suggestion := rule.RuleSuggestion{
							Message: buildReplaceTsIgnoreWithTsExpectErrorMessage(),
							FixesArr: []rule.RuleFix{
								{
									Text: replacementText,
									Range: commentRange.TextRange,
								},
							},
						}

						ctx.ReportRangeWithSuggestions(commentRange.TextRange, buildTsIgnoreInsteadOfExpectErrorMessage(), suggestion)
					} else {
						ctx.ReportRange(commentRange.TextRange, buildTsDirectiveCommentMessage(directive.Directive))
					}
					return
				}
			}

			// Handle allow-with-description mode
			if mode == "allow-with-description" || mode == "description-format" {
				trimmedDescription := strings.TrimSpace(directive.Description)
				descriptionLength := getStringLength(trimmedDescription)

				// Check minimum description length
				if descriptionLength < *opts.MinimumDescriptionLength {
					ctx.ReportRange(commentRange.TextRange, buildTsDirectiveCommentRequiresDescriptionMessage(directive.Directive, *opts.MinimumDescriptionLength))
					return
				}

				// Check description format if specified
				if mode == "description-format" && descFormat != "" {
					regex := descriptionFormats[directiveName]
					if regex != nil && !regex.MatchString(directive.Description) {
						ctx.ReportRange(commentRange.TextRange, buildTsDirectiveCommentDescriptionNotMatchPatternMessage(directive.Directive, descFormat))
					}
				}
			}
		}

		// Process all comments directly in the Run function
		sourceFile := ctx.SourceFile
		sourceText := string(sourceFile.Text())

		// Get the first statement in the file for ts-nocheck handling
		var firstStatement *ast.Node
		if sourceFile.Statements != nil && len(sourceFile.Statements.Nodes) > 0 {
			firstStatement = sourceFile.Statements.Nodes[0]
		}

		if len(sourceText) > 0 {
			// Use GetCommentsInRange to get all comments in the entire file
			fileRange := core.NewTextRange(0, len(sourceText))
			processedComments := make(map[string]bool) // Use string key to avoid duplicates
			
			for commentRange := range utils.GetCommentsInRange(sourceFile, fileRange) {
				// Create a unique key based on position and end to avoid processing the same comment twice
				commentKey := fmt.Sprintf("%d-%d", commentRange.Pos(), commentRange.End())
				if !processedComments[commentKey] {
					processedComments[commentKey] = true
					processComment(commentRange, sourceFile, sourceText, firstStatement)
				}
			}
			
			// Check for comments in if (false) blocks if needed, but use a more efficient approach
			if strings.Contains(sourceText, "if (false)") && strings.Contains(sourceText, "@ts-") {
				// Use a simple regex-based approach instead of manual parsing
				// This is much faster than the previous character-by-character approach
				// Look for if (false) { ... } blocks containing @ts- comments
				ifFalsePattern := regexp.MustCompile(`if\s*\(\s*false\s*\)\s*\{[^}]*@ts-[^}]*\}`)
				matches := ifFalsePattern.FindAllString(sourceText, -1)
				for _, match := range matches {
					// Find the position of this match in the source
					matchPos := strings.Index(sourceText, match)
					if matchPos != -1 {
						// Process comments in this block using GetCommentsInRange
						blockRange := core.NewTextRange(matchPos, matchPos+len(match))
						for commentRange := range utils.GetCommentsInRange(sourceFile, blockRange) {
							commentKey := fmt.Sprintf("%d-%d", commentRange.Pos(), commentRange.End())
							if !processedComments[commentKey] {
								processedComments[commentKey] = true
								processComment(commentRange, sourceFile, sourceText, firstStatement)
							}
						}
					}
				}
			}
		}

		// Return empty listeners since we've already processed everything
		return rule.RuleListeners{}
	},
}
