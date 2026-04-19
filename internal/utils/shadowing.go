package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// IsNameShadowedBetween walks from `node` up to (but not including) `boundary`,
// returning true if any intermediate scope introduces a binding for `name`.
//
// Scopes examined:
//   - Function-like parameter lists (covers all 7 function-like kinds).
//   - Block-scoped declarations (var/let/const/function/class inside a Block).
//   - Catch-clause variable bindings (including destructured patterns).
//   - For-statement `let`/`const` init (scoped to the loop).
//   - For-in / for-of `let`/`const` init (scoped to the loop).
//   - Class declaration/expression names (scoped to the class body).
//
// Use this when a rule tracks a specific declaration (e.g. a parameter, class,
// or function name) and needs to ignore references that were shadowed before
// they reached the declaration site. For scope walks that should also examine
// the SourceFile or module boundary, use `IsShadowed` instead.
func IsNameShadowedBetween(node *ast.Node, boundary *ast.Node, name string) bool {
	for current := node.Parent; current != nil && current != boundary; current = current.Parent {
		if ast.IsFunctionLikeDeclaration(current) {
			if HasShadowingParameter(current, name) {
				return true
			}
		}
		switch current.Kind {
		case ast.KindBlock:
			if HasShadowingDeclaration(current, name) {
				return true
			}
		case ast.KindCatchClause:
			cc := current.AsCatchClause()
			if cc != nil && cc.VariableDeclaration != nil {
				vd := cc.VariableDeclaration.AsVariableDeclaration()
				if vd != nil && vd.Name() != nil && HasNameInBindingPattern(vd.Name(), name) {
					return true
				}
			}
		case ast.KindForStatement:
			forStmt := current.AsForStatement()
			if forStmt != nil && forStmt.Initializer != nil &&
				forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(forStmt.Initializer, name) {
				return true
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := current.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer != nil &&
				stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(stmt.Initializer, name) {
				return true
			}
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
	}
	return false
}

// GetReferenceSymbol returns the variable symbol for the given identifier.
// When the identifier is the key of a shorthand property assignment in a
// destructuring pattern (e.g., `({foo} = bar)`), it returns the value-binding
// symbol rather than the property symbol that `GetSymbolAtLocation` would
// otherwise produce.
func GetReferenceSymbol(node *ast.Node, typeChecker *checker.Checker) *ast.Symbol {
	if node == nil || typeChecker == nil {
		return nil
	}
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			if symbol := typeChecker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
				return symbol
			}
		}
	}
	return typeChecker.GetSymbolAtLocation(node)
}

// IsShadowed checks whether the given identifier name is shadowed by a local
// declaration at the usage site. It walks from node up to the SourceFile,
// checking every scope boundary for variable/function/class/enum/import
// declarations, function parameters, catch variables, and hoisted var
// declarations.
func IsShadowed(node *ast.Node, name string) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile:
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				if HasLocalDeclarationInStatements(sf.Statements.Nodes, name) {
					return true
				}
			}
			if HasHoistedVarDeclaration(current, name) {
				return true
			}
			return false

		case ast.KindBlock:
			if HasShadowingDeclaration(current, name) {
				return true
			}

		case ast.KindCatchClause:
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if HasNameInBindingPattern(varDecl.Name(), name) {
						return true
					}
				}
			}

		case ast.KindClassDeclaration, ast.KindClassExpression:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}

		case ast.KindEnumDeclaration:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}

		case ast.KindForStatement:
			forStmt := current.AsForStatement()
			if forStmt != nil && forStmt.Initializer != nil &&
				forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(forStmt.Initializer, name) {
				return true
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := current.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer != nil &&
				stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				HasVarDeclListWithName(stmt.Initializer, name) {
				return true
			}

		default:
			if ast.IsFunctionLikeDeclaration(current) {
				if HasShadowingParameter(current, name) {
					return true
				}
				// Function declarations and expressions can shadow via their own name.
				if current.Kind == ast.KindFunctionDeclaration || current.Kind == ast.KindFunctionExpression {
					if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
						return true
					}
				}
				if body := current.Body(); body != nil && HasHoistedVarDeclaration(body, name) {
					return true
				}
			}
		}
		current = current.Parent
	}
	return false
}

// HasShadowingParameter checks if a function-like node has a parameter
// whose binding name matches the given name (supports destructuring patterns).
func HasShadowingParameter(node *ast.Node, name string) bool {
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		nameNode := param.Name()
		if nameNode == nil {
			continue
		}
		found := false
		CollectBindingNames(nameNode, func(_ *ast.Node, n string) {
			if n == name {
				found = true
			}
		})
		if found {
			return true
		}
	}
	return false
}

// HasShadowingDeclaration checks if a block contains a variable, function,
// class, enum, or namespace declaration whose name matches the given name.
func HasShadowingDeclaration(node *ast.Node, name string) bool {
	if node.Kind != ast.KindBlock {
		return false
	}

	block := node.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}

	for _, stmt := range block.Statements.Nodes {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindVariableStatement:
			varStmt := stmt.AsVariableStatement()
			if varStmt == nil || varStmt.DeclarationList == nil {
				continue
			}
			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			if declList == nil || declList.Declarations == nil {
				continue
			}
			for _, decl := range declList.Declarations.Nodes {
				if decl == nil || decl.Kind != ast.KindVariableDeclaration {
					continue
				}
				varDecl := decl.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					continue
				}
				found := false
				CollectBindingNames(varDecl.Name(), func(_ *ast.Node, n string) {
					if n == name {
						found = true
					}
				})
				if found {
					return true
				}
			}
		case ast.KindFunctionDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindClassDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindEnumDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindModuleDeclaration:
			// `namespace X {}` / `module X {}` introduce a value binding when
			// named by an identifier. Skip ambient modules (`declare module "x"`),
			// which use a string literal name and don't bind a variable.
			modDecl := stmt.AsModuleDeclaration()
			if modDecl != nil && modDecl.Name() != nil &&
				modDecl.Name().Kind == ast.KindIdentifier && modDecl.Name().Text() == name {
				return true
			}
		}
	}

	return false
}

