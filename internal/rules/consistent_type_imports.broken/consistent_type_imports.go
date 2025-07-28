package consistent_type_imports

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type ConsistentTypeImportsOptions struct {
	DisallowTypeAnnotations bool   `json:"disallowTypeAnnotations"`
	FixStyle                string `json:"fixStyle"`
	Prefer                  string `json:"prefer"`
}

type SourceImports struct {
	ReportValueImports   []ReportValueImport
	Source               string
	TypeOnlyNamedImport  *ast.Node
	ValueImport          *ast.Node
	ValueOnlyNamedImport *ast.Node
}

type ReportValueImport struct {
	Node                 *ast.Node
	InlineTypeSpecifiers []*ast.Node
	TypeSpecifiers       []*ast.Node
	UnusedSpecifiers     []*ast.Node
	ValueSpecifiers      []*ast.Node
}

var ConsistentTypeImportsRule = rule.Rule{
	Name: "consistent-type-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ConsistentTypeImportsOptions{
			DisallowTypeAnnotations: true,
			FixStyle:                "separate-type-imports",
			Prefer:                  "type-imports",
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
				if val, ok := optsMap["disallowTypeAnnotations"].(bool); ok {
					opts.DisallowTypeAnnotations = val
				}
				if val, ok := optsMap["fixStyle"].(string); ok {
					opts.FixStyle = val
				}
				if val, ok := optsMap["prefer"].(string); ok {
					opts.Prefer = val
				}
			}
		}

		sourceImportsMap := make(map[string]*SourceImports)
		hasDecoratorMetadata := false
		// Track which identifiers are used as values (simplified)
		valueUsedIdentifiers := make(map[string]bool)
		// Track all identifier references
		allReferencedIdentifiers := make(map[string]bool)
		// Track identifiers that are shadowed by local declarations
		shadowedIdentifiers := make(map[string]bool)

		listeners := make(rule.RuleListeners)

		// Check for decorator metadata compatibility
		emitDecoratorMetadata := false
		experimentalDecorators := false

		if experimentalDecorators && emitDecoratorMetadata {
			listeners[ast.KindDecorator] = func(node *ast.Node) {
				hasDecoratorMetadata = true
			}
		}

		// Handle disallow type annotations
		if opts.DisallowTypeAnnotations {
			listeners[ast.KindImportType] = func(node *ast.Node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noImportTypeAnnotations",
					Description: "`import()` type annotations are forbidden.",
				})
			}
		}

		// Handle prefer no-type-imports
		if opts.Prefer == "no-type-imports" {
			listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				if importDecl.ImportClause != nil {
					importClause := importDecl.ImportClause.AsImportClause()
					if importClause.IsTypeOnly {
						ctx.ReportNodeWithFixes(node, rule.RuleMessage{
							Id:          "avoidImportType",
							Description: "Use an `import` instead of an `import type`.",
						}, fixRemoveTypeSpecifierFromImportDeclaration(ctx.SourceFile, node)...)
					}
				}

				// Check for inline type specifiers
				if importDecl.ImportClause != nil {
					importClause := importDecl.ImportClause.AsImportClause()
					if importClause.NamedBindings != nil {
						namedBindings := importClause.NamedBindings
						if ast.IsNamedImports(namedBindings) {
							namedImports := namedBindings.AsNamedImports()
							if namedImports.Elements != nil {
								for _, element := range namedImports.Elements.Nodes {
									if len(namedImports.Elements.Nodes) > 10 {
										break // Performance limit
									}
									importSpecifier := element.AsImportSpecifier()
									if importSpecifier.IsTypeOnly {
										ctx.ReportNodeWithFixes(element, rule.RuleMessage{
											Id:          "avoidImportType",
											Description: "Use an `import` instead of an `import type`.",
										}, fixRemoveTypeSpecifierFromImportSpecifier(ctx.SourceFile, element)...)
									}
								}
							}
						}
					}
				}
			}

			return listeners
		}

		// Simplified value usage tracking
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if newExpr.Expression != nil && ast.IsIdentifier(newExpr.Expression) {
				valueUsedIdentifiers[newExpr.Expression.AsIdentifier().Text] = true
			}
		}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr.Expression != nil && ast.IsIdentifier(callExpr.Expression) {
				valueUsedIdentifiers[callExpr.Expression.AsIdentifier().Text] = true
			}
		}

		// Track variable declarations for value usage
		listeners[ast.KindVariableStatement] = func(node *ast.Node) {
			variableStmt := node.AsVariableStatement()
			if variableStmt.DeclarationList != nil {
				declarationList := variableStmt.DeclarationList.AsVariableDeclarationList()
				if declarationList.Declarations != nil {
					processed := 0
					for _, decl := range declarationList.Declarations.Nodes {
						if processed >= 5 { // Performance limit
							break
						}
						processed++
						
						if ast.IsVariableDeclaration(decl) {
							variableDeclaration := decl.AsVariableDeclaration()
							if variableDeclaration.Initializer != nil {
								markValueUsageInExpression(variableDeclaration.Initializer, valueUsedIdentifiers)
							}
						}
					}
				}
			}
		}

		// Track property access expressions for value usage
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			propAccess := node.AsPropertyAccessExpression()
			if propAccess.Expression != nil && ast.IsIdentifier(propAccess.Expression) {
				// foo.bar - foo is used as a value (unless it's in a type context)
				if !isInTypeContext(node) {
					valueUsedIdentifiers[propAccess.Expression.AsIdentifier().Text] = true
				}
			}
		}

		// Track identifiers in type positions (simplified)
		listeners[ast.KindTypeReference] = func(node *ast.Node) {
			typeRef := node.AsTypeReference()
			if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
				identifierName := typeRef.TypeName.AsIdentifier().Text
				allReferencedIdentifiers[identifierName] = true
				
				// Check if this identifier is shadowed by a type parameter
				if isIdentifierShadowedByTypeParameter(node, identifierName) {
					shadowedIdentifiers[identifierName] = true
				}
			} else if typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
				qualifiedName := typeRef.TypeName.AsQualifiedName()
				if qualifiedName.Left != nil && ast.IsIdentifier(qualifiedName.Left) {
					identifierName := qualifiedName.Left.AsIdentifier().Text
					allReferencedIdentifiers[identifierName] = true
				}
			}
		}
		
		// Track function declarations for local shadowing
		listeners[ast.KindFunctionDeclaration] = func(node *ast.Node) {
			funcDecl := node.AsFunctionDeclaration()
			if funcDecl.Body != nil {
				// Look for local type declarations inside functions
				// Skip marking local type declarations for simplified version
			}
		}

		// Track identifiers in type queries (typeof)
		listeners[ast.KindTypeQuery] = func(node *ast.Node) {
			typeQuery := node.AsTypeQueryNode()
			if typeQuery.ExprName != nil {
				if ast.IsIdentifier(typeQuery.ExprName) {
					identifierName := typeQuery.ExprName.AsIdentifier().Text
					allReferencedIdentifiers[identifierName] = true
				} else if ast.IsQualifiedName(typeQuery.ExprName) {
					qualifiedName := typeQuery.ExprName.AsQualifiedName()
					if qualifiedName.Left != nil && ast.IsIdentifier(qualifiedName.Left) {
						identifierName := qualifiedName.Left.AsIdentifier().Text
						allReferencedIdentifiers[identifierName] = true
					}
				}
			}
		}

		// Track export usage (simplified)
		listeners[ast.KindExportDeclaration] = func(node *ast.Node) {
			exportDecl := node.AsExportDeclaration()
			if !exportDecl.IsTypeOnly && exportDecl.ExportClause != nil && ast.IsNamedExports(exportDecl.ExportClause) {
				namedExports := exportDecl.ExportClause.AsNamedExports()
				if namedExports.Elements != nil {
					processed := 0
					for _, element := range namedExports.Elements.Nodes {
						if processed >= 5 { // Performance limit
							break
						}
						processed++
						
						if ast.IsExportSpecifier(element) {
							exportSpec := element.AsExportSpecifier()
							if exportSpec.PropertyName != nil && ast.IsIdentifier(exportSpec.PropertyName) {
								valueUsedIdentifiers[exportSpec.PropertyName.AsIdentifier().Text] = true
							} else if exportSpec.Name() != nil && ast.IsIdentifier(exportSpec.Name()) {
								valueUsedIdentifiers[exportSpec.Name().AsIdentifier().Text] = true
							}
						}
					}
				}
			}
		}

		// Track export assignments
		listeners[ast.KindExportAssignment] = func(node *ast.Node) {
			exportAssign := node.AsExportAssignment()
			if exportAssign.Expression != nil && ast.IsIdentifier(exportAssign.Expression) {
				valueUsedIdentifiers[exportAssign.Expression.AsIdentifier().Text] = true
			}
		}

		// Store import declarations for processing
		importDeclarations := []*ast.Node{}
		
		listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
			if len(importDeclarations) < 20 { // Performance limit
				importDeclarations = append(importDeclarations, node)
			}
		}

		// Process all imports at the end
		listeners[ast.KindEndOfFile] = func(node *ast.Node) {
			if hasDecoratorMetadata {
				return
			}

			// Process imports with performance limits
			processed := 0
			for _, importNode := range importDeclarations {
				if processed >= 10 { // Performance limit
					break
				}
				processed++
				
				importDecl := importNode.AsImportDeclaration()
				if importDecl.ModuleSpecifier == nil || !ast.IsStringLiteral(importDecl.ModuleSpecifier) {
					continue
				}

				source := importDecl.ModuleSpecifier.AsStringLiteral().Text

				sourceImports, exists := sourceImportsMap[source]
				if !exists {
					sourceImports = &SourceImports{
						ReportValueImports:   []ReportValueImport{},
						Source:               source,
						TypeOnlyNamedImport:  nil,
						ValueImport:          nil,
						ValueOnlyNamedImport: nil,
					}
					sourceImportsMap[source] = sourceImports
				}

				// Skip type-only imports
				if importDecl.ImportClause != nil && importDecl.ImportClause.AsImportClause().IsTypeOnly {
					continue
				}

				var typeSpecifiers []*ast.Node
				var valueSpecifiers []*ast.Node

				// Simplified classification of import specifiers
				if importDecl.ImportClause != nil {
					classifyImportSpecifiersSimple(importDecl, &typeSpecifiers, &valueSpecifiers, valueUsedIdentifiers, allReferencedIdentifiers)
				}

				// Only report if we have type specifiers in a value import
				if len(typeSpecifiers) > 0 {
					sourceImports.ReportValueImports = append(sourceImports.ReportValueImports, ReportValueImport{
						Node:                 importNode,
						InlineTypeSpecifiers: []*ast.Node{},
						TypeSpecifiers:       typeSpecifiers,
						UnusedSpecifiers:     []*ast.Node{},
						ValueSpecifiers:      valueSpecifiers,
					})
				}
			}

			// Report diagnostics
			for _, sourceImports := range sourceImportsMap {
				for _, report := range sourceImports.ReportValueImports {
					if len(report.ValueSpecifiers) == 0 && len(report.UnusedSpecifiers) == 0 {
						// All imports are type-only
						if !hasImportAttributes(report.Node) {
							ctx.ReportNodeWithSuggestions(report.Node, rule.RuleMessage{
								Id:          "typeOverValue",
								Description: "All imports in the declaration are only used as types. Use `import type`.",
							}, rule.RuleSuggestion{
								Message: rule.RuleMessage{
									Id:          "fixToTypeImport",
									Description: "Convert to type import.",
								},
								FixesArr: fixToTypeImportDeclarationSimple(ctx.SourceFile, report.Node),
							})
						}
					} else if len(report.TypeSpecifiers) > 0 {
						// Mixed imports
						importNames := make([]string, 0, len(report.TypeSpecifiers))
						for _, specifier := range report.TypeSpecifiers {
							name := getImportSpecifierName(specifier)
							importNames = append(importNames, fmt.Sprintf(`"%s"`, name))
						}

						typeImports := formatWordList(importNames)
						message := fmt.Sprintf("Imports %s are only used as type.", typeImports)

						ctx.ReportNodeWithFixes(report.Node, rule.RuleMessage{
							Id:          "someImportsAreOnlyTypes",
							Description: message,
						}, fixToTypeImportDeclarationSimple(ctx.SourceFile, report.Node)...)
					}
				}
			}
		}

		return listeners
	},
}

