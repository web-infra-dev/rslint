// cspell:ignore unscopables

package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// EvalValueIfNoSideEffects folds a value only when evaluation has no side
// effects or potentially mutable member reads. Unlike control-flow evaluation,
// this also accepts unmodified let/var bindings, matching Unicorn's merge rules.
func (staticEvaluator *StaticStringEvaluator) EvalValueIfNoSideEffects(node *ast.Node) (any, bool) {
	if staticEvaluator == nil || node == nil {
		return nil, false
	}
	safety := staticControlFlowSafety{evaluator: staticEvaluator, visiting: make(map[*ast.Symbol]bool)}
	if !safety.safeValue(node, false) {
		return nil, false
	}
	return staticEvaluator.EvalValue(node)
}

// EvalControlFlowValue mirrors unicorn's getStaticValueForControlFlow: reject
// side effects and unsupported member reads throughout the expression, but
// reject mutable bindings only on paths that can actually be evaluated.
// Keep this policy outside ordinary evaluation; changing evaluator state could
// also change the shared write-reference scan and affect later evaluations.
func (staticEvaluator *StaticStringEvaluator) EvalControlFlowValue(node *ast.Node) (any, bool) {
	if staticEvaluator == nil || node == nil {
		return nil, false
	}
	safety := staticControlFlowSafety{evaluator: staticEvaluator, visiting: make(map[*ast.Symbol]bool)}
	if !safety.safeValue(node, false) || safety.hasMutableBinding(node) {
		return nil, false
	}
	return staticEvaluator.EvalValue(node)
}

// EvalControlFlowArrayValue shares the value evaluator's safety policy while
// keeping its private aggregate representation out of rule implementations.
func (staticEvaluator *StaticStringEvaluator) EvalControlFlowArrayValue(node *ast.Node) (isArray bool, known bool) {
	value, ok := staticEvaluator.EvalControlFlowValue(node)
	if !ok {
		return false, false
	}
	_, isArray = value.(*staticArrayValue)
	return isArray, true
}

// EvalControlFlowTruthiness selects a branch without requiring its value to fold.
func (staticEvaluator *StaticStringEvaluator) EvalControlFlowTruthiness(node *ast.Node) (truthy bool, known bool) {
	value, ok := staticEvaluator.EvalControlFlowValue(node)
	if !ok {
		return false, false
	}
	return staticValueTruthy(value)
}

// EvalControlFlowNullish selects a branch of `??` or `?.` without requiring the
// value to fold: nullish reports whether the node is null or undefined.
func (staticEvaluator *StaticStringEvaluator) EvalControlFlowNullish(node *ast.Node) (nullish bool, known bool) {
	value, ok := staticEvaluator.EvalControlFlowValue(node)
	if !ok {
		return false, false
	}
	return staticValueNullish(value), true
}

type staticControlFlowSafety struct {
	evaluator *StaticStringEvaluator
	visiting  map[*ast.Symbol]bool
}

func (safety *staticControlFlowSafety) safeValue(node *ast.Node, considerGetters bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	// Only a complete, non-optional Object.freeze/seal/preventExtensions call
	// with one statically known argument is a side-effect exception. Calls
	// nested in a larger expression do not inherit that exception.
	if argument, ok := safety.passThroughArgument(node); ok {
		return safety.safeValue(argument, false) && safety.evaluator.evalValue(argument).ok
	}
	return !safety.evaluator.HasSideEffect(node, considerGetters) && safety.safeReferencesAndMembers(node)
}

func (safety *staticControlFlowSafety) passThroughArgument(node *ast.Node) (*ast.Node, bool) {
	if node.Kind != ast.KindCallExpression || ast.IsOptionalChainRoot(node) {
		return nil, false
	}
	call := node.AsCallExpression()
	callee := SkipAssertionsAndParens(call.Expression)
	if call.QuestionDotToken != nil || len(node.Arguments()) != 1 || callee == nil ||
		callee.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChainRoot(callee) ||
		callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		return nil, false
	}
	return safety.evaluator.objectPassThroughArgument(node)
}

