// Package valid_expect holds the framework-neutral async-assertion machinery
// shared by jest/valid-expect and rstest/valid-expect: walking out of a chained
// .then/.catch, spotting a Promise.all([...]) wrapper, deciding whether the
// result is awaited or returned, and the async/await fixes.
//
// Only the parts whose correctness does not depend on the framework live here.
// The rules keep their own bodies: how a broken chain's reason is derived, what
// counts as an excess expect argument and which matcher shapes exist all differ
// between jest and rstest, and merging those would mean framework branches
// inside a shared body.
package valid_expect

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func isGlobalPromiseObject(node *ast.Node) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text == "Promise" && !internalUtils.IsShadowed(node, "Promise")
	}
	if ast.IsAccessExpression(node) {
		name, ok := accessStaticNameThroughAssertions(node)
		if !ok || name != "Promise" {
			return false
		}
		obj := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(node))
		return obj != nil && obj.Kind == ast.KindIdentifier &&
			obj.AsIdentifier().Text == "globalThis" &&
			!internalUtils.IsShadowed(obj, "globalThis")
	}
	return false
}

func accessStaticNameThroughAssertions(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if name, ok := internalUtils.AccessExpressionStaticName(node); ok {
		return name, true
	}
	if node.Kind == ast.KindElementAccessExpression {
		arg := internalUtils.SkipAssertionsAndParens(node.AsElementAccessExpression().ArgumentExpression)
		return internalUtils.GetStaticStringLiteralValue(arg)
	}
	return "", false
}

func promiseMethodName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindCallExpression {
		return "", false
	}

	callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if !ast.IsAccessExpression(callee) {
		return "", false
	}

	if !isGlobalPromiseObject(internalUtils.AccessExpressionObject(callee)) {
		return "", false
	}

	// Only statically named Promise methods (dot access or string/template
	// literal keys) have known semantics. Computed identifier keys such as
	// `Promise[resolve]([...])` where `const resolve = 'all'` return ("", true)
	// so they are not mistaken for `Promise.resolve`.
	method, _ := accessStaticNameThroughAssertions(callee)
	return method, true
}

// outermostTransparentExpression walks upward across parentheses and
// TypeScript assertion wrappers (`as`, `<T>`, `satisfies`, `!`) that are erased
// at runtime and preserve the expression's Promise value.
func outermostTransparentExpression(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			if paren := parent.AsParenthesizedExpression(); paren != nil && paren.Expression == current {
				current = parent
				continue
			}
		} else if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == current {
			current = parent
			continue
		}
		break
	}
	return current
}

func getPromiseCallExpressionNode(node *ast.Node, directArg *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	if node.Kind == ast.KindArrayLiteralExpression {
		outer := outermostTransparentExpression(node)
		if outer == nil || outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression {
			return nil
		}
		call := outer.Parent
		args := call.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) == 0 || args.Nodes[0] != outer {
			return nil
		}
		method, ok := promiseMethodName(call)
		// `Promise.resolve([...])`, `Promise.reject([...])`, and
		// `Promise.withResolvers([...])` do not await array elements, unlike
		// `Promise.all` / `allSettled` / `race` / `any` (or custom combinators
		// such as `Promise.x`).
		if !ok || method == "resolve" || method == "reject" || method == "withResolvers" {
			return nil
		}
		return call
	}

	method, ok := promiseMethodName(node)
	if !ok || method == "withResolvers" {
		return nil
	}
	if directArg != nil {
		switch method {
		case "resolve", "reject", "all", "allSettled", "race", "any":
			args := node.AsCallExpression().Arguments
			if args == nil || len(args.Nodes) == 0 || args.Nodes[0] != directArg {
				return nil
			}
		}
	}
	return node
}

func findPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	current := outermostTransparentExpression(node)
	for current != nil && current.Parent != nil {
		parent := current.Parent
		// A conditional branch or comma-operator RHS inside an array/call
		// (`Promise.all([cond ? expect(...) : ...])`, `Promise.all([(log(), expect(...))])`)
		// forwards the assertion promise to the enclosing container.
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == current || cond.WhenFalse == current) {
				current = outermostTransparentExpression(parent)
				continue
			}
			return nil
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == current {
			current = outermostTransparentExpression(parent)
			continue
		}
		if parent.Kind == ast.KindCallExpression {
			if parent.AsCallExpression().Expression == current {
				return nil
			}
			return getPromiseCallExpressionNode(parent, current)
		}
		if parent.Kind == ast.KindArrayLiteralExpression {
			return getPromiseCallExpressionNode(parent, nil)
		}
		return nil
	}
	return nil
}

func getParentIfPromiseChained(node *ast.Node) *ast.Node {
	outer := outermostTransparentExpression(node)
	for outer != nil && outer.Parent != nil {
		parent := outer.Parent
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == outer || cond.WhenFalse == outer) {
				outer = outermostTransparentExpression(parent)
				continue
			}
			return node
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == outer {
			outer = outermostTransparentExpression(parent)
			continue
		}
		break
	}
	if outer == nil || outer.Parent == nil {
		return node
	}

	member := outer.Parent
	if !ast.IsAccessExpression(member) || internalUtils.AccessExpressionObject(member) != outer {
		return node
	}

	outerMember := outermostTransparentExpression(member)
	grandParent := outerMember.Parent
	if grandParent == nil ||
		grandParent.Kind != ast.KindCallExpression ||
		grandParent.AsCallExpression().Expression != outerMember {
		return node
	}

	last, ok := accessStaticNameThroughAssertions(member)
	if ok && (last == "then" || last == "catch" || last == "finally") {
		return getParentIfPromiseChained(grandParent)
	}

	return node
}

func isPromiseChainMethodCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return false
	}
	name, ok := accessStaticNameThroughAssertions(callee)
	return ok && (name == "then" || name == "catch" || name == "finally")
}

func isAcceptableReturnNode(child, parent *ast.Node, allowReturn bool) bool {
	if parent == nil {
		return false
	}

	if allowReturn && parent.Kind == ast.KindReturnStatement {
		return parent.AsReturnStatement().Expression == child
	}
	if parent.Kind == ast.KindConditionalExpression {
		cond := parent.AsConditionalExpression()
		// Only a value-producing branch can forward the assertion promise;
		// testing the promise in the condition discards it as always-truthy.
		if cond == nil || (cond.WhenTrue != child && cond.WhenFalse != child) {
			return false
		}
		outer := outermostTransparentExpression(parent)
		return isAcceptableReturnNode(outer, outer.Parent, allowReturn)
	}
	if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == child {
		outer := outermostTransparentExpression(parent)
		return isAcceptableReturnNode(outer, outer.Parent, allowReturn)
	}
	if parent.Kind == ast.KindArrowFunction {
		return parent.AsArrowFunction().Body == child
	}
	if parent.Kind == ast.KindAwaitExpression {
		return parent.AsAwaitExpression().Expression == child
	}

	return false
}

func isTerminalFunctionReturn(ret *ast.Node) bool {
	if ret == nil || ret.Kind != ast.KindReturnStatement || ret.Parent == nil || ret.Parent.Kind != ast.KindBlock {
		return false
	}
	block := ret.Parent
	fn := block.Parent
	if fn == nil || !testFramework.IsFunction(fn) {
		return false
	}
	if block.AsBlock() == nil || block.AsBlock().Statements == nil {
		return false
	}
	stmts := block.AsBlock().Statements.Nodes
	return len(stmts) > 0 && stmts[len(stmts)-1] == ret
}

