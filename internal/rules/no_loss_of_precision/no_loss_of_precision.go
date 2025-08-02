package no_loss_of_precision

import (
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

func buildNoLossOfPrecisionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noLossOfPrecision",
		Description: "This number literal will lose precision at runtime.",
	}
}

var NoLossOfPrecisionRule = rule.Rule{
	Name: "no-loss-of-precision",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				numericLiteral := node.AsNumericLiteral()
				text := numericLiteral.Text
				tokenFlags := numericLiteral.TokenFlags

				// Skip if it has invalid separators or other issues
				if tokenFlags&ast.TokenFlagsContainsInvalidSeparator != 0 {
					return
				}

				// Remove underscores (numeric separators)
				cleanText := strings.ReplaceAll(text, "_", "")

				// Parse the numeric value - since TypeScript has already converted 
				// hex/binary/octal to decimal in the text, just parse as float
				value, err := strconv.ParseFloat(cleanText, 64)
				if err != nil {
					// Parsing error - skip this literal
					return
				}

				// Check if the value loses precision
				if isLossOfPrecision(value, text, tokenFlags) {
					ctx.ReportNode(node, buildNoLossOfPrecisionMessage())
				}
			},
		}
	},
}

// isLossOfPrecision checks if a numeric value loses precision when represented as a float64
func isLossOfPrecision(value float64, originalText string, tokenFlags ast.TokenFlags) bool {
	// If it's not finite, there's no precision loss to check
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return false
	}

	const maxSafeInteger = 9007199254740991 // 2^53 - 1
	
	// Remove underscores for parsing
	cleanText := strings.ReplaceAll(originalText, "_", "")
	
	// Since TypeScript has already converted the literal to decimal format in the text,
	// we can directly check if the value exceeds MAX_SAFE_INTEGER for integer literals
	// or use other precision loss detection methods
	
	if tokenFlags&ast.TokenFlagsScientific != 0 || strings.Contains(cleanText, "e") || strings.Contains(cleanText, "E") {
		// Scientific notation
		
		// If the tokenFlags indicate scientific notation was in the original source,
		// check if it represents a precise integer that exceeds MAX_SAFE_INTEGER
		if tokenFlags&ast.TokenFlagsScientific != 0 {
			// User wrote explicit scientific notation like 9.007199254740993e3
			if math.Abs(value) > maxSafeInteger && value == math.Trunc(value) {
				return true
			}
		} else {
			// TypeScript auto-converted to scientific notation (like 1.23e+25)
			// These are generally acceptable as they represent very large numbers
			// that JavaScript naturally represents in scientific notation
			return false
		}
		return false
	} else {
		// For all other number formats (hex, binary, octal, decimal)
		// TypeScript has already converted them to decimal representation
		// Check if it's an integer that exceeds MAX_SAFE_INTEGER
		if value == math.Trunc(value) {
			// It's an integer value
			if math.Abs(value) > maxSafeInteger {
				return true
			}
		}
	}
	
	return false
}