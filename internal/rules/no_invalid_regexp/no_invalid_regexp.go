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

			// No arguments means RegExp() which is valid (produces /(?:)/)
			if args == nil || len(args.Nodes) == 0 {
				return
			}

			// Get flags (second argument)
			var flags *string // nil means unknown (non-literal)
			if len(args.Nodes) >= 2 {
				flagsNode := args.Nodes[1]
				if flagsNode.Kind == ast.KindStringLiteral {
					f := flagsNode.AsStringLiteral().Text
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
			patternNode := args.Nodes[0]
			if patternNode.Kind != ast.KindStringLiteral {
				return
			}
			pattern := patternNode.AsStringLiteral().Text

			if flags != nil {
				// Known flags: validate pattern with those flags
				if msg := validatePattern(pattern, *flags); msg != "" {
					report(node, msg)
				}
			} else {
				// Unknown flags: only report if pattern is invalid in ALL
				// flag combinations (unicode, unicodeSets, neither)
				msg1 := validatePattern(pattern, "u")
				msg2 := validatePattern(pattern, "v")
				msg3 := validatePattern(pattern, "")
				if msg1 != "" && msg2 != "" && msg3 != "" {
					report(node, msg3)
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
func validateFlags(flags string, allowConstructorFlags map[byte]bool) string {
	allFlags := make(map[byte]bool)
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
	var duplicates []byte
	for i := range len(remaining) {
		if allFlags[remaining[i]] {
			duplicates = append(duplicates, remaining[i])
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

// validatePattern tries to compile the pattern using regexp2 with ECMAScript mode.
// Returns an error message or empty string if valid.
// Note: regexp2 is a .NET regex port and does not fully support ECMAScript semantics.
// See no_invalid_regexp.md "Known Limitations" for details on misalignments with ESLint.
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
