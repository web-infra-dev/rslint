package ban_ts_comment

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/rivo/uniseg"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type DirectiveConfig struct {
	Enabled              bool   // Whether the directive is enabled (true means banned)
	AllowWithDescription bool   // Whether to allow with description
	DescriptionFormat    string // Regex pattern for description format
	descriptionRegex     *regexp.Regexp
}

type BanTsCommentOptions struct {
	TsExpectError            interface{} `json:"ts-expect-error"`
	TsIgnore                 interface{} `json:"ts-ignore"`
	TsNocheck                interface{} `json:"ts-nocheck"`
	TsCheck                  interface{} `json:"ts-check"`
	MinimumDescriptionLength int         `json:"minimumDescriptionLength"`
}

type directiveKind uint8

const (
	directiveExpectError directiveKind = iota
	directiveIgnore
	directiveNocheck
	directiveCheck
	directiveCount
)

type directiveConfigs [directiveCount]DirectiveConfig

var (
	tsIgnoreInsteadOfExpectErrorMessage = rule.RuleMessage{
		Id:          "tsIgnoreInsteadOfExpectError",
		Description: "Prefer '@ts-expect-error' over '@ts-ignore' as it requires error to be present in next line.",
	}
	replaceTsIgnoreWithTsExpectErrorMessage = rule.RuleMessage{
		Id:          "replaceTsIgnoreWithTsExpectError",
		Description: "Replace '@ts-ignore' with '@ts-expect-error'.",
	}
)

// BanTsCommentRule implements the ban-ts-comment rule
// Bans @ts-<directive> comments or requires descriptions after directive
var BanTsCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-ts-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, _options []any) rule.RuleListeners {
	text := ctx.SourceFile.Text()
	if !strings.Contains(text, "@ts-") {
		return rule.RuleListeners{}
	}

	opts := parseOptions(_options)
	configs := directiveConfigs{
		directiveExpectError: parseDirectiveConfig(opts.TsExpectError),
		directiveIgnore:      parseDirectiveConfig(opts.TsIgnore),
		directiveNocheck:     parseDirectiveConfig(opts.TsNocheck),
		directiveCheck:       parseDirectiveConfig(opts.TsCheck),
	}
	if !containsEnabledDirective(text, &configs) {
		return rule.RuleListeners{}
	}

	// Determine token position of the first statement (excluding leading trivia)
	// only when an enabled ts-nocheck pragma could need the position check.
	firstStatementPos := -1
	if configs[directiveNocheck].Enabled && strings.Contains(text, "@ts-nocheck") &&
		ctx.SourceFile.Statements != nil && len(ctx.SourceFile.Statements.Nodes) > 0 {
		firstStatementPos = scanner.GetTokenPosOfNode(ctx.SourceFile.Statements.Nodes[0], ctx.SourceFile, false)
	}

	processComments(ctx, text, &configs, opts.MinimumDescriptionLength, firstStatementPos)

	return rule.RuleListeners{}
}

func parseOptions(options []any) BanTsCommentOptions {
	opts := BanTsCommentOptions{
		TsExpectError:            "allow-with-description",
		TsIgnore:                 true,
		TsNocheck:                true,
		TsCheck:                  false,
		MinimumDescriptionLength: 3,
	}

	if len(options) == 0 {
		return opts
	}

	option := rule.LegacyUnwrapOptions(options)
	// Preserve compatibility with callers that still pass the old nested
	// array shape after the shared legacy-options shim has unwrapped Run's
	// ESLint-style context.options.
	if optionArray, ok := option.([]interface{}); ok && len(optionArray) > 0 {
		option = optionArray[0]
	}
	optsMap, ok := option.(map[string]interface{})
	if !ok {
		return opts
	}

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

	return opts
}

// parseDirectiveConfig converts the option value to DirectiveConfig
func parseDirectiveConfig(value interface{}) DirectiveConfig {
	config := DirectiveConfig{}

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
	if config.DescriptionFormat != "" {
		config.descriptionRegex, _ = regexp.Compile(config.DescriptionFormat)
	}

	return config
}

// containsEnabledDirective is a source-text gate before the shared comment
// store is materialized. It is deliberately conservative about lexical
// context (a marker in a string may reach the comment scan), but exact about
// directive spelling, word boundaries, and whether that directive is enabled.
func containsEnabledDirective(text string, configs *directiveConfigs) bool {
	for searchStart := 0; searchStart < len(text); {
		offset := strings.Index(text[searchStart:], "@ts-")
		if offset < 0 {
			return false
		}
		markerStart := searchStart + offset
		if directive, _, ok := matchDirectiveNameAt(text, markerStart); ok && configs[directive].Enabled {
			return true
		}
		searchStart = markerStart + len("@ts-")
	}
	return false
}