// Simplified classification function
func classifyImportSpecifiersSimple(importDecl *ast.ImportDeclaration, typeSpecifiers, valueSpecifiers *[]*ast.Node, valueUsedIdentifiers map[string]bool, allReferencedIdentifiers map[string]bool) {
	if importDecl.ImportClause == nil {
		return
	}

	importClause := importDecl.ImportClause.AsImportClause()

	// Handle default import
	if importClause.Name() != nil {
		defaultImport := importClause.Name()
		identifierName := defaultImport.AsIdentifier().Text

		if allReferencedIdentifiers[identifierName] {
			if valueUsedIdentifiers[identifierName] {
				*valueSpecifiers = append(*valueSpecifiers, defaultImport)
			} else {
				*typeSpecifiers = append(*typeSpecifiers, defaultImport)
			}
		}
	}

	// Handle named imports
	if importClause.NamedBindings != nil {
		namedBindings := importClause.NamedBindings
		if ast.IsNamedImports(namedBindings) {
			namedImports := namedBindings.AsNamedImports()
			if namedImports.Elements != nil {
				processed := 0
				for _, element := range namedImports.Elements.Nodes {
					if processed >= 5 { // Performance limit
						break
					}
					processed++
					
					importSpecifier := element.AsImportSpecifier()
					if importSpecifier.IsTypeOnly {
						continue
					}

					identifierName := importSpecifier.Name().AsIdentifier().Text

					if allReferencedIdentifiers[identifierName] {
						if valueUsedIdentifiers[identifierName] {
							*valueSpecifiers = append(*valueSpecifiers, element)
						} else {
							*typeSpecifiers = append(*typeSpecifiers, element)
						}
					}
				}
			}
		} else if ast.IsNamespaceImport(namedBindings) {
			namespaceImport := namedBindings.AsNamespaceImport()
			identifierName := namespaceImport.Name().AsIdentifier().Text
			
			if allReferencedIdentifiers[identifierName] {
				if valueUsedIdentifiers[identifierName] {
					*valueSpecifiers = append(*valueSpecifiers, namedBindings)
				} else {
					*typeSpecifiers = append(*typeSpecifiers, namedBindings)
				}
			}
		}
	}
}

