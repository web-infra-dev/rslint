package no_plusplus

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_plusplus.schema.json
var schemaJSON []byte

type options struct {
	allowForLoopAfterthoughts bool
}

func parseOptions(raw []any) options {
	opts := options{allowForLoopAfterthoughts: false}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]any)
	if v, ok := m["allowForLoopAfterthoughts"].(bool); ok {
		opts.allowForLoopAfterthoughts = v
	}
	return opts
}

func buildUnexpectedUnaryOpMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedUnaryOp",
		Description: fmt.Sprintf("Unary operator '%s' used.", operator),
		Data: map[string]string{
			"operator": operator,
		},
	}
}

// isForLoopAfterthought mirrors upstream's isForLoopAfterthought(node), which
// walks up through SequenceExpression ancestors before checking
// isForStatementUpdate. tsgo has no SequenceExpression node — the comma
// operator collapses into a BinaryExpression (see AST_PATTERNS.md), and
// parentheses are explicit ParenthesizedExpression wrappers instead of being
// stripped from the tree — so the walk skips both.
func isForLoopAfterthought(node *ast.Node) bool {
	current := utils.OutermostParenthesizedExpression(node)
	parent := current.Parent
	if utils.IsCommaOperator(parent) {
		return isForLoopAfterthought(parent)
	}
	return parent != nil && parent.Kind == ast.KindForStatement &&
		parent.AsForStatement().Incrementor == current
}

func operatorText(kind ast.Kind) (string, bool) {
	switch kind {
	case ast.KindPlusPlusToken:
		return "++", true
	case ast.KindMinusMinusToken:
		return "--", true
	default:
		return "", false
	}
}

// NoPlusplusRule disallows the unary operators ++ and --.
//
// https://eslint.org/docs/latest/rules/no-plusplus
var NoPlusplusRule = rule.Rule{
	Name:   "no-plusplus",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		check := func(node *ast.Node, operatorKind ast.Kind) {
			operator, ok := operatorText(operatorKind)
			if !ok {
				return
			}
			if opts.allowForLoopAfterthoughts && isForLoopAfterthought(node) {
				return
			}
			ctx.ReportNode(node, buildUnexpectedUnaryOpMessage(operator))
		}

		return rule.RuleListeners{
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				check(node, node.AsPrefixUnaryExpression().Operator)
			},
			ast.KindPostfixUnaryExpression: func(node *ast.Node) {
				check(node, node.AsPostfixUnaryExpression().Operator)
			},
		}
	},
}
