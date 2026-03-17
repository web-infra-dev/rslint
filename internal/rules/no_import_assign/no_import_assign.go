package no_import_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// importedBinding holds information about a single imported binding.
type importedBinding struct {
	name        string
	isNamespace bool
	nameNode    *ast.Node
	symbol      *ast.Symbol
}

// isMemberWrite checks if an identifier is the object in a member-write expression.
// For namespace imports like `import * as ns`, member writes include:
//   - ns.prop = val
//   - ns.prop++
//   - ns["prop"] = val
//   - delete ns.prop
//   - Object.assign(ns, ...)
//   - Object.defineProperty(ns, ...)
func isMemberWrite(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check ns.prop = val, ns.prop++, ns["prop"] = val, etc.
	if parent.Kind == ast.KindPropertyAccessExpression {
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression == node {
			return utils.IsWriteReference(parent)
		}
	}

	if parent.Kind == ast.KindElementAccessExpression {
		elemAccess := parent.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.Expression == node {
			return utils.IsWriteReference(parent)
		}
	}

	// Check delete ns.prop or delete ns["prop"]
	if parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression {
		grandparent := parent.Parent
		if grandparent != nil && grandparent.Kind == ast.KindDeleteExpression {
			// Verify ns is the expression being accessed
			if parent.Kind == ast.KindPropertyAccessExpression {
				propAccess := parent.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Expression == node {
					return true
				}
			}
			if parent.Kind == ast.KindElementAccessExpression {
				elemAccess := parent.AsElementAccessExpression()
				if elemAccess != nil && elemAccess.Expression == node {
					return true
				}
			}
		}
	}

	// Check spread into destructuring assignment target: ({...ns} = obj)
	if parent.Kind == ast.KindSpreadAssignment {
		return utils.IsInDestructuringAssignment(parent)
	}

	// Check for...in/of: for (ns.prop in ...) or for (ns.prop of ...)
	// These are caught by the PropertyAccessExpression + isWriteReference path above.

	return false
}

// isImportBindingName checks if the identifier is a declaration name within an import.
func isImportBindingName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportSpecifier:
		return true
	}
	return false
}

// NoImportAssignRule disallows assigning to imported bindings.
var NoImportAssignRule = rule.Rule{
	Name: "no-import-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				if importDecl == nil || importDecl.ImportClause == nil {
					return
				}

				importClause := importDecl.ImportClause.AsImportClause()
				if importClause == nil {
					return
				}

				var bindings []importedBinding

				// Default import: import foo from 'mod'
				if importClause.Name() != nil {
					nameNode := importClause.Name()
					name := nameNode.Text()
					var sym *ast.Symbol
					if ctx.TypeChecker != nil {
						sym = ctx.TypeChecker.GetSymbolAtLocation(nameNode)
					}
					bindings = append(bindings, importedBinding{
						name:        name,
						isNamespace: false,
						nameNode:    nameNode,
						symbol:      sym,
					})
				}

				// Named or namespace bindings
				if importClause.NamedBindings != nil {
					nb := importClause.NamedBindings

					switch nb.Kind {
					case ast.KindNamespaceImport:
						nsImport := nb.AsNamespaceImport()
						if nsImport != nil && nsImport.Name() != nil {
							nameNode := nsImport.Name()
							name := nameNode.Text()
							var sym *ast.Symbol
							if ctx.TypeChecker != nil {
								sym = ctx.TypeChecker.GetSymbolAtLocation(nameNode)
							}
							bindings = append(bindings, importedBinding{
								name:        name,
								isNamespace: true,
								nameNode:    nameNode,
								symbol:      sym,
							})
						}
					case ast.KindNamedImports:
						namedImports := nb.AsNamedImports()
						if namedImports != nil && namedImports.Elements != nil {
							for _, elem := range namedImports.Elements.Nodes {
								importSpec := elem.AsImportSpecifier()
								if importSpec != nil && importSpec.Name() != nil {
									nameNode := importSpec.Name()
									name := nameNode.Text()
									var sym *ast.Symbol
									if ctx.TypeChecker != nil {
										sym = ctx.TypeChecker.GetSymbolAtLocation(nameNode)
									}
									bindings = append(bindings, importedBinding{
										name:        name,
										isNamespace: false,
										nameNode:    nameNode,
										symbol:      sym,
									})
								}
							}
						}
					}
				}

				if len(bindings) == 0 {
					return
				}

				// Walk the source file looking for write references to the imported bindings.
				sourceFile := ctx.SourceFile
				var walk func(*ast.Node)
				walk = func(n *ast.Node) {
					if n == nil {
						return
					}

					if n.Kind == ast.KindIdentifier {
						for _, binding := range bindings {
							if n.Text() != binding.name {
								continue
							}

							// Skip the import declaration names themselves
							if isImportBindingName(n) {
								continue
							}

							// Verify symbol identity when type checker is available
							if binding.symbol != nil && ctx.TypeChecker != nil {
								refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
								if refSym != binding.symbol {
									continue
								}
							}

							if utils.IsWriteReference(n) {
								ctx.ReportNode(n, rule.RuleMessage{
									Id:          "readonly",
									Description: "'" + binding.name + "' is read-only.",
								})
							} else if binding.isNamespace && isMemberWrite(n) {
								ctx.ReportNode(n, rule.RuleMessage{
									Id:          "readonlyMember",
									Description: "The members of '" + binding.name + "' are read-only.",
								})
							}
						}
					}

					n.ForEachChild(func(child *ast.Node) bool {
						walk(child)
						return false
					})
				}
				walk(sourceFile.AsNode())
			},
		}
	},
}