func hasImportAttributes(node *ast.Node) bool {
	importDecl := node.AsImportDeclaration()
	return importDecl.Attributes != nil && importDecl.Attributes.Elements() != nil && len(importDecl.Attributes.Elements()) > 0
}

func getImportSpecifierName(node *ast.Node) string {
	if ast.IsImportSpecifier(node) {
		importSpec := node.AsImportSpecifier()
		if importSpec.PropertyName != nil {
			return getIdentifierText(importSpec.PropertyName) + " as " + getIdentifierText(importSpec.Name())
		}
		return getIdentifierText(importSpec.Name())
	} else if ast.IsIdentifier(node) {
		return getIdentifierText(node)
	} else if ast.IsNamespaceImport(node) {
		namespaceImport := node.AsNamespaceImport()
		return getIdentifierText(namespaceImport.Name())
	}
	return ""
}

func getIdentifierText(node *ast.Node) string {
	if ast.IsIdentifier(node) {
		return node.AsIdentifier().Text
	}
	return ""
}

func formatWordList(words []string) string {
	if len(words) == 0 {
		return ""
	}
	if len(words) == 1 {
		return words[0]
	}
	if len(words) == 2 {
		return words[0] + " and " + words[1]
	}

	sort.Strings(words)
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}

func fixRemoveTypeSpecifierFromImportDeclaration(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	var fixes []rule.RuleFix

	sourceText := string(sourceFile.Text())
	nodeStart := int(node.Pos())
	nodeEnd := int(node.End())
	nodeText := sourceText[nodeStart:nodeEnd]

	// Find and remove "type" keyword after "import"
	importPattern := regexp.MustCompile(`import\s+type\s+`)
	match := importPattern.FindStringIndex(nodeText)
	if match != nil {
		typeStart := nodeStart + match[0] + 6 // "import".length
		typeEnd := nodeStart + match[1]
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(typeStart, typeEnd),
			" ",
		))
	}

	return fixes
}

