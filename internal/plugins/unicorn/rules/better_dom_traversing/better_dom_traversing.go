package better_dom_traversing

import (
	"math"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var BetterDomTraversingRule = rule.Rule{
	Name:   "unicorn/better-dom-traversing",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindElementAccessExpression:  func(node *ast.Node) { checkIndexedChild(&ctx, node) },
			ast.KindPropertyAccessExpression: func(node *ast.Node) { checkParentChain(&ctx, node) },
			ast.KindQualifiedName:            func(node *ast.Node) { checkParentChain(&ctx, node) },
			ast.KindCallExpression:           func(node *ast.Node) { checkSelectorChain(&ctx, node) },
		}
	},
}

func checkIndexedChild(ctx *rule.RuleContext, node *ast.Node) {
	rawCollection := node.Expression()
	collection := chainOperand(rawCollection)
	receiver, name := dotMember(collection)
	if name != "childNodes" && name != "children" {
		return
	}
	index, ok := collectionIndex(node)
	if !ok || (name == "childNodes" && index != 0) || unicornutil.IsReferenceTarget(node) {
		return
	}
	if unicornutil.IsNonDOMNodeExpression(receiver) {
		return
	}
	if name == "children" {
		_, receiverName := dotMember(receiver)
		if (receiver != nil && ast.IsIdentifier(receiver) && receiver.Text() == "props") || receiverName == "props" {
			return
		}
	}
	parent := enclosingMember(node)
	if _, name := dotMember(parent); name == "children" {
		if _, ok := collectionIndex(enclosingMember(parent)); ok {
			return
		}
	}

	if index != 0 {
		ctx.ReportNode(node, rule.RuleMessage{
			Id: "query-selector", Description: "Consider `.querySelector()` over positional child traversal.",
		})
		return
	}

	id, replacement := "first-child", "firstChild"
	if name == "children" {
		id, replacement = "first-element-child", "firstElementChild"
	}
	ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
		Id: id, Description: "Prefer `." + replacement + "` over `." + name + "[0]`.",
	}, func() []rule.RuleSuggestion {
		if hasComments(ctx, node) {
			return nil
		}
		indexReplacement := ""
		// Removing brackets can join the new property to an adjacent word
		// operator, as in element.children[0]in nodes.
		if rawCollection == collection {
			if next, ok := utils.TokenAtOrAfter(ctx.SourceFile, node.End()); ok &&
				next.Start == node.End() && !utils.CanTokenTextsBeAdjacent(replacement, next.Text) {
				indexReplacement = " "
			}
		}
		_, property := utils.MemberExpressionParts(collection)
		return []rule.RuleSuggestion{{
			Message: rule.RuleMessage{Id: "suggestion-" + id, Description: "Switch to `." + replacement + "`."},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplace(ctx.SourceFile, property, replacement),
				// Keep any parentheses around the collection.
				rule.RuleFixReplaceRange(core.NewTextRange(rawCollection.End(), node.End()), indexReplacement),
			},
		}}
	})
}

func collectionIndex(node *ast.Node) (float64, bool) {
	if node == nil || !ast.IsElementAccessExpression(node) || ast.IsOptionalChainRoot(node) {
		return 0, false
	}
	index := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
	if index == nil || !ast.IsNumericLiteral(index) {
		return 0, false
	}
	const maxSafeInteger = 1<<53 - 1
	value, ok := ecmascript.StringToNumber(index.Text())
	return value, ok && value >= 0 && value <= maxSafeInteger && math.Trunc(value) == value
}

func checkParentChain(ctx *rule.RuleContext, node *ast.Node) {
	parent, name := dotMember(node)
	if name != "parentElement" {
		return
	}
	root, name := dotMember(parent)
	if name != "parentElement" {
		return
	}
	if _, outer := dotMember(enclosingMember(node)); outer == "parentElement" || utils.IsInJsxTagName(node) {
		return
	}
	for {
		receiver, name := dotMember(root)
		if name != "parentElement" {
			break
		}
		root = receiver
	}
	if !unicornutil.IsNonDOMNodeExpression(root) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id: "closest", Description: "Consider `.closest()` over chaining `.parentElement`.",
		})
	}
}

