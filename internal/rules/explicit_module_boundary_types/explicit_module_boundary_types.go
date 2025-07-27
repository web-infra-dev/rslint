package explicit_module_boundary_types

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type ExplicitModuleBoundaryTypesOptions struct {
	AllowArgumentsExplicitlyTypedAsAny           bool     `json:"allowArgumentsExplicitlyTypedAsAny"`
	AllowDirectConstAssertionInArrowFunctions    bool     `json:"allowDirectConstAssertionInArrowFunctions"`
	AllowedNames                                 []string `json:"allowedNames"`
	AllowHigherOrderFunctions                    bool     `json:"allowHigherOrderFunctions"`
	AllowTypedFunctionExpressions                bool     `json:"allowTypedFunctionExpressions"`
	AllowOverloadFunctions                       bool     `json:"allowOverloadFunctions"`
}

type functionInfo struct {
	node    *ast.Node
	returns []*ast.Node
}

// Message builders
func buildAnyTypedArgMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "anyTypedArg",
		Description: fmt.Sprintf("Argument '%s' should be typed with a non-any type.", name),
	}
}

func buildAnyTypedArgUnnamedMessage(paramType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "anyTypedArgUnnamed",
		Description: fmt.Sprintf("%s argument should be typed with a non-any type.", paramType),
	}
}

func buildMissingArgTypeMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingArgType",
		Description: fmt.Sprintf("Argument '%s' should be typed.", name),
	}
}

func buildMissingArgTypeUnnamedMessage(paramType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingArgTypeUnnamed",
		Description: fmt.Sprintf("%s argument should be typed.", paramType),
	}
}

func buildMissingReturnTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnType",
		Description: "Missing return type on function.",
	}
}

// Helper to check if a function has a return type
func hasReturnType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Type != nil
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Type != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Type != nil
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration().Type != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Type != nil
	case ast.KindConstructor:
		// Constructors don't need return type
		return true
	case ast.KindSetAccessor:
		// Set accessors don't need return type
		return true
	}
	return false
}

// Check if node is a function
func isFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return true
	}
	return false
}

