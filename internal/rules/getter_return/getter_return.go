package getter_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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

// checkGetterReturn checks if a getter function has proper return statements
func checkGetterReturn(ctx rule.RuleContext, node *ast.Node, opts Options) {
	if node == nil {
		return
	}

	body := node.Body()
	if body == nil {
		return
	}

	// If allowImplicit is true, we don't check for return values
	if opts.AllowImplicit {
		return
	}

	// Perform control flow analysis
	result := analyzeReturnPaths(body)

	// Report on the getter node itself
	if result.hasNoReturns {
		// No return statements at all, or only empty returns
		ctx.ReportNode(node, buildExpectedMessage())
	} else if !result.allPathsReturn {
		// Some paths return a value, but not all paths do
		ctx.ReportNode(node, buildExpectedAlwaysMessage())
	}
}

// returnAnalysisResult holds the result of control flow analysis
type returnAnalysisResult struct {
	hasNoReturns    bool // true if there are no return statements with values
	allPathsReturn  bool // true if all code paths return a value
}

// analyzeReturnPaths performs control flow analysis on a function body
func analyzeReturnPaths(body *ast.Node) returnAnalysisResult {
	if body == nil {
		return returnAnalysisResult{hasNoReturns: true, allPathsReturn: false}
	}

	hasReturnWithValue := false
	hasReturnWithoutValue := false

	// Use ForEachReturnStatement to find all return statements
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		expr := stmt.Expression()
		if expr != nil {
			hasReturnWithValue = true
		} else {
			hasReturnWithoutValue = true
		}
		return false // Continue iterating
	})

	// Determine if this is a simple case or complex case
	// Simple case: body is a block with a single return statement
	isSingleReturn := false
	if body.Kind == ast.KindBlock {
		statements := body.Statements()
		if len(statements) == 1 && statements[0].Kind == ast.KindReturnStatement {
			isSingleReturn = true
		}
	}

	// If we have no return with value at all, report "expected"
	if !hasReturnWithValue {
		return returnAnalysisResult{
			hasNoReturns:   true,
			allPathsReturn: false,
		}
	}

	// Heuristic for determining if all paths return:
	// 1. If it's a single return statement, yes
	// 2. If we have both return with value and return without value, no (inconsistent)
	// 3. If we have only returns with values and no control flow, yes
	// 4. If we have only returns with values and control flow, we need more analysis
	//    For now, we'll be conservative: assume all paths return if there are multiple returns with values
	//    and no empty returns (this handles if-else cases)

	countReturnsWithValue := 0
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		if stmt.Expression() != nil {
			countReturnsWithValue++
		}
		return false
	})

	allPathsReturn := isSingleReturn ||
		(!hasReturnWithoutValue && (isSimpleBody(body) || countReturnsWithValue >= 2))

	return returnAnalysisResult{
		hasNoReturns:   false,
		allPathsReturn: allPathsReturn,
	}
}

// isSimpleBody checks if a function body is simple enough that we can assume all paths return
// if there's at least one return with value
func isSimpleBody(body *ast.Node) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return true
	}

	statements := body.Statements()
	if len(statements) == 0 {
		return true
	}

	// Check for control flow statements (if, switch, loops, etc.)
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindIfStatement, ast.KindSwitchStatement,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement,
			ast.KindTryStatement:
			// Has control flow - not simple
			return false
		}
	}

	return true
}

// GetterReturnRule enforces return statements in getters
var GetterReturnRule = rule.CreateRule(rule.Rule{
	Name: "getter-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindGetAccessor: func(node *ast.Node) {
				checkGetterReturn(ctx, node, opts)
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
				// This handles both regular PropertyAccessExpression and optional chaining (checked via flags)
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
})

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
				if propNameNode.Kind == ast.KindIdentifier {
					propName = propNameNode.Text()
				} else if propNameNode.Kind == ast.KindStringLiteral {
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
				if prop.Kind == ast.KindPropertyAssignment {
					getterFunc = prop.Initializer()
				} else if prop.Kind == ast.KindMethodDeclaration {
					getterFunc = prop
				}

				if getterFunc != nil {
					if getterFunc.Kind == ast.KindFunctionExpression ||
						getterFunc.Kind == ast.KindArrowFunction ||
						getterFunc.Kind == ast.KindMethodDeclaration {
						// Report on the 'get' property name, not the function
						checkGetterReturnInDescriptor(ctx, getterFunc, propNameNode, opts)
					}
				}
			}
		}
	}
}

// checkGetterReturnInDescriptor is like checkGetterReturn but reports on a specific node
func checkGetterReturnInDescriptor(ctx rule.RuleContext, funcNode *ast.Node, reportNode *ast.Node, opts Options) {
	if funcNode == nil {
		return
	}

	body := funcNode.Body()
	if body == nil {
		return
	}

	// If allowImplicit is true, we don't check for return values
	if opts.AllowImplicit {
		return
	}

	// Perform control flow analysis
	result := analyzeReturnPaths(body)

	// Use the reportNode for error reporting (e.g., the 'get' property name)
	// Fall back to function node if reportNode is nil
	if reportNode == nil {
		reportNode = funcNode
	}

	if result.hasNoReturns {
		ctx.ReportNode(reportNode, buildExpectedMessage())
	} else if !result.allPathsReturn {
		ctx.ReportNode(reportNode, buildExpectedAlwaysMessage())
	}
}
