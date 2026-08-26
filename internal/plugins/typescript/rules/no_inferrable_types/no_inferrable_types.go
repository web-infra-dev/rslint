package no_inferrable_types

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_inferrable_types.schema.json
var schemaJSON []byte

type Options struct {
	IgnoreParameters bool
	IgnoreProperties bool
}

type inferrableType uint8

const (
	inferrableTypeNone inferrableType = iota
	inferrableTypeBigInt
	inferrableTypeBoolean
	inferrableTypeNumber
	inferrableTypeString
	inferrableTypeNull
	inferrableTypeRegExp
	inferrableTypeUndefined
	inferrableTypeSymbol
)

var noInferrableTypesMessages = [...]rule.RuleMessage{
	inferrableTypeBigInt: {
		Id:          "noInferrableType",
		Description: "Type bigint trivially inferred from a bigint literal, remove type annotation.",
	},
	inferrableTypeBoolean: {
		Id:          "noInferrableType",
		Description: "Type boolean trivially inferred from a boolean literal, remove type annotation.",
	},
	inferrableTypeNumber: {
		Id:          "noInferrableType",
		Description: "Type number trivially inferred from a number literal, remove type annotation.",
	},
	inferrableTypeString: {
		Id:          "noInferrableType",
		Description: "Type string trivially inferred from a string literal, remove type annotation.",
	},
	inferrableTypeNull: {
		Id:          "noInferrableType",
		Description: "Type null trivially inferred from a null literal, remove type annotation.",
	},
	inferrableTypeRegExp: {
		Id:          "noInferrableType",
		Description: "Type RegExp trivially inferred from a RegExp literal, remove type annotation.",
	},
	inferrableTypeUndefined: {
		Id:          "noInferrableType",
		Description: "Type undefined trivially inferred from a undefined literal, remove type annotation.",
	},
	inferrableTypeSymbol: {
		Id:          "noInferrableType",
		Description: "Type symbol trivially inferred from a symbol literal, remove type annotation.",
	},
}

func parseOptions(options []any) Options {
	// Default values match typescript-eslint defaults
	opts := Options{
		IgnoreParameters: false,
		IgnoreProperties: false,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if v, ok := optsMap["ignoreParameters"].(bool); ok {
		opts.IgnoreParameters = v
	}
	if v, ok := optsMap["ignoreProperties"].(bool); ok {
		opts.IgnoreProperties = v
	}
	return opts
}

func getInferrableTypeAnnotation(typeNode *ast.Node) inferrableType {
	if typeNode == nil {
		return inferrableTypeNone
	}
	if typeNode.Kind == ast.KindParenthesizedType {
		typeNode = ast.SkipTypeParentheses(typeNode)
		if typeNode == nil {
			return inferrableTypeNone
		}
	}

	switch typeNode.Kind {
	case ast.KindBigIntKeyword:
		return inferrableTypeBigInt
	case ast.KindBooleanKeyword:
		return inferrableTypeBoolean
	case ast.KindNumberKeyword:
		return inferrableTypeNumber
	case ast.KindStringKeyword:
		return inferrableTypeString
	case ast.KindNullKeyword:
		return inferrableTypeNull
	case ast.KindLiteralType:
		// Depending on the parser path, `null` can be represented directly or
		// wrapped in a LiteralType node.
		litType := typeNode.AsLiteralTypeNode()
		if litType != nil && litType.Literal != nil && litType.Literal.Kind == ast.KindNullKeyword {
			return inferrableTypeNull
		}
	case ast.KindUndefinedKeyword:
		return inferrableTypeUndefined
	case ast.KindSymbolKeyword:
		return inferrableTypeSymbol
	case ast.KindTypeReference:
		typeRef := typeNode.AsTypeReferenceNode()
		if typeRef != nil && isIdentifierNamed(typeRef.TypeName, "RegExp") {
			return inferrableTypeRegExp
		}
	}
	return inferrableTypeNone
}

func identifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	identifier := node.AsIdentifier()
	if identifier == nil {
		return ""
	}
	return identifier.Text
}

