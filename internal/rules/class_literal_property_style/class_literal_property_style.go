package class_literal_property_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type propertiesInfo struct {
	excludeSet map[string]bool
	properties []*ast.Node
}

func buildPreferFieldStyleMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFieldStyle",
		Description: "Literals should be exposed using readonly fields.",
	}
}

func buildPreferFieldStyleSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFieldStyleSuggestion",
		Description: "Replace the literals with readonly fields.",
	}
}

func buildPreferGetterStyleMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferGetterStyle",
		Description: "Literals should be exposed using getters.",
	}
}

func buildPreferGetterStyleSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferGetterStyleSuggestion",
		Description: "Replace the literals with getters.",
	}
}

func printNodeModifiers(node *ast.Node, final string) string {
	var modifiers []string

	flags := ast.GetCombinedModifierFlags(node)

	if flags&ast.ModifierFlagsPublic != 0 {
		modifiers = append(modifiers, "public")
	} else if flags&ast.ModifierFlagsPrivate != 0 {
		modifiers = append(modifiers, "private")
	} else if flags&ast.ModifierFlagsProtected != 0 {
		modifiers = append(modifiers, "protected")
	}

	if flags&ast.ModifierFlagsStatic != 0 {
		modifiers = append(modifiers, "static")
	}

	modifiers = append(modifiers, final)

	result := strings.Join(modifiers, " ")
	if result != "" {
		result += " "
	}
	return result
}

func isSupportedLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindTemplateExpression:
		// Only support template literals with no interpolation
		template := node.AsTemplateExpression()
		return template != nil && len(template.TemplateSpans.Nodes) == 0
	case ast.KindNoSubstitutionTemplateLiteral:
		return true
	case ast.KindTaggedTemplateExpression:
		// Support tagged template expressions only with no interpolation
		tagged := node.AsTaggedTemplateExpression()
		if tagged.Template.Kind == ast.KindNoSubstitutionTemplateLiteral {
			return true
		}
		if tagged.Template.Kind == ast.KindTemplateExpression {
			template := tagged.Template.AsTemplateExpression()
			return template != nil && len(template.TemplateSpans.Nodes) == 0
		}
		return false
	default:
		return false
	}
}

func getStaticMemberAccessValue(ctx rule.RuleContext, node *ast.Node) string {
	// Get the name of a class member
	var nameNode *ast.Node

	if ast.IsPropertyDeclaration(node) {
		nameNode = node.AsPropertyDeclaration().Name()
	} else if ast.IsMethodDeclaration(node) {
		nameNode = node.AsMethodDeclaration().Name()
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
	// Handle computed property names
	if nameNode.Kind == ast.KindComputedPropertyName {
		computed := nameNode.AsComputedPropertyName()
		// For computed properties, get the name from the expression itself
		return extractPropertyNameFromExpression(ctx, computed.Expression)
	}

	// Handle regular identifiers
	if nameNode.Kind == ast.KindIdentifier {
		return nameNode.AsIdentifier().Text
	}

	// Handle string literals as property names
	if ast.IsLiteralExpression(nameNode) {
		text := nameNode.Text()
		// Remove quotes for string literals to normalize the name
		if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
			return text[1 : len(text)-1]
		}
		return text
	}

	return ""
}

func extractPropertyNameFromExpression(ctx rule.RuleContext, expr *ast.Node) string {
	// Handle string/numeric literals
	if ast.IsLiteralExpression(expr) {
		text := expr.Text()
		// Remove quotes for string literals to normalize the name
		if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
			return text[1 : len(text)-1]
		}
		return text
	}

	// Handle identifiers (like variable references)
	if expr.Kind == ast.KindIdentifier {
		// For identifiers in computed properties, we return a special marker
		// to indicate this is a dynamic property name
		return "[" + expr.AsIdentifier().Text + "]"
	}

	return ""
}

func isStaticMemberAccessOfValue(ctx rule.RuleContext, node *ast.Node, name string) bool {
	return getStaticMemberAccessValue(ctx, node) == name
}

func isAssignee(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check if this is the left side of an assignment
	if ast.IsBinaryExpression(parent) {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			return binary.Left == node
		}
	}

	return false
}

func isFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return ast.IsFunctionDeclaration(node) ||
		ast.IsFunctionExpression(node) ||
		ast.IsArrowFunction(node) ||
		ast.IsMethodDeclaration(node) ||
		ast.IsGetAccessorDeclaration(node) ||
		ast.IsSetAccessorDeclaration(node) ||
		ast.IsConstructorDeclaration(node)
}

