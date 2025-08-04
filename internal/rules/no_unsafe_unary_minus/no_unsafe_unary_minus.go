package no_unsafe_unary_minus

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnaryMinusMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unaryMinus",
		Description: fmt.Sprintf("Argument of unary negation should be assignable to number | bigint but is %v instead.", t),
	}
}

var NoUnsafeUnaryMinusRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-unary-minus",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				expr := node.AsPrefixUnaryExpression()

				if expr.Operator != ast.KindMinusToken {
					return
				}

				argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Operand)

				for _, t := range utils.UnionTypeParts(argType) {
					if !utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsNever|checker.TypeFlagsBigIntLike|checker.TypeFlagsNumberLike) {
						ctx.ReportNode(node, buildUnaryMinusMessage(ctx.TypeChecker.TypeToString(t)))
						break
					}
				}
			},
		}
	},
})
