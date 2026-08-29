package utils

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/evaluator"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// StaticStringEvaluator folds expressions to string constants. It wraps tsgo's
// evaluator and adds stable local variable resolution through the TypeChecker.
// Create one evaluator per linted file; it keeps write-reference and recursion
// state for that file.
type StaticStringEvaluator struct {
	typeChecker            *checker.Checker
	sourceFile             *ast.SourceFile
	referenceResolver      StaticReferenceResolver
	evaluator              evaluator.Evaluator
	resolveIdentifiers     bool
	resolving              map[*ast.Symbol]bool
	referenceFlagsComputed bool
	referenceFlags         map[*ast.Symbol]staticReferenceFlags
}

func NewStaticStringEvaluator(typeChecker *checker.Checker) *StaticStringEvaluator {
	return NewStaticStringEvaluatorWithSourceFile(typeChecker, nil)
}

func NewStaticStringEvaluatorWithSourceFile(typeChecker *checker.Checker, sourceFile *ast.SourceFile) *StaticStringEvaluator {
	return NewStaticStringEvaluatorWithReferenceResolver(typeChecker, sourceFile, nil)
}

// StaticReferenceResolver is the subset of rule.RefStore used by the evaluator.
// Supplying it reuses the linter's binder-based symbol resolution while the
// evaluator performs its single mutation scan.
type StaticReferenceResolver interface {
	Resolve(node *ast.Node) *ast.Symbol
}

func NewStaticStringEvaluatorWithReferenceResolver(
	typeChecker *checker.Checker,
	sourceFile *ast.SourceFile,
	referenceResolver StaticReferenceResolver,
) *StaticStringEvaluator {
	staticEvaluator := &StaticStringEvaluator{
		typeChecker:        typeChecker,
		sourceFile:         sourceFile,
		referenceResolver:  referenceResolver,
		resolveIdentifiers: true,
		resolving:          map[*ast.Symbol]bool{},
	}
	staticEvaluator.evaluator = evaluator.NewEvaluator(staticEvaluator.evaluateEntity, ast.OEKAssertions)
	return staticEvaluator
}

// NewStaticStringEvaluatorWithoutScope mirrors eslint-utils getStaticValue
// when no initial scope is supplied: literal expressions can still be folded,
// but identifiers (including built-in globals) cannot be resolved.
func NewStaticStringEvaluatorWithoutScope() *StaticStringEvaluator {
	staticEvaluator := NewStaticStringEvaluator(nil)
	staticEvaluator.resolveIdentifiers = false
	return staticEvaluator
}

type staticNullValue struct{}
type staticUndefinedValue struct{}

// staticNumberValue is a number this evaluator computed itself. tsgo hands
// numbers back as jsnum.Number, a type internal/utils cannot import; the IsNaN
// method both types carry is what tells a folded number from a folded bigint.
type staticNumberValue float64

func (value staticNumberValue) IsNaN() bool {
	return math.IsNaN(float64(value))
}

// staticStringNode keeps literal strings backed by their existing AST node so
// nested aggregate evaluation doesn't allocate an interface box per literal.
type staticStringNode ast.Node

// staticObjectValue and staticArrayValue hold folded object and array literals
// so member access and JavaScript's default string coercion can be modeled.
type staticObjectValue struct {
	property        staticObjectProperty
	extraProperties []staticObjectProperty
	propertyCount   int
	prototype       any
	prototypeSet    bool
}

type staticObjectProperty struct {
	name  string
	value any
}

type staticArrayValue struct {
	length   int
	inline   [2]any
	overflow []any
}

type staticReferenceFlags uint8

const (
	staticReferenceWrite staticReferenceFlags = 1 << iota
	staticReferencePropertyMutation
)

// objectPassThroughMethods are the `Object` methods that return their argument
// unchanged, so folding can see through them.
var objectPassThroughMethods = map[string]bool{
	"freeze":            true,
	"preventExtensions": true,
	"seal":              true,
}

// Static evaluation must not turn a compact expression into an unbounded
// allocation. The limit is deliberately much larger than a useful diagnostic
// message while still bounding nested Array#join coercion.
const maxStaticStringLength = 1 << 20

type staticEvalResult struct {
	value any
	ok    bool
}

// StaticEvalValueKind classifies a static evaluation result for callers that
// only need to distinguish strings from known non-strings and unknown values.
type StaticEvalValueKind uint8

const (
	StaticEvalUnknown StaticEvalValueKind = iota
	StaticEvalString
	StaticEvalNonString
)

// Eval returns the static string value of node, if it can be determined. It
// covers the string-producing subset of ESLint's getStaticValue that rules use
// for computed property names, key arguments, constructor arguments, and other
// string-only rule inputs. It includes nested conditionals, logical
// short-circuiting, String(), String.raw, and local variables with stable
// initializers.
func (staticEvaluator *StaticStringEvaluator) Eval(node *ast.Node) (string, bool) {
	if staticEvaluator == nil || node == nil {
		return "", false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return "", false
	}
	return staticValueAsString(result.value)
}

// EvalToString returns JavaScript's String conversion of a statically known
// value. This matches eslint-utils getStringIfConstant after getStaticValue has
// evaluated an expression.
func (staticEvaluator *StaticStringEvaluator) EvalToString(node *ast.Node) (string, bool) {
	if staticEvaluator == nil || node == nil {
		return "", false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return "", false
	}
	return staticValueToString(result.value)
}

// EvalValue returns the static value of node if it can be determined, regardless
// of its type. It allows rules to check if a value is statically known to be a
// non-string (like a boolean or number).
func (staticEvaluator *StaticStringEvaluator) EvalValue(node *ast.Node) (any, bool) {
	if staticEvaluator == nil || node == nil {
		return nil, false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return nil, false
	}
	if value, ok := staticValueAsString(result.value); ok {
		return value, true
	}
	return result.value, true
}

// EvalArrayValue classifies a statically known value by whether it is an
// array. The second result is false when the expression cannot be folded. It
// keeps the evaluator's private aggregate representation encapsulated while
// allowing callers that mirror eslint-utils' getStaticValue + Array.isArray
// pattern to distinguish arrays from other known values.
func (staticEvaluator *StaticStringEvaluator) EvalArrayValue(node *ast.Node) (isArray bool, known bool) {
	if staticEvaluator == nil || node == nil {
		return false, false
	}
	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return false, false
	}
	_, isArray = result.value.(*staticArrayValue)
	return isArray, true
}

