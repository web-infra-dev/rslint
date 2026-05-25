package ban_ts_comment

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/rivo/uniseg"
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

// Regexes for single-line comments.
// ts-expect-error and ts-ignore: match any number of leading slashes (original: /^\/*\s*@ts-.../)
// ts-check and ts-nocheck: only match 2 or 3 slashes (pragma comments: /^\/\/\/?\s*@ts-.../)
var (
	// For ts-expect-error / ts-ignore in single-line comments: any number of leading slashes
	singleLineDirectiveRegex = regexp.MustCompile(`^/{2,}\s*@ts-(expect-error|ignore)\b`)

	// For ts-check / ts-nocheck in single-line comments: exactly 2 or 3 slashes (pragma style)
	singleLinePragmaRegex = regexp.MustCompile(`^///?\s*@ts-(check|nocheck)\b`)

	// For ts-expect-error / ts-ignore on the last line of a block comment
	// Matches optional whitespace, optional leading * or /, then the directive
	multiLineLastLineRegex = regexp.MustCompile(`^\s*(?:[/*])*\s*@ts-(expect-error|ignore)\b`)
)

// BanTsCommentRule implements the ban-ts-comment rule
// Bans @ts-<directive> comments or requires descriptions after directive
var BanTsCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-ts-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := BanTsCommentOptions{
		TsExpectError:            "allow-with-description",
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

	// Determine token position of the first statement (excluding leading trivia)
	// for ts-nocheck pragma check: only report if comment is before first real code
	firstStatementPos := -1
	if ctx.SourceFile.Statements != nil && len(ctx.SourceFile.Statements.Nodes) > 0 {
		firstStatementPos = scanner.GetTokenPosOfNode(ctx.SourceFile.Statements.Nodes[0], ctx.SourceFile, false)
	}

	processComments(ctx, ctx.SourceFile.Text(), configs, opts.MinimumDescriptionLength, firstStatementPos)

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
func processComments(ctx rule.RuleContext, text string, configs map[string]*DirectiveConfig, minDescLength int, firstStatementPos int) {
	utils.ForEachComment(ctx.SourceFile.AsNode(), func(comment *ast.CommentRange) {
		if comment == nil {
			return
		}
		if comment.Pos() < 0 || comment.End() > len(text) || comment.Pos() >= comment.End() {
			return
		}

		commentText := text[comment.Pos():comment.End()]

		switch comment.Kind {
		case ast.KindSingleLineCommentTrivia:
			checkSingleLineComment(ctx, commentText, comment.Pos(), configs, minDescLength, firstStatementPos)
		case ast.KindMultiLineCommentTrivia:
			checkMultiLineComment(ctx, commentText, comment.Pos(), configs, minDescLength)
		}
	}, ctx.SourceFile)
}

// checkSingleLineComment handles single-line comments (// or /// style).
func checkSingleLineComment(ctx rule.RuleContext, commentText string, commentStart int, configs map[string]*DirectiveConfig, minDescLength int, firstStatementPos int) {
	// Try matching ts-expect-error / ts-ignore (any number of leading slashes)
	if match := singleLineDirectiveRegex.FindStringSubmatch(commentText); match != nil {
		directiveType := match[1] // "expect-error" or "ignore"
		directiveName := "ts-" + directiveType
		description := extractSingleLineDescription(commentText, "@ts-"+directiveType)
		reportDirective(ctx, commentText, commentStart, configs, directiveName, directiveType, description, minDescLength)
		return
	}

	// Try matching ts-check / ts-nocheck (only 2 or 3 slashes, pragma style)
	if match := singleLinePragmaRegex.FindStringSubmatch(commentText); match != nil {
		directiveType := match[1] // "check" or "nocheck"
		directiveName := "ts-" + directiveType

		// ts-nocheck after the first statement is not effective and should not be reported
		if directiveType == "nocheck" && firstStatementPos >= 0 && commentStart >= firstStatementPos {
			return
		}

		description := extractSingleLineDescription(commentText, "@ts-"+directiveType)
		reportDirective(ctx, commentText, commentStart, configs, directiveName, directiveType, description, minDescLength)
	}
}

// checkMultiLineComment handles block comments (/* */ style).
// Only ts-expect-error and ts-ignore are checked in block comments.
// ts-check and ts-nocheck are pragma-only (single-line).
// The directive must appear on the LAST line of the comment to be recognized.
func checkMultiLineComment(ctx rule.RuleContext, commentText string, commentStart int, configs map[string]*DirectiveConfig, minDescLength int) {
	// Extract content between /* and */
	if len(commentText) < 4 {
		return
	}
	content := commentText[2 : len(commentText)-2]

	// Get the last line of the comment content
	lastLineStart := strings.LastIndexByte(content, '\n')
	var lastLine string
	if lastLineStart == -1 {
		lastLine = content
	} else {
		lastLine = content[lastLineStart+1:]
	}

	// Check if the last line contains a directive (ts-expect-error or ts-ignore only)
	match := multiLineLastLineRegex.FindStringSubmatch(lastLine)
	if match == nil {
		return
	}

	directiveType := match[1] // "expect-error" or "ignore"
	directiveName := "ts-" + directiveType

	// Extract raw description from the last line after the directive
	directivePattern := "@ts-" + directiveType
	idx := strings.Index(lastLine, directivePattern)
	if idx == -1 {
		return
	}
	rawDescription := lastLine[idx+len(directivePattern):]

	reportDirective(ctx, commentText, commentStart, configs, directiveName, directiveType, rawDescription, minDescLength)
}

// extractSingleLineDescription extracts the raw description from a single-line comment after the directive.
// The returned string is NOT trimmed — callers decide how to use it (raw for format check, trimmed for length).
func extractSingleLineDescription(commentText string, directive string) string {
	idx := strings.Index(commentText, directive)
	if idx == -1 {
		return ""
	}
	return commentText[idx+len(directive):]
}

// reportDirective reports a directive violation if applicable.
// rawDescription is the text after the directive keyword (untrimmed), used for format matching.
func reportDirective(ctx rule.RuleContext, commentText string, commentStart int, configs map[string]*DirectiveConfig, directiveName string, directiveType string, rawDescription string, minDescLength int) {
	config := configs[directiveName]
	if config == nil || !config.Enabled {
		return
	}

	commentRange := core.NewTextRange(commentStart, commentStart+len(commentText))

	// Special case: for ts-ignore with no description allowed, suggest ts-expect-error
	// BUT only if ts-expect-error is not also completely banned — otherwise it's contradictory
	// to suggest something that's also forbidden.
	expectErrorConfig := configs["ts-expect-error"]
	expectErrorAlsoBanned := expectErrorConfig != nil && expectErrorConfig.Enabled && !expectErrorConfig.AllowWithDescription
	if directiveType == "ignore" && !config.AllowWithDescription && !expectErrorAlsoBanned {
		msg := rule.RuleMessage{
			Id:          "tsIgnoreInsteadOfExpectError",
			Description: "Prefer '@ts-expect-error' over '@ts-ignore' as it requires error to be present in next line.",
		}
		// Provide a suggestion (not autofix): replace @ts-ignore with @ts-expect-error.
		// This is intentionally a suggestion rather than an autofix because @ts-expect-error
		// requires the next line to actually have a type error — auto-replacing could break code.
		idx := strings.LastIndex(commentText, "@ts-ignore")
		if idx >= 0 {
			fixRange := core.NewTextRange(commentStart+idx, commentStart+idx+len("@ts-ignore"))
			fix := rule.RuleFixReplaceRange(fixRange, "@ts-expect-error")
			ctx.ReportRangeWithSuggestions(commentRange, msg, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "replaceTsIgnoreWithTsExpectError",
					Description: "Replace '@ts-ignore' with '@ts-expect-error'.",
				},
				FixesArr: []rule.RuleFix{fix},
			})
		} else {
			ctx.ReportRange(commentRange, msg)
		}
		return
	}

	// If the directive is completely banned (no description allowed)
	if !config.AllowWithDescription {
		ctx.ReportRange(
			commentRange,
			rule.RuleMessage{
				Id:          "tsDirectiveComment",
				Description: "Do not use '@" + directiveName + "' because it alters compilation errors.",
			},
		)
		return
	}

	// Trimmed description for length check; raw description for format check
	trimmedDescription := strings.TrimSpace(rawDescription)

	// Check minimum length using grapheme cluster count on the trimmed description
	descLength := graphemeLength(trimmedDescription)
	if descLength < minDescLength {
		ctx.ReportRange(
			commentRange,
			rule.RuleMessage{
				Id:          "tsDirectiveCommentRequiresDescription",
				Description: "Include a description after the '@" + directiveName + "' directive to explain why the '@" + directiveName + "' is necessary. The description must be " + formatInt(minDescLength) + " characters or longer.",
			},
		)
		return
	}

	// Check description format against raw (untrimmed) description
	if config.DescriptionFormat != "" {
		formatRegex, err := regexp.Compile(config.DescriptionFormat)
		if err == nil {
			if !formatRegex.MatchString(rawDescription) {
				ctx.ReportRange(
					commentRange,
					rule.RuleMessage{
						Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
						Description: "The description for the '@" + directiveName + "' directive must match the format '" + config.DescriptionFormat + "'.",
					},
				)
			}
		}
	}
}

// graphemeLength returns the number of grapheme clusters in a string.
// Uses proper Unicode grapheme cluster segmentation (UAX#29) via rivo/uniseg.
func graphemeLength(s string) int {
	return uniseg.GraphemeClusterCount(s)
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
