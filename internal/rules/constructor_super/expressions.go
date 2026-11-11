package constructor_super

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (a *analyzer) expression(n *ast.Node, s paths) paths {
	if n == nil || s == 0 || ast.IsTypeNode(n) {
		return s
	}
	if n.Flags&ast.NodeFlagsOptionalChain != 0 && (n.Kind == ast.KindCallExpression || n.Kind == ast.KindPropertyAccessExpression || n.Kind == ast.KindElementAccessExpression) {
		active, skipped := a.optionalChain(n, s)
		return active | skipped
	}
	if utils.IsFunctionLikeContainer(n) {
		if name := n.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
			return a.expression(name.Expression(), s)
		}
		return s
	}
	switch n.Kind {
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
		return a.expression(n.Expression(), s)
	case ast.KindClassDeclaration, ast.KindClassExpression:
		s = a.expression(superClass(n), s)
		for _, member := range n.Members() {
			if name := member.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				s = a.expression(name.Expression(), s)
			}
		}
		return s
	case ast.KindVariableStatement:
		return a.expression(n.AsVariableStatement().DeclarationList, s)
	case ast.KindVariableDeclarationList:
		for _, decl := range n.AsVariableDeclarationList().Declarations.Nodes {
			s = a.expression(decl, s)
		}
		return s
	case ast.KindVariableDeclaration:
		d := n.AsVariableDeclaration()
		s = a.expression(d.Initializer, s)
		return a.bindingValue(d.Name(), s, true, valueFact(d.Initializer))
	case ast.KindConditionalExpression:
		c := n.AsConditionalExpression()
		yes, no := a.condition(c.Condition, s, true)
		return a.expression(c.WhenTrue, yes) | a.expression(c.WhenFalse, no)
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		op := b.OperatorToken.Kind
		if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken {
			yes, no := a.condition(n, s, true)
			return yes | no
		}
		if ast.IsAssignmentOperator(op) {
			return a.assignment(b, s)
		}
		s = a.expression(b.Left, s)
		s = a.expression(b.Right, s)
		if op != ast.KindCommaToken {
			a.mayThrow(s)
		}
		return s
	case ast.KindPropertyAccessExpression:
		s = a.expression(n.AsPropertyAccessExpression().Expression, s)
		a.mayThrow(s)
		return s
	case ast.KindElementAccessExpression:
		e := n.AsElementAccessExpression()
		s = a.expression(e.Expression, s)
		s = a.expression(e.ArgumentExpression, s)
		a.mayThrow(s)
		return s
	case ast.KindCallExpression, ast.KindNewExpression:
		s = a.expression(n.Expression(), s)
		for _, arg := range n.Arguments() {
			s = a.expression(arg, s)
		}
		if s == 0 {
			return 0
		}
		// Argument evaluation and constructing the base happen before this binds.
		a.mayThrow(s)
		if ast.IsSuperCall(n) {
			a.reportCall(n, s)
			s = called
			if a.instanceElements {
				a.mayThrow(s)
			}
		}
		return s
	case ast.KindObjectLiteralExpression:
		for _, p := range n.AsObjectLiteralExpression().Properties.Nodes {
			if name := p.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				s = a.expression(name.Expression(), s)
			}
			switch p.Kind {
			case ast.KindPropertyAssignment:
				s = a.expression(p.AsPropertyAssignment().Initializer, s)
			case ast.KindShorthandPropertyAssignment:
				s = a.expression(p.Name(), s)
			case ast.KindSpreadAssignment:
				s = a.expression(p.Expression(), s)
				a.mayThrow(s)
			}
		}
		return s
	case ast.KindArrayLiteralExpression:
		for _, e := range n.AsArrayLiteralExpression().Elements.Nodes {
			s = a.expression(e, s)
		}
		return s
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		s = a.expression(n.Expression(), s)
		a.mayThrow(s)
		return s
	case ast.KindIdentifier:
		if !utils.IsNonReferenceIdentifier(n) {
			a.mayThrow(s)
		}
		return s
	case ast.KindPrefixUnaryExpression:
		u := n.AsPrefixUnaryExpression()
		if u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken {
			s = a.reference(u.Operand, s)
			a.mayThrow(s)
			return s
		}
		s = a.expression(u.Operand, s)
		if u.Operator != ast.KindExclamationToken {
			a.mayThrow(s)
		}
		return s
	case ast.KindPostfixUnaryExpression:
		s = a.reference(n.AsPostfixUnaryExpression().Operand, s)
		a.mayThrow(s)
		return s
	case ast.KindVoidExpression, ast.KindTypeOfExpression, ast.KindAwaitExpression, ast.KindYieldExpression, ast.KindDeleteExpression:
		s = a.expression(n.Expression(), s)
		if n.Kind == ast.KindAwaitExpression || n.Kind == ast.KindDeleteExpression {
			a.mayThrow(s)
		}
		return s
	case ast.KindTaggedTemplateExpression:
		t := n.AsTaggedTemplateExpression()
		s = a.expression(t.Tag, s)
		s = a.expression(t.Template, s)
		a.mayThrow(s)
		return s
	}
	n.ForEachChild(func(child *ast.Node) bool { s = a.expression(child, s); return false })
	return s
}

