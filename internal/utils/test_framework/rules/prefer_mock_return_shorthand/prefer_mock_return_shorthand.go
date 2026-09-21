// Package prefer_mock_return_shorthand implements the body shared by the
// test-framework rules that prefer `mockReturnValue` over a `mockImplementation`
// whose callback does nothing but return a value.
//
// The rule matches on method shape alone: any `<receiver>.mockImplementation(...)`
// qualifies, with no attempt to prove the receiver is a mock created by the
// framework. Both reference plugins make the same choice, and narrowing it would
// silently drop user-defined mock-like objects, so there is nothing
// framework-specific left to configure beyond the rule's name.
package prefer_mock_return_shorthand

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type Config struct {
	Name string
}

const (
	implementationMethod = "mockImplementation"
	returnValueMethod    = "mockReturnValue"
	onceSuffix           = "Once"
)

func useMockShorthandMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useMockShorthand",
		Description: "Prefer " + replacement,
	}
}

// calledAccessor returns the accessor node naming the method a call reaches and
// the name it spells, for the dotted, quoted and template forms alike.
//
// A computed identifier key (`mock[name](...)`) names a variable rather than a
// method, so it is rejected here rather than reported without a fix: the rule
// cannot know the call is `mockImplementation` at all.
func calledAccessor(callee *ast.Node) (*ast.Node, string, bool) {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return nil, "", false
	}

	var accessor *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		accessor = callee.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		accessor = ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
	default:
		return nil, "", false
	}

	if accessor == nil || !testFramework.IsAccessorNode(accessor) ||
		testFramework.IsComputedIdentifierAccessor(accessor) {
		return nil, "", false
	}

	name, ok := utils.AccessExpressionStaticName(callee)
	if !ok {
		return nil, "", false
	}
	return accessor, name, true
}

// singleReturnExpression returns the expression a zero-argument callback hands
// back, for both the concise-body and the `{ return … }` forms.
//
// Upstream only looks at the first statement of a block body, so a block that
// starts with anything else — including a block whose `return` comes after other
// work — is left alone. That is deliberate: those callbacks do something on every
// call, which is exactly what the shorthand cannot express.
func singleReturnExpression(fn *ast.Node) *ast.Node {
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

// isEvaluatedOnceSafe reports whether collapsing the callback into a value that
// is evaluated once, at the point the mock is configured, preserves what the
// callback did on every call.
//
// Two things break that equivalence, and both are checked over the whole
// expression rather than only its outermost node:
//
//   - a write (`value++`, `n += 1`, `obj.x = 1`) would run once instead of once
//     per call. Both reference plugins guard only a top-level update expression,
//     so `() => (n += 1)` is rewritten and a mock that counted 1, 2, 3 across
//     three calls returns 1 every time. The guard is widened here to every write
//     in the expression.
//   - a read of a binding that can be reassigned would freeze the value the
//     binding happened to hold at configuration time.
//
// Writes are not looked for inside a nested function body, because that body does
// not run while the expression is evaluated. Reads are, because a closure the
// expression hands back still observes the binding later.
func isEvaluatedOnceSafe(ctx rule.RuleContext, expression *ast.Node) bool {
	return !containsWrite(expression) && !readsMutableBinding(ctx, expression)
}

// buildsRejectedPromise reports whether the expression constructs a rejected
// promise, which the shorthand would construct once, when the mock is configured,
// instead of once per call.
//
// A rejected promise nobody has awaited yet is an unhandled rejection. While the
// callback builds it, the promise only exists after the mock is called, and the
// call site that triggered it is also what awaits it. Moved into
// `mockReturnValue`, it exists from the moment the mock is set up, so a mock that
// ends up not being called — the mock configured in `beforeEach` for a branch this
// test does not take, say — leaves a rejection nobody handles, which a test runner
// reports as a run-level error even though every test passed.
//
// Both reference plugins rewrite this shape anyway, and one of them documents
// `mockReturnValue(Promise.reject(...))` as an acceptable result. It is not
// reported here at all, rather than reported without a fix,
// because the advice itself is the problem: the safe rewrite of a rejecting mock
// is `mockRejectedValue`, which builds the promise per call, and that belongs to
// the promise-shorthand rule rather than this one.
func buildsRejectedPromise(ctx rule.RuleContext, expression *ast.Node) bool {
	call := ast.SkipParentheses(expression)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}

	access := ast.SkipParentheses(call.AsCallExpression().Expression)
	if access == nil || !ast.IsAccessExpression(access) {
		return false
	}
	// `Promise.reject`, `Promise['reject']` and `` Promise[`reject`] `` are the
	// same call. The rest of this rule already treats the three accessor
	// spellings as one, and a guard that only recognized the dotted form would
	// leave the other two rewritten.
	if name, ok := utils.AccessExpressionStaticName(access); !ok || name != "reject" {
		return false
	}

	object := ast.SkipParentheses(access.Expression())
	if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "Promise" {
		return false
	}
	// A locally declared `Promise` is some other object whose `reject` this rule
	// knows nothing about.
	return ctx.Refs == nil ||
		!utils.IsValueSymbolDeclaredInFile(ctx.Refs.Resolve(object), ctx.SourceFile)
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
// test that reassigns an import it also mocks is already beyond what this rule
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

// replacementArgumentText renders the expression as the shorthand's argument.
//
// The outer parentheses of `() => ({ a: 1 })` are dropped, matching upstream,
// whose AST does not record them in the first place. A comma expression keeps
// one pair: without it `mockReturnValue((0, value))` would collapse into two
// arguments and return the wrong operand, which is what both reference plugins
// currently emit.
func replacementArgumentText(ctx rule.RuleContext, expression *ast.Node) string {
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

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					arguments := node.Arguments()
					if len(arguments) == 0 {
						return
					}

					accessor, name, ok := calledAccessor(node.AsCallExpression().Expression)
					if !ok {
						return
					}

					once := strings.HasSuffix(name, onceSuffix)
					if name != withOnce(implementationMethod, once) {
						return
					}

					callback := ast.SkipParentheses(arguments[0])
					if !isCollapsibleCallback(callback) {
						return
					}

					expression := singleReturnExpression(callback)
					if expression == nil || !isEvaluatedOnceSafe(ctx, expression) ||
						buildsRejectedPromise(ctx, expression) ||
						readsCallSiteBinding(callback, expression) {
						return
					}

					replacement := withOnce(returnValueMethod, once)
					ctx.ReportNodeWithDeferredFixes(accessor, useMockShorthandMessage(replacement), func() []rule.RuleFix {
						accessorRange, accessorText, ok := testFramework.AccessorReplacement(ctx.SourceFile, accessor, replacement)
						if !ok || dropsComment(ctx, callback, expression) {
							return nil
						}
						return []rule.RuleFix{
							rule.RuleFixReplaceRange(accessorRange, accessorText),
							rule.RuleFixReplace(ctx.SourceFile, callback, replacementArgumentText(ctx, expression)),
						}
					})
				},
			}
		},
	}
}

