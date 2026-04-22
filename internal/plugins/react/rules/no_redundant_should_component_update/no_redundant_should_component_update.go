package no_redundant_should_component_update

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// extendsPureComponent reports whether `classNode`'s extends clause matches
// upstream's `componentUtil.isPureComponent` regex
// `/^(<pragma>\.)?PureComponent$/`. Distinct from
// `reactutil.ExtendsReactComponent` (which also matches plain `Component`).
//
// Parens around the extends expression are skipped — matching ESLint's
// behavior where `getText(node.superClass)` returns the bare member-expression
// text (range excludes surrounding parens) and the regex still matches.
func extendsPureComponent(classNode *ast.Node, pragma string) bool {
	if classNode == nil {
		return false
	}
	heritage := ast.GetClassExtendsHeritageElement(classNode)
	if heritage == nil {
		return false
	}
	hc := heritage.AsExpressionWithTypeArguments()
	if hc == nil || hc.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(hc.Expression)
	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text == "PureComponent"
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == "PureComponent"
	}
	return false
}

// hasShouldComponentUpdate mirrors upstream's
// `astUtil.getComponentProperties(node).some(p => getPropertyName(p) === 'shouldComponentUpdate')`.
//
// Upstream `getPropertyName` returns `nameNode.name`, which is populated for
// Identifier and PrivateIdentifier (per spec, PrivateIdentifier exposes its
// name without the leading `#`) — so `#shouldComponentUpdate` also matches,
// matching upstream. Other key shapes (StringLiteral / NumericLiteral /
// ComputedPropertyName) yield undefined `.name` upstream and never match.
func hasShouldComponentUpdate(members []*ast.Node) bool {
	for _, m := range members {
		if m == nil {
			continue
		}
		key := m.Name()
		if key == nil {
			continue
		}
		var name string
		switch key.Kind {
		case ast.KindIdentifier:
			name = key.AsIdentifier().Text
		case ast.KindPrivateIdentifier:
			name = strings.TrimPrefix(key.AsPrivateIdentifier().Text, "#")
		default:
			continue
		}
		if name == "shouldComponentUpdate" {
			return true
		}
	}
	return false
}

// classDisplayName mirrors upstream's `getNodeName`:
//
//   - ClassDeclaration / named ClassExpression → its own Identifier name
//     (e.g. `class Foo …` → "Foo", `class Bar extends …` returned from a
//     factory → "Bar").
//   - Anonymous ClassExpression assigned via `var Foo = class …` → "Foo"
//     (parent VariableDeclaration's binding Identifier).
//   - Anything else (e.g. `export default class extends …`) → "".
func classDisplayName(node *ast.Node) string {
	name := node.Name()
	if name != nil && name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text
	}
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindVariableDeclaration {
		binding := parent.AsVariableDeclaration().Name()
		if binding != nil && binding.Kind == ast.KindIdentifier {
			return binding.AsIdentifier().Text
		}
	}
	return ""
}

var NoRedundantShouldComponentUpdateRule = rule.Rule{
	Name: "react/no-redundant-should-component-update",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)

		check := func(node *ast.Node) {
			if !extendsPureComponent(node, pragma) {
				return
			}
			if !hasShouldComponentUpdate(node.Members()) {
				return
			}
			name := classDisplayName(node)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noShouldCompUpdate",
				Description: name + " does not need shouldComponentUpdate when extending React.PureComponent.",
			})
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: check,
			ast.KindClassExpression:  check,
		}
	},
}
