// Package no_unneeded_async_expect_function implements the framework-independent
// body shared by test-framework rules that report an async function wrapper
// handed to expect() when the wrapper's only job is to await one call.
//
// The package owns the wrapper's shape: which functions qualify, which call
// the assertion could receive instead, and whether replacing the wrapper with
// that call keeps its meaning. Deciding that an assertion reaches the wrapper
// at all is framework-specific and belongs to Runtime.ParseExpect, and so is
// how the edit is offered, because the plugins disagree on whether unwrapping
// is safe enough to apply automatically; Config.Report owns it.
package no_unneeded_async_expect_function

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Match describes one reportable wrapper.
type Match struct {
	// HeadCall is the expect(...) call that receives the wrapper.
	HeadCall *ast.Node
	// Wrapper is the first argument exactly as written, parentheses included,
	// so replacing its range replaces everything the wrapper contributes.
	Wrapper *ast.Node
	// Awaited is the call expression the wrapper awaits.
	Awaited *ast.Node
}

// Runtime carries the per-file state a plugin needs to recognize its own
// expect calls.
type Runtime struct {
	// ParseExpect returns the expect(...) call whose first argument this rule
	// may report, or nil for every other node. Implementations decide which
	// assertion factories and which modifier chains qualify.
	ParseExpect func(*ast.Node) *ast.Node
}

// Config configures NewRule for one plugin.
type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
	Report  func(rule.RuleContext, Match)
}

// Message is the diagnostic both plugins report.
var Message = rule.RuleMessage{
	Id:          "noAsyncWrapperForExpectedPromise",
	Description: "Avoid wrapping asynchronous expectations in an unnecessary async function.",
}

// IsAsyncFunctionExpression reports whether node is an async function
// expression or async arrow function. ast.IsAsyncFunction already rejects
// async generators, whose call result is an async iterator rather than a
// promise.
func IsAsyncFunctionExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	return node != nil &&
		ast.IsFunctionExpressionOrArrowFunction(node) &&
		ast.IsAsyncFunction(node)
}

func functionBody(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	default:
		return nil
	}
}

// singleStatementExpression returns the expression a function evaluates when
// its body is either a concise arrow body or a block holding exactly one
// expression statement.
func singleStatementExpression(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return body
	}

	block := body.AsBlock()
	if block == nil || block.Statements == nil || len(block.Statements.Nodes) != 1 {
		return nil
	}

	stmt := block.Statements.Nodes[0]
	if stmt == nil || stmt.Kind != ast.KindExpressionStatement {
		return nil
	}

	return stmt.AsExpressionStatement().Expression
}

// AwaitedCall returns the call an async wrapper awaits as its only work, or
// nil when the wrapper does anything else. Everything else in the body — a
// second statement, a declaration, a loop, an array literal around the await,
// a non-call operand — keeps the wrapper meaningful.
func AwaitedCall(fn *ast.Node) *ast.Node {
	expr := singleStatementExpression(functionBody(fn))
	if expr == nil {
		return nil
	}
	expr = ast.SkipParentheses(expr)
	if expr == nil || expr.Kind != ast.KindAwaitExpression {
		return nil
	}

	awaited := expr.AsAwaitExpression().Expression
	if awaited == nil {
		return nil
	}
	awaited = ast.SkipParentheses(awaited)
	if awaited == nil || awaited.Kind != ast.KindCallExpression {
		return nil
	}

	return awaited
}

// keepsWrapperBindings reports whether the wrapper's own bindings survive the
// unwrap.
//
// `rejects` calls the wrapper with no arguments and no receiver, so a
// parameter is always `undefined` inside it, but the name the awaited call
// reads disappears with the wrapper and would resolve to something else — or
// to nothing — at the assertion's own scope. Type parameters, the name of a
// named function expression, `this`, `arguments` and `new.target` are bound
// the same way. Arrow functions take `this`, `arguments` and `new.target` from
// the enclosing scope already, so unwrapping leaves them pointing at the same
// bindings.
func keepsWrapperBindings(fn *ast.Node, awaited *ast.Node) bool {
	if len(fn.Parameters()) != 0 || len(fn.TypeParameters()) != 0 {
		return false
	}
	if fn.Kind != ast.KindFunctionExpression {
		return true
	}
	if fn.Name() != nil {
		return false
	}
	return !referencesCallerBindings(awaited)
}

