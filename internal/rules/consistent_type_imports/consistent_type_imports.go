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

		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
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
		// Track which identifiers are used as values
		valueUsedIdentifiers := make(map[string]bool)
		// Track all identifier references (for detecting unused imports)
		allReferencedIdentifiers := make(map[string]bool)
		// Track all identifier reference nodes
		allReferencedNodes := make(map[string][]*ast.Node)
		// Track identifiers that are shadowed by local declarations
		shadowedIdentifiers := make(map[string]bool)
		// Track all local declarations for shadowing analysis
		localDeclarations := make(map[string][]*ast.Node)

		listeners := make(rule.RuleListeners)

		// Check for decorator metadata compatibility
		emitDecoratorMetadata := false
		experimentalDecorators := false
		// Note: For now, we'll skip the compiler options check as the API may not be available
		// This is a simplification for the Go port

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

		// Track all local declarations for shadowing analysis
		listeners[ast.KindTypeAliasDeclaration] = func(node *ast.Node) {
			typeAlias := node.AsTypeAliasDeclaration()
			if typeAlias.Name() != nil {
				name := typeAlias.Name().AsIdentifier().Text
				localDeclarations[name] = append(localDeclarations[name], node)
			}
		}

		listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl.Name() != nil {
				name := interfaceDecl.Name().AsIdentifier().Text
				localDeclarations[name] = append(localDeclarations[name], node)
			}
		}

		listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			if classDecl.Name() != nil {
				name := classDecl.Name().AsIdentifier().Text
				localDeclarations[name] = append(localDeclarations[name], node)
			}
		}

		listeners[ast.KindFunctionDeclaration] = func(node *ast.Node) {
			funcDecl := node.AsFunctionDeclaration()
			if funcDecl.Name() != nil {
				name := funcDecl.Name().AsIdentifier().Text
				localDeclarations[name] = append(localDeclarations[name], node)
			}
		}

		// Track value usage in variable declarations
		listeners[ast.KindVariableStatement] = func(node *ast.Node) {
			variableStmt := node.AsVariableStatement()
			if variableStmt.DeclarationList != nil {
				declarationList := variableStmt.DeclarationList.AsVariableDeclarationList()
				if declarationList.Declarations != nil {
					// Limit processing to prevent performance issues
					maxDecls := 50
					processed := 0
					for _, decl := range declarationList.Declarations.Nodes {
						if processed >= maxDecls {
							break
						}
						processed++
						
						if ast.IsVariableDeclaration(decl) {
							variableDeclaration := decl.AsVariableDeclaration()
							// Track local declarations
							if variableDeclaration.Name() != nil && ast.IsIdentifier(variableDeclaration.Name()) {
								name := variableDeclaration.Name().AsIdentifier().Text
								localDeclarations[name] = append(localDeclarations[name], node)
							}
							// Track value usage in initializers
							if variableDeclaration.Initializer != nil {
								if ast.IsIdentifier(variableDeclaration.Initializer) {
									// const d = c - c is used as a value
									valueUsedIdentifiers[variableDeclaration.Initializer.AsIdentifier().Text] = true
								} else if variableDeclaration.Initializer.Kind == ast.KindAsExpression {
									// const d = {} as Type - check the expression part
									asExpr := variableDeclaration.Initializer.AsAsExpression()
									if asExpr.Expression != nil && ast.IsIdentifier(asExpr.Expression) {
										valueUsedIdentifiers[asExpr.Expression.AsIdentifier().Text] = true
									}
								}
							}
						}
					}
				}
			}
		}

		listeners[ast.KindEnumDeclaration] = func(node *ast.Node) {
			enumDecl := node.AsEnumDeclaration()
			if enumDecl.Name() != nil {
				name := enumDecl.Name().AsIdentifier().Text
				localDeclarations[name] = append(localDeclarations[name], node)
			}
		}

		listeners[ast.KindModuleDeclaration] = func(node *ast.Node) {
			moduleDecl := node.AsModuleDeclaration()
			if moduleDecl.Name() != nil && ast.IsIdentifier(moduleDecl.Name()) {
				name := moduleDecl.Name().AsIdentifier().Text
				localDeclarations[name] = append(localDeclarations[name], node)
			}
		}

		// Track all identifier references in type positions
		listeners[ast.KindTypeReference] = func(node *ast.Node) {
			typeRef := node.AsTypeReference()
			if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
				identifierName := typeRef.TypeName.AsIdentifier().Text
				allReferencedIdentifiers[identifierName] = true
				// Store reference for shadowing analysis
				refs := allReferencedNodes[identifierName]
				if len(refs) < 50 {
					allReferencedNodes[identifierName] = append(refs, typeRef.TypeName)
				}
			} else if typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
				// Handle qualified names like foo.Bar
				qualifiedName := typeRef.TypeName.AsQualifiedName()
				if qualifiedName.Left != nil && ast.IsIdentifier(qualifiedName.Left) {
					identifierName := qualifiedName.Left.AsIdentifier().Text
					allReferencedIdentifiers[identifierName] = true
					refs := allReferencedNodes[identifierName]
					if len(refs) < 50 {
						allReferencedNodes[identifierName] = append(refs, qualifiedName.Left)
					}
				}
			}
		}

		// Track identifiers in type queries (typeof)
		listeners[ast.KindTypeQuery] = func(node *ast.Node) {
			typeQuery := node.AsTypeQueryNode()
			if typeQuery.ExprName != nil && ast.IsIdentifier(typeQuery.ExprName) {
				identifierName := typeQuery.ExprName.AsIdentifier().Text
				allReferencedIdentifiers[identifierName] = true
				refs := allReferencedNodes[identifierName]
				if len(refs) < 50 {
					allReferencedNodes[identifierName] = append(refs, typeQuery.ExprName)
				}
			}
		}

		// Track identifiers in expression statements that aren't covered by other listeners
		listeners[ast.KindExpressionStatement] = func(node *ast.Node) {
			exprStmt := node.AsExpressionStatement()
			if exprStmt.Expression != nil && ast.IsIdentifier(exprStmt.Expression) {
				// Bare identifier expression - mark as value usage
				identifierName := exprStmt.Expression.AsIdentifier().Text
				allReferencedIdentifiers[identifierName] = true
				valueUsedIdentifiers[identifierName] = true
			}
		}

		// Track value usages of identifiers
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			// new Foo() - Foo is used as a value
			newExpr := node.AsNewExpression()
			if newExpr.Expression != nil && ast.IsIdentifier(newExpr.Expression) {
				identifierName := newExpr.Expression.AsIdentifier().Text
				valueUsedIdentifiers[identifierName] = true
			}
		}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			// Foo() - Foo is used as a value
			callExpr := node.AsCallExpression()
			if callExpr.Expression != nil && ast.IsIdentifier(callExpr.Expression) {
				valueUsedIdentifiers[callExpr.Expression.AsIdentifier().Text] = true
			}
		}

		// Track export statements as value usage
		listeners[ast.KindExportDeclaration] = func(node *ast.Node) {
			exportDecl := node.AsExportDeclaration()
			// Only track as value usage if it's not a type-only export
			if !exportDecl.IsTypeOnly && exportDecl.ExportClause != nil && ast.IsNamedExports(exportDecl.ExportClause) {
				namedExports := exportDecl.ExportClause.AsNamedExports()
				if namedExports.Elements != nil {
					// Limit number of exports to process
					maxExports := 50
					processed := 0
					for _, element := range namedExports.Elements.Nodes {
						if processed >= maxExports {
							break
						}
						processed++
						
						if ast.IsExportSpecifier(element) {
							exportSpec := element.AsExportSpecifier()
							if exportSpec.PropertyName != nil {
								// export { foo as bar } - foo is used as a value
								if ast.IsIdentifier(exportSpec.PropertyName) {
									valueUsedIdentifiers[exportSpec.PropertyName.AsIdentifier().Text] = true
								}
							} else if exportSpec.Name() != nil {
								// export { foo } - foo is used as a value
								if ast.IsIdentifier(exportSpec.Name()) {
									valueUsedIdentifiers[exportSpec.Name().AsIdentifier().Text] = true
								}
							}
						}
					}
				}
			}
		}

		// Track default exports as value usage
		listeners[ast.KindExportAssignment] = func(node *ast.Node) {
			exportAssign := node.AsExportAssignment()
			if exportAssign.Expression != nil {
				if ast.IsIdentifier(exportAssign.Expression) {
					// export = Foo - Foo is used as a value
					valueUsedIdentifiers[exportAssign.Expression.AsIdentifier().Text] = true
				} else if exportAssign.Expression.Kind == ast.KindAsExpression {
					// export = {} as A - check if this is a type-only usage
					asExpr := exportAssign.Expression.AsAsExpression()
					if asExpr.Expression != nil && ast.IsObjectLiteralExpression(asExpr.Expression) {
						// This is export = {} as A - A is only used as a type
						// Don't mark it as value usage
					} else if asExpr.Expression != nil && ast.IsIdentifier(asExpr.Expression) {
						// export = something as A - something is used as value
						valueUsedIdentifiers[asExpr.Expression.AsIdentifier().Text] = true
					}
				}
			}
		}
		
		// Handle ES6 export default
		// TODO: Find the correct way to handle export default
		// Currently commented out as KindExportDefault doesn't exist
		// listeners[ast.KindExportDefault] = func(node *ast.Node) {
		// 	// This catches export default Foo
		// 	parent := node.Parent
		// 	if parent != nil && ast.IsExportDeclaration(parent) {
		// 		exportDecl := parent.AsExportDeclaration()
		// 		if exportDecl.Expression != nil && ast.IsIdentifier(exportDecl.Expression) {
		// 			valueUsedIdentifiers[exportDecl.Expression.AsIdentifier().Text] = true
		// 		}
		// 	}
		// }



		// Track property access as value usage
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			propAccess := node.AsPropertyAccessExpression()
			if propAccess.Expression != nil && ast.IsIdentifier(propAccess.Expression) {
				// foo.bar - foo is used as a value
				valueUsedIdentifiers[propAccess.Expression.AsIdentifier().Text] = true
			}
		}

		// Store all import declarations for later processing
		importDeclarations := []*ast.Node{}

		// Collect import declarations
		listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
			importDeclarations = append(importDeclarations, node)
		}

		listeners[ast.KindEndOfFile] = func(node *ast.Node) {
			if hasDecoratorMetadata {
				return
			}

			// Process all import declarations now that we've collected value usages
			for _, importNode := range importDeclarations {
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

				if importDecl.ImportClause != nil && importDecl.ImportClause.AsImportClause().IsTypeOnly {
					if sourceImports.TypeOnlyNamedImport == nil && hasOnlyNamedImports(importDecl) {
						sourceImports.TypeOnlyNamedImport = importNode
					}
				} else if sourceImports.ValueOnlyNamedImport == nil && hasOnlyNamedImports(importDecl) {
					sourceImports.ValueOnlyNamedImport = importNode
					sourceImports.ValueImport = importNode
				} else if sourceImports.ValueImport == nil && hasDefaultImport(importDecl) {
					sourceImports.ValueImport = importNode
				}

				var typeSpecifiers []*ast.Node
				var inlineTypeSpecifiers []*ast.Node
				var valueSpecifiers []*ast.Node
				var unusedSpecifiers []*ast.Node

				// Analyze which imports are shadowed by local declarations
				analyzeShadowing(importDecl, localDeclarations, shadowedIdentifiers)

				if importDecl.ImportClause != nil {
					classifyImportSpecifiers(ctx, importDecl, &typeSpecifiers, &inlineTypeSpecifiers, &valueSpecifiers, &unusedSpecifiers, valueUsedIdentifiers, allReferencedIdentifiers, allReferencedNodes, shadowedIdentifiers)
				}

				if importDecl.ImportClause != nil && !importDecl.ImportClause.AsImportClause().IsTypeOnly && len(typeSpecifiers) > 0 {
					sourceImports.ReportValueImports = append(sourceImports.ReportValueImports, ReportValueImport{
						Node:                 importNode,
						InlineTypeSpecifiers: inlineTypeSpecifiers,
						TypeSpecifiers:       typeSpecifiers,
						UnusedSpecifiers:     unusedSpecifiers,
						ValueSpecifiers:      valueSpecifiers,
					})
				}
			}

			// Now report diagnostics for each source
			for _, sourceImports := range sourceImportsMap {
				if len(sourceImports.ReportValueImports) == 0 {
					continue
				}

				for _, report := range sourceImports.ReportValueImports {
					if len(report.ValueSpecifiers) == 0 && len(report.UnusedSpecifiers) == 0 && !isTypeOnlyImport(report.Node) {
						// All imports are only used as types (no value specifiers and no unused)
						// Check for import attributes/assertions
						if !hasImportAttributes(report.Node) {
							// Report entire import statement
							ctx.ReportNodeWithSuggestions(report.Node, rule.RuleMessage{
								Id:          "typeOverValue",
								Description: "All imports in the declaration are only used as types. Use `import type`.",
							}, rule.RuleSuggestion{
								Message: rule.RuleMessage{
									Id:          "fixToTypeImport",
									Description: "Convert to type import.",
								},
								FixesArr: fixToTypeImportDeclaration(ctx.SourceFile, report, sourceImports, opts.FixStyle),
							})
						}
					} else if len(report.TypeSpecifiers) > 0 {
						// Mixed type/value imports or has unused - some imports are only used as types
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
						}, fixToTypeImportDeclaration(ctx.SourceFile, report, sourceImports, opts.FixStyle)...)
					}
				}
			}
		}

		return listeners
	},
}

