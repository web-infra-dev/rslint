package prefer_enum_initializers

import (
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildDefineInitializerMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "defineInitializer",
		Description: "The value of the member '" + name + "' should be explicitly defined.",
		Data:        map[string]string{"name": name},
	}
}

func buildDefineInitializerSuggestionMessage(name, suggested string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "defineInitializerSuggestion",
		Description: "Can be fixed to " + name + " = " + suggested,
		Data:        map[string]string{"name": name, "suggested": suggested},
	}
}

var PreferEnumInitializersRule = rule.CreateRule(rule.Rule{
	Name: "prefer-enum-initializers",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				for index, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil || member.Initializer != nil {
						continue
					}

					memberRange := utils.TrimNodeTextRange(ctx.SourceFile, memberNode)
					name := sourceText[memberRange.Pos():memberRange.End()]

					ctx.ReportRangeWithDeferredSuggestions(memberRange, buildDefineInitializerMessage(name), func() []rule.RuleSuggestion {
						indexStr := strconv.Itoa(index)
						nextStr := strconv.Itoa(index + 1)
						stringSuggested := "'" + name + "'"
						// Keep the three single-fix suggestions in one backing array.
						fixes := [...]rule.RuleFix{
							rule.RuleFixReplaceRange(memberRange, name+" = "+indexStr),
							rule.RuleFixReplaceRange(memberRange, name+" = "+nextStr),
							rule.RuleFixReplaceRange(memberRange, name+" = "+stringSuggested),
						}
						return []rule.RuleSuggestion{
							{
								Message:  buildDefineInitializerSuggestionMessage(name, indexStr),
								FixesArr: fixes[0:1:1],
							},
							{
								Message:  buildDefineInitializerSuggestionMessage(name, nextStr),
								FixesArr: fixes[1:2:2],
							},
							{
								Message:  buildDefineInitializerSuggestionMessage(name, stringSuggested),
								FixesArr: fixes[2:3:3],
							},
						}
					})
				}
			},
		}
	},
})
