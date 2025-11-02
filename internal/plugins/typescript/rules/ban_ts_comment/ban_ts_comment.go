package ban_ts_comment

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type DirectiveConfig struct {
	Enabled              bool   // Whether the directive is enabled (true means banned)
	AllowWithDescription bool   // Whether to allow with description
	DescriptionFormat    string // Regex pattern for description format
}

type BanTsCommentOptions struct {
	TsExpectError          interface{} `json:"ts-expect-error"`
	TsIgnore               interface{} `json:"ts-ignore"`
	TsNocheck              interface{} `json:"ts-nocheck"`
	TsCheck                interface{} `json:"ts-check"`
	MinimumDescriptionLength int        `json:"minimumDescriptionLength"`
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
		TsExpectError:          true,
		TsIgnore:               true,
		TsNocheck:              true,
		TsCheck:                false,
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

	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process the text to find comments
	processComments(ctx, text, configs, opts.MinimumDescriptionLength)

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

// processComments scans the source text for comments and checks for banned directives
func processComments(ctx rule.RuleContext, text string, configs map[string]*DirectiveConfig, minDescLength int) {
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
				checkComment(ctx, commentText, commentStart, configs, minDescLength, false)
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
				checkComment(ctx, commentText, commentStart, configs, minDescLength, true)
				pos = commentEnd
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
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

	// Check if there's a description
	description := strings.TrimSpace(afterDirective)

	// Remove leading separators (: -- etc.)
	description = strings.TrimLeft(description, ": \t-")
	description = strings.TrimSpace(description)

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

	// If description is required
	if config.AllowWithDescription {
		// Check if description exists
		if len(description) == 0 {
			ctx.ReportRange(
				core.NewTextRange(commentStart, commentStart+len(commentText)),
				rule.RuleMessage{
					Id:          "tsDirectiveCommentRequiresDescription",
					Description: "Include a description after the '@" + directiveName + "' directive to explain why the '@" + directiveName + "' is necessary. The description must be " + formatMinimumDescLength(minDescLength) + " characters long.",
				},
			)
			return
		}

		// Check minimum description length (counting grapheme clusters for Unicode)
		descLength := graphemeLength(description)
		if descLength < minDescLength {
			ctx.ReportRange(
				core.NewTextRange(commentStart, commentStart+len(commentText)),
				rule.RuleMessage{
					Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
					Description: "The description for the '@" + directiveName + "' directive must be " + formatMinimumDescLength(minDescLength) + " characters long.",
				},
			)
			return
		}

		// Check description format if specified
		if config.DescriptionFormat != "" {
			formatRegex, err := regexp.Compile(config.DescriptionFormat)
			if err == nil {
				// For format checking, we need to check the original afterDirective text
				// to preserve the exact format (including leading colons, etc.)
				checkText := strings.TrimSpace(afterDirective)
				if !formatRegex.MatchString(checkText) {
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
