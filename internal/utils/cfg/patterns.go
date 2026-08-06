package cfg

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// patternReads emits what an identifier or member assignment target evaluates
// before the assigned expression: the receiver and computed key of a member
// target, and the throwable fork ESLint records for naming the target inside a
// `try` block. Destructuring targets are laid out by patternBind instead.
func (b *Builder[E]) patternReads(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		// A compound-assignment or update-expression target reads the
		// variable before writing it; a plain assignment target does not,
		// but ESLint still counts naming it as a point where the enclosing
		// `try` block can throw — so the handler sees the value the target
		// had before the assignment.
		b.read(node)
		if isThrowableIdentifier(node) {
			b.firstThrowableFork()
		}

	case ast.KindParenthesizedExpression:
		b.patternReads(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternReads(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternReads(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternReads(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternReads(node.AsTypeAssertion().Expression)

	default:
		// A member target such as `obj.x` or `obj[k]` evaluates its receiver.
		b.expr(node)
	}
}

// patternWrites emits the write of an identifier assignment target after its
// assigned expression. Member targets write no variable.
func (b *Builder[E]) patternWrites(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		b.write(node)

	case ast.KindParenthesizedExpression:
		b.patternWrites(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternWrites(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternWrites(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternWrites(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternWrites(node.AsTypeAssertion().Expression)
	}
}

// patternBind lays out a destructuring target the way it evaluates at run
// time, after the assigned expression has produced its value: each element in
// source order evaluates its computed key, forks around its default, and
// writes its target.
func (b *Builder[E]) patternBind(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		b.read(node)
		if isThrowableIdentifier(node) {
			b.firstThrowableFork()
		}
		b.write(node)

	case ast.KindOmittedExpression:
		// An array hole binds nothing.

	case ast.KindParenthesizedExpression:
		b.patternBind(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternBind(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternBind(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternBind(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternBind(node.AsTypeAssertion().Expression)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		pattern := node.AsBindingPattern()
		if pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			binding := element.AsBindingElement()
			if binding == nil {
				continue
			}
			if binding.PropertyName != nil && binding.PropertyName.Kind == ast.KindComputedPropertyName {
				b.expr(binding.PropertyName.AsComputedPropertyName().Expression)
			}
			b.bindWithDefault(binding.Name(), binding.Initializer)
		}

	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				assignment := property.AsPropertyAssignment()
				if name := assignment.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					b.expr(name.AsComputedPropertyName().Expression)
				}
				b.patternBind(assignment.Initializer)
			case ast.KindShorthandPropertyAssignment:
				shorthand := property.AsShorthandPropertyAssignment()
				b.bindWithDefault(shorthand.Name(), shorthand.ObjectAssignmentInitializer)
			case ast.KindSpreadAssignment:
				b.patternBind(property.AsSpreadAssignment().Expression)
			}
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			b.patternBind(element)
		}

	case ast.KindSpreadElement:
		b.patternBind(node.AsSpreadElement().Expression)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			b.bindWithDefault(binary.Left, binary.Right)
			return
		}
		b.expr(node)

	default:
		// A member target such as `obj.x` or `obj[k]` evaluates its receiver.
		b.expr(node)
	}
}

// bindWithDefault lays out one pattern element: a fork around the default
// value, which only runs when the incoming value is undefined, then the
// target's own binding.
func (b *Builder[E]) bindWithDefault(target *ast.Node, fallback *ast.Node) {
	if fallback != nil && b.cur != nil {
		join := b.newBlock()
		b.link(b.cur, join)
		withFallback := b.newBlock()
		b.link(b.cur, withFallback)

		b.enter(withFallback)
		b.expr(fallback)
		b.link(b.cur, join)

		b.enter(join)
	}
	b.patternBind(target)
}

// isDestructuringTarget reports whether an assignment's left-hand side is an
// object or array pattern.
func isDestructuringTarget(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil &&
		(node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindArrayLiteralExpression)
}