func hasOnlyNamedImports(importDecl *ast.ImportDeclaration) bool {
	if importDecl.ImportClause == nil {
		return false
	}

	// Check if there's no default import and only named imports
	importClause := importDecl.ImportClause.AsImportClause()
	hasDefault := importClause.Name() != nil
	hasNamed := importClause.NamedBindings != nil && ast.IsNamedImports(importClause.NamedBindings)

	return !hasDefault && hasNamed
}

func hasDefaultImport(importDecl *ast.ImportDeclaration) bool {
	return importDecl.ImportClause != nil && importDecl.ImportClause.AsImportClause().Name() != nil
}

func isTypeOnlyImport(node *ast.Node) bool {
	importDecl := node.AsImportDeclaration()
	return importDecl.ImportClause != nil && importDecl.ImportClause.AsImportClause().IsTypeOnly
}

func hasImportAttributes(node *ast.Node) bool {
	importDecl := node.AsImportDeclaration()
	return importDecl.Attributes != nil && importDecl.Attributes.Elements() != nil && len(importDecl.Attributes.Elements()) > 0
}

func classifyImportSpecifiers(ctx rule.RuleContext, importDecl *ast.ImportDeclaration, typeSpecifiers, inlineTypeSpecifiers, valueSpecifiers, unusedSpecifiers *[]*ast.Node, valueUsedIdentifiers map[string]bool, allReferencedIdentifiers map[string]bool, allReferencedNodes map[string][]*ast.Node, shadowedIdentifiers map[string]bool) {
	if importDecl.ImportClause == nil {
		return
	}

	// Cast ImportClause to access its properties
	importClause := importDecl.ImportClause.AsImportClause()

	// Handle default import
	if importClause.Name() != nil {
		defaultImport := importClause.Name()
		identifierName := defaultImport.AsIdentifier().Text

		// Check if this identifier is referenced at all or shadowed by local declarations
		if !allReferencedIdentifiers[identifierName] || shadowedIdentifiers[identifierName] {
			// Not referenced anywhere or shadowed by local declarations - it's unused
			*unusedSpecifiers = append(*unusedSpecifiers, defaultImport)
		} else if valueUsedIdentifiers[identifierName] {
			// Used as a value (in new expression, call expression, etc.)
			*valueSpecifiers = append(*valueSpecifiers, defaultImport)
		} else {
			// Referenced but not as a value - check if all references are shadowed by type parameters
			if areAllReferencesTypeParameterShadowed(identifierName, allReferencedNodes) {
				// All references are shadowed by type parameters - treat as unused
				*unusedSpecifiers = append(*unusedSpecifiers, defaultImport)
			} else {
				// Referenced but not as a value - it's only used as a type
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
				for _, element := range namedImports.Elements.Nodes {
					importSpecifier := element.AsImportSpecifier()

					if importSpecifier.IsTypeOnly {
						*inlineTypeSpecifiers = append(*inlineTypeSpecifiers, element)
						continue
					}

					identifierName := importSpecifier.Name().AsIdentifier().Text

					// Check if this identifier is referenced at all or shadowed by local declarations
					if !allReferencedIdentifiers[identifierName] || shadowedIdentifiers[identifierName] {
						// Not referenced anywhere or shadowed by local declarations - it's unused
						*unusedSpecifiers = append(*unusedSpecifiers, element)
					} else if valueUsedIdentifiers[identifierName] {
						// Used as a value
						*valueSpecifiers = append(*valueSpecifiers, element)
					} else {
						// Referenced but not as a value - check if all references are shadowed by type parameters
					if areAllReferencesTypeParameterShadowed(identifierName, allReferencedNodes) {
						// All references are shadowed by type parameters - treat as unused
						*unusedSpecifiers = append(*unusedSpecifiers, element)
						} else {
							// Referenced but not as a value - it's only used as a type
							*typeSpecifiers = append(*typeSpecifiers, element)
						}
					}
				}
			}
		} else if ast.IsNamespaceImport(namedBindings) {
			namespaceImport := namedBindings.AsNamespaceImport()
			identifierName := namespaceImport.Name().AsIdentifier().Text
			
			// Check if this identifier is referenced at all or shadowed by local declarations
			if !allReferencedIdentifiers[identifierName] || shadowedIdentifiers[identifierName] {
				// Not referenced anywhere or shadowed by local declarations - it's unused
				*unusedSpecifiers = append(*unusedSpecifiers, namedBindings)
			} else if valueUsedIdentifiers[identifierName] {
				// Used as a value
				*valueSpecifiers = append(*valueSpecifiers, namedBindings)
			} else {
				// Referenced but not as a value - check if all references are shadowed by type parameters
				if areAllReferencesTypeParameterShadowed(identifierName, allReferencedNodes) {
					// All references are shadowed by type parameters - treat as unused
					*unusedSpecifiers = append(*unusedSpecifiers, namedBindings)
				} else {
					// Referenced but not as a value - it's only used as a type
					*typeSpecifiers = append(*typeSpecifiers, namedBindings)
				}
			}
		}
	}
}


