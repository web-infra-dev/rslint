package prefer_enum_initializers

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildDefineInitializerMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "defineInitializer",
		Description: fmt.Sprintf("The value of the member '%s' should be explicitly defined.", name),
		Data:        map[string]string{"name": name},
	}
}

func buildDefineInitializerSuggestionMessage(name, suggested string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "defineInitializerSuggestion",
		Description: fmt.Sprintf("Can be fixed to %s = %s", name, suggested),
		Data:        map[string]string{"name": name, "suggested": suggested},
	}
}

var PreferEnumInitializersRule = rule.CreateRule(rule.Rule{
	Name: "prefer-enum-initializers",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

					name := utils.TrimmedNodeText(ctx.SourceFile, memberNode)
					indexStr := strconv.Itoa(index)
					nextStr := strconv.Itoa(index + 1)
					stringSuggested := "'" + name + "'"

					ctx.ReportNodeWithSuggestions(memberNode, buildDefineInitializerMessage(name),
						rule.RuleSuggestion{
							Message: buildDefineInitializerSuggestionMessage(name, indexStr),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, memberNode, name+" = "+indexStr),
							},
						},
						rule.RuleSuggestion{
							Message: buildDefineInitializerSuggestionMessage(name, nextStr),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, memberNode, name+" = "+nextStr),
							},
						},
						rule.RuleSuggestion{
							Message: buildDefineInitializerSuggestionMessage(name, stringSuggested),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, memberNode, name+" = "+stringSuggested),
							},
						},
					)
				}
			},
		}
	},
})
