package consistent_type_definitions

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func buildInterfaceOverTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "interfaceOverType",
		Description: "Use an `interface` instead of a `type`.",
	}
}

func buildTypeOverInterfaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeOverInterface",
		Description: "Use a `type` instead of an `interface`.",
	}
}

// Get node text with proper range handling
func getNodeText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
}

// Unwrap parentheses around a type to get the actual type literal
func unwrapParentheses(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	current := node
	for current.Kind == ast.KindParenthesizedType {
		parenthesized := current.AsParenthesizedTypeNode()
		current = parenthesized.Type
		if current == nil {
			break
		}
	}

	return current
}

// Check if a character is valid in an identifier
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '$'
}

// Check if a node is within a declare global module declaration
func isWithinDeclareGlobalModule(ctx rule.RuleContext, node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindModuleDeclaration {
			moduleDecl := current.AsModuleDeclaration()
			// Check if this is a global module declaration with declare modifier
			if moduleDecl.Name() != nil &&
			   ast.IsIdentifier(moduleDecl.Name()) &&
			   moduleDecl.Name().AsIdentifier().Text == "global" {
				// Check for declare modifier
				if moduleDecl.Modifiers() != nil {
					for _, modifier := range moduleDecl.Modifiers().Nodes {
						if modifier.Kind == ast.KindDeclareKeyword {
							return true
						}
					}
				}
			}
		}
		current = current.Parent
	}
	return false
}

// Convert type alias with type literal to interface
func convertTypeToInterface(ctx rule.RuleContext, node *ast.Node) []rule.RuleFix {
	typeAlias := node.AsTypeAliasDeclaration()
	sourceFile := ctx.SourceFile
	text := string(sourceFile.Text())

	// Get the actual type literal node, potentially unwrapping parentheses
	actualTypeLiteral := unwrapParentheses(typeAlias.Type)
	if actualTypeLiteral.Kind != ast.KindTypeLiteral {
		return []rule.RuleFix{}
	}

	// Find 'type' keyword
	typeStart := int(node.Pos())
	nameStart := int(typeAlias.Name().Pos())

	typeKeywordStart := -1
	for i := typeStart; i < nameStart; i++ {
		if i+4 <= len(text) && text[i:i+4] == "type" {
			// Make sure it's a word boundary
			if (i == 0 || !isIdentifierChar(text[i-1])) &&
			   (i+4 >= len(text) || !isIdentifierChar(text[i+4])) {
				typeKeywordStart = i
				break
			}
		}
	}

	if typeKeywordStart == -1 {
		return []rule.RuleFix{}
	}

	// Find the end position to replace from (after name/type params)
	replaceFromPos := int(typeAlias.Name().End())
	if typeAlias.TypeParameters != nil && len(typeAlias.TypeParameters.Nodes) > 0 {
		replaceFromPos = int(typeAlias.TypeParameters.End())
	}

	// Find the equals sign position
	equalsStart := -1
	equalsEnd := -1
	for i := replaceFromPos; i < len(text) && i < int(typeAlias.Type.Pos()); i++ {
		if text[i] == '=' {
			// Find start of whitespace before equals
			equalsStart = i
			for equalsStart > replaceFromPos && (text[equalsStart-1] == ' ' || text[equalsStart-1] == '\t') {
				equalsStart--
			}

			// Find end after equals (including trailing whitespace)
			equalsEnd = i + 1
			for equalsEnd < len(text) && (text[equalsEnd] == ' ' || text[equalsEnd] == '\t') {
				equalsEnd++
			}
			break
		}
	}

	// Find the end positions we need
	statementEnd := int(node.End())

	fixes := []rule.RuleFix{
		// Replace 'type' with 'interface'
		{
			Text:  "interface",
			Range: core.TextRange{}.WithPos(typeKeywordStart).WithEnd(typeKeywordStart + 4),
		},
	}

	// Replace equals and everything up to the actual type literal
	if equalsStart >= 0 {
		// Replace the equals and everything up to the type (including parentheses) with just a space,
		// then insert the type literal content
		fixes = append(fixes, rule.RuleFix{
			Text:  " " + getNodeText(ctx, actualTypeLiteral),
			Range: core.TextRange{}.WithPos(equalsStart).WithEnd(int(typeAlias.Type.End())),
		})
	}

	// Remove everything from end of original type to end of statement (semicolon, etc.)
	if int(typeAlias.Type.End()) < statementEnd {
		fixes = append(fixes, rule.RuleFix{
			Text:  "",
			Range: core.TextRange{}.WithPos(int(typeAlias.Type.End())).WithEnd(statementEnd),
		})
	}

	return fixes
}

