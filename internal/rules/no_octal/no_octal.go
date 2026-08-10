package no_octal

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const legacyOctalTokenFlags = ast.TokenFlagsOctal | ast.TokenFlagsContainsLeadingZero

var noOctalMessage = rule.RuleMessage{
	Id:          "noOctal",
	Description: "Octal literals should not be used.",
}

// https://eslint.org/docs/latest/rules/no-octal
var NoOctalRule = rule.Rule{
	Name:   "no-octal",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				// NumericLiteral.Text is normalized, but the parser preserves whether
				// the raw token was a legacy octal or a leading-zero decimal in flags.
				if node.AsNumericLiteral().TokenFlags&legacyOctalTokenFlags == 0 {
					return
				}

				// Node.Pos may include leading trivia. Skip only that trivia on the
				// rare report path instead of rescanning every numeric token.
				ctx.ReportRange(
					node.Loc.WithPos(scanner.SkipTrivia(sourceText, node.Pos())),
					noOctalMessage,
				)
			},
		}
	},
}
