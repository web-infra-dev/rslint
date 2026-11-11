package constructor_super

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Facts describe the value on successful evaluation; they never skip effects.
// An unknown incoming value keeps both default branches. In particular, absent
// properties of an ordinary object can be supplied by its mutable prototype.
type inputValue struct {
	node                    *ast.Node
	undefined, nonUndefined bool
	safeGet                 bool
}

func valueFact(n *ast.Node) inputValue {
	if n == nil {
		return inputValue{}
	}
	n = ast.SkipOuterExpressions(n, ast.OEKAll)
	v := inputValue{node: n}
	switch n.Kind {
	case ast.KindVoidExpression:
		v.undefined = true
	case ast.KindNullKeyword, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword,
		ast.KindRegularExpressionLiteral, ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression, ast.KindNewExpression, ast.KindTypeOfExpression:
		v.nonUndefined = true
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		if b.OperatorToken.Kind == ast.KindCommaToken || b.OperatorToken.Kind == ast.KindEqualsToken {
			return valueFact(b.Right)
		}
	}
	return v
}
func (a *analyzer) propertyInput(v inputValue, key string) inputValue {
	if v.node == nil || v.node.Kind != ast.KindObjectLiteralExpression {
		return inputValue{}
	}
	result := inputValue{}
	for _, p := range v.node.AsObjectLiteralExpression().Properties.Nodes {
		if p.Kind == ast.KindSpreadAssignment {
			result = inputValue{}
			continue
		}
		name := p.Name()
		if name == nil {
			continue
		}
		text, known := a.propertyKey(name)
		if !known {
			result = inputValue{}
			continue
		}
		if text != key {
			continue
		}
		// A non-computed __proto__ property sets the prototype rather than
		// defining an own data property. A computed one is ordinary data.
		if text == "__proto__" && name.Kind != ast.KindComputedPropertyName && p.Kind == ast.KindPropertyAssignment {
			result = inputValue{}
			continue
		}
		if p.Kind == ast.KindPropertyAssignment {
			result = valueFact(p.AsPropertyAssignment().Initializer)
			result.safeGet = true
		} else {
			result = inputValue{}
		}
	}
	return result
}
func (a *analyzer) binding(n *ast.Node, s paths, declaration bool) paths {
	return a.bindingValue(n, s, declaration, inputValue{})
}
func (a *analyzer) bindingDefault(n, init *ast.Node, s paths, declaration bool) paths {
	return a.bindDefault(n, init, s, declaration, inputValue{})
}
func (a *analyzer) bindDefault(n, init *ast.Node, s paths, declaration bool, v inputValue) paths {
	if init == nil || v.nonUndefined {
		return a.bindingValue(n, s, declaration, v)
	}
	fallback := a.expression(init, s)
	fallback = a.bindingValue(n, fallback, declaration, valueFact(init))
	if v.undefined {
		return fallback
	}
	return a.bindingValue(n, s, declaration, v) | fallback
}

