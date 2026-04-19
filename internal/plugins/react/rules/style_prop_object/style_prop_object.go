package style_prop_object

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var StylePropObjectRule = rule.Rule{
	Name: "react/style-prop-object",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Parse the `allow` option: list of component names to skip
		var allowedComponents map[string]bool
		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			if allowList, ok := optsMap["allow"]; ok {
				if arr, ok := allowList.([]interface{}); ok {
					allowedComponents = make(map[string]bool)
					for _, item := range arr {
						if name, ok := item.(string); ok {
							allowedComponents[name] = true
						}
					}
				}
			}
		}

		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "stylePropNotObject",
				Description: "Style prop value must be an object",
			})
		}

		// checkIdentifier resolves an identifier to its declaration initializer
		// and reports if the initializer is a non-object literal.
		// This mirrors ESLint's checkIdentifiers which uses variableUtil.getVariableFromContext
		// to resolve variable.defs[0].node.init and checks isNonNullaryLiteral.
		// We use TypeChecker symbol resolution which is more accurate (handles cross-file, type info).
		checkIdentifier := func(expr *ast.Node, reportNode *ast.Node) {
			decl := utils.GetDeclaration(ctx.TypeChecker, expr)
			if decl == nil {
				return
			}
			if decl.Kind != ast.KindVariableDeclaration {
				return
			}
			init := decl.AsVariableDeclaration().Initializer
			if init == nil {
				return
			}
			if isNonObjectExpression(init) {
				report(reportNode)
			}
		}

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()

				// Check if the attribute name is "style"
				name := attr.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				if name.AsIdentifier().Text != "style" {
					return
				}

				// Check if the parent component is in the allow list
				if len(allowedComponents) > 0 {
					componentName := getParentComponentName(ctx.SourceFile, node)
					if allowedComponents[componentName] {
						return
					}
				}

				initializer := attr.Initializer
				if initializer == nil {
					// style without a value is fine (boolean shorthand)
					return
				}

				// If the value is a string literal, it's not an object
				if initializer.Kind == ast.KindStringLiteral {
					report(node)
					return
				}

				// If the value is a JSX expression container, check its expression
				if initializer.Kind == ast.KindJsxExpression {
					expr := initializer.AsJsxExpression().Expression
					if expr == nil {
						return
					}

					// If the expression is a literal (string, number, bool) but not null, report
					if isNonObjectExpression(expr) {
						report(node)
						return
					}

					// If the expression is an identifier, resolve its declaration initializer
					if expr.Kind == ast.KindIdentifier {
						checkIdentifier(expr, node)
					}
				}
			},

			// Handle <pragma>.createElement('div', { style: ... })
			// NOTE: Destructured createElement (e.g. import { createElement } from 'react')
			// and @jsx comment pragmas are not supported.
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateElementCall(call.Expression, reactutil.GetReactPragma(ctx.Settings)) {
					return
				}
				args := call.Arguments
				if args == nil || len(args.Nodes) < 2 {
					return
				}

				// Check allow list: if first arg is an identifier in the allowed set, skip
				firstArg := args.Nodes[0]
				if firstArg.Kind == ast.KindIdentifier {
					if allowedComponents[firstArg.AsIdentifier().Text] {
						return
					}
				}

				// Second arg must be an ObjectExpression
				secondArg := args.Nodes[1]
				if secondArg.Kind != ast.KindObjectLiteralExpression {
					return
				}

				// Find the "style" property (non-computed)
				obj := secondArg.AsObjectLiteralExpression()
				if obj.Properties == nil {
					return
				}
				for _, prop := range obj.Properties.Nodes {
					switch prop.Kind {
					case ast.KindPropertyAssignment:
						pa := prop.AsPropertyAssignment()
						nameNode := pa.Name()
						if nameNode == nil || nameNode.Kind == ast.KindComputedPropertyName {
							continue
						}
						if nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "style" {
							continue
						}
						value := pa.Initializer
						if value == nil {
							continue
						}
						if value.Kind == ast.KindIdentifier {
							checkIdentifier(value, value)
						} else if isNonObjectExpression(value) {
							report(value)
						}
						return
					case ast.KindShorthandPropertyAssignment:
						spa := prop.AsShorthandPropertyAssignment()
						nameNode := spa.Name()
						if nameNode == nil || nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "style" {
							continue
						}
						// For shorthand { style }, resolve the value symbol to the variable declaration
						valueSymbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(prop)
						if valueSymbol != nil && len(valueSymbol.Declarations) > 0 {
							decl := valueSymbol.Declarations[0]
							if decl.Kind == ast.KindVariableDeclaration {
								init := decl.AsVariableDeclaration().Initializer
								if init != nil && isNonObjectExpression(init) {
									report(nameNode)
								}
							}
						}
						return
					default:
						continue
					}
				}
			},
		}
	},
}

// isNonObjectExpression returns true if the expression is a literal value that is not an object.
// This mirrors ESLint's isNonNullaryLiteral: expression.type === 'Literal' && expression.value !== null.
// Our check covers StringLiteral, NumericLiteral, TrueKeyword, FalseKeyword.
// NOTE: ESLint's Literal type also includes RegExp and BigInt literals, but in practice
// no one writes style={/regex/} or style={100n}, so these edge cases are not covered.
func isNonObjectExpression(expr *ast.Node) bool {
	switch expr.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword:
		return true
	case ast.KindNullKeyword:
		// null is allowed (it's a common pattern to conditionally apply styles)
		return false
	}
	return false
}

// getParentComponentName gets the component name from the parent JSX element.
func getParentComponentName(sourceFile *ast.SourceFile, attrNode *ast.Node) string {
	// Walk up: JsxAttribute -> JsxAttributes -> JsxOpeningElement/JsxSelfClosingElement
	parent := attrNode.Parent
	if parent == nil {
		return ""
	}
	grandParent := parent.Parent
	if grandParent == nil {
		return ""
	}

	var tagName *ast.Node
	switch grandParent.Kind {
	case ast.KindJsxOpeningElement:
		tagName = grandParent.AsJsxOpeningElement().TagName
	case ast.KindJsxSelfClosingElement:
		tagName = grandParent.AsJsxSelfClosingElement().TagName
	}

	if tagName == nil {
		return ""
	}

	if tagName.Kind == ast.KindIdentifier {
		return tagName.AsIdentifier().Text
	}

	// For property access expressions, get the full text
	trimmed := utils.TrimNodeTextRange(sourceFile, tagName)
	return sourceFile.Text()[trimmed.Pos():trimmed.End()]
}