// EvalStringValue returns a string without boxing it into an interface, or
// classifies the result as a known non-string or an unknown value.
func (staticEvaluator *StaticStringEvaluator) EvalStringValue(node *ast.Node) (string, StaticEvalValueKind) {
	if staticEvaluator == nil || node == nil {
		return "", StaticEvalUnknown
	}

	node = SkipAssertionsAndParens(node)
	if node == nil {
		return "", StaticEvalUnknown
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, StaticEvalString
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, StaticEvalString
	case ast.KindCallExpression:
		if value, matched, ok := staticEvaluator.evalArrayJoin(node); matched {
			if ok {
				return value, StaticEvalString
			}
			return "", StaticEvalUnknown
		}
	}

	result := staticEvaluator.evalValue(node)
	if !result.ok {
		return "", StaticEvalUnknown
	}
	if value, ok := staticValueAsString(result.value); ok {
		return value, StaticEvalString
	}
	return "", StaticEvalNonString
}

func (staticEvaluator *StaticStringEvaluator) evalValue(node *ast.Node) staticEvalResult {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return staticEvalResult{}
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: (*staticStringNode)(node), ok: true}
	case ast.KindTrueKeyword:
		return staticEvalResult{value: true, ok: true}
	case ast.KindFalseKeyword:
		return staticEvalResult{value: false, ok: true}
	case ast.KindNullKeyword:
		return staticEvalResult{value: staticNullValue{}, ok: true}
	case ast.KindUndefinedKeyword:
		if !staticEvaluator.resolveIdentifiers {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindIdentifier:
		if !staticEvaluator.resolveIdentifiers {
			return staticEvalResult{}
		}
		identifier := node.AsIdentifier()
		if identifier != nil && identifier.Text == "undefined" && !IsShadowed(node, "undefined") {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvaluator.evalIdentifier(node)
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(node)
	case ast.KindBinaryExpression:
		return staticEvaluator.evalBinaryExpression(node)
	case ast.KindPrefixUnaryExpression:
		return staticEvaluator.evalPrefixUnaryExpression(node)
	case ast.KindConditionalExpression:
		return staticEvaluator.evalConditionalExpression(node)
	case ast.KindVoidExpression:
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case ast.KindObjectLiteralExpression:
		if result := staticEvaluator.evalObjectLiteral(node); result.ok {
			return result
		}
	case ast.KindArrayLiteralExpression:
		if result := staticEvaluator.evalArrayLiteral(node); result.ok {
			return result
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if result := staticEvaluator.evalMemberAccess(node); result.ok {
			return result
		}
	case ast.KindCallExpression:
		if result := staticEvaluator.evalBuiltinStaticCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalStringCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalArrayJoinCall(node); result.ok {
			return result
		}
		if result := staticEvaluator.evalObjectPassThroughCall(node); result.ok {
			return result
		}
	case ast.KindTaggedTemplateExpression:
		if result := staticEvaluator.evalStringRawTag(node); result.ok {
			return result
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

func (staticEvaluator *StaticStringEvaluator) evalIdentifier(node *ast.Node) staticEvalResult {
	if !staticEvaluator.resolveIdentifiers {
		return staticEvalResult{}
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || staticEvaluator.resolving[symbol] {
		return staticEvalResult{}
	}

	staticEvaluator.resolving[symbol] = true
	defer delete(staticEvaluator.resolving, symbol)

	if isAggregateLiteral(initializer) && staticEvaluator.hasPropertyMutation(symbol) {
		return staticEvalResult{}
	}
	result := staticEvaluator.evalValue(initializer)
	if result.ok && staticValueIsAggregate(result.value) &&
		!isAggregateLiteral(initializer) && staticEvaluator.hasPropertyMutation(symbol) {
		return staticEvalResult{}
	}
	return result
}

func isAggregateLiteral(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	return node != nil && (node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindArrayLiteralExpression)
}

// ResolveIdentifierInitializer resolves an identifier to a stable variable
// initializer: const bindings, or let/var bindings with no write references in
// the current source file. It returns false for destructuring, using bindings,
// ambiguous symbols, or files where let/var writes cannot be checked.
func (staticEvaluator *StaticStringEvaluator) ResolveIdentifierInitializer(node *ast.Node) (*ast.Node, bool) {
	if staticEvaluator == nil {
		return nil, false
	}
	initializer, _, ok := staticEvaluator.resolveIdentifierInitializer(node)
	return initializer, ok
}

func (staticEvaluator *StaticStringEvaluator) resolveIdentifierInitializer(node *ast.Node) (*ast.Node, *ast.Symbol, bool) {
	if staticEvaluator == nil || (staticEvaluator.typeChecker == nil && staticEvaluator.referenceResolver == nil) {
		return nil, nil, false
	}

	expr := SkipAssertionsAndParens(node)
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return nil, nil, false
	}

	symbol := staticEvaluator.referenceSymbol(expr)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil, false
	}

	declarationNode := symbol.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return nil, nil, false
	}

	declaration := declarationNode.AsVariableDeclaration()
	if declaration == nil || declaration.Initializer == nil || !isIdentifierWithText(declaration.Name(), expr.AsIdentifier().Text) {
		return nil, nil, false
	}

	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return nil, nil, false
	}
	if ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList) {
		return nil, nil, false
	}
	if !ast.IsVarConst(declarationList) && staticEvaluator.hasWrites(symbol) {
		return nil, nil, false
	}
	return declaration.Initializer, symbol, true
}

