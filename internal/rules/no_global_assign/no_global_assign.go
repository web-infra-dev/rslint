package no_global_assign

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// builtinGlobals contains names of known read-only built-in globals.
var builtinGlobals = map[string]bool{
	"Array": true, "ArrayBuffer": true, "Atomics": true,
	"BigInt": true, "BigInt64Array": true, "BigUint64Array": true,
	"Boolean": true, "DataView": true, "Date": true,
	"decodeURI": true, "decodeURIComponent": true, "encodeURI": true,
	"encodeURIComponent": true, "Error": true, "EvalError": true,
	"Float32Array": true, "Float64Array": true, "Function": true,
	"globalThis": true, "Infinity": true, "Int8Array": true,
	"Int16Array": true, "Int32Array": true, "isFinite": true,
	"isNaN": true, "JSON": true, "Map": true, "Math": true,
	"NaN": true, "Number": true, "Object": true, "parseFloat": true,
	"parseInt": true, "Promise": true, "Proxy": true, "RangeError": true,
	"ReferenceError": true, "Reflect": true, "RegExp": true,
	"Set": true, "SharedArrayBuffer": true, "String": true,
	"Symbol": true, "SyntaxError": true, "TypeError": true,
	"Uint8Array": true, "Uint8ClampedArray": true, "Uint16Array": true,
	"Uint32Array": true, "URIError": true, "undefined": true,
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

// getIdentifierName extracts the name from an identifier node
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.Text()
}

// isWriteReference checks if a node is a write reference (assignment target)
func isWriteReference(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken,
			ast.KindPlusEqualsToken,
			ast.KindMinusEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken,
			ast.KindAsteriskEqualsToken,
			ast.KindSlashEqualsToken,
			ast.KindPercentEqualsToken,
			ast.KindLessThanLessThanEqualsToken,
			ast.KindGreaterThanGreaterThanEqualsToken,
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
			ast.KindAmpersandEqualsToken,
			ast.KindBarEqualsToken,
			ast.KindCaretEqualsToken,
			ast.KindBarBarEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken,
			ast.KindQuestionQuestionEqualsToken:
			return binary.Left == node
		}

	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}
		switch postfix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return postfix.Operand == node
		}

	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		switch prefix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return prefix.Operand == node
		}

	case ast.KindObjectBindingPattern:
		return isBindingPatternInAssignment(parent)

	case ast.KindArrayBindingPattern:
		return isBindingPatternInAssignment(parent)

	case ast.KindBindingElement:
		return isWriteReference(parent)

	case ast.KindShorthandPropertyAssignment:
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindPropertyAssignment:
		propAssignment := parent.AsPropertyAssignment()
		if propAssignment != nil && propAssignment.Initializer == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindArrayLiteralExpression:
		return isInDestructuringAssignment(parent)

	case ast.KindObjectLiteralExpression:
		return isInDestructuringAssignment(parent)

	case ast.KindParenthesizedExpression:
		return isWriteReference(parent)

	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return isWriteReference(parent)

	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := parent.AsForInOrOfStatement()
		if stmt != nil {
			return stmt.Initializer == node
		}
	}

	return false
}

// isBindingPatternInAssignment checks if a binding pattern is the left side of an assignment
func isBindingPatternInAssignment(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent

	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}

	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			leftNode := binary.Left
			for leftNode != nil && leftNode.Kind == ast.KindParenthesizedExpression {
				parenExpr := leftNode.AsParenthesizedExpression()
				if parenExpr != nil {
					leftNode = parenExpr.Expression
				} else {
					break
				}
			}
			return leftNode == node
		}
	}

	return false
}