// Convert interface to type alias
func convertInterfaceToType(ctx rule.RuleContext, node *ast.Node) []rule.RuleFix {
	interfaceDecl := node.AsInterfaceDeclaration()
	text := string(ctx.SourceFile.Text())

	// Find 'interface' keyword
	nodeStart := int(node.Pos())
	nameStart := int(interfaceDecl.Name().Pos())

	interfaceKeywordStart := -1
	for i := nodeStart; i < nameStart; i++ {
		if i+9 <= len(text) && text[i:i+9] == "interface" {
			// Make sure it's a word boundary
			if (i == 0 || !isIdentifierChar(text[i-1])) &&
			   (i+9 >= len(text) || !isIdentifierChar(text[i+9])) {
				interfaceKeywordStart = i
				break
			}
		}
	}

	if interfaceKeywordStart == -1 {
		return []rule.RuleFix{}
	}

	// Find the opening brace by searching forward from the name
	nameEnd := int(interfaceDecl.Name().End())
	if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
		nameEnd = int(interfaceDecl.TypeParameters.End())
	}

	// Find the opening brace
	openBracePos := -1
	for i := nameEnd; i < len(text); i++ {
		if text[i] == '{' {
			openBracePos = i
			break
		}
	}

	if openBracePos == -1 {
		return []rule.RuleFix{} // Can't find opening brace
	}

	// Find position to start replacement from (should be right after name/type params)
	replaceFromPos := nameEnd

	// If we have type parameters, we need to ensure we're after the closing >
	if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
		// Search forward to find the closing >
		for i := nameEnd; i < openBracePos && i < len(text); i++ {
			if text[i] == '>' {
				replaceFromPos = i + 1
				break
			}
		}
	}

	// Find the opening brace position
	bodyEnd := int(interfaceDecl.Members.End())

	// Find the actual closing brace character
	closeBracePos := -1
	for i := bodyEnd - 1; i >= openBracePos && i < len(text); i++ {
		if text[i] == '}' {
			closeBracePos = i + 1 // Position after the closing brace
			break
		}
	}

	fixes := []rule.RuleFix{
		// Replace 'interface' with 'type'
		{
			Text:  "type",
			Range: core.TextRange{}.WithPos(interfaceKeywordStart).WithEnd(interfaceKeywordStart + 9),
		},
	}

	// Insert ' = ' before the opening brace
	if openBracePos >= 0 {
		fixes = append(fixes, rule.RuleFix{
			Text:  " = ",
			Range: core.TextRange{}.WithPos(replaceFromPos).WithEnd(openBracePos),
		})
	}

	// Handle extends clauses - convert to intersection types
	if interfaceDecl.HeritageClauses != nil {
		for _, clause := range interfaceDecl.HeritageClauses.Nodes {
			if clause.Kind == ast.KindHeritageClause {
				heritageClause := clause.AsHeritageClause()
				if heritageClause.Token == ast.KindExtendsKeyword && len(heritageClause.Types.Nodes) > 0 {
					// Add intersection types after the body
					intersectionText := ""
					for _, heritageType := range heritageClause.Types.Nodes {
						typeText := getNodeText(ctx, heritageType)
						intersectionText += fmt.Sprintf(" & %s", typeText)
					}

					// Insert all intersection types at once after the closing brace
					insertPos := bodyEnd
					if closeBracePos >= 0 {
						insertPos = closeBracePos
					}
					fixes = append(fixes, rule.RuleFix{
						Text:  intersectionText,
						Range: core.TextRange{}.WithPos(insertPos).WithEnd(insertPos),
					})
				}
			}
		}
	}

	// Handle export default interfaces by checking if the text starts with "export default"
	interfaceNodeStart := int(node.Pos())
	isExportDefault := false

	// Look backwards from interface keyword to see if we have "export default"
	searchStart := interfaceNodeStart - 20
	if searchStart < 0 {
		searchStart = 0
	}

	textBefore := text[searchStart:interfaceKeywordStart]
	if strings.Contains(textBefore, "export") && strings.Contains(textBefore, "default") {
		isExportDefault = true
	}

	if isExportDefault {
		// Find the start of "export"
		exportStart := -1
		for i := searchStart; i < interfaceKeywordStart; i++ {
			if i+6 <= len(text) && text[i:i+6] == "export" {
				exportStart = i
				break
			}
		}

		if exportStart >= 0 {
			// Remove "export default " before interface
			fixes = append(fixes, rule.RuleFix{
				Text:  "",
				Range: core.TextRange{}.WithPos(exportStart).WithEnd(interfaceKeywordStart),
			})

			// Add export default after the type declaration
			interfaceName := ""
			if interfaceDecl.Name() != nil && ast.IsIdentifier(interfaceDecl.Name()) {
				interfaceName = interfaceDecl.Name().AsIdentifier().Text
			}

			insertPos := bodyEnd
			if closeBracePos >= 0 {
				insertPos = closeBracePos
			}
			fixes = append(fixes, rule.RuleFix{
				Text:  fmt.Sprintf("\nexport default %s", interfaceName),
				Range: core.TextRange{}.WithPos(insertPos).WithEnd(insertPos),
			})
		}
	}

	return fixes
}

var ConsistentTypeDefinitionsRule = rule.Rule{
	Name: "consistent-type-definitions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default option is "interface"
		option := "interface"

		// Parse options
		if options != nil {
			// Handle different option formats
			switch v := options.(type) {
			case string:
				// Direct string option
				option = v
			case []interface{}:
				// Array format - take first element
				if len(v) > 0 {
					if optStr, ok := v[0].(string); ok {
						option = optStr
					}
				}
			}
		}

		listeners := rule.RuleListeners{}

		if option == "interface" {
			// Report type aliases with type literals that should be interfaces
			listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()

				// Check if the type is a type literal (object type), potentially wrapped in parentheses
				if typeAlias.Type != nil {
					actualType := unwrapParentheses(typeAlias.Type)
					if actualType != nil && actualType.Kind == ast.KindTypeLiteral {
						// Report on the identifier, not the whole declaration
						fixes := convertTypeToInterface(ctx, node)
						ctx.ReportNodeWithFixes(typeAlias.Name(), buildInterfaceOverTypeMessage(), fixes...)
					}
				}
			}
		} else if option == "type" {
			// Report interfaces that should be type aliases
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()

				// Check if this is within a declare global module - if so, don't provide fixes
				withinDeclareGlobal := isWithinDeclareGlobalModule(ctx, node)

				if withinDeclareGlobal {
					// Report without fixes
					ctx.ReportNode(interfaceDecl.Name(), buildTypeOverInterfaceMessage())
				} else {
					// Report with fixes
					fixes := convertInterfaceToType(ctx, node)
					ctx.ReportNodeWithFixes(interfaceDecl.Name(), buildTypeOverInterfaceMessage(), fixes...)
				}
			}
		}

		return listeners
	},
}
