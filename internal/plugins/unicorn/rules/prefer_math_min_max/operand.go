package prefer_math_min_max

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func skipOperand(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, operand *ast.Node) bool {
	node := utils.ESTreeRuntimeExpression(operand)
	if node == nil {
		return false
	}
	if isDateConstruction(node) {
		return true
	}
	// Check the outer authored cast, if any. Non-null assertions do not
	// contribute a type annotation.
	cast := ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKNonNullAssertions)
	if ast.IsAsExpression(cast) || ast.IsTypeAssertion(cast) {
		annotation := ast.SkipTypeParentheses(cast.Type())
		boxedNumber := annotation.Kind == ast.KindTypeReference &&
			unicornutil.NodeMatchesPath(annotation.AsTypeReferenceNode().TypeName, "Number")
		if annotation.Kind != ast.KindNumberKeyword && !boxedNumber {
			return true
		}
	}

	binding := ctx.Refs.ResolveInFile(unwrapOperand(node))
	if binding == nil {
		return unicornutil.IsBigInt(ctx, evaluator, operand)
	}
	for _, definition := range binding.Declarations {
		// ESTree's variable definition points to the entire declarator even
		// when the identifier is part of a destructuring pattern.
		declaration := ast.GetRootDeclaration(definition)
		switch declaration.Kind {
		case ast.KindVariableDeclaration:
			if ast.IsCatchClauseVariableDeclarationOrBindingElement(declaration) {
				continue
			}
		case ast.KindParameter:
			// The type of a destructured or rest parameter belongs to the
			// pattern, while a destructured name can have its own default.
			declaration = definition
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				continue
			}
		default:
			continue
		}
		annotation := declaration.Type()
		if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) && ast.SkipTypeParentheses(annotation).Kind != ast.KindNumberKeyword {
			return true
		}
		initializer := utils.ESTreeRuntimeExpression(declaration.Initializer())
		if nonNumberLiteral(initializer) || unicornutil.IsBigInt(ctx, evaluator, initializer) ||
			(declaration.Kind == ast.KindVariableDeclaration && isDateConstruction(initializer)) {
			return true
		}
	}
	return unicornutil.IsBigInt(ctx, evaluator, operand)
}

func isDateConstruction(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindNewExpression &&
		unicornutil.NodeMatchesPath(node.Expression(), "Date")
}

func nonNumberLiteral(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
		(utils.IsESTreeLiteralKind(node.Kind) && node.Kind != ast.KindNumericLiteral))
}