// isInDestructuringAssignment checks if a node is part of a destructuring assignment pattern
func isInDestructuringAssignment(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindObjectLiteralExpression || current.Kind == ast.KindArrayLiteralExpression {
			parent := current.Parent

			for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
				parent = parent.Parent
			}

			if parent != nil && parent.Kind == ast.KindBinaryExpression {
				binary := parent.AsBinaryExpression()
				if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
					leftNode := binary.Left
					for leftNode != nil && leftNode.Kind == ast.KindParenthesizedExpression {
						parenExpr := leftNode.AsParenthesizedExpression()
						if parenExpr != nil {
							leftNode = parenExpr.Expression
						} else {
							break
						}
					}
					if leftNode == current {
						return true
					}
				}
			}
			return false
		}
		current = current.Parent
	}
	return false
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
			return false

		case ast.KindBlock:
			if hasShadowingVariable(current, name) {
				return true
			}

		case ast.KindFunctionDeclaration:
			if hasShadowingParameter(current, name) {
				return true
			}
			// Also check if a function declaration itself has this name
			funcDecl := current.AsFunctionDeclaration()
			if funcDecl != nil && funcDecl.Name() != nil && getIdentifierName(funcDecl.Name()) == name {
				return true
			}

		case ast.KindFunctionExpression:
			if hasShadowingParameter(current, name) {
				return true
			}

		case ast.KindArrowFunction:
			if hasShadowingParameter(current, name) {
				return true
			}

		case ast.KindMethodDeclaration:
			if hasShadowingParameter(current, name) {
				return true
			}

		case ast.KindConstructor:
			if hasShadowingParameter(current, name) {
				return true
			}

		case ast.KindGetAccessor:
			if hasShadowingParameter(current, name) {
				return true
			}

		case ast.KindSetAccessor:
			if hasShadowingParameter(current, name) {
				return true
			}

		case ast.KindCatchClause:
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if getIdentifierName(varDecl.Name()) == name {
						return true
					}
				}
			}

		case ast.KindClassDeclaration:
			classDecl := current.AsClassDeclaration()
			if classDecl != nil && classDecl.Name() != nil && getIdentifierName(classDecl.Name()) == name {
				return true
			}
		}
		current = current.Parent
	}
	return false
}

// hasLocalDeclarationInStatements checks if a list of statements contains a
// variable, function, class, or import declaration with the given name.
func hasLocalDeclarationInStatements(stmts []*ast.Node, name string) bool {
	for _, stmt := range stmts {
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
								if getIdentifierName(varDecl.Name()) == name {
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
				if getIdentifierName(funcDecl.Name()) == name {
					return true
				}
			}

		case ast.KindClassDeclaration:
			classDecl := stmt.AsClassDeclaration()
			if classDecl != nil && classDecl.Name() != nil {
				if getIdentifierName(classDecl.Name()) == name {
					return true
				}
			}

		case ast.KindImportDeclaration:
			importDecl := stmt.AsImportDeclaration()
			if importDecl != nil && importDecl.ImportClause != nil {
				importClause := importDecl.ImportClause.AsImportClause()
				if importClause != nil {
					// Default import: import X from 'mod'
					if importClause.Name() != nil && getIdentifierName(importClause.Name()) == name {
						return true
					}
					// Named imports: import { X } from 'mod' or namespace: import * as X from 'mod'
					if importClause.NamedBindings != nil {
						switch importClause.NamedBindings.Kind {
						case ast.KindNamespaceImport:
							nsImport := importClause.NamedBindings.AsNamespaceImport()
							if nsImport != nil && nsImport.Name() != nil && getIdentifierName(nsImport.Name()) == name {
								return true
							}
						case ast.KindNamedImports:
							namedImports := importClause.NamedBindings.AsNamedImports()
							if namedImports != nil && namedImports.Elements != nil {
								for _, elem := range namedImports.Elements.Nodes {
									if elem != nil {
										importSpec := elem.AsImportSpecifier()
										if importSpec != nil && importSpec.Name() != nil && getIdentifierName(importSpec.Name()) == name {
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

// hasShadowingParameter checks if a function has a parameter with the given name
func hasShadowingParameter(node *ast.Node, name string) bool {
	var params []*ast.Node

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl != nil && funcDecl.Parameters != nil {
			params = funcDecl.Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr != nil && funcExpr.Parameters != nil {
			params = funcExpr.Parameters.Nodes
		}
	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		if arrowFunc != nil && arrowFunc.Parameters != nil {
			params = arrowFunc.Parameters.Nodes
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Parameters != nil {
			params = method.Parameters.Nodes
		}
	case ast.KindConstructor:
		constructor := node.AsConstructorDeclaration()
		if constructor != nil && constructor.Parameters != nil {
			params = constructor.Parameters.Nodes
		}
	case ast.KindGetAccessor:
		getter := node.AsGetAccessorDeclaration()
		if getter != nil && getter.Parameters != nil {
			params = getter.Parameters.Nodes
		}
	case ast.KindSetAccessor:
		setter := node.AsSetAccessorDeclaration()
		if setter != nil && setter.Parameters != nil {
			params = setter.Parameters.Nodes
		}
	}

	for _, param := range params {
		if param != nil && param.Kind == ast.KindParameter {
			paramDecl := param.AsParameterDeclaration()
			if paramDecl != nil && paramDecl.Name() != nil && getIdentifierName(paramDecl.Name()) == name {
				return true
			}
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
								if getIdentifierName(varDecl.Name()) == name {
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

				if !isWriteReference(node) {
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