// Check if arrow function directly returns as const
func hasDirectConstAssertion(node *ast.Node) bool {
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
		// Check if the expression itself is an as const
		expr := satisfiesExpr.Expression
		if expr.Kind == ast.KindAsExpression {
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

// Check if function immediately returns another function expression
func doesImmediatelyReturnFunctionExpression(info functionInfo) bool {
	returns := info.returns
	
	// Should have exactly one return statement
	if len(returns) != 1 {
		return false
	}
	
	returnStatement := returns[0]
	if returnStatement.AsReturnStatement().Expression == nil {
		return false
	}
	
	expr := returnStatement.AsReturnStatement().Expression
	return isFunction(expr)
}

// Check if ancestor has return type
func ancestorHasReturnType(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if isFunction(parent) && hasReturnType(parent) {
			return true
		}
		parent = parent.Parent
	}
	return false
}

// Check if function expression is typed
func isTypedFunctionExpression(node *ast.Node, options ExplicitModuleBoundaryTypesOptions) bool {
	if !options.AllowTypedFunctionExpressions {
		return false
	}
	
	parent := node.Parent
	if parent == nil {
		return false
	}
	
	// Variable declarator with type annotation
	if parent.Kind == ast.KindVariableDeclaration {
		varDecl := parent.AsVariableDeclaration()
		return varDecl.Type != nil
	}
	
	// As expression
	if parent.Kind == ast.KindAsExpression {
		return true
	}
	
	// Property with type annotation
	if parent.Kind == ast.KindPropertyAssignment || parent.Kind == ast.KindPropertyDeclaration {
		// Check if the parent object/class has a type
		grandParent := parent.Parent
		if grandParent != nil {
			// Object literal in typed context
			if grandParent.Kind == ast.KindObjectLiteralExpression {
				ggParent := grandParent.Parent
				if ggParent != nil {
					if ggParent.Kind == ast.KindAsExpression {
						return true
					}
					if ggParent.Kind == ast.KindVariableDeclaration {
						varDecl := ggParent.AsVariableDeclaration()
						return varDecl.Type != nil
					}
				}
			}
		}
	}
	
	// Property/method declaration with explicit type
	if parent.Kind == ast.KindPropertyDeclaration {
		propDecl := parent.AsPropertyDeclaration()
		return propDecl.Type != nil
	}
	
	if parent.Kind == ast.KindMethodDeclaration {
		methodDecl := parent.AsMethodDeclaration()
		return methodDecl.Type != nil
	}
	
	return false
}

// Check if function has overload signatures
func hasOverloadSignatures(node *ast.Node, ctx rule.RuleContext) bool {
	// For function declarations, check if there are other declarations with the same name
	if node.Kind == ast.KindFunctionDeclaration {
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() == nil {
			return false
		}
		
		// Check parent (usually SourceFile or Block) for other functions with same name
		parent := node.Parent
		if parent != nil {
			siblings := getChildren(parent)
			if !ast.IsIdentifier(funcDecl.Name()) {
				return false
			}
			funcName := funcDecl.Name().AsIdentifier().Text
			overloadCount := 0
			
			for _, sibling := range siblings {
				if sibling.Kind == ast.KindFunctionDeclaration {
					siblingFunc := sibling.AsFunctionDeclaration()
					if siblingFunc.Name() != nil && ast.IsIdentifier(siblingFunc.Name()) && siblingFunc.Name().AsIdentifier().Text == funcName {
						overloadCount++
						if overloadCount > 1 {
							return true
						}
					}
				}
			}
		}
	}
	
	// For method declarations, check class body
	if node.Kind == ast.KindMethodDeclaration {
		methodDecl := node.AsMethodDeclaration()
		
		// Get method name
		var methodName string
		if ast.IsIdentifier(methodDecl.Name()) {
			methodName = methodDecl.Name().AsIdentifier().Text
		} else {
			return false
		}
		
		// Check class body for other methods with same name
		classBody := node.Parent
		if classBody != nil && classBody.Kind == ast.KindClassStaticBlockDeclaration {
			classDecl := classBody.Parent
			if classDecl != nil && (classDecl.Kind == ast.KindClassDeclaration || classDecl.Kind == ast.KindClassExpression) {
				members := classDecl.Members()
				overloadCount := 0
				
				for _, member := range members {
					if member.Kind == ast.KindMethodDeclaration {
						memberMethod := member.AsMethodDeclaration()
						if ast.IsIdentifier(memberMethod.Name()) && memberMethod.Name().AsIdentifier().Text == methodName {
							overloadCount++
							if overloadCount > 1 {
								return true
							}
						}
					}
				}
			}
		}
	}
	
	return false
}

// Check if function name is in allowed names list
func isAllowedName(node *ast.Node, options ExplicitModuleBoundaryTypesOptions, sourceFile *ast.SourceFile) bool {
	if len(options.AllowedNames) == 0 {
		return false
	}
	
	var name string
	
	switch node.Kind {
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl.Name() != nil && ast.IsIdentifier(varDecl.Name()) {
			name = varDecl.Name().AsIdentifier().Text
		}
		
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Name() != nil && ast.IsIdentifier(funcDecl.Name()) {
			name = funcDecl.Name().AsIdentifier().Text
		}
		
	case ast.KindMethodDeclaration:
		methodDecl := node.AsMethodDeclaration()
		if ast.IsIdentifier(methodDecl.Name()) {
			name = methodDecl.Name().AsIdentifier().Text
		}
		
	case ast.KindPropertyDeclaration, ast.KindPropertyAssignment:
		var memberName *ast.Node
		if node.Kind == ast.KindPropertyDeclaration {
			memberName = node.AsPropertyDeclaration().Name()
		} else {
			memberName = node.AsPropertyAssignment().Name()
		}
		if memberName != nil {
			propertyName, _ := utils.GetNameFromMember(sourceFile, memberName)
			name = propertyName
		}
		
	case ast.KindGetAccessor, ast.KindSetAccessor:
		// For accessors, check the name
		var accessorName *ast.Node
		if node.Kind == ast.KindGetAccessor {
			accessorName = node.AsGetAccessorDeclaration().Name()
		} else {
			accessorName = node.AsSetAccessorDeclaration().Name()
		}
		if ast.IsIdentifier(accessorName) {
			name = accessorName.AsIdentifier().Text
		}
	}
	
	return name != "" && slices.Contains(options.AllowedNames, name)
}

// Get all children nodes of a parent
func getChildren(parent *ast.Node) []*ast.Node {
	switch parent.Kind {
	case ast.KindSourceFile:
		return parent.AsSourceFile().Statements.Nodes
	case ast.KindBlock:
		return parent.AsBlock().Statements.Nodes
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return parent.Members()
	case ast.KindObjectLiteralExpression:
		return parent.AsObjectLiteralExpression().Properties.Nodes
	case ast.KindArrayLiteralExpression:
		return parent.AsArrayLiteralExpression().Elements.Nodes
	case ast.KindExportDeclaration:
		exportDecl := parent.AsExportDeclaration()
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			return exportDecl.ExportClause.AsNamedExports().Elements.Nodes
		}
	}
	return nil
}

