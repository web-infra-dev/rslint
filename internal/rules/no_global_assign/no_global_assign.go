package no_global_assign

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// builtinGlobals contains names of known read-only built-in globals.
var builtinGlobals = map[string]bool{
	"AggregateError": true, "Array": true, "ArrayBuffer": true, "AsyncDisposableStack": true,
	"AsyncIterator": true, "Atomics": true,
	"BigInt": true, "BigInt64Array": true, "BigUint64Array": true,
	"Boolean": true, "DataView": true, "Date": true,
	"decodeURI": true, "decodeURIComponent": true, "DisposableStack": true,
	"encodeURI": true, "encodeURIComponent": true,
	"Error": true, "escape": true, "EvalError": true,
	"FinalizationRegistry": true, "Float32Array": true, "Float64Array": true, "Function": true,
	"globalThis": true, "Infinity": true, "Int8Array": true,
	"Int16Array": true, "Int32Array": true, "Intl": true, "isFinite": true,
	"isNaN": true, "Iterator": true, "JSON": true, "Map": true, "Math": true,
	"NaN": true, "Number": true, "Object": true, "parseFloat": true,
	"parseInt": true, "Promise": true, "Proxy": true, "RangeError": true,
	"ReferenceError": true, "Reflect": true, "RegExp": true,
	"Set": true, "SharedArrayBuffer": true, "String": true, "SuppressedError": true,
	"Symbol": true, "SyntaxError": true, "TypeError": true,
	"Uint8Array": true, "Uint8ClampedArray": true, "Uint16Array": true,
	"Uint32Array": true, "unescape": true, "URIError": true, "undefined": true,
	"WeakMap": true, "WeakRef": true, "WeakSet": true,
}

type options struct {
	exceptions map[string]bool
}

func parseOptions(opts any) options {
	result := options{exceptions: make(map[string]bool)}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if exceptions, ok := optsMap["exceptions"].([]interface{}); ok {
			for _, e := range exceptions {
				if s, ok := e.(string); ok {
					result.exceptions[s] = true
				}
			}
		}
	}
	return result
}

// isShadowed checks whether the given identifier name is shadowed by a local declaration.
// It walks up from the node looking for variable declarations, function declarations,
// function/method parameters, class declarations, import declarations, and catch variables.
func isShadowed(node *ast.Node, name string) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile:
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				if hasLocalDeclarationInStatements(sf.Statements.Nodes, name) {
					return true
				}
			}
			if hasHoistedVarDeclaration(current, name) {
				return true
			}
			return false

		case ast.KindBlock:
			if hasShadowingVariable(current, name) {
				return true
			}

		case ast.KindCatchClause:
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if hasNameInBindingPattern(varDecl.Name(), name) {
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
			if hasForInitializerDeclaration(current, name) {
				return true
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			if hasForInOfDeclaration(current, name) {
				return true
			}

		default:
			if ast.IsFunctionLikeDeclaration(current) {
				if hasShadowingParameter(current, name) {
					return true
				}
				// FunctionDeclaration and FunctionExpression can shadow via their own name
				if current.Kind == ast.KindFunctionDeclaration || current.Kind == ast.KindFunctionExpression {
					if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
						return true
					}
				}
				if body := current.Body(); body != nil && hasHoistedVarDeclaration(body, name) {
					return true
				}
			}
		}
		current = current.Parent
	}
	return false
}

