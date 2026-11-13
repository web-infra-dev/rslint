package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

const staticCallEscapeFlags = staticReferencePropertyMutation | staticReferenceNestedMutation | staticReferenceCallEscape

// Share expression results only while the mutation facts remain unchanged.
// This avoids repeatedly evaluating long call chains during the reference scan,
// without letting its partial facts or values escape into ordinary evaluation.
func (staticEvaluator *StaticStringEvaluator) invalidateMutationValues() {
	if staticEvaluator.evaluation != nil {
		staticEvaluator.evaluation.values = make(map[*ast.Node]staticEvalResult)
		staticEvaluator.evaluation.optionalShortCircuits = make(map[*ast.Node]bool)
	}
}

// Unknown calls may retain and mutate the receiver and every reference passed
// to them. No function-body execution is needed: invalidate reachable values
// while retaining primitive bindings and operations already proven static.
func (staticEvaluator *StaticStringEvaluator) markCallEscapes(node *ast.Node) {
	if staticEvaluator.evalValue(node).ok {
		return
	}
	if node.Kind == ast.KindTaggedTemplateExpression {
		tagged := node.AsTaggedTemplateExpression()
		staticEvaluator.markCallReceiver(tagged.Tag, staticCallEscapeFlags|staticReferenceCallReceiver)
		if tagged.Template.Kind == ast.KindTemplateExpression {
			for _, span := range tagged.Template.AsTemplateExpression().TemplateSpans.Nodes {
				staticEvaluator.markReference(span.AsTemplateSpan().Expression, staticCallEscapeFlags)
			}
		}
		return
	}
	callee := SkipAssertionsAndParens(node.Expression())
	if node.Kind == ast.KindCallExpression && (staticEvaluator.isKnownObjectOperation(node, callee) || staticEvaluator.isStaticArraySlice(node, callee) ||
		staticEvaluator.isEmptyFunctionCall(callee, make(map[*ast.Symbol]bool))) {
		return
	}
	staticEvaluator.markCallReceiver(callee, staticCallEscapeFlags|staticReferenceCallReceiver)
	for _, argument := range node.Arguments() {
		if ast.IsSpreadElement(argument) {
			staticEvaluator.markCopiedReferences(argument.Expression(), staticCallEscapeFlags)
		} else {
			staticEvaluator.markReference(argument, staticCallEscapeFlags)
		}
	}
}

func (staticEvaluator *StaticStringEvaluator) markCallReceiver(callee *ast.Node, flag staticReferenceFlags) {
	callee = SkipAssertionsAndParens(callee)
	if ast.IsAccessExpression(callee) {
		staticEvaluator.markReference(AccessExpressionObject(callee), flag)
	} else {
		staticEvaluator.markReference(callee, flag)
	}
}

func (staticEvaluator *StaticStringEvaluator) isStaticArraySlice(call, callee *ast.Node) bool {
	if !ast.IsAccessExpression(callee) {
		return false
	}
	name, known := staticEvaluator.evalAccessExpressionKey(callee)
	if !known || name != "slice" || staticEvaluator.arrayPrototypeMutated {
		return false
	}
	value := staticEvaluator.evalValue(AccessExpressionObject(callee))
	if _, array := value.value.(*staticArrayValue); !value.ok || !array {
		return false
	}
	for _, argument := range call.Arguments() {
		value := staticEvaluator.evalValue(argument)
		if ast.IsSpreadElement(argument) || !value.ok || !IsPrimitiveStaticValue(value.value) {
			return false
		}
	}
	return true
}

// Retain the existing target-only mutation policy for these known operations.
// Aliases must resolve to the actual built-in, not just a matching method name.
func (staticEvaluator *StaticStringEvaluator) isKnownObjectOperation(call, callee *ast.Node) bool {
	for _, builtin := range []string{"Object", "Reflect"} {
		name, known := staticEvaluator.builtinMethodName(callee, builtin, make(map[*ast.Symbol]bool))
		if !known {
			continue
		}
		if builtin == "Object" && objectPassThroughMethods[name] {
			return true
		}
		switch name {
		case "defineProperty", "setPrototypeOf":
		case "defineProperties", "assign":
			if builtin != "Object" {
				return false
			}
		case "deleteProperty", "set":
			if builtin != "Reflect" {
				return false
			}
		default:
			return false
		}
		staticEvaluator.markMutationArgument(call.Arguments(), 0)
		if builtin == "Reflect" && name == "set" {
			staticEvaluator.markMutationArgument(call.Arguments(), 3)
		}
		return true
	}
	return false
}

