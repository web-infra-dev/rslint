package no_octal

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-octal
var NoOctalRule = rule.Rule{
	Name: "no-octal",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				// tsgo normalizes NumericLiteral.Text at parse time (e.g. "01234" -> "668"),
				// so we must read the raw source to distinguish a legacy octal from a decimal.
				raw := utils.TrimmedNodeText(ctx.SourceFile, node)
				if isOctalLiteralRaw(raw) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noOctal",
						Description: "Octal literals should not be used.",
					})
				}
			},
		}
	},
}

// isOctalLiteralRaw reports whether a NumericLiteral's raw source text denotes a
// legacy octal literal or a leading-zero decimal (e.g. `08`, `09.1`).
// Matches ESLint's `^0\d` check on `node.raw` — hex (`0x`), modern octal (`0o`),
// binary (`0b`), and fractional literals (`0.1`) are excluded by construction.
func isOctalLiteralRaw(raw string) bool {
	return len(raw) >= 2 && raw[0] == '0' && raw[1] >= '0' && raw[1] <= '9'
}
