package no_redundant_should_component_update

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// hasShouldComponentUpdate mirrors upstream's
// `astUtil.getComponentProperties(node).some(p => getPropertyName(p) === 'shouldComponentUpdate')`.
//
// Includes PropertyDeclaration (class field arrow `shouldComponentUpdate = () => {}`)
// because upstream's `getComponentProperties` returns class fields too.
// Contrast `reactutil.ClassHasMethodNamed`, which excludes PropertyDeclaration
// to match MethodDefinition-listener semantics — wrong oracle for this rule.
func hasShouldComponentUpdate(members []*ast.Node) bool {
	for _, m := range members {
		if m == nil {
			continue
		}
		if reactutil.IdentifierOrPrivateName(m.Name()) == "shouldComponentUpdate" {
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
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)

		check := func(node *ast.Node) {
			if !reactutil.ExtendsReactPureComponent(node, pragma) {
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