func (a *analyzer) bindingValue(n *ast.Node, s paths, declaration bool, v inputValue) paths {
	if n == nil || s == 0 {
		return s
	}
	n = ast.SkipOuterExpressions(n, ast.OEKAll)
	switch n.Kind {
	case ast.KindIdentifier:
		if !declaration {
			a.mayThrow(s)
		}
		return s
	case ast.KindOmittedExpression:
		return s
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		if v.undefined || v.node != nil && v.node.Kind == ast.KindNullKeyword {
			a.mayThrow(s)
			return 0
		}
		// Arrays use an iterator even for literal values. Object bindings
		// reject null/undefined, then Get each property after its key.
		if n.Kind == ast.KindArrayBindingPattern || !v.nonUndefined {
			a.mayThrow(s)
		}
		for _, element := range n.AsBindingPattern().Elements.Nodes {
			if element.Kind != ast.KindBindingElement {
				continue
			}
			b := element.AsBindingElement()
			incoming := inputValue{}
			if n.Kind == ast.KindObjectBindingPattern {
				key := b.PropertyName
				if key == nil {
					key = b.Name()
				}
				if key != nil && key.Kind == ast.KindComputedPropertyName {
					s = a.expression(key.Expression(), s)
				}
				if text, known := a.propertyKey(key); known {
					incoming = a.propertyInput(v, text)
				}
			}
			if n.Kind == ast.KindObjectBindingPattern && !incoming.safeGet {
				a.mayThrow(s)
			}
			if b.Initializer != nil {
				s = a.bindDefault(b.Name(), b.Initializer, s, declaration, incoming)
			} else {
				s = a.bindingValue(b.Name(), s, declaration, incoming)
			}
		}
		return s
	case ast.KindObjectLiteralExpression:
		if v.undefined || v.node != nil && v.node.Kind == ast.KindNullKeyword {
			a.mayThrow(s)
			return 0
		}
		if !v.nonUndefined {
			a.mayThrow(s)
		}
		for _, p := range n.AsObjectLiteralExpression().Properties.Nodes {
			incoming := inputValue{}
			if name := p.Name(); name != nil {
				if name.Kind == ast.KindComputedPropertyName {
					s = a.expression(name.Expression(), s)
				}
				if text, known := a.propertyKey(name); known {
					incoming = a.propertyInput(v, text)
				}
			}
			switch p.Kind {
			case ast.KindPropertyAssignment:
				s = a.assignmentElement(p.AsPropertyAssignment().Initializer, s, incoming)
			case ast.KindShorthandPropertyAssignment:
				x := p.AsShorthandPropertyAssignment()
				if !incoming.safeGet {
					a.mayThrow(s)
				}
				s = a.bindDefault(x.Name(), x.ObjectAssignmentInitializer, s, false, incoming)
			case ast.KindSpreadAssignment:
				s = a.assignmentElement(p.Expression(), s, inputValue{})
			}
		}
		return s
	case ast.KindArrayLiteralExpression:
		a.mayThrow(s)
		if v.undefined || v.node != nil && v.node.Kind == ast.KindNullKeyword {
			return 0
		}
		for _, element := range n.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement {
				element = element.Expression()
			}
			s = a.assignmentElement(element, s, inputValue{})
		}
		return s
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		if b.OperatorToken.Kind == ast.KindEqualsToken {
			return a.bindDefault(b.Left, b.Right, s, false, v)
		}
	case ast.KindSpreadElement:
		return a.binding(n.Expression(), s, declaration)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		s = a.reference(n, s)
		a.mayThrow(s)
		return s
	}
	return a.expression(n, s)
}

func (a *analyzer) propertyKey(n *ast.Node) (string, bool) {
	if n == nil {
		return "", false
	}
	if n.Kind == ast.KindIdentifier || n.Kind == ast.KindStringLiteral {
		return n.Text(), true
	}
	if a.truth == nil {
		a.truth = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return a.truth.EvalPropertyName(n)
}

// Assignment patterns prepare a scalar reference before getting the element
// value or evaluating its default. Nested patterns bind after that value.
func (a *analyzer) assignmentElement(n *ast.Node, s paths, v inputValue) paths {
	if n == nil || s == 0 || n.Kind == ast.KindOmittedExpression {
		return s
	}
	target := ast.SkipOuterExpressions(n, ast.OEKAll)
	var init *ast.Node
	if target.Kind == ast.KindBinaryExpression && target.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		init = target.AsBinaryExpression().Right
		target = ast.SkipOuterExpressions(target.AsBinaryExpression().Left, ast.OEKAll)
	}
	if target.Kind != ast.KindIdentifier && target.Kind != ast.KindPropertyAccessExpression && target.Kind != ast.KindElementAccessExpression {
		if !v.safeGet {
			a.mayThrow(s)
		}
		return a.bindDefault(target, init, s, false, v)
	}
	s = a.reference(target, s)
	if !v.safeGet {
		a.mayThrow(s)
	}
	if init != nil && !v.nonUndefined {
		fallback := a.expression(init, s)
		if v.undefined {
			s = fallback
		} else {
			s |= fallback
		}
	}
	a.mayThrow(s)
	return s
}
