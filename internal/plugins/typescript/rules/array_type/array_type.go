package array_type

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed array_type.schema.json
var schemaJSON []byte

type ArrayTypeOptions struct {
	Default  string `json:"default"`
	Readonly string `json:"readonly,omitempty"`
}

func parseOptions(options []any) ArrayTypeOptions {
	opts := ArrayTypeOptions{Default: "array"}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if defaultVal, ok := optsMap["default"].(string); ok {
		opts.Default = defaultVal
	}
	if readonlyVal, ok := optsMap["readonly"].(string); ok {
		opts.Readonly = readonlyVal
	}
	return opts
}

// Check whatever node can be considered as simple
func isSimpleType(node *ast.Node) bool {
	node = ast.SkipTypeParentheses(node)
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
	case ast.KindLiteralType:
		return node.AsLiteralTypeNode().Literal.Kind == ast.KindNullKeyword
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
	node = ast.SkipTypeParentheses(node)
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

func buildArrayMessage(id, className, readonlyPrefix, typeStr string) rule.RuleMessage {
	qualifier := ""
	if id == "errorStringArraySimple" || id == "errorStringArraySimpleReadonly" {
		qualifier = " for simple types"
	}
	brackets := ""
	if id == "errorStringArray" || id == "errorStringArraySimple" {
		brackets = "[]"
	}
	return rule.RuleMessage{
		Id: id,
		Description: "Array type using '" + className + "<" + typeStr + ">' is forbidden" +
			qualifier + ". Use '" + readonlyPrefix + typeStr + brackets + "' instead.",
	}
}

func buildGenericMessage(id, readonlyPrefix, typeStr, className string) rule.RuleMessage {
	qualifier := ""
	if id == "errorStringGenericSimple" {
		qualifier = " for non-simple types"
	}
	return rule.RuleMessage{
		Id: id,
		Description: "Array type using '" + readonlyPrefix + typeStr + "[]' is forbidden" +
			qualifier + ". Use '" + className + "<" + typeStr + ">' instead.",
	}
}

func nodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	nodeRange := utils.TrimNodeTextRange(sourceFile, ast.SkipTypeParentheses(node))
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

	return []rule.RuleFix{
		rule.RuleFixReplace(sourceFile, errorNode, className+"<"+elementTypeText+">"),
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
			!ast.IsParenthesizedTypeNode(node.Parent.AsArrayTypeNode().ElementType)
	}

	typeParamText := nodeText(sourceFile, typeParam)

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
	Name:   "array-type",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		defaultOption := opts.Default
		readonlyOption := opts.Readonly
		if readonlyOption == "" {
			readonlyOption = defaultOption
		}

		// Acquire models scope-manager's declaration sets, including value-only
		// shadows in type positions. ResolveTypeOrNamespace would miss those.
		var scopes *scope.Manager
		isShadowed := func(node *ast.Node, name string) bool {
			if scopes == nil {
				scopes = scopeanalysis.Declarations(ctx)
			}
			nonGlobalProgram := ctx.LanguageOptions.EffectiveSourceType() == "module"
			if ctx.Refs != nil {
				nonGlobalProgram = ctx.Refs.HasNonGlobalProgramScope()
			}
			for current := scopes.Acquire(node); current != nil; current = current.Parent {
				if current == scopes.Global && !nonGlobalProgram {
					break
				}
				if len(current.Declarations(name)) != 0 {
					return true
				}
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindArrayType: func(node *ast.Node) {
				arrayType := node.AsArrayTypeNode()
				if arrayType == nil {
					return
				}

				parent := ast.WalkUpParenthesizedTypes(node.Parent)
				isReadonly := false
				if parent != nil && parent.Kind == ast.KindTypeOperator {
					typeOp := parent.AsTypeOperatorNode()
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
					errorNode = parent
				}

				typeStr := messageType(ctx.SourceFile, arrayType.ElementType)
				className := "Array"
				readonlyPrefix := ""
				if isReadonly {
					className = "ReadonlyArray"
					readonlyPrefix = "readonly "
				}

				message := buildGenericMessage(messageId, readonlyPrefix, typeStr, className)

				ctx.ReportNodeWithDeferredFixes(errorNode, message, func() []rule.RuleFix {
					return buildGenericFixes(ctx.SourceFile, errorNode, arrayType.ElementType, className)
				})
			},

			ast.KindTypeReference: func(node *ast.Node) {
				// Heritage targets must remain entity names. Their type arguments
				// are visited separately and can still use array syntax.
				if typescriptutil.IsClassImplementsOrInterfaceExtends(node) {
					return
				}
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

				typeParams := typeRef.TypeArguments
				if typeParams == nil || len(typeParams.Nodes) != 1 {
					return
				}
				typeParam := ast.SkipTypeParentheses(typeParams.Nodes[0])

				// Handle Readonly<T[]> case
				if typeName == "Readonly" {
					if typeParam.Kind != ast.KindArrayType {
						return
					}
				}

				isReadonlyWithGenericArrayType := typeName == "Readonly"

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

				var messageId string
				switch currentOption {
				case "array":
					if isReadonlyWithGenericArrayType {
						messageId = "errorStringArrayReadonly"
					} else {
						messageId = "errorStringArray"
					}
				case "array-simple":
					if !isSimpleType(typeParam) {
						return
					}

					if isReadonlyArrayType && typeName != "ReadonlyArray" {
						messageId = "errorStringArraySimpleReadonly"
					} else {
						messageId = "errorStringArraySimple"
					}
				}

				if isShadowed(node, typeName) {
					return
				}

				typeStr := messageType(ctx.SourceFile, typeParam)
				className := typeName
				if !isReadonlyArrayType {
					className = "Array"
				}

				message := buildArrayMessage(messageId, className, readonlyPrefix, typeStr)

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
