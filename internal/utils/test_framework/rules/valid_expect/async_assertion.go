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

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
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

func getPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	if outer := internalUtils.OutermostParenthesizedExpression(node); node.Kind == ast.KindArrayLiteralExpression && outer.Parent != nil && outer.Parent.Kind == ast.KindCallExpression {
		node = outer.Parent
	}

	if isPromiseMethodCall(node) {
		return node
	}

	return nil
}

func findPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	node = internalUtils.OutermostParenthesizedExpression(node)
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	if node.Parent.Kind != ast.KindCallExpression && node.Parent.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	return getPromiseCallExpressionNode(node.Parent)
}

func getParentIfPromiseChained(node *ast.Node) *ast.Node {
	outer := internalUtils.OutermostParenthesizedExpression(node)
	if outer == nil || outer.Parent == nil || outer.Parent.Parent == nil {
		return node
	}

	grandParent := outer.Parent.Parent
	if grandParent.Kind != ast.KindCallExpression || !ast.IsAccessExpression(grandParent.AsCallExpression().Expression) {
		return node
	}

	member := grandParent.AsCallExpression().Expression
	entries := testFramework.GetMemberEntries(member)
	if len(entries) == 0 {
		return node
	}

	last := entries[len(entries)-1].Name
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
		return isAcceptableReturnNode(node.Parent, allowReturn)
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
	outer := internalUtils.OutermostParenthesizedExpression(node)
	if alwaysAwait && outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement {
		ret := outer.Parent
		retRange := internalUtils.TrimNodeTextRange(sourceFile, ret)
		nodeRange := internalUtils.TrimNodeTextRange(sourceFile, outer)
		return rule.RuleFixReplaceRange(core.NewTextRange(retRange.Pos(), nodeRange.Pos()), "await ")
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
	outerChained := internalUtils.OutermostParenthesizedExpression(promiseChainedAssertionNode)
	insideAssertionArray = outerChained.Parent != nil && outerChained.Parent.Kind == ast.KindArrayLiteralExpression
	reportNode = promiseChainedAssertionNode
	if promiseCallNode := findPromiseCallExpressionNode(promiseChainedAssertionNode); promiseCallNode != nil {
		reportNode = promiseCallNode
		promiseWrapped = true
	}

	// Parentheses around the assertion are nodes in the TypeScript AST but not
	// in ESTree, so the await or return upstream sees as the direct parent sits
	// one or more wrappers up here. reportNode itself stays the report range and
	// the insertion point for the fix.
	outermost := internalUtils.OutermostParenthesizedExpression(reportNode)
	if outermost.Parent == nil || isAcceptableReturnNode(outermost.Parent, !alwaysAwait) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	return reportNode, promiseWrapped, insideAssertionArray, true
}