// processComments scans real comment trivia and checks for banned directives.
func processComments(ctx rule.RuleContext, text string, configs *directiveConfigs, minDescLength int, firstStatementPos int) {
	for _, comment := range ctx.Comments.All() {
		if comment == nil {
			continue
		}
		if comment.Pos() < 0 || comment.End() > len(text) || comment.Pos() >= comment.End() {
			continue
		}

		commentText := text[comment.Pos():comment.End()]

		switch comment.Kind {
		case ast.KindSingleLineCommentTrivia:
			checkSingleLineComment(ctx, commentText, comment.Pos(), configs, minDescLength, firstStatementPos)
		case ast.KindMultiLineCommentTrivia:
			checkMultiLineComment(ctx, commentText, comment.Pos(), configs, minDescLength)
		}
	}
}

// checkSingleLineComment handles single-line comments (// or /// style).
func checkSingleLineComment(ctx rule.RuleContext, commentText string, commentStart int, configs *directiveConfigs, minDescLength int, firstStatementPos int) {
	directive, descriptionStart, ok := matchSingleLineDirective(commentText)
	if !ok {
		return
	}

	// ts-nocheck after the first statement is not effective and should not be reported.
	if directive == directiveNocheck && firstStatementPos >= 0 && commentStart >= firstStatementPos {
		return
	}

	reportDirective(ctx, commentText, commentStart, configs, directive, commentText[descriptionStart:], minDescLength)
}

// checkMultiLineComment handles block comments (/* */ style).
// Only ts-expect-error and ts-ignore are checked in block comments.
// ts-check and ts-nocheck are pragma-only (single-line).
// The directive must appear on the LAST line of the comment to be recognized.
func checkMultiLineComment(ctx rule.RuleContext, commentText string, commentStart int, configs *directiveConfigs, minDescLength int) {
	directive, descriptionStart, contentEnd, ok := matchMultiLineDirective(commentText)
	if !ok {
		return
	}

	reportDirective(ctx, commentText, commentStart, configs, directive, commentText[descriptionStart:contentEnd], minDescLength)
}

func matchSingleLineDirective(commentText string) (directiveKind, int, bool) {
	slashCount := 0
	for slashCount < len(commentText) && commentText[slashCount] == '/' {
		slashCount++
	}
	if slashCount < 2 {
		return 0, 0, false
	}

	directiveStart := skipDirectiveWhitespace(commentText, slashCount)
	directive, descriptionStart, ok := matchDirectiveNameAt(commentText, directiveStart)
	if !ok {
		return 0, 0, false
	}
	if directive == directiveExpectError || directive == directiveIgnore {
		return directive, descriptionStart, true
	}
	if slashCount <= 3 && (directive == directiveCheck || directive == directiveNocheck) {
		return directive, descriptionStart, true
	}
	return 0, 0, false
}

func matchMultiLineDirective(commentText string) (directiveKind, int, int, bool) {
	if len(commentText) < 4 {
		return 0, 0, 0, false
	}

	contentStart := 2
	contentEnd := len(commentText) - 2
	if lastLineBreak := strings.LastIndexByte(commentText[contentStart:contentEnd], '\n'); lastLineBreak >= 0 {
		contentStart += lastLineBreak + 1
	}

	directiveStart := skipDirectiveWhitespace(commentText, contentStart)
	for directiveStart < contentEnd && (commentText[directiveStart] == '/' || commentText[directiveStart] == '*') {
		directiveStart++
	}
	directiveStart = skipDirectiveWhitespace(commentText[:contentEnd], directiveStart)

	directive, descriptionStart, ok := matchDirectiveNameAt(commentText[:contentEnd], directiveStart)
	if !ok || (directive != directiveExpectError && directive != directiveIgnore) {
		return 0, 0, 0, false
	}
	return directive, descriptionStart, contentEnd, true
}

func matchDirectiveNameAt(text string, start int) (directiveKind, int, bool) {
	if start < 0 || start+len("@ts-") > len(text) || text[start:start+len("@ts-")] != "@ts-" {
		return 0, 0, false
	}

	nameStart := start + len("@ts-")
	var directive directiveKind
	var name string
	if nameStart < len(text) {
		switch text[nameStart] {
		case 'e':
			directive, name = directiveExpectError, "expect-error"
		case 'i':
			directive, name = directiveIgnore, "ignore"
		case 'n':
			directive, name = directiveNocheck, "nocheck"
		case 'c':
			directive, name = directiveCheck, "check"
		}
	}
	nameEnd := nameStart + len(name)
	if name == "" || nameEnd > len(text) || text[nameStart:nameEnd] != name {
		return 0, 0, false
	}
	if nameEnd < len(text) && isASCIIWordByte(text[nameEnd]) {
		return 0, 0, false
	}
	return directive, nameEnd, true
}

