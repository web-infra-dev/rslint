package consistent_type_exports

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type ConsistentTypeExportsOptions struct {
	FixMixedExportsWithInlineTypeSpecifier bool `json:"fixMixedExportsWithInlineTypeSpecifier"`
}

type SourceExports struct {
	ReportValueExports  []ReportValueExport
	Source              string
	TypeOnlyNamedExport *ast.Node
	ValueOnlyNamedExport *ast.Node
}

type ReportValueExport struct {
	Node                 *ast.Node
	InlineTypeSpecifiers []*ast.Node
	TypeBasedSpecifiers  []*ast.Node
	ValueSpecifiers      []*ast.Node
}

var (
	exportPattern = regexp.MustCompile(`export`)
	typePattern   = regexp.MustCompile(`type\s+`)
)

var ConsistentTypeExportsRule = rule.Rule{
	Name: "consistent-type-exports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ConsistentTypeExportsOptions{
			FixMixedExportsWithInlineTypeSpecifier: false,
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
				if val, ok := optsMap["fixMixedExportsWithInlineTypeSpecifier"].(bool); ok {
					opts.FixMixedExportsWithInlineTypeSpecifier = val
				}
			}
		}

		sourceExportsMap := make(map[string]*SourceExports)

		isSymbolTypeBased := func(symbol *ast.Symbol) (bool, bool) {
			if symbol == nil {
				return false, false
			}

			// Check if this symbol is from an external module first
			if symbol.Declarations != nil && len(symbol.Declarations) > 0 {
				decl := symbol.Declarations[0]
				if decl != nil {
					sourceFile := ast.GetSourceFileOfNode(decl)
					if sourceFile != nil {
						fileName := sourceFile.FileName()
						// Skip external modules completely
						if strings.Contains(fileName, "node_modules") ||
						   strings.HasPrefix(fileName, "/usr/") ||
						   strings.HasPrefix(fileName, "C:\\") ||
						   !strings.HasSuffix(fileName, ".ts") && !strings.HasSuffix(fileName, ".tsx") && !strings.HasSuffix(fileName, ".js") && !strings.HasSuffix(fileName, ".jsx") {
							return false, false
						}
					}
				}
			}

			aliasedSymbol := symbol
			if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
				// Only resolve aliases for local symbols
				aliasedSymbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
			}

			if aliasedSymbol == nil || ctx.TypeChecker.IsUnknownSymbol(aliasedSymbol) {
				return false, false
			}

			isType := !utils.IsSymbolFlagSet(aliasedSymbol, ast.SymbolFlagsValue)
			return isType, true
		}

		processExports := func() {
			// Process all collected exports at the end
			for source, sourceExports := range sourceExportsMap {
				// Skip processing if no exports to report
				if sourceExports == nil || len(sourceExports.ReportValueExports) == 0 {
					continue
				}
				
				// Additional safety check to prevent processing problematic sources
				if source == "" {
					continue
				}

				for _, report := range sourceExports.ReportValueExports {
					if len(report.ValueSpecifiers) == 0 {
						// Export is all type-only; convert entire export to export type
						ctx.ReportNodeWithFixes(report.Node, rule.RuleMessage{
							Id:          "typeOverValue",
							Description: "All exports in the declaration are only used as types. Use `export type`.",
						}, fixExportInsertType(ctx.SourceFile, report.Node)...)
						continue
					}

					// We have both type and value violations
					allExportNames := make([]string, 0, len(report.TypeBasedSpecifiers))
					for _, specifier := range report.TypeBasedSpecifiers {
						if specifier == nil {
							continue
						}
						exportSpecifier := specifier.AsExportSpecifier()
						name := getExportSpecifierName(exportSpecifier)
						if name != "" {
							allExportNames = append(allExportNames, name)
						}
					}

					if len(allExportNames) == 1 {
						exportNames := allExportNames[0]
						message := fmt.Sprintf("Type export %s is not a value and should be exported using `export type`.", exportNames)

						var fixes []rule.RuleFix
						if opts.FixMixedExportsWithInlineTypeSpecifier {
							fixes = fixAddTypeSpecifierToNamedExports(ctx.SourceFile, report)
						} else {
							fixes = fixSeparateNamedExports(ctx.SourceFile, report)
						}

						ctx.ReportNodeWithFixes(report.Node, rule.RuleMessage{
							Id:          "singleExportIsType",
							Description: message,
						}, fixes...)
					} else {
						exportNames := formatWordList(allExportNames)
						message := fmt.Sprintf("Type exports %s are not values and should be exported using `export type`.", exportNames)

						var fixes []rule.RuleFix
						if opts.FixMixedExportsWithInlineTypeSpecifier {
							fixes = fixAddTypeSpecifierToNamedExports(ctx.SourceFile, report)
						} else {
							fixes = fixSeparateNamedExports(ctx.SourceFile, report)
						}

						ctx.ReportNodeWithFixes(report.Node, rule.RuleMessage{
							Id:          "multipleExportsAreTypes",
							Description: message,
						}, fixes...)
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				source := getSourceFromExport(node)

				// Skip export * from '...' and export * as name from '...' declarations
				// These require complex module resolution which can cause issues
				if exportDecl.ModuleSpecifier != nil && (exportDecl.ExportClause == nil || ast.IsNamespaceExport(exportDecl.ExportClause)) {
					// Skip export * declarations to avoid timeouts
					return
				}

				sourceExports, exists := sourceExportsMap[source]
				if !exists {
					sourceExports = &SourceExports{
						ReportValueExports:   []ReportValueExport{},
						Source:               source,
						TypeOnlyNamedExport:  nil,
						ValueOnlyNamedExport: nil,
					}
					sourceExportsMap[source] = sourceExports
				}

				if exportDecl.IsTypeOnly {
					if sourceExports.TypeOnlyNamedExport == nil {
						sourceExports.TypeOnlyNamedExport = node
					}
				} else {
					if sourceExports.ValueOnlyNamedExport == nil {
						sourceExports.ValueOnlyNamedExport = node
					}
				}

				var typeBasedSpecifiers []*ast.Node
				var inlineTypeSpecifiers []*ast.Node
				var valueSpecifiers []*ast.Node

				if !exportDecl.IsTypeOnly {
					exportClause := exportDecl.ExportClause
					if exportClause != nil && ast.IsNamedExports(exportClause) {
						namedExports := exportClause.AsNamedExports()
						elements := namedExports.Elements

						if elements != nil {
							for _, specifier := range elements.Nodes {
								exportSpecifier := specifier.AsExportSpecifier()

								if exportSpecifier.IsTypeOnly {
									inlineTypeSpecifiers = append(inlineTypeSpecifiers, specifier)
									continue
								}

								propertyName := exportSpecifier.PropertyName
								var identifierNode *ast.Node
								if propertyName != nil {
									identifierNode = propertyName
								} else {
									identifierNode = exportSpecifier.Name()
								}

								symbol := ctx.TypeChecker.GetSymbolAtLocation(identifierNode)
								isType, hasSymbol := isSymbolTypeBased(symbol)

								if hasSymbol && isType {
									typeBasedSpecifiers = append(typeBasedSpecifiers, specifier)
								} else if hasSymbol {
									valueSpecifiers = append(valueSpecifiers, specifier)
								}
							}
						}
					}
				}

				if (exportDecl.IsTypeOnly && len(valueSpecifiers) > 0) ||
					(!exportDecl.IsTypeOnly && len(typeBasedSpecifiers) > 0) {
					sourceExports.ReportValueExports = append(sourceExports.ReportValueExports, ReportValueExport{
						Node:                 node,
						InlineTypeSpecifiers: inlineTypeSpecifiers,
						TypeBasedSpecifiers:  typeBasedSpecifiers,
						ValueSpecifiers:      valueSpecifiers,
					})
				}
			},

			ast.KindEndOfFile: func(node *ast.Node) {
				processExports()
			},
		}
	},
}

