package require_yield

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMissingYieldMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingYield",
		Description: "This generator function does not have 'yield'.",
	}
}

func isGenerator(node *ast.Node) bool {
	return ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator != 0
}

func hasNonEmptyBody(node *ast.Node) bool {
	body := node.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}
	block := body.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}
	return len(block.Statements.Nodes) > 0
}

// bodyLikeRange returns the [pos, end) of the "execution scope" of a
// scope-bearing node. Yield expressions whose position falls inside this
// range are attributed to this scope; yields outside it (e.g. in a
// computed property key or a decorator expression) pass through to an
// outer scope.
//
// Required because tsgo performs error-recovery parsing: an illegally
// placed `yield` (e.g. inside a non-generator function body or in a
// parameter default value of a non-generator) still produces a
// KindYieldExpression AST node. Without position-aware attribution, such
// a yield would bubble up to the nearest enclosing generator on the
// stack and cause a false negative.
//
// For function-like nodes the scope starts at the parameter list, so
// that yield in parameter default values is attributed here rather than
// leaking to an outer generator.
func bodyLikeRange(node *ast.Node) (int, int, bool) {
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindMethodDeclaration,
		ast.KindArrowFunction,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor:
		body := node.Body()
		if body == nil {
			return 0, 0, false
		}
		pos := body.Pos()
		if params := node.ParameterList(); params != nil {
			pos = params.Loc.Pos()
		}
		return pos, body.End(), true
	case ast.KindPropertyDeclaration:
		init := node.AsPropertyDeclaration().Initializer
		if init == nil {
			return 0, 0, false
		}
		return init.Pos(), init.End(), true
	case ast.KindClassStaticBlockDeclaration:
		body := node.AsClassStaticBlockDeclaration().Body
		if body == nil {
			return 0, 0, false
		}
		return body.Pos(), body.End(), true
	}
	return 0, 0, false
}

type stackFrame struct {
	node  *ast.Node
	count int
}

// https://eslint.org/docs/latest/rules/require-yield
var RequireYieldRule = rule.Rule{
	Name: "require-yield",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		stack := make([]stackFrame, 0, 8)

		enter := func(node *ast.Node) {
			stack = append(stack, stackFrame{node: node})
		}

		exit := func(node *ast.Node) {
			n := len(stack)
			if n == 0 {
				return
			}
			top := stack[n-1]
			stack = stack[:n-1]
			// PropertyDeclaration initializer and ClassStaticBlockDeclaration
			// both represent implicit-constructor-like scopes that are never
			// themselves generators.
			if top.node.Kind == ast.KindPropertyDeclaration ||
				top.node.Kind == ast.KindClassStaticBlockDeclaration {
				return
			}
			if isGenerator(top.node) && top.count == 0 && hasNonEmptyBody(top.node) {
				ctx.ReportRange(
					utils.GetFunctionHeadLoc(ctx.SourceFile, top.node),
					buildMissingYieldMessage(),
				)
			}
		}

		countYield := func(node *ast.Node) {
			for i := len(stack) - 1; i >= 0; i-- {
				bp, be, ok := bodyLikeRange(stack[i].node)
				if !ok {
					continue
				}
				if node.Pos() >= bp && node.End() <= be {
					stack[i].count++
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              enter,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         exit,
			ast.KindFunctionExpression:                               enter,
			rule.ListenerOnExit(ast.KindFunctionExpression):          exit,
			ast.KindMethodDeclaration:                                enter,
			rule.ListenerOnExit(ast.KindMethodDeclaration):           exit,
			ast.KindArrowFunction:                                    enter,
			rule.ListenerOnExit(ast.KindArrowFunction):               exit,
			ast.KindGetAccessor:                                      enter,
			rule.ListenerOnExit(ast.KindGetAccessor):                 exit,
			ast.KindSetAccessor:                                      enter,
			rule.ListenerOnExit(ast.KindSetAccessor):                 exit,
			ast.KindConstructor:                                      enter,
			rule.ListenerOnExit(ast.KindConstructor):                 exit,
			ast.KindPropertyDeclaration:                              enter,
			rule.ListenerOnExit(ast.KindPropertyDeclaration):         exit,
			ast.KindClassStaticBlockDeclaration:                      enter,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): exit,

			ast.KindYieldExpression: countYield,
		}
	},
}