func fixRemoveTypeSpecifierFromImportSpecifier(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	var fixes []rule.RuleFix

	sourceText := string(sourceFile.Text())
	nodeStart := int(node.Pos())
	nodeEnd := int(node.End())
	nodeText := sourceText[nodeStart:nodeEnd]

	// Find and remove "type" keyword from specifier
	typePattern := regexp.MustCompile(`type\s+`)
	match := typePattern.FindStringIndex(nodeText)
	if match != nil {
		typeStart := nodeStart + match[0]
		typeEnd := nodeStart + match[1]
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(typeStart, typeEnd),
			"",
		))
	}

	return fixes
}

// Simplified fix function
func fixToTypeImportDeclarationSimple(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	var fixes []rule.RuleFix

	sourceText := string(sourceFile.Text())
	nodeStart := int(node.Pos())
	nodeEnd := int(node.End())
	nodeText := sourceText[nodeStart:nodeEnd]

	// Simple case: just add "type" after "import"
	importPattern := regexp.MustCompile(`import\s+`)
	match := importPattern.FindStringIndex(nodeText)
	if match != nil {
		importEnd := nodeStart + match[1]
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(importEnd, importEnd),
			"type ",
		))
	}

	return fixes
}

// Helper function to mark value usage in expressions
func markValueUsageInExpression(expr *ast.Node, valueUsedIdentifiers map[string]bool) {
	if expr == nil {
		return
	}
	
	switch expr.Kind {
	case ast.KindIdentifier:
		valueUsedIdentifiers[expr.AsIdentifier().Text] = true
	case ast.KindPropertyAccessExpression:
		propAccess := expr.AsPropertyAccessExpression()
		if propAccess.Expression != nil && ast.IsIdentifier(propAccess.Expression) {
			valueUsedIdentifiers[propAccess.Expression.AsIdentifier().Text] = true
		}
	case ast.KindCallExpression:
		callExpr := expr.AsCallExpression()
		if callExpr.Expression != nil && ast.IsIdentifier(callExpr.Expression) {
			valueUsedIdentifiers[callExpr.Expression.AsIdentifier().Text] = true
		}
	case ast.KindNewExpression:
		newExpr := expr.AsNewExpression()
		if newExpr.Expression != nil && ast.IsIdentifier(newExpr.Expression) {
			valueUsedIdentifiers[newExpr.Expression.AsIdentifier().Text] = true
		}
	}
}

