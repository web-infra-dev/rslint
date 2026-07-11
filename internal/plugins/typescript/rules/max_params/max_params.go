package max_params

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MaxParamsRule enforces a maximum number of parameters in function
// definitions. Mirrors @typescript-eslint/max-params, which extends ESLint
// core's max-params with the `countVoidThis` option.
//
// https://typescript-eslint.io/rules/max-params
var MaxParamsRule = rule.CreateRule(rule.Rule{
	Name: "max-params",
	Run:  run,
})

const defaultMax = 3

type ruleOptions struct {
	max           int
	countVoidThis bool
}

func run(ctx rule.RuleContext, _options []any) rule.RuleListeners {
	options := rule.LegacyUnwrapOptions(_options)
	opts := parseOptions(options)

	check := func(node *ast.Node) {
		params := node.Parameters()
		effective := len(params)
		if !opts.countVoidThis && utils.IsThisVoidParameter(ast.GetThisParameter(node)) {
			effective--
		}
		if effective <= opts.max {
			return
		}
		name := utils.UpperCaseFirstASCII(utils.GetFunctionNameWithKind(node))
		ctx.ReportRange(
			utils.GetFunctionHeadLoc(ctx.SourceFile, node),
			rule.RuleMessage{
				Id: "exceed",
				Description: fmt.Sprintf(
					"%s has too many parameters (%d). Maximum allowed is %d.",
					name, effective, opts.max,
				),
			},
		)
	}

	// In ESTree, FunctionExpression wraps class methods, constructors,
	// getters, setters, and object methods (via MethodDefinition.value /
	// Property.value). tsgo represents each as its own kind, so listen
	// explicitly to keep parity with ESLint.
	return rule.RuleListeners{
		ast.KindFunctionDeclaration: check,
		ast.KindFunctionExpression:  check,
		ast.KindArrowFunction:       check,
		ast.KindMethodDeclaration:   check,
		ast.KindConstructor:         check,
		ast.KindGetAccessor:         check,
		ast.KindSetAccessor:         check,
		ast.KindFunctionType:        check,
	}
}

// parseOptions mirrors @typescript-eslint/max-params' schema: a single
// optional object with `max`, `maximum`, and `countVoidThis`. Other shapes
// (bare integer, integer-in-array) fall through to defaults — upstream
// rejects them at schema validation, and rslint has no schema layer to
// reproduce that, so we treat them as "no option" rather than guess.
//
// `option.maximum || option.max` follows JS truthy coercion: a present-but-
// zero `maximum` falls through to `max`; a present-but-falsy value with no
// fallback leaves `numParams = undefined` upstream, which makes every
// `count > undefined` comparison false. The shared legacy max helper models
// that disabled state with MaxInt.
func parseOptions(o any) ruleOptions {
	out := ruleOptions{max: defaultMax}
	m := utils.GetOptionsMap(o)
	if m == nil {
		return out
	}
	out.max = utils.ResolveLegacyMaxOption(m, defaultMax)
	if v, ok := m["countVoidThis"]; ok {
		if b, ok := v.(bool); ok {
			out.countVoidThis = b
		}
	}
	return out
}
