package no_invalid_regexp

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed no_invalid_regexp.schema.json
var schemaJSON []byte

// validFlags is the set of standard RegExp flags.
var validFlags = map[rune]bool{
	'd': true,
	'g': true,
	'i': true,
	'm': true,
	's': true,
	'u': true,
	'v': true,
	'y': true,
}

// https://eslint.org/docs/latest/rules/no-invalid-regexp
var NoInvalidRegexpRule = rule.Rule{
	Name:   "no-invalid-regexp",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		report := func(node *ast.Node, msg string) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "regexMessage",
				Description: msg + ".",
			})
		}

		check := func(node *ast.Node, callee *ast.Node, args *ast.NodeList) {
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "RegExp" {
				return
			}

			if !ctx.Globals.Access("RegExp").IsDeclared() {
				return
			}
			if ctx.Refs != nil {
				if !ctx.Refs.IsGlobalReference(callee) {
					return
				}
			} else if utils.IsShadowed(callee, "RegExp") {
				return
			}

			// No arguments means RegExp() which is valid (produces /(?:)/)
			if args == nil || len(args.Nodes) == 0 {
				return
			}

			// Get flags (second argument)
			var flags *string // nil means unknown (non-literal)
			if len(args.Nodes) >= 2 {
				flagsNode := ast.SkipParentheses(args.Nodes[1])
				if flagsNode.Kind == ast.KindStringLiteral {
					f := stringLiteralValue(ctx, flagsNode)
					flags = &f
				}
			} else {
				empty := ""
				flags = &empty
			}

			// Validate flags if known
			if flags != nil {
				if msg := validateFlags(*flags, opts.allowConstructorFlags); msg != "" {
					report(node, msg)
					return
				}
			}

			// Validate pattern only if first argument is a string literal
			patternNode := ast.SkipParentheses(args.Nodes[0])
			if patternNode.Kind != ast.KindStringLiteral {
				return
			}
			pattern := stringLiteralValue(ctx, patternNode)

			if flags != nil {
				// Known flags: validate pattern with those flags
				if msg := validatePattern(pattern, *flags); msg != "" {
					report(node, msg)
				}
			} else {
				// Unknown flags: only report if pattern is invalid in ALL
				// flag combinations (unicode, unicodeSets, neither)
				if validatePattern(pattern, "u") != "" && validatePattern(pattern, "v") != "" {
					if msg := validatePattern(pattern, ""); msg != "" {
						report(node, msg)
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				check(node, callExpr.Expression, callExpr.Arguments)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				check(node, newExpr.Expression, newExpr.Arguments)
			},
		}
	},
}

// validateFlags checks for invalid or duplicate flags, and mutually exclusive u/v.
// Returns an error message or empty string if valid.
// Follows ESLint's validation order: strip valid flags → check u+v → check duplicates → check invalid.
func validateFlags(flags string, allowConstructorFlags map[rune]bool) string {
	allFlags := make(map[rune]bool)
	for k, v := range validFlags {
		allFlags[k] = v
	}
	for k, v := range allowConstructorFlags {
		allFlags[k] = v
	}

	// Strip all valid/allowed flags (one occurrence each) to get flagsToCheck
	remaining := flags
	for ch := range allFlags {
		remaining = strings.Replace(remaining, string(ch), "", 1)
	}

	// Collect duplicate flags (remaining chars that are in allFlags)
	var duplicates []rune
	for _, ch := range remaining {
		if allFlags[ch] {
			duplicates = append(duplicates, ch)
		}
	}

	// Check mutually exclusive u and v first
	if strings.Contains(flags, "u") && strings.Contains(flags, "v") {
		return "Regex 'u' and 'v' flags cannot be used together"
	}

	// Then check duplicates
	if len(duplicates) > 0 {
		return fmt.Sprintf("Duplicate flags ('%s') supplied to RegExp constructor", string(duplicates))
	}

	// Then check invalid flags
	if remaining != "" {
		return fmt.Sprintf("Invalid flags supplied to RegExp constructor '%s'", remaining)
	}

	return ""
}

// validatePattern validates syntax without invoking a matching engine. Only u/v
// affect the grammar; custom constructor flags have already been checked.
func validatePattern(pattern string, flags string) string {
	syntaxFlags := ""
	if strings.Contains(flags, "u") {
		syntaxFlags = "u"
	}
	if strings.Contains(flags, "v") {
		syntaxFlags = "v"
	}
	if err := esregexp.Validate(pattern, syntaxFlags); err != nil {
		return err.Error()
	}
	return ""
}

// The AST's decoded text does not decode legacy octal escapes in sloppy scripts.
func stringLiteralValue(ctx rule.RuleContext, node *ast.Node) string {
	if utils.IsInStrictModeWithSourceType(node, ctx.SourceFile, ctx.LanguageOptions.EffectiveSourceType()) {
		return node.AsStringLiteral().Text
	}
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	units := utils.ParseJSStringLiteralSource(ctx.SourceFile.Text()[r.Pos():r.End()])
	if units == nil {
		return node.AsStringLiteral().Text
	}
	values := make([]uint16, len(units))
	for i, unit := range units {
		values[i] = uint16(unit.Value)
	}
	return ecmascript.StringFromCodeUnits(values)
}

type noInvalidRegexpOptions struct {
	allowConstructorFlags map[rune]bool
}

func parseOptions(options []any) noInvalidRegexpOptions {
	opts := noInvalidRegexpOptions{
		allowConstructorFlags: make(map[rune]bool),
	}

	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]any)
	if allowArr, ok := m["allowConstructorFlags"].([]interface{}); ok {
		for _, item := range allowArr {
			if str, ok := item.(string); ok {
				for _, ch := range str {
					opts.allowConstructorFlags[ch] = true
				}
			}
		}
	}

	return opts
}