func isIdentifierNamed(node *ast.Node, name string) bool {
	return identifierName(node) == name
}

func isNumberConstant(node *ast.Node) bool {
	name := identifierName(node)
	return name == "Infinity" || name == "NaN"
}

// isESTreeLiteral covers the tsgo kinds represented by a TSESTree Literal.
func isESTreeLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindBigIntLiteral ||
		node.Kind == ast.KindFalseKeyword ||
		node.Kind == ast.KindNullKeyword ||
		node.Kind == ast.KindNumericLiteral ||
		node.Kind == ast.KindRegularExpressionLiteral ||
		node.Kind == ast.KindStringLiteral ||
		node.Kind == ast.KindTrueKeyword
}

func skipParentheses(node *ast.Node) *ast.Node {
	if node != nil && node.Kind == ast.KindParenthesizedExpression {
		return ast.SkipParentheses(node)
	}
	return node
}

func isCallTo(node *ast.Node, name string) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	return call != nil && isIdentifierNamed(skipParentheses(call.Expression), name)
}

func isNewRegExp(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindNewExpression {
		return false
	}
	newExpression := node.AsNewExpression()
	return newExpression != nil && isIdentifierNamed(skipParentheses(newExpression.Expression), "RegExp")
}

// isInferrableInitializer is deliberately keyed by the annotation first. Most
// declarations use user-defined types, so they are rejected before inspecting
// the initializer at all.
func isInferrableInitializer(init *ast.Node, expectedType inferrableType) bool {
	if init == nil {
		return false
	}
	init = skipParentheses(init)

	switch expectedType {
	case inferrableTypeBigInt:
		if isESTreeLiteral(init) || isCallTo(init, "BigInt") {
			return true
		}
		if init.Kind == ast.KindPrefixUnaryExpression {
			unary := init.AsPrefixUnaryExpression()
			if unary == nil {
				return false
			}
			operand := skipParentheses(unary.Operand)
			return operand != nil &&
				unary.Operator == ast.KindMinusToken &&
				(isESTreeLiteral(operand) || isCallTo(operand, "BigInt"))
		}

	case inferrableTypeBoolean:
		if init.Kind == ast.KindTrueKeyword || init.Kind == ast.KindFalseKeyword || isCallTo(init, "Boolean") {
			return true
		}
		if init.Kind == ast.KindPrefixUnaryExpression {
			unary := init.AsPrefixUnaryExpression()
			return unary != nil && unary.Operator == ast.KindExclamationToken
		}

	case inferrableTypeNumber:
		if init.Kind == ast.KindNumericLiteral ||
			isNumberConstant(init) ||
			isCallTo(init, "Number") {
			return true
		}
		if init.Kind == ast.KindPrefixUnaryExpression {
			unary := init.AsPrefixUnaryExpression()
			if unary == nil {
				return false
			}
			operand := skipParentheses(unary.Operand)
			return operand != nil &&
				(unary.Operator == ast.KindPlusToken || unary.Operator == ast.KindMinusToken) &&
				(operand.Kind == ast.KindNumericLiteral ||
					isNumberConstant(operand) ||
					isCallTo(operand, "Number"))
		}

	case inferrableTypeString:
		return init.Kind == ast.KindStringLiteral ||
			init.Kind == ast.KindNoSubstitutionTemplateLiteral ||
			init.Kind == ast.KindTemplateExpression ||
			isCallTo(init, "String")

	case inferrableTypeNull:
		return init.Kind == ast.KindNullKeyword

	case inferrableTypeRegExp:
		return init.Kind == ast.KindRegularExpressionLiteral ||
			isCallTo(init, "RegExp") ||
			isNewRegExp(init)

	case inferrableTypeUndefined:
		return init.Kind == ast.KindVoidExpression || isIdentifierNamed(init, "undefined")

	case inferrableTypeSymbol:
		return isCallTo(init, "Symbol")
	}

	return false
}

