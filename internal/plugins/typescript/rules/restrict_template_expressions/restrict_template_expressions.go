package restrict_template_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type RestrictTemplateExpressionsOptions struct {
	AllowNumber  bool     `json:"allowNumber"`
	AllowBoolean bool     `json:"allowBoolean"`
	AllowAny     bool     `json:"allowAny"`
	AllowNullish bool     `json:"allowNullish"`
	AllowRegExp  bool     `json:"allowRegExp"`
	AllowNever   bool     `json:"allowNever"`
	AllowArray   bool     `json:"allowArray"`
	Allow        []string `json:"allow"`
}

// RestrictTemplateExpressionsRule implements the restrict-template-expressions rule
// Enforce template literal expressions to be of string type
var RestrictTemplateExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "restrict-template-expressions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := RestrictTemplateExpressionsOptions{
		AllowNumber:  false,
		AllowBoolean: false,
		AllowAny:     false,
		AllowNullish: false,
		AllowRegExp:  false,
		AllowNever:   false,
		AllowArray:   false,
		Allow:        []string{},
	}

	// Parse options
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
			if allowNumber, ok := optsMap["allowNumber"].(bool); ok {
				opts.AllowNumber = allowNumber
			}
			if allowBoolean, ok := optsMap["allowBoolean"].(bool); ok {
				opts.AllowBoolean = allowBoolean
			}
			if allowAny, ok := optsMap["allowAny"].(bool); ok {
				opts.AllowAny = allowAny
			}
			if allowNullish, ok := optsMap["allowNullish"].(bool); ok {
				opts.AllowNullish = allowNullish
			}
			if allowRegExp, ok := optsMap["allowRegExp"].(bool); ok {
				opts.AllowRegExp = allowRegExp
			}
			if allowNever, ok := optsMap["allowNever"].(bool); ok {
				opts.AllowNever = allowNever
			}
			if allowArray, ok := optsMap["allowArray"].(bool); ok {
				opts.AllowArray = allowArray
			}
			if allow, ok := optsMap["allow"].([]interface{}); ok {
				for _, item := range allow {
					if str, ok := item.(string); ok {
						opts.Allow = append(opts.Allow, str)
					}
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindTemplateExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			templateExpr := node.AsTemplateExpression()
			if templateExpr == nil || templateExpr.TemplateSpans == nil {
				return
			}

			// Check each template span's expression
			for _, span := range templateExpr.TemplateSpans.Nodes {
				templateSpan := span.AsTemplateSpan()
				if templateSpan == nil || templateSpan.Expression == nil {
					continue
				}

				// TODO: Use TypeChecker to check if expression type is allowed
				// For now, this is a placeholder that will need proper type checking implementation
				// Example:
				// typ := ctx.TypeChecker.GetTypeAtLocation(templateSpan.Expression)
				// if !isAllowedType(typ, opts) {
				//     ctx.ReportNode(templateSpan.Expression, rule.RuleMessage{
				//         Id:          "invalidType",
				//         Description: "Invalid type in template expression",
				//     })
				// }
			}
		},
	}
}
