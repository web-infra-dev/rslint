package class_methods_use_this

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type Options struct {
	EnforceForClassFields                     *bool
	ExceptMethods                            []string
	IgnoreClassesThatImplementAnInterface    interface{} // can be bool or string "public-fields"
	IgnoreOverrideMethods                    *bool
}

type StackInfo struct {
	Class    *ast.Node
	Member   *ast.Node
	Parent   *StackInfo
	UsesThis bool
}

func buildMissingThisMessage(functionName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingThis",
		Description: fmt.Sprintf("Expected 'this' to be used by class %s.", functionName),
	}
}

func getFunctionNameWithKind(ctx rule.RuleContext, node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionExpression:
		if node.AsFunctionExpression().Name() != nil {
			return "function '" + node.AsFunctionExpression().Name().AsIdentifier().Text + "'"
		}
		return "function"
	case ast.KindArrowFunction:
		return "arrow function"
	case ast.KindMethodDeclaration:
		if node.AsMethodDeclaration().Name() != nil {
			return "method '" + getMethodName(node) + "'"
		}
		return "method"
	case ast.KindGetAccessor:
		if node.AsGetAccessorDeclaration().Name() != nil {
			name := extractPropertyName(ctx, node.AsGetAccessorDeclaration().Name())
			if name != "" {
				return "getter '" + name + "'"
			}
		}
		return "getter"
	case ast.KindSetAccessor:
		if node.AsSetAccessorDeclaration().Name() != nil {
			name := extractPropertyName(ctx, node.AsSetAccessorDeclaration().Name())
			if name != "" {
				return "setter '" + name + "'"
			}
		}
		return "setter"
	default:
		return "function"
	}
}

func getMethodName(node *ast.Node) string {
	if !ast.IsMethodDeclaration(node) {
		return ""
	}

	method := node.AsMethodDeclaration()
	nameNode := method.Name()
	if nameNode == nil {
		return ""
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text
	case ast.KindStringLiteral:
		return nameNode.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return nameNode.AsNumericLiteral().Text
	case ast.KindPrivateIdentifier:
		return nameNode.AsPrivateIdentifier().Text
	case ast.KindComputedPropertyName:
		// For computed properties, try to get the inner name
		computed := nameNode.AsComputedPropertyName()
		if computed.Expression != nil && ast.IsIdentifier(computed.Expression) {
			return computed.Expression.AsIdentifier().Text
		}
		return ""
	default:
		return ""
	}
}

func getStaticMemberAccessValue(ctx rule.RuleContext, node *ast.Node) string {
	var nameNode *ast.Node

	if ast.IsMethodDeclaration(node) {
		nameNode = node.AsMethodDeclaration().Name()
	} else if ast.IsPropertyDeclaration(node) {
		nameNode = node.AsPropertyDeclaration().Name()
	} else if ast.IsGetAccessorDeclaration(node) {
		nameNode = node.AsGetAccessorDeclaration().Name()
	} else if ast.IsSetAccessorDeclaration(node) {
		nameNode = node.AsSetAccessorDeclaration().Name()
	} else {
		return ""
	}

	if nameNode == nil {
		return ""
	}

	return extractPropertyName(ctx, nameNode)
}

func extractPropertyName(ctx rule.RuleContext, nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text
	case ast.KindStringLiteral:
		text := nameNode.AsStringLiteral().Text
		// Remove quotes
		if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
			return text[1 : len(text)-1]
		}
		return text
	case ast.KindNumericLiteral:
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		return string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
	case ast.KindComputedPropertyName:
		computed := nameNode.AsComputedPropertyName()
		if computed.Expression != nil {
			return extractPropertyName(ctx, computed.Expression)
		}
		return ""
	case ast.KindPrivateIdentifier:
		return nameNode.AsPrivateIdentifier().Text
	default:
		return ""
	}
}

func isPublicField(node *ast.Node) bool {
	if node == nil {
		return true
	}

	flags := ast.GetCombinedModifierFlags(node)

	// If no explicit modifier or public modifier, it's public
	if flags&(ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected) == 0 {
		return true
	}

	return false
}

