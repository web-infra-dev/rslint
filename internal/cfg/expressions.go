package cfg

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

func (b *Builder[E]) expr(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}

	switch node.Kind {
	case ast.KindIdentifier:
		b.read(node)
		if isThrowableIdentifier(node) {
			b.firstThrowableFork()
		}

	case ast.KindParenthesizedExpression:
		b.expr(node.AsParenthesizedExpression().Expression)

	case ast.KindBinaryExpression:
		b.binaryExpression(node)

	case ast.KindConditionalExpression:
		b.conditionalExpression(node)

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			b.updateExpression(unary.Operand)
			return
		}
		b.expr(unary.Operand)

	case ast.KindPostfixUnaryExpression:
		b.updateExpression(node.AsPostfixUnaryExpression().Operand)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		b.accessOrCall(node)

	case ast.KindYieldExpression:
		b.expr(node.AsYieldExpression().Expression)
		b.makeYield()

	case ast.KindClassExpression:
		b.classLike(node)

	default:
		if IsRoot(node) {
			b.nestedFunction(node)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			b.visitUnknown(child)
			return false
		})
	}
}

// visitUnknown routes a child of a node the builder has no dedicated handler
// for. Statements and expressions are handled by their own walkers so control
// flow inside them is still modelled.
func (b *Builder[E]) visitUnknown(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	if ast.IsStatement(node) {
		b.statement(node)
		return
	}
	b.expr(node)
}

func (b *Builder[E]) binaryExpression(node *ast.Node) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind

	switch {
	case operator == ast.KindAmpersandAmpersandToken ||
		operator == ast.KindBarBarToken ||
		operator == ast.KindQuestionQuestionToken:
		b.expr(binary.Left)
		join := b.newBlock()
		b.link(b.cur, join)
		right := b.newBlock()
		b.link(b.cur, right)
		b.enter(right)
		b.expr(binary.Right)
		b.link(b.cur, join)
		b.enter(join)

	case ast.IsLogicalOrCoalescingAssignmentOperator(operator):
		// `a ||= b` reads a, then conditionally evaluates b and writes a.
		b.patternReads(binary.Left)
		join := b.newBlock()
		b.link(b.cur, join)
		right := b.newBlock()
		b.link(b.cur, right)
		b.enter(right)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)
		b.link(b.cur, join)
		b.enter(join)

	case operator == ast.KindEqualsToken:
		if isDestructuringTarget(binary.Left) {
			// Run-time evaluation order: the right-hand side produces the
			// value first, then the pattern binds element by element.
			b.expr(binary.Right)
			b.patternBind(binary.Left)
			return
		}
		b.patternReads(binary.Left)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)

	case ast.IsAssignmentOperator(operator):
		// Compound assignment reads the target before writing it.
		b.patternReads(binary.Left)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)

	default:
		b.expr(binary.Left)
		b.expr(binary.Right)
	}
}

func (b *Builder[E]) conditionalExpression(node *ast.Node) {
	cond := node.AsConditionalExpression()
	b.expr(cond.Condition)
	testEnd := b.cur
	join := b.newBlock()

	whenTrue := b.newBlock()
	b.link(testEnd, whenTrue)
	b.enter(whenTrue)
	b.expr(cond.WhenTrue)
	b.link(b.cur, join)

	whenFalse := b.newBlock()
	b.link(testEnd, whenFalse)
	b.enter(whenFalse)
	b.expr(cond.WhenFalse)
	b.link(b.cur, join)

	b.enter(join)
}

func (b *Builder[E]) updateExpression(operand *ast.Node) {
	b.patternReads(operand)
	b.patternWrites(operand)
}

