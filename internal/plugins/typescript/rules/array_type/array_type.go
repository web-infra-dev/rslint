package array_type

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
		typeRef := node.AsTypeReferenceNode()
		if typeRef == nil {
			return false
		}
		if ast.IsIdentifier(typeRef.TypeName) {
			identifier := typeRef.TypeName.AsIdentifier()
			if identifier == nil {
				return false
			}
			if identifier.Text == "Array" {
				if typeRef.TypeArguments == nil {
					return true
				}
				if len(typeRef.TypeArguments.Nodes) == 1 {
					return isSimpleType(typeRef.TypeArguments.Nodes[0])
				}
			} else {
				return typeRef.TypeArguments == nil
			}
		} else if ast.IsQualifiedName(typeRef.TypeName) {
			// TypeReference with a QualifiedName (e.g., fooName.BarType) is simple if it has no type arguments
			return typeRef.TypeArguments == nil
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
		typeRef := node.AsTypeReferenceNode()
		if typeRef == nil {
			return false
		}
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
		if identifier == nil {
			return false
		}
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
		Description: fmt.Sprintf("Array type using '%s<%s>' is forbidden. Use '%s%s[]' instead.", className, typeStr, readonlyPrefix, typeStr),
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
		Description: fmt.Sprintf("Array type using '%s<%s>' is forbidden for simple types. Use '%s%s[]' instead.", className, typeStr, readonlyPrefix, typeStr),
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

func nodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

func messageType(sourceFile *ast.SourceFile, node *ast.Node) string {
	if isSimpleType(node) {
		return nodeText(sourceFile, node)
	}
	return "T"
}

func buildGenericFixes(
	sourceFile *ast.SourceFile,
	errorNode *ast.Node,
	elementType *ast.Node,
	className string,
) []rule.RuleFix {
	elementTypeText := nodeText(sourceFile, elementType)

	// When converting T[] -> Array<T>, remove unnecessary parentheses.
	if ast.IsParenthesizedTypeNode(elementType) {
		parenType := elementType.AsParenthesizedTypeNode()
		if parenType != nil && parenType.Type != nil {
			elementTypeText = nodeText(sourceFile, parenType.Type)
		}
	}

	return []rule.RuleFix{
		rule.RuleFixReplace(sourceFile, errorNode, className+"<"+elementTypeText+">"),
	}
}

func buildAnyArrayFixes(
	sourceFile *ast.SourceFile,
	node *ast.Node,
	readonlyPrefix string,
) []rule.RuleFix {
	return []rule.RuleFix{
		rule.RuleFixReplace(sourceFile, node, readonlyPrefix+"any[]"),
	}
}

func buildArrayReplacement(
	typeParamText string,
	readonlyPrefix string,
	typeParens bool,
	parentParens bool,
	appendArrayBrackets bool,
) string {
	switch {
	case parentParens && typeParens && appendArrayBrackets:
		return "(" + readonlyPrefix + "(" + typeParamText + ")[])"
	case parentParens && typeParens:
		return "(" + readonlyPrefix + "(" + typeParamText + "))"
	case parentParens && appendArrayBrackets:
		return "(" + readonlyPrefix + typeParamText + "[])"
	case parentParens:
		return "(" + readonlyPrefix + typeParamText + ")"
	case typeParens && appendArrayBrackets:
		return readonlyPrefix + "(" + typeParamText + ")[]"
	case typeParens:
		return readonlyPrefix + "(" + typeParamText + ")"
	case appendArrayBrackets:
		return readonlyPrefix + typeParamText + "[]"
	default:
		return readonlyPrefix + typeParamText
	}
}

func buildArrayFixes(
	sourceFile *ast.SourceFile,
	node *ast.Node,
	typeParam *ast.Node,
	currentOption string,
	readonlyPrefix string,
	isReadonlyWithGenericArrayType bool,
) []rule.RuleFix {
	// Converting Array<T> -> T[] may require parentheses around T or
	// around the whole readonly form when it is nested in another array.
	var typeParens bool
	var parentParens bool
	if currentOption == "array" || currentOption == "array-simple" {
		typeParens = typeNeedsParentheses(typeParam)
		parentParens = readonlyPrefix != "" &&
			node.Parent != nil &&
			node.Parent.Kind == ast.KindArrayType &&
			!isParenthesized(node.Parent.AsArrayTypeNode().ElementType)
	}

	typeParamText := nodeText(sourceFile, typeParam)
	if currentOption == "array-simple" && ast.IsParenthesizedTypeNode(typeParam) {
		parenType := typeParam.AsParenthesizedTypeNode()
		if parenType != nil && parenType.Type != nil {
			typeParamText = nodeText(sourceFile, parenType.Type)
		}
	}

	return []rule.RuleFix{
		rule.RuleFixReplace(
			sourceFile,
			node,
			buildArrayReplacement(
				typeParamText,
				readonlyPrefix,
				typeParens,
				parentParens,
				!isReadonlyWithGenericArrayType,
			),
		),
	}
}

var ArrayTypeRule = rule.CreateRule(rule.Rule{
	Name: "array-type",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := ArrayTypeOptions{
			Default: "array",
		}
		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
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

		return rule.RuleListeners{
			ast.KindArrayType: func(node *ast.Node) {
				arrayType := node.AsArrayTypeNode()
				if arrayType == nil {
					return
				}

				isReadonly := false
				if node.Parent != nil && node.Parent.Kind == ast.KindTypeOperator {
					typeOp := node.Parent.AsTypeOperatorNode()
					if typeOp != nil {
						isReadonly = typeOp.Operator == ast.KindReadonlyKeyword
					}
				}

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

				typeStr := messageType(ctx.SourceFile, arrayType.ElementType)
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

				ctx.ReportNodeWithDeferredFixes(errorNode, message, func() []rule.RuleFix {
					return buildGenericFixes(ctx.SourceFile, errorNode, arrayType.ElementType, className)
				})
			},

			ast.KindTypeReference: func(node *ast.Node) {
				typeRef := node.AsTypeReferenceNode()
				if typeRef == nil {
					return
				}

				if !ast.IsIdentifier(typeRef.TypeName) {
					return
				}

				identifier := typeRef.TypeName.AsIdentifier()
				if identifier == nil {
					return
				}
				typeName := identifier.Text

				if typeName != "Array" && typeName != "ReadonlyArray" && typeName != "Readonly" {
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
				switch currentOption {
				case "array":
					if isReadonlyWithGenericArrayType {
						messageId = "errorStringArrayReadonly"
					} else {
						messageId = "errorStringArray"
					}
				case "array-simple":
					// For array-simple mode, determine if we have type parameters to check
					// 'any' (no type params) is considered simple
					isSimple := typeParams == nil || len(typeParams.Nodes) == 0 ||
						(len(typeParams.Nodes) == 1 && isSimpleType(typeParams.Nodes[0]))

					// For array-simple mode, only report errors if the type is simple
					if !isSimple {
						return
					}

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

					ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
						return buildAnyArrayFixes(ctx.SourceFile, node, readonlyPrefix)
					})
					return
				}

				if len(typeParams.Nodes) != 1 {
					return
				}

				typeParam := typeParams.Nodes[0]

				typeStr := messageType(ctx.SourceFile, typeParam)
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

				ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
					return buildArrayFixes(
						ctx.SourceFile,
						node,
						typeParam,
						currentOption,
						readonlyPrefix,
						isReadonlyWithGenericArrayType,
					)
				})
			},
		}
	},
})