// isSafeTerminalReturnStrip reports whether `return expr` can safely be
// rewritten to `await expr` (dropping `return`) rather than `return await expr`.
// Only anonymous callbacks passed directly as call arguments (such as
// `test("...", () => { return expect(...); })`) returning a plain assertion
// promise strip `return`. Value-returning helpers, assigned functions,
// object/class methods (e.g. `expect.extend`), `.then()`/`Promise.*` wrappers
// that may resolve to a non-undefined value, and functions with explicit return
// type annotations keep `return await`.
func isSafeTerminalReturnStrip(ret *ast.Node, expr *ast.Node) bool {
	if !isTerminalFunctionReturn(ret) {
		return false
	}
	fn := ret.Parent.Parent
	if fn == nil || fn.Type() != nil {
		return false
	}
	if fn.Kind != ast.KindArrowFunction && fn.Kind != ast.KindFunctionExpression {
		return false
	}
	fnOuter := outermostTransparentExpression(fn)
	if fnOuter == nil || fnOuter.Parent == nil ||
		fnOuter.Parent.Kind != ast.KindCallExpression ||
		fnOuter.Parent.AsCallExpression().Expression == fnOuter {
		return false
	}
	unwrapped := internalUtils.SkipAssertionsAndParens(expr)
	if _, isPromiseWrapper := promiseMethodName(unwrapped); isPromiseWrapper {
		return false
	}
	if isPromiseChainMethodCall(unwrapped) {
		return false
	}
	return true
}

// FindFixableContainingFunction returns the enclosing function that can safely
// receive `await` (and `async` if not already async) for node. It returns
// (nil, false) when node sits inside a scope where `await` or `async` is
// illegal or would break callers (constructors, getters, setters, default
// parameters, class field initializers, class static blocks, sync generators,
// Annex B single-statement function declarations, directly constructed
// `new (function () {})()`, functions using `await` as an identifier, or
// non-async functions with explicit non-Promise return type annotations).
func FindFixableContainingFunction(node *ast.Node) (*ast.Node, bool) {
	return FindFixableContainingFunctionCached(node, nil)
}

// FindFixableContainingFunctionCached is the memoized form of
// FindFixableContainingFunction. Passing a per-file cache map avoids rescanning
// a large synchronous function body for `await` identifiers on every diagnostic.
func FindFixableContainingFunctionCached(node *ast.Node, fnAllowed map[*ast.Node]bool) (*ast.Node, bool) {
	if node == nil {
		return nil, false
	}
	prev := node
	for current := node.Parent; current != nil; prev, current = current, current.Parent {
		switch current.Kind {
		case ast.KindClassStaticBlockDeclaration:
			return nil, false
		case ast.KindPropertyDeclaration:
			if prop := current.AsPropertyDeclaration(); prop != nil && prop.Initializer == prev {
				return nil, false
			}
		case ast.KindParameter:
			// `await` is a syntax error inside parameter default initializers
			// even when the function itself is async.
			return nil, false
		case ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindMethodSignature, ast.KindCallSignature, ast.KindConstructSignature,
			ast.KindFunctionType, ast.KindConstructorType:
			// Decorators and computed property keys evaluate in the enclosing
			// scope rather than inside this member's body.
			if prev.Kind == ast.KindDecorator || prev == current.Name() {
				continue
			}
			return nil, false
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
			if prev.Kind == ast.KindDecorator || prev == current.Name() {
				continue
			}
			if current.Body() != prev {
				return nil, false
			}
			if allowed, cached := fnAllowed[current]; cached {
				if !allowed {
					return nil, false
				}
				return current, true
			}
			allowed := isFunctionSafelyAsyncFixable(current)
			if fnAllowed != nil {
				fnAllowed[current] = allowed
			}
			if !allowed {
				return nil, false
			}
			return current, true
		}
	}
	return nil, false
}

