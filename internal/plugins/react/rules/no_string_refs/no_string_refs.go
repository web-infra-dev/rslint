package no_string_refs

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoStringRefsRule = rule.Rule{
	Name: "react/no-string-refs",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		detectTemplateLiterals := false
		if optsMap := utils.GetOptionsMap(options); optsMap != nil {
			if v, ok := optsMap["noTemplateLiterals"].(bool); ok {
				detectTemplateLiterals = v
			}
		}

		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		// `this.refs` is writable in React 18.3.0 and later, so only check on versions < 18.3.0.
		// When `react.version` is absent the setting defaults to "latest" (999.999.999),
		// which disables the check.
		checkRefsUsage := reactutil.ReactVersionLessThan(ctx.Settings, 18, 3, 0)

		reportStringRef := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "stringInRefDeprecated",
				Description: "Using string literals in ref attributes is deprecated.",
			})
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if !checkRefsUsage {
					return
				}
				prop := node.AsPropertyAccessExpression()
				if ast.SkipParentheses(prop.Expression).Kind != ast.KindThisKeyword {
					return
				}
				name := prop.Name()
				if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "refs" {
					return
				}
				if !reactutil.IsInsideReactComponent(node, pragma, createClass) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "thisRefsDeprecated",
					Description: "Using this.refs is deprecated.",
				})
			},

			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				name := attr.Name()
				if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "ref" {
					return
				}
				init := attr.Initializer
				if init == nil {
					return
				}
				switch init.Kind {
				case ast.KindStringLiteral:
					reportStringRef(node)
				case ast.KindJsxExpression:
					expr := init.AsJsxExpression().Expression
					if expr == nil {
						return
					}
					// Unwrap parens so `ref={('x')}` reports (ESTree flattens
					// parens; tsgo preserves them). We deliberately do NOT
					// unwrap TypeScript wrappers such as AsExpression /
					// NonNullExpression / SatisfiesExpression: eslint-plugin-react
					// (under @typescript-eslint/parser) only matches when
					// `expression.type === 'Literal'`, and those wrappers make
					// the outer type something else, so upstream does not report
					// `ref={'x' as string}` / `ref={'x'!}`. We align with that.
					expr = ast.SkipParentheses(expr)
					switch expr.Kind {
					case ast.KindStringLiteral:
						reportStringRef(node)
					case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
						if detectTemplateLiterals {
							reportStringRef(node)
						}
					}
				}
			},
		}
	},
}
