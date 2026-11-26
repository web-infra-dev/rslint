// Package valid_expect holds the framework-neutral async-assertion machinery
// shared by jest/valid-expect and rstest/valid-expect: walking out of a chained
// .then/.catch, spotting a Promise.all([...]) wrapper, deciding whether the
// result is awaited or returned, and the async/await fixes.
//
// Only the parts whose correctness does not depend on the framework live here.
// The rules keep their own bodies: how a broken chain's reason is derived, what
// counts as an excess expect argument and which matcher shapes exist all differ
// between jest and rstest, and merging those would mean framework branches
// inside a shared body.
package valid_expect

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func isPromiseMethodCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}

	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !ast.IsAccessExpression(callee) {
		return false
	}

	return testFramework.CalleeChainName(internalUtils.AccessExpressionObject(callee)) == "Promise"
}

// outermostTransparentExpression walks upward across parentheses and
// TypeScript assertion wrappers (`as`, `<T>`, `satisfies`, `!`) that are erased
// at runtime and preserve the expression's Promise value.
func outermostTransparentExpression(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			if paren := parent.AsParenthesizedExpression(); paren != nil && paren.Expression == current {
				current = parent
				continue
			}
		} else if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == current {
			current = parent
			continue
		}
		break
	}
	return current
}

func getPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	if outer := outermostTransparentExpression(node); node.Kind == ast.KindArrayLiteralExpression && outer.Parent != nil && outer.Parent.Kind == ast.KindCallExpression {
		node = outer.Parent
	}

	if isPromiseMethodCall(node) {
		return node
	}

	return nil
}

func findPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	node = outermostTransparentExpression(node)
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	if node.Parent.Kind != ast.KindCallExpression && node.Parent.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	return getPromiseCallExpressionNode(node.Parent)
}

func getParentIfPromiseChained(node *ast.Node) *ast.Node {
	outer := outermostTransparentExpression(node)
	if outer == nil || outer.Parent == nil {
		return node
	}

	member := outer.Parent
	if !ast.IsAccessExpression(member) || internalUtils.AccessExpressionObject(member) != outer {
		return node
	}

	outerMember := outermostTransparentExpression(member)
	grandParent := outerMember.Parent
	if grandParent == nil ||
		grandParent.Kind != ast.KindCallExpression ||
		grandParent.AsCallExpression().Expression != outerMember {
		return node
	}

	_, last := testFramework.AccessorPropertyAndName(member)
	if last == "then" || last == "catch" {
		return getParentIfPromiseChained(grandParent)
	}

	return node
}

func isAcceptableReturnNode(node *ast.Node, allowReturn bool) bool {
	if node == nil {
		return false
	}

	if allowReturn && node.Kind == ast.KindReturnStatement {
		return true
	}
	if node.Kind == ast.KindConditionalExpression {
		outer := outermostTransparentExpression(node)
		return isAcceptableReturnNode(outer.Parent, allowReturn)
	}

	return node.Kind == ast.KindArrowFunction || node.Kind == ast.KindAwaitExpression
}

// PromiseArrayExceptionKey identifies a Promise.all([...]) call by source range,
// so the assertions inside its array report once between them.
func PromiseArrayExceptionKey(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	return fmt.Sprintf("%d:%d", r.Pos(), r.End())
}

// AsyncInsertFix marks the function containing an async assertion as async.
func AsyncInsertFix(sourceFile *ast.SourceFile, fn *ast.Node) rule.RuleFix {
	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		head := internalUtils.GetFunctionHeadLoc(sourceFile, fn)
		return rule.RuleFixReplaceRange(core.NewTextRange(head.Pos(), head.Pos()), "async ")
	default:
		return rule.RuleFixInsertBefore(sourceFile, fn, "async ")
	}
}

// AwaitFix awaits an async assertion, replacing a `return` with `await` when
// returning it is not accepted.
func AwaitFix(sourceFile *ast.SourceFile, node *ast.Node, alwaysAwait bool) rule.RuleFix {
	outer := outermostTransparentExpression(node)
	if alwaysAwait && outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement {
		ret := outer.Parent
		retRange := internalUtils.TrimNodeTextRange(sourceFile, ret)
		outerRange := internalUtils.TrimNodeTextRange(sourceFile, outer)
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return rule.RuleFixReplaceRange(
				core.NewTextRange(retRange.Pos(), outerRange.End()),
				"await ("+sourceFile.Text()[outerRange.Pos():outerRange.End()]+")",
			)
		}
		return rule.RuleFixReplaceRange(core.NewTextRange(retRange.Pos(), outerRange.Pos()), "await ")
	}
	// If the assertion is wrapped in a TypeScript assertion (`as`, `<T>`,
	// `satisfies`, `!`), await the outermost wrapper so `await` binds to the
	// cast Promise rather than splitting `expect(...)` from `as Promise<void>`.
	// Plain parentheses keep inserting before `node` to preserve `(await ...)`.
	if outer != internalUtils.OutermostParenthesizedExpression(node) {
		outerRange := internalUtils.TrimNodeTextRange(sourceFile, outer)
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return rule.RuleFixReplaceRange(
				outerRange,
				"await ("+sourceFile.Text()[outerRange.Pos():outerRange.End()]+")",
			)
		}
		return rule.RuleFixInsertBefore(sourceFile, outer, "await ")
	}
	return rule.RuleFixInsertBefore(sourceFile, node, "await ")
}

// ResolveAsyncAssertionReportNode decides where an async assertion is reported
// and whether it has to be reported at all. assertionNode must be the outermost
// expression of the assertion chain: for a Chai chain carrying several matchers,
// such as expect(p).resolves.to.be.a("string").that.contains("x"), an inner
// matcher's call would be inspected against the wrong parent.
//
// reportNode is the node the report and the await fix apply to. promiseWrapped
// tells the caller to use the Promise-specific message, and insideAssertionArray
// marks a Promise.all([...]) array whose assertions report only once.
func ResolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	if assertionNode == nil {
		return nil, false, false, false
	}

	promiseChainedAssertionNode := getParentIfPromiseChained(assertionNode)
	outerChained := outermostTransparentExpression(promiseChainedAssertionNode)
	insideAssertionArray = outerChained.Parent != nil && outerChained.Parent.Kind == ast.KindArrayLiteralExpression
	reportNode = promiseChainedAssertionNode
	if promiseCallNode := findPromiseCallExpressionNode(promiseChainedAssertionNode); promiseCallNode != nil {
		reportNode = promiseCallNode
		promiseWrapped = true
	}

	// Parentheses and TypeScript assertion wrappers around the assertion are
	// erased at runtime, so the await or return that consumes the Promise may
	// sit above one or more wrappers. reportNode itself stays the report range.
	outermost := outermostTransparentExpression(reportNode)
	if outermost.Parent == nil || isAcceptableReturnNode(outermost.Parent, !alwaysAwait) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	return reportNode, promiseWrapped, insideAssertionArray, true
}
