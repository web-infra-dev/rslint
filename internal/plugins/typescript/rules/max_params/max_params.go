package max_params

import (
	"fmt"
	"math"

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

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	check := func(node *ast.Node) {
		params := node.Parameters()
		effective := len(params)
		if !opts.countVoidThis && len(params) > 0 && isThisVoidParam(params[0]) {
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

// isThisVoidParam reports whether `p` is a `this: void` parameter, which
// the typescript-eslint rule strips from the parameter list before counting
// (unless `countVoidThis: true`).
func isThisVoidParam(p *ast.Node) bool {
	if !ast.IsThisParameter(p) {
		return false
	}
	t := p.Type()
	return t != nil && t.Kind == ast.KindVoidKeyword
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
// `count > undefined` comparison false (effectively disabling the rule).
// We model that case with MaxInt.
func parseOptions(o any) ruleOptions {
	out := ruleOptions{max: defaultMax}
	m := utils.GetOptionsMap(o)
	if m == nil {
		return out
	}
	_, hasMaximum := m["maximum"]
	_, hasMax := m["max"]
	if hasMaximum || hasMax {
		hasNum := false
		if hasMaximum {
			if n, ok := utils.CoerceInt(m["maximum"]); ok && n != 0 {
				out.max = n
				hasNum = true
			}
		}
		if !hasNum && hasMax {
			if n, ok := utils.CoerceInt(m["max"]); ok {
				out.max = n
				hasNum = true
			}
		}
		if !hasNum {
			out.max = math.MaxInt
		}
	}
	if v, ok := m["countVoidThis"]; ok {
		if b, ok := v.(bool); ok {
			out.countVoidThis = b
		}
	}
	return out
}

