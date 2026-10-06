package no_meaningless_void_operator

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_meaningless_void_operator.schema.json
var schemaJSON []byte

func buildMeaninglessVoidOperatorMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "meaninglessVoidOperator",
		Description: fmt.Sprintf("void operator shouldn't be used on %v; it should convey that a return value is being ignored", t),
	}
}
func buildRemoveVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeVoid",
		Description: "Remove 'void'",
	}
}

type NoMeaninglessVoidOperatorOptions struct {
	CheckNever bool
}

func parseOptions(options []any) NoMeaninglessVoidOperatorOptions {
	opts := NoMeaninglessVoidOperatorOptions{}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["checkNever"].(bool); ok {
		opts.CheckNever = value
	}
	return opts
}

var NoMeaninglessVoidOperatorRule = rule.CreateRule(rule.Rule{
	Name:             "no-meaningless-void-operator",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindVoidExpression: func(node *ast.Node) {
				arg := node.AsVoidExpression().Expression
				argType := ctx.TypeChecker.GetTypeAtLocation(arg)

				unionParts := utils.UnionTypeParts(argType)
				isVoidLike := utils.Every(unionParts, func(t *checker.Type) bool {
					return utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike)
				})
				if !isVoidLike && (!opts.CheckNever || !utils.Every(unionParts, func(t *checker.Type) bool {
					return utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike|checker.TypeFlagsNever)
				})) {
					return
				}

				reportRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				message := buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType))
				buildFixes := func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixRemoveRange(reportRange.WithEnd(utils.TrimNodeTextRange(ctx.SourceFile, arg).Pos()))}
				}
				if isVoidLike {
					ctx.ReportRangeWithDeferredFixes(reportRange, message, buildFixes)
				} else {
					ctx.ReportRangeWithDeferredSuggestions(reportRange, message, func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  buildRemoveVoidMessage(),
							FixesArr: buildFixes(),
						}}
					})
				}
			},
		}
	},
})
