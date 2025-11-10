// Package explicit_function_return_type implements the @typescript-eslint/explicit-function-return-type rule.
// This rule enforces explicit return type annotations on functions and methods,
// improving code documentation and type safety by requiring developers to explicitly
// declare what types their functions return.
package explicit_function_return_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ExplicitFunctionReturnTypeOptions struct {
	AllowExpressions                                  bool     `json:"allowExpressions"`
	AllowTypedFunctionExpressions                     bool     `json:"allowTypedFunctionExpressions"`
	AllowHigherOrderFunctions                         bool     `json:"allowHigherOrderFunctions"`
	AllowDirectConstAssertionInArrowFunctions         bool     `json:"allowDirectConstAssertionInArrowFunctions"`
	AllowConciseArrowFunctionExpressionsStartingWithVoid bool `json:"allowConciseArrowFunctionExpressionsStartingWithVoid"`
	AllowFunctionsWithoutTypeParameters               bool     `json:"allowFunctionsWithoutTypeParameters"`
	AllowIIFEs                                        bool     `json:"allowIIFEs"`
	AllowedNames                                      []string `json:"allowedNames"`
}

func parseOptions(options any) ExplicitFunctionReturnTypeOptions {
	opts := ExplicitFunctionReturnTypeOptions{
		AllowExpressions:                                  false,
		AllowTypedFunctionExpressions:                     true,
		AllowHigherOrderFunctions:                         true,
		AllowDirectConstAssertionInArrowFunctions:         true,
		AllowConciseArrowFunctionExpressionsStartingWithVoid: false,
		AllowFunctionsWithoutTypeParameters:               false,
		AllowIIFEs:                                        false,
		AllowedNames:                                      []string{},
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
		if m, ok := optsArray[0].(map[string]interface{}); ok {
			optsMap = m
		}
	} else if m, ok := options.(map[string]interface{}); ok {
		optsMap = m
	}

	if optsMap != nil {
		if v, ok := optsMap["allowExpressions"].(bool); ok {
			opts.AllowExpressions = v
		}
		if v, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
			opts.AllowTypedFunctionExpressions = v
		}
		if v, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
			opts.AllowHigherOrderFunctions = v
		}
		if v, ok := optsMap["allowDirectConstAssertionInArrowFunctions"].(bool); ok {
			opts.AllowDirectConstAssertionInArrowFunctions = v
		}
		if v, ok := optsMap["allowConciseArrowFunctionExpressionsStartingWithVoid"].(bool); ok {
			opts.AllowConciseArrowFunctionExpressionsStartingWithVoid = v
		}
		if v, ok := optsMap["allowFunctionsWithoutTypeParameters"].(bool); ok {
			opts.AllowFunctionsWithoutTypeParameters = v
		}
		if v, ok := optsMap["allowIIFEs"].(bool); ok {
			opts.AllowIIFEs = v
		}
		if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
			for _, name := range allowedNames {
				if str, ok := name.(string); ok {
					opts.AllowedNames = append(opts.AllowedNames, str)
				}
			}
		}
	}

	return opts
}

func buildMissingReturnTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnType",
		Description: "Missing return type on function.",
	}
}

// Check if a function has an explicit return type
func hasReturnType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		return fn != nil && fn.Type != nil
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		return fn != nil && fn.Type != nil
	case ast.KindArrowFunction:
		fn := node.AsArrowFunction()
		return fn != nil && fn.Type != nil
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		return method != nil && method.Type != nil
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		return accessor != nil && accessor.Type != nil
	}
	return false
}

// Check if function has type parameters
func hasTypeParameters(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		return fn != nil && fn.TypeParameters != nil && fn.TypeParameters.Nodes != nil && len(fn.TypeParameters.Nodes) > 0
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		return fn != nil && fn.TypeParameters != nil && fn.TypeParameters.Nodes != nil && len(fn.TypeParameters.Nodes) > 0
	case ast.KindArrowFunction:
		fn := node.AsArrowFunction()
		return fn != nil && fn.TypeParameters != nil && fn.TypeParameters.Nodes != nil && len(fn.TypeParameters.Nodes) > 0
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		return method != nil && method.TypeParameters != nil && method.TypeParameters.Nodes != nil && len(method.TypeParameters.Nodes) > 0
	}
	return false
}

// Check if arrow function body is a const assertion
func isConstAssertion(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for direct const assertion: () => x as const
	if node.Kind == ast.KindAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr != nil && asExpr.Type != nil && asExpr.Type.Kind == ast.KindTypeReference {
			typeRef := asExpr.Type.AsTypeReference()
			if typeRef != nil && ast.IsIdentifier(typeRef.TypeName) {
				ident := typeRef.TypeName.AsIdentifier()
				if ident != nil && ident.Text == "const" {
					return true
				}
			}
		}
	}

	// Check for satisfies with const: () => x as const satisfies R
	if node.Kind == ast.KindSatisfiesExpression {
		satisfiesExpr := node.AsSatisfiesExpression()
		if satisfiesExpr != nil && satisfiesExpr.Expression != nil {
			return isConstAssertion(satisfiesExpr.Expression)
		}
	}

	return false
}

