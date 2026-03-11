package consistent_type_definitions

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type DefinitionStyle string

const (
	DefinitionStyleInterface DefinitionStyle = "interface"
	DefinitionStyleType      DefinitionStyle = "type"
)

type ConsistentTypeDefinitionsOptions struct {
	Style DefinitionStyle `json:"style"`
}

// ConsistentTypeDefinitionsRule enforces consistent type definitions
var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-definitions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeDefinitionsOptions{
		Style: DefinitionStyleInterface,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if str, ok := optArray[0].(string); ok {
				opts.Style = DefinitionStyle(str)
			}
		} else if str, ok := options.(string); ok {
			opts.Style = DefinitionStyle(str)
		}
	}

	sourceText := ctx.SourceFile.Text()

	// Helper to get source text for a range
	getSourceRange := func(start, end int) string {
		return sourceText[start:end]
	}

	// Helper to check if a type is an object type literal
	isObjectTypeLiteral := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}
		return typeNode.Kind == ast.KindTypeLiteral
	}

	// Helper to check if a type alias is a simple object type (not a union, intersection, etc.)
	// Unwraps any number of parenthesized type wrappers before checking.
	isSimpleObjectType := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}

		// Unwrap all layers of parenthesized types
		unwrapped := ast.SkipTypeParentheses(typeNode)
		return isObjectTypeLiteral(unwrapped)
	}

	// Helper to check if a modifier list includes a specific modifier kind
	hasModifier := func(modifiers *ast.ModifierList, kind ast.Kind) bool {
		if modifiers == nil {
			return false
		}
		for _, mod := range modifiers.Nodes {
			if mod.Kind == kind {
				return true
			}
		}
		return false
	}

	// Helper to check if interface is in a declare global module
	isInDeclareGlobal := func(node *ast.Node) bool {
		current := node.Parent
		for current != nil {
			if current.Kind == ast.KindModuleDeclaration {
				moduleDecl := current.AsModuleDeclaration()
				if moduleDecl != nil && moduleDecl.Name() != nil {
					if ast.IsIdentifier(moduleDecl.Name()) {
						ident := moduleDecl.Name().AsIdentifier()
						if ident != nil && ident.Text == "global" {
							// Only return true if module has 'declare' keyword
							if hasModifier(moduleDecl.Modifiers(), ast.KindDeclareKeyword) {
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

	// Helper to build modifiers prefix text (e.g. "export ", "export declare ")
	// Returns the prefix and whether it has export/default/declare.
	// Note: hasDefault is intentionally excluded from prefix because
	// "export default interface" requires special handling (split into type decl + separate default export).
	getModifiersInfo := func(modifiers *ast.ModifierList) (prefix string, hasExport bool, hasDefault bool, hasDeclare bool) {
		if modifiers == nil {
			return "", false, false, false
		}
		for _, mod := range modifiers.Nodes {
			switch mod.Kind {
			case ast.KindExportKeyword:
				hasExport = true
			case ast.KindDefaultKeyword:
				hasDefault = true
			case ast.KindDeclareKeyword:
				hasDeclare = true
			}
		}
		var parts []string
		if hasExport {
			parts = append(parts, "export")
		}
		if hasDeclare {
			parts = append(parts, "declare")
		}
		if len(parts) > 0 {
			prefix = strings.Join(parts, " ") + " "
		}
		return
	}

	// Helper to get type parameters text including angle brackets
	getTypeParamsText := func(typeParams *ast.NodeList) string {
		if typeParams == nil || len(typeParams.Nodes) == 0 {
			return ""
		}
		firstParam := typeParams.Nodes[0]
		lastParam := typeParams.Nodes[len(typeParams.Nodes)-1]
		firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
		lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
		// Include the angle brackets
		start := firstRange.Pos() - 1
		end := lastRange.End() + 1
		return getSourceRange(start, end)
	}

	checkTypeAlias := func(node *ast.Node) {
		if opts.Style != DefinitionStyleInterface {
			return
		}

		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias == nil {
			return
		}

		// Only report if it's a simple object type literal
		if !isSimpleObjectType(typeAlias.Type) {
			return
		}

		// Build the fix: convert type alias to interface
		// Get name text
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.Name())
		nameText := getSourceRange(nameRange.Pos(), nameRange.End())

		// Get type parameters
		typeParamsText := getTypeParamsText(typeAlias.TypeParameters)

		// Get modifiers prefix
		prefix, _, _, _ := getModifiersInfo(typeAlias.Modifiers())

		// Get the body text from the unwrapped type literal
		unwrapped := ast.SkipTypeParentheses(typeAlias.Type)
		bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, unwrapped)
		bodyText := getSourceRange(bodyRange.Pos(), bodyRange.End())

		// Check for comments between name/type-params and '='
		// We need to preserve any comments that exist between the name and the body
		var commentText string
		nameOrTypeParamsEnd := nameRange.End()
		if typeAlias.TypeParameters != nil && len(typeAlias.TypeParameters.Nodes) > 0 {
			lastParam := typeAlias.TypeParameters.Nodes[len(typeAlias.TypeParameters.Nodes)-1]
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
			nameOrTypeParamsEnd = lastRange.End() + 1 // +1 for '>'
		}

		// Check for comments in the region between name and '='
		typeRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.Type)
		betweenText := getSourceRange(nameOrTypeParamsEnd, typeRange.Pos())
		// Look for /* ... */ style comments
		if idx := strings.Index(betweenText, "/*"); idx >= 0 {
			endIdx := strings.Index(betweenText, "*/")
			if endIdx >= 0 {
				commentText = " " + strings.TrimSpace(betweenText[idx:endIdx+2]) + " "
			}
		}

		// Build the replacement
		var replacement string
		if commentText != "" {
			replacement = prefix + "interface " + nameText + typeParamsText + commentText + bodyText
		} else {
			replacement = prefix + "interface " + nameText + typeParamsText + " " + bodyText
		}

		// Determine the range to replace: from start of declaration to end
		// Need to handle trailing semicolon
		declRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

		ctx.ReportNodeWithFixes(typeAlias.Name(), rule.RuleMessage{
			Id:          "interfaceOverType",
			Description: "Use an interface instead of a type literal.",
		}, rule.RuleFixReplaceRange(declRange, replacement))
	}

	checkInterface := func(node *ast.Node) {
		if opts.Style != DefinitionStyleType {
			return
		}

		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil {
			return
		}

		msg := rule.RuleMessage{
			Id:          "typeOverInterface",
			Description: "Use a type literal instead of an interface.",
		}

		// Don't fix interfaces in declare global modules (see typescript-eslint #2707)
		if isInDeclareGlobal(node) {
			ctx.ReportNode(interfaceDecl.Name(), msg)
			return
		}

		// Get name text
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
		nameText := getSourceRange(nameRange.Pos(), nameRange.End())

		// Get type parameters
		typeParamsText := getTypeParamsText(interfaceDecl.TypeParameters)

		// Get modifiers info
		prefix, hasExport, hasDefault, _ := getModifiersInfo(interfaceDecl.Modifiers())

		// Get the body text: from '{' to '}'
		// The Members list gives us the members, but we need the '{...}' block
		// We'll scan for the opening '{' after the interface name/type-params/extends
		declRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

		// Find the opening '{' by scanning from after extends clause or name
		var bodyStartScanPos int
		if interfaceDecl.HeritageClauses != nil && len(interfaceDecl.HeritageClauses.Nodes) > 0 {
			lastClause := interfaceDecl.HeritageClauses.Nodes[len(interfaceDecl.HeritageClauses.Nodes)-1]
			bodyStartScanPos = lastClause.End()
		} else if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
			lastParam := interfaceDecl.TypeParameters.Nodes[len(interfaceDecl.TypeParameters.Nodes)-1]
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
			bodyStartScanPos = lastRange.End() + 1 // +1 for '>'
		} else {
			bodyStartScanPos = nameRange.End()
		}

		// Scan to find the opening '{'
		s := scanner.GetScannerForSourceFile(ctx.SourceFile, bodyStartScanPos)
		openBracePos := -1
		for s.TokenStart() < declRange.End() {
			if s.Token() == ast.KindOpenBraceToken {
				openBracePos = s.TokenStart()
				break
			}
			s.Scan()
		}
		if openBracePos == -1 {
			// Fallback: just report without fix
			ctx.ReportNode(interfaceDecl.Name(), msg)
			return
		}

		// Body is from '{' to the end of the interface declaration
		bodyText := getSourceRange(openBracePos, declRange.End())

		// Get extends types for intersection
		var extendsTypes []string
		if interfaceDecl.HeritageClauses != nil {
			for _, clause := range interfaceDecl.HeritageClauses.Nodes {
				heritageClause := clause.AsHeritageClause()
				if heritageClause == nil || heritageClause.Token != ast.KindExtendsKeyword {
					continue
				}
				if heritageClause.Types != nil {
					for _, typeNode := range heritageClause.Types.Nodes {
						typeRange := utils.TrimNodeTextRange(ctx.SourceFile, typeNode)
						extendsTypes = append(extendsTypes, getSourceRange(typeRange.Pos(), typeRange.End()))
					}
				}
			}
		}

		// Handle export default interface
		if hasExport && hasDefault {
			// Convert to: type Name = { ... }\nexport default Name
			replacement := "type " + nameText + typeParamsText + " = " + bodyText
			if len(extendsTypes) > 0 {
				replacement += " & " + strings.Join(extendsTypes, " & ")
			}
			replacement += "\nexport default " + nameText

			ctx.ReportNodeWithFixes(interfaceDecl.Name(), msg,
				rule.RuleFixReplaceRange(declRange, replacement))
			return
		}

		// Build replacement: type Name = { ... }
		var replacement string
		replacement = prefix + "type " + nameText + typeParamsText + " = " + bodyText
		if len(extendsTypes) > 0 {
			replacement += " & " + strings.Join(extendsTypes, " & ")
		}

		// Check for trailing semicolon after the interface declaration
		afterDeclEnd := declRange.End()
		if afterDeclEnd < len(sourceText) && sourceText[afterDeclEnd] == ';' {
			// Include the trailing semicolon in the range so it gets replaced
			// The semicolon after interface closing brace should be preserved in the output
			fixRange := declRange.WithEnd(afterDeclEnd + 1)
			replacement += ";"
			ctx.ReportNodeWithFixes(interfaceDecl.Name(), msg,
				rule.RuleFixReplaceRange(fixRange, replacement))
			return
		}

		ctx.ReportNodeWithFixes(interfaceDecl.Name(), msg,
			rule.RuleFixReplaceRange(declRange, replacement))
	}

	return rule.RuleListeners{
		ast.KindTypeAliasDeclaration: checkTypeAlias,
		ast.KindInterfaceDeclaration: checkInterface,
	}
}
