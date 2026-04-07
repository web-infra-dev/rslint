package ban_ts_comment

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type DirectiveConfig struct {
	Enabled              bool   // Whether the directive is enabled (true means banned)
	AllowWithDescription bool   // Whether to allow with description
	DescriptionFormat    string // Regex pattern for description format
}

type BanTsCommentOptions struct {
	TsExpectError            interface{} `json:"ts-expect-error"`
	TsIgnore                 interface{} `json:"ts-ignore"`
	TsNocheck                interface{} `json:"ts-nocheck"`
	TsCheck                  interface{} `json:"ts-check"`
	MinimumDescriptionLength int         `json:"minimumDescriptionLength"`
}

// Regular expressions for matching TypeScript directives
var (
	// Matches single-line comments: // @ts-<directive>
	singleLineDirectiveRegex = regexp.MustCompile(`^\/\/\/?\s*@ts-(expect-error|ignore|nocheck|check)\b`)

	// Matches multi-line comments: /* @ts-<directive> */
	// Matches if the directive appears anywhere in the comment (including after newlines)
	// Uses [\s*]* to match whitespace (including newlines) and asterisks
	multiLineDirectiveRegex = regexp.MustCompile(`^\/\*[\s*]*@ts-(expect-error|ignore|nocheck|check)\b`)
)

// BanTsCommentRule implements the ban-ts-comment rule
// Bans @ts-<directive> comments or requires descriptions after directive
var BanTsCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-ts-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := BanTsCommentOptions{
		TsExpectError:            true,
		TsIgnore:                 true,
		TsNocheck:                true,
		TsCheck:                  false,
		MinimumDescriptionLength: 3,
	}

	// Parse options
	if options != nil {
		var optsMap map[string]interface{}
		var ok bool

		// Handle array format: [{ option: value }]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			optsMap, ok = optArray[0].(map[string]interface{})
		} else {
			// Handle direct object format: { option: value }
			optsMap, ok = options.(map[string]interface{})
		}

		if ok {
			if val, exists := optsMap["ts-expect-error"]; exists {
				opts.TsExpectError = val
			}
			if val, exists := optsMap["ts-ignore"]; exists {
				opts.TsIgnore = val
			}
			if val, exists := optsMap["ts-nocheck"]; exists {
				opts.TsNocheck = val
			}
			if val, exists := optsMap["ts-check"]; exists {
				opts.TsCheck = val
			}
			if val, ok := optsMap["minimumDescriptionLength"].(float64); ok {
				opts.MinimumDescriptionLength = int(val)
			} else if val, ok := optsMap["minimumDescriptionLength"].(int); ok {
				opts.MinimumDescriptionLength = val
			}
		}
	}

	// Parse directive configurations
	configs := map[string]*DirectiveConfig{
		"ts-expect-error": parseDirectiveConfig(opts.TsExpectError),
		"ts-ignore":       parseDirectiveConfig(opts.TsIgnore),
		"ts-nocheck":      parseDirectiveConfig(opts.TsNocheck),
		"ts-check":        parseDirectiveConfig(opts.TsCheck),
	}

	processComments(ctx, ctx.SourceFile.Text(), configs, opts.MinimumDescriptionLength)

	return rule.RuleListeners{}
}

// parseDirectiveConfig converts the option value to DirectiveConfig
func parseDirectiveConfig(value interface{}) *DirectiveConfig {
	config := &DirectiveConfig{}

	switch v := value.(type) {
	case bool:
		config.Enabled = v
		config.AllowWithDescription = false
	case string:
		if v == "allow-with-description" {
			config.Enabled = true
			config.AllowWithDescription = true
		}
	case map[string]interface{}:
		if descFormat, ok := v["descriptionFormat"].(string); ok {
			config.Enabled = true
			config.AllowWithDescription = true
			config.DescriptionFormat = descFormat
		}
	}

	return config
}

// processComments scans real comment trivia and checks for banned directives.
// Using the TS scanner avoids matching comment-like text inside strings/template literals,
// for example `const c = "// @ts-ignore"` should stay a plain string and must not be linted
// as if it were an actual comment directive.
func processComments(ctx rule.RuleContext, text string, configs map[string]*DirectiveConfig, minDescLength int) {
	utils.ForEachComment(ctx.SourceFile.AsNode(), func(comment *ast.CommentRange) {
		if comment == nil {
			return
		}
		if comment.Pos() < 0 || comment.End() > len(text) || comment.Pos() >= comment.End() {
			return
		}

		commentText := text[comment.Pos():comment.End()]
		checkComment(
			ctx,
			commentText,
			comment.Pos(),
			configs,
			minDescLength,
			comment.Kind == ast.KindMultiLineCommentTrivia,
		)
	}, ctx.SourceFile)
}