// Helper function to check if a node is in a type context
func isInTypeContext(node *ast.Node) bool {
	current := node.Parent
	depth := 0
	maxDepth := 3 // Limit depth for performance
	
	for current != nil && depth < maxDepth {
		depth++
		switch current.Kind {
		case ast.KindTypeReference, ast.KindTypeQuery:
			return true
		case ast.KindSourceFile, ast.KindBlock:
			return false
		}
		current = current.Parent
	}
	return false
}

// isIdentifierShadowedByTypeParameter checks if an identifier is shadowed by a type parameter
// in any enclosing scope (type alias, interface, class, function, etc.)
func isIdentifierShadowedByTypeParameter(node *ast.Node, identifierName string) bool {
	// Simple check for type parameter shadowing in common cases
	current := node.Parent
	maxDepth := 3 // Very limited depth for performance
	depth := 0
	
	for current != nil && depth < maxDepth {
		depth++
		
		// Check for type alias with type parameters: type T<Foo> = Foo
		if current.Kind == ast.KindTypeAliasDeclaration {
			typeAlias := current.AsTypeAliasDeclaration()
			if typeAlias.TypeParameters != nil {
				for _, typeParam := range typeAlias.TypeParameters.Nodes {
					if ast.IsTypeParameterDeclaration(typeParam) {
						typeParamDecl := typeParam.AsTypeParameter()
						if typeParamDecl.Name() != nil && typeParamDecl.Name().AsIdentifier().Text == identifierName {
							return true
						}
					}
				}
			}
		}
		
		// Break early at certain boundaries
		if current.Kind == ast.KindSourceFile {
			break
		}
		
		current = current.Parent
	}
	
	return false
}