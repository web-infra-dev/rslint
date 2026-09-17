// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ReportUnnecessaryLengthArgument checks the second argument of an already
// matched two-argument slice/splice call. Callers own receiver restrictions.
func ReportUnnecessaryLengthArgument(ctx rule.RuleContext, call DotMethodCall, messageID, argumentName string) {
	raw := call.Call.Arguments()[1]
	argument := utils.ESTreeRuntimeExpression(raw)
	description := lengthOrInfinityDescription(ctx, argument, utils.ESTreeRuntimeExpression(call.Object))
	if description == "" {
		return
	}
	message := rule.RuleMessage{Id: messageID, Description: "Passing `" + description + "` as the `" + argumentName + "` argument is unnecessary.", Data: map[string]string{"description": description, "argumentName": argumentName}}
	ctx.ReportNodeWithDeferredFixes(argument, message, func() []rule.RuleFix {
		// Preserve comments before the separator and any existing trailing comma.
		comma, ok := utils.TokenBeforePosition(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, raw).Pos())
		if !ok || comma.Kind != ast.KindCommaToken {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, raw.End()))}
	})
}
func lengthOrInfinityDescription(ctx rule.RuleContext, argument, object *ast.Node) string {
	if ast.IsIdentifier(argument) && argument.Text() == "Infinity" && ctx.Globals.Access("Infinity").IsDeclared() && ctx.Refs.IsGlobalReference(argument) {
		return "Infinity"
	}
	if !ast.IsPropertyAccessExpression(argument) {
		return ""
	}
	member := argument.AsPropertyAccessExpression()
	receiver := utils.ESTreeRuntimeExpression(member.Expression)
	if !ast.IsIdentifier(member.Name()) {
		return ""
	}
	if member.Name().Text() == "POSITIVE_INFINITY" && !ast.IsOptionalChain(argument) && ast.IsIdentifier(receiver) && receiver.Text() == "Number" && ctx.Globals.Access("Number").IsDeclared() && ctx.Refs.IsGlobalReference(receiver) {
		return "Number.POSITIVE_INFINITY"
	}
	if member.Name().Text() != "length" || !utils.IsSameReference(object, receiver, false) {
		return ""
	}
	name := "…"
	if ast.IsIdentifier(object) {
		name = object.Text()
	}
	if ast.IsOptionalChain(argument) {
		return name + "?.length"
	}
	return name + ".length"
}