func getSourceFromExport(node *ast.Node) string {
	exportDecl := node.AsExportDeclaration()
	moduleSpecifier := exportDecl.ModuleSpecifier
	if moduleSpecifier != nil && ast.IsStringLiteral(moduleSpecifier) {
		return moduleSpecifier.AsStringLiteral().Text
	}
	return "undefined"
}

func getExportSpecifierName(specifier *ast.ExportSpecifier) string {
	if specifier == nil {
		return ""
	}
	
	// In TypeScript AST:
	// - Name returns the exported name (what shows up after 'as' or the identifier if no 'as')
	// - PropertyName returns the local name (what appears before 'as', if present)

	exported := specifier.Name()
	local := specifier.PropertyName

	// If no propertyName, then local and exported are the same
	if local == nil {
		return getIdentifierName(exported)
	}

	exportedName := getIdentifierName(exported)
	localName := getIdentifierName(local)
	
	// Handle empty names
	if localName == "" || exportedName == "" {
		return getIdentifierName(exported)
	}

	return fmt.Sprintf("%s as %s", localName, exportedName)
}

func getIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if ast.IsIdentifier(node) {
		return node.AsIdentifier().Text
	} else if ast.IsStringLiteral(node) {
		return node.AsStringLiteral().Text
	}
	// Default to empty string if node type is unexpected
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

	// Create a copy to avoid modifying the original slice
	wordsCopy := make([]string, len(words))
	copy(wordsCopy, words)
	sort.Strings(wordsCopy)
	return strings.Join(wordsCopy[:len(wordsCopy)-1], ", ") + ", and " + wordsCopy[len(wordsCopy)-1]
}

