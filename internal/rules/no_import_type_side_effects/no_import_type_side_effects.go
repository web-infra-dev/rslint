package no_import_type_side_effects

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoImportTypeSideEffectsRule = rule.Rule{
	Name: "no-import-type-side-effects",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				
				// Skip if it's already a type-only import
				if importDecl.ImportClause != nil && importDecl.ImportClause.AsImportClause().IsTypeOnly {
					return
				}
				
				// Skip if no specifiers (side-effect import like: import 'mod';)
				if importDecl.ImportClause == nil {
					return
				}
				
				importClause := importDecl.ImportClause.AsImportClause()
				
				// We need to check if all named imports have inline type qualifiers
				// Skip if there's a default import or namespace import
				if importClause.Name() != nil {
					return
				}
				if importClause.NamedBindings != nil && ast.IsNamespaceImport(importClause.NamedBindings) {
					return
				}
				if importClause.NamedBindings == nil {
					return
				}
				
				// Only handle named imports
				if !ast.IsNamedImports(importClause.NamedBindings) {
					return
				}
				
				namedImports := importClause.NamedBindings.AsNamedImports()
				if namedImports.Elements == nil || len(namedImports.Elements.Nodes) == 0 {
					return
				}
				
				// Check if all specifiers have type-only imports
				var allTypeOnly = true
				var typeOnlySpecifiers []*ast.ImportSpecifier
				
				for _, element := range namedImports.Elements.Nodes {
					if !ast.IsImportSpecifier(element) {
						allTypeOnly = false
						break
					}
					
					specifier := element.AsImportSpecifier()
					// Check if specifier has type modifier
					if !specifier.IsTypeOnly {
						allTypeOnly = false
						break
					}
					
					typeOnlySpecifiers = append(typeOnlySpecifiers, specifier)
				}
				
				if !allTypeOnly || len(typeOnlySpecifiers) == 0 {
					return
				}
				
				// Report the issue
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{
					Id:          "useTopLevelQualifier",
					Description: "TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import.",
				}, createFix(ctx, node, typeOnlySpecifiers)...)
			},
		}
	},
}

func createFix(ctx rule.RuleContext, importNode *ast.Node, specifiers []*ast.ImportSpecifier) []rule.RuleFix {
	fixes := []rule.RuleFix{}
	
	// First, remove all inline type keywords
	for _, specifier := range specifiers {
		// Find the "type" keyword position
		// In TypeScript AST, the type keyword appears before the imported name
		// We need to remove "type " (including the space) from each specifier
		
		// Get the text range of the specifier  
		specifierRange := utils.TrimNodeTextRange(ctx.SourceFile, specifier.AsNode())
		
		// The PropertyName (if exists) or Name contains the identifier after "type"
		var identifierNode *ast.Node
		if specifier.PropertyName != nil {
			identifierNode = specifier.PropertyName
		} else {
			identifierNode = specifier.Name()
		}
		
		if identifierNode != nil {
			identifierRange := utils.TrimNodeTextRange(ctx.SourceFile, identifierNode)
			// The "type " keyword is between the specifier start and the identifier
			// Calculate the range to remove
			removeStart := specifierRange.Pos()
			removeEnd := identifierRange.Pos()
			
			if removeEnd > removeStart {
				fixes = append(fixes, rule.RuleFix{
					Range: core.NewTextRange(removeStart, removeEnd),
					Text: "",
				})
			}
		}
	}
	
	// Then, add "type" after the import keyword
	// Find the position after "import" keyword
	// The import keyword is at the beginning of the import declaration
	importStart := importNode.Pos()
	
	// We want to insert " type" after "import"
	// "import" is 6 characters long
	insertPos := importStart + 6
	
	fixes = append(fixes, rule.RuleFix{
		Range: core.NewTextRange(insertPos, insertPos),
		Text: " type",
	})
	
	return fixes
}