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
	hasNoReturns   bool // true if there are no return statements with values
	allPathsReturn bool // true if all code paths return a value
}

// pathResult represents how a code path terminates
type pathResult int

const (
	pathFallthrough  pathResult = iota // path falls through (no return/throw)
	pathReturns                        // path returns a value
	pathTerminates                     // path terminates without returning a value (throw, empty return)
	pathEmptyReturn                    // path has an explicit empty return
)

// analyzeReturnPaths performs control flow analysis on a function body
func analyzeReturnPaths(body *ast.Node) returnAnalysisResult {
	if body == nil {
		return returnAnalysisResult{hasNoReturns: true, allPathsReturn: false}
	}

	hasReturnWithValue := false
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		if stmt.Expression() != nil {
			hasReturnWithValue = true
		}
		return false
	})

	// Analyze control flow
	if body.Kind != ast.KindBlock {
		if hasReturnWithValue {
			return returnAnalysisResult{hasNoReturns: false, allPathsReturn: true}
		}
		return returnAnalysisResult{hasNoReturns: true, allPathsReturn: false}
	}

	result := analyzeStatements(body.Statements())

	if !hasReturnWithValue {
		// No return with value: valid only if all paths terminate (e.g., all throw)
		if result == pathTerminates {
			return returnAnalysisResult{hasNoReturns: false, allPathsReturn: true}
		}
		return returnAnalysisResult{hasNoReturns: true, allPathsReturn: false}
	}

	// Has return with value: check if all paths return or terminate
	allPathsReturn := result == pathReturns || result == pathTerminates

	return returnAnalysisResult{
		hasNoReturns:   false,
		allPathsReturn: allPathsReturn,
	}
}

// analyzeStatements analyzes a list of statements to determine how the path terminates
func analyzeStatements(statements []*ast.Node) pathResult {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		r := analyzeStatement(stmt)
		if r != pathFallthrough {
			return r
		}
	}
	return pathFallthrough
}

// analyzeStatement analyzes a single statement to determine how its path terminates
func analyzeStatement(stmt *ast.Node) pathResult {
	if stmt == nil {
		return pathFallthrough
	}

	switch stmt.Kind {
	case ast.KindReturnStatement:
		if stmt.Expression() != nil {
			return pathReturns
		}
		return pathEmptyReturn

	case ast.KindThrowStatement:
		return pathTerminates

	case ast.KindIfStatement:
		return analyzeIfStatement(stmt)

	case ast.KindSwitchStatement:
		return analyzeSwitchStatement(stmt)

	case ast.KindTryStatement:
		return analyzeTryStatement(stmt)

	case ast.KindBlock:
		return analyzeStatements(stmt.Statements())
	}

	return pathFallthrough
}

// analyzeIfStatement analyzes if/else branches
func analyzeIfStatement(stmt *ast.Node) pathResult {
	ifStmt := stmt.AsIfStatement()
	if ifStmt == nil {
		return pathFallthrough
	}

	thenResult := analyzeBranch(ifStmt.ThenStatement)

	if ifStmt.ElseStatement == nil {
		// No else: can't guarantee all paths are covered
		return pathFallthrough
	}

	elseResult := analyzeBranch(ifStmt.ElseStatement)

	return combineResults(thenResult, elseResult)
}

// analyzeSwitchStatement analyzes switch/case branches
func analyzeSwitchStatement(stmt *ast.Node) pathResult {
	switchStmt := stmt.AsSwitchStatement()
	if switchStmt == nil || switchStmt.CaseBlock == nil {
		return pathFallthrough
	}

	// Get clauses from case block using ForEachChild
	var clauses []*ast.Node
	switchStmt.CaseBlock.ForEachChild(func(child *ast.Node) bool {
		if child != nil && (child.Kind == ast.KindCaseClause || child.Kind == ast.KindDefaultClause) {
			clauses = append(clauses, child)
		}
		return false
	})

	if len(clauses) == 0 {
		return pathFallthrough
	}

	hasDefault := false
	combined := pathReturns // start optimistic

	for _, clause := range clauses {
		if clause == nil {
			continue
		}
		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
		}

		caseOrDefault := clause.AsCaseOrDefaultClause()
		if caseOrDefault == nil || caseOrDefault.Statements == nil || len(caseOrDefault.Statements.Nodes) == 0 {
			// Empty clause falls through to next
			continue
		}

		clauseResult := analyzeStatements(caseOrDefault.Statements.Nodes)
		combined = combineResults(combined, clauseResult)
	}

	if !hasDefault {
		// Without default, not all paths are covered
		return pathFallthrough
	}

	return combined
}

// analyzeTryStatement analyzes try/catch/finally
func analyzeTryStatement(stmt *ast.Node) pathResult {
	tryStmt := stmt.AsTryStatement()
	if tryStmt == nil {
		return pathFallthrough
	}

	// If there's a finally block that terminates, the whole thing terminates
	if tryStmt.FinallyBlock != nil {
		finallyResult := analyzeStatements(tryStmt.FinallyBlock.Statements())
		if finallyResult == pathReturns || finallyResult == pathTerminates {
			return finallyResult
		}
	}

	tryResult := pathFallthrough
	if tryStmt.TryBlock != nil {
		tryResult = analyzeStatements(tryStmt.TryBlock.Statements())
	}

	if tryStmt.CatchClause != nil {
		catchBlock := tryStmt.CatchClause.AsCatchClause().Block
		catchResult := pathFallthrough
		if catchBlock != nil {
			catchResult = analyzeStatements(catchBlock.Statements())
		}
		return combineResults(tryResult, catchResult)
	}

	return tryResult
}

// analyzeBranch analyzes a branch (statement or block)
func analyzeBranch(stmt *ast.Node) pathResult {
	if stmt == nil {
		return pathFallthrough
	}

	if stmt.Kind == ast.KindBlock {
		return analyzeStatements(stmt.Statements())
	}

	return analyzeStatement(stmt)
}

// combineResults combines results from two branches (e.g., if/else)
// Both branches must return for the combined result to be "returns"
func combineResults(a, b pathResult) pathResult {
	if a == pathFallthrough || b == pathFallthrough {
		return pathFallthrough
	}
	if a == pathEmptyReturn || b == pathEmptyReturn {
		return pathEmptyReturn
	}
	// Both branches either return or terminate
	if a == pathReturns && b == pathReturns {
		return pathReturns
	}
	// At least one terminates (throw), the other returns or terminates
	if a == pathReturns || b == pathReturns {
		return pathReturns
	}
	return pathTerminates
}

// GetterReturnRule enforces return statements in getters
var GetterReturnRule = rule.Rule{
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