// HasLocalDeclarationInStatements checks if a list of top-level statements
// (typically from a SourceFile) contains a variable, function, class, enum,
// module, or import declaration with the given name.
func HasLocalDeclarationInStatements(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindVariableStatement:
			varStmt := stmt.AsVariableStatement()
			if varStmt != nil && varStmt.DeclarationList != nil {
				declList := varStmt.DeclarationList.AsVariableDeclarationList()
				if declList != nil && declList.Declarations != nil {
					for _, decl := range declList.Declarations.Nodes {
						if decl != nil && decl.Kind == ast.KindVariableDeclaration {
							varDecl := decl.AsVariableDeclaration()
							if varDecl != nil && varDecl.Name() != nil {
								if HasNameInBindingPattern(varDecl.Name(), name) {
									return true
								}
							}
						}
					}
				}
			}

		case ast.KindFunctionDeclaration:
			if n := stmt.Name(); n != nil && n.Text() == name {
				return true
			}

		case ast.KindClassDeclaration:
			if n := stmt.Name(); n != nil && n.Text() == name {
				return true
			}

		case ast.KindEnumDeclaration:
			if n := stmt.Name(); n != nil && n.Text() == name {
				return true
			}

		case ast.KindModuleDeclaration:
			// Ambient module (`declare module "x"`) uses a string-literal name
			// and doesn't bind a variable — only identifier-named namespaces do.
			modDecl := stmt.AsModuleDeclaration()
			if modDecl != nil && modDecl.Name() != nil &&
				modDecl.Name().Kind == ast.KindIdentifier && modDecl.Name().Text() == name {
				return true
			}

		case ast.KindImportEqualsDeclaration:
			importEquals := stmt.AsImportEqualsDeclaration()
			if importEquals != nil && importEquals.Name() != nil && importEquals.Name().Text() == name {
				return true
			}

		case ast.KindImportDeclaration:
			importDecl := stmt.AsImportDeclaration()
			if importDecl != nil && importDecl.ImportClause != nil {
				importClause := importDecl.ImportClause.AsImportClause()
				if importClause != nil {
					// Default import: import X from 'mod'
					if importClause.Name() != nil && importClause.Name().Text() == name {
						return true
					}
					// Named/namespace imports
					if importClause.NamedBindings != nil {
						switch importClause.NamedBindings.Kind {
						case ast.KindNamespaceImport:
							nsImport := importClause.NamedBindings.AsNamespaceImport()
							if nsImport != nil && nsImport.Name() != nil && nsImport.Name().Text() == name {
								return true
							}
						case ast.KindNamedImports:
							namedImports := importClause.NamedBindings.AsNamedImports()
							if namedImports != nil && namedImports.Elements != nil {
								for _, elem := range namedImports.Elements.Nodes {
									if elem != nil {
										importSpec := elem.AsImportSpecifier()
										if importSpec != nil && importSpec.Name() != nil && importSpec.Name().Text() == name {
											return true
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

// HasNameInBindingPattern checks if a binding pattern (including nested
// destructuring) contains a binding with the given name.
func HasNameInBindingPattern(pattern *ast.Node, name string) bool {
	found := false
	CollectBindingNames(pattern, func(_ *ast.Node, n string) {
		if n == name {
			found = true
		}
	})
	return found
}

// HasVarDeclListWithName checks if a VariableDeclarationList contains a
// declaration with the given name (supports destructuring).
func HasVarDeclListWithName(node *ast.Node, name string) bool {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return false
	}
	for _, decl := range declList.Declarations.Nodes {
		if decl != nil && decl.Kind == ast.KindVariableDeclaration {
			varDecl := decl.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && HasNameInBindingPattern(varDecl.Name(), name) {
				return true
			}
		}
	}
	return false
}

// IsVarKeyword returns true if a VariableDeclarationList uses `var`
// (not `let`/`const`/`using`/`await using`).
func IsVarKeyword(node *ast.Node) bool {
	return node.Flags&ast.NodeFlagsBlockScoped == 0
}

// HasHoistedVarDeclaration recursively scans a subtree for `var` declarations
// with the given name. It stops at nested function boundaries and module
// declarations since `var` does not hoist past them.
func HasHoistedVarDeclaration(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt != nil && varStmt.DeclarationList != nil && IsVarKeyword(varStmt.DeclarationList) {
			if HasVarDeclListWithName(varStmt.DeclarationList, name) {
				return true
			}
		}
		return false

	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil && forStmt.Initializer != nil &&
			forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
			IsVarKeyword(forStmt.Initializer) &&
			HasVarDeclListWithName(forStmt.Initializer, name) {
			return true
		}

	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := node.AsForInOrOfStatement()
		if stmt != nil && stmt.Initializer != nil &&
			stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
			IsVarKeyword(stmt.Initializer) &&
			HasVarDeclListWithName(stmt.Initializer, name) {
			return true
		}

	// var does not hoist past function, static block, or module boundaries.
	case ast.KindModuleDeclaration:
		return false
	default:
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) {
			return false
		}
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if HasHoistedVarDeclaration(child, name) {
			found = true
			return true
		}
		return false
	})
	return found
}
