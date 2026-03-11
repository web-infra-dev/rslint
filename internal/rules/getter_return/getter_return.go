package getter_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options for getter-return rule
type Options struct {
	AllowImplicit bool `json:"allowImplicit"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowImplicit: false,
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support (handles both array and object formats)
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
		if v, ok := optsMap["allowImplicit"].(bool); ok {
			opts.AllowImplicit = v
		}
	}
	return opts
}

func buildExpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expected",
		Description: "Expected to return a value in getter.",
	}
}

func buildExpectedAlwaysMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedAlways",
		Description: "Expected getter to always return a value.",
	}
}

// reportGetterReturn reports getter-return diagnostics based on flow analysis.
func reportGetterReturn(ctx rule.RuleContext, funcNode *ast.Node, reportNode *ast.Node, opts Options) {
	if funcNode == nil {
		return
	}

	body := funcNode.Body()
	if body == nil {
		return
	}

	// ESLint only checks getters with block bodies (BlockStatement).
	// Arrow functions with expression bodies (e.g., `get: () => value`) are
	// implicitly returning and should not be flagged.
	if body.Kind != ast.KindBlock {
		return
	}

	if opts.AllowImplicit {
		return
	}

	analysis := utils.AnalyzeFunctionReturns(funcNode)

	if reportNode == nil {
		reportNode = funcNode
	}

	if !analysis.HasReturnWithValue {
		// No return-with-value anywhere. Valid only if all paths throw.
		// EndReachable means some paths fall through; HasEmptyReturn means explicit "return;"
		if analysis.EndReachable || analysis.HasEmptyReturn {
			ctx.ReportNode(reportNode, buildExpectedMessage())
		}
	} else {
		// Has at least one return-with-value. All paths must return a value.
		if analysis.EndReachable || analysis.HasEmptyReturn {
			ctx.ReportNode(reportNode, buildExpectedAlwaysMessage())
		}
	}
}

// GetterReturnRule enforces return statements in getters
var GetterReturnRule = rule.Rule{
	Name: "getter-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindGetAccessor: func(node *ast.Node) {
				reportGetterReturn(ctx, node, node, opts)
			},

			// Handle Object.defineProperty, Reflect.defineProperty
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.Expression()
				if expr == nil {
					return
				}

				var objectName, methodName string

				// Handle optional chaining: Object?.defineProperty or (Object?.defineProperty)
				actualExpr := expr

				// Unwrap ParenthesizedExpression
				for actualExpr != nil && actualExpr.Kind == ast.KindParenthesizedExpression {
					actualExpr = actualExpr.Expression()
				}

				// Check for Object.defineProperty, Reflect.defineProperty
				if actualExpr != nil && actualExpr.Kind == ast.KindPropertyAccessExpression {
					obj := actualExpr.Expression()
					if obj != nil && obj.Kind == ast.KindIdentifier {
						objectName = obj.Text()
					}
					name := actualExpr.Name()
					if name != nil && name.Kind == ast.KindIdentifier {
						methodName = name.Text()
					}
				}

				args := node.Arguments()
				if args == nil {
					return
				}

				var descriptorArg *ast.Node

				// Object.defineProperty(obj, 'prop', { get: function() {} })
				if (objectName == "Object" && methodName == "defineProperty") ||
					(objectName == "Reflect" && methodName == "defineProperty") {
					if len(args) >= 3 {
						descriptorArg = args[2]
					}
				}

				// Object.defineProperties(obj, { prop: { get: function() {} } })
				if objectName == "Object" && methodName == "defineProperties" {
					if len(args) >= 2 {
						propsArg := args[1]
						if propsArg != nil && propsArg.Kind == ast.KindObjectLiteralExpression {
							props := propsArg.Properties()
							for _, prop := range props {
								if prop != nil && (prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindShorthandPropertyAssignment) {
									init := prop.Initializer()
									if init != nil && init.Kind == ast.KindObjectLiteralExpression {
										checkDescriptorForGetter(ctx, init, opts)
									}
								}
							}
						}
					}
					return
				}

				// Object.create(proto, { prop: { get: function() {} } })
				if objectName == "Object" && methodName == "create" {
					if len(args) >= 2 {
						descriptorArg = args[1]
						if descriptorArg != nil && descriptorArg.Kind == ast.KindObjectLiteralExpression {
							props := descriptorArg.Properties()
							for _, prop := range props {
								if prop != nil && (prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindShorthandPropertyAssignment) {
									init := prop.Initializer()
									if init != nil && init.Kind == ast.KindObjectLiteralExpression {
										checkDescriptorForGetter(ctx, init, opts)
									}
								}
							}
						}
					}
					return
				}

				if descriptorArg != nil && descriptorArg.Kind == ast.KindObjectLiteralExpression {
					checkDescriptorForGetter(ctx, descriptorArg, opts)
				}
			},
		}
	},
}

// checkDescriptorForGetter checks property descriptors for get functions
func checkDescriptorForGetter(ctx rule.RuleContext, descriptor *ast.Node, opts Options) {
	if descriptor == nil || descriptor.Kind != ast.KindObjectLiteralExpression {
		return
	}

	props := descriptor.Properties()
	for _, prop := range props {
		if prop == nil {
			continue
		}

		// Look for 'get' property
		if prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindMethodDeclaration {
			var propName string
			var propNameNode *ast.Node
			if prop.Name() != nil {
				propNameNode = prop.Name()
				switch propNameNode.Kind {
				case ast.KindIdentifier:
					propName = propNameNode.Text()
				case ast.KindStringLiteral:
					propName = propNameNode.Text()
					// Remove quotes
					if len(propName) >= 2 {
						propName = propName[1 : len(propName)-1]
					}
				}
			}

			if propName == "get" {
				// Found a getter
				var getterFunc *ast.Node
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					getterFunc = prop.Initializer()
				case ast.KindMethodDeclaration:
					getterFunc = prop
				}

				if getterFunc != nil {
					if getterFunc.Kind == ast.KindFunctionExpression ||
						getterFunc.Kind == ast.KindArrowFunction ||
						getterFunc.Kind == ast.KindMethodDeclaration {
						reportGetterReturn(ctx, getterFunc, propNameNode, opts)
					}
				}
			}
		}
	}
}
