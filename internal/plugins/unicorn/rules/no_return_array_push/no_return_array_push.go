package no_return_array_push

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoReturnArrayPushRule = rule.Rule{
	Name:   "unicorn/no-return-array-push",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		minimumArguments := 1
		methodOptions := unicornutil.DotMethodCallOptions{
			Methods:             []string{"push", "unshift"},
			MinimumArguments:    &minimumArguments,
			AllowOptionalCall:   true,
			AllowOptionalMember: true,
		}
		receiverOptions := unicornutil.ArrayReceiverOptions{CheckArrayProducers: true, ReadJSDoc: true}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, methodOptions)
				if !ok || isDiscardedOrChained(node) {
					return
				}
				method := call.Property.Text()
				if method == "push" && isIgnoredPushReceiver(call.Object) &&
					!unicornutil.IsArrayWithOptions(ctx, call.Object, receiverOptions) {
					return
				}
				if unicornutil.IsKnownNonIndexedCollectionWithOptions(ctx, call.Object, receiverOptions) {
					return
				}

				message := rule.RuleMessage{
					Id:          "no-return-array-push/error",
					Description: "Do not use the return value of `." + method + "(…)`.",
				}
				statement := utils.ESTreeParent(node)
				// An optional call has a ChainExpression between it and return in ESTree.
				if !ast.IsOptionalChain(node) && statement != nil && ast.IsReturnStatement(statement) && ast.IsBlock(statement.Parent) {
					ctx.ReportNodeWithDeferredSuggestions(call.Property, message, func() []rule.RuleSuggestion {
						return separateReturn(ctx, statement, node, method)
					})
					return
				}
				ctx.ReportNode(call.Property, message)
			},
		}
	},
}

func isDiscardedOrChained(node *ast.Node) bool {
	// Parentheses and TypeScript wrappers preserve how the result is used.
	const wrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments
	for parent := node.Parent; parent != nil && ast.IsOuterExpression(parent, wrappers) && parent.Expression() == node; parent = node.Parent {
		node = parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		return true
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// Chaining a member off the result is common for custom push APIs.
		return parent.Expression() == node
	default:
		return false
	}
}

func isIgnoredPushReceiver(node *ast.Node) bool {
	// A parenthesized optional chain is a distinct receiver in ESTree.
	node = utils.ESTreeCallCallee(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text() == "stream" || node.Text() == "router"
	case ast.KindThisKeyword:
		return true
	case ast.KindPropertyAccessExpression:
		root := utils.ESTreeCallCallee(node.Expression())
		name := node.Name()
		if root == nil || name == nil || !ast.IsIdentifier(name) {
			return false
		}
		if root.Kind == ast.KindThisKeyword {
			return name.Text() == "stream" || name.Text() == "router" || name.Text() == "$router"
		}
		return ast.IsIdentifier(root) && root.Text() == "process" &&
			(name.Text() == "stdin" || name.Text() == "stdout" || name.Text() == "stderr")
	default:
		return false
	}
}

func separateReturn(ctx rule.RuleContext, statement, call *ast.Node, method string) []rule.RuleSuggestion {
	source := ctx.SourceFile
	statementRange := utils.TrimNodeTextRange(source, statement)
	callRange := utils.TrimNodeTextRange(source, call)
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, statementRange.Pos(), callRange.Pos()) ||
		utils.HasCommentInSpan(comments, callRange.End(), statementRange.End()) {
		return nil
	}

	callText := source.Text()[callRange.Pos():callRange.End()]
	if unicornutil.StartsWithDeclarationLikeToken(source, call) {
		callText = "(" + callText + ")"
	}
	if previous, ok := utils.TokenBeforePosition(source, statementRange.Pos()); ok &&
		unicornutil.NeedsSemicolonAfter(source, previous, callText, false) {
		callText = ";" + callText
	}
	return []rule.RuleSuggestion{
		{
			Message: rule.RuleMessage{
				Id:          "no-return-array-push/suggestion",
				Description: "Separate the `" + method + "()` call from `return`.",
				Data:        map[string]string{"method": method},
			},
			FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(statementRange, callText+"; return;")},
		},
	}
}
