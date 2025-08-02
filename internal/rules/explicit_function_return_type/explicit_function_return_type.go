package explicit_function_return_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type ExplicitFunctionReturnTypeOptions struct {
	AllowConciseArrowFunctionExpressionsStartingWithVoid bool     `json:"allowConciseArrowFunctionExpressionsStartingWithVoid"`
	AllowDirectConstAssertionInArrowFunctions            bool     `json:"allowDirectConstAssertionInArrowFunctions"`
	AllowedNames                                          []string `json:"allowedNames"`
	AllowExpressions                                      bool     `json:"allowExpressions"`
	AllowFunctionsWithoutTypeParameters                   bool     `json:"allowFunctionsWithoutTypeParameters"`
	AllowHigherOrderFunctions                             bool     `json:"allowHigherOrderFunctions"`
	AllowIIFEs                                            bool     `json:"allowIIFEs"`
	AllowTypedFunctionExpressions                         bool     `json:"allowTypedFunctionExpressions"`
}

type functionInfo struct {
	node    *ast.Node
	returns []*ast.Node
}

func buildMissingReturnTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnType",
		Description: "Missing return type on function.",
	}
}

// Check if a function is an IIFE (Immediately Invoked Function Expression)
func isIIFE(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindCallExpression
}

// Check if the function has a return type annotation
func hasReturnType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		return arrowFunc.Type != nil
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		return funcDecl.Type != nil
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		return funcExpr.Type != nil
	}
	return false
}

// Check if arrow function starts with void operator
func startsWithVoid(node *ast.Node) bool {
	if node.Kind != ast.KindArrowFunction {
		return false
	}
	
	arrowFunc := node.AsArrowFunction()
	if arrowFunc.Body == nil || arrowFunc.Body.Kind != ast.KindBlock {
		// Check if it's a concise arrow function with void expression
		body := arrowFunc.Body
		if body != nil && body.Kind == ast.KindPrefixUnaryExpression {
			unary := body.AsPrefixUnaryExpression()
			return unary.Operator == ast.KindVoidKeyword
		}
	}
	return false
}

// Check if arrow function directly returns as const
func hasDirectConstAssertion(node *ast.Node, ctx rule.RuleContext) bool {
	if node.Kind != ast.KindArrowFunction {
		return false
	}
	
	arrowFunc := node.AsArrowFunction()
	if arrowFunc.Body == nil || arrowFunc.Body.Kind == ast.KindBlock {
		return false
	}
	
	// Check for as const expression
	body := arrowFunc.Body
	if body.Kind == ast.KindAsExpression {
		asExpr := body.AsAsExpression()
		if asExpr.Type != nil && asExpr.Type.Kind == ast.KindTypeReference {
			typeRef := asExpr.Type.AsTypeReference()
			if ast.IsIdentifier(typeRef.TypeName) {
				typeName := typeRef.TypeName.AsIdentifier()
				return typeName.Text == "const"
			}
		}
	}
	
	// Check for satisfies ... as const pattern
	if body.Kind == ast.KindSatisfiesExpression {
		satisfiesExpr := body.AsSatisfiesExpression()
		expr := satisfiesExpr.Expression
		for expr != nil && expr.Kind == ast.KindSatisfiesExpression {
			expr = expr.AsSatisfiesExpression().Expression
		}
		if expr != nil && expr.Kind == ast.KindAsExpression {
			asExpr := expr.AsAsExpression()
			if asExpr.Type != nil && asExpr.Type.Kind == ast.KindTypeReference {
				typeRef := asExpr.Type.AsTypeReference()
				if ast.IsIdentifier(typeRef.TypeName) {
					typeName := typeRef.TypeName.AsIdentifier()
					return typeName.Text == "const"
				}
			}
		}
	}
	
	return false
}

// Get the function name from various contexts
func getFunctionName(node *ast.Node, ctx rule.RuleContext) string {
	// Check if function has direct name
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil {
			return funcDecl.Name().Text()
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr.Name() != nil {
			return funcExpr.Name().Text()
		}
	}
	
	// Check parent context for name
	parent := node.Parent
	if parent == nil {
		return ""
	}
	
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		if varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier {
			return varDecl.Name().AsIdentifier().Text
		}
	case ast.KindMethodDeclaration, ast.KindPropertyDeclaration, ast.KindPropertyAssignment:
		name, _ := utils.GetNameFromMember(ctx.SourceFile, parent)
		if name != "" {
			return name
		}
	}
	
	return ""
}