// checkComment checks a single comment for banned directives
func checkComment(ctx rule.RuleContext, commentText string, commentStart int, configs map[string]*DirectiveConfig, minDescLength int, isMultiLine bool) {
	var matches []string
	var directiveType string

	if isMultiLine {
		match := multiLineDirectiveRegex.FindStringSubmatch(commentText)
		if match != nil {
			matches = match
			directiveType = match[1]
		}
	} else {
		match := singleLineDirectiveRegex.FindStringSubmatch(commentText)
		if match != nil {
			matches = match
			directiveType = match[1]
		}
	}

	if matches == nil {
		return
	}

	// Get the directive config
	directiveName := "ts-" + directiveType
	config := configs[directiveName]

	// If directive is disabled, don't report
	if !config.Enabled {
		return
	}

	// Extract the part after the directive
	directivePattern := `@ts-` + directiveType
	idx := strings.Index(commentText, directivePattern)
	if idx == -1 {
		return
	}

	afterDirective := commentText[idx+len(directivePattern):]

	// For multi-line comments, check if there's meaningful content after the directive on subsequent lines
	// If there is, this is not a directive comment (it's just a comment that mentions the directive)
	if isMultiLine {
		// Remove the trailing */
		withoutClosing := strings.TrimSuffix(afterDirective, "*/")

		// Find the first newline after the directive
		firstNewline := strings.Index(withoutClosing, "\n")
		if firstNewline != -1 {
			// Get content after the first newline
			afterFirstLine := withoutClosing[firstNewline+1:]

			// Check if there's any meaningful content after the directive line
			// (excluding whitespace, asterisks, and description separators)
			lines := strings.Split(afterFirstLine, "\n")
			for _, line := range lines {
				trimmed := strings.TrimLeft(line, " \t*")
				trimmed = strings.TrimSpace(trimmed)
				// If this line has content that's not just separators, it's not a directive comment
				if len(trimmed) > 0 && !isOnlyDescriptionSeparator(trimmed) {
					return
				}
			}
		}

		afterDirective = withoutClosing
	}

	// Description is everything after the directive, trimmed of whitespace only.
	// Separators like ':' and '--' are part of the description (included in length).
	description := strings.TrimSpace(afterDirective)

	// Special case: for ts-ignore with no description allowed, suggest ts-expect-error
	if directiveType == "ignore" && !config.AllowWithDescription {
		ctx.ReportRange(
			core.NewTextRange(commentStart, commentStart+len(commentText)),
			rule.RuleMessage{
				Id:          "tsIgnoreInsteadOfExpectError",
				Description: "Prefer '@ts-expect-error' over '@ts-ignore' as it requires error to be present in next line.",
			},
		)
		return
	}

	// If the directive is completely banned (no description allowed)
	if !config.AllowWithDescription {
		ctx.ReportRange(
			core.NewTextRange(commentStart, commentStart+len(commentText)),
			rule.RuleMessage{
				Id:          "tsDirectiveComment",
				Description: "Do not use '@" + directiveName + "' because it alters compilation errors.",
			},
		)
		return
	}

	// If description is required, check minimum length (handles both empty and too-short)
	if config.AllowWithDescription {
		descLength := graphemeLength(description)
		if descLength < minDescLength {
			ctx.ReportRange(
				core.NewTextRange(commentStart, commentStart+len(commentText)),
				rule.RuleMessage{
					Id:          "tsDirectiveCommentRequiresDescription",
					Description: "Include a description after the '@" + directiveName + "' directive to explain why the '@" + directiveName + "' is necessary. The description must be " + formatMinimumDescLength(minDescLength) + " characters long.",
				},
			)
			return
		}

		// Check description format if specified
		if config.DescriptionFormat != "" {
			formatRegex, err := regexp.Compile(config.DescriptionFormat)
			if err == nil {
				if !formatRegex.MatchString(description) {
					ctx.ReportRange(
						core.NewTextRange(commentStart, commentStart+len(commentText)),
						rule.RuleMessage{
							Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
							Description: "The description for the '@" + directiveName + "' directive must match the format '" + config.DescriptionFormat + "'.",
						},
					)
					return
				}
			}
		}
	}
}

// isOnlyDescriptionSeparator checks if a string contains only description separator characters
func isOnlyDescriptionSeparator(s string) bool {
	for _, ch := range s {
		if ch != ':' && ch != ' ' && ch != '\t' && ch != '-' {
			return false
		}
	}
	return len(s) > 0
}

// graphemeLength returns the number of grapheme clusters in a string
// This properly handles Unicode characters including emojis
func graphemeLength(s string) int {
	// For simplicity, we'll count runes, which is close enough for most cases
	// A more accurate implementation would use a grapheme cluster library
	return utf8.RuneCountInString(s)
}

// formatMinimumDescLength formats the minimum description length message
func formatMinimumDescLength(length int) string {
	if length == 1 {
		return "at least 1 character"
	}
	return "at least " + formatInt(length) + " characters"
}

// formatInt converts an integer to a string
func formatInt(n int) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return formatInt(n/10) + string(rune('0'+n%10))
}