// areAllReferencesTypeParameterShadowed checks if all references to an identifier
// are shadowed by type parameters in their respective scopes
func areAllReferencesTypeParameterShadowed(identifierName string, allReferencedNodes map[string][]*ast.Node) bool {
	references, exists := allReferencedNodes[identifierName]
	if !exists || len(references) == 0 {
		return false
	}

	// Limit checking to prevent performance issues
	maxChecks := 10
	checked := 0
	
	// Check if all references are shadowed by type parameters
	for _, ref := range references {
		if checked >= maxChecks {
			// Too many references, assume not all are shadowed
			return false
		}
		checked++
		
		if !isIdentifierShadowedByTypeParameter(ref, identifierName) {
			// Found a reference that is not shadowed by a type parameter
			return false
		}
	}

	// All references are shadowed by type parameters
	return true
}



func getImportSpecifierName(node *ast.Node) string {
	if ast.IsImportSpecifier(node) {
		importSpec := node.AsImportSpecifier()
		if importSpec.PropertyName != nil {
			// import { foo as bar } - return "foo as bar"
			return getIdentifierText(importSpec.PropertyName) + " as " + getIdentifierText(importSpec.Name())
		}
		return getIdentifierText(importSpec.Name())
	} else if ast.IsIdentifier(node) {
		return getIdentifierText(node)
	} else if ast.IsNamespaceImport(node) {
		// import * as Foo
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

func fixToTypeImportDeclaration(sourceFile *ast.SourceFile, report ReportValueImport, sourceImports *SourceImports, fixStyle string) []rule.RuleFix {
	var fixes []rule.RuleFix

	importDecl := report.Node.AsImportDeclaration()
	sourceText := string(sourceFile.Text())

	// Check if this is a simple case where we can just add "type"
	if len(report.ValueSpecifiers) == 0 && len(report.UnusedSpecifiers) == 0 {
		// All imports are type-only, convert entire import to type import
		nodeStart := int(report.Node.Pos())
		nodeEnd := int(report.Node.End())
		nodeText := sourceText[nodeStart:nodeEnd]

		importPattern := regexp.MustCompile(`import\s+`)
		match := importPattern.FindStringIndex(nodeText)
		if match != nil {
			importEnd := nodeStart + match[1]
			fixes = append(fixes, rule.RuleFixReplaceRange(
				core.NewTextRange(importEnd, importEnd),
				"type ",
			))
		}

		// Remove inline type specifiers
		if importDecl.ImportClause != nil {
			importClause := importDecl.ImportClause.AsImportClause()
			if importClause.NamedBindings != nil && ast.IsNamedImports(importClause.NamedBindings) {
				namedImports := importClause.NamedBindings.AsNamedImports()
				if namedImports.Elements != nil {
					for _, element := range namedImports.Elements.Nodes {
						importSpecifier := element.AsImportSpecifier()
						if importSpecifier.IsTypeOnly {
							fixes = append(fixes, fixRemoveTypeSpecifierFromImportSpecifier(sourceFile, element)...)
						}
					}
				}
			}
		}

		return fixes
	}

	// Mixed imports - need to separate or inline type imports
	if fixStyle == "inline-type-imports" {
		// Add type keywords to type specifiers  
		for _, specifier := range report.TypeSpecifiers {
			if ast.IsImportSpecifier(specifier) {
				fixes = append(fixes, rule.RuleFixReplaceRange(
					core.NewTextRange(int(specifier.Pos()), int(specifier.Pos())),
					"type ",
				))
			}
		}
	} else {
		// Separate type imports - handle different import types
		if importDecl.ImportClause != nil {
			importClause := importDecl.ImportClause.AsImportClause()
			
			// Categorize the import specifiers
			var defaultImport *ast.Node
			var namespaceImport *ast.Node
			var namedImports []*ast.Node
			
			if importClause.Name() != nil {
				defaultImport = importClause.Name()
			}
			
			if importClause.NamedBindings != nil {
				if ast.IsNamespaceImport(importClause.NamedBindings) {
					namespaceImport = importClause.NamedBindings
				} else if ast.IsNamedImports(importClause.NamedBindings) {
					namedImportsNode := importClause.NamedBindings.AsNamedImports()
					if namedImportsNode.Elements != nil {
						namedImports = namedImportsNode.Elements.Nodes
					}
				}
			}
			
			// Check which parts are type-only
			defaultIsType := defaultImport != nil && isInSpecifierList(defaultImport, report.TypeSpecifiers)
			namespaceIsType := namespaceImport != nil && isInSpecifierList(namespaceImport, report.TypeSpecifiers)
			
			// Collect type and value named imports
			var typeNamedImports []string
			var valueNamedImports []string
			
			for _, namedImport := range namedImports {
				if isInSpecifierList(namedImport, report.TypeSpecifiers) {
					typeNamedImports = append(typeNamedImports, getImportSpecifierText(sourceFile, namedImport))
				} else if isInSpecifierList(namedImport, report.ValueSpecifiers) {
					valueNamedImports = append(valueNamedImports, getImportSpecifierText(sourceFile, namedImport))
				}
			}
			
			// Generate new import statements
			moduleSpecifier := sourceText[int(importDecl.ModuleSpecifier.Pos()):int(importDecl.ModuleSpecifier.End())]
			var newImports []string
			
			// Add type imports
			if len(typeNamedImports) > 0 {
				if sourceImports.TypeOnlyNamedImport != nil {
					// Merge with existing type import
					fixes = append(fixes, mergeIntoExistingTypeImport(sourceFile, sourceImports.TypeOnlyNamedImport, typeNamedImports)...)
				} else {
					newImports = append(newImports, fmt.Sprintf("import type { %s } from %s;", strings.Join(typeNamedImports, ", "), moduleSpecifier))
				}
			}
			
			if defaultIsType {
				defaultText := sourceText[int(defaultImport.Pos()):int(defaultImport.End())]
				newImports = append(newImports, fmt.Sprintf("import type %s from %s;", defaultText, moduleSpecifier))
			}
			
			if namespaceIsType {
				namespaceText := sourceText[int(namespaceImport.Pos()):int(namespaceImport.End())]
				newImports = append(newImports, fmt.Sprintf("import type %s from %s;", namespaceText, moduleSpecifier))
			}
			
			// Add new imports before the current import
			if len(newImports) > 0 {
				newImportText := strings.Join(newImports, "\n") + "\n"
				fixes = append(fixes, rule.RuleFixReplaceRange(
					core.NewTextRange(int(report.Node.Pos()), int(report.Node.Pos())),
					newImportText,
				))
			}
			
			// Generate the remaining value import
			if len(report.ValueSpecifiers) > 0 || len(report.UnusedSpecifiers) > 0 {
				// Need to reconstruct the import with only value specifiers
				var remainingParts []string
				
				if defaultImport != nil && !defaultIsType {
					remainingParts = append(remainingParts, sourceText[int(defaultImport.Pos()):int(defaultImport.End())])
				}
				
				if namespaceImport != nil && !namespaceIsType {
					namespaceText := sourceText[int(namespaceImport.Pos()):int(namespaceImport.End())]
					if len(remainingParts) > 0 {
						remainingParts[0] = remainingParts[0] + ", " + namespaceText
					} else {
						remainingParts = append(remainingParts, namespaceText)
					}
				}
				
				if len(valueNamedImports) > 0 {
					namedPart := "{ " + strings.Join(valueNamedImports, ", ") + " }"
					if len(remainingParts) > 0 {
						remainingParts[0] = remainingParts[0] + ", " + namedPart
					} else {
						remainingParts = append(remainingParts, namedPart)
					}
				}
				
				if len(remainingParts) > 0 {
					newImport := fmt.Sprintf("import %s from %s;", strings.Join(remainingParts, ""), moduleSpecifier)
					fixes = append(fixes, rule.RuleFixReplaceRange(
						core.NewTextRange(int(report.Node.Pos()), int(report.Node.End())),
						newImport,
					))
				} else {
					// Remove the entire import
					fixes = append(fixes, rule.RuleFixReplaceRange(
						core.NewTextRange(int(report.Node.Pos()), int(report.Node.End())+1), // +1 to include newline
						"",
					))
				}
			} else {
				// All specifiers are type-only, remove the original import
				fixes = append(fixes, rule.RuleFixReplaceRange(
					core.NewTextRange(int(report.Node.Pos()), int(report.Node.End())+1), // +1 to include newline  
					"",
				))
			}
		}
	}

	return fixes
}

// Helper functions for fix generation
func isInSpecifierList(node *ast.Node, list []*ast.Node) bool {
	for _, item := range list {
		if item == node {
			return true
		}
	}
	return false
}

func getImportSpecifierText(sourceFile *ast.SourceFile, node *ast.Node) string {
	sourceText := string(sourceFile.Text())
	if ast.IsImportSpecifier(node) {
		importSpec := node.AsImportSpecifier()
		if importSpec.PropertyName != nil {
			// import { foo as bar }
			propertyText := sourceText[int(importSpec.PropertyName.Pos()):int(importSpec.PropertyName.End())]
			nameText := sourceText[int(importSpec.Name().Pos()):int(importSpec.Name().End())]
			return propertyText + " as " + nameText
		}
		// import { foo }
		return sourceText[int(importSpec.Name().Pos()):int(importSpec.Name().End())]
	}
	return sourceText[int(node.Pos()):int(node.End())]
}

func mergeIntoExistingTypeImport(sourceFile *ast.SourceFile, existingImport *ast.Node, newSpecifiers []string) []rule.RuleFix {
	var fixes []rule.RuleFix
	
	// Find the closing brace of the existing import
	importDecl := existingImport.AsImportDeclaration()
	if importDecl.ImportClause != nil {
		importClause := importDecl.ImportClause.AsImportClause()
		if importClause.NamedBindings != nil && ast.IsNamedImports(importClause.NamedBindings) {
			namedImports := importClause.NamedBindings.AsNamedImports()
			// Insert before the closing brace
			closingBracePos := int(namedImports.End()) - 1
			newContent := ", " + strings.Join(newSpecifiers, ", ")
			fixes = append(fixes, rule.RuleFixReplaceRange(
				core.NewTextRange(closingBracePos, closingBracePos),
				newContent,
			))
		}
	}
	
	return fixes
}

// isIdentifierShadowedByTypeParameter checks if an identifier is shadowed by a type parameter
// in any enclosing scope (type alias, interface, class, function, etc.)
func isIdentifierShadowedByTypeParameter(node *ast.Node, identifierName string) bool {
	current := node.Parent
	// Add a safety limit to prevent infinite loops
	maxDepth := 20 // Reduced from 50
	depth := 0
	
	for current != nil && depth < maxDepth {
		depth++
		
		// Break early if we reach the source file or module
		switch current.Kind {
		case ast.KindSourceFile, ast.KindModuleBlock:
			return false
		}
		
		// Check if the current node has type parameters that might shadow the identifier
		var typeParameters *ast.TypeParameterList

		switch current.Kind {
		case ast.KindTypeAliasDeclaration:
			typeAlias := current.AsTypeAliasDeclaration()
			typeParameters = typeAlias.TypeParameters
		case ast.KindInterfaceDeclaration:
			typeParameters = current.AsInterfaceDeclaration().TypeParameters
		case ast.KindClassDeclaration, ast.KindClassExpression:
			typeParameters = current.ClassLikeData().TypeParameters
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
			typeParameters = current.FunctionLikeData().TypeParameters
		case ast.KindConstructorType, ast.KindFunctionType:
			if ast.IsConstructorTypeNode(current) {
				typeParameters = current.AsConstructorTypeNode().TypeParameters
			} else if ast.IsFunctionTypeNode(current) {
				typeParameters = current.AsFunctionTypeNode().TypeParameters
			}
		}

		if typeParameters != nil && len(typeParameters.Nodes) > 0 {
			for _, typeParam := range typeParameters.Nodes {
				if ast.IsTypeParameterDeclaration(typeParam) {
					typeParamDecl := typeParam.AsTypeParameter()
					if typeParamDecl.Name() != nil {
						paramName := typeParamDecl.Name().AsIdentifier().Text
						if paramName == identifierName {
							// This identifier is shadowed by a type parameter
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

// analyzeShadowing analyzes which imports are shadowed by local declarations
func analyzeShadowing(importDecl *ast.ImportDeclaration, localDeclarations map[string][]*ast.Node, shadowedIdentifiers map[string]bool) {
	if importDecl.ImportClause == nil {
		return
	}
	
	importClause := importDecl.ImportClause.AsImportClause()

	// Check default import
	if importClause.Name() != nil {
		defaultImportName := importClause.Name().AsIdentifier().Text
		if isImportShadowedByLocalDeclaration(importDecl, defaultImportName, localDeclarations) {
			shadowedIdentifiers[defaultImportName] = true
		}
	}

	// Check named imports
	if importClause.NamedBindings != nil {
		if ast.IsNamedImports(importClause.NamedBindings) {
			namedImports := importClause.NamedBindings.AsNamedImports()
			if namedImports.Elements != nil {
				for _, element := range namedImports.Elements.Nodes {
					importSpecifier := element.AsImportSpecifier()
					importName := importSpecifier.Name().AsIdentifier().Text
					if isImportShadowedByLocalDeclaration(importDecl, importName, localDeclarations) {
						shadowedIdentifiers[importName] = true
					}
				}
			}
		} else if ast.IsNamespaceImport(importClause.NamedBindings) {
			// Check namespace imports too
			namespaceImport := importClause.NamedBindings.AsNamespaceImport()
			importName := namespaceImport.Name().AsIdentifier().Text
			if isImportShadowedByLocalDeclaration(importDecl, importName, localDeclarations) {
				shadowedIdentifiers[importName] = true
			}
		}
	}
}

// isImportShadowedByLocalDeclaration checks if an import is shadowed by any local declaration
// This only considers actual local declarations (variables, functions, classes, etc.)
// NOT type parameters, which have more limited scope
func isImportShadowedByLocalDeclaration(importDecl *ast.ImportDeclaration, importName string, localDeclarations map[string][]*ast.Node) bool {
	declarations, exists := localDeclarations[importName]
	if !exists {
		return false
	}

	// Check if any declaration shadows this import
	for _, decl := range declarations {
		// Only consider declarations that are at the same scope level as the import
		// or in a parent scope that would affect the import's visibility
		if isDeclarationShadowingImport(importDecl, decl) {
			return true
		}
	}

	return false
}

// isDeclarationShadowingImport checks if a declaration actually shadows an import
func isDeclarationShadowingImport(importDecl *ast.ImportDeclaration, decl *ast.Node) bool {
	// Only consider declarations that are in the same module scope as the import
	// Type parameters and other scoped declarations don't shadow imports globally

	// Check if the declaration is a module-level declaration that would shadow the import
	switch decl.Kind {
	case ast.KindVariableStatement, ast.KindFunctionDeclaration, ast.KindClassDeclaration,
		 ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration,
		 ast.KindModuleDeclaration:
		// These are module-level declarations that can shadow imports
		// We're simplifying by assuming all such declarations are at module scope
		// which is typically true for these node types
		return true
	default:
		// Other kinds of declarations (like type parameters) don't shadow imports
		return false
	}
}



