package no_meaningless_void_operator

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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
	CheckNever *bool `json:"checkNever"`
}

func parseOptions(options []any) NoMeaninglessVoidOperatorOptions {
	opts := NoMeaninglessVoidOperatorOptions{}
	if len(options) > 0 {
		if optsMap, ok := options[0].(map[string]interface{}); ok {
			if optsJSON, err := json.Marshal(optsMap); err == nil {
				_ = json.Unmarshal(optsJSON, &opts)
			}
		}
	}
	if opts.CheckNever == nil {
		opts.CheckNever = utils.Ref(false)
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

				mask := checker.TypeFlagsVoidLike | checker.TypeFlagsNever

				for _, t := range utils.UnionTypeParts(argType) {
					mask &= checker.Type_flags(t)
				}

				fixRemoveVoidKeyword := func() rule.RuleFix {
					return rule.RuleFixRemoveRange(utils.TrimNodeTextRange(ctx.SourceFile, node).WithEnd(arg.Pos()))
				}

				if mask&checker.TypeFlagsVoidLike != 0 {
					ctx.ReportNodeWithFixes(node, buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType)), fixRemoveVoidKeyword())
				} else if *opts.CheckNever && mask&checker.TypeFlagsNever != 0 {
					ctx.ReportNodeWithSuggestions(node, buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType)), rule.RuleSuggestion{
						Message:  buildRemoveVoidMessage(),
						FixesArr: []rule.RuleFix{fixRemoveVoidKeyword()},
					})
				}
			},
		}
	},
})