// Constant initializers get upstream's stricter getter check. A direct literal
// member may be safe while an alias initialized by a member read remains unknown.
func (safety *staticControlFlowSafety) safeReferencesAndMembers(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) {
		return true
	}
	switch node.Kind {
	case ast.KindIdentifier:
		if IsNonReferenceIdentifier(node) {
			return true
		}
		initializer, symbol, ok := safety.evaluator.resolveIdentifierInitializer(node)
		if !ok || !ast.IsVarConst(symbol.Declarations[0].Parent) {
			return true
		}
		if safety.visiting[symbol] {
			return false
		}
		safety.visiting[symbol] = true
		defer delete(safety.visiting, symbol)
		return safety.safeValue(initializer, true)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if !safety.safeMember(node) {
			return false
		}
	case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression:
		return true
	}
	return !node.ForEachChild(func(child *ast.Node) bool { return !safety.safeReferencesAndMembers(child) })
}

func (safety *staticControlFlowSafety) propertyName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindPropertyAccessExpression {
		return AccessExpressionStaticName(node)
	}
	argument := node.AsElementAccessExpression().ArgumentExpression
	value := safety.evaluator.evalValue(argument)
	if !value.ok {
		return "", false
	}
	switch staticValueKindOf(value.value) {
	case staticKindString, staticKindNumber:
		return staticValueToString(value.value)
	default:
		return "", false
	}
}

func (safety *staticControlFlowSafety) safeMember(node *ast.Node) bool {
	key, ok := safety.propertyName(node)
	if !ok {
		return false
	}
	object := SkipAssertionsAndParens(AccessExpressionObject(node))
	if object == nil {
		return false
	}
	if safeControlFlowGlobalMember(node, object, key) {
		return true
	}
	switch object.Kind {
	case ast.KindObjectLiteralExpression:
		found := false
		for _, property := range object.AsObjectLiteralExpression().Properties.Nodes {
			if (property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment) ||
				property.Name().Kind == ast.KindComputedPropertyName {
				return false
			}
			name, ok := safety.evaluator.evalPropertyKey(property.Name())
			if !ok || name == "__proto__" {
				return false
			}
			found = found || ecmascript.CompareStrings(name, key) == 0
		}
		return found
	case ast.KindArrayLiteralExpression:
		elements := object.AsArrayLiteralExpression().Elements.Nodes
		for _, element := range elements {
			if element.Kind == ast.KindSpreadElement {
				return false
			}
		}
		if key == "length" {
			return true
		}
		index, ok := staticArrayIndex(key)
		return ok && index < len(elements) && elements[index].Kind != ast.KindOmittedExpression
	case ast.KindIdentifier:
		initializer, symbol, ok := safety.evaluator.resolveIdentifierInitializer(object)
		if !ok || !ast.IsVarConst(symbol.Declarations[0].Parent) {
			return false
		}
		object = initializer
	}
	// Only literal strings (or const bindings initialized by one) are safe;
	// aggregate aliases can be mutated or have their properties redefined.
	if object.Kind != ast.KindStringLiteral {
		return false
	}
	if key == "length" {
		return true
	}
	index, ok := staticArrayIndex(key)
	return ok && index < ecmascript.StringCodeUnitCount(object.Text())
}

func safeControlFlowGlobalMember(node, object *ast.Node, key string) bool {
	if !ast.IsIdentifier(object) || ast.IsOptionalChainRoot(node) || IsShadowed(object, object.Text()) {
		return false
	}
	if node.Kind == ast.KindElementAccessExpression &&
		node.AsElementAccessExpression().ArgumentExpression.Kind != ast.KindStringLiteral {
		return false
	}
	if _, ok := staticGlobalNumber(object, key); ok {
		return true
	}
	if object.Text() == "String" {
		return key == "raw"
	}
	if object.Text() == "Symbol" {
		switch key {
		case "asyncIterator", "hasInstance", "isConcatSpreadable", "iterator", "match", "matchAll",
			"replace", "search", "species", "split", "toPrimitive", "toStringTag", "unscopables":
			return true
		}
	}
	return false
}

