package max_params

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type MaxParamsOptions struct {
	Max           int  `json:"max"`
	CountVoidThis bool `json:"countVoidThis"`
}

func parseNumericOption(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func parseOptions(options any) MaxParamsOptions {
	opts := MaxParamsOptions{
		Max:           3,
		CountVoidThis: false,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if arr, ok := options.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap == nil {
		return opts
	}

	hasMax := false
	if value, ok := optsMap["max"]; ok {
		if maxValue, ok := parseNumericOption(value); ok {
			opts.Max = maxValue
			hasMax = true
		}
	}
	if !hasMax {
		if value, ok := optsMap["maximum"]; ok {
			if maxValue, ok := parseNumericOption(value); ok {
				opts.Max = maxValue
			}
		}
	}
	if value, ok := optsMap["countVoidThis"]; ok {
		if flag, ok := value.(bool); ok {
			opts.CountVoidThis = flag
		}
	}

	return opts
}

func isVoidThisParameter(param *ast.Node) bool {
	if param == nil || !ast.IsParameter(param) {
		return false
	}

	decl := param.AsParameterDeclaration()
	if decl == nil {
		return false
	}

	name := decl.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}

	if name.AsIdentifier().Text != "this" {
		return false
	}

	return decl.Type != nil && decl.Type.Kind == ast.KindVoidKeyword
}

func buildExceedMessage(count int, maxCount int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceed",
		Description: fmt.Sprintf("Function has too many parameters (%d). Maximum allowed is %d.", count, maxCount),
	}
}

var MaxParamsRule = rule.CreateRule(rule.Rule{
	Name: "max-params",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkParameters := func(node *ast.Node) {
			params := node.Parameters()
			if params == nil {
				return
			}

			count := len(params)
			if !opts.CountVoidThis && count > 0 && isVoidThisParameter(params[0]) {
				count--
			}

			if count > opts.Max {
				ctx.ReportNode(node, buildExceedMessage(count, opts.Max))
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkParameters,
			ast.KindFunctionExpression:  checkParameters,
			ast.KindArrowFunction:       checkParameters,
			ast.KindMethodDeclaration:   checkParameters,
			ast.KindMethodSignature:     checkParameters,
			ast.KindConstructor:         checkParameters,
			ast.KindGetAccessor:         checkParameters,
			ast.KindSetAccessor:         checkParameters,
			ast.KindCallSignature:       checkParameters,
			ast.KindConstructSignature:  checkParameters,
			ast.KindFunctionType:        checkParameters,
			ast.KindConstructorType:     checkParameters,
		}
	},
})
