package no_mutable_exports

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/no-mutable-exports.js
var NoMutableExportsRule = rule.Rule{
	Name: "import/no-mutable-exports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// checkDeclarationList reports if the VariableDeclarationList uses `let` or `var`.
		checkDeclarationList := func(declList *ast.Node) {
			if declList == nil || !ast.IsVariableDeclarationList(declList) {
				return
			}
			if ast.IsVarConstLike(declList) {
				return
			}
			kind := "var"
			if ast.IsVarLet(declList) {
				kind = "let"
			}
			ctx.ReportNode(declList, rule.RuleMessage{
				Id:          "noMutableExports",
				Description: fmt.Sprintf("Exporting mutable '%s' binding, use 'const' instead.", kind),
			})
		}

		// resolveVarDeclListBySymbol uses TypeChecker to resolve an identifier to
		// its VariableDeclarationList. SkipAlias is needed because export specifier
		// names resolve to alias symbols pointing at the specifier itself.
		resolveVarDeclListBySymbol := func(identNode *ast.Node) *ast.Node {
			if ctx.TypeChecker == nil {
				return nil
			}
			sym := ctx.TypeChecker.GetSymbolAtLocation(identNode)
			if sym == nil {
				return nil
			}
			resolved := ctx.TypeChecker.SkipAlias(sym)
			if resolved == nil || len(resolved.Declarations) == 0 {
				return nil
			}
			decl := resolved.Declarations[0]
			if !ast.IsVariableDeclaration(decl) {
				return nil
			}
			declList := decl.Parent
			if declList != nil && ast.IsVariableDeclarationList(declList) {
				return declList
			}
			return nil
		}

		// findVarDeclListByWalk is the fallback when TypeChecker is unavailable.
		// It walks top-level VariableStatements to find a declaration by name.
		findVarDeclListByWalk := func(name string) *ast.Node {
			if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil {
				return nil
			}
			for _, stmt := range ctx.SourceFile.Statements.Nodes {
				if !ast.IsVariableStatement(stmt) {
					continue
				}
				declList := stmt.AsVariableStatement().DeclarationList
				if declList == nil || !ast.IsVariableDeclarationList(declList) {
					continue
				}
				vdl := declList.AsVariableDeclarationList()
				if vdl.Declarations == nil {
					continue
				}
				for _, decl := range vdl.Declarations.Nodes {
					if !ast.IsVariableDeclaration(decl) {
						continue
					}
					declName := decl.AsVariableDeclaration().Name()
					if declName == nil {
						continue
					}
					if utils.HasNameInBindingPattern(declName, name) {
						return declList
					}
				}
			}
			return nil
		}

		// findVarDeclarationList resolves an identifier to its VariableDeclarationList.
		// TypeChecker when available, otherwise top-level AST walk as fallback.
		findVarDeclarationList := func(identNode *ast.Node) *ast.Node {
			if declList := resolveVarDeclListBySymbol(identNode); declList != nil {
				return declList
			}
			if identNode.Kind == ast.KindIdentifier {
				return findVarDeclListByWalk(identNode.AsIdentifier().Text)
			}
			return nil
		}

		// isModuleLevelNode returns true if the node is a direct child of the
		// SourceFile. Exports inside namespace/module blocks are TypeScript
		// namespace exports, not ES module exports, and should not be reported.
		isModuleLevelNode := func(node *ast.Node) bool {
			return node.Parent != nil && node.Parent.Kind == ast.KindSourceFile
		}

		return rule.RuleListeners{
			// Handle: export let x = 1 / export var x = 1
			ast.KindVariableStatement: func(node *ast.Node) {
				if !isModuleLevelNode(node) {
					return
				}
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
					return
				}
				checkDeclarationList(node.AsVariableStatement().DeclarationList)
			},
			// Handle: export { x } / export { x as y }
			ast.KindExportDeclaration: func(node *ast.Node) {
				if !isModuleLevelNode(node) {
					return
				}
				exportDecl := node.AsExportDeclaration()
				// Skip re-exports (export { x } from './foo')
				if exportDecl.ModuleSpecifier != nil {
					return
				}
				if exportDecl.IsTypeOnly {
					return
				}
				if exportDecl.ExportClause == nil || exportDecl.ExportClause.Kind != ast.KindNamedExports {
					return
				}
				namedExports := exportDecl.ExportClause.AsNamedExports()
				if namedExports.Elements == nil {
					return
				}
				for _, spec := range namedExports.Elements.Nodes {
					if spec.Kind != ast.KindExportSpecifier {
						continue
					}
					exportSpec := spec.AsExportSpecifier()
					if exportSpec.IsTypeOnly {
						continue
					}
					// The local name is PropertyName if present (export { local as exported }),
					// otherwise Name (export { name })
					localNode := exportSpec.PropertyName
					if localNode == nil {
						localNode = exportSpec.Name()
					}
					if localNode == nil {
						continue
					}
					if declList := findVarDeclarationList(localNode); declList != nil {
						checkDeclarationList(declList)
					}
				}
			},
			// Handle: export default x
			ast.KindExportAssignment: func(node *ast.Node) {
				if !isModuleLevelNode(node) {
					return
				}
				exportAssign := node.AsExportAssignment()
				if exportAssign.IsExportEquals {
					return
				}
				expr := ast.SkipParentheses(exportAssign.Expression)
				if expr == nil || expr.Kind != ast.KindIdentifier {
					return
				}
				if declList := findVarDeclarationList(expr); declList != nil {
					checkDeclarationList(declList)
				}
			},
		}
	},
}
