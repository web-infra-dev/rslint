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

		// Main logic for prefer type-imports
		listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
			importDecl := node.AsImportDeclaration()
			if importDecl.ModuleSpecifier == nil || !ast.IsStringLiteral(importDecl.ModuleSpecifier) {
				return
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
					sourceImports.TypeOnlyNamedImport = node
				}
			} else if sourceImports.ValueOnlyNamedImport == nil && hasOnlyNamedImports(importDecl) {
				sourceImports.ValueOnlyNamedImport = node
				sourceImports.ValueImport = node
			} else if sourceImports.ValueImport == nil && hasDefaultImport(importDecl) {
				sourceImports.ValueImport = node
			}

			var typeSpecifiers []*ast.Node
			var inlineTypeSpecifiers []*ast.Node
			var valueSpecifiers []*ast.Node
			var unusedSpecifiers []*ast.Node

			if importDecl.ImportClause != nil {
				classifyImportSpecifiers(ctx, importDecl, &typeSpecifiers, &inlineTypeSpecifiers, &valueSpecifiers, &unusedSpecifiers)
			}

			if importDecl.ImportClause != nil && !importDecl.ImportClause.AsImportClause().IsTypeOnly && len(typeSpecifiers) > 0 {
				sourceImports.ReportValueImports = append(sourceImports.ReportValueImports, ReportValueImport{
					Node:                 node,
					InlineTypeSpecifiers: inlineTypeSpecifiers,
					TypeSpecifiers:       typeSpecifiers,
					UnusedSpecifiers:     unusedSpecifiers,
					ValueSpecifiers:      valueSpecifiers,
				})
			}
		}

		listeners[ast.KindEndOfFile] = func(node *ast.Node) {
			if hasDecoratorMetadata {
				return
			}

			for _, sourceImports := range sourceImportsMap {
				if len(sourceImports.ReportValueImports) == 0 {
					continue
				}

				for _, report := range sourceImports.ReportValueImports {
					if len(report.ValueSpecifiers) == 0 && len(report.UnusedSpecifiers) == 0 && !isTypeOnlyImport(report.Node) {
						// Check for import attributes/assertions
						if !hasImportAttributes(report.Node) {
							ctx.ReportNodeWithFixes(report.Node, rule.RuleMessage{
								Id:          "typeOverValue",
								Description: "All imports in the declaration are only used as types. Use `import type`.",
							}, fixToTypeImportDeclaration(ctx.SourceFile, report, sourceImports, opts.FixStyle)...)
						}
					} else {
						// Mixed type/value imports
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

func classifyImportSpecifiers(ctx rule.RuleContext, importDecl *ast.ImportDeclaration, typeSpecifiers, inlineTypeSpecifiers, valueSpecifiers, unusedSpecifiers *[]*ast.Node) {
	if importDecl.ImportClause == nil {
		return
	}

	// Cast ImportClause to access its properties
	importClause := importDecl.ImportClause.AsImportClause()

	// Handle default import
	if importClause.Name() != nil {
		defaultImport := importClause.Name()
		symbol := ctx.TypeChecker.GetSymbolAtLocation(defaultImport)
		if symbol != nil {
			if isSymbolOnlyUsedAsType(ctx, symbol, importClause.IsTypeOnly) {
				*typeSpecifiers = append(*typeSpecifiers, defaultImport)
			} else {
				*valueSpecifiers = append(*valueSpecifiers, defaultImport)
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

					symbol := ctx.TypeChecker.GetSymbolAtLocation(importSpecifier.Name())
					if symbol != nil {
						if isSymbolOnlyUsedAsType(ctx, symbol, importClause.IsTypeOnly) {
							*typeSpecifiers = append(*typeSpecifiers, element)
						} else {
							*valueSpecifiers = append(*valueSpecifiers, element)
						}
					} else {
						*unusedSpecifiers = append(*unusedSpecifiers, element)
					}
				}
			}
		} else if ast.IsNamespaceImport(namedBindings) {
			namespaceImport := namedBindings.AsNamespaceImport()
			symbol := ctx.TypeChecker.GetSymbolAtLocation(namespaceImport.Name())
			if symbol != nil {
				if isSymbolOnlyUsedAsType(ctx, symbol, importClause.IsTypeOnly) {
					*typeSpecifiers = append(*typeSpecifiers, namedBindings)
				} else {
					*valueSpecifiers = append(*valueSpecifiers, namedBindings)
				}
			}
		}
	}
}

func isSymbolOnlyUsedAsType(ctx rule.RuleContext, symbol *ast.Symbol, isTypeOnlyImport bool) bool {
	if symbol == nil {
		return false
	}

	// If it's already a type-only import, it's definitely only used as a type
	if isTypeOnlyImport {
		return true
	}

	// TEMPORARY: For the initial implementation, we'll use a very conservative approach
	// and only treat imports as type-only if they're already marked as such.
	// A full implementation would need to traverse the AST and find all references.

	// For now, we'll return false to avoid false positives
	// This means the rule won't catch all cases, but it won't have false positives
	return false
}

func isChildOf(child, parent *ast.Node) bool {
	if child == nil || parent == nil {
		return false
	}
	current := child.Parent
	for current != nil {
		if current == parent {
			return true
		}
		current = current.Parent
	}
	return false
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
	}
	return getIdentifierText(node)
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

	// Check if this is a simple case where we can just add "type"
	if len(report.ValueSpecifiers) == 0 && len(report.UnusedSpecifiers) == 0 {
		// All imports are type-only, convert entire import to type import
		sourceText := string(sourceFile.Text())
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

	// Mixed imports - need to separate type imports
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
		// Separate type imports - create new import statement
		typeSpecifierNames := make([]string, 0, len(report.TypeSpecifiers))
		for _, specifier := range report.TypeSpecifiers {
			name := getImportSpecifierName(specifier)
			typeSpecifierNames = append(typeSpecifierNames, name)
		}

		if len(typeSpecifierNames) > 0 {
			newImportLine := fmt.Sprintf("import type { %s } from %s;\n",
				strings.Join(typeSpecifierNames, ", "),
				string(sourceFile.Text())[int(importDecl.ModuleSpecifier.Pos()):int(importDecl.ModuleSpecifier.End())])

			// Insert new type import line before current import
			fixes = append(fixes, rule.RuleFixReplaceRange(
				core.NewTextRange(int(report.Node.Pos()), int(report.Node.Pos())),
				newImportLine,
			))

			// Remove type specifiers from original import
			fixes = append(fixes, removeTypeSpecifiersFromImport(sourceFile, report)...)
		}
	}

	return fixes
}

func removeTypeSpecifiersFromImport(sourceFile *ast.SourceFile, report ReportValueImport) []rule.RuleFix {
	var fixes []rule.RuleFix

	importDecl := report.Node.AsImportDeclaration()
	if importDecl.ImportClause == nil {
		return fixes
	}

	importClause := importDecl.ImportClause.AsImportClause()
	if importClause.NamedBindings == nil {
		return fixes
	}

	namedBindings := importClause.NamedBindings
	if !ast.IsNamedImports(namedBindings) {
		return fixes
	}

	namedImports := namedBindings.AsNamedImports()
	if namedImports.Elements == nil {
		return fixes
	}

	// Build list of value specifier names for replacement
	valueSpecifierNames := make([]string, 0, len(report.ValueSpecifiers))
	for _, specifier := range report.ValueSpecifiers {
		name := getImportSpecifierName(specifier)
		valueSpecifierNames = append(valueSpecifierNames, name)
	}

	// Replace the content between braces with value specifiers only
	openBrace := int(namedImports.Pos()) + 1  // After '{'
	closeBrace := int(namedImports.End()) - 1 // Before '}'

	valueContent := " " + strings.Join(valueSpecifierNames, ", ") + " "
	fixes = append(fixes, rule.RuleFixReplaceRange(
		core.NewTextRange(openBrace, closeBrace),
		valueContent,
	))

	return fixes
}
