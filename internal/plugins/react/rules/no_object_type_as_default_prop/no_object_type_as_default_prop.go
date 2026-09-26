package no_object_type_as_default_prop

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	scopeAnalysis "github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

var NoObjectTypeAsDefaultPropRule = rule.Rule{
	Name:   "react/no-object-type-as-default-prop",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		scopes := scopeAnalysis.For(ctx)

		check := func(node *ast.Node) {
			params := utils.ESTreeParameters(node)
			if len(params) == 0 {
				return
			}
			param := params[0].AsParameterDeclaration()
			if param.Initializer != nil || param.DotDotDotToken != nil || param.Name().Kind != ast.KindObjectBindingPattern {
				return
			}
			componentChecked := false
			for _, element := range param.Name().Elements() {
				binding := element.AsBindingElement()
				if binding.Initializer == nil || binding.DotDotDotToken != nil {
					continue
				}
				kind := forbiddenDefaultType(binding.Initializer)
				if kind == "" {
					continue
				}
				// Inspect the function body only when a default could be reported.
				if !componentChecked {
					if !reactutil.IsDetectedStatelessComponent(node, pragma, ctx.TypeChecker, wrappers, scopes) {
						return
					}
					componentChecked = true
				}
				key := ast.TryGetPropertyNameOfBindingOrAssignmentElement(element)
				if key != nil && key.Kind == ast.KindComputedPropertyName {
					key = utils.ESTreeRuntimeExpression(key.AsComputedPropertyName().Expression)
				}
				// Upstream reads Property.key.name, including "undefined" for
				// literal keys and computed expressions without an identifier name.
				propName := "undefined"
				if key != nil && key.Kind == ast.KindIdentifier {
					propName = key.Text()
				}
				// ESTree reports AssignmentPattern, which excludes an aliased key.
				start := utils.TrimNodeTextRange(ctx.SourceFile, binding.Name()).Pos()
				ctx.ReportRange(core.NewTextRange(start, element.End()), rule.RuleMessage{
					Id: "forbiddenTypeDefaultParam",
					Description: fmt.Sprintf("%s has a/an %s as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of %s.",
						propName, kind, kind),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
			ast.KindMethodDeclaration:   check,
			ast.KindGetAccessor:         check,
			ast.KindSetAccessor:         check,
		}
	},
}

func forbiddenDefaultType(node *ast.Node) string {
	node = utils.ESTreeRuntimeExpression(node)
	switch node.Kind {
	case ast.KindArrowFunction:
		return "arrow function"
	case ast.KindFunctionExpression:
		return "function expression"
	case ast.KindObjectLiteralExpression:
		return "object literal"
	case ast.KindArrayLiteralExpression:
		return "array literal"
	case ast.KindClassExpression:
		return "class expression"
	case ast.KindNewExpression:
		return "construction expression"
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return "JSX element"
	case ast.KindRegularExpressionLiteral:
		return "regex literal"
	case ast.KindCallExpression:
		callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
		if !ast.IsOptionalChain(node) && callee != nil && callee.Kind == ast.KindIdentifier && callee.Text() == "Symbol" {
			return "Symbol literal"
		}
	}
	return ""
}
