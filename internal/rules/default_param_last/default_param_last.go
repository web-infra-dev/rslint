package default_param_last

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func shouldBeLastMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "shouldBeLast",
		Description: "Default parameters should be last.",
	}
}

// isOptionalParam checks if node is optional parameter
func isOptionalParam(node *ast.Node) bool {
	if node == nil || !ast.IsParameter(node) {
		return false
	}

	param := node.AsParameterDeclaration()
	return param.QuestionToken != nil
}

// isDefaultParam checks if node is a parameter with default value
func isDefaultParam(node *ast.Node) bool {
	if node == nil || !ast.IsParameter(node) {
		return false
	}
	
	param := node.AsParameterDeclaration()
	return param.Initializer != nil
}

// isRestParam checks if node is a rest parameter
func isRestParam(node *ast.Node) bool {
	if node == nil || !ast.IsParameter(node) {
		return false
	}
	return utils.IsRestParameterDeclaration(node)
}

// isPlainParam checks if node is plain parameter (not optional, not default, not rest)
func isPlainParam(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return !isOptionalParam(node) && !isDefaultParam(node) && !isRestParam(node)
}

func checkDefaultParamLast(ctx rule.RuleContext, functionNode *ast.Node) {
	var params []*ast.Node

	// Extract parameters based on function type
	switch functionNode.Kind {
	case ast.KindArrowFunction:
		if functionNode.AsArrowFunction().Parameters != nil {
			params = functionNode.AsArrowFunction().Parameters.Nodes
		}
	case ast.KindFunctionDeclaration:
		if functionNode.AsFunctionDeclaration().Parameters != nil {
			params = functionNode.AsFunctionDeclaration().Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		if functionNode.AsFunctionExpression().Parameters != nil {
			params = functionNode.AsFunctionExpression().Parameters.Nodes
		}
	case ast.KindMethodDeclaration:
		if functionNode.AsMethodDeclaration().Parameters != nil {
			params = functionNode.AsMethodDeclaration().Parameters.Nodes
		}
	case ast.KindConstructor:
		if functionNode.AsConstructorDeclaration().Parameters != nil {
			params = functionNode.AsConstructorDeclaration().Parameters.Nodes
		}
	case ast.KindGetAccessor:
		if functionNode.AsGetAccessorDeclaration().Parameters != nil {
			params = functionNode.AsGetAccessorDeclaration().Parameters.Nodes
		}
	case ast.KindSetAccessor:
		if functionNode.AsSetAccessorDeclaration().Parameters != nil {
			params = functionNode.AsSetAccessorDeclaration().Parameters.Nodes
		}
	default:
		return
	}

	hasSeenPlainParam := false
	var violatingParams []*ast.Node
	
	// Iterate through parameters from right to left to find violations
	for i := len(params) - 1; i >= 0; i-- {
		current := params[i]
		if current == nil {
			continue
		}

		if isPlainParam(current) {
			hasSeenPlainParam = true
			continue
		}

		if hasSeenPlainParam && (isOptionalParam(current) || isDefaultParam(current)) {
			violatingParams = append(violatingParams, current)
		}
	}
	
	// Report violations in forward order (left to right)
	for i := len(violatingParams) - 1; i >= 0; i-- {
		ctx.ReportNode(violatingParams[i], shouldBeLastMessage())
	}
}

var DefaultParamLastRule = rule.Rule{
	Name: "default-param-last",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindArrowFunction: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
			ast.KindConstructor: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				checkDefaultParamLast(ctx, node)
			},
		}
	},
}