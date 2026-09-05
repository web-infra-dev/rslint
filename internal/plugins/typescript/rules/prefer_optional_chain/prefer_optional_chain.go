package prefer_optional_chain

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed prefer_optional_chain.schema.json
var schemaJSON []byte

type PreferOptionalChainOptions struct {
	AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing bool
	CheckAny                                                           bool
	CheckUnknown                                                       bool
	CheckString                                                        bool
	CheckNumber                                                        bool
	CheckBoolean                                                       bool
	CheckBigInt                                                        bool
	RequireNullish                                                     bool
}

func parseOptions(options []any) PreferOptionalChainOptions {
	opts := PreferOptionalChainOptions{
		CheckAny:     true,
		CheckUnknown: true,
		CheckString:  true,
		CheckNumber:  true,
		CheckBoolean: true,
		CheckBigInt:  true,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing"].(bool); ok {
		opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing = value
	}
	if value, ok := optsMap["checkAny"].(bool); ok {
		opts.CheckAny = value
	}
	if value, ok := optsMap["checkUnknown"].(bool); ok {
		opts.CheckUnknown = value
	}
	if value, ok := optsMap["checkString"].(bool); ok {
		opts.CheckString = value
	}
	if value, ok := optsMap["checkNumber"].(bool); ok {
		opts.CheckNumber = value
	}
	if value, ok := optsMap["checkBoolean"].(bool); ok {
		opts.CheckBoolean = value
	}
	if value, ok := optsMap["checkBigInt"].(bool); ok {
		opts.CheckBigInt = value
	}
	if value, ok := optsMap["requireNullish"].(bool); ok {
		opts.RequireNullish = value
	}
	return opts
}

func buildPreferOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferOptionalChain",
		Description: "Prefer using an optional chain expression instead, as it's more concise and easier to read.",
	}
}

func buildOptionalChainSuggestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "optionalChainSuggest",
		Description: "Change to an optional chain.",
	}
}

var PreferOptionalChainRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-optional-chain",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		analyzer := NewOperandAnalyzer(ctx, opts)
		chainAnalyzer := NewChainAnalyzer(ctx, opts)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				op := bin.OperatorToken.Kind

				// Only handle && and || chains
				if op != ast.KindAmpersandAmpersandToken && op != ast.KindBarBarToken {
					return
				}

				// Skip if this node is already part of a larger chain of the same operator
				// We only want to process from the topmost binary expression of each chain
				// Walk up through parenthesized expressions to handle cases like `a && (a.b && a.b.c)`
				ancestor := node.Parent
				for ancestor != nil && ast.IsParenthesizedExpression(ancestor) {
					ancestor = ancestor.Parent
				}
				if ancestor != nil && ast.IsBinaryExpression(ancestor) {
					parentBin := ancestor.AsBinaryExpression()
					if parentBin.OperatorToken.Kind == op {
						return
					}
				}

				// Gather and classify operands
				operands, chainOp := analyzer.GatherLogicalOperands(node)
				if len(operands) < 2 {
					return
				}

				// Analyze and report chains
				chainAnalyzer.AnalyzeChain(operands, chainOp, node)
			},

			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				// Check for (foo || {}).bar pattern
				chainAnalyzer.AnalyzeOrEmptyObjectPattern(node)
			},

			ast.KindElementAccessExpression: func(node *ast.Node) {
				// Check for (foo || {})[bar] pattern
				chainAnalyzer.AnalyzeOrEmptyObjectPattern(node)
			},
		}
	},
})
