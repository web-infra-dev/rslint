package prefer_optional_chain

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

type NodeComparisonResult int

const (
	NodeComparisonInvalid NodeComparisonResult = iota
	NodeComparisonEqual
	NodeComparisonSubset
)

func skipDownwards(node *ast.Node) *ast.Node {
	n := node
	for {
		if ast.IsNonNullExpression(n) {
			n = n.Expression()
			continue
		}
		if ast.IsParenthesizedExpression(n) {
			n = n.Expression()
			continue
		}
		break
	}
	return n
}

func nodeText(node *ast.Node) string {
	if ast.IsIdentifier(node) {
		return node.Text()
	}
	return ""
}

func compareNodesUncached(nodeA *ast.Node, nodeB *ast.Node) NodeComparisonResult {
	a := skipDownwards(nodeA)
	b := skipDownwards(nodeB)

	if a.Kind != b.Kind {
		// Check if a is a prefix/subset of b's chain
		if isChainPrefix(a, b) {
			return NodeComparisonSubset
		}
		return NodeComparisonInvalid
	}

	switch a.Kind {
	case ast.KindIdentifier:
		if nodeText(a) == nodeText(b) {
			return NodeComparisonEqual
		}
		return NodeComparisonInvalid

	case ast.KindThisKeyword, ast.KindSuperKeyword,
		// Type keyword nodes used in type arguments (e.g., foo<string>()).
		// Two nodes of the same keyword kind are always structurally equal.
		ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindBooleanKeyword,
		ast.KindAnyKeyword, ast.KindUnknownKeyword, ast.KindVoidKeyword,
		ast.KindNeverKeyword, ast.KindObjectKeyword, ast.KindUndefinedKeyword,
		ast.KindNullKeyword, ast.KindBigIntKeyword, ast.KindSymbolKeyword:
		return NodeComparisonEqual

	case ast.KindPropertyAccessExpression:
		propA := a.AsPropertyAccessExpression()
		propB := b.AsPropertyAccessExpression()

		if nodeText(propA.Name()) != nodeText(propB.Name()) {
			// Names differ - but a might still be a prefix of b
			if isChainPrefix(a, b) {
				return NodeComparisonSubset
			}
			return NodeComparisonInvalid
		}

		// Parenthesized optional chains like (foo?.a).b have different semantics
		// from foo?.a.b because the parens break the optional chain propagation.
		if isParenthesizedOptionalChain(propA.Expression) != isParenthesizedOptionalChain(propB.Expression) {
			return NodeComparisonInvalid
		}

		exprResult := compareNodesUncached(propA.Expression, propB.Expression)
		if exprResult != NodeComparisonEqual {
			return exprResult
		}
		return NodeComparisonEqual

	case ast.KindAwaitExpression:
		return compareNodesUncached(a.AsAwaitExpression().Expression, b.AsAwaitExpression().Expression)

	case ast.KindElementAccessExpression:
		elemA := a.AsElementAccessExpression()
		elemB := b.AsElementAccessExpression()

		if isParenthesizedOptionalChain(elemA.Expression) != isParenthesizedOptionalChain(elemB.Expression) {
			return NodeComparisonInvalid
		}

		// Check if a is a prefix of b first (different arguments)
		argResult := compareNodesUncached(elemA.ArgumentExpression, elemB.ArgumentExpression)
		if argResult != NodeComparisonEqual {
			if isChainPrefix(a, b) {
				return NodeComparisonSubset
			}
			return NodeComparisonInvalid
		}

		exprResult := compareNodesUncached(elemA.Expression, elemB.Expression)
		if exprResult != NodeComparisonEqual {
			return exprResult
		}
		return NodeComparisonEqual

	case ast.KindCallExpression:
		callA := a.AsCallExpression()
		callB := b.AsCallExpression()

		if isParenthesizedOptionalChain(callA.Expression) != isParenthesizedOptionalChain(callB.Expression) {
			return NodeComparisonInvalid
		}

		// Compare type arguments (e.g., foo<a>() vs foo<a, b>())
		var typeArgsA, typeArgsB []*ast.Node
		if callA.TypeArguments != nil {
			typeArgsA = callA.TypeArguments.Nodes
		}
		if callB.TypeArguments != nil {
			typeArgsB = callB.TypeArguments.Nodes
		}
		if len(typeArgsA) != len(typeArgsB) {
			if isChainPrefix(a, b) {
				return NodeComparisonSubset
			}
			return NodeComparisonInvalid
		}
		for i := range typeArgsA {
			taResult := compareNodesUncached(typeArgsA[i], typeArgsB[i])
			if taResult != NodeComparisonEqual {
				if isChainPrefix(a, b) {
					return NodeComparisonSubset
				}
				return NodeComparisonInvalid
			}
		}

		var argsA, argsB []*ast.Node
		if callA.Arguments != nil {
			argsA = callA.Arguments.Nodes
		}
		if callB.Arguments != nil {
			argsB = callB.Arguments.Nodes
		}
		if len(argsA) != len(argsB) {
			if isChainPrefix(a, b) {
				return NodeComparisonSubset
			}
			return NodeComparisonInvalid
		}

		for i := range argsA {
			argResult := compareNodesUncached(argsA[i], argsB[i])
			if argResult != NodeComparisonEqual {
				if isChainPrefix(a, b) {
					return NodeComparisonSubset
				}
				return NodeComparisonInvalid
			}
		}

		exprResult := compareNodesUncached(callA.Expression, callB.Expression)
		if exprResult != NodeComparisonEqual {
			return exprResult
		}
		return NodeComparisonEqual

	case ast.KindMetaProperty:
		// MetaProperty: new.target, import.meta
		metaA := a.AsMetaProperty()
		metaB := b.AsMetaProperty()
		if metaA.KeywordToken == metaB.KeywordToken &&
			metaA.Name().Text() == metaB.Name().Text() {
			return NodeComparisonEqual
		}
		return NodeComparisonInvalid

	case ast.KindTypeReference:
		refA := a.AsTypeReferenceNode()
		refB := b.AsTypeReferenceNode()
		nameResult := compareNodesUncached(refA.TypeName, refB.TypeName)
		if nameResult != NodeComparisonEqual {
			return NodeComparisonInvalid
		}
		var typeArgsA, typeArgsB []*ast.Node
		if refA.TypeArguments != nil {
			typeArgsA = refA.TypeArguments.Nodes
		}
		if refB.TypeArguments != nil {
			typeArgsB = refB.TypeArguments.Nodes
		}
		if len(typeArgsA) != len(typeArgsB) {
			return NodeComparisonInvalid
		}
		for i := range typeArgsA {
			if compareNodesUncached(typeArgsA[i], typeArgsB[i]) != NodeComparisonEqual {
				return NodeComparisonInvalid
			}
		}
		return NodeComparisonEqual

	case ast.KindQualifiedName:
		qualA := a.AsQualifiedName()
		qualB := b.AsQualifiedName()
		leftResult := compareNodesUncached(qualA.Left, qualB.Left)
		if leftResult != NodeComparisonEqual {
			return NodeComparisonInvalid
		}
		return compareNodesUncached(qualA.Right, qualB.Right)

	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if a.Text() == b.Text() {
			return NodeComparisonEqual
		}
		return NodeComparisonInvalid

	case ast.KindTemplateExpression,
		ast.KindBinaryExpression,
		ast.KindNewExpression,
		ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindYieldExpression:
		return NodeComparisonInvalid
	}

	return NodeComparisonInvalid
}

