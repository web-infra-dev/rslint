package require_array_join_separator

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "require-array-join-separator"

var missingSeparatorMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Missing the separator argument.",
}

func isIdentifierNamed(node *ast.Node, name string) bool {
	return node != nil && ast.IsIdentifier(node) && node.AsIdentifier().Text == name
}

// isArrayPrototypeProperty mirrors unicorn's isArrayPrototypeProperty helper.
// It intentionally accepts only dotted, non-optional member access.
func isArrayPrototypeProperty(node *ast.Node, property string) bool {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return false
	}

	propertyAccess := node.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.QuestionDotToken != nil ||
		!isIdentifierNamed(propertyAccess.Name(), property) {
		return false
	}

	object := ast.SkipParentheses(propertyAccess.Expression)
	if object == nil {
		return false
	}
	if ast.IsEmptyArrayLiteral(object) {
		return true
	}

	if !ast.IsPropertyAccessExpression(object) {
		return false
	}
	prototypeAccess := object.AsPropertyAccessExpression()
	if prototypeAccess == nil || prototypeAccess.QuestionDotToken != nil ||
		!isIdentifierNamed(prototypeAccess.Name(), "prototype") {
		return false
	}

	return isIdentifierNamed(ast.SkipParentheses(prototypeAccess.Expression), "Array")
}

func callParenthesesRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, core.TextRange, core.TextRange, ast.Kind, bool) {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return core.TextRange{}, core.TextRange{}, core.TextRange{}, ast.KindUnknown, false
	}

	scanStart := call.Expression.End()
	if call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0 {
		scanStart = call.TypeArguments.End()
	}

	s := scanner.GetScannerForSourceFile(sourceFile, scanStart)
	var opening core.TextRange
	var previous core.TextRange
	previousKind := ast.KindUnknown
	depth := 0
	foundOpening := false
	for s.TokenStart() < node.End() {
		switch s.Token() {
		case ast.KindOpenParenToken:
			if !foundOpening {
				opening = s.TokenRange()
				foundOpening = true
			}
			depth++
		case ast.KindCloseParenToken:
			if !foundOpening {
				return core.TextRange{}, core.TextRange{}, core.TextRange{}, ast.KindUnknown, false
			}
			depth--
			if depth == 0 {
				return opening, s.TokenRange(), previous, previousKind, true
			}
		default:
			if foundOpening {
				previous = s.TokenRange()
				previousKind = s.Token()
			}
		}
		s.Scan()
	}

	return core.TextRange{}, core.TextRange{}, core.TextRange{}, ast.KindUnknown, false
}

func appendSeparatorFix(closing core.TextRange, penultimateKind ast.Kind, isPrototypeMethod bool) rule.RuleFix {
	separator := "','"
	if isPrototypeMethod {
		if penultimateKind == ast.KindCommaToken {
			// A trailing comma is already present; retain it and add the new
			// argument using the same spacing as unicorn's appendArgument.
			return rule.RuleFixReplaceRange(core.NewTextRange(closing.Pos(), closing.Pos()), " "+separator+",")
		}
		return rule.RuleFixReplaceRange(core.NewTextRange(closing.Pos(), closing.Pos()), ", "+separator)
	}

	return rule.RuleFixReplaceRange(core.NewTextRange(closing.Pos(), closing.Pos()), separator)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/require-array-join-separator.md
var RequireArrayJoinSeparatorRule = rule.Rule{
	Name: "unicorn/require-array-join-separator",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		zeroArguments := 0
		oneArgument := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, directMethod := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "join",
					ArgumentsLength:     &zeroArguments,
					AllowOptionalMember: true,
				})
				if !directMethod {
					call, directMethod = unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
						Method:          "call",
						ArgumentsLength: &oneArgument,
					})
					arguments := node.Arguments()
					if !directMethod || len(arguments) != 1 || ast.IsSpreadElement(arguments[0]) ||
						!isArrayPrototypeProperty(call.Object, "join") {
						return
					}
				}

				opening, closing, penultimate, penultimateKind, ok := callParenthesesRange(ctx.SourceFile, node)
				if !ok {
					return
				}

				isPrototypeMethod := len(node.Arguments()) == 1
				start := opening.Pos()
				if isPrototypeMethod {
					start = penultimate.End()
				}
				reportRange := core.NewTextRange(start, closing.End())
				ctx.ReportRangeWithFixes(
					reportRange,
					missingSeparatorMessage,
					appendSeparatorFix(closing, penultimateKind, isPrototypeMethod),
				)
			},
		}
	},
}
