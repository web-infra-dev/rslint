package prefer_at

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var domCollectionTypes = unicornutil.TypeClassifierOptions{
	TargetTypeNames:            utils.NewSetFromItems("NodeList", "NodeListOf", "HTMLCollection", "HTMLCollectionOf"),
	HeritageSymbolFlags:        ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
	AllowNullishInMixedUnion:   true,
	PreferKnownAnnotationNames: true,
}

func isDOMCollection(ctx rule.RuleContext, node *ast.Node) bool {
	// Direct DOM expressions remain excluded even if cast to an array.
	if isDOMExpression(unwrapped(node)) {
		return true
	}
	if class := classifyDOMExpression(ctx, node, make(map[*ast.Symbol]bool)); class != unicornutil.TypeUnknown {
		return class == unicornutil.TypeTarget
	}
	return ctx.TypeChecker != nil &&
		unicornutil.ClassifyType(ctx, ctx.TypeChecker.GetTypeAtLocation(node), domCollectionTypes) == unicornutil.TypeTarget
}

func classifyDOMAnnotation(ctx rule.RuleContext, annotation *ast.Node) unicornutil.TypeClass {
	class := unicornutil.ClassifyTypeAnnotation(ctx, annotation, domCollectionTypes)
	if class == unicornutil.TypeTarget && ctx.TypeChecker != nil && annotation.Parent != nil && !utils.IsJSDocSyntaxNode(annotation) {
		// TypeScript can erase a native intersection to any through an import
		// or a type argument that the source annotation walk cannot resolve.
		resolved := ctx.TypeChecker.GetTypeAtLocation(annotation)
		if utils.IsTypeAnyType(resolved) && !utils.IsIntrinsicErrorType(resolved) {
			return unicornutil.TypeUnknown
		}
	}
	return class
}

func classifyDOMExpression(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Symbol]bool) unicornutil.TypeClass {
	node = ast.SkipParentheses(node)
	if node == nil {
		return unicornutil.TypeUnknown
	}
	switch node.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		if class := classifyDOMAnnotation(ctx, node.Type()); class != unicornutil.TypeUnknown {
			return class
		}
		return classifyDOMExpression(ctx, node.Expression(), visiting)
	case ast.KindNonNullExpression, ast.KindSatisfiesExpression, ast.KindExpressionWithTypeArguments:
		return classifyDOMExpression(ctx, node.Expression(), visiting)
	case ast.KindIdentifier:
		if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 && !visiting[symbol] {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration || declaration.Kind == ast.KindParameter {
				annotation, _ := unicornutil.ValueTypeAnnotation(ctx, declaration)
				if class := classifyDOMAnnotation(ctx, annotation); class != unicornutil.TypeUnknown {
					return class
				}
			}
			if initializer := constInitializer(declaration); initializer != nil {
				visiting[symbol] = true
				class := classifyDOMExpression(ctx, initializer, visiting)
				delete(visiting, symbol)
				return class
			}
		}
	case ast.KindCallExpression:
		if annotation := callReturnAnnotation(ctx, node); annotation != nil {
			return classifyDOMAnnotation(ctx, annotation)
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if classifyDOMExpression(ctx, conditional.WhenTrue, visiting) == unicornutil.TypeTarget &&
			classifyDOMExpression(ctx, conditional.WhenFalse, visiting) == unicornutil.TypeTarget {
			return unicornutil.TypeTarget
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindCommaToken:
			return classifyDOMExpression(ctx, binary.Right, visiting)
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			if classifyDOMExpression(ctx, binary.Left, visiting) == unicornutil.TypeTarget &&
				classifyDOMExpression(ctx, binary.Right, visiting) == unicornutil.TypeTarget {
				return unicornutil.TypeTarget
			}
		}
	}
	if isDOMExpression(node) {
		return unicornutil.TypeTarget
	}
	return unicornutil.TypeUnknown
}

func callReturnAnnotation(ctx rule.RuleContext, call *ast.Node) *ast.Node {
	callee := utils.ESTreeCallCallee(call.Expression())
	if callee == nil || !ast.IsIdentifier(callee) {
		return nil
	}
	symbol := ctx.Refs.ResolveInFile(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference) {
		return nil
	}
	declaration := symbol.Declarations[0]
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
	case ast.KindParameter, ast.KindVariableDeclaration:
		// Upstream requires type services for TypeScript callable bindings.
		// JSDoc function types and directly declared function results also
		// work with the syntax-only parser.
		if annotation := declaration.Type(); annotation != nil && !utils.IsJSDocSyntaxNode(annotation) {
			return nil
		}
		if initializer := declaration.Initializer(); ast.IsVariableDeclaration(declaration) && initializer != nil {
			initializer = ast.SkipOuterExpressions(initializer, ast.OEKParentheses|ast.OEKSatisfies|ast.OEKNonNullAssertions)
			if function := utils.ESTreeRuntimeExpression(initializer); ast.IsFunctionLike(function) {
				declaration = function
			}
		}
	default:
		return nil
	}
	if ast.IsFunctionLike(declaration) && (len(declaration.TypeParameters()) != 0 ||
		ast.GetFunctionFlags(declaration)&ast.FunctionFlagsAsyncGenerator != 0) {
		return nil
	}
	return unicornutil.CallableReturnTypeAnnotation(ctx, declaration)
}

func isDOMExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsPropertyAccessExpression(node) && ast.IsIdentifier(node.Name()) {
		switch node.Name().Text() {
		case "children", "childNodes":
			return true
		}
	}
	_, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{
			"getElementsByClassName", "getElementsByName", "getElementsByTagName", "getElementsByTagNameNS", "querySelectorAll",
		},
		AllowOptionalCall: true, AllowOptionalMember: true,
	})
	return ok
}

func isObviouslyUnsupportedReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	node = unwrapped(node)
	if node != nil && ast.IsIdentifier(node) {
		if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
			node = unwrapped(constInitializer(symbol.Declarations[0]))
		}
	}
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression,
		ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

func constInitializer(declaration *ast.Node) *ast.Node {
	if declaration == nil || !ast.IsVariableDeclaration(declaration) || !ast.IsVarConst(declaration) || !ast.IsIdentifier(declaration.Name()) {
		return nil
	}
	return declaration.AsVariableDeclaration().Initializer
}
