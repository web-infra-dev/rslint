package no_useless_else

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Add a possible throw for operations the graph misses. Analyze the resulting
// consequent as a whole so break, continue, and finally keep their context.
func (a *exitAnalysis) withUnmodeledExceptions(consequent *ast.Node) *ast.Node {
	if !containsTryStatement(consequent) {
		return consequent
	}

	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{
		OnUpdate: func(updated, original *ast.Node) {
			if ast.IsLabeledStatement(updated) && updated.Statement() != original.Statement() {
				// The graph reads labels from the statement's parent chain.
				updated.Statement().Parent = updated
			}
		},
	})
	withPossibleThrow := func(block *ast.Node) *ast.Node {
		// The omitted expression lets the graph explore both outcomes.
		branch := factory.NewIfStatement(factory.NewOmittedExpression(),
			factory.NewThrowStatement(nil), nil)
		original := block.AsBlock()
		body := block.Statements()
		statements := make([]*ast.Node, 1, len(body)+1)
		statements[0] = branch
		statements = append(statements, body...)
		return factory.UpdateBlock(original, factory.NewNodeList(statements), original.MultiLine)
	}
	visitor := ast.NewNodeVisitor(nil, factory, ast.NodeVisitorHooks{})
	visitor.Visit = func(node *ast.Node) *ast.Node {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) {
			return node
		}
		updated := node.VisitEachChild(visitor)
		if !ast.IsTryStatement(node) {
			return updated
		}

		original := node.AsTryStatement()
		stmt := updated.AsTryStatement()
		tryBlock, catchClause := stmt.TryBlock, stmt.CatchClause
		if a.mayThrowOutsideGraph(original.TryBlock) {
			// A possible throw at entry conservatively makes the handler reachable.
			tryBlock = withPossibleThrow(tryBlock)
		}
		if a.mayThrowOutsideGraph(original.CatchClause) {
			catch := catchClause.AsCatchClause()
			catchClause = factory.UpdateCatchClause(catch, catch.VariableDeclaration, withPossibleThrow(catch.Block))
		}
		return factory.UpdateTryStatement(stmt, tryBlock, catchClause, stmt.FinallyBlock)
	}
	return visitor.VisitNode(consequent)
}

func containsTryStatement(node *ast.Node) bool {
	return ast.IsTryStatement(node) ||
		!ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) && node.ForEachChild(containsTryStatement)
}

func (a *exitAnalysis) mayThrowOutsideGraph(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsFunctionLike(node) {
		name := node.Name()
		return name != nil && name.Kind == ast.KindComputedPropertyName &&
			a.mayThrowOutsideGraph(name.Expression())
	}
	switch node.Kind {
	case ast.KindBlock:
		for _, stmt := range node.Statements() {
			if a.mayThrowOutsideGraph(stmt) {
				return true
			}
			switch stmt.Kind {
			case ast.KindReturnStatement, ast.KindThrowStatement, ast.KindBreakStatement, ast.KindContinueStatement:
				return false
			case ast.KindExpressionStatement:
				if a.simpleExitExpression(stmt.Expression()) {
					return false
				}
			}
		}
		return false
	case ast.KindForInStatement, ast.KindWithStatement:
		// Enumeration and with-object coercion happen before the body.
		return !a.simpleExitExpression(node.Expression())
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return a.jsxMayThrow(node)
	case ast.KindBinaryExpression:
		if ast.IsDestructuringAssignment(node) {
			return !a.simpleExitExpression(node.AsBinaryExpression().Right)
		}
	case ast.KindCatchClause:
		decl := node.AsCatchClause().VariableDeclaration
		if decl != nil && decl.Name() != nil && ast.IsBindingPattern(decl.Name()) {
			return true
		}
	case ast.KindThisKeyword:
		// Arrows can capture the constructor's still-uninitialized this.
		container := ast.GetThisContainer(node, false, false)
		return container.Kind == ast.KindConstructor && container.Parent != nil &&
			ast.GetClassExtendsHeritageElement(container.Parent) != nil
	case ast.KindTemplateExpression:
		// Untagged templates coerce each value before the next substitution.
		// A tag receives the substitution values directly.
		if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
			break
		}
		for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
			expression := span.Expression()
			if a.simpleExitExpression(expression) {
				return false
			}
			if !a.safeTemplateSubstitution(expression) {
				return true
			}
		}
		return false
	case ast.KindIfStatement:
		stmt := node.AsIfStatement()
		if truthy, known := a.truthiness(stmt.Expression); known {
			if truthy {
				return a.mayThrowOutsideGraph(stmt.ThenStatement)
			}
			return a.mayThrowOutsideGraph(stmt.ElseStatement)
		}
	}
	return node.ForEachChild(a.mayThrowOutsideGraph)
}

func (a *exitAnalysis) jsxMayThrow(node *ast.Node) bool {
	var tag *ast.Node
	var attributes *ast.Node
	switch node.Kind {
	case ast.KindJsxElement:
		element := node.AsJsxElement()
		if element.OpeningElement != nil {
			opening := element.OpeningElement.AsJsxOpeningElement()
			tag, attributes = opening.TagName, opening.Attributes
		}
	case ast.KindJsxSelfClosingElement:
		element := node.AsJsxSelfClosingElement()
		tag, attributes = element.TagName, element.Attributes
	}
	// A component reference can throw before its props are evaluated.
	if tag != nil && (tag.Kind != ast.KindIdentifier || !scanner.IsIntrinsicJsxName(tag.Text())) {
		return true
	}
	// The generated factory call runs after its arguments. A terminal prop
	// reached after safe props prevents the factory from running.
	if attributes != nil {
		props := attributes.AsJsxAttributes().Properties
		if props != nil {
			for _, prop := range props.Nodes {
				if prop.Kind != ast.KindJsxAttribute {
					break
				}
				value := prop.Initializer()
				if value == nil || value.Kind == ast.KindStringLiteral {
					continue
				}
				if value.Kind != ast.KindJsxExpression {
					break
				}
				expression := value.Expression()
				if a.simpleExitExpression(expression) {
					return false
				}
				if !a.safePrimitiveRead(expression) {
					break
				}
			}
		}
	}
	return true
}

func (a *exitAnalysis) safeTemplateSubstitution(expression *ast.Node) bool {
	if !a.safePrimitiveRead(expression) {
		return false
	}
	value, known := a.evaluator().EvalControlFlowValue(expression)
	return known && utils.IsPrimitiveStaticValue(value)
}
