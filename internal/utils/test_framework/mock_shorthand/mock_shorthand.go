// Package mock_shorthand holds the analysis shared by the test-framework rules
// that replace a mock configuration call with one of its shorthands:
// `mockImplementation(() => value)` with `mockReturnValue(value)`, and
// `Promise.resolve(value)` / `Promise.reject(value)` passed to either of them with
// `mockResolvedValue(value)` / `mockRejectedValue(value)`.
//
// Every shorthand moves a value out of a callback that runs on each call and into
// an argument evaluated once, when the mock is configured. Most of what lives here
// decides whether that move preserves what the mock did; the rest renders the
// rewrite without losing what the author wrote around it.
//
// Both rules match on method shape alone: any `<receiver>.mockImplementation(...)`
// qualifies, with no attempt to prove the receiver is a mock created by the
// framework. Both reference plugins make the same choice, and narrowing it would
// silently drop user-defined mock-like objects.
package mock_shorthand

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const onceSuffix = "Once"

// WithOnce appends the `Once` suffix that selects the single-call variant of a
// mock method.
func WithOnce(name string, once bool) string {
	if once {
		return name + onceSuffix
	}
	return name
}

// CalledMethod returns the accessor node naming the method a call reaches, the
// name it spells, and whether that name ends in `Once`, for the dotted, quoted
// and template forms alike.
//
// A computed identifier key (`mock[name](...)`) names a variable rather than a
// method, so it is rejected here rather than reported without a fix: the rule
// cannot know which method the call reaches at all.
func CalledMethod(callee *ast.Node) (accessor *ast.Node, name string, once bool, ok bool) {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return nil, "", false, false
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		accessor = callee.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		accessor = ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
	default:
		return nil, "", false, false
	}

	if accessor == nil || !testFramework.IsAccessorNode(accessor) ||
		testFramework.IsComputedIdentifierAccessor(accessor) {
		return nil, "", false, false
	}

	name, ok = utils.AccessExpressionStaticName(callee)
	if !ok {
		return nil, "", false, false
	}
	return accessor, name, strings.HasSuffix(name, onceSuffix), true
}

// IsCollapsibleCallback reports whether the callback can be replaced by the value
// it returns.
//
// A callback that declares parameters cannot: it reads what the mock was called
// with. That covers a TypeScript `this` parameter too, which the shorthand has no
// way to bind.
//
// A callback that declares type parameters — `<T>(): T => null as T` — has no
// parameters and a single return, but the rewrite drops the `<T>` declaration
// and leaves `T` dangling in the argument. And a generator callback does not
// return what its `return` names: calling it produces an iterator, so rewriting
// `function* () { return 1; }` to `mockReturnValue(1)` changes what every caller
// of the mock receives. Both reference plugins perform both rewrites.
//
// Whether an `async` callback qualifies depends on the shorthand, so the caller
// decides.
func IsCollapsibleCallback(callback *ast.Node, allowAsync bool) bool {
	if callback == nil || !ast.IsFunctionExpressionOrArrowFunction(callback) {
		return false
	}
	if callback.Kind == ast.KindFunctionExpression &&
		callback.AsFunctionExpression().AsteriskToken != nil {
		return false
	}
	return len(callback.Parameters()) == 0 &&
		len(callback.TypeParameters()) == 0 &&
		(allowAsync || !ast.IsAsyncFunction(callback))
}

// SingleReturnExpression returns the expression a callback hands back, for both
// the concise-body and the `{ return … }` forms.
//
// Upstream only looks at the first statement of a block body, so a block that
// starts with anything else — including a block whose `return` comes after other
// work — is left alone. That is deliberate: those callbacks do something on every
// call, which is exactly what a shorthand cannot express.
func SingleReturnExpression(fn *ast.Node) *ast.Node {
	body := fn.Body()
	if body == nil {
		return nil
	}
	if body.Kind != ast.KindBlock {
		return body
	}

	statements := body.AsBlock().Statements
	if statements == nil || len(statements.Nodes) == 0 {
		return nil
	}
	first := statements.Nodes[0]
	if first.Kind != ast.KindReturnStatement {
		return nil
	}
	return first.AsReturnStatement().Expression
}

