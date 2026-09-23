package consistent_type_exports

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_type_exports.schema.json
var schemaJSON []byte

type ConsistentTypeExportsOptions struct {
	FixMixedExportsWithInlineTypeSpecifier bool `json:"fixMixedExportsWithInlineTypeSpecifier"`
}

// ConsistentTypeExportsRule enforces consistent type exports.
var ConsistentTypeExportsRule = rule.CreateRule(rule.Rule{
	Name:             "consistent-type-exports",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run:              run,
})

func parseOptions(options []any) ConsistentTypeExportsOptions {
	opts := ConsistentTypeExportsOptions{}
	if len(options) != 0 {
		optMap, _ := options[0].(map[string]any)
		opts.FixMixedExportsWithInlineTypeSpecifier, _ = optMap["fixMixedExportsWithInlineTypeSpecifier"].(bool)
	}
	return opts
}

func typeOverValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeOverValue",
		Description: "All exports in the declaration are only used as types. Use `export type`.",
	}
}

// Resolve one alias at a time: an intermediate type-only import or export can
// hide a value even when the final symbol is a class or another runtime value.
func isSymbolTypeBased(typeChecker *checker.Checker, symbol *ast.Symbol) (isType, known bool) {
	var seen collections.Set[*ast.Symbol]
	for symbol != nil && !typeChecker.IsUnknownSymbol(symbol) {
		for _, declaration := range symbol.Declarations {
			if ast.IsTypeOnlyImportOrExportDeclaration(declaration) {
				return true, true
			}
		}
		if symbol.Flags&ast.SymbolFlagsValue != 0 {
			return false, true
		}
		if symbol.Flags&ast.SymbolFlagsAlias == 0 {
			return true, true
		}
		// Immediate aliases can form cycles. Resolving the final symbol first
		// would lose type-only imports whose module cannot be resolved.
		next := typeChecker.GetImmediateAliasedSymbol(symbol)
		// A terminal target cannot cycle; avoid allocating a set for the common
		// single-hop export of a local type or value.
		if next != nil && next.Flags&ast.SymbolFlagsAlias != 0 && !seen.AddIfAbsent(symbol) {
			return false, false
		}
		symbol = next
	}
	return false, false
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)
	return rule.RuleListeners{
		ast.KindExportDeclaration: func(node *ast.Node) {
			declaration := node.AsExportDeclaration()
			if declaration.IsTypeOnly {
				return
			}
			if declaration.ExportClause == nil || declaration.ExportClause.Kind == ast.KindNamespaceExport {
				if declaration.ModuleSpecifier == nil {
					return
				}
				// Use the Program's resolved source, not a raw exports table: star
				// exports and type-only star re-exports are resolved by the checker.
				// Upstream resolves without a usage mode, including in NodeNext
				// files; the import/require-specific cached resolution can differ.
				resolved := ctx.Program().ResolveModuleName(declaration.ModuleSpecifier.Text(), ctx.SourceFile.FileName(), core.ResolutionModeNone)
				if resolved == nil {
					return
				}
				source := ctx.Program().GetSourceFileForResolvedModule(resolved.ResolvedFileName)
				if source == nil {
					return
				}
				symbol := ctx.TypeChecker.GetSymbolAtLocation(source.AsNode())
				if symbol == nil {
					return
				}
				moduleType := ctx.TypeChecker.GetTypeOfSymbol(symbol)
				for _, property := range ctx.TypeChecker.GetPropertiesOfType(moduleType) {
					// GetPropertyOfType excludes values reached through export type *.
					if ctx.TypeChecker.GetPropertyOfType(moduleType, property.Name) != nil {
						return
					}
				}
				ctx.ReportNodeWithDeferredFixes(node, typeOverValueMessage(), func() []rule.RuleFix {
					exportToken, _ := utils.TokenAtOrAfter(ctx.SourceFile, node.Pos())
					asterisk, _ := utils.TokenAtOrAfter(ctx.SourceFile, exportToken.End)
					return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(asterisk.Start, asterisk.Start), "type ")}
				})
				return
			}

			var typeSpecifiers []*ast.Node
			hasValue := false
			for _, specifier := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
				if specifier.IsTypeOnly() {
					continue
				}
				isType, known := isSymbolTypeBased(ctx.TypeChecker, ctx.TypeChecker.GetSymbolAtLocation(specifier.Name()))
				if !known {
					continue
				}
				if isType {
					typeSpecifiers = append(typeSpecifiers, specifier)
				} else {
					hasValue = true
				}
			}
			if len(typeSpecifiers) == 0 {
				return
			}
			if !hasValue {
				ctx.ReportNodeWithDeferredFixes(node, typeOverValueMessage(), func() []rule.RuleFix {
					return fixExportType(ctx.SourceFile, node)
				})
				return
			}

			names := make([]string, 0, len(typeSpecifiers))
			for _, specifier := range typeSpecifiers {
				names = append(names, specifier.PropertyNameOrName().Text())
			}
			exportNames := names[0]
			message := rule.RuleMessage{Id: "singleExportIsType"}
			if len(names) == 1 {
				message.Description = "Type export " + exportNames + " is not a value and should be exported using `export type`."
			} else {
				exportNames = strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
				message.Id = "multipleExportsAreTypes"
				message.Description = "Type exports " + exportNames + " are not values and should be exported using `export type`."
			}
			message.Data = map[string]string{"exportNames": exportNames}
			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				if opts.FixMixedExportsWithInlineTypeSpecifier {
					fixes := make([]rule.RuleFix, 0, len(typeSpecifiers))
					for _, specifier := range typeSpecifiers {
						fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, specifier, "type "))
					}
					return fixes
				}
				return fixSeparateExports(ctx, node, typeSpecifiers)
			})
		},
	}
}