func buildNoInferrableTypesFixes(sourceText string, typeAnnotation *ast.Node, postfixToken *ast.Node) []rule.RuleFix {
	// parseTypeAnnotation records the type node's full start immediately after
	// the colon token, even when comments or whitespace precede the type.
	colonPos := typeAnnotation.Pos() - 1
	if colonPos < 0 || colonPos >= len(sourceText) || sourceText[colonPos] != ':' {
		// Recovery/synthesized ASTs should still get a diagnostic, but no
		// potentially corrupting edit.
		return nil
	}

	typeFix := rule.RuleFixRemoveRange(core.NewTextRange(colonPos, typeAnnotation.End()))
	if postfixToken == nil {
		return []rule.RuleFix{typeFix}
	}

	// Preserve comments around `?`/`!` exactly like the upstream two-edit fix:
	// remove the token itself separately from the type annotation.
	postfixPos := postfixToken.End() - 1
	if postfixPos < 0 || postfixPos >= len(sourceText) {
		return nil
	}
	postfix := sourceText[postfixPos]
	if postfix != '?' && postfix != '!' {
		return nil
	}
	return []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(postfixPos, postfixToken.End())),
		typeFix,
	}
}

var NoInferrableTypesRule = rule.CreateRule(rule.Rule{
	Name:   "no-inferrable-types",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		sourceText := ctx.SourceFile.Text()

		checkDeclaration := func(reportStartNode *ast.Node, reportEndNode *ast.Node, typeAnnotation *ast.Node, initializer *ast.Node, postfixToken *ast.Node) {
			if reportStartNode == nil || reportEndNode == nil || typeAnnotation == nil || initializer == nil {
				return
			}
			if utils.IsJSDocSyntaxNode(typeAnnotation) {
				return
			}

			inferrableType := getInferrableTypeAnnotation(typeAnnotation)
			if inferrableType == inferrableTypeNone || !isInferrableInitializer(initializer, inferrableType) {
				return
			}

			reportRange := core.NewTextRange(
				scanner.SkipTrivia(sourceText, reportStartNode.Pos()),
				reportEndNode.End(),
			)
			ctx.ReportRangeWithDeferredFixes(
				reportRange,
				noInferrableTypesMessages[inferrableType],
				func() []rule.RuleFix {
					return buildNoInferrableTypesFixes(sourceText, typeAnnotation, postfixToken)
				},
			)
		}

		listeners := rule.RuleListeners{
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
				checkDeclaration(reportNode, varDecl.Initializer, varDecl.Type, varDecl.Initializer, nil)
			},
		}

		if !opts.IgnoreParameters {
			listeners[ast.KindParameter] = func(node *ast.Node) {
				param := node.AsParameterDeclaration()
				if param == nil {
					return
				}
				// Report on the name node for correct column position
				reportNode := param.Name()
				if reportNode == nil {
					reportNode = node
				}
				checkDeclaration(reportNode, param.Initializer, param.Type, param.Initializer, param.QuestionToken)
			}
		}

		if !opts.IgnoreProperties {
			listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
				prop := node.AsPropertyDeclaration()
				if prop == nil || prop.Type == nil || prop.Initializer == nil {
					return
				}

				// Optional properties and readonly properties must keep their
				// annotations because removing them can change their type.
				if prop.PostfixToken != nil && prop.PostfixToken.Kind == ast.KindQuestionToken {
					return
				}
				modifierFlags := node.ModifierFlags()
				if modifierFlags&ast.ModifierFlagsReadonly != 0 {
					return
				}

				// Upstream reports the complete property definition, including
				// decorators, modifiers, and a trailing semicolon.
				checkDeclaration(node, node, prop.Type, prop.Initializer, prop.PostfixToken)
			}
		}

		return listeners
	},
})