func isChainPrefix(prefix *ast.Node, chain *ast.Node) bool {
	p := skipDownwards(prefix)
	c := skipDownwards(chain)

	// Walk down the left spine of the chain until we find a match
	for {
		switch c.Kind {
		case ast.KindPropertyAccessExpression:
			prop := c.AsPropertyAccessExpression()
			if compareNodesUncached(p, prop.Expression) == NodeComparisonEqual {
				return true
			}
			c = skipDownwards(prop.Expression)
			continue

		case ast.KindElementAccessExpression:
			elem := c.AsElementAccessExpression()
			if compareNodesUncached(p, elem.Expression) == NodeComparisonEqual {
				return true
			}
			c = skipDownwards(elem.Expression)
			continue

		case ast.KindCallExpression:
			call := c.AsCallExpression()
			if compareNodesUncached(p, call.Expression) == NodeComparisonEqual {
				return true
			}
			c = skipDownwards(call.Expression)
			continue
		}
		break
	}
	return false
}

// isParenthesizedOptionalChain checks if a node is a ParenthesizedExpression
// wrapping an optional chain access (e.g., `(foo?.bar)`). Parentheses break
// optional chain propagation, so `(foo?.a).b` has different semantics from
// `foo?.a.b` and they should not be compared as equal.
func isParenthesizedOptionalChain(node *ast.Node) bool {
	if !ast.IsParenthesizedExpression(node) {
		return false
	}
	inner := node.Expression()
	switch inner.Kind {
	case ast.KindPropertyAccessExpression:
		return inner.AsPropertyAccessExpression().QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		return inner.AsElementAccessExpression().QuestionDotToken != nil
	case ast.KindCallExpression:
		return inner.AsCallExpression().QuestionDotToken != nil
	}
	return false
}
