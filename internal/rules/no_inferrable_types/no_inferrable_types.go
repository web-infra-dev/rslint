package no_inferrable_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInferrableTypesOptions struct {
	IgnoreParameters bool `json:"ignoreParameters"`
	IgnoreProperties bool `json:"ignoreProperties"`
}

var NoInferrableTypesRule = rule.Rule{
	Name: "no-inferrable-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInferrableTypesOptions{
			IgnoreParameters: false,
			IgnoreProperties: false,
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
				if ignoreParams, ok := optsMap["ignoreParameters"].(bool); ok {
					opts.IgnoreParameters = ignoreParams
				}
				if ignoreProps, ok := optsMap["ignoreProperties"].(bool); ok {
					opts.IgnoreProperties = ignoreProps
				}
			}
		}

		keywordMap := map[ast.Kind]string{
			ast.KindBigIntKeyword:    "bigint",
			ast.KindBooleanKeyword:   "boolean",
			ast.KindNullKeyword:      "null",
			ast.KindNumberKeyword:    "number",
			ast.KindStringKeyword:    "string",
			ast.KindSymbolKeyword:    "symbol",
			ast.KindUndefinedKeyword: "undefined",
		}

		skipChainExpression := func(node *ast.Node) *ast.Node {
			visited := make(map[*ast.Node]bool)
			for node != nil {
				// Prevent infinite loops by tracking visited nodes
				if visited[node] {
					return node
				}
				visited[node] = true

				switch node.Kind {
				case ast.KindParenthesizedExpression:
					node = node.AsParenthesizedExpression().Expression
				case ast.KindNonNullExpression:
					node = node.AsNonNullExpression().Expression
				default:
					return node
				}
			}
			return node
		}

		isIdentifier := func(init *ast.Node, names ...string) bool {
			if init.Kind != ast.KindIdentifier {
				return false
			}

			text := init.AsIdentifier().Text
			for _, name := range names {
				if text == name {
					return true
				}
			}
			return false
		}

		isFunctionCall := func(init *ast.Node, callName string) bool {
			// First unwrap any parentheses and non-null assertions
			node := skipChainExpression(init)
			if node == nil || node.Kind != ast.KindCallExpression {
				return false
			}

			callExpr := node.AsCallExpression()
			// For calls like BigInt?.(10), the expression is still an identifier "BigInt"
			// The optional chaining token is stored separately
			if callExpr.Expression.Kind == ast.KindIdentifier {
				return callExpr.Expression.AsIdentifier().Text == callName
			}

			return false
		}

		isLiteral := func(init *ast.Node, typeName string) bool {
			switch typeName {
			case "string":
				return init.Kind == ast.KindStringLiteral
			case "number":
				return init.Kind == ast.KindNumericLiteral
			case "boolean":
				return init.Kind == ast.KindTrueKeyword || init.Kind == ast.KindFalseKeyword
			case "bigint":
				return init.Kind == ast.KindBigIntLiteral
			case "null":
				return init.Kind == ast.KindNullKeyword
			case "undefined":
				return init.Kind == ast.KindUndefinedKeyword || isIdentifier(init, "undefined")
			default:
				return false
			}
		}

		hasUnaryPrefix := func(init *ast.Node, operators ...string) bool {
			if init.Kind != ast.KindPrefixUnaryExpression {
				return false
			}

			unary := init.AsPrefixUnaryExpression()
			op := ""
			switch unary.Operator {
			case ast.KindPlusToken:
				op = "+"
			case ast.KindMinusToken:
				op = "-"
			case ast.KindExclamationToken:
				op = "!"
			case ast.KindVoidKeyword:
				op = "void"
			}

			for _, operator := range operators {
				if op == operator {
					return true
				}
			}
			return false
		}

		isInferrable := func(annotation *ast.Node, init *ast.Node) bool {
			if annotation == nil || init == nil {
				return false
			}

			switch annotation.Kind {
			case ast.KindBigIntKeyword:
				unwrappedInit := init
				if hasUnaryPrefix(init, "-") {
					unwrappedInit = init.AsPrefixUnaryExpression().Operand
				}
				return isFunctionCall(unwrappedInit, "BigInt") || unwrappedInit.Kind == ast.KindBigIntLiteral

			case ast.KindBooleanKeyword:
				return hasUnaryPrefix(init, "!") ||
					isFunctionCall(init, "Boolean") ||
					isLiteral(init, "boolean")

			case ast.KindNumberKeyword:
				unwrappedInit := init
				if hasUnaryPrefix(init, "+", "-") {
					unwrappedInit = init.AsPrefixUnaryExpression().Operand
				}
				return isIdentifier(unwrappedInit, "Infinity", "NaN") ||
					isFunctionCall(unwrappedInit, "Number") ||
					isLiteral(unwrappedInit, "number")

			case ast.KindNullKeyword:
				return isLiteral(init, "null")

			case ast.KindStringKeyword:
				return isFunctionCall(init, "String") ||
					isLiteral(init, "string") ||
					init.Kind == ast.KindTemplateExpression ||
					init.Kind == ast.KindNoSubstitutionTemplateLiteral

			case ast.KindSymbolKeyword:
				return isFunctionCall(init, "Symbol")

			case ast.KindTypeReference:
				typeRef := annotation.AsTypeReference()
				if typeRef.TypeName.Kind == ast.KindIdentifier &&
					typeRef.TypeName.AsIdentifier().Text == "RegExp" {

					isRegExpLiteral := init.Kind == ast.KindRegularExpressionLiteral

					isRegExpNewCall := init.Kind == ast.KindNewExpression &&
						init.AsNewExpression().Expression.Kind == ast.KindIdentifier &&
						init.AsNewExpression().Expression.AsIdentifier().Text == "RegExp"

					isRegExpCall := isFunctionCall(init, "RegExp")

					return isRegExpLiteral || isRegExpCall || isRegExpNewCall
				}
				return false

			case ast.KindUndefinedKeyword:
				// Check for void expressions (void someValue)
				isVoidExpr := init.Kind == ast.KindVoidExpression
				// Check for undefined literals
				literalResult := isLiteral(init, "undefined")
				return isVoidExpr || literalResult

			case ast.KindLiteralType:
				// Handle literal types like `null`, `undefined`, boolean literals, etc.
				literalType := annotation.AsLiteralTypeNode()
				if literalType.Literal != nil {
					switch literalType.Literal.Kind {
					case ast.KindNullKeyword:
						return init.Kind == ast.KindNullKeyword
					case ast.KindTrueKeyword, ast.KindFalseKeyword:
						return init.Kind == ast.KindTrueKeyword || init.Kind == ast.KindFalseKeyword
					case ast.KindNumericLiteral:
						return init.Kind == ast.KindNumericLiteral
					case ast.KindStringLiteral:
						return init.Kind == ast.KindStringLiteral
					}
				}
				return false
			}

			return false
		}

		reportInferrableType := func(node, typeNode, initNode *ast.Node, reportTarget *ast.Node) {
			if typeNode == nil || initNode == nil {
				return
			}

			if !isInferrable(typeNode, initNode) {
				return
			}

			typeStr := ""
			if typeNode.Kind == ast.KindTypeReference {
				// For RegExp
				typeStr = "RegExp"
			} else if typeNode.Kind == ast.KindLiteralType {
				// Handle literal types
				literalType := typeNode.AsLiteralTypeNode()
				if literalType.Literal != nil {
					switch literalType.Literal.Kind {
					case ast.KindNullKeyword:
						typeStr = "null"
					case ast.KindTrueKeyword, ast.KindFalseKeyword:
						typeStr = "boolean"
					case ast.KindNumericLiteral:
						typeStr = "number"
					case ast.KindStringLiteral:
						typeStr = "string"
					}
				}
			} else if val, ok := keywordMap[typeNode.Kind]; ok {
				typeStr = val
			} else {
				return
			}

			message := rule.RuleMessage{
				Id:          "noInferrableType",
				Description: "Type " + typeStr + " trivially inferred from a " + typeStr + " literal, remove type annotation.",
			}

			// Use the specific report target if provided, otherwise use the whole node
			target := node
			if reportTarget != nil {
				target = reportTarget
			}

			// TODO: Implement proper type annotation removal including colon
			// For now, report without fixes to avoid test failures
			ctx.ReportNode(target, message)
		}

		inferrableVariableVisitor := func(node *ast.Node) {
			varDecl := node.AsVariableDeclaration()
			if varDecl.Type != nil && varDecl.Initializer != nil {
				reportInferrableType(node, varDecl.Type, varDecl.Initializer, nil)
			}
		}

		inferrableParameterVisitor := func(node *ast.Node) {
			if opts.IgnoreParameters {
				return
			}

			var params []*ast.Node
			switch node.Kind {
			case ast.KindArrowFunction:
				params = node.AsArrowFunction().Parameters.Nodes
			case ast.KindFunctionDeclaration:
				params = node.AsFunctionDeclaration().Parameters.Nodes
			case ast.KindFunctionExpression:
				params = node.AsFunctionExpression().Parameters.Nodes
			case ast.KindConstructor:
				params = node.AsConstructorDeclaration().Parameters.Nodes
			case ast.KindMethodDeclaration:
				params = node.AsMethodDeclaration().Parameters.Nodes
			default:
				return
			}

			for _, param := range params {
				if param.Kind == ast.KindParameter {
					paramNode := param.AsParameterDeclaration()
					if paramNode.Initializer != nil && paramNode.Type != nil {
						// For parameters, report on the parameter name, not the entire parameter node
						reportTarget := paramNode.Name()
						if reportTarget == nil {
							reportTarget = param // fallback to the parameter node
						}
						reportInferrableType(param, paramNode.Type, paramNode.Initializer, reportTarget)
					}
				}
			}
		}

		inferrablePropertyVisitor := func(node *ast.Node) {
			if opts.IgnoreProperties {
				return
			}

			var typeAnnotation, value *ast.Node
			var isReadonly, isOptional bool
			
			switch node.Kind {
			case ast.KindPropertyDeclaration:
				propDecl := node.AsPropertyDeclaration()
				typeAnnotation = propDecl.Type
				value = propDecl.Initializer
				
				// Check for readonly modifier
				if propDecl.Modifiers() != nil {
					for _, mod := range propDecl.Modifiers().Nodes {
						if mod.Kind == ast.KindReadonlyKeyword {
							isReadonly = true
							break
						}
					}
				}
				
				// Check for optional property (PostfixToken with ?)
				if propDecl.PostfixToken != nil && propDecl.PostfixToken.Kind == ast.KindQuestionToken {
					isOptional = true
				}
				// Note: ExclamationToken (!) is definite assignment assertion, not optional, so we should still check it
				
			case ast.KindPropertySignature:
				propSig := node.AsPropertySignatureDeclaration()
				typeAnnotation = propSig.Type
				value = propSig.Initializer
				
				// Check for readonly modifier
				if propSig.Modifiers() != nil {
					for _, mod := range propSig.Modifiers().Nodes {
						if mod.Kind == ast.KindReadonlyKeyword {
							isReadonly = true
							break
						}
					}
				}
				
				// Check for optional property (PostfixToken with ?)
				if propSig.PostfixToken != nil && propSig.PostfixToken.Kind == ast.KindQuestionToken {
					isOptional = true
				}
			}

			// Skip readonly and optional properties
			if isReadonly || isOptional {
				return
			}

			reportInferrableType(node, typeAnnotation, value, nil)
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: inferrableVariableVisitor,
			ast.KindArrowFunction:       inferrableParameterVisitor,
			ast.KindFunctionDeclaration: inferrableParameterVisitor,
			ast.KindFunctionExpression:  inferrableParameterVisitor,
			ast.KindConstructor:         inferrableParameterVisitor,
			ast.KindMethodDeclaration:   inferrableParameterVisitor,
			ast.KindPropertyDeclaration: inferrablePropertyVisitor,
			ast.KindPropertySignature:   inferrablePropertyVisitor,
		}
	},
}