func (staticEvaluator *StaticStringEvaluator) evalTemplateExpression(node *ast.Node) staticEvalResult {
	template := node.AsTemplateExpression()
	if template == nil {
		return staticEvalResult{}
	}

	var builder strings.Builder
	if template.Head != nil {
		builder.WriteString(template.Head.Text())
	}
	if template.TemplateSpans != nil {
		for _, spanNode := range template.TemplateSpans.Nodes {
			span := spanNode.AsTemplateSpan()
			if span == nil {
				return staticEvalResult{}
			}
			result := staticEvaluator.evalValue(span.Expression)
			if !result.ok {
				return staticEvalResult{}
			}
			value, ok := staticValueToString(result.value)
			if !ok {
				return staticEvalResult{}
			}
			builder.WriteString(value)
			if span.Literal != nil {
				builder.WriteString(span.Literal.Text())
			}
		}
	}
	return staticEvalResult{value: builder.String(), ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalBinaryExpression(node *ast.Node) staticEvalResult {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return staticEvalResult{}
	}

	switch binary.OperatorToken.Kind {
	case ast.KindCommaToken:
		// A sequence expression evaluates to its final operand. This mirrors
		// eslint-utils getStaticValue for the expression shapes that consume the
		// result rather than its side effects.
		return staticEvaluator.evalValue(binary.Right)
	case ast.KindEqualsToken:
		result := staticEvaluator.evalValue(binary.Right)
		if result.ok && staticValueIsAggregate(result.value) {
			// The assigned aggregate can be mutated by another operand or call
			// argument before the containing expression consumes it.
			return staticEvalResult{}
		}
		return result
	case ast.KindBarBarToken:
		left := staticEvaluator.evalValue(binary.Left)
		if !left.ok {
			return staticEvalResult{}
		}
		truthy, ok := staticValueTruthy(left.value)
		if !ok {
			return staticEvalResult{}
		}
		if truthy {
			return left
		}
		return staticEvaluator.evalValue(binary.Right)
	case ast.KindAmpersandAmpersandToken:
		left := staticEvaluator.evalValue(binary.Left)
		if !left.ok {
			return staticEvalResult{}
		}
		truthy, ok := staticValueTruthy(left.value)
		if !ok {
			return staticEvalResult{}
		}
		if !truthy {
			return left
		}
		return staticEvaluator.evalValue(binary.Right)
	case ast.KindQuestionQuestionToken:
		left := staticEvaluator.evalValue(binary.Left)
		if !left.ok {
			return staticEvalResult{}
		}
		if staticValueNullish(left.value) {
			return staticEvaluator.evalValue(binary.Right)
		}
		return left
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		left := staticEvaluator.evalValue(binary.Left)
		right := staticEvaluator.evalValue(binary.Right)
		if !left.ok || !right.ok {
			return staticEvalResult{}
		}
		equal, ok := staticValuesStrictEqual(left.value, right.value)
		if !ok {
			return staticEvalResult{}
		}
		if binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			equal = !equal
		}
		return staticEvalResult{value: equal, ok: true}
	case ast.KindPlusToken:
		left := staticEvaluator.evalValue(binary.Left)
		right := staticEvaluator.evalValue(binary.Right)
		if !left.ok || !right.ok {
			return staticEvalResult{}
		}
		if staticValueIsString(left.value) {
			return staticEvaluator.concatStaticValues(left.value, right.value)
		}
		if staticValueIsString(right.value) {
			return staticEvaluator.concatStaticValues(left.value, right.value)
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

func (staticEvaluator *StaticStringEvaluator) concatStaticValues(left any, right any) staticEvalResult {
	leftString, ok := staticValueToString(left)
	if !ok {
		return staticEvalResult{}
	}
	rightString, ok := staticValueToString(right)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: leftString + rightString, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalPrefixUnaryExpression(node *ast.Node) staticEvalResult {
	prefix := node.AsPrefixUnaryExpression()
	if prefix == nil {
		return staticEvaluator.evalWithTsgo(node)
	}

	switch prefix.Operator {
	case ast.KindExclamationToken:
		operand := staticEvaluator.evalValue(prefix.Operand)
		if !operand.ok {
			return staticEvalResult{}
		}
		truthy, ok := staticValueTruthy(operand.value)
		if !ok {
			return staticEvalResult{}
		}
		return staticEvalResult{value: !truthy, ok: true}
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken:
		if result := staticEvaluator.evalNumericPrefix(prefix); result.ok {
			return result
		}
	}

	return staticEvaluator.evalWithTsgo(node)
}

// evalNumericPrefix folds `+`, `-` and `~` over any static operand. tsgo's
// evaluator folds these operators only when the operand is already a number,
// so string, boolean and nullish operands reach here unfolded.
func (staticEvaluator *StaticStringEvaluator) evalNumericPrefix(prefix *ast.PrefixUnaryExpression) staticEvalResult {
	operand := staticEvaluator.evalValue(prefix.Operand)
	if !operand.ok {
		return staticEvalResult{}
	}
	number, ok := staticValueToNumber(operand.value)
	if !ok {
		return staticEvalResult{}
	}

	switch prefix.Operator {
	case ast.KindPlusToken:
		return staticEvalResult{value: staticNumberValue(number), ok: true}
	case ast.KindMinusToken:
		return staticEvalResult{value: staticNumberValue(-number), ok: true}
	case ast.KindTildeToken:
		return staticEvalResult{value: staticNumberValue(^toInt32(number)), ok: true}
	}
	return staticEvalResult{}
}

func (staticEvaluator *StaticStringEvaluator) evalConditionalExpression(node *ast.Node) staticEvalResult {
	conditional := node.AsConditionalExpression()
	if conditional == nil {
		return staticEvalResult{}
	}

	condition := staticEvaluator.evalValue(conditional.Condition)
	if !condition.ok {
		return staticEvalResult{}
	}
	truthy, ok := staticValueTruthy(condition.value)
	if !ok {
		return staticEvalResult{}
	}
	if truthy {
		return staticEvaluator.evalValue(conditional.WhenTrue)
	}
	return staticEvaluator.evalValue(conditional.WhenFalse)
}

func (staticEvaluator *StaticStringEvaluator) evalObjectLiteral(node *ast.Node) staticEvalResult {
	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return staticEvalResult{}
	}

	value := &staticObjectValue{}
	for _, property := range object.Properties.Nodes {
		var valueNode *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			valueNode = property.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			valueNode = property.Name()
		default:
			return staticEvalResult{}
		}

		if isObjectLiteralPrototypeSetter(property) {
			propertyValue := staticEvaluator.evalValue(valueNode)
			if !propertyValue.ok {
				return staticEvalResult{}
			}
			switch propertyValue.value.(type) {
			case *staticObjectValue, *staticArrayValue:
				value.prototype = propertyValue.value
				value.prototypeSet = true
			case staticNullValue:
				value.prototype = nil
				value.prototypeSet = true
			default:
				// JavaScript ignores primitive prototype-setter values.
			}
			continue
		}

		key, ok := staticEvaluator.evalPropertyKey(property.Name())
		if !ok {
			return staticEvalResult{}
		}
		propertyValue := staticEvaluator.evalValue(valueNode)
		if !propertyValue.ok {
			return staticEvalResult{}
		}
		value.addProperty(key, propertyValue.value)
	}
	return staticEvalResult{value: value, ok: true}
}

func (value *staticObjectValue) addProperty(name string, propertyValue any) {
	property := staticObjectProperty{name: name, value: propertyValue}
	if value.propertyCount == 0 {
		value.property = property
	} else {
		value.extraProperties = append(value.extraProperties, property)
	}
	value.propertyCount++
}

func isObjectLiteralPrototypeSetter(property *ast.Node) bool {
	return property != nil && property.Kind == ast.KindPropertyAssignment && ast.IsProtoSetter(property.Name())
}

func (staticEvaluator *StaticStringEvaluator) evalObjectLiteralMember(node *ast.Node, key string) staticEvalResult {
	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return staticEvalResult{}
	}

	var ownValue staticEvalResult
	ownFound := false
	prototypeValue := staticEvalResult{value: staticUndefinedValue{}, ok: true}
	prototypeSet := false
	for _, property := range object.Properties.Nodes {
		var valueNode *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			valueNode = property.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			valueNode = property.Name()
		default:
			return staticEvalResult{}
		}

		if isObjectLiteralPrototypeSetter(property) {
			prototypeNode := SkipAssertionsAndParens(valueNode)
			if prototypeNode != nil && prototypeNode.Kind == ast.KindObjectLiteralExpression {
				prototypeValue = staticEvaluator.evalObjectLiteralMember(prototypeNode, key)
				if !prototypeValue.ok {
					return staticEvalResult{}
				}
				prototypeSet = true
				continue
			}

			value := staticEvaluator.evalValue(valueNode)
			if !value.ok {
				return staticEvalResult{}
			}
			switch value.value.(type) {
			case *staticObjectValue, *staticArrayValue:
				prototypeValue = staticMemberValue(value.value, key)
				if !prototypeValue.ok {
					return staticEvalResult{}
				}
				prototypeSet = true
			case staticNullValue:
				prototypeValue = staticEvalResult{value: staticUndefinedValue{}, ok: true}
				prototypeSet = true
			default:
				// JavaScript ignores primitive prototype-setter values.
			}
			continue
		}

		propertyKey, ok := staticEvaluator.evalPropertyKey(property.Name())
		if !ok {
			return staticEvalResult{}
		}
		value := staticEvaluator.evalValue(valueNode)
		if !value.ok {
			return staticEvalResult{}
		}
		if propertyKey == key {
			ownValue = value
			ownFound = true
		}
	}

	if ownFound {
		return ownValue
	}
	if prototypeSet {
		return prototypeValue
	}
	return staticEvalResult{value: staticUndefinedValue{}, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalArrayLiteral(node *ast.Node) staticEvalResult {
	array := node.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return staticEvalResult{}
	}

	arrayValue := &staticArrayValue{length: len(array.Elements.Nodes)}
	if arrayValue.length > len(arrayValue.inline) {
		arrayValue.overflow = make([]any, arrayValue.length-len(arrayValue.inline))
	}
	for i, element := range array.Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			arrayValue.set(i, staticUndefinedValue{})
			continue
		}
		if element.Kind == ast.KindSpreadElement {
			return staticEvalResult{}
		}
		elementValue := staticEvaluator.evalValue(element)
		if !elementValue.ok {
			return staticEvalResult{}
		}
		arrayValue.set(i, elementValue.value)
	}
	return staticEvalResult{value: arrayValue, ok: true}
}