// accessOrCall walks a member access or call and forks around each `?.` link so
// the rest of the chain can be skipped.
func (b *Builder[E]) accessOrCall(node *ast.Node) {
	outermost := ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node)
	if outermost {
		b.chainJoins = append(b.chainJoins, b.newBlock())
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		b.expr(node.AsPropertyAccessExpression().Expression)
		b.forkOptionalChain(node)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		b.expr(access.Expression)
		b.forkOptionalChain(node)
		b.expr(access.ArgumentExpression)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		b.expr(call.Expression)
		b.forkOptionalChain(node)
		b.typeArguments(node)
		if call.Arguments != nil {
			for _, arg := range call.Arguments.Nodes {
				b.expr(arg)
			}
		}
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		b.expr(newExpr.Expression)
		b.typeArguments(node)
		if newExpr.Arguments != nil {
			for _, arg := range newExpr.Arguments.Nodes {
				b.expr(arg)
			}
		}
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		b.expr(tagged.Tag)
		b.typeArguments(node)
		b.expr(tagged.Template)
	}

	b.firstThrowableFork()

	if outermost {
		join := b.chainJoins[len(b.chainJoins)-1]
		b.chainJoins = b.chainJoins[:len(b.chainJoins)-1]
		b.link(b.cur, join)
		b.enter(join)
	}
}

func (b *Builder[E]) forkOptionalChain(node *ast.Node) {
	if b.cur == nil || len(b.chainJoins) == 0 || !ast.IsOptionalChainRoot(node) {
		return
	}
	b.link(b.cur, b.chainJoins[len(b.chainJoins)-1])
	next := b.newBlock()
	b.link(b.cur, next)
	b.cur = next
}

// typeArguments walks an explicit `<T>` argument list, so a `typeof x` inside
// it still counts as a read even though nothing in it runs.
func (b *Builder[E]) typeArguments(node *ast.Node) {
	for _, typeArg := range node.TypeArguments() {
		b.expr(typeArg)
	}
}

// typeParameters walks a `<T extends U = D>` declaration list, so a `typeof x`
// inside a constraint or default still counts as a read even though nothing in
// it runs.
func (b *Builder[E]) typeParameters(node *ast.Node) {
	for _, typeParam := range node.TypeParameters() {
		param := typeParam.AsTypeParameterDeclaration()
		if param == nil {
			continue
		}
		b.expr(param.Constraint)
		b.expr(param.DefaultType)
	}
}

// nestedFunction keeps evaluating the parts of a nested function that belong to
// the enclosing code path — a computed member name and decorators — while its
// parameters and body form their own code path root.
func (b *Builder[E]) nestedFunction(node *ast.Node) {
	b.memberHeader(node)
}

func (b *Builder[E]) classLike(node *ast.Node) {
	class := node.ClassLikeData()
	if class == nil {
		return
	}
	b.decorators(node)
	b.typeParameters(node)
	if class.HeritageClauses != nil {
		for _, clause := range class.HeritageClauses.Nodes {
			heritage := clause.AsHeritageClause()
			if heritage == nil || heritage.Types == nil {
				continue
			}
			for _, typeNode := range heritage.Types.Nodes {
				b.expr(typeNode)
			}
		}
	}
	if class.Members != nil {
		for _, member := range class.Members.Nodes {
			b.memberHeader(member)
		}
	}
}

// memberHeader evaluates the parts of a class or object member that belong to
// the enclosing code path: its decorators, its computed key, and its type
// annotation. A member with a body of its own — a method, accessor, or
// property initializer — forms its own code path, so its parameters and
// return type are walked there instead; an index signature has no body, so
// its parameter and return types are walked here.
func (b *Builder[E]) memberHeader(node *ast.Node) {
	b.decorators(node)
	if ast.IsFunctionLikeDeclaration(node) {
		for _, parameter := range node.Parameters() {
			b.decorators(parameter)
		}
	}
	name := node.Name()
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		b.expr(name.AsComputedPropertyName().Expression)
	}
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		b.expr(node.AsPropertyDeclaration().Type)
	case ast.KindIndexSignature:
		for _, parameter := range node.Parameters() {
			b.expr(parameter.Type())
		}
		b.expr(node.Type())
	}
}

func (b *Builder[E]) decorators(node *ast.Node) {
	for _, decorator := range node.Decorators() {
		b.expr(decorator.AsDecorator().Expression)
	}
}