// Check if function is allowed based on options
func isAllowedFunction(node *ast.Node, opts ExplicitFunctionReturnTypeOptions) bool {
	// Check allowFunctionsWithoutTypeParameters
	if opts.AllowFunctionsWithoutTypeParameters {
		hasTypeParams := false
		switch node.Kind {
		case ast.KindArrowFunction:
			hasTypeParams = node.AsArrowFunction().TypeParameters != nil
		case ast.KindFunctionDeclaration:
			hasTypeParams = node.AsFunctionDeclaration().TypeParameters != nil
		case ast.KindFunctionExpression:
			hasTypeParams = node.AsFunctionExpression().TypeParameters != nil
		}
		if !hasTypeParams {
			return true
		}
	}
	
	// Check allowIIFEs
	if opts.AllowIIFEs && isIIFE(node) {
		return true
	}
	
	// Check allowedNames
	if len(opts.AllowedNames) > 0 {
		// Note: This would need context to get function name properly
		// For now, skip this check as it requires refactoring
	}
	
	return false
}

// Check if the function expression has valid return type through its context
func isValidFunctionExpressionReturnType(node *ast.Node, opts ExplicitFunctionReturnTypeOptions, ctx rule.RuleContext) bool {
	if !opts.AllowTypedFunctionExpressions {
		return false
	}
	
	// Already has return type
	if hasReturnType(node) {
		return true
	}
	
	parent := node.Parent
	if parent == nil {
		return false
	}
	
	checker := ctx.TypeChecker
	if checker == nil {
		return false
	}
	
	// Check various parent contexts for type information
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		// Check if variable has type annotation
		varDecl := parent.AsVariableDeclaration()
		if varDecl.Type != nil {
			return true
		}
		// Check if variable has initializer with type assertion
		if varDecl.Initializer != nil && (varDecl.Initializer.Kind == ast.KindAsExpression || varDecl.Initializer.Kind == ast.KindTypeAssertionExpression) {
			return true
		}
		
	case ast.KindPropertyDeclaration, ast.KindPropertyAssignment:
		// Check if property has type annotation
		if parent.Kind == ast.KindPropertyDeclaration {
			propDecl := parent.AsPropertyDeclaration()
			if propDecl.Type != nil {
				return true
			}
		}
		
		// Check if parent object has type
		grandParent := parent.Parent
		if grandParent != nil {
			switch grandParent.Kind {
			case ast.KindObjectLiteralExpression:
				// Check if object literal is in typed context
				objParent := grandParent.Parent
				if objParent != nil {
					switch objParent.Kind {
					case ast.KindVariableDeclaration:
						varDecl := objParent.AsVariableDeclaration()
						return varDecl.Type != nil
					case ast.KindAsExpression, ast.KindTypeAssertionExpression:
						return true
					case ast.KindReturnStatement:
						// Check if we're in a typed function
						return isInTypedContext(objParent, checker)
					}
				}
			}
		}
		
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return true
		
	case ast.KindCallExpression:
		// Check if it's a typed function parameter
		return isTypedFunctionParameter(node, parent, checker)
		
	case ast.KindArrayLiteralExpression:
		// Check if array is in typed context
		return isInTypedContext(parent, checker)
		
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		// JSX props are typed
		return true
		
	case ast.KindJsxExpression:
		// JSX expression container
		if parent.Parent != nil && (parent.Parent.Kind == ast.KindJsxElement || parent.Parent.Kind == ast.KindJsxSelfClosingElement) {
			return true
		}
	}
	
	return false
}

// Check if we're in a typed context (has contextual type)
func isInTypedContext(node *ast.Node, checker *checker.Checker) bool {
	contextualType := checker.GetContextualType(node, 0)
	return contextualType != nil
}

// Check if function is a parameter to a typed call
func isTypedFunctionParameter(funcNode, callNode *ast.Node, checker *checker.Checker) bool {
	call := callNode.AsCallExpression()
	
	// Find which argument position this function is in
	argIndex := -1
	for i, arg := range call.Arguments.Nodes {
		if arg == funcNode {
			argIndex = i
			break
		}
	}
	
	if argIndex == -1 {
		return false
	}
	
	// Get the signature of the called function
	signature := checker.GetResolvedSignature(callNode)
	if signature == nil {
		return false
	}
	
	// Check if the parameter at this position expects a function type
	params := signature.Parameters()
	if argIndex < len(params) {
		paramType := checker.GetTypeOfSymbol(params[argIndex])
		if paramType != nil {
			// Enhanced function type detection using TypeScript's type checking
			// Check if the parameter type is a function type by looking for call signatures
			callSignatures := checker.GetSignaturesOfType(paramType, 0) // SignatureKindCall = 0
			if len(callSignatures) > 0 {
				// This parameter expects a function - return type information is available
				return true
			}
			
			// Check if it's a constructor type  
			constructSignatures := checker.GetSignaturesOfType(paramType, 1) // SignatureKindConstruct = 1
			if len(constructSignatures) > 0 {
				return true
			}
		}
	}
	
	return false
}

