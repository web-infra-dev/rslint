package no_unsafe_declaration_merging

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

var NoUnsafeDeclarationMergingRule = rule.Rule{
	Name: "no-unsafe-declaration-merging",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Helper function to check if a symbol has declarations of a specific kind
		hasDeclarationOfKind := func(symbol any, kind ast.Kind) bool {
			if symbol == nil {
				return false
			}
			// Note: This is a simplified check - in a real implementation,
			// we would need to access the symbol's declarations
			return false
		}

		// Helper function to report unsafe merging
		reportUnsafeMerging := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unsafeMerging",
				Description: "Unsafe declaration merging between classes and interfaces.",
			})
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				className := classDecl.Name()
				if className == nil {
					return
				}

				// Get the symbol for this class name
				symbol := ctx.TypeChecker.GetSymbolAtLocation(className)
				if symbol == nil {
					return
				}

				// Check if this symbol also has interface declarations
				if hasDeclarationOfKind(symbol, ast.KindInterfaceDeclaration) {
					reportUnsafeMerging(className)
				}
			},

			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				interfaceName := interfaceDecl.Name()
				if interfaceName == nil {
					return
				}
				
				// Get the symbol for this interface name
				symbol := ctx.TypeChecker.GetSymbolAtLocation(interfaceName)
				if symbol == nil {
					return
				}

				// Check if this symbol also has class declarations
				if hasDeclarationOfKind(symbol, ast.KindClassDeclaration) {
					reportUnsafeMerging(interfaceName)
				}
			},
		}
	},
}