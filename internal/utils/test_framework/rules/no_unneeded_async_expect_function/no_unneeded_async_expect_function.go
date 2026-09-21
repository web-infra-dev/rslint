// Package no_unneeded_async_expect_function implements the framework-independent
// body shared by test-framework rules that report an async function wrapper
// handed to expect() when the wrapper's only job is to await one call.
//
// The package owns the wrapper's shape: which functions qualify and which call
// the assertion could receive instead. Deciding that an assertion reaches the
// wrapper at all is framework-specific and belongs to Runtime.ParseExpect, and
// so is the edit, because the plugins disagree on whether unwrapping is safe
// enough to apply automatically; Config.Report owns it.
package no_unneeded_async_expect_function

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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