// Check if any ancestor provides return type information
func ancestorHasReturnType(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression:
			if hasReturnType(parent) {
				return true
			}
		case ast.KindVariableDeclaration:
			varDecl := parent.AsVariableDeclaration()
			if varDecl.Type != nil {
				return true
			}
		case ast.KindPropertyDeclaration:
			propDecl := parent.AsPropertyDeclaration()
			if propDecl.Type != nil {
				return true
			}
		case ast.KindMethodDeclaration:
			methodDecl := parent.AsMethodDeclaration()
			if methodDecl.Type != nil {
				return true
			}
		}
		parent = parent.Parent
	}
	return false
}

// Check if the function is a higher-order function that immediately returns another function
func isHigherOrderFunction(info *functionInfo) bool {
	// Function must have exactly one return statement
	if len(info.returns) != 1 {
		return false
	}
	
	returnStmt := info.returns[0]
	if returnStmt.Kind != ast.KindReturnStatement {
		return false
	}
	
	returnNode := returnStmt.AsReturnStatement()
	if returnNode.Expression == nil {
		return false
	}
	
	// Check if return expression is a function
	expr := returnNode.Expression
	return expr.Kind == ast.KindArrowFunction ||
		expr.Kind == ast.KindFunctionExpression
}

// Get the location to report the error
func getReportLocation(node *ast.Node, ctx rule.RuleContext) (int, int) {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		// Report at "function" keyword
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil {
			return node.Pos(), funcDecl.Name().Pos()
		}
		return node.Pos(), node.Pos() + 8 // "function" length
		
	case ast.KindFunctionExpression:
		// Report at "function" keyword
		funcExpr := node.AsFunctionExpression()
		if funcExpr.Name() != nil {
			return node.Pos(), funcExpr.Name().End()
		}
		return node.Pos(), node.Pos() + 8 // "function" length
		
	case ast.KindArrowFunction:
		// Report at arrow
		arrow := node.AsArrowFunction()
		// Find the arrow position (after parameters)
		if arrow.Parameters != nil && len(arrow.Parameters.Nodes) > 0 {
			lastParam := arrow.Parameters.Nodes[len(arrow.Parameters.Nodes)-1]
			// Look for arrow after last parameter
			arrowPos := lastParam.End()
			// Skip whitespace and find =>
			text := ctx.SourceFile.Text()
			for i := arrowPos; i < len(text)-1; i++ {
				if text[i] == '=' && text[i+1] == '>' {
					return i, i + 2
				}
			}
		}
		// Fallback to node position
		return node.Pos(), node.Pos() + 2
	}
	
	return node.Pos(), node.End()
}