// Check if arrow function starts with void
func startsWithVoid(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindArrowFunction {
		return false
	}

	arrowFn := node.AsArrowFunction()
	if arrowFn == nil || arrowFn.Body == nil {
		return false
	}

	// Check if body is a void expression
	if arrowFn.Body.Kind == ast.KindVoidExpression {
		return true
	}

	return false
}

// Check if node is a typed function expression
func isTypedFunctionExpression(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check for variable declaration with type: const x: Foo = () => {}
	if parent.Kind == ast.KindVariableDeclaration {
		varDecl := parent.AsVariableDeclaration()
		if varDecl != nil && varDecl.Type != nil {
			return true
		}
	}

	// Check for type assertion: (() => {}) as Foo or <Foo>(() => {})
	if parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindTypeAssertionExpression {
		return true
	}

	// Check for property assignment in typed object literal
	if parent.Kind == ast.KindPropertyAssignment {
		// Walk up to find if object literal has type assertion
		for p := parent.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindAsExpression || p.Kind == ast.KindTypeAssertionExpression {
				return true
			}
			if p.Kind == ast.KindVariableDeclaration {
				varDecl := p.AsVariableDeclaration()
				if varDecl != nil && varDecl.Type != nil {
					return true
				}
			}
			// Stop at certain boundaries
			if p.Kind == ast.KindSourceFile || p.Kind == ast.KindBlock {
				break
			}
		}
	}

	// Check for property declaration with type: private method: MethodType = () => {}
	if parent.Kind == ast.KindPropertyDeclaration {
		propDecl := parent.AsPropertyDeclaration()
		if propDecl != nil && propDecl.Type != nil {
			return true
		}
	}

	// Note: Without full type system access, we cannot accurately determine if a function
	// passed as an argument to a call expression or used in JSX has a typed signature
	// from the parameter's type annotation. This may lead to false positives where
	// functions that should be flagged are incorrectly considered typed.
	// Examples that would require type information:
	// - setTimeout(() => {}) - callback is NOT typed
	// - array.map((x) => {}) - callback parameter type comes from array element type
	// - <Component onClick={() => {}} /> - prop type comes from component definition

	return false
}

// Check if node is a higher-order function (returns a function)
// Note: This only checks for direct returns of functions. It does not handle:
// - Conditional returns: () => condition ? () => {} : null
// - Wrapped returns: () => Promise.resolve(() => {})
// - Multiple return paths with different types
// These limitations may cause some higher-order functions to not be detected.
func isHigherOrderFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindArrowFunction:
		arrowFn := node.AsArrowFunction()
		if arrowFn == nil || arrowFn.Body == nil {
			return false
		}

		// Direct return of arrow or function: () => () => {}
		bodyKind := arrowFn.Body.Kind
		if bodyKind == ast.KindArrowFunction || bodyKind == ast.KindFunctionExpression {
			return true
		}

		// Block with return statement
		if bodyKind == ast.KindBlock {
			block := arrowFn.Body.AsBlock()
			if block != nil && block.Statements != nil && block.Statements.Nodes != nil {
				for _, stmt := range block.Statements.Nodes {
					if stmt.Kind == ast.KindReturnStatement {
						retStmt := stmt.AsReturnStatement()
						if retStmt != nil && retStmt.Expression != nil {
							exprKind := retStmt.Expression.Kind
							if exprKind == ast.KindArrowFunction || exprKind == ast.KindFunctionExpression {
								return true
							}
						}
					}
				}
			}
		}

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		var body *ast.Node
		if node.Kind == ast.KindFunctionDeclaration {
			fn := node.AsFunctionDeclaration()
			if fn != nil {
				body = fn.Body
			}
		} else {
			fn := node.AsFunctionExpression()
			if fn != nil {
				body = fn.Body
			}
		}

		if body == nil || body.Kind != ast.KindBlock {
			return false
		}

		block := body.AsBlock()
		if block != nil && block.Statements != nil && block.Statements.Nodes != nil {
			for _, stmt := range block.Statements.Nodes {
				if stmt.Kind == ast.KindReturnStatement {
					retStmt := stmt.AsReturnStatement()
					if retStmt != nil && retStmt.Expression != nil {
						exprKind := retStmt.Expression.Kind
						if exprKind == ast.KindArrowFunction || exprKind == ast.KindFunctionExpression {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// Check if node is an IIFE (Immediately Invoked Function Expression)
func isIIFE(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check if parent is a call expression
	if parent.Kind == ast.KindCallExpression {
		callExpr := parent.AsCallExpression()
		if callExpr != nil && callExpr.Expression == node {
			return true
		}
	}

	// Check for parenthesized IIFE: (function() {})()
	if parent.Kind == ast.KindParenthesizedExpression {
		grandparent := parent.Parent
		if grandparent != nil && grandparent.Kind == ast.KindCallExpression {
			return true
		}
	}

	return false
}

// Check if function is used as an expression
func isExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	// Function declarations are not expressions
	if node.Kind == ast.KindFunctionDeclaration {
		return false
	}

	// Check various expression contexts
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindNewExpression:
		return true
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindParenthesizedExpression:
		return true
	case ast.KindBinaryExpression:
		return true
	case ast.KindConditionalExpression:
		return true
	case ast.KindExportAssignment:
		return true
	case ast.KindReturnStatement:
		return true
	case ast.KindJsxExpression, ast.KindJsxAttribute:
		return true
	case ast.KindPropertyAssignment:
		return true
	}

	return false
}

// Get function name for reporting
func getFunctionName(ctx rule.RuleContext, node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil && fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
			ident := fn.Name().AsIdentifier()
			if ident != nil {
				return ident.Text
			}
		}
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		if fn != nil && fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
			ident := fn.Name().AsIdentifier()
			if ident != nil {
				return ident.Text
			}
		}
		// Check parent for name
		if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
			varDecl := node.Parent.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier {
				ident := varDecl.Name().AsIdentifier()
				if ident != nil {
					return ident.Text
				}
			}
		}
	case ast.KindArrowFunction:
		// Check parent for name
		if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
			varDecl := node.Parent.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier {
				ident := varDecl.Name().AsIdentifier()
				if ident != nil {
					return ident.Text
				}
			}
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
			return name
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil && accessor.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
			return name
		}
	}
	return ""
}