func (value *staticArrayValue) set(index int, element any) {
	if index < len(value.inline) {
		value.inline[index] = element
		return
	}
	value.overflow[index-len(value.inline)] = element
}

func (value *staticArrayValue) element(index int) any {
	if index < len(value.inline) {
		return value.inline[index]
	}
	return value.overflow[index-len(value.inline)]
}

func (staticEvaluator *StaticStringEvaluator) evalPropertyKey(name *ast.Node) (string, bool) {
	if name == nil {
		return "", false
	}

	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text, true
	case ast.KindStringLiteral:
		return name.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return name.AsNoSubstitutionTemplateLiteral().Text, true
	case ast.KindNumericLiteral:
		// Not the source text: `{1e3: x}` has the key "1000", which is what an
		// index expression folds to.
		result := staticEvaluator.evalValue(name)
		if !result.ok {
			return "", false
		}
		return staticValueToString(result.value)
	case ast.KindComputedPropertyName:
		computed := name.AsComputedPropertyName()
		if computed == nil {
			return "", false
		}
		result := staticEvaluator.evalValue(computed.Expression)
		if !result.ok {
			return "", false
		}
		return staticValueToString(result.value)
	}
	return "", false
}

func (staticEvaluator *StaticStringEvaluator) evalMemberAccess(node *ast.Node) staticEvalResult {
	objectNode := AccessExpressionObject(node)
	if objectNode == nil {
		return staticEvalResult{}
	}
	key, ok := staticEvaluator.evalAccessExpressionKey(node)
	if !ok {
		return staticEvalResult{}
	}
	if literal := SkipAssertionsAndParens(objectNode); literal != nil && literal.Kind == ast.KindObjectLiteralExpression {
		return staticEvaluator.evalObjectLiteralMember(literal, key)
	}

	object := staticEvaluator.evalValue(objectNode)
	if !object.ok {
		return staticEvalResult{}
	}
	if key == "length" {
		if text, ok := staticValueAsString(object.value); ok {
			return staticEvalResult{value: staticNumberValue(len(utf16.Encode([]rune(text)))), ok: true}
		}
	}
	return staticMemberValue(object.value, key)
}

func (staticEvaluator *StaticStringEvaluator) evalAccessExpressionKey(node *ast.Node) (string, bool) {
	if key, ok := AccessExpressionStaticName(node); ok {
		return key, true
	}
	if node == nil || node.Kind != ast.KindElementAccessExpression {
		return "", false
	}
	argument := staticEvaluator.evalValue(node.AsElementAccessExpression().ArgumentExpression)
	if !argument.ok {
		return "", false
	}
	return staticValueToString(argument.value)
}

