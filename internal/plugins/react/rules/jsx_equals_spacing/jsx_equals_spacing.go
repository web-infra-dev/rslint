package jsx_equals_spacing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var JsxEqualsSpacingRule = rule.Rule{
	Name: "react/jsx-equals-spacing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		mode := "never" // default mode

		// Parse options: first element is the mode string
		if arr, ok := options.([]interface{}); ok {
			if len(arr) > 0 {
				if m, ok := arr[0].(string); ok {
					mode = m
				}
			}
		} else if m, ok := options.(string); ok {
			mode = m
		}

		checkAttributes := func(node *ast.Node) {
			var attrs *ast.Node
			if node.Kind == ast.KindJsxOpeningElement {
				attrs = node.AsJsxOpeningElement().Attributes
			} else {
				attrs = node.AsJsxSelfClosingElement().Attributes
			}

			if attrs == nil {
				return
			}

			jsxAttrs := attrs.AsJsxAttributes()
			if jsxAttrs.Properties == nil {
				return
			}

			text := ctx.SourceFile.Text()

			for _, prop := range jsxAttrs.Properties.Nodes {
				if prop.Kind != ast.KindJsxAttribute {
					continue
				}

				attr := prop.AsJsxAttribute()
				if attr.Initializer == nil {
					continue
				}

				nameNode := attr.Name()
				if nameNode == nil {
					continue
				}

				// Use trimmed ranges to exclude leading trivia
				nameTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				initTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, attr.Initializer)
				nameEnd := nameTrimmed.End()
				initStart := initTrimmed.Pos()

				// Find the `=` position in the source text
				equalsPos := -1
				for i := nameEnd; i < initStart && i < len(text); i++ {
					if text[i] == '=' {
						equalsPos = i
						break
					}
				}

				if equalsPos < 0 {
					continue
				}

				if mode == "never" {
					if equalsPos > nameEnd {
						ctx.ReportRangeWithFixes(
							core.NewTextRange(nameEnd, equalsPos),
							rule.RuleMessage{
								Id:          "noSpaceBefore",
								Description: "There should be no space before '='",
							},
							rule.RuleFix{
								Text:  "",
								Range: core.NewTextRange(nameEnd, equalsPos),
							},
						)
					}

					afterEquals := equalsPos + 1
					if afterEquals < initStart {
						ctx.ReportRangeWithFixes(
							core.NewTextRange(afterEquals, initStart),
							rule.RuleMessage{
								Id:          "noSpaceAfter",
								Description: "There should be no space after '='",
							},
							rule.RuleFix{
								Text:  "",
								Range: core.NewTextRange(afterEquals, initStart),
							},
						)
					}
				} else {
					// mode == "always"
					if equalsPos == nameEnd {
						insertPos := core.NewTextRange(equalsPos, equalsPos)
						ctx.ReportRangeWithFixes(
							insertPos,
							rule.RuleMessage{
								Id:          "needSpaceBefore",
								Description: "A space is required before '='",
							},
							rule.RuleFix{
								Text:  " ",
								Range: insertPos,
							},
						)
					}

					afterEquals := equalsPos + 1
					if afterEquals == initStart {
						insertPos := core.NewTextRange(afterEquals, afterEquals)
						ctx.ReportRangeWithFixes(
							insertPos,
							rule.RuleMessage{
								Id:          "needSpaceAfter",
								Description: "A space is required after '='",
							},
							rule.RuleFix{
								Text:  " ",
								Range: insertPos,
							},
						)
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     checkAttributes,
			ast.KindJsxSelfClosingElement: checkAttributes,
		}
	},
}
