package ban_ts_comment

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
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

func findDirectiveInComment(commentRange ast.CommentRange, sourceText string) *MatchedTSDirective {
	commentText := sourceText[commentRange.Pos():commentRange.End()]

	singleLinePragmaRegEx := regexp.MustCompile(`^\/\/\/?\s*@ts-(?P<directive>check|nocheck)(?P<description>.*)$`)
	commentDirectiveRegExSingleLine := regexp.MustCompile(`^\/*\s*@ts-(?P<directive>expect-error|ignore)(?P<description>.*)`)
	commentDirectiveRegExMultiLine := regexp.MustCompile(`^\s*(?:\/|\*)*\s*@ts-(?P<directive>expect-error|ignore)(?P<description>.*)`)

	if commentRange.Kind == ast.KindSingleLineCommentTrivia {
		if matchedPragma := execDirectiveRegEx(singleLinePragmaRegEx, commentText); matchedPragma != nil {
			return matchedPragma
		}

		commentValue := commentText[2:] // Remove "//"
		return execDirectiveRegEx(commentDirectiveRegExSingleLine, commentValue)
	}

	commentValue := commentText[2 : len(commentText)-2] // Remove "/*" and "*/"
	commentLines := strings.Split(commentValue, "\n")
	lastLine := commentLines[len(commentLines)-1]
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

		// Process comments immediately, regardless of AST nodes
		sourceFile := ctx.SourceFile
		sourceText := string(sourceFile.Text())
		
		// Check if we have any directives
		if !strings.Contains(sourceText, "@ts-") {
			return rule.RuleListeners{}
		}
		
		// Process all @ts- directives immediately
		pos := 0
		for {
			tsPos := strings.Index(sourceText[pos:], "@ts-")
			if tsPos == -1 {
				break
			}
			tsPos += pos
			
			// Find the start of the comment
			commentStart := strings.LastIndex(sourceText[:tsPos], "//")
			if commentStart == -1 {
				commentStart = strings.LastIndex(sourceText[:tsPos], "/*")
			}
			if commentStart == -1 {
				pos = tsPos + 4
				continue
			}
			
			// Extract the directive
			directiveEnd := tsPos + 4
			for directiveEnd < len(sourceText) && (sourceText[directiveEnd] >= 'a' && sourceText[directiveEnd] <= 'z' || sourceText[directiveEnd] == '-') {
				directiveEnd++
			}
			
			directive := sourceText[tsPos+4:directiveEnd]
			description := ""
			if directiveEnd < len(sourceText) {
				commentEnd := strings.Index(sourceText[directiveEnd:], "\n")
				if commentEnd == -1 {
					commentEnd = strings.Index(sourceText[directiveEnd:], "*/")
					if commentEnd == -1 {
						commentEnd = len(sourceText) - directiveEnd
					}
				}
				description = sourceText[directiveEnd:directiveEnd+commentEnd]
			}
			
			// Process the directive
			fullDirective := "ts-" + directive
			config := directiveConfigs[fullDirective]
			enabled, mode, _ := parseDirectiveConfig(config)
			
			if enabled && mode == "" {
				// Find comment end for proper range
				commentEnd := strings.Index(sourceText[commentStart:], "\n")
				if commentEnd == -1 {
					commentEnd = strings.Index(sourceText[commentStart:], "*/")
					if commentEnd != -1 {
						commentEnd += 2 // Include */
					} else {
						commentEnd = len(sourceText) - commentStart
					}
				}
				commentEnd += commentStart
				
				// Create range and report
				textRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(commentStart).WithEnd(commentEnd)
				
				if directive == "ignore" {
					ctx.ReportRange(textRange, buildTsIgnoreInsteadOfExpectErrorMessage())
				} else {
					ctx.ReportRange(textRange, buildTsDirectiveCommentMessage(directive))
				}
			}
			
			if enabled && (mode == "allow-with-description" || mode == "description-format") {
				minimumDescriptionLength := defaultMinimumDescriptionLength
				if opts.MinimumDescriptionLength != nil {
					minimumDescriptionLength = *opts.MinimumDescriptionLength
				}

				regex := descriptionFormats[fullDirective]
				trimmedDescription := strings.TrimSpace(description)
				
				if getStringLength(trimmedDescription) < minimumDescriptionLength {
					commentEnd := strings.Index(sourceText[commentStart:], "\n")
					if commentEnd == -1 {
						commentEnd = strings.Index(sourceText[commentStart:], "*/")
						if commentEnd != -1 {
							commentEnd += 2
						} else {
							commentEnd = len(sourceText) - commentStart
						}
					}
					commentEnd += commentStart
					
					textRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(commentStart).WithEnd(commentEnd)
					ctx.ReportRange(textRange, buildTsDirectiveCommentRequiresDescriptionMessage(directive, minimumDescriptionLength))
				} else if regex != nil && !regex.MatchString(description) {
					commentEnd := strings.Index(sourceText[commentStart:], "\n")
					if commentEnd == -1 {
						commentEnd = strings.Index(sourceText[commentStart:], "*/")
						if commentEnd != -1 {
							commentEnd += 2
						} else {
							commentEnd = len(sourceText) - commentStart
						}
					}
					commentEnd += commentStart
					
					textRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(commentStart).WithEnd(commentEnd)
					ctx.ReportRange(textRange, buildTsDirectiveCommentDescriptionNotMatchPatternMessage(directive, regex.String()))
				}
			}
			
			pos = tsPos + 4
		}
		
		// Return empty listeners since we've already processed everything
		return rule.RuleListeners{}
	},
}