func fixExportType(source *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	exportToken, _ := utils.TokenAtOrAfter(source, node.Pos())
	fixes := []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(exportToken.End, exportToken.End), " type")}
	for _, specifier := range node.AsExportDeclaration().ExportClause.AsNamedExports().Elements.Nodes {
		if specifier.IsTypeOnly() {
			typeToken, _ := utils.TokenAtOrAfter(source, specifier.Pos())
			end := scanner.SkipTriviaEx(source.Text(), typeToken.End, &scanner.SkipTriviaOptions{StopAtComments: true})
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(typeToken.Start, end)))
		}
	}
	return fixes
}

func fixSeparateExports(ctx rule.RuleContext, node *ast.Node, typeSpecifiers []*ast.Node) []rule.RuleFix {
	declaration := node.AsExportDeclaration()
	typeNames := make([]string, 0, len(typeSpecifiers))
	for _, specifier := range typeSpecifiers {
		typeNames = append(typeNames, specifierText(ctx.SourceFile, specifier))
	}
	var valueNames []string
	// Materialize the remaining specifiers only when an autofix is requested.
	// Upstream puts inferred types before existing inline type specifiers.
	for _, specifier := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
		if specifier.IsTypeOnly() {
			typeNames = append(typeNames, specifierText(ctx.SourceFile, specifier))
		} else if isType, known := isSymbolTypeBased(ctx.TypeChecker, ctx.TypeChecker.GetSymbolAtLocation(specifier.Name())); known && !isType {
			valueNames = append(valueNames, specifierText(ctx.SourceFile, specifier))
		}
	}
	exportText := "export type { " + strings.Join(typeNames, ", ") + " }"
	if declaration.ModuleSpecifier != nil && declaration.ModuleSpecifier.Text() != "" {
		source := "'" + declaration.ModuleSpecifier.Text() + "'"
		if strings.ContainsAny(declaration.ModuleSpecifier.Text(), "'\\\r\n") {
			// Reuse the parsed literal rather than interpreting its escapes again
			// or introducing an unescaped quote in the new export declaration.
			source = utils.TrimmedNodeText(ctx.SourceFile, declaration.ModuleSpecifier)
		}
		exportText += " from " + source
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(utils.BracedNodeInnerRange(ctx.SourceFile, declaration.ExportClause), " "+strings.Join(valueNames, ", ")+" "),
		rule.RuleFixInsertBefore(ctx.SourceFile, node, exportText+";\n"),
	}
}

func specifierText(source *ast.SourceFile, specifier *ast.Node) string {
	nameText := func(name *ast.Node) string {
		if ast.IsStringLiteral(name) {
			return utils.TrimmedNodeText(source, name)
		}
		return name.Text()
	}
	local := nameText(specifier.PropertyNameOrName())
	exported := nameText(specifier.Name())
	if local != exported {
		return local + " as " + exported
	}
	return local
}
