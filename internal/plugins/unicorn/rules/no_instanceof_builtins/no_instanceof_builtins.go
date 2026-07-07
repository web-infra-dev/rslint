package no_instanceof_builtins

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDNoInstanceofBuiltins = "no-instanceof-builtins"
	messageIDSwitchToTypeOf       = "switch-to-type-of"
)

var primitiveWrappers = utils.NewSetFromItems("String", "Number", "Boolean", "BigInt", "Symbol")

var strictStrategyConstructors = utils.NewSetFromItems(
	// Error types
	"Error",
	"EvalError",
	"RangeError",
	"ReferenceError",
	"SyntaxError",
	"TypeError",
	"URIError",
	"AggregateError",
	"SuppressedError",

	// Collection types
	"Map",
	"Set",
	"WeakMap",
	"WeakRef",
	"WeakSet",

	// Arrays and Typed Arrays
	"ArrayBuffer",
	"Int8Array",
	"Uint8Array",
	"Uint8ClampedArray",
	"Int16Array",
	"Uint16Array",
	"Int32Array",
	"Uint32Array",
	"Float16Array",
	"Float32Array",
	"Float64Array",
	"BigInt64Array",
	"BigUint64Array",

	// Data types
	"Object",

	// Regular Expressions
	"RegExp",

	// Async and functions
	"Promise",
	"Proxy",

	// Other
	"DataView",
	"Date",
	"SharedArrayBuffer",
	"FinalizationRegistry",
)

type options struct {
	useErrorIsError bool
	strategy        string
	include         *utils.Set[string]
	exclude         *utils.Set[string]
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/no-instanceof-builtins.md
var NoInstanceofBuiltinsRule = rule.Rule{
	Name: "unicorn/no-instanceof-builtins",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.UnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				checkBinaryExpression(ctx, node, opts)
			},
		}
	},
}

func checkBinaryExpression(ctx rule.RuleContext, node *ast.Node, opts options) {
	binary := node.AsBinaryExpression()
	if binary == nil || !ast.IsInstanceOfExpression(node) {
		return
	}

	right := ast.SkipParentheses(binary.Right)
	if right == nil || !ast.IsIdentifier(right) {
		return
	}

	constructorName := right.AsIdentifier().Text
	if opts.exclude.Has(constructorName) {
		return
	}

	message := messageNoInstanceofBuiltins()
	switch {
	case constructorName == "Array":
		ctx.ReportNodeWithFixes(
			node,
			message,
			replaceWithFunctionCall(ctx.SourceFile, binary, "Array.isArray")...,
		)
	case constructorName == "Error" && opts.useErrorIsError:
		ctx.ReportNodeWithFixes(
			node,
			message,
			replaceWithFunctionCall(ctx.SourceFile, binary, "Error.isError")...,
		)
	case constructorName == "Function":
		ctx.ReportNodeWithFixes(
			node,
			message,
			replaceWithTypeOfExpression(ctx.SourceFile, binary, constructorName)...,
		)
	case primitiveWrappers.Has(constructorName):
		ctx.ReportNodeWithSuggestions(
			node,
			message,
			rule.RuleSuggestion{
				Message:  messageSwitchToTypeOf(constructorName),
				FixesArr: replaceWithTypeOfExpression(ctx.SourceFile, binary, constructorName),
			},
		)
	case opts.forbids(constructorName):
		ctx.ReportNode(node, message)
	}
}

func parseOptions(rawOptions any) options {
	opts := options{
		useErrorIsError: false,
		strategy:        "loose",
		include:         utils.NewSetFromItems[string](),
		exclude:         utils.NewSetFromItems[string](),
	}

	optsMap := utils.GetOptionsMap(rawOptions)
	if optsMap == nil {
		return opts
	}

	if useErrorIsError, ok := optsMap["useErrorIsError"].(bool); ok {
		opts.useErrorIsError = useErrorIsError
	}
	if strategy, ok := optsMap["strategy"].(string); ok && (strategy == "loose" || strategy == "strict") {
		opts.strategy = strategy
	}
	opts.include = parseStringSet(optsMap["include"])
	opts.exclude = parseStringSet(optsMap["exclude"])

	return opts
}

func (opts options) forbids(constructorName string) bool {
	if opts.strategy == "strict" && strictStrategyConstructors.Has(constructorName) {
		return true
	}
	return opts.include.Has(constructorName)
}

func parseStringSet(value any) *utils.Set[string] {
	if array, ok := value.([]string); ok {
		return utils.NewSetFromItems(array...)
	}
	return utils.NewSetFromItems(utils.ToStringSlice(value)...)
}