func isIncludedInstanceMethod(ctx rule.RuleContext, node *ast.Node, options *Options) bool {
	if node == nil {
		return false
	}

	// Check if static
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
		return false
	}

	// Check if abstract
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract) {
		return false
	}

	// Check if constructor
	if ast.IsConstructorDeclaration(node) {
		return false
	}

	// Skip methods with computed property names
	if hasComputedPropertyName(node) {
		return false
	}

	// Check if method definition with constructor kind (but constructors are handled above)
	// This check is redundant since constructors have their own node type

	// Check enforceForClassFields option for property declarations and accessors
	if ast.IsPropertyDeclaration(node) || ast.IsGetAccessorDeclaration(node) || ast.IsSetAccessorDeclaration(node) {
		if options.EnforceForClassFields == nil || !*options.EnforceForClassFields {
			return false
		}
	}

	// Check if method is in except list
	if len(options.ExceptMethods) > 0 {
		name := getStaticMemberAccessValue(ctx, node)
		if name != "" {
			// For private identifiers, the name already includes "#"
			for _, exceptMethod := range options.ExceptMethods {
				if exceptMethod == name {
					return false
				}
			}
		}
	}

	return true
}

func hasComputedPropertyName(node *ast.Node) bool {
	var nameNode *ast.Node

	if ast.IsMethodDeclaration(node) {
		nameNode = node.AsMethodDeclaration().Name()
	} else if ast.IsPropertyDeclaration(node) {
		nameNode = node.AsPropertyDeclaration().Name()
	} else if ast.IsGetAccessorDeclaration(node) {
		nameNode = node.AsGetAccessorDeclaration().Name()
	} else if ast.IsSetAccessorDeclaration(node) {
		nameNode = node.AsSetAccessorDeclaration().Name()
	}

	return nameNode != nil && nameNode.Kind == ast.KindComputedPropertyName
}

func hasPrivateIdentifier(node *ast.Node) bool {
	var nameNode *ast.Node

	if ast.IsMethodDeclaration(node) {
		nameNode = node.AsMethodDeclaration().Name()
	} else if ast.IsPropertyDeclaration(node) {
		nameNode = node.AsPropertyDeclaration().Name()
	} else if ast.IsGetAccessorDeclaration(node) {
		nameNode = node.AsGetAccessorDeclaration().Name()
	} else if ast.IsSetAccessorDeclaration(node) {
		nameNode = node.AsSetAccessorDeclaration().Name()
	}

	return nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier
}

func isNodeOrDescendant(ancestor *ast.Node, descendant *ast.Node) bool {
	if ancestor == descendant {
		return true
	}

	current := descendant.Parent
	for current != nil {
		if current == ancestor {
			return true
		}
		current = current.Parent
	}
	return false
}

func shouldIgnoreMethod(stackContext *StackInfo, options *Options) bool {
	if stackContext == nil || stackContext.Member == nil || stackContext.Class == nil {
		return true
	}

	// Check if method uses this
	if stackContext.UsesThis {
		return true
	}

	// Check if method has override modifier
	if options.IgnoreOverrideMethods != nil && *options.IgnoreOverrideMethods {
		if ast.HasSyntacticModifier(stackContext.Member, ast.ModifierFlagsOverride) {
			return true
		}
	}

	// Check ignoreClassesThatImplementAnInterface option
	if options.IgnoreClassesThatImplementAnInterface != nil {
		var classDecl *ast.ClassDeclaration
		if ast.IsClassDeclaration(stackContext.Class) {
			classDecl = stackContext.Class.AsClassDeclaration()
		} else if ast.IsClassExpression(stackContext.Class) {
			// Class expressions don't have implements clauses in the same way
			return false
		} else {
			return false
		}

		if classDecl != nil {
			hasImplements := false
			if classDecl.HeritageClauses != nil && len(classDecl.HeritageClauses.Nodes) > 0 {
				for _, clause := range classDecl.HeritageClauses.Nodes {
					if clause.AsHeritageClause().Token == ast.KindImplementsKeyword {
						hasImplements = true
						break
					}
				}
			}
			if hasImplements {
				switch v := options.IgnoreClassesThatImplementAnInterface.(type) {
				case bool:
					if v {
						return true
					}
				case string:
					if v == "public-fields" && isPublicField(stackContext.Member) {
						return true
					}
				}
			}
		}
	}

	return false
}

