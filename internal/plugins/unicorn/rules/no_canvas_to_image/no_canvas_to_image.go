// Ported from eslint-plugin-unicorn v80.0.0. See LICENSE.
package no_canvas_to_image

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var NoCanvasToImageRule = rule.Rule{
	Name:   "unicorn/no-canvas-to-image",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		oneArgument := 1
		singleReferences := make(map[*ast.Symbol]*ast.Node)
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Methods: []string{"drawImage", "putImageData"}, MinimumArguments: &oneArgument,
			})
			if !ok || node.Arguments()[0].Kind == ast.KindSpreadElement {
				return
			}
			argument := node.Arguments()[0]
			expression := directExpressionOrOneUseConst(ctx, argument, singleReferences)
			if expression == nil {
				return
			}
			if call.Property.Text() == "drawImage" {
				if !isCanvasImage(expression) {
					return
				}
			} else if _, ok := unicornutil.MatchDotMethodCall(expression, unicornutil.DotMethodCallOptions{
				Method: "getImageData",
			}); !ok {
				return
			}
			ctx.ReportNode(utils.ESTreeRuntimeExpression(argument), rule.RuleMessage{
				Id:          "no-canvas-to-image",
				Description: "Prefer drawing the canvas directly with `drawImage()`.",
			})
		}}
	},
}

func withoutAwait(node *ast.Node) *ast.Node {
	node = utils.ESTreeRuntimeExpression(node)
	for node != nil && node.Kind == ast.KindAwaitExpression {
		node = utils.ESTreeRuntimeExpression(node.Expression())
	}
	if node != nil && ast.IsOptionalChain(node) {
		return nil
	}
	return node
}

func directExpressionOrOneUseConst(ctx rule.RuleContext, node *ast.Node, singleReferences map[*ast.Symbol]*ast.Node) *ast.Node {
	node = withoutAwait(node)
	if node == nil {
		return nil
	}
	if !ast.IsIdentifier(node) {
		return node
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if !ast.IsVariableDeclaration(declaration) || !ast.IsIdentifier(declaration.Name()) ||
		!ast.IsVarConst(declaration) || declaration.Name().Pos() > node.Pos() {
		return nil
	}
	// RefStore excludes the declaration's initializing write. Upstream counts
	// that write and the single use as two references.
	// Cache the sole non-JSDoc reference, or nil for a multi-use binding.
	// RefStore owns the slice; stop scanning as soon as a second use is found.
	singleReference, cached := singleReferences[symbol]
	if !cached {
		for _, reference := range ctx.Refs.References(symbol) {
			if isJSDocReference(reference) {
				continue
			}
			if singleReference != nil {
				singleReference = nil
				break
			}
			singleReference = reference
		}
		singleReferences[symbol] = singleReference
	}
	if singleReference != node {
		return nil
	}
	return withoutAwait(declaration.AsVariableDeclaration().Initializer)
}

func isJSDocReference(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if utils.IsJSDocSyntaxNode(current) {
			return true
		}
	}
	return false
}

func isCanvasImage(node *ast.Node) bool {
	if !ast.IsCallExpression(node) || ast.IsOptionalChainRoot(node) || len(node.Arguments()) == 0 ||
		node.Arguments()[0].Kind == ast.KindSpreadElement {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "loadImage" {
		return false
	}
	url := utils.ESTreeRuntimeExpression(node.Arguments()[0])
	if url == nil || ast.IsOptionalChain(url) {
		return false
	}
	call, ok := unicornutil.MatchDotMethodCall(url, unicornutil.DotMethodCallOptions{
		Method: "toDataURL",
	})
	if !ok {
		return false
	}
	receiver := utils.ESTreeCallCallee(call.Object)
	if receiver == nil {
		return false
	}
	if ast.IsPropertyAccessExpression(receiver) {
		receiver = receiver.Name()
	}
	return ast.IsIdentifier(receiver) && strings.Contains(ecmascript.StringToLowerCase(receiver.Text()), "canvas")
}