// Preserve the existing empty-function case without assuming a method is pure
// merely because its name or TypeScript signature resembles a built-in.
func (staticEvaluator *StaticStringEvaluator) isEmptyFunctionCall(node *ast.Node, visiting map[*ast.Symbol]bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		symbol := staticEvaluator.referenceSymbol(node)
		if symbol == nil || visiting[symbol] || len(symbol.Declarations) != 1 || staticEvaluator.hasWrites(symbol) {
			return false
		}
		visiting[symbol] = true
		defer delete(visiting, symbol)
		node = symbol.Declarations[0]
		if node.Kind == ast.KindVariableDeclaration {
			return staticEvaluator.isEmptyFunctionCall(node.AsVariableDeclaration().Initializer, visiting)
		}
	}
	if ast.IsAccessExpression(node) {
		name, known := staticEvaluator.evalAccessExpressionKey(node)
		object := SkipAssertionsAndParens(AccessExpressionObject(node))
		if !known {
			return false
		}
		for ast.IsIdentifier(object) {
			initializer, symbol, known := staticEvaluator.resolveIdentifierInitializer(object)
			if !known || visiting[symbol] || staticEvaluator.hasPropertyMutation(symbol) {
				return false
			}
			visiting[symbol] = true
			object = SkipAssertionsAndParens(initializer)
		}
		if object == nil || object.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		// Duplicate keys and spreads can replace the method at runtime.
		var method *ast.Node
		for _, property := range object.AsObjectLiteralExpression().Properties.Nodes {
			key, known := staticEvaluator.EvalPropertyName(property.Name())
			if !known || property.Kind == ast.KindSpreadAssignment {
				return false
			}
			if key == name {
				if method != nil {
					return false
				}
				method = property
				if property.Kind == ast.KindPropertyAssignment {
					method = SkipAssertionsAndParens(property.AsPropertyAssignment().Initializer)
				}
			}
		}
		return staticEvaluator.isEmptyFunctionCall(method, visiting)
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
		body := node.Body()
		if body == nil || body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 0 {
			return false
		}
		for _, parameter := range node.Parameters() {
			p := parameter.AsParameterDeclaration()
			if !ast.IsIdentifier(p.Name()) || p.Initializer != nil || len(parameter.Decorators()) != 0 {
				return false
			}
		}
		return true
	}
	return false
}

// Shallow copies retain nested references, but never the source container or
// its primitive elements. This also keeps primitive spread arguments stable.
func (staticEvaluator *StaticStringEvaluator) hasOnlyPrimitiveChildren(node *ast.Node) bool {
	value := staticEvaluator.evalValue(node)
	if !value.ok {
		return false
	}
	switch value := value.value.(type) {
	case *staticArrayValue:
		for i := range value.length {
			if element := value.element(i); element != nil && !IsPrimitiveStaticValue(element) {
				return false
			}
		}
		return true
	case *staticObjectValue:
		if value.propertyCount > 0 && !IsPrimitiveStaticValue(value.property.value) {
			return false
		}
		for _, property := range value.extraProperties {
			if !IsPrimitiveStaticValue(property.value) {
				return false
			}
		}
		return true
	}
	return IsPrimitiveStaticValue(value.value)
}

func (staticEvaluator *StaticStringEvaluator) markCopiedReferences(node *ast.Node, flag staticReferenceFlags) {
	if flag&staticReferenceCallEscape != 0 && staticEvaluator.hasOnlyPrimitiveChildren(node) {
		return
	}
	staticEvaluator.markReference(node, flag)
}

