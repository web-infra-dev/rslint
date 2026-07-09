package use_memo

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	expectedMemoizationFunctionReason = "Expected a callback function to be passed to useMemo"
	expectedInlineFunctionReason      = "Expected the first argument to be an inline function expression"
	expectedArrayLiteralReason        = "Expected the dependency list for useMemo to be an array literal"
	expectedSimpleDepsReason          = "Expected the dependency list to be an array of simple expressions (e.g. `x`, `x.y.z`, `x?.y?.z`)"
	callbackParamsReason              = "useMemo() callbacks may not accept parameters"
	asyncGeneratorReason              = "useMemo() callbacks may not be async or generator functions"
	reassignOuterVariableReason       = "useMemo() callbacks may not reassign variables declared outside of the callback"
)

const (
	expectedMemoizationFunctionDescription = "The first argument to useMemo() must be a function that calculates a result to cache"
	expectedInlineFunctionDescription      = expectedInlineFunctionReason
	expectedArrayLiteralDescription        = expectedArrayLiteralReason
	expectedSimpleDepsDescription          = expectedSimpleDepsReason
	callbackParamsDescription              = "useMemo() callbacks are called by React to cache calculations across re-renders. They should not take parameters. Instead, directly reference the props, state, or local variables needed for the computation"
	asyncGeneratorDescription              = "useMemo() callbacks are called once and must synchronously return a value"
	reassignOuterVariableDescription       = "useMemo() callbacks must be pure functions and cannot reassign variables defined outside of the callback function"
)

type useMemoState struct {
	ctx rule.RuleContext
}

// UseMemoRule is the rslint port of upstream `react-hooks/use-memo`.
//
// Upstream emits this rule from React Compiler's UseMemo diagnostic category.
// This port validates the published call-shape contract locally: useMemo must
// receive an inline function, dependency lists must be array literals of simple
// dependency expressions, callbacks cannot accept parameters or be async /
// generator functions, and callbacks cannot reassign variables captured from
// outside the callback.
var UseMemoRule = rule.Rule{
	Name: "react-hooks/use-memo",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &useMemoState{ctx: ctx}
		return rule.RuleListeners{
			ast.KindCallExpression: state.processCallExpression,
		}
	},
}

func (state *useMemoState) processCallExpression(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || !react_hooksutil.IsManualUseMemoCallee(call.Expression, state.ctx.TypeChecker) {
		return
	}

	args := call.Arguments
	if args == nil || len(args.Nodes) == 0 {
		state.ctx.ReportNode(node, buildUseMemoMessage("expectedMemoizationFunction", expectedMemoizationFunctionReason, expectedMemoizationFunctionDescription, "Expected a memoization function"))
		return
	}

	callback := ast.SkipParentheses(args.Nodes[0])
	if callback != nil && callback.Kind == ast.KindSpreadElement {
		return
	}
	if callback == nil || !ast.IsFunctionExpressionOrArrowFunction(callback) {
		reportNode := node
		if callback != nil {
			reportNode = callback
		}
		state.ctx.ReportNode(reportNode, buildUseMemoMessage("expectedInlineFunction", expectedInlineFunctionReason, expectedInlineFunctionDescription, "Expected the first argument to be an inline function expression"))
		return
	}

	state.checkCallback(callback)
	state.checkDependencyList(args.Nodes)
}

func (state *useMemoState) checkCallback(callback *ast.Node) {
	params := callback.Parameters()
	if len(params) > 0 {
		reportNode := params[0]
		if name := params[0].Name(); name != nil {
			reportNode = name
		}
		state.ctx.ReportNode(reportNode, buildUseMemoMessage("noCallbackParameters", callbackParamsReason, callbackParamsDescription, "Callbacks with parameters are not supported"))
	}
	flags := ast.GetFunctionFlags(callback)
	if flags&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) != 0 {
		state.ctx.ReportNode(callback, buildUseMemoMessage("noAsyncOrGenerator", asyncGeneratorReason, asyncGeneratorDescription, "Async and generator functions are not supported"))
	}
	state.checkExternalReassignments(callback)
}