var ClassLiteralPropertyStyleRule = rule.Rule{
	Name: "class-literal-property-style",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		style := "fields" // default option

		// Parse options - handle both string and array formats
		if options != nil {
			switch opts := options.(type) {
			case string:
				style = opts
			case []interface{}:
				if len(opts) > 0 {
					if s, ok := opts[0].(string); ok {
						style = s
					}
				}
			}
		}

		var propertiesInfoStack []*propertiesInfo

		listeners := rule.RuleListeners{}

		// Only add the getter check when style is "fields"
		if style == "fields" {
			listeners[ast.KindGetAccessor] = func(node *ast.Node) {
				getter := node.AsGetAccessorDeclaration()

				// Skip if getter has override modifier
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsOverride) {
					return
				}

				if getter.Body == nil {
					return
				}

				if !ast.IsBlock(getter.Body) {
					return
				}

				block := getter.Body.AsBlock()
				if block == nil || len(block.Statements.Nodes) == 0 {
					return
				}

				// Check if it's a single return statement with a literal
				if len(block.Statements.Nodes) != 1 {
					return
				}

				stmt := block.Statements.Nodes[0]
				if !ast.IsReturnStatement(stmt) {
					return
				}

				returnStmt := stmt.AsReturnStatement()
				if returnStmt.Expression == nil || !isSupportedLiteral(returnStmt.Expression) {
					return
				}

				name := getStaticMemberAccessValue(ctx, node)

				// Check if there's a corresponding setter
				if name != "" && node.Parent != nil {
					members := node.Parent.Members()
					if members != nil {
						for _, member := range members {
							if ast.IsSetAccessorDeclaration(member) && isStaticMemberAccessOfValue(ctx, member, name) {
								return // Skip if there's a setter with the same name
							}
						}
					}
				}

				// Report with suggestion to convert to readonly field
				// For the fix text, we need to get the actual text of the name and value
				nameNode := getter.Name()
				var nameText string
				if nameNode.Kind == ast.KindComputedPropertyName {
					// For computed properties, get the full text including brackets
					nameText = strings.TrimSpace(string(ctx.SourceFile.Text()[nameNode.Pos():nameNode.End()]))
				} else {
					// For regular identifiers, just get the text
					nameText = nameNode.Text()
				}
				
				valueText := strings.TrimSpace(string(ctx.SourceFile.Text()[returnStmt.Expression.Pos():returnStmt.Expression.End()]))

				var fixText string
				fixText += printNodeModifiers(node, "readonly")
				fixText += nameText
				fixText += fmt.Sprintf(" = %s;", valueText)

				// Report on the property name (node.key in TypeScript-ESLint)
				// For computed properties, report on the inner expression rather than the bracket
				reportNode := getter.Name()
				if reportNode.Kind == ast.KindComputedPropertyName {
					computed := reportNode.AsComputedPropertyName()
					if computed.Expression != nil {
						reportNode = computed.Expression
					}
				}
				ctx.ReportNodeWithSuggestions(reportNode, buildPreferFieldStyleMessage(),
					rule.RuleSuggestion{
						Message: buildPreferFieldStyleSuggestionMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, fixText),
						},
					})
			}
		}

		if style == "getters" {
			enterClassBody := func() {
				propertiesInfoStack = append(propertiesInfoStack, &propertiesInfo{
					excludeSet: make(map[string]bool),
					properties: []*ast.Node{},
				})
			}

			exitClassBody := func() {
				if len(propertiesInfoStack) == 0 {
					return
				}

				info := propertiesInfoStack[len(propertiesInfoStack)-1]
				propertiesInfoStack = propertiesInfoStack[:len(propertiesInfoStack)-1]

				for _, node := range info.properties {
					property := node.AsPropertyDeclaration()
					if property.Initializer == nil || !isSupportedLiteral(property.Initializer) {
						continue
					}

					name := getStaticMemberAccessValue(ctx, node)
					if name != "" && info.excludeSet[name] {
						continue
					}

					// Report with suggestion to convert to getter
					// Get the name and value text for the fix
					nameNode := property.Name()
					var nameText string
					if nameNode.Kind == ast.KindComputedPropertyName {
						// For computed properties, get the full text including brackets
						nameText = strings.TrimSpace(string(ctx.SourceFile.Text()[nameNode.Pos():nameNode.End()]))
					} else {
						// For regular identifiers, just get the text
						nameText = nameNode.Text()
					}
					
					valueText := strings.TrimSpace(string(ctx.SourceFile.Text()[property.Initializer.Pos():property.Initializer.End()]))

					var fixText string
					fixText += printNodeModifiers(node, "get")
					fixText += nameText
					fixText += fmt.Sprintf("() { return %s; }", valueText)

					// For computed property names, report on the inner expression rather than the bracket
					// For regular property names, report on the property name
					reportNode := property.Name()
					if reportNode.Kind == ast.KindComputedPropertyName {
						computed := reportNode.AsComputedPropertyName()
						if computed.Expression != nil {
							reportNode = computed.Expression
						}
					}

					ctx.ReportNodeWithSuggestions(reportNode, buildPreferGetterStyleMessage(),
						rule.RuleSuggestion{
							Message: buildPreferGetterStyleSuggestionMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, node, fixText),
							},
						})
				}
			}

			// Track class declarations and expressions to match TypeScript-ESLint ClassBody behavior
			// Since Go AST doesn't have a separate ClassBody node, use the class nodes themselves
			listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
				enterClassBody()
			}
			listeners[rule.ListenerOnExit(ast.KindClassDeclaration)] = func(node *ast.Node) {
				exitClassBody()
			}
			listeners[ast.KindClassExpression] = func(node *ast.Node) {
				enterClassBody()
			}
			listeners[rule.ListenerOnExit(ast.KindClassExpression)] = func(node *ast.Node) {
				exitClassBody()
			}

			// ThisExpression pattern matching for constructor exclusions
			// This matches the TypeScript-ESLint pattern: 'MethodDefinition[kind="constructor"] ThisExpression'
			listeners[ast.KindThisKeyword] = func(node *ast.Node) {
				// Check if this is inside a member expression (this.property or this['property'])
				if node.Parent == nil || (!ast.IsPropertyAccessExpression(node.Parent) && !ast.IsElementAccessExpression(node.Parent)) {
					return
				}

				memberExpr := node.Parent
				var propName string
				
				if ast.IsPropertyAccessExpression(memberExpr) {
					propAccess := memberExpr.AsPropertyAccessExpression()
					propName = extractPropertyName(ctx, propAccess.Name())
				} else if ast.IsElementAccessExpression(memberExpr) {
					elemAccess := memberExpr.AsElementAccessExpression()
					if ast.IsLiteralExpression(elemAccess.ArgumentExpression) {
						propName = extractPropertyName(ctx, elemAccess.ArgumentExpression)
					}
				}

				if propName == "" {
					return
				}

				// Walk up to find the containing function
				parent := memberExpr.Parent
				for parent != nil && !isFunction(parent) {
					parent = parent.Parent
				}

				// Check if this function is a constructor by checking its parent
				if parent != nil && parent.Parent != nil {
					if ast.IsMethodDeclaration(parent.Parent) {
						method := parent.Parent.AsMethodDeclaration()
						if method.Kind == ast.KindConstructorKeyword {
							// We're in a constructor - exclude this property
							if len(propertiesInfoStack) > 0 {
								info := propertiesInfoStack[len(propertiesInfoStack)-1]
								info.excludeSet[propName] = true
							}
						}
					} else if ast.IsConstructorDeclaration(parent.Parent) {
						// Direct constructor declaration
						if len(propertiesInfoStack) > 0 {
							info := propertiesInfoStack[len(propertiesInfoStack)-1]
							info.excludeSet[propName] = true
						}
					}
				}
			}

			// Track property assignments in constructors (keeping existing logic as fallback)
			listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}

				// Check if left side is a this.property or this['property'] access
				left := binary.Left
				if !ast.IsPropertyAccessExpression(left) && !ast.IsElementAccessExpression(left) {
					return
				}

				var thisExpr *ast.Node
				var propName string

				if ast.IsPropertyAccessExpression(left) {
					propAccess := left.AsPropertyAccessExpression()
					if propAccess.Expression.Kind == ast.KindThisKeyword {
						thisExpr = propAccess.Expression
						propName = extractPropertyName(ctx, propAccess.Name())
					}
				} else if ast.IsElementAccessExpression(left) {
					elemAccess := left.AsElementAccessExpression()
					if elemAccess.Expression.Kind == ast.KindThisKeyword {
						thisExpr = elemAccess.Expression
						if ast.IsLiteralExpression(elemAccess.ArgumentExpression) {
							propName = extractPropertyName(ctx, elemAccess.ArgumentExpression)
						}
					}
				}

				if thisExpr == nil || propName == "" {
					return
				}

				// Find the constructor by walking up the tree, but stop if we encounter another function
				current := node.Parent
				for current != nil && !ast.IsConstructorDeclaration(current) {
					// If we encounter another function declaration before reaching the constructor,
					// then this assignment is inside a nested function, not directly in the constructor
					if isFunction(current) && !ast.IsConstructorDeclaration(current) {
						return
					}
					current = current.Parent
				}

				if current != nil && len(propertiesInfoStack) > 0 {
					info := propertiesInfoStack[len(propertiesInfoStack)-1]
					info.excludeSet[propName] = true
				}
			}

			// Track readonly properties
			listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
					return // Not readonly
				}
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) {
					return // Declare modifier
				}
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsOverride) {
					return // Override modifier
				}

				if len(propertiesInfoStack) > 0 {
					info := propertiesInfoStack[len(propertiesInfoStack)-1]
					info.properties = append(info.properties, node)
				}
			}
		}

		return listeners
	},
}
