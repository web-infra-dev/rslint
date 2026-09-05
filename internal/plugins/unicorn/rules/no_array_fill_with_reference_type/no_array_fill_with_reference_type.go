package no_array_fill_with_reference_type

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-array-fill-with-reference-type"

var referenceFillValueMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not use a reference value as the fill value.",
}

// NoArrayFillWithReferenceTypeRule mirrors eslint-plugin-unicorn's
// no-array-fill-with-reference-type rule. It is syntactic-first: reporting is
// decided by the shape of the fill value, and the type checker is only
// consulted (when available) to skip receivers that are known non-arrays. On JS
// / gap files (`ctx.TypeChecker == nil`) that check degrades to "unknown",
// exactly like upstream when no parser services are present — so the rule stays
// useful and MUST NOT declare RequiresTypeInfo. `.fill()` called with an object
// literal, array literal, class expression or `new` expression has no
// non-array receiver in the platform to confuse it with: typed arrays take
// numbers and `CanvasRenderingContext2D#fill()` takes no reference value.
var NoArrayFillWithReferenceTypeRule = rule.Rule{
	Name:   "unicorn/no-array-fill-with-reference-type",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "fill",
					MinimumArguments:    &oneArgument,
					AllowOptionalCall:   false,
					AllowOptionalMember: false,
				})
				if !ok {
					return
				}

				fillValue := node.Arguments()[0]
				if !isReferenceFillValue(ctx, fillValue) ||
					unicornutil.IsKnownNonArray(ctx, call.Object) {
					return
				}

				// ESTree does not preserve parentheses, so report the inner
				// expression while retaining TypeScript assertion wrappers.
				ctx.ReportNode(ast.SkipParentheses(fillValue), referenceFillValueMessage)
			},
		}
	},
}

func isReferenceFillValue(ctx rule.RuleContext, node *ast.Node) bool {
	if isReferenceExpression(ctx, node) {
		return true
	}
	return isReferenceExpression(ctx, getConstVariableInitializer(ctx, node))
}

func isReferenceExpression(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindClassExpression:
		return true
	case ast.KindNewExpression:
		return !isGlobalRegExpConstruction(ctx, node)
	default:
		return false
	}
}

func isGlobalRegExpConstruction(ctx rule.RuleContext, node *ast.Node) bool {
	newExpression := node.AsNewExpression()
	if newExpression == nil {
		return false
	}
	callee := ast.SkipParentheses(newExpression.Expression)
	if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "RegExp" ||
		!ctx.Globals.Access("RegExp").IsDeclared() {
		return false
	}

	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(callee); symbol != nil {
			return !utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return !utils.IsShadowed(callee, "RegExp")
}

// getConstVariableInitializer mirrors the rule-local upstream helper rather
// than the broader shared utility: only a plain identifier with exactly one
// direct const VariableDeclaration definition is followed, and aliases are not
// chased recursively.
func getConstVariableInitializer(ctx rule.RuleContext, node *ast.Node) *ast.Node {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || !ast.IsIdentifier(node) || ctx.Refs == nil {
		return nil
	}

	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || !ast.IsVariableDeclaration(declaration) {
		return nil
	}

	declarationList := declaration.Parent
	if declarationList == nil || !ast.IsVariableDeclarationList(declarationList) ||
		declarationList.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	variableDeclaration := declaration.AsVariableDeclaration()
	name := variableDeclaration.Name()
	if name == nil || !ast.IsIdentifier(name) ||
		name.AsIdentifier().Text != node.AsIdentifier().Text {
		return nil
	}

	return utils.SkipAssertionsAndParens(variableDeclaration.Initializer)
}
