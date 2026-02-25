package no_inferrable_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type Options struct {
	IgnoreParameters bool
	IgnoreProperties bool
}

func parseOptions(options any) Options {
	// Default values match typescript-eslint defaults
	opts := Options{
		IgnoreParameters: false,
		IgnoreProperties: false,
	}
	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			optsMap, _ = arr[0].(map[string]interface{})
		}
	} else {
		// Handle direct object format
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["ignoreParameters"].(bool); ok {
			opts.IgnoreParameters = v
		}
		if v, ok := optsMap["ignoreProperties"].(bool); ok {
			opts.IgnoreProperties = v
		}
	}
	return opts
}

func buildNoInferrableTypesMessage(typeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noInferrableType",
		Description: "Type " + typeName + " trivially inferred from a " + typeName + " literal, remove type annotation.",
	}
}

// getInferrableType checks if the initializer is an inferrable type
// Returns the type name if inferrable, empty string otherwise
func getInferrableType(init *ast.Node) string {
	if init == nil {
		return ""
	}

	switch init.Kind {
	case ast.KindBigIntLiteral:
		return "bigint"

	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return "boolean"

	case ast.KindNumericLiteral:
		return "number"

	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return "string"

	case ast.KindNullKeyword:
		return "null"

	case ast.KindRegularExpressionLiteral:
		return "RegExp"

	case ast.KindIdentifier:
		id := init.AsIdentifier()
		if id != nil {
			switch id.Text {
			case "undefined":
				return "undefined"
			case "Infinity", "NaN":
				return "number"
			}
		}

	case ast.KindPrefixUnaryExpression:
		unary := init.AsPrefixUnaryExpression()
		if unary != nil {
			switch unary.Operator {
			case ast.KindExclamationToken:
				// !x is boolean
				return "boolean"
			case ast.KindPlusToken, ast.KindMinusToken:
				// +x, -x with number/bigint literals or function calls
				if unary.Operand != nil {
					if unary.Operand.Kind == ast.KindNumericLiteral {
						return "number"
					}
					if unary.Operand.Kind == ast.KindBigIntLiteral {
						return "bigint"
					}
					// Check for Infinity, NaN identifiers
					if unary.Operand.Kind == ast.KindIdentifier {
						id := unary.Operand.AsIdentifier()
						if id != nil && (id.Text == "Infinity" || id.Text == "NaN") {
							return "number"
						}
					}
					// Check for function calls like -BigInt(10), -Number('1')
					if unary.Operand.Kind == ast.KindCallExpression {
						call := unary.Operand.AsCallExpression()
						if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier {
							funcName := call.Expression.AsIdentifier()
							if funcName != nil {
								switch funcName.Text {
								case "BigInt":
									return "bigint"
								case "Number":
									return "number"
								}
							}
						}
					}
				}
			}
		}

	case ast.KindVoidExpression:
		return "undefined"

	case ast.KindCallExpression:
		call := init.AsCallExpression()
		if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier {
			funcName := call.Expression.AsIdentifier()
			if funcName != nil {
				switch funcName.Text {
				case "BigInt":
					return "bigint"
				case "Boolean":
					return "boolean"
				case "Number":
					return "number"
				case "String":
					return "string"
				case "Symbol":
					return "symbol"
				case "RegExp":
					return "RegExp"
				}
			}
		}

	case ast.KindNewExpression:
		newExpr := init.AsNewExpression()
		if newExpr != nil && newExpr.Expression != nil && newExpr.Expression.Kind == ast.KindIdentifier {
			className := newExpr.Expression.AsIdentifier()
			if className != nil && className.Text == "RegExp" {
				return "RegExp"
			}
		}
	}

	return ""
}

// matchesTypeAnnotation checks if the type annotation matches the expected type
func matchesTypeAnnotation(typeNode *ast.Node, expectedType string) bool {
	if typeNode == nil {
		return false
	}

	switch expectedType {
	case "bigint":
		return typeNode.Kind == ast.KindBigIntKeyword
	case "boolean":
		return typeNode.Kind == ast.KindBooleanKeyword
	case "number":
		return typeNode.Kind == ast.KindNumberKeyword
	case "string":
		return typeNode.Kind == ast.KindStringKeyword
	case "null":
		// null type can be KindNullKeyword directly or wrapped in LiteralType
		if typeNode.Kind == ast.KindNullKeyword {
			return true
		}
		if typeNode.Kind == ast.KindLiteralType {
			litType := typeNode.AsLiteralTypeNode()
			if litType != nil && litType.Literal != nil {
				return litType.Literal.Kind == ast.KindNullKeyword
			}
		}
		return false
	case "undefined":
		return typeNode.Kind == ast.KindUndefinedKeyword
	case "symbol":
		return typeNode.Kind == ast.KindSymbolKeyword
	case "RegExp":
		if typeNode.Kind == ast.KindTypeReference {
			typeRef := typeNode.AsTypeReference()
			if typeRef != nil && typeRef.TypeName != nil && typeRef.TypeName.Kind == ast.KindIdentifier {
				return typeRef.TypeName.AsIdentifier().Text == "RegExp"
			}
		}
	}
	return false
}

var NoInferrableTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-inferrable-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkDeclaration := func(reportNode *ast.Node, typeAnnotation *ast.Node, initializer *ast.Node, hasQuestionToken bool, hasExclamationToken bool) {
			if typeAnnotation == nil || initializer == nil {
				return
			}

			inferrableType := getInferrableType(initializer)
			if inferrableType == "" {
				return
			}

			if matchesTypeAnnotation(typeAnnotation, inferrableType) {
				// Report with fix to remove type annotation
				// Find the colon position by scanning backwards from typeAnnotation
				colonPos := typeAnnotation.Pos()
				sourceText := ctx.SourceFile.Text()
				for i := colonPos - 1; i >= 0; i-- {
					if sourceText[i] == ':' {
						colonPos = i
						break
					}
				}

				// If there's a question token (optional) or exclamation token (definite assignment),
				// we need to remove it too and adjust the position
				if hasQuestionToken || hasExclamationToken {
					for i := colonPos - 1; i >= 0; i-- {
						ch := sourceText[i]
						if ch == '?' || ch == '!' {
							colonPos = i
							break
						} else if ch != ' ' && ch != '\t' {
							break
						}
					}
				}

				fixRange := core.NewTextRange(colonPos, typeAnnotation.End())
				// Report range spans from the name node (trimmed) to the end of the initializer
				// This matches typescript-eslint's behavior of reporting on the full declaration
				trimmedStartPos := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, reportNode.Pos()).Pos()
				reportRange := core.NewTextRange(trimmedStartPos, initializer.End())
				ctx.ReportRangeWithFixes(
					reportRange,
					buildNoInferrableTypesMessage(inferrableType),
					rule.RuleFixRemoveRange(fixRange),
				)
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}
				// Report on the name node for correct column position
				reportNode := varDecl.Name()
				if reportNode == nil {
					reportNode = node
				}
				checkDeclaration(reportNode, varDecl.Type, varDecl.Initializer, false, false)
			},

			ast.KindParameter: func(node *ast.Node) {
				if opts.IgnoreParameters {
					return
				}
				param := node.AsParameterDeclaration()
				if param == nil {
					return
				}
				// Check for optional parameter with question token
				hasQuestionToken := param.QuestionToken != nil
				// Report on the name node for correct column position
				reportNode := param.Name()
				if reportNode == nil {
					reportNode = node
				}
				checkDeclaration(reportNode, param.Type, param.Initializer, hasQuestionToken, false)
			},

			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if opts.IgnoreProperties {
					return
				}
				prop := node.AsPropertyDeclaration()
				if prop == nil {
					return
				}
				// Check if this is an auto-accessor property
				isAutoAccessor := false
				if prop.Modifiers() != nil {
					for _, mod := range prop.Modifiers().Nodes {
						// Skip readonly properties
						if mod.Kind == ast.KindReadonlyKeyword {
							return
						}
						// Check for accessor modifier
						if mod.Kind == ast.KindAccessorKeyword {
							isAutoAccessor = true
						}
					}
				}
				// Check for optional property with question token - skip these
				if prop.PostfixToken != nil && prop.PostfixToken.Kind == ast.KindQuestionToken {
					return
				}
				// Check for definite assignment assertion (!) - these should be reported with fix
				hasExclamationToken := prop.PostfixToken != nil && prop.PostfixToken.Kind == ast.KindExclamationToken
				// For auto-accessor properties, report from the accessor keyword (full node)
				// For regular properties, report from the name node
				var reportNode *ast.Node
				if isAutoAccessor {
					// Use the full property declaration node for auto-accessors
					reportNode = node
				} else {
					reportNode = prop.Name()
					if reportNode == nil {
						reportNode = node
					}
				}
				checkDeclaration(reportNode, prop.Type, prop.Initializer, false, hasExclamationToken)
			},
		}
	},
})