// HasSideEffect checks explicit effects, including decorator and tag calls,
// optionally treating member reads as getters. Known String.raw tags remain
// safe unless their children have effects. Implicit coercions and property
// getters are otherwise ignored, as upstream does.
func (staticEvaluator *StaticStringEvaluator) HasSideEffect(node *ast.Node, considerGetters bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	// Class extends clauses and TS instantiation expressions share a type-node
	// kind, but evaluate their expression. Implements/interface heritage does not.
	if node.Kind == ast.KindExpressionWithTypeArguments {
		return !ast.IsPartOfTypeNode(node) && staticEvaluator.HasSideEffect(node.Expression(), considerGetters)
	}
	if ast.IsTypeNode(node) {
		return false
	}
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindAwaitExpression,
		ast.KindYieldExpression, ast.KindDeleteExpression, ast.KindPostfixUnaryExpression,
		ast.KindDecorator:
		return true
	case ast.KindSpreadElement:
		if staticEvaluator != nil && staticEvaluator.hasModifiedIterator(nil) {
			value := staticEvaluator.evalValue(node.Expression())
			if !value.ok || staticEvaluator.hasModifiedIterator(value.value) {
				return true
			}
		}
	case ast.KindTaggedTemplateExpression:
		if staticEvaluator == nil || !staticEvaluator.isStringRawTag(node.AsTaggedTemplateExpression().Tag) {
			return true
		}
	case ast.KindBinaryExpression:
		if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			return true
		}
	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			return true
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if considerGetters {
			return true
		}
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return false
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		for _, decorator := range node.Decorators() {
			if staticEvaluator.HasSideEffect(decorator, considerGetters) {
				return true
			}
		}
		// Legacy parameter decorators run during class definition. Parameter
		// initializers and the function body still wait until invocation.
		for _, parameter := range node.Parameters() {
			if len(parameter.Decorators()) > 0 {
				return true
			}
		}
		return staticEvaluator.HasSideEffect(node.Name(), considerGetters)
	}
	return node.ForEachChild(func(child *ast.Node) bool { return staticEvaluator.HasSideEffect(child, considerGetters) })
}

func (safety *staticControlFlowSafety) hasMutableBinding(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		if IsNonReferenceIdentifier(node) {
			return false
		}
		symbol := safety.evaluator.referenceSymbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil ||
			declaration.Parent.Kind != ast.KindVariableDeclarationList {
			return false
		}
		if !ast.IsVarConst(declaration.Parent) {
			return true
		}
		if declaration.End() > node.Pos() {
			return true
		}
		if safety.visiting[symbol] {
			return false
		}
		safety.visiting[symbol] = true
		defer delete(safety.visiting, symbol)
		return safety.hasMutableBinding(declaration.AsVariableDeclaration().Initializer)
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if safety.hasMutableBinding(conditional.Condition) {
			return true
		}
		condition := safety.evaluator.evalValue(conditional.Condition)
		if condition.ok {
			if truthy, known := staticValueTruthy(condition.value); known {
				if truthy {
					return safety.hasMutableBinding(conditional.WhenTrue)
				}
				return safety.hasMutableBinding(conditional.WhenFalse)
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken || operator == ast.KindQuestionQuestionToken {
			if safety.hasMutableBinding(binary.Left) {
				return true
			}
			left := safety.evaluator.evalValue(binary.Left)
			if left.ok {
				truthy, known := staticValueTruthy(left.value)
				if known && ((operator == ast.KindAmpersandAmpersandToken && !truthy) ||
					(operator == ast.KindBarBarToken && truthy) ||
					(operator == ast.KindQuestionQuestionToken && !staticValueNullish(left.value))) {
					return false
				}
			}
			return safety.hasMutableBinding(binary.Right)
		}
	case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression:
		return false
	}
	return node.ForEachChild(safety.hasMutableBinding)
}