var ExplicitModuleBoundaryTypesRule = rule.Rule{
	Name: "explicit-module-boundary-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ExplicitModuleBoundaryTypesOptions{
			AllowArgumentsExplicitlyTypedAsAny:        false,
			AllowDirectConstAssertionInArrowFunctions: true,
			AllowedNames:                              []string{},
			AllowHigherOrderFunctions:                 true,
			AllowTypedFunctionExpressions:             true,
			AllowOverloadFunctions:                    false,
		}
		
		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
				if val, ok := optsMap["allowArgumentsExplicitlyTypedAsAny"].(bool); ok {
					opts.AllowArgumentsExplicitlyTypedAsAny = val
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
				if val, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
					opts.AllowHigherOrderFunctions = val
				}
				if val, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
					opts.AllowTypedFunctionExpressions = val
				}
				if val, ok := optsMap["allowOverloadFunctions"].(bool); ok {
					opts.AllowOverloadFunctions = val
				}
			}
		}
		
		// Track functions we've already checked
		checkedFunctions := make(map[*ast.Node]bool)
		
		// Track function stack for nested functions
		functionStack := []*ast.Node{}
		
		// Map functions to their return statements
		functionReturnsMap := make(map[*ast.Node][]*ast.Node)
		
		// Track all visited nodes to avoid cycles
		alreadyVisited := make(map[*ast.Node]bool)
		
		// Helper to get returns for a function
		getReturnsInFunction := func(node *ast.Node) []*ast.Node {
			if returns, ok := functionReturnsMap[node]; ok {
				return returns
			}
			return []*ast.Node{}
		}
		
		// Helper to check if exported higher order function
		isExportedHigherOrderFunction := func(info functionInfo) bool {
			current := info.node.Parent
			for current != nil {
				if current.Kind == ast.KindReturnStatement {
					// Skip block statement parent
					current = current.Parent.Parent
					continue
				}
				
				if !isFunction(current) {
					return false
				}
				
				returns := getReturnsInFunction(current)
				funcInfo := functionInfo{node: current, returns: returns}
				if !doesImmediatelyReturnFunctionExpression(funcInfo) {
					return false
				}
				
				if checkedFunctions[current] {
					return true
				}
				
				current = current.Parent
			}
			return false
		}
		
		// Check a single parameter
		var checkParameter func(param *ast.Node)
		checkParameter = func(param *ast.Node) {
			report := func(namedMessageId, unnamedMessageId func(string) rule.RuleMessage) {
				switch param.Kind {
				case ast.KindIdentifier:
					ctx.ReportNode(param, namedMessageId(param.AsIdentifier().Text))
					
				case ast.KindArrayBindingPattern:
					ctx.ReportNode(param, unnamedMessageId("Array pattern"))
					
				case ast.KindObjectBindingPattern:
					ctx.ReportNode(param, unnamedMessageId("Object pattern"))
					
				case ast.KindBindingElement:
					restElem := param.AsBindingElement()
					if restElem.Name() != nil && ast.IsIdentifier(restElem.Name()) {
						ctx.ReportNode(param, namedMessageId(restElem.Name().AsIdentifier().Text))
					} else {
						ctx.ReportNode(param, unnamedMessageId("Rest"))
					}
				}
			}
			
			switch param.Kind {
			case ast.KindArrayBindingPattern, ast.KindIdentifier, ast.KindObjectBindingPattern, ast.KindBindingElement:
				hasType := false
				isAnyType := false
				
				// Check if parameter has type annotation
				if param.Kind == ast.KindIdentifier {
					// For identifiers, check parent for type annotation
					parent := param.Parent
					if parent != nil && parent.Kind == ast.KindParameter {
						paramNode := parent.AsParameterDeclaration()
						if paramNode.Type != nil {
							hasType = true
							// Check if it's any type
							if paramNode.Type.Kind == ast.KindAnyKeyword {
								isAnyType = true
							}
						}
					}
				} else {
					// For patterns, they may have type annotation directly
					// This depends on the AST structure
				}
				
				if !hasType {
					report(buildMissingArgTypeMessage, buildMissingArgTypeUnnamedMessage)
				} else if isAnyType && !opts.AllowArgumentsExplicitlyTypedAsAny {
					report(buildAnyTypedArgMessage, buildAnyTypedArgUnnamedMessage)
				}
				
			case ast.KindParameter:
				// Handle TSParameterProperty
				nameNode := param.AsParameterDeclaration().Name()
				if nameNode != nil {
					checkParameter(nameNode)
				}
				
			case ast.KindShorthandPropertyAssignment:
				// Assignment patterns have default values, ignore
				return
			}
		}
		
		// Check function parameters
		checkParameters := func(node *ast.Node) {
			var params []*ast.Node
			
			switch node.Kind {
			case ast.KindArrowFunction:
				params = node.AsArrowFunction().Parameters.Nodes
			case ast.KindFunctionDeclaration:
				params = node.AsFunctionDeclaration().Parameters.Nodes
			case ast.KindFunctionExpression:
				params = node.AsFunctionExpression().Parameters.Nodes
			case ast.KindMethodDeclaration:
				params = node.AsMethodDeclaration().Parameters.Nodes
			case ast.KindConstructor:
				params = node.AsConstructorDeclaration().Parameters.Nodes
			case ast.KindSetAccessor:
				params = node.AsSetAccessorDeclaration().Parameters.Nodes
			case ast.KindGetAccessor:
				params = node.AsGetAccessorDeclaration().Parameters.Nodes
			}
			
			for _, param := range params {
				checkParameter(param)
			}
		}
		
		// Check function expression
		checkFunctionExpression := func(info functionInfo) {
			node := info.node
			if checkedFunctions[node] {
				return
			}
			checkedFunctions[node] = true
			
			if isAllowedName(node.Parent, opts, ctx.SourceFile) ||
				isTypedFunctionExpression(node, opts) ||
				ancestorHasReturnType(node) {
				return
			}
			
			if opts.AllowOverloadFunctions &&
				node.Parent != nil &&
				node.Parent.Kind == ast.KindMethodDeclaration &&
				hasOverloadSignatures(node.Parent, ctx) {
				return
			}
			
			// Check return type
			if !hasReturnType(node) {
				// Special handling for arrow functions with direct const assertion
				if node.Kind == ast.KindArrowFunction &&
					opts.AllowDirectConstAssertionInArrowFunctions &&
					hasDirectConstAssertion(node) {
					// Still need to check parameters
					checkParameters(node)
					return
				}
				
				// Report missing return type
				loc := core.NewTextRange(node.Pos(), node.Pos() + 1)
				if node.Kind == ast.KindArrowFunction {
					arrowFunc := node.AsArrowFunction()
					if arrowFunc.EqualsGreaterThanToken != nil {
						loc = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, arrowFunc.EqualsGreaterThanToken.Pos())
					}
				} else if node.Kind == ast.KindFunctionExpression {
					funcExpr := node.AsFunctionExpression()
					if funcExpr.Name() != nil {
						nameRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, funcExpr.Name().Pos())
						loc = core.NewTextRange(nameRange.Pos(), nameRange.End())
					} else {
						// Anonymous function, use "function" keyword
						loc = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
					}
				}
				
				ctx.ReportRange(loc, buildMissingReturnTypeMessage())
			}
			
			checkParameters(node)
		}
		
		// Check function declaration
		checkFunction := func(info functionInfo) {
			node := info.node
			if checkedFunctions[node] {
				return
			}
			checkedFunctions[node] = true
			
			if isAllowedName(node, opts, ctx.SourceFile) || ancestorHasReturnType(node) {
				return
			}
			
			if opts.AllowOverloadFunctions && hasOverloadSignatures(node, ctx) {
				return
			}
			
			// Check return type
			if !hasReturnType(node) {
				// Get location for error
				funcDecl := node.AsFunctionDeclaration()
				loc := core.NewTextRange(node.Pos(), node.Pos() + 1)
				if funcDecl.Name() != nil {
					nameRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, funcDecl.Name().Pos())
					loc = core.NewTextRange(nameRange.Pos(), nameRange.End())
				} else {
					// Anonymous function, use "function" keyword
					loc = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
				}
				
				ctx.ReportRange(loc, buildMissingReturnTypeMessage())
			}
			
			checkParameters(node)
		}
		
		// Note: checkEmptyBodyFunctionExpression removed as it was unused
		
		// Follow reference to check exported identifiers
		followReference := func(node *ast.Node) {
			if node.Kind != ast.KindIdentifier {
				return
			}
			
			// In a real implementation, we would use the type checker to resolve references
			// For now, we'll do a simplified version
			// This would need proper scope analysis
		}
		
		// Main check node function
		var checkNode func(node *ast.Node)
		checkNode = func(node *ast.Node) {
			if node == nil || alreadyVisited[node] {
				return
			}
			alreadyVisited[node] = true
			
			switch node.Kind {
			case ast.KindArrowFunction, ast.KindFunctionExpression:
				returns := getReturnsInFunction(node)
				checkFunctionExpression(functionInfo{node: node, returns: returns})
				
			case ast.KindArrayLiteralExpression:
				for _, elem := range node.AsArrayLiteralExpression().Elements.Nodes {
					checkNode(elem)
				}
				
			case ast.KindPropertyDeclaration, ast.KindMethodDeclaration:
				// Skip private members
				if node.Kind == ast.KindPropertyDeclaration {
					prop := node.AsPropertyDeclaration()
					if prop.Modifiers() != nil {
						for _, mod := range prop.Modifiers().Nodes {
							if mod.Kind == ast.KindPrivateKeyword {
								return
							}
						}
					}
				} else {
					method := node.AsMethodDeclaration()
					if method.Modifiers() != nil {
						for _, mod := range method.Modifiers().Nodes {
							if mod.Kind == ast.KindPrivateKeyword {
								return
							}
						}
					}
				}
				
				// Check the value/implementation
				if node.Kind == ast.KindPropertyDeclaration {
					prop := node.AsPropertyDeclaration()
					if prop.Initializer != nil {
						checkNode(prop.Initializer)
					}
				} else {
					checkNode(node)
				}
				
			case ast.KindClassDeclaration, ast.KindClassExpression:
				// Check all class members
				for _, member := range node.Members() {
					checkNode(member)
				}
				
			case ast.KindFunctionDeclaration:
				returns := getReturnsInFunction(node)
				checkFunction(functionInfo{node: node, returns: returns})
				
			case ast.KindIdentifier:
				followReference(node)
				
			case ast.KindObjectLiteralExpression:
				for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
					checkNode(prop)
				}
				
			case ast.KindPropertyAssignment:
				checkNode(node.AsPropertyAssignment().Initializer)
				
			case ast.KindVariableDeclarationList:
				for _, decl := range node.AsVariableDeclarationList().Declarations.Nodes {
					checkNode(decl)
				}
				
			case ast.KindVariableDeclaration:
				varDecl := node.AsVariableDeclaration()
				if varDecl.Initializer != nil {
					checkNode(varDecl.Initializer)
				}
			}
		}
		
		return rule.RuleListeners{
			// Track function enters/exits
			ast.KindArrowFunction: func(node *ast.Node) {
				functionStack = append(functionStack, node)
				functionReturnsMap[node] = []*ast.Node{}
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				functionStack = append(functionStack, node)
				functionReturnsMap[node] = []*ast.Node{}
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				functionStack = append(functionStack, node)
				functionReturnsMap[node] = []*ast.Node{}
			},
			
			// Track return statements
			ast.KindReturnStatement: func(node *ast.Node) {
				if len(functionStack) > 0 {
					current := functionStack[len(functionStack)-1]
					functionReturnsMap[current] = append(functionReturnsMap[current], node)
				}
			},
			
			// Handle export declarations
			ast.KindExportAssignment: func(node *ast.Node) {
				// Check the expression being exported
				exportDefault := node.AsExportAssignment()
				checkNode(exportDefault.Expression)
			},
			
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ModuleSpecifier == nil { // Not re-export
					if exportDecl.ExportClause != nil {
						// export { foo, bar }
						if exportDecl.ExportClause.Kind == ast.KindNamedExports {
							for _, spec := range exportDecl.ExportClause.AsNamedExports().Elements.Nodes {
								if spec.Kind == ast.KindExportSpecifier {
									specNode := spec.AsExportSpecifier()
									nameNode := specNode.Name()
									followReference(nameNode)
								}
							}
						}
					}
					// Note: Declaration field handling removed as API changed
				}
			},
			
			// Handle function exits
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
			
			// Program exit - check for exported higher-order functions
			rule.ListenerOnExit(ast.KindSourceFile): func(node *ast.Node) {
				for funcNode, returns := range functionReturnsMap {
					info := functionInfo{node: funcNode, returns: returns}
					if isExportedHigherOrderFunction(info) {
						checkNode(funcNode)
					}
				}
			},
		}
	},
}