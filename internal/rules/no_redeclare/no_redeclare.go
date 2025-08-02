package no_redeclare

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type NoRedeclareOptions struct {
	BuiltinGlobals         bool `json:"builtinGlobals"`
	IgnoreDeclarationMerge bool `json:"ignoreDeclarationMerge"`
}

type DeclarationInfo struct {
	Node     *ast.Node
	DeclType string // "syntax", "builtin", "comment"
	Kind     ast.Kind
}

type VariableInfo struct {
	Name         string
	Declarations []DeclarationInfo
}

type ScopeInfo struct {
	Node      *ast.Node
	Variables map[string]*VariableInfo
	Parent    *ScopeInfo
	Children  []*ScopeInfo
}

var NoRedeclareRule = rule.Rule{
	Name: "no-redeclare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoRedeclareOptions{
			BuiltinGlobals:         false,
			IgnoreDeclarationMerge: true,
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
				if val, ok := optsMap["builtinGlobals"].(bool); ok {
					opts.BuiltinGlobals = val
				}
				if val, ok := optsMap["ignoreDeclarationMerge"].(bool); ok {
					opts.IgnoreDeclarationMerge = val
				}
			}
		}

		// Built-in globals to check
		builtinGlobals := map[string]bool{
			"Object":     true,
			"Array":      true,
			"Function":   true,
			"String":     true,
			"Number":     true,
			"Boolean":    true,
			"Symbol":     true,
			"BigInt":     true,
			"Promise":    true,
			"Error":      true,
			"Map":        true,
			"Set":        true,
			"WeakMap":    true,
			"WeakSet":    true,
			"Date":       true,
			"RegExp":     true,
			"JSON":       true,
			"Math":       true,
			"console":    true,
			"window":     true,
			"document":   true,
			"global":     true,
			"globalThis": true,
			"undefined":  true,
			"Infinity":   true,
			"NaN":        true,
			"eval":       true,
			"parseInt":   true,
			"parseFloat": true,
			"isNaN":      true,
			"isFinite":   true,
			"top":        true,
			"self":       true,
			// TypeScript lib types
			"NodeListOf": true,
		}

		// Track scopes
		rootScope := &ScopeInfo{
			Node:      ctx.SourceFile.AsNode(),
			Variables: make(map[string]*VariableInfo),
		}
		currentScope := rootScope
		scopeStack := []*ScopeInfo{rootScope}

		// Track which scopes we've processed
		processedScopes := make(map[*ast.Node]bool)


		// Helper to enter a new scope
		enterScope := func(node *ast.Node) {
			if _, processed := processedScopes[node]; processed {
				return
			}
			processedScopes[node] = true

			newScope := &ScopeInfo{
				Node:      node,
				Variables: make(map[string]*VariableInfo),
				Parent:    currentScope,
			}
			currentScope.Children = append(currentScope.Children, newScope)
			currentScope = newScope
			scopeStack = append(scopeStack, newScope)
		}

		// Helper to exit scope
		exitScope := func() {
			if len(scopeStack) > 1 {
				scopeStack = scopeStack[:len(scopeStack)-1]
				currentScope = scopeStack[len(scopeStack)-1]
			}
		}

		// Helper to get identifier name
		getIdentifierName := func(node *ast.Node) string {
			if ast.IsIdentifier(node) {
				return node.AsIdentifier().Text
			}
			return ""
		}


		// Helper to add declaration
		addDeclaration := func(name string, nameNode *ast.Node, declType string, declNode *ast.Node) {
			if name == "" {
				return
			}
			

			varInfo, exists := currentScope.Variables[name]
			if !exists {
				varInfo = &VariableInfo{
					Name:         name,
					Declarations: []DeclarationInfo{},
				}
				currentScope.Variables[name] = varInfo
			}

			varInfo.Declarations = append(varInfo.Declarations, DeclarationInfo{
				Node:     nameNode,  // Use nameNode for reporting location
				DeclType: declType,
				Kind:     declNode.Kind,  // Use declNode.Kind for merging logic
			})
			
			// Check for redeclaration immediately
			if len(varInfo.Declarations) > 1 {
				if opts.IgnoreDeclarationMerge {
					// When declaration merging is enabled, we need more nuanced handling
					classCounts := 0
					functionCounts := 0
					enumCounts := 0
					namespaceCounts := 0
					interfaceCounts := 0
					otherCounts := 0
					
					for _, decl := range varInfo.Declarations {
						switch decl.Kind {
						case ast.KindClassDeclaration:
							classCounts++
						case ast.KindFunctionDeclaration:
							functionCounts++
						case ast.KindEnumDeclaration:
							enumCounts++
						case ast.KindModuleDeclaration:
							namespaceCounts++
						case ast.KindInterfaceDeclaration:
							interfaceCounts++
						default:
							otherCounts++
						}
					}
					
					currentKind := declNode.Kind
					shouldReport := false
					
					// Check if this combination is allowed for declaration merging
					// If there are any non-mergeable types (like variables), report conflict
					if otherCounts > 0 {
						// Variables and other non-mergeable types conflict with everything
						shouldReport = true
					} else {
						// Handle mergeable declaration types
						switch currentKind {
						case ast.KindClassDeclaration:
							// Report only when we encounter exactly the second class
							shouldReport = classCounts == 2
						case ast.KindFunctionDeclaration:
							// Report only when we encounter exactly the second function
							shouldReport = functionCounts == 2
						case ast.KindEnumDeclaration:
							// Report only when we encounter exactly the second enum
							shouldReport = enumCounts == 2
						case ast.KindModuleDeclaration:
							// Namespaces can merge with classes/functions/enums
							if classCounts > 1 || functionCounts > 1 || enumCounts > 1 {
								// If there are already duplicate classes/functions/enums, 
								// don't report on the namespace
								shouldReport = false
							} else if namespaceCounts > 1 && classCounts == 0 && functionCounts == 0 && enumCounts == 0 && interfaceCounts == 0 {
								// Multiple standalone namespaces are allowed to merge
								shouldReport = false
							} else {
								// Single namespace merging with single class/function/enum is allowed
								shouldReport = false
							}
						case ast.KindInterfaceDeclaration:
							// Interfaces can always merge with each other
							// Interfaces can merge with classes and namespaces
							if classCounts <= 1 && functionCounts == 0 && enumCounts == 0 {
								shouldReport = false
							} else {
								shouldReport = true
							}
						default:
							// For other mergeable types we haven't explicitly handled
							shouldReport = true
						}
					}
					
					if shouldReport {
						// Check if this is a builtin global conflict
						firstDecl := varInfo.Declarations[0]
						var messageId string
						if firstDecl.DeclType == "builtin" {
							messageId = "redeclaredAsBuiltin"
						} else if firstDecl.DeclType == "comment" && declType == "syntax" {
							messageId = "redeclaredBySyntax"
						} else {
							messageId = "redeclared"
						}

						var description string
						switch messageId {
						case "redeclaredAsBuiltin":
							description = fmt.Sprintf("'%s' is already defined as a built-in global variable.", name)
						case "redeclaredBySyntax":
							description = fmt.Sprintf("'%s' is already defined by a variable declaration.", name)
						default:
							description = fmt.Sprintf("'%s' is already defined.", name)
						}

						ctx.ReportNode(nameNode, rule.RuleMessage{
							Id:          messageId,
							Description: description,
						})
					}
				} else {
					// When declaration merging is disabled, report all redeclarations
					firstDecl := varInfo.Declarations[0]
					
					var messageId string
					if firstDecl.DeclType == "builtin" {
						messageId = "redeclaredAsBuiltin"
					} else if firstDecl.DeclType == "comment" && declType == "syntax" {
						messageId = "redeclaredBySyntax"
					} else {
						messageId = "redeclared"
					}

					var description string
					switch messageId {
					case "redeclaredAsBuiltin":
						description = fmt.Sprintf("'%s' is already defined as a built-in global variable.", name)
					case "redeclaredBySyntax":
						description = fmt.Sprintf("'%s' is already defined by a variable declaration.", name)
					default:
						description = fmt.Sprintf("'%s' is already defined.", name)
					}

					ctx.ReportNode(nameNode, rule.RuleMessage{
						Id:          messageId,
						Description: description,
					})
				}
			}
		}

		// Add built-in globals to root scope if enabled
		if opts.BuiltinGlobals {
			// Only add globals to the actual global scope (not module scope)
			// Check if this is a module by looking for exports or imports
			isModuleScope := false
			ctx.SourceFile.ForEachChild(func(node *ast.Node) bool {
				if ast.IsImportDeclaration(node) || ast.IsExportDeclaration(node) || ast.IsExportAssignment(node) {
					isModuleScope = true
					return true
				}
				return false
			})
			if !isModuleScope {
				for global := range builtinGlobals {
					addDeclaration(global, ctx.SourceFile.AsNode(), "builtin", ctx.SourceFile.AsNode())
				}
			}
		}
		

		return rule.RuleListeners{
			// Variable statements (containing variable declarations)
			ast.KindVariableStatement: func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt.DeclarationList == nil {
					return
				}
				declList := varStmt.DeclarationList.AsVariableDeclarationList()
				for _, decl := range declList.Declarations.Nodes {
					varDecl := decl.AsVariableDeclaration()
					if ast.IsIdentifier(varDecl.Name()) {
						name := getIdentifierName(varDecl.Name())
						addDeclaration(name, varDecl.Name(), "syntax", node)
					} else if ast.IsObjectBindingPattern(varDecl.Name()) || ast.IsArrayBindingPattern(varDecl.Name()) {
						// Handle destructuring patterns
						var processBindingPattern func(*ast.Node)
						processBindingPattern = func(pattern *ast.Node) {
							if ast.IsObjectBindingPattern(pattern) || ast.IsArrayBindingPattern(pattern) {
								bindingPattern := pattern.AsBindingPattern()
								for _, element := range bindingPattern.Elements.Nodes {
									if element != nil && ast.IsBindingElement(element) {
										bindingElement := element.AsBindingElement()
										if ast.IsIdentifier(bindingElement.Name()) {
											name := getIdentifierName(bindingElement.Name())
											addDeclaration(name, bindingElement.Name(), "syntax", node)
										} else {
											// Nested patterns
											processBindingPattern(bindingElement.Name())
										}
									}
								}
							}
						}
						processBindingPattern(varDecl.Name())
					}
				}
			},

			// Function declarations
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl.Name() != nil {
					name := getIdentifierName(funcDecl.Name())
					addDeclaration(name, funcDecl.Name(), "syntax", node)
				}
				enterScope(node)
			},

			// Function expressions  
			ast.KindFunctionExpression: func(node *ast.Node) {
				enterScope(node)
			},

			// Arrow functions
			ast.KindArrowFunction: func(node *ast.Node) {
				enterScope(node)
			},

			// Class declarations
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl.Name() != nil {
					name := getIdentifierName(classDecl.Name())
					addDeclaration(name, classDecl.Name(), "syntax", node)
				}
			},

			// TypeScript declarations
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				name := getIdentifierName(interfaceDecl.Name())
				addDeclaration(name, interfaceDecl.Name(), "syntax", node)
			},

			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				name := getIdentifierName(typeAlias.Name())
				addDeclaration(name, typeAlias.Name(), "syntax", node)
			},

			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				name := getIdentifierName(enumDecl.Name())
				addDeclaration(name, enumDecl.Name(), "syntax", node)
			},

			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if ast.IsIdentifier(moduleDecl.Name()) {
					name := getIdentifierName(moduleDecl.Name())
					addDeclaration(name, moduleDecl.Name(), "syntax", node)
				}
			},

			// Block statements create new scopes
			ast.KindBlock: func(node *ast.Node) {
				// Don't create scope for function bodies (already created)
				parent := node.Parent
				if parent != nil {
					switch parent.Kind {
					case ast.KindFunctionDeclaration,
						ast.KindFunctionExpression,
						ast.KindArrowFunction:
						return
					}
				}
				enterScope(node)
			},

			// For statements
			ast.KindForStatement: func(node *ast.Node) {
				enterScope(node)
			},

			ast.KindForInStatement: func(node *ast.Node) {
				enterScope(node)
			},

			ast.KindForOfStatement: func(node *ast.Node) {
				enterScope(node)
			},

			// Switch statements
			ast.KindSwitchStatement: func(node *ast.Node) {
				enterScope(node)
			},

			// Exit listeners
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				exitScope()
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				exitScope()
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				exitScope()
			},
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				parent := node.Parent
				if parent != nil {
					switch parent.Kind {
					case ast.KindFunctionDeclaration,
						ast.KindFunctionExpression,
						ast.KindArrowFunction:
						return
					}
				}
				exitScope()
			},
			rule.ListenerOnExit(ast.KindForStatement): func(node *ast.Node) {
				exitScope()
			},
			rule.ListenerOnExit(ast.KindForInStatement): func(node *ast.Node) {
				exitScope()
			},
			rule.ListenerOnExit(ast.KindForOfStatement): func(node *ast.Node) {
				exitScope()
			},
			rule.ListenerOnExit(ast.KindSwitchStatement): func(node *ast.Node) {
				exitScope()
			},

		}
	},
}