func (state *useMemoState) checkDependencyList(args []*ast.Node) {
	if len(args) < 2 {
		return
	}
	deps := ast.SkipParentheses(args[1])
	if deps != nil && deps.Kind == ast.KindSpreadElement {
		return
	}
	if deps == nil || deps.Kind != ast.KindArrayLiteralExpression {
		reportNode := args[1]
		if deps != nil {
			reportNode = deps
		}
		state.ctx.ReportNode(reportNode, buildUseMemoMessage("expectedArrayLiteral", expectedArrayLiteralReason, expectedArrayLiteralDescription, expectedArrayLiteralReason))
		return
	}
	array := deps.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return
	}
	for _, elem := range array.Elements.Nodes {
		if elem == nil {
			continue
		}
		if elem.Kind == ast.KindSpreadElement || elem.Kind == ast.KindOmittedExpression {
			state.ctx.ReportNode(deps, buildUseMemoMessage("expectedArrayLiteral", expectedArrayLiteralReason, expectedArrayLiteralDescription, expectedArrayLiteralReason))
			return
		}
		expr := skipDependencyParensAndNonNull(elem)
		if expr == nil || !isSimpleDependencyExpression(expr) {
			state.ctx.ReportNode(elem, buildUseMemoMessage("expectedSimpleDependencies", expectedSimpleDepsReason, expectedSimpleDepsDescription, expectedSimpleDepsReason))
		}
	}
}

func isSimpleDependencyExpression(node *ast.Node) bool {
	node = skipDependencyParensAndNonNull(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword:
		return true
	case ast.KindPropertyAccessExpression:
		return isSimpleDependencyExpression(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if access == nil || !isSimpleDependencyExpression(access.Expression) {
			return false
		}
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key == nil {
			return false
		}
		switch key.Kind {
		case ast.KindStringLiteral, ast.KindNumericLiteral:
			return true
		}
		return false
	}
	return false
}

func skipDependencyParensAndNonNull(node *ast.Node) *ast.Node {
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKNonNullAssertions)
}

func (state *useMemoState) checkExternalReassignments(callback *ast.Node) {
	reported := map[*ast.Node]bool{}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != callback && react_hooksutil.IsCompilerFunctionKind(node) {
			return
		}

		switch node.Kind {
		case ast.KindBinaryExpression:
			if ast.IsAssignmentExpression(node, false) {
				binary := node.AsBinaryExpression()
				if binary != nil {
					state.reportExternalAssignmentTargets(callback, binary.Left, reported)
				}
			}
		case ast.KindPrefixUnaryExpression:
			prefix := node.AsPrefixUnaryExpression()
			if prefix != nil && (prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken) {
				state.reportExternalAssignmentTargets(callback, prefix.Operand, reported)
			}
		case ast.KindPostfixUnaryExpression:
			postfix := node.AsPostfixUnaryExpression()
			if postfix != nil && (postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken) {
				state.reportExternalAssignmentTargets(callback, postfix.Operand, reported)
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(callback)
}

func (state *useMemoState) reportExternalAssignmentTargets(callback *ast.Node, target *ast.Node, reported map[*ast.Node]bool) {
	for _, assignmentTarget := range react_hooksutil.CollectAssignmentTargetIdentifiersThroughAssertions(target) {
		id := assignmentTarget.Identifier
		if id == nil || reported[id] || !state.isDeclaredOutsideCallback(callback, id) {
			continue
		}
		reported[id] = true
		state.ctx.ReportNode(id, buildUseMemoMessage("noExternalReassignment", reassignOuterVariableReason, reassignOuterVariableDescription, "Cannot reassign variable"))
	}
}

func (state *useMemoState) isDeclaredOutsideCallback(callback *ast.Node, id *ast.Node) bool {
	if id == nil || id.Kind != ast.KindIdentifier || state.ctx.TypeChecker == nil {
		return false
	}
	sym := utils.GetReferenceSymbol(id, state.ctx.TypeChecker)
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil || ast.GetSourceFileOfNode(decl) != state.ctx.SourceFile {
			continue
		}
		if react_hooksutil.ContainsNode(callback, decl) {
			return false
		}
		return true
	}
	return false
}

func buildUseMemoMessage(id, reason, description, detail string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          id,
		Description: fmt.Sprintf("Error: %s\n\n%s.", reason, description),
		Data: map[string]string{
			"detail":      detail,
			"description": description,
			"reason":      reason,
		},
	}
}
