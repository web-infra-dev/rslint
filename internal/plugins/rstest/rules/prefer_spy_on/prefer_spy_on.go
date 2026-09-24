package prefer_spy_on

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
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
// in removed text, an argument beyond the implementation, a spread argument,
// an argument that would be dropped although evaluating it may have an
// effect, or a `super` property, which cannot be passed to `spyOn`.
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

	// A spread may expand to no argument at all, which would leave the spy
	// without an implementation, so it cannot be copied as one.
	if implementation != nil && implementation.Kind == ast.KindSpreadElement {
		return nil
	}

	keepsImplementation := !chainsMockImplementation(target.FnCall)
	if implementation != nil {
		switch {
		case isEmptyImplementation(ctx, implementation):
			// `rs.fn(undefined)` is `rs.fn()`, and handing the spy
			// `undefined` would make it call the original method.
			implementation = nil
		case !keepsImplementation && !isSideEffectFree(ctx, implementation):
			// The chained `mockImplementation()` replaces the argument, so
			// the fix drops it, and with it whatever evaluating it did.
			return nil
		case keepsImplementation && isVoidExpression(implementation):
			// `void expr` is `undefined` too, but replacing it would skip
			// evaluating `expr`.
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
	} else if isCommaExpression(target.Property) {
		// A comma inside the brackets would separate `spyOn` arguments.
		spy.WriteString("(" + utils.TrimmedNodeText(source, target.Property) + ")")
	} else {
		spy.WriteString(utils.TrimmedNodeText(source, target.Property))
	}
	spy.WriteString(")")
	if keepsImplementation {
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
	//
	// The rewritten statement starts with whatever now follows the removed
	// target. When that is an opening parenthesis on a line of its own, a
	// preceding statement without a semicolon would take it as a call, so
	// a semicolon is written in front of it.
	var fixes []rule.RuleFix
	var removed []core.TextRange
	if chain == target.FnCall {
		prefix := semicolonBefore(ctx, target.Assignment, spy.String()[0])
		removed = append(removed, core.NewTextRange(start, right.End()))
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, right.End()), prefix+spy.String()))
	} else {
		chainStart := utils.TrimNodeTextRange(source, chain).Pos()
		first := source.Text()[chainStart]
		if fnCallRange.Pos() == chainStart {
			first = spy.String()[0]
		}
		prefix := semicolonBefore(ctx, target.Assignment, first)
		removed = append(removed, core.NewTextRange(start, chainStart))
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, chainStart), prefix))
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
	if implementation != nil && keepsImplementation {
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

// isEmptyImplementation reports whether an implementation argument is
// `undefined` itself: the global `undefined`, or `void` applied to a literal,
// possibly wrapped in parentheses or TypeScript assertions.
func isEmptyImplementation(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindVoidExpression {
		return testFramework.IsSideEffectFreeLiteral(node.AsVoidExpression().Expression)
	}
	return isGlobalUndefined(ctx, node)
}

// isGlobalUndefined reports whether node is `undefined` and no declaration in
// this file, an import included, gives the name another value.
func isGlobalUndefined(ctx rule.RuleContext, node *ast.Node) bool {
	if !utils.IsUndefinedIdentifier(node) {
		return false
	}
	symbol := ctx.Refs.Resolve(node)
	if symbol == nil {
		return true
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil && ast.GetSourceFileOfNode(declaration) == ctx.SourceFile {
			return false
		}
	}
	return true
}

// isVoidExpression reports whether node is a `void` expression, possibly
// wrapped in parentheses or TypeScript assertions.
func isVoidExpression(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	return node != nil && node.Kind == ast.KindVoidExpression
}

// isSideEffectFree reports whether evaluating an implementation argument can
// be skipped: a function, a variable read, a literal or `undefined`.
func isSideEffectFree(ctx rule.RuleContext, node *ast.Node) bool {
	inner := utils.SkipAssertionsAndParens(node)
	if inner == nil {
		return false
	}
	switch inner.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindIdentifier:
		return true
	}
	return testFramework.IsSideEffectFreeLiteral(inner) || isEmptyImplementation(ctx, inner)
}

// isCommaExpression reports whether node is a comma expression not already
// wrapped in parentheses.
func isCommaExpression(node *ast.Node) bool {
	return node.Kind == ast.KindBinaryExpression &&
		node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken
}

// semicolonBefore returns the semicolon the rewritten statement needs when it
// will start with first, or nothing. Only an assignment that begins an
// expression statement can meet the previous statement: anywhere else the
// text before it belongs to the same expression.
func semicolonBefore(ctx rule.RuleContext, assignment *ast.Node, first byte) string {
	if first != '(' && first != '[' && first != '`' {
		return ""
	}
	start := utils.TrimNodeTextRange(ctx.SourceFile, assignment).Pos()
	node := assignment
	for node.Parent != nil && node.Parent.Kind != ast.KindExpressionStatement {
		if utils.TrimNodeTextRange(ctx.SourceFile, node.Parent).Pos() != start {
			return ""
		}
		node = node.Parent
	}
	statement := node.Parent
	if statement == nil || utils.TrimNodeTextRange(ctx.SourceFile, statement).Pos() != start {
		return ""
	}
	if !utils.NeedsPrecedingSemicolon(ctx.SourceFile, statement) {
		return ""
	}
	return ";"
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
