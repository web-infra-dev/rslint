package no_this_alias

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoThisAliasOptions struct {
	AllowDestructuring bool     `json:"allowDestructuring"`
	AllowedNames       []string `json:"allowedNames"`
}

// aliasTarget returns the ESTree-equivalent VariableDeclarator id or
// AssignmentExpression left-hand side for a `this` value. ESTree elides
// parentheses, while the TypeScript AST preserves them.
func aliasTarget(node *ast.Node) *ast.Node {
	value := node
	for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
		value = value.Parent
	}

	parent := value.Parent
	if parent == nil {
		return nil
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration != nil && declaration.Initializer == value {
			return declaration.Name()
		}
	case ast.KindBinaryExpression:
		if !ast.IsAssignmentExpression(parent, false) {
			return nil
		}
		expression := parent.AsBinaryExpression()
		if expression != nil && expression.Right == value {
			return ast.SkipParentheses(expression.Left)
		}
	}

	return nil
}

var NoThisAliasRule = rule.CreateRule(rule.Rule{
	Name: "no-this-alias",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := NoThisAliasOptions{
			AllowDestructuring: true,
			AllowedNames:       []string{},
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowDestructuring, ok := optsMap["allowDestructuring"].(bool); ok {
					opts.AllowDestructuring = allowDestructuring
				}
				if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
					for _, name := range allowedNames {
						if nameStr, ok := name.(string); ok {
							opts.AllowedNames = append(opts.AllowedNames, nameStr)
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindThisKeyword: func(node *ast.Node) {
				target := aliasTarget(node)
				if target == nil {
					return
				}

				if opts.AllowDestructuring && target.Kind != ast.KindIdentifier {
					return
				}

				if target.Kind == ast.KindIdentifier {
					name := target.AsIdentifier().Text
					for _, allowedName := range opts.AllowedNames {
						if name == allowedName {
							return
						}
					}
					ctx.ReportNode(target, rule.RuleMessage{
						Id:          "thisAssignment",
						Description: "Unexpected aliasing of `this` to local variable.",
					})
					return
				}

				ctx.ReportNode(target, rule.RuleMessage{
					Id:          "thisDestructure",
					Description: "Destructuring `this` is not allowed.",
				})
			},
		}
	},
})
