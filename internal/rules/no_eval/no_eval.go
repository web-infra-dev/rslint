package no_eval

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var globalObjects = []string{"window", "global", "globalThis"}

// https://eslint.org/docs/latest/rules/no-eval
var NoEvalRule = rule.Rule{
	Name: "no-eval",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		allowIndirect := false
		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			if v, ok := optsMap["allowIndirect"].(bool); ok {
				allowIndirect = v
			}
		}

		msg := rule.RuleMessage{
			Id:          "unexpected",
			Description: "`eval` can be harmful.",
		}

		if allowIndirect {
			// Only flag direct eval() calls
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					call := node.AsCallExpression()
					// Optional calls eval?.() are not direct eval
					if call.QuestionDotToken != nil {
						return
					}
					callee := call.Expression
					if callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "eval" {
						ctx.ReportNode(callee, msg)
					}
				},
			}
		}

		// Default mode: flag all eval usage
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := call.Expression
				if callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "eval" {
					ctx.ReportNode(callee, msg)
				}
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				propAccess := node.AsPropertyAccessExpression()
				name := propAccess.Name()
				if name == nil || name.Text() != "eval" {
					return
				}

				obj := ast.SkipParentheses(propAccess.Expression)
				if obj == nil {
					return
				}

				// Check for this.eval
				if obj.Kind == ast.KindThisKeyword {
					if isThisReferringToGlobal(obj, ctx.SourceFile) {
						ctx.ReportNode(name, msg)
					}
					return
				}

				// Check for window.eval, global.eval, globalThis.eval
				// and chained forms like window.window.eval
				if isGlobalObjectChain(obj, ctx.TypeChecker) {
					ctx.ReportNode(name, msg)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				elemAccess := node.AsElementAccessExpression()
				argExpr := elemAccess.ArgumentExpression
				if argExpr == nil {
					return
				}

				// Check if accessing ['eval'] or [`eval`]
				if utils.GetStaticStringValue(argExpr) != "eval" {
					return
				}

				obj := ast.SkipParentheses(elemAccess.Expression)
				if obj == nil {
					return
				}

				// Check for this['eval']
				if obj.Kind == ast.KindThisKeyword {
					if isThisReferringToGlobal(obj, ctx.SourceFile) {
						ctx.ReportNode(argExpr, msg)
					}
					return
				}

				if isGlobalObjectChain(obj, ctx.TypeChecker) {
					ctx.ReportNode(argExpr, msg)
				}
			},
			ast.KindIdentifier: func(node *ast.Node) {
				if node.AsIdentifier().Text != "eval" {
					return
				}

				parent := node.Parent
				if parent == nil {
					return
				}

				// Skip if this is a direct call callee (handled by KindCallExpression)
				if ast.IsCallExpression(parent) && parent.AsCallExpression().Expression == node {
					return
				}

				// Skip if this is a property name in property access (handled above)
				if ast.IsPropertyAccessExpression(parent) && parent.AsPropertyAccessExpression().Name() == node {
					return
				}

				// Skip non-reference identifiers.
				if utils.IsNonReferenceIdentifier(node) {
					return
				}

				// Skip if eval is shadowed by a local variable/parameter
				if utils.IsShadowed(node, "eval") {
					return
				}

				// Non-call reference to eval (e.g., var x = eval, func(eval))
				ctx.ReportNode(node, msg)
			},
		}
	},
}

// isGlobalObjectChain checks if a node represents a reference to a global object,
// potentially through chaining (e.g., window.window).
// When TypeChecker is available, it verifies the root identifier actually resolves
// to a known global (e.g., window from lib.dom.d.ts) rather than a local variable.
func isGlobalObjectChain(node *ast.Node, tc *checker.Checker) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	if ast.IsIdentifier(node) {
		if !slices.Contains(globalObjects, node.AsIdentifier().Text) {
			return false
		}
		// When TypeChecker is available, verify the identifier is actually declared
		// (from lib.dom.d.ts, @types/node, etc.). If it doesn't resolve, it's an
		// unknown variable — not a known global object.
		if tc != nil {
			if tc.GetSymbolAtLocation(node) == nil {
				return false
			}
		}
		return true
	}

	// Handle window.window, global.global, globalThis.globalThis, etc.
	if ast.IsPropertyAccessExpression(node) {
		prop := node.AsPropertyAccessExpression()
		name := prop.Name()
		if name != nil && slices.Contains(globalObjects, name.Text()) {
			return isGlobalObjectChain(prop.Expression, tc)
		}
	}

	// Handle window['window'], global['global'], etc.
	if ast.IsElementAccessExpression(node) {
		elem := node.AsElementAccessExpression()
		argText := utils.GetStaticStringValue(elem.ArgumentExpression)
		if slices.Contains(globalObjects, argText) {
			return isGlobalObjectChain(ast.SkipParentheses(elem.Expression), tc)
		}
	}

	return false
}

// isThisReferringToGlobal checks if 'this' at the given position refers to the global object.
// It uses ast.GetThisContainer to find the enclosing "this scope" (skipping arrow functions
// and class computed property names) and then determines whether 'this' is the global object.
func isThisReferringToGlobal(thisNode *ast.Node, sourceFile *ast.SourceFile) bool {
	// GetThisContainer with includeArrowFunctions=false skips arrow functions.
	// With includeClassComputedPropertyName=false, computed property names in
	// classes are transparent — the walker jumps past them to the outer scope.
	container := ast.GetThisContainer(thisNode, false /*includeArrowFunctions*/, false /*includeClassComputedPropertyName*/)

	switch container.Kind {
	case ast.KindSourceFile:
		// Top level of script — 'this' is always global (even in strict mode).
		// In modules, 'this' is undefined.
		return !ast.IsExternalModule(sourceFile)

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		// In strict mode, 'this' is undefined — not global.
		if utils.IsInStrictMode(thisNode, sourceFile) {
			return false
		}
		// Check how the function is used to determine if 'this' defaults to global.
		return utils.IsDefaultThisBinding(container)

	default:
		// MethodDeclaration, Constructor, GetAccessor, SetAccessor,
		// PropertyDeclaration (field value), ClassStaticBlockDeclaration, etc.
		// In all these cases 'this' refers to the instance/class, not global.
		return false
	}
}
