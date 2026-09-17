// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package explicit_timer_delay

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed explicit_timer_delay.schema.json
var schemaJSON []byte

var ExplicitTimerDelayRule = rule.Rule{
	Name:   "unicorn/explicit-timer-delay",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		removeZero := len(options) > 0 && options[0] == "never"
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ast.IsOptionalChainRoot(node) {
					return
				}
				name := timerName(ctx, node)
				if name == "" {
					return
				}
				args := node.Arguments()
				// Runtime argument positions are unknown after a leading spread.
				if len(args) == 0 || args[0].Kind == ast.KindSpreadElement {
					return
				}
				if !removeZero && len(args) == 1 {
					ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
						Id: "missing-delay", Description: "`" + name + "` should have an explicit delay argument.", Data: map[string]string{"name": name},
					}, func() []rule.RuleFix {
						end := args[0].End()
						return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(end, end), ", 0")}
					})
				}
				if removeZero && len(args) == 2 && isZeroDelay(args[1]) {
					ctx.ReportNodeWithDeferredFixes(utils.ESTreeRuntimeExpression(args[1]), rule.RuleMessage{
						Id: "redundant-delay", Description: "`" + name + "` should not have an explicit delay of `0`.", Data: map[string]string{"name": name},
					}, func() []rule.RuleFix {
						return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(args[0].End(), args[1].End()), "")}
					})
				}
			},
		}
	},
}

func timerName(ctx rule.RuleContext, node *ast.Node) string {
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee == nil {
		return ""
	}
	if ast.IsIdentifier(callee) {
		if (callee.Text() == "setTimeout" || callee.Text() == "setInterval") && ctx.Globals.Access(callee.Text()).IsDeclared() && ctx.Refs.IsGlobalReference(callee) {
			return callee.Text()
		}
		return ""
	}
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Methods: []string{"setTimeout", "setInterval"}, AllowOptionalMember: true})
	if !ok {
		return ""
	}
	object := utils.ESTreeRuntimeExpression(call.Object)
	if !ast.IsIdentifier(object) {
		return ""
	}
	switch object.Text() {
	case "window", "globalThis", "global", "self":
		if ctx.Globals.Access(object.Text()).IsDeclared() && ctx.Refs.IsGlobalReference(object) {
			return call.Property.Text()
		}
	}
	return ""
}

func isZeroDelay(node *ast.Node) bool {
	for {
		node = utils.ESTreeRuntimeExpression(node)
		if node.Kind != ast.KindPrefixUnaryExpression {
			return ast.IsNumericLiteral(node) && node.Text() == "0"
		}
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindPlusToken && unary.Operator != ast.KindMinusToken {
			return false
		}
		node = unary.Operand
	}
}