// Check if function name is in allowed list
func isAllowedName(ctx rule.RuleContext, node *ast.Node, allowedNames []string) bool {
	if len(allowedNames) == 0 {
		return false
	}

	name := getFunctionName(ctx, node)
	if name == "" {
		return false
	}

	for _, allowed := range allowedNames {
		if name == allowed {
			return true
		}
	}
	return false
}

// Get the node to report (the function signature part)
func getReportNode(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil && fn.Name() != nil {
			return fn.Name()
		}
	case ast.KindFunctionExpression:
		// Return the "function" keyword position
		return node
	case ast.KindArrowFunction:
		// Return the arrow function node itself for positioning
		return node
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Name() != nil {
			return method.Name()
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil && accessor.Name() != nil {
			return accessor.Name()
		}
	}
	return node
}

// Check if function should be skipped based on options
func shouldSkipFunction(ctx rule.RuleContext, node *ast.Node, opts ExplicitFunctionReturnTypeOptions) bool {
	// Check allowedNames option
	if isAllowedName(ctx, node, opts.AllowedNames) {
		return true
	}

	// Check allowFunctionsWithoutTypeParameters option
	if opts.AllowFunctionsWithoutTypeParameters && !hasTypeParameters(node) {
		return true
	}

	// Check allowExpressions option
	if opts.AllowExpressions && isExpression(node) {
		return true
	}

	// Check allowTypedFunctionExpressions option
	if opts.AllowTypedFunctionExpressions &&
		(node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction) &&
		isTypedFunctionExpression(ctx, node) {
		return true
	}

	// Check allowHigherOrderFunctions option
	if opts.AllowHigherOrderFunctions && isHigherOrderFunction(node) {
		return true
	}

	// Check allowDirectConstAssertionInArrowFunctions option
	if opts.AllowDirectConstAssertionInArrowFunctions && node.Kind == ast.KindArrowFunction {
		arrowFn := node.AsArrowFunction()
		if arrowFn != nil && isConstAssertion(arrowFn.Body) {
			return true
		}
	}

	// Check allowConciseArrowFunctionExpressionsStartingWithVoid option
	if opts.AllowConciseArrowFunctionExpressionsStartingWithVoid && startsWithVoid(node) {
		return true
	}

	// Check allowIIFEs option
	if opts.AllowIIFEs && isIIFE(node) {
		return true
	}

	return false
}

var ExplicitFunctionReturnTypeRule = rule.CreateRule(rule.Rule{
	Name: "explicit-function-return-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkFunction := func(node *ast.Node) {
			// Skip if already has return type
			if hasReturnType(node) {
				return
			}

			// Skip if any of the options indicate this function should be ignored
			if shouldSkipFunction(ctx, node, opts) {
				return
			}

			// Report the missing return type
			reportNode := getReportNode(node)
			ctx.ReportNode(reportNode, buildMissingReturnTypeMessage())
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,
			ast.KindMethodDeclaration:   checkFunction,
			ast.KindGetAccessor:         checkFunction,
		}
	},
})
