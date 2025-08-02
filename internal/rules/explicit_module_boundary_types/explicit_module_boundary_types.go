package explicit_module_boundary_types

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ExplicitModuleBoundaryTypesOptions struct {
	AllowArgumentsExplicitlyTypedAsAny        bool     `json:"allowArgumentsExplicitlyTypedAsAny"`
	AllowDirectConstAssertionInArrowFunctions bool     `json:"allowDirectConstAssertionInArrowFunctions"`
	AllowedNames                              []string `json:"allowedNames"`
	AllowHigherOrderFunctions                 bool     `json:"allowHigherOrderFunctions"`
	AllowTypedFunctionExpressions             bool     `json:"allowTypedFunctionExpressions"`
	AllowOverloadFunctions                    bool     `json:"allowOverloadFunctions"`
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
	node := info.node
	returns := info.returns

	// For arrow functions, check if body is directly a function
	if node.Kind == ast.KindArrowFunction {
		arrowFunc := node.AsArrowFunction()
		if arrowFunc.Body != nil && arrowFunc.Body.Kind != ast.KindBlock {
			// Direct expression body
			return isFunction(arrowFunc.Body)
		}
	}

	// For regular functions or arrow functions with block bodies, check return statements
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
	depth := 0
	for parent != nil && depth < 10 {
		if isFunction(parent) && hasReturnType(parent) {
			return true
		}
		parent = parent.Parent
		depth++
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
			AllowedNames:                  []string{},
			AllowHigherOrderFunctions:     true,
			AllowTypedFunctionExpressions: true,
			AllowOverloadFunctions:        false,
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
				if val, exists := optsMap["allowArgumentsExplicitlyTypedAsAny"]; exists {
					if boolVal, ok := val.(bool); ok {
						opts.AllowArgumentsExplicitlyTypedAsAny = boolVal
					}
				}
				if val, exists := optsMap["allowDirectConstAssertionInArrowFunctions"]; exists {
					if boolVal, ok := val.(bool); ok {
						opts.AllowDirectConstAssertionInArrowFunctions = boolVal
					}
				}
				if val, exists := optsMap["allowedNames"]; exists {
					if arrayVal, ok := val.([]interface{}); ok {
						opts.AllowedNames = make([]string, 0, len(arrayVal))
						for _, v := range arrayVal {
							if s, ok := v.(string); ok {
								opts.AllowedNames = append(opts.AllowedNames, s)
							}
						}
					}
				}
				if val, exists := optsMap["allowHigherOrderFunctions"]; exists {
					if boolVal, ok := val.(bool); ok {
						opts.AllowHigherOrderFunctions = boolVal
					}
				}
				if val, exists := optsMap["allowTypedFunctionExpressions"]; exists {
					if boolVal, ok := val.(bool); ok {
						opts.AllowTypedFunctionExpressions = boolVal
					}
				}
				if val, exists := optsMap["allowOverloadFunctions"]; exists {
					if boolVal, ok := val.(bool); ok {
						opts.AllowOverloadFunctions = boolVal
					}
				}
			}
		}

		// Track return statements for functions
		functionReturnsMap := make(map[*ast.Node][]*ast.Node)
		functionStack := []*ast.Node{}

		// Helper to check if a node is exported
		isExported := func(node *ast.Node) bool {
			// Direct export function
			if node.Kind == ast.KindFunctionDeclaration {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl.Modifiers() != nil {
					for _, mod := range funcDecl.Modifiers().Nodes {
						if mod.Kind == ast.KindExportKeyword {
							return true
						}
					}
				}
			}

			// Check if it's in an export statement - limit depth to avoid infinite loops
			parent := node.Parent
			depth := 0
			for parent != nil && depth < 10 {
				if parent.Kind == ast.KindExportAssignment ||
					parent.Kind == ast.KindExportDeclaration {
					return true
				}
				parent = parent.Parent
				depth++
			}

			return false
		}

		// Removed unused checkParameters function

		// Check if function should be allowed
		checkFunction := func(node *ast.Node) {
			// Only check exported functions
			if !isExported(node) {
				return
			}

			// Simple check for return type
			if !hasReturnType(node) {
				if node.Kind == ast.KindFunctionDeclaration {
					funcDecl := node.AsFunctionDeclaration()
					if funcDecl.Name() != nil {
						ctx.ReportNode(funcDecl.Name(), buildMissingReturnTypeMessage())
					} else {
						ctx.ReportNode(node, buildMissingReturnTypeMessage())
					}
				} else {
					ctx.ReportNode(node, buildMissingReturnTypeMessage())
				}
			}

			// Simple parameter check
			var params []*ast.Node
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				params = node.AsFunctionDeclaration().Parameters.Nodes
			case ast.KindArrowFunction:
				params = node.AsArrowFunction().Parameters.Nodes
			case ast.KindFunctionExpression:
				params = node.AsFunctionExpression().Parameters.Nodes
			}

			for _, param := range params {
				if param.Kind == ast.KindParameter {
					paramNode := param.AsParameterDeclaration()
					if paramNode.Type == nil {
						nameNode := paramNode.Name()
						if nameNode != nil && ast.IsIdentifier(nameNode) {
							ctx.ReportNode(nameNode, buildMissingArgTypeMessage(nameNode.AsIdentifier().Text))
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			// Track function enters for return statement collection
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
			ast.KindMethodDeclaration: func(node *ast.Node) {
				functionStack = append(functionStack, node)
				functionReturnsMap[node] = []*ast.Node{}
			},

			// Track return statements
			ast.KindReturnStatement: func(node *ast.Node) {
				if len(functionStack) > 0 {
					current := functionStack[len(functionStack)-1]
					if functionReturnsMap[current] != nil {
						functionReturnsMap[current] = append(functionReturnsMap[current], node)
					}
				}
			},

			// Check functions on exit
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				checkFunction(node)
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				checkFunction(node)
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				checkFunction(node)
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(node *ast.Node) {
				// Only check public methods in exported classes
				method := node.AsMethodDeclaration()
				isPrivate := false
				if method.Modifiers() != nil {
					for _, mod := range method.Modifiers().Nodes {
						if mod.Kind == ast.KindPrivateKeyword {
							isPrivate = true
							break
						}
					}
				}
				if !isPrivate {
					checkFunction(node)
				}
				if len(functionStack) > 0 {
					functionStack = functionStack[:len(functionStack)-1]
				}
			},
		}
	},
}
