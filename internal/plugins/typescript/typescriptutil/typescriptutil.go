// Package typescriptutil collects helpers shared across rules in the
// @typescript-eslint plugin. Anything specific to a single rule stays with
// the rule; whatever a second rule needs (or would naturally re-implement
// with the same semantics) lives here.
package typescriptutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsReferenceToGlobalIdentifier mirrors typescript-eslint's
// `isReferenceToGlobalFunction` helper: an identifier counts as referring to
// a lib.d.ts-provided global only when no user-source declaration in the
// *current file* provides a binding for it.
//
// The check is layered to match the upstream ESLint scope-manager behavior
// across the cases tsgo's TypeChecker handles inconsistently:
//
//  1. The TypeChecker side — if `GetSymbolAtLocation` resolves the
//     identifier to a symbol whose declarations include any from the current
//     source file, treat the reference as locally bound. This covers
//     declaration merging (e.g. a same-file `interface Function {...}` that
//     augments the lib type) and value-side bindings (`class`, `let`,
//     `function`).
//
//  2. The lexical scope side — if any enclosing block, function body,
//     module declaration, or SourceFile statement list directly declares a
//     `type` alias, `interface`, `class`, `enum`, `namespace`, `var`,
//     `let`, `const`, `function`, or `import` with the same identifier
//     name, treat the reference as locally bound. This catches the
//     script-scope cases where a `type Number = 0 | 1` collides with the
//     lib `interface Number` and tsgo's checker resolves to the lib symbol
//     anyway — upstream's ESLint scope manager sees the local def and
//     stays silent.
//
// Returns true only when neither path finds a local binding (or the
// TypeChecker is unavailable AND the lexical walk also fails), matching
// upstream's "unresolved means global" fallback.
func IsReferenceToGlobalIdentifier(ctx rule.RuleContext, ident *ast.Node) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return true
	}
	name := ident.AsIdentifier().Text

	if ctx.TypeChecker != nil {
		if sym := ctx.TypeChecker.GetSymbolAtLocation(ident); sym != nil {
			for _, decl := range sym.Declarations {
				if decl == nil {
					continue
				}
				if ast.GetSourceFileOfNode(decl) == ctx.SourceFile {
					return false
				}
			}
		}
	}

	if hasLocalTypeOrValueDeclaration(ident, name) {
		return false
	}

	return true
}

// hasLocalTypeOrValueDeclaration walks from `ident` to the SourceFile,
// returning true when any enclosing scope holds a declaration that binds
// `name`. Unlike `utils.IsShadowed`, this also recognizes TypeScript-only
// type bindings — `type` aliases and `interface` declarations — that
// upstream's ESLint scope manager treats as shadows but tsgo's TypeChecker
// can miss when they conflict with a lib-provided global.
func hasLocalTypeOrValueDeclaration(ident *ast.Node, name string) bool {
	if utils.IsShadowed(ident, name) {
		return true
	}
	for current := ident.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindSourceFile:
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				if hasTypeOrInterfaceInStatements(sf.Statements.Nodes, name) {
					return true
				}
			}
			return false
		case ast.KindBlock:
			block := current.AsBlock()
			if block != nil && block.Statements != nil {
				if hasTypeOrInterfaceInStatements(block.Statements.Nodes, name) {
					return true
				}
			}
		case ast.KindModuleBlock:
			modBlock := current.AsModuleBlock()
			if modBlock != nil && modBlock.Statements != nil {
				if hasTypeOrInterfaceInStatements(modBlock.Statements.Nodes, name) {
					return true
				}
			}
		}
	}
	return false
}

// hasTypeOrInterfaceInStatements scans a flat statement list for a
// `type X = ...` or `interface X { ... }` declaration with the given name.
// All other shadow shapes (var/let/const/function/class/enum/module/import)
// are already covered by `utils.IsShadowed`; this helper is the type-only
// extension.
func hasTypeOrInterfaceInStatements(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindTypeAliasDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindInterfaceDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
	}
	return false
}

