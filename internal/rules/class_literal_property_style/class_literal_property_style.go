package class_literal_property_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
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
		if ast.IsLiteralExpression(computed.Expression) {
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, computed.Expression)
			text := string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
			// Remove quotes for string literals to normalize the name
			if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
				return text[1 : len(text)-1]
			}
			return text
		}
		// Handle identifier expressions in computed properties
		if computed.Expression.Kind == ast.KindIdentifier {
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, computed.Expression)
			return string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
		}
		return ""
	}

	// Handle regular identifiers
	if nameNode.Kind == ast.KindIdentifier {
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		return string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
	}

	// Handle string literals as property names
	if ast.IsLiteralExpression(nameNode) {
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		text := string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
		// Remove quotes for string literals to normalize the name
		if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
			return text[1 : len(text)-1]
		}
		return text
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

		// Parse options
		if options != nil {
			if optionsSlice, ok := options.([]interface{}); ok && len(optionsSlice) > 0 {
				if s, ok := optionsSlice[0].(string); ok {
					style = s
				}
			}
		}

		var propertiesInfoStack []*propertiesInfo

		listeners := rule.RuleListeners{}

		if style == "fields" {
			// When preferring fields, check getters that return literals
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

				name, _ := utils.GetNameFromMember(ctx.SourceFile, getter.Name())

				// Check if there's a corresponding setter
				if name != "" && node.Parent != nil {
					members := node.Parent.Members()
					if members != nil {
						for _, member := range members {
							if ast.IsSetAccessorDeclaration(member) {
								setterName, _ := utils.GetNameFromMember(ctx.SourceFile, member.AsSetAccessorDeclaration().Name())
								if setterName == name {
									return // Skip if there's a setter with the same name
								}
							}
						}
					}
				}

				// Report with suggestion to convert to readonly field
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, getter.Name())
				valueRange := utils.TrimNodeTextRange(ctx.SourceFile, returnStmt.Expression)

				nameText := string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
				valueText := string(ctx.SourceFile.Text()[valueRange.Pos():valueRange.End()])

				isComputed := getter.Name().Kind == ast.KindComputedPropertyName

				var fixText string
				fixText += printNodeModifiers(node, "readonly")
				if isComputed {
					fixText += nameText // nameText already includes brackets for computed properties
				} else {
					fixText += nameText
				}
				fixText += fmt.Sprintf(" = %s;", valueText)

				// For computed property names, report on the opening bracket
				// For regular property names, report on the getter name
				var reportNode *ast.Node
				if isComputed {
					// For computed properties, report on the bracket
					reportNode = getter.Name()
				} else {
					reportNode = getter.Name()
				}

				ctx.ReportNodeWithSuggestions(reportNode, buildPreferFieldStyleMessage(),
					rule.RuleSuggestion{
						Message: buildPreferFieldStyleSuggestionMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, fixText),
						},
					})
			}
		} else if style == "getters" {
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
					nameRange := utils.TrimNodeTextRange(ctx.SourceFile, property.Name())
					valueRange := utils.TrimNodeTextRange(ctx.SourceFile, property.Initializer)

					nameText := string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])
					valueText := string(ctx.SourceFile.Text()[valueRange.Pos():valueRange.End()])

					isComputed := property.Name().Kind == ast.KindComputedPropertyName

					var fixText string
					fixText += printNodeModifiers(node, "get")
					if isComputed {
						fixText += nameText // nameText already includes brackets for computed properties
					} else {
						fixText += nameText
					}
					fixText += fmt.Sprintf("() { return %s; }", valueText)

					// For computed property names, report on the entire property
					// For regular property names, report on the property name
					reportNode := property.Name()

					ctx.ReportNodeWithSuggestions(reportNode, buildPreferGetterStyleMessage(),
						rule.RuleSuggestion{
							Message: buildPreferGetterStyleSuggestionMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, node, fixText),
							},
						})
				}
			}

			// When preferring getters, track readonly properties and exclude assigned ones
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

			// Track property assignments in constructors
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