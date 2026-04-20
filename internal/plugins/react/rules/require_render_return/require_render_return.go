package require_render_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isFunctionLikeExpression mirrors eslint-plugin-react's astUtil.isFunctionLikeExpression:
// an expression whose value is a function (FunctionExpression or ArrowFunction),
// not a regular function declaration.
func isFunctionLikeExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	}
	return false
}

// memberName returns the Identifier text of a class/object member's key, or "".
//
// Mirrors eslint-plugin-react's astUtil.getPropertyName, which reads only
// `nameNode.name`. Consequently, non-Identifier keys — string literals
// (`"render": fn`), numeric keys, computed keys (`['render']()`,
// `` [`render`]() ``, `[tag`render`]()`), etc. — all map to "" and never
// match `render`. Only a bare Identifier key `render` qualifies.
func memberName(member *ast.Node) string {
	if member == nil {
		return ""
	}
	switch member.Kind {
	case ast.KindMethodDeclaration,
		ast.KindPropertyAssignment,
		ast.KindShorthandPropertyAssignment,
		ast.KindPropertyDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
	default:
		return ""
	}
	name := member.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

// findAllRenderMembers returns every member whose Identifier key is `render`
// and whose value is function-like (accessor/shorthand method / FE / arrow).
//
// Mirrors the *filter* in eslint-plugin-react's findRenderMethod, which does
// not short-circuit on the first match — the rule's Program:exit still
// considers the component "OK" if ANY render-named member has a return
// statement (tracked via `markReturnStatementPresent`). So we must examine
// every qualifying member for returns, not just the first.
func findAllRenderMembers(members []*ast.Node) []*ast.Node {
	var out []*ast.Node
	for _, m := range members {
		if memberName(m) != "render" {
			continue
		}
		switch m.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			out = append(out, m)
		case ast.KindPropertyAssignment:
			if isFunctionLikeExpression(m.AsPropertyAssignment().Initializer) {
				out = append(out, m)
			}
		case ast.KindPropertyDeclaration:
			if isFunctionLikeExpression(m.AsPropertyDeclaration().Initializer) {
				out = append(out, m)
			}
		}
	}
	return out
}

// renderFunctionOf returns the function-like node that carries the render
// body: the MethodDeclaration / accessor itself for class/object shorthand
// methods, or the FunctionExpression / ArrowFunction initializer for
// property-style render.
func renderFunctionOf(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return member
	case ast.KindPropertyAssignment:
		return member.AsPropertyAssignment().Initializer
	case ast.KindPropertyDeclaration:
		return member.AsPropertyDeclaration().Initializer
	}
	return nil
}

// hasReturnInBody walks `body` (a Block) looking for a ReturnStatement at
// "depth ≤ 1" in ESLint's sense — i.e. not inside a nested function-like
// boundary.
//
// eslint-plugin-react's depth regex `/Function(Expression|Declaration)$/`
// is UNANCHORED, so it matches both `FunctionExpression` and
// `ArrowFunctionExpression` (suffix match). Consequently arrows are NOT
// transparent — a return inside a nested arrow inside `render()` does not
// count as render's return. `MethodDefinition` / accessor wrappers don't
// match the regex directly, but in tsgo those kinds are themselves the
// function node (no ESTree FunctionExpression child wrapper), so stopping
// at them is the correct equivalent of ESLint counting the inner
// FunctionExpression.
func hasReturnInBody(body *ast.Node) bool {
	if body == nil {
		return false
	}
	found := false
	var walk ast.Visitor
	walk = func(n *ast.Node) bool {
		if found || n == nil {
			return found
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			found = true
			return true
		case ast.KindFunctionExpression,
			ast.KindFunctionDeclaration,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			// Do not descend: a return inside these would bump ESLint's
			// depth counter past 1 and therefore not mark the outer render.
			return false
		}
		n.ForEachChild(walk)
		return found
	}
	walk(body)
	return found
}

// renderReturns reports whether the render function-like value either has an
// implicit return (arrow with expression body) or contains a qualifying
// ReturnStatement in its block body.
func renderReturns(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	switch fn.Kind {
	case ast.KindArrowFunction:
		af := fn.AsArrowFunction()
		if af.Body == nil {
			return false
		}
		if af.Body.Kind != ast.KindBlock {
			// Implicit return: `render = () => <div/>`.
			return true
		}
		return hasReturnInBody(af.Body)
	case ast.KindFunctionExpression:
		return hasReturnInBody(fn.AsFunctionExpression().Body)
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		var body *ast.Node
		switch fn.Kind {
		case ast.KindMethodDeclaration:
			body = fn.AsMethodDeclaration().Body
		case ast.KindGetAccessor:
			body = fn.AsGetAccessorDeclaration().Body
		case ast.KindSetAccessor:
			body = fn.AsSetAccessorDeclaration().Body
		}
		if body == nil {
			// Overload signature / abstract / ambient method — not a concrete
			// body we'd reasonably demand a return from. Mirrors ESLint's
			// default behavior, which never reaches this state in practice.
			return true
		}
		return hasReturnInBody(body)
	}
	return false
}

var RequireRenderReturnRule = rule.Rule{
	Name: "react/require-render-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		report := func(member *ast.Node) {
			ctx.ReportNode(member, rule.RuleMessage{
				Id:          "noRenderReturn",
				Description: "Your render method should have a return statement",
			})
		}

		checkComponent := func(members []*ast.Node) {
			renders := findAllRenderMembers(members)
			if len(renders) == 0 {
				return
			}
			for _, m := range renders {
				if renderReturns(renderFunctionOf(m)) {
					return
				}
			}
			// None of the render-named members has a return. Report on the
			// first one — matches eslint-plugin-react's `findRenderMethod`,
			// which reports on the first qualifying member.
			report(renders[0])
		}

		checkClass := func(classNode *ast.Node) {
			if !reactutil.ExtendsReactComponent(classNode, pragma) {
				return
			}
			checkComponent(classNode.Members())
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: checkClass,
			ast.KindClassExpression:  checkClass,

			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateClassCall(call, pragma, createClass) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				arg := ast.SkipParentheses(call.Arguments.Nodes[0])
				if arg.Kind != ast.KindObjectLiteralExpression {
					return
				}
				checkComponent(arg.AsObjectLiteralExpression().Properties.Nodes)
			},
		}
	},
}