var ExplicitFunctionReturnTypeRule = rule.Rule{
	Name: "explicit-function-return-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Initialize options with defaults
		opts := ExplicitFunctionReturnTypeOptions{
			AllowConciseArrowFunctionExpressionsStartingWithVoid: false,
			AllowDirectConstAssertionInArrowFunctions:            true,
			AllowedNames:                                          []string{},
			AllowExpressions:                                      false,
			AllowFunctionsWithoutTypeParameters:                   false,
			AllowHigherOrderFunctions:                             true,
			AllowIIFEs:                                            false,
			AllowTypedFunctionExpressions:                         true,
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
				if val, ok := optsMap["allowConciseArrowFunctionExpressionsStartingWithVoid"].(bool); ok {
					opts.AllowConciseArrowFunctionExpressionsStartingWithVoid = val
				}
				if val, ok := optsMap["allowDirectConstAssertionInArrowFunctions"].(bool); ok {
					opts.AllowDirectConstAssertionInArrowFunctions = val
				}
				if val, ok := optsMap["allowedNames"].([]interface{}); ok {
					opts.AllowedNames = make([]string, 0, len(val))
					for _, v := range val {
						if s, ok := v.(string); ok {
							opts.AllowedNames = append(opts.AllowedNames, s)
						}
					}
				}
				if val, ok := optsMap["allowExpressions"].(bool); ok {
					opts.AllowExpressions = val
				}
				if val, ok := optsMap["allowFunctionsWithoutTypeParameters"].(bool); ok {
					opts.AllowFunctionsWithoutTypeParameters = val
				}
				if val, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
					opts.AllowHigherOrderFunctions = val
				}
				if val, ok := optsMap["allowIIFEs"].(bool); ok {
					opts.AllowIIFEs = val
				}
				if val, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
					opts.AllowTypedFunctionExpressions = val
				}
			}
		}
		
		// Stack to track function information
		functionStack := make([]*functionInfo, 0)
		
		// Helper to push function onto stack
		enterFunction := func(node *ast.Node) {
			functionStack = append(functionStack, &functionInfo{
				node:    node,
				returns: make([]*ast.Node, 0),
			})
		}
		
		// Helper to pop function from stack
		exitFunction := func() *functionInfo {
			if len(functionStack) == 0 {
				return nil
			}
			info := functionStack[len(functionStack)-1]
			functionStack = functionStack[:len(functionStack)-1]
			return info
		}
		
		// Helper to get current function info
		currentFunction := func() *functionInfo {
			if len(functionStack) == 0 {
				return nil
			}
			return functionStack[len(functionStack)-1]
		}
		
		// Check function expression (arrow or function expression)
		checkFunctionExpression := func(node *ast.Node) {
			info := exitFunction()
			if info == nil {
				return
			}
			
			// Special case: arrow function with void
			if opts.AllowConciseArrowFunctionExpressionsStartingWithVoid && startsWithVoid(node) {
				return
			}
			
			// Special case: arrow function with as const
			if opts.AllowDirectConstAssertionInArrowFunctions && hasDirectConstAssertion(node, ctx) {
				return
			}
			
			// Check if function is allowed
			if isAllowedFunction(node, opts) {
				return
			}
			
			// Check if it's a typed function expression
			if opts.AllowTypedFunctionExpressions &&
				(isValidFunctionExpressionReturnType(node, opts, ctx) || ancestorHasReturnType(node)) {
				return
			}
			
			// Check if it's a higher-order function
			if opts.AllowHigherOrderFunctions && isHigherOrderFunction(info) {
				return
			}
			
			// Check if expressions are allowed
			if opts.AllowExpressions {
				parent := node.Parent
				if parent != nil {
					switch parent.Kind {
					case ast.KindCallExpression,
						ast.KindArrayLiteralExpression,
						ast.KindParenthesizedExpression,
						ast.KindExpressionStatement:
						return
					}
				}
			}
			
			// Report missing return type
			start, end := getReportLocation(node, ctx)
			ctx.ReportRange(core.NewTextRange(start, end), buildMissingReturnTypeMessage())
		}
		
		// Check function declaration
		checkFunctionDeclaration := func(node *ast.Node) {
			info := exitFunction()
			if info == nil {
				return
			}
			
			// Check if function is allowed
			if isAllowedFunction(node, opts) {
				return
			}
			
			// Function declarations with return type are always ok
			if hasReturnType(node) {
				return
			}
			
			// Check if typed function expressions are allowed (for consistency)
			if opts.AllowTypedFunctionExpressions && hasReturnType(node) {
				return
			}
			
			// Check if it's a higher-order function
			if opts.AllowHigherOrderFunctions && isHigherOrderFunction(info) {
				return
			}
			
			// Check if expressions are allowed (export default)
			if opts.AllowExpressions {
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindExportAssignment {
					return
				}
			}
			
			// Report missing return type
			start, end := getReportLocation(node, ctx)
			ctx.ReportRange(core.NewTextRange(start, end), buildMissingReturnTypeMessage())
		}
		
		return rule.RuleListeners{
			ast.KindArrowFunction: enterFunction,
			ast.KindFunctionDeclaration: enterFunction,
			ast.KindFunctionExpression: enterFunction,
			
			rule.ListenerOnExit(ast.KindArrowFunction): checkFunctionExpression,
			rule.ListenerOnExit(ast.KindFunctionExpression): checkFunctionExpression,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): checkFunctionDeclaration,
			
			ast.KindReturnStatement: func(node *ast.Node) {
				if info := currentFunction(); info != nil {
					info.returns = append(info.returns, node)
				}
			},
		}
	},
}