// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package prefer_boolean_return

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferBooleanReturnRule = rule.Rule{
	Name:   "unicorn/prefer-boolean-return",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindIfStatement: func(node *ast.Node) {
			statement := node.AsIfStatement()
			test := utils.ESTreeRuntimeExpression(statement.Expression)
			if test == nil || test.Kind == ast.KindConditionalExpression {
				return
			}
			consequent := returnedBoolean(unwrapBranch(statement.ThenStatement))
			if consequent == ast.KindUnknown {
				return
			}

			alternate := unwrapBranch(statement.ElseStatement)
			end := node.End()
			if statement.ElseStatement == nil {
				parent := node.Parent
				if parent == nil || !parent.CanHaveStatements() {
					return
				}
				siblings := parent.Statements()
				index := ast.IndexOfNode(siblings, node)
				if index < 0 || index+1 == len(siblings) || siblings[index] != node {
					return
				}
				alternate = siblings[index+1]
				if !ast.IsReturnStatement(alternate) {
					return
				}
				if index > 0 && returnsSameBoolean(siblings[index-1], consequent) {
					return
				}
				end = alternate.End()
			}
			if value := returnedBoolean(alternate); value == ast.KindUnknown || value == consequent {
				return
			}

			diagnosticRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			ctx.ReportRangeWithDeferredFixes(diagnosticRange, rule.RuleMessage{
				Id:          "prefer-boolean-return",
				Description: "Return the boolean expression directly.",
			}, func() []rule.RuleFix {
				return fixReturn(ctx, node, test, alternate.End(), diagnosticRange.WithEnd(end), consequent == ast.KindFalseKeyword)
			})
		}}
	},
}

// unwrapBranch follows single-statement blocks, ignoring stray semicolons.
func unwrapBranch(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindBlock {
		var only *ast.Node
		for _, statement := range node.Statements() {
			if statement.Kind == ast.KindEmptyStatement {
				continue
			}
			if only != nil {
				return nil
			}
			only = statement
		}
		node = only
	}
	return node
}

func returnedBoolean(node *ast.Node) ast.Kind {
	if node == nil || node.Kind != ast.KindReturnStatement || node.Expression() == nil {
		return ast.KindUnknown
	}
	value := ast.SkipOuterExpressions(node.Expression(), ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
	if value != nil && ast.IsBooleanLiteral(value) {
		return value.Kind
	}
	return ast.KindUnknown
}

func returnsSameBoolean(node *ast.Node, value ast.Kind) bool {
	if node.Kind != ast.KindIfStatement {
		return false
	}
	statement := node.AsIfStatement()
	return statement.ElseStatement == nil && returnedBoolean(unwrapBranch(statement.ThenStatement)) == value
}

func isRuntimeBoolean(ctx rule.RuleContext, node *ast.Node, trustGlobals bool) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindDeleteExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return unicornutil.IsLogicalNot(node)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if unicornutil.IsComparisonOperator(operator) || utils.IsLooseEqualityOperator(operator) {
			return true
		}
		switch operator {
		case ast.KindInKeyword, ast.KindInstanceOfKeyword:
			return true
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return isRuntimeBoolean(ctx, binary.Left, trustGlobals) && isRuntimeBoolean(ctx, binary.Right, trustGlobals)
		}
	case ast.KindCallExpression:
		if !trustGlobals || ast.IsOptionalChain(node) {
			return false
		}
		if ctx.Globals.Access("Boolean") != utils.GlobalAccessOff && unicornutil.IsGlobalBooleanCall(ctx, node) {
			return true
		}
		if call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: "isArray"}); ok {
			object := utils.ESTreeRuntimeExpression(call.Object)
			return object != nil && ast.IsIdentifier(object) && object.Text() == "Array" &&
				ctx.Globals.Access("Array").IsDeclared() && unicornutil.IsGlobalReference(ctx, object)
		}
	}
	return false
}

func fixReturn(ctx rule.RuleContext, node, test *ast.Node, alternateEnd int, replacement core.TextRange, negate bool) []rule.RuleFix {
	sf := ctx.SourceFile
	condition := utils.TrimNodeTextRange(sf, node.AsIfStatement().Expression)
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, replacement.Pos(), condition.Pos()) ||
		utils.HasCommentInSpan(comments, condition.End(), replacement.End()) ||
		hasTrailingComment(sf, comments, alternateEnd) {
		return nil
	}

	// A with object can provide its own Boolean or Array even when there is no
	// lexical binding for either name.
	trustGlobals := ctx.LanguageOptions.EffectiveSourceType() == "module" || !insideWithBody(node)
	text := sf.Text()[condition.Pos():condition.End()]
	if negate || !isRuntimeBoolean(ctx, test, trustGlobals) {
		if !negate && (!trustGlobals || !booleanNameAvailable(ctx, node)) {
			return nil
		}
		inner := utils.TrimNodeTextRange(sf, test)
		if !utils.HasCommentInSpan(comments, condition.Pos(), inner.Pos()) &&
			!utils.HasCommentInSpan(comments, inner.End(), condition.End()) {
			text = sf.Text()[inner.Pos():inner.End()]
			if (negate && negationNeedsParens(test)) || (!negate && utils.IsCommaOperator(test)) {
				text = "(" + text + ")"
			}
		}
		if negate {
			text = "!" + text
		} else {
			text = "Boolean(" + text + ")"
		}
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(replacement, "return "+text+";")}
}

func hasTrailingComment(sf *ast.SourceFile, comments []*ast.CommentRange, end int) bool {
	after := utils.CommentsInSpan(comments, end, len(sf.Text()))
	if len(after) == 0 || !utils.IsSameLine(sf, end, after[0].Pos()) {
		return false
	}
	return scanner.SkipTriviaEx(sf.Text(), end, &scanner.SkipTriviaOptions{StopAtComments: true}) == after[0].Pos()
}

func negationNeedsParens(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindTypeAssertionExpression, ast.KindPostfixUnaryExpression:
		return true
	case ast.KindNonNullExpression:
		return !ast.IsOptionalChain(ast.SkipOuterExpressions(node, ast.OEKNonNullAssertions))
	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
	default:
		return ast.GetExpressionPrecedence(node) < ast.OperatorPrecedenceUnary
	}
}

func insideWithBody(node *ast.Node) bool {
	return ast.FindAncestor(node.Parent, func(ancestor *ast.Node) bool {
		return ast.IsWithStatement(ancestor) && ancestor.Statement() != nil && ancestor.Statement().Loc.Contains(node.Pos())
	}) != nil
}

func booleanNameAvailable(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Globals.Access("Boolean") == utils.GlobalAccessOff {
		return false
	}
	if ctx.Refs == nil {
		return !utils.IsShadowed(node, "Boolean")
	}
	// An import type can also make an inserted Boolean(...) fail type checking.
	return !ctx.Refs.IsNameDefinedInFileWithMeaning(node, "Boolean", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
}
