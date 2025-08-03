package no_useless_empty_export

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func isEmptyExport(node *ast.Node) bool {
	if node.Kind != ast.KindExportDeclaration {
		return false
	}

	exportDecl := node.AsExportDeclaration()
	// Empty export is either:
	// 1. export {} - no export clause and no module specifier
	// 2. export {} from "module" would have module specifier
	if exportDecl.ModuleSpecifier != nil {
		return false
	}

	// Check if it's specifically an empty export {}
	// For export {}, ExportClause might be a NamedExports with zero elements
	if exportDecl.ExportClause == nil {
		// Could be export declaration with embedded declaration like:
		// export const _ = {} or export function foo() {}
		// These are NOT empty exports
		return false
	}

	// If there's an export clause, check if it's empty
	if exportDecl.ExportClause.Kind == ast.KindNamedExports {
		namedExports := exportDecl.ExportClause.AsNamedExports()
		return len(namedExports.Elements.Nodes) == 0
	}

	return false
}

func isExportStatement(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		// Type-only exports don't count
		if exportDecl.IsTypeOnly {
			return false
		}
		// Empty exports are handled separately
		if isEmptyExport(node) {
			return false
		}
		// Any other export declaration is a real export
		// This includes:
		// - export { foo }
		// - export { foo } from 'bar'
		// - export * from 'bar'
		// But apparently NOT export const/let/var/function/class
		return true
	case ast.KindExportAssignment:
		return true
	case ast.KindVariableStatement:
		// Check if variable statement has export modifier
		// Skip if it has declare modifier
		if hasDeclareModifier(node) {
			return false
		}
		varStmt := node.AsVariableStatement()
		if varStmt.Modifiers() != nil {
			for _, mod := range varStmt.Modifiers().Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		// Skip if it has declare modifier
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		// Type-only declarations don't count as runtime exports
		return false
	case ast.KindEnumDeclaration:
		// Enums are runtime values (unless they have declare modifier)
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	case ast.KindModuleDeclaration:
		// Module declarations with declare are ambient
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	}
	return false
}

func hasExportModifier(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.Modifiers() != nil {
			for _, mod := range varStmt.Modifiers().Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindFunctionDeclaration:
		if m := node.AsFunctionDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindClassDeclaration:
		if m := node.AsClassDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindInterfaceDeclaration:
		if m := node.AsInterfaceDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindTypeAliasDeclaration:
		if m := node.AsTypeAliasDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindEnumDeclaration:
		if m := node.AsEnumDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	case ast.KindModuleDeclaration:
		if m := node.AsModuleDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindExportKeyword {
					return true
				}
			}
		}
	}
	return false
}

func hasDeclareModifier(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt.Modifiers() != nil {
			for _, mod := range varStmt.Modifiers().Nodes {
				if mod.Kind == ast.KindDeclareKeyword {
					return true
				}
			}
		}
	case ast.KindFunctionDeclaration:
		if m := node.AsFunctionDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindDeclareKeyword {
					return true
				}
			}
		}
	case ast.KindClassDeclaration:
		if m := node.AsClassDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindDeclareKeyword {
					return true
				}
			}
		}
	case ast.KindEnumDeclaration:
		if m := node.AsEnumDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindDeclareKeyword {
					return true
				}
			}
		}
	case ast.KindModuleDeclaration:
		if m := node.AsModuleDeclaration().Modifiers(); m != nil {
			for _, mod := range m.Nodes {
				if mod.Kind == ast.KindDeclareKeyword {
					return true
				}
			}
		}
	}
	return false
}

func isExportOrImportStatement(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		// Check if it's a type-only export
		if exportDecl.IsTypeOnly {
			return false
		}
		// export * from 'module'
		if exportDecl.ModuleSpecifier != nil && exportDecl.ExportClause == nil {
			return true
		}
		// export * as ns from 'module'
		if exportDecl.ModuleSpecifier != nil && exportDecl.ExportClause != nil && ast.IsNamespaceExport(exportDecl.ExportClause) {
			return true
		}
		// export { x } or export { x } from 'module'
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if len(namedExports.Elements.Nodes) > 0 {
				return true
			}
		}
		return false
	case ast.KindExportAssignment:
		// This covers export = and possibly export default
		return true
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		// Skip type-only imports
		if importDecl.ImportClause != nil && importDecl.ImportClause.IsTypeOnly() {
			return false
		}
		return true
	case ast.KindImportEqualsDeclaration:
		return true
	case ast.KindVariableStatement:
		// Check for export const _ = {}
		// Skip if it has declare modifier
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		// Skip if it has declare modifier
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		// Type-only declarations don't count as runtime exports
		return false
	case ast.KindEnumDeclaration:
		// Enums are runtime values (unless they have declare modifier)
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	case ast.KindModuleDeclaration:
		// Module declarations with declare are ambient
		if hasDeclareModifier(node) {
			return false
		}
		return hasExportModifier(node)
	case ast.KindExpressionStatement:
		// ExpressionStatement by itself is not an export
		// Export default expressions would be handled by KindExportAssignment
		return false
	}
	return false
}

func isDefinitionFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".d.ts")
}

var NoUselessEmptyExportRule = rule.Rule{
	Name: "no-useless-empty-export",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// In a definition file, export {} is necessary to make the module properly
		// encapsulated, even when there are other exports
		if isDefinitionFile(ctx.SourceFile.FileName()) {
			return rule.RuleListeners{}
		}

		// First pass: collect all statements to check for exports
		var emptyExports []*ast.Node
		hasOtherExports := false

		// Check all statements upfront
		for _, statement := range ctx.SourceFile.Statements.Nodes {
			if isEmptyExport(statement) {
				emptyExports = append(emptyExports, statement)
			} else if isExportOrImportStatement(statement) {
				hasOtherExports = true
			}
		}

		// If there are other exports, report the empty exports as useless
		if hasOtherExports {
			for _, emptyExport := range emptyExports {
				ctx.ReportNodeWithFixes(emptyExport, rule.RuleMessage{
					Id:          "uselessExport",
					Description: "Empty export does nothing and can be removed.",
				}, rule.RuleFixRemove(ctx.SourceFile, emptyExport))
			}
		}

		// Return empty listeners since we already processed everything
		return rule.RuleListeners{}
	},
}
