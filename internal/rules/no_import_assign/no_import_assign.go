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

// makeImportedBinding creates an importedBinding, resolving the symbol via the type checker if available.
func makeImportedBinding(nameNode *ast.Node, isNamespace bool, ctx *rule.RuleContext) importedBinding {
	name := nameNode.Text()
	var sym *ast.Symbol
	if ctx.TypeChecker != nil {
		sym = ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	}
	return importedBinding{
		name:        name,
		isNamespace: isNamespace,
		nameNode:    nameNode,
		symbol:      sym,
	}
}

// wellKnownMutationMethods maps global object names to their mutation method names.
var wellKnownMutationMethods = map[string]map[string]bool{
	"Object": {
		"assign":           true,
		"defineProperty":   true,
		"defineProperties": true,
		"freeze":           true,
		"setPrototypeOf":   true,
	},
	"Reflect": {
		"defineProperty": true,
		"deleteProperty": true,
		"set":            true,
		"setPrototypeOf": true,
	},
}

// isArgumentOfWellKnownMutationFunction checks if a given node is the first argument
// of a well-known mutation function such as Object.assign, Object.defineProperty,
// Reflect.set, Reflect.setPrototypeOf, etc.
// It skips the check when Object/Reflect is locally shadowed (e.g. var Object).
func isArgumentOfWellKnownMutationFunction(node *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// The parent must be a CallExpression
	if parent.Kind != ast.KindCallExpression {
		return false
	}

	callExpr := parent.AsCallExpression()
	if callExpr == nil || callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
		return false
	}

	// The node must be the first argument
	if callExpr.Arguments.Nodes[0] != node {
		return false
	}

	// The callee must be a PropertyAccessExpression like Object.assign or Reflect.set
	// Unwrap parentheses: (Object?.defineProperty)(ns, ...) → Object?.defineProperty
	callee := callExpr.Expression
	for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
	}
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	propAccess := callee.AsPropertyAccessExpression()
	if propAccess == nil || propAccess.Expression == nil || propAccess.Name() == nil {
		return false
	}

	// The object must be a simple identifier (Object or Reflect)
	if propAccess.Expression.Kind != ast.KindIdentifier {
		return false
	}

	objectName := propAccess.Expression.Text()
	methodName := propAccess.Name().Text()

	methods, ok := wellKnownMutationMethods[objectName]
	if !ok {
		return false
	}
	if !methods[methodName] {
		return false
	}

	// If Object/Reflect is locally shadowed, skip the check (same as no-console pattern).
	if ctx.TypeChecker != nil {
		sym := ctx.TypeChecker.GetSymbolAtLocation(propAccess.Expression)
		if sym != nil {
			for _, decl := range sym.Declarations {
				declSF := ast.GetSourceFileOfNode(decl)
				if declSF != nil && declSF == ctx.SourceFile {
					return false
				}
			}
		}
	}

	return true
}

// isWrappedInTypeAssertion checks whether a node is wrapped in a type assertion
// (e.g. `as any`, `<any>`, or `!`) before reaching the actual write position.
// When a developer writes `(ns.prop as any) = value`, the `as any` is an intentional
// type-level escape hatch; ESLint does not flag such cases, so we skip them too.
func isWrappedInTypeAssertion(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindParenthesizedExpression:
			current = current.Parent
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
			return true
		default:
			return false
		}
	}
	return false
}

// isMemberExpressionWrite checks if a member expression (PropertyAccess or ElementAccess)
// is a write target: assignment left side, update expression operand, delete operand,
// or for-in/of initializer.
func isMemberExpressionWrite(memberExpr *ast.Node) bool {
	// Type assertion wrappers (as any, <any>, !) indicate intentional bypass — skip.
	if isWrappedInTypeAssertion(memberExpr) {
		return false
	}
	if utils.IsWriteReference(memberExpr) {
		return true
	}
	// IsWriteReference does not handle delete expressions, so check separately.
	if memberExpr.Parent != nil && memberExpr.Parent.Kind == ast.KindDeleteExpression {
		return true
	}
	return false
}

// isMemberWrite checks if an identifier is the object in a member-write expression.
// For namespace imports like `import * as ns`, member writes include:
//   - ns.prop = val
//   - ns.prop++
//   - ns["prop"] = val
//   - delete ns.prop
//   - Object.assign(ns, ...)
//   - Object.defineProperty(ns, ...)
//   - Reflect.set(ns, ...) etc.
func isMemberWrite(node *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check ns.prop = val, ns.prop++, ns["prop"] = val, delete ns.prop, etc.
	if parent.Kind == ast.KindPropertyAccessExpression {
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression == node {
			return isMemberExpressionWrite(parent)
		}
	}

	if parent.Kind == ast.KindElementAccessExpression {
		elemAccess := parent.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.Expression == node {
			return isMemberExpressionWrite(parent)
		}
	}

	// Check spread into destructuring assignment target: ({...ns} = obj)
	if parent.Kind == ast.KindSpreadAssignment {
		return utils.IsInDestructuringAssignment(parent)
	}

	// Check if the identifier is the first argument of a well-known mutation function
	// e.g., Object.assign(ns, ...), Reflect.set(ns, ...), etc.
	if isArgumentOfWellKnownMutationFunction(node, ctx) {
		return true
	}

	// Check for...in/of: for (ns.prop in ...) or for (ns.prop of ...)
	// These are caught by the PropertyAccessExpression + isMemberExpressionWrite path above
	// since IsWriteReference handles for-in/of initializers.

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
					bindings = append(bindings, makeImportedBinding(importClause.Name(), false, &ctx))
				}

				// Named or namespace bindings
				if importClause.NamedBindings != nil {
					nb := importClause.NamedBindings

					switch nb.Kind {
					case ast.KindNamespaceImport:
						nsImport := nb.AsNamespaceImport()
						if nsImport != nil && nsImport.Name() != nil {
							bindings = append(bindings, makeImportedBinding(nsImport.Name(), true, &ctx))
						}
					case ast.KindNamedImports:
						namedImports := nb.AsNamedImports()
						if namedImports != nil && namedImports.Elements != nil {
							for _, elem := range namedImports.Elements.Nodes {
								importSpec := elem.AsImportSpecifier()
								if importSpec != nil && importSpec.Name() != nil {
									bindings = append(bindings, makeImportedBinding(importSpec.Name(), false, &ctx))
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
							} else if binding.isNamespace && isMemberWrite(n, &ctx) {
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
