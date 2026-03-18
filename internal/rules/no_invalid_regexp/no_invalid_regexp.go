package no_invalid_regexp

import (
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// validFlags is the set of standard RegExp flags.
var validFlags = map[byte]bool{
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
	Name: "no-invalid-regexp",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		check := func(node *ast.Node, callee *ast.Node, args *ast.NodeList) {
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "RegExp" {
				return
			}

			// No arguments means RegExp() which is valid (produces /(?:)/)
			if args == nil || len(args.Nodes) == 0 {
				return
			}

			// Get pattern (first argument) - only validate string literals
			patternNode := args.Nodes[0]
			if patternNode.Kind != ast.KindStringLiteral {
				return
			}
			pattern := patternNode.AsStringLiteral().Text

			// Get flags (second argument) - only validate string literals
			var flags string
			var hasFlags bool
			flagsAreKnown := true
			if len(args.Nodes) >= 2 {
				flagsNode := args.Nodes[1]
				if flagsNode.Kind == ast.KindStringLiteral {
					flags = flagsNode.AsStringLiteral().Text
					hasFlags = true
				} else {
					// Non-literal flags — we still validate the pattern
					// by trying without special flags. If the pattern is invalid
					// regardless of flags, we report it.
					flagsAreKnown = false
				}
			}

			// Validate flags if present and known
			if hasFlags && flagsAreKnown {
				if msg := validateFlags(flags, opts.allowConstructorFlags); msg != "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "regexMessage",
						Description: msg + ".",
					})
					return
				}
			}

			// Validate pattern
			if flagsAreKnown {
				// Known flags: validate pattern with those flags
				if msg := validatePattern(pattern, flags); msg != "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "regexMessage",
						Description: msg + ".",
					})
				}
			} else {
				// Unknown flags: only report if the pattern is invalid
				// regardless of flags (try without any special flags)
				if msg := validatePattern(pattern, ""); msg != "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "regexMessage",
						Description: msg + ".",
					})
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
func validateFlags(flags string, allowConstructorFlags map[byte]bool) string {
	seen := make(map[byte]bool)
	for i := range len(flags) {
		ch := flags[i]
		if !validFlags[ch] && !allowConstructorFlags[ch] {
			return fmt.Sprintf("Invalid flags supplied to RegExp constructor '%s'", flags)
		}
		if seen[ch] {
			return fmt.Sprintf("Invalid flags supplied to RegExp constructor '%s'", flags)
		}
		seen[ch] = true
	}
	// Check mutually exclusive u and v
	if seen['u'] && seen['v'] {
		return fmt.Sprintf("Invalid flags supplied to RegExp constructor '%s'", flags)
	}
	return ""
}

// validatePattern tries to compile the pattern using regexp2 with ECMAScript mode.
// Returns an error message or empty string if valid.
func validatePattern(pattern string, flags string) string {
	var regexpFlags regexp2.RegexOptions = regexp2.ECMAScript
	if strings.Contains(flags, "i") {
		regexpFlags |= regexp2.IgnoreCase
	}
	if strings.Contains(flags, "m") {
		regexpFlags |= regexp2.Multiline
	}
	if strings.Contains(flags, "s") {
		regexpFlags |= regexp2.Singleline
	}
	if strings.Contains(flags, "u") || strings.Contains(flags, "v") {
		regexpFlags |= regexp2.Unicode
	}

	_, err := regexp2.Compile(pattern, regexpFlags)
	if err != nil {
		// Extract a clean error message
		errMsg := err.Error()
		// regexp2 error messages often start with "error parsing regexp:"
		if idx := strings.Index(errMsg, ": "); idx != -1 {
			errMsg = errMsg[idx+2:]
		}
		return fmt.Sprintf("Invalid regular expression: /%s/: %s", pattern, errMsg)
	}
	return ""
}

type noInvalidRegexpOptions struct {
	allowConstructorFlags map[byte]bool
}

func parseOptions(opts any) noInvalidRegexpOptions {
	result := noInvalidRegexpOptions{
		allowConstructorFlags: make(map[byte]bool),
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if allowArr, ok := optsMap["allowConstructorFlags"].([]interface{}); ok {
			for _, item := range allowArr {
				if str, ok := item.(string); ok {
					for i := range len(str) {
						result.allowConstructorFlags[str[i]] = true
					}
				}
			}
		}
		// Also handle string format: "allowConstructorFlags": "a"
		if allowStr, ok := optsMap["allowConstructorFlags"].(string); ok {
			for i := range len(allowStr) {
				result.allowConstructorFlags[allowStr[i]] = true
			}
		}
	}

	return result
}
