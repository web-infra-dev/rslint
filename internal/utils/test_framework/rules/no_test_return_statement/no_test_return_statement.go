// Package no_test_return_statement holds the framework-neutral checks shared
// by jest/no-test-return-statement and rstest/no-test-return-statement.
//
// The rule reports the first `return` statement written directly in the block
// body of a test callback. Returns nested in other statements are left alone:
// an early `if (!supported) return;` guard is a common, deliberate way to end a
// test, and a nested function's return belongs to that function. A concise
// arrow body is not a return statement and is never reported.
//
// A callback passed by name is reported only while the name is used for
// nothing but test callbacks. A function that is also called directly,
// exported or otherwise referenced may rely on its return value in those other
// uses, so removing the return would break working code.
//
// Everything a framework decides lives in the injected Runtime: which calls
// register a test and which function runs as that registration's callback.
package no_test_return_statement

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

type Runtime struct {
	// TestCallback returns the function a call runs as a test callback: a
	// function literal among its arguments, or the local function a callback
	// identifier is bound to. It returns nil when the call is not a test
	// registration or when the callback cannot be resolved to one function.
	TestCallback func(call *ast.Node) *ast.Node
}

type Config struct {
	Name    string
	Message rule.RuleMessage
	// Prepare creates the framework adapter once per file.
	Prepare func(ctx rule.RuleContext) Runtime
}

// outermostWrapper climbs from node through parentheses and TypeScript
// assertions, which leave the wrapped value unchanged at runtime.
func outermostWrapper(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		parent := node.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			node = parent
			continue
		}
		if _, ok := internalUtils.TransparentExpression(parent); ok {
			node = parent
			continue
		}
		break
	}
	return node
}

// isArgumentOf reports whether node, through wrappers, is an argument of call.
func isArgumentOf(node *ast.Node, call *ast.Node) bool {
	wrapper := outermostWrapper(node)
	if wrapper.Parent != call || call.Kind != ast.KindCallExpression {
		return false
	}
	arguments := call.AsCallExpression().Arguments
	return arguments != nil && slices.Contains(arguments.Nodes, wrapper)
}

// callbackDeclaration returns the declaration binding a named callback: the
// function declaration itself, or the variable declaration a function
// expression initializes. It returns nil for any other function.
func callbackDeclaration(function *ast.Node) *ast.Node {
	if function.Kind == ast.KindFunctionDeclaration {
		return function
	}
	container := outermostWrapper(function).Parent
	if container != nil && container.Kind == ast.KindVariableDeclaration {
		return container
	}
	return nil
}

// firstDirectReturn returns the first return statement of a block body,
// without descending into nested statements or functions.
func firstDirectReturn(function *ast.Node) *ast.Node {
	body := function.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	statements := body.AsBlock().Statements
	if statements == nil {
		return nil
	}
	for _, statement := range statements.Nodes {
		if statement.Kind == ast.KindReturnStatement {
			return statement
		}
	}
	return nil
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			checked := map[*ast.Node]bool{}

			// onlyTestCallback reports whether every reference to the binding
			// declared by declaration passes it as the callback of a test that
			// resolves to function.
			onlyTestCallback := func(declaration *ast.Node, function *ast.Node) bool {
				if ctx.Refs == nil ||
					ast.GetCombinedModifierFlags(declaration)&(ast.ModifierFlagsExport|ast.ModifierFlagsDefault) != 0 {
					return false
				}
				references := ctx.Refs.References(declaration.Symbol())
				if len(references) == 0 {
					return false
				}
				for _, reference := range references {
					call := outermostWrapper(reference).Parent
					if call == nil || !isArgumentOf(reference, call) ||
						runtime.TestCallback(call) != function {
						return false
					}
				}
				return true
			}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					function := runtime.TestCallback(node)
					if function == nil || checked[function] {
						return
					}
					checked[function] = true

					if !isArgumentOf(function, node) {
						declaration := callbackDeclaration(function)
						if declaration == nil || !onlyTestCallback(declaration, function) {
							return
						}
					}
					if statement := firstDirectReturn(function); statement != nil {
						ctx.ReportNode(statement, config.Message)
					}
				},
			}
		},
	}
}
