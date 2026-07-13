package no_empty_static_block

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-empty-static-block
var NoEmptyStaticBlockRule = rule.Rule{
	Name: "no-empty-static-block",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindClassStaticBlockDeclaration: func(node *ast.Node) {
				staticBlock := node.AsClassStaticBlockDeclaration()
				if staticBlock == nil || staticBlock.Body == nil {
					return
				}

				body := staticBlock.Body
				block := body.AsBlock()
				if block == nil {
					return
				}
				if block.Statements != nil && len(block.Statements.Nodes) != 0 {
					return
				}

				if utils.HasCommentInsideNode(ctx.SourceFile, body) {
					return
				}

				bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
				ctx.ReportRangeWithSuggestions(bodyRange, unexpectedMessage(), rule.RuleSuggestion{
					Message:  suggestCommentMessage(),
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(utils.BracedNodeInnerRange(ctx.SourceFile, body), " /* empty */ ")},
				})
			},
		}
	},
}

func unexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected empty static block.",
	}
}

func suggestCommentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestComment",
		Description: "Add comment inside empty static block.",
	}
}
