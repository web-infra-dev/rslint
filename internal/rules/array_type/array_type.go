package array_type

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/rslint/internal/rule"
)

// rewrite of https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/array-type.ts

func GetMessageType(node *ast.Node) string {
	if isSimpleType(node) {
		return scanner.GetTextOfNode(node)
	} else {
		return "T"
	}
}

/**
 * Check whatever node can be considered as simple
 * @param node the node to be evaluated.
 */
func isSimpleType(node *ast.Node) bool {
	var tyKind = node.Kind
	switch tyKind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindBooleanKeyword,
		ast.KindNeverKeyword, ast.KindNumberKeyword, ast.KindBigIntKeyword,
		ast.KindObjectKeyword, ast.KindStringKeyword, ast.KindSymbolKeyword,
		ast.KindUnknownKeyword, ast.KindVoidKeyword, ast.KindNullKeyword,
		ast.KindUndefinedKeyword, ast.KindThisType, ast.KindQualifiedName,
		ast.KindArrayType:
		return true
	case ast.KindTypeReference:
		typeName := node.AsTypeReference().TypeName
		if typeName.Kind == ast.KindIdentifier && typeName.AsIdentifier().Text == "Array" {
			typeArgs := node.AsTypeReference().TypeArguments
			if typeArgs == nil {
				return true
			} else if len(typeArgs.Nodes) == 1 {
				return isSimpleType(typeArgs.Nodes[0])
			}
		} else {
			if typeArgs := node.AsTypeReference().TypeArguments; typeArgs != nil {
				return false
			}
			return isSimpleType(typeName)
		}
		return false
	default:
		return false
	}
}

/*
*

	Check if the type needs parentheses when used in an array type.
	* @param node the node to be evaluated.
	* @returns true if the type needs parentheses, false otherwise.
*/
func typeNeedsParentheses(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindTypeReference:
		return typeNeedsParentheses(node.AsTypeReference().TypeName)
	case ast.KindUnionType, ast.KindFunctionType, ast.KindIntersectionType,
		ast.KindTypeOperator, ast.KindInferType, ast.KindConstructorType,
		ast.KindConditionalType:
		return true
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "ReadonlyArray"
	default:
		return false
	}
}

type ArrayTypeOptions struct {
	Default  string `json:"default"`
	Readonly string `json:"readonly,omitempty"`
}

func buildErrorStringGenericMessage(readonlyPrefix string, typeName string, className string) rule.RuleMessage {

	return rule.RuleMessage{
		Id:          "errorStringGeneric",
		Description: fmt.Sprintf("Array type using '%v%v[]' is forbidden. Use '%v<%v>' instead.", readonlyPrefix, typeName, className, typeName),
	}
}

func buildErrorStringArrayMessage(readonlyPrefix string, typeName string, className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringArray",
		Description: fmt.Sprintf("Array type using '%v<%v>' is forbidden. Use '%v%v[]' instead.", className, typeName, readonlyPrefix, typeName),
	}
}

var ArrayTypeRule = rule.Rule{
	Name: "array-type",
	// Add any additional properties or methods for the rule here
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		
		parsedOptions, ok := options.(ArrayTypeOptions)
		var defaultOption = "array"
		var readonlyOption = "readonlyArray"
		if ok {
			defaultOption = parsedOptions.Default
			readonlyOption = parsedOptions.Readonly
		}
		return rule.RuleListeners{
			ast.KindArrayType: func(node *ast.Node) {

				isReadOnly := node.Parent.Type().Kind == ast.KindTypeOperator && node.Parent.AsTypeOperatorNode().Operator == ast.KindReadonlyKeyword

				var currentOption string
				if isReadOnly {
					currentOption = readonlyOption
				} else {
					currentOption = defaultOption
				}

				elementType := node.AsArrayTypeNode().ElementType
				if currentOption == "array" ||
					(currentOption == "array-simple" && isSimpleType(elementType)) {
					return
				}

				ctx.ReportNodeWithFixes(node, buildErrorStringGenericMessage("", "", ""), rule.RuleFixReplace(ctx.SourceFile, node, currentOption+"<"+elementType.Text()+">"))
			},
			ast.KindTypeReference: func(node *ast.Node) {
				typeName := node.AsTypeReference().TypeName
				// don't process Array<T> ReadonlyArray<T> Readonly<Array<T>>

				if typeName.Kind != ast.KindIdentifier {
					return
				}

				identifier := typeName.AsIdentifier()
				if !(identifier.Text == "Array" || identifier.Text == "ReadonlyArray" || identifier.Text == "Readonly") {
					return
				}

				if identifier.Text == "Readonly" {
					typeArgs := node.AsTypeReference().TypeArguments
					if typeArgs == nil || len(typeArgs.Nodes) == 0 || typeArgs.Nodes[0].Kind != ast.KindArrayType {
						return
					}
				}

				var typeArgs1 = node.AsTypeReference().TypeArguments

				ctx.ReportNodeWithFixes(node, buildErrorStringArrayMessage("", GetMessageType(typeArgs1.Nodes[0]), identifier.Text), rule.RuleFixReplace(ctx.SourceFile, node, "Array<"+typeName.Text()+">"))
			},
		}
	},
}
