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

func directCallParenthesesRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, core.TextRange, bool) {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return core.TextRange{}, core.TextRange{}, false
	}

	scanStart := call.Expression.End()
	if call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0 {
		scanStart = call.TypeArguments.End()
	}

	s := scanner.GetScannerForSourceFile(sourceFile, scanStart)
	for s.Token() != ast.KindOpenParenToken && s.TokenStart() < node.End() {
		s.Scan()
	}
	if s.Token() != ast.KindOpenParenToken {
		return core.TextRange{}, core.TextRange{}, false
	}
	opening := s.TokenRange()
	s.Scan()
	if s.Token() != ast.KindCloseParenToken || s.TokenEnd() != node.End() {
		return core.TextRange{}, core.TextRange{}, false
	}

	return opening, s.TokenRange(), true
}

func prototypeCallSuffix(sourceFile *ast.SourceFile, node *ast.Node, argument *ast.Node) (core.TextRange, int, ast.Kind, bool) {
	if argument == nil {
		return core.TextRange{}, 0, ast.KindUnknown, false
	}

	// Start after the argument so regex literals and nested calls inside it are
	// never re-scanned as ordinary tokens. Only trivia, an optional trailing
	// comma, and the call's closing parenthesis can occur in this suffix.
	s := scanner.GetScannerForSourceFile(sourceFile, argument.End())
	penultimateEnd := argument.End()
	penultimateKind := ast.KindUnknown
	if s.Token() == ast.KindCommaToken {
		penultimateEnd = s.TokenEnd()
		penultimateKind = ast.KindCommaToken
		s.Scan()
	}
	if s.Token() != ast.KindCloseParenToken || s.TokenEnd() != node.End() {
		return core.TextRange{}, 0, ast.KindUnknown, false
	}

	return s.TokenRange(), penultimateEnd, penultimateKind, true
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
				var prototypeArgument *ast.Node
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
					if !directMethod || ast.IsSpreadElement(arguments[0]) ||
						!isArrayPrototypeProperty(call.Object, "join") {
						return
					}
					prototypeArgument = arguments[0]
				}

				isPrototypeMethod := prototypeArgument != nil
				var closing core.TextRange
				var start int
				penultimateKind := ast.KindUnknown
				if isPrototypeMethod {
					var ok bool
					closing, start, penultimateKind, ok = prototypeCallSuffix(ctx.SourceFile, node, prototypeArgument)
					if !ok {
						return
					}
				} else {
					opening, directClosing, ok := directCallParenthesesRange(ctx.SourceFile, node)
					if !ok {
						return
					}
					closing = directClosing
					start = opening.Pos()
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