// isCollapsibleCallback reports whether the callback can be replaced by the value
// it returns.
//
// A callback that declares parameters cannot: it reads what the mock was called
// with. That covers a TypeScript `this` parameter too, which the shorthand has no
// way to bind. An `async` callback cannot either, since its result is a promise
// the shorthand would not create — `mockResolvedValue` is that rewrite, and it
// belongs to a different rule.
//
// Two more exclusions sit on top of what upstream checks. A callback that
// declares type parameters — `<T>(): T => null as T` — has no parameters and a
// single return, but the rewrite drops the `<T>` declaration and leaves `T`
// dangling in the argument. And a generator callback does not return what its
// `return` names: calling it produces an iterator, so rewriting
// `function* () { return 1; }` to `mockReturnValue(1)` changes what every caller
// of the mock receives. Both reference plugins perform that rewrite.
func isCollapsibleCallback(callback *ast.Node) bool {
	if callback == nil || !ast.IsFunctionExpressionOrArrowFunction(callback) {
		return false
	}
	if callback.Kind == ast.KindFunctionExpression &&
		callback.AsFunctionExpression().AsteriskToken != nil {
		return false
	}
	return len(callback.Parameters()) == 0 &&
		len(callback.TypeParameters()) == 0 &&
		!ast.IsAsyncFunction(callback)
}

// readsCallSiteBinding reports whether a `function` callback's returned
// expression reads something the call site binds rather than the surrounding
// code: `this`, `arguments`, or `new.target`.
//
// A `function` mock implementation is invoked with the mock's receiver and
// arguments, and a mock can be called with `new`, so
// `function () { return this.id; }` returns a different value per call. Lifted
// into `mockReturnValue`, the same `this`, `arguments` and `new.target` name
// whatever encloses the configuration site, which is a different value or a
// syntax error — `new.target` outside any function does not parse at all. Both
// reference plugins perform that rewrite.
//
// An arrow callback is not affected: it has no `this` or `arguments` of its own,
// so both already name the enclosing scope. Nested arrows inside a `function`
// callback do inherit its bindings and are therefore searched; a nested
// `function` rebinds them and is not.
func readsCallSiteBinding(callback *ast.Node, expression *ast.Node) bool {
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
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
		if ast.IsClassLike(node) {
			return false
		}
		return node.ForEachChild(visit)
	}
	return visit(expression)
}

// dropsComment reports whether replacing the callback with its returned
// expression would delete a comment the author wrote inside the callback —
// around the `=>`, before the `return`, or after the expression. The comment is
// not part of the expression's text, so the rewrite would silently discard it.
// The diagnostic still stands; only the fix is withheld.
func dropsComment(ctx rule.RuleContext, callback *ast.Node, expression *ast.Node) bool {
	callbackRange := utils.TrimNodeTextRange(ctx.SourceFile, callback)
	expressionRange := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(expression))
	return utils.HasCommentInSpan(ctx.Comments.All(), callbackRange.Pos(), expressionRange.Pos()) ||
		utils.HasCommentInSpan(ctx.Comments.All(), expressionRange.End(), callbackRange.End())
}

func withOnce(name string, once bool) string {
	if once {
		return name + onceSuffix
	}
	return name
}