// staticMemberValue reads key off a folded object or array literal. Reading a
// key that isn't there yields `undefined`, as it would at runtime. A key an
// array inherits — `length`, `join` — stays unresolved, because folding it
// would need a numeric or callable value representation this evaluator doesn't
// carry.
func staticMemberValue(object any, key string) staticEvalResult {
	switch object := object.(type) {
	case *staticObjectValue:
		if value, ok := staticObjectOwnProperty(object, key); ok {
			return staticEvalResult{value: value, ok: true}
		}
		if object.prototypeSet && object.prototype != nil {
			return staticMemberValue(object.prototype, key)
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case *staticArrayValue:
		index, ok := staticArrayIndex(key)
		if !ok {
			if staticNumberShapedKey(key) {
				return staticEvalResult{value: staticUndefinedValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		if index >= object.length {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: object.element(index), ok: true}
	}
	return staticEvalResult{}
}

func staticObjectOwnProperty(object *staticObjectValue, key string) (any, bool) {
	for i := len(object.extraProperties) - 1; i >= 0; i-- {
		if object.extraProperties[i].name == key {
			return object.extraProperties[i].value, true
		}
	}
	if object.propertyCount > 0 && object.property.name == key {
		return object.property.value, true
	}
	return nil, false
}

// staticNumberShapedKey reports whether key opens the way a number's string
// form can. Such a key is neither an own property of an array literal nor one
// it inherits, so reading it yields `undefined`.
func staticNumberShapedKey(key string) bool {
	if key == "" {
		return false
	}
	switch key[0] {
	case '-', '+', '.':
		return true
	default:
		return key[0] >= '0' && key[0] <= '9'
	}
}

// staticArrayIndex parses key as a canonical array index string: decimal
// digits only, with no leading zero unless key is exactly "0". JavaScript
// treats any other shape (`"00"`, `"1e3"`, `"-1"`) as an ordinary property
// name, which array literals don't have.
func staticArrayIndex(key string) (int, bool) {
	if key == "0" {
		return 0, true
	}
	if key == "" || key[0] < '1' || key[0] > '9' {
		return 0, false
	}
	for i := 1; i < len(key); i++ {
		if key[i] < '0' || key[i] > '9' {
			return 0, false
		}
	}
	index, err := strconv.Atoi(key)
	if err != nil {
		return 0, false
	}
	return index, true
}

// evalObjectPassThroughCall folds `Object.freeze(x)` and its siblings to the
// value of x.
func (staticEvaluator *StaticStringEvaluator) evalObjectPassThroughCall(node *ast.Node) staticEvalResult {
	if !staticEvaluator.resolveIdentifiers {
		return staticEvalResult{}
	}
	argument, ok := staticEvaluator.objectPassThroughArgument(node)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvaluator.evalValue(argument)
}

func (staticEvaluator *StaticStringEvaluator) objectPassThroughArgument(node *ast.Node) (*ast.Node, bool) {
	call := node.AsCallExpression()
	if call == nil {
		return nil, false
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return nil, false
	}
	methodName, ok := staticEvaluator.evalAccessExpressionKey(callee)
	if !ok || !objectPassThroughMethods[methodName] {
		return nil, false
	}

	object := SkipAssertionsAndParens(AccessExpressionObject(callee))
	if !isIdentifierWithText(object, "Object") || IsShadowed(object, "Object") {
		return nil, false
	}

	args := node.Arguments()
	if len(args) == 0 || ast.IsSpreadElement(args[0]) {
		return nil, false
	}
	return args[0], true
}

func (staticEvaluator *StaticStringEvaluator) evalArrayJoinCall(node *ast.Node) staticEvalResult {
	value, _, ok := staticEvaluator.evalArrayJoin(node)
	return staticEvalResult{value: value, ok: ok}
}

func (staticEvaluator *StaticStringEvaluator) evalArrayJoin(node *ast.Node) (value string, matched bool, ok bool) {
	call := node.AsCallExpression()
	if call == nil {
		return "", false, false
	}

	callee := SkipAssertionsAndParens(call.Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return "", false, false
	}
	methodName, nameOK := staticEvaluator.evalAccessExpressionKey(callee)
	if !nameOK || methodName != "join" {
		return "", false, false
	}

	receiverNode := AccessExpressionObject(callee)
	if literal := SkipAssertionsAndParens(receiverNode); literal != nil && literal.Kind == ast.KindArrayLiteralExpression {
		arrayLiteral := literal.AsArrayLiteralExpression()
		if arrayLiteral != nil && arrayLiteral.Elements != nil &&
			len(arrayLiteral.Elements.Nodes) <= len((staticArrayValue{}).inline) {
			array, arrayOK := staticEvaluator.evalInlineArrayLiteral(literal)
			if !arrayOK {
				return "", true, false
			}
			separator, separatorOK := staticEvaluator.evalArrayJoinSeparator(node)
			if !separatorOK {
				return "", true, false
			}
			value, ok := staticArrayJoin(&array, separator)
			return value, true, ok
		}
	}

	receiver := staticEvaluator.evalValue(receiverNode)
	if !receiver.ok {
		return "", true, false
	}
	array, arrayOK := receiver.value.(*staticArrayValue)
	if !arrayOK {
		return "", true, false
	}

	separator, separatorOK := staticEvaluator.evalArrayJoinSeparator(node)
	if !separatorOK {
		return "", true, false
	}
	value, ok = staticArrayJoin(array, separator)
	return value, true, ok
}

func (staticEvaluator *StaticStringEvaluator) evalInlineArrayLiteral(node *ast.Node) (staticArrayValue, bool) {
	literal := node.AsArrayLiteralExpression()
	if literal == nil || literal.Elements == nil || len(literal.Elements.Nodes) > len((staticArrayValue{}).inline) {
		return staticArrayValue{}, false
	}

	value := staticArrayValue{length: len(literal.Elements.Nodes)}
	for i, element := range literal.Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			value.set(i, staticUndefinedValue{})
			continue
		}
		if element.Kind == ast.KindSpreadElement {
			return staticArrayValue{}, false
		}
		elementValue := staticEvaluator.evalValue(element)
		if !elementValue.ok {
			return staticArrayValue{}, false
		}
		value.set(i, elementValue.value)
	}
	return value, true
}

func (staticEvaluator *StaticStringEvaluator) evalArrayJoinSeparator(node *ast.Node) (string, bool) {
	separator := ","
	for i, argumentNode := range node.Arguments() {
		if ast.IsSpreadElement(argumentNode) {
			return "", false
		}
		argument := staticEvaluator.evalValue(argumentNode)
		if !argument.ok {
			return "", false
		}
		if i == 0 && !staticValueUndefined(argument.value) {
			var ok bool
			separator, ok = staticValueToString(argument.value)
			if !ok {
				return "", false
			}
		}
	}
	return separator, true
}

func (staticEvaluator *StaticStringEvaluator) evalStringCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil {
		return staticEvalResult{}
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if !staticEvaluator.isBuiltinStringValue(callee, map[*ast.Symbol]bool{}) {
		return staticEvalResult{}
	}

	args := node.Arguments()
	if len(args) == 0 {
		return staticEvalResult{value: "", ok: true}
	}
	if ast.IsSpreadElement(args[0]) {
		return staticEvalResult{}
	}
	arg := staticEvaluator.evalValue(args[0])
	if !arg.ok {
		return staticEvalResult{}
	}
	value, ok := staticValueToString(arg.value)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: value, ok: true}
}

// evalBuiltinStaticCall folds the narrow built-in call subset that
// eslint-utils' getStaticValue permits and that string-consuming native rules
// need. It intentionally recognizes only the real String / Array constructors
// (or stable aliases) and only fully static arguments, never arbitrary methods
// named like a built-in.
func (staticEvaluator *StaticStringEvaluator) evalBuiltinStaticCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil {
		return staticEvalResult{}
	}

	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return staticEvalResult{}
	}
	method, ok := staticEvaluator.evalAccessExpressionKey(callee)
	if !ok {
		return staticEvalResult{}
	}
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}

	receiverNode := AccessExpressionObject(callee)
	if staticEvaluator.isBuiltinStringValue(receiverNode, map[*ast.Symbol]bool{}) {
		switch method {
		case "fromCharCode":
			return staticStringFromCharCode(arguments)
		}
	}
	if staticEvaluator.isBuiltinArrayValue(receiverNode, map[*ast.Symbol]bool{}) && method == "of" {
		return staticEvalResult{value: staticArrayFromValues(arguments), ok: true}
	}

	receiver := staticEvaluator.evalValue(receiverNode)
	text, ok := staticValueAsString(receiver.value)
	if !receiver.ok || !ok {
		return staticEvalResult{}
	}

	switch method {
	case "indexOf":
		return staticStringIndexOf(text, arguments)
	case "toUpperCase":
		return staticEvalResult{value: ecmascript.StringToUpperCase(text), ok: true}
	case "toLowerCase":
		return staticEvalResult{value: ecmascript.StringToLowerCase(text), ok: true}
	case "slice":
		return staticStringSlice(text, arguments)
	case "substring":
		return staticStringSubstring(text, arguments)
	case "substr":
		return staticStringSubstr(text, arguments)
	case "charAt":
		return staticStringCharAt(text, arguments)
	case "concat":
		return staticStringConcat(text, arguments)
	}

	return staticEvalResult{}
}

