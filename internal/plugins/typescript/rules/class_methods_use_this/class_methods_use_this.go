package class_methods_use_this

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ClassMethodsUseThisOptions struct {
	ExceptMethods                         []string `json:"exceptMethods"`
	EnforceForClassFields                 bool     `json:"enforceForClassFields"`
	IgnoreClassesThatImplementAnInterface interface{} `json:"ignoreClassesThatImplementAnInterface"`
	IgnoreOverrideMethods                 bool     `json:"ignoreOverrideMethods"`
}

type scopeInfo struct {
	hasThis bool
	node    *ast.Node
	upper   *scopeInfo
}

var ClassMethodsUseThisRule = rule.CreateRule(rule.Rule{
	Name: "class-methods-use-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ClassMethodsUseThisOptions{
			ExceptMethods:         []string{},
			EnforceForClassFields: true,
			IgnoreClassesThatImplementAnInterface: false,
			IgnoreOverrideMethods: false,
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if opts, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = opts
				}
			} else if opts, ok := options.(map[string]interface{}); ok {
				optsMap = opts
			}

			if optsMap != nil {
				if exceptMethods, ok := optsMap["exceptMethods"].([]interface{}); ok {
					for _, method := range exceptMethods {
						if str, ok := method.(string); ok {
							opts.ExceptMethods = append(opts.ExceptMethods, str)
						}
					}
				}
				if enforceForClassFields, ok := optsMap["enforceForClassFields"].(bool); ok {
					opts.EnforceForClassFields = enforceForClassFields
				}
				if ignoreClasses, ok := optsMap["ignoreClassesThatImplementAnInterface"]; ok {
					opts.IgnoreClassesThatImplementAnInterface = ignoreClasses
				}
				if ignoreOverride, ok := optsMap["ignoreOverrideMethods"].(bool); ok {
					opts.IgnoreOverrideMethods = ignoreOverride
				}
			}
		}

		// Helper to check if a method name is excepted
		isExceptedMethod := func(methodName string) bool {
			for _, name := range opts.ExceptMethods {
				if name == methodName {
					return true
				}
			}
			return false
		}

		// Helper to check if node is inside a class
		isInClass := func(node *ast.Node) bool {
			current := node.Parent
			for current != nil {
				if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
					return true
				}
				current = current.Parent
			}
			return false
		}

		// Helper to get the parent class node
		getParentClass := func(node *ast.Node) *ast.Node {
			current := node.Parent
			for current != nil {
				if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
					return current
				}
				current = current.Parent
			}
			return nil
		}

		// Helper to check if a class implements an interface
		classImplementsInterface := func(classNode *ast.Node) bool {
			if classNode == nil {
				return false
			}

			heritageClauses := utils.GetHeritageClauses(classNode)
			if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
				return false
			}

			for _, clauseNode := range heritageClauses.Nodes {
				clause := clauseNode.AsHeritageClause()
				if clause != nil && clause.Token == ast.KindImplementsKeyword {
					return true
				}
			}
			return false
		}

		// Helper to check if member should be ignored based on ignoreClassesThatImplementAnInterface option
		shouldIgnoreInterfaceImpl := func(node *ast.Node) bool {
			if opts.IgnoreClassesThatImplementAnInterface == nil || opts.IgnoreClassesThatImplementAnInterface == false {
				return false
			}

			classNode := getParentClass(node)
			if !classImplementsInterface(classNode) {
				return false
			}

			// If option is true, ignore all members of classes that implement interfaces
			if boolVal, ok := opts.IgnoreClassesThatImplementAnInterface.(bool); ok && boolVal {
				return true
			}

			// If option is "public-fields", only ignore public members
			if strVal, ok := opts.IgnoreClassesThatImplementAnInterface.(string); ok && strVal == "public-fields" {
				// Check if the member is private or protected
				hasPrivateModifier := ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate)
				hasProtectedModifier := ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected)
				isPrivateName := false

				// Check if it's a private name (starts with #)
				if node.Kind == ast.KindMethodDeclaration {
					if method := node.AsMethodDeclaration(); method != nil && method.Name() != nil {
						_, nameType := utils.GetNameFromMember(ctx.SourceFile, method.Name())
						isPrivateName = nameType == utils.MemberNameTypePrivate
					}
				} else if node.Kind == ast.KindPropertyDeclaration {
					if prop := node.AsPropertyDeclaration(); prop != nil && prop.Name() != nil {
						_, nameType := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
						isPrivateName = nameType == utils.MemberNameTypePrivate
					}
				} else if node.Kind == ast.KindGetAccessor {
					if accessor := node.AsGetAccessorDeclaration(); accessor != nil && accessor.Name() != nil {
						_, nameType := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
						isPrivateName = nameType == utils.MemberNameTypePrivate
					}
				} else if node.Kind == ast.KindSetAccessor {
					if accessor := node.AsSetAccessorDeclaration(); accessor != nil && accessor.Name() != nil {
						_, nameType := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
						isPrivateName = nameType == utils.MemberNameTypePrivate
					}
				}

				// If it's private or protected, don't ignore it (check it)
				if hasPrivateModifier || hasProtectedModifier || isPrivateName {
					return false
				}

				// It's a public member, so ignore it
				return true
			}

			return false
		}

		// Helper to check if member has override modifier
		hasOverrideModifier := func(node *ast.Node) bool {
			return ast.HasSyntacticModifier(node, ast.ModifierFlagsOverride)
		}

		// Get method name for display
		getMethodName := func(node *ast.Node) string {
			if node.Kind == ast.KindMethodDeclaration {
				method := node.AsMethodDeclaration()
				if method != nil && method.Name() != nil {
					name, nameType := utils.GetNameFromMember(ctx.SourceFile, method.Name())
					if nameType == utils.MemberNameTypePrivate {
						if method.Kind == ast.KindGetAccessor {
							return "private getter " + name
						} else if method.Kind == ast.KindSetAccessor {
							return "private setter " + name
						} else if method.AsteriskToken != nil {
							return "private generator method " + name
						}
						return "private method " + name
					}

					if method.Kind == ast.KindGetAccessor {
						if name == "" {
							return "getter"
						}
						return "getter '" + name + "'"
					} else if method.Kind == ast.KindSetAccessor {
						if name == "" {
							return "setter"
						}
						return "setter '" + name + "'"
					} else if method.AsteriskToken != nil {
						if name == "" {
							return "generator method"
						}
						return "generator method '" + name + "'"
					}
					if name == "" {
						return "method"
					}
					return "method '" + name + "'"
				}
				return "method"
			} else if node.Kind == ast.KindGetAccessor {
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, nameType := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					if nameType == utils.MemberNameTypePrivate {
						return "private getter " + name
					}
					if name == "" {
						return "getter"
					}
					return "getter '" + name + "'"
				}
				return "getter"
			} else if node.Kind == ast.KindSetAccessor {
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, nameType := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					if nameType == utils.MemberNameTypePrivate {
						return "private setter " + name
					}
					if name == "" {
						return "setter"
					}
					return "setter '" + name + "'"
				}
				return "setter"
			} else if node.Kind == ast.KindPropertyDeclaration {
				prop := node.AsPropertyDeclaration()
				if prop != nil && prop.Name() != nil {
					name, nameType := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
					if nameType == utils.MemberNameTypePrivate {
						return "private method " + name
					}
					if name == "" {
						return "method"
					}
					return "method '" + name + "'"
				}
				return "method"
			}
			return "method"
		}

		var currentScope *scopeInfo

		// Enter a method or property
		enterMethod := func(node *ast.Node) {
			// Skip constructors
			if node.Kind == ast.KindConstructor {
				return
			}

			// Skip static methods
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
				return
			}

			// Skip abstract methods
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract) {
				return
			}

			// Skip if not in a class
			if !isInClass(node) {
				return
			}

			// Skip if has override modifier and ignoreOverrideMethods is true
			if opts.IgnoreOverrideMethods && hasOverrideModifier(node) {
				return
			}

			// Skip if in a class that implements an interface and should be ignored
			if shouldIgnoreInterfaceImpl(node) {
				return
			}

			// Check if method is in except list
			var methodName string
			if node.Kind == ast.KindMethodDeclaration {
				method := node.AsMethodDeclaration()
				if method != nil && method.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
					methodName = name
				}
			} else if node.Kind == ast.KindGetAccessor {
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					methodName = name
				}
			} else if node.Kind == ast.KindSetAccessor {
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					methodName = name
				}
			}

			if methodName != "" && isExceptedMethod(methodName) {
				return
			}

			// Create a new scope
			currentScope = &scopeInfo{
				hasThis: false,
				node:    node,
				upper:   currentScope,
			}
		}

		// Exit a method
		exitMethod := func(node *ast.Node) {
			if currentScope != nil && currentScope.node == node {
				// Check if we used 'this' or 'super'
				if !currentScope.hasThis {
					displayName := getMethodName(node)
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "missingThis",
						Description: "Expected 'this' to be used by class " + displayName + ".",
					})
				}

				// Pop the scope
				currentScope = currentScope.upper
			}
		}

		// Enter a property initializer (function or arrow function)
		enterPropertyInit := func(node *ast.Node) {
			// Only process if this is a function/arrow that's a direct child of a property declaration
			parent := node.Parent
			if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
				return
			}

			prop := parent.AsPropertyDeclaration()
			if prop == nil || prop.Initializer != node {
				return
			}

			// Skip if enforceForClassFields is false
			if !opts.EnforceForClassFields {
				return
			}

			// Skip static properties
			if ast.HasSyntacticModifier(parent, ast.ModifierFlagsStatic) {
				return
			}

			// Skip if not in a class
			if !isInClass(parent) {
				return
			}

			// Skip if has override modifier and ignoreOverrideMethods is true
			if opts.IgnoreOverrideMethods && hasOverrideModifier(parent) {
				return
			}

			// Skip if in a class that implements an interface and should be ignored
			if shouldIgnoreInterfaceImpl(parent) {
				return
			}

			// Check if property name is in except list
			if prop.Name() != nil {
				name, _ := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
				if name != "" && isExceptedMethod(name) {
					return
				}
			}

			// Create a new scope for the initializer
			currentScope = &scopeInfo{
				hasThis: false,
				node:    node,
				upper:   currentScope,
			}
		}

		// Exit a property initializer
		exitPropertyInit := func(node *ast.Node) {
			if currentScope != nil && currentScope.node == node {
				// Check if we used 'this' or 'super'
				if !currentScope.hasThis {
					// Find the parent property declaration
					parent := node.Parent
					if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
						displayName := getMethodName(parent)
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "missingThis",
							Description: "Expected 'this' to be used by class " + displayName + ".",
						})
					}
				}

				// Pop the scope
				currentScope = currentScope.upper
			}
		}

		// Mark that we found 'this' or 'super'
		markAsHasThis := func() {
			if currentScope != nil {
				currentScope.hasThis = true
			}
		}

		// Enter a function expression or arrow function (to create a boundary for nested functions)
		enterNestedFunction := func(node *ast.Node) {
			// Check if this is a property initializer - if so, skip (handled separately)
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
				prop := parent.AsPropertyDeclaration()
				if prop != nil && prop.Initializer == node {
					return
				}
			}

			// Don't check nested regular functions for 'this'
			// Create a boundary scope for nested functions
			if node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindFunctionDeclaration {
				currentScope = &scopeInfo{
					hasThis: true, // Mark as having 'this' so we don't report
					node:    node,
					upper:   currentScope,
				}
			}
		}

		// Exit a nested function
		exitNestedFunction := func(node *ast.Node) {
			// Skip if this is a property initializer (handled by exitPropertyInit)
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
				prop := parent.AsPropertyDeclaration()
				if prop != nil && prop.Initializer == node {
					return
				}
			}

			if currentScope != nil && currentScope.node == node {
				currentScope = currentScope.upper
			}
		}

		return rule.RuleListeners{
			// Method listeners
			ast.KindMethodDeclaration:                      enterMethod,
			rule.ListenerOnExit(ast.KindMethodDeclaration): exitMethod,
			ast.KindGetAccessor:                            enterMethod,
			rule.ListenerOnExit(ast.KindGetAccessor):       exitMethod,
			ast.KindSetAccessor:                            enterMethod,
			rule.ListenerOnExit(ast.KindSetAccessor):       exitMethod,

			// Function expression/arrow function listeners
			// These handle both property initializers and nested functions
			ast.KindFunctionExpression: func(node *ast.Node) {
				enterPropertyInit(node)
				enterNestedFunction(node)
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				exitPropertyInit(node)
				exitNestedFunction(node)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				enterNestedFunction(node)
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				exitNestedFunction(node)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				enterPropertyInit(node)
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				exitPropertyInit(node)
			},

			// This/super keyword listeners
			ast.KindThisKeyword:  func(node *ast.Node) { markAsHasThis() },
			ast.KindSuperKeyword: func(node *ast.Node) { markAsHasThis() },
		}
	},
})
