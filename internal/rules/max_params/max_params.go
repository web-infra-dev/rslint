package max_params

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_params.schema.json
var schemaJSON []byte

// MaxParamsRule enforces a maximum number of parameters in function definitions.
// https://eslint.org/docs/latest/rules/max-params
var MaxParamsRule = rule.Rule{
	Name:   "max-params",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		check := func(node *ast.Node) {
			params := utils.ESTreeParameters(node)
			effective := len(params)
			if effective <= opts.max {
				return
			}
			if effective != 0 {
				switch opts.countThis {
				case countThisNever:
					if ast.IsThisParameter(params[0]) {
						effective--
					}
				case countThisExceptVoid:
					if utils.IsThisVoidParameter(params[0]) {
						effective--
					}
				}
			}
			if effective <= opts.max {
				return
			}

			name := utils.UpperCaseFirstASCII(utils.GetFunctionNameWithKindCore(node))
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

		// ESLint listens for FunctionExpression, which covers methods,
		// constructors, getters, and setters through ESTree's wrapper nodes.
		// tsgo represents those function-likes directly, so each kind is wired.
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
	},
}

const (
	defaultMax          = 3
	countThisAlways     = "always"
	countThisNever      = "never"
	countThisExceptVoid = "except-void"
)

type ruleOptions struct {
	max       int
	countThis string
}

// parseOptions keeps max parsing in a shared helper; only countThis /
// countVoidThis are specific to this rule.
func parseOptions(options []any) ruleOptions {
	out := ruleOptions{
		max:       defaultMax,
		countThis: countThisExceptVoid,
	}
	if len(options) == 0 {
		return out
	}
	out.max = utils.ResolveLegacyMaxOption(options[0], defaultMax)
	m, _ := options[0].(map[string]interface{})
	if m == nil {
		return out
	}

	if v, ok := m["countThis"].(string); ok && v != "" {
		out.countThis = v
	} else if v, ok := m["countVoidThis"]; ok {
		if b, ok := v.(bool); ok && b {
			out.countThis = countThisAlways
		} else {
			out.countThis = countThisExceptVoid
		}
	}

	return out
}