func isFunctionSafelyAsyncFixable(fn *ast.Node) bool {
	if ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync) {
		return true
	}
	if functionAsteriskToken(fn) != nil {
		// Converting a synchronous generator into an async generator changes
		// its runtime return protocol.
		return false
	}
	if hasExplicitNonPromiseReturnType(fn) {
		return false
	}
	if fn.Kind == ast.KindFunctionDeclaration && !isValidAsyncFunctionDeclarationParent(fn.Parent) {
		return false
	}
	if fn.Kind == ast.KindFunctionExpression {
		fnOuter := outermostTransparentExpression(fn)
		if fnOuter != nil && fnOuter.Parent != nil &&
			fnOuter.Parent.Kind == ast.KindNewExpression &&
			fnOuter.Parent.AsNewExpression().Expression == fnOuter {
			return false
		}
	}
	if containsIdentifierNamedAwait(fn) {
		return false
	}
	return true
}

func isValidAsyncFunctionDeclarationParent(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return true
	default:
		return false
	}
}

// containsIdentifierNamedAwait reports whether a non-async function uses
// `await` as its own name, an inner declaration name, a parameter, a binding,
// or an expression identifier in its own scope, where adding `async` would
// turn `await` into a reserved keyword syntax error.
func containsIdentifierNamedAwait(fn *ast.Node) bool {
	if fn.Kind == ast.KindFunctionExpression || fn.Kind == ast.KindFunctionDeclaration {
		if n := fn.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "await" {
			return true
		}
	}
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if node != fn {
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				// The function's name is declared in `fn`'s own scope, while
				// its parameters and body form a nested non-async scope.
				if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "await" {
					found = true
				}
				return
			case ast.KindFunctionExpression, ast.KindArrowFunction:
				return
			case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				// Decorators and computed property keys of inner methods
				// evaluate in `fn`'s own scope.
				if mods := node.Modifiers(); mods != nil {
					for _, mod := range mods.Nodes {
						if mod != nil && mod.Kind == ast.KindDecorator {
							visit(mod)
						}
					}
				}
				if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					visit(name)
				}
				return
			case ast.KindTypeReference, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
				return
			}
		}
		if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "await" {
			parent := node.Parent
			if parent != nil {
				switch parent.Kind {
				case ast.KindPropertyAccessExpression:
					if parent.AsPropertyAccessExpression().Name() == node {
						return
					}
				case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
					ast.KindGetAccessor, ast.KindSetAccessor, ast.KindBindingElement:
					if parent.Name() == node && parent.Kind != ast.KindBindingElement {
						return
					}
					if parent.Kind == ast.KindBindingElement && parent.AsBindingElement().PropertyName == node {
						return
					}
				}
			}
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	for _, param := range fn.Parameters() {
		visit(param)
	}
	visit(fn.Body())
	return found
}

func functionAsteriskToken(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		return fn.AsFunctionDeclaration().AsteriskToken
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().AsteriskToken
	case ast.KindMethodDeclaration:
		return fn.AsMethodDeclaration().AsteriskToken
	default:
		return nil
	}
}

// IsAsyncFunctionLike reports whether fn already carries the `async` modifier
// (including `async function*` and `async *m()`, unlike `ast.IsAsyncFunction`).
func IsAsyncFunctionLike(fn *ast.Node) bool {
	return fn != nil && ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
}

func hasExplicitNonPromiseReturnType(fn *ast.Node) bool {
	retType := fn.Type()
	if retType == nil {
		return false
	}
	if retType.Kind == ast.KindTypeReference {
		ref := retType.AsTypeReferenceNode()
		if ref != nil && ref.TypeName != nil {
			if ref.TypeName.Kind == ast.KindIdentifier && ref.TypeName.AsIdentifier().Text == "Promise" {
				return false
			}
			if ref.TypeName.Kind == ast.KindQualifiedName {
				qn := ref.TypeName.AsQualifiedName()
				if qn != nil && qn.Left != nil && qn.Left.Kind == ast.KindIdentifier &&
					qn.Left.AsIdentifier().Text == "globalThis" &&
					qn.Right != nil && qn.Right.Text() == "Promise" {
					return false
				}
			}
		}
	}
	return true
}