// referencesCallerBindings reports whether the expression reads `this`,
// `arguments` or `new.target`, which a function expression binds and an
// expression in the assertion's scope does not.
func referencesCallerBindings(node *ast.Node) bool {
	found := false
	var walk func(*ast.Node) bool
	walk = func(child *ast.Node) bool {
		if found || child == nil {
			return true
		}
		switch child.Kind {
		case ast.KindThisKeyword:
			found = true
			return true
		case ast.KindIdentifier:
			if child.Text() == "arguments" {
				found = true
				return true
			}
		case ast.KindMetaProperty:
			if child.AsMetaProperty().KeywordToken == ast.KindNewKeyword {
				found = true
				return true
			}
		}
		return child.ForEachChild(walk)
	}
	walk(node)
	return found
}

// awaitsInWrapper reports whether the awaited call contains an `await` of its
// own, such as `run(await load())`. That `await` belongs to the wrapper, so the
// unwrap would move it into the assertion's scope: a syntax error when that
// scope is not async, and otherwise an await that runs before expect() is
// called, so a rejection it produces escapes the assertion instead of reaching
// `rejects`. An `await` inside a nested function belongs to that function and
// moves with it; only a computed name of such a function is evaluated in the
// wrapper.
func awaitsInWrapper(node *ast.Node) bool {
	found := false
	var walk func(*ast.Node) bool
	walk = func(child *ast.Node) bool {
		if found || child == nil {
			return true
		}
		if child.Kind == ast.KindAwaitExpression {
			found = true
			return true
		}
		if ast.IsFunctionLike(child) {
			if name := child.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				walk(name)
			}
			return false
		}
		return child.ForEachChild(walk)
	}
	walk(node)
	return found
}

// keepsComments reports whether every comment the wrapper carries survives the
// unwrap. Only the awaited call's own text is kept, so a comment written
// anywhere else inside the wrapper would be deleted.
func keepsComments(ctx rule.RuleContext, wrapper *ast.Node, awaited *ast.Node) bool {
	wrapperRange := utils.TrimNodeTextRange(ctx.SourceFile, wrapper)
	awaitedRange := utils.TrimNodeTextRange(ctx.SourceFile, awaited)
	comments := ctx.Comments.All()
	return !utils.HasCommentInSpan(comments, wrapperRange.Pos(), awaitedRange.Pos()) &&
		!utils.HasCommentInSpan(comments, awaitedRange.End(), wrapperRange.End())
}

// UnwrapFix returns the edit that replaces the wrapper with the awaited call,
// or nil when the unwrapped call would not mean the same thing.
func UnwrapFix(ctx rule.RuleContext, match Match) *rule.RuleFix {
	// expect<T>() pins the asserted value's type, and the wrapper is what T
	// describes; the unwrapped call has the awaited value's type instead.
	if match.HeadCall.AsCallExpression().TypeArguments != nil {
		return nil
	}
	fn := ast.SkipParentheses(match.Wrapper)
	if !keepsWrapperBindings(fn, match.Awaited) ||
		awaitsInWrapper(match.Awaited) ||
		!keepsComments(ctx, match.Wrapper, match.Awaited) {
		return nil
	}
	fix := rule.RuleFixReplace(
		ctx.SourceFile,
		match.Wrapper,
		utils.TrimmedNodeText(ctx.SourceFile, match.Awaited),
	)
	return &fix
}

// NewRule builds the rule for one plugin.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					if runtime.ParseExpect == nil || config.Report == nil {
						return
					}
					headCall := runtime.ParseExpect(node)
					if headCall == nil {
						return
					}

					arguments := headCall.Arguments()
					if len(arguments) == 0 || !IsAsyncFunctionExpression(arguments[0]) {
						return
					}
					awaited := AwaitedCall(arguments[0])
					if awaited == nil {
						return
					}

					config.Report(ctx, Match{
						HeadCall: headCall,
						Wrapper:  arguments[0],
						Awaited:  awaited,
					})
				},
			}
		},
	}
}