func skipDirectiveWhitespace(text string, start int) int {
	for start < len(text) {
		switch text[start] {
		case ' ', '\t', '\n', '\f', '\r':
			start++
		default:
			return start
		}
	}
	return start
}

func isASCIIWordByte(value byte) bool {
	return value == '_' || value >= '0' && value <= '9' || value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

// reportDirective reports a directive violation if applicable.
// rawDescription is the text after the directive keyword (untrimmed), used for format matching.
func reportDirective(ctx rule.RuleContext, commentText string, commentStart int, configs *directiveConfigs, directive directiveKind, rawDescription string, minDescLength int) {
	config := &configs[directive]
	if !config.Enabled {
		return
	}

	commentRange := core.NewTextRange(commentStart, commentStart+len(commentText))

	// Special case: for ts-ignore with no description allowed, suggest ts-expect-error
	// BUT only if ts-expect-error is not also completely banned — otherwise it's contradictory
	// to suggest something that's also forbidden.
	expectErrorConfig := &configs[directiveExpectError]
	expectErrorAlsoBanned := expectErrorConfig.Enabled && !expectErrorConfig.AllowWithDescription
	if directive == directiveIgnore && !config.AllowWithDescription && !expectErrorAlsoBanned {
		// Provide a suggestion (not autofix): replace @ts-ignore with @ts-expect-error.
		// This is intentionally a suggestion rather than an autofix because @ts-expect-error
		// requires the next line to actually have a type error — auto-replacing could break code.
		ctx.ReportRangeWithDeferredSuggestions(commentRange, tsIgnoreInsteadOfExpectErrorMessage, func() []rule.RuleSuggestion {
			idx := strings.LastIndex(commentText, "@ts-ignore")
			if idx < 0 {
				return nil
			}
			fixRange := core.NewTextRange(commentStart+idx, commentStart+idx+len("@ts-ignore"))
			return []rule.RuleSuggestion{{
				Message:  replaceTsIgnoreWithTsExpectErrorMessage,
				FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, "@ts-expect-error")},
			}}
		})
		return
	}

	// If the directive is completely banned (no description allowed)
	if !config.AllowWithDescription {
		ctx.ReportRange(commentRange, directiveCommentMessage(directive))
		return
	}
	fullDirectiveName := directiveName(directive)

	// Trimmed description for length check; raw description for format check
	trimmedDescription := strings.TrimSpace(rawDescription)

	// Check minimum length using grapheme cluster count on the trimmed description
	descLength := graphemeLength(trimmedDescription)
	if descLength < minDescLength {
		ctx.ReportRange(
			commentRange,
			rule.RuleMessage{
				Id:          "tsDirectiveCommentRequiresDescription",
				Description: "Include a description after the '@" + fullDirectiveName + "' directive to explain why the '@" + fullDirectiveName + "' is necessary. The description must be " + strconv.Itoa(minDescLength) + " characters or longer.",
			},
		)
		return
	}

	// Check description format against raw (untrimmed) description
	if config.descriptionRegex != nil && !config.descriptionRegex.MatchString(rawDescription) {
		ctx.ReportRange(
			commentRange,
			rule.RuleMessage{
				Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
				Description: "The description for the '@" + fullDirectiveName + "' directive must match the format '" + config.DescriptionFormat + "'.",
			},
		)
	}
}

func directiveName(directive directiveKind) string {
	switch directive {
	case directiveExpectError:
		return "ts-expect-error"
	case directiveIgnore:
		return "ts-ignore"
	case directiveNocheck:
		return "ts-nocheck"
	case directiveCheck:
		return "ts-check"
	default:
		return ""
	}
}

func directiveCommentMessage(directive directiveKind) rule.RuleMessage {
	switch directive {
	case directiveExpectError:
		return rule.RuleMessage{Id: "tsDirectiveComment", Description: "Do not use '@ts-expect-error' because it alters compilation errors."}
	case directiveIgnore:
		return rule.RuleMessage{Id: "tsDirectiveComment", Description: "Do not use '@ts-ignore' because it alters compilation errors."}
	case directiveNocheck:
		return rule.RuleMessage{Id: "tsDirectiveComment", Description: "Do not use '@ts-nocheck' because it alters compilation errors."}
	case directiveCheck:
		return rule.RuleMessage{Id: "tsDirectiveComment", Description: "Do not use '@ts-check' because it alters compilation errors."}
	default:
		return rule.RuleMessage{Id: "tsDirectiveComment"}
	}
}

// graphemeLength returns the number of grapheme clusters in a string.
// Uses proper Unicode grapheme cluster segmentation (UAX#29) via rivo/uniseg.
func graphemeLength(s string) int {
	return uniseg.GraphemeClusterCount(s)
}
