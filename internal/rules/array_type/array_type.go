package array_type

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type ArrayTypeOptions struct {
	Default  string `json:"default"`
	Readonly string `json:"readonly,omitempty"`
}

// Check whatever node can be considered as simple
func isSimpleType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindAnyKeyword,
		ast.KindBooleanKeyword,
		ast.KindNeverKeyword,
		ast.KindNumberKeyword,
		ast.KindBigIntKeyword,
		ast.KindObjectKeyword,
		ast.KindStringKeyword,
		ast.KindSymbolKeyword,
		ast.KindUnknownKeyword,
		ast.KindVoidKeyword,
		ast.KindNullKeyword,
		ast.KindArrayType,
		ast.KindUndefinedKeyword,
		ast.KindThisType,
		ast.KindQualifiedName:
		return true
	case ast.KindTypeReference:
		typeRef := node.AsTypeReference()
		if ast.IsIdentifier(typeRef.TypeName) {
			identifier := typeRef.TypeName.AsIdentifier()
			if identifier.Text == "Array" {
				if typeRef.TypeArguments == nil {
					return true
				}
				if len(typeRef.TypeArguments.Nodes) == 1 {
					return isSimpleType(typeRef.TypeArguments.Nodes[0])
				}
			} else {
				if typeRef.TypeArguments != nil {
					return false
				}
				return isSimpleType(typeRef.TypeName)
			}
		}
		return false
	default:
		return false
	}
}

// Check if node needs parentheses
func typeNeedsParentheses(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindTypeReference:
		typeRef := node.AsTypeReference()
		return typeNeedsParentheses(typeRef.TypeName)
	case ast.KindUnionType,
		ast.KindFunctionType,
		ast.KindIntersectionType,
		ast.KindTypeOperator,
		ast.KindInferType,
		ast.KindConstructorType,
		ast.KindConditionalType:
		return true
	case ast.KindIdentifier:
		identifier := node.AsIdentifier()
		return identifier.Text == "ReadonlyArray"
	default:
		return false
	}
}

func isParenthesized(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	
	// Simple check - if the parent is a parenthesized type expression
	return ast.IsParenthesizedTypeNode(parent)
}

func buildErrorStringArrayMessage(className, readonlyPrefix, typeStr string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringArray",
		Description: fmt.Sprintf("Array type using '%s<%s>' is forbidden. Use '%s%s[]' instead.", className, typeStr, readonlyPrefix, typeStr),
	}
}

func buildErrorStringArrayReadonlyMessage(className, readonlyPrefix, typeStr string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringArrayReadonly",
		Description: fmt.Sprintf("Array type using '%s<%s>' is forbidden. Use '%s%s' instead.", className, typeStr, readonlyPrefix, typeStr),
	}
}

func buildErrorStringArraySimpleMessage(className, readonlyPrefix, typeStr string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringArraySimple",
		Description: fmt.Sprintf("Array type using '%s<%s>' is forbidden for simple types. Use '%s%s[]' instead.", className, typeStr, readonlyPrefix, typeStr),
	}
}

func buildErrorStringArraySimpleReadonlyMessage(className, readonlyPrefix, typeStr string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringArraySimpleReadonly",
		Description: fmt.Sprintf("Array type using '%s<%s>' is forbidden for simple types. Use '%s%s' instead.", className, typeStr, readonlyPrefix, typeStr),
	}
}

func buildErrorStringGenericMessage(readonlyPrefix, typeStr, className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringGeneric",
		Description: fmt.Sprintf("Array type using '%s%s[]' is forbidden. Use '%s<%s>' instead.", readonlyPrefix, typeStr, className, typeStr),
	}
}

func buildErrorStringGenericSimpleMessage(readonlyPrefix, typeStr, className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorStringGenericSimple",
		Description: fmt.Sprintf("Array type using '%s%s[]' is forbidden for non-simple types. Use '%s<%s>' instead.", readonlyPrefix, typeStr, className, typeStr),
	}
}