// PromiseArrayExceptionKey identifies a Promise.all([...]) call by source range,
// so the assertions inside its array report once between them.
func PromiseArrayExceptionKey(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	return fmt.Sprintf("%d:%d", r.Pos(), r.End())
}

func needsLeadingSpace(text string, pos int) bool {
	if pos <= 0 || pos > len(text) {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(text[:pos])
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

// AsyncInsertFix marks the function containing an async assertion as async.
func AsyncInsertFix(sourceFile *ast.SourceFile, fn *ast.Node) rule.RuleFix {
	text := sourceFile.Text()
	var pos int
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		pos = internalUtils.FindFunctionKeywordPos(sourceFile, fn)
	case ast.KindMethodDeclaration:
		// Insert `async ` after all modifiers/decorators (e.g. `static`,
		// `public`, `override`), right before `*` or the method name.
		if asterisk := functionAsteriskToken(fn); asterisk != nil {
			pos = scanner.SkipTrivia(text, asterisk.Pos())
		} else if fn.Name() != nil {
			pos = scanner.SkipTrivia(text, fn.Name().Pos())
		} else {
			pos = internalUtils.GetFunctionHeadLoc(sourceFile, fn).Pos()
		}
	default:
		pos = internalUtils.TrimNodeTextRange(sourceFile, fn).Pos()
	}
	insertion := "async "
	if needsLeadingSpace(text, pos) {
		insertion = " async "
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), insertion)
}

