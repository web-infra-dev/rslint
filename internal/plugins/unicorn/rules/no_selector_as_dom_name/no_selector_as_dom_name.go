// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package no_selector_as_dom_name

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The upstream pattern uses ASCII word characters.
var simpleSelector = regexp.MustCompile("^[.#]-?[A-Za-z_][A-Za-z0-9_-]*$")

var methods = []string{
	"add", "remove", "contains", "toggle", "replace", "getElementById", "getElementsByClassName",
}

var selectorMessage = rule.RuleMessage{
	Id:          "no-selector-as-dom-name",
	Description: "Do not use selector syntax in DOM names.",
}

const runtimeWrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

var NoSelectorAsDomNameRule = rule.Rule{
	Name:   "unicorn/no-selector-as-dom-name",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Methods: methods, AllowOptionalCall: true, AllowOptionalMember: true,
			})
			if !ok {
				return
			}
			args, prefix := domNameArguments(call)
			for _, arg := range args {
				if arg.Kind == ast.KindSpreadElement {
					continue
				}
				literal := ast.SkipOuterExpressions(arg, runtimeWrappers)
				value := utils.GetStaticStringValue(literal)
				if literal.Kind == ast.KindTemplateExpression {
					value = literal.AsTemplateExpression().Head.Text()
				}
				if value == "" || (value[0] != '.' && value[0] != '#') {
					continue
				}
				ctx.ReportNodeWithDeferredFixes(literal, selectorMessage, func() []rule.RuleFix {
					return removeSelectorPrefix(ctx.SourceFile, literal, value, prefix)
				})
			}
		}}
	},
}

func domNameArguments(call unicornutil.DotMethodCall) ([]*ast.Node, byte) {
	args := call.Call.Arguments()
	method := call.Property.Text()
	if method == "getElementById" || method == "getElementsByClassName" {
		if len(args) != 1 || isNonDOMReceiver(call.Object) {
			return nil, 0
		}
		if method == "getElementById" {
			return args, '#'
		}
		return args, '.'
	}

	classList := utils.ESTreeCallCallee(call.Object)
	if classList == nil || !ast.IsPropertyAccessExpression(classList) ||
		!ast.IsIdentifier(classList.Name()) || classList.Name().Text() != "classList" ||
		isNonDOMReceiver(classList.Expression()) {
		return nil, 0
	}
	switch method {
	case "add", "remove":
		return args, '.'
	case "replace":
		// A leading spread leaves the position of the next argument unknown.
		if len(args) > 0 && args[0].Kind == ast.KindSpreadElement {
			return nil, 0
		}
		return args[:min(len(args), 2)], '.'
	default:
		return args[:min(len(args), 1)], '.'
	}
}

func isNonDOMReceiver(node *ast.Node) bool {
	return unicornutil.IsNonDOMNodeExpression(ast.SkipOuterExpressions(node, runtimeWrappers))
}

func removeSelectorPrefix(source *ast.SourceFile, literal *ast.Node, value string, prefix byte) []rule.RuleFix {
	textRange := utils.TrimNodeTextRange(source, literal)
	start := textRange.Pos() + 1
	// A cooked prefix can come from an escape. Only remove a literal byte.
	if value[0] != prefix || start >= textRange.End() || source.Text()[start] != prefix {
		return nil
	}
	if literal.Kind == ast.KindTemplateExpression {
		spans := literal.AsTemplateExpression().TemplateSpans.Nodes
		if len(value) != 1 || len(spans) != 1 || spans[0].AsTemplateSpan().Literal.Text() != "" {
			return nil
		}
	} else if !simpleSelector.MatchString(value) {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, start+1))}
}
