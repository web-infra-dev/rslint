package cfg

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// deadWritesFromBinder reuses the bound flow graph for one tracked variable.
// Searching backwards from all reads stops at the first write on each path;
// those writes are live. A shared visited set bounds the search by the graph's
// size, including loop back edges. Multiple variables use the basic-block
// liveness solver, which avoids repeating this search for every variable.
//
// Binder flow describes type narrowing, not every JavaScript read/write event.
// Only a restricted function body can use it directly. Unsupported syntax or
// missing flow data returns no result so the caller can build CFG.
func deadWritesFromBinder(
	root *ast.Node,
	reads map[*ast.Node]int,
	writes []VariableWrite,
) ([]*ast.Node, bool) {
	if root.Kind != ast.KindFunctionDeclaration || root.AsFunctionDeclaration().AsteriskToken != nil ||
		ast.HasSyntacticModifier(root, ast.ModifierFlagsAsync) || root.ForEachChild(unsupportedBinderFlow) {
		return nil, false
	}
	for _, write := range writes {
		node := write.Node
		if node.AsIdentifier().FlowNode == nil {
			return nil, false
		}
		parent := node.Parent
		if parent.Name() == node && (parent.Kind == ast.KindParameter ||
			parent.Kind == ast.KindVariableDeclaration && parent.Initializer() == nil) {
			// These bindings have no corresponding binder assignment flow.
			return nil, false
		}
		for parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		if parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindNonNullExpression ||
			parent.Kind == ast.KindTypeAssertionExpression || parent.Kind == ast.KindSatisfiesExpression {
			return nil, false
		}
	}
	pending := make([]*ast.FlowNode, 0, len(reads))
	for read := range reads {
		flow := read.AsIdentifier().FlowNode
		if flow == nil {
			return nil, false
		}
		pending = append(pending, flow)
	}
	visited := make(map[*ast.FlowNode]struct{})
	live := make(map[*ast.Node]bool, len(writes))
	for _, write := range writes {
		live[write.Node] = false
	}
	for len(pending) != 0 {
		flow := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if flow == nil || flow.Flags&(ast.FlowFlagsStart|ast.FlowFlagsUnreachable) != 0 {
			continue
		}
		if _, seen := visited[flow]; seen {
			continue
		}
		visited[flow] = struct{}{}
		switch {
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			target := flow.Node
			if target == nil {
				return nil, false
			}
			if target.Kind == ast.KindVariableDeclaration {
				target = target.Name()
			}
			if _, ok := live[ast.SkipParentheses(target)]; ok {
				live[ast.SkipParentheses(target)] = true
				continue
			}
			pending = append(pending, flow.Antecedent)
		case flow.Flags&ast.FlowFlagsLabel != 0:
			for antecedent := flow.Antecedents; antecedent != nil; antecedent = antecedent.Next {
				pending = append(pending, antecedent.Flow)
			}
		case flow.Flags&(ast.FlowFlagsCondition|ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
			pending = append(pending, flow.Antecedent)
		default:
			// In particular, a ReduceLabel needs finally-context handling.
			return nil, false
		}
	}
	var dead []*ast.Node
	for node, used := range live {
		if !used && node.AsIdentifier().FlowNode.Flags&ast.FlowFlagsUnreachable == 0 {
			dead = append(dead, node)
		}
	}
	return dead, true
}

// Admit only syntax whose reads, writes and reachable paths agree with CFG.
// Unknown syntax falls back automatically. Among the excluded forms are
// destructuring (binder groups its writes), literal boolean conditions (binder
// prunes branches), nested flow roots, and try/finally (different path models).
// Short-circuit conditions can join differently, and CFG treats for-loop
// increment expressions as reachable before considering the body's exits.
func unsupportedBinderFlow(node *ast.Node) bool {
	if ast.IsOptionalChain(node) {
		return true
	}
	switch node.Kind {
	case ast.KindUnknown, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindTypeQuery:
		return true
	case ast.KindParameter:
		if node.Initializer() != nil {
			return true
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) ||
			ast.IsLogicalOrCoalescingAssignmentOperator(binary.OperatorToken.Kind) {
			return true
		}
		left := ast.SkipParentheses(binary.Left)
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
			(left.Kind == ast.KindArrayLiteralExpression || left.Kind == ast.KindObjectLiteralExpression) {
			return true
		}
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if node.Label() != nil {
			return true
		}
	case ast.KindBlock, ast.KindEmptyStatement, ast.KindDebuggerStatement,
		ast.KindVariableStatement, ast.KindVariableDeclarationList, ast.KindVariableDeclaration,
		ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindThrowStatement,
		ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement,
		ast.KindParenthesizedExpression, ast.KindConditionalExpression,
		ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindCallExpression, ast.KindNewExpression,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindDeleteExpression, ast.KindTypeOfExpression, ast.KindVoidExpression,
		ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression,
		ast.KindArrayLiteralExpression, ast.KindSpreadElement, ast.KindOmittedExpression,
		ast.KindObjectLiteralExpression, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindSpreadAssignment, ast.KindComputedPropertyName,
		ast.KindTemplateExpression, ast.KindTemplateSpan,
		ast.KindTypeParameter, ast.KindTypeAliasDeclaration:
	default:
		if node.Kind > ast.KindLastToken && (node.Kind < ast.KindFirstTypeNode || node.Kind > ast.KindLastTypeNode) {
			return true
		}
	}
	return node.ForEachChild(unsupportedBinderFlow)
}