// CanSafelyAwaitFix reports whether inserting `await` at node preserves the
// test's control flow, syntax validity, and downstream value expectations.
func CanSafelyAwaitFix(node *ast.Node) bool {
	if node == nil {
		return false
	}
	outer := outermostTransparentExpression(node)
	if outer == nil || outer.Parent == nil {
		return true
	}

	// Walk upward through value-propagating expressions to detect member/call/
	// syntax-forbidden positions, destructuring, comparisons, truthiness tests,
	// switch discriminants, or short-circuiting operators where turning a
	// Promise into its resolved value (`undefined`) breaks syntax or semantics.
	cur := outer
	for cur != nil && cur.Parent != nil {
		parent := cur.Parent
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if internalUtils.AccessExpressionObject(parent) == cur {
				return false
			}
			return true
		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression == cur {
				return false
			}
			return true
		case ast.KindNewExpression:
			if parent.AsNewExpression().Expression == cur {
				return false
			}
			return true
		case ast.KindTaggedTemplateExpression:
			if parent.AsTaggedTemplateExpression().Tag == cur {
				return false
			}
			return true
		case ast.KindExpressionWithTypeArguments, ast.KindDecorator,
			ast.KindPostfixUnaryExpression, ast.KindTypeOfExpression:
			return false
		case ast.KindVariableDeclaration:
			vd := parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == cur && vd.Name() != nil &&
				(vd.Name().Kind == ast.KindObjectBindingPattern || vd.Name().Kind == ast.KindArrayBindingPattern) {
				return false
			}
			return true
		case ast.KindConditionalExpression:
			cond := parent.AsConditionalExpression()
			if cond == nil {
				return true
			}
			if cond.Condition == cur {
				return false
			}
			if cond.WhenTrue == cur || cond.WhenFalse == cur {
				cur = outermostTransparentExpression(parent)
				continue
			}
			return true
		case ast.KindIfStatement:
			if stmt := parent.AsIfStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindWhileStatement:
			if stmt := parent.AsWhileStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindDoStatement:
			if stmt := parent.AsDoStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindForStatement:
			if stmt := parent.AsForStatement(); stmt != nil && stmt.Condition == cur {
				return false
			}
			return true
		case ast.KindSwitchStatement:
			if stmt := parent.AsSwitchStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindCaseClause:
			if clause := parent.AsCaseOrDefaultClause(); clause != nil && clause.Expression == cur {
				return false
			}
			return true
		case ast.KindPrefixUnaryExpression:
			if unary := parent.AsPrefixUnaryExpression(); unary != nil && unary.Operator == ast.KindExclamationToken && unary.Operand == cur {
				return false
			}
			return true
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return true
			}
			switch bin.OperatorToken.Kind {
			case ast.KindAsteriskAsteriskToken:
				if bin.Left == cur {
					return false
				}
			case ast.KindInKeyword, ast.KindInstanceOfKeyword,
				ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
				ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
				ast.KindLessThanToken, ast.KindGreaterThanToken,
				ast.KindLessThanEqualsToken, ast.KindGreaterThanEqualsToken:
				return false
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				if bin.Left == cur {
					return false
				}
				if bin.Right == cur {
					cur = outermostTransparentExpression(parent)
					continue
				}
			case ast.KindEqualsToken:
				if bin.Right == cur {
					lhs := ast.SkipParentheses(bin.Left)
					if lhs != nil && (lhs.Kind == ast.KindObjectLiteralExpression || lhs.Kind == ast.KindArrayLiteralExpression) {
						return false
					}
					cur = outermostTransparentExpression(parent)
					continue
				}
			case ast.KindCommaToken,
				ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
				if bin.Right == cur {
					cur = outermostTransparentExpression(parent)
					continue
				}
			}
			return true
		case ast.KindArrayLiteralExpression:
			arrOuter := outermostTransparentExpression(parent)
			if arrOuter != nil && arrOuter.Parent != nil {
				switch arrOuter.Parent.Kind {
				case ast.KindVariableDeclaration:
					vd := arrOuter.Parent.AsVariableDeclaration()
					if vd != nil && vd.Initializer == arrOuter && vd.Name() != nil &&
						(vd.Name().Kind == ast.KindArrayBindingPattern || vd.Name().Kind == ast.KindObjectBindingPattern) {
						return false
					}
				case ast.KindBinaryExpression:
					bin := arrOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == arrOuter {
						lhs := ast.SkipParentheses(bin.Left)
						if lhs != nil && (lhs.Kind == ast.KindArrayLiteralExpression || lhs.Kind == ast.KindObjectLiteralExpression) {
							return false
						}
					}
				case ast.KindCallExpression:
					call := arrOuter.Parent
					if method, ok := promiseMethodName(call); ok && (method == "resolve" || method == "reject") {
						callOuter := outermostTransparentExpression(call)
						if callOuter != nil && callOuter.Parent != nil && callOuter.Parent.Kind == ast.KindAwaitExpression {
							callOuter = outermostTransparentExpression(callOuter.Parent)
						}
						if callOuter == nil || callOuter.Parent == nil || callOuter.Parent.Kind != ast.KindExpressionStatement {
							return false
						}
						if method == "reject" && isInsideTryWithCatch(callOuter.Parent) {
							return false
						}
					}
				}
			}
			return true
		default:
			return true
		}
	}
	return true
}

func isInsideTryWithCatch(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if testFramework.IsFunction(current) {
			return false
		}
		if current.Parent != nil && current.Parent.Kind == ast.KindTryStatement {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt != nil && tryStmt.TryBlock == current && tryStmt.CatchClause != nil {
				return true
			}
		}
	}
	return false
}

func insertAwaitBefore(sourceFile *ast.SourceFile, node *ast.Node) rule.RuleFix {
	text := sourceFile.Text()
	pos := internalUtils.TrimNodeTextRange(sourceFile, node).Pos()
	insertion := "await "
	if needsLeadingSpace(text, pos) {
		insertion = " await "
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), insertion)
}

func wrapAwaitWithParens(sourceFile *ast.SourceFile, node *ast.Node) rule.RuleFix {
	text := sourceFile.Text()
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	prefix := "await ("
	if needsLeadingSpace(text, r.Pos()) {
		prefix = " await ("
	}
	return rule.RuleFixReplaceRange(r, prefix+text[r.Pos():r.End()]+")")
}

