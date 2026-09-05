package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

type destructuringStep struct {
	index *int
	key   string
}

// StaticBindingForValue maps a value in a static array/object literal
// initializer to the identifier that receives it through a destructuring
// declaration or assignment. Dynamic spreads, computed keys, rest bindings,
// and ambiguous duplicate object keys return nil.
func StaticBindingForValue(value *ast.Node) *ast.Node {
	if value == nil {
		return nil
	}
	current := ast.SkipParentheses(value)
	var steps []destructuringStep

	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent

		case ast.KindArrayLiteralExpression:
			index, ok := staticArrayElementIndex(parent, current)
			if !ok {
				return nil
			}
			indexCopy := index
			steps = append(steps, destructuringStep{index: &indexCopy})
			current = parent

		case ast.KindPropertyAssignment:
			property := parent.AsPropertyAssignment()
			if property == nil || ast.SkipParentheses(property.Initializer) != current ||
				parent.Parent == nil || parent.Parent.Kind != ast.KindObjectLiteralExpression {
				return nil
			}
			key, ok := staticPropertyName(property.Name())
			if !ok || !objectLiteralKeyIsUnambiguous(parent.Parent, parent, key) {
				return nil
			}
			steps = append(steps, destructuringStep{key: key})
			current = parent.Parent

		case ast.KindVariableDeclaration:
			declaration := parent.AsVariableDeclaration()
			if declaration == nil ||
				ast.SkipParentheses(declaration.Initializer) != current ||
				len(steps) == 0 {
				return nil
			}
			return applyDestructuringSteps(declaration.Name(), steps)

		default:
			if ast.IsAssignmentExpression(parent, true) {
				assignment := parent.AsBinaryExpression()
				if assignment == nil ||
					assignment.OperatorToken.Kind != ast.KindEqualsToken ||
					ast.SkipParentheses(assignment.Right) != current ||
					len(steps) == 0 {
					return nil
				}
				return applyDestructuringSteps(assignment.Left, steps)
			}
			return nil
		}
	}
	return nil
}

func staticArrayElementIndex(arrayNode, value *ast.Node) (int, bool) {
	array := arrayNode.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return 0, false
	}
	index := -1
	for elementIndex, element := range array.Elements.Nodes {
		if element == nil || element.Kind == ast.KindSpreadElement {
			return 0, false
		}
		if ast.SkipParentheses(element) == value {
			index = elementIndex
		}
	}
	return index, index >= 0
}

func objectLiteralKeyIsUnambiguous(objectNode, propertyNode *ast.Node, key string) bool {
	object := objectNode.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return false
	}
	count := 0
	for _, property := range object.Properties.Nodes {
		if property == nil || property.Kind == ast.KindSpreadAssignment {
			return false
		}
		switch property.Kind {
		case ast.KindPropertyAssignment:
			name, ok := staticPropertyName(property.AsPropertyAssignment().Name())
			if !ok {
				return false
			}
			if name == key {
				count++
				if property != propertyNode {
					return false
				}
			}
		case ast.KindShorthandPropertyAssignment:
			name := property.AsShorthandPropertyAssignment().Name()
			if name == nil {
				return false
			}
			if name.Text() == key {
				return false
			}
		default:
			return false
		}
	}
	return count == 1
}

func applyDestructuringSteps(target *ast.Node, steps []destructuringStep) *ast.Node {
	current := ast.SkipParentheses(target)
	for index := len(steps) - 1; index >= 0; index-- {
		step := steps[index]
		if step.index != nil {
			current = arrayTargetAt(current, *step.index)
		} else {
			current = objectTargetAt(current, step.key)
		}
		if current == nil {
			return nil
		}
	}
	current = unwrapAssignmentDefault(current)
	if current != nil && current.Kind == ast.KindIdentifier {
		return current
	}
	return nil
}

func arrayTargetAt(pattern *ast.Node, index int) *ast.Node {
	pattern = ast.SkipParentheses(pattern)
	if pattern == nil {
		return nil
	}
	switch pattern.Kind {
	case ast.KindArrayBindingPattern:
		bindingPattern := pattern.AsBindingPattern()
		if bindingPattern == nil || bindingPattern.Elements == nil ||
			index < 0 || index >= len(bindingPattern.Elements.Nodes) {
			return nil
		}
		element := bindingPattern.Elements.Nodes[index]
		if element == nil || element.Kind != ast.KindBindingElement {
			return nil
		}
		binding := element.AsBindingElement()
		if binding == nil || binding.DotDotDotToken != nil {
			return nil
		}
		return binding.Name()
	case ast.KindArrayLiteralExpression:
		array := pattern.AsArrayLiteralExpression()
		if array == nil || array.Elements == nil ||
			index < 0 || index >= len(array.Elements.Nodes) {
			return nil
		}
		element := ast.SkipParentheses(array.Elements.Nodes[index])
		if element == nil || element.Kind == ast.KindOmittedExpression ||
			element.Kind == ast.KindSpreadElement {
			return nil
		}
		return unwrapAssignmentDefault(element)
	default:
		return nil
	}
}

func objectTargetAt(pattern *ast.Node, key string) *ast.Node {
	pattern = ast.SkipParentheses(pattern)
	if pattern == nil {
		return nil
	}
	switch pattern.Kind {
	case ast.KindObjectBindingPattern:
		bindingPattern := pattern.AsBindingPattern()
		if bindingPattern == nil || bindingPattern.Elements == nil {
			return nil
		}
		for _, element := range bindingPattern.Elements.Nodes {
			if element == nil || element.Kind != ast.KindBindingElement {
				continue
			}
			binding := element.AsBindingElement()
			if binding == nil || binding.DotDotDotToken != nil {
				continue
			}
			nameNode := binding.PropertyName
			if nameNode == nil {
				nameNode = binding.Name()
			}
			name, ok := staticPropertyName(nameNode)
			if ok && name == key {
				return binding.Name()
			}
		}
	case ast.KindObjectLiteralExpression:
		object := pattern.AsObjectLiteralExpression()
		if object == nil || object.Properties == nil {
			return nil
		}
		for _, property := range object.Properties.Nodes {
			if property == nil {
				continue
			}
			switch property.Kind {
			case ast.KindPropertyAssignment:
				assignment := property.AsPropertyAssignment()
				name, ok := staticPropertyName(assignment.Name())
				if ok && name == key {
					return unwrapAssignmentDefault(assignment.Initializer)
				}
			case ast.KindShorthandPropertyAssignment:
				shorthand := property.AsShorthandPropertyAssignment()
				if shorthand == nil || shorthand.Name() == nil ||
					shorthand.Name().Text() != key {
					continue
				}
				return shorthand.Name()
			case ast.KindSpreadAssignment:
				return nil
			}
		}
	}
	return nil
}

func unwrapAssignmentDefault(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return node
	}
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
		return node
	}
	return ast.SkipParentheses(binary.Left)
}

func staticPropertyName(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return "", false
	}
	computed := false
	if node.Kind == ast.KindComputedPropertyName {
		computed = true
		node = ast.SkipParentheses(node.AsComputedPropertyName().Expression)
	}
	if node == nil {
		return "", false
	}
	if value, ok := internalUtils.GetStaticStringLiteralValue(node); ok {
		return value, true
	}
	if computed {
		if node.Kind == ast.KindNumericLiteral {
			return internalUtils.NormalizeNumericLiteral(node.AsNumericLiteral().Text), true
		}
		return "", false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text(), true
	case ast.KindNumericLiteral:
		return internalUtils.NormalizeNumericLiteral(node.AsNumericLiteral().Text), true
	}
	return "", false
}