// IsClassImplementsOrInterfaceExtends reports whether the given
// `ExpressionWithTypeArguments` sits in a heritage position the
// typescript-eslint rules listen to as `TSClassImplements` or
// `TSInterfaceHeritage`:
//
//   - `class X implements ...`
//   - `interface X extends ...`
//
// `class X extends ...` is intentionally excluded: upstream does not register
// a listener for it. Non-heritage uses of `ExpressionWithTypeArguments` (e.g.
// JSDoc-style type contexts) are also rejected because upstream's listener
// set never matches them.
func IsClassImplementsOrInterfaceExtends(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || !ast.IsHeritageClause(parent) {
		return false
	}
	grand := parent.Parent
	switch {
	case grand == nil:
		return false
	case ast.IsInterfaceDeclaration(grand):
		return parent.AsHeritageClause().Token == ast.KindExtendsKeyword
	case ast.IsClassLike(grand):
		return parent.AsHeritageClause().Token == ast.KindImplementsKeyword
	}
	return false
}

// FirstSpreadIndex returns the index of the first SpreadElement argument in a
// call expression, or -1 when there is none. Arguments before the first spread
// can still be mapped to parameters by index; arguments at or past it cannot.
//
// Extracted from no_unnecessary_condition because strict_boolean_expressions
// needs the same predicate-parameter mapping; mirrors typescript-eslint's
// `firstSpreadIndex` helper.
func FirstSpreadIndex(callExpr *ast.CallExpression) int {
	for i, arg := range callExpr.Arguments.Nodes {
		if ast.IsSpreadElement(arg) {
			return i
		}
	}
	return -1
}

// FindTruthinessAssertedArgument returns the argument node corresponding to a
// truthiness-assertion call (`function f(x): asserts x`), or nil when the call
// is not such an assertion. Mirrors typescript-eslint's
// `findTruthinessAssertedArgument`:
//   - the resolved signature must carry an `asserts` type predicate with no
//     `Type()` (a truthiness assertion, not a type guard);
//   - the asserted parameter index must map to an argument before any spread.
//
// Extracted from no_unnecessary_condition to satisfy the plugin's
// duplicate-across-rules rule (strict_boolean_expressions consumes the same
// helper for its CallExpression listener).
func FindTruthinessAssertedArgument(tc *checker.Checker, callExpr *ast.CallExpression) *ast.Node {
	sig := checker.Checker_getResolvedSignature(tc, callExpr.AsNode(), nil, checker.CheckModeNormal)
	if sig == nil {
		return nil
	}

	predicate := tc.GetTypePredicateOfSignature(sig)
	if predicate == nil {
		return nil
	}

	if predicate.Type() != nil {
		return nil
	}

	if predicate.Kind() != checker.TypePredicateKindAssertsIdentifier {
		return nil
	}

	paramIndex := predicate.ParameterIndex()
	spreadIdx := FirstSpreadIndex(callExpr)
	if spreadIdx >= 0 && int(paramIndex) >= spreadIdx {
		return nil
	}
	if int(paramIndex) >= len(callExpr.Arguments.Nodes) {
		return nil
	}
	return callExpr.Arguments.Nodes[paramIndex]
}

// IsWeakPrecedenceParent mirrors typescript-eslint's `isWeakPrecedenceParent`
// inside `getWrappingFixer`: it asks whether `node`'s parent might silently
// rebind operator precedence after `node` is rewrapped, so the caller can
// decide whether to wrap the replacement in parens. The list is kept in sync
// with upstream's set:
//
//   - any unary/binary/conditional/await/typeof/void/delete parent
//   - a property access whose object IS `node`
//   - an element access whose expression IS `node`
//   - a call/new whose callee IS `node`
//   - a tagged template whose tag IS `node`
//
// Pairs with utils.IsStrongPrecedenceNode (which decides the inverse: whether
// the INNER expression already binds tight enough to skip its inner paren).
func IsWeakPrecedenceParent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindBinaryExpression,
		ast.KindConditionalExpression,
		ast.KindAwaitExpression,
		ast.KindTypeOfExpression,
		ast.KindVoidExpression,
		ast.KindDeleteExpression:
		return true
	}
	if ast.IsPropertyAccessExpression(parent) {
		return parent.AsPropertyAccessExpression().Expression == node
	}
	if ast.IsElementAccessExpression(parent) {
		return parent.AsElementAccessExpression().Expression == node
	}
	if ast.IsCallExpression(parent) || ast.IsNewExpression(parent) {
		return parent.Expression() == node
	}
	if ast.IsTaggedTemplateExpression(parent) {
		return parent.AsTaggedTemplateExpression().Tag == node
	}
	return false
}