// Use token-level edits so comments around `instanceof` and the right operand
// are preserved the same way as upstream's fixers.
func replaceWithFunctionCall(sourceFile *ast.SourceFile, binary *ast.BinaryExpression, functionName string) []rule.RuleFix {
	leftRange := utils.TrimNodeTextRange(sourceFile, binary.Left)
	insertBefore := functionName + "("
	if utils.NeedsLeadingSpaceForReplacement(sourceFile.Text(), leftRange.Pos(), insertBefore) {
		insertBefore = " " + insertBefore
	}

	fixes := []rule.RuleFix{
		insertAt(leftRange.Pos(), insertBefore),
		insertAt(leftRange.End(), ")"),
		removeRangeAndSpacesBefore(sourceFile, utils.TrimNodeTextRange(sourceFile, binary.OperatorToken)),
	}
	fixes = append(fixes, removeNodeSyntaxAndSpacesBefore(sourceFile, binary.Right)...)
	return fixes
}

func replaceWithTypeOfExpression(sourceFile *ast.SourceFile, binary *ast.BinaryExpression, constructorName string) []rule.RuleFix {
	leftRange := utils.TrimNodeTextRange(sourceFile, binary.Left)
	insertBefore := "typeof "
	if utils.NeedsLeadingSpaceForReplacement(sourceFile.Text(), leftRange.Pos(), insertBefore) {
		insertBefore = " " + insertBefore
	}

	return []rule.RuleFix{
		insertAt(leftRange.Pos(), insertBefore),
		rule.RuleFixReplaceRange(utils.TrimNodeTextRange(sourceFile, binary.OperatorToken), "==="),
		rule.RuleFixReplaceRange(utils.TrimNodeTextRange(sourceFile, binary.Right), fmt.Sprintf("'%s'", strings.ToLower(constructorName))),
	}
}

func insertAt(pos int, text string) rule.RuleFix {
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), text)
}

func removeNodeSyntaxAndSpacesBefore(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	if node == nil {
		return nil
	}

	if node.Kind == ast.KindParenthesizedExpression {
		nodeRange := utils.TrimNodeTextRange(sourceFile, node)
		expression := node.AsParenthesizedExpression().Expression
		fixes := make([]rule.RuleFix, 0, 3)
		if text := sourceFile.Text(); nodeRange.Pos() < nodeRange.End() && text[nodeRange.Pos()] == '(' {
			fixes = append(fixes, removeRangeAndSpacesBefore(sourceFile, core.NewTextRange(nodeRange.Pos(), nodeRange.Pos()+1)))
		}
		fixes = append(fixes, removeNodeSyntaxAndSpacesBefore(sourceFile, expression)...)

		closeParenEnd := utils.SkipTrailingWhitespace(sourceFile.Text(), nodeRange.Pos(), nodeRange.End())
		if closeParenEnd > nodeRange.Pos() && sourceFile.Text()[closeParenEnd-1] == ')' {
			fixes = append(fixes, removeRangeAndSpacesBefore(sourceFile, core.NewTextRange(closeParenEnd-1, closeParenEnd)))
		}
		return fixes
	}

	return []rule.RuleFix{removeRangeAndSpacesBefore(sourceFile, utils.TrimNodeTextRange(sourceFile, node))}
}

func removeRangeAndSpacesBefore(sourceFile *ast.SourceFile, textRange core.TextRange) rule.RuleFix {
	text := sourceFile.Text()
	start := textRange.Pos()
	trimmedStart := utils.SkipTrailingWhitespace(text, 0, start)
	return rule.RuleFixReplaceRange(
		textRange.WithPos(trimmedStart),
		firstLineBreak(text[trimmedStart:start]),
	)
}

func firstLineBreak(text string) string {
	for index := range len(text) {
		switch text[index] {
		case '\r':
			if index+1 < len(text) && text[index+1] == '\n' {
				return "\r\n"
			}
			return "\r"
		case '\n':
			return "\n"
		case 0xE2:
			if index+2 < len(text) && text[index+1] == 0x80 {
				switch text[index+2] {
				case 0xA8:
					return "\u2028"
				case 0xA9:
					return "\u2029"
				}
			}
		}
	}
	return ""
}

func messageNoInstanceofBuiltins() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDNoInstanceofBuiltins,
		Description: "Avoid using `instanceof` for type checking as it can lead to unreliable results.",
	}
}

func messageSwitchToTypeOf(constructorName string) rule.RuleMessage {
	typeName := strings.ToLower(constructorName)
	return rule.RuleMessage{
		Id:          messageIDSwitchToTypeOf,
		Description: fmt.Sprintf("Switch to `typeof … === '%s'`.", typeName),
		Data: map[string]string{
			"type": typeName,
		},
	}
}
