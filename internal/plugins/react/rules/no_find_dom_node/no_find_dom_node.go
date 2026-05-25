package no_find_dom_node

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const noFindDOMNodeMessage = "Do not use findDOMNode. It doesn’t work with function components and is deprecated in StrictMode. See https://reactjs.org/docs/react-dom.html#finddomnode"

// matchesFindDOMNodeCallee mirrors ESLint's two-branch check on `callee`:
//
//	('name' in callee && callee.name === 'findDOMNode') ||
//	('property' in callee && callee.property && 'name' in callee.property && callee.property.name === 'findDOMNode')
//
// The first branch matches a bare `findDOMNode(...)` identifier call. The
// second matches `<anything>.findDOMNode(...)` (property-access; bracket /
// element access is excluded because `Literal` has no `.name` in ESTree).
// ESTree's `PrivateIdentifier.name` has no leading `#`, so `this.#findDOMNode()`
// also matches upstream — we mirror that here.
func matchesFindDOMNodeCallee(callee *ast.Node) bool {
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == "findDOMNode"
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name == nil {
			return false
		}
		switch name.Kind {
		case ast.KindIdentifier:
			return name.AsIdentifier().Text == "findDOMNode"
		case ast.KindPrivateIdentifier:
			return name.AsPrivateIdentifier().Text == "#findDOMNode"
		}
	}
	return false
}

var NoFindDomNodeRule = rule.Rule{
	Name: "react/no-find-dom-node",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if !matchesFindDOMNodeCallee(callee) {
					return
				}
				ctx.ReportNode(callee, rule.RuleMessage{
					Id:          "noFindDOMNode",
					Description: noFindDOMNodeMessage,
				})
			},
		}
	},
}