// IsEvaluatedOnceSafe reports whether collapsing the callback into a value that
// is evaluated once, at the point the mock is configured, preserves what the
// callback did on every call.
//
// Two things break that equivalence, and both are checked over the whole
// expression rather than only its outermost node:
//
//   - a write (`value++`, `n += 1`, `obj.x = 1`) would run once instead of once
//     per call. The reference plugins guard at most a top-level update
//     expression, so `() => (n += 1)` is rewritten and a mock that counted 1, 2,
//     3 across three calls returns 1 every time. The guard is widened here to
//     every write in the expression.
//   - a read of a binding that can be reassigned would freeze the value the
//     binding happened to hold at configuration time.
//
// Writes are not looked for inside a nested function body, because that body does
// not run while the expression is evaluated — but they are in the parts of a
// class or method that do, such as a static initializer or a computed name.
// Reads are looked for everywhere, because a closure the expression hands back
// still observes the binding later.
func IsEvaluatedOnceSafe(ctx rule.RuleContext, expression *ast.Node) bool {
	return !containsWrite(expression) && !readsMutableBinding(ctx, expression)
}

func containsWrite(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindPrefixUnaryExpression:
		if isIncrementOrDecrement(node.AsPrefixUnaryExpression().Operator) {
			return true
		}
	case ast.KindPostfixUnaryExpression:
		if isIncrementOrDecrement(node.AsPostfixUnaryExpression().Operator) {
			return true
		}
	case ast.KindBinaryExpression:
		if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			return true
		}
	case ast.KindDeleteExpression:
		return true
	}

	if testFramework.IsFunction(node) || ast.IsClassLike(node) {
		for _, part := range creationTimeParts(node, true) {
			if containsWrite(part) {
				return true
			}
		}
		return false
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsWrite(child) {
			found = true
			return true
		}
		return false
	})
	return found
}

func isIncrementOrDecrement(operator ast.Kind) bool {
	return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
}

func readsMutableBinding(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Refs == nil || node == nil {
		return false
	}

	found := false
	var visit func(*ast.Node) bool
	visit = func(child *ast.Node) bool {
		if child.Kind == ast.KindIdentifier {
			if scope.IsReferenceIdentifier(child) && !scope.InTypePosition(child) &&
				isMutableBinding(ctx.Refs.Resolve(child), ctx.SourceFile) {
				found = true
				return true
			}
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	return found
}

// isMutableBinding reports whether a resolved symbol was declared in this file
// with `let` or `var` — the only declarations both reference plugins treat as
// unsafe.
//
// Everything else stays fixable, including a parameter and an imported binding.
// Both can be assigned to in principle, and upstream reports them anyway: a
// test that reassigns an import it also mocks is already beyond what these rules
// can reason about, and refusing parameters would drop the common
// `makeMock(value) => rs.fn().mockImplementation(() => value)` shape.
//
// The declaration has to be in this file. Every standard-library global is
// declared `declare var` — `Promise`, `Error`, `Array` — so a symbol from a
// `.d.ts` or another file would otherwise make `() => Promise.resolve(42)`
// unfixable, when upstream's scope manager never resolves those to a definition
// at all. An identifier that resolves to nothing is treated as immutable for the
// same reason.
func isMutableBinding(symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if ast.GetSourceFileOfNode(declaration) != sourceFile {
			continue
		}
		if declarationList := enclosingVariableDeclarationList(declaration); declarationList != nil &&
			declarationList.Flags&ast.NodeFlagsConst == 0 {
			return true
		}
	}
	return false
}

// enclosingVariableDeclarationList walks out of a binding pattern to the
// declaration list that carries the `const` / `let` / `var` flag, and returns nil
// for a declaration that is not a variable at all — a parameter, a function, a
// class or an import.
func enclosingVariableDeclarationList(declaration *ast.Node) *ast.Node {
	node := declaration
	for node != nil {
		switch node.Kind {
		case ast.KindVariableDeclarationList:
			return node
		case ast.KindBindingElement, ast.KindObjectBindingPattern,
			ast.KindArrayBindingPattern, ast.KindVariableDeclaration:
			node = node.Parent
		default:
			return nil
		}
	}
	return nil
}

// ContainsAwait reports whether evaluating the expression awaits something, which
// only an `async` callback can do. Lifted out of the callback, the `await` would
// either fail to parse or turn into a top-level await that runs once, while the
// module loads.
//
// A nested function body does not run while the expression is evaluated, and its
// own `await` belongs to it, so only the parts of a function or class evaluated
// with the expression are searched.
func ContainsAwait(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindAwaitExpression {
		return true
	}
	if testFramework.IsFunction(node) || ast.IsClassLike(node) {
		for _, part := range creationTimeParts(node, false) {
			if ContainsAwait(part) {
				return true
			}
		}
		return false
	}
	return node.ForEachChild(ContainsAwait)
}

// ReadsCallSiteBinding reports whether a `function` callback's returned
// expression reads something the call site binds rather than the surrounding
// code: `this`, `arguments`, or `new.target`.
//
// A `function` mock implementation is invoked with the mock's receiver and
// arguments, and a mock can be called with `new`, so
// `function () { return this.id; }` returns a different value per call. Lifted
// into a shorthand, the same `this`, `arguments` and `new.target` name whatever
// encloses the configuration site, which is a different value or a syntax error
// — `new.target` outside any function does not parse at all. Both reference
// plugins perform that rewrite.
//
// An arrow callback is not affected: it has no `this` or `arguments` of its own,
// so both already name the enclosing scope. Nested arrows inside a `function`
// callback do inherit its bindings and are therefore searched; a nested
// `function` rebinds them and is not. A nested method or class rebinds them only
// in its bodies, so its computed names, decorators and `extends` expression are
// still searched.
func ReadsCallSiteBinding(callback *ast.Node, expression *ast.Node) bool {
	if callback.Kind != ast.KindFunctionExpression {
		return false
	}

	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindThisKeyword:
			return true
		case ast.KindIdentifier:
			return node.Text() == "arguments" && scope.IsReferenceIdentifier(node)
		case ast.KindMetaProperty:
			// `import.meta` is the other meta property, and it names the module
			// rather than the call, so it travels with the expression.
			return node.AsMetaProperty().KeywordToken == ast.KindNewKeyword
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration:
			return false
		}
		if ast.IsMethodOrAccessor(node) || ast.IsClassLike(node) {
			for _, part := range creationTimeParts(node, false) {
				if visit(part) {
					return true
				}
			}
			return false
		}
		return node.ForEachChild(visit)
	}
	return visit(expression)
}

