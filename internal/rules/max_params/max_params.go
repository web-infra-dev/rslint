package max_params

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type MaxParamsOptions struct {
	Max           int  `json:"max"`
	Maximum       int  `json:"maximum"` // Deprecated alias for max
	CountVoidThis bool `json:"countVoidThis"`
}

func buildExceedMessage(name string, count int, max int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceed",
		Description: fmt.Sprintf("%s has too many parameters (%d). Maximum allowed is %d.", name, count, max),
	}
}

// Check if a parameter is a `this` parameter with void type annotation
func isVoidThisParam(param *ast.Node) bool {
	if param == nil || param.Kind != ast.KindParameter {
		return false
	}

	paramNode := param.AsParameterDeclaration()
	if paramNode.Name() == nil || !ast.IsIdentifier(paramNode.Name()) {
		return false
	}

	identifier := paramNode.Name().AsIdentifier()
	if identifier.Text != "this" {
		return false
	}

	// Check if it has a void type annotation
	if paramNode.Type == nil {
		return false
	}

	return paramNode.Type.Kind == ast.KindVoidKeyword
}

// Get function name for error message
func getFunctionName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil {
			return "Function '" + funcDecl.Name().AsIdentifier().Text + "'"
		}
		return "Function"
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr.Name() != nil {
			return "Function '" + funcExpr.Name().AsIdentifier().Text + "'"
		}
		return "Function"
	case ast.KindArrowFunction:
		return "Arrow function"
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method.Name() != nil {
			if ast.IsIdentifier(method.Name()) {
				return "Method '" + method.Name().AsIdentifier().Text + "'"
			}
		}
		return "Method"
	case ast.KindConstructor:
		return "Constructor"
	case ast.KindGetAccessor:
		getter := node.AsGetAccessorDeclaration()
		if getter.Name() != nil {
			if ast.IsIdentifier(getter.Name()) {
				return "Getter '" + getter.Name().AsIdentifier().Text + "'"
			}
		}
		return "Getter"
	case ast.KindSetAccessor:
		setter := node.AsSetAccessorDeclaration()
		if setter.Name() != nil {
			if ast.IsIdentifier(setter.Name()) {
				return "Setter '" + setter.Name().AsIdentifier().Text + "'"
			}
		}
		return "Setter"
	default:
		return "Function"
	}
}

// Get parameters from function-like node
func getParameters(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Parameters.Nodes
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Parameters.Nodes
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Parameters.Nodes
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Parameters.Nodes
	case ast.KindConstructor:
		return node.AsConstructorDeclaration().Parameters.Nodes
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration().Parameters.Nodes
	case ast.KindSetAccessor:
		return node.AsSetAccessorDeclaration().Parameters.Nodes
	case ast.KindFunctionType:
		return node.AsFunctionTypeNode().Parameters.Nodes
	case ast.KindCallSignature:
		return node.AsCallSignatureDeclaration().Parameters.Nodes
	case ast.KindConstructSignature:
		return node.AsConstructSignatureDeclaration().Parameters.Nodes
	default:
		return nil
	}
}

var MaxParamsRule = rule.Rule{
	Name: "max-params",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := MaxParamsOptions{
			Max:           3,
			CountVoidThis: false,
		}

		if options != nil {
			var optsMap map[string]interface{}

			// Handle both direct map format and array format
			if directMap, ok := options.(map[string]interface{}); ok {
				optsMap = directMap
			} else if optsSlice, ok := options.([]interface{}); ok && len(optsSlice) > 0 {
				// Handle array format like ["error", {max: 4}]
				if len(optsSlice) > 0 {
					if arrayMap, ok := optsSlice[0].(map[string]interface{}); ok {
						optsMap = arrayMap
					}
				}
			}

			if optsMap != nil {
				// Parse max option (support both int and float64)
				if maxVal, ok := optsMap["max"]; ok {
					switch v := maxVal.(type) {
					case float64:
						opts.Max = int(v)
					case int:
						opts.Max = v
					}
				}
				// Parse maximum option (deprecated alias)
				if maximumVal, ok := optsMap["maximum"]; ok {
					switch v := maximumVal.(type) {
					case float64:
						opts.Max = int(v)
					case int:
						opts.Max = v
					}
				}
				// Parse countVoidThis option
				if countVoidThis, ok := optsMap["countVoidThis"].(bool); ok {
					opts.CountVoidThis = countVoidThis
				}
			}
		}

		checkFunction := func(node *ast.Node) {
			params := getParameters(node)
			if params == nil {
				return
			}

			// Count parameters, potentially skipping void this
			paramCount := len(params)
			if !opts.CountVoidThis && paramCount > 0 && isVoidThisParam(params[0]) {
				paramCount--
			}

			if paramCount > opts.Max {
				funcName := getFunctionName(node)
				ctx.ReportNode(node, buildExceedMessage(funcName, paramCount, opts.Max))
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,
			ast.KindMethodDeclaration:   checkFunction,
			ast.KindConstructor:         checkFunction,
			ast.KindGetAccessor:         checkFunction,
			ast.KindSetAccessor:         checkFunction,
			ast.KindFunctionType:        checkFunction,
			ast.KindCallSignature:       checkFunction,
			ast.KindConstructSignature:  checkFunction,
		}
	},
}