var ArrayTypeRule = rule.Rule{
	Name: "array-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ArrayTypeOptions{
			Default: "array",
		}
		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
				if defaultVal, ok := optsMap["default"].(string); ok {
					opts.Default = defaultVal
				}
				if readonlyVal, ok := optsMap["readonly"].(string); ok {
					opts.Readonly = readonlyVal
				}
			}
		}

		defaultOption := opts.Default
		readonlyOption := opts.Readonly
		if readonlyOption == "" {
			readonlyOption = defaultOption
		}

		getMessageType := func(node *ast.Node) string {
			if isSimpleType(node) {
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				return string(ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()])
			}
			return "T"
		}

		return rule.RuleListeners{
			ast.KindArrayType: func(node *ast.Node) {
				arrayType := node.AsArrayTypeNode()
				
				isReadonly := node.Parent != nil &&
					node.Parent.Kind == ast.KindTypeOperator &&
					node.Parent.AsTypeOperatorNode().Operator == ast.KindReadonlyKeyword

				currentOption := defaultOption
				if isReadonly {
					currentOption = readonlyOption
				}

				if currentOption == "array" ||
					(currentOption == "array-simple" && isSimpleType(arrayType.ElementType)) {
					return
				}

				var messageId string
				if currentOption == "generic" {
					messageId = "errorStringGeneric"
				} else {
					messageId = "errorStringGenericSimple"
				}

				errorNode := node
				if isReadonly {
					errorNode = node.Parent
				}

				typeStr := getMessageType(arrayType.ElementType)
				className := "Array"
				readonlyPrefix := ""
				if isReadonly {
					className = "ReadonlyArray"
					readonlyPrefix = "readonly "
				}

				var message rule.RuleMessage
				if messageId == "errorStringGeneric" {
					message = buildErrorStringGenericMessage(readonlyPrefix, typeStr, className)
				} else {
					message = buildErrorStringGenericSimpleMessage(readonlyPrefix, typeStr, className)
				}

				// Get the exact text of the element type to preserve formatting
				elementTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, arrayType.ElementType)
				elementTypeText := string(ctx.SourceFile.Text()[elementTypeRange.Pos():elementTypeRange.End()])
				
				// When converting T[] -> Array<T>, remove unnecessary parentheses
				if ast.IsParenthesizedTypeNode(arrayType.ElementType) {
					// For parenthesized types, get the inner type to avoid double parentheses
					innerType := arrayType.ElementType.AsParenthesizedTypeNode().Type
					innerTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, innerType)
					elementTypeText = string(ctx.SourceFile.Text()[innerTypeRange.Pos():innerTypeRange.End()])
				}
				
				newText := fmt.Sprintf("%s<%s>", className, elementTypeText)
				ctx.ReportNodeWithFixes(errorNode, message,
					rule.RuleFixReplace(ctx.SourceFile, errorNode, newText))
			},

			ast.KindTypeReference: func(node *ast.Node) {
				typeRef := node.AsTypeReference()
				
				if !ast.IsIdentifier(typeRef.TypeName) {
					return
				}

				identifier := typeRef.TypeName.AsIdentifier()
				typeName := identifier.Text
				
				if !(typeName == "Array" || typeName == "ReadonlyArray" || typeName == "Readonly") {
					return
				}

				// Handle Readonly<T[]> case
				if typeName == "Readonly" {
					if typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Nodes) == 0 {
						return
					}
					if typeRef.TypeArguments.Nodes[0].Kind != ast.KindArrayType {
						return
					}
				}

				isReadonlyWithGenericArrayType := typeName == "Readonly" &&
					typeRef.TypeArguments != nil &&
					len(typeRef.TypeArguments.Nodes) > 0 &&
					typeRef.TypeArguments.Nodes[0].Kind == ast.KindArrayType

				isReadonlyArrayType := typeName == "ReadonlyArray" || isReadonlyWithGenericArrayType

				currentOption := defaultOption
				if isReadonlyArrayType {
					currentOption = readonlyOption
				}

				if currentOption == "generic" {
					return
				}

				readonlyPrefix := ""
				if isReadonlyArrayType {
					readonlyPrefix = "readonly "
				}

				typeParams := typeRef.TypeArguments
				var messageId string
				if currentOption == "array" {
					if isReadonlyWithGenericArrayType {
						messageId = "errorStringArrayReadonly"
					} else {
						messageId = "errorStringArray"
					}
				} else {
					if isReadonlyArrayType && typeName != "ReadonlyArray" {
						messageId = "errorStringArraySimpleReadonly"
					} else {
						messageId = "errorStringArraySimple"
					}
				}

				if typeParams == nil || len(typeParams.Nodes) == 0 {
					// Create an 'any' array
					className := "Array"
					if isReadonlyArrayType {
						className = "ReadonlyArray"
					}

					var message rule.RuleMessage
					switch messageId {
					case "errorStringArray":
						message = buildErrorStringArrayMessage(className, readonlyPrefix, "any")
					case "errorStringArrayReadonly":
						message = buildErrorStringArrayReadonlyMessage(className, readonlyPrefix, "any")
					case "errorStringArraySimple":
						message = buildErrorStringArraySimpleMessage(className, readonlyPrefix, "any")
					case "errorStringArraySimpleReadonly":
						message = buildErrorStringArraySimpleReadonlyMessage(className, readonlyPrefix, "any")
					}

					ctx.ReportNodeWithFixes(node, message,
						rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%sany[]", readonlyPrefix)))
					return
				}

				if len(typeParams.Nodes) != 1 ||
					(currentOption == "array-simple" && !isSimpleType(typeParams.Nodes[0])) {
					return
				}

				typeParam := typeParams.Nodes[0]
				
				// Only add parentheses when converting Array<T> -> T[] if T needs them
				// Never add parentheses when converting T[] -> Array<T>
				var typeParens bool
				var parentParens bool
				
				if currentOption == "array" || currentOption == "array-simple" {
					// Converting Array<T> -> T[] - may need parentheses
					typeParens = typeNeedsParentheses(typeParam)
					parentParens = readonlyPrefix != "" &&
						node.Parent != nil &&
						node.Parent.Kind == ast.KindArrayType &&
						!isParenthesized(node.Parent.AsArrayTypeNode().ElementType)
				}
				// If converting T[] -> Array<T>, don't add parentheses

				start := ""
				if parentParens {
					start += "("
				}
				start += readonlyPrefix
				if typeParens {
					start += "("
				}

				end := ""
				if typeParens {
					end += ")"
				}
				if !isReadonlyWithGenericArrayType {
					end += "[]"
				}
				if parentParens {
					end += ")"
				}

				typeStr := getMessageType(typeParam)
				className := typeName
				if !isReadonlyArrayType {
					className = "Array"
				}

				var message rule.RuleMessage
				switch messageId {
				case "errorStringArray":
					message = buildErrorStringArrayMessage(className, readonlyPrefix, typeStr)
				case "errorStringArrayReadonly":
					message = buildErrorStringArrayReadonlyMessage(className, readonlyPrefix, typeStr)
				case "errorStringArraySimple":
					message = buildErrorStringArraySimpleMessage(className, readonlyPrefix, typeStr)
				case "errorStringArraySimpleReadonly":
					message = buildErrorStringArraySimpleReadonlyMessage(className, readonlyPrefix, typeStr)
				}

				// Get the exact text of the type parameter to preserve formatting
				typeParamRange := utils.TrimNodeTextRange(ctx.SourceFile, typeParam)
				typeParamText := string(ctx.SourceFile.Text()[typeParamRange.Pos():typeParamRange.End()])

				// When converting from array-simple mode, we're converting T[] -> Array<T>
				// In this case, if T is a parenthesized type, we should remove the parentheses
				if (currentOption == "array-simple") && ast.IsParenthesizedTypeNode(typeParam) {
					// For parenthesized types, get the inner type to avoid double parentheses
					innerType := typeParam.AsParenthesizedTypeNode().Type
					innerTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, innerType)
					typeParamText = string(ctx.SourceFile.Text()[innerTypeRange.Pos():innerTypeRange.End()])
				}

				ctx.ReportNodeWithFixes(node, message,
					rule.RuleFixReplace(ctx.SourceFile, node, start + typeParamText + end))
			},
		}
	},
}