func checkSelectorChain(ctx *rule.RuleContext, node *ast.Node) {
	call, selector, ok := staticQuerySelector(node)
	if !ok {
		return
	}
	if member := enclosingMember(node); member != nil {
		if next, _, ok := staticQuerySelector(utils.ESTreeParent(member)); ok && next.Callee == member {
			return
		}
	}

	selectors := []string{selector}
	for {
		inner, selector, ok := staticQuerySelector(chainOperand(call.Object))
		if !ok {
			break
		}
		selectors = append(selectors, selector)
		call = inner
	}
	rawRoot := call.Object
	root := chainOperand(rawRoot)
	if len(selectors) < 2 || root == nil || unicornutil.IsNonDOMNodeExpression(root) || isInOptionalChain(node) {
		return
	}

	ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
		Id: "merge-query-selector", Description: "Consider merging chained `.querySelector()` calls.",
	}, func() []rule.RuleSuggestion {
		if hasComments(ctx, node) {
			return nil
		}
		for _, selector := range selectors {
			if strings.Contains(selector, ",") || strings.Contains(selector, ":scope") {
				return nil
			}
		}
		slices.Reverse(selectors)
		merged := strings.Join(selectors, " ")
		if !isDocument(root) {
			merged = ":scope " + merged
		}
		quote := byte('\'')
		firstSelector := utils.ESTreeRuntimeExpression(call.Call.Arguments()[0])
		if ast.IsStringLiteral(firstSelector) && strings.HasPrefix(utils.TrimmedNodeText(ctx.SourceFile, firstSelector), "\"") {
			quote = '"'
		}
		rootText := utils.TrimmedNodeText(ctx.SourceFile, rawRoot)
		if rawRoot == root && unicornutil.ShouldAddParenthesesToMemberExpressionObjectAt(ctx.SourceFile, root, node) {
			rootText = "(" + rootText + ")"
		}
		replacement := rootText + ".querySelector(" + unicornutil.QuoteString(merged, quote) + ")"
		replacement = utils.SafeReplacementText(ctx.SourceFile, node, replacement)
		return []rule.RuleSuggestion{{
			Message:  rule.RuleMessage{Id: "suggestion-merge-query-selector", Description: "Merge the `.querySelector()` calls."},
			FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
		}}
	})
}

func staticQuerySelector(node *ast.Node) (unicornutil.DotMethodCall, string, bool) {
	oneArgument := 1
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method: "querySelector", ArgumentsLength: &oneArgument, RejectSpreadElement: true,
	})
	if !ok {
		return call, "", false
	}
	value, ok := utils.GetStaticStringLiteralValue(utils.ESTreeRuntimeExpression(node.Arguments()[0]))
	return call, value, ok
}

func isDocument(node *ast.Node) bool {
	if ast.IsIdentifier(node) {
		return node.Text() == "document"
	}
	receiver, name := dotMember(node)
	return name == "document" && receiver != nil && ast.IsIdentifier(receiver) &&
		(receiver.Text() == "window" || receiver.Text() == "globalThis")
}

// Parentheses normally disappear in ESTree. Around an optional chain they
// separate the outer receiver from the chain's inner member or call.
func chainOperand(raw *ast.Node) *ast.Node {
	node := utils.ESTreeRuntimeExpression(raw)
	if node != nil && node != raw && ast.IsOptionalChain(node) {
		return nil
	}
	return node
}

func enclosingMember(node *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(node)
	object, _ := utils.MemberExpressionParts(parent)
	if chainOperand(object) == node {
		return parent
	}
	return nil
}

func dotMember(node *ast.Node) (*ast.Node, string) {
	if node == nil || ast.IsElementAccessExpression(node) || ast.IsOptionalChainRoot(node) {
		return nil, ""
	}
	object, name := utils.MemberExpressionParts(node)
	if name != nil && ast.IsIdentifier(name) {
		return chainOperand(object), name.Text()
	}
	return nil, ""
}

func isInOptionalChain(node *ast.Node) bool {
	for {
		if ast.IsOptionalChain(node) {
			return true
		}
		parent := utils.ESTreeParent(node)
		if parent == nil || (!ast.IsAccessExpression(parent) && !ast.IsCallExpression(parent)) ||
			utils.ESTreeRuntimeExpression(parent.Expression()) != node {
			return false
		}
		node = parent
	}
}

func hasComments(ctx *rule.RuleContext, node *ast.Node) bool {
	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()[span.Pos():span.End()]
	// Build the file's comment index only when this expression can contain a comment.
	return (strings.Contains(text, "//") || strings.Contains(text, "/*")) &&
		utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End())
}
