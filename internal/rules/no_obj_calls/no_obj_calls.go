package no_obj_calls

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var nonCallableGlobals = map[string]bool{
	"Math": true, "JSON": true, "Reflect": true, "Atomics": true, "Intl": true,
}

// skipAssertionsAndParens strips parentheses and all TS assertion wrappers
// (as, satisfies, !, <T>) from an expression using tsgo's built-in utility.
func skipAssertionsAndParens(node *ast.Node) *ast.Node {
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
}

// https://eslint.org/docs/latest/rules/no-obj-calls
var NoObjCallsRule = rule.Rule{
	Name:             "no-obj-calls",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// isGlobalSymbol returns true if the symbol comes from lib.d.ts
		// (none of its declarations are in the current source file).
		isGlobalSymbol := func(symbol *ast.Symbol) bool {
			return !utils.IsSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}

		// resolveExprToGlobal walks through an expression to find if any
		// branch resolves to a non-callable global. It handles:
		// - Identifiers and property access (including multi-hop variable chains)
		// - Conditional expressions (ternary)
		// - Logical operators (||, &&, ??)
		// - Comma operator (last expression)
		// - TS type assertions (as, satisfies, !, <T>) via SkipOuterExpressions
		var resolveExprToGlobal func(node *ast.Node) (string, bool)
		resolveExprToGlobal = func(node *ast.Node) (string, bool) {
			node = skipAssertionsAndParens(node)
			switch node.Kind {
			case ast.KindIdentifier, ast.KindPropertyAccessExpression:
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil {
					return "", false
				}
				if nonCallableGlobals[symbol.Name] && isGlobalSymbol(symbol) {
					return symbol.Name, true
				}
				// Multi-hop: trace through variable declarations
				for _, decl := range symbol.Declarations {
					if ast.IsVariableDeclaration(decl) {
						varDecl := decl.AsVariableDeclaration()
						if varDecl.Initializer != nil {
							if name, ok := resolveExprToGlobal(varDecl.Initializer); ok {
								return name, true
							}
						}
					}
				}
			case ast.KindConditionalExpression:
				cond := node.AsConditionalExpression()
				if name, ok := resolveExprToGlobal(cond.WhenTrue); ok {
					return name, true
				}
				return resolveExprToGlobal(cond.WhenFalse)
			case ast.KindBinaryExpression:
				bin := node.AsBinaryExpression()
				switch bin.OperatorToken.Kind {
				case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
					if name, ok := resolveExprToGlobal(bin.Left); ok {
						return name, true
					}
					return resolveExprToGlobal(bin.Right)
				case ast.KindCommaToken:
					return resolveExprToGlobal(bin.Right)
				}
			}
			return "", false
		}

		// getCalleeName returns the display name for the callee expression.
		// For identifiers, returns the name; for member expressions, the property
		// name; for anything else (e.g. TS assertions), returns "undefined"
		// to match ESLint's getReportNodeName behavior.
		getCalleeName := func(node *ast.Node) string {
			node = ast.SkipParentheses(node)
			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := node.AsPropertyAccessExpression()
				if propAccess.Name() != nil {
					return propAccess.Name().Text()
				}
			case ast.KindIdentifier:
				return node.AsIdentifier().Text
			}
			return "undefined"
		}

		checkCallee := func(node *ast.Node, calleeNode *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}

			callee := skipAssertionsAndParens(calleeNode)
			symbol := ctx.TypeChecker.GetSymbolAtLocation(callee)
			if symbol == nil {
				return
			}

			name := getCalleeName(calleeNode)

			// Direct or TS-assertion-wrapped callee resolves to a non-callable global.
			if nonCallableGlobals[symbol.Name] && isGlobalSymbol(symbol) {
				if name == symbol.Name {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedCall",
						Description: fmt.Sprintf("'%s' is not a function.", name),
					})
				} else {
					// Callee name differs from global (e.g. TS assertion wrapping).
					// ESLint reports this as an indirect reference.
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedRefCall",
						Description: fmt.Sprintf("'%s' is reference to '%s', which is not a function.", name, symbol.Name),
					})
				}
				return
			}

			// Indirect: callee is a local variable whose initializer
			// (possibly through multiple hops) references a non-callable global.
			for _, decl := range symbol.Declarations {
				if ast.IsVariableDeclaration(decl) {
					varDecl := decl.AsVariableDeclaration()
					if varDecl.Initializer != nil {
						if globalName, ok := resolveExprToGlobal(varDecl.Initializer); ok {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          "unexpectedRefCall",
								Description: fmt.Sprintf("'%s' is reference to '%s', which is not a function.", name, globalName),
							})
							return
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				checkCallee(node, callExpr.Expression)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkCallee(node, newExpr.Expression)
			},
		}
	},
}
