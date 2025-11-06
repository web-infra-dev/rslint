package default_param_last

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// DefaultParamLastRule enforces default parameters to be last
var DefaultParamLastRule = rule.CreateRule(rule.Rule{
	Name: "default-param-last",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Helper function to check if a parameter is optional
	isOptionalParam := func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		// Check if parameter has optional modifier
		if node.Kind == ast.KindParameter {
			param := node.AsParameterDeclaration()
			if param != nil && param.QuestionToken != nil {
				return true
			}
		}

		return false
	}

	// Helper function to check if a parameter is a rest parameter
	isRestParam := func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		// Check for rest parameter (...)
		if node.Kind == ast.KindParameter {
			param := node.AsParameterDeclaration()
			if param != nil && param.DotDotDotToken != nil {
				return true
			}
		}

		return false
	}

	// Helper function to check if a parameter is plain (no default, not rest, not optional)
	isPlainParam := func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		// Rest parameter is not plain
		if isRestParam(node) {
			return false
		}

		// Parameter with default value is not plain
		if node.Kind == ast.KindParameter {
			param := node.AsParameterDeclaration()
			if param != nil {
				// Has initializer (default value)
				if param.Initializer != nil {
					return false
				}
				// Is optional
				if param.QuestionToken != nil {
					return false
				}
			}
		}

		return true
	}

	// Check function for default parameter positioning
	checkDefaultParamLast := func(node *ast.Node) {
		var params []*ast.Node

		// Get parameters based on node type
		switch node.Kind {
		case ast.KindFunctionDeclaration:
			funcDecl := node.AsFunctionDeclaration()
			if funcDecl != nil && funcDecl.Parameters != nil {
				params = funcDecl.Parameters.Nodes
			}
		case ast.KindFunctionExpression:
			funcExpr := node.AsFunctionExpression()
			if funcExpr != nil && funcExpr.Parameters != nil {
				params = funcExpr.Parameters.Nodes
			}
		case ast.KindArrowFunction:
			arrowFunc := node.AsArrowFunction()
			if arrowFunc != nil && arrowFunc.Parameters != nil {
				params = arrowFunc.Parameters.Nodes
			}
		case ast.KindMethodDeclaration:
			methodDecl := node.AsMethodDeclaration()
			if methodDecl != nil && methodDecl.Parameters != nil {
				params = methodDecl.Parameters.Nodes
			}
		case ast.KindConstructor:
			constructor := node.AsConstructorDeclaration()
			if constructor != nil && constructor.Parameters != nil {
				params = constructor.Parameters.Nodes
			}
		default:
			return
		}

		if len(params) == 0 {
			return
		}

		// Iterate backward through parameters
		hasSeenPlainParam := false
		for i := len(params) - 1; i >= 0; i-- {
			current := params[i]
			if current == nil {
				continue
			}

			// Get the actual parameter (unwrap if it's a parameter property)
			param := current
			if current.Kind == ast.KindParameter {
				p := current.AsParameterDeclaration()
				if p != nil && p.Name() != nil {
					// Check if parameter has modifiers (public/private/protected/readonly)
					// which would make it a parameter property
					hasModifiers := ast.GetCombinedModifierFlags(current)&(ast.ModifierFlagsPublic|ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected|ast.ModifierFlagsReadonly) != 0
					if hasModifiers {
						// For parameter properties, check the parameter itself
						param = current
					}
				}
			}

			// Skip rest parameters - they can come after defaults
			if isRestParam(param) {
				continue
			}

			if isPlainParam(param) {
				hasSeenPlainParam = true
				continue
			}

			// Check if this is a default or optional parameter that comes before a plain parameter
			if hasSeenPlainParam {
				if param.Kind == ast.KindParameter {
					paramDecl := param.AsParameterDeclaration()
					isDefaultParam := paramDecl != nil && paramDecl.Initializer != nil
					isOptional := isOptionalParam(param)

					if isDefaultParam || isOptional {
						ctx.ReportNode(current, rule.RuleMessage{
							Id:          "shouldBeLast",
							Description: "Default parameters should be last.",
						})
					}
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: checkDefaultParamLast,
		ast.KindFunctionExpression:  checkDefaultParamLast,
		ast.KindArrowFunction:       checkDefaultParamLast,
		ast.KindMethodDeclaration:   checkDefaultParamLast,
		ast.KindConstructor:         checkDefaultParamLast,
	}
}