// Escaped callbacks and function results can expose references held in their
// closure. Follow return values, not unrelated reads or nested function bodies.
func (staticEvaluator *StaticStringEvaluator) markReturnedReferences(node *ast.Node, flag staticReferenceFlags) {
	if node == nil {
		return
	}
	flag &^= staticReferenceCallReceiver
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		for _, member := range node.Members() {
			if member.Kind == ast.KindPropertyDeclaration {
				staticEvaluator.markReference(member.AsPropertyDeclaration().Initializer, flag)
			} else {
				staticEvaluator.markReturnedReferences(member, flag)
			}
		}
		return
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindConstructor:
	default:
		return
	}
	body := node.Body()
	if body == nil {
		return
	}
	if body.Kind != ast.KindBlock {
		staticEvaluator.markReference(body, flag)
		return
	}
	var visit func(*ast.Node)
	visit = func(child *ast.Node) {
		if child.Kind == ast.KindReturnStatement || child.Kind == ast.KindThrowStatement || child.Kind == ast.KindYieldExpression {
			staticEvaluator.markReference(child.Expression(), flag)
			return
		}
		if child.Kind == ast.KindBinaryExpression {
			binary := child.AsBinaryExpression()
			if ast.IsAccessExpression(SkipAssertionsAndParens(binary.Left)) && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				staticEvaluator.markReference(binary.Right, flag)
			}
		}
		if child != body && child.Body() != nil {
			return
		}
		child.ForEachChild(func(next *ast.Node) bool { visit(next); return false })
	}
	visit(body)
}

// Index references stored after an object's initializer, including writes made
// through aliases. The index is built once with the other reference facts.
func (staticEvaluator *StaticStringEvaluator) recordStoredReferences(target, value *ast.Node) {
	visited := make(map[*ast.Symbol]bool)
	var visit func(*ast.Node, bool)
	visit = func(node *ast.Node, nested bool) {
		node = SkipAssertionsAndParens(node)
		if node == nil {
			return
		}
		if ast.IsIdentifier(node) {
			symbol := staticEvaluator.referenceSymbol(node)
			if symbol == nil || visited[symbol] {
				return
			}
			visited[symbol] = true
			if staticEvaluator.storedReferences == nil {
				staticEvaluator.storedReferences = make(map[*ast.Symbol][]*ast.Node)
			}
			staticEvaluator.storedReferences[symbol] = append(staticEvaluator.storedReferences[symbol], value)
			if len(symbol.Declarations) == 1 {
				declaration := symbol.Declarations[0]
				switch declaration.Kind {
				case ast.KindVariableDeclaration:
					visit(declaration.AsVariableDeclaration().Initializer, nested)
				case ast.KindBindingElement:
					for binding := declaration; binding.Kind == ast.KindBindingElement; binding = binding.Parent.Parent {
						element := binding.AsBindingElement()
						if element.DotDotDotToken != nil && !nested {
							return
						}
						visit(element.Initializer, nested)
						nested = true
					}
					if root := EnclosingVariableDeclarationOfBindingElement(declaration); root != nil {
						visit(root.AsVariableDeclaration().Initializer, true)
					}
				}
			}
			return
		}
		switch node.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			visit(AccessExpressionObject(node), true)
		case ast.KindConditionalExpression:
			visit(node.AsConditionalExpression().WhenTrue, nested)
			visit(node.AsConditionalExpression().WhenFalse, nested)
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindCommaToken, ast.KindEqualsToken:
				visit(binary.Right, nested)
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				visit(binary.Left, nested)
				visit(binary.Right, nested)
			}
		case ast.KindObjectLiteralExpression:
			if nested {
				for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
					switch property.Kind {
					case ast.KindPropertyAssignment:
						visit(property.AsPropertyAssignment().Initializer, true)
					case ast.KindShorthandPropertyAssignment:
						visit(property.Name(), true)
					}
				}
			}
		case ast.KindArrayLiteralExpression:
			if nested {
				for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
					visit(element, true)
				}
			}
		}
	}
	visit(target, false)
}
