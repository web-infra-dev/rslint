package no_return_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildRedundantUseOfAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redundantUseOfAwait",
		Description: "Redundant use of `await` on a return value.",
	}
}

func buildRemoveAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAwait",
		Description: "Remove redundant `await`.",
	}
}

// effectiveParent returns the nearest ancestor of node that is not a
// ParenthesizedExpression, together with the child that ancestor actually sees.
// ESTree drops parentheses entirely, so upstream's `node.parent` checks compare
// against the unwrapped expression; tsgo keeps them as real nodes.
func effectiveParent(node *ast.Node) (child *ast.Node, parent *ast.Node) {
	child = utils.OutermostParenthesizedExpression(node)
	return child, child.Parent
}

// hasErrorHandler determines whether a thrown error from this node will be
// caught/handled within this function rather than immediately halting it. A
// statement in a `try` block always has an error handler; a statement in a
// `catch` clause only has one when a `finally` block follows.
func hasErrorHandler(node *ast.Node) bool {
	for ancestor := node; ancestor != nil && !ast.IsFunctionOrSourceFile(ancestor); ancestor = ancestor.Parent {
		parent := ancestor.Parent
		if parent == nil || !ast.IsTryStatement(parent) {
			continue
		}

		try := parent.AsTryStatement()
		if ancestor == try.TryBlock {
			return true
		}
		if ancestor == try.CatchClause && try.FinallyBlock != nil {
			return true
		}
	}
	return false
}

// isInTailCallPosition reports whether the value of node is the value the
// enclosing function resolves with. `return` arguments (and concise arrow
// bodies) can be complex expressions, so an `await` is only redundant when it
// sits on the branch that actually produces the returned value.
func isInTailCallPosition(node *ast.Node) bool {
	child, parent := effectiveParent(node)
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindArrowFunction:
		return parent.Body() == child
	case ast.KindReturnStatement:
		return !hasErrorHandler(parent)
	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		if conditional.WhenTrue == child || conditional.WhenFalse == child {
			return isInTailCallPosition(parent)
		}
	case ast.KindBinaryExpression:
		// ESTree's LogicalExpression and SequenceExpression both collapse into
		// BinaryExpression here, so branch on the operator. A comma chain is
		// left-associative, which makes "is the last expression" equivalent to
		// "is the right operand".
		binary := parent.AsBinaryExpression()
		if binary.Right != child {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken,
			ast.KindBarBarToken,
			ast.KindQuestionQuestionToken,
			ast.KindCommaToken:
			return isInTailCallPosition(parent)
		}
	}

	return false
}

// https://eslint.org/docs/latest/rules/no-return-await
var NoReturnAwaitRule = rule.Rule{
	Name:   "no-return-await",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceFile := ctx.SourceFile

		buildRemoveAwaitSuggestion := func(node *ast.Node) []rule.RuleSuggestion {
			text := sourceFile.Text()
			awaitRange := scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos())

			// Upstream compares the first two tokens of the await expression;
			// removing `await` across a line break would join the two lines.
			lineMap := sourceFile.ECMALineMap()
			awaitLine := scanner.ComputeLineOfPosition(lineMap, awaitRange.Pos())
			operandLine := scanner.ComputeLineOfPosition(lineMap, scanner.SkipTrivia(text, awaitRange.End()))
			if awaitLine != operandLine {
				return nil
			}

			// A single space right after the keyword is consumed along with it,
			// so `return await bar()` becomes `return bar()` rather than
			// `return  bar()`.
			end := awaitRange.End()
			if end < len(text) && text[end] == ' ' {
				end++
			}

			return []rule.RuleSuggestion{{
				Message:  buildRemoveAwaitMessage(),
				FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(awaitRange.Pos(), end))},
			}}
		}

		return rule.RuleListeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				if !isInTailCallPosition(node) || hasErrorHandler(node) {
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(node, buildRedundantUseOfAwaitMessage(), func() []rule.RuleSuggestion {
					return buildRemoveAwaitSuggestion(node)
				})
			},
		}
	},
}