// creationTimeParts returns the parts of a method or class that run when the node
// itself is evaluated rather than when the method is called or the class is
// instantiated: computed member names, decorators and the `extends` expression.
// Those parts see the enclosing function's `this`, `arguments` and
// `new.target`.
//
// With staticBodies, it also returns static field initializers and static
// blocks. They run at the same moment, but bind `this` to the class itself, so
// they matter to a search for writes and not to one for call-site bindings.
func creationTimeParts(node *ast.Node, staticBodies bool) []*ast.Node {
	var parts []*ast.Node
	addEvaluatedHeader := func(member *ast.Node) {
		if modifiers := member.Modifiers(); modifiers != nil {
			for _, modifier := range modifiers.Nodes {
				if ast.IsDecorator(modifier) {
					parts = append(parts, modifier)
				}
			}
		}
		if name := member.Name(); name != nil && ast.IsComputedPropertyName(name) {
			parts = append(parts, name)
		}
	}

	switch {
	case ast.IsMethodOrAccessor(node):
		addEvaluatedHeader(node)
	case ast.IsClassLike(node):
		addEvaluatedHeader(node)
		if heritage := ast.GetClassExtendsHeritageElement(node); heritage != nil {
			parts = append(parts, heritage)
		}
		for _, member := range node.Members() {
			addEvaluatedHeader(member)
			if !staticBodies || !ast.IsStatic(member) {
				continue
			}
			switch member.Kind {
			case ast.KindClassStaticBlockDeclaration:
				parts = append(parts, member.AsClassStaticBlockDeclaration().Body)
			case ast.KindPropertyDeclaration:
				if initializer := member.Initializer(); initializer != nil {
					parts = append(parts, initializer)
				}
			}
		}
	}
	return parts
}

