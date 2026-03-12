package jsx_boolean_value

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var JsxBooleanValueRule = rule.Rule{
	Name: "react/jsx-boolean-value",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		mode := "never" // default mode
		var exceptions map[string]bool
		assumeUndefinedIsFalse := false

		// Parse options: first element is mode string, second is options object
		if arr, ok := options.([]interface{}); ok {
			if len(arr) > 0 {
				if m, ok := arr[0].(string); ok {
					mode = m
				}
			}
			if len(arr) > 1 {
				if optsObj, ok := arr[1].(map[string]interface{}); ok {
					// ESLint schema: when mode is "always", exceptions are in "never" key and vice versa
					exKey := "always"
					if mode == "always" {
						exKey = "never"
					}
					if exList, ok := optsObj[exKey]; ok {
						if exArr, ok := exList.([]interface{}); ok {
							exceptions = make(map[string]bool)
							for _, item := range exArr {
								if name, ok := item.(string); ok {
									exceptions[name] = true
								}
							}
						}
					}
					if v, ok := optsObj["assumeUndefinedIsFalse"]; ok {
						if b, ok := v.(bool); ok {
							assumeUndefinedIsFalse = b
						}
					}
				}
			}
		} else {
			// Also try single string option
			if m, ok := options.(string); ok {
				mode = m
			}
		}

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				nameNode := attr.Name()
				if nameNode == nil {
					return
				}

				propName := ""
				if nameNode.Kind == ast.KindIdentifier {
					propName = nameNode.AsIdentifier().Text
				} else {
					trimmed := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
					propName = ctx.SourceFile.Text()[trimmed.Pos():trimmed.End()]
				}

				// Determine effective mode for this prop (considering exceptions)
				effectiveMode := mode
				if exceptions != nil && exceptions[propName] {
					if mode == "always" {
						effectiveMode = "never"
					} else {
						effectiveMode = "always"
					}
				}

				initializer := attr.Initializer

				if effectiveMode == "never" {
					// In "never" mode: attribute should NOT have ={true}
					if initializer != nil && isExpressionTrue(initializer) {
						// Fix: remove the ={true} part
						// The initializer starts at `=` and ends after `}`
						nameEnd := nameNode.End()
						initEnd := initializer.End()
						fixRange := core.NewTextRange(nameEnd, initEnd)

						ctx.ReportNodeWithFixes(node, rule.RuleMessage{
							Id:          "omitBoolean",
							Description: "Value must be omitted for boolean attribute `" + propName + "`",
						}, rule.RuleFix{
							Text:  "",
							Range: fixRange,
						})
					}

					// In "never" mode with assumeUndefinedIsFalse: if ={false}, suggest removing the entire prop
					if assumeUndefinedIsFalse && initializer != nil && isExpressionFalse(initializer) {
						trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)

						// Also remove any leading whitespace before the attribute
						text := ctx.SourceFile.Text()
						start := trimmed.Pos()
						for start > 0 && (text[start-1] == ' ' || text[start-1] == '\t') {
							start--
						}
						fixRange := core.NewTextRange(start, trimmed.End())

						ctx.ReportNodeWithFixes(node, rule.RuleMessage{
							Id:          "omitPropAndBoolean",
							Description: "Value must be omitted for `false` attribute: `" + propName + "`",
						}, rule.RuleFix{
							Text:  "",
							Range: fixRange,
						})
					}
				} else {
					// In "always" mode: attribute should have ={true}
					if initializer == nil {
						// Fix: add ={true} after the attribute name
						nameEnd := nameNode.End()
						fixRange := core.NewTextRange(nameEnd, nameEnd)

						ctx.ReportNodeWithFixes(node, rule.RuleMessage{
							Id:          "setBoolean",
							Description: "Value must be set for boolean attribute `" + propName + "`",
						}, rule.RuleFix{
							Text:  "={true}",
							Range: fixRange,
						})
					}
				}
			},
		}
	},
}

// isExpressionTrue checks if the initializer is `={true}`.
func isExpressionTrue(initializer *ast.Node) bool {
	if initializer.Kind != ast.KindJsxExpression {
		return false
	}
	expr := initializer.AsJsxExpression().Expression
	if expr == nil {
		return false
	}
	return expr.Kind == ast.KindTrueKeyword
}

// isExpressionFalse checks if the initializer is `={false}`.
func isExpressionFalse(initializer *ast.Node) bool {
	if initializer.Kind != ast.KindJsxExpression {
		return false
	}
	expr := initializer.AsJsxExpression().Expression
	if expr == nil {
		return false
	}
	return expr.Kind == ast.KindFalseKeyword
}
