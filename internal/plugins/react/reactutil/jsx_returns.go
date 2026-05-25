package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// FunctionReturnsJSXOrNull reports whether the function's body contains a
// `return <jsx/>` / `return null` / `return <pragma>.createElement(...)` at
// depth ≤ 1 (nested functions excluded), OR — for an arrow with expression
// body — whether that expression qualifies under the same rules.
// ConditionalExpression is traversed so `return cond ? <jsx/> : null`
// qualifies.
//
// Identifier returns (`return view` where `view` is bound to a JSX value)
// are resolved structurally via a local block scan. Use
// `FunctionReturnsJSXOrNullWithChecker` for full TypeChecker-based scope
// resolution.
//
// Mirrors upstream jsxUtil.isReturningJSX invoked with default arguments
// (which accept JSX, `null`, and `<pragma>.createElement(...)` returns).
// Pass an empty pragma to default to "React".
func FunctionReturnsJSXOrNull(fn *ast.Node, pragma string) bool {
	return functionReturnsJSXInternal(fn, true, pragma, nil)
}

// FunctionReturnsJSXOrNullWithChecker is the TypeChecker-aware variant of
// FunctionReturnsJSXOrNull. When `tc` is non-nil, Identifier returns are
// resolved through `GetSymbolAtLocation` → `Declarations[0]` →
// `VariableDeclaration.Initializer`, matching upstream's `findVariableByName`
// scope walk semantically (any binding the TS resolver can reach is
// considered, not just bindings in the immediately-enclosing block). When
// `tc` is nil, falls back to the local-block scan.
func FunctionReturnsJSXOrNullWithChecker(fn *ast.Node, pragma string, tc *checker.Checker) bool {
	return functionReturnsJSXInternal(fn, true, pragma, tc)
}

// FunctionReturnsJSX is the strict sibling of FunctionReturnsJSXOrNull:
// a `null` return does NOT qualify on its own. `<pragma>.createElement(...)`
// calls still qualify. Mirrors upstream jsxUtil.isReturningJSX invoked with
// `strict=true, ignoreNull=true`. `<pragma>.createElement(...)` calls still
// qualify. Pass an empty pragma to default to "React".
func FunctionReturnsJSX(fn *ast.Node, pragma string) bool {
	return functionReturnsJSXInternal(fn, false, pragma, nil)
}

// FunctionReturnsJSXWithChecker is the TypeChecker-aware strict variant.
// See FunctionReturnsJSXOrNullWithChecker for the resolution semantics.
func FunctionReturnsJSXWithChecker(fn *ast.Node, pragma string, tc *checker.Checker) bool {
	return functionReturnsJSXInternal(fn, false, pragma, tc)
}

func functionReturnsJSXInternal(fn *ast.Node, acceptNull bool, pragma string, tc *checker.Checker) bool {
	if fn == nil {
		return false
	}
	var body *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		body = fn.AsArrowFunction().Body
		if body != nil && body.Kind != ast.KindBlock {
			return isJSXExpression(body, acceptNull, pragma, tc)
		}
	case ast.KindMethodDeclaration:
		body = fn.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		body = fn.AsGetAccessorDeclaration().Body
	case ast.KindSetAccessor:
		body = fn.AsSetAccessorDeclaration().Body
	}
	if body == nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if found || n == nil {
			return found
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			rs := n.AsReturnStatement()
			if rs.Expression != nil && isJSXExpression(rs.Expression, acceptNull, pragma, tc) {
				found = true
				return true
			}
		case ast.KindFunctionExpression,
			ast.KindFunctionDeclaration,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			return false
		}
		n.ForEachChild(visit)
		return found
	}
	visit(body)
	return found
}

// isJSXExpression reports whether `expr` may evaluate to JSX (or to `null`
// when `acceptNull` is true) on at least one control-flow path. Walks through
// ParenthesizedExpression, TS expression wrappers (`as` / `satisfies` / `<T>x`
// / `x!`), ConditionalExpression and LogicalExpression (NON-strict semantics:
// either side qualifying is enough), comma-sequence right-most operands, and
// optional chains. A `<pragma>.createElement(...)` CallExpression also
// qualifies — upstream's jsxUtil.isReturningJSX treats `createElement` calls
// as JSX returns. An Identifier resolves through its declaring
// VariableDeclaration initializer when present, mirroring upstream's
// `findVariableByName` lookup but limited to const/let initializers within
// the same scope (no re-binding analysis).
//
// Strict semantics note: upstream's jsxUtil.isReturningJSX accepts a
// `strict` parameter that, when true, requires BOTH branches of a
// Conditional / LogicalExpression to qualify. Every call site in upstream
// `Components.js` (rev 7.x) passes `strict=undefined` (falsy = non-strict),
// so the strict mode is effectively unreachable through this rule and the
// no-unstable-nested-components rule itself. We therefore match upstream's
// observable behavior (non-strict for all current consumers) and omit the
// strict parameter; if a future rule needs strict mode it should be added
// then with the corresponding test coverage.
//
// Pass `acceptNull=true` for `isReturningJSXOrNull`-style gates and `false`
// for the strict `isReturningJSX` (ignoreNull=true) gates. Pass `tc` (the
// active TypeChecker) when scope-resolved Identifier lookup is desired;
// pass nil to fall back to a local-block initializer scan.
//
// Identifier-via-initializer resolution is one-step only — matching
// upstream's `isJSXValue → findVariableByName → isJSX(variable)` chain
// where `isJSX` accepts ONLY a JSXElement / JSXFragment node and does not
// recurse. No depth bookkeeping needed because the function does not
// recurse on Identifier; the only recursion sites (Conditional / comma /
// `&&` / `||` / `??`) walk strictly smaller AST subtrees.
func isJSXExpression(expr *ast.Node, acceptNull bool, pragma string, tc *checker.Checker) bool {
	expr = SkipExpressionWrappers(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	case ast.KindNullKeyword:
		return acceptNull
	case ast.KindCallExpression:
		return IsCreateElementCallWithChecker(expr.AsCallExpression().Expression, pragma, tc)
	case ast.KindIdentifier:
		// Upstream's `isJSXValue` for the Identifier case calls
		// `findVariableByName` and then `isJSX(variable)` — and `isJSX`
		// ONLY accepts JSXElement / JSXFragment. It does NOT recurse
		// into ConditionalExpression / LogicalExpression / CallExpression
		// (`createElement`) / nested Identifiers. We mirror that here:
		// resolve the initializer one step, accept iff the resolved node
		// is itself a JSX element/fragment. Anything else returns false.
		init := ResolveIdentifierInitializer(expr, tc)
		if init == nil {
			return false
		}
		init = SkipExpressionWrappers(init)
		switch init.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
			return true
		}
		return false
	case ast.KindConditionalExpression:
		ce := expr.AsConditionalExpression()
		return isJSXExpression(ce.WhenTrue, acceptNull, pragma, tc) || isJSXExpression(ce.WhenFalse, acceptNull, pragma, tc)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindCommaToken:
			return isJSXExpression(bin.Right, acceptNull, pragma, tc)
		case ast.KindAmpersandAmpersandToken,
			ast.KindBarBarToken,
			ast.KindQuestionQuestionToken:
			return isJSXExpression(bin.Left, acceptNull, pragma, tc) || isJSXExpression(bin.Right, acceptNull, pragma, tc)
		}
	}
	return false
}
