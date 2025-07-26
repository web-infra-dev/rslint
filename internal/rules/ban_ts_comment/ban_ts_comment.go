package ban_ts_comment

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

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
	return utf8.RuneCountInString(s)
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
	singleLinePragmaRegEx              = regexp.MustCompile(`^\/\/\/?\s*@ts-(?P<directive>check|nocheck)(?P<description>.*)$`)
	commentDirectiveRegExSingleLine    = regexp.MustCompile(`^\/*\s*@ts-(?P<directive>expect-error|ignore)(?P<description>.*)`)
	commentDirectiveRegExMultiLine     = regexp.MustCompile(`^\s*(?:\/|\*)*\s*@ts-(?P<directive>expect-error|ignore)(?P<description>.*)`)
	singleLinePragmaRegExForMultiLine  = regexp.MustCompile(`^\s*(?:\/|\*)*\s*@ts-(?P<directive>check|nocheck)(?P<description>.*)`)
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
		// First check for pragma comments (check/nocheck)
		if matchedPragma := execDirectiveRegEx(singleLinePragmaRegEx, commentText); matchedPragma != nil {
			return matchedPragma
		}

		// Then check for directive comments (expect-error/ignore)
		// Single line comments should start with "//"
		if len(commentText) >= 2 && strings.HasPrefix(commentText, "//") {
			commentValue := commentText[2:] // Remove "//"
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
	
	// Split into lines and check the last line
	commentLines := strings.Split(commentValue, "\n")
	if len(commentLines) == 0 {
		return nil
	}
	
	lastLine := commentLines[len(commentLines)-1]

	// Check for pragma comments in the last line of multi-line comments
	if matchedPragma := execDirectiveRegEx(singleLinePragmaRegExForMultiLine, lastLine); matchedPragma != nil {
		return matchedPragma
	}

	return execDirectiveRegEx(commentDirectiveRegExMultiLine, lastLine)
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

			// Special handling for ts-nocheck: skip if it appears before the first statement
			if directive.Directive == "nocheck" && firstStatement != nil {
				if commentRange.End() <= firstStatement.Pos() {
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

			// Handle when directive is enabled (true) - should report error
			if enabled && mode == "" {
				// Special handling for ts-ignore
				if directive.Directive == "ignore" {
					// Get the comment text for the fix
					commentText := sourceText[commentRange.Pos():commentRange.End()]

					// Create suggestion to replace ts-ignore with ts-expect-error
					suggestion := rule.RuleSuggestion{
						Message: buildReplaceTsIgnoreWithTsExpectErrorMessage(),
						FixesArr: []rule.RuleFix{
							{
								Text: strings.Replace(commentText, "@ts-ignore", "@ts-expect-error", 1),
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

			// Handle allow-with-description mode
			if mode == "allow-with-description" || mode == "description-format" {
				descriptionLength := getStringLength(strings.TrimSpace(directive.Description))

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

		return rule.RuleListeners{
			ast.KindSourceFile: func(node *ast.Node) {
				sourceFile := node.AsSourceFile()
				sourceText := string(sourceFile.Text())

				// Get the first statement in the file for ts-nocheck handling
				var firstStatement *ast.Node
				if sourceFile.Statements != nil && len(sourceFile.Statements.Nodes) > 0 {
					firstStatement = sourceFile.Statements.Nodes[0]
				}

				if len(sourceText) == 0 {
					return
				}

				// Use GetCommentsInRange to get all comments in the entire file
				// This is more reliable than trying to scan from specific positions
				fileRange := core.NewTextRange(0, len(sourceText))
				
				for commentRange := range utils.GetCommentsInRange(sourceFile, fileRange) {
					processComment(commentRange, sourceFile, sourceText, firstStatement)
				}
			},
		}
	},
}