// AwaitFix awaits an async assertion, replacing a terminal `return` in a void
// callback with `await` when returning it is not accepted. Early returns and
// value-returning helpers keep `return await` to preserve control flow and
// return values.
func AwaitFix(sourceFile *ast.SourceFile, node *ast.Node, alwaysAwait bool) rule.RuleFix {
	outer := outermostTransparentExpression(node)
	if alwaysAwait && outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement {
		ret := outer.Parent
		retRange := internalUtils.TrimNodeTextRange(sourceFile, ret)
		outerRange := internalUtils.TrimNodeTextRange(sourceFile, outer)
		if isSafeTerminalReturnStrip(ret, outer) {
			text := sourceFile.Text()
			retKeywordEnd := retRange.Pos() + len("return")
			between := ""
			if retKeywordEnd <= outerRange.Pos() && outerRange.Pos() <= len(text) {
				between = text[retKeywordEnd:outerRange.Pos()]
			}
			if between == "" {
				between = " "
			}
			if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
				return rule.RuleFixReplaceRange(
					core.NewTextRange(retRange.Pos(), outerRange.End()),
					"await"+between+"("+text[outerRange.Pos():outerRange.End()]+")",
				)
			}
			return rule.RuleFixReplaceRange(core.NewTextRange(retRange.Pos(), outerRange.Pos()), "await"+between)
		}
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return wrapAwaitWithParens(sourceFile, outer)
		}
		return insertAwaitBefore(sourceFile, outer)
	}
	// If the assertion is wrapped in a TypeScript assertion (`as`, `<T>`,
	// `satisfies`, `!`), await the outermost wrapper so `await` binds to the
	// cast Promise rather than splitting `expect(...)` from `as Promise<void>`.
	// Plain parentheses keep inserting before `node` to preserve `(await ...)`.
	if outer != internalUtils.OutermostParenthesizedExpression(node) {
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return wrapAwaitWithParens(sourceFile, outer)
		}
		return insertAwaitBefore(sourceFile, outer)
	}
	return insertAwaitBefore(sourceFile, node)
}

// ResolveAsyncAssertionReportNode decides where an async assertion is reported
// and whether it has to be reported at all. assertionNode must be the outermost
// expression of the assertion chain: for a Chai chain carrying several matchers,
// such as expect(p).resolves.to.be.a("string").that.contains("x"), an inner
// matcher's call would be inspected against the wrong parent.
//
// reportNode is the node the report and the await fix apply to. promiseWrapped
// tells the caller to use the Promise-specific message, and insideAssertionArray
// marks a Promise.all([...]) array whose assertions report only once.
func ResolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	if assertionNode == nil {
		return nil, false, false, false
	}

	reportNode = assertionNode
	// Walk outward to a fixed point across `.then`/`.catch`/`.finally` chains
	// and enclosing `Promise.*` wrappers (including `Promise.all([...]).then(...)`
	// and nested `Promise.all([Promise.resolve(...)])`).
	for {
		chained := getParentIfPromiseChained(reportNode)
		outerChained := outermostTransparentExpression(chained)
		if promiseCallNode := findPromiseCallExpressionNode(chained); promiseCallNode != nil {
			reportNode = promiseCallNode
			promiseWrapped = true
			if outerChained != nil && outerChained.Parent != nil && outerChained.Parent.Kind == ast.KindArrayLiteralExpression {
				insideAssertionArray = true
			}
			continue
		}
		if chained != reportNode {
			reportNode = chained
			continue
		}
		break
	}

	// Parentheses and TypeScript assertion wrappers around the assertion are
	// erased at runtime, so the await or return that consumes the Promise may
	// sit above one or more wrappers. reportNode itself stays the report range.
	outermost := outermostTransparentExpression(reportNode)
	if outermost.Parent == nil || isAcceptableReturnNode(outermost, outermost.Parent, !alwaysAwait) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	return reportNode, promiseWrapped, insideAssertionArray, true
}