// fixExportAllInsertType is commented out until export * handling is re-enabled
// func fixExportAllInsertType(sourceFile *ast.SourceFile, node *ast.Node) rule.RuleFix {
// 	// Find the asterisk token
// 	sourceText := string(sourceFile.Text())
// 	nodeStart := int(node.Pos())
// 	nodeEnd := int(node.End())
// 	nodeText := sourceText[nodeStart:nodeEnd]
//
// 	// Find the position of the asterisk
// 	match := asteriskPattern.FindStringIndex(nodeText)
// 	if match != nil {
// 		asteriskPos := nodeStart + match[0]
// 		// Insert "type " before the asterisk
// 		return rule.RuleFixReplaceRange(core.NewTextRange(asteriskPos, asteriskPos), "type ")
// 	}
//
// 	return rule.RuleFixReplaceRange(core.NewTextRange(int(node.Pos()), int(node.Pos())), "")
// }

func fixExportInsertType(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	var fixes []rule.RuleFix

	// Insert "type" after "export"
	sourceText := string(sourceFile.Text())
	nodeStart := int(node.Pos())
	nodeEnd := int(node.End())
	nodeText := sourceText[nodeStart:nodeEnd]

	match := exportPattern.FindStringIndex(nodeText)
	if match != nil {
		exportEndPos := nodeStart + match[1]
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(exportEndPos, exportEndPos),
			" type",
		))
	}

	// Remove inline "type" specifiers
	exportDecl := node.AsExportDeclaration()
	exportClause := exportDecl.ExportClause
	if exportClause != nil && ast.IsNamedExports(exportClause) {
		namedExports := exportClause.AsNamedExports()
		elements := namedExports.Elements

		if elements != nil {
			for _, specifier := range elements.Nodes {
				exportSpecifier := specifier.AsExportSpecifier()

				if exportSpecifier.IsTypeOnly {
					// Remove "type" keyword from specifier
					specifierStart := int(specifier.Pos())
					specifierText := sourceText[specifierStart:int(specifier.End())]

					typeMatch := typePattern.FindStringIndex(specifierText)
					if typeMatch != nil {
						typeStart := specifierStart + typeMatch[0]
						typeEnd := specifierStart + typeMatch[1]
						fixes = append(fixes, rule.RuleFixReplaceRange(
							core.NewTextRange(typeStart, typeEnd),
							"",
						))
					}
				}
			}
		}
	}

	return fixes
}

func fixSeparateNamedExports(sourceFile *ast.SourceFile, report ReportValueExport) []rule.RuleFix {
	var fixes []rule.RuleFix

	typeSpecifiers := append(report.TypeBasedSpecifiers, report.InlineTypeSpecifiers...)
	source := getSourceFromExport(report.Node)

	// Build type specifier names
	typeSpecifierNames := make([]string, 0, len(typeSpecifiers))
	for _, specifier := range typeSpecifiers {
		if specifier == nil {
			continue
		}
		exportSpecifier := specifier.AsExportSpecifier()
		name := getExportSpecifierName(exportSpecifier)
		if name != "" {
			typeSpecifierNames = append(typeSpecifierNames, name)
		}
	}

	// Build value specifier names
	valueSpecifierNames := make([]string, 0, len(report.ValueSpecifiers))
	for _, specifier := range report.ValueSpecifiers {
		if specifier == nil {
			continue
		}
		exportSpecifier := specifier.AsExportSpecifier()
		name := getExportSpecifierName(exportSpecifier)
		if name != "" {
			valueSpecifierNames = append(valueSpecifierNames, name)
		}
	}

	// Find braces and replace content
	exportDecl := report.Node.AsExportDeclaration()
	exportClause := exportDecl.ExportClause
	if exportClause != nil && ast.IsNamedExports(exportClause) {
		namedExports := exportClause.AsNamedExports()

		// Replace the content between braces with value specifiers only
		openBrace := int(namedExports.Pos()) + 1  // After '{'
		closeBrace := int(namedExports.End()) - 1   // Before '}'

		valueContent := " " + strings.Join(valueSpecifierNames, ", ") + " "
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(openBrace, closeBrace),
			valueContent,
		))
	}

	// Insert new type export line above
	typeSpecifierText := strings.Join(typeSpecifierNames, ", ")
	var newExportLine string
	if source != "undefined" && source != "" {
		newExportLine = fmt.Sprintf("export type { %s } from '%s';\n", typeSpecifierText, source)
	} else {
		newExportLine = fmt.Sprintf("export type { %s };\n", typeSpecifierText)
	}

	fixes = append(fixes, rule.RuleFixReplaceRange(
		core.NewTextRange(int(report.Node.Pos()), int(report.Node.Pos())),
		newExportLine,
	))

	return fixes
}

func fixAddTypeSpecifierToNamedExports(sourceFile *ast.SourceFile, report ReportValueExport) []rule.RuleFix {
	var fixes []rule.RuleFix

	if report.Node.AsExportDeclaration().IsTypeOnly {
		return fixes
	}

	for _, specifier := range report.TypeBasedSpecifiers {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(int(specifier.Pos()), int(specifier.Pos())),
			"type ",
		))
	}

	return fixes
}