func staticStringIndexOf(text string, arguments []any) staticEvalResult {
	if len(arguments) == 0 {
		return staticEvalResult{}
	}
	needle, ok := staticValueToString(arguments[0])
	if !ok {
		return staticEvalResult{}
	}

	start := 0
	if len(arguments) > 1 {
		number, ok := staticValueToNumber(arguments[1])
		if !ok {
			return staticEvalResult{}
		}
		if !math.IsNaN(number) {
			start = int(math.Trunc(number))
		}
	}

	textUnits := utf16.Encode([]rune(text))
	needleUnits := utf16.Encode([]rune(needle))
	start = min(max(start, 0), len(textUnits))
	for i := start; i+len(needleUnits) <= len(textUnits); i++ {
		if slices.Equal(textUnits[i:i+len(needleUnits)], needleUnits) {
			return staticEvalResult{value: staticNumberValue(i), ok: true}
		}
	}
	return staticEvalResult{value: staticNumberValue(-1), ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evalCallArguments(node *ast.Node) ([]any, bool) {
	values := make([]any, 0, len(node.Arguments()))
	for _, argumentNode := range node.Arguments() {
		if ast.IsSpreadElement(argumentNode) {
			spread := staticEvaluator.evalValue(argumentNode.AsSpreadElement().Expression)
			array, ok := spread.value.(*staticArrayValue)
			if !spread.ok || !ok {
				return nil, false
			}
			for index := range array.length {
				values = append(values, array.element(index))
			}
			continue
		}
		argument := staticEvaluator.evalValue(argumentNode)
		if !argument.ok {
			return nil, false
		}
		values = append(values, argument.value)
	}
	return values, true
}

func staticArrayFromValues(values []any) *staticArrayValue {
	array := &staticArrayValue{length: len(values)}
	if array.length > len(array.inline) {
		array.overflow = make([]any, array.length-len(array.inline))
	}
	for index, value := range values {
		array.set(index, value)
	}
	return array
}

func staticStringFromCharCode(arguments []any) staticEvalResult {
	units := make([]uint16, 0, len(arguments))
	for _, argument := range arguments {
		number, ok := staticValueToNumber(argument)
		if !ok {
			return staticEvalResult{}
		}
		units = append(units, uint16(toUint32(number)))
	}
	return staticEvalResult{value: string(utf16.Decode(units)), ok: true}
}

func staticStringSlice(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	end, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := utf16.Encode([]rune(text))
	length := len(units)
	from := normalizeSliceIndex(start, length)
	to := normalizeSliceIndex(end, length)
	if to < from {
		to = from
	}
	return staticEvalResult{value: string(utf16.Decode(units[from:to])), ok: true}
}

func staticStringSubstring(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	end, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := utf16.Encode([]rune(text))
	length := len(units)
	from := clampSubstringIndex(start, length)
	to := clampSubstringIndex(end, length)
	if from > to {
		from, to = to, from
	}
	return staticEvalResult{value: string(utf16.Decode(units[from:to])), ok: true}
}

// staticStringSubstr implements the Annex B String#substr, whose second
// argument is a length rather than an end index.
func staticStringSubstr(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	count, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := utf16.Encode([]rune(text))
	length := len(units)
	from := normalizeSliceIndex(start, length)
	if count <= 0 {
		return staticEvalResult{value: "", ok: true}
	}
	to := length
	if count < length-from {
		to = from + count
	}
	return staticEvalResult{value: string(utf16.Decode(units[from:to])), ok: true}
}

func staticStringCharAt(text string, arguments []any) staticEvalResult {
	index, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	units := utf16.Encode([]rune(text))
	if index < 0 || index >= len(units) {
		return staticEvalResult{value: "", ok: true}
	}
	return staticEvalResult{value: string(utf16.Decode(units[index : index+1])), ok: true}
}

func staticStringConcat(text string, arguments []any) staticEvalResult {
	var builder strings.Builder
	builder.WriteString(text)
	for _, argument := range arguments {
		part, ok := staticValueToString(argument)
		if !ok {
			return staticEvalResult{}
		}
		if len(part) > maxStaticStringLength-builder.Len() {
			return staticEvalResult{}
		}
		builder.WriteString(part)
	}
	return staticEvalResult{value: builder.String(), ok: true}
}

func staticArgumentInteger(arguments []any, index int, defaultValue int) (int, bool) {
	if index >= len(arguments) || staticValueUndefined(arguments[index]) {
		return defaultValue, true
	}
	number, ok := staticValueToNumber(arguments[index])
	if !ok {
		return 0, false
	}
	if math.IsNaN(number) || number == 0 {
		return 0, true
	}
	if math.IsInf(number, 1) {
		return math.MaxInt, true
	}
	if math.IsInf(number, -1) {
		return math.MinInt, true
	}
	if number >= float64(math.MaxInt) {
		return math.MaxInt, true
	}
	if number <= float64(math.MinInt) {
		return math.MinInt, true
	}
	return int(math.Trunc(number)), true
}

func normalizeSliceIndex(index, length int) int {
	if index < 0 {
		return max(length+index, 0)
	}
	return min(index, length)
}

func clampSubstringIndex(index, length int) int {
	return min(max(index, 0), length)
}

func toUint32(number float64) uint32 {
	if math.IsNaN(number) || math.IsInf(number, 0) || number == 0 {
		return 0
	}
	remainder := math.Mod(math.Trunc(number), 1<<32)
	if remainder < 0 {
		remainder += 1 << 32
	}
	return uint32(remainder)
}

func toInt32(number float64) int32 {
	return int32(toUint32(number))
}

func staticValueToNumber(value any) (float64, bool) {
	switch value := value.(type) {
	case staticNumberValue:
		return float64(value), true
	case bool:
		if value {
			return 1, true
		}
		return 0, true
	case staticNullValue:
		return 0, true
	case staticUndefinedValue:
		return math.NaN(), true
	case string:
		return parseStaticNumber(value)
	case *staticStringNode:
		text, _ := staticValueAsString(value)
		return parseStaticNumber(text)
	default:
		text, ok := staticValueToString(value)
		if !ok {
			return 0, false
		}
		return parseStaticNumber(text)
	}
}

func parseStaticNumber(text string) (float64, bool) {
	value, ok := ecmascript.StringToNumber(text)
	if !ok {
		// StringToNumber returns NaN for an invalid numeric string. NaN is still
		// a statically known value and downstream coercions must be allowed to
		// observe it (for example, String.fromCharCode(NaN) produces a NUL).
		return math.NaN(), true
	}
	return value, true
}

func (staticEvaluator *StaticStringEvaluator) evalStringRawTag(node *ast.Node) staticEvalResult {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Template == nil || !staticEvaluator.isStringRawTag(tagged.Tag) {
		return staticEvalResult{}
	}

	switch tagged.Template.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return staticEvalResult{value: tagged.Template.Text(), ok: true}
	case ast.KindTemplateExpression:
		return staticEvaluator.evalTemplateExpression(tagged.Template)
	default:
		return staticEvalResult{}
	}
}

func (staticEvaluator *StaticStringEvaluator) isStringRawTag(tag *ast.Node) bool {
	tag = ast.SkipOuterExpressions(tag, ast.OEKParentheses|ast.OEKAssertions)
	if tag == nil || !ast.IsAccessExpression(tag) {
		return false
	}
	propertyName, ok := staticEvaluator.evalAccessExpressionKey(tag)
	if !ok || propertyName != "raw" {
		return false
	}
	return staticEvaluator.isBuiltinStringValue(AccessExpressionObject(tag), map[*ast.Symbol]bool{})
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinStringValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	if !staticEvaluator.resolveIdentifiers {
		return false
	}
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if isIdentifierWithText(node, "String") && !IsShadowed(node, "String") {
		return true
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || resolvingAliases[symbol] {
		return false
	}
	resolvingAliases[symbol] = true
	defer delete(resolvingAliases, symbol)
	return staticEvaluator.isBuiltinStringValue(initializer, resolvingAliases)
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinArrayValue(node *ast.Node, resolvingAliases map[*ast.Symbol]bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if isIdentifierWithText(node, "Array") && !IsShadowed(node, "Array") {
		return true
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || resolvingAliases[symbol] {
		return false
	}
	resolvingAliases[symbol] = true
	defer delete(resolvingAliases, symbol)
	return staticEvaluator.isBuiltinArrayValue(initializer, resolvingAliases)
}

func (staticEvaluator *StaticStringEvaluator) evalWithTsgo(node *ast.Node) (result staticEvalResult) {
	defer func() {
		if recover() != nil {
			result = staticEvalResult{}
		}
	}()

	value := staticEvaluator.evaluator(node, node).Value
	if value == nil {
		return staticEvalResult{}
	}
	return staticEvalResult{value: value, ok: true}
}

func (staticEvaluator *StaticStringEvaluator) evaluateEntity(expr *ast.Node, location *ast.Node) evaluator.Result {
	result := staticEvaluator.evalIdentifier(expr)
	if !result.ok {
		return evaluator.Result{}
	}
	if value, ok := staticValueAsString(result.value); ok {
		return evaluator.Result{Value: value}
	}
	if !staticValueIsTsgoSafe(result.value) {
		return evaluator.Result{}
	}

	return evaluator.Result{Value: result.value}
}

func (staticEvaluator *StaticStringEvaluator) hasWrites(symbol *ast.Symbol) bool {
	if staticEvaluator.sourceFile == nil ||
		(staticEvaluator.typeChecker == nil && staticEvaluator.referenceResolver == nil) {
		return true
	}
	return staticEvaluator.referenceFlagsFor(symbol)&staticReferenceWrite != 0
}

func (staticEvaluator *StaticStringEvaluator) hasPropertyMutation(symbol *ast.Symbol) bool {
	if staticEvaluator.sourceFile == nil ||
		(staticEvaluator.typeChecker == nil && staticEvaluator.referenceResolver == nil) {
		return true
	}
	return staticEvaluator.referenceFlagsFor(symbol)&staticReferencePropertyMutation != 0
}

func (staticEvaluator *StaticStringEvaluator) referenceFlagsFor(symbol *ast.Symbol) staticReferenceFlags {
	if !staticEvaluator.referenceFlagsComputed {
		staticEvaluator.computeReferenceFlags()
	}
	return staticEvaluator.referenceFlags[symbol]
}

type staticMutationCandidate struct {
	reference *ast.Node
	access    *ast.Node
}

func (staticEvaluator *StaticStringEvaluator) computeReferenceFlags() {
	staticEvaluator.referenceFlagsComputed = true
	staticEvaluator.referenceFlags = nil
	var mutationCandidates []staticMutationCandidate

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && !IsNonReferenceIdentifier(node) {
			if IsWriteReference(node) {
				staticEvaluator.markReference(node, staticReferenceWrite)
			}
			parent := node.Parent
			mayStartAccess := parent != nil && (ast.IsAccessExpression(parent) ||
				ast.IsOuterExpression(parent, skipTransparentKinds))
			if mayStartAccess {
				access, outer := outermostAccessFromReference(node)
				if access != nil && outer.Parent != nil {
					if ast.IsAssignmentTarget(outer) {
						staticEvaluator.markReference(node, staticReferencePropertyMutation)
					} else if outer.Parent.Kind == ast.KindCallExpression {
						call := outer.Parent.AsCallExpression()
						if call.Expression == outer {
							if methodName, ok := AccessExpressionStaticName(access); ok {
								if isMutatingArrayMethod(methodName) {
									staticEvaluator.markReference(node, staticReferencePropertyMutation)
								}
							} else {
								mutationCandidates = append(mutationCandidates, staticMutationCandidate{reference: node, access: access})
							}
						}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(&staticEvaluator.sourceFile.Node)

	for _, candidate := range mutationCandidates {
		if methodName, ok := staticEvaluator.evalAccessExpressionKey(candidate.access); ok && isMutatingArrayMethod(methodName) {
			staticEvaluator.markReference(candidate.reference, staticReferencePropertyMutation)
		}
	}
}

func (staticEvaluator *StaticStringEvaluator) markReference(node *ast.Node, flag staticReferenceFlags) {
	if symbol := staticEvaluator.referenceSymbol(node); symbol != nil {
		if staticEvaluator.referenceFlags == nil {
			staticEvaluator.referenceFlags = make(map[*ast.Symbol]staticReferenceFlags, 8)
		}
		staticEvaluator.referenceFlags[symbol] |= flag
	}
}

func (staticEvaluator *StaticStringEvaluator) referenceSymbol(node *ast.Node) *ast.Symbol {
	if staticEvaluator.referenceResolver != nil {
		if symbol := staticEvaluator.referenceResolver.Resolve(node); symbol != nil {
			return symbol
		}
	}
	return GetReferenceSymbol(node, staticEvaluator.typeChecker)
}

func outermostAccessFromReference(node *ast.Node) (access *ast.Node, outer *ast.Node) {
	outer = node
	for outer.Parent != nil {
		parent := outer.Parent
		if ast.IsOuterExpression(parent, skipTransparentKinds) && parent.Expression() == outer {
			outer = parent
			continue
		}
		if ast.IsAccessExpression(parent) && AccessExpressionObject(parent) == outer {
			access = parent
			outer = parent
			continue
		}
		break
	}
	return access, outer
}

func isMutatingArrayMethod(name string) bool {
	switch name {
	case "copyWithin", "fill", "pop", "push", "reverse", "shift", "sort", "splice", "unshift":
		return true
	default:
		return false
	}
}

func staticValueIsTsgoSafe(value any) bool {
	switch value.(type) {
	case bool, staticNullValue, staticUndefinedValue, staticNumberValue, *staticStringNode, *staticObjectValue, *staticArrayValue:
		return false
	default:
		return value != nil
	}
}

func staticValueAsString(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case *staticStringNode:
		node := (*ast.Node)(value)
		switch node.Kind {
		case ast.KindStringLiteral:
			return node.AsStringLiteral().Text, true
		case ast.KindNoSubstitutionTemplateLiteral:
			return node.AsNoSubstitutionTemplateLiteral().Text, true
		}
	}
	return "", false
}

func staticValueIsString(value any) bool {
	_, ok := staticValueAsString(value)
	return ok
}

func staticValueNullish(value any) bool {
	switch value.(type) {
	case staticNullValue, staticUndefinedValue:
		return true
	default:
		return false
	}
}

func staticValueUndefined(value any) bool {
	_, ok := value.(staticUndefinedValue)
	return ok
}

// staticValueKind is the JavaScript typeof-like classification `===` needs.
// Objects and arrays compare by identity, which folding does not model, and
// tsgo's bigint representation cannot be inspected from this package; both stay
// unknown.
type staticValueKind uint8

const (
	staticKindUnknown staticValueKind = iota
	staticKindString
	staticKindNumber
	staticKindBoolean
	staticKindNull
	staticKindUndefined
)

func staticValueKindOf(value any) staticValueKind {
	switch value.(type) {
	case bool:
		return staticKindBoolean
	case staticNullValue:
		return staticKindNull
	case staticUndefinedValue:
		return staticKindUndefined
	}
	if staticValueIsString(value) {
		return staticKindString
	}
	if _, isNumber := value.(interface{ IsNaN() bool }); isNumber {
		return staticKindNumber
	}
	return staticKindUnknown
}

// staticValuesStrictEqual implements the strict equality comparison over folded
// values. Comparing across kinds is always false, which keeps `1n === 1` right
// even though bigint values themselves stay unknown.
func staticValuesStrictEqual(left any, right any) (equal bool, ok bool) {
	leftKind := staticValueKindOf(left)
	rightKind := staticValueKindOf(right)
	if leftKind == staticKindUnknown || rightKind == staticKindUnknown {
		return false, false
	}
	if leftKind != rightKind {
		return false, true
	}

	switch leftKind {
	case staticKindString:
		leftText, _ := staticValueAsString(left)
		rightText, _ := staticValueAsString(right)
		return leftText == rightText, true
	case staticKindNumber:
		leftNumber, leftOk := staticValueToNumber(left)
		rightNumber, rightOk := staticValueToNumber(right)
		if !leftOk || !rightOk {
			return false, false
		}
		return leftNumber == rightNumber, true
	case staticKindBoolean:
		leftFlag, leftOk := left.(bool)
		rightFlag, rightOk := right.(bool)
		if !leftOk || !rightOk {
			return false, false
		}
		return leftFlag == rightFlag, true
	}
	return true, true
}

func staticValueIsAggregate(value any) bool {
	switch value.(type) {
	case *staticObjectValue, *staticArrayValue:
		return true
	default:
		return false
	}
}

func staticValueTruthy(value any) (truthy bool, ok bool) {
	switch value := value.(type) {
	case string:
		return value != "", true
	case staticNumberValue:
		return value != 0 && !value.IsNaN(), true
	case *staticStringNode:
		stringValue, _ := staticValueAsString(value)
		return stringValue != "", true
	case bool:
		return value, true
	case staticNullValue, staticUndefinedValue:
		return false, true
	case *staticObjectValue, *staticArrayValue:
		return true, true
	default:
		return evaluatorBool(func() bool { return evaluator.IsTruthy(value) })
	}
}

func staticValueToString(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case staticNumberValue:
		return ecmascript.NumberToString(float64(value)), true
	case *staticStringNode:
		return staticValueAsString(value)
	case bool:
		if value {
			return "true", true
		}
		return "false", true
	case staticNullValue:
		return "null", true
	case staticUndefinedValue:
		return "undefined", true
	case *staticArrayValue:
		return staticArrayToString(value)
	case *staticObjectValue:
		if _, overridesToString := staticObjectOwnProperty(value, "toString"); overridesToString || value.prototypeSet {
			return "", false
		}
		return "[object Object]", true
	default:
		return evaluatorString(func() string { return evaluator.AnyToString(value) })
	}
}

// staticArrayToString implements `Array.prototype.join(',')`, which is what
// an array coerces to via the default ToString. Nullish elements fold to "",
// not "null"/"undefined" as a direct string conversion would give.
func staticArrayToString(value *staticArrayValue) (string, bool) {
	return staticArrayJoin(value, ",")
}

func staticArrayJoin(value *staticArrayValue, separator string) (string, bool) {
	if value == nil || value.length < 0 {
		return "", false
	}
	var builder strings.Builder
	capacity := 0
	if value.length > 1 {
		separatorCount := value.length - 1
		if len(separator) > 0 && separatorCount > maxStaticStringLength/len(separator) {
			return "", false
		}
		capacity = separatorCount * len(separator)
	}
	for i := range value.length {
		element := value.element(i)
		elementLength := 0
		switch element := element.(type) {
		case string:
			elementLength = len(element)
		case *staticStringNode:
			stringValue, _ := staticValueAsString(element)
			elementLength = len(stringValue)
		case bool:
			if element {
				elementLength = len("true")
			} else {
				elementLength = len("false")
			}
		}
		if elementLength > maxStaticStringLength-capacity {
			return "", false
		}
		capacity += elementLength
	}
	builder.Grow(capacity)
	for i := range value.length {
		element := value.element(i)
		if i > 0 {
			if len(separator) > maxStaticStringLength-builder.Len() {
				return "", false
			}
			builder.WriteString(separator)
		}
		if staticValueNullish(element) {
			continue
		}
		part, ok := staticValueToString(element)
		if !ok {
			return "", false
		}
		if len(part) > maxStaticStringLength-builder.Len() {
			return "", false
		}
		builder.WriteString(part)
	}
	return builder.String(), true
}

func evaluatorBool(fn func() bool) (value bool, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return fn(), true
}

func evaluatorString(fn func() string) (value string, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return fn(), true
}

func isIdentifierWithText(node *ast.Node, text string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == text
}