// hasLocalDeclarationInStatements checks if a list of statements contains a
// variable, function, class, or import declaration with the given name.
func hasLocalDeclarationInStatements(statements []*ast.Node, name string) bool {
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
								if hasNameInBindingPattern(varDecl.Name(), name) {
									return true
								}
							}
						}
					}
				}
			}

		case ast.KindFunctionDeclaration:
			funcDecl := stmt.AsFunctionDeclaration()
			if funcDecl != nil && funcDecl.Name() != nil {
				if funcDecl.Name().Text() == name {
					return true
				}
			}

		case ast.KindClassDeclaration:
			classDecl := stmt.AsClassDeclaration()
			if classDecl != nil && classDecl.Name() != nil {
				if classDecl.Name().Text() == name {
					return true
				}
			}

		case ast.KindEnumDeclaration:
			enumDecl := stmt.AsEnumDeclaration()
			if enumDecl != nil && enumDecl.Name() != nil {
				if enumDecl.Name().Text() == name {
					return true
				}
			}

		case ast.KindModuleDeclaration:
			modDecl := stmt.AsModuleDeclaration()
			if modDecl != nil && modDecl.Name() != nil && modDecl.Name().Text() == name {
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
					// Named imports: import { X } from 'mod' or namespace: import * as X from 'mod'
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

// hasNameInBindingPattern checks if a binding pattern (including nested patterns)
// contains a binding with the given name.
func hasNameInBindingPattern(pattern *ast.Node, name string) bool {
	found := false
	utils.CollectBindingNames(pattern, func(_ *ast.Node, n string) {
		if n == name {
			found = true
		}
	})
	return found
}

// hasShadowingParameter checks if a function-like node has a parameter with the given name
func hasShadowingParameter(node *ast.Node, name string) bool {
	params := node.Parameters()
	for _, param := range params {
		if param == nil {
			continue
		}
		nameNode := param.Name()
		if nameNode == nil {
			continue
		}
		found := false
		utils.CollectBindingNames(nameNode, func(_ *ast.Node, n string) {
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

// hasShadowingVariable checks if a block contains a variable declaration with the given name
func hasShadowingVariable(node *ast.Node, name string) bool {
	if node.Kind != ast.KindBlock {
		return false
	}

	block := node.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}

	for _, stmt := range block.Statements.Nodes {
		if stmt != nil && stmt.Kind == ast.KindVariableStatement {
			varStmt := stmt.AsVariableStatement()
			if varStmt != nil && varStmt.DeclarationList != nil {
				declList := varStmt.DeclarationList.AsVariableDeclarationList()
				if declList != nil && declList.Declarations != nil {
					for _, decl := range declList.Declarations.Nodes {
						if decl != nil && decl.Kind == ast.KindVariableDeclaration {
							varDecl := decl.AsVariableDeclaration()
							if varDecl != nil && varDecl.Name() != nil {
								if hasNameInBindingPattern(varDecl.Name(), name) {
									return true
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

// hasForInitializerDeclaration checks if a for-statement has a variable declaration
// in its initializer that matches the given name.
func hasForInitializerDeclaration(node *ast.Node, name string) bool {
	forStmt := node.AsForStatement()
	if forStmt == nil || forStmt.Initializer == nil {
		return false
	}
	if forStmt.Initializer.Kind != ast.KindVariableDeclarationList {
		return false
	}
	return hasVarDeclListWithName(forStmt.Initializer, name)
}

// hasForInOfDeclaration checks if a for-in/for-of statement has a variable
// declaration in its initializer that matches the given name.
func hasForInOfDeclaration(node *ast.Node, name string) bool {
	stmt := node.AsForInOrOfStatement()
	if stmt == nil || stmt.Initializer == nil {
		return false
	}
	if stmt.Initializer.Kind != ast.KindVariableDeclarationList {
		return false
	}
	return hasVarDeclListWithName(stmt.Initializer, name)
}

// hasVarDeclListWithName checks if a VariableDeclarationList contains a
// declaration with the given name.
func hasVarDeclListWithName(node *ast.Node, name string) bool {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return false
	}
	for _, decl := range declList.Declarations.Nodes {
		if decl != nil && decl.Kind == ast.KindVariableDeclaration {
			varDecl := decl.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && hasNameInBindingPattern(varDecl.Name(), name) {
				return true
			}
		}
	}
	return false
}

// isVarKeyword returns true if a VariableDeclarationList uses `var` (not `let`/`const`/`using`/`await using`).
func isVarKeyword(node *ast.Node) bool {
	return node.Flags&ast.NodeFlagsBlockScoped == 0
}

// hasHoistedVarDeclaration recursively scans a subtree for `var` declarations
// with the given name. It stops at nested function boundaries since `var` does
// not hoist past them.
func hasHoistedVarDeclaration(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt != nil && varStmt.DeclarationList != nil && isVarKeyword(varStmt.DeclarationList) {
			if hasVarDeclListWithName(varStmt.DeclarationList, name) {
				return true
			}
		}
		return false

	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil && forStmt.Initializer != nil &&
			forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
			isVarKeyword(forStmt.Initializer) &&
			hasVarDeclListWithName(forStmt.Initializer, name) {
			return true
		}

	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := node.AsForInOrOfStatement()
		if stmt != nil && stmt.Initializer != nil &&
			stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
			isVarKeyword(stmt.Initializer) &&
			hasVarDeclListWithName(stmt.Initializer, name) {
			return true
		}

	// Do not descend into nested function-like nodes, static blocks, or module declarations —
	// var does not hoist past them.
	case ast.KindModuleDeclaration:
		return false
	default:
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) {
			return false
		}
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if hasHoistedVarDeclaration(child, name) {
			found = true
			return true // stop iteration
		}
		return false
	})
	return found
}

// isWriteThroughTypeAssertion checks if the identifier reaches its assignment target
// through an AsExpression or TypeAssertionExpression. ESLint's scope analysis does not
// track writes through these TS-specific wrappers, so we skip them to match ESLint.
func isWriteThroughTypeAssertion(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			return true
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			current = current.Parent
			continue
		default:
			return false
		}
	}
	return false
}

// NoGlobalAssignRule disallows assignments to native objects or read-only global variables
var NoGlobalAssignRule = rule.Rule{
	Name: "no-global-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				if !builtinGlobals[name] || opts.exceptions[name] {
					return
				}

				if !utils.IsWriteReference(node) {
					return
				}

				if isWriteThroughTypeAssertion(node) {
					return
				}

				if isShadowed(node, name) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "globalShouldNotBeModified",
					Description: fmt.Sprintf("Read-only global '%s' should not be modified.", name),
				})
			},
		}
	},
}
