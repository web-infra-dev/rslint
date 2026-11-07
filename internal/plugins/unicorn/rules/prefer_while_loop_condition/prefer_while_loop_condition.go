// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package prefer_while_loop_condition

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var PreferWhileLoopConditionRule = rule.Rule{
	Name:   "unicorn/prefer-while-loop-condition",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(loop *ast.Node) {
			body, test := infiniteLoopParts(loop)
			if body == nil || body.Kind != ast.KindBlock || loop.Parent.Kind == ast.KindLabeledStatement {
				return
			}
			statements := body.AsBlock().Statements.Nodes
			if len(statements) == 0 || statements[0].Kind != ast.KindIfStatement {
				return
			}
			first := statements[0]
			guard := first.AsIfStatement()
			if guard.ElseStatement != nil || !onlyUnlabeledBreak(guard.ThenStatement) ||
				slices.ContainsFunc(statements[1:], hasUnlabeledBreak) || hasUnsafeLiftReference(ctx, guard.Expression, body) {
				return
			}

			ctx.ReportNodeWithDeferredFixes(first, rule.RuleMessage{
				Id:          "prefer-while-loop-condition",
				Description: "Prefer putting the condition in the `while` statement.",
			}, func() []rule.RuleFix {
				if hasDirectEval(guard.Expression) || hasUnpreservedComment(ctx, loop, body, first) {
					return nil
				}
				condition := loopConditionText(ctx, guard.Expression)
				var fixes []rule.RuleFix
				switch loop.Kind {
				case ast.KindWhileStatement:
					fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, utils.ESTreeRuntimeExpression(test), condition))
				case ast.KindForStatement:
					fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(
						utils.TrimNodeTextRange(ctx.SourceFile, loop).Pos(), utils.TrimNodeTextRange(ctx.SourceFile, body).Pos(),
					), "while ("+condition+") "))
				case ast.KindDoStatement:
					start := utils.TrimNodeTextRange(ctx.SourceFile, loop).Pos()
					fixes = append(fixes,
						rule.RuleFixReplaceRange(core.NewTextRange(start, start+len("do")), "while ("+condition+")"),
						rule.RuleFixRemoveRange(core.NewTextRange(body.End(), loop.End())),
					)
				}
				return append(fixes, rule.RuleFixRemoveRange(unicornutil.StatementRemovalRange(
					ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, first),
				)))
			})
		}
		return rule.RuleListeners{
			ast.KindWhileStatement: check,
			ast.KindForStatement:   check,
			ast.KindDoStatement:    check,
		}
	},
}

func infiniteLoopParts(loop *ast.Node) (body, test *ast.Node) {
	if loop.Kind == ast.KindForStatement {
		statement := loop.AsForStatement()
		if statement.Initializer != nil || statement.Incrementor != nil {
			return nil, nil
		}
		test = statement.Condition
	} else {
		test = loop.Expression()
	}
	if test != nil && utils.ESTreeRuntimeExpression(test).Kind != ast.KindTrueKeyword {
		return nil, nil
	}
	return loop.Statement(), test
}

func onlyUnlabeledBreak(node *ast.Node) bool {
	if node.Kind == ast.KindBlock {
		statements := node.AsBlock().Statements.Nodes
		if len(statements) != 1 {
			return false
		}
		node = statements[0]
	}
	return node.Kind == ast.KindBreakStatement && node.AsBreakStatement().Label == nil
}

func hasUnlabeledBreak(node *ast.Node) bool {
	if node.Kind == ast.KindBreakStatement {
		return node.AsBreakStatement().Label == nil
	}
	if ast.IsFunctionLikeDeclaration(node) || ast.IsIterationStatement(node, false) || node.Kind == ast.KindSwitchStatement {
		return false
	}
	return node.ForEachChild(hasUnlabeledBreak)
}

func hasUnsafeLiftReference(ctx rule.RuleContext, test, body *ast.Node) bool {
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			if symbol := ctx.Refs.ResolveInFile(node); symbol != nil {
				for _, declaration := range symbol.Declarations {
					root := ast.GetRootDeclaration(declaration)
					if root.Kind == ast.KindVariableDeclaration && !ast.IsBlockOrCatchScoped(declaration) {
						continue
					}
					if name := declaration.Name(); name != nil && name.Pos() >= bodyRange.Pos() && name.End() <= bodyRange.End() {
						return true
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	return visit(test)
}

func hasDirectEval(node *ast.Node) bool {
	if ast.IsCallExpression(node) && node.AsCallExpression().QuestionDotToken == nil {
		callee := ast.SkipOuterExpressions(node.Expression(), ast.OEKAll)
		// Even a local binding named eval can hold the intrinsic eval function.
		if ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "eval" {
			return true
		}
	}
	return node.ForEachChild(hasDirectEval)
}

func loopConditionText(ctx rule.RuleContext, test *ast.Node) string {
	test = utils.ESTreeRuntimeExpression(test)
	if unicornutil.IsLogicalNot(test) {
		operand := utils.ESTreeRuntimeExpression(test.AsPrefixUnaryExpression().Operand)
		textRange := utils.TrimNodeTextRange(ctx.SourceFile, operand)
		return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	}
	text := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, test).Pos():test.End()]
	needsParentheses := ast.GetExpressionPrecedence(test) < ast.OperatorPrecedenceUnary
	switch test.Kind {
	case ast.KindPostfixUnaryExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
		needsParentheses = true
	case ast.KindPrefixUnaryExpression:
		operator := test.AsPrefixUnaryExpression().Operator
		needsParentheses = operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
	}
	if needsParentheses {
		text = "(" + text + ")"
	}
	return "!" + text
}

func hasUnpreservedComment(ctx rule.RuleContext, loop, body, first *ast.Node) bool {
	comments := ctx.Comments.All()
	if len(comments) == 0 {
		return false
	}
	start := utils.TrimNodeTextRange(ctx.SourceFile, loop).Pos()
	if utils.HasCommentInSpan(comments, start, first.End()) {
		return true
	}
	preservedStart := trailingCommentEnd(ctx.SourceFile, comments, first.End())
	end := loop.End()
	if loop.Kind == ast.KindDoStatement {
		end = trailingCommentEnd(ctx.SourceFile, comments, end)
	}
	for _, comment := range utils.CommentsInSpan(comments, first.End(), end) {
		if comment.Pos() < preservedStart || comment.End() > body.End() {
			return true
		}
	}
	return false
}

func trailingCommentEnd(sourceFile *ast.SourceFile, comments []*ast.CommentRange, end int) int {
	next, ok := utils.TokenAtOrAfter(sourceFile, end)
	limit := len(sourceFile.Text())
	if ok {
		limit = next.Start
	}
	last := end
	for _, comment := range utils.CommentsInSpan(comments, end, limit) {
		if ecmascript.ContainsLineTerminator(sourceFile.Text(), end, comment.Pos()) {
			break
		}
		last = comment.End()
	}
	return last
}