// A reference evaluates a member's receiver/key before an assignment's RHS,
// but an identifier target is not read. PutValue can throw after the RHS.
func (a *analyzer) reference(n *ast.Node, s paths) paths {
	if n == nil || s == 0 {
		return s
	}
	n = ast.SkipOuterExpressions(n, ast.OEKAll)
	if n.Kind == ast.KindIdentifier {
		return s
	}
	if n.Kind == ast.KindPropertyAccessExpression {
		return a.expression(n.AsPropertyAccessExpression().Expression, s)
	}
	if n.Kind == ast.KindElementAccessExpression {
		e := n.AsElementAccessExpression()
		s = a.expression(e.Expression, s)
		return a.expression(e.ArgumentExpression, s)
	}
	return a.expression(n, s)
}
func (a *analyzer) assignment(b *ast.BinaryExpression, s paths) paths {
	op := b.OperatorToken.Kind
	left := ast.SkipOuterExpressions(b.Left, ast.OEKAll)
	if op == ast.KindEqualsToken && (left.Kind == ast.KindObjectLiteralExpression || left.Kind == ast.KindArrayLiteralExpression) {
		s = a.expression(b.Right, s)
		return a.bindingValue(left, s, false, valueFact(b.Right))
	}
	s = a.reference(b.Left, s)
	if op != ast.KindEqualsToken {
		a.mayThrow(s)
	}
	if ast.IsLogicalOrCoalescingAssignmentOperator(op) {
		skip, write := s, s
		if op == ast.KindQuestionQuestionEqualsToken {
			if nullish, known := a.nullish(b.Left); known {
				if nullish {
					skip = 0
				} else {
					write = 0
				}
			}
		} else if truth, known := a.truthy(b.Left); known {
			needs := op == ast.KindAmpersandAmpersandEqualsToken && truth || op == ast.KindBarBarEqualsToken && !truth
			if needs {
				skip = 0
			} else {
				write = 0
			}
		}
		write = a.expression(b.Right, write)
		a.mayThrow(write)
		return skip | write
	}
	s = a.expression(b.Right, s)
	a.mayThrow(s)
	return s
}

