package consistent_type_exports

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentTypeExportsOptions struct {
	FixMixedExportsWithInlineTypeSpecifier bool `json:"fixMixedExportsWithInlineTypeSpecifier"`
}

// ConsistentTypeExportsRule enforces consistent type exports
var ConsistentTypeExportsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-exports",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeExportsOptions{
		FixMixedExportsWithInlineTypeSpecifier: false,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optMap, ok := optArray[0].(map[string]interface{}); ok {
				if fixMixed, ok := optMap["fixMixedExportsWithInlineTypeSpecifier"].(bool); ok {
					opts.FixMixedExportsWithInlineTypeSpecifier = fixMixed
				}
			}
		} else if optMap, ok := options.(map[string]interface{}); ok {
			if fixMixed, ok := optMap["fixMixedExportsWithInlineTypeSpecifier"].(bool); ok {
				opts.FixMixedExportsWithInlineTypeSpecifier = fixMixed
			}
		}
	}

	// Helper to check if a symbol is type-only
	// Returns: true = type-only, false = value-based, nil = unknown/unresolved
	isSymbolTypeBased := func(symbol *ast.Symbol) *bool {
		if symbol == nil {
			return nil
		}

		// Follow alias chain
		for symbol != nil && (symbol.Flags&ast.SymbolFlagsAlias) != 0 {
			symbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
			if symbol == nil {
				return nil
			}

			// Check if any declaration in the chain is type-only
			declarations := symbol.Declarations
			for _, decl := range declarations {
				// Use the Node's IsTypeOnly() method which handles all type-only checks
				if decl.IsTypeOnly() {
					trueVal := true
					return &trueVal
				}
			}
		}

		// Check if the symbol is unknown
		if symbol == nil || ctx.TypeChecker.IsUnknownSymbol(symbol) {
			return nil
		}

		// Check if symbol has Value flag - if not, it's type-only
		hasValue := (symbol.Flags & ast.SymbolFlagsValue) != 0
		isType := !hasValue
		return &isType
	}

	checkExportDeclaration := func(node *ast.Node) {
		exportDecl := node.AsExportDeclaration()
		if exportDecl == nil {
			return
		}

		// Skip if already marked as type-only
		if exportDecl.IsTypeOnly {
			return
		}

		// Handle export * from 'module'
		if exportDecl.ExportClause == nil && exportDecl.ModuleSpecifier != nil {
			// Check if the entire module exports only types
			moduleSpecifier := exportDecl.ModuleSpecifier
			moduleSymbol := ctx.TypeChecker.GetSymbolAtLocation(moduleSpecifier)

			if moduleSymbol != nil {
				// Get the exports of the module symbol
				if moduleSymbol.Exports != nil {
					hasValueExport := false
					hasAnyExport := false

					// Check each export from the module
					for _, exportSymbol := range moduleSymbol.Exports {
						hasAnyExport = true
						// Use our helper function to determine if this export is type-only
						isType := isSymbolTypeBased(exportSymbol)
						if isType != nil && !*isType {
							// This export is a value
							hasValueExport = true
							break
						}
					}

					// If all exports are type-only, report it
					if hasAnyExport && !hasValueExport {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "typeOverValue",
							Description: "All exports in the declaration are only used as types. Use `export type`.",
						})
					}
				}
			}
			return
		}

		// Handle named exports: export { x, y, z } or export { x, y, z } from 'module'
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports == nil || len(namedExports.Elements.Nodes) == 0 {
				return
			}

			var typeSpecifiers []*ast.Node
			var valueSpecifiers []*ast.Node
			var inlineTypeSpecifiers []*ast.Node

			for _, element := range namedExports.Elements.Nodes {
				exportSpecifier := element.AsExportSpecifier()
				if exportSpecifier == nil {
					continue
				}

				// Check if this specifier is already marked as type-only (inline type)
				if exportSpecifier.IsTypeOnly {
					inlineTypeSpecifiers = append(inlineTypeSpecifiers, element)
					continue
				}

				// Get the symbol being exported
				var symbol *ast.Symbol
				// For local exports, we check the property name (what's being exported)
				// For re-exports, we check the name (what's being imported from the module)
				if exportSpecifier.PropertyName != nil {
					symbol = ctx.TypeChecker.GetSymbolAtLocation(exportSpecifier.PropertyName)
				} else {
					symbol = ctx.TypeChecker.GetSymbolAtLocation(exportSpecifier.Name())
				}

				isType := isSymbolTypeBased(symbol)
				// Skip if we can't determine the type (unknown symbol)
				if isType == nil {
					continue
				}

				if *isType {
					typeSpecifiers = append(typeSpecifiers, element)
				} else {
					valueSpecifiers = append(valueSpecifiers, element)
				}
			}

			// All specifiers are type-only
			if len(typeSpecifiers) > 0 && len(valueSpecifiers) == 0 && len(inlineTypeSpecifiers) == 0 {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "typeOverValue",
					Description: "All exports in the declaration are only used as types. Use `export type`.",
				})
				return
			}

			// Mixed: some types, some values
			if len(typeSpecifiers) > 0 && len(valueSpecifiers) > 0 {
				// If fixMixedExportsWithInlineTypeSpecifier is enabled and there are already
				// inline type specifiers, don't report an error (the code is already following
				// the preferred inline style)
				if opts.FixMixedExportsWithInlineTypeSpecifier && len(inlineTypeSpecifiers) > 0 {
					return
				}

				if len(typeSpecifiers) == 1 {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "singleExportIsType",
						Description: "Type export should use `export type`.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "multipleExportsAreTypes",
						Description: "Type exports should use `export type`.",
					})
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindExportDeclaration: checkExportDeclaration,
	}
}