var ClassMethodsUseThisRule = rule.Rule{
	Name: "class-methods-use-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Parse options
		opts := &Options{
			EnforceForClassFields: func() *bool { b := true; return &b }(),
			ExceptMethods:         []string{},
			IgnoreClassesThatImplementAnInterface: false,
			IgnoreOverrideMethods: func() *bool { b := false; return &b }(),
		}

		if options != nil {
			var optMap map[string]interface{}
			
			// Handle both direct map and array of maps
			if m, ok := options.(map[string]interface{}); ok {
				optMap = m
			} else if arr, ok := options.([]interface{}); ok && len(arr) > 0 {
				if m, ok := arr[0].(map[string]interface{}); ok {
					optMap = m
				}
			}
			
			if optMap != nil {
				if val, exists := optMap["enforceForClassFields"]; exists {
					if b, ok := val.(bool); ok {
						opts.EnforceForClassFields = &b
					}
				}
				if val, exists := optMap["exceptMethods"]; exists {
					if methods, ok := val.([]interface{}); ok {
						opts.ExceptMethods = make([]string, len(methods))
						for i, method := range methods {
							if s, ok := method.(string); ok {
								opts.ExceptMethods[i] = s
							}
						}
					}
				}
				if val, exists := optMap["ignoreClassesThatImplementAnInterface"]; exists {
					opts.IgnoreClassesThatImplementAnInterface = val
				}
				if val, exists := optMap["ignoreOverrideMethods"]; exists {
					if b, ok := val.(bool); ok {
						opts.IgnoreOverrideMethods = &b
					}
				}
			}
		}

		var stack *StackInfo

		pushContext := func(member *ast.Node) {
			if member != nil && member.Parent != nil {
				// Check if the parent is a class declaration or expression
				classNode := member.Parent
				if classNode != nil && (classNode.Kind == ast.KindClassDeclaration || classNode.Kind == ast.KindClassExpression) {
					stack = &StackInfo{
						Class:    classNode,
						Member:   member,
						Parent:   stack,
						UsesThis: false,
					}
				} else {
					stack = &StackInfo{
						Class:    nil,
						Member:   nil,
						Parent:   stack,
						UsesThis: false,
					}
				}
			} else {
				stack = &StackInfo{
					Class:    nil,
					Member:   nil,
					Parent:   stack,
					UsesThis: false,
				}
			}
		}

		popContext := func() *StackInfo {
			oldStack := stack
			if stack != nil {
				stack = stack.Parent
			}
			return oldStack
		}

		enterFunction := func(node *ast.Node) {
			// Simplified context detection to match TypeScript behavior
			// Check if the immediate parent is a class member (direct relationship)
			if node.Parent != nil {
				parent := node.Parent
				switch parent.Kind {
				case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
					pushContext(parent)
					return
				case ast.KindPropertyDeclaration:
					// Check if this function is the direct initializer of the property
					propDecl := parent.AsPropertyDeclaration()
					if propDecl.Initializer == node {
						pushContext(parent)
						return
					}
				// Note: AccessorProperty doesn't exist in current AST, handled via PropertyDeclaration with accessor modifier
				}
			}
			// If not a direct child of a class member, push nil context
			pushContext(nil)
		}

		exitFunction := func(node *ast.Node) {
			stackContext := popContext()

			if shouldIgnoreMethod(stackContext, opts) {
				return
			}

			if stackContext.Member != nil && isIncludedInstanceMethod(ctx, stackContext.Member, opts) {
				functionName := getFunctionNameWithKind(ctx, node)
				ctx.ReportNode(node, buildMissingThisMessage(functionName))
			}
		}

		markAsUsesThis := func() {
			if stack != nil {
				stack.UsesThis = true
			}
		}

		listeners := rule.RuleListeners{
			// Function declarations have their own `this` context
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				pushContext(nil)
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				popContext()
			},

			// Function expressions
			ast.KindFunctionExpression:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression): exitFunction,

			// Method declarations
			ast.KindMethodDeclaration: func(node *ast.Node) {
				pushContext(node)
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(node *ast.Node) {
				stackContext := popContext()

				if shouldIgnoreMethod(stackContext, opts) {
					return
				}

				if stackContext.Member != nil && isIncludedInstanceMethod(ctx, stackContext.Member, opts) {
					functionName := getFunctionNameWithKind(ctx, node)
					// For methods, getters, and setters, report on the name node for better error positioning
					var reportNode *ast.Node
					switch node.Kind {
					case ast.KindMethodDeclaration:
						if node.AsMethodDeclaration().Name() != nil {
							reportNode = node.AsMethodDeclaration().Name()
						}
					case ast.KindGetAccessor:
						if node.AsGetAccessorDeclaration().Name() != nil {
							reportNode = node.AsGetAccessorDeclaration().Name()
						}
					case ast.KindSetAccessor:
						if node.AsSetAccessorDeclaration().Name() != nil {
							reportNode = node.AsSetAccessorDeclaration().Name()
						}
					}
					if reportNode == nil {
						reportNode = node
					}
					ctx.ReportNode(reportNode, buildMissingThisMessage(functionName))
				}
			},

			// Get accessors
			ast.KindGetAccessor: func(node *ast.Node) {
				pushContext(node)
			},
			rule.ListenerOnExit(ast.KindGetAccessor): func(node *ast.Node) {
				stackContext := popContext()

				if shouldIgnoreMethod(stackContext, opts) {
					return
				}

				if stackContext.Member != nil && isIncludedInstanceMethod(ctx, stackContext.Member, opts) {
					functionName := getFunctionNameWithKind(ctx, node)
					// For methods, getters, and setters, report on the name node for better error positioning
					var reportNode *ast.Node
					switch node.Kind {
					case ast.KindMethodDeclaration:
						if node.AsMethodDeclaration().Name() != nil {
							reportNode = node.AsMethodDeclaration().Name()
						}
					case ast.KindGetAccessor:
						if node.AsGetAccessorDeclaration().Name() != nil {
							reportNode = node.AsGetAccessorDeclaration().Name()
						}
					case ast.KindSetAccessor:
						if node.AsSetAccessorDeclaration().Name() != nil {
							reportNode = node.AsSetAccessorDeclaration().Name()
						}
					}
					if reportNode == nil {
						reportNode = node
					}
					ctx.ReportNode(reportNode, buildMissingThisMessage(functionName))
				}
			},

			// Set accessors
			ast.KindSetAccessor: func(node *ast.Node) {
				pushContext(node)
			},
			rule.ListenerOnExit(ast.KindSetAccessor): func(node *ast.Node) {
				stackContext := popContext()

				if shouldIgnoreMethod(stackContext, opts) {
					return
				}

				if stackContext.Member != nil && isIncludedInstanceMethod(ctx, stackContext.Member, opts) {
					functionName := getFunctionNameWithKind(ctx, node)
					// For methods, getters, and setters, report on the name node for better error positioning
					var reportNode *ast.Node
					switch node.Kind {
					case ast.KindMethodDeclaration:
						if node.AsMethodDeclaration().Name() != nil {
							reportNode = node.AsMethodDeclaration().Name()
						}
					case ast.KindGetAccessor:
						if node.AsGetAccessorDeclaration().Name() != nil {
							reportNode = node.AsGetAccessorDeclaration().Name()
						}
					case ast.KindSetAccessor:
						if node.AsSetAccessorDeclaration().Name() != nil {
							reportNode = node.AsSetAccessorDeclaration().Name()
						}
					}
					if reportNode == nil {
						reportNode = node
					}
					ctx.ReportNode(reportNode, buildMissingThisMessage(functionName))
				}
			},


			// Static blocks have their own `this` context
			ast.KindClassStaticBlockDeclaration: func(node *ast.Node) {
				pushContext(nil)
			},
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): func(node *ast.Node) {
				popContext()
			},

			// Mark `this` usage
			ast.KindThisKeyword: func(node *ast.Node) {
				markAsUsesThis()
			},
			ast.KindSuperKeyword: func(node *ast.Node) {
				markAsUsesThis()
			},
		}

		// Arrow functions - but only handle them if they're not already handled as property initializers
		listeners[ast.KindArrowFunction] = func(node *ast.Node) {
			// Check if this arrow function is a property initializer
			if node.Parent != nil && node.Parent.Kind == ast.KindPropertyDeclaration {
				propDecl := node.Parent.AsPropertyDeclaration()
				if propDecl.Initializer == node {
					// This is handled by PropertyDeclaration logic, skip here
					return
				}
			}
			enterFunction(node)
		}
		listeners[rule.ListenerOnExit(ast.KindArrowFunction)] = func(node *ast.Node) {
			// Check if this arrow function is a property initializer
			if node.Parent != nil && node.Parent.Kind == ast.KindPropertyDeclaration {
				propDecl := node.Parent.AsPropertyDeclaration()
				if propDecl.Initializer == node {
					// This is handled by PropertyDeclaration logic, skip here
					return
				}
			}
			exitFunction(node)
		}

		// Add specific handling for PropertyDefinition > ArrowFunctionExpression when enforceForClassFields is enabled
		if opts.EnforceForClassFields != nil && *opts.EnforceForClassFields {
			listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				if property.Initializer != nil && property.Initializer.Kind == ast.KindArrowFunction {
					// This is a property with arrow function initializer - treat it like a method
					pushContext(node)
				}
			}
			listeners[rule.ListenerOnExit(ast.KindPropertyDeclaration)] = func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				if property.Initializer != nil && property.Initializer.Kind == ast.KindArrowFunction {
					// Exit the context for property with arrow function
					stackContext := popContext()
					
					if shouldIgnoreMethod(stackContext, opts) {
						return
					}
					
					if stackContext.Member != nil && isIncludedInstanceMethod(ctx, stackContext.Member, opts) {
						// Report the error on the arrow function for better positioning
						var reportNode *ast.Node
						if property.Initializer != nil && property.Initializer.Kind == ast.KindArrowFunction {
							reportNode = property.Initializer
						} else if property.Name() != nil {
							reportNode = property.Name()
						} else {
							reportNode = node
						}
						functionName := "property '" + getStaticMemberAccessValue(ctx, node) + "'"
						ctx.ReportNode(reportNode, buildMissingThisMessage(functionName))
					}
				}
			}
		}

		// Handle accessor properties (PropertyDeclaration with accessor modifier)
		// Since enforceForClassFields affects accessor properties when enabled
		if opts.EnforceForClassFields != nil && *opts.EnforceForClassFields {
			// Enhance existing PropertyDeclaration listeners to also handle accessor properties
			existingEnterListener := listeners[ast.KindPropertyDeclaration]
			listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				
				// Handle arrow function initializers first
				if existingEnterListener != nil {
					existingEnterListener(node)
				}
				
				// Also handle accessor properties
				if ast.HasAccessorModifier(node) && (property.Initializer == nil || property.Initializer.Kind != ast.KindArrowFunction) {
					// This is an accessor property without arrow function - treat it like a method
					pushContext(node)
				}
			}
			
			existingExitListener := listeners[rule.ListenerOnExit(ast.KindPropertyDeclaration)]
			listeners[rule.ListenerOnExit(ast.KindPropertyDeclaration)] = func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				
				// Handle arrow function initializers first
				if existingExitListener != nil {
					existingExitListener(node)
				}
				
				// Also handle accessor properties
				if ast.HasAccessorModifier(node) && (property.Initializer == nil || property.Initializer.Kind != ast.KindArrowFunction) {
					// Exit the context for accessor property
					stackContext := popContext()
					
					if shouldIgnoreMethod(stackContext, opts) {
						return
					}
					
					if stackContext.Member != nil && isIncludedInstanceMethod(ctx, stackContext.Member, opts) {
						// Report the error on the property name
						var reportNode *ast.Node
						if property.Name() != nil {
							reportNode = property.Name()
						} else {
							reportNode = node
						}
						functionName := "accessor '" + getStaticMemberAccessValue(ctx, node) + "'"
						ctx.ReportNode(reportNode, buildMissingThisMessage(functionName))
					}
				}
			}
		}

		return listeners
	},
}