// PromiseFactoryCall recognizes `Promise.resolve(...)` and `Promise.reject(...)`
// and returns the call and the method name.
//
// `Promise.resolve`, `Promise['resolve']`, the template-literal key and
// `Promise?.resolve` are the same call. Type assertions and instantiation
// expressions are erased at run time, so `Promise.reject(e) as Promise<never>`,
// `(Promise as any).reject(e)` and `(Promise.reject<never>)(e)` build the same
// promise too. wrapped reports whether such a wrapper sits around the call itself,
// which a rewrite that keeps only the call's argument would drop.
//
// A locally declared `Promise` is some other object whose methods these rules know
// nothing about, so it is not recognized.
func PromiseFactoryCall(ctx rule.RuleContext, expression *ast.Node) (call *ast.Node, method string, wrapped bool, ok bool) {
	call = ast.SkipOuterExpressions(expression, ast.OEKAll)
	if call == nil || call.Kind != ast.KindCallExpression {
		return nil, "", false, false
	}

	access := ast.SkipOuterExpressions(call.AsCallExpression().Expression, ast.OEKAll)
	if access == nil || !ast.IsAccessExpression(access) {
		return nil, "", false, false
	}
	method, ok = utils.AccessExpressionStaticName(access)
	if !ok || (method != "resolve" && method != "reject") {
		return nil, "", false, false
	}

	object := ast.SkipOuterExpressions(access.Expression(), ast.OEKAll)
	if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "Promise" {
		return nil, "", false, false
	}
	if ctx.Refs != nil && utils.IsValueSymbolDeclaredInFile(ctx.Refs.Resolve(object), ctx.SourceFile) {
		return nil, "", false, false
	}
	return call, method, ast.SkipParentheses(expression) != call, true
}

// DropsCallbackScope reports whether replacing the callback with an expression
// taken from its first `return` would delete a declaration the program still
// relies on.
//
// The return statement is the callback's first statement, so the only things
// the callback declares outside the expression are what follows the return —
// a hoisted function or `var` the expression may call or read — and, for a
// named function expression, its own name. Either is gone once the callback is,
// so `function impl() { return impl; }` would become a reference to an
// undeclared `impl`.
func DropsCallbackScope(ctx rule.RuleContext, callback *ast.Node, expression *ast.Node) bool {
	if body := callback.Body(); body != nil && body.Kind == ast.KindBlock &&
		body.AsBlock().Statements != nil && len(body.AsBlock().Statements.Nodes) > 1 {
		return true
	}

	name := callback.Name()
	if name == nil || expression == nil {
		return false
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && node.Text() == name.Text() &&
			scope.IsReferenceIdentifier(node) && resolvesTo(ctx, node, callback) {
			return true
		}
		return node.ForEachChild(visit)
	}
	return visit(expression)
}

// resolvesTo reports whether the identifier names the given declaration. Without
// a reference index it assumes it does, which only withholds a fix.
func resolvesTo(ctx rule.RuleContext, identifier *ast.Node, declaration *ast.Node) bool {
	if ctx.Refs == nil {
		return true
	}
	symbol := ctx.Refs.Resolve(identifier)
	return symbol != nil && slices.Contains(symbol.Declarations, declaration)
}

// DropsComment reports whether replacing replaced with the text of kept would
// delete a comment the author wrote — around a callback's `=>`, before its
// `return`, inside a `Promise.resolve(` wrapper, or after the kept expression. The
// comment is not part of the kept text, so the rewrite would silently discard it.
// With nothing kept, any comment inside replaced counts.
func DropsComment(ctx rule.RuleContext, replaced *ast.Node, kept *ast.Node) bool {
	replacedRange := utils.TrimNodeTextRange(ctx.SourceFile, replaced)
	if kept == nil {
		return utils.HasCommentInSpan(ctx.Comments.All(), replacedRange.Pos(), replacedRange.End())
	}
	keptRange := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(kept))
	return utils.HasCommentInSpan(ctx.Comments.All(), replacedRange.Pos(), keptRange.Pos()) ||
		utils.HasCommentInSpan(ctx.Comments.All(), keptRange.End(), replacedRange.End())
}

// ArgumentText renders the expression as a shorthand's argument.
//
// Outer parentheses — the ones in `() => ({ a: 1 })` — are dropped, matching
// upstream, whose AST does not record them in the first place. A comma
// expression keeps one pair: without it `mockReturnValue((0, value))` would
// collapse into two arguments and return the wrong operand, which is what both
// reference plugins currently emit.
func ArgumentText(ctx rule.RuleContext, expression *ast.Node) string {
	inner := ast.SkipParentheses(expression)
	if inner == nil {
		inner = expression
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, inner)
	if inner.Kind == ast.KindBinaryExpression &&
		inner.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		return "(" + text + ")"
	}
	return text
}
