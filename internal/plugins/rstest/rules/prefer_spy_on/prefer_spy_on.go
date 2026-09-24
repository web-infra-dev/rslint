package prefer_spy_on

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	sharedPreferSpyOn "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_spy_on"
)

// emptyImplementation is what the fix installs when the mock factory is called
// without an implementation. A spy whose implementation is unset calls through
// to the original method (`implementation || spyState.getOriginal()` in
// packages/core/src/runtime/api/spy.ts), so an argument-less
// `.mockImplementation()` would run the real code where `rs.fn()` returned
// `undefined`. An explicit function keeps the mock returning `undefined`.
const emptyImplementation = "() => undefined"

func buildUseRsSpyOnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useRsSpyOn",
		Description: "Use `rs.spyOn` instead",
	}
}

// isRstestFnCall matches a call of the utilities object's `fn`, reached through
// any binding of `rs` or `rstest`. A bracketed key is the same member when it is
// a static string.
func isRstestFnCall(ctx rule.RuleContext) func(*ast.Node) (*ast.Node, bool) {
	return func(call *ast.Node) (*ast.Node, bool) {
		callee := ast.SkipParentheses(call.AsCallExpression().Expression)
		if !sharedPreferSpyOn.IsMemberAccessNode(callee) {
			return nil, false
		}
		if name, ok := utils.AccessExpressionStaticName(callee); !ok || name != "fn" {
			return nil, false
		}
		receiver := utils.AccessExpressionObject(callee)
		if !rstestUtils.IsUtilitiesObject(ctx, receiver) {
			return nil, false
		}
		return receiver, true
	}
}

// buildSpyOnFixes rewrites `obj.prop = rs.fn(impl)` to
// `rs.spyOn(obj, 'prop').mockImplementation(impl)`.
//
// The spy is written in place of the mock factory call, so any chain or
// assertion that follows it stays attached to the spy, and the assignment
// target in front of it is removed. `spyOn` is called on the receiver the
// factory was called on, which the rule has already resolved to the utilities
// object, so an aliased or namespaced receiver keeps working.
//
// No fix is offered when the rewrite cannot keep the source intact: a comment
// in removed text, an argument beyond the implementation, or a `super`
// property, which cannot be passed to `spyOn`.
func buildSpyOnFixes(ctx rule.RuleContext, target sharedPreferSpyOn.Target) []rule.RuleFix {
	source := ctx.SourceFile
	if target.Object.Kind == ast.KindSuperKeyword {
		return nil
	}

	call := target.FnCall.AsCallExpression()
	var implementation *ast.Node
	if call.Arguments != nil {
		switch len(call.Arguments.Nodes) {
		case 0:
		case 1:
			implementation = call.Arguments.Nodes[0]
		default:
			return nil
		}
	}

	var spy strings.Builder
	spy.WriteString(utils.TrimmedNodeText(source, target.Receiver))
	spy.WriteString(".spyOn(")
	spy.WriteString(utils.TrimmedNodeText(source, target.Object))
	spy.WriteString(", ")
	if target.Left.Kind == ast.KindPropertyAccessExpression {
		spy.WriteString("'" + target.Property.Text() + "'")
	} else {
		spy.WriteString(utils.TrimmedNodeText(source, target.Property))
	}
	spy.WriteString(")")
	if !chainsMockImplementation(target.FnCall) {
		spy.WriteString(".mockImplementation(")
		if implementation != nil {
			spy.WriteString(utils.TrimmedNodeText(source, implementation))
		} else {
			spy.WriteString(emptyImplementation)
		}
		spy.WriteString(")")
	}

	assignment := target.Assignment.AsBinaryExpression()
	start := utils.TrimNodeTextRange(source, assignment.Left).Pos()
	right := assignment.Right
	chain := ast.SkipParentheses(right)
	fnCall := utils.OutermostParenthesizedExpression(target.FnCall)
	fnCallRange := utils.TrimNodeTextRange(source, fnCall)

	// Everything from the assignment target to the spy's position is
	// removed, together with the parentheses around the whole right-hand
	// side and around the factory call itself: the spy is a call expression,
	// so it binds at least as tightly as whatever those parentheses grouped.
	var fixes []rule.RuleFix
	var removed []core.TextRange
	if chain == target.FnCall {
		removed = append(removed, core.NewTextRange(start, right.End()))
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, right.End()), spy.String()))
	} else {
		chainStart := utils.TrimNodeTextRange(source, chain).Pos()
		removed = append(removed, core.NewTextRange(start, chainStart))
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(start, chainStart)))
		for node := right; node != chain; node = node.AsParenthesizedExpression().Expression {
			closing := core.NewTextRange(node.AsParenthesizedExpression().Expression.End(), node.End())
			removed = append(removed, closing)
			fixes = append(fixes, rule.RuleFixRemoveRange(closing))
		}
		removed = append(removed, fnCallRange)
		fixes = append(fixes, rule.RuleFixReplaceRange(fnCallRange, spy.String()))
	}

	// The target, receiver and implementation are copied into the spy, so
	// only comments outside those copies would be lost.
	kept := []*ast.Node{target.Object, target.Receiver}
	if target.Left.Kind == ast.KindElementAccessExpression {
		kept = append(kept, target.Property)
	}
	if implementation != nil && !chainsMockImplementation(target.FnCall) {
		kept = append(kept, implementation)
	}
	comments := ctx.Comments.All()
	for _, span := range removed {
		if hasCommentOutside(source, comments, span, kept) {
			return nil
		}
	}

	return fixes
}

// chainsMockImplementation reports whether the factory call's result has its
// implementation set right away, which makes the one passed to the factory
// redundant.
func chainsMockImplementation(fnCall *ast.Node) bool {
	access := utils.OutermostParenthesizedExpression(fnCall).Parent
	if access == nil || !sharedPreferSpyOn.IsMemberAccessNode(access) {
		return false
	}
	if name, ok := utils.AccessExpressionStaticName(access); !ok || name != "mockImplementation" {
		return false
	}
	parent := access.Parent
	return parent != nil && parent.Kind == ast.KindCallExpression &&
		parent.AsCallExpression().Expression == access
}

// hasCommentOutside reports whether span holds a comment that is not inside
// one of the kept nodes.
func hasCommentOutside(source *ast.SourceFile, comments []*ast.CommentRange, span core.TextRange, kept []*ast.Node) bool {
	for _, comment := range comments {
		if comment.Pos() < span.Pos() || comment.End() > span.End() {
			continue
		}
		inside := false
		for _, node := range kept {
			nodeRange := utils.TrimNodeTextRange(source, node)
			if comment.Pos() >= nodeRange.Pos() && comment.End() <= nodeRange.End() {
				inside = true
				break
			}
		}
		if !inside {
			return true
		}
	}
	return false
}

var PreferSpyOnRule = sharedPreferSpyOn.NewRule(sharedPreferSpyOn.Config{
	Name:                 "rstest/prefer-spy-on",
	Message:              buildUseRsSpyOnMessage(),
	PlainAssignmentOnly:  true,
	UnwrapTypeAssertions: true,
	Prepare:              isRstestFnCall,
	Fix:                  buildSpyOnFixes,
})