// Keep the skipped paths separate through a whole chain. Parentheses end that
// chain; a following ordinary call still evaluates all its arguments.
func (a *analyzer) optionalChain(n *ast.Node, s paths) (paths, paths) {
	receiver := n.Expression()
	var active, skipped paths
	if receiver != nil && receiver.Flags&ast.NodeFlagsOptionalChain != 0 {
		active, skipped = a.optionalChain(receiver, s)
	} else {
		active = a.expression(receiver, s)
	}
	var token *ast.Node
	switch n.Kind {
	case ast.KindCallExpression:
		token = n.AsCallExpression().QuestionDotToken
	case ast.KindPropertyAccessExpression:
		token = n.AsPropertyAccessExpression().QuestionDotToken
	case ast.KindElementAccessExpression:
		token = n.AsElementAccessExpression().QuestionDotToken
	}
	if token != nil {
		isNull, known := a.nullish(receiver)
		if !known || isNull {
			skipped |= active
		}
		if known && isNull {
			active = 0
		}
	}
	switch n.Kind {
	case ast.KindCallExpression:
		for _, arg := range n.Arguments() {
			active = a.expression(arg, active)
		}
	case ast.KindElementAccessExpression:
		active = a.expression(n.AsElementAccessExpression().ArgumentExpression, active)
	}
	a.mayThrow(active)
	return active, skipped
}

// Conditions carry their successful truthy/falsy paths. Logical short-circuit
// branches must not be joined before deciding which operand executes.
func (a *analyzer) condition(n *ast.Node, s paths, prune bool) (paths, paths) {
	if s == 0 {
		return 0, 0
	}
	if n == nil {
		return s, 0
	}
	n = ast.SkipOuterExpressions(n, ast.OEKAll)
	switch n.Kind {
	case ast.KindPrefixUnaryExpression:
		u := n.AsPrefixUnaryExpression()
		if u.Operator == ast.KindExclamationToken {
			yes, no := a.condition(u.Operand, s, true)
			return no, yes
		}
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		switch b.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			yes, no := a.condition(b.Left, s, true)
			ry, rn := a.condition(b.Right, yes, true)
			return ry, no | rn
		case ast.KindBarBarToken:
			yes, no := a.condition(b.Left, s, true)
			ry, rn := a.condition(b.Right, no, true)
			return yes | ry, rn
		case ast.KindQuestionQuestionToken:
			left := a.expression(b.Left, s)
			take, skip := left, left
			if nilValue, known := a.nullish(b.Left); known {
				if nilValue {
					skip = 0
				} else {
					take = 0
				}
			}
			ry, rn := a.condition(b.Right, take, true)
			if truth, known := a.truthy(b.Left); known {
				if truth {
					ry |= skip
				} else {
					rn |= skip
				}
			} else {
				ry |= skip
				rn |= skip
			}
			return ry, rn
		}
	case ast.KindConditionalExpression:
		c := n.AsConditionalExpression()
		yes, no := a.condition(c.Condition, s, true)
		ty, tn := a.condition(c.WhenTrue, yes, true)
		fy, fn := a.condition(c.WhenFalse, no, true)
		return ty | fy, tn | fn
	}
	s = a.expression(n, s)
	if prune {
		if truth, known := a.truthy(n); known {
			if truth {
				return s, 0
			}
			return 0, s
		}
	}
	return s, s
}
func (a *analyzer) truthy(n *ast.Node) (bool, bool) {
	if n == nil {
		return false, false
	}
	n = ast.SkipOuterExpressions(n, ast.OEKAll)
	// Calls remain opaque under the pinned rule contract, including super().
	if n.Kind == ast.KindCallExpression {
		return false, false
	}
	if n.Kind == ast.KindTrueKeyword {
		return true, true
	}
	if n.Kind == ast.KindFalseKeyword || n.Kind == ast.KindNullKeyword || n.Kind == ast.KindVoidExpression {
		return false, true
	}
	if a.truth == nil {
		a.truth = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return a.truth.EvalControlFlowTruthiness(n)
}
func (a *analyzer) nullish(n *ast.Node) (bool, bool) {
	fact := valueFact(n)
	if fact.undefined {
		return true, true
	}
	if fact.node != nil && fact.node.Kind == ast.KindNullKeyword {
		return true, true
	}
	if fact.nonUndefined && fact.node != nil && fact.node.Kind != ast.KindCallExpression {
		return false, true
	}
	if n == nil {
		return false, false
	}
	if a.truth == nil {
		a.truth = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return a.truth.EvalControlFlowNullish(n)
}
