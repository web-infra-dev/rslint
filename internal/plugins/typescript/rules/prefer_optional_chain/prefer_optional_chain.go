package prefer_optional_chain

import (
	"encoding/json"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type PreferOptionalChainOptions struct {
	AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing *bool `json:"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing"`
	CheckAny                                                           *bool `json:"checkAny"`
	CheckUnknown                                                       *bool `json:"checkUnknown"`
	CheckString                                                        *bool `json:"checkString"`
	CheckNumber                                                        *bool `json:"checkNumber"`
	CheckBoolean                                                       *bool `json:"checkBoolean"`
	CheckBigInt                                                        *bool `json:"checkBigInt"`
	RequireNullish                                                     *bool `json:"requireNullish"`
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
		Description: "Change to an optional chain expression. This may change the return type of the expression and some TypeScript errors may occur.",
	}
}

var PreferOptionalChainRule = rule.CreateRule(rule.Rule{
	Name: "prefer-optional-chain",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(PreferOptionalChainOptions)
		if !ok {
			opts = PreferOptionalChainOptions{}
			if options != nil {
				// For IPC mode, options come as []interface{} or map[string]interface{}
				raw := options
				if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
					raw = optArray[0]
				}
				if jsonBytes, err := json.Marshal(raw); err == nil {
					_ = json.Unmarshal(jsonBytes, &opts)
				}
			}
		}

		// Set defaults
		if opts.CheckAny == nil {
			opts.CheckAny = boolPtr(true)
		}
		if opts.CheckUnknown == nil {
			opts.CheckUnknown = boolPtr(true)
		}
		if opts.CheckString == nil {
			opts.CheckString = boolPtr(true)
		}
		if opts.CheckNumber == nil {
			opts.CheckNumber = boolPtr(true)
		}
		if opts.CheckBoolean == nil {
			opts.CheckBoolean = boolPtr(true)
		}
		if opts.CheckBigInt == nil {
			opts.CheckBigInt = boolPtr(true)
		}
		if opts.RequireNullish == nil {
			opts.RequireNullish = boolPtr(false)
		}
		if opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing == nil {
			opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing = boolPtr(false)
		}

		analyzer := NewOperandAnalyzer(ctx, opts)
		chainAnalyzer := NewChainAnalyzer(ctx, opts)

		// Track visited binary expressions to avoid duplicate reports
		visited := make(map[*ast.Node]struct{})

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

				// Skip if already visited
				if _, ok := visited[node]; ok {
					return
				}
				visited[node] = struct{}{}

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

func boolPtr(b bool) *bool {
	return &b
}

func derefBoolDefault(